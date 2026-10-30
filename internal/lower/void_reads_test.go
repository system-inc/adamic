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
		// tsc folds void | undefined to void, so an optional call of a void function is a void read too.
		{"optional call", `const maybe: (() => void) | undefined = make; const got = maybe?.(); console.log(` + "`${got === undefined}`" + `);`, "as an operand of ==="},
		{"cast and read", `const seen = make() as unknown; console.log(typeof seen);`, "as a binding's initializer"},
		// satisfies keeps the binding void, so the read is refused where the binding is read.
		{"satisfies and read", `const seen = make() satisfies unknown; console.log(typeof seen);`, "as typeof's operand"},
		// ??=, ||= and &&= read their target before they write: Node tests 42, and both backends tested
		// undefined, so ??= and ||= ran their right where Node didn't and &&= skipped it where Node ran it.
		{"nullish assignment target", `let v = make(); let n = 0; v ??= (n++, undefined); console.log(String(n));`, "as the target of ??="},
		{"or assignment target", `let v = make(); let n = 0; v ||= (n++, undefined); console.log(String(n));`, "as the target of ||="},
		{"and assignment target", `let v = make(); let n = 0; v &&= (n++, undefined); console.log(String(n));`, "as the target of &&="},
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
// initializer and incrementor, a cast or satisfies whose own place discards, #se1zcgr's discards
// (return f() and a concise arrow body in an exactly void function, an async one included), a void
// binding nobody reads, and relating a value-returning function to a void function type.
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
		// A bare cast or satisfies statement waits on stage 0 (TestVoidUnderAWrapperStatementIsNoRead),
		// so the pass-through is held to Node under void and the left of a comma.
		{"cast under void", `void (make() as unknown); console.log("done");`},
		{"satisfies under void", `void (make() satisfies unknown); console.log("done");`},
		{"cast left of a comma", "let n = 0; n = (make() as unknown, 1); console.log(`${n}`);"},
		{"satisfies left of a comma", "let n = 0; n = (make() satisfies unknown, 1); console.log(`${n}`);"},
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

// A cast, a satisfies and a .ts non-null hand on the value they wrap, so a void call under one as a
// statement is a discard, never H2's refusal (#chmg7wr). Each still waits on its own stage-0 gap (a
// cast as a statement, void! as never), which is a NotYet, a promise H2's Refused would have broken.
func TestVoidUnderAWrapperStatementIsNoRead(t *testing.T) {
	t.Parallel()
	for _, statement := range []string{`make() as unknown;`, `make() satisfies void;`, `make()!;`} {
		t.Run(statement, func(t *testing.T) {
			t.Parallel()
			_, err := lowerTypeScriptAssertionSource(t, voidReadMake+"\n"+statement+"\nconsole.log(\"done\");")
			var refused *Refused
			if errors.As(err, &refused) {
				t.Fatalf("want no refusal for a discarded void value, got %v", err)
			}
		})
	}
}

// ! hands on its value as a cast does, so binding it wider still reads it. Only .ts checks a non-null
// assertion; .a refuses ! itself.
func TestVoidReadThroughNonNullIsRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerVoidProbe(t, ".ts", voidReadMake+"\n"+`const kept: unknown = make()!; console.log(typeof kept);`)
	var refused *Refused
	if !errors.As(err, &refused) || refused.What != "a read of a void value, as a binding's initializer" {
		t.Fatalf("want a void read refused as a binding's initializer, got %v", err)
	}
}
