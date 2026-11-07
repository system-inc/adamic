package lower

import (
	"errors"
	"strings"
	"testing"
)

// A function written in place whose every return is provably undefined isn't a lying void function, so
// it widens freely, then's and catch's slots included (#9fj31n7, Adamic's ruling on #tfycfvw).
func TestVoidProvablyUndefinedWidens(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source string }{
		{
			"forEach with no parameter",
			"const mark: (i: number) => void = (i) => i;\nconst all: void[] = [1, 2].map((i) => mark(i));\nlet n = 0;\nall.forEach(() => { n += 1; });\nconsole.log(`${n}`);",
		},
		{"bare return", `const run: () => unknown = () => { console.log("ran"); return; }; run();`},
		{"return undefined", `const run: () => unknown = () => { console.log("ran"); return undefined; }; run();`},
		{"concise void", `const run: () => unknown = () => void console.log("ran"); run();`},
		{
			"nested return is its own",
			`const run: () => unknown = () => { const inner = (): number => 1; console.log(String(inner())); }; run();`,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			lowersAndAgreesWithNode(test, probe.source)
		})
	}
}

// A body that could return a value or a thenable stays refused: Node would keep the value, or follow a
// thenable's then and take its rejecting branch.
func TestVoidCouldReturnStaysRefused(test *testing.T) {
	test.Parallel()
	lying := `const make: () => void = () => ({ then: (ok: (v: unknown) => void, no: (e: unknown) => void): void => { no("lied"); } });
async function done(): Promise<void> {}
`
	for _, probe := range []struct{ name, source string }{
		{"concise call", lying + `await done().then(() => make()).catch(() => { console.log("rejected"); });`},
		{"function value", lying + `await done().then(make).catch(() => { console.log("rejected"); });`},
		{
			"block returning a call",
			lying + `await done().then(() => { return make(); }).catch(() => { console.log("rejected"); });`,
		},
	} {
		for _, extension := range []string{".a", ".ts"} {
			test.Run(probe.name+extension, func(test *testing.T) {
				test.Parallel()
				_, lowerVoidProbeError := lowerVoidProbe(test, extension, probe.source)
				var refused *Refused
				if !errors.As(lowerVoidProbeError, &refused) ||
					!strings.Contains(refused.Fix, "adamic/no-void-widening") &&
						!strings.Contains(refused.Fix, "adamic/no-void-read") {
					test.Fatalf("want it refused as a void function seen wider, got %v", lowerVoidProbeError)
				}
			})
		}
	}
}

// then's and catch's callbacks written in place are admitted by this rule too. Today those programs
// stop first on then and catch read through the library's prototype (unbound-method), which isn't this
// rule's, so the move is held where the rule stands.
func TestVoidProvablyUndefinedThenAndCatchAreNotWidened(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"then", `async function done(): Promise<void> {}
await done().then(() => { console.log("then"); });
console.log("after");`},
		{"catch", `async function fail(): Promise<void> { throw new Error("no"); }
await fail().catch((e) => { console.log("caught"); });
console.log("after");`},
	} {
		for _, extension := range []string{".a", ".ts"} {
			test.Run(probe.name+extension, func(test *testing.T) {
				test.Parallel()
				_, lowerVoidProbeError := lowerVoidProbe(test, extension, probe.source)
				var refused *Refused
				if errors.As(lowerVoidProbeError, &refused) &&
					strings.Contains(refused.Fix, "adamic/no-void-widening") {
					test.Fatalf("want the callback admitted, got %v", lowerVoidProbeError)
				}
			})
		}
	}
}
