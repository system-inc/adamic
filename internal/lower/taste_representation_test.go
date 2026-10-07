package lower

import (
	"errors"
	"strings"
	"testing"
)

func TestTasteRepresentationLimitsStayExplicit(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, reason string }{
		{"optional string insertion", `const box: {value?: string} = {}; box.value = "x";`, "writing a possibly absent optional own field"},
		{"optional array insertion", `const box: {value?: string[]} = {}; box.value = [];`, "writing a possibly absent optional own field"},
		{"optional boolean insertion", `const box: {value?: boolean} = {}; box.value = true;`, "writing a possibly absent optional own field"},
		{"union scalar field", `const box: {value: number | string} = {value: 1}; box.value = 3; console.log(String(box.value));`, "a narrowed scalar in a boxed union field"},
		{"union array search", `const values: (number | string)[] = [1, "x"]; console.log(String(values.includes(1)));`, "searching boxed union array elements"},
		{"evolving different objects", `let value; value = {a: 1}; value = {b: "b"};`, "a value of type any"},
		{"explicit any", `let value: any; value = {a: 1};`, "a value of type any"},
		{"computed enum", `function next(): number { return 1; } enum Code {Value = next()}`, "a computed enum member"},
		{"merged enum", `enum Code {First = 1} enum Code {Second = 2}`, "merged enum declarations"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("want NotYet %q, got %v", probe.reason, err)
			}
		})
	}
}

func TestTastePresentOptionalReferenceWrite(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `const box: {value?: string} = {value: "old"}; box.value = "new"; console.log(box.value);`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTasteEnumPrototypeRefusal(t *testing.T) {
	_, err := lowerSource(t, `enum Code {__proto__ = 1}`)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "enum member name that changes the prototype") {
		t.Fatalf("want prototype refusal, got %v", err)
	}
}
