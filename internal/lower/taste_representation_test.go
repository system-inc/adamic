package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestTasteRepresentationLimitsStayExplicit(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"union scalar field", `const box: {value: number | string} = {value: 1}; box.value = 3; console.log(String(box.value));`, "a narrowed scalar in a boxed union field"},
		{"evolving different objects", `let value; value = {a: 1}; value = {b: "b"};`, "a value of type any"},
		{"explicit any", `let value: any; value = {a: 1};`, "explicit any in .a"},
		{"computed enum", `function next(): number { return 1; } enum Code {Value = next()}`, "a computed enum member"},
		{"enum prototype", `enum Code {__proto__ = 1}`, "an enum member name that changes the prototype"},
		{"merged enum", `enum Code {First = 1} enum Code {Second = 2}`, "merged enum declarations"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			var refused *Refused
			kind := errors.As(err, &notYet)
			if probe.name == "enum prototype" || probe.name == "explicit any" {
				kind = errors.As(err, &refused)
			}
			if !kind || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want NotYet %q, got %v", probe.reason, err)
			}
		})
	}
}
