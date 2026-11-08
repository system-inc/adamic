package lower

import (
	"errors"
	"strings"
	"testing"
)

// A view the program declares that names a lib-declared member the runtime shape lacks is NotYet at
// the relation, at any depth (#sn788gx, Adamic's ruling at 23:3xZ Nov 7). lib declares stack and cause
// on Error, and the runtime's Error shape holds name and message alone, so on main a real Error seen as
// { readonly stack?: string } read stack as absent natively, printing none where Node printed the stack,
// in .ts and .a both. A view naming the members the shape holds, and the Error type itself, still lower.
// Mutant: libraryUnstoredMember answering "" for every member, and every row lowers again, the silent none among them.
func TestLibraryUnstoredMemberViewIsNotYet(t *testing.T) {
	t.Parallel()
	const made = "const made = new Error(\"m\");\n"
	for _, probe := range []struct{ name, source, line, member string }{
		{"stack through a declared view", "interface Traced { readonly stack?: string }\n" + made + "const seen: Traced = made;\nconst text = seen.stack ?? \"none\";\nconsole.log(text.startsWith(\"Error: m\") ? \"stack\" : text);\n", ":3:", "stack"},
		{"stack one field deep", "interface Traced { readonly stack?: string }\ninterface Holder { readonly inner: Traced }\n" + made + "const holder: Holder = { inner: made };\nconsole.log(holder.inner.stack ?? \"none\");\n", ":4:", "stack"},
		{"stack through a parameter", "interface Traced { readonly stack?: string }\nfunction show(seen: Traced): string { return seen.stack ?? \"none\"; }\n" + made + "console.log(show(made));\n", ":4:", "stack"},
		{"cause through a declared view", "interface Caused { readonly cause?: unknown }\n" + made + "const seen: Caused = made;\nconsole.log(seen.cause === undefined ? \"none\" : \"cause\");\n", ":3:", "cause"},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := "stage 0 can't lower a view naming " + probe.member + ", which lib declares on Error but the runtime's Error doesn't hold (leave " + probe.member + " out of the view's type) yet"
				var pending *NotYet
				if !errors.As(err, &pending) || !strings.HasSuffix(err.Error(), want) || !strings.Contains(err.Error(), "main"+extension+probe.line) {
					t.Fatalf("want NotYet %q at %s, got %v", want, probe.line, err)
				}
			})
		}
	}
	// The members Error's shape holds, and Error seen as itself, still lower and agree with Node.
	lowersAndAgreesWithNodeAs(t, "interface Said { readonly name: string; readonly message: string }\n"+made+"const seen: Said = made;\nconsole.log(seen.name + \": \" + seen.message);\nconst same: Error = made;\nconsole.log(same.message);\n", ".a")
}

// A method a library interface declares lives on the prototype, never in a runtime shape, so a view the
// program declares naming it is NotYet at the relation too (#sn788gx's census, Cloud job
// 01M6X4EQS643Y5NGD8S9PB7272 on main 4dd0aaad, .ts native against Node): a Date seen as
// { readonly getTime?: () => number } printed absent where Node prints present, and getTime called
// through { getTime(): number }, or next through a view of a matchAll iterator, exited 70 with "compiler
// bug: a method the checker proved is there is missing" where Node prints 0 and value.
// Mutant: libraryUnstoredMember's method arm (`if declaration.Kind == ast.KindMethodSignature {`) becomes
// `if false {`, and every row lowers again, the silent absent among them.
func TestLibraryMethodViewIsNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, line, member, owner string }{
		{"a method read as an optional field", "const made = new Date(0);\nconst holder: { readonly getTime?: () => number } = made;\nconsole.log(holder.getTime === undefined ? \"absent\" : \"present\");\n", ":2:", "getTime", "Date"},
		{"a method called through a view", "const made = new Date(0);\nconst clock: { getTime(): number } = made;\nconsole.log(`${clock.getTime()}`);\n", ":2:", "getTime", "Date"},
		{"a method one field deep", "interface Clock { getTime(): number }\ninterface Holder { readonly clock: Clock }\nconst holder: Holder = { clock: new Date(0) };\nconsole.log(`${holder.clock.getTime()}`);\n", ":3:", "getTime", "Date"},
		{"next through a view of a matchAll iterator", "const it: { next(): IteratorResult<RegExpExecArray> } = \"aa\".matchAll(/a/g);\nconst step = it.next();\nconsole.log(step.done ? \"done\" : \"value\");\n", ":1:", "next", ""},
	} {
		for _, extension := range []string{".a", ".ts"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				want := "stage 0 can't lower a view naming " + probe.member + ", which lib declares on " + probe.owner
				var pending *NotYet
				if !errors.As(err, &pending) || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "main"+extension+probe.line) {
					t.Fatalf("want NotYet %q at %s, got %v", want, probe.line, err)
				}
			})
		}
	}
	// A view naming only the Date's value and no method lowers: a Date seen as itself.
	lowersAndAgreesWithNodeAs(t, "const made = new Date(0);\nconst same: Date = made;\nconsole.log(`${same.getTime()}`);\n", ".a")
}

// The rule reads only a view the program declares (Compiler's must on #sn788gx). A lib type seen as
// another lib type is lowering's own library dispatch: a RegExp handed to String.prototype.replace is
// related to lib's { [Symbol.replace](...) } parameter, and a Map seen as ReadonlyMap is still a Map, so
// both lower and agree with Node. A program view naming a well-known symbol method is named as the source
// spells it, never by the checker's internal key.
// Mutant: the method arm's `(!method || programDeclared(property))` becomes `true`, and the replace row
// stops with "a view naming [Symbol.replace]".
func TestLibraryOwnViewsLower(t *testing.T) {
	t.Parallel()
	lowersAndAgreesWithNodeAs(t, "console.log(\"a-b\".replace(/-/g, \"_\"));\nconsole.log(\"a\\nb\".split(/\\n/).join(\",\"));\n", ".a")
	lowersAndAgreesWithNodeAs(t, "const seen: ReadonlyMap<string, number> = new Map<string, number>([[\"a\", 1]]);\nseen.forEach((value: number, key: string): void => { console.log(`${key}=${value}`); });\nconsole.log(seen.has(\"a\") ? \"has\" : \"none\");\n", ".a")
	for _, extension := range []string{".a", ".ts"} {
		t.Run("a symbol-named method through a program view"+extension, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, "const made = new Set<number>([1]);\nconst seen: { [Symbol.iterator](): Iterator<number> } = made;\nconsole.log(seen[Symbol.iterator]().next().done ? \"done\" : \"value\");\n", extension)
			const want = "stage 0 can't lower a view naming [Symbol.iterator], which lib declares on Set"
			var pending *NotYet
			if !errors.As(err, &pending) || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "main"+extension+":2:") {
				t.Fatalf("want NotYet %q at :2:, got %v", want, err)
			}
		})
	}
}
