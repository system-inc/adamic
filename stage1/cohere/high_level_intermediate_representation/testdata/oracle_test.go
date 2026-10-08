//go:build lintoracle

// Overlay beside the Go package. No file in cohere is changed.
package high_level_intermediate_representation

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/system-inc/cohere/static_single_assignment"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
	"unicode/utf16"
)

func oraclePlace(p Place) string {
	return fmt.Sprintf("%d:%s:%d:%d:%d", p.Identifier, p.Effect, oracleBool(p.Reactive), p.Range.Pos(), p.Range.End())
}
func oracleBool(b bool) int {
	if b {
		return 1
	}
	return 0
}
func oracleDump(f *Function) string {
	var out strings.Builder
	fmt.Fprintf(&out, "hir-v1\nfunction %s %s entry=%d async=%d generator=%d\n", f.Name, f.Kind, f.Entry, oracleBool(f.IsAsync), oracleBool(f.IsGenerator))
	for _, row := range []struct {
		name   string
		places []Place
	}{{"params", f.Params}, {"context", f.Context}} {
		places := []string{}
		for _, p := range row.places {
			places = append(places, oraclePlace(p))
		}
		fmt.Fprintf(&out, "%s %s\n", row.name, strings.Join(places, ","))
	}
	fmt.Fprintf(&out, "returns %s\n", oraclePlace(f.Returns))
	for _, id := range f.Identifiers {
		fmt.Fprintf(&out, "identifier %d %d %s\n", id.Id, id.Declaration, id.Name)
	}
	seen := map[InstructionId]bool{}
	for _, b := range f.Blocks {
		preds := []string{}
		for _, p := range b.Predecessors {
			preds = append(preds, fmt.Sprint(p))
		}
		fmt.Fprintf(&out, "block %d %s predecessors=%s\n", b.Id, b.Kind, strings.Join(preds, ","))
		for _, phi := range b.Phis {
			fmt.Fprintf(&out, "phi %s", oraclePlace(phi.Place))
			for _, operand := range phi.Operands {
				fmt.Fprintf(&out, " %d=%s", operand.Predecessor, oraclePlace(operand.Place))
			}
			out.WriteByte('\n')
		}
		for _, id := range b.Instructions {
			i := f.Instruction(id)
			seen[i.Id] = true
			out.WriteString(oracleInstruction(i))
		}
		if terminal, ok := b.Terminal.(*Return); ok {
			fmt.Fprintf(&out, "terminal %d Return %s\n", terminal.Order, oraclePlace(terminal.Value))
		} else {
			fmt.Fprintf(&out, "terminal %d %s %s\n", TerminalOrder(b.Terminal), reflect.TypeOf(b.Terminal).Elem().Name(), oraclePayload(b.Terminal))
		}
	}
	for _, instruction := range f.Instructions {
		if !seen[instruction.Id] {
			out.WriteString("orphan ")
			out.WriteString(oracleInstruction(instruction))
		}
	}
	if len(f.ContextDeclarations) > 0 {
		ids := []int{}
		for id, yes := range f.ContextDeclarations {
			if yes {
				ids = append(ids, int(id))
			}
		}
		sort.Ints(ids)
		fmt.Fprintf(&out, "context-declarations %v\n", ids)
	}
	if len(f.Outlined) > 0 {
		ids := []int{}
		for id := range f.Outlined {
			ids = append(ids, int(id))
		}
		sort.Ints(ids)
		for _, id := range ids {
			fmt.Fprintf(&out, "outlined %d %d\n", id, f.Outlined[static_single_assignment.IdentifierId(id)])
		}
	}
	for index, nested := range f.Functions {
		fmt.Fprintf(&out, "nested %d\n%s", index, oracleDump(nested))
	}
	out.WriteString("scopes -\nend\n")
	return out.String()
}
func TestStage1HIRDump(t *testing.T) {
	input, err := os.ReadFile(os.Getenv("HIR_CORPUS"))
	if err != nil {
		t.Fatal(err)
	}
	var output strings.Builder
	for _, code := range strings.Split(string(input), "\n") {
		if code == "" {
			continue
		}
		source := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/test.tsx", PathKey: "/test.tsx"}, code, core.ScriptKindTSX)
		var node *ast.Node
		source.AsNode().ForEachChild(func(n *ast.Node) bool {
			if ast.IsFunctionLike(n) {
				node = n
				return true
			}
			return false
		})
		if node == nil {
			t.Fatalf("no function: %s", code)
		}
		f := Lower(node, nil)
		Construct(f)
		output.WriteString(oracleDump(f))
	}
	if err := os.WriteFile(os.Getenv("HIR_OUTPUT"), []byte(output.String()), 0600); err != nil {
		t.Fatal(err)
	}
}

// Fixed UTF-16 escape alphabet shared with the stage 1 parser's written().
func oracleText(text string) string {
	var out strings.Builder
	for _, code := range utf16.Encode([]rune(text)) {
		if code >= 32 && code <= 126 && code != 92 {
			out.WriteRune(rune(code))
		} else {
			fmt.Fprintf(&out, "\\u%04x", code)
		}
	}
	return out.String()
}

// All currently declared payload variants are observed, including places in patterns.
// AST nodes are represented by stable kind/span/text handles, never object addresses.
func oracleValue(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		return oracleValue(v.Elem())
	}
	if v.CanInterface() {
		switch x := v.Interface().(type) {
		case *StartMemoize:
			var deps any
			if x.Deps != nil {
				deps = oracleValue(reflect.ValueOf(x.Deps))
			}
			return map[string]any{"ManualMemoId": x.ManualMemoId, "Deps": deps}
		case Place:
			return oraclePlace(x)
		case core.TextRange:
			return fmt.Sprintf("%d:%d", x.Pos(), x.End())
		case *ast.Node:
			if x == nil {
				return nil
			}
			return map[string]any{"kind": x.Kind.String(), "pos": x.Pos(), "end": x.End()}
		}
	}
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil
		}
		return oracleValue(v.Elem())
	}
	switch v.Kind() {
	case reflect.Struct:
		fields := map[string]any{}
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if field.PkgPath != "" || field.Name == "Order" {
				continue
			}
			fields[field.Name] = oracleValue(v.Field(i))
		}
		return fields
	case reflect.Slice, reflect.Array:
		values := []any{}
		for i := 0; i < v.Len(); i++ {
			values = append(values, oracleValue(v.Index(i)))
		}
		return values
	case reflect.Map:
		fields := map[string]any{}
		for _, key := range v.MapKeys() {
			fields[fmt.Sprint(key.Interface())] = oracleValue(v.MapIndex(key))
		}
		return fields
	default:
		return v.Interface()
	}
}
func oraclePayload(value any) string {
	data, err := json.Marshal(oracleValue(reflect.ValueOf(value)))
	if err != nil {
		panic(err)
	}
	return string(data)
}

func oracleInstruction(i *Instruction) string {
	var out strings.Builder
	kind, payload := "", ""
	switch v := i.Value.(type) {
	case *Primitive:
		kind = "Primitive"
		switch x := v.Value.(type) {
		case nil:
			payload = "nil"
		case string:
			payload = "string:" + oracleText(x)
		case bool:
			payload = fmt.Sprintf("bool:%t", x)
		default:
			panic("literal outside slice")
		}
	case *LoadLocal:
		kind = "LoadLocal"
		payload = oraclePlace(v.Place)
	case *UnaryExpression:
		kind = "UnaryExpression"
		payload = v.Operator + " " + oraclePlace(v.Value)
	case *BinaryExpression:
		kind = "BinaryExpression"
		payload = oraclePlace(v.Left) + " " + v.Operator + " " + oraclePlace(v.Right)
	default:
		kind = reflect.TypeOf(v).Elem().Name()
		payload = oraclePayload(v)
	}
	fmt.Fprintf(&out, "instruction %d %d %s %d:%d %s %s\n", i.Id, i.Order, oraclePlace(i.LValue), i.Range.Pos(), i.Range.End(), kind, payload)
	return out.String()
}

func TestStage1CentralInstructionVariants(t *testing.T) {
	destination := os.Getenv("HIR_CENSUS")
	if destination == "" {
		t.Skip("census exporter only")
	}
	p := Place{}
	values := []InstructionValue{
		&DeclareContext{LValue: p, Kind: 0},
		&StartMemoize{ManualMemoId: 1},
		&StartMemoize{ManualMemoId: 2, Deps: []ManualMemoDependency{}},
		&StartMemoize{ManualMemoId: 3, Deps: []ManualMemoDependency{{Root: ManualMemoRoot{Place: p}, Path: []DependencyPathEntry{{Property: "x", Optional: true}}}, {Root: ManualMemoRoot{IsGlobal: true, Place: p, Name: "React"}, Path: []DependencyPathEntry{}}}},
		&FinishMemoize{ManualMemoId: 3, Value: p},
		&FinishMemoize{ManualMemoId: 2, Value: p, Pruned: true},
	}
	var out strings.Builder
	for id, value := range values {
		out.WriteString(oracleInstruction(&Instruction{Id: InstructionId(id), LValue: p, Value: value}))
	}
	if err := os.MkdirAll(destination, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination+"/central-instructions.dump", []byte(out.String()), 0600); err != nil {
		t.Fatal(err)
	}
}
