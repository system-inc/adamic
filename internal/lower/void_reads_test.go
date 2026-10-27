package lower

import (
	"errors"
	"strings"
	"testing"
)

// make returns 42 through a void type, so Node reads 42 wherever a void value is read and Adamic, which
// releases a void call's result (#f418fbz), would read undefined. Each read form is refused in .a and
// .ts with the role it plays (#x2gt39p).
const voidReadMake = `const make: () => void = () => 42;`

func TestVoidReadsAreRefused(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, role string }{
		// Language's probe: Node prints false, and both of our backends printed true.
		{"bound and read", `const value = make(); console.log(` + "`${value === undefined}`" + `);`, "as an operand of ==="},
		{"compared", "console.log(`${make() === undefined}`);", "as an operand of ==="},
		{"passed", `function show(x: unknown): void { console.log(typeof x); } show(make());`, "as an argument"},
		{"typeof", `console.log(typeof make());`, "as typeof's operand"},
		{"template", "console.log(`${make()}`);", "as a template literal part"},
		{"assigned", `let slot: unknown = 0; slot = make(); console.log(typeof slot);`, "as an assigned value"},
		{"bound wider", `const kept: unknown = make(); console.log(typeof kept);`, "as a binding's initializer"},
		{"array element", `const all = [make()]; console.log(` + "`${all.length}`" + `);`, "as a value"},
		{"conditional branch", `const pick = (go: boolean): unknown => go ? make() : 1; console.log(typeof pick(true));`, "as a returned value"},
		{"async awaited", `async function later(): Promise<void> { return make(); } const got = await later(); console.log(typeof got);`, "as typeof's operand"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerVoidProbe(t, extension, voidReadMake+"\n"+probe.source)
				var refused *Refused
				if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, "a read of a void value") || !strings.Contains(refused.What, probe.role) || !strings.Contains(refused.Fix, "call it as a statement, or give the function a return type") {
					t.Fatalf("want a void read refused %s, got %v", probe.role, err)
				}
			})
		}
	}
}

// What stays legal: a void call as a statement, void's operand, the left of a comma, a loop's
// initializer and incrementor, #se1zcgr's discards (return f() and a concise arrow body in an exactly
// void function, an async one included), a void binding nobody reads, and relating a value-returning
// function to a void function type.
func TestVoidDiscardsStayLegal(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"statement", `make(); console.log("done");`},
		{"void operator", `void make(); console.log("done");`},
		{"comma", "let n = 0; n = (make(), 1); console.log(`${n}`);"},
		{"loop", `for (make(); false; make()) { } console.log("done");`},
		{"return in void", `function run(): void { return make(); } run(); console.log("done");`},
		{"concise arrow in void", `const go = (): void => make(); go(); console.log("done");`},
		{"async return in void", `async function later(): Promise<void> { return make(); } await later(); console.log("done");`},
		{"unread binding", `const value = make(); console.log("done");`},
		// The relation H2 keeps: push returns a number, and forEach's callback type returns void.
		{"callback relation", "const kept: number[] = []; [1, 2].forEach((x) => kept.push(x)); console.log(`${kept.length}`);"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNode(t, voidReadMake+"\n"+probe.source)
		})
	}
}

// A primitive brand's member is void so the brand stays phantom: a primitive carries no own property,
// so reading it is undefined on Node and here, and no void call ever stood behind it.
func TestVoidBrandReadStaysLegal(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNode(t, `type Escaped = string & { __escapedIdentifier: void }; const name = ('na' + 'me') as Escaped; console.log(String(name.__escapedIdentifier));`)
}

func lowerVoidProbe(t *testing.T, extension, source string) (any, error) {
	t.Helper()
	if extension == ".ts" {
		return lowerTypeScriptAssertionSource(t, source)
	}
	return lowerSource(t, source)
}
