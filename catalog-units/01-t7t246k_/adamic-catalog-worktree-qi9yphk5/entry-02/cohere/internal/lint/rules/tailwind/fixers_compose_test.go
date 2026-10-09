package tailwind

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The Tailwind fixers compose: a shuffled, duplicated, badly spaced literal reaches the plugin's text
// through the real edit engine, in a bounded number of passes, with nothing refused.
//
// Each fixer used to rewrite the whole literal, so on one literal only one of them landed per pass and
// the rest were refused as overlapping: 6,847 refusals on the perturbed ahra corpus and 8,517 on
// www-phi-health (#vf1hd6j), every one landing a pass later. Now each edits only its own bytes
// (class_tokens.go). The order rule waits for the repeat to go, so the shape is two passes that land
// and a third that finds nothing.
//
// The wanted text is what Prettier's Tailwind plugin wrote for the same input, run on 2026-10-02 with
// ahra's pinned prettier-plugin-tailwindcss 0.8.1 and its own theme.css:
//
//	"  items-center flex p-2 flex   peer  "  ->  "peer flex items-center p-2"
//
// Mutation check, run when this was written: putting `no-unnecessary-whitespace` back to rewriting
// its whole segment turns this red with overlap refusals against `no-duplicate-classes`.
func TestTailwindFixersComposeInOnePass(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name, source, want string
	}{
		{
			name:   "shuffled, duplicated and padded",
			source: `export const element = <div className="  items-center flex p-2 flex   peer  " />;`,
			want:   `export const element = <div className="peer flex items-center p-2" />;`,
		},
		{
			// A template with holes: each run sorted on its own, the hole's string sorted as a literal,
			// and the padding at the template's outer edges gone. The plugin wrote, for the same input,
			//	`flex items-center ${open ? "m-2 p-4" : ""} block gap-2`
			// and the quotes are its formatter's choice, not its sorter's.
			name:   "a padded, shuffled template with a string in its hole",
			source: "export const element = <div className={`  items-center flex ${open ? 'p-4 m-2' : ''} gap-2 block  `} />;",
			want:   "export const element = <div className={`flex items-center ${open ? 'm-2 p-4' : ''} block gap-2`} />;",
		},
		{
			// Repeats inside both runs of a template: removed in the first pass, the runs ordered in
			// the second. The plugin wrote exactly this for the same input.
			name:   "a template whose runs hold repeats",
			source: "export const element = <div className={`items-center flex  flex ${size} gap-2 block gap-2`} />;",
			want:   "export const element = <div className={`flex items-center ${size} block gap-2`} />;",
		},
		{
			// A deprecated class that has to move: the rename lands first and the order after it.
			name:   "a deprecated class that moves, beside a repeat and padding",
			source: `export const element = <div className=" items-center flex-grow  flex flex" />;`,
			want:   `export const element = <div className="flex grow items-center" />;`,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result, err := edit.FixText("/Component.tsx", testCase.source, proposeFromTailwindFixers(t), edit.DefaultMaxPasses)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			for _, rejection := range result.Rejected {
				t.Errorf("refused %s (conflicts with %s): %s",
					rejection.Proposal.RuleName, rejection.ConflictsWith, rejection.Reason)
			}
			t.Logf("%d passes, %d fixes applied by %v", result.Passes, len(result.Applied), appliedRuleNames(result.Applied))
			if !result.Converged {
				t.Fatalf("did not converge in %d passes", result.Passes)
			}
			if result.Text != testCase.want {
				t.Fatalf("the fixers produced\n  %q\nwant the plugin's\n  %q", result.Text, testCase.want)
			}
			if result.Passes > 4 {
				t.Errorf("took %d passes; the rename, the repeat and the padding land together, then the order", result.Passes)
			}
		})
	}
}

// proposeFromTailwindFixers runs the four fixing rules over the current text and collects their fixes.
//
// The order rule reads the design system off a real program, so its pass goes through the same
// fixture runner its own tests use, rebuilt on every pass because the text changed. The other three
// are syntactic and run on a parse.
func proposeFromTailwindFixers(t *testing.T) edit.Propose {
	t.Helper()
	return func(fileName string, text string) ([]edit.Proposal, error) {
		baseName := filepath.Base(fileName)
		diagnostics := []rule.Diagnostic{}

		ordered := runClassOrderFixture(t, baseName, text)
		diagnostics = append(diagnostics, ordered.Diagnostics...)

		for _, subject := range []rule.Rule{NoDuplicateClasses, NoUnnecessaryWhitespace, NoDeprecatedClasses} {
			diagnostics = append(diagnostics, rule_testing.Run(t, subject, baseName, text).Diagnostics...)
		}

		// The typed fixture writes the file trimmed, with one newline after it, so its offsets are the
		// engine's only while the text has no leading whitespace. Asserted rather than assumed: a
		// mismatch would move every order fix.
		if ordered.SourceFile == nil || !strings.HasPrefix(ordered.SourceFile.Text(), text) {
			t.Fatalf("the order rule read different text from the engine's")
		}
		return edit.ProposalsFrom(diagnostics), nil
	}
}

// appliedRuleNames names the rules whose fixes landed, so the log shows the composition rather than
// only its result.
func appliedRuleNames(applied []edit.Proposal) []string {
	names := []string{}
	for _, proposal := range applied {
		names = append(names, strings.TrimPrefix(proposal.RuleName, "better-tailwindcss/"))
	}
	return names
}
