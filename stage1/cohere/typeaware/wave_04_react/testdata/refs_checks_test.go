// Loaded by an owned overlay; all judgments come from production Go helpers.
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

func TestWave04RefsChecks(t *testing.T) {
	// Not parallel: the isolated oracle writes one caller-supplied output path.
	type value struct{ Kind, Value, Function int }
	rows := []value{{0, -1, -1}, {1, -1, -1}, {2, -1, -1}, {3, -1, -1}, {4, -1, -1}, {5, 4, -1}, {5, -1, 0}, {5, -1, 1}, {5, 5, -1}, {5, 3, -1}, {5, 0, 0}}
	values := make([]*refsAccessType, len(rows))
	for i, r := range rows {
		values[i] = &refsAccessType{Kind: refsAccessKind(r.Kind)}
	}
	values[3].RefId, values[3].HasRefId = 11, true
	values[4].RefId, values[4].HasRefId = 11, true
	values[4].Span, values[4].HasSpan = 12, true
	functions := []*refsFunctionType{{ReadRefEffect: true}, {ReturnType: values[4]}}
	for i, r := range rows {
		if r.Value >= 0 {
			values[i].Value = values[r.Value]
		}
		if r.Function >= 0 {
			values[i].Function = functions[r.Function]
		}
	}
	nodes := []*ast.Node{{Kind: ast.KindIdentifier}, {Kind: ast.KindIdentifier}}
	var out strings.Builder
	for index := -1; index < len(rows); index++ {
		env := newRefsEnvironment()
		if index >= 0 {
			env.data[7] = values[index]
		}
		env.accessNodes[7] = nodes[0]
		for _, seed := range []bool{false, true} {
			for _, method := range []string{"direct", "value", "passed", "update"} {
				var findings []refsFinding
				if seed {
					findings = append(findings, refsFinding{Kind: refsFindingFunctionAccessesRef, Value: 99})
				}
				switch method {
				case "direct":
					findings = refsCheckDirectValueAccess(env, 7, findings)
				case "value":
					findings = refsCheckValueAccess(env, 7, findings)
				case "passed":
					findings = refsCheckPassedToFunction(env, 7, findings)
				case "update":
					findings = refsCheckUpdate(env, 7, nodes[1], findings)
				}
				fmt.Fprintf(&out, "case %d %t %s %d\n", index, seed, method, len(findings))
				for _, f := range findings {
					node := -1
					if f.Node == nodes[0] {
						node = 0
					} else if f.Node == nodes[1] {
						node = 1
					}
					fmt.Fprintf(&out, "%d:%d:%t:%d\n", f.Kind, hir.IdentifierId(f.Value), f.HasValue, node)
				}
			}
		}
	}
	data, err := json.Marshal(struct {
		Values   []value
		Expected string
	}{rows, out.String()})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("WAVE04_REFS_CHECKS"), data, 0600); err != nil {
		t.Fatal(err)
	}
}
