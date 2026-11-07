package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestTasteRepresentationLimitsStayExplicit(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"evolving different objects", `let value; value = {a: 1}; value = {b: "b"};`, "a value of type any"},
		{"explicit any", `let value: any; value = {a: 1};`, "a value of type any"},
		{"computed enum", `function next(): number { return 1; } enum Code {Value = next()}`, "a computed enum member"},
		{"enum prototype", `enum Code {__proto__ = 1}`, "an enum member name that changes the prototype"},
		{"merged enum", `enum Code {First = 1} enum Code {Second = 2}`, "merged enum declarations"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			var notYet *NotYet
			var refused *Refused
			kind := errors.As(lowerSourceError, &notYet)
			if probe.name == "enum prototype" {
				kind = errors.As(lowerSourceError, &refused)
			}
			if !kind || !strings.Contains(lowerSourceError.Error(), probe.reason) {
				test.Fatalf("want NotYet %q, got %v", probe.reason, lowerSourceError)
			}
		})
	}
}

func TestTasteBoxedScalarFieldNowLowers(test *testing.T) {
	test.Parallel()
	if _, lowerSourceError := lowerSource(
		test,
		`const box: {value: number | string} = {value: 1}; box.value = 3; console.log(String(box.value));`,
	); lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
}
