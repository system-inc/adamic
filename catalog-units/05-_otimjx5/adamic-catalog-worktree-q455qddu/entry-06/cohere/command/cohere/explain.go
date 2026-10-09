package main

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"

	"github.com/system-inc/cohere/internal/types/program"
)

// printExplanation renders what every rule did on one file.
//
// The counterpart to --timing. Timing says a rule is expensive across the tree; this says what it
// did on the file where it was expensive, and why it did or did not run at all.
//
// Rules that did nothing are printed rather than omitted, which is the whole reason this exists. A
// rule absent from the output and a rule that looked and found nothing are indistinguishable, and
// four rules in the gate this tool replaces were silently dead for months underneath exactly that.
func printExplanation(out io.Writer, explanation *program.Explanation) {
	fmt.Fprintf(out, "\nexplain: %s\n", explanation.FileName)

	if explanation.FileIgnored {
		fmt.Fprintf(out, "  no rules ran: ignorePattern %q excludes this file\n", explanation.IgnoredBy)
		return
	}

	if len(explanation.Accounts) == 0 {
		// An explanation with no rows measured nothing and must not read as a clean file.
		fmt.Fprintf(out, "  no rules were offered this file, so this explains nothing\n")
		return
	}

	listened, declinedBySelf, declinedByConfig := partitionAccounts(explanation.Accounts)

	fmt.Fprintf(out, "  %d rules listened, %d declined the file, %d were off by config — %s total\n",
		len(listened), len(declinedBySelf), len(declinedByConfig), formatMilliseconds(explanation.TotalDuration))

	for _, account := range listened {
		fmt.Fprintf(out, "\n  %s  %s, %d nodes\n",
			account.Name, formatMilliseconds(account.Duration), account.NodesOffered)
		fmt.Fprintf(out, "    listens for: %s\n", strings.Join(account.NodeKinds, ", "))

		for _, finding := range account.Findings {
			fmt.Fprintf(out, "    %d:%d  %s  %s\n",
				finding.Line, finding.Column, finding.MessageId, finding.Message)
		}
		// Suppressed findings are shown rather than hidden. A run that withheld a finding and a run
		// that had none to withhold are the same output otherwise, and on a single file the reader
		// is usually asking precisely why they did not see something.
		for _, finding := range account.Suppressed {
			fmt.Fprintf(out, "    %d:%d  %s  [suppressed]  %s\n",
				finding.Line, finding.Column, finding.MessageId, finding.Message)
		}
		if len(account.Findings) == 0 && len(account.Suppressed) == 0 {
			fmt.Fprintf(out, "    no findings\n")
		}
	}

	// Declining is the cheapest and most valuable thing a rule can do, so the rules that declined
	// are listed compactly rather than dropped: seeing that a rule looked is the point.
	if len(declinedBySelf) > 0 {
		fmt.Fprintf(out, "\n  declined this file (the rule looked and opted out): %s\n",
			strings.Join(namesOf(declinedBySelf), ", "))
	}

	// Config declines are separated and each carries its reason, because "somebody turned this off
	// here" and "nobody has said whether this should run" are different facts.
	for _, account := range declinedByConfig {
		fmt.Fprintf(out, "  %s did not run: %s\n", account.Name, account.DeclinedBecause)
	}
}

// partitionAccounts splits rules by what happened to them, keeping the three cases apart because
// they are three different answers to "why did this rule not report anything".
func partitionAccounts(accounts []program.RuleAccount) (listened, declinedBySelf, declinedByConfig []program.RuleAccount) {
	for _, account := range accounts {
		switch {
		case account.DeclinedBecause != "":
			declinedByConfig = append(declinedByConfig, account)
		case account.Listened:
			listened = append(listened, account)
		default:
			declinedBySelf = append(declinedBySelf, account)
		}
	}
	return listened, declinedBySelf, declinedByConfig
}

func namesOf(accounts []program.RuleAccount) []string {
	names := make([]string, 0, len(accounts))
	for _, account := range accounts {
		names = append(names, account.Name)
	}
	return names
}

// findExplainSubject matches a file the user named against the program's own file list.
//
// Matched by suffix so a reader can pass a short relative path rather than the absolute one the
// compiler uses, which is what anyone actually types. The match must fall on a path separator, or
// `Api.ts` would match `McpApi.ts` and explain a file nobody asked about, which is the quiet kind of
// wrong: a complete, correct-looking explanation of the wrong subject.
func findExplainSubject(files []*ast.SourceFile, named string) *ast.SourceFile {
	wanted := filepath.ToSlash(named)

	for _, candidate := range files {
		name := filepath.ToSlash(candidate.FileName().AsString())
		if name == wanted {
			return candidate
		}
	}
	for _, candidate := range files {
		name := filepath.ToSlash(candidate.FileName().AsString())
		if strings.HasSuffix(name, "/"+wanted) {
			return candidate
		}
	}
	return nil
}
