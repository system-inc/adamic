package lower

import "testing"

// x ?? await g() is split by NormalizeAsync into a test and two arms, since the await suspends only when x
// is missing. The test has to be ?? 's own idea of missing: undefined or null. It once tested undefined
// alone, so a null took the present arm, whose Coalesce fell back to the very local it was assigning, and
// native printed 0 or "" where Node printed g()'s value (#wkfwexm, found by Lower on #smqr4vn). Each
// program below runs on Node, on the JavaScript backend and as sanitized native, and the three must print
// the same lines. Mutant: the split's test back to IsUndefined alone, and every null row reds natively.
func TestAsyncCoalesceSplitTreatsNullAsMissing(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"a number or null", `async function g(): Promise<number> { return 5; }
async function h(x: number | null): Promise<string> { const r = x ?? await g(); return String(r); }
console.log(await h(null));
console.log(await h(3));
console.log(await h(0));
`, "5\n3\n0\n"},
		{"a number, null or undefined", `async function g(): Promise<number> { return 5; }
async function h(x: number | null | undefined): Promise<string> { const r = x ?? await g(); return String(r); }
console.log(await h(null));
console.log(await h(undefined));
console.log(await h(3));
`, "5\n5\n3\n"},
		// The control the miscompile never reached: a number pair holds no null.
		{"a number or undefined", `async function g(): Promise<number> { return 5; }
async function h(x: number | undefined): Promise<string> { const r = x ?? await g(); return String(r); }
console.log(await h(undefined));
console.log(await h(3));
`, "5\n3\n"},
		{"a string or null", `async function g(): Promise<string> { return "fallback"; }
async function h(x: string | null): Promise<string> { const r = x ?? await g(); return r; }
console.log(await h(null));
console.log(await h("given"));
console.log(await h("") + "|");
`, "fallback\ngiven\n|\n"},
		{"a boolean or null", `async function g(): Promise<boolean> { return true; }
async function h(x: boolean | null): Promise<string> { const r = x ?? await g(); return String(r); }
console.log(await h(null));
console.log(await h(false));
`, "true\nfalse\n"},
		{"an object or null", `interface Box { readonly n: number; }
async function g(): Promise<Box> { return { n: 5 }; }
async function h(x: Box | null): Promise<string> { const r = x ?? await g(); return String(r.n); }
console.log(await h(null));
console.log(await h({ n: 3 }));
`, "5\n3\n"},
		// The same split inside a loop's body, where each pass runs it again.
		{"inside a loop", `async function g(): Promise<number> { return 5; }
async function h(xs: (number | null)[]): Promise<string> {
	let text = "";
	for (const x of xs) {
		const r = x ?? await g();
		text += String(r) + ",";
	}
	return text;
}
console.log(await h([null, 2, null, 0]));
`, "5,2,5,0,\n"},
		// A nested ??: the outer split's present arm holds the inner one's result.
		{"a nested coalesce", `async function g(): Promise<number> { return 5; }
async function h(x: number | null, y: number | null): Promise<string> { const r = x ?? y ?? await g(); return String(r); }
console.log(await h(null, null));
console.log(await h(null, 7));
console.log(await h(1, null));
`, "5\n7\n1\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNode(t, probe.source)
			code, stdout, stderr := brandedNative(t, program)
			if code != 0 || stdout != probe.want || stderr != "" {
				t.Fatalf("native: exit %d stdout %q stderr %q; want exit 0 and %q, as Node prints", code, stdout, stderr, probe.want)
			}
		})
	}
}
