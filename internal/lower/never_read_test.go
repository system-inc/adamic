package lower

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Language's probes from #p6vawef (language/probe-never-between, probe/never-between) open with the
// same kinds: N exhausted by a check of each kind, with the read past both typed never.
const neverReadKinds = `enum K { A = 1, B = 2 }
interface NodeA { readonly kind: K.A; readonly a: number }
interface NodeB { readonly kind: K.B; readonly b: number }
type N = NodeA | NodeB;
function stray(value: unknown): string { return "stray " + typeof value; }
`

// neverReadHolder is the path road: holder.node, which flip writes through another name.
const neverReadHolder = neverReadKinds + `interface Holder { node: N }
function flip(holder: Holder): void { holder.node = { kind: K.A, a: 1 }; }
function describe(holder: Holder): string {
	if (holder.node.kind === K.A) return 'a';
`

// neverReadLet is the let road: node, which a closure the function calls between the checks writes.
const neverReadLet = neverReadKinds + `function describe(start: N): string {
	let node: N = start;
	const flip = (): void => { node = { kind: K.A, a: 1 }; };
	if (node.kind === K.A) return 'a';
	flip();
	if (node.kind === K.B) return 'b';
`

// neverReadWhere is where the last occurrence of read starts in source, as main.ts:line:column.
func neverReadWhere(t *testing.T, source, read string) string {
	t.Helper()
	index := strings.LastIndex(source, read)
	if index < 0 {
		t.Fatalf("%q is not in the source", read)
	}
	return "main.ts:" + strconv.Itoa(strings.Count(source[:index], "\n")+1) + ":" + strconv.Itoa(index-strings.LastIndex(source[:index], "\n"))
}

// neverReadTraps is whether a lowered program holds a never read's trap: a never_read function that panics.
func neverReadTraps(program *ir.Program) bool {
	for _, function := range program.Functions {
		if function.Name != "never_read" {
			continue
		}
		for _, statement := range function.Body {
			if _, panics := statement.(ir.Panic); panics {
				return true
			}
		}
	}
	return false
}

// A read the checker types never, with something that can write its place between the earliest
// narrowing the never is made of and the read, is rulings.md:151's case and stops as not yet at the read
// (never_read.go). nv1 and nv5 hold the path road's call between the two checks, nv3 its call after the
// last check, which only the earliest narrowing sees; nv2 and nv6 to nv9 the let road, which
// narrowedNeverReadGap also stops. Each pins where it stops and that it's a not yet, never the words, so
// the earliest-narrowing rule alone holds every row: never_read.mutants.md's third mutant unregisters
// narrowedNeverReadGap and runs this test.
func TestNeverReadAcrossAWriteIsNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, read string }{
		{"nv1", neverReadHolder + "\tflip(holder);\n\tif (holder.node.kind === K.B) return 'b';\n\t{ const seen: unknown = holder.node; return stray(seen); }\n}\nconsole.log(describe({ node: { kind: K.B, b: 2 } }));\n", "holder.node; return"},
		{"nv3", neverReadHolder + "\tif (holder.node.kind === K.B) {\n\t\tflip(holder);\n\t\tif (holder.node.kind === K.B) return 'still b';\n\t\t{ const seen: unknown = holder.node; return stray(seen); }\n\t}\n\treturn 'none';\n}\nconsole.log(describe({ node: { kind: K.B, b: 2 } }));\n", "holder.node; return"},
		{"nv5", neverReadHolder + "\tflip(holder);\n\tif (holder.node.kind === K.B) return 'b';\n\t{ const count: number = holder.node; return \"count \" + String(count + 1); }\n}\nconsole.log(describe({ node: { kind: K.B, b: 2 } }));\n", "holder.node; return"},
		{"nv2", neverReadLet + "\t{ const seen: unknown = node; return stray(seen); }\n}\nconsole.log(describe({ kind: K.B, b: 2 }));\n", "node; return"},
		{"nv6", neverReadLet + "\t{ const count: number = node; return \"count \" + String(count + 1); }\n}\nconsole.log(describe({ kind: K.B, b: 2 }));\n", "node; return"},
		{"nv7", neverReadLet + "\t{ const other: { readonly label: string } = node; return other.label.toUpperCase(); }\n}\nconsole.log(describe({ kind: K.B, b: 2 }));\n", "node; return"},
		{"nv8", neverReadLet + "\t{ const other: { readonly label: string } = node; return \"label \" + String(other.label === undefined); }\n}\nconsole.log(describe({ kind: K.B, b: 2 }));\n", "node; return"},
		{"nv9", neverReadLet + "\t{ const other: { readonly kind: string } = node; return other.kind.toUpperCase(); }\n}\nconsole.log(describe({ kind: K.B, b: 2 }));\n", "node; return"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".ts")
			var notYet *NotYet
			if want := neverReadWhere(t, probe.source, probe.read); !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, want) {
				t.Fatalf("got %v, want a not yet at %s", err, want)
			}
		})
	}
}

// A never read with nothing that can write its place between the earliest narrowing and the read is the
// checker's proof, so it lowers, and the read is a loud trap (rulings.md:74): nv4, Language's exhaustive
// control, a parameter two checks exhaust with nothing between; an exhaustive switch's never handed to a
// never-returning call, the domain promise tsc's Debug.assertNever keeps (class (a) in Lower's scout); a
// const an aliased condition narrows, which the walk can't place but nothing but its declaration writes;
// and the switch inside a for-of loop, whose walk stops at the loop's own assignment, before its label.
// never_read.mutants.md's first mutant, the between started at the last narrowing, lowers nv1 this way.
func TestNeverReadWithNothingBetweenTraps(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"nv4", neverReadKinds + "function describe(node: N): string {\n\tif (node.kind === K.A) return 'a';\n\tif (node.kind === K.B) return 'b';\n\t{ const seen: unknown = node; return stray(seen); }\n}\nconsole.log(describe({ kind: K.B, b: 2 }));\nconsole.log(describe({ kind: K.A, a: 1 }));\n"},
		{"switch", neverReadShapes + "console.log(String(corners(\"circle\")) + \" \" + String(corners(\"square\")));\n"},
		{"alias", "function show(value: unknown): string { return typeof value; }\nconst big = 'x';\nconst flag = big === 'x';\nconsole.log(flag ? 'yes' : show(big));\n"},
		{"for_of", "type Shape = \"circle\" | \"square\";\nfunction fail(shape: unknown): never { throw new Error(\"unexpected shape \" + String(shape)); }\nfunction corners(shapes: readonly Shape[]): number {\n\tlet total = 0;\n\tfor (const shape of shapes) {\n\t\tswitch (shape) {\n\t\t\tcase \"circle\": break;\n\t\t\tcase \"square\": total += 4; break;\n\t\t\tdefault: fail(shape);\n\t\t}\n\t}\n\treturn total;\n}\nconsole.log(String(corners([\"circle\", \"square\"])));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, probe.source, ".ts")
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			if !neverReadTraps(program) {
				t.Fatal("want the never read trapped, and no never_read function panics")
			}
		})
	}
}

// neverReadShapes is class (a): an exhaustive switch's never handed to a never-returning call, with
// nothing between the switch and the read. fail takes unknown, since a parameter typed never is not yet on
// its own (a value of type never, at the parameter), before any read reaches it.
const neverReadShapes = `type Shape = "circle" | "square";
function fail(shape: unknown): never { throw new Error("unexpected shape " + String(shape)); }
function corners(shape: Shape): number {
	switch (shape) {
		case "circle": return 0;
		case "square": return 4;
	}
	return fail(shape);
}
`

// tsc's two never reads Lower's scout found reached on valid input (#p6vawef), in their shapes: n1 is
// checker.ts's lastAssignmentPos, a property a call writes after an early !== undefined return, and n2
// esnextAnd2015.ts's importRequireStatements, a captured let a visitor assigns. Each stops before its
// never read, at the condition's read narrowed to undefined (narrowedNullishReadGap), in .ts and .a, so
// each pins where it stops and that it's a not yet. When #54vt253 reads them through the declared type,
// these lower and agree with Node (5 and 5).
func TestNeverReadTscShapesStopAtTheCondition(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, read string }{
		{"n1", "interface Sym { lastAssignmentPos: number | undefined }\nfunction markNodeAssignments(symbol: Sym): void { symbol.lastAssignmentPos = 5; }\nfunction isSymbolAssignedDefinitely(symbol: Sym): number {\n\tif (symbol.lastAssignmentPos !== undefined) return -1;\n\tmarkNodeAssignments(symbol);\n\treturn symbol.lastAssignmentPos !== undefined ? symbol.lastAssignmentPos : 0;\n}\nconsole.log(String(isSymbolAssignedDefinitely({ lastAssignmentPos: undefined })));\n", "symbol.lastAssignmentPos !== undefined ?"},
		{"n2", "function createRequireCall(): number { return 1; }\nfunction use(statements: readonly [number, number]): number { return statements[0] + statements[1]; }\nfunction transformSourceFile(count: number): number {\n\tlet importRequireStatements: [number, number] | undefined = undefined;\n\tconst visit = (): void => { importRequireStatements = [createRequireCall(), count]; };\n\tvisit();\n\tif (importRequireStatements) return use(importRequireStatements);\n\treturn 0;\n}\nconsole.log(String(transformSourceFile(4)));\n", "importRequireStatements) return"},
	} {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var notYet *NotYet
				if want := strings.Replace(neverReadWhere(t, probe.source, probe.read), "main.ts", "main"+extension, 1); !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, want) {
					t.Fatalf("got %v, want a not yet at %s", err, want)
				}
			})
		}
	}
}

// neverReadOwnShape is a local its own function narrows to never by two checks: writer is the statement
// after its declaration, and between the one between the two checks.
func neverReadOwnShape(writer, between string) string {
	return "type Shape = \"circle\" | \"square\";\nfunction stray(value: unknown): string { return \"stray \" + String(value); }\nconst source: { v: Shape; shape: Shape } = { v: \"circle\", shape: \"circle\" };\nconst pair: [Shape] = [\"circle\"];\nfunction describe(start: Shape): string {\n\tlet shape: Shape = start;\n\t" + writer + "\n\tif (shape === \"circle\") return \"circle\";\n\t" + between + "\n\tif (shape === \"square\") return \"square\";\n\t{ const seen: unknown = shape; return stray(seen); }\n}\nconsole.log(describe(\"square\"));\n"
}

// neverReadOwnCount is neverReadOwnShape for ++ and --: a number or string local typeof checks exhaust.
func neverReadOwnCount(update string) string {
	return "function stray(value: unknown): string { return \"stray \" + String(value); }\nfunction describe(start: number | string): string {\n\tlet count: number | string = start;\n\tconst bump = (): void => { if (typeof count === \"number\") " + update + "; };\n\tif (typeof count === \"string\") return \"string\";\n\tbump();\n\tif (typeof count === \"number\") return \"number\";\n\t{ const seen: unknown = count; return stray(seen); }\n}\nconsole.log(describe(1));\n"
}

// The edges of "only its own function writes" (Adamic on #p6vawef, counted by references with
// IsAssignmentTarget as rulings.md:338 does, writtenAcrossFunctions): a local its own function narrows
// to never is trusted only when no other code writes it. An arrow declared in that function that writes
// it and runs between the checks, a destructuring write in one (an object pattern, a shorthand and an
// array), and ++ or -- in one are each another function's write, so each read stops as not yet at the
// read, in .ts and .a, never a trap on a valid run (Node prints stray circle for the first five). The
// control writes the same local by a destructuring in its own function, which the checker's flow sees:
// between the checks the never vanishes and the read lowers as a read, and before them the never holds
// and lowers to the trap.
func TestNeverReadOwnFunctionEdges(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, read string }{
		{"arrow", neverReadOwnShape("const flip = (): void => { shape = \"circle\"; };", "flip();"), "shape; return"},
		{"destructured_object", neverReadOwnShape("const flip = (): void => { ({ v: shape } = source); };", "flip();"), "shape; return"},
		{"destructured_shorthand", neverReadOwnShape("const flip = (): void => { ({ shape } = source); };", "flip();"), "shape; return"},
		{"destructured_array", neverReadOwnShape("const flip = (): void => { [shape] = pair; };", "flip();"), "shape; return"},
		{"increment", neverReadOwnCount("count++"), "count; return"},
		{"decrement", neverReadOwnCount("count--"), "count; return"},
	} {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(probe.name+extension, func(t *testing.T) {
				t.Parallel()
				_, err := lowerSourceAs(t, probe.source, extension)
				var notYet *NotYet
				if want := strings.Replace(neverReadWhere(t, probe.source, probe.read), "main.ts", "main"+extension, 1); !errors.As(err, &notYet) || !strings.HasSuffix(notYet.Where, want) {
					t.Fatalf("got %v, want a not yet at %s", err, want)
				}
			})
		}
	}
	for _, control := range []struct {
		name, source string
		traps        bool
	}{
		{"control_between", neverReadOwnShape("", "({ v: shape } = source);"), false},
		{"control_before", neverReadOwnShape("({ v: shape } = source);", ""), true},
	} {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(control.name+extension, func(t *testing.T) {
				t.Parallel()
				program, err := lowerSourceAs(t, control.source, extension)
				if err != nil {
					t.Fatalf("want it to lower, got %v", err)
				}
				if neverReadTraps(program) != control.traps {
					t.Fatalf("trapped %v, want %v", neverReadTraps(program), control.traps)
				}
			})
		}
	}
}
