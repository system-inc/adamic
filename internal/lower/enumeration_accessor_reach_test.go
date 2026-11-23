package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
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

// enumerationReachRefusal is the words an accessor that reaches the operand stops with, as before.
const enumerationReachRefusal = "Object enumeration with getter or accessor descriptors; use own data properties for this first reflection implementation"

// enumerationReachTemperature is a getter nothing enumerated can hold: Temperature is assignable to no
// type any enumeration below reads.
const enumerationReachTemperature = "export class Temperature { kelvin = 300; get celsius(): number { return this.kelvin - 273; } }\n"

// enumerationReachLower writes held.a and main.a to one directory and lowers main.a, returning its path.
func enumerationReachLower(t *testing.T, held, main string) (string, *ir.Program, error) {
	t.Helper()
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "held.a"), []byte(held), 0o644); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "main.a")
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
	path, program, err := enumerationReachLower(t, held, main)
	if err != nil {
		t.Fatalf("enumeration reach row %s was refused: %v", name, err)
	}
	agreeEntry(t, path, program, 0, nil, nil)
	if code, out, errs := brandedNative(t, program); code != 0 || out != stdout {
		t.Fatalf("enumeration reach row %s natively: exit=%d stdout=%q stderr=%q, want %q", name, code, out, errs, stdout)
	}
}

// enumerationReachRefused holds a row to the accessor refusal's words.
func enumerationReachRefused(t *testing.T, name, held, main string) {
	t.Helper()
	_, _, err := enumerationReachLower(t, held, main)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), enumerationReachRefusal) {
		t.Fatalf("enumeration reach row %s was admitted, want %q, got %v", name, enumerationReachRefusal, err)
	}
}

// Object.keys and Object.entries of a const literal, Object.values of a parameter (the checked
// enumeration, which asks the rule a second time) and encodeJson, beside a class whose getter no
// enumerated value can hold: each lowers and agrees with Node, where the program-wide scan refused all
// four. Mutant WholeGraph (every accessor reaches) turns it red.
func TestEnumerationReachUnrelatedGetterLowers(t *testing.T) {
	t.Parallel()
	const main = `import { encodeJson } from 'adamic';
import { Temperature } from './held.a';
interface Point { x: number; y: number }
const point = { x: 1, y: 2 };
function total(shape: Point): number {
  let sum = 0;
  for (const value of Object.values(shape)) {
    sum += value;
  }
  return sum;
}
const pairs = Object.entries(point);
console.log(Object.keys(point).join(','));
console.log(String(total(point)));
console.log(pairs.map((pair) => pair[0] + '=' + String(pair[1])).join(','));
const reading: Point = { x: 3, y: 4 };
console.log(encodeJson<Point>(reading));
console.log(String(new Temperature().celsius));
`
	enumerationReachAgrees(t, "unrelated", enumerationReachTemperature, main, "x,y\n3\nx=1,y=2\n{\"x\":3,\"y\":4}\n27\n")
}

// Object.keys of a parameter typed as a union of two object shapes, beside the unrelated getter: each
// member is classified and read, so the union is known and lowers. Mutant UnionMembers (a union's members
// are left unread) makes the union a type that can't be classified, which fails closed to the program-wide
// scan, and turns it red.
func TestEnumerationReachUnionLowers(t *testing.T) {
	t.Parallel()
	const main = `import { Temperature } from './held.a';
type Shape = { kind: string; width: number } | { kind: string; radius: number };
function names(shape: Shape): string {
  return Object.keys(shape).join(',');
}
console.log(names({ kind: 'box', width: 2 }));
console.log(names({ kind: 'disc', radius: 1 }));
console.log(String(new Temperature().celsius));
`
	enumerationReachAgrees(t, "union", enumerationReachTemperature, main, "kind,width\nkind,radius\n27\n")
}

// Each row's operand can hold a getter-bearing value, so each stays refused in today's words: an object
// literal's getter seen through a type it's assignable to, a class instance, the interface a getter's
// class implements, a union holding that interface, a class's static getter, a getter a subclass
// inherits (the base isn't assignable to the subclass, its instances are), Object.values' checked road,
// and a getter of a generic class, whose owner relates for no instantiation in particular.
func TestEnumerationReachReachingGetterRefused(t *testing.T) {
	t.Parallel()
	const sensor = "export interface Reading { readonly celsius: number }\nexport class Sensor implements Reading { get celsius(): number { return 21; } }\n"
	for _, row := range []struct{ name, held, main string }{
		{"literal", "export const literal = { get celsius(): number { return 21; } };\n", "import { literal } from './held.a';\nconst reading: { readonly celsius: number } = literal;\nconsole.log(Object.keys(reading).join(','));\n"},
		{"class", sensor, "import { Sensor } from './held.a';\nconst sensor = new Sensor();\nconsole.log(Object.keys(sensor).join(','));\n"},
		{"interface", sensor, "import { type Reading, Sensor } from './held.a';\nconst reading: Reading = new Sensor();\nconsole.log(Object.keys(reading).join(','));\n"},
		{"union", sensor, "import { type Reading, Sensor } from './held.a';\ntype Either = { label: string } | Reading;\nconst either: Either = new Sensor();\nconsole.log(Object.keys(either).join(','));\n"},
		{"static", "export class Sensor { static get celsius(): number { return 21; } }\n", "import { Sensor } from './held.a';\nconsole.log(Object.keys(Sensor).join(','));\n"},
		{"subclass", "export class Base { get celsius(): number { return 21; } }\nexport class Sub extends Base { label = 's'; }\n", "import { Sub } from './held.a';\nconst sub = new Sub();\nconsole.log(Object.keys(sub).join(','));\n"},
		{"values", sensor, "import { type Reading, Sensor } from './held.a';\nfunction read(reading: Reading): number[] { return Object.values(reading); }\nconsole.log(read(new Sensor()).join(','));\n"},
		{"generic", "export class Box<T> { held: T; constructor(held: T) { this.held = held; } get value(): T { return this.held; } }\n", "import { Box } from './held.a';\nconst reading: { readonly value: number } = { value: 1 };\nconsole.log(Object.keys(reading).join(','));\nconsole.log(String(new Box<number>(2).value));\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			enumerationReachRefused(t, row.name, row.held, row.main)
		})
	}
}

// encodeJson reads every field at any depth, so a getter-bearing value under the operand reaches it: a
// field typed as the interface a getter's class implements, and an array of it. Mutant DeepReach (encodeJson
// reads only the operand's own type) turns it red.
func TestEnumerationReachDeepGetterRefused(t *testing.T) {
	t.Parallel()
	const gauge = "export interface Inner { readonly value: number }\nexport class Gauge implements Inner { get value(): number { return 3; } }\n"
	for _, row := range []struct{ name, main string }{
		{"field", "import { encodeJson } from 'adamic';\nimport { type Inner, Gauge } from './held.a';\ninterface Outer { inner: Inner; label: string }\nconst outer: Outer = { inner: new Gauge(), label: 'g' };\nconsole.log(encodeJson<Outer>(outer));\n"},
		{"element", "import { encodeJson } from 'adamic';\nimport { type Inner, Gauge } from './held.a';\ninterface Outer { items: Inner[] }\nconst outer: Outer = { items: [new Gauge()] };\nconsole.log(encodeJson<Outer>(outer));\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			enumerationReachRefused(t, row.name, gauge, row.main)
		})
	}
}

// encodeJson of a union whose one member alone has a field typed as the interface a getter's class
// implements: the union's own properties are only what every member shares, so the field is reached
// through the member, and stays refused.
func TestEnumerationReachDeepUnionGetterRefused(t *testing.T) {
	t.Parallel()
	const gauge = "export interface Inner { readonly value: number }\nexport class Gauge implements Inner { get value(): number { return 3; } }\n"
	enumerationReachRefused(t, "deep union", gauge, "import { encodeJson } from 'adamic';\nimport { type Inner, Gauge } from './held.a';\ntype Shape = { kind: 'plain'; size: number } | { kind: 'held'; inner: Inner };\nconst shape: Shape = { kind: 'held', inner: new Gauge() };\nconsole.log(encodeJson<Shape>(shape));\n")
}

// An operand typed object or {} is any object at all, so its reach is unknown and the program-wide scan
// of before applies: the unrelated getter still refuses it.
func TestEnumerationReachUnknownKeepsProgramScan(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, main string }{
		{"object", "import { Temperature } from './held.a';\nfunction keys(value: object): string { return Object.keys(value).join(','); }\nconsole.log(keys({ a: 1 }));\nconsole.log(String(new Temperature().celsius));\n"},
		{"empty", "import { Temperature } from './held.a';\nfunction keys(value: {}): string { return Object.keys(value).join(','); }\nconsole.log(keys({ a: 1 }));\nconsole.log(String(new Temperature().celsius));\n"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			enumerationReachRefused(t, row.name, enumerationReachTemperature, row.main)
		})
	}
}
