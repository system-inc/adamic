package lower

import (
	"os"
	"testing"
)

func TestClockGenericReturnsT03(t *testing.T) {
	source, err := os.ReadFile("../oracle/testdata/clock_generic_returns_t_03.a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lowerSource(t, string(source)); err != nil {
		t.Fatal(err)
	}
}

func TestClockGenericReturnsT03Unsupported(t *testing.T) {
	for name, source := range map[string]string{
		"indexed":           `interface A { [key: string]: string; } interface B { readonly values: readonly string[]; } function f(): A & B { throw "stop"; } f();`,
		"callable":          `interface A { (): string; } interface B { readonly values: readonly string[]; } function f(): A & B { throw "stop"; } f();`,
		"container":         `interface B { readonly values: readonly string[]; } function f(): Map<string, string> & B { throw "stop"; } f();`,
		"unproven elements": `interface A { readonly name: string; } interface B { readonly values: readonly object[]; } function f(): A & B { throw "stop"; } f();`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := lowerSource(t, source); err == nil {
				t.Fatal("unsupported result lowered")
			}
		})
	}
}
