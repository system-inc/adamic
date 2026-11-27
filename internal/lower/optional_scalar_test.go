package lower

import (
	"os"
	"slices"
	"strings"
	"testing"
)

// A ?. read whose result the checker calls a number lowers as the maybe it is (optional_scalar.go, #0d56xpd):
// each use that observes it keeps it with no check, and each use that needs the number takes the check whose
// words name the hatch. Main stopped every row at lowering ("?. to a number, which would be number |
// undefined"). The programs are the oracle's, which holds each to Node (internal/oracle/optional_scalar_test.go).
func TestOptionalScalarReadLowers(t *testing.T) {
	t.Parallel()
	const absent = "a ?. read whose receiver the checker calls present was absent, through a non-null assertion"
	for name, check := range map[string]string{
		"u1_switch":       "",
		"u2_strict_equal": "",
		"u3_coalesce":     "",
		"u4_or":           "",
		"u5_bitwise":      "",
		"u6_field":        "",
		"u7_chain":        "",
		"s1_assertion":    absent + ": node.parent?.kind",
		"s1_return":       absent + ": node.parent?.kind",
		"s1_argument":     absent + ": node.parent?.kind",
		"s1_field":        absent + ": node.parent?.kind",
		"s1_element":      absent + ": node.parent?.kind",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("../oracle/testdata/optional_scalar/" + name + ".ts")
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerSourceAs(t, string(source), ".ts")
			if err != nil {
				t.Fatalf("want it lowered, got %v", err)
			}
			checked := slices.IndexFunc(program.Strings, func(text string) bool { return strings.HasPrefix(text, absent) })
			if check == "" && checked >= 0 {
				t.Fatalf("an observing use took the check: %q", program.Strings[checked])
			}
			if check != "" && !slices.Contains(program.Strings, check) {
				t.Fatalf("want the check %q among the program's strings, got %q", check, program.Strings)
			}
		})
	}
}
