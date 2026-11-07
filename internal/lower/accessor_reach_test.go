package lower

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

// Accessor names are judged by what can reach what, never program-wide by name (Compiler and Lower,
// Nov 1, #c7w6j0y): a getter refuses a checked field contract only where an allocation owning it
// can reach the view's origin, or the reach is unknown, and a read becomes the accessor dispatcher
// only where its receiver can reach such an allocation, or its reach is unknown.

const accessorReachNodes = `interface Node { readonly kind: number; readonly pos: number; readonly end: number; }
interface Identifier extends Node { readonly kind: 80; readonly text: string; }
function idText(node: Node): string {
    if (node.kind === 80) { const id = node as Identifier; return id.text + "@" + id.pos; }
    return "?";
}
`

// The decoder's shape, a literal with get pos that never reaches a Node, beside a checked Node view
// whose contract has pos, lowers and agrees with Node.
const decoderBesideNodeView = accessorReachNodes + `interface MappingsDecoder { readonly pos: number; next(): boolean; }
function decodeMappings(mappings: string): MappingsDecoder {
    let pos = 0;
    return {
        get pos() { return pos; },
        next: () => { pos++; return pos < mappings.length; },
    };
}
const ident = { kind: 80 as const, pos: 3, end: 7, text: "x" };
console.log(idText(ident));
const decoder = decodeMappings("abc");
let count = 0;
while (decoder.next()) { count++; }
console.log(String(decoder.pos) + " " + String(count));
`

func TestDecoderGetterBesideCheckedNodeViewAgrees(test *testing.T) {
	test.Parallel()
	program := lowersAndAgreesWithNodeAs(test, decoderBesideNodeView, ".ts")
	accessorReachNative(test, program, "x@3\n3 2\n")
	// What behavior can't observe: the Node's pos read keeps its check. Node prints the same
	// whether the read is checked, unchecked or dispatched, since the value is a plain field.
	viewed, dispatched := accessorReachReads(program, "pos")
	if viewed != 1 {
		test.Fatalf("want the Node pos read kept as one checked ir.Property, got %d", viewed)
	}
	if dispatched != 1 {
		test.Fatalf("want only the decoder's pos read dispatched, got %d dispatcher calls", dispatched)
	}
}

// sourcemap.ts:474's own shape, as close to tsc's decodeMappings as a small program allows: getters
// for pos, state and error over the decoder's locals, next as a function field (an accessor literal
// takes no method, so no Symbol.iterator either), its use walking the decoder and reading its error,
// beside a checked Node view.
const sourcemapDecoder = accessorReachNodes + `interface Mapping { generatedLine: number; generatedCharacter: number; }
interface MappingsDecoder {
    readonly pos: number;
    readonly error: string;
    readonly state: Mapping;
    next(): boolean;
}
function decodeMappings(mappings: string): MappingsDecoder {
    let done = false;
    let pos = 0;
    let generatedLine = 0;
    let generatedCharacter = 0;
    let error = "";
    return {
        get pos() { return pos; },
        get state() { return captureMapping(); },
        get error() { return error; },
        next: () => {
            while (!done && pos < mappings.length) {
                const ch = mappings.charCodeAt(pos);
                if (ch === 59) { generatedLine++; generatedCharacter = 0; pos++; continue; }
                if (ch === 44) { pos++; continue; }
                generatedCharacter++;
                pos++;
                return true;
            }
            if (pos > mappings.length) { error = "Invalid mappings"; }
            done = true;
            return false;
        },
    };
    function captureMapping(): Mapping {
        return { generatedLine, generatedCharacter };
    }
}
function isSourceMapping(decoder: MappingsDecoder): string {
    let last = "";
    while (decoder.next()) { last = decoder.state.generatedLine + ":" + decoder.state.generatedCharacter; }
    return decoder.error === "" ? last + " at " + decoder.pos : decoder.error;
}
const ident = { kind: 80 as const, pos: 3, end: 7, text: "x" };
console.log(idText(ident));
console.log(isSourceMapping(decodeMappings("A,B;C")));
`

func TestSourcemapDecoderShapeAgrees(test *testing.T) {
	test.Parallel()
	program := lowersAndAgreesWithNodeAs(test, sourcemapDecoder, ".ts")
	accessorReachNative(test, program, "x@3\n1:1 at 5\n")
	// What behavior can't observe: the Node's pos read keeps its check (as above).
	if viewed, _ := accessorReachReads(program, "pos"); viewed != 1 {
		test.Fatalf("want the Node pos read kept as one checked ir.Property, got %d", viewed)
	}
}

// A read whose receiver is a callback's parameter, which the graph can't follow, may reach the
// getter, so it stays the dispatcher's call and reads the getter's value, not its slot, while a
// read of a literal no getter owns keeps its plain load.
func TestGetterThroughUnknownReceiverAgrees(test *testing.T) {
	test.Parallel()
	program := lowersAndAgreesWithNodeAs(test, `interface Positioned { readonly pos: number; }
let at = 3;
const decoder = { get pos() { return at; } };
const plain = { pos: 5 };
const all: Positioned[] = [decoder, plain];
all.forEach((each) => console.log(String(each.pos)));
at = 4;
console.log(String(decoder.pos) + " " + String(plain.pos));
`, ".ts")
	accessorReachNative(test, program, "3\n5\n4 5\n")
}

// accessorReachNative builds program natively, under the sanitizers as getter_structural_read_test.go
// does, and wants want on its stdout, what Node prints: the JavaScript backend's property read
// runs a getter itself, so it can't tell a slot load from the dispatcher's call, and native can.
func accessorReachNative(test *testing.T, program *ir.Program, want string) {
	test.Helper()
	binary := filepath.Join(test.TempDir(), "native")
	if buildError := native.Build(native.C(program), binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	var out, errs bytes.Buffer
	command.Stdout, command.Stderr = &out, &errs
	if runError := command.Run(); runError != nil {
		test.Fatalf("native: %v, stdout %q stderr %q", runError, out.String(), errs.String())
	}
	if out.String() != want || errs.Len() != 0 {
		test.Fatalf("native: stdout %q stderr %q; want %q, as Node prints", out.String(), errs.String(), want)
	}
}

// accessorReachReads counts the program's checked ir.Property reads of name and its calls to
// name's getter dispatcher.
func accessorReachReads(program *ir.Program, name string) (int, int) {
	viewed, dispatched := 0, 0
	count := func(node any) bool {
		switch expression := node.(type) {
		case ir.Property:
			if expression.Name == name &&
				expression.View != "" &&
				program.CheckedField(expression.ViewReceiverTypeID, name) {
				viewed++
			}
		case ir.Call:
			for _, target := range program.CallTargets(expression) {
				if program.Functions[target].Name == "property_get:"+name {
					dispatched++
				}
			}
		}
		return true
	}
	walk(program.Main, count)
	for _, function := range program.Functions {
		walk(function.Body, count)
	}
	return viewed, dispatched
}

func TestGetterReachingCheckedViewStaysNotYet(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ name, source, what string }{
		// The literal owning get pos is the value the view checks.
		{"reaching", accessorReachNodes + `let at = 3;
const ident = { kind: 80 as const, get pos() { return at; }, end: 7, text: "x" };
console.log(idText(ident));
`, "a getter in a checked field contract"},
		// The view's origin is a callback's parameter, which the graph can't follow, so a get pos
		// anywhere may reach it.
		{"unknown reach", `interface Node { readonly kind: number; readonly pos: number; }
interface Identifier extends Node { readonly kind: 80; readonly text: string; }
const nodes: Node[] = [{ kind: 80, pos: 3, text: "x" } as Node];
nodes.forEach((node) => { if (node.kind === 80) { const id = node as Identifier; console.log(id.text); } });
let at = 3;
const decoder = { get pos() { return at; } };
console.log(String(decoder.pos));
`, "a getter in a checked field contract"},
		// The cast's origin reaches no getter, but a checked Identifier read through a callback's
		// parameter may, so finishAccessors would route it through the dispatcher's unchecked load.
		{"guard", accessorReachNodes + `const ident = { kind: 80 as const, pos: 3, end: 7, text: "x" };
console.log(idText(ident));
let at = 3;
const decoder = { get pos() { return at; } };
console.log(String(decoder.pos));
const ids: Identifier[] = [ident];
ids.forEach((id) => console.log(String(id.pos)));
`, "a checked view read whose receiver may own an accessor of the same name"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerTypeScriptAssertionSourceError := lowerTypeScriptAssertionSource(test, probe.source)
			var notYet *NotYet
			if !errors.As(lowerTypeScriptAssertionSourceError, &notYet) ||
				!strings.Contains(lowerTypeScriptAssertionSourceError.Error(), probe.what) {
				test.Fatalf("want NotYet %q, got %v", probe.what, lowerTypeScriptAssertionSourceError)
			}
		})
	}
}

// A class's static getter lives on the class value, whose object is an allocation with the static
// class's layout and accessors (class_static.go's staticDeclaration), so the class value seen
// through a checked contract naming pos reaches an owner: a cast's view stays NotYet at the getter,
// as the old scan by name had it, and so does a checked predicate's, whose origin is unknown.
func TestStaticGetterInCheckedContractStaysNotYet(test *testing.T) {
	test.Parallel()
	prelude := `interface Positioned { readonly kind: number; readonly pos: number; }
interface Clocked extends Positioned { readonly kind: 1; }
class Clock { static kind = 1; static get pos(): number { return 3; } }
`
	for _, probe := range []struct{ name, source string }{
		{"cast of a parameter", prelude +
			`function show(value: Positioned): string { if (value.kind === 1) { const c = value as Clocked; return String(c.pos); } return "?"; }
console.log(show(Clock));
`},
		{"cast of the class value", prelude + `const c = Clock as Positioned as Clocked;
console.log(String(c.pos));
`},
		{
			"checked predicate",
			prelude + `function isClocked(value: Positioned): value is Clocked { return value.kind === 1; }
function show(value: Positioned): string { if (isClocked(value)) { return String(value.pos); } return "?"; }
console.log(show(Clock));
`,
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerTypeScriptAssertionSourceError := lowerTypeScriptAssertionSource(test, probe.source)
			var notYet *NotYet
			if !errors.As(lowerTypeScriptAssertionSourceError, &notYet) ||
				!strings.Contains(
					lowerTypeScriptAssertionSourceError.Error(),
					"a getter in a checked field contract",
				) ||
				!strings.HasSuffix(notYet.Where, "main.ts:3:32") {
				test.Fatalf(
					"want NotYet at the static getter, main.ts:3:32, got %v",
					lowerTypeScriptAssertionSourceError,
				)
			}
		})
	}
}

// A checked write the accessor rewrite would route through the setter dispatcher is NotYet where it
// is written, at its receiver, as a checked read's is at the read.
func TestCheckedWriteThroughAccessorRewriteIsLocated(test *testing.T) {
	test.Parallel()
	_, lowerTypeScriptAssertionSourceError := lowerTypeScriptAssertionSource(
		test,
		`interface Node { readonly kind: number; pos: number; }
interface Identifier extends Node { readonly kind: 80; readonly text: string; }
function idText(node: Node): string { if (node.kind === 80) { const id = node as Identifier; return id.text; } return "?"; }
const ident = { kind: 80 as const, pos: 3, text: "x" };
console.log(idText(ident));
let at = 3;
const decoder = { get pos() { return at; }, set pos(value: number) { at = value; } };
console.log(String(decoder.pos));
const ids: Identifier[] = [ident];
ids.forEach((id) => { id.pos = 4; });
`,
	)
	var notYet *NotYet
	if !errors.As(lowerTypeScriptAssertionSourceError, &notYet) || !strings.Contains(
		lowerTypeScriptAssertionSourceError.Error(),
		"a checked view write of pos whose receiver may own an accessor of the same name",
	) || !strings.HasSuffix(notYet.Where, "main.ts:10:23") {
		test.Fatalf("want the checked write NotYet at main.ts:10:23, got %v", lowerTypeScriptAssertionSourceError)
	}
}

// The getter's owner is judged per allocation over every module: a literal with get pos in another
// module, reaching the viewed value through an import, stays NotYet.
func TestGetterInAnotherModuleReachingViewStaysNotYet(test *testing.T) {
	test.Parallel()
	directory := test.TempDir()
	for name, source := range map[string]string{
		"nodes.ts": `export interface Node { readonly kind: number; readonly pos: number; }
export interface Identifier extends Node { readonly kind: 80; readonly text: string; }
let at = 3;
export const ident = { kind: 80 as const, get pos() { return at; }, text: "x" };
`,
		"main.ts": `import type { Node, Identifier } from "./nodes";
import { ident } from "./nodes";
function idText(node: Node): string {
    if (node.kind === 80) { const id = node as Identifier; return id.text + "@" + id.pos; }
    return "?";
}
console.log(idText(ident));
`,
	} {
		if writeError := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	program, loadError := load.Load([]string{filepath.Join(directory, "main.ts")})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := Lower(context.Background(), program)
	var notYet *NotYet
	if !errors.As(lowerError, &notYet) ||
		!strings.Contains(lowerError.Error(), "a getter in a checked field contract") ||
		!strings.Contains(notYet.Where, "nodes.ts") {
		test.Fatalf("want NotYet at nodes.ts's getter, got %v", lowerError)
	}
}

// Readiness only removes a read's view, never adds one, so a read finishAccessors keeps without a
// view stays without one: here a record literal's field makes pos checked by name in readiness
// itself, the one place after finishAccessors that widens what is checked.
func TestReadinessNeverAddsAView(test *testing.T) {
	test.Parallel()
	read := func(view string) ir.Statement {
		return ir.Evaluate{Value: ir.Property{
			Object:    ir.Read{Local: 0, Of: ir.Object},
			Name:      "pos",
			Of:        ir.Number,
			View:      view,
			ViewWhere: view,
		}}
	}
	program := &ir.Program{
		Locals: []ir.Local{{Name: "record", Type: ir.Object, Function: -1}},
		Main: []ir.Statement{
			ir.Declare{Local: 0, Value: ir.ObjectLiteral{
				Record: true,
				Fields: []ir.Field{{Name: "pos", Value: ir.NumberConstant{Value: 1}}},
			}},
			read(""),
			read("record.pos"),
			ir.Evaluate{Value: ir.Property{
				Object: ir.Read{Local: 0, Of: ir.Object},
				Name:   "end",
				Of:     ir.Number,
				View:   "record.end",
			}},
		},
	}
	readiness(program)
	views := []string{}
	walk(program.Main, func(node any) bool {
		if property, ok := node.(ir.Property); ok {
			views = append(views, property.Name+"="+property.View)
		}
		return true
	})
	if got, want := strings.Join(views, " "), "pos= pos=record.pos end="; got != want {
		test.Fatalf("want views %q after readiness, got %q", want, got)
	}
}
