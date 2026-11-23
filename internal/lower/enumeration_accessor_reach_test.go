package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Family 17, Adamic's ruling of Nov 22 (#1atvfp5): an enumeration is refused beside a getter or setter
// only when the accessor's object can reach the enumerated operand by the type graph
// (enumeration_accessor_reach.go). Each program keeps its accessors in held.a, beside main.a: rule 700
// (enumerationDescriptorRefusal) refuses a getter in a module that itself calls Object.values or
// Object.entries, in its own words, so a getter in main.a would stop there before this rule is asked.
//
// Every row reaches the guard, enumerationDescriptors, through a road that calls it: Object.values and
// Object.entries ask it in objectCallArguments (library_object.go:100) before any check of the
// operand's shape or value type, a checked one asks again (library_object.go:197), and encodeJson asks
// it once its schema is proven (library_json_encode.go:39). Object.keys mostly doesn't: objectKeys
// (class_features.go) lowers a class instance, a static side and a plain view straight to ir.ObjectKeys
// and reaches objectCall only for a node array, a RegExp, a possibly undefined operand or a shape with
// an optional field, so only the unrelated row's optional field enumerates with it.

// enumerationReachRefusal is the words an accessor that reaches the operand stops with, as before.
const enumerationReachRefusal = "Object enumeration with getter or accessor descriptors; use own data properties for this first reflection implementation"

// enumerationReachTemperature is a getter nothing enumerated can hold: Temperature is assignable to no
// type any enumeration below reads, and none of them is assignable to it.
const enumerationReachTemperature = "export class Temperature { kelvin = 300; get celsius(): number { return this.kelvin - 273; } }\n"

// enumerationReachSensor is a class whose getter implements an interface's readonly field.
const enumerationReachSensor = "export interface Reading { readonly celsius: number }\nexport class Sensor implements Reading { get celsius(): number { return 21; } }\n"

// enumerationReachGauge is a class whose getter implements an interface's readonly field, read under
// encodeJson's operand.
const enumerationReachGauge = "export interface Inner { readonly value: number }\nexport class Gauge implements Inner { get value(): number { return 3; } }\n"

// enumerationReachLower writes held and main, each named by extension, to one directory and lowers
// main, returning its path.
func enumerationReachLower(t *testing.T, extension, held, main string) (string, *ir.Program, error) {
	t.Helper()
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "held"+extension), []byte(held), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "main"+extension)
	if err := os.WriteFile(path, []byte(main), 0o644); err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	program, err := Lower(context.Background(), checked)
	return path, program, err
}

// enumerationReachAgrees lowers a row that must lower, and holds it to Node on the JavaScript backend
// and to stdout natively.
func enumerationReachAgrees(t *testing.T, name, held, main, stdout string) {
	t.Helper()
	path, program, err := enumerationReachLower(t, ".a", held, main)
	if err != nil {
		t.Fatalf("enumeration reach row %s was refused: %v", name, err)
	}
	agreeEntry(t, path, program, 0, nil, nil)
	if code, out, errs := brandedNative(t, program); code != 0 || out != stdout {
		t.Fatalf("enumeration reach row %s natively: exit=%d stdout=%q stderr=%q, want %q", name, code, out, errs, stdout)
	}
}

// enumerationReachRefused holds a row, held.a beside main.a, to the accessor refusal's words.
func enumerationReachRefused(t *testing.T, name, held, main string) {
	t.Helper()
	enumerationReachRefusedAs(t, name, ".a", held, main)
}

// enumerationReachRefusedAs holds a row in files of the given extension to the accessor refusal's words.
func enumerationReachRefusedAs(t *testing.T, name, extension, held, main string) {
	t.Helper()
	_, _, err := enumerationReachLower(t, extension, held, main)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), enumerationReachRefusal) {
		t.Fatalf("enumeration reach row %s was admitted, want %q, got %v", name, enumerationReachRefusal, err)
	}
}

// Object.values and Object.entries of a const literal, Object.values of a parameter (the checked
// enumeration, which asks the rule a second time), encodeJson, and Object.keys of a shape with an
// optional field (the fallback road objectKeys hands to objectCall), beside a class whose getter no
// enumerated value can hold: each lowers and agrees with Node, where the program-wide scan refused all
// five. Mutant WholeGraph (every accessor reaches) turns it red.
func TestEnumerationReachUnrelatedGetterLowers(t *testing.T) {
	t.Parallel()
	const main = `import { encodeJson } from 'adamic';
import { Temperature } from './held.a';
interface Point { x: number; y: number }
const point = { x: 1, y: 2 };
function total(shape: { readonly x: number; readonly y: number }): number {
  let sum = 0;
  for (const value of Object.values(shape)) {
    sum += value;
  }
  return sum;
}
console.log(Object.values(point).join(','));
for (const [key, value] of Object.entries(point)) console.log(` + "`${key}:${value}`" + `);
console.log(String(total(point)));
const reading: Point = { x: 3, y: 4 };
console.log(encodeJson<Point>(reading));
const partial: { x: number; y?: number } = { x: 5 };
console.log(Object.keys(partial).join(','));
console.log(String(new Temperature().celsius));
`
	enumerationReachAgrees(t, "unrelated", enumerationReachTemperature, main, "1,2\nx:1\ny:2\n3\n{\"x\":3,\"y\":4}\nx\n27\n")
}

// Object.values of a parameter typed as a union of two object shapes with number values, beside the
// unrelated getter: each member is classified and read, so the union is known and lowers. Mutant
// UnionMembers (a union's members are left unread) makes the union a type that can't be classified,
// which fails closed to the program-wide scan, and turns it red.
func TestEnumerationReachUnionLowers(t *testing.T) {
	t.Parallel()
	const main = `import { Temperature } from './held.a';
type Shape = { width: number; height: number } | { radius: number };
function sizes(shape: Shape): string {
  return Object.values(shape).join(',');
}
console.log(sizes({ width: 2, height: 3 }));
console.log(sizes({ radius: 1 }));
console.log(String(new Temperature().celsius));
`
	enumerationReachAgrees(t, "union", enumerationReachTemperature, main, "2,3\n1\n27\n")
}

// Each row's operand can hold a getter-bearing value, so each stays refused in today's words: an object
// literal's getter seen through a type it's assignable to, a class instance (the owner found is
// Sensor's instance side), the interface a getter's class implements (by Object.entries), a parameter
// typed as a union holding that interface, a class's static side (the operand is the constructor
// itself, and the owner found is typeof Sensor, whose static getter Object.values reads), a getter a
// subclass inherits (the owners found are Base and Sub, whose property is Base's getter; the base isn't
// assignable to the subclass, its instances are), a parameter typed as the interface, a parameter typed
// as a discriminated union a class is assignable to as a whole, by its discriminant, though it relates
// to neither member either way (Lower's witness, so the union is read as itself as well as by its
// members; mutant DiscriminatedUnion turns it red), a getter of a generic class, whose owner taken at
// T = any reaches a { readonly value: number } view (mutant AnyInstance turns it red), and a literal
// annotated Holder<T>, whose owner stays unknown.
//
// tsc types Object.values of a class instance, a constructor or a union with an interface member as
// any[], since only a type literal has an implicit index signature, and the call read as join's
// receiver stops at "an array of any" (elementType, object.go) before it's lowered. Those rows return
// the call from a function typed number[] instead, as the parameter row does, so the call is lowered
// with a contextual array type and asks the guard at library_object.go:100, the operand still typed as
// the class, the constructor or the union.
func TestEnumerationReachReachingGetterRefused(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, held, main string }{
		{"literal", "export const literal = { get celsius(): number { return 21; } };\n", "import { literal } from './held.a';\nconst reading: { readonly celsius: number } = literal;\nconsole.log(Object.values(reading).join(','));\n"},
		{"class", enumerationReachSensor, "import { Sensor } from './held.a';\nfunction read(sensor: Sensor): number[] { return Object.values(sensor); }\nconsole.log(read(new Sensor()).join(','));\n"},
		{"interface", enumerationReachSensor, "import { type Reading, Sensor } from './held.a';\nconst reading: Reading = new Sensor();\nconsole.log(String(Object.entries(reading).length));\n"},
		{"union", enumerationReachSensor, "import { type Reading, Sensor } from './held.a';\ntype Either = { readonly label: number } | Reading;\nfunction read(either: Either): number[] { return Object.values(either); }\nconsole.log(read(new Sensor()).join(','));\n"},
		{"static", "export class Sensor { static get celsius(): number { return 21; } }\n", "import { Sensor } from './held.a';\nfunction read(): number[] { return Object.values(Sensor); }\nconsole.log(read().join(','));\n"},
		{"subclass", "export class Base { get celsius(): number { return 21; } }\nexport class Sub extends Base { count = 1; }\n", "import { Sub } from './held.a';\nfunction read(sub: Sub): number[] { return Object.values(sub); }\nconsole.log(read(new Sub()).join(','));\n"},
		{"parameter", enumerationReachSensor, "import { type Reading, Sensor } from './held.a';\nfunction read(reading: Reading): number[] { return Object.values(reading); }\nconsole.log(read(new Sensor()).join(','));\n"},
		{"discriminated", "export class Coded { kind: 1 | 2 = 1; x = 1; get g(): number { return 2; } }\n", "import { Coded } from './held.a';\ntype Tagged = { kind: 1; g: number } | { kind: 2; g: number };\nfunction read(tagged: Tagged): number[] { return Object.values(tagged); }\nconsole.log(read(new Coded()).join(','));\n"},
		{"held", "export interface Holder<T> { readonly held: T }\nfunction make<T>(value: T): Holder<T> { return { get held(): T { return value; } }; }\nexport function hold<T>(value: T): T { const box: Holder<T> = make(value); return box.held; }\n", "import { hold } from './held.a';\nconst point = { x: 1, y: 2 };\nconsole.log(Object.values(point).join(','));\nconsole.log(String(hold<number>(3)));\n"},
		{"generic", "export class Box<T> { held: T; constructor(held: T) { this.held = held; } get value(): T { return this.held; } }\n", "import { Box } from './held.a';\nconst reading: { readonly value: number } = { value: 1 };\nconsole.log(Object.values(reading).join(','));\nconsole.log(String(new Box<number>(2).value));\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			enumerationReachRefused(t, row.name, row.held, row.main)
		})
	}
}

// A getter-bearing object literal built inside a generic function, whose own type mentions no type
// parameter, has one owner for every instantiation, as tsc's createTransformationContext literal inside
// transformNodes<T> does: beside Object.values of an unrelated literal, it lowers and agrees with Node.
// Mutant GenericLiteral (any enclosing type parameters make the literal's owner unknown again) turns it
// red. The literal whose type does mention T is the refused row held.
func TestEnumerationReachGenericLiteralLowers(t *testing.T) {
	t.Parallel()
	const held = "export function doubled<T>(value: T): number { const box = { count: 4, get twice(): number { return 8; } }; return value === undefined ? box.count : box.twice; }\n"
	const main = "import { doubled } from './held.a';\nconst point = { x: 1, y: 2 };\nconsole.log(Object.values(point).join(','));\nconsole.log(String(doubled<string>('s')));\n"
	enumerationReachAgrees(t, "generic literal", held, main, "1,2\n8\n")
}

// A generic class whose getter no enumerated value can hold, read at its type parameter taken at any:
// tsc's IdentifierNameMap<V> shape, beside Object.values of an unrelated literal, lowers and agrees with
// Node. Mutant GenericClass (a class's own type parameters make its owner unknown again) turns it red.
// The generic row, a Box<T> whose getter does reach a { readonly value: number } view at T = any, stays
// refused, and mutant AnyInstance (the class read uninstantiated, where T relates to nothing) turns that red.
func TestEnumerationReachGenericClassLowers(t *testing.T) {
	t.Parallel()
	const held = "export class Tally<V> { private count = 0; add(value: V): number { this.count += 1; return this.count; } get size(): number { return this.count; } }\n"
	const main = "import { Tally } from './held.a';\nconst point = { x: 1, y: 2 };\nconsole.log(Object.values(point).join(','));\nconst tally = new Tally<string>();\nconsole.log(String(tally.add('a')));\nconsole.log(String(tally.size));\n"
	enumerationReachAgrees(t, "generic class", held, main, "1,2\n1\n1\n")
}

// Getter-bearing literals inside generic functions whose own types mention T, owned by their contextual
// types instead: one returned from a function declared to return Cursor (tsc's decodeMappings shape,
// returned as a MappingsDecoder), one passed to a parameter typed Cursor. Beside Object.values of an
// unrelated literal each lowers and agrees with Node. Mutant Contextual (the contextual type is never
// taken) turns it red; mutant ContextualScope (a contextual type that mentions T is taken anyway) turns
// the held row, a literal returned as Holder<T>, red.
func TestEnumerationReachContextualLiteralLowers(t *testing.T) {
	t.Parallel()
	const held = "export interface Cursor { tag: string; readonly pos: number }\nfunction measure(cursor: Cursor): number { return cursor.pos; }\nexport function open<T extends string>(tag: T): Cursor { return { tag, get pos(): number { return 1; } }; }\nexport function probe<T extends string>(tag: T): number { return measure({ tag, get pos(): number { return 2; } }); }\n"
	const main = "import { open, probe } from './held.a';\nconst point = { x: 1, y: 2 };\nconsole.log(Object.values(point).join(','));\nconsole.log(String(open('a').pos));\nconsole.log(String(probe('b')));\n"
	enumerationReachAgrees(t, "contextual literal", held, main, "1,2\n1\n2\n")
}

// A getter-bearing literal inside a generic function whose own type mentions T, through a field holding
// the function's value (T extends string, so the field has a native slot), which initializes a variable
// declared as a non-generic interface: the literal is reached only through that variable, so the
// interface is its owner, and beside Object.values of an unrelated literal it lowers and agrees with
// Node. That's tsc's createTransformationContext shape (a literal typed TransformationContext inside
// transformNodes<T>). Counter also holds a chain nested past eachReached's depth, as NodeFactory does,
// so the contextual road's walk can't clear it and the annotation alone does: mutant DeclaredOwner (the
// declared type is never taken) turns it red. The held row, whose annotation is Holder<T>, stays refused.
func TestEnumerationReachDeclaredLiteralLowers(t *testing.T) {
	t.Parallel()
	var held strings.Builder
	for level := range enumerationReachLevels - 1 {
		held.WriteString("export interface Link" + strconv.Itoa(level) + " { next: Link" + strconv.Itoa(level+1) + " }\n")
	}
	held.WriteString("export interface Link" + strconv.Itoa(enumerationReachLevels-1) + " { value: number }\n")
	held.WriteString("const chain: Link0 = " + strings.Repeat("{ next: ", enumerationReachLevels-1) + "{ value: 1 }" + strings.Repeat(" }", enumerationReachLevels-1) + ";\n")
	held.WriteString("export interface Counter { count: number; readonly twice: number; tag: string; link: Link0 }\n")
	held.WriteString("export function doubled<T extends string>(value: T): number { const box: Counter = { count: 4, get twice(): number { return 8; }, tag: value, link: chain }; return box.twice; }\n")
	const main = "import { doubled } from './held.a';\nconst point = { x: 1, y: 2 };\nconsole.log(Object.values(point).join(','));\nconsole.log(String(doubled<string>('s')));\n"
	enumerationReachAgrees(t, "declared literal", held.String(), main, "1,2\n8\n")
}

// encodeJson reads every field at any depth, so a getter-bearing value under the operand reaches it: a
// field typed as the interface a getter's class implements, and an array of it. Mutant DeepReach (encodeJson
// reads only the operand's own type) turns it red.
func TestEnumerationReachDeepGetterRefused(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, main string }{
		{"field", "import { encodeJson } from 'adamic';\nimport { type Inner, Gauge } from './held.a';\ninterface Outer { inner: Inner; label: string }\nconst outer: Outer = { inner: new Gauge(), label: 'g' };\nconsole.log(encodeJson<Outer>(outer));\n"},
		{"element", "import { encodeJson } from 'adamic';\nimport { type Inner, Gauge } from './held.a';\ninterface Outer { items: Inner[] }\nconst outer: Outer = { items: [new Gauge()] };\nconsole.log(encodeJson<Outer>(outer));\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			enumerationReachRefused(t, row.name, enumerationReachGauge, row.main)
		})
	}
}

// encodeJson of a union whose one member alone has a field typed as the interface a getter's class
// implements: the union's own properties are only what every member shares, so the field is reached
// through the member, and stays refused.
func TestEnumerationReachDeepUnionGetterRefused(t *testing.T) {
	t.Parallel()
	enumerationReachRefused(t, "deep union", enumerationReachGauge, "import { encodeJson } from 'adamic';\nimport { type Inner, Gauge } from './held.a';\ntype Shape = { kind: 'plain'; size: number } | { kind: 'held'; inner: Inner };\nconst shape: Shape = { kind: 'held', inner: new Gauge() };\nconsole.log(encodeJson<Shape>(shape));\n")
}

// enumerationReachLevels is how many interfaces deep the chain row nests, one past enumerationReachDepth's
// last level read.
const enumerationReachLevels = enumerationReachDepth + 2

// enumerationReachChain is encodeJson of a value whose interfaces nest enumerationReachLevels deep, beside
// the unrelated getter: Level0's next is Level1, and so on to a last level holding a number.
func enumerationReachChain() string {
	var source strings.Builder
	source.WriteString("import { encodeJson } from 'adamic';\nimport { Temperature } from './held.a';\n")
	for level := range enumerationReachLevels - 1 {
		source.WriteString("interface Level" + strconv.Itoa(level) + " { next: Level" + strconv.Itoa(level+1) + " }\n")
	}
	source.WriteString("interface Level" + strconv.Itoa(enumerationReachLevels-1) + " { value: number }\n")
	source.WriteString("const chain: Level0 = " + strings.Repeat("{ next: ", enumerationReachLevels-1) + "{ value: 1 }" + strings.Repeat(" }", enumerationReachLevels-1) + ";\n")
	source.WriteString("console.log(encodeJson<Level0>(chain));\nconsole.log(String(new Temperature().celsius));\n")
	return source.String()
}

// An operand whose reach can't be told keeps the program-wide scan of before, so the unrelated getter
// still refuses it: a parameter typed object (a non-primitive type), one typed {} (any object at all),
// one typed as a readonly string index whose values are objects (the program doesn't hold it as a
// record, so recordElement is nil and the call isn't taken by the record road), each by Object.values,
// which asks the guard before its value type is checked; encodeJson of an empty interface's literal;
// and encodeJson of a chain nested past enumerationReachDepth.
//
// The other causes enumerationReadTypes names are fail-closed and unreachable today, each stopped
// first elsewhere. Any: a parameter typed any is refused as "a value of type any" (refuseAnyValue,
// never_refused.go). A type parameter: the guard is asked, but a generic body is lowered per
// instantiation and l.concrete maps T to its type argument first, so a T extends object operand reads
// as { a: number } and the call stops at Object.values' homogeneity check (library_object.go:168). An
// index signature under encodeJson: jsonDecodeSchema refuses every index signature but an array's or a
// tuple's as "cannot prove ... is JSON data" before the guard. A mutable string index under
// Object.values or Object.entries: recordElement is its value type, so recordObjectCallArguments
// (library_object.go:24) lowers it as a record call without asking the guard, as at main.
func TestEnumerationReachUnknownKeepsProgramScan(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, main string }{
		{"object", "import { Temperature } from './held.a';\nfunction count(value: object): number { return Object.values(value).length; }\nconsole.log(String(count({ a: 1 })));\nconsole.log(String(new Temperature().celsius));\n"},
		{"empty", "import { Temperature } from './held.a';\nfunction count(value: {}): number { return Object.values(value).length; }\nconsole.log(String(count({ a: 1 })));\nconsole.log(String(new Temperature().celsius));\n"},
		{"index", "import { Temperature } from './held.a';\nfunction count(bag: { readonly [key: string]: { label: string } }): number { return Object.values(bag).length; }\nconsole.log(String(count({ a: { label: 'x' } })));\nconsole.log(String(new Temperature().celsius));\n"},
		{"empty interface", "import { encodeJson } from 'adamic';\nimport { Temperature } from './held.a';\ninterface Empty {}\nconsole.log(encodeJson<Empty>({}));\nconsole.log(String(new Temperature().celsius));\n"},
		{"depth", enumerationReachChain()},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			enumerationReachRefused(t, row.name, enumerationReachTemperature, row.main)
		})
	}
}

// A checked downcast in .ts keeps the same object at a slot its owner type isn't assignable to: a class
// instance with a getter, seen as a narrower shape and cast back down to a wider one that is assignable
// to the class but that the class isn't assignable to. The read type is assignable to the owner, so the
// accessor reaches, and Object.values stays refused at the guard. The class is in held.ts, since rule 700
// refuses a getter in the module that calls Object.values. Mutant OneWay (owner to read only) turns it
// red.
func TestEnumerationReachDowncastRefused(t *testing.T) {
	t.Parallel()
	enumerationReachRefusedAs(t, "downcast", ".ts", "export class Meter { a = 1; get g(): number { return 2; } }\n", "import { Meter } from './held';\nconst narrow: { a: number } = new Meter();\nconst wide = narrow as { a: number; readonly g: number; c: number };\nconsole.log(Object.values(wide).join(','));\n")
}

// enumerationReachShared is a getter-bearing literal held as Shared, a type it's assignable to, which
// main.ts casts down to a view related to the literal neither way.
const enumerationReachShared = "type Shared = { b: number };\nconst lit = { b: 2, d: 4, get a(): number { return 1; } };\nexport const shared: Shared = lit;\n"

// A checked view checks at a read, never at its cast, so a .ts downcast puts the literal, getter and
// all, at a view whose type relates to the literal's neither way, where the guard admits Object.values
// of it: Node printed 2,1,4 and native exited 70 on a field the checker proved there (J-m9g0n72).
// checkEnumerationReachViews widens the read types by each view's source once lowering is done, so each
// row stops in the guard's words: the sibling (Other declares d: 4 where the literal's d is a number),
// Language's sibling (Other declares a field c the literal lacks), and a chain of two casts, Shared to
// Middle to Other, where neither Middle nor Other relates to the literal and only the second step of
// the widening reaches Shared. The control with no cast in the program is the unrelated row, which
// lowers and agrees with Node. Mutant ViewPass (the post-pass returns at its top) turns the sibling red.
// A getter-bearing literal written under as, owned by the asserted type it's reached through (Lower's
// should on 5eacc825 asked whether that owner can miss): enumerated through the asserted type in another
// module, it stays refused. It's refused by the asserted owner directly and, were the literal's own type
// the owner instead, by the post-pass following the assertion's view, so no mutant of the owner choice can
// turn it red (J-rr89gsv); it pins the behavior, and an asserted downcast to an incomparable type is
// TS2352, never a program (J-8w9vc55). The upcast rows are Lower's: an asserted type that hides the getter
// ({ b, get a } as { b: number }), enumerated by Object.values and Object.entries in another module, stays
// refused at the head as at main (J-y0fk1wv, J-bb91pr3).
func TestEnumerationReachAssertedLiteralRefused(t *testing.T) {
	t.Parallel()
	const upcast = "export const asserted = { b: 1, get a(): number { return 2; } } as { b: number };\n"
	for _, row := range []struct{ name, held, main string }{
		{"asserted literal", "export type View = { b: number; d: 4; e?: number };\nexport const asserted = { b: 1, d: 4, get a(): number { return 1; } } as View;\n", "import { type View, asserted } from './held';\nfunction read(view: View): number[] { return Object.values(view); }\nconsole.log(read(asserted).join(','));\n"},
		{"upcast values", upcast, "import { asserted } from './held';\nfunction read(view: { b: number }): number[] { return Object.values(view); }\nconsole.log(read(asserted).join(','));\n"},
		{"upcast entries", upcast, "import { asserted } from './held';\nfunction show(view: { b: number }): void { for (const [key, value] of Object.entries(view)) console.log(key + '=' + String(value)); }\nshow(asserted);\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			enumerationReachRefusedAs(t, row.name, ".ts", row.held, row.main)
		})
	}
}

func TestEnumerationReachViewRefused(t *testing.T) {
	t.Parallel()
	const read = "function read(view: Other): number[] { return Object.values(view); }\nconsole.log(read(other).join(','));\n"
	for _, row := range []struct{ name, main string }{
		{"sibling", "import { shared } from './held';\ntype Other = { b: number; d: 4 };\nconst other = shared as Other;\n" + read},
		{"absent field", "import { shared } from './held';\ntype Other = { b: number; c: number };\nconst other = shared as Other;\n" + read},
		{"chain", "import { shared } from './held';\ntype Middle = { b: number; m: number };\ntype Other = { b: number; m: number; c: number };\nconst middle = shared as Middle;\nconst other = middle as Other;\n" + read},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			enumerationReachRefusedAs(t, row.name, ".ts", enumerationReachShared, row.main)
		})
	}
}
