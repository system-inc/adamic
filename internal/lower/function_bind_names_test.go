package lower

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

const boundNamePrelude = `type Unary = (x: number) => number;
function one(x: number): number { return x + 1; }
function two(x: number, y: number = 1): number { return x + y; }
const held: Unary = two;
console.log(String(held(1)));
const otherReceiver = {};
const host: { f: Unary } = { f: (x: number): number => x + 1 };
`

const boundNameCount = "bind of a function that declares 2 parameters where the type it's bound as names 1"

// The bound receiver is deliberately separate from the writes: a name's union includes writes
// elsewhere, proven fresh writes included. The closure call in the prelude makes reach unknown.
func TestBoundNameReach(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, writes string
		stops        bool
	}{
		{"no writes", "", false},
		{"undefined write", `const elsewhere: { f: Unary | undefined } = { f: undefined }; elsewhere.f = undefined;`, false},
		{"null write", `const elsewhere: { f: Unary | null } = { f: null }; elsewhere.f = null;`, false},
		{"field literal", `const elsewhere = { f: (x: number): number => x + 1 }; elsewhere.f = (x: number, y: number = 1): number => x + y;`, true},
		{"named function", `const elsewhere = { f: (x: number): number => x + 1 }; elsewhere.f = two;`, true},
		{"literal in function", `function elsewhere(): { f: Unary } { return { f: (x: number, y: number = 1): number => x + y }; } console.log(String(elsewhere().f(1)));`, true},
		{"unreadable alias", `const elsewhere = { f: (x: number): number => x + 1 }; elsewhere.f = held;`, true},
		{"computed union", `const elsewhere: { f: Unary; g: Unary } = { f: (x: number): number => x + 1, g: one }; function put(key: 'f' | 'g'): void { elsewhere[key] ||= held; } put('f');`, true},
		{"assign result", `const target = { f: (x: number): number => x + 1 }; Object.assign(target, {});`, true},
		{"spread", `const original = { f: (x: number): number => x + 1 }; const copied = { ...original }; console.log(String(copied.f(1)));`, true},
		{"other name spread", `const original = { g: one }; const copied = { ...original }; console.log(String(copied.g(1)));`, false},
		{"other name computed union", `const elsewhere = { g: one, h: one }; function put(key: 'g' | 'h'): void { elsewhere[key] ||= held; } put('g');`, false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, boundNamePrelude+probe.writes+"\nconst bound = host.f.bind(otherReceiver); console.log(String(bound(1)));\n", ".ts")
			if !probe.stops {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != boundNameCount {
				t.Fatalf("want NotYet %q, got %v", boundNameCount, err)
			}
		})
	}
}

func TestBoundNameHostMember(t *testing.T) {
	t.Parallel()
	_, err := lowerModules(t, map[string]string{
		"host.d.ts": "export interface Host { f: (x: number) => number; }\n",
		"main.ts":   "import type { Host } from './host.d.ts';\n" + strings.Replace(boundNamePrelude, "const host: { f: Unary }", "const host: Host", 1) + "const bound = host.f.bind(otherReceiver); console.log(String(bound(1)));\n",
	})
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != boundNameCount {
		t.Fatalf("want host-member fallback NotYet %q, got %v", boundNameCount, err)
	}
}

func TestBoundNameSubclassOverride(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, `class Base { m(x: number, y: number = 0): number { return x + y; } }
class Derived extends Base { override m(x: number, y: number = 1): number { return x + y; } }
const host: { m(x: number): number } = new Derived();
const bound = host.m.bind(host);
console.log(String(bound(1)));
`, ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != boundNameCount {
		t.Fatalf("want override NotYet %q, got %v", boundNameCount, err)
	}
}

// These operations stop before producing a shaped field write. Keep those stops explicit rather
// than pretending fresh sees a WriteUnknown for an operation that never reaches it.
func TestBoundNameEarlierWriteStops(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, writes, words string }{
		{"assign", `Object.assign(host, { f: (x: number): number => x + 1 });`, "Object.assign on a shape not proven"},
		{"any", `function put(receiver: any, key: string): void { receiver[key] = held; } put(host, 'f');`, "any"},
		{"defineProperty", `function put(key: string): void { Object.defineProperty(host, key, { value: held }); } put('f');`, "property descriptors"},
		{"override count", `class Base { m(x: number): number { return x; } } class Derived extends Base { override m(x: number, y: number = 1): number { return x + y; } } const derived = new Derived(); console.log(String(derived.m(1)));`, "an override with a different parameter count"},
		{"Reflect.set", `function put(key: string): void { Reflect.set(host, key, held); } put('f');`, "namespace member without a lowered binding"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, boundNamePrelude+probe.writes+"\nconst bound = host.f.bind(otherReceiver); console.log(String(bound(1)));\n", ".ts")
			if err == nil || !strings.Contains(err.Error(), probe.words) {
				t.Fatalf("want %q, got %v", probe.words, err)
			}
		})
	}
}

func TestBoundNameExistingPins(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"narrow_type", "absent_property", "present_property", "absent_fixed_property", "null_property"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile("../oracle/testdata/bind_extra_parameters/" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSourceAs(t, string(source), ".ts")
			if name == "narrow_type" {
				var notYet *NotYet
				if !errors.As(err, &notYet) || notYet.What != boundNameCount {
					t.Fatalf("b01: want NotYet %q, got %v", boundNameCount, err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// These types decide which names an unresolved key can affect, even when the particular write
// syntax still stops before IR. Record storage cannot overwrite a shaped field of the same name.
func TestBoundNameReceiverTypes(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, parameters, receiver string
		holds                      bool
	}{
		{"any", "", "any", true},
		{"unknown", "", "unknown", true},
		{"shaped f", "", "{ f: number }", true},
		{"shaped g", "", "{ g: number }", false},
		{"union includes f", "", "{ f: number } | { g: number }", true},
		{"union omits f", "", "{ g: number } | { h: number }", false},
		{"record", "", "Record<string, number>", false},
		{"declared indexed f", "", "{ f: number; [name: string]: number }", false},
		{"unconstrained", "<T>", "T", true},
		{"constrained f", "<T extends { f: number }>", "T", true},
		{"constrained g", "<T extends { g: number }>", "T", false},
		{"constrained record", "<T extends Record<string, number>>", "T", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source := "function put" + probe.parameters + "(receiver: " + probe.receiver + "): void {}\n"
			program, err := load.Load([]string{writeAgreementSource(t, "main.ts", source)})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			l := &lowering{program: program, checker: checked}
			parameter := file.Statements.Nodes[0].Parameters()[0]
			if got := l.boundReceiverHolds(checked.GetTypeAtLocation(parameter.Name()), "f"); got != probe.holds {
				t.Fatalf("receiver %s holds f: got %t, want %t", probe.receiver, got, probe.holds)
			}
		})
	}
}

// A future operation and a lost write site must not look like an absent field. Hand-built IR
// isolates these boundaries without relying on source syntax that lowering already refuses.
func TestBoundNameIncompleteWrites(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name string
		body []ir.Statement
	}{
		{"unknown", []ir.Statement{ir.Evaluate{Value: ir.NodeBufferCall{Function: "future_buffer", Returns: ir.Object}}}},
		{"siteless", []ir.Statement{ir.SetProperty{Object: ir.ObjectLiteral{}, Name: "f", Value: ir.MakeClosure{Function: 1}}}},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := load.Load([]string{writeAgreementSource(t, "main.ts", boundNamePrelude+"host.f.bind(otherReceiver);\n")})
			if err != nil {
				t.Fatal(err)
			}
			file := program.Files()[0]
			checked, release := program.Checker(context.Background(), file)
			defer release()
			var bind *ast.Node
			var visit ast.Visitor
			visit = func(node *ast.Node) bool {
				if boundMemberTarget(node) != nil {
					bind = node
				}
				node.ForEachChild(visit)
				return false
			}
			file.AsNode().ForEachChild(visit)
			if bind == nil {
				t.Fatal("no property bind in the witness")
			}
			result := &ir.Program{Functions: make([]ir.Function, 2), Main: probe.body}
			result.Functions[0].Source.Role = "bound_function"
			l := &lowering{program: program, checker: checked, result: result, closureRecords: []closureRecord{{function: 0, node: bind}}}
			if functions, readable := l.boundNameReach()(boundMemberTarget(bind)); readable {
				t.Fatalf("incomplete writes were readable as %v", functions)
			}
		})
	}
}

// A named read of record storage is still a record read. Its literal entries and indexed writes
// do not enter the shaped-field table, so an unknown record projection must keep the old count.
func TestBoundNameRecordTargetKeepsTypeWideFallback(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, `type Unary = (x: number) => number;
function two(x: number, y: number = 1): number { return x + y; }
const host: Record<string, Unary> = { f: two };
const otherReceiver = {};
const bound = host.f?.bind(otherReceiver);
if (bound !== undefined) console.log(String(bound(1)));
`, ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != boundNameCount {
		t.Fatalf("want record-target NotYet %q, got %v", boundNameCount, err)
	}
}
