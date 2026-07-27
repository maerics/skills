package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type runOpts struct {
	claudeBin string
	permMode  string
	judge     bool
	keep      bool
}

type caseResult struct {
	Case         string          `json:"case"`
	Pass         bool            `json:"pass"`
	CostUSD      float64         `json:"cost_usd"`
	AgentMessage string          `json:"agent_message,omitempty"`
	Validate     json.RawMessage `json:"validate,omitempty"`
	ValidatePass bool            `json:"validate_pass"`
	JudgePass    *bool           `json:"judge_pass,omitempty"`
	JudgeReasons []string        `json:"judge_reasons,omitempty"`
	WorkDir      string          `json:"work_dir,omitempty"`
	Error        string          `json:"error,omitempty"`
}

// agentTurn is the subset of `claude -p --output-format json` we need.
type agentTurn struct {
	IsError      bool    `json:"is_error"`
	Result       string  `json:"result"`
	TotalCostUSD float64 `json:"total_cost_usd"`
}

func runCase(name string, opts runOpts) caseResult {
	res := caseResult{Case: name}
	dir := filepath.Join(casesDir, name)

	prompt, err := os.ReadFile(filepath.Join(dir, "prompt.txt"))
	if err != nil {
		res.Error = err.Error()
		return res
	}

	work, err := os.MkdirTemp("", "okf-eval-"+name+"-")
	if err != nil {
		res.Error = err.Error()
		return res
	}
	if opts.keep {
		res.WorkDir = work
	} else {
		defer os.RemoveAll(work)
	}

	if err := copyTree(filepath.Join(dir, "fixture"), work); err != nil {
		res.Error = fmt.Sprintf("copying fixture: %s", err)
		return res
	}
	if err := installSkill(work); err != nil {
		res.Error = fmt.Sprintf("installing skill: %s", err)
		return res
	}
	if err := runSetup(dir, work); err != nil {
		res.Error = fmt.Sprintf("running setup: %s", err)
		return res
	}

	turn, err := runAgent(opts.claudeBin, work, strings.TrimSpace(string(prompt)), opts.permMode)
	if err != nil {
		res.Error = fmt.Sprintf("running agent: %s", err)
		return res
	}
	res.CostUSD = turn.TotalCostUSD
	res.AgentMessage = turn.Result
	if turn.IsError {
		res.Error = "agent turn reported is_error"
		return res
	}

	bundle := filepath.Join(work, ".okf")
	report, conformant, err := runValidate(bundle)
	if err != nil {
		res.Error = fmt.Sprintf("running validator: %s", err)
		return res
	}
	res.Validate = report
	res.ValidatePass = conformant

	res.Pass = res.ValidatePass
	if opts.judge {
		rubric, err := os.ReadFile(filepath.Join(dir, "rubric.md"))
		if err != nil {
			res.Error = fmt.Sprintf("reading rubric: %s", err)
			return res
		}
		pass, reasons, cost, err := runJudge(opts.claudeBin, work, string(rubric), turn.Result)
		if err != nil {
			res.Error = fmt.Sprintf("running judge: %s", err)
			return res
		}
		res.CostUSD += cost
		res.JudgePass = &pass
		res.JudgeReasons = reasons
		res.Pass = res.Pass && pass
	}

	return res
}

// installSkill makes the okf skill available to the agent under test as a
// project-scoped skill, the same way it'd be found in a real repo.
func installSkill(work string) error {
	skills := filepath.Join(work, ".claude", "skills")
	if err := os.MkdirAll(skills, 0o755); err != nil {
		return err
	}
	return os.Symlink(skillDir, filepath.Join(skills, "okf"))
}

// runSetup runs caseDir/setup.sh, if present, with work as its cwd. Cases
// use this to build git history (e.g. commit dates the recency gate reads)
// that can't be expressed as static fixture files.
func runSetup(caseDir, work string) error {
	script := filepath.Join(caseDir, "setup.sh")
	if _, err := os.Stat(script); err != nil {
		return nil
	}
	cmd := exec.Command("sh", script)
	cmd.Dir = work
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, out)
	}
	return nil
}

func runAgent(claudeBin, dir, prompt, permMode string) (agentTurn, error) {
	cmd := exec.Command(claudeBin, "-p", prompt,
		"--output-format", "json",
		"--permission-mode", permMode)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return agentTurn{}, fmt.Errorf("%w (stderr unavailable via Output)", err)
	}
	var turn agentTurn
	if err := json.Unmarshal(out, &turn); err != nil {
		return agentTurn{}, fmt.Errorf("parsing agent output: %w", err)
	}
	return turn, nil
}

// runJudge asks a fresh agent to grade the case's rubric against the
// resulting working directory and the tested agent's final message.
func runJudge(claudeBin, work, rubric, agentMessage string) (pass bool, reasons []string, cost float64, err error) {
	prompt := fmt.Sprintf(`You are grading another agent's work against a rubric. The
working directory (also available to you as a tool) contains the result.
Inspect it as needed, then answer.

# Rubric
%s

# Agent's final message
%s

Respond with ONLY a JSON object, no other text: {"pass": bool, "reasons": [string, ...]}`,
		rubric, agentMessage)

	turn, err := runAgent(claudeBin, work, prompt, "acceptEdits")
	if err != nil {
		return false, nil, 0, err
	}
	var verdict struct {
		Pass    bool     `json:"pass"`
		Reasons []string `json:"reasons"`
	}
	text := strings.TrimSpace(turn.Result)
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &verdict); err != nil {
		return false, nil, turn.TotalCostUSD, fmt.Errorf("parsing judge verdict %q: %w", turn.Result, err)
	}
	return verdict.Pass, verdict.Reasons, turn.TotalCostUSD, nil
}

func runValidate(bundle string) (report json.RawMessage, conformant bool, err error) {
	out, runErr := exec.Command(validateBin, "-strict", bundle).Output()
	if len(out) == 0 {
		return nil, false, runErr
	}
	var parsed struct {
		Counts struct {
			Errors   int `json:"errors"`
			Warnings int `json:"warnings"`
		} `json:"counts"`
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		return json.RawMessage(out), false, fmt.Errorf("parsing validator output: %w", err)
	}
	// -strict is always passed above, so warnings fail the case too.
	pass := parsed.Counts.Errors == 0 && parsed.Counts.Warnings == 0
	return json.RawMessage(out), pass, nil
}
