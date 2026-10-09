package high_level_intermediate_representation

import (
	"testing"
)

// TestRefCurrentReadsTruncateToTheirRoot pins that `.current` is not a dependency.
//
// Upstream's `visitDependency` truncates such a path to empty with the comment "ref.current access
// is not a valid dep" (`PropagateScopeDependenciesHIR.ts:536`). A scope must depend on the ref
// OBJECT rather than on the mutable slot inside it: `ref.current` changes without the ref changing,
// so a dependency naming it would claim a stability React does not offer.
//
// # This diverges from upstream and the divergence is deliberate
//
// Upstream's guard is `isUseRefType(identifier) && path[0].property === 'current'`. The type half is
// not expressible here -- the IR carries no type, which is what `DependencyGapTypeExclusions`
// records -- so only the property name is tested.
//
// The cost is that a value named `current` on a non-ref object is truncated where upstream keeps the
// path. That is the safe direction: the scope then depends on the whole object and invalidates more
// often than needed, never less. Measured corpus-wide it moves five dependencies from carrying a
// path to carrying none, and the dependency oracle is unchanged at 78 matched of 88, so none of the
// five was an answer upstream names.
//
// The fixture below is upstream's own `repro-maybe-invalid-useCallback-read-maybeRef.ts`, whose
// parameter is deliberately untyped -- "maybe" a ref. Upstream compiles it with no error and a bare
// `maybeRef` dependency; before this, our nested-function walk produced `maybeRef.current` and the
// comparison reported a `RefAccessDifference` against the source's `[maybeRef]`.
func TestRefCurrentReadsTruncateToTheirRoot(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name       string
		source     string
		want       string
		mustNotSee string
	}{
		{
			name: "a current read through a callback capture",
			source: `
				import {useCallback} from 'react';
				function useHook(maybeRef) {
					return useCallback(() => {
						return [maybeRef.current];
					}, [maybeRef]);
				}
			`,
			want:       "maybeRef",
			mustNotSee: "maybeRef.current",
		},
		{
			name: "a deeper path through current still truncates to the root",
			source: `
				import {useCallback} from 'react';
				function useHook(maybeRef) {
					return useCallback(() => {
						return [maybeRef.current.value];
					}, [maybeRef]);
				}
			`,
			want:       "maybeRef",
			mustNotSee: "maybeRef.current.value",
		},
		{
			name: "an ordinary path is untouched",
			source: `
				import {useCallback} from 'react';
				function useHook(props) {
					return useCallback(() => {
						return [props.a.b];
					}, [props.a.b]);
				}
			`,
			want: "props.a.b",
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			deps, ok := inferredDependencyStrings(t, testCase.source)
			if !ok {
				t.Fatal("the fixture did not lower, so this test asserts nothing about the " +
					"dependency it claims to check")
			}
			if len(deps) == 0 {
				t.Fatal("no dependency was collected; the fixture stopped exercising the collector")
			}
			found := false
			for _, dependency := range deps {
				if dependency == testCase.want {
					found = true
				}
				if testCase.mustNotSee != "" && dependency == testCase.mustNotSee {
					t.Errorf("collected %q, which reads through a ref's mutable slot; the scope "+
						"must depend on the ref object instead", dependency)
				}
			}
			if !found {
				t.Errorf("collected %v, which does not contain %q", deps, testCase.want)
			}
		})
	}
}
