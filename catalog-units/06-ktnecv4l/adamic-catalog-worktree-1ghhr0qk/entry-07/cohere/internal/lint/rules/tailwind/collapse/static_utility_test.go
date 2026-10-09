package tailwind

import (
	"sort"
	"testing"
)

// TestFrameworkStaticsReproduceTheBaseTable is the gate #1mwetnq will delete rows behind.
//
// Every reading in `baseStatics` must be reproduced by walking the declarations in
// FrameworkStaticDeclarations, class for class rather than as a rate. A rate would let a row that
// stopped being answered hide behind 868 that still are, and a row silently dropped is the failure
// this whole port is guarded against.
func TestFrameworkStaticsReproduceTheBaseTable(t *testing.T) {
	t.Parallel()
	var missing, disagreeing []string

	for name, expected := range baseStatics {
		actual, found := FrameworkStaticReading(name)
		if !found {
			missing = append(missing, name)
			continue
		}
		if !readingsEqual(expected, actual) {
			disagreeing = append(disagreeing, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(disagreeing)

	if len(missing) > 0 {
		t.Errorf("%d of %d base statics have no framework registration, so deleting their rows would delete an answer: %v",
			len(missing), len(baseStatics), truncate(missing, 12))
	}
	for _, name := range disagreeing {
		expected := baseStatics[name]
		actual, _ := FrameworkStaticReading(name)
		t.Errorf("%s: base table reads %v#%d, walking its declarations reads %v#%d",
			name, expected.Order, expected.Count, actual.Order, actual.Count)
	}

	// The positive assertion, stated separately. A run where baseStatics were empty would pass every
	// check above by comparing nothing, which is the vacuous-sweep shape this slice has been bitten
	// by before.
	if len(baseStatics) < 800 {
		t.Fatalf("baseStatics holds %d entries; expected around 869, so this comparison measured far less than it appears to", len(baseStatics))
	}
	t.Logf("reproduced %d of %d base statics by walking declarations; the framework table holds %d",
		len(baseStatics)-len(missing)-len(disagreeing), len(baseStatics), len(FrameworkStaticDeclarations))
}

// TestFrameworkStaticsCoverTheDeprecatedRegistrations is the finding this component was built on.
//
// 21 statics are registered by the engine and absent from `baseStatics`, all of them deprecated
// utilities that `getClassList()` does not advertise. Three are written in the ahra tree today.
// Asserting them by name rather than by count is deliberate: a count keeps passing while the set
// underneath it changes, and the point is which classes were missing.
func TestFrameworkStaticsCoverTheDeprecatedRegistrations(t *testing.T) {
	t.Parallel()
	deprecated := map[string]Reading{
		"bg-gradient-to-r":  {Order: []int{199, 200}, Count: 2},
		"bg-gradient-to-br": {Order: []int{199, 200}, Count: 2},
		"bg-left-bottom":    {Order: []int{254}, Count: 1},
		"object-right-top":  {Order: []int{269}, Count: 1},
		"decoration-clone":  {Order: []int{250}, Count: 2},
		"break-words":       {Order: []int{291}, Count: 1},
		"overflow-ellipsis": {Order: []int{293}, Count: 1},
		"max-w-screen":      {Order: []int{46}, Count: 1},
	}

	for name, expected := range deprecated {
		if _, inBase := baseStatics[name]; inBase {
			t.Errorf("%s is now in baseStatics; this test asserts the gap the framework table closes, so the gap may have been closed elsewhere and this should be re-measured rather than loosened", name)
		}
		actual, found := FrameworkStaticReading(name)
		if !found {
			t.Errorf("%s is not in the framework table, so the deprecated registrations are no longer covered", name)
			continue
		}
		if !readingsEqual(expected, actual) {
			t.Errorf("%s reads %v#%d, expected %v#%d from the engine", name, actual.Order, actual.Count, expected.Order, expected.Count)
		}
	}
}

// TestFrameworkStaticsExcludeRepositoryUtilities is the other half of the split.
//
// `markdown-content` and `typing-dots` are ahra's own `@utility` blocks. A framework table carrying
// them would be one repository's tokens shipped as framework facts, which is the exact defect this
// whole port exists to remove, so it is asserted rather than left to the generator's refusal.
func TestFrameworkStaticsExcludeRepositoryUtilities(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"markdown-content", "typing-dots", "synthetic-static"} {
		if _, found := FrameworkStaticDeclarations[name]; found {
			t.Errorf("%q is a repository utility and must not be in the framework table", name)
		}
	}
}

// TestFrameworkStaticDeclarationsCountsAbsentValues pins the field that decides arity.
//
// PropertySort skips a declaration whose value is absent and counts one whose value is the empty
// string, because `--tw-foo:;` is valid CSS. Collapsing the two onto Go's empty string would
// undercount, and count is half the sort key, so the distinction is asserted on the real walk rather
// than trusted from the struct definition.
func TestFrameworkStaticDeclarationsCountsAbsentValues(t *testing.T) {
	t.Parallel()
	present := PropertySort(nodesFromStaticDeclarations([]StaticDeclaration{
		{Property: "--tw-probe", Value: "", ValuePresent: true},
	}))
	absent := PropertySort(nodesFromStaticDeclarations([]StaticDeclaration{
		{Property: "--tw-probe", Value: "", ValuePresent: false},
	}))

	if present.Count != 1 {
		t.Errorf("a declaration with an empty but present value counted %d, expected 1", present.Count)
	}
	if absent.Count != 0 {
		t.Errorf("a declaration with an absent value counted %d, expected 0", absent.Count)
	}
}

func readingsEqual(left Reading, right Reading) bool {
	if left.Count != right.Count || len(left.Order) != len(right.Order) {
		return false
	}
	for index := range left.Order {
		if left.Order[index] != right.Order[index] {
			return false
		}
	}
	return true
}

func truncate(values []string, count int) []string {
	if len(values) < count {
		return values
	}
	return values[:count]
}
