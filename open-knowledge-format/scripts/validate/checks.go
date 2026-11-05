package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// validateBundle walks every markdown file under root and runs the rule
// checks in this file against it, returning a summarized Result.
func validateBundle(root string) (Result, error) {
	files, err := walkMarkdownFiles(root)
	if err != nil {
		return Result{}, err
	}

	findings := []Finding{}
	haveRootIndex := false
	haveRootLog := false
	var rootIndexFM *frontmatter

	for _, rel := range files {
		full := filepath.Join(root, filepath.FromSlash(rel))
		raw, err := os.ReadFile(full)
		if err != nil {
			return Result{}, err
		}
		fm, body, present, closed, bodyLineOffset := splitFrontmatter(raw)
		isRoot := isBundleRootRelPath(rel)

		switch filepath.Base(rel) {
		case "index.md":
			if isRoot {
				haveRootIndex = true
			}
			findings = append(findings, checkIndexFile(rel, isRoot, fm, body, present, closed)...)
			if isRoot && present && closed {
				pf := parseFrontmatter(fm)
				rootIndexFM = &pf
			}
		case "log.md":
			if isRoot {
				haveRootLog = true
			}
			findings = append(findings, checkLogFile(rel, body, present, closed, bodyLineOffset)...)
		default:
			findings = append(findings, checkConceptFile(rel, fm, present, closed)...)
		}

		findings = append(findings, checkLinks(root, rel, body, bodyLineOffset)...)
	}

	if !haveRootIndex {
		findings = append(findings, Finding{
			Code: "root-index-missing", Level: LevelWarn, Spec: "SKILL.md bootstrap convention",
			Path: "index.md", Message: "bundle root has no index.md",
		})
	} else if rootIndexFM != nil {
		findings = append(findings, checkRootIndexManifest(*rootIndexFM)...)
	}
	if !haveRootLog {
		findings = append(findings, Finding{
			Code: "log-missing", Level: LevelWarn, Spec: "SKILL.md bootstrap convention",
			Path: "log.md", Message: "bundle root has no log.md",
		})
	}

	sortFindings(findings)

	res := Result{Bundle: root, Findings: findings}
	res.Counts.Files = len(files)
	for _, f := range findings {
		if f.Level == LevelError {
			res.Counts.Errors++
		} else {
			res.Counts.Warnings++
		}
	}
	res.Conformant = res.Counts.Errors == 0
	return res, nil
}

func sortFindings(f []Finding) {
	sort.SliceStable(f, func(i, j int) bool {
		if f[i].Path != f[j].Path {
			return f[i].Path < f[j].Path
		}
		if f[i].Line != f[j].Line {
			return f[i].Line < f[j].Line
		}
		return f[i].Code < f[j].Code
	})
}

func exitCode(r Result, strict bool) int {
	if r.Counts.Errors > 0 {
		return 1
	}
	if strict && r.Counts.Warnings > 0 {
		return 2
	}
	return 0
}

// checkConceptFile applies SPEC §9.1/§9.2's frontmatter/type rules, which
// apply only to non-reserved (concept) markdown files.
func checkConceptFile(rel string, fm []byte, present, closed bool) []Finding {
	if !present {
		return []Finding{{
			Code: "frontmatter-missing", Level: LevelError, Spec: "SPEC §4.1, §9.1",
			Path: rel, Line: 1, Message: "no frontmatter block found at the start of the file",
		}}
	}
	if !closed {
		return []Finding{{
			Code: "frontmatter-invalid", Level: LevelError, Spec: "SPEC §4.1, §9.1",
			Path: rel, Line: 1, Message: "frontmatter block is never closed with a matching '---'",
		}}
	}

	var out []Finding
	pf := parseFrontmatter(fm)
	if pf.Malformed {
		out = append(out, Finding{
			Code: "frontmatter-invalid", Level: LevelError, Spec: "SPEC §4.1, §9.1",
			Path: rel, Line: 1, Message: "frontmatter block contains a line that could not be parsed",
		})
	}

	typeVal, hasType := pf.Keys["type"]
	switch {
	case !hasType || (!typeVal.IsBlock && strings.TrimSpace(typeVal.Raw) == ""):
		out = append(out, Finding{
			Code: "type-missing", Level: LevelError, Spec: "SPEC §4.1, §9.2",
			Path: rel, Line: fmLineToFileLine(typeVal.Line), Message: "frontmatter has no non-empty 'type' field",
		})
	case typeVal.IsBlock:
		out = append(out, Finding{
			Code: "type-invalid", Level: LevelError, Spec: "SPEC §4.1, §9.2",
			Path: rel, Line: fmLineToFileLine(typeVal.Line), Message: "'type' must be a scalar string, not a list or mapping",
		})
	}
	return out
}

// checkIndexFile applies SPEC §6/§9.3/§11's index.md rules: frontmatter is
// permitted only on the bundle-root index.md, and the body should be
// organized into heading sections.
func checkIndexFile(rel string, isRoot bool, fm, body []byte, present, closed bool) []Finding {
	var out []Finding
	if present {
		switch {
		case !isRoot:
			out = append(out, Finding{
				Code: "index-frontmatter", Level: LevelError, Spec: "SPEC §6, §9.3, §11",
				Path: rel, Line: 1, Message: "index.md must not contain frontmatter except at the bundle root",
			})
		case !closed:
			out = append(out, Finding{
				Code: "frontmatter-invalid", Level: LevelError, Spec: "SPEC §4.1, §9.1",
				Path: rel, Line: 1, Message: "frontmatter block is never closed with a matching '---'",
			})
		default:
			if pf := parseFrontmatter(fm); pf.Malformed {
				out = append(out, Finding{
					Code: "frontmatter-invalid", Level: LevelError, Spec: "SPEC §4.1, §9.1",
					Path: rel, Line: 1, Message: "frontmatter block contains a line that could not be parsed",
				})
			}
		}
	}
	if !hasHeadingSection(body) {
		out = append(out, Finding{
			Code: "index-no-sections", Level: LevelWarn, Spec: "SPEC §6 (descriptive)",
			Path: rel, Message: "index.md body has no '#' heading sections",
		})
	}
	return out
}

// checkLogFile applies SPEC §7/§9.3's log.md rules: no frontmatter (an
// extension beyond literal spec text, applied for consistency with
// index.md — see design notes in the implementation plan), and ISO 8601
// date headings.
func checkLogFile(rel string, body []byte, present, closed bool, bodyLineOffset int) []Finding {
	var out []Finding
	if present {
		out = append(out, Finding{
			Code: "log-frontmatter", Level: LevelError, Spec: "SPEC §7, §9.3 (extension)",
			Path: rel, Line: 1, Message: "log.md must not contain a frontmatter block",
		})
		if !closed {
			return out
		}
	}

	dates, dateFindings := scanLogDateHeadings(rel, body, bodyLineOffset)
	out = append(out, dateFindings...)
	if !isNonIncreasing(dates) {
		out = append(out, Finding{
			Code: "log-not-newest-first", Level: LevelWarn, Spec: "SPEC §7 (descriptive)",
			Path: rel, Message: "log.md date headings are not in newest-first order",
		})
	}
	return out
}

var headingRE = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)

func scanLogDateHeadings(rel string, body []byte, bodyLineOffset int) ([]time.Time, []Finding) {
	if body == nil {
		return nil, nil
	}
	var dates []time.Time
	var findings []Finding
	for i, line := range strings.Split(string(body), "\n") {
		m := headingRE.FindStringSubmatch(strings.TrimRight(line, "\r"))
		if m == nil || m[1] != "##" {
			continue
		}
		text := strings.TrimSpace(m[2])
		t, err := time.Parse("2006-01-02", text)
		if err != nil {
			findings = append(findings, Finding{
				Code: "log-bad-date-heading", Level: LevelError, Spec: "SPEC §7, §9.3",
				Path: rel, Line: bodyLineOffset + i,
				Message: fmt.Sprintf("log.md date heading %q is not a valid ISO 8601 YYYY-MM-DD date", text),
			})
			continue
		}
		dates = append(dates, t)
	}
	return dates, findings
}

func isNonIncreasing(dates []time.Time) bool {
	for i := 1; i < len(dates); i++ {
		if dates[i].After(dates[i-1]) {
			return false
		}
	}
	return true
}

func hasHeadingSection(body []byte) bool {
	if body == nil {
		return false
	}
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			return true
		}
	}
	return false
}

// checkRootIndexManifest applies the SKILL.md house convention that a
// bundle-root index.md's frontmatter (when present) should include the
// manifest fields that let an agent distinguish co-located bundles without
// opening each one.
func checkRootIndexManifest(pf frontmatter) []Finding {
	var out []Finding
	for _, key := range []string{"okf_version", "type", "title", "tags"} {
		v, ok := pf.Keys[key]
		empty := !ok || (!v.IsBlock && strings.TrimSpace(v.Raw) == "")
		if empty {
			out = append(out, Finding{
				Code: "root-index-missing-field", Level: LevelWarn, Spec: "SKILL.md manifest convention",
				Path: "index.md", Message: fmt.Sprintf("root index.md frontmatter is missing recommended field %q", key),
			})
		}
	}
	return out
}

// --- Link scanning (SPEC §5) ---

type linkKind int

const (
	linkRelative linkKind = iota
	linkAbsolute
	linkExternal
	linkAnchor
)

var linkRE = regexp.MustCompile(`\[[^\]]*\]\(([^)\s]+)(?:\s+"[^"]*")?\)`)

func classifyLinkTarget(target string) linkKind {
	switch {
	case strings.HasPrefix(target, "#"):
		return linkAnchor
	case strings.Contains(target, "://"):
		return linkExternal
	case strings.HasPrefix(target, "mailto:"):
		return linkExternal
	case strings.HasPrefix(target, "//"):
		return linkExternal
	case strings.HasPrefix(target, "/"):
		return linkAbsolute
	default:
		return linkRelative
	}
}

// linkTargetExists resolves target (already classified as absolute or
// relative) against bundleRoot / the linking file's directory and reports
// whether it exists on disk.
func linkTargetExists(bundleRoot, rel, target string, kind linkKind) bool {
	clean := target
	if idx := strings.IndexAny(clean, "#?"); idx >= 0 {
		clean = clean[:idx]
	}
	if clean == "" {
		return true
	}
	var full string
	if kind == linkAbsolute {
		full = filepath.Join(bundleRoot, filepath.FromSlash(strings.TrimPrefix(clean, "/")))
	} else {
		full = filepath.Join(bundleRoot, filepath.FromSlash(filepath.Dir(rel)), filepath.FromSlash(clean))
	}
	_, err := os.Stat(full)
	return err == nil
}

// checkLinks scans a file's body for markdown links and applies SPEC
// §5.1/§5.3 and the skill's relative-link preference. It is a line-based
// regex scan, not a full CommonMark parser, so it can miss links split
// across lines or ones appearing inside fenced code blocks — a known,
// documented limitation acceptable for this tool's purpose.
func checkLinks(bundleRoot, rel string, body []byte, bodyLineOffset int) []Finding {
	if body == nil {
		return nil
	}
	var out []Finding
	for i, line := range strings.Split(string(body), "\n") {
		for _, m := range linkRE.FindAllStringSubmatch(line, -1) {
			target := m[1]
			kind := classifyLinkTarget(target)
			if kind == linkExternal || kind == linkAnchor {
				continue
			}
			fileLine := bodyLineOffset + i
			if kind == linkAbsolute {
				out = append(out, Finding{
					Code: "absolute-link", Level: LevelWarn, Spec: "SPEC §5.1 vs. SKILL.md",
					Path: rel, Line: fileLine,
					Message: fmt.Sprintf("link target %q uses the absolute bundle-relative form; the skill prefers relative links", target),
				})
			}
			if !linkTargetExists(bundleRoot, rel, target, kind) {
				out = append(out, Finding{
					Code: "broken-link", Level: LevelWarn, Spec: "SPEC §5.3",
					Path: rel, Line: fileLine,
					Message: fmt.Sprintf("link target %q does not resolve to a file in the bundle", target),
				})
			}
		}
	}
	return out
}
