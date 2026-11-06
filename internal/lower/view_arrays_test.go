package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestViewArrayReadonlyDescriptor(t *testing.T) {
	t.Parallel()
	for _, spelling := range []string{"readonly Node[]", "ReadonlyArray<Node>", "Node[]"} {
		t.Run(spelling, func(t *testing.T) {
			t.Parallel()
			l, target, release := mixedUnionLowering(t, "interface Node {readonly kind:number} type Target="+spelling+";")
			defer release()
			node := l.program.Files()[0].Statements.Nodes[0]
			id, err := l.viewContract(node, target)
			if err != nil {
				t.Fatal(err)
			}
			contract := l.result.ViewContracts[id-1]
			wantReadonly := spelling != "Node[]"
			if contract.Kind != ir.ViewArray || contract.Element == 0 || contract.ArrayReadonly != wantReadonly {
				t.Fatalf("lost array contract: %#v", contract)
			}
		})
	}
}

func TestViewMutableArrayFieldNotYet(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"items: Node[]", "readonly items: Node[]", "readonly items: Node[] | undefined"} {
		t.Run(field, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, `interface Node {readonly kind:number}
interface Holder {readonly kind:'Holder'; `+field+`}
const raw = {kind:'Holder' as const, items:[{kind:1}]};
const base:{readonly kind:'Holder'}=raw;
const holder = base as Holder;
console.log(String(holder.items !== undefined));`, ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != "a checked view of mutable array field items" {
				t.Fatalf("want mutable-array NotYet at read, got %v", err)
			}
		})
	}
}

func TestViewReadonlyArrayLengthIsRaw(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, `interface Node {readonly kind:number}
interface Holder {readonly kind:'Holder'; readonly items:readonly Node[]}
const raw = {kind:'Holder' as const, items:[{kind:1}]};
const base:{readonly kind:'Holder'}=raw;
const holder = base as Holder;
console.log(String(holder.items.length));`, ".ts")
	if err != nil {
		t.Fatalf("length reads no element: %v", err)
	}
}

func TestViewMutableArrayFieldUndemanded(t *testing.T) {
	t.Parallel()
	_, err := lowerSourceAs(t, `interface Node {readonly kind:number}
interface Holder {readonly kind:'Holder'; readonly items:Node[]}
const raw = {kind:'Holder' as const, items:[{kind:1}]};
const base:{readonly kind:'Holder'}=raw;
const holder = base as Holder;
console.log(holder.kind);`, ".ts")
	if err != nil {
		t.Fatalf("unused mutable field refused at cast: %v", err)
	}
}

// The source inventory grows with IR, while each admitted handout names a real
// oracle pin and must actually lower to a checked load or callback annotation.
func TestViewArrayAllowlist(t *testing.T) {
	t.Parallel()
	listed := map[string]bool{}
	for kind := range viewArrayRoads {
		listed[kind.Name()] = true
	}
	for name := range irOperandHolders(t) {
		if strings.HasPrefix(name, "Array") && !listed[name] {
			t.Errorf("ir.%s has no array-road classification", name)
		}
	}
	pins := map[string]string{"index": "index-good", "at": "at-lie", "for-of": "for-of-good", "values": "values-lie", "forEach": "forEach-good", "map": "map-good", "filter": "filter-good", "some": "some-good", "every": "every-good", "find": "find-good", "findIndex": "findIndex-good", "findLast": "findLast-good", "findLastIndex": "findLastIndex-good", "reduce": "reduce-good"}
	for road, pin := range pins {
		t.Run(road, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("../oracle/testdata/views_v3", pin+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerSourceAs(t, string(source), ".ts")
			if err != nil {
				t.Fatal(err)
			}
			found := false
			inspect := func(node any) bool {
				switch value := node.(type) {
				case ir.ViewElement:
					found = found || ((road == "index" || road == "at" || road == "values") && (road != "at" || value.Relative))
				case ir.ForOf:
					found = found || (road == "for-of" && value.ElementCheck != nil)
				case ir.ArrayMap:
					found = found || (road == "map" && value.ElementCheck != nil)
				case ir.ArrayReduce:
					found = found || (road == "reduce" && value.ElementCheck != nil)
				case ir.ArrayVisit:
					found = found || (road == value.Method && value.ElementCheck != nil)
				}
				return true
			}
			walk(program.Main, inspect)
			for _, fn := range program.Functions {
				walk(fn.Body, inspect)
			}
			if !found {
				t.Fatalf("%s does not exercise the checked %s road", pin, road)
			}
		})
	}
	for kind, road := range viewArrayRoads {
		if road.handout || road.noElement {
			continue
		}
		t.Run(kind.Name(), func(t *testing.T) {
			source := ir.Read{Of: ir.Array}
			value := reflect.New(kind).Elem()
			field := value.FieldByName("Array")
			if kind == reflect.TypeFor[ir.ParallelMap]() {
				field = value.FieldByName("Items")
			}
			if kind == reflect.TypeFor[ir.ArrayFrom]() {
				if viewReadOf(nil, value.Interface().(ir.Expression)) != viewReadNotYet {
					t.Fatal("Array.from must fail closed")
				}
				return
			}
			if !field.IsValid() {
				t.Fatalf("denied road has no array operand: %v", kind)
			}
			field.Set(reflect.ValueOf(source))
			program := &ir.Program{ViewOrigins: []ir.Expression{source}, ViewContracts: []ir.ViewContract{{Kind: ir.ViewArray, ArrayReadonly: true}}, Main: []ir.Statement{ir.Evaluate{Value: value.Interface().(ir.Expression)}}}
			l := &lowering{result: program}
			var stop *NotYet
			if err := l.checkViewArrays(); !errors.As(err, &stop) || !strings.Contains(stop.What, road.member) {
				t.Fatalf("unchecked %s road did not fail closed: %v", road.member, err)
			}
		})
	}
}

// keyedProgram is a program lowered with its element key kept, so a test can read the key and ask
// it about the program's own reads.
type keyedProgram struct {
	key      *viewElementKey
	lowering *lowering
	entry    *ast.SourceFile
}

// lowerKeyed lowers the file at path once, with no second lowering when the key misses, which is
// lowerCheckedWithCycleTypes' fallback. The lowering is nil when it fails.
func lowerKeyed(t *testing.T, path string) (keyedProgram, error) {
	t.Helper()
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	t.Cleanup(release)
	key := &viewElementKey{}
	built, err := lowerWithElementKey(key, loaded, check, entry, nil, false, false, nil)
	return keyedProgram{key: key, lowering: built, entry: entry}, err
}

// elementAccesses is every array[index] in a file, in source order.
func elementAccesses(entry *ast.SourceFile) []*ast.Node {
	var found []*ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindElementAccessExpression {
			found = append(found, node)
		}
		node.ForEachChild(visit)
		return false
	}
	entry.AsNode().ForEachChild(visit)
	return found
}

// loweredElementChecks is the ViewType of every checked element read a lowered program keeps, and
// how many implicit handouts keep an element check.
func loweredElementChecks(program *ir.Program) ([]string, int) {
	var reads []string
	handouts := 0
	inspect := func(node any) bool {
		switch value := node.(type) {
		case ir.ViewElement:
			reads = append(reads, value.ViewType)
		case ir.ForOf:
			if value.ElementCheck != nil {
				handouts++
			}
		case ir.ArrayMap:
			if value.ElementCheck != nil {
				handouts++
			}
		case ir.ArrayVisit:
			if value.ElementCheck != nil {
				handouts++
			}
		case ir.ArrayReduce:
			if value.ElementCheck != nil {
				handouts++
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	return reads, handouts
}

// The key checks an element read only where the program views an array whose element type is
// comparable to the read's (Lower's ruling on #6wfvfhm). A program that views no array, here one
// that views a plain Holder, asks no read for a check, so none is built and it lowers as main does.
// Node shares no field with Holder, so a Holder view can't be held as a Node, and main's V2 rule for
// a read through a type a view may reach doesn't stop node.id.
func TestViewElementKeyViewsNoArray(t *testing.T) {
	t.Parallel()
	keyed, err := lowerKeyed(t, writeAgreementSource(t, "main.ts", `interface Holder { readonly kind: 'Holder'; readonly count: number; }
interface Node { readonly id: number; }
const raw = {kind: 'Holder' as const, count: 2};
const base: {readonly kind: 'Holder'} = raw;
const holder = base as Holder;
const tokens: number[] = [3, 4];
const nodes: Node[] = [{id: 1}];
let total = 0;
for (const token of tokens) total += token;
const kinds = nodes.map((node: Node) => node.id);
const first = nodes[0];
console.log(String(tokens[1]));
console.log(String(kinds.length));
console.log(String(total));
console.log(String(first?.id));
console.log(String(holder.count));
`))
	if err != nil {
		t.Fatal(err)
	}
	if keyed.key.every || len(keyed.key.order) != 0 || len(keyed.key.misses) != 0 {
		t.Fatalf("a program that views no array keyed every %t, viewed %d, missed %v", keyed.key.every, len(keyed.key.order), keyed.key.misses)
	}
	for _, access := range elementAccesses(keyed.entry) {
		check, err := keyed.lowering.viewArrayElementCheck(access, access.AsElementAccessExpression().Expression)
		if err != nil || check != nil {
			t.Errorf("%s: a read in a program that views no array was given a check: %#v, %v", keyed.lowering.program.Where(access), check, err)
		}
	}
	if reads, handouts := loweredElementChecks(keyed.lowering.result); len(reads) != 0 || handouts != 0 {
		t.Fatalf("a program that views no array kept checked reads %v and %d checked handouts", reads, handouts)
	}
}

// A program that views Node[] checks its Node reads, and marks its number[] reads unrelated: they
// can't hand out a viewed element, so they read raw and never stop as unchecked.
func TestViewElementKeyMarksUnrelatedReads(t *testing.T) {
	t.Parallel()
	keyed, err := lowerKeyed(t, writeAgreementSource(t, "main.ts", `interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw = {tag:1 as const, _children:[{kind:1}]};
const base: Base = raw;
const view = base as SyntaxList;
const tokens: number[] = [3, 4];
let total = 0;
for (const token of tokens) total += token;
const child = view._children[0];
console.log(String(tokens[1]));
console.log(String(total));
console.log(String(child?.kind));
`))
	if err != nil {
		t.Fatal(err)
	}
	if keyed.key.every || len(keyed.key.order) == 0 || len(keyed.key.misses) != 0 {
		t.Fatalf("keyed every %t, viewed %d, missed %v", keyed.key.every, len(keyed.key.order), keyed.key.misses)
	}
	for _, access := range elementAccesses(keyed.entry) {
		check, err := keyed.lowering.viewArrayElementCheck(access, access.AsElementAccessExpression().Expression)
		if err != nil || check == nil {
			t.Fatalf("%s: no check or mark: %v", keyed.lowering.program.Where(access), err)
		}
		if unrelated := check.ViewType == "number"; check.Unrelated != unrelated || !unrelated && check.Contract == 0 {
			t.Errorf("%s: a read of %s keyed %#v", keyed.lowering.program.Where(access), check.ViewType, check)
		}
	}
	reads, handouts := loweredElementChecks(keyed.lowering.result)
	if len(reads) != 1 || reads[0] != "Node" || handouts != 0 {
		t.Fatalf("want the one checked Node read, got reads %v and %d checked handouts", reads, handouts)
	}
}

// filter with a type predicate hands a viewed Node out as an Identifier, and Node isn't assignable
// to Identifier, so the key relates them both ways and the Identifier read stays checked. With the
// relation one way, T assignable to E only, this goes red (view_elements.mutants.md).
func TestViewElementKeyPredicateFilter(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("../oracle/testdata/views_v3/filter-predicate-good.ts")
	if err != nil {
		t.Fatal(err)
	}
	keyed, err := lowerKeyed(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if keyed.key.every || len(keyed.key.misses) != 0 {
		t.Fatalf("keyed every %t, missed %v", keyed.key.every, keyed.key.misses)
	}
	reads, _ := loweredElementChecks(keyed.lowering.result)
	if !strings.Contains(strings.Join(reads, ","), "Identifier") {
		t.Fatalf("the predicate-narrowed read is unchecked: checked reads %v", reads)
	}
}

// Lower's condition on the key: every array contract lowering makes is one the pre-pass saw. One it
// didn't is a read the key may have left raw, and lowerCheckedWithCycleTypes then lowers again with
// every read checked. Each views_v3 fixture must decide and need no second lowering.
func TestViewElementKeyAgreesWithLowering(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("../oracle/testdata/views_v3/*.ts")
	if err != nil || len(paths) == 0 {
		t.Fatalf("no fixtures: %v", err)
	}
	for _, path := range paths {
		// console.ts stops at the prelude's check, before lowering (internal/oracle's TestViewArrayConsole).
		if filepath.Base(path) == "console.ts" {
			continue
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			keyed, _ := lowerKeyed(t, path)
			if keyed.key.every {
				t.Fatal("the pre-pass couldn't decide, so every read is checked")
			}
			if len(keyed.key.misses) != 0 {
				t.Fatalf("lowering made array contracts the pre-pass never saw: %v", keyed.key.misses)
			}
		})
	}
}

// The agreement's own guard: an array contract made where the pre-pass can't see it is a miss.
// With the miss unrecorded, this goes red (view_elements.mutants.md).
func TestViewElementKeyMissesAnUnseenArray(t *testing.T) {
	t.Parallel()
	keyed, err := lowerKeyed(t, writeAgreementSource(t, "main.ts", `interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw = {tag:1 as const, _children:[{kind:1}]};
const base: Base = raw;
const view = base as SyntaxList;
const names: readonly string[] = ['a'];
console.log(String(view._children[0]?.kind));
console.log(String(names.length));
`))
	if err != nil {
		t.Fatal(err)
	}
	if len(keyed.key.misses) != 0 {
		t.Fatalf("missed before the unseen contract: %v", keyed.key.misses)
	}
	var names *ast.Node
	for _, statement := range keyed.entry.Statements.Nodes {
		if statement.Kind == ast.KindVariableStatement {
			if declaration := statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0]; declaration.Name().Text() == "names" {
				names = declaration
			}
		}
	}
	if names == nil {
		t.Fatal("no names declaration")
	}
	if _, err := keyed.lowering.viewContract(names, keyed.lowering.checker.GetTypeAtLocation(names.Name())); err != nil {
		t.Fatal(err)
	}
	if len(keyed.key.misses) != 1 {
		t.Fatalf("an array contract the pre-pass never saw passed unnoticed: misses %v", keyed.key.misses)
	}
}

// An unrelated read isn't an unreachable one (Lower's must on #6wfvfhm): a road that hands a viewed
// Node[] to an Other[] that relates to Node neither way leaves the key no check to build. So
// checkViewArrays reads an unrelated read raw only where its array can't reach a viewed allocation, and
// stops it where it can. Today every such road already stops before the read, at the hop itself, so
// these rows pin those stops: the head that lifts one meets this rule and its read stops. The rule's
// mutant (the put-back above demanded) has no live road to turn red until then
// (view_elements.mutants.md).
func TestViewElementUnrelatedButReachableStops(t *testing.T) {
	t.Parallel()
	const prelude = `interface Node { readonly kind: number; }
interface Other { readonly name: string; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw = {tag: 1 as const, _children: [{kind: 1}]};
const base: Base = raw;
const view = base as SyntaxList;
`
	for _, row := range []struct{ name, source, want string }{
		{"through any", prelude + "const a: any = view._children;\nconst others: readonly Other[] = a;\nconsole.log(String(others[0]?.name.length));\n", "a value of type any"},
		{"through as unknown as", prelude + "const others = view._children as unknown as readonly Other[];\nconsole.log(String(others[0]?.name.length));\n", ""},
		{"through object", prelude + "const wide: object = view._children;\nconst others = wide as readonly Other[];\nconsole.log(String(others[0]?.name.length));\n", ""},
	} {
		_, err := lowerSourceAs(t, row.source, ".ts")
		var notYet *NotYet
		var refused *Refused
		if !errors.As(err, &notYet) && !errors.As(err, &refused) {
			t.Errorf("%s: got %v, want a lowering stop before an Other[] reads a viewed Node raw", row.name, err)
			continue
		}
		t.Logf("%s: %v", row.name, err)
		if row.want != "" && !strings.Contains(err.Error(), row.want) {
			t.Errorf("%s: got %v, want a stop containing %q", row.name, err, row.want)
		}
	}
}

// A copy road names its read's line and column, as every other stop does (Language's should on
// #6wfvfhm, probes p01, p03, p07 and p15). slice, a spread, new Set and new Map have no checked form
// yet, so a viewed array copied through one stops, and the stop is recorded where the copy's array is
// read. Without the record, checkViewArrays reaches these roads only through IR nodes that carry no
// location, and the stop names the file alone.
func TestViewCopyRoadsStopAtTheirRead(t *testing.T) {
	t.Parallel()
	const prelude = `interface Node { readonly kind: number; }
interface Base { readonly tag: 1; }
interface SyntaxList extends Base { readonly _children: readonly Node[]; }
const raw = { tag: 1 as const, _children: [7] };
const base: Base = raw;
const view = base as SyntaxList;
console.log('before');
`
	for _, row := range []struct{ road, line, where string }{
		{"slice", "const copy = view._children.slice();\nconsole.log(String(copy[0]?.kind));\n", "main.ts:8:14"},
		{"spread", "const copy = [...view._children];\nconsole.log(String(copy[0]?.kind));\n", "main.ts:8:18"},
		{"Set constructor", "const seen = new Set(view._children);\nfor (const child of seen) {\n    console.log(String(child.kind));\n}\n", "main.ts:8:22"},
		{"Map constructor", "const byKind = new Map(view._children.map((child, index) => [index, child] as const));\nconsole.log(String(byKind.get(0)?.kind));\n", "main.ts:8:24"},
	} {
		t.Run(row.road, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, prelude+row.line+"console.log('after');\n", ".ts")
			var notYet *NotYet
			want := "a checked view's array read through " + row.road + " (not checked yet)"
			if !errors.As(err, &notYet) || notYet.What != want || !strings.HasSuffix(notYet.Where, row.where) {
				t.Fatalf("got %v, want %q at %s", err, want, row.where)
			}
		})
	}
}
