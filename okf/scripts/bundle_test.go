package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitFrontmatter(t *testing.T) {
	cases := []struct {
		name           string
		raw            string
		wantPresent    bool
		wantClosed     bool
		wantFM         string
		wantBody       string
		wantLineOffset int
	}{
		{
			name:           "no frontmatter",
			raw:            "# Hello\n\nBody text.\n",
			wantPresent:    false,
			wantClosed:     false,
			wantBody:       "# Hello\n\nBody text.\n",
			wantLineOffset: 1,
		},
		{
			name:           "well formed",
			raw:            "---\ntype: Thing\n---\n\nBody.\n",
			wantPresent:    true,
			wantClosed:     true,
			wantFM:         "type: Thing",
			wantBody:       "\nBody.\n",
			wantLineOffset: 4,
		},
		{
			name:        "unclosed",
			raw:         "---\ntype: Thing\n\nBody with no closing delimiter.\n",
			wantPresent: true,
			wantClosed:  false,
		},
		{
			name:           "CRLF line endings",
			raw:            "---\r\ntype: Thing\r\n---\r\n\r\nBody.\r\n",
			wantPresent:    true,
			wantClosed:     true,
			wantFM:         "type: Thing\r",
			wantLineOffset: 4,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fm, body, present, closed, lineOffset := splitFrontmatter([]byte(tc.raw))
			if present != tc.wantPresent {
				t.Fatalf("present = %v, want %v", present, tc.wantPresent)
			}
			if closed != tc.wantClosed {
				t.Fatalf("closed = %v, want %v", closed, tc.wantClosed)
			}
			if tc.wantFM != "" && string(fm) != tc.wantFM {
				t.Fatalf("fm = %q, want %q", fm, tc.wantFM)
			}
			if tc.wantBody != "" && string(body) != tc.wantBody {
				t.Fatalf("body = %q, want %q", body, tc.wantBody)
			}
			if tc.wantLineOffset != 0 && lineOffset != tc.wantLineOffset {
				t.Fatalf("lineOffset = %d, want %d", lineOffset, tc.wantLineOffset)
			}
		})
	}
}

func TestParseFrontmatter(t *testing.T) {
	t.Run("flat scalars", func(t *testing.T) {
		pf := parseFrontmatter([]byte("type: Thing\ntitle: A Title\n"))
		if pf.Malformed {
			t.Fatalf("unexpected malformed")
		}
		if pf.Keys["type"].Raw != "Thing" {
			t.Fatalf("type = %q, want Thing", pf.Keys["type"].Raw)
		}
		if pf.Keys["type"].IsBlock {
			t.Fatalf("type should not be a block")
		}
	})

	t.Run("inline list", func(t *testing.T) {
		pf := parseFrontmatter([]byte("type: Thing\ntags: [a, b, c]\n"))
		if pf.Keys["tags"].IsBlock {
			t.Fatalf("inline list should not be marked as block")
		}
		if pf.Keys["tags"].Raw != "[a, b, c]" {
			t.Fatalf("tags = %q", pf.Keys["tags"].Raw)
		}
	})

	t.Run("block list", func(t *testing.T) {
		pf := parseFrontmatter([]byte("type:\n  - a\n  - b\ntitle: X\n"))
		if !pf.Keys["type"].IsBlock {
			t.Fatalf("type should be marked as block")
		}
	})

	t.Run("empty scalar", func(t *testing.T) {
		pf := parseFrontmatter([]byte("type: \ntitle: X\n"))
		if pf.Keys["type"].Raw != "" || pf.Keys["type"].IsBlock {
			t.Fatalf("type should be an empty, non-block scalar")
		}
	})

	t.Run("missing key", func(t *testing.T) {
		pf := parseFrontmatter([]byte("title: X\n"))
		if _, ok := pf.Keys["type"]; ok {
			t.Fatalf("type should be absent")
		}
	})

	t.Run("malformed line", func(t *testing.T) {
		pf := parseFrontmatter([]byte("type: Thing\nnot a key value pair\n"))
		if !pf.Malformed {
			t.Fatalf("expected malformed")
		}
	})

	t.Run("comments and blanks ignored", func(t *testing.T) {
		pf := parseFrontmatter([]byte("# comment\n\ntype: Thing\n"))
		if pf.Malformed {
			t.Fatalf("unexpected malformed")
		}
		if pf.Keys["type"].Raw != "Thing" {
			t.Fatalf("type = %q", pf.Keys["type"].Raw)
		}
	})
}

func TestResolveBundleRoot(t *testing.T) {
	tmp := t.TempDir()

	t.Run("no .okf subdirectory", func(t *testing.T) {
		root, err := resolveBundleRoot(tmp)
		if err != nil {
			t.Fatal(err)
		}
		if root != tmp {
			t.Fatalf("root = %q, want %q", root, tmp)
		}
	})

	t.Run("with .okf subdirectory", func(t *testing.T) {
		container := t.TempDir()
		okfDir := filepath.Join(container, ".okf")
		if err := os.Mkdir(okfDir, 0o755); err != nil {
			t.Fatal(err)
		}
		root, err := resolveBundleRoot(container)
		if err != nil {
			t.Fatal(err)
		}
		if root != okfDir {
			t.Fatalf("root = %q, want %q", root, okfDir)
		}
	})

	t.Run("nonexistent path", func(t *testing.T) {
		if _, err := resolveBundleRoot(filepath.Join(tmp, "does-not-exist")); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestWalkMarkdownFiles(t *testing.T) {
	tmp := t.TempDir()
	mustWrite(t, filepath.Join(tmp, "index.md"), "")
	mustWrite(t, filepath.Join(tmp, "a.md"), "")
	mustWrite(t, filepath.Join(tmp, "notes.txt"), "")
	if err := os.MkdirAll(filepath.Join(tmp, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(tmp, "sub", "b.md"), "")
	if err := os.MkdirAll(filepath.Join(tmp, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	mustWrite(t, filepath.Join(tmp, ".git", "ignored.md"), "")

	files, err := walkMarkdownFiles(tmp)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a.md", "index.md", "sub/b.md"}
	if len(files) != len(want) {
		t.Fatalf("files = %v, want %v", files, want)
	}
	for i := range want {
		if files[i] != want[i] {
			t.Fatalf("files = %v, want %v", files, want)
		}
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
