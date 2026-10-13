package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestObjectReplacedBindingsUseChecks(t *testing.T) {
	t.Parallel()
	for _, write := range []string{
		`value = hidden;`,
		`[value] = [hidden];`,
		`({ value } = { value: hidden });`,
		`for (value of [hidden]) { }`,
	} {
		source := `let value = { n: 1 }; const hidden = { n: 2, extra: 'wrong' }; ` + write + ` Object.values(value);`
		t.Run(write, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, source)
			if write == `value = hidden;` || write == `[value] = [hidden];` {
				if err != nil {
					t.Fatal(err)
				}
				checked := false
				walk(program.Main, func(node any) bool {
					if call, ok := node.(ir.ObjectCall); ok && call.Method == "values" {
						checked = call.Checked
					}
					return true
				})
				if !checked {
					t.Fatal("replaced binding lost its enumeration check")
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("got %v, want NotYet", err)
			}
		})
	}
}

// A replacement can supply conversion code or descriptors absent from the view.
// The literal initializer proves neither after that replacement.
func TestObjectReplacedBindingCannotSupplyACompleteShape(t *testing.T) {
	t.Parallel()
	for _, observed := range []string{
		"Object.getOwnPropertyNames(value);",
		"Number(value);",
		"Object.hasOwn(value, 'n');",
	} {
		source := "let value = { n: 1 }; class Hidden { n: number = 2; toString(): string { return '7'; } } value = new Hidden(); " + observed
		_, err := lowerSource(t, source)
		var gap *NotYet
		if !errors.As(err, &gap) {
			t.Errorf("%s: got %v, want NotYet for a replaced shape", observed, err)
		}
	}
}

func TestObjectConversionGetterStaysRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `let calls = 0; const value = { get toString(): boolean { calls++; return false; }, valueOf: () => 7 }; console.log(String(value)); console.log(String(calls));`)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.Error(), "conversion getter") {
		t.Fatalf("got %v, want NotYet for a conversion getter", err)
	}
}
