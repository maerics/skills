// Command okf-validate checks an OKF (Open Knowledge Format) knowledge
// bundle for conformance with SPEC.md and the okf skill's house directives.
package main

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Level is the severity of a Finding.
type Level string

const (
	LevelError Level = "error"
	LevelWarn  Level = "warn"
)

// Finding is a single conformance or house-convention issue found in a
// bundle.
type Finding struct {
	Code    string `json:"code"`
	Level   Level  `json:"level"`
	Spec    string `json:"spec"`
	Path    string `json:"path"`
	Line    int    `json:"line,omitempty"`
	Message string `json:"message"`
}

// Counts summarizes a Result's findings.
type Counts struct {
	Files    int `json:"files"`
	Errors   int `json:"errors"`
	Warnings int `json:"warnings"`
}

// Result is the outcome of validating a bundle.
type Result struct {
	Bundle     string    `json:"bundle"`
	Conformant bool      `json:"conformant"`
	Counts     Counts    `json:"counts"`
	Findings   []Finding `json:"findings"`
}

func isBundleRootRelPath(relPath string) bool {
	return filepath.Dir(filepath.ToSlash(relPath)) == "."
}

// resolveBundleRoot resolves path to a bundle root directory: if
// <path>/.okf is a directory, that is the bundle root; otherwise path
// itself is treated as the bundle root.
func resolveBundleRoot(path string) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", &os.PathError{Op: "resolveBundleRoot", Path: path, Err: os.ErrInvalid}
	}
	okfDir := filepath.Join(path, ".okf")
	if st, err := os.Stat(okfDir); err == nil && st.IsDir() {
		return okfDir, nil
	}
	return path, nil
}

// walkMarkdownFiles returns every *.md file under root, as sorted,
// root-relative, forward-slash paths. Directories other than root itself
// whose base name starts with "." are skipped.
func walkMarkdownFiles(root string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if p != root && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(out)
	return out, nil
}

// splitFrontmatter splits raw into a leading YAML frontmatter block and the
// remaining body, per SPEC.md §4: a frontmatter block is delimited by a
// line that is exactly "---" as the very first line of the file, and a
// matching "---" line that closes it.
//
// present is false if the file does not start with a "---" line at all. If
// present is true but closed is false, the block was opened but never
// closed before EOF. bodyLineOffset is the 1-based line number (within
// raw) at which body begins, used for accurate Finding.Line reporting; it
// is 0 when there is no body to report against (present && !closed).
func splitFrontmatter(raw []byte) (fm, body []byte, present, closed bool, bodyLineOffset int) {
	lines := strings.Split(string(raw), "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r") != "---" {
		return nil, raw, false, false, 1
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r") == "---" {
			fmText := strings.Join(lines[1:i], "\n")
			bodyText := strings.Join(lines[i+1:], "\n")
			return []byte(fmText), []byte(bodyText), true, true, i + 2
		}
	}
	return []byte(strings.Join(lines[1:], "\n")), nil, true, false, 0
}

// fmValue is one top-level frontmatter key's parsed value.
type fmValue struct {
	Raw     string // trimmed inline scalar text; empty if the value is a block
	IsBlock bool   // true if the value continues as an indented/"-" block on following lines
	Line    int    // 1-based line number within the frontmatter block where the key appears
}

// frontmatter is the result of scanning a frontmatter block with the
// minimal parser in parseFrontmatter.
type frontmatter struct {
	Keys      map[string]fmValue
	Order     []string
	Malformed bool
}

var fmKeyRE = regexp.MustCompile(`^([A-Za-z0-9_-]+):(.*)$`)

// parseFrontmatter scans a frontmatter block (the text between the "---"
// delimiters, as returned by splitFrontmatter) using a minimal, YAML-subset
// scanner: it recognizes flat top-level "key: value" pairs whose value is
// either an inline scalar/list on the same line, or a block that continues
// on following indented or "-"-prefixed lines. It does not implement full
// YAML grammar (quoting edge cases, flow mappings, anchors, multi-line
// block scalars, etc.) — OKF frontmatter (SPEC §4.1) is always a flat set
// of key/value pairs in practice, and this tool only needs to know whether
// a key exists and whether its value is a non-empty scalar.
func parseFrontmatter(fm []byte) frontmatter {
	result := frontmatter{Keys: map[string]fmValue{}}
	lastKey := ""
	for i, line := range strings.Split(string(fm), "\n") {
		trimmed := strings.TrimRight(line, "\r")
		stripped := strings.TrimSpace(trimmed)
		if stripped == "" || strings.HasPrefix(stripped, "#") {
			continue
		}
		if !strings.HasPrefix(trimmed, " ") && !strings.HasPrefix(trimmed, "\t") && !strings.HasPrefix(trimmed, "-") {
			m := fmKeyRE.FindStringSubmatch(trimmed)
			if m == nil {
				result.Malformed = true
				continue
			}
			key := m[1]
			result.Keys[key] = fmValue{Raw: strings.TrimSpace(m[2]), Line: i + 1}
			result.Order = append(result.Order, key)
			lastKey = key
			continue
		}
		// Continuation line (indented, or a top-level "-" list item):
		// marks the previous key's value as a non-scalar block.
		if lastKey == "" {
			result.Malformed = true
			continue
		}
		fv := result.Keys[lastKey]
		fv.IsBlock = true
		result.Keys[lastKey] = fv
	}
	return result
}

func fmLineToFileLine(fmLine int) int {
	if fmLine == 0 {
		return 1
	}
	return fmLine + 1
}
