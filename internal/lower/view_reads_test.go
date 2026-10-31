package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Every IR expression that holds an operand is one of viewRead's three classes, and
// TestEveryOperandIsClassedForViews finds them in internal/ir's source rather than among values some
// program made, as internal/flow's writes_test.go does for writes (#8krmqk4), so an expression added
// to the IR fails it until someone says which. A class can turn on the value: a field read checks the
// view under a checked receiver and is NotYet under any other, and an object literal copies whole only
// with a spread. Each example is that value, and viewReadOf must give it its class.

// viewReadProgram checks text under receiver type 1, so a Property's class turns on its receiver.
var viewReadProgram = &ir.Program{CheckedFields: ir.SortedMap[string, bool]{ir.CheckedFieldKey(1, "text"): true}}

type viewReadExample struct {
	expression ir.Expression
	why        string
}

// viewChecks read each field they read through a check.
var viewChecks = []viewReadExample{
	{ir.Property{ViewReceiverTypeID: 1, Name: "text"}, "a field a view checks under its receiver's type: emitted through the view's check (checkLazyViewReads, unemittableCheckedViewRead)"},
	{ir.CheckedCast{}, "reads the discriminant through the cast's own check"},
	{ir.ObjectCall{Method: "values", Checked: true}, "checks each value against the declared element at run time"},
	{ir.ObjectCall{Method: "entries", Checked: true}, "checks each value against the declared element at run time"},
	{ir.DynamicProperty{}, "a dynamic read of an unknown or a Function, by the allocation's own runtime shape, never a type's claim"},
}

// viewNotYet read fields raw, and are NotYet where their operand may hold a viewed value.
var viewNotYet = []viewReadExample{
	{ir.Property{ViewReceiverTypeID: 2, Name: "text"}, "a field read through a receiver type no view checks it under (checkViewWidenedReads)"},
	{ir.ObjectLiteral{Spread: ir.Read{}}, "copies its spread source's fields whole (checkViewCopies)"},
	{ir.ObjectCall{Method: "assign"}, "copies each source's fields whole (checkViewCopies)"},
	{ir.JSONEncode{}, "reads every field of the value and of what it holds, by its schema (checkViewWidenedReads' whole-value readers)"},
}

// viewNoField read no field of an object operand, each with why.
var viewNoField = []viewReadExample{
	{ir.ArrayConcat{}, "a new array of the arrays' elements, moved"},
	{ir.ArrayFill{}, "writes a value into an array's elements"},
	{ir.ArrayFrom{}, "a new array from a length and a callback, whose reads are classed where they are"},
	{ir.ArrayIndex{}, "reads an array's element, which is no field; a tuple's is a Property"},
	{ir.ArrayIsArray{}, "tests a value's brand"},
	{ir.ArrayJoin{}, "joins scalar elements; an array of objects is NotYet before it lowers"},
	{ir.ArrayLiteral{}, "a new array of the elements it's given, moved"},
	{ir.ArrayMap{}, "calls back with the elements, moved; the callback's reads are classed where they are"},
	{ir.ArrayPop{}, "moves an element out"},
	{ir.ArrayPush{}, "moves a value in"},
	{ir.ArrayReduce{}, "calls back with the elements, moved"},
	{ir.ArrayReverse{}, "moves the elements"},
	{ir.ArraySearch{}, "compares elements by identity or SameValueZero"},
	{ir.ArraySlice{}, "a new array of some of the elements, moved"},
	{ir.ArraySort{}, "calls back with the elements, moved"},
	{ir.ArraySplice{}, "moves elements out and in"},
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
	{ir.CollectionIterator{}, "an iterator over a Map or a Set, which holds no field a view checks"},
	{ir.Comma{}, "its operands, in order"},
	{ir.Concat{}, "a new string"},
	{ir.Conditional{}, "one of its branches"},
	{ir.Defined{}, "the value itself, checked for undefined"},
	{ir.Effects{}, "runs statements; a field read in them is a Property, classed on its own"},
	{ir.FileStatus{}, "reads a path from the host"},
	{ir.HasAccessor{}, "tests for an accessor"},
	{ir.HasOwn{}, "tests for an own field without reading its value"},
	{ir.HasProperty{}, "tests for a field without reading its value"},
	{ir.InstanceOf{}, "tests a class chain"},
	{ir.IsNull{}, "tests a reference"},
	{ir.IsUndefined{}, "tests a reference"},
	{ir.JSONDecode{}, "parses a string"},
	{ir.JSONStringify{}, "a scalar or an array of scalars; an object reference is NotYet before it lowers"},
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
	{ir.MapNew{}, "a new map from pairs, each pair's key and value moved"},
	{ir.MapSet{}, "moves a value in"},
	{ir.MapSize{}, "a map's size"},
	{ir.MapValues{}, "a new array of a map's values, moved"},
	{ir.MathCall{}, "a function of numbers"},
	{ir.MaybeOf{}, "a number or a boolean that may be missing"},
	{ir.MaybeToString{}, "a new string"},
	{ir.Narrow{}, "the value itself, as one member of its union"},
	{ir.NodeBufferCall{}, "a host operation on buffers, strings and numbers"},
	{ir.NodeFSFile{}, "a host operation on strings, numbers and host-made handles, which no view admits"},
	{ir.NodeHostCall{}, "a host operation on strings and numbers"},
	{ir.NumberCall{}, "a function of numbers and strings"},
	{ir.NumberFormat{}, "a new string"},
	{ir.NumberToString{}, "a new string"},
	{ir.ObjectCall{Method: "keys"}, "an object's own keys, no field's value"},
	{ir.ObjectCall{Method: "hasOwn"}, "tests for an own field without reading its value"},
	{ir.ObjectCall{Method: "is"}, "compares by SameValue"},
	{ir.ObjectCall{Method: "isFrozen"}, "tests a flag"},
	{ir.ObjectCall{Method: "freeze"}, "sets a flag on a shape a literal proves"},
	{ir.ObjectCall{Method: "values"}, "unchecked only for a literal's own const binding (enumerationProven), whose fields no cast claimed"},
	{ir.ObjectKeys{}, "an object's own keys, no field's value"},
	{ir.ObjectLiteral{}, "a new object of the values it's given, moved"},
	{ir.ParallelMap{}, "moves items to workers, whose reads are classed where they are"},
	{ir.PhantomMember{}, "a member read from a primitive"},
	{ir.PlaceholderUse{}, "the placeholder's value itself, once its readiness flag is checked (slice 5)"},
	{ir.ProcessCall{}, "a host operation on the prelude's process"},
	{ir.PromiseValue{}, "the value itself, settled"},
	{ir.ReadDirectory{}, "reads a path from the host"},
	{ir.ReadTextFile{}, "reads a path from the host"},
	{ir.RecordCall{}, "a record's own entries; a fixed object is no record"},
	{ir.RecordCoalesce{}, "a record's own entries"},
	{ir.RecordLiteral{}, "a record's own entries; a fixed object spread into one is NotYet before it lowers"},
	{ir.RegExpCall{}, "a pattern on a string"},
	{ir.RegExpGroup{}, "a match's named group, which no view claims"},
	{ir.RegExpNew{}, "a new pattern"},
	{ir.RegExpProperty{}, "a match's own field, which no view claims"},
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
	{ir.TypeOf{}, "tests a value"},
	{ir.TypedArrayFill{}, "writes a typed array"},
	{ir.TypedArrayNew{}, "a new typed array"},
	{ir.TypedArraySet{}, "writes a typed array"},
	{ir.TypedArraySort{}, "sorts a typed array"},
	{ir.TypedArraySubarray{}, "a view of a typed array"},
	{ir.Unary{}, "an operator on a number or a boolean"},
	{ir.UnionToString{}, "a scalar union's spelling; a union with an object in it is NotYet before it lowers"},
	{ir.Unwrap{}, "the value itself, checked present"},
	{ir.Utf8At{}, "reads a string"},
	{ir.Utf8Length{}, "reads a string"},
	{ir.Void{}, "evaluates its operand and reads nothing of it"},
	{ir.WeakOf{}, "the target's weak handle"},
	{ir.WeakTarget{}, "reads a weak reference"},
	{ir.WriteTextFile{}, "writes a string to the host"},
}

func TestEveryOperandIsClassedForViews(t *testing.T) {
	t.Parallel()
	holders := irOperandHolders(t)
	name := func(expression ir.Expression) string { return reflect.TypeOf(expression).Name() }
	classed := map[string]bool{}
	for _, list := range []struct {
		class    viewRead
		named    string
		examples []viewReadExample
	}{
		{viewReadChecks, "checks the view", viewChecks},
		{viewReadsNoField, "reads no field", viewNoField},
		{viewReadNotYet, "NotYet on viewed reach", viewNotYet},
	} {
		for _, example := range list.examples {
			classed[name(example.expression)] = true
			if strings.TrimSpace(example.why) == "" {
				t.Errorf("ir.%s is listed as %s without saying why", name(example.expression), list.named)
			}
			if got := viewReadOf(viewReadProgram, example.expression); got != list.class {
				t.Errorf("viewReadOf(ir.%s %+v) = %d, want %s", name(example.expression), example.expression, got, list.named)
			}
		}
	}
	expressions := slices.Sorted(maps.Keys(holders))
	for _, expression := range expressions {
		if holders[expression] && !classed[expression] {
			t.Errorf("ir.%s holds an operand and has no class for checked views: give it one in viewReadOf (view_reads.go) and list it in view_reads_test.go with why", expression)
		}
	}
	for listed := range classed {
		if _, ok := holders[listed]; !ok {
			t.Errorf("ir.%s is classed but isn't an expression in internal/ir", listed)
		}
	}
	operands := 0
	for _, holds := range holders {
		if holds {
			operands++
		}
	}
	t.Logf("%d expressions, %d with an operand: %d examples check the view, %d read no field, %d are NotYet on viewed reach", len(expressions), operands, len(viewChecks), len(viewNoField), len(viewNotYet))
}

// irOperandHolders is every type internal/ir declares that implements ir.Expression, from its source,
// and whether one of its fields can hold an operand: an Expression or a Statement, directly or in a
// slice, an array, a pointer or an ir struct that holds one. With Expression one method, Type() Type,
// that's every type with that method, as writes_test.go's irExpressions relies on.
func irOperandHolders(t *testing.T) map[string]bool {
	t.Helper()
	// Reads: internal/ir/*.go
	paths, err := filepath.Glob("../ir/*.go")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	var expressions []string
	structs := map[string][]ast.Expr{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(files, path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.TypeSpec:
				if structure, ok := node.Type.(*ast.StructType); ok {
					for _, field := range structure.Fields.List {
						if len(field.Names) == 0 {
							t.Fatalf("%s: ir.%s embeds a type, whose fields this enumeration can't see", files.Position(field.Pos()), node.Name.Name)
						}
						structs[node.Name.Name] = append(structs[node.Name.Name], field.Type)
					}
				}
			case *ast.FuncDecl:
				if node.Recv == nil || node.Name.Name != "Type" || node.Type.Params.NumFields() != 0 || node.Type.Results.NumFields() != 1 {
					return false
				}
				if result, ok := node.Type.Results.List[0].Type.(*ast.Ident); !ok || result.Name != "Type" {
					return false
				}
				identifier, ok := node.Recv.List[0].Type.(*ast.Ident)
				if !ok {
					t.Fatalf("%s: a Type method on %T: only a value receiver is enumerated", files.Position(node.Pos()), node.Recv.List[0].Type)
				}
				expressions = append(expressions, identifier.Name)
				return false
			}
			return true
		})
	}
	if len(expressions) < 100 {
		t.Fatalf("found %d expressions in ../ir: the enumeration no longer finds the IR", len(expressions))
	}
	var holds func(ast.Expr, map[string]bool) bool
	holds = func(held ast.Expr, seen map[string]bool) bool {
		switch held := held.(type) {
		case *ast.ArrayType:
			return holds(held.Elt, seen)
		case *ast.StarExpr:
			return holds(held.X, seen)
		case *ast.MapType:
			return holds(held.Key, seen) || holds(held.Value, seen)
		case *ast.Ident:
			if held.Name == "Expression" || held.Name == "Statement" {
				return true
			}
			if seen[held.Name] {
				return false
			}
			seen[held.Name] = true
			for _, field := range structs[held.Name] {
				if holds(field, seen) {
					return true
				}
			}
		}
		return false
	}
	result := map[string]bool{}
	for _, expression := range expressions {
		holder := false
		for _, field := range structs[expression] {
			holder = holder || holds(field, map[string]bool{expression: true})
		}
		result[expression] = holder
	}
	return result
}
