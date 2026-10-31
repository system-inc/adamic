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

// A non-literal-key write is admitted only when the value's widened type is each named property's type
// itself (#npqr3sx, Language's w8 and w10): tsc relates the value to a named property structurally, and
// recordNamedWrite's list missed nominal ancestry, so native ran Point's describe on an Other and printed
// POINT where Node prints OTHER (w8), and void widening, so a () => void was read as () => unknown (w10).
// A record literal's computed name is the same write, made as the literal is: w8's and w10's shapes
// through { names: ..., [key]: ... } are not yet too. An honest write, a number into number properties,
// still lowers. Language's w12 to w14 reach the same write through an alias, a chained assignment and
// a parameter, and w10's literal in a slot boxed by a number is judged too, since a declared function
// read from a boxed slot is checked only by its tag. Mutants: the identity check dropped lowers w8,
// w10, both literals, w12 to w14 and the boxed w10, and w8's print POINT natively; recordComputedEntry
// skipped lowers w8's literal, which prints POINT natively, and w10's two literals.
func TestNamedRecordWritesAreThePropertyTypeItself(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"w8", `interface Shape { x: number; describe(): string }
class Point { x = 1; describe(): string { return "point"; } }
class Other { x = 2; label = "other"; describe(): string { return this.label; } }
type Places = { names: Point; [k: string]: Point | Shape };
const places: Places = { names: new Point() };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
places[key] = new Other();
console.log(places.names.describe().toUpperCase());
`},
		{"w10", `type Handlers = { names: () => unknown; [k: string]: () => void };
const keys: string[] = ["names", "other"];
const quiet: () => void = () => "ab" + String(keys.length);
const first = (): unknown => undefined;
const handlers: Handlers = { names: first };
const key = keys[0] ?? "other";
handlers[key] = quiet;
const seen = handlers.names();
console.log(typeof seen === "string" ? seen.toUpperCase() : "none");
`},
		{"w8 construction", `interface Shape { x: number; describe(): string }
class Point { x = 1; describe(): string { return "point"; } }
class Other { x = 2; label = "other"; describe(): string { return this.label; } }
type Places = { names: Point; [k: string]: Point | Shape };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
const places: Places = { names: new Point(), [key]: new Other() };
console.log(places.names.describe().toUpperCase());
`},
		{"w10 construction", `type Handlers = { names: () => unknown; [k: string]: () => void };
const keys: string[] = ["names", "other"];
const quiet: () => void = () => "ab" + String(keys.length);
const first = (): unknown => undefined;
const key = keys[0] ?? "other";
const handlers: Handlers = { names: first, [key]: quiet };
const seen = handlers.names();
console.log(typeof seen === "string" ? seen.toUpperCase() : "none");
`},
		{"w12 alias receiver", `interface Shape { x: number; describe(): string }
class Point { x = 1; describe(): string { return "point"; } }
class Other { x = 2; label = "other"; describe(): string { return this.label; } }
type Places = { names: Point; [k: string]: Point | Shape };
const places: Places = { names: new Point() };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
const both: Places[] = [places];
const slot = both[0];
if (slot !== undefined) slot[key] = new Other();
console.log(places.names.describe().toUpperCase());
`},
		{"w13 chained", `interface Shape { x: number; describe(): string }
class Point { x = 1; describe(): string { return "point"; } }
class Other { x = 2; label = "other"; describe(): string { return this.label; } }
type Places = { names: Point; [k: string]: Point | Shape };
const places: Places = { names: new Point() };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
let spare: Point | Shape = new Point();
spare = places[key] = new Other();
console.log(places.names.describe().toUpperCase());
`},
		{"w14 parameter", `interface Shape { x: number; describe(): string }
class Point { x = 1; describe(): string { return "point"; } }
class Other { x = 2; label = "other"; describe(): string { return this.label; } }
type Places = { names: Point; [k: string]: Point | Shape };
const places: Places = { names: new Point() };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
const write = (target: Places, at: string): void => { target[at] = new Other(); };
write(places, key);
console.log(places.names.describe().toUpperCase());
`},
		{"w10 construction boxed", `type Handlers = { names: () => unknown; [k: string]: (() => void) | number };
const keys: string[] = ["names", "other"];
const quiet: () => void = () => "ab" + String(keys.length);
const first = (): unknown => undefined;
const key = keys[0] ?? "other";
const handlers: Handlers = { names: first, [key]: quiet };
const seen = handlers.names();
console.log(typeof seen === "string" ? seen.toUpperCase() : "none");
`},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.source)
			var ny *NotYet
			if !errors.As(err, &ny) || !strings.Contains(err.Error(), "type itself") {
				t.Fatalf("got %v, want the NotYet for a value that isn't the named property's type itself", err)
			}
		})
	}
	// Stops of their own that the boxed slot and a computed method rely on: w8's literal in a slot boxed
	// by a number in the index reads names as an object kind with no checked tag, and a computed method
	// is no data entry. Each goes red if the code it relies on stops holding.
	for _, probe := range []struct{ name, source, reason string }{
		{"w8 construction boxed", `interface Shape { x: number; describe(): string }
class Point { x = 1; describe(): string { return "point"; } }
class Other { x = 2; label = "other"; describe(): string { return this.label; } }
type Places = { names: Point; [k: string]: Point | Shape | number };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
const places: Places = { names: new Point(), [key]: new Other() };
console.log(places.names.describe().toUpperCase());
`, "narrowed to an object kind without a checked runtime tag"},
		{"w10 computed method", `type Handlers = { names: () => unknown; [k: string]: () => void };
const keys: string[] = ["names", "other"];
const first = (): unknown => undefined;
const key = keys[0] ?? "other";
const handlers: Handlers = { names: first, [key]() { return "x"; } };
const seen = handlers.names();
console.log(typeof seen === "string" ? seen.toUpperCase() : "none");
`, "a record literal member other than a data entry"},
	} {
		t.Run(probe.name, func(t *testing.T) {
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
