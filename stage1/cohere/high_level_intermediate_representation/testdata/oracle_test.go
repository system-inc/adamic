//go:build lintoracle

// Overlay beside the Go package. No file in cohere is changed.
package high_level_intermediate_representation

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"os"
	"strings"
	"testing"
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
	for _, b := range f.Blocks {
		preds := []string{}
		for _, p := range b.Predecessors {
			preds = append(preds, fmt.Sprint(p))
		}
		fmt.Fprintf(&out, "block %d %s predecessors=%s\n", b.Id, b.Kind, strings.Join(preds, ","))
		if len(b.Phis) > 0 {
			panic("slice cannot dump phi")
		}
		for _, id := range b.Instructions {
			i := f.Instruction(id)
			kind, payload := "", ""
			switch v := i.Value.(type) {
			case *Primitive:
				kind = "Primitive"
				switch x := v.Value.(type) {
				case nil:
					payload = "nil"
				case string:
					payload = "string:" + x
				case bool:
					payload = fmt.Sprintf("bool:%t", x)
				default:
					panic("literal outside slice")
				}
			case *LoadLocal:
				kind = "LoadLocal"
				payload = oraclePlace(v.Place)
			default:
				panic(fmt.Sprintf("instruction outside slice: %T", v))
			}
			fmt.Fprintf(&out, "instruction %d %d %s %d:%d %s %s\n", i.Id, i.Order, oraclePlace(i.LValue), i.Range.Pos(), i.Range.End(), kind, payload)
		}
		terminal, ok := b.Terminal.(*Return)
		if !ok {
			panic("terminal outside slice")
		}
		fmt.Fprintf(&out, "terminal %d Return %s\n", terminal.Order, oraclePlace(terminal.Value))
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
