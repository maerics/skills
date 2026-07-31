// Command okf-evals runs behavioral scenarios against the okf skill: each
// case gives an agent a fixture repo and a task prompt, then grades the
// result against the OKF validator (SPEC conformance) and a rubric (agent
// judgment calls the validator can't check, graded by a second agent).
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// evalsDir is this source file's own directory, resolved at compile time so
// case discovery works regardless of the caller's cwd (same trick the
// shell validate wrapper uses, but free in Go via runtime.Caller).
var evalsDir = func() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Dir(file)
}()

var (
	skillDir    = filepath.Dir(evalsDir)           // okf/
	casesDir    = filepath.Join(evalsDir, "cases") // okf/evals/cases/
	validateBin = filepath.Join(skillDir, "scripts", "validate")
)

func main() {
	agent := flag.String("agent", "claude", "agent to invoke: claude or codex")
	claudeBin := flag.String("claude", "claude", "claude CLI binary to invoke")
	codexBin := flag.String("codex", "codex", "codex CLI binary to invoke")
	only := flag.String("case", "", "comma-separated case names to run (default: all)")
	permMode := flag.String("permission-mode", "acceptEdits", "--permission-mode passed to the agent under test")
	noJudge := flag.Bool("no-judge", false, "skip LLM-judge grading, report validator conformance only")
	keep := flag.Bool("keep", false, "keep each case's working directory instead of deleting it")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [flags]\n\n"+
			"Run behavioral evals for the okf skill. Each case under cases/ gets a\n"+
			"fresh copy of its fixture, an agent run against its prompt.txt, an OKF\n"+
			"validator check, and (unless -no-judge) a second agent grading the\n"+
			"result against its rubric.md. Prints a JSON report and exits nonzero\n"+
			"if any case fails.\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if *agent != "claude" && *agent != "codex" {
		fmt.Fprintf(os.Stderr, "okf-evals: unknown agent %q (want claude or codex)\n", *agent)
		os.Exit(1)
	}

	names, err := caseNames(*only)
	if err != nil {
		fmt.Fprintln(os.Stderr, "okf-evals:", err)
		os.Exit(1)
	}

	opts := runOpts{agent: *agent, claudeBin: *claudeBin, codexBin: *codexBin, permMode: *permMode, judge: !*noJudge, keep: *keep}
	results := make([]caseResult, 0, len(names))
	allPass := true
	for _, name := range names {
		r := runCase(name, opts)
		results = append(results, r)
		allPass = allPass && r.Pass
	}

	out, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stderr, "okf-evals:", err)
		os.Exit(1)
	}
	printJSON(out)

	if !allPass {
		os.Exit(1)
	}
}

// caseNames lists case directories under casesDir, filtered by a
// comma-separated allowlist when csv is non-empty.
func caseNames(csv string) ([]string, error) {
	entries, err := os.ReadDir(casesDir)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", casesDir, err)
	}
	var want map[string]bool
	if csv != "" {
		want = map[string]bool{}
		for _, n := range strings.Split(csv, ",") {
			want[strings.TrimSpace(n)] = true
		}
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if want == nil || want[e.Name()] {
			names = append(names, e.Name())
		}
	}
	if want != nil && len(names) != len(want) {
		return nil, fmt.Errorf("-case named an unknown case (have: %v)", names)
	}
	return names, nil
}

// printJSON pipes b through jq for readability when available, falling back
// to raw stdout, matching this repo's CLI conventions.
func printJSON(b []byte) {
	jq, err := exec.LookPath("jq")
	if err != nil {
		os.Stdout.Write(append(b, '\n'))
		return
	}
	cmd := exec.Command(jq, ".")
	cmd.Stdin = strings.NewReader(string(b))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if cmd.Run() != nil {
		os.Stdout.Write(append(b, '\n'))
	}
}

// copyTree recursively copies src into dst, creating dst.
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}
