package oracle

// A memory probe of #w8nanzw's slot writes, never to land: each row is a program whose optional field
// write the head may now admit, run on Node, the JavaScript backend, native under the sanitizers, a
// release build and the counted build, and every outcome logged on a line starting PROBE.

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

type slotProbeRow struct {
	name   string
	source string
}

const slotProbeElement = `interface Node { readonly kind: number; readonly pos: number; }
interface Declaration extends Node { _declarationBrand: undefined; }
`

var slotProbeRows = []slotProbeRow{
	// Control: the head's own class_view.ts.
	{"q0_control_class_view", slotProbeElement + `interface Element extends Declaration { _elementBrand: undefined; name?: string; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function rename(n: Node): void { if (isElement(n)) { n.name = "w"; console.log(String(n.name)); } }
class Box { readonly kind = 171; readonly pos = 1; name?: string; label = "l"; }
const box = new Box();
rename(box);
console.log(String(box.name));
`},
	// Q1: a fresh function's literal in a statement region, written with a reference by a consumer.
	{"q1_region_plain_reference", `interface Box { kind: number; next?: Box }
function make(kind: number): Box { return { kind: kind, next: undefined }; }
function consume(b: Box): number { b.next = { kind: b.kind + 1, next: undefined }; return b.kind; }
const lone: Box = { kind: 0 };
let total = 0;
for (let index = 0; index < 300; index++) { total += consume(make(index)); }
console.log(String(total) + " " + String(lone.next === undefined));
`},
	{"q1_region_view_string", slotProbeElement + `interface Element extends Declaration { _elementBrand: undefined; tag?: string | undefined; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
interface Named extends Node { tag?: string | undefined; }
function make(pos: number): Named { return { kind: 171, pos: pos, tag: undefined }; }
function consume(n: Node): number { if (isElement(n)) { n.tag = "t" + String(n.pos); } return n.pos; }
let total = 0;
for (let index = 0; index < 300; index++) { total += consume(make(index)); }
console.log(String(total));
`},
	{"q1_region_construction_cast", `interface Full { kind: number; name: string }
function make(kind: number): Full { return { kind: kind } as Full; }
function consume(f: Full): number { f.name = "n" + String(f.kind); return f.kind; }
let total = 0;
for (let index = 0; index < 300; index++) { total += consume(make(index)); }
console.log(String(total));
`},
	{"q1_region_control_present", `interface Box { kind: number; next: Box | undefined }
function make(kind: number): Box { return { kind: kind, next: undefined }; }
function consume(b: Box): number { b.next = { kind: b.kind + 1, next: undefined }; return b.kind; }
let total = 0;
for (let index = 0; index < 300; index++) { total += consume(make(index)); }
console.log(String(total));
`},
	// Q2: a boolean written through a checked view into an optional boolean field.
	{"q2_class_boolean_true", slotProbeElement + `interface Element extends Declaration { _elementBrand: undefined; flag?: boolean; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function mark(n: Node): void { if (isElement(n)) { n.flag = true; console.log(String(n.flag)); } }
class Box { readonly kind = 171; readonly pos = 1; flag?: boolean; label = "l"; }
const box = new Box();
console.log(String(box.flag));
mark(box);
console.log(String(box.flag) + " " + String(box.flag === true));
`},
	{"q2_class_boolean_false", slotProbeElement + `interface Element extends Declaration { _elementBrand: undefined; flag?: boolean; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function mark(n: Node): void { if (isElement(n)) { n.flag = false; console.log(String(n.flag)); } }
class Box { readonly kind = 171; readonly pos = 1; flag?: boolean; label = "l"; }
const box = new Box();
mark(box);
console.log(String(box.flag) + " " + String(box.flag === false) + " " + String(box.flag === undefined));
`},
	{"q2_literal_boolean_value", slotProbeElement + `interface Element extends Declaration { _elementBrand: undefined; flag?: boolean; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function mark(n: Node, value: boolean): void { if (isElement(n)) { n.flag = value; console.log(String(n.flag)); } }
interface Named extends Node { flag?: boolean; }
const raw: Named = { kind: 171, pos: 1, flag: true };
mark(raw, false);
console.log(String(raw.flag) + " " + String(raw.flag === false));
mark(raw, true);
console.log(String(raw.flag) + " " + String(raw.flag === true));
`},
	{"q2_literal_boolean_undefined", slotProbeElement + `interface Element extends Declaration { _elementBrand: undefined; flag?: boolean | undefined; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function mark(n: Node, value: boolean): void { if (isElement(n)) { n.flag = value; console.log(String(n.flag)); } }
interface Named extends Node { flag?: boolean | undefined; }
const raw: Named = { kind: 171, pos: 1, flag: undefined };
mark(raw, false);
console.log(String(raw.flag) + " " + String(raw.flag === false));
mark(raw, true);
console.log(String(raw.flag) + " " + String(raw.flag === true));
`},
	{"q2_union_boolean", `enum SyntaxKind { Identifier = 80, PropertyAccessExpression = 211, ElementAccessExpression = 212 }
interface Node { readonly kind: SyntaxKind; readonly pos: number }
interface PropertyAccessExpression extends Node { readonly kind: SyntaxKind.PropertyAccessExpression; flag?: boolean | undefined }
interface ElementAccessExpression extends Node { readonly kind: SyntaxKind.ElementAccessExpression; flag?: boolean | undefined }
function mark(node: Node, value: boolean): void { (node as PropertyAccessExpression | ElementAccessExpression).flag = value; }
const odd: ElementAccessExpression = { kind: SyntaxKind.ElementAccessExpression, pos: 0, flag: undefined };
mark(odd, false);
console.log(String(odd.flag) + " " + String(odd.flag === false));
mark(odd, true);
console.log(String(odd.flag) + " " + String(odd.flag === true));
`},
	{"q2_control_class_number", slotProbeElement + `interface Element extends Declaration { _elementBrand: undefined; count?: number; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function put(n: Node, value: number): void { if (isElement(n)) { n.count = value; console.log(String(n.count)); } }
class Box { readonly kind = 171; readonly pos = 1; count?: number; label = "l"; }
const box = new Box();
put(box, 0);
console.log(String(box.count) + " " + String(box.count === 0));
put(box, 5);
console.log(String(box.count));
`},
	// Q3: graph programs, an object on a cycle with a never-written optional reference field written.
	{"q3_graph_literal_plain", `interface Link { kind: number; next?: Link; other?: Link }
function link(x: Link, y: Link): void { x.other = y; }
function turn(index: number): number {
  const a: Link = { kind: index, next: undefined, other: undefined };
  const b: Link = { kind: 1, next: a, other: undefined };
  a.next = b;
  link(a, b);
  link(b, b);
  return a.other === b && b.other === b ? 1 : 0;
}
const lone: Link = { kind: 0 };
let total = 0;
for (let index = 0; index < 1000; index++) { total += turn(index); }
console.log(String(total) + " " + String(lone.other === undefined));
`},
	{"q3_graph_class_view", slotProbeElement + `interface Element extends Declaration { _elementBrand: undefined; other?: Link; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
class Link { readonly kind = 171; readonly pos = 1; next?: Link; other?: Link; label = "l"; }
function link(n: Node, to: Link): void { if (isElement(n)) { n.other = to; } }
function turn(index: number): number {
  const a = new Link();
  const b = new Link();
  a.next = b;
  b.next = a;
  link(a, b);
  link(b, b);
  return a.other === b && b.other === b ? 1 : 0;
}
let total = 0;
for (let index = 0; index < 1000; index++) { total += turn(index); }
console.log(String(total));
`},
	{"q3_graph_class_plain", `interface Linked { other?: Link }
class Link { kind = 1; next?: Link; other?: Link; }
function link(x: Linked, to: Link): void { x.other = to; }
function turn(index: number): number {
  const a = new Link();
  const b = new Link();
  a.next = b;
  b.next = a;
  link(a, b);
  link(b, a);
  return a.other === b && b.other === a ? 1 : 0;
}
const empty: Linked = {};
let total = 0;
for (let index = 0; index < 1000; index++) { total += turn(index); }
console.log(String(total) + " " + String(empty.other === undefined));
`},
	// Q4: a class declaring f?: T never assigned, written through a view of an interface declaring f? too.
	{"q4_class_object_view", slotProbeElement + `class Holder { value = 1; }
interface Element extends Declaration { _elementBrand: undefined; parent?: Holder; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function attach(n: Node, h: Holder): number { if (isElement(n)) { n.parent = h; return n.parent.value; } return 0; }
class Box { readonly kind = 171; readonly pos = 1; parent?: Holder; label = "l"; }
let total = 0;
for (let index = 0; index < 200; index++) {
  const box = new Box();
  total += attach(box, new Holder());
  total += attach(box, new Holder());
}
console.log(String(total));
`},
	{"q4_class_string_view_twice", slotProbeElement + `interface Element extends Declaration { _elementBrand: undefined; name?: string; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function rename(n: Node, to: string): void { if (isElement(n)) { n.name = to; } }
class Box { readonly kind = 171; readonly pos = 1; name?: string; label = "l"; }
let last = "";
for (let index = 0; index < 200; index++) {
  const box = new Box();
  rename(box, "a" + String(index));
  rename(box, "b" + String(index));
  last = String(box.name);
}
console.log(last);
`},
	{"q4_base_class_object_view", slotProbeElement + `class Holder { value = 2; }
interface Element extends Declaration { _elementBrand: undefined; parent?: Holder; label: string; }
function isElement(n: Node): n is Element { return n.kind === 171; }
function attach(n: Node, h: Holder): void { if (isElement(n)) { n.parent = h; } }
class Base { readonly kind = 171; readonly pos = 1; parent?: Holder; }
class Derived extends Base { label = "l"; }
let total = 0;
for (let index = 0; index < 200; index++) {
  const derived = new Derived();
  attach(derived, new Holder());
  total += derived.parent === undefined ? 0 : derived.parent.value;
}
console.log(String(total));
`},
}

// slotProbeMarkers is the emitted C lines that say where a row's store goes: a checked view's write, a
// region literal's unzeroed allocation, a construction, and every scalar slot store.
var slotProbeMarkers = []string{"adamic_object_view_write(", "adamic_object_new_filled_in", "adamic_object_new_in(", "adamic_object_construct(", "->boolean = ", "->maybe_boolean = ", "adamic_region_end(", "adamic_object_new_graph"}

func slotProbeQuote(text []byte) string {
	lines := strings.Split(strings.TrimRight(string(text), "\n"), "\n")
	if len(lines) > 8 {
		lines = lines[:8]
	}
	return strings.Join(lines, " | ")
}

func TestZZSlotProbe(t *testing.T) {
	t.Parallel()
	for _, row := range slotProbeRows {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), row.name+".ts")
			if err := os.WriteFile(path, []byte(row.source), 0o644); err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			t.Logf("PROBE %s node: exit %d stdout %q stderr %q", row.name, node.exitCode, node.stdout, slotProbeQuote(node.stderr))
			program, err := lowered(t, path)
			if err != nil {
				var notYet *lower.NotYet
				if errors.As(err, &notYet) {
					t.Logf("PROBE %s lowering: STOPPED NotYet: %s", row.name, err)
				} else {
					t.Logf("PROBE %s lowering: STOPPED other: %s", row.name, err)
				}
				return
			}
			t.Logf("PROBE %s lowering: ADMITTED graphTypes %d", row.name, len(program.GraphTypes))
			code := native.C(program)
			counts := map[string]int{}
			var shown []string
			for _, line := range strings.Split(code, "\n") {
				for _, marker := range slotProbeMarkers {
					if strings.Contains(line, marker) {
						counts[marker]++
						if counts[marker] <= 3 {
							shown = append(shown, strings.TrimSpace(line))
						}
					}
				}
			}
			t.Logf("PROBE %s C markers: %v", row.name, counts)
			for _, line := range shown {
				t.Logf("PROBE %s C line: %s", row.name, line)
			}
			sanitized, _ := nativelyUncached(t, program)
			t.Logf("PROBE %s sanitized: exit %d stdout %q stderr %q agree=%q", row.name, sanitized.exitCode, sanitized.stdout, slotProbeQuote(sanitized.stderr), disagreement(node, sanitized))
			release := releasedUncached(t, program)
			t.Logf("PROBE %s release: exit %d stdout %q stderr %q agree=%q", row.name, release.exitCode, release.stdout, slotProbeQuote(release.stderr), disagreement(node, release))
			javascript := onJavaScriptBackend(t, program)
			t.Logf("PROBE %s javascript: exit %d stdout %q agree=%q", row.name, javascript.exitCode, javascript.stdout, disagreement(node, javascript))
			binary := filepath.Join(t.TempDir(), "counted")
			if err := native.Build(code, binary, native_runtime.Options{Count: true}); err != nil {
				t.Logf("PROBE %s counted: build failed: %v", row.name, err)
				return
			}
			counted := execute(t, binary)
			report := leakcheck.Unbalanced(leakcheck.Run{Stdout: counted.stdout, Stderr: counted.stderr, ExitCode: counted.exitCode})
			t.Logf("PROBE %s counted: exit %d stdout %q stderr %q unbalanced=%q", row.name, counted.exitCode, counted.stdout, slotProbeQuote(counted.stderr), report)
		})
	}
}
