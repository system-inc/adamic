package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNamedRecordReadTypes(t *testing.T) {
	t.Parallel()
	p, err := lowerSource(t, `interface Options { target?:number; [key:string]:number|string|undefined } function named(o:Options):number|undefined{return o.target;} function indexed(o:Options,k:string):number|string|undefined{return o[k];}`)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range p.Functions {
		if f.Name == "named" {
			r := f.Body[len(f.Body)-1].(ir.Return)
			if r.Value.Type() != ir.MaybeNumber {
				t.Fatalf("named read used index type %v", r.Value.Type())
			}
		}
		if f.Name == "indexed" {
			r := f.Body[len(f.Body)-1].(ir.Return)
			if r.Value.Type() != ir.Union {
				t.Fatalf("indexed read lost index type %v", r.Value.Type())
			}
		}
	}
}

// A slot whose union holds null boxes it, as tsc's CompilerOptions does. A declared read still unboxes through the
// checked tag, since a write through the index signature can put any value of the union under a declared key.
func TestNullableRecordSlots(t *testing.T) {
	t.Parallel()
	p, err := lowerSource(t, `interface R {named?:number;[key:string]:number|string|boolean|null|undefined} function named(r:R):number|undefined{return r.named;} function indexed(r:R,k:string):number|string|boolean|null|undefined{return r[k];}`)
	if err != nil {
		t.Fatal(err)
	}
	checked := false
	for _, f := range p.Functions {
		switch f.Name {
		case "named":
			if r := f.Body[len(f.Body)-1].(ir.Return); r.Value.Type() != ir.MaybeNumber {
				t.Fatalf("named read used the slot type %v", r.Value.Type())
			}
		case "indexed":
			if r := f.Body[len(f.Body)-1].(ir.Return); r.Value.Type() != ir.Union {
				t.Fatalf("indexed read lost the slot type %v", r.Value.Type())
			}
		case "record_declared_member":
			checked = true
		}
	}
	if !checked {
		t.Fatal("the declared read skipped the checked unbox")
	}
}

// typeof says object for a dictionary and a fixed object alike, so a local narrowed to one of them can't be checked
// against the other after a call writes it again. Without the dictionary among the kinds that stop, native read the
// record's storage as a Box and printed 4.6e-310 where Node prints 5.
func TestNarrowedUnionMemberBesideRecord(t *testing.T) {
	t.Parallel()
	for _, declared := range []string{"Box | Record<string, number>", "Box | Record<string, number> | null"} {
		t.Run(declared, func(t *testing.T) {
			_, err := lowerSource(t, `class Box { v = 1; }
let x: `+declared+` = new Box();
function swap(): void { const d: Record<string, number> = {}; d["v"] = 5; x = d; }
if (x instanceof Box) { swap(); console.log(String(x.v)); }`)
			var ny *NotYet
			if !errors.As(err, &ny) || !strings.Contains(err.Error(), "object tag cannot be checked with typeof") {
				t.Fatalf("got %v, want the object-tag NotYet", err)
			}
		})
	}
}

func TestNamedRecordRefusals(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{`interface R {named?:number[];[key:string]:number[]|string|undefined} function read(r:R):number|undefined {return r.named?.length;}`, "dictionary member narrowed to an object kind"},
		{`interface R {named?:number;[key:string]:number|string|undefined} function write(r:R,k:string):void {r[k]="wrong";}`, "named property named"},
		{`interface R {named?:number;[key:string]:number|string|undefined} const r:R={}; const view:Record<string,number|string|undefined>=r;`, "seen as"},
		{`interface R {readonly named:number;[key:string]:number} const r:R={named:1}; const view:Record<string,number>=r;`, "seen as"},
		{`interface R {named:number;[key:string]:number} const fixed={named:1}; const r:R=fixed;`, "seen as"},
		{`interface R {named?:number;readonly [key:string]:number|undefined} const r:R={};`, "signature"},
		{`interface R {named:number;[key:number]:number} const r:R={named:1};`, "signature"},
	} {
		t.Run(probe.source, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var ny *NotYet
			if !errors.As(err, &ny) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want NotYet %s", err, probe.reason)
			}
		})
	}
}

// A non-literal-key write lands under whichever named property the key hits, so it's judged by the whole
// refusal registry with each named property's type as its slot (#b6m00y7, #npqr3sx), and stops as the
// direct write to that property does: tsc relates the value to a named property structurally, and
// recordNamedWrite's own list missed nominal ancestry, so native ran Point's describe on an Other and
// printed POINT where Node prints OTHER (Language's w8), and void widening, so a () => void was read as
// () => unknown (w10). A record literal's computed name is the same write, made as the literal is. An
// honest write, a number into number properties, still lowers. The literal's twin, { names: quiet }, is
// stopped a level up, at the whole literal's initializer through .names, so a construction is held to its
// twin's kind and rule (the fix it names) rather than to where it's said. Language's w12 to w14 reach the
// same write through an alias, a chained assignment and a parameter. In a slot boxed by a number in the
// index, a declared property is left to its read only when that read is proven to stop on a lie
// (readGuardedInUnionSlot): w10's literal there (b10) is judged, since a function's tag says only
// "function", and a plain object's (b8), an array's (a8), a Map's (p8) and a dictionary's (r8) reads stop.
// Mutants: slotType ignoring the slot a caller set lowers w8, which then prints POINT natively;
// recordComputedEntry skipped lowers w8's literal and b10; refuseWidening's rules skipped in refuseInSlot's
// run lowers the invariance and enum pins.
func TestNamedRecordWritesAreJudgedAsTheirDirectTwins(t *testing.T) {
	t.Parallel()
	w8 := `interface Shape { x: number; describe(): string }
class Point { x = 1; describe(): string { return "point"; } }
class Other { x = 2; label = "other"; describe(): string { return this.label; } }
type Places = { names: Point; [k: string]: Point | Shape };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
`
	w10 := `type Handlers = { names: () => unknown; [k: string]: () => void };
const keys: string[] = ["names", "other"];
const quiet: () => void = () => "ab" + String(keys.length);
const first = (): unknown => undefined;
const key = keys[0] ?? "other";
`
	// The same shapes in a slot boxed by a number in the index.
	boxed := func(source string) string {
		return strings.Replace(strings.Replace(source, "[k: string]: Point | Shape }", "[k: string]: Point | Shape | number }", 1), "[k: string]: () => void }", "[k: string]: (() => void) | number }", 1)
	}
	for _, probe := range []struct {
		name, source, write, direct string
		byRule                      bool
	}{
		{"w8", w8, "const places: Places = { names: new Point() };\nplaces[key] = new Other();", "const places: Places = { names: new Point() };\nplaces.names = new Other();", false},
		{"w10", w10, "const handlers: Handlers = { names: first };\nhandlers[key] = quiet;", "const handlers: Handlers = { names: first };\nhandlers.names = quiet;", false},
		// The list's widened and enumAssignable went with it, since the registry judges both at the direct
		// write (Lower's and Compiler's pins on 82ae5693). The index type is readonly Animal[], which can't write,
		// so the site's own view of the index stops nothing, and isn't boxed as a union, which recordComputedEntry
		// leaves to its reads; only pets can stop it: a Dog[] under pets: Animal[] could be written
		// an Animal the dogs are read as (adamic/invariant-mutable), and a number under kind: Kind.Text
		// breaks the member's literal (#w7w2asn's slot refusal). Each through a key and through a literal.
		{"invariance", "class Animal { name = \"a\"; }\nclass Dog extends Animal { bark(): string { return \"woof\"; } }\ntype Kennels = { pets: Animal[]; [k: string]: readonly Animal[] };\nconst keys: string[] = [\"pets\", \"other\"];\nconst key = keys[0] ?? \"other\";\nconst dogs: Dog[] = [];\n", "const kennels: Kennels = { pets: [] };\nkennels[key] = dogs;", "const kennels: Kennels = { pets: [] };\nkennels.pets = dogs;", false},
		{"invariance construction", "class Animal { name = \"a\"; }\nclass Dog extends Animal { bark(): string { return \"woof\"; } }\ntype Kennels = { pets: Animal[]; [k: string]: readonly Animal[] };\nconst keys: string[] = [\"pets\", \"other\"];\nconst key = keys[0] ?? \"other\";\nconst dogs: Dog[] = [];\n", "const kennels: Kennels = { pets: [new Animal()], [key]: dogs };", "const kennels: Kennels = { pets: dogs };", true},
		{"an open enum's member", "enum Kind { Text, Count }\ntype Nodes = { kind: Kind.Text; [k: string]: number };\nconst keys: string[] = [\"kind\", \"other\"];\nconst key = keys[0] ?? \"other\";\nconst n: number = 7;\n", "const nodes: Nodes = { kind: Kind.Text };\nnodes[key] = n;", "const nodes: Nodes = { kind: Kind.Text };\nnodes.kind = n;", false},
		{"an open enum's member construction", "enum Kind { Text, Count }\ntype Nodes = { kind: Kind.Text; [k: string]: number };\nconst keys: string[] = [\"kind\", \"other\"];\nconst key = keys[0] ?? \"other\";\nconst n: number = 7;\n", "const nodes: Nodes = { kind: Kind.Text, [key]: n };", "const nodes: Nodes = { kind: n };", true},
		// A host object under a declared property: lowering's host-view check sees the index type the value
		// is stored as (Clock, whose now is registered), so the declared property's view (View, which reaches
		// the unregistered eventLoopUtilization) has to be judged by the write itself (Lower's probe).
		{"a node host object", "import { performance } from \"node:perf_hooks\";\ntype Clock = { now(): number };\ntype View = { now(): number; eventLoopUtilization(): { readonly idle: number } };\ntype Places = { names: View; [k: string]: Clock };\nconst keys: string[] = [\"names\"];\nconst key = keys[0] ?? \"names\";\nconst initial: View = { now: () => 1, eventLoopUtilization: () => ({ idle: 1 }) };\nconst places: Places = { names: initial };\n", "places[key] = performance;\nconsole.log(String(places.names.eventLoopUtilization().idle >= 0));", "places.names = performance;\nconsole.log(String(places.names.eventLoopUtilization().idle >= 0));", false},
		{"w8 construction", w8, "const places: Places = { names: new Point(), [key]: new Other() };", "const places: Places = { names: new Other() };", true},
		{"w10 construction", w10, "const handlers: Handlers = { names: first, [key]: quiet };", "const handlers: Handlers = { names: quiet };", true},
		{"w12 alias receiver", w8, "const places: Places = { names: new Point() };\nconst both: Places[] = [places];\nconst slot = both[0];\nif (slot !== undefined) slot[key] = new Other();", "const places: Places = { names: new Point() };\nconst both: Places[] = [places];\nconst slot = both[0];\nif (slot !== undefined) slot.names = new Other();", false},
		{"w13 chained", w8, "const places: Places = { names: new Point() };\nlet spare: Point | Shape = new Point();\nspare = places[key] = new Other();", "const places: Places = { names: new Point() };\nlet spare: Point | Shape = new Point();\nspare = places.names = new Other();", false},
		{"w14 parameter", w8, "const places: Places = { names: new Point() };\nconst write = (target: Places, at: string): void => { target[at] = new Other(); };\nwrite(places, key);", "const places: Places = { names: new Point() };\nconst write = (target: Places, at: string): void => { target.names = new Other(); };\nwrite(places, key);", false},
		{"w10 construction boxed", boxed(w10), "const handlers: Handlers = { names: first, [key]: quiet };", "const handlers: Handlers = { names: quiet };", true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, written := lowerSource(t, probe.source+probe.write+"\nconsole.log('ran');\n")
			_, direct := lowerSource(t, probe.source+probe.direct+"\nconsole.log('ran');\n")
			if direct == nil {
				t.Fatalf("the direct twin %q lowers, so it proves nothing about the write", probe.direct)
			}
			same := written != nil && stoppedFor(written) == stoppedFor(direct)
			if probe.byRule {
				same = written != nil && ruleOf(written) == ruleOf(direct)
			}
			if !same {
				t.Fatalf("the write stopped as %v, want what its direct twin stops as: %v", written, direct)
			}
		})
	}
	// The reads that hold readGuardedInUnionSlot: a lie under a declared plain object (b8, w8's literal
	// boxed), array (a8), Map (p8) or dictionary (r8) in a boxed slot is left to its read, which stops as
	// an object kind with no checked tag. Each goes red if the skip stops naming its kind (the write's
	// stop comes first) or the read stops stopping. A computed method is no data entry, a stop of its own.
	for _, probe := range []struct{ name, source, reason string }{
		{"b8 object construction boxed", boxed(w8) + "const places: Places = { names: new Point(), [key]: new Other() };\nconsole.log(places.names.describe().toUpperCase());\n", "narrowed to an object kind without a checked runtime tag"},
		{"a8 array construction boxed", `type Lists = { names: string[]; [k: string]: string[] | number[] | number };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
const lists: Lists = { names: ["a"], [key]: [1, 2] };
console.log((lists.names[0] ?? "").toUpperCase());
`, "narrowed to an object kind without a checked runtime tag"},
		{"p8 map construction boxed", `type Tables = { names: Map<string, string>; [k: string]: Map<string, string> | Map<string, number> | number };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
const tables: Tables = { names: new Map<string, string>([["a", "b"]]), [key]: new Map<string, number>([["a", 1]]) };
console.log((tables.names.get("a") ?? "").toUpperCase());
`, "narrowed to an object kind without a checked runtime tag"},
		{"r8 dictionary construction boxed", `interface MapLike<T> { [index: string]: T; }
type Tables = { names: MapLike<string>; [k: string]: MapLike<string> | MapLike<number> | number };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
const tables: Tables = { names: { a: "b" }, [key]: { a: 1 } };
console.log((tables.names["a"] ?? "").toUpperCase());
`, "narrowed to an object kind without a checked runtime tag"},
		{"w10 computed method", w10 + "const handlers: Handlers = { names: first, [key]() { return \"x\"; } };\nconsole.log('ran');\n", "a record literal member other than a data entry"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var ny *NotYet
			if !errors.As(err, &ny) || !strings.Contains(err.Error(), probe.reason) {
				t.Fatalf("got %v, want the NotYet %s", err, probe.reason)
			}
		})
	}
	if _, err := lowerSource(t, `interface Options { target?: number; required: number; [option: string]: number | undefined }
function write(options: Options, key: string): void { options[key] = 9; }`); err != nil {
		t.Fatalf("an honest write: %v", err)
	}
	// A computed name strings joined by + make is judged as that name: tsc's ["" + "constructor"] beside
	// the keywords its literal declares names none of them (01_scanner_keywords.a's shape).
	if _, err := lowerSource(t, "const SyntaxKind = { AbstractKeyword: 128, AccessorKeyword: 129, ConstructorKeyword: 137 } as const;\ntype KeywordSyntaxKind = (typeof SyntaxKind)[keyof typeof SyntaxKind];\ninterface MapLike<T> { [index: string]: T; }\nconst kinds: MapLike<KeywordSyntaxKind> = {\n    abstract: SyntaxKind.AbstractKeyword,\n    accessor: SyntaxKind.AccessorKeyword,\n    [\"\" + \"constructor\"]: SyntaxKind.ConstructorKeyword,\n};\nconsole.log(`${String(kinds[\"abstract\"])} ${Object.keys(kinds).join(\",\")}`);\n"); err != nil {
		t.Fatalf("a folded name no named property takes: %v", err)
	}
}

// ruleOf is whether an error is a refusal or a NotYet, and the fix a refusal names, which names its rule.
func ruleOf(err error) string {
	var refused *Refused
	if errors.As(err, &refused) {
		return "Refused: " + refused.Fix
	}
	return stoppedFor(err)
}
