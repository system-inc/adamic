package main

import (
	"fmt"
	"io"

	"github.com/system-inc/cohere/internal/types/program"
)

// reportBugsAt is where a user reports a bug in cohere. The README's "Reporting bugs" names the same
// place, and the two change together.
const reportBugsAt = "https://github.com/system-inc/cohere/issues"

// runCrash is a rule that could not finish a file, or a file nothing could check, because cohere crashed.
//
// A crash is cohere's bug, never the user's, and a verdict it leaves missing fails the run (#v1ah2qq).
// It used to be named only under --verbose while the run ended green: TanStack/query's six react-hooks
// crashes printed a footer marker and nothing else, so a missing verdict read like a clean one. The README
// already promised that a run which could not check a file never exits 0.
type runCrash struct {
	Path string `json:"path"`
	// Rule is the rule that crashed, empty when the whole file was lost.
	Rule  string `json:"rule,omitempty"`
	Cause string `json:"cause"`
}

// crashesFrom is a walk's crashes, whole files first, each group in the order the coverage names them.
func crashesFrom(coverage program.Coverage) []runCrash {
	crashes := make([]runCrash, 0, len(coverage.FilesCrashed)+len(coverage.RulesCrashed))
	for _, crash := range coverage.FilesCrashed {
		crashes = append(crashes, runCrash{Path: crash.FileName, Cause: fmt.Sprint(crash.Cause)})
	}
	for _, crash := range coverage.RulesCrashed {
		crashes = append(crashes, runCrash{Path: crash.FileName, Rule: crash.RuleName, Cause: fmt.Sprint(crash.Cause)})
	}
	return crashes
}

// message says what crashed, that it is cohere's bug, what is missing because of it, and where to report it.
func (crash runCrash) message() string {
	what := fmt.Sprintf("cohere's rule %s crashed on this file (%s). This is a bug in cohere, not in your code: "+
		"the rule's verdict on this file is missing, and the file's other rules ran.", crash.Rule, crash.Cause)
	if crash.Rule == "" {
		what = fmt.Sprintf("cohere crashed on this file (%s). This is a bug in cohere, not in your code: "+
			"nothing in this file was checked.", crash.Cause)
	}
	return what + fmt.Sprintf(" Please report it at %s with the output of `cohere --version`, the rule, "+
		"the code that crashed it, and this message.", reportBugsAt)
}

// printCrash prints one crash in the view the run asked for: a line in the human view, a line of JSON under
// `--json`. `--verbose` names crashes in its coverage notes, in the same words.
func printCrash(out io.Writer, crash runCrash) {
	switch activeOutput.Mode {
	case outputJSON:
		writeJSONLine(out, crashJSON{Kind: "crash", runCrash: crash})
	case outputHuman:
		fmt.Fprintln(out, crashLine(crash, activeOutput.Style))
	}
}

// crashLine is a crash in the human view, laid out as a finding is, without a line or column: a crash
// belongs to the file, not to a place in it.
func crashLine(crash runCrash, style textStyle) string {
	rule := crash.Rule
	if rule == "" {
		rule = "cohere"
	}
	return fmt.Sprintf("%s %s %s %s", crash.Path, style.red("crash"), style.dim(rule), crash.message())
}

// crashJSON is one crash's line under `--json`.
type crashJSON struct {
	Kind string `json:"kind"`
	runCrash
}
