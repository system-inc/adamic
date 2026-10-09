package high_level_intermediate_representation

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// scopesFor lowers one source and returns the outermost function with its scopes.
//
// Typed for the reason `disjointFor` is: a checker-less lowering never emits `StoreContext`, so a
// plain harness would make several assertions below pass vacuously.
func scopesFor(t *testing.T, source string) (*Function, *ReactiveScopes) {
	t.Helper()
	function, ranges := rangesFor(t, source)
	set := FindDisjointMutableValuesWithRanges(function, ranges)
	return function, AssignReactiveScopesWithSets(function, ranges, set)
}

// scopeOfName returns the scope holding the LAST value a named binding takes.
func scopeOfName(t *testing.T, function *Function, scopes *ReactiveScopes, name string) ScopeId {
	t.Helper()
	var target static_single_assignment.IdentifierId
	found := false
	for id, identifier := range function.Identifiers {
		if identifier != nil && identifier.Name == name {
			if !found || static_single_assignment.IdentifierId(id) > target {
				target = static_single_assignment.IdentifierId(id)
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("no identifier named %q, so this test measured nothing", name)
	}
	return scopes.ScopeOf(target)
}

// ---------------------------------------------------------------------------
// The hull, measured against React's own merge loop
// ---------------------------------------------------------------------------

// TestScopeHullMatchesReact pins the range merge against React's real implementation.
//
// These ten cases were not invented. They were run through React 7.1.1's own merge loop -- the body
// of `inferReactiveScopeVariables` at bundle line 32235, reached by copying the development bundle
// and appending an export inside its IIFE -- and this table is the output that produced.
//
// The cases exist to separate implementations rather than to confirm one. A plain
// `min(start), max(end)` passes `simpleHull`, `singleton`, `identicalRanges` and
// `laterStartsHigher`, and FAILS every case involving an unset member, because zero is the unset
// sentinel and a naive minimum drags the hull's start to zero. `firstUnset` and `runOfUnsets`
// separate the adoption branch from the skip branch, which are two different behaviours sharing one
// `if/else if`.
func TestScopeHullMatchesReact(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		ranges []mutation_aliasing.MutableRange
		want   mutation_aliasing.MutableRange
	}{
		{"simpleHull", []mutation_aliasing.MutableRange{{Start: 3, End: 5}, {Start: 1, End: 9}, {Start: 7, End: 8}}, mutation_aliasing.MutableRange{Start: 1, End: 9}},
		{"firstUnset", []mutation_aliasing.MutableRange{{Start: 0, End: 0}, {Start: 4, End: 6}, {Start: 2, End: 9}}, mutation_aliasing.MutableRange{Start: 2, End: 9}},
		{"middleUnset", []mutation_aliasing.MutableRange{{Start: 4, End: 6}, {Start: 0, End: 0}, {Start: 2, End: 9}}, mutation_aliasing.MutableRange{Start: 2, End: 9}},
		{"allUnset", []mutation_aliasing.MutableRange{{Start: 0, End: 0}, {Start: 0, End: 0}}, mutation_aliasing.MutableRange{Start: 0, End: 0}},
		{"runOfUnsets", []mutation_aliasing.MutableRange{{Start: 0, End: 0}, {Start: 0, End: 0}, {Start: 5, End: 7}}, mutation_aliasing.MutableRange{Start: 5, End: 7}},
		{"singleton", []mutation_aliasing.MutableRange{{Start: 3, End: 7}}, mutation_aliasing.MutableRange{Start: 3, End: 7}},
		{"identicalRanges", []mutation_aliasing.MutableRange{{Start: 2, End: 4}, {Start: 2, End: 4}}, mutation_aliasing.MutableRange{Start: 2, End: 4}},
		{"laterStartsHigher", []mutation_aliasing.MutableRange{{Start: 2, End: 4}, {Start: 6, End: 9}}, mutation_aliasing.MutableRange{Start: 2, End: 9}},
		{"trailingUnsetOnly", []mutation_aliasing.MutableRange{{Start: 2, End: 4}, {Start: 0, End: 0}}, mutation_aliasing.MutableRange{Start: 2, End: 4}},
		{"unsetEndOnlyGuard", []mutation_aliasing.MutableRange{{Start: 5, End: 9}, {Start: 0, End: 0}, {Start: 0, End: 0}}, mutation_aliasing.MutableRange{Start: 5, End: 9}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			// Build a function whose identifiers carry exactly these ranges, and a set unioning
			// them into one class. Synthetic rather than lowered, because the point is to pin the
			// merge arithmetic against React's, and a lowered input cannot be made to produce an
			// arbitrary range table.
			function := &Function{}
			ranges := &mutation_aliasing.MutableRanges{}
			var members []static_single_assignment.IdentifierId
			for index, memberRange := range testCase.ranges {
				id := static_single_assignment.IdentifierId(index + 1)
				function.Identifiers = append(function.Identifiers, nil)
				ranges.Set(id, memberRange)
				members = append(members, id)
			}
			function.Identifiers = append(function.Identifiers, nil)

			set := &DisjointSet{}
			set.Union(members)

			scopes := AssignReactiveScopesWithSets(function, ranges, set)
			if scopes.Len() != 1 {
				t.Fatalf("got %d scopes, want exactly 1; the class did not form", scopes.Len())
			}
			got := scopes.RangeOf(scopes.Ids()[0])
			if got != testCase.want {
				t.Errorf("hull = [%d,%d), want [%d,%d) -- this is React's own answer for this input",
					got.Start, got.End, testCase.want.Start, testCase.want.End)
			}
		})
	}
}

// TestScopeHullIsNotAPlainMinimum is the negative control for the case above.
//
// The whole reason the merge is written as an `if/else if` rather than as `min` is the unset
// sentinel, and a reader simplifying it would not obviously be wrong. This asserts the specific
// input on which the two spellings differ, so the simplification cannot land silently.
func TestScopeHullIsNotAPlainMinimum(t *testing.T) {
	t.Parallel()

	function := &Function{Identifiers: []*Identifier{nil, nil, nil}}
	ranges := &mutation_aliasing.MutableRanges{}
	ranges.Set(1, mutation_aliasing.MutableRange{Start: 4, End: 6})
	ranges.Set(2, mutation_aliasing.MutableRange{Start: 0, End: 0})

	set := &DisjointSet{}
	set.Union([]static_single_assignment.IdentifierId{1, 2})

	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	got := scopes.RangeOf(scopes.Ids()[0])
	if got.Start == 0 {
		t.Fatal("the unset member dragged the hull's start to zero, which is a plain minimum " +
			"rather than upstream's guarded merge; zero is the UNSET sentinel, not a position")
	}
	if got != (mutation_aliasing.MutableRange{Start: 4, End: 6}) {
		t.Errorf("hull = [%d,%d), want [4,6)", got.Start, got.End)
	}
}

// ---------------------------------------------------------------------------
// One scope per class, which is the whole shape of this pass
// ---------------------------------------------------------------------------

// TestScopeIsOnePerEquivalenceClass pins the central claim: a scope names a class, not a value.
//
// Upstream keys its scope map on the disjoint set's GROUP representative (bundle line 32235), so
// every member of a class reaches the same scope object. A pass that minted one scope per value
// would compile, would assign a scope to everything, and would be wrong in the way that matters --
// the consumer asks "are these two memoized together" and would always hear no.
func TestScopeIsOnePerEquivalenceClass(t *testing.T) {
	t.Parallel()

	function := &Function{Identifiers: []*Identifier{nil, nil, nil, nil, nil, nil}}
	ranges := &mutation_aliasing.MutableRanges{}
	for id := static_single_assignment.IdentifierId(1); id <= 5; id++ {
		ranges.Set(id, mutation_aliasing.MutableRange{Start: static_single_assignment.EvaluationOrder(id), End: static_single_assignment.EvaluationOrder(id) + 3})
	}

	set := &DisjointSet{}
	set.Union([]static_single_assignment.IdentifierId{1, 2, 3})
	set.Union([]static_single_assignment.IdentifierId{4, 5})

	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	if scopes.Len() != 2 {
		t.Fatalf("got %d scopes for 2 classes over 5 values; a scope must name a class", scopes.Len())
	}
	if a, b, c := scopes.ScopeOf(1), scopes.ScopeOf(2), scopes.ScopeOf(3); a != b || b != c {
		t.Errorf("values 1,2,3 are one class but got scopes %d,%d,%d", a, b, c)
	}
	if d, e := scopes.ScopeOf(4), scopes.ScopeOf(5); d != e {
		t.Errorf("values 4,5 are one class but got scopes %d,%d", d, e)
	}
	if scopes.ScopeOf(1) == scopes.ScopeOf(4) {
		t.Error("two separate classes were given one scope, which is a false merge")
	}
}

// TestScopeAbsentValueHasNoScope pins that a value outside every class gets zero rather than a scope.
//
// Upstream's `identifier.scope` stays null for a value `findDisjointMutableValues` never admitted,
// and the consumer at bundle line 45732 branches on `identifier.scope == null`. Zero is that null
// here, which is why zero is not a valid ScopeId.
func TestScopeAbsentValueHasNoScope(t *testing.T) {
	t.Parallel()

	function := &Function{Identifiers: []*Identifier{nil, nil, nil, nil}}
	ranges := &mutation_aliasing.MutableRanges{}
	ranges.Set(1, mutation_aliasing.MutableRange{Start: 1, End: 4})
	ranges.Set(2, mutation_aliasing.MutableRange{Start: 2, End: 5})

	set := &DisjointSet{}
	set.Union([]static_single_assignment.IdentifierId{1, 2})

	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	if got := scopes.ScopeOf(3); got != 0 {
		t.Errorf("an unscoped value answered scope %d, want 0; zero is the absent scope", got)
	}
	if scopes.ScopeOf(1) == 0 {
		t.Error("a scoped value answered zero, which makes the absent sentinel unreadable")
	}
}

// TestScopeMemberRangesAreRewrittenToTheHull pins upstream's second write.
//
// `identifier.mutableRange = scope.range` at bundle line 32261 is the line a reader does not expect,
// and it is not incidental: after this pass a member reports its CLASS's span rather than its own,
// because entangled values are memoized as one unit and therefore live as one unit. A port that
// assigned scopes and left ranges alone would satisfy every other test in this file.
func TestScopeMemberRangesAreRewrittenToTheHull(t *testing.T) {
	t.Parallel()

	function := &Function{Identifiers: []*Identifier{nil, nil, nil, nil}}
	ranges := &mutation_aliasing.MutableRanges{}
	ranges.Set(1, mutation_aliasing.MutableRange{Start: 3, End: 5})
	ranges.Set(2, mutation_aliasing.MutableRange{Start: 1, End: 9})

	set := &DisjointSet{}
	set.Union([]static_single_assignment.IdentifierId{1, 2})

	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	rewritten := scopes.MemberRanges()

	want := mutation_aliasing.MutableRange{Start: 1, End: 9}
	for _, id := range []static_single_assignment.IdentifierId{1, 2} {
		if got := rewritten.Get(id); got != want {
			t.Errorf("value %d reports [%d,%d) after entanglement, want the hull [%d,%d)",
				id, got.Start, got.End, want.Start, want.End)
		}
	}

	// The INPUT table is deliberately not modified, which is where this differs from upstream.
	if got := ranges.Get(1); got != (mutation_aliasing.MutableRange{Start: 3, End: 5}) {
		t.Errorf("the input range table was mutated to [%d,%d); this pass must leave it alone so a "+
			"caller can still see the pre-entanglement range", got.Start, got.End)
	}
	// A value with no scope is absent from the rewritten view.
	if got := rewritten.Get(3); got.IsSet() {
		t.Errorf("an unscoped value appears in the rewritten ranges as [%d,%d)", got.Start, got.End)
	}
}

// ---------------------------------------------------------------------------
// Determinism and termination
// ---------------------------------------------------------------------------

// TestScopeAssignmentIsDeterministic pins that scope ids do not move between runs.
//
// The classes come from `DisjointSet.Sets`, which sorts precisely so that a Go map's randomised
// iteration cannot leak into the answer. Ids are handed out in that order, so a scope id is stable.
// A pass whose partition is right but whose keys move between runs produces a table that differs
// every time, which is the shape that makes a cache non-reproducible without ever being wrong.
func TestScopeAssignmentIsDeterministic(t *testing.T) {
	t.Parallel()

	const source = `
		function Component(props) {
			const items = [];
			const other = {};
			items.push(props.a);
			other.value = props.b;
			return <div>{items}{other}</div>;
		}
	`
	_, first := scopesFor(t, source)
	if first.Len() == 0 {
		t.Fatal("no scopes formed, so this test measured nothing")
	}
	for attempt := 0; attempt < 8; attempt++ {
		_, again := scopesFor(t, source)
		if again.Len() != first.Len() {
			t.Fatalf("run %d produced %d scopes, first run produced %d", attempt, again.Len(), first.Len())
		}
		for index, id := range first.Ids() {
			if again.Ids()[index] != id {
				t.Fatalf("run %d assigned scope ids in a different order", attempt)
			}
			firstRange, againRange := first.RangeOf(id), again.RangeOf(id)
			if firstRange != againRange {
				t.Fatalf("run %d gave scope %d range [%d,%d), first run gave [%d,%d)",
					attempt, id, againRange.Start, againRange.End, firstRange.Start, firstRange.End)
			}
			if len(first.MembersOf(id)) != len(again.MembersOf(id)) {
				t.Fatalf("run %d gave scope %d a different membership", attempt, id)
			}
		}
	}
}

// TestScopeAssignmentIsASinglePass proves the termination argument empirically.
//
// The claim in the package comment is that there is no fixpoint here at all: one visit per class and
// one per member, both over slices fixed before the walk begins. That is a weak claim compared to
// the union-find upstream of it, and a weak claim is still worth pinning, because the failure mode
// of getting it wrong is not a wrong answer but a hang -- the same silent shape `ranges.go` and
// `disjoint.go` both prove empirically rather than assert.
//
// Counted rather than argued: every member appears in exactly one scope, and the members summed
// across scopes equal the set's size. A pass that revisited would double-count; one that iterated to
// convergence could not finish this test at all.
func TestScopeAssignmentIsASinglePass(t *testing.T) {
	t.Parallel()

	function := &Function{}
	ranges := &mutation_aliasing.MutableRanges{}
	set := &DisjointSet{}

	// 2,000 values in 200 classes of 10. Large enough that a quadratic or iterating implementation
	// is visible as a hang rather than as a slow test.
	const classes, perClass = 200, 10
	function.Identifiers = append(function.Identifiers, nil)
	for class := 0; class < classes; class++ {
		var members []static_single_assignment.IdentifierId
		for member := 0; member < perClass; member++ {
			id := static_single_assignment.IdentifierId(class*perClass + member + 1)
			function.Identifiers = append(function.Identifiers, nil)
			ranges.Set(id, mutation_aliasing.MutableRange{Start: static_single_assignment.EvaluationOrder(id), End: static_single_assignment.EvaluationOrder(id) + 2})
			members = append(members, id)
		}
		set.Union(members)
	}

	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	if scopes.Len() != classes {
		t.Fatalf("got %d scopes for %d classes", scopes.Len(), classes)
	}

	seen := map[static_single_assignment.IdentifierId]ScopeId{}
	total := 0
	for _, id := range scopes.Ids() {
		for _, member := range scopes.MembersOf(id) {
			if previous, already := seen[member]; already {
				t.Fatalf("value %d appears in scope %d and again in scope %d; the walk revisited",
					member, previous, id)
			}
			seen[member] = id
			total++
		}
	}
	if total != set.Size() {
		t.Errorf("scopes hold %d members but the set has %d; the walk is not one pass over each",
			total, set.Size())
	}
}

// TestScopeAssignmentIsIdempotent pins that running the pass twice changes nothing.
//
// `Construct` is not idempotent in this package, which is a trap a caller can walk into; this pass
// is, because it reads two tables and writes a third rather than mutating its input. Asserted rather
// than assumed, because the property is what makes the input table safe to reuse.
func TestScopeAssignmentIsIdempotent(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
		function Component(props) {
			const items = [];
			items.push(props.value);
			return <div>{items}</div>;
		}
	`)
	set := FindDisjointMutableValuesWithRanges(function, ranges)

	first := AssignReactiveScopesWithSets(function, ranges, set)
	second := AssignReactiveScopesWithSets(function, ranges, set)

	if first.Len() != second.Len() {
		t.Fatalf("second run produced %d scopes, first produced %d", second.Len(), first.Len())
	}
	for _, id := range first.Ids() {
		if first.RangeOf(id) != second.RangeOf(id) {
			t.Errorf("scope %d moved between runs", id)
		}
	}
}

// TestScopeHandlesANilFunction pins the degenerate inputs.
func TestScopeHandlesANilFunction(t *testing.T) {
	t.Parallel()

	if got := AssignReactiveScopes(nil); got == nil || got.Len() != 0 {
		t.Error("a nil function must produce an empty table rather than a nil one or a panic")
	}
	var empty *ReactiveScopes
	if empty.ScopeOf(1) != 0 || empty.Len() != 0 || empty.MembersOf(1) != nil {
		t.Error("the nil table must read as empty rather than panicking")
	}
	if got := empty.RangeOf(1); got.IsSet() {
		t.Error("the nil table must answer the unset range")
	}
	if got := empty.MemberRanges(); got == nil || got.Len() != 0 {
		t.Error("the nil table's rewritten ranges must be empty rather than nil")
	}
	if got := ValidateScopes(nil, nil); got != nil {
		t.Errorf("validating nothing returned %v", got)
	}
}

// TestScopeMembersOfIsStableAcrossCalls pins that a caller reading membership twice sees the same
// slice contents, and documents why the defensive copy in the pass is NOT load-bearing.
//
// # A mutation removing that copy SURVIVES, and it is a genuine equivalence rather than a gap
//
// The first version of this test wrote through `MembersOf` and then asserted the disjoint set was
// unharmed. It passed, and it could not have failed: `DisjointSet.Sets` allocates a fresh outer
// slice and fresh inner slices on EVERY call, so nothing a caller does to the result can reach the
// set. The test was asserting a property no implementation could violate, which is the shape the
// brief names as a fixture that is weaker than the property it claims to guard.
//
// The distinguishing input would have to be two readers sharing one `Sets()` result.
// `AssignReactiveScopesWithSets` calls `Sets()` itself (scopes.go, the `classes :=` line), so each
// invocation gets its own slices and no second holder exists. There is no such input, so the copy
// is equivalent to the alias and no fixture can score it.
//
// The copy is KEPT anyway, for the reason `disjoint.go` keeps its `PhiOperandsInOrder` read after
// measuring the same verdict: it costs one allocation per scope, it is what a caller reading
// `MembersOf` would assume, and the equivalence rests entirely on `Sets` continuing to allocate. A
// change there would make this load-bearing again with nothing to announce it. This test pins what
// IS observable -- that membership is stable and complete -- rather than pretending to pin the copy.
func TestScopeMembersOfIsStableAcrossCalls(t *testing.T) {
	t.Parallel()

	function := &Function{Identifiers: []*Identifier{nil, nil, nil, nil}}
	ranges := &mutation_aliasing.MutableRanges{}
	ranges.Set(1, mutation_aliasing.MutableRange{Start: 1, End: 4})
	ranges.Set(2, mutation_aliasing.MutableRange{Start: 2, End: 5})
	ranges.Set(3, mutation_aliasing.MutableRange{Start: 3, End: 6})
	set := &DisjointSet{}
	set.Union([]static_single_assignment.IdentifierId{1, 2, 3})

	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	scope := scopes.Ids()[0]

	first := scopes.MembersOf(scope)
	if len(first) != 3 {
		t.Fatalf("got %d members, want 3", len(first))
	}
	second := scopes.MembersOf(scope)
	for index := range first {
		if first[index] != second[index] {
			t.Errorf("membership moved between reads at index %d", index)
		}
	}
	// Membership must be complete and must agree with the per-value index, which is the property a
	// consumer actually depends on.
	for _, member := range first {
		if scopes.ScopeOf(member) != scope {
			t.Errorf("value %d is listed in scope %d but indexes to %d",
				member, scope, scopes.ScopeOf(member))
		}
	}
}

// ---------------------------------------------------------------------------
// The corpus: the measurements this stage exists to make
// ---------------------------------------------------------------------------

// corpusScopeStats walks the corpus once and returns what every measurement below reads.
func corpusScopeStats(t *testing.T, limit int) (functions, scopes, members int, widths, sizes []int,
	invalidHulls, emptyHulls, pastMax, someUnset, allUnset, widthOneMulti int) {
	t.Helper()

	var files []string
	err := filepath.Walk(corpusRoot(t), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the corpus: %v", err)
	}
	sort.Strings(files)
	if len(files) > limit {
		files = files[:limit]
	}
	if len(files) < 10 {
		t.Fatalf("the corpus holds only %d files; the path is probably wrong", len(files))
	}

	for _, path := range files {
		contents, readErr := os.ReadFile(path)
		if readErr != nil {
			continue
		}
		fileName := "/corpus/" + filepath.Base(path)
		probe := rule.Rule{
			Name:             "scopes-corpus",
			NeedsTypeChecker: true,
			Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{
					ast.KindSourceFile: func(node *ast.Node) {
						if ctx.TypeChecker == nil {
							t.Fatal("the typed harness handed this probe a nil checker, so every " +
								"count below would be a fact about the harness rather than the code")
						}
						forEachFunctionLike(node, func(functionNode *ast.Node) {
							function := Lower(functionNode, ctx.TypeChecker)
							if function == nil {
								return
							}
							Construct(function)
							functions++

							ranges := InferMutableRanges(function)
							set := FindDisjointMutableValuesWithRanges(function, ranges)
							assigned := AssignReactiveScopesWithSets(function, ranges, set)

							pastMax += len(ValidateScopes(function, assigned))

							for _, id := range assigned.Ids() {
								scopes++
								memberList := assigned.MembersOf(id)
								members += len(memberList)
								sizes = append(sizes, len(memberList))

								setCount := 0
								for _, member := range memberList {
									if ranges.Get(member).IsSet() {
										setCount++
									}
								}
								if setCount == 0 {
									allUnset++
								} else if setCount != len(memberList) {
									someUnset++
								}

								hull := assigned.RangeOf(id)
								switch {
								case hull.Start == 0 || hull.End == 0:
									emptyHulls++
								case hull.End <= hull.Start:
									invalidHulls++
								default:
									width := int(hull.End - hull.Start)
									widths = append(widths, width)
									if width == 1 && len(memberList) > 1 {
										widthOneMulti++
									}
								}
							}
						})
					},
				}
			},
		}
		rule_testing.RunTypedFiles(t, probe, map[string]string{fileName: string(contents)}, fileName)
	}
	return
}

// TestScopeWidthAndMembershipAgree is this stage's headline, pinned as a test.
//
// A count cannot distinguish genuine structure from catastrophic over-merging, so the measurement is
// a CROSS-TABULATION of scope width against scope membership. Over 400 corpus files and 685
// outermost functions:
//
//	                       width 1      width > 1
//	one member               1,903             80
//	many members                 0          1,323
//
// The empty cell is the result. Not one scope of several values hulls to a single instruction, which
// is exactly the shape over-merging would produce -- values swept together into a span too narrow to
// hold them. Two dimensions measured independently, and they agree everywhere.
//
// The counterfactual identified which mechanism produces which population. Neutralising
// `mayAllocate` to false collapses the width-one singletons from 1,903 to 193 while the wide
// multi-member scopes hold at 1,203 of 1,323. So allocation mints the narrow single-value scopes and
// the range and mutability gates produce the entangled wide ones: two populations, two causes,
// neither an artifact of the other.
//
// The thresholds below are far looser than those numbers. This guards the SHAPE, not the corpus,
// which moves when the tree does.
func TestScopeWidthAndMembershipAgree(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)

	functions, scopes, members, widths, sizes, _, _, _, _, _, widthOneMulti := corpusScopeStats(t, 400)

	sort.Ints(widths)
	sort.Ints(sizes)
	t.Logf("functions=%d scopes=%d members=%d", functions, scopes, members)
	if len(widths) > 0 {
		t.Logf("width p50=%d p90=%d max=%d", widths[len(widths)/2], widths[len(widths)*90/100],
			widths[len(widths)-1])
	}
	if len(sizes) > 0 {
		t.Logf("size p50=%d p90=%d max=%d", sizes[len(sizes)/2], sizes[len(sizes)*90/100],
			sizes[len(sizes)-1])
	}

	if functions < 100 {
		t.Fatalf("only %d functions were lowered; this test measured almost nothing", functions)
	}
	if scopes == 0 {
		t.Fatal("no scope was created, so the pass is inert")
	}

	// The headline. A multi-member scope whose hull is one instruction wide is the signature of
	// over-merging: several values cannot genuinely share a span too narrow to contain them.
	if widthOneMulti != 0 {
		t.Errorf("%d scopes hold several values in a one-instruction span, which is the shape of "+
			"catastrophic over-merging: values swept together into a hull too narrow to hold them",
			widthOneMulti)
	}

	// Two-sided. A pass that merged everything into one enormous scope would satisfy the check
	// above trivially, so the distribution has to show both populations really exist.
	if len(widths) > 0 && widths[len(widths)-1] < 5 {
		t.Errorf("the widest scope spans %d instructions; real code entangles over longer spans, "+
			"which suggests hulls are not being computed", widths[len(widths)-1])
	}
	if len(sizes) > 0 && sizes[len(sizes)/2] > 3 {
		t.Errorf("the median scope holds %d values; most values in real code are independent, so "+
			"this suggests over-merging", sizes[len(sizes)/2])
	}
	if len(sizes) > 0 && sizes[len(sizes)-1] < 5 {
		t.Errorf("the largest scope holds %d values, which suggests classes are not reaching this "+
			"pass intact", sizes[len(sizes)-1])
	}
}

// TestScopeRangesSatisfyUpstreamsInvariant pins the zero that made this stage safe to build.
//
// React ends the pass with a hard `CompilerError.invariant` (bundle line 32271) rejecting any scope
// whose range starts at zero, ends at zero, or ends past the last instruction. There is no branch
// that skips such a scope: upstream treats it as a compiler bug and stops.
//
// That is a real risk here rather than a formality. `RangeGapLoopCarriedInversion` leaves 48 values
// across 400 files with an unset range, and a class composed entirely of those would hull to [0,0)
// and trip the invariant. Measured: 3,278 classes have every member's range set, 28 have SOME member
// unset, and ZERO have every member unset -- so the invariant holds everywhere, and the 28 mixed
// classes are exactly what upstream's `start === 0` branch exists to handle.
//
// This test is the guard against a future change to `InferMutableRanges` widening that unset
// population until a whole class falls into it. It fails loudly rather than the pass emitting a
// scope whose range describes nothing.
func TestScopeRangesSatisfyUpstreamsInvariant(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)

	functions, scopes, _, _, _, invalidHulls, emptyHulls, pastMax, someUnset, allUnset, _ :=
		corpusScopeStats(t, 400)

	t.Logf("functions=%d scopes=%d someMemberUnset=%d allMembersUnset=%d",
		functions, scopes, someUnset, allUnset)
	t.Logf("emptyHulls=%d invalidHulls=%d pastMaxInstruction=%d", emptyHulls, invalidHulls, pastMax)

	if scopes == 0 {
		t.Fatal("no scope was created, so this test measured nothing")
	}
	if emptyHulls != 0 {
		t.Errorf("%d scopes hull to an empty range, which upstream treats as a compiler bug and "+
			"raises on; every member of those classes must have had an unset range", emptyHulls)
	}
	if invalidHulls != 0 {
		t.Errorf("%d scopes hull to an inverted range (end <= start), which no combination of valid "+
			"member ranges can produce", invalidHulls)
	}
	if pastMax != 0 {
		t.Errorf("%d scopes violate upstream's invariant as ValidateScopes reports it", pastMax)
	}
	// Two-sided: the mixed classes are what the `start === 0` branch exists for, and a corpus with
	// none of them would leave that branch unmeasured rather than proven correct.
	if someUnset == 0 {
		t.Log("no class mixes set and unset member ranges on this corpus, so upstream's " +
			"`start === 0` adoption branch is exercised only by TestScopeHullMatchesReact")
	}
}

// TestValidateScopesCatchesAnEmptyHull pins the branch the corpus cannot reach today.
//
// # This fixture exists because a mutation neutralising the zero-start check SURVIVED
//
// `TestScopeRangesSatisfyUpstreamsInvariant` measures zero empty hulls over 400 corpus files, which
// is the good news and is also why nothing in the suite could see that branch break. The branch is
// not dead: it is the guard that fires if `RangeGapLoopCarriedInversion`'s unset population ever
// grows until a whole equivalence class falls into it, at which point the hull is [0,0) and
// describes nothing.
//
// So the input is built by hand rather than lowered, because the corpus cannot produce it and the
// verdict "unreachable" would expire the moment ranges change. That is the distinguishing input the
// brief asks to be named before a fixture is written: a class every member of which has an unset
// range, which React's own merge loop hulls to [0,0) -- confirmed against the 7.1.1 bundle as the
// `allUnset` case in TestScopeHullMatchesReact.
//
// Upstream RAISES here (`CompilerError.invariant`, bundle line 32271). This returns the offenders
// instead, so the assertion is on which scopes are named rather than on a panic.
func TestValidateScopesCatchesAnEmptyHull(t *testing.T) {
	t.Parallel()

	// A function with real instructions, so maxInstruction is non-zero and the third condition is
	// not what fires. Without this the test would pass for the wrong reason.
	function, _ := rangesFor(t, `
		function Component(props) {
			const items = [];
			items.push(props.value);
			return <div>{items}</div>;
		}
	`)

	// Two values whose ranges are unset, unioned into one class. This is the shape an all-unset
	// loop-carried class would take.
	unsetA, unsetB := static_single_assignment.IdentifierId(1), static_single_assignment.IdentifierId(2)
	handmade := &mutation_aliasing.MutableRanges{}
	handmade.Set(unsetA, mutation_aliasing.MutableRange{Start: 0, End: 0})
	handmade.Set(unsetB, mutation_aliasing.MutableRange{Start: 0, End: 0})
	set := &DisjointSet{}
	set.Union([]static_single_assignment.IdentifierId{unsetA, unsetB})

	scopes := AssignReactiveScopesWithSets(function, handmade, set)
	if scopes.Len() != 1 {
		t.Fatalf("got %d scopes, want 1", scopes.Len())
	}
	if got := scopes.RangeOf(scopes.Ids()[0]); got.IsSet() {
		t.Fatalf("the all-unset class hulled to [%d,%d); React hulls it to [0,0)", got.Start, got.End)
	}

	invalid := ValidateScopes(function, scopes)
	if len(invalid) != 1 {
		t.Fatalf("ValidateScopes named %d scopes, want 1; the empty hull must be reported because "+
			"upstream treats it as a compiler bug and raises", len(invalid))
	}
	if invalid[0] != scopes.Ids()[0] {
		t.Errorf("ValidateScopes named scope %d, want %d", invalid[0], scopes.Ids()[0])
	}

	// Two-sided: a valid hull over the same function must NOT be named, or the check above would
	// pass for an implementation that condemns everything.
	valid := &mutation_aliasing.MutableRanges{}
	valid.Set(unsetA, mutation_aliasing.MutableRange{Start: 1, End: 3})
	valid.Set(unsetB, mutation_aliasing.MutableRange{Start: 2, End: 4})
	validSet := &DisjointSet{}
	validSet.Union([]static_single_assignment.IdentifierId{unsetA, unsetB})
	if named := ValidateScopes(function, AssignReactiveScopesWithSets(function, valid, validSet)); len(named) != 0 {
		t.Errorf("a valid scope was named as invalid: %v", named)
	}
}

// TestValidateScopesCatchesAStartOfZeroWithARealEnd pins the START half of the zero check.
//
// # This is the SECOND fixture written for one survivor, and the first one was wrong
//
// A mutation neutralising `scopeRange.Start == 0` survived, and it survived AGAIN after
// `TestValidateScopesCatchesAnEmptyHull` was added for it. That test hulls to [0,0), where
// `End == 0` fires first and short-circuits the whole disjunction, so the start check never gets to
// decide the case and neutralising it changes nothing. The hypothesis about why the mutant survived
// was right about the mechanism and wrong about which input could expose it, which is precisely the
// failure the brief names.
//
// The distinguishing input is a hull whose start is zero and whose end is NOT: [0,5). Only the start
// check can condemn it. That is reachable through the merge whenever a member carries a range of
// [0,n) -- start at the unset sentinel, end real -- which `MutableRange.IsSet` reports as a set
// range, so nothing upstream of here rejects it.
//
// Upstream carries the two tests as independent disjuncts for this reason, and reproducing them as
// one collapsed check would be a silent narrowing.
func TestValidateScopesCatchesAStartOfZeroWithARealEnd(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `
		function Component(props) {
			const items = [];
			items.push(props.value);
			return <div>{items}</div>;
		}
	`)

	ranges := &mutation_aliasing.MutableRanges{}
	ranges.Set(1, mutation_aliasing.MutableRange{Start: 0, End: 5})
	set := &DisjointSet{}
	set.Union([]static_single_assignment.IdentifierId{1})

	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	hull := scopes.RangeOf(scopes.Ids()[0])
	if hull.Start != 0 || hull.End != 5 {
		t.Fatalf("hull = [%d,%d), want [0,5); this test needs a zero start with a real end, which "+
			"is the only shape the start check alone can condemn", hull.Start, hull.End)
	}

	named := ValidateScopes(function, scopes)
	if len(named) != 1 {
		t.Fatalf("ValidateScopes named %d scopes for a hull of [0,5), want 1; a scope starting at "+
			"the unset sentinel describes no position and upstream raises on it", len(named))
	}
}

// TestValidateScopesCatchesAnEndOfZero pins the second half of the zero check separately.
//
// A hull with a real start and a zero end cannot arise from the merge -- the end only ever grows --
// but the condition is upstream's and is written as two independent tests, so it is scored as two.
func TestValidateScopesCatchesAnEndOfZero(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `
		function Component(props) {
			const items = [];
			items.push(props.value);
			return <div>{items}</div>;
		}
	`)
	ranges := &mutation_aliasing.MutableRanges{}
	ranges.Set(1, mutation_aliasing.MutableRange{Start: 5, End: 0})
	set := &DisjointSet{}
	set.Union([]static_single_assignment.IdentifierId{1})

	scopes := AssignReactiveScopesWithSets(function, ranges, set)
	if named := ValidateScopes(function, scopes); len(named) != 1 {
		t.Errorf("ValidateScopes named %d scopes for a hull ending at zero, want 1", len(named))
	}
}

// TestScopeGapsAreDeclared pins the gap list, so closing one is a visible event.
func TestScopeGapsAreDeclared(t *testing.T) {
	t.Parallel()

	gaps := ScopeGaps()
	if len(gaps) != 1 || gaps[0] != ScopeGapPostAlignmentWidening {
		t.Errorf("ScopeGaps() = %v; a change here means a gap opened or closed and the package "+
			"comment plus the consumers must be re-read", gaps)
	}
}

// TestScopeMatchesTheDisjointSetPartition pins that this pass adds no merging of its own.
//
// The entanglement question is answered entirely upstream of this file. This pass must reproduce
// that partition exactly: same number of classes, same membership. A pass that merged two classes
// while hulling them would produce plausible ranges and a wrong answer, and no range assertion could
// see it.
func TestScopeMatchesTheDisjointSetPartition(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)

	var files []string
	filepath.Walk(corpusRoot(t), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
			files = append(files, path)
		}
		return nil
	})
	sort.Strings(files)
	if len(files) > 120 {
		files = files[:120]
	}

	compared := 0
	for _, path := range files {
		contents, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		fileName := "/corpus/" + filepath.Base(path)
		probe := rule.Rule{
			Name:             "scopes-partition",
			NeedsTypeChecker: true,
			Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{
					ast.KindSourceFile: func(node *ast.Node) {
						if ctx.TypeChecker == nil {
							t.Fatal("nil checker: this comparison would be vacuous")
						}
						forEachFunctionLike(node, func(functionNode *ast.Node) {
							function := Lower(functionNode, ctx.TypeChecker)
							if function == nil {
								return
							}
							Construct(function)
							ranges := InferMutableRanges(function)
							set := FindDisjointMutableValuesWithRanges(function, ranges)
							assigned := AssignReactiveScopesWithSets(function, ranges, set)

							classes := set.Sets()
							if len(classes) != assigned.Len() {
								t.Fatalf("%s: %d classes became %d scopes", fileName,
									len(classes), assigned.Len())
							}
							for index, class := range classes {
								scope := assigned.Ids()[index]
								got := assigned.MembersOf(scope)
								if len(got) != len(class) {
									t.Fatalf("%s: class %d holds %d values, its scope holds %d",
										fileName, index, len(class), len(got))
								}
								for member := range class {
									if got[member] != class[member] {
										t.Fatalf("%s: scope %d membership diverged from its class",
											fileName, scope)
									}
									if assigned.ScopeOf(class[member]) != scope {
										t.Fatalf("%s: value %d is in class %d but indexes to scope %d",
											fileName, class[member], index, assigned.ScopeOf(class[member]))
									}
								}
								compared++
							}
						})
					},
				}
			},
		}
		rule_testing.RunTypedFiles(t, probe, map[string]string{fileName: string(contents)}, fileName)
	}

	if compared < 100 {
		t.Fatalf("only %d classes were compared; this test measured almost nothing", compared)
	}
	t.Logf("compared %d classes against their scopes", compared)
}
