package lower

import (
	"context"
	"errors"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// Every IR expression that holds an operand is one of rangeUnion's three classes, and
// TestEveryOperandIsClassedForRangeUnions finds them in internal/ir's source (irOperandHolders,
// view_reads_test.go), so an expression added to the IR fails it until someone says what it does with a range
// type's Union (#hae423v, Lower's Must 2). A class can turn on the value: a narrow asks the kind only when
// checked, an object literal reads a layout only through a spread. Each example is that value, and
// rangeUnionOf must give it its class; a stopping one's operand is set, and rangeUnionStopped must name it.
//
// Mutant (scratch-ha1/range_unions.go.class_removed.mutant): TypeOf taken out of the table's dispatches, and
// this goes red naming it.

type rangeUnionExample struct {
	expression ir.Expression
	why        string
}

// rangeHeld is a stopping example's operand, held as a Union.
var rangeHeld = ir.Read{Of: ir.Union}

// rangeDispatches ask the kind at run time on both backends before they read any layout.
var rangeDispatches = []rangeUnionExample{
	{ir.Narrow{Checked: true}, "a checked narrow tests the member's kind and panics on any other (native union.go's checked arm, JavaScript's adamicPlaceholderNarrow)"},
	{ir.ArrayIsArray{}, "tests a value's brand"},
	{ir.CheckedCast{}, "reads the discriminant through the cast's own check"},
	{ir.DynamicKeyedProperty{}, "a run-time key read, by the allocation's own runtime shape"},
	{ir.DynamicProperty{}, "a dynamic read switches on the value's kind (union.c's dynamic_property), and JavaScript reads value[key]"},
	{ir.HasProperty{}, "an in test switches on the value's kind (union.c's has_property answers an array from its index, side record and methods), and JavaScript's in handles an array"},
	{ir.ErrorToString{}, "taken only where UnknownKind said the value is an Error"},
	{ir.IsCollectionIterator{}, "reads the runtime object's own heap kind"},
	{ir.IsGenerator{}, "reads the runtime object's own heap kind"},
	{ir.IsRegExpGroups{}, "adamic_regexp_groups_is asks the heap kind first, then an object's shape, and JavaScript asks the real value's prototype"},
	{ir.JSONStringify{}, "json_objects.c reads by the allocation's own representation bytes, an array's as an array's"},
	{ir.ObjectTag{}, "adamic_object_tag asks the heap kind first, then an object's shape address, and JavaScript asks the real value's Object.prototype.toString"},
	{ir.OwnFunction{}, "taken only where UnknownKind said the value is an object, and reads by its own runtime shape"},
	{ir.TypeOf{}, "adamic_union_typeof switches on the kind"},
	{ir.UnionToString{}, "a scalar union's spelling, or to_string_hint for an unknown, read by its own runtime tags"},
	{ir.UnknownKind{}, "tests a value's own runtime tags"},
	{ir.ViewElement{}, "checks an element against its declared contract at the read"},
}

// rangeStops read their object operand's layout raw, and are NotYet where it's held as a range type's Union.
var rangeStops = []rangeUnionExample{
	{ir.Narrow{Value: rangeHeld, To: ir.Object}, "an unchecked narrow to an object takes the union's word for its kind (native's raw cast), which an array in it breaks"},
	{ir.Narrow{Value: rangeHeld, To: ir.Array}, "an unchecked narrow to an array takes the union's word for its kind, which an object in it breaks"},
	{ir.ObjectLiteral{Spread: rangeHeld}, "a spread copies its source's fields by its layout"},
	{ir.RecordLiteral{Spread: rangeHeld}, "a spread into a record copies its source's fields by its layout"},
	{ir.ObjectCall{Method: "keys", Arguments: []ir.Expression{rangeHeld}}, "Object.keys reads an object's own keys from its layout"},
	{ir.ObjectCall{Method: "assign", Arguments: []ir.Expression{rangeHeld}}, "Object.assign copies each source's fields by its layout"},
	{ir.BoundMethod{Object: rangeHeld}, "reads its method through raw dispatch"},
	{ir.HasAccessor{Object: rangeHeld}, "reads the object's class chain"},
	{ir.HasOwn{Object: rangeHeld}, "reads the object's own field table"},
	{ir.InstanceOf{Value: rangeHeld}, "an exact test reads the object's class word after a null test only (emit_expressions.go)"},
	{ir.JSONEncode{Value: rangeHeld}, "encodeJson reads every field by its schema, raw"},
	{ir.ObjectKeys{Object: rangeHeld}, "Object.keys and a for-in read the object's own keys from its layout"},
	{ir.Property{Object: rangeHeld, Name: "pos"}, "a field read through the object's layout; a read of pos or end through a range type is a dispatch whose arms take checked narrows (range_fields.go)"},
	{ir.RecordStorage{Value: rangeHeld}, "reads an object's storage word"},
}

// rangeMoves read no object's layout.
var rangeMoves = []rangeUnionExample{
	{ir.Narrow{To: ir.String}, "an unchecked narrow to a scalar or a closure, which a range type's union never holds a member of"},
	{ir.ObjectLiteral{}, "a new object of the values it's given, moved"},
	{ir.RecordLiteral{}, "a new record of the entries it's given, moved"},
	{ir.ObjectCall{Method: "is"}, "compares by SameValue"},
	{ir.ArrayConcat{}, "an array road: an array operand, its elements moved"},
	{ir.ArrayFill{}, "an array road: an array operand, its elements moved"},
	{ir.ArrayFrom{}, "an array road: an array operand, its elements moved"},
	{ir.ArrayIndex{}, "reads an array's element"},
	{ir.ArrayJoin{}, "an array road: an array operand"},
	{ir.ArrayLiteral{}, "a new array of the elements it's given, moved, a spread's iterable an array"},
	{ir.ArrayMap{}, "calls back with the elements, moved"},
	{ir.ArrayOfScalars{}, "tests an array's elements"},
	{ir.ArrayPop{}, "an array road: an array operand, its element moved"},
	{ir.ArrayPush{}, "an array road: the elements moved in"},
	{ir.ArrayReduce{}, "calls back with the elements, moved"},
	{ir.ArrayReverse{}, "an array road: an array operand"},
	{ir.ArraySearch{}, "an array road: compares elements by identity or value"},
	{ir.ArraySlice{}, "an array road: an array operand, its elements moved"},
	{ir.ArraySort{}, "an array road: an array operand, its elements moved"},
	{ir.ArraySplice{}, "an array road: an array operand, its elements moved"},
	{ir.ArrayVisit{}, "calls back with the elements, moved"},
	{ir.Await{}, "the settled value itself"},
	{ir.Binary{}, "an operator on numbers, booleans or strings, or an identity comparison"},
	{ir.BooleanToString{}, "a boolean's spelling"},
	{ir.Box{}, "the value itself, where a union goes"},
	{ir.Call{}, "moves its arguments into the callee's parameters, where the reads are classed"},
	{ir.CallClosure{}, "moves its arguments into the closure's parameters, where the reads are classed"},
	{ir.CharCodeAt{}, "reads a string"},
	{ir.Coalesce{}, "one of its operands"},
	{ir.CodePoints{}, "a new array of a string's code points"},
	{ir.CollectionIterator{}, "an iterator over an array, a Map or a Set"},
	{ir.Comma{}, "its operands, in order"},
	{ir.Concat{}, "a new string"},
	{ir.Conditional{}, "one of its branches"},
	{ir.Defined{}, "the value itself, tested for undefined"},
	{ir.Effects{}, "runs statements; a read in them is classed on its own"},
	{ir.FileStatus{}, "reads a path from the host"},
	{ir.FunctionLength{}, "a closure's arity"},
	{ir.GeneratorDone{}, "reads whether the generator's body finished"},
	{ir.GeneratorNext{}, "the value its body yields, moved"},
	{ir.IntlCollatorCompareValue{}, "the compare closure of a collator"},
	{ir.IntlCollatorNew{}, "a new collator from its locale, a string"},
	{ir.IsNull{}, "tests a reference"},
	{ir.IsUndefined{}, "tests a reference"},
	{ir.IterableOf{}, "a generator where an Iterable<T> slot holds it"},
	{ir.IteratorNext{}, "a hand-driven next()'s result, moved"},
	{ir.JSONDecode{}, "parses a string"},
	{ir.JoinEnter{}, "compares the array's identity with the arrays being joined"},
	{ir.Length{}, "an array's length"},
	{ir.Logical{}, "one of its operands"},
	{ir.LogicalAssignment{}, "its read is a Property and its write a SetProperty, each classed where it is"},
	{ir.MakeError{}, "a new Error from strings"},
	{ir.MapClear{}, "empties a map"},
	{ir.MapDelete{}, "removes an entry"},
	{ir.MapEntries{}, "a new array of a map's entries, moved"},
	{ir.MapForEach{}, "calls back with a map's entries, moved"},
	{ir.MapGet{}, "moves an entry's value out"},
	{ir.MapHas{}, "tests for an entry"},
	{ir.MapKeys{}, "a new array of a map's keys"},
	{ir.MapNew{}, "a new map from pairs, each moved"},
	{ir.MapSet{}, "moves a value in"},
	{ir.MapSize{}, "a map's size"},
	{ir.MapValues{}, "a new array of a map's values, moved"},
	{ir.MathCall{}, "a function of numbers"},
	{ir.MaybeOf{}, "a number or a boolean that may be missing"},
	{ir.MaybeToString{}, "a new string"},
	{ir.NodeBufferCall{}, "a host operation on buffers, strings and numbers"},
	{ir.NodeFSFile{}, "a host operation on strings, numbers and host-made handles"},
	{ir.NodeHostCall{}, "a host operation on strings and numbers"},
	{ir.NumberCall{}, "a function of numbers and strings"},
	{ir.NumberFormat{}, "a new string"},
	{ir.NumberToString{}, "a new string"},
	{ir.ParallelMap{}, "calls workers with the elements, moved"},
	{ir.PhantomMember{}, "a member read from a primitive"},
	{ir.PlaceholderUse{}, "the placeholder's value itself, once its readiness flag is checked"},
	{ir.ProcessCall{}, "a host operation on the prelude's process"},
	{ir.PromiseValue{}, "the value itself, settled"},
	{ir.ReadDirectory{}, "reads a path from the host"},
	{ir.ReadTextFile{}, "reads a path from the host"},
	{ir.RecordCall{}, "a record's own entries, its values moved"},
	{ir.RecordCoalesce{}, "a record's own entry, its value moved"},
	{ir.RegExpCall{}, "a pattern on a string"},
	{ir.RegExpGroup{}, "a match's named group"},
	{ir.RegExpNew{}, "a new pattern"},
	{ir.RegExpProperty{}, "a match's own field"},
	{ir.SetAdd{}, "moves a value in"},
	{ir.SetNew{}, "a new set of an array's elements, moved"},
	{ir.SetValues{}, "a new array of a set's elements, moved"},
	{ir.StringCall{}, "a string's method"},
	{ir.StringFromCodes{}, "a new string"},
	{ir.StringIndex{}, "reads a string"},
	{ir.StringIterator{}, "an iterator over a string"},
	{ir.StringLength{}, "reads a string"},
	{ir.ToFixed{}, "a new string"},
	{ir.Trim{}, "a new string"},
	{ir.Truthy{}, "tests a value"},
	{ir.TypedArrayFill{}, "writes a typed array"},
	{ir.TypedArrayNew{}, "a new typed array"},
	{ir.TypedArraySet{}, "writes a typed array"},
	{ir.TypedArraySort{}, "sorts a typed array"},
	{ir.TypedArraySubarray{}, "a view of a typed array"},
	{ir.Unary{}, "an operator on a number or a boolean"},
	{ir.Unwrap{}, "the value itself, checked present"},
	{ir.Utf8At{}, "reads a string"},
	{ir.Utf8Length{}, "reads a string"},
	{ir.Void{}, "evaluates its operand and reads nothing of it"},
	{ir.WeakCollectionDelete{}, "drops the entry keyed by the key's weak handle, by identity"},
	{ir.WeakCollectionHas{}, "tests for the entry keyed by the key's weak handle, by identity"},
	{ir.WeakMapGet{}, "the entry keyed by the key's weak handle, by identity; its value is a scalar"},
	{ir.WeakMapSet{}, "keys an entry by the key's weak handle, by identity; its value is a scalar"},
	{ir.WeakOf{}, "the target's weak handle"},
	{ir.WeakSetAdd{}, "keys an entry by the key's weak handle, by identity"},
	{ir.WeakTarget{}, "reads a weak reference"},
	{ir.WriteTextFile{}, "writes a string to the host"},
	{ir.Yield{}, "moves the value out to the consumer"},
}

func TestEveryOperandIsClassedForRangeUnions(t *testing.T) {
	t.Parallel()
	holders := irOperandHolders(t)
	name := func(expression ir.Expression) string { return reflect.TypeOf(expression).Name() }
	classed := map[string]bool{}
	for _, list := range []struct {
		class    rangeUnion
		named    string
		examples []rangeUnionExample
	}{
		{rangeUnionDispatches, "dispatches", rangeDispatches},
		{rangeUnionMoves, "moves", rangeMoves},
		{rangeUnionStops, "stops", rangeStops},
	} {
		for _, example := range list.examples {
			classed[name(example.expression)] = true
			if strings.TrimSpace(example.why) == "" {
				t.Errorf("ir.%s is listed as %s without saying why", name(example.expression), list.named)
			}
			if got := rangeUnionOf(example.expression); got != list.class {
				t.Errorf("rangeUnionOf(ir.%s %+v) = %d, want %s", name(example.expression), example.expression, got, list.named)
			}
			stopped := rangeUnionStopped(example.expression)
			if list.class == rangeUnionStops && !slices.ContainsFunc(stopped, func(operand ir.Expression) bool { return operand == ir.Expression(rangeHeld) }) {
				t.Errorf("rangeUnionStopped(ir.%s) = %v, want its operand, so checkRangeUnionReads asks about it", name(example.expression), stopped)
			}
			if list.class != rangeUnionStops && len(stopped) != 0 {
				t.Errorf("rangeUnionStopped(ir.%s) = %v for a class that doesn't stop", name(example.expression), stopped)
			}
		}
	}
	expressions := slices.Sorted(maps.Keys(holders))
	for _, expression := range expressions {
		if holders[expression] && !classed[expression] {
			t.Errorf("ir.%s holds an operand and has no class for a range type's union: give it one in rangeUnionOf (range_unions.go) and list it in range_unions_test.go with why", expression)
		}
	}
	for listed := range classed {
		if _, ok := holders[listed]; !ok {
			t.Errorf("ir.%s is classed but isn't an expression in internal/ir", listed)
		}
	}
	t.Logf("%d expression examples: %d dispatch, %d move, %d stop", len(rangeDispatches)+len(rangeMoves)+len(rangeStops), len(rangeDispatches), len(rangeMoves), len(rangeStops))
}

// rangeProgram loads source as main.ts and gives back its checker and each parameter's name node, by name.
func rangeProgram(t *testing.T, source string) (*load.Program, *checker.Checker, map[string]*ast.Node) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	checked, release := program.Checker(context.Background(), program.Files()[0])
	t.Cleanup(release)
	parameters := map[string]*ast.Node{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindParameter && ast.IsIdentifier(node.Name()) {
			parameters[node.Name().Text()] = node.Name()
		}
		return node.ForEachChild(visit)
	}
	program.Files()[0].AsNode().ForEachChild(visit)
	return program, checked, parameters
}

const rangeTextRange = "interface TextRange { pos: number; end: number }\ninterface Node { kind: number }\n"

const rangeNodeArray = "interface NodeArray<T extends Node> extends ReadonlyArray<T>, TextRange { readonly hasTrailingComma: boolean; transformFlags: number }\n"

// A closed range type is held as a Union in a program that declares an interface with the NodeArray layout,
// wherever it's declared and whether or not any NodeArray value goes near the range type, and as an object in
// any other program, so the oracle corpus, which declares none, doesn't move (Lower's Must 1 on #hae423v). The
// NodeArray here is generic, so its declared type has no layout of its own to compute: the bit asks its
// ancestry and its pos and end, never its element. A Node stays an object either way.
//
// Mutant (scratch-ha1/range_types.go.walk_misses.mutant): the bit walks only a file's top-level statements, so
// a NodeArray declared in a function's body is missed, and "a NodeArray in a function's body" goes red.
func TestRangeTypeHeldAsUnionWithNodeArrayLayout(t *testing.T) {
	t.Parallel()
	width := "export function width(range: TextRange, node: Node): number { return node.kind + range.end - range.pos; }\n"
	for _, probe := range []struct {
		name, source string
		want         ir.Type
	}{
		{"a NodeArray interface", rangeTextRange + rangeNodeArray + width, ir.Union},
		{"a NodeArray in a function's body", rangeTextRange + "export function made(): number {\n\t" + rangeNodeArray + "\treturn 0;\n}\n" + width, ir.Union},
		{"no NodeArray interface", rangeTextRange + width, ir.Object},
		{"an array interface without pos and end", rangeTextRange + "interface Names extends ReadonlyArray<string> { readonly hasTrailingComma: boolean }\n" + width, ir.Object},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, checked, parameters := rangeProgram(t, probe.source)
			l := &lowering{program: program, checker: checked, result: &ir.Program{}}
			if of, known := l.representation(checked.GetTypeAtLocation(parameters["range"])); !known || of != probe.want {
				t.Errorf("TextRange is held as %v (known %t), want %v", of, known, probe.want)
			}
			if of, known := l.representation(checked.GetTypeAtLocation(parameters["node"])); !known || of != ir.Object {
				t.Errorf("Node is held as %v (known %t), want an object", of, known)
			}
		})
	}
}

// A field read of pos or end through a range type held as a Union is a dispatch (range_fields.go): a helper that
// asks ArrayIsArray and reads each arm through a checked narrow to its kind, with no unchecked narrow to an
// object or an array left anywhere in the program. An optional chain through one still stops. The same program
// with no NodeArray interface lowers as it did, its reads plain object reads and no helper made.
func TestRangeUnionFieldReadDispatches(t *testing.T) {
	t.Parallel()
	width := "export function width(range: TextRange): number { return range.end - range.pos; }\nconsole.log('ok');\n"
	helpers := func(program *ir.Program) (made int, unasked int) {
		inspect := func(node any) bool {
			if narrow, isNarrow := node.(ir.Narrow); isNarrow && !narrow.Checked && (narrow.To == ir.Object || narrow.To == ir.Array) {
				unasked++
			}
			return true
		}
		for _, function := range program.Functions {
			if function.Name == "range_field_read" {
				made++
				asks := false
				walk(function.Body, func(node any) bool {
					_, isTest := node.(ir.ArrayIsArray)
					asks = asks || isTest
					return true
				})
				if !asks {
					t.Errorf("a range_field_read helper never asks ArrayIsArray")
				}
			}
			walk(function.Body, inspect)
		}
		walk(program.Main, inspect)
		return made, unasked
	}
	program, err := lowerSourceAs(t, rangeTextRange+rangeNodeArray+width, ".ts")
	if err != nil {
		t.Fatalf("with a NodeArray: got %v, want it lowered", err)
	}
	if made, unasked := helpers(program); made != 2 || unasked != 0 {
		t.Errorf("with a NodeArray: %d range_field_read helpers and %d unchecked narrows to an object or an array, want 2 and 0", made, unasked)
	}
	_, err = lowerSourceAs(t, rangeTextRange+rangeNodeArray+"export function start(range: TextRange | undefined): number { return range?.pos ?? -1; }\nconsole.log('ok');\n", ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.Contains(notYet.Where, "main.ts:4:") || !strings.Contains(notYet.What, "a read of field pos through a range type held as a union") {
		t.Errorf("an optional chain: got %v, want the range field read's stop at main.ts:4", err)
	}
	plain, err := lowerSourceAs(t, rangeTextRange+width, ".ts")
	if err != nil {
		t.Fatalf("without a NodeArray: got %v, want it lowered", err)
	}
	if made, _ := helpers(plain); made != 0 {
		t.Errorf("without a NodeArray: %d range_field_read helpers, want none", made)
	}
}

// checkRangeUnionReads stops an unchecked narrow to an object of a local the checker types as a range type,
// passes one of a local typed otherwise, and fails closed on a local the IR keeps no type for and nothing
// writes. Each lowering is built by hand over one program's checker, its IR a single narrow.
func TestRangeUnionNarrowStops(t *testing.T) {
	t.Parallel()
	program, checked, parameters := rangeProgram(t, rangeTextRange+rangeNodeArray+"export function width(range: TextRange, node: Node): number { return node.kind + range.end - range.pos; }\n")
	narrowed := func(typed *checker.Type) error {
		l := &lowering{program: program, checker: checked, result: &ir.Program{Source: "main.ts"}}
		l.result.Locals = []ir.Local{{Name: "held", Type: ir.Union, Function: -1}}
		if typed != nil {
			l.localTypes = map[int]*checker.Type{0: typed}
		}
		l.result.Main = []ir.Statement{ir.Evaluate{Value: ir.Narrow{Value: ir.Read{Local: 0, Of: ir.Union}, To: ir.Object}}}
		if held, known := l.representation(checked.GetTypeAtLocation(parameters["range"])); !known || held != ir.Union {
			t.Fatalf("TextRange is held as %v, want a union", held)
		}
		return l.checkRangeUnionReads()
	}
	stops := func(name string, err error) {
		var notYet *NotYet
		if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "a narrow of a union taken unasked in the program's body, whose operand may be a range type's value") {
			t.Errorf("%s: got %v, want the narrow's stop", name, err)
		}
	}
	stops("a range local", narrowed(checked.GetTypeAtLocation(parameters["range"])))
	if err := narrowed(checked.GetTypeAtLocation(parameters["node"])); err != nil {
		t.Errorf("a Node local: got %v, want nil", err)
	}
	stops("an untyped local nothing writes", narrowed(nil))
}

// A NodeArray's in test, with the bit on, lowers because the in test asks the kind, never because its boxed
// operand is followed: a box of an array still counts as a range value to every raw-layout reader (Lower's
// condition on #hae423v). node-array-keys.a's 'pos' in nodes is this shape. The program also reads a TextRange,
// so a type is held as a range Union and checkRangeUnionReads runs; with none it reads nothing. Mutant M5 in
// scratch-ha1, the in test classed as stopping again, stops it at the in test.
func TestRangeUnionInTestOfANodeArrayLowers(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, rangeTextRange+rangeNodeArray+"export function width(range: TextRange): number { return range.end - range.pos; }\nexport function has(nodes: NodeArray<Node>): boolean { return 'pos' in nodes; }\nconsole.log('ok');\n", ".ts"); err != nil {
		t.Fatalf("got %v, want a NodeArray's in test lowered", err)
	}
}
