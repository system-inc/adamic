package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

// TestCompareManualMemoDependenciesAsymmetry pins the direction, which is the whole rule.
//
// A longer inferred path is fine: reading `props.a.b` where the source declared `props.a` recomputes
// no more often than promised, because any change to `props.a.b` is a change to `props.a`. A shorter
// inferred path is not fine, for the mirror reason.
//
// Getting that backwards produces a rule that accepts every under-specified dependency and reports
// every over-specified one -- exactly inverted, and passing any test that only checks "unequal paths
// disagree".
func TestCompareManualMemoDependenciesAsymmetry(t *testing.T) {
	t.Parallel()

	local := func(identifier static_single_assignment.IdentifierId, properties ...string) ManualMemoDependency {
		path := make([]DependencyPathEntry, 0, len(properties))
		for _, property := range properties {
			path = append(path, DependencyPathEntry{Property: property})
		}
		return ManualMemoDependency{
			Root: ManualMemoRoot{Place: Place{Identifier: identifier}},
			Path: path,
		}
	}

	for _, testCase := range []struct {
		name     string
		inferred ManualMemoDependency
		source   ManualMemoDependency
		want     CompareDependencyResult
	}{
		{
			name:     "identical paths match",
			inferred: local(1, "a", "b"),
			source:   local(1, "a", "b"),
			want:     CompareDependencyOk,
		},
		{
			name:     "inferred deeper than source is acceptable",
			inferred: local(1, "a", "b"),
			source:   local(1, "a"),
			want:     CompareDependencyOk,
		},
		{
			name:     "inferred shallower than source is not",
			inferred: local(1, "a"),
			source:   local(1, "a", "b"),
			want:     CompareDependencySubpath,
		},
		{
			name:     "diverging property is a path difference",
			inferred: local(1, "a", "b"),
			source:   local(1, "a", "c"),
			want:     CompareDependencyPathDifference,
		},
		{
			name:     "different root",
			inferred: local(1, "a"),
			source:   local(2, "a"),
			want:     CompareDependencyRootDifference,
		},
		{
			name:     "both empty paths match",
			inferred: local(1),
			source:   local(1),
			want:     CompareDependencyOk,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := CompareManualMemoDependencies(testCase.inferred, testCase.source); got != testCase.want {
				t.Errorf("got %v, want %v", got, testCase.want)
			}
		})
	}
}

// TestCompareManualMemoDependenciesWithdrawsTheRuleForRefs covers the exception.
//
// A ref is not immutable: `ref_prev === ref_new` does not imply
// `ref_prev.current === ref_new.current`. So the longer-is-fine rule is withdrawn wherever either
// path reads `current`, and a case that would otherwise be `Ok` becomes a ref-access difference.
//
// The control matters here more than usual: the same shape with a different property name must be
// `Ok`, or the case below passes because deeper paths are rejected generally.
func TestCompareManualMemoDependenciesWithdrawsTheRuleForRefs(t *testing.T) {
	t.Parallel()

	path := func(properties ...string) []DependencyPathEntry {
		entries := make([]DependencyPathEntry, 0, len(properties))
		for _, property := range properties {
			entries = append(entries, DependencyPathEntry{Property: property})
		}
		return entries
	}
	root := ManualMemoRoot{Place: Place{Identifier: 1}}

	// Inferred reads deeper AND through `current`: the rule is withdrawn.
	refCase := CompareManualMemoDependencies(
		ManualMemoDependency{Root: root, Path: path("ref", "current", "x")},
		ManualMemoDependency{Root: root, Path: path("ref")},
	)
	if refCase != CompareDependencyRefAccessDifference {
		t.Errorf("a deeper inferred path through `current` returned %v, want a ref-access "+
			"difference; a prefix match through a ref proves nothing", refCase)
	}

	// The control: identical shape, no `current`. Must be Ok, or the case above is not testing refs.
	control := CompareManualMemoDependencies(
		ManualMemoDependency{Root: root, Path: path("ref", "value", "x")},
		ManualMemoDependency{Root: root, Path: path("ref")},
	)
	if control != CompareDependencyOk {
		t.Fatalf("the control returned %v, want ok; the ref case above is therefore not testing "+
			"the ref exception but the depth rule", control)
	}

	// Exact-length equality through `current` is still Ok: no prefix matching is involved.
	exact := CompareManualMemoDependencies(
		ManualMemoDependency{Root: root, Path: path("ref", "current")},
		ManualMemoDependency{Root: root, Path: path("ref", "current")},
	)
	if exact != CompareDependencyOk {
		t.Errorf("identical paths through `current` returned %v, want ok; the exception is about "+
			"partial matches, not about refs generally", exact)
	}
}

// TestCompareManualMemoDependenciesComparesOptionality covers the flag that changes the answer.
//
// An optional inferred read where the source declared a non-optional one is less precise: `a?.b`
// recomputes on a different set of changes than `a.b`. Upstream reports it immediately rather than
// letting the depth rule decide, because it is a difference regardless of what the rest of the path
// does.
func TestCompareManualMemoDependenciesComparesOptionality(t *testing.T) {
	t.Parallel()

	root := ManualMemoRoot{Place: Place{Identifier: 1}}
	optional := ManualMemoDependency{Root: root, Path: []DependencyPathEntry{{Property: "a", Optional: true}}}
	plain := ManualMemoDependency{Root: root, Path: []DependencyPathEntry{{Property: "a"}}}

	if got := CompareManualMemoDependencies(optional, plain); got != CompareDependencyPathDifference {
		t.Errorf("an optional inferred read against a non-optional source returned %v, want a path "+
			"difference", got)
	}
	if got := CompareManualMemoDependencies(plain, optional); got != CompareDependencyPathDifference {
		t.Errorf("the reverse returned %v, want a path difference", got)
	}
	if got := CompareManualMemoDependencies(optional, optional); got != CompareDependencyOk {
		t.Errorf("two optional reads of the same property returned %v, want ok", got)
	}
}

// TestCompareManualMemoDependenciesGlobalRoots covers the root kinds.
//
// A global is compared by name and a local by identifier, and the two kinds never match each other
// -- a global has no identifier in this function to compare against.
func TestCompareManualMemoDependenciesGlobalRoots(t *testing.T) {
	t.Parallel()

	globalRoot := func(name string) ManualMemoDependency {
		return ManualMemoDependency{Root: ManualMemoRoot{IsGlobal: true, Name: name}}
	}
	localRoot := ManualMemoDependency{Root: ManualMemoRoot{Place: Place{Identifier: 1}}}

	if got := CompareManualMemoDependencies(globalRoot("Math"), globalRoot("Math")); got != CompareDependencyOk {
		t.Errorf("two globals of the same name returned %v, want ok", got)
	}
	if got := CompareManualMemoDependencies(globalRoot("Math"), globalRoot("Date")); got != CompareDependencyRootDifference {
		t.Errorf("two globals of different names returned %v, want a root difference", got)
	}
	if got := CompareManualMemoDependencies(globalRoot("Math"), localRoot); got != CompareDependencyRootDifference {
		t.Errorf("a global against a local returned %v, want a root difference", got)
	}
	// A local whose identifier happens to be zero must not match a global by accident.
	zeroLocal := ManualMemoDependency{Root: ManualMemoRoot{Place: Place{Identifier: 0}}}
	if got := CompareManualMemoDependencies(zeroLocal, globalRoot("")); got != CompareDependencyRootDifference {
		t.Errorf("a zero-identifier local against an unnamed global returned %v, want a root "+
			"difference; the kinds must not collapse on zero values", got)
	}
}

// TestMergeCompareDependencyResultsTakesTheMostSpecific pins the merge.
//
// A dependency is compared against every source entry and reported only if none matched. The
// reported reason is the strongest disagreement seen, so a root difference against one entry does
// not mask a ref-access difference against another.
func TestMergeCompareDependencyResultsTakesTheMostSpecific(t *testing.T) {
	t.Parallel()

	results := []CompareDependencyResult{
		CompareDependencyOk,
		CompareDependencyRootDifference,
		CompareDependencyPathDifference,
		CompareDependencySubpath,
		CompareDependencyRefAccessDifference,
	}
	for _, first := range results {
		for _, second := range results {
			want := first
			if second > first {
				want = second
			}
			if got := MergeCompareDependencyResults(first, second); got != want {
				t.Errorf("merge(%v, %v) = %v, want %v", first, second, got, want)
			}
			if got := MergeCompareDependencyResults(second, first); got != want {
				t.Errorf("merge is not commutative for (%v, %v)", first, second)
			}
		}
	}

	// Each result prints distinctly, since these strings reach a developer in the diagnostic.
	seen := map[string]bool{}
	for _, result := range results {
		name := result.String()
		if seen[name] {
			t.Errorf("two results print as %q", name)
		}
		seen[name] = true
	}
}
