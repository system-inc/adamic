package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// falsyArms keeps every arm of an && left that a primitive can be a value of and drops the arms that can only
// be objects (#y1mw6wj): a type with a required member no primitive has (Builder, an array, through pop), a
// function, and a type parameter whose constraint is one of those. `{}`, `unknown`, a structural type whose
// required members a primitive has (`{ length: number }`), a weak type and a type of index signatures, Language's
// two-hop Weakish and its intersection with Sized, a type parameter bound by Weakish, an unconstrained or
// `{}`-bound type parameter and a brand stay judged. Mutants: leave string out of primitiveHas, and the
// structural and sizedFoo rows are dropped; count an optional property as required, and the weak and weakish
// rows are dropped.
func TestAndLeftFalsyArms(t *testing.T) {
	t.Parallel()
	const source = `interface Program { name: string }
interface Builder { readonly name: string; readonly program: Program | undefined }
interface Sized { length: number }
interface Weakish { foo?: number }
function probe<B extends Builder, S extends { readonly length: number }, U, E extends {}, W extends Weakish, V extends W>(
	builder: Builder | undefined, generic: B | undefined, list: string[] | undefined, call: (() => void) | undefined,
	text: string | Builder, zero: 0 | Builder, flag: boolean | Builder, big: bigint | Builder, empty: {} | Builder,
	opaque: unknown, sized: S | Builder, bare: U | Builder, loose: E | Builder, structural: Sized | Builder,
	brand: (string & { readonly brand: true }) | Builder, whole: Builder, weak: { a?: number } | Builder,
	indexed: { [key: string]: number } | Builder, weakish: Weakish | Builder, sizedFoo: { length: number; foo?: number } | Builder,
	both: (Weakish & { length: number }) | Builder, boundWeak: W | Builder, chained: V | Builder): void {}
console.log("probe");
`
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	t.Cleanup(release)
	l := &lowering{program: program, checker: checked, result: &ir.Program{}, this: -1, functionIndex: -1}
	got := map[string]string{}
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node.Kind == ast.KindParameter && ast.IsIdentifier(node.Name()) {
			kept := l.falsyArms(checked.GetTypeAtLocation(node.Name()))
			got[node.Name().Text()] = "nothing"
			if kept != nil {
				got[node.Name().Text()] = checked.TypeToString(kept)
			}
		}
		return node.ForEachChild(walk)
	}
	file.AsNode().ForEachChild(walk)
	for name, want := range map[string]string{
		"builder": "undefined", "generic": "undefined", "list": "undefined", "call": "undefined", "whole": "nothing",
		"text": "string", "zero": "0", "flag": "boolean", "big": "bigint", "empty": "{}", "opaque": "unknown",
		"sized": "S", "bare": "U", "loose": "E", "structural": "Sized",
		"brand": "string & { readonly brand: true; }", "weak": "{ a?: number; }", "indexed": "{ [key: string]: number; }",
		"weakish": "Weakish", "sizedFoo": "{ length: number; foo?: number; }", "both": "Weakish & { length: number; }",
		"boundWeak": "W", "chained": "V",
	} {
		if got[name] != want {
			t.Errorf("%s: falsy arms %q, want %q", name, got[name], want)
		}
	}
}

// tsc's two sites, as minimal programs: builder.ts:1601's object literal property and watchPublic.ts:610's
// generic return, each `x && x.program` where x is an object type or undefined. The left reaches the result only
// as undefined, so the Builder whose readonly name a Program could write is no view of the result, and each lowers
// and agrees with Node, the program read back through the result written. Mutant: drop the andLeftOperand check
// in refuseWidening, and each is refused again, "a value of type Builder | undefined seen as Program | undefined,
// whose readonly field name becomes writable".
func TestAndLeftObjectArmsLower(t *testing.T) {
	t.Parallel()
	const shapes = "interface Program { name: string }\ninterface Builder { readonly name: string; readonly program: Program | undefined }\nconst program: Program = { name: \"p\" };\nconst builder: Builder = { name: \"b\", program };\n"
	for name, source := range map[string]string{
		"builder.ts property":    shapes + "function options(old: Builder | undefined): { oldProgram: Program | undefined } {\n\treturn { oldProgram: old && old.program };\n}\nconst none = options(undefined).oldProgram;\nconsole.log(none === undefined ? \"none\" : none.name);\nconst made = options(builder).oldProgram;\nif (made !== undefined) made.name = \"q\";\nconsole.log(program.name);\n",
		"watchPublic.ts generic": shapes + "function current<T extends Builder>(watched: T | undefined): Program | undefined {\n\treturn watched && watched.program;\n}\nconst none = current<Builder>(undefined);\nconsole.log(none === undefined ? \"none\" : none.name);\nconst got = current(builder);\nif (got !== undefined) got.name = \"r\";\nconsole.log(program.name);\n",
		"parenthesized":          shapes + "function read(old: Builder | undefined): Program | undefined {\n\treturn ((old)) && old.program;\n}\nconst none = read(undefined);\nconsole.log(none === undefined ? \"none\" : none.name);\nconsole.log(read(builder)?.name ?? \"none\");\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			lowersAndAgreesWithNodeAs(t, source, ".ts")
		})
	}
}

// An arm a primitive can be stays judged, so a falsy primitive never reaches a slot that reads it as an object:
// `""` bound to `T extends { readonly length: number }` on the left of && would leave "" in a Sized slot, the
// silent miscompile this rule must not open. The row with no primitive bound is refused at the && left in
// invariant-mutable's words, as on main; the row binding "" is refused too. Mutant: leave string out of
// primitiveHas, and the constraint's length proves it an object, so the first row lowers.
func TestAndLeftKeepsPrimitiveArms(t *testing.T) {
	t.Parallel()
	const first = "interface Sized { length: number }\nfunction first<T extends { readonly length: number }>(x: T, y: Sized): Sized {\n\treturn x && y;\n}\n"
	for _, row := range []struct{ name, source, want string }{
		{"object bound", first + "console.log(String(first({ length: 2 }, { length: 3 }).length));\n", "(adamic/invariant-mutable)"},
		{"empty string bound", first + "const got = first(\"\", { length: 3 });\ngot.length = 4;\nconsole.log(String(got.length));\n", ""},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, row.source, ".ts")
			var refused *Refused
			var notYet *NotYet
			if !errors.As(err, &refused) && !errors.As(err, &notYet) {
				t.Fatalf("want a stop, got %v", err)
			}
			if !strings.HasSuffix(err.Error(), row.want) {
				t.Fatalf("want a stop ending %q, got %v", row.want, err)
			}
		})
	}
}
