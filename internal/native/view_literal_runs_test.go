package native

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// viewLiteralCheck is the line a checked view's read of node.kind, its literal set allowed, holds its
// value to, emitted from a receiver already evaluated, and all the read emitted. The slot is the read's
// first temporary.
func viewLiteralCheck(t *testing.T, of ir.Type, allowed []ir.Expression, constants []string) (string, string) {
	t.Helper()
	e := &emitter{program: &ir.Program{Strings: constants}}
	e.viewFieldObject(ir.Property{Name: "kind", Of: of, View: "node.kind", ViewType: "Kind", ViewAllowed: allowed}, "receiver")
	checks := []string{}
	for _, line := range strings.Split(e.out.String(), "\n") {
		if strings.Contains(line, "adamic_view_literal_failure(") {
			checks = append(checks, strings.TrimSpace(line))
		}
	}
	if len(checks) != 1 {
		t.Fatalf("want one literal check, got %d:\n%s", len(checks), e.out.String())
	}
	return checks[0], e.out.String()
}

// viewLiteralLine is the check line as the read writes it around tests, joined with ||.
func viewLiteralLine(of ir.Type, tests []string) string {
	return fmt.Sprintf("if (!(%s)) adamic_view_literal_failure(%s, %s, %d, adamic_temporary_1);", strings.Join(tests, " || "), cString("node.kind"), cString("Kind"), of)
}

func numberLiterals(values ...float64) []ir.Expression {
	literals := []ir.Expression{}
	for _, value := range values {
		literals = append(literals, ir.NumberConstant{Value: value})
	}
	return literals
}

// sameNumber is one equality test of the chain, as the read spells it for a number literal.
func sameNumber(value float64) string {
	return fmt.Sprintf("adamic_same_number(adamic_temporary_1.number, %s)", cNumber(value))
}

// numberRun is one run's range test, its bounds written as every number constant is.
func numberRun(lowest, highest float64) string {
	return fmt.Sprintf("(adamic_temporary_1.number >= %s && adamic_temporary_1.number <= %s && adamic_temporary_1.number == trunc(adamic_temporary_1.number))", cNumber(lowest), cNumber(highest))
}

// A checked view's literal set of three or more consecutive integers is one range test per run
// (#zpzm9ta): 400 contiguous members, given in reverse, are one run and no equality at all, and a
// holed set is two runs. A set with a run and stragglers is sorted, the stragglers' equalities in
// place between the runs. Mutants: viewLiteralRuns returning nil at once turns every case here red;
// the run's length test written > 3 turns the three-member run red.
func TestViewLiteralRunsAreRangeTests(t *testing.T) {
	t.Parallel()
	contiguous := []float64{}
	for value := 399; value >= 0; value-- {
		contiguous = append(contiguous, float64(value))
	}
	holed := []float64{}
	for value := 1; value <= 20; value++ {
		if value != 11 {
			holed = append(holed, float64(value))
		}
	}
	for _, probe := range []struct {
		name    string
		allowed []float64
		tests   []string
	}{
		{"400 contiguous", contiguous, []string{numberRun(0, 399)}},
		{"a hole at 11", holed, []string{numberRun(1, 10), numberRun(12, 20)}},
		{"a run and stragglers", []float64{9, 2, 1, 3, 5, 6, 2}, []string{numberRun(1, 3), sameNumber(5), sameNumber(6), sameNumber(9)}},
		{"negative and zero", []float64{0, -1, -2, 4}, []string{numberRun(-2, 0), sameNumber(4)}},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			check, emitted := viewLiteralCheck(t, ir.Number, numberLiterals(probe.allowed...), nil)
			if want := viewLiteralLine(ir.Number, probe.tests); check != want {
				t.Fatalf("got\n%s\nwant\n%s", check, want)
			}
			if len(probe.tests) == 1 && strings.Contains(emitted, "adamic_same_number") {
				t.Fatalf("a single run still emitted an equality:\n%s", emitted)
			}
		})
	}
}

// A set with no run of three or more emits today's chain byte for byte, in the set's order: under
// three members, three that aren't consecutive, a run broken by a non-integer, NaN or an infinity,
// and a string set. Mutant: viewLiteralRuns sorting and splitting without asking whether a run
// formed reorders {5, 3, 1} and turns it red.
func TestViewLiteralSetsWithoutARunKeepTheChain(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name   string
		values []float64
	}{
		{"two", []float64{1, 2}},
		{"odd", []float64{1, 3, 5}},
		{"odd, given backwards", []float64{5, 3, 1}},
		{"a non-integer", []float64{1, 2, 3, 3.5}},
		{"NaN", []float64{1, 2, 3, math.NaN()}},
		{"an infinity", []float64{1, 2, 3, math.Inf(1)}},
		{"past the safe integers", []float64{9007199254740990, 9007199254740991, 9007199254740992}},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			check, _ := viewLiteralCheck(t, ir.Number, numberLiterals(probe.values...), nil)
			tests := []string{}
			for _, value := range probe.values {
				tests = append(tests, sameNumber(value))
			}
			if want := viewLiteralLine(ir.Number, tests); check != want {
				t.Fatalf("got\n%s\nwant\n%s", check, want)
			}
		})
	}
	t.Run("strings", func(t *testing.T) {
		t.Parallel()
		literals := []string{"b", "a", "c"}
		allowed := []ir.Expression{}
		tests := []string{}
		for index, literal := range literals {
			allowed = append(allowed, ir.StringConstant{Index: index})
			tests = append(tests, fmt.Sprintf("adamic_string_equal(adamic_temporary_1.reference, &%s)", stringName(literal)))
		}
		check, _ := viewLiteralCheck(t, ir.String, allowed, literals)
		if want := viewLiteralLine(ir.String, tests); check != want {
			t.Fatalf("got\n%s\nwant\n%s", check, want)
		}
	})
}
