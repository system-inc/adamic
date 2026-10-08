package main

import "strings"

// These records measure the actual 90 active syntax listeners' converging edit
// engine. The full 93-rule all-fixes cell remains checker-blocked. Retaining a
// separate measured subset prevents that coverage block from hiding a known fix
// disagreement. No bytes inside the selected protocol records are normalized.
func syntaxFixes(a WorkAnswer) WorkAnswer {
	var out strings.Builder
	fixed := false
	for _, line := range strings.SplitAfter(a.Output, "\n") {
		if strings.HasPrefix(line, "fixed\t") {
			fixed = true
			out.WriteString(line)
		} else if strings.HasPrefix(line, "rejected ") || strings.HasPrefix(line, "unconverged\t") {
			out.WriteString(line)
		}
	}
	a.Output = out.String()
	if !fixed && a.Error == "" {
		a.Error = "missing fixed-source protocol record"
	}
	return a
}
func divergences(r Receipt) []string {
	names := append([]string{}, r.Divergences...)
	if r.GoLint != nil && r.NodeLint != nil {
		a, b := syntaxFixes(*r.GoLint), syntaxFixes(*r.NodeLint)
		if status, _ := stateCode(a, b, false); status == "diverge" {
			found := false
			for _, name := range names {
				found = found || name == "lint/syntax-fixes"
			}
			if !found {
				names = append(names, "lint/syntax-fixes")
			}
		}
	}
	return names
}
func comparison(a WorkAnswer, name string) WorkAnswer {
	if name == "lint/syntax-fixes" {
		return syntaxFixes(a)
	}
	if strings.HasPrefix(name, "format/") {
		return a
	}
	a.Output = ruleAnswer(a.execution(), name).Output
	return a
}
