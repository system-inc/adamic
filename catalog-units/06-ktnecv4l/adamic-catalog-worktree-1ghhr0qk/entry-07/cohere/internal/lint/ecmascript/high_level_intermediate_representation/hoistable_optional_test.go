package high_level_intermediate_representation

import (
	"testing"
)

// TestOptionalLoadsProveNothingAboutTheirObject pins that `a?.b` does not make `a` hoistable.
//
// An optional load is the guard AGAINST its object being null, so reading it is not evidence the
// object is non-null -- it is evidence the program expects it might be. Recording it as non-null
// makes the dependency tree's cursor non-null at that step, which takes the
// `hoistableCursor.nonNull` arm in `addDependency` and flattens `arg?.items` to `arg.items`. The
// comparison then reports an optionality mismatch against a source that wrote `arg?.items`, which is
// a disagreement the analysis manufactured.
//
// # Why this is read from a flag rather than ported
//
// Upstream keys its hoistable set by optional BLOCK: `collectOptionalChainSidemap` builds
// `optionalBlock -> baseObject?.a` (`CollectHoistablePropertyLoads.ts:94`), so the guard survives
// into the tree and the optional arm fires. That pass is 418 lines and is driven by `Optional`
// TERMINALS. This lowering does not produce them -- measured on the fixture below, the only
// terminals are `Return` -- so the same fact is read from `PropertyLoad.Optional`, which upstream
// does not have because its optionality lives in control flow.
//
// The two halves are atomic. Measured, neither alone moves anything: the hoistable change without
// the flag reaching the tree is inert, and the flag without this change is flattened at the arm.
func TestOptionalLoadsProveNothingAboutTheirObject(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		source     string
		wantSilent bool
	}{
		{
			// The developer writes `arg?.items` and the callback reads `arg?.items`. Those agree,
			// and reporting them is the defect this pins.
			name: "an optional dependency read optionally",
			source: `
				import {useMemo} from 'react';
				function Component({arg}) {
					return useMemo(() => {
						const x = [];
						x.push(arg?.items);
						return x;
					}, [arg?.items]);
				}
			`,
			wantSilent: true,
		},
		{
			// The control. An unguarded read does prove the object non-null, so the path stays deep
			// and nothing here should change it.
			name: "an unconditional dependency is unaffected",
			source: `
				import {useMemo} from 'react';
				function Component({arg}) {
					return useMemo(() => {
						const x = [];
						x.push(arg.items);
						return x;
					}, [arg.items]);
				}
			`,
			wantSilent: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			deps, ok := inferredDependencyStrings(t, testCase.source)
			if !ok {
				t.Fatal("the fixture did not lower, so this test asserts nothing")
			}
			if len(deps) == 0 {
				t.Fatal("no dependency was collected; the fixture stopped exercising the collector")
			}
			t.Logf("collected %v", deps)
		})
	}
}

// TestOptionalHoistableIsNotRecorded is the direct assertion, at the analysis rather than the rule.
//
// Kept separate from the fixture-level test above because it fails for one reason: an optional load
// contributing a non-null fact. A rule-level assertion can go green for several.
func TestOptionalHoistableIsNotRecorded(t *testing.T) {
	t.Parallel()

	optional := &PropertyLoad{
		Object:   Place{Identifier: 7},
		Property: "items",
		Optional: true,
	}
	if _, recorded := maybeNonNullInInstruction(optional, temporaries{}); recorded {
		t.Error("an optional load contributed a non-null fact about its object; `a?.b` is the " +
			"guard against `a` being null and proves the opposite")
	}

	unconditional := &PropertyLoad{
		Object:   Place{Identifier: 7},
		Property: "items",
	}
	path, recorded := maybeNonNullInInstruction(unconditional, temporaries{})
	if !recorded {
		t.Fatal("an unconditional load contributed nothing; that is the whole input to this " +
			"analysis and the assertion above would pass vacuously without it")
	}
	if path.Identifier != 7 {
		t.Errorf("recorded identifier %d, want 7", path.Identifier)
	}
}
