package main

import (
	"testing"
	"time"
)

func TestClassifyLinkTarget(t *testing.T) {
	cases := map[string]linkKind{
		"./other.md":              linkRelative,
		"../parent.md":            linkRelative,
		"sibling.md":              linkRelative,
		"/tables/orders.md":       linkAbsolute,
		"https://example.com":     linkExternal,
		"http://example.com/x.md": linkExternal,
		"mailto:someone@x.com":    linkExternal,
		"//cdn.example.com/x.js":  linkExternal,
		"#section":                linkAnchor,
	}
	for target, want := range cases {
		if got := classifyLinkTarget(target); got != want {
			t.Errorf("classifyLinkTarget(%q) = %v, want %v", target, got, want)
		}
	}
}

func TestScanLogDateHeadings(t *testing.T) {
	t.Run("valid dates", func(t *testing.T) {
		body := []byte("# Log\n\n## 2026-02-01\n* second\n\n## 2026-01-01\n* first\n")
		dates, findings := scanLogDateHeadings("log.md", body, 1)
		if len(findings) != 0 {
			t.Fatalf("unexpected findings: %v", findings)
		}
		if len(dates) != 2 {
			t.Fatalf("dates = %v, want 2 entries", dates)
		}
		if !isNonIncreasing(dates) {
			t.Fatalf("expected newest-first order to hold")
		}
	})

	t.Run("bad date", func(t *testing.T) {
		body := []byte("# Log\n\n## 2026/01/01\n* bad\n")
		_, findings := scanLogDateHeadings("log.md", body, 1)
		if len(findings) != 1 || findings[0].Code != "log-bad-date-heading" {
			t.Fatalf("findings = %+v, want one log-bad-date-heading", findings)
		}
	})

	t.Run("out of order", func(t *testing.T) {
		dates := []time.Time{mustDate(t, "2026-01-01"), mustDate(t, "2026-02-01")}
		if isNonIncreasing(dates) {
			t.Fatalf("expected out-of-order dates to fail the newest-first check")
		}
	})
}

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse("2006-01-02", s)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

func TestHasHeadingSection(t *testing.T) {
	if hasHeadingSection([]byte("just a paragraph\n")) {
		t.Fatalf("expected no heading section")
	}
	if !hasHeadingSection([]byte("intro\n\n# Section\n\ntext\n")) {
		t.Fatalf("expected a heading section")
	}
}

func TestExitCode(t *testing.T) {
	cases := []struct {
		name     string
		errors   int
		warnings int
		strict   bool
		want     int
	}{
		{"clean", 0, 0, false, 0},
		{"clean strict", 0, 0, true, 0},
		{"warnings only, non-strict", 0, 2, false, 0},
		{"warnings only, strict", 0, 2, true, 2},
		{"errors, non-strict", 1, 0, false, 1},
		{"errors, strict", 1, 3, true, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := Result{Counts: Counts{Errors: tc.errors, Warnings: tc.warnings}}
			if got := exitCode(r, tc.strict); got != tc.want {
				t.Fatalf("exitCode = %d, want %d", got, tc.want)
			}
		})
	}
}

// findingCodes returns the set of finding codes present in findings.
func findingCodes(findings []Finding) map[string]int {
	m := map[string]int{}
	for _, f := range findings {
		m[f.Code]++
	}
	return m
}

func TestValidateBundleFixtures(t *testing.T) {
	cases := []struct {
		dir          string
		wantErrors   int
		wantWarnings int
		wantCodes    []string
	}{
		{"testdata/valid-minimal", 0, 0, nil},
		{"testdata/valid-nested", 0, 0, nil},
		{"testdata/err-frontmatter-missing", 1, 0, []string{"frontmatter-missing"}},
		{"testdata/err-frontmatter-invalid", 1, 0, []string{"frontmatter-invalid"}},
		{"testdata/err-type-missing", 1, 0, []string{"type-missing"}},
		{"testdata/err-type-invalid", 1, 0, []string{"type-invalid"}},
		{"testdata/err-index-frontmatter-nonroot", 1, 0, []string{"index-frontmatter"}},
		{"testdata/err-log-frontmatter", 1, 0, []string{"log-frontmatter"}},
		{"testdata/err-log-bad-date", 1, 0, []string{"log-bad-date-heading"}},
		{"testdata/warn-absolute-link", 0, 1, []string{"absolute-link"}},
		{"testdata/warn-broken-link", 0, 1, []string{"broken-link"}},
		{"testdata/warn-root-index-missing-fields", 0, 2, []string{"root-index-missing-field"}},
		{"testdata/warn-root-index-missing", 0, 1, []string{"root-index-missing"}},
		{"testdata/warn-log-missing", 0, 1, []string{"log-missing"}},
		{"testdata/warn-index-no-sections", 0, 1, []string{"index-no-sections"}},
		{"testdata/warn-log-not-newest-first", 0, 1, []string{"log-not-newest-first"}},
	}

	for _, tc := range cases {
		t.Run(tc.dir, func(t *testing.T) {
			res, err := validateBundle(tc.dir)
			if err != nil {
				t.Fatal(err)
			}
			if res.Counts.Errors != tc.wantErrors {
				t.Errorf("errors = %d, want %d (findings: %+v)", res.Counts.Errors, tc.wantErrors, res.Findings)
			}
			if res.Counts.Warnings != tc.wantWarnings {
				t.Errorf("warnings = %d, want %d (findings: %+v)", res.Counts.Warnings, tc.wantWarnings, res.Findings)
			}
			if res.Conformant != (res.Counts.Errors == 0) {
				t.Errorf("conformant = %v, inconsistent with errors = %d", res.Conformant, res.Counts.Errors)
			}
			codes := findingCodes(res.Findings)
			for _, want := range tc.wantCodes {
				if codes[want] == 0 {
					t.Errorf("expected finding code %q, got codes %v", want, codes)
				}
			}
		})
	}
}

func TestValidateBundleStrictWarningsOnly(t *testing.T) {
	res, err := validateBundle("testdata/strict-warnings-only")
	if err != nil {
		t.Fatal(err)
	}
	if res.Counts.Errors != 0 || res.Counts.Warnings == 0 {
		t.Fatalf("expected zero errors and at least one warning, got %+v", res.Counts)
	}
	if got := exitCode(res, false); got != 0 {
		t.Fatalf("exitCode(strict=false) = %d, want 0", got)
	}
	if got := exitCode(res, true); got != 2 {
		t.Fatalf("exitCode(strict=true) = %d, want 2", got)
	}
}
