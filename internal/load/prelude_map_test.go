package load

import (
	"strings"
	"testing"
)

// The prelude's argumentless MapConstructor overload is merged ahead of the bundled Map<any, any>
// one, so an uncontextual new Map() is Map<unknown, unknown>, and it can't be handed on as a typed
// map. Mutant: the prelude's MapConstructor removed, which gives Map<any, any> and admits the typed
// slot below.
func TestPreludeArgumentlessMapIsUnknownWithoutContext(t *testing.T) {
	t.Parallel()
	byName := declarationsByName(t, writeProgram(t, [2]string{"main.ts", `const loose = new Map();
`}))
	if got := byName["variable loose"].Type; got != "Map<unknown, unknown>" {
		t.Errorf("variable loose: got %q, want %q", got, "Map<unknown, unknown>")
	}
	diagnostics := checkErrors(t, writeProgram(t, [2]string{"main.ts", `const loose = new Map();
const typed: Map<string, number> = loose;
`}))
	if len(diagnostics) != 1 || !strings.Contains(diagnostics[0], "main.ts:2:7: error TS2322: Type 'Map<unknown, unknown>' is not assignable to type 'Map<string, number>'.") {
		t.Errorf("got %q, want one TS2322 at main.ts:2:7", diagnostics)
	}
}

// With a slot to fill, new Map() takes its arguments from the slot: a declared local, a return, a
// class field and ??= into an optional map each load clean, as they do on the bundled overload.
func TestPreludeArgumentlessMapTakesItsSlotsArguments(t *testing.T) {
	t.Parallel()
	byName := declarationsByName(t, writeProgram(t, [2]string{"main.ts", `const declared: Map<string, number> = new Map();
function make(): Map<string, boolean> { return new Map(); }
class Holder { cache: Map<string, number> = new Map(); }
let optional: Map<string, number[]> | undefined;
(optional ??= new Map()).set('k', [1]);
`}))
	if got := byName["variable declared"].Type; got != "Map<string, number>" {
		t.Errorf("variable declared: got %q, want %q", got, "Map<string, number>")
	}
}
