package main

import (
	"io"
	"os"
	"testing"
)

// TestPrintResultFallback verifies that printResult falls back to a plain
// stdout print when jq is unavailable, per the house convention of piping
// through jq but never hard-failing if it's missing.
func TestPrintResultFallback(t *testing.T) {
	t.Setenv("PATH", "")

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	input := []byte(`{"conformant":true}` + "\n")
	printResult(input)

	w.Close()
	got, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(input) {
		t.Fatalf("printResult fallback output = %q, want %q", got, input)
	}
}
