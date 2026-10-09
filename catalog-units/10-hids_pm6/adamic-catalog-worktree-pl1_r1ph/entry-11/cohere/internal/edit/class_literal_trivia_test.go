package edit

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rules/nexus"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
)

// One rule's bad range costs every rule's fixes in that file, and a class literal after trivia was
// that bad range.
//
// The parse guard refuses a pass whole, which is the right guarantee and has a price: a single fix
// whose range is wrong takes every other fix in the same file down with it. classLiteralFrom took a
// literal's Loc, whose start sits before the leading whitespace, so for a literal on its own line
// the quote strip cut the opening quote instead of the first class character. The rewrite stopped
// parsing and the engine refused the whole pass, the import fix below included. Measured by
// @system_cohere_format_markdown on perturbed input (#vf1hd6j): 174 files and 1,755 refused fixes in
// ahra, 175 files and 1,932 in www-phi-health. Fixed in cohere c079219.
//
// The rule-level fixture in the tailwind package proves the quote survives. This proves the outcome
// the bug was measured by: two rules, one file, both fixes landing, nothing refused. Mutation check,
// run when this was written: restoring `contentRange := literal.Loc` in classLiteralFrom turns this
// red with both fixes refused for a parse failure, and the file left as it was.
func TestATailwindFixAfterTriviaNoLongerTakesTheFileBatchDown(t *testing.T) {
	t.Parallel()
	source := "import * as NodeFileSystem from 'fs';\n" +
		"\n" +
		"export const merged = cn(\n" +
		"    'flex flex',\n" +
		");\n" +
		"export const roots = [NodeFileSystem];\n"
	want := "import * as NodeFileSystem from 'node:fs';\n" +
		"\n" +
		"export const merged = cn(\n" +
		"    'flex',\n" +
		");\n" +
		"export const roots = [NodeFileSystem];\n"

	result, err := FixText("Component.tsx", source, proposeFromRules(tailwind.NoDuplicateClasses, nexus.ImportRequireNodeNamespace), DefaultMaxPasses)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, rejection := range result.Rejected {
		t.Errorf("refused %s: %s", rejection.Proposal.RuleName, rejection.Reason)
	}
	if result.Text != want {
		t.Fatalf("the file batch did not land:\n  want %q\n  got  %q", want, result.Text)
	}

	// Both rules must be among what landed, or the test has stopped being about a shared batch.
	landed := strings.Join(ruleNamesOf(result.Applied), ",")
	for _, ruleName := range []string{tailwind.NoDuplicateClasses.Name, nexus.ImportRequireNodeNamespace.Name} {
		if !strings.Contains(landed, ruleName) {
			t.Fatalf("%s did not land, so this no longer proves a shared batch (landed: %s)", ruleName, landed)
		}
	}
}
