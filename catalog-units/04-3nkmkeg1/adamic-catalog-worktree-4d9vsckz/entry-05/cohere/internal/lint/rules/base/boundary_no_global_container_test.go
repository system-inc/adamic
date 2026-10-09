package base

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

// boundaryNoGlobalContainerFile names the fixture file. The rule reads no path and gates on no extension.
const boundaryNoGlobalContainerFile = "/repository/source/Container.ts"

// asTheNoGlobalContainerHarnessWroteIt transforms a fixture the way the harness transforms input.
//
// rule_testing writes each typed fixture as strings.TrimSpace(contents)+"\n". These cases are
// written without leading whitespace so the transform is an identity, and it is applied anyway so
// a later case that does carry indentation cannot silently shift every span in the table.
func asTheNoGlobalContainerHarnessWroteIt(text string) string {
	return strings.TrimSpace(text) + "\n"
}

// TestNoGlobalContainer measures this port against the SOURCE rule, shape by shape.
//
// These rules are ours rather than an upstream port, so there is no published corpus to import and
// no fixture-provenance question. That inverts the usual guidance: a fixture invented here would
// encode the same belief as the rule, and pass for exactly the reason the rule is wrong.
//
// So every row was measured by DRIVING the real rule from api-phi-health, loaded directly with its
// @nexus alias supplied by a loader hook, over the exact source text below. Two of them taught this
// port something it would not have invented: a property key reports, and a locally declared
// function of that name reports too.
func TestNoGlobalContainer(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		why        string
		sourceText string
		wantIds    []string
		wantSpans  []string
	}{
		{
			name:       "bare-call",
			why:        "the ordinary shape the rule exists for",
			sourceText: "const c = getGlobalContainer();\n",
			wantIds:    []string{"boundaryNoGlobalContainer"},
			wantSpans:  []string{"getGlobalContainer"},
		},
		{
			name:       "member-on-result",
			why:        "member access on the call's result, which the source rule catches through the same identifier visitor rather than a second one",
			sourceText: "getGlobalContainer().resolve('x');\n",
			wantIds:    []string{"boundaryNoGlobalContainer"},
			wantSpans:  []string{"getGlobalContainer"},
		},
		{
			name:       "import-specifier",
			why:        "an unaliased import plus a call. The SOURCE RULE reports three times here and this port reports twice, because its parser gives the specifier two identifier nodes at one span and ours gives one. Recorded as the port's one divergence",
			sourceText: "import { getGlobalContainer } from './c';\nconst c = getGlobalContainer();\n",
			wantIds:    []string{"boundaryNoGlobalContainer", "boundaryNoGlobalContainer"},
			wantSpans:  []string{"getGlobalContainer", "getGlobalContainer"},
		},
		{
			name:       "import-only",
			why:        "the import alone, which is where the doubling is clearest: the source rule reports twice at the identical span and this reports once",
			sourceText: "import { getGlobalContainer } from './c';\n",
			wantIds:    []string{"boundaryNoGlobalContainer"},
			wantSpans:  []string{"getGlobalContainer"},
		},
		{
			name:       "renamed-import",
			why:        "a renamed import, where even the source rule reports once, because the local binding is a different name",
			sourceText: "import { getGlobalContainer as gc } from './c';\nconst c = gc();\n",
			wantIds:    []string{"boundaryNoGlobalContainer"},
			wantSpans:  []string{"getGlobalContainer"},
		},
		{
			name:       "property-key",
			why:        "a property KEY of that name, which reports. The rule tests the name rather than resolving it, so anything spelled this way is refused",
			sourceText: "const o = { getGlobalContainer: 1 };\n",
			wantIds:    []string{"boundaryNoGlobalContainer"},
			wantSpans:  []string{"getGlobalContainer"},
		},
		{
			name:       "property-access",
			why:        "a member access of that name on an unrelated object, which also reports for the same reason",
			sourceText: "declare const o: any;\no.getGlobalContainer;\n",
			wantIds:    []string{"boundaryNoGlobalContainer"},
			wantSpans:  []string{"getGlobalContainer"},
		},
		{
			name:       "string-literal",
			why:        "a string holding the name, which is silent: there is no identifier",
			sourceText: "const s = 'getGlobalContainer';\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "local-declaration",
			why:        "a LOCAL function of that name, which still reports. A rule that resolved the name would exempt this, and a local shadow of the global container is exactly as bad as the global one",
			sourceText: "function getGlobalContainer() { return 1; }\n",
			wantIds:    []string{"boundaryNoGlobalContainer"},
			wantSpans:  []string{"getGlobalContainer"},
		},
		{
			name:       "type-position",
			why:        "the name in a type query, reported in both the declaration and the query",
			sourceText: "declare const getGlobalContainer: () => number;\ntype T = typeof getGlobalContainer;\n",
			wantIds:    []string{"boundaryNoGlobalContainer", "boundaryNoGlobalContainer"},
			wantSpans:  []string{"getGlobalContainer", "getGlobalContainer"},
		},
		{
			name:       "comment-only",
			why:        "the name in a comment, silent",
			sourceText: "// getGlobalContainer is banned\nconst x = 1;\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "similar-name",
			why:        "a longer name that merely starts with it, silent, so the test is equality rather than a prefix",
			sourceText: "const c = getGlobalContainerThing();\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
		{
			name:       "control-clean",
			why:        "the control: an ordinary call, so a rule that had stopped looking would still be visible here",
			sourceText: "const c = somethingElse();\n",
			wantIds:    []string{},
			wantSpans:  []string{},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, BoundaryNoGlobalContainer, boundaryNoGlobalContainerFile, testCase.sourceText)
			// A row expecting nothing goes through ExpectClean rather than through ExpectFindings
			// with an empty list. The two are the same assertion to a reader and not to the
			// fixture-pair guard, which reads the call by name: a suite that only ever calls
			// ExpectFindings proves the rule can detect and never that it can stay quiet, and a
			// violation-only corpus is exactly what that guard exists to catch.
			if len(testCase.wantIds) == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			rule_testing.ExpectFindings(t, result, testCase.wantIds...)
			if len(result.Diagnostics) != len(testCase.wantSpans) {
				t.Fatalf("wanted %d findings, got %d, on the case that covers: %s",
					len(testCase.wantSpans), len(result.Diagnostics), testCase.why)
			}
			written := asTheNoGlobalContainerHarnessWroteIt(testCase.sourceText)
			for findingIndex, wantSpan := range testCase.wantSpans {
				reported := result.Diagnostics[findingIndex]
				gotSpan := written[reported.Range.Pos():reported.Range.End()]
				if gotSpan != wantSpan {
					t.Errorf("finding %d points at %q, wanted %q (%s)",
						findingIndex, gotSpan, wantSpan, testCase.why)
				}
			}
		})
	}
}

// TestNoGlobalContainerRendersTheSourceRulesMessage asserts what a reader is told.
//
// ExpectFindings compares ids and count and nothing else, so the text is asserted separately and
// against a literal typed here rather than against the rule's own constant, which would move with
// any mutation of it.
func TestNoGlobalContainerRendersTheSourceRulesMessage(t *testing.T) {
	t.Parallel()

	result := rule_testing.Run(t, BoundaryNoGlobalContainer, boundaryNoGlobalContainerFile,
		"const c = getGlobalContainer();\n")
	if len(result.Diagnostics) != 1 {
		t.Fatalf("wanted one finding, got %d", len(result.Diagnostics))
	}
	reported := result.Diagnostics[0]
	if reported.Message.Id != "boundaryNoGlobalContainer" {
		t.Errorf("reported id %q, wanted %q", reported.Message.Id, "boundaryNoGlobalContainer")
	}
	const wantMessage = "Usage of 'getGlobalContainer' is not allowed."
	if reported.Message.Description != wantMessage {
		t.Errorf("reported message %q, wanted %q", reported.Message.Description, wantMessage)
	}
}
