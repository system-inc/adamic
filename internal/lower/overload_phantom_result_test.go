package lower

import (
	"errors"
	"strings"
	"testing"
)

// overloadPhantomPrelude is tsc's Path and PathPathComponents as adaptation 40 shapes them: a phantom
// string brand, and an Array<Path> extension whose one member is a phantom brand (path.ts:457 as an
// interface, #z5d0jye step 3 (a)). The extension takes its element as a defaulted type argument, so a
// reference to it has its base's arguments and is held as its base (unbranded.go): a non-generic
// `extends Array<Path>` has none of its own, which lowering reads elements from.
const overloadPhantomPrelude = `type Path = string & { __pathBrand: undefined };
interface PathPathComponents<T extends Path = Path> extends Array<T> { __pathComponensBrand: undefined }
`

// tsc's getPathComponents (path.ts:493) over pathComponents (path.ts:449): the first overload promises
// PathPathComponents over an implementation returning a fresh string[], so only phantom brands differ,
// on the array and on each element (rulings.md:34 with :32). Mutant: overloadPhantomResult answering no,
// and the declaration is refused.
func TestOverloadPhantomResultServesGetPathComponents(t *testing.T) {
	t.Parallel()
	source := overloadPhantomPrelude + `function pathComponents(path: string, rootLength: number): string[] {
    const root = path.substring(0, rootLength);
    const rest = path.substring(rootLength).split("/");
    if (rest.length > 0 && rest[rest.length - 1] === "") rest.pop();
    return [root, ...rest];
}
function getPathComponents(path: Path): PathPathComponents;
function getPathComponents(path: string, currentDirectory?: string): string[];
function getPathComponents(path: string, currentDirectory = ""): string[] {
    path = currentDirectory + path;
    return pathComponents(path, path.startsWith("/") ? 1 : 0);
}
const parts = getPathComponents("/a/b/" as Path);
console.log(String(parts.length));
console.log(parts.join("|"));
console.log(getPathComponents("c/d", "/x/").join("|"));
`
	program := lowersAndAgreesWithNodeAs(t, source, ".ts")
	if code, stdout, stderr := brandedNative(t, program); code != 0 || stdout != "3\n/|a|b\n/|x|c|d\n" {
		t.Fatalf("native: exit %d, stdout %q, stderr %q; want 3, /|a|b and /|x|c|d, as Node prints", code, stdout, stderr)
	}
}

// Each of these differs from its implementation by more than phantom brands, or by a brand on an array
// someone else may hold, and stays refused at the declaration.
func TestOverloadPhantomResultKeepsRefusing(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, what string }{
		// The intersection spelling isn't an extension, so :32's presence refusals don't see its brand
		// (Language's must). Mutant: phantomStripped taking an intersection's array part.
		{"an intersection-spelled array brand", `type Path = string & { __pathBrand: undefined };
type Components = Path[] & { __components: undefined };
function comps(path: Path): Components;
function comps(path: string): string[];
function comps(path: string): string[] { return [path]; }
console.log(String(comps("a" as Path).length));
`, "overload 1 of comps result"},
		// An extension of an object, not an array: :32's presence refusals don't see its brand, so `in` on
		// the result would lower (Language's must on efa2009c). Mutant: phantomStripped stripping any
		// extension, and this is accepted.
		{"a phantom extension of an object", `interface Node { readonly kind: number }
interface TaggedNode extends Node { __tagged: undefined }
function tag(node: Node, sure: true): TaggedNode;
function tag(node: Node, sure: boolean): Node;
function tag(node: Node, sure: boolean): Node { return node; }
console.log(String(tag({ kind: 1 }, true).kind));
`, "overload 1 of tag result"},
		// Only the array's own brand differs, over an array the body also stores in a number[] field, so an
		// alias's push breaks what the brand claims (Language's Sorted program on efa2009c). Mutant: the
		// array-brand-only branch answering yes without freshness, and this is accepted.
		{"an array-only brand over an array stored elsewhere", `interface Sorted<T> extends Array<T> { __sorted: undefined }
const store: { last: number[] } = { last: [] };
function sorted(xs: number[], sure: true): Sorted<number>;
function sorted(xs: number[], sure: boolean): number[];
function sorted(xs: number[], sure: boolean): number[] { const ys = [...xs]; ys.sort((a, b) => a - b); store.last = ys; return ys; }
console.log(String(sorted([2, 1], true).length));
`, "overload 1 of sorted result"},
		// An element that differs by more than a brand. Mutant: the stripped elements' identity test dropped.
		{"an element narrowed by a literal", overloadPhantomPrelude + `function comps(path: Path): ("a" | "b")[];
function comps(path: string): string[];
function comps(path: string): string[] { return [path]; }
console.log(String(comps("a" as Path).length));
`, "overload 1 of comps result"},
		// A readonly array promised as a mutable brand. Mutant: the mutability comparison dropped.
		{"a readonly result promised mutable", overloadPhantomPrelude + `function comps(path: Path): PathPathComponents;
function comps(path: string): readonly string[];
function comps(path: string): readonly string[] { return [path]; }
console.log(String(comps("a" as Path).length));
`, "overload 1 of comps result"},
		// The returned array is also held as string[] elsewhere, so a plain string pushed there reads back
		// as an unminted Path (Compiler's condition). Mutant: overloadReturnsFreshArray answering yes.
		{"an array also stored in a string[] field", overloadPhantomPrelude + `const store: { last: string[] } = { last: [] };
function comps(path: Path): PathPathComponents;
function comps(path: string): string[];
function comps(path: string): string[] { const parts = path.split("/"); store.last = parts; return parts; }
console.log(String(comps("a/b" as Path).length));
`, "overload 1 of comps result"},
		{"a returned parameter", overloadPhantomPrelude + `function comps(path: Path, parts: string[]): PathPathComponents;
function comps(path: string, parts: string[]): string[];
function comps(path: string, parts: string[]): string[] { return parts; }
console.log(String(comps("a" as Path, ["x"]).length));
`, "overload 1 of comps result"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.What, probe.what) {
				t.Fatalf("want %q refused, got %v", probe.what, err)
			}
		})
	}
}
