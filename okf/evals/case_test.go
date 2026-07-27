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
