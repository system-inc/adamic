package lower

import (
	"errors"
	"strings"
	"testing"
)

// #k8ed1f2, Adamic's ruling: a function value written at a view site, an arrow or a function expression with
// no other handle, is fresh for what each call returns when every value it returns is made in its own body
// and stored nowhere, as an object literal's own method is (#wjqtypf). Its results' own slots are then
// covariant, and only what they hold is a view. A native map whose callback is such a function makes each
// element of its copy the same way. Each positive pin below refused on main, at the function or the map,
// and writes through the wide view after lowering, so a wrong lowering disagrees with Node.

// arrowInvariant is the invariant rule's fix, the end of every refusal these pins expect.
const arrowInvariant = "; make the wider type readonly (readonly T[], ReadonlyMap, readonly fields), which can't write; or copy the value ([...items], { ...item }) (adamic/invariant-mutable)"

// arrowPets is a Dog that's an Animal with a field of its own added, never one narrowed, so a Dog seen as
// an Animal is no refusal of its own: Animal's one field is the same string on both. Two lines.
const arrowPets = "class Animal { name: string; constructor(name: string) { this.name = name; } }\n" +
	"class Dog extends Animal { bark: string; constructor(name: string, bark: string) { super(name); this.bark = bark; } }\n"

// arrowShapes is arrowPets as interfaces, for a positive pin whose values are object literals: a class's
// instance is nominal, and a literal seen as one is refused first as no instance (adamic/nominal-class).
const arrowShapes = "interface Animal { name: string }\n" +
	"interface Dog extends Animal { bark: string }\n"

// arrowRefusedAt fails t unless source, lowered as main plus extension, is refused first at site
// (line:column) in the invariant rule's words.
func arrowRefusedAt(t *testing.T, source, extension, site string) {
	t.Helper()
	_, err := lowerSourceAs(t, source, extension)
	where := "main" + extension + ":" + site + ": Adamic 0.1 refuses a value of type "
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), where) || !strings.HasSuffix(err.Error(), arrowInvariant) {
		t.Fatalf("got %v, want a refusal at %q ending %q", err, where, arrowInvariant)
	}
}

// P1, symbolWalker.ts's shape with an expression body: a literal's walk, () => ({ tag: "a" }), seen as
// () => { tag: string | undefined }, makes a new object at every call, so the write of undefined through the
// wide view reaches nothing else. On main the arrow's inferred () => { tag: string } was walked as a view
// and refused. Mutant M4: freshOrWidened's arrow arm disabled, and this refuses again.
func TestFreshArrowExpressionBodyLowers(t *testing.T) {
	t.Parallel()
	source := "interface Walker {\n" +
		"\twalk: () => { tag: string | undefined };\n" +
		"}\n" +
		"const walker: Walker = { walk: () => ({ tag: \"a\" }) };\n" +
		"const step = walker.walk();\n" +
		"step.tag = undefined;\n" +
		"console.log(String(step.tag));\n" +
		"console.log(String(walker.walk().tag));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, extension)
		})
	}
}

// P2, symbolWalker.ts's own shape: a block body returns a literal inside try, and the finally mutates a
// captured record. The finally can't name the value the return built, so it's still new at every call.
// Refused on main as P1 was. Mutant M4 refuses it again.
func TestFreshArrowTryFinallyLowers(t *testing.T) {
	t.Parallel()
	source := "interface Walker {\n" +
		"\twalk: () => { tag: string | undefined };\n" +
		"}\n" +
		"const record = { calls: 0 };\n" +
		"const walker: Walker = {\n" +
		"\twalk: () => {\n" +
		"\t\ttry {\n" +
		"\t\t\treturn { tag: \"a\" };\n" +
		"\t\t} finally {\n" +
		"\t\t\trecord.calls = record.calls + 1;\n" +
		"\t\t}\n" +
		"\t},\n" +
		"};\n" +
		"const step = walker.walk();\n" +
		"step.tag = undefined;\n" +
		"console.log(String(step.tag));\n" +
		"console.log(String(walker.walk().tag));\n" +
		"console.log(String(record.calls));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, extension)
		})
	}
}

// P3, declarations.ts's shape: an arrow assigned with =, whose fresh result holds a held Dog seen as an
// Animal. The result's own node slot is new, so writing a cat into it reaches no Dog, and the Dog itself is
// walked as a view of Animal, which can't write what a Dog can't hold. Refused on main, since node was a
// writable Dog slot seen as an Animal one. Mutant M4 refuses it again.
func TestFreshArrowAssignedHoldingAHeldValueLowers(t *testing.T) {
	t.Parallel()
	source := arrowShapes +
		"const dog: Dog = { name: \"Rex\", bark: \"woof\" };\n" +
		"const cat: Animal = { name: \"Tom\" };\n" +
		"let get: () => { node: Animal } = () => ({ node: cat });\n" +
		"get = () => ({ node: dog });\n" +
		"get().node = cat;\n" +
		"console.log(dog.bark);\n" +
		"console.log(dog.name);\n" +
		"console.log(get().node.name);\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, extension)
		})
	}
}

// P4, moduleSpecifiers.ts's shape: a native map whose callback makes each element, n => ({ tag: n }), into
// { tag: string | undefined }[]. Each element is a new object, so writing undefined into one reaches
// nothing else. On main the copy's element type, { tag: string }, was walked as a view and refused. Mutant
// M5: freshOrWidened passes no callback to freshContainerWidened, and this refuses again.
func TestFreshMapLiteralElementsLower(t *testing.T) {
	t.Parallel()
	source := "const names: string[] = [\"a\", \"b\"];\n" +
		"const rows: { tag: string | undefined }[] = names.map(n => ({ tag: n }));\n" +
		"const first = rows[0];\n" +
		"if (first !== undefined) first.tag = undefined;\n" +
		"for (const row of rows) console.log(String(row.tag));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, extension)
		})
	}
}

// P5, utilities.ts's shape: names.map(() => []) into string[][]. Each element is a new empty array, so a
// string pushed into one reaches no other. On main the copy's never[] elements were walked as writable
// views of string[] and refused. Mutant M5 refuses it again.
func TestFreshMapEmptyArrayElementsLower(t *testing.T) {
	t.Parallel()
	source := "const names: string[] = [\"a\", \"b\"];\n" +
		"const results: string[][] = names.map(() => []);\n" +
		"const first = results[0];\n" +
		"if (first !== undefined) first.push(\"x\");\n" +
		"for (const result of results) console.log(String(result.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, extension)
		})
	}
}

// R1, condition 1: an arrow returning a held { inner: Dog[] } as { inner: Animal[] } gives every caller the
// same object, so get().inner.push(cat) writes a cat among the dogs. Refused on main and after, at the arrow
// (4:38, the arrow's open parenthesis), since a name isn't made at the site and the arrow takes widened's
// path. Mutant M1 (condition 1 dropped) still refuses it at the same site in the same words, because
// madeAtSiteWidened judges a name by its type (freshOrWidened), so R1 is no witness for M1 (R11 is).
func TestFreshArrowReturningAHeldValueStaysRefused(t *testing.T) {
	t.Parallel()
	source := arrowPets +
		"const held: { inner: Dog[] } = { inner: [new Dog(\"Rex\", \"woof\")] };\n" +
		"let get: () => { inner: Animal[] } = () => held;\n" +
		"console.log(String(get().inner.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "4:38")
		})
	}
}

// R2, Language's attack: captured ?? { inner: [] } gives back the captured { inner: Dog[] } whenever it's
// there, the same object at every call. ?? is no value made at the site, so the arrow takes widened's path,
// refused on main and after at the arrow (5:38, its open parenthesis).
func TestFreshArrowNullishOfACapturedValueStaysRefused(t *testing.T) {
	t.Parallel()
	source := arrowPets +
		"function pick(n: number): { inner: Dog[] } | undefined { return n > 0 ? { inner: [new Dog(\"Rex\", \"woof\")] } : undefined; }\n" +
		"const captured = pick(1);\n" +
		"let get: () => { inner: Animal[] } = () => captured ?? { inner: [] };\n" +
		"console.log(String(get().inner.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "5:38")
		})
	}
}

// R3: x => x gives back its own argument, so f(kept).tag = undefined writes undefined where kept.tag reads
// a string. A parameter isn't made at the site, so it's refused on main and after at the arrow (1:64, the
// parameter x that starts it).
func TestFreshArrowReturningItsParameterStaysRefused(t *testing.T) {
	t.Parallel()
	source := "const f: (x: { tag: string }) => { tag: string | undefined } = x => x;\n" +
		"const kept = { tag: \"a\" };\n" +
		"const seen = f(kept);\n" +
		"seen.tag = undefined;\n" +
		"console.log(String(kept.tag.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "1:64")
		})
	}
}

// R4: the arrow pushes the object it returns into a closed-over { tag: string }[] before returning it, so
// a write of undefined through the wide result reaches stash's string. It returns a name, not a value made
// in the return, so it's refused on main and after at the arrow (2:49, its open parenthesis).
func TestFreshArrowStashingItsResultStaysRefused(t *testing.T) {
	t.Parallel()
	source := "const stash: { tag: string }[] = [];\n" +
		"const make: () => { tag: string | undefined } = () => {\n" +
		"\tconst o = { tag: \"a\" };\n" +
		"\tstash.push(o);\n" +
		"\treturn o;\n" +
		"};\n" +
		"const seen = make();\n" +
		"seen.tag = undefined;\n" +
		"console.log(String(stash.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "2:49")
		})
	}
}

// R5: a function expression recursing through its own name returns its own call's result, a call no
// condition reads across. It fails condition 1 (a call) and condition 3 (its own name) both, so it's
// refused on main and after at the function expression (1:59, its function keyword), and neither M1 nor M6
// alone lets it through.
func TestFreshFunctionRecursingThroughItsNameStaysRefused(t *testing.T) {
	t.Parallel()
	source := "const count: (n: number) => { tag: string | undefined } = function self(n: number): { tag: string } {\n" +
		"\treturn n > 0 ? self(n - 1) : { tag: \"a\" };\n" +
		"};\n" +
		"const seen = count(1);\n" +
		"seen.tag = undefined;\n" +
		"console.log(String(seen.tag));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "1:59")
		})
	}
}

// R6, condition 2: the returned literal's method is handed the literal as this, and its this parameter
// stores it as a { tag: string }, so seen.keep() then seen.tag = undefined leaves stash holding undefined
// where it reads a string. The method takes an explicit this because without one, strict typing gives this
// the literal's contextual type, the wide view, and stash.push(this) into { tag: string }[] doesn't check.
// Refused on main and after at the arrow (2:63, its open parenthesis), since a literal with a method lends
// itself. Mutant M2 (lendsItself dropped): the arm judges the literal fresh, and this lowers.
func TestFreshArrowLiteralLendingItselfStaysRefused(t *testing.T) {
	t.Parallel()
	source := "const stash: { tag: string }[] = [];\n" +
		"const make: () => { tag: string | undefined; keep(): void } = () => ({ tag: \"a\", keep(this: { tag: string }) { stash.push(this); } });\n" +
		"const seen = make();\n" +
		"seen.keep();\n" +
		"seen.tag = undefined;\n" +
		"console.log(String(stash.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "2:63")
		})
	}
}

// R7, the hole map's inference opens: in holders.map(h => ({ inner: h.dogs })), tsc takes map's element
// type from the callback's return, { inner: Dog[] }, so the literal's own sites judge h.dogs against Dog[]
// and pass, while the copy is written into { inner: Animal[] }[], whose inner can push a cat among a
// holder's dogs. madeAtSiteWidened judges the literal against the slot's element instead, and h.dogs as a
// view of Animal[] is refused. On main the copy's element type was walked as a view and refused. Both
// refuse at the map call (4:37, holders). Mutant M3 (the literal case judges nothing): the call lowers.
func TestFreshMapLiteralHoldingAHeldArrayStaysRefused(t *testing.T) {
	t.Parallel()
	source := arrowPets +
		"const holders: { dogs: Dog[] }[] = [{ dogs: [new Dog(\"Rex\", \"woof\")] }];\n" +
		"const tags: { inner: Animal[] }[] = holders.map(h => ({ inner: h.dogs }));\n" +
		"console.log(String(tags.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "4:37")
		})
	}
}

// R8: kennels.map(k => k) holds the kennels' own objects, so out's undefined lands where a kennel reads a
// string. k isn't made at the site, so the copy is walked as a view of its elements, refused on main and
// after at the map call (2:44, kennels).
func TestFreshMapOfHeldElementsStaysRefused(t *testing.T) {
	t.Parallel()
	source := "const kennels: { tag: string }[] = [{ tag: \"a\" }];\n" +
		"const out: { tag: string | undefined }[] = kennels.map(k => k);\n" +
		"const first = out[0];\n" +
		"if (first !== undefined) first.tag = undefined;\n" +
		"console.log(String(kennels.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "2:44")
		})
	}
}

// R9: a fresh result can still hold a held part: get().inner is dogs itself, seen as a writable Animal[].
// Refused on main and after at the arrow (4:38, its open parenthesis): madeAtSiteWidened walks dogs as a
// view of Animal[]. Mutant M3: the arrow passes, and the refusal moves in to the literal's own site for
// dogs (4:54), which isn't the arrow's.
func TestFreshArrowLiteralHoldingAHeldArrayStaysRefused(t *testing.T) {
	t.Parallel()
	source := arrowPets +
		"const dogs: Dog[] = [new Dog(\"Rex\", \"woof\")];\n" +
		"let get: () => { inner: Animal[] } = () => ({ inner: dogs });\n" +
		"console.log(String(get().inner.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "4:38")
		})
	}
}

// R10, condition 3 alone: a function expression that stores itself by its own name, whose every return is
// a literal with no member of its own. Only its name keeps it from being fresh, so it's refused on main and
// after at the function expression (2:49, its function keyword), by widened's path. Mutant M6 (namesItself
// dropped): the arm judges it fresh, and this lowers. R5 can't witness M6, since its call fails condition 1.
func TestFreshFunctionStoringItselfStaysRefused(t *testing.T) {
	t.Parallel()
	source := "const stash: (() => { tag: string })[] = [];\n" +
		"const make: () => { tag: string | undefined } = function self(): { tag: string } {\n" +
		"\tstash.push(self);\n" +
		"\treturn { tag: \"a\" };\n" +
		"};\n" +
		"console.log(String(make().tag));\n" +
		"console.log(String(stash.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "2:49")
		})
	}
}

// R11, condition 1 where it carries weight: h.spare ?? { inner: h.dogs } is the literal whenever spare is
// undefined, but ?? is no value made at the site. freshOrWidened lets a literal operand of ?? through to its
// own sites, which inside a map callback see map's inferred { inner: Dog[] } (R7's hole), so only condition
// 1, sending this copy down widened's path, refuses it. Refused on main and after at the map call (4:37,
// holders). Mutant M1 (condition 1 dropped): the literal operand is judged nowhere against the slot, and
// this lowers.
func TestFreshMapNullishOfALiteralStaysRefused(t *testing.T) {
	t.Parallel()
	source := arrowPets +
		"const holders: { dogs: Dog[]; spare: undefined }[] = [{ dogs: [new Dog(\"Rex\", \"woof\")], spare: undefined }];\n" +
		"const tags: { inner: Animal[] }[] = holders.map(h => h.spare ?? { inner: h.dogs });\n" +
		"console.log(String(tags.length));\n"
	for _, extension := range []string{".a", ".ts"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			arrowRefusedAt(t, source, extension, "4:37")
		})
	}
}
