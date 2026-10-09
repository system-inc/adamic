package main

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/core"
	"github.com/system-inc/cohere/internal/types/program"
)

/*
 * A disable comment withholds the repair as well as the finding.
 *
 * The lint path filters every diagnostic through `directives.Suppresses` before reporting it
 * (`internal/types/program/walk.go:500`). The fix path did not build that index at all, so a rule's
 * finding was correctly withheld and its fix was applied anyway.
 *
 * That combination is worse than either half alone. Both linters report nothing on the disabled
 * line, both fixers rewrite it, and the file changes with no diagnostic explaining why. The only
 * evidence is in `git diff`, which is the last place anyone looks after a formatting pass.
 *
 * It cost a real repair rather than being theoretical.
 * `libraries/structure/source/router/hooks/useRouter.ts` carries a deliberate disable for
 * `nexus/import-no-forbidden-source`, because that module is the implementation the navigation
 * barrel re-exports and importing through the barrel closes a cycle that makes the hook's own type
 * `any`. A fixer run reverted it and reintroduced 108 findings across five rules.
 *
 * # Why this rule and not a more interesting one
 *
 * `proposalsForText` re-parses the text it is given and does not wire parent pointers, so any rule
 * consulting `node.Parent` reports nothing here regardless of suppression. `no-else-return` was the
 * first choice and failed for exactly that reason, with its unsuppressed control going red, which
 * is what a control is for. `no-debugger` asks nothing of the tree above the node, so it fires on a
 * bare parse and isolates the suppression question from the parsing one.
 *
 * # The unsuppressed cases are the load-bearing half
 *
 * A repair that withheld every fix, or that failed to build the index and returned nothing, would
 * satisfy a table written only from the suppressed cases.
 */
func TestFixPathHonoursADisableComment(t *testing.T) {
	t.Parallel()

	const body = "function f() {\n    debugger;\n}\n"

	for name, testCase := range map[string]struct {
		text          string
		wantProposals bool
	}{
		"unsuppressed": {
			text:          body,
			wantProposals: true,
		},
		"suppressedForTheFile": {
			text:          "/* eslint-disable no-debugger */\n" + body,
			wantProposals: false,
		},
		"suppressedForTheLine": {
			text:          "function f() {\n    // eslint-disable-next-line no-debugger\n    debugger;\n}\n",
			wantProposals: false,
		},
		"suppressedForADifferentRule": {
			// A disable naming another rule must not withhold this one's repair. Without this case
			// an implementation that suppressed everything would pass.
			text:          "/* eslint-disable no-console */\n" + body,
			wantProposals: true,
		},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			proposals, err := proposalsForText("/Probe.ts", testCase.text, &program.Graph{},
				[]rule.Rule{core.NoDebugger})
			if err != nil {
				t.Fatalf("proposalsForText: %v", err)
			}
			if got := len(proposals) > 0; got != testCase.wantProposals {
				t.Errorf("proposals = %d, want %s", len(proposals),
					map[bool]string{true: "at least one", false: "none"}[testCase.wantProposals])
			}
		})
	}
}
