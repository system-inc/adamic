package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNamedRecordReadTypes(test *testing.T) {
	test.Parallel()
	p, lowerSourceError := lowerSource(
		test,
		`interface Options { target?:number; [key:string]:number|string|undefined } function named(o:Options):number|undefined{return o.target;} function indexed(o:Options,k:string):number|string|undefined{return o[k];}`,
	)
	if lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
	for _, f := range p.Functions {
		if f.Name == "named" {
			r := f.Body[len(f.Body)-1].(ir.Return)
			if r.Value.Type() != ir.MaybeNumber {
				test.Fatalf("named read used index type %v", r.Value.Type())
			}
		}
		if f.Name == "indexed" {
			r := f.Body[len(f.Body)-1].(ir.Return)
			if r.Value.Type() != ir.Union {
				test.Fatalf("indexed read lost index type %v", r.Value.Type())
			}
		}
	}
}

// A slot whose union holds null boxes it, as tsc's CompilerOptions does. A declared read still unboxes through the
// checked tag, since a write through the index signature can put any value of the union under a declared key.
func TestNullableRecordSlots(test *testing.T) {
	test.Parallel()
	p, lowerSourceError := lowerSource(
		test,
		`interface R {named?:number;[key:string]:number|string|boolean|null|undefined} function named(r:R):number|undefined{return r.named;} function indexed(r:R,k:string):number|string|boolean|null|undefined{return r[k];}`,
	)
	if lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
	checked := false
	for _, f := range p.Functions {
		switch f.Name {
		case "named":
			if r := f.Body[len(f.Body)-1].(ir.Return); r.Value.Type() != ir.MaybeNumber {
				test.Fatalf("named read used the slot type %v", r.Value.Type())
			}
		case "indexed":
			if r := f.Body[len(f.Body)-1].(ir.Return); r.Value.Type() != ir.Union {
				test.Fatalf("indexed read lost the slot type %v", r.Value.Type())
			}
		case "record_declared_member":
			checked = true
		}
	}
	if !checked {
		test.Fatal("the declared read skipped the checked unbox")
	}
}

// typeof says object for a dictionary and a fixed object alike, so a local narrowed to one of them can't be checked
// against the other after a call writes it again. Without the dictionary among the kinds that stop, native read the
// record's storage as a Box and printed 4.6e-310 where Node prints 5.
func TestNarrowedUnionMemberBesideRecord(test *testing.T) {
	test.Parallel()
	for _, declared := range []string{"Box | Record<string, number>", "Box | Record<string, number> | null"} {
		test.Run(declared, func(test *testing.T) {
			_, lowerSourceError := lowerSource(test, `class Box { v = 1; }
let x: `+declared+` = new Box();
function swap(): void { const d: Record<string, number> = {}; d["v"] = 5; x = d; }
if (x instanceof Box) { swap(); console.log(String(x.v)); }`)
			var ny *NotYet
			if !errors.As(lowerSourceError, &ny) ||
				!strings.Contains(lowerSourceError.Error(), "object tag cannot be checked with typeof") {
				test.Fatalf("got %v, want the object-tag NotYet", lowerSourceError)
			}
		})
	}
}

// A fixed object or a dictionary read off a union slot is proven by a heap test only when it's the one
// member of its runtime class in the slot's union (recordSlotMemberUnique): tsc's configFile and paths
// read so. An array beside other arrays (rootDirs), two fixed objects, two dictionaries and a member whose
// heap class isn't vouched for (object, which may be either) keep the NotYet.
func TestNamedRecordHeapMemberReads(test *testing.T) {
	test.Parallel()
	options := func(index string) string {
		return `interface Source { text: string; diagnostics: string[] }
interface Other { label: string }
interface Options { configFile?: Source; paths?: { [index: string]: string[] }; rootDirs?: string[]; [option: string]: ` + index + ` }
`
	}
	tscShape := "string | number | boolean | string[] | number[] | { [index: string]: string[] } | Source | undefined"
	for _, probe := range []struct{ name, source string }{
		{"configFile", options(tscShape) +
			"function read(o: Options): string { const file = o.configFile; return file ? file.text : \"none\"; }"},
		{"paths", options(tscShape) +
			"function read(o: Options): string { const paths = o.paths; return paths ? Object.keys(paths).join(\",\") : \"none\"; }"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			p, lowerSourceError := lowerSource(test, probe.source)
			if lowerSourceError != nil {
				test.Fatal(lowerSourceError)
			}
			checked := false
			for _, f := range p.Functions {
				if f.Name == "record_declared_member" {
					narrow, ok := f.Body[len(f.Body)-1].(ir.Return).Value.(ir.Narrow)
					checked = checked || ok && narrow.Checked
				}
			}
			if !checked {
				test.Fatal("the declared read skipped the checked heap narrow")
			}
		})
	}
	// An array read stops under its own reason until the element-checked view lands (#y852c3c).
	_, lowerSourceError := lowerSource(test, options(tscShape)+
		"function read(o: Options): number { const roots = o.rootDirs; return roots ? roots.length : 0; }")
	var arrayStop *NotYet
	if !errors.As(lowerSourceError, &arrayStop) ||
		!strings.Contains(
			lowerSourceError.Error(),
			"a dictionary member narrowed to an array without an element-checked view",
		) {
		test.Fatalf("an array beside arrays: got %v, want the array NotYet", lowerSourceError)
	}
	for _, probe := range []struct{ name, source string }{
		{"two fixed objects", options(
			"string | string[] | { [index: string]: string[] } | Source | Other | undefined",
		) + "function read(o: Options): string { const file = o.configFile; return file ? file.text : \"none\"; }"},
		{"two dictionaries", options(
			"string | string[] | { [index: string]: string[] } | { [index: string]: number } | Source | undefined",
		) +
			"function read(o: Options): string { const paths = o.paths; return paths ? Object.keys(paths).join(\",\") : \"none\"; }"},
		{"an object member", options(
			"string | string[] | { [index: string]: string[] } | Source | object | undefined",
		) + "function read(o: Options): string { const file = o.configFile; return file ? file.text : \"none\"; }"},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceError := lowerSource(test, probe.source)
			var ny *NotYet
			if !errors.As(lowerSourceError, &ny) || !strings.Contains(
				lowerSourceError.Error(),
				"dictionary member narrowed to an object kind without a checked runtime tag",
			) {
				test.Fatalf("got %v, want the object-kind NotYet", lowerSourceError)
			}
		})
	}
}

func TestNamedRecordRefusals(test *testing.T) {
	test.Parallel()
	// Each probe stays not yet in .ts. In .a, the write and the view a declared property's type could
	// lie through are refused instead (refusedInA, TestNamedRecordLiesAreRefusedInA). The view drops a
	// property held as a union, whose read comes back unchecked: one whose read is checked lowers in .ts
	// (#6c4swvt G3a, internal/oracle/testdata/records_options_base).
	for _, probe := range []struct {
		source, reason string
		refusedInA     bool
	}{
		{
			`interface R {named?:number[];[key:string]:number[]|string|undefined} function read(r:R):number|undefined {return r.named?.length;}`,
			"dictionary member narrowed to an array without an element-checked view",
			false,
		},
		{
			`interface R {named?:number;[key:string]:number|string|undefined} function write(r:R,k:string):void {r[k]="wrong";}`,
			"named property named",
			true,
		},
		{
			`interface R {named?:number|string;[key:string]:number|string|boolean|undefined} const r:R={}; const view:Record<string,number|string|boolean|undefined>=r;`,
			"seen as",
			true,
		},
		{
			`interface R {readonly named:number;[key:string]:number} const r:R={named:1}; const view:Record<string,number>=r;`,
			"seen as",
			false,
		},
		{`interface R {named:number;[key:string]:number} const fixed={named:1}; const r:R=fixed;`, "seen as", false},
		{`interface R {named?:number;readonly [key:string]:number|undefined} const r:R={};`, "signature", false},
		{`interface R {named:number;[key:number]:number} const r:R={named:1};`, "signature", false},
	} {
		for _, extension := range []string{".a", ".ts"} {
			if extension == ".a" && probe.refusedInA {
				continue
			}
			test.Run(extension+" "+probe.source, func(test *testing.T) {
				_, lowerSourceAsError := lowerSourceAs(test, probe.source, extension)
				var ny *NotYet
				if !errors.As(lowerSourceAsError, &ny) || !strings.Contains(lowerSourceAsError.Error(), probe.reason) {
					test.Fatalf("got %v, want NotYet %s", lowerSourceAsError, probe.reason)
				}
			})
		}
	}
}

// In .a, a write by a key that isn't one literal, a view through a wider index signature, and a type
// argument whose constraint is such a view, either way, are refused where a declared property's type
// could lie (#npqr3sx).
func TestNamedRecordLiesAreRefusedInA(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ source, reason string }{
		{
			`interface R {named?:number;[key:string]:number|string|undefined} function write(r:R,k:string):void {r[k]="wrong";}`,
			"a write into a record with declared properties through a key that is not one literal",
		},
		{
			`interface R {named?:number;[key:string]:number|string|undefined} const r:R={}; const view:Record<string,number|string|undefined>=r;`,
			"a value with declared properties seen through a wider index signature",
		},
		{
			`interface R {named?:number;[key:string]:number|string|undefined} type D={[key:string]:number|string|undefined}; function put<T extends D>(t:T,k:string):void {const w:D=t; w[k]="wrong";} const r:R={}; put(r,"named");`,
			"a type argument with declared properties bound to a type parameter constrained to a wider index signature",
		},
		{
			`interface R {named?:number;[key:string]:number|string|undefined} type D={[key:string]:number|string|undefined}; function read<T extends R>(t:T):number|undefined {return t.named;} const d:D={}; read(d);`,
			"a dictionary bound as a type argument whose constraint reads it through a declared property narrower than its values",
		},
	} {
		test.Run(probe.source, func(test *testing.T) {
			_, lowerSourceError := lowerSource(test, probe.source)
			var refused *Refused
			if !errors.As(lowerSourceError, &refused) || !strings.HasPrefix(refused.What, probe.reason) {
				test.Fatalf("got %v, want refused %s", lowerSourceError, probe.reason)
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
// Each row stops so in .a and .ts alike: in .a the declared-key write judges the value by the same
// registry first (records_declared_keys.go). Mutants: slotType ignoring the slot a caller set lowers
// w8, which then prints POINT natively; recordComputedEntry skipped lowers w8's literal and b10 in .ts;
// refuseWidening's rules skipped in refuseInSlot's run lowers the invariance and enum pins.
func TestNamedRecordWritesAreJudgedAsTheirDirectTwins(test *testing.T) {
	test.Parallel()
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
		return strings.Replace(
			strings.Replace(source, "[k: string]: Point | Shape }", "[k: string]: Point | Shape | number }", 1),
			"[k: string]: () => void }",
			"[k: string]: (() => void) | number }",
			1,
		)
	}
	for _, probe := range []struct {
		name, source, write, direct string
		byRule                      bool
	}{
		{
			"w8",
			w8,
			"const places: Places = { names: new Point() };\nplaces[key] = new Other();",
			"const places: Places = { names: new Point() };\nplaces.names = new Other();",
			false,
		},
		{
			"w10",
			w10,
			"const handlers: Handlers = { names: first };\nhandlers[key] = quiet;",
			"const handlers: Handlers = { names: first };\nhandlers.names = quiet;",
			false,
		},
		// The list's widened and enumAssignable went with it, since the registry judges both at the direct
		// write (Lower's and Compiler's pins on 82ae5693). The index type is readonly Animal[], which can't write,
		// so the site's own view of the index stops nothing, and isn't boxed as a union, which recordComputedEntry
		// leaves to its reads; only pets can stop it: a Dog[] under pets: Animal[] could be written
		// an Animal the dogs are read as (adamic/invariant-mutable), and a number under kind: Kind.Text
		// breaks the member's literal (#w7w2asn's slot refusal). Each through a key and through a literal.
		{
			"invariance",
			"class Animal { name = \"a\"; }\nclass Dog extends Animal { bark(): string { return \"woof\"; } }\ntype Kennels = { pets: Animal[]; [k: string]: readonly Animal[] };\nconst keys: string[] = [\"pets\", \"other\"];\nconst key = keys[0] ?? \"other\";\nconst dogs: Dog[] = [];\n",
			"const kennels: Kennels = { pets: [] };\nkennels[key] = dogs;",
			"const kennels: Kennels = { pets: [] };\nkennels.pets = dogs;",
			false,
		},
		{
			"invariance construction",
			"class Animal { name = \"a\"; }\nclass Dog extends Animal { bark(): string { return \"woof\"; } }\ntype Kennels = { pets: Animal[]; [k: string]: readonly Animal[] };\nconst keys: string[] = [\"pets\", \"other\"];\nconst key = keys[0] ?? \"other\";\nconst dogs: Dog[] = [];\n",
			"const kennels: Kennels = { pets: [new Animal()], [key]: dogs };",
			"const kennels: Kennels = { pets: dogs };",
			true,
		},
		{
			"an open enum's member",
			"enum Kind { Text, Count }\ntype Nodes = { kind: Kind.Text; [k: string]: number };\nconst keys: string[] = [\"kind\", \"other\"];\nconst key = keys[0] ?? \"other\";\nconst n: number = 7;\n",
			"const nodes: Nodes = { kind: Kind.Text };\nnodes[key] = n;",
			"const nodes: Nodes = { kind: Kind.Text };\nnodes.kind = n;",
			false,
		},
		{
			"an open enum's member construction",
			"enum Kind { Text, Count }\ntype Nodes = { kind: Kind.Text; [k: string]: number };\nconst keys: string[] = [\"kind\", \"other\"];\nconst key = keys[0] ?? \"other\";\nconst n: number = 7;\n",
			"const nodes: Nodes = { kind: Kind.Text, [key]: n };",
			"const nodes: Nodes = { kind: n };",
			true,
		},
		// A host object under a declared property: lowering's host-view check sees the index type the value
		// is stored as (Clock, whose now is registered), so the declared property's view (View, which reaches
		// the unregistered eventLoopUtilization) has to be judged by the write itself (Lower's probe).
		{
			"a node host object",
			"import { performance } from \"node:perf_hooks\";\ntype Clock = { now(): number };\ntype View = { now(): number; eventLoopUtilization(): { readonly idle: number } };\ntype Places = { names: View; [k: string]: Clock };\nconst keys: string[] = [\"names\"];\nconst key = keys[0] ?? \"names\";\nconst initial: View = { now: () => 1, eventLoopUtilization: () => ({ idle: 1 }) };\nconst places: Places = { names: initial };\n",
			"places[key] = performance;\nconsole.log(String(places.names.eventLoopUtilization().idle >= 0));",
			"places.names = performance;\nconsole.log(String(places.names.eventLoopUtilization().idle >= 0));",
			false,
		},
		{
			"w8 construction",
			w8,
			"const places: Places = { names: new Point(), [key]: new Other() };",
			"const places: Places = { names: new Other() };",
			true,
		},
		{
			"w10 construction",
			w10,
			"const handlers: Handlers = { names: first, [key]: quiet };",
			"const handlers: Handlers = { names: quiet };",
			true,
		},
		{
			"w12 alias receiver",
			w8,
			"const places: Places = { names: new Point() };\nconst both: Places[] = [places];\nconst slot = both[0];\nif (slot !== undefined) slot[key] = new Other();",
			"const places: Places = { names: new Point() };\nconst both: Places[] = [places];\nconst slot = both[0];\nif (slot !== undefined) slot.names = new Other();",
			false,
		},
		{
			"w13 chained",
			w8,
			"const places: Places = { names: new Point() };\nlet spare: Point | Shape = new Point();\nspare = places[key] = new Other();",
			"const places: Places = { names: new Point() };\nlet spare: Point | Shape = new Point();\nspare = places.names = new Other();",
			false,
		},
		{
			"w14 parameter",
			w8,
			"const places: Places = { names: new Point() };\nconst write = (target: Places, at: string): void => { target[at] = new Other(); };\nwrite(places, key);",
			"const places: Places = { names: new Point() };\nconst write = (target: Places, at: string): void => { target.names = new Other(); };\nwrite(places, key);",
			false,
		},
		{
			"w10 construction boxed",
			boxed(w10),
			"const handlers: Handlers = { names: first, [key]: quiet };",
			"const handlers: Handlers = { names: quiet };",
			true,
		},
	} {
		for _, extension := range []string{".a", ".ts"} {
			test.Run(probe.name+" "+extension, func(test *testing.T) {
				test.Parallel()
				_, written := lowerSourceAs(test, probe.source+probe.write+"\nconsole.log('ran');\n", extension)
				_, direct := lowerSourceAs(test, probe.source+probe.direct+"\nconsole.log('ran');\n", extension)
				if direct == nil {
					test.Fatalf("the direct twin %q lowers, so it proves nothing about the write", probe.direct)
				}
				same := written != nil && stoppedFor(written) == stoppedFor(direct)
				if probe.byRule {
					same = written != nil && ruleOf(written) == ruleOf(direct)
				}
				if !same {
					test.Fatalf("the write stopped as %v, want what its direct twin stops as: %v", written, direct)
				}
			})
		}
	}
	// The reads that hold readGuardedInUnionSlot, in .ts: a lie under a declared plain object (b8, w8's
	// literal boxed), array (a8), Map (p8) or dictionary (r8) in a boxed slot is left to its read, which
	// stops as an object kind with no checked tag, or an array with no element-checked view (#y852c3c). Each goes red
	// if the skip stops naming its kind (the write's stop comes first) or the read stops stopping. A computed method is
	// no data entry, a stop of its own. In .a the declared-key write comes first and refuses each by the fix inA names:
	// b8 by its direct twin's nominal rule, the others as a value names can't take, and the computed method,
	// whose function is no one expression's value, is not yet.
	for _, probe := range []struct{ name, source, reason, inA string }{
		{
			"b8 object construction boxed",
			boxed(w8) +
				"const places: Places = { names: new Point(), [key]: new Other() };\nconsole.log(places.names.describe().toUpperCase());\n",
			"narrowed to an object kind without a checked runtime tag",
			"Refused: construct that class or a subclass; use an interface for structural values (adamic/nominal-class)",
		},
		{"a8 array construction boxed", `type Lists = { names: string[]; [k: string]: string[] | number[] | number };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
const lists: Lists = { names: ["a"], [key]: [1, 2] };
console.log((lists.names[0] ?? "").toUpperCase());
`, "narrowed to an array without an element-checked view", "Refused: " + declaredKeysFix},
		{
			"p8 map construction boxed",
			`type Tables = { names: Map<string, string>; [k: string]: Map<string, string> | Map<string, number> | number };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
const tables: Tables = { names: new Map<string, string>([["a", "b"]]), [key]: new Map<string, number>([["a", 1]]) };
console.log((tables.names.get("a") ?? "").toUpperCase());
`,
			"narrowed to an object kind without a checked runtime tag",
			"Refused: " + declaredKeysFix,
		},
		{"r8 dictionary construction boxed", `interface MapLike<T> { [index: string]: T; }
type Tables = { names: MapLike<string>; [k: string]: MapLike<string> | MapLike<number> | number };
const keys: string[] = ["names", "other"];
const key = keys[0] ?? "other";
const tables: Tables = { names: { a: "b" }, [key]: { a: 1 } };
console.log((tables.names["a"] ?? "").toUpperCase());
`, "narrowed to an object kind without a checked runtime tag", "Refused: " + declaredKeysFix},
		{
			"w10 computed method",
			w10 + "const handlers: Handlers = { names: first, [key]() { return \"x\"; } };\nconsole.log('ran');\n",
			"a record literal member other than a data entry",
			"NotYet: a write into a record with declared properties through a key that is not one literal, of no one expression's value, whose type isn't declared property names's type itself",
		},
	} {
		test.Run(probe.name, func(test *testing.T) {
			test.Parallel()
			_, lowerSourceAsError := lowerSourceAs(test, probe.source, ".ts")
			var ny *NotYet
			if !errors.As(lowerSourceAsError, &ny) || !strings.Contains(lowerSourceAsError.Error(), probe.reason) {
				test.Fatalf("got %v, want the NotYet %s", lowerSourceAsError, probe.reason)
			}
			if _, lowerSourceError := lowerSource(
				test,
				probe.source,
			); lowerSourceError == nil || ruleOf(lowerSourceError) != probe.inA {
				test.Fatalf("in .a: got %v, want %s", lowerSourceError, probe.inA)
			}
		})
	}
	for _, extension := range []string{".a", ".ts"} {
		if _, lowerSourceAsError := lowerSourceAs(
			test,
			`interface Options { target?: number; required: number; [option: string]: number | undefined }
function write(options: Options, key: string): void { options[key] = 9; }`,
			extension,
		); lowerSourceAsError != nil {
			test.Fatalf("an honest write in %s: %v", extension, lowerSourceAsError)
		}
	}
	// A computed name strings joined by + make is judged as that name: tsc's ["" + "constructor"] beside
	// the keywords its literal declares names none of them (01_scanner_keywords.a's shape).
	if _, lowerSourceError := lowerSource(
		test,
		"const SyntaxKind = { AbstractKeyword: 128, AccessorKeyword: 129, ConstructorKeyword: 137 } as const;\ntype KeywordSyntaxKind = (typeof SyntaxKind)[keyof typeof SyntaxKind];\ninterface MapLike<T> { [index: string]: T; }\nconst kinds: MapLike<KeywordSyntaxKind> = {\n    abstract: SyntaxKind.AbstractKeyword,\n    accessor: SyntaxKind.AccessorKeyword,\n    [\"\" + \"constructor\"]: SyntaxKind.ConstructorKeyword,\n};\nconsole.log(`${String(kinds[\"abstract\"])} ${Object.keys(kinds).join(\",\")}`);\n",
	); lowerSourceError != nil {
		test.Fatalf("a folded name no named property takes: %v", lowerSourceError)
	}
}

// ruleOf is whether an error is a refusal or a NotYet, and the fix a refusal names, which names its rule.
func ruleOf(sourceError error) string {
	var refused *Refused
	if errors.As(sourceError, &refused) {
		return "Refused: " + refused.Fix
	}
	return stoppedFor(sourceError)
}
