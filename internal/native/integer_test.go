package native

import (
	"fmt"
	"math"
	"math/rand/v2"
	"path/filepath"
	"strings"
	"testing"
)

// Read binary64 bits, avoiding decimal parsing as a second source of rounding. Node's
// Object.is observes the zero sign. Arbitrary exponents exercise both cast guards and libm.
func TestIntegerRuntimeMatchesNode(t *testing.T) {
	t.Parallel()
	operations := []struct{ name, native, javascript string }{
		{"remainder", "adamic_remainder(left, right)", "left % right"},
		{"and", "adamic_bitwise_and(left, right)", "left & right"},
		{"or", "adamic_bitwise_or(left, right)", "left | right"},
		{"xor", "adamic_bitwise_xor(left, right)", "left ^ right"},
		{"not", "adamic_bitwise_not(left)", "~left"},
		{"left", "adamic_shift_left(left, right)", "left << right"},
		{"right", "adamic_shift_right(left, right)", "left >> right"},
		{"unsigned", "adamic_shift_right_unsigned(left, right)", "left >>> right"},
	}
	var dispatch, cases strings.Builder
	for _, operation := range operations {
		fmt.Fprintf(&dispatch, "if (strcmp(operation, %q) == 0) { number(%s); } else ", operation.name, operation.native)
		fmt.Fprintf(&cases, "case %q: out.push(number(%s)); break;\n", operation.name, operation.javascript)
	}
	harness := mathHarness[:strings.Index(mathHarness, "\t\tif (strcmp(operation,")] + dispatch.String() + "{ return 1; }\n\t}\n\treturn 0;\n}\n"
	oracle := mathOracle[:strings.Index(mathOracle, "\tswitch (operation)")] + "switch (operation) {\n" + cases.String() + "}\n}\nprocess.stdout.write(out.join('\\n') + '\\n');\n"
	values := []float64{0, math.Copysign(0, -1), 1, -1, 2, -2, .5, -.5, 1.5, -1.5, math.NaN(), math.Inf(1), math.Inf(-1), math.SmallestNonzeroFloat64, math.MaxFloat64}
	for _, power := range []float64{1 << 31, 1 << 32, 1 << 53} {
		for _, offset := range []float64{-2, -1, -.5, 0, .5, 1, 2} {
			values = append(values, power+offset, -power-offset)
		}
	}
	var input strings.Builder
	ask := func(left, right float64) {
		for _, operation := range operations {
			fmt.Fprintf(&input, "%s %016x %016x\n", operation.name, math.Float64bits(left), math.Float64bits(right))
		}
	}
	for _, left := range values {
		for _, right := range values {
			ask(left, right)
		}
	}
	random := rand.New(rand.NewPCG(31, 53))
	for range 10000 {
		ask(math.Float64frombits(random.Uint64()), math.Float64frombits(random.Uint64()))
		ask(float64(random.Int64N(1<<54)-(1<<53)), float64(random.Int64N(1<<54)-(1<<53)))
	}
	binary := filepath.Join(t.TempDir(), "integer")
	if err := Build(harness, binary, Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := runWithInput(t, input.String(), binary)
	want := runWithInput(t, input.String(), "node", "--eval", oracle)
	if got != want {
		actual, expected := strings.Split(got, "\n"), strings.Split(want, "\n")
		questions := strings.Split(input.String(), "\n")
		if len(actual) != len(expected) {
			t.Fatalf("native %d lines, Node %d", len(actual), len(expected))
		}
		for index := range expected {
			if actual[index] != expected[index] {
				t.Fatalf("%s: native %q, Node %q", questions[index], actual[index], expected[index])
			}
		}
	}
	t.Logf("%d answers match Node byte for byte", strings.Count(want, "\n"))
}
