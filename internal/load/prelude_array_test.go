package load

import (
	"strings"
	"testing"
)

// The prelude's isArray overload is merged ahead of lib.es5.d.ts's any[] predicate, so it is the one
// a call resolves to: unknown narrows to readonly unknown[], and a member that is already an array
// keeps its own type. Mutant: the prelude's ArrayConstructor removed, which narrows unknown to any[]
// and loads the typed read below.
func TestPreludeIsArrayNarrowsUnknownToReadonlyUnknown(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.ts", `function narrow(value: unknown, text: string | string[], view: string | readonly number[]): void {
	if (Array.isArray(value)) {
		const unknownArray = value;
	}
	if (Array.isArray(text)) {
		const strings = text;
	} else {
		const single = text;
	}
	if (Array.isArray(view)) {
		const numbers = view;
	}
}
`})
	byName := declarationsByName(t, paths)
	for name, want := range map[string]string{
		"variable unknownArray": "readonly unknown[]",
		"variable strings":      "string[]",
		"variable single":       "string",
		"variable numbers":      "readonly number[]",
	} {
		if got := byName[name].Type; got != want {
			t.Errorf("%s: got %q, want %q", name, got, want)
		}
	}
	diagnostics := checkErrors(t, writeProgram(t, [2]string{"main.ts", `function read(value: unknown): void {
	if (Array.isArray(value)) {
		const strings: string[] = value;
	}
}
`}))
	if len(diagnostics) != 1 || !strings.HasSuffix(diagnostics[0], "main.ts:3:9: error TS4104: The type 'readonly unknown[]' is 'readonly' and cannot be assigned to the mutable type 'string[]'.") {
		t.Errorf("got %q, want one TS4104 at main.ts:3:9", diagnostics)
	}
}
