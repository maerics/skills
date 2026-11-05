package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCaseStructure checks the cheap, deterministic part of each case: the
// files an eval run depends on actually exist. It does not invoke claude.
func TestCaseStructure(t *testing.T) {
	names, err := caseNames("")
	if err != nil {
		t.Fatal(err)
	}
	if len(names) == 0 {
		t.Fatal("no cases found under cases/")
	}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(casesDir, name)
			for _, req := range []string{"prompt.txt", "rubric.md", "fixture"} {
				if _, err := os.Stat(filepath.Join(dir, req)); err != nil {
					t.Errorf("missing %s: %s", req, err)
				}
			}
			prompt, err := os.ReadFile(filepath.Join(dir, "prompt.txt"))
			if err != nil {
				t.Fatal(err)
			}
			if len(prompt) == 0 {
				t.Error("prompt.txt is empty")
			}
		})
	}
}

func TestInstallSkillUsesProviderPath(t *testing.T) {
	for _, agent := range []string{"claude", "codex"} {
		t.Run(agent, func(t *testing.T) {
			work := t.TempDir()
			if err := installSkill(work, agent); err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(work, "."+agent, "skills", "open-knowledge-format")
			info, err := os.Lstat(path)
			if err != nil {
				t.Fatal(err)
			}
			if info.Mode()&os.ModeSymlink == 0 {
				t.Fatalf("%s is not a symlink", path)
			}
		})
	}
}

func TestRunCodexCapturesLastMessage(t *testing.T) {
	for _, tc := range []struct {
		name    string
		judge   bool
		sandbox string
	}{
		{name: "agent", sandbox: "workspace-write"},
		{name: "judge", judge: true, sandbox: "read-only"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bin := filepath.Join(t.TempDir(), "codex")
			script := "#!/bin/sh\n" +
				"[ \"$1\" = exec ] || exit 10\n" +
				"last= sandbox=\n" +
				"while [ $# -gt 0 ]; do\n" +
				"  case \"$1\" in\n" +
				"    --output-last-message) last=\"$2\"; shift 2;;\n" +
				"    --sandbox) sandbox=\"$2\"; shift 2;;\n" +
				"    *) shift;;\n" +
				"  esac\n" +
				"done\n" +
				"[ \"$sandbox\" = \"" + tc.sandbox + "\" ] || exit 11\n" +
				"printf 'final response' > \"$last\"\n"
			if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
				t.Fatal(err)
			}

			turn, err := runCodex(bin, t.TempDir(), "test prompt", tc.judge)
			if err != nil {
				t.Fatal(err)
			}
			if turn.Result != "final response" {
				t.Fatalf("result = %q", turn.Result)
			}
			if turn.CostUSD != nil {
				t.Fatalf("CostUSD = %v, want nil", *turn.CostUSD)
			}
		})
	}
}
