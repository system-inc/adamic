package lower

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
)

// functionConstructorHonest is tsc's node constructor, small: utilities.ts's `function Node(this: Mutable<Node>,
// kind, pos, end)`, writing every field its this declares, in the order its fields are laid out.
const functionConstructorHonest = "function NodeObject(this: Nd, kind: number, pos: number, end: number): void {\n" +
	"    this.pos = pos;\n" +
	"    this.end = end;\n" +
	"    this.kind = kind;\n" +
	"}\n"

// functionConstructorProgram is tsc's road from objectAllocator to a node, small (#837zjkt): constructor, an
// allocator const whose getter hands it out `as any` (utilities.ts:8553), extra declarations, a module cache
// filled by the || at the new (parser.ts:433), and two nodes made through it and read, then tail.
func functionConstructorProgram(constructor, extra, tail string) string {
	return "interface Nd { kind: number; pos: number; end: number }\n" +
		"interface Allocator { getNodeConstructor(): new (kind: number, pos: number, end: number) => Nd }\n" +
		constructor +
		"const objectAllocator: Allocator = {\n    getNodeConstructor: () => NodeObject as any,\n};\n" +
		extra +
		"let NodeConstructor: (new (kind: number, pos: number, end: number) => Nd) | undefined;\n" +
		"function createBaseNode(kind: number): Nd {\n" +
		"    return new (NodeConstructor || (NodeConstructor = objectAllocator.getNodeConstructor()))(kind, -1, -1);\n" +
		"}\n" +
		"const first = createBaseNode(7);\n" +
		"const second = createBaseNode(9);\n" +
		"console.log(String(first.kind) + ' ' + String(first.pos) + ' ' + String(first.end));\n" +
		"console.log(String(second.kind));\n" +
		tail
}

// A function constructor handed out by an allocator getter `as any` and newed through the cache its || fills
// lowers as its synthesized class, its object born as a checked view of Nd, and agrees with Node on the
// JavaScript backend and natively (Adamic's ruling on #837zjkt).
// Mutant: m1, allocatorGetterShape admitting nothing, stops this at the getter's cast.
func TestFunctionConstructorThroughAllocatorAgrees(t *testing.T) {
	t.Parallel()
	program := lowersAndAgreesWithNodeAs(t, functionConstructorProgram(functionConstructorHonest, "", ""), ".ts")
	code, out, errs := brandedNative(t, program)
	if code != 0 || out != "7 -1 -1\n9\n" {
		t.Fatalf("native: exit %d stdout %q stderr %q; want exit 0 and Node's %q", code, out, errs, "7 -1 -1\n9\n")
	}
}

// The lie (#837zjkt's pin): a function constructor whose this claims a field its body never writes, newed
// through the road. Node reads the missing end as undefined; Adamic's object is born as a checked view of Nd,
// so the read of end stops with exit 70, before the line that reads it prints, on both backends.
// Mutant: m5, newClassValue giving the object no view, lets the JavaScript backend print Node's lie.
func TestFunctionConstructorLieStopsAtTheRead(t *testing.T) {
	t.Parallel()
	lie := "function NodeObject(this: Nd, kind: number, pos: number, end: number): void {\n" +
		"    this.pos = pos;\n" +
		"    this.kind = kind;\n" +
		"}\n"
	path := writeAgreementSource(t, "main.ts", functionConstructorProgram(lie, "", ""))
	node := runAgreementNode(t, path)
	if node.code != 0 || string(node.stdout) != "7 -1 undefined\n9\n" {
		t.Fatalf("source Node: exit %d stdout %q, want exit 0 and the lie read", node.code, node.stdout)
	}
	program := lowerAgreementEntry(t, path)
	generated := filepath.Join(t.TempDir(), "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	backend := runAgreementNode(t, generated)
	if backend.code != 70 || string(backend.stdout) != "" || !strings.HasPrefix(string(backend.stderr), "adamic: panic: ") {
		t.Fatalf("JavaScript backend: exit %d stdout %q stderr %q; want exit 70 at the checked read of end, before anything prints", backend.code, backend.stdout, backend.stderr)
	}
	code, out, errs := brandedNative(t, program)
	if code != 70 || out != "" || !strings.HasPrefix(errs, "adamic: panic: ") {
		t.Fatalf("native: exit %d stdout %q stderr %q; want exit 70 at the checked read of end, before anything prints", code, out, errs)
	}
}

// Every other road stops (#837zjkt). The premise pin: a program that writes, aliases or hands on the allocator,
// each its own row, turns its new NotYet, since what reaches the new's constructor isn't known. A use of the
// any's value other than as a new's constructor, a comparison or a pass of the cache or a pass of a getter's
// result, gives back the cast's own refusal, as does a cast outside the getter's shape. A constructor whose
// prototype the program reads, or that returns a value, reads this, hands this to a call or writes a field on
// one path is no function constructor, so its getter's cast keeps its refusal. A new that may be a class or a
// function constructor is NotYet. In .a the road stays refused at its cast.
//
// The assign and store rows are pinned by reading, not yet by a run: Object.assign and a method store may stop
// at their own words before the new does, which is fail closed too, and the row then takes those words.
// Mutants: m2, every allocator sealed, lowers the store, pass and alias rows; m3, the uses always admitted, lowers
// the comparison, cache pass and getter pass rows; m4, a mixed dispatch admitted, lowers the mixed row's new.
func TestFunctionConstructorOtherRoadsStop(t *testing.T) {
	t.Parallel()
	const refused = "a cast the runtime can't check"
	for _, probe := range []struct {
		name, source, extension, words string
	}{
		{"Object.assign into the allocator", functionConstructorProgram(functionConstructorHonest, "const other: Allocator = {\n    getNodeConstructor: () => NodeObject as any,\n};\nfunction setObjectAllocator(alloc: Allocator): void {\n    Object.assign(objectAllocator, alloc);\n}\n", "setObjectAllocator(other);\n"), ".ts", functionConstructorUnsealed},
		{"a store through the allocator", functionConstructorProgram(functionConstructorHonest, "", "objectAllocator.getNodeConstructor = objectAllocator.getNodeConstructor;\n"), ".ts", functionConstructorUnsealed},
		{"a pass of the allocator", functionConstructorProgram(functionConstructorHonest, "function patch(allocator: Allocator): Allocator {\n    return allocator;\n}\n", "patch(objectAllocator);\n"), ".ts", functionConstructorUnsealed},
		{"an alias of the allocator", functionConstructorProgram(functionConstructorHonest, "", "const alias: Allocator = objectAllocator;\n"), ".ts", functionConstructorUnsealed},
		{"a comparison of the cache", functionConstructorProgram(functionConstructorHonest, "", "console.log(String(NodeConstructor === undefined));\n"), ".ts", refused},
		{"a pass of the cache", functionConstructorProgram(functionConstructorHonest, "function keep(value: (new (kind: number, pos: number, end: number) => Nd) | undefined): void {}\n", "keep(NodeConstructor);\n"), ".ts", refused},
		{"a pass of a getter's result", functionConstructorProgram(functionConstructorHonest, "function keep(value: new (kind: number, pos: number, end: number) => Nd): void {}\n", "keep(objectAllocator.getNodeConstructor());\n"), ".ts", refused},
		{"a cast outside the getter", functionConstructorProgram(functionConstructorHonest, "", "const length: number = (NodeObject as any).length;\n"), ".ts", refused},
		{"a prototype read", functionConstructorProgram(functionConstructorHonest, "", "console.log(typeof NodeObject.prototype);\n"), ".ts", refused},
		{"a constructor returning a value", functionConstructorProgram("function NodeObject(this: Nd, kind: number, pos: number, end: number): Nd {\n    this.pos = pos;\n    this.end = end;\n    this.kind = kind;\n    return this;\n}\n", "", ""), ".ts", refused},
		{"a constructor reading this", functionConstructorProgram("function NodeObject(this: Nd, kind: number, pos: number, end: number): void {\n    this.pos = pos;\n    this.end = this.pos;\n    this.kind = kind;\n}\n", "", ""), ".ts", refused},
		{"a constructor handing this to a call", functionConstructorProgram("function kindOf(node: Nd): number {\n    return node.pos;\n}\nfunction NodeObject(this: Nd, kind: number, pos: number, end: number): void {\n    this.pos = pos;\n    this.end = end;\n    this.kind = kindOf(this);\n}\n", "", ""), ".ts", refused},
		{"a constructor writing on one path", functionConstructorProgram("function NodeObject(this: Nd, kind: number, pos: number, end: number): void {\n    this.pos = pos;\n    this.end = end;\n    if (kind > 0) {\n        this.kind = kind;\n    } else {\n        this.kind = 0;\n    }\n}\n", "", ""), ".ts", refused},
		{"a new that may be a class", functionConstructorProgram(functionConstructorHonest, "class Real {\n    static readonly label = 'real';\n    kind: number;\n    pos: number;\n    end: number;\n    constructor(kind: number, pos: number, end: number) {\n        this.kind = kind;\n        this.pos = pos;\n        this.end = end;\n    }\n}\n", "console.log(Real.label);\n"), ".ts", "a constructor value that may be a class or a function constructor"},
		{"in .a", functionConstructorProgram(functionConstructorHonest, "", ""), ".a", refused},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			if err == nil || !strings.Contains(err.Error(), probe.words) {
				t.Fatalf("want a stop holding %q, got %v", probe.words, err)
			}
		})
	}
}
