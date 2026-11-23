package lower

import (
	"errors"
	"strings"
	"testing"
)

// A step of a chain optionalChain guards has the type it has in the guard's present arm (#cdpxjmb): a
// number field read on is a number, so counted?.n.toFixed(1) calls a number's toFixed, where the field's
// type in tsc's chain is number | undefined and read as that it would be a maybe no number method takes;
// and a function value's call read on, make?.().x, reads its result's field. Each agrees with Node on the
// JavaScript backend and natively under ASan and UBSan (the oracle's testdata/optional_chain holds the
// ruling's shapes). Mutant: typeOf's clause for a guarded chain's inner step dropped, and the toFixed row
// stops or disagrees.
func TestOptionalChainStepsHaveTheirOwnTypes(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"a number field read on", "interface Counted {\n\treadonly n: number;\n}\nfunction fixed(counted: Counted | undefined): string {\n\treturn String(counted?.n.toFixed(1));\n}\nconsole.log(`${fixed(undefined)} ${fixed({ n: 2 })}`);\n", "undefined 2.0\n"},
		{"a function value's call read on", "interface Box {\n\treadonly x: number;\n}\nfunction run(make: (() => Box) | undefined): string {\n\treturn String(make?.().x);\n}\nconsole.log(`${run(undefined)} ${run((): Box => ({ x: 1 }))}`);\n", "undefined 1\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := lowersAndAgreesWithNodeAs(t, probe.source, ".a")
			// Agreement can't tell the guard from another road that happened to print the same, so the IR is
			// read for the guard's held receiver: these rows exist to reach optionalChain.
			guarded := false
			for _, local := range program.Locals {
				guarded = guarded || local.Name == "optional_chain_receiver"
			}
			if !guarded {
				t.Fatal("no optional chain receiver held: the chain took another road")
			}
			if code, out, errs := brandedNative(t, program); code != 0 || out != probe.want || errs != "" {
				t.Fatalf("native: exit %d stdout %q stderr %q; want exit 0 and %q, as Node prints", code, out, errs, probe.want)
			}
		})
	}
}

// What optionalChain doesn't take keeps its stop: an optional method's call read on, whose method may be
// missing where its object is there, so its guard would have to test the lookup, in property's words as
// before the guard; and an await inside a chain, which the JavaScript backend's arrow function around the
// guard can't hold, in the guard's own.
func TestOptionalChainGapsStop(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"an optional method's call read on", "interface Inner {\n\treadonly x: number;\n}\ninterface Maker {\n\tmake?(): Inner;\n}\nfunction made(maker: Maker | undefined): string {\n\treturn String(maker?.make?.().x);\n}\nconsole.log(made({ make: (): Inner => ({ x: 1 }) }));\n", "an optional chain longer than one step"},
		{"an await inside a chain", "interface Box {\n\treadonly items: number[];\n}\nasync function one(): Promise<number> {\n\treturn 1;\n}\nasync function grow(box: Box | undefined): Promise<string> {\n\treturn String(box?.items.push(await one()));\n}\nconsole.log(await grow({ items: [] }));\n", "an await or a yield inside an optional chain"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".a")
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.want) {
				t.Fatalf("want NotYet %q, got %v", probe.want, err)
			}
		})
	}
}
