package high_level_intermediate_representation

import (
	"fmt"
	"testing"
)

func TestBranchLocalHoistableFactsRespectDominatingReactiveScope(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		source       string
		wantFindings int
	}{
		{
			name: "top-level branch scope keeps branch facts",
			source: `
				import {useMemo} from 'react';
				function Component({propA, propB}) {
					return useMemo(() => {
						if (propA) return {value: propB.x.y};
					}, [propA, propB.x.y]);
				}
			`,
			wantFindings: 0,
		},
		{
			name: "dominating mutable scope keeps entry facts",
			source: `
				import {useMemo} from 'react';
				import {mutate} from 'shared-runtime';
				function Component({propA, propB}) {
					return useMemo(() => {
						const x = {};
						if (propA?.a) {
							mutate(x);
							return {value: propB.x.y};
						}
					}, [propA?.a, propB.x.y]);
				}
			`,
			// The harness lowers both the component and its memo callback as function-like nodes.
			// React's reverse-postorder places both mutable scopes before their continuations, so
			// each lowering now reaches the same check rather than one being hidden by block order.
			wantFindings: 2,
		},
	}

	for _, testCase := range testCases {
		for _, enabled := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/preservation=%t", testCase.name, enabled), func(t *testing.T) {
				t.Parallel()
				source, wantFindings := testCase.source, 0
				if !enabled {
					source = "// @enablePreserveExistingMemoizationGuarantees:false\n" + source
					wantFindings = testCase.wantFindings
				}
				findings, lowered := findingsForSource(t, source)
				if !lowered {
					t.Fatal("fixture did not lower")
				}
				if len(findings) != wantFindings {
					t.Fatalf("got %d preserved-memoization findings, want %d: %+v",
						len(findings), wantFindings, findings)
				}
			})
		}
	}
}
