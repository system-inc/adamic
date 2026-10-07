// Owned overlay in the production React package, calling unmodified helpers.
package react

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	hir "github.com/system-inc/cohere/internal/lint/ecmascript/high_level_intermediate_representation"
	"os"
	"strings"
	"testing"
)

type wave04EnvBinding struct {
	Name        string
	Declaration int
	Present     bool
}
type wave04EnvOp struct {
	Kind       string
	Key, Value int
	Text       string
}

func TestWave04RefsEnvironment(t *testing.T) {
	// Not parallel: this isolated oracle writes to the caller's single output path.
	bindings := []wave04EnvBinding{{"ref", 0, true}, {"", 0, true}, {"current", 7, true}, {"current", 7, true}, {"other", 8, true}, {"", 0, false}, {"other", 8, true}, {"", 0, true}}
	scenarios := [][]wave04EnvOp{
		{{Kind: "set", Key: 0, Value: 0}, {Kind: "reset"}, {Kind: "note", Key: 0}, {Kind: "set", Key: 0, Value: 1}, {Kind: "reset"}, {Kind: "set", Key: 0, Value: 2}, {Kind: "reset"}, {Kind: "set", Key: 0, Value: 0}, {Kind: "mint"}},
		{{Kind: "note", Key: 2}, {Kind: "note", Key: 3}, {Kind: "set", Key: 2, Value: 1}, {Kind: "reset"}, {Kind: "set", Key: 3, Value: 3}, {Kind: "define", Key: 3, Value: 7}, {Kind: "set", Key: 3, Value: 1}, {Kind: "note", Key: 4}, {Kind: "note", Key: 6}, {Kind: "define", Key: 6, Value: 4}, {Kind: "set", Key: 6, Value: 3}},
		{{Kind: "access", Key: 0, Value: 0}, {Kind: "define", Key: 1, Value: 0}, {Kind: "define", Key: 7, Value: 1}, {Kind: "access", Key: 1, Value: 1}, {Kind: "define", Key: 2, Value: 1}, {Kind: "carry", Key: 3, Value: 1}, {Kind: "access", Key: 4, Value: 1}, {Kind: "define", Key: 4, Value: 0}, {Kind: "carry", Key: 6, Value: 4}, {Kind: "define", Key: 0, Value: 5}, {Kind: "carry", Key: 8, Value: 1}},
		{{Kind: "name", Key: 7, Value: 0}, {Kind: "name", Key: 7, Value: 1}, {Kind: "name2", Key: 0, Text: "alias"}, {Kind: "name2", Key: 1, Text: "alias"}, {Kind: "name2", Key: 1, Text: ""}, {Kind: "property", Key: 1, Text: "ref"}, {Kind: "property", Key: 1, Text: ""}, {Kind: "name", Key: 8, Value: 5}, {Kind: "note", Key: 5}, {Kind: "note", Key: 9}},
	}
	function := &hir.Function{}
	for _, b := range bindings {
		if b.Present {
			function.Identifiers = append(function.Identifiers, &hir.Identifier{Name: b.Name, Declaration: hir.DeclarationId(b.Declaration)})
		} else {
			function.Identifiers = append(function.Identifiers, nil)
		}
	}
	values := []*refsAccessType{{Kind: refsNone}, {Kind: refsRef, RefId: 11, HasRefId: true}, {Kind: refsRef, RefId: 12, HasRefId: true}, {Kind: refsRefValue, RefId: 11, HasRefId: true, Span: 9, HasSpan: true}}
	nodes := []*ast.Node{{Kind: ast.KindIdentifier}, {Kind: ast.KindIdentifier}}
	signature := func(v *refsAccessType) string {
		if v == nil {
			return "nil"
		}
		return fmt.Sprintf("%d:%d:%t:%d:%t:%d:%t:(nil):(nil)", v.Kind, v.RefId, v.HasRefId, v.Span, v.HasSpan, v.RefSpan, v.HasRefSpan)
	}
	var out strings.Builder
	for si, ops := range scenarios {
		env := newRefsEnvironment()
		for step, op := range ops {
			k, v := hir.IdentifierId(op.Key), hir.IdentifierId(op.Value)
			switch op.Kind {
			case "set":
				env.set(k, values[op.Value])
			case "reset":
				env.changed = false
			case "note":
				env.noteDeclaration(function, k)
			case "define":
				env.define(k, v)
			case "access":
				env.accessNodes[k] = nodes[op.Value]
			case "carry":
				env.carryAccessNode(k, v)
			case "name":
				env.setName(function, k, v)
			case "name2":
				env.setName2(k, op.Text)
			case "property":
				env.setPropertyName(k, op.Text)
			case "mint":
				env.nextRefId()
			default:
				t.Fatal(op.Kind)
			}
			fmt.Fprintf(&out, "step %d %d %t %d\n", si, step, env.changed, env.refIdSeed)
			for id := 0; id < 10; id++ {
				key := hir.IdentifierId(id)
				node := -1
				if env.accessNodes[key] == nodes[0] {
					node = 0
				} else if env.accessNodes[key] == nodes[1] {
					node = 1
				}
				fmt.Fprintf(&out, "key %d %d %s %s %s %d\n", id, env.operandId(key), signature(env.get(key)), env.nameOf(function, key), env.propertyNames[key], node)
			}
		}
	}
	data, err := json.Marshal(struct {
		Bindings  []wave04EnvBinding
		Scenarios [][]wave04EnvOp
		Expected  string
	}{bindings, scenarios, out.String()})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("WAVE04_REFS_ENV"), data, 0600); err != nil {
		t.Fatal(err)
	}
}
