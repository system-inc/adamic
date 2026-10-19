package flow

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

// Every IR expression is one of these five to the effect inference, and TestEveryExpressionIsClassified
// finds them in internal/ir's source rather than among values some program made, so an expression
// added to the IR fails it until someone says which. A node that isn't one of the first four, and
// whose value can't be mutated, gives nothing (value's last lines), so a write whose result can't be
// mutated (map.clear() is void) reaches no range unless it's one of the first four.

// writeExamples are what writes reports, each as a value it reports.
var writeExamples = []ir.Expression{
	ir.ArraySplice{}, ir.ArrayFill{}, ir.ArraySort{}, ir.MapSet{}, ir.MapDelete{}, ir.MapClear{}, ir.SetAdd{},
	ir.NodeBufferCall{Function: "buffer_set"}, ir.NodeFSFile{Operation: "read_sync"}, ir.ObjectCall{Method: "assign"},
	ir.RegExpCall{Method: "test"},
	ir.NodeFSFile{Operation: "host_env_set"}, ir.NodeFSFile{Operation: "host_env_delete"},
}

// callBackExamples are what callsBack reports, each as a value it reports.
var callBackExamples = []ir.Expression{
	ir.ArrayMap{}, ir.ArrayVisit{}, ir.ArrayReduce{}, ir.ArrayFrom{}, ir.ArraySort{}, ir.MapForEach{}, ir.ParallelMap{},
	ir.RegExpCall{Replacement: &ir.RegExpReplacement{}},
	ir.JSONEncode{}, // Fail closed: lowering refuses getter-bearing operands before JSON encoding.
}

// valueCases write into what they're handed, or call, through a case of value's own, which makes
// the store or the call itself.
var valueCases = []ir.Expression{
	ir.TypedArraySort{}, ir.TypedArrayFill{}, ir.TypedArraySet{}, ir.RecordCoalesce{}, ir.RecordCall{}, ir.ArrayPush{},
	ir.ArrayPop{}, ir.ArrayReverse{}, ir.Call{}, ir.CallClosure{},
}

// holdsStatements write through the statements they hold, which operands makes as run1 makes a
// statement, a write into a field or an element being a store.
var holdsStatements = []ir.Expression{ir.Effects{}, ir.LogicalAssignment{}}

// writesNothing write into nothing they're handed and call nothing, each with why.
var writesNothing = []struct {
	expression ir.Expression
	why        string
}{
	{ir.ProcessCall{}, "host operation does not write a program value or call a callback"},
	{ir.NodeHostCall{}, "owned host result; may share argument string storage through a counted share"},
	{ir.ArrayConcat{}, "a new array of the arrays' elements"},
	{ir.ArrayIndex{}, "reads an element"},
	{ir.ArrayIsArray{}, "tests a value's brand"},
	{ir.ArrayJoin{}, "reads the elements into a new string"},
	{ir.ArrayLiteral{}, "a new array"},
	{ir.ArraySearch{}, "reads the elements"},
	{ir.ArraySlice{}, "a new array of some of the elements"},
	{ir.Await{}, "a suspension, which Build makes a Suspend terminal, where run mutates every escaped value"},
	{ir.Binary{}, "an operator on numbers, booleans or strings"},
	{ir.BooleanConstant{}, "a constant"},
	{ir.BooleanToString{}, "a new string"},
	{ir.Box{}, "the value itself, where a union goes"},
	{ir.CharCodeAt{}, "reads a string"},
	{ir.CheckedCast{}, "the value itself, its discriminant checked"},
	{ir.ClosureSelf{}, "the running closure itself"},
	{ir.Coalesce{}, "one of its operands"},
	{ir.CodePoints{}, "a new array of a string's code points"},
	{ir.CollectionIterator{}, "a new iterator holding the collection, whose next is called as a function value"},
	{ir.Comma{}, "its operands, in order"},
	{ir.Concat{}, "a new string"},
	{ir.Conditional{}, "one of its branches"},
	{ir.Defined{}, "the value itself, checked for undefined"},
	{ir.DynamicProperty{}, "reads a field"},
	{ir.FileStatus{}, "reads a path from the host"},
	{ir.HasAccessor{}, "tests for an accessor"},
	{ir.HasOwn{}, "tests for an own field"},
	{ir.HasProperty{}, "tests for a field without reading it"},
	{ir.InstanceOf{}, "tests a class chain"},
	{ir.IsNull{}, "tests a reference"},
	{ir.IsUndefined{}, "tests a reference"},
	{ir.JSONDecode{}, "parses a string into a new value without calling user code"},
	{ir.JSONNull{}, "a constant inside stringify's input"},
	{ir.JSONStringify{}, "reads a value into a new string; lowering refuses a replacer function"},
	{ir.Length{}, "reads an array's length"},
	{ir.LibraryGlobal{}, "a built-in's identity"},
	{ir.Logical{}, "one of its operands"},
	{ir.MakeClosure{}, "a new function value; a call of it is CallClosure"},
	{ir.MakeError{}, "a new Error"},
	{ir.MapEntries{}, "a new array of a map's entries"},
	{ir.MapGet{}, "reads an entry"},
	{ir.MapHas{}, "tests for an entry"},
	{ir.MapKeys{}, "a new array of a map's keys"},
	{ir.MapNew{}, "a new map, reading the pairs it's made from"},
	{ir.MapSize{}, "reads a map's size"},
	{ir.MapValues{}, "a new array of a map's values"},
	{ir.MathCall{}, "a function of numbers"},
	{ir.MaybeOf{}, "a number or a boolean that may be missing"},
	{ir.MaybeToString{}, "a new string"},
	{ir.Narrow{}, "the value itself, as one member of its union"},
	{ir.Null{}, "a constant"},
	{ir.NumberCall{}, "a function of numbers and strings"},
	{ir.NumberConstant{}, "a constant"},
	{ir.NumberFormat{}, "a new string"},
	{ir.NumberToString{}, "a new string"},
	{ir.ObjectKeys{}, "a new array of an object's keys"},
	{ir.ObjectLiteral{}, "a new object"},
	{ir.PhantomMember{}, "undefined read from a primitive, or a TypeError made when the value is missing"},
	{ir.ProgramArguments{}, "a new array of the arguments"},
	{ir.PromiseValue{}, "a new settled promise"},
	{ir.Property{}, "reads a field; a getter is a Call"},
	{ir.Read{}, "reads a variable"},
	{ir.ReadDirectory{}, "reads a path from the host"},
	{ir.ReadTextFile{}, "reads a path from the host"},
	{ir.RecordLiteral{}, "a new record"},
	{ir.RegExpGroup{}, "reads a match's named group"},
	{ir.RegExpNew{}, "a new pattern"},
	{ir.RegExpProperty{}, "reads a match's field"},
	{ir.SetNew{}, "a new set, reading the array it's made from"},
	{ir.SetValues{}, "a new array of a set's elements"},
	{ir.StringCall{}, "a string's method; a string can't be mutated"},
	{ir.StringConstant{}, "a constant"},
	{ir.StringFromCodes{}, "a new string"},
	{ir.StringIndex{}, "reads a string"},
	{ir.StringIterator{}, "a new iterator holding an immutable string"},
	{ir.StringLength{}, "reads a string"},
	{ir.ToFixed{}, "a new string"},
	{ir.Trim{}, "a new string"},
	{ir.Truthy{}, "tests a value"},
	{ir.TypedArrayNew{}, "a new typed array, reading what it's made from"},
	{ir.TypedArraySubarray{}, "a view of the array, which value makes the array itself"},
	{ir.TypeOf{}, "tests a value"},
	{ir.Unary{}, "an operator on a number or a boolean"},
	{ir.Undefined{}, "a constant"},
	{ir.UnionToString{}, "a new string"},
	{ir.Unwrap{}, "the value itself, checked present"},
	{ir.Utf8At{}, "reads a string"},
	{ir.Utf8Length{}, "reads a string"},
	{ir.Void{}, "evaluates its operand"},
	{ir.WeakOf{}, "the target's weak handle, which changes nothing the program reads of the target"},
	{ir.WeakTarget{}, "reads a weak reference"},
	{ir.WriteTextFile{}, "writes a file from a string"},
}

// regExpMethods is every Method lowering gives a RegExpCall (internal/lower/regexp.go,
// library_regexp_replace.go, object.go), with whether it moves a pattern's lastIndex. writes counts
// every method but the three it names, so a method added later is a write until someone says
// otherwise; this pins each one that's here now.
var regExpMethods = []struct {
	method string
	writes bool
}{
	{"exec", true}, {"test", true}, {"next", true}, {"match", true}, {"search", true},
	{"replace", true}, {"replaceAll", true},
	{"split", false}, {"matchAll", false}, {"iteratorDone", false},
}

func TestRegExpCallWritesByMethod(t *testing.T) {
	t.Parallel()
	for _, entry := range regExpMethods {
		if got := writes(ir.RegExpCall{Method: entry.method}); got != entry.writes {
			t.Errorf("writes(RegExpCall %s) = %t, want %t", entry.method, got, entry.writes)
		}
	}
}

func TestEveryExpressionIsClassified(t *testing.T) {
	t.Parallel()
	statements := irExpressions(t)
	expressions := slices.Sorted(maps.Keys(statements))
	cases := valueCaseNames(t)
	name := func(expression ir.Expression) string { return reflect.TypeOf(expression).Name() }
	classified := map[string]string{}
	classify := func(expression ir.Expression, as string) {
		if before, ok := classified[name(expression)]; ok && before != as && !(before == "writes" && as == "callsBack") {
			t.Errorf("ir.%s is both %s and %s", name(expression), before, as)
		}
		classified[name(expression)] = as
	}
	for _, expression := range writeExamples {
		classify(expression, "writes")
		if !writes(expression) {
			t.Errorf("writes doesn't report ir.%s %+v", name(expression), expression)
		}
	}
	for _, expression := range callBackExamples {
		classify(expression, "callsBack")
		if !callsBack(expression) {
			t.Errorf("callsBack doesn't report ir.%s %+v", name(expression), expression)
		}
	}
	for _, expression := range append(writeExamples, callBackExamples...) {
		if cases[name(expression)] {
			t.Errorf("ir.%s has a case of its own in value, which returns before writes or callsBack is asked", name(expression))
		}
	}
	for _, expression := range valueCases {
		classify(expression, "a case of value's own")
		if !cases[name(expression)] {
			t.Errorf("ir.%s is listed as handled by a case of value's own, and value has none", name(expression))
		}
	}
	for _, expression := range holdsStatements {
		classify(expression, "holding statements")
		if !statements[name(expression)] {
			t.Errorf("ir.%s is listed as holding statements, and holds none", name(expression))
		}
	}
	for _, entry := range writesNothing {
		classify(entry.expression, "writing nothing")
		if strings.TrimSpace(entry.why) == "" {
			t.Errorf("ir.%s is listed as writing nothing without saying why", name(entry.expression))
		}
		if writes(entry.expression) || callsBack(entry.expression) {
			t.Errorf("ir.%s is listed as writing nothing, and writes or callsBack reports it", name(entry.expression))
		}
	}
	for _, expression := range expressions {
		if _, ok := classified[expression]; !ok {
			t.Errorf("ir.%s isn't classified: add it to writes or callsBack in infer.go, give it a case in value that stores or calls, or list it in holdsStatements or in writesNothing with why (writes_test.go)", expression)
		}
	}
	for listed := range classified {
		if !slices.Contains(expressions, listed) {
			t.Errorf("ir.%s is classified but isn't an expression in internal/ir", listed)
		}
	}
	t.Logf("%d expressions: %d write, %d call back, %d have a case of value's own, %d hold statements, %d write nothing", len(expressions), len(writeExamples), len(callBackExamples), len(valueCases), len(holdsStatements), len(writesNothing))
}

// irExpressions is every type internal/ir declares that implements ir.Expression, from its source,
// and whether it has a field holding statements. With Expression one method, Type() Type, that's
// every type with that method, as long as no struct embeds a type (whose methods it would take on
// without declaring them).
func irExpressions(t *testing.T) map[string]bool {
	t.Helper()
	// Reads: internal/ir/*.go
	paths, err := filepath.Glob("../ir/*.go")
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	var expressions []string
	statements := map[string]bool{}
	interfaces := 0
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
				if node.Name.Name == "Expression" {
					interfaces++
					if !typeMethodOnly(node.Type) {
						t.Fatalf("%s: ir.Expression is no longer the one method Type() Type, which this enumeration relies on", files.Position(node.Pos()))
					}
				}
				if structure, ok := node.Type.(*ast.StructType); ok {
					for _, field := range structure.Fields.List {
						if len(field.Names) == 0 {
							t.Fatalf("%s: ir.%s embeds a type, whose methods this enumeration can't see", files.Position(field.Pos()), node.Name.Name)
						}
						held := field.Type
						if list, ok := held.(*ast.ArrayType); ok {
							held = list.Elt
						}
						if identifier, ok := held.(*ast.Ident); ok && identifier.Name == "Statement" {
							statements[node.Name.Name] = true
						}
					}
				}
			case *ast.FuncDecl:
				if node.Recv == nil || node.Name.Name != "Type" || !typeSignature(node.Type) {
					return false
				}
				receiver := node.Recv.List[0].Type
				identifier, ok := receiver.(*ast.Ident)
				if !ok {
					t.Fatalf("%s: a Type method on %T: only a value receiver is enumerated", files.Position(node.Pos()), receiver)
				}
				expressions = append(expressions, identifier.Name)
				return false
			}
			return true
		})
	}
	if interfaces != 1 || len(expressions) < 100 {
		t.Fatalf("found %d Expression interfaces and %d expressions in ../ir: the enumeration no longer finds the IR", interfaces, len(expressions))
	}
	result := map[string]bool{}
	for _, expression := range expressions {
		result[expression] = statements[expression]
	}
	return result
}

// typeMethodOnly reports whether an interface is exactly Type() Type.
func typeMethodOnly(expression ast.Expr) bool {
	methods, ok := expression.(*ast.InterfaceType)
	if !ok || len(methods.Methods.List) != 1 {
		return false
	}
	method := methods.Methods.List[0]
	signature, ok := method.Type.(*ast.FuncType)
	return ok && len(method.Names) == 1 && method.Names[0].Name == "Type" && typeSignature(signature)
}

// typeSignature reports whether a function takes nothing and gives one Type.
func typeSignature(signature *ast.FuncType) bool {
	if signature.Params.NumFields() != 0 || signature.Results.NumFields() != 1 {
		return false
	}
	result, ok := signature.Results.List[0].Type.(*ast.Ident)
	return ok && result.Name == "Type"
}

// valueCaseNames is every IR type the switch in inference.value gives a case of its own.
func valueCaseNames(t *testing.T) map[string]bool {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "infer.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv == nil || function.Name.Name != "value" {
			continue
		}
		for _, statement := range function.Body.List {
			typeSwitch, ok := statement.(*ast.TypeSwitchStmt)
			if !ok {
				continue
			}
			for _, clause := range typeSwitch.Body.List {
				for _, listed := range clause.(*ast.CaseClause).List {
					if selector, ok := listed.(*ast.SelectorExpr); ok {
						names[selector.Sel.Name] = true
					}
				}
			}
		}
	}
	if len(names) == 0 {
		t.Fatal("found no cases in inference.value's switch")
	}
	return names
}

func TestProcessEnvironmentWrites(t *testing.T) {
	t.Parallel()
	for _, one := range []struct {
		operation string
		want      bool
	}{
		{"host_env", false}, {"host_env_set", true}, {"host_env_delete", true},
	} {
		t.Run(one.operation, func(t *testing.T) {
			t.Parallel()
			if got := writes(ir.NodeFSFile{Operation: one.operation}); got != one.want {
				t.Fatalf("writes(%s)=%t, want %t", one.operation, got, one.want)
			}
		})
	}
}
