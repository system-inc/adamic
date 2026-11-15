package lower

import (
	"fmt"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// Scratch probe, never lands: for each road P row, every h.m read's parent chain and what
// presenceConsumer and truthinessOnly answer for it.
func TestRoadPProbe(t *testing.T) {
	bodies := map[string]string{
		"if":       "\tif (flag && h.m) {\n\t\treturn 'has';\n\t}\n\treturn 'none';",
		"not":      "\treturn String(!(flag && h.m));",
		"while":    "\twhile (flag && h.m) {\n\t\treturn 'has';\n\t}\n\treturn 'none';",
		"nested-r": "\tif (flag && (other && h.m)) {\n\t\treturn 'has';\n\t}\n\treturn 'none';",
	}
	for name, body := range bodies {
		for _, form := range roadPForms {
			for _, extension := range []string{".a", ".ts"} {
				path := writeAgreementSource(t, "main"+extension, roadPProgram(form.declared, body))
				loaded, err := load.Load([]string{path})
				if err != nil {
					fmt.Printf("PROBE %s|%s|%s load: %v\n", name, form.name, extension, err)
					continue
				}
				for _, file := range loaded.Files() {
					var visit func(*ast.Node) bool
					visit = func(node *ast.Node) bool {
						if node.Kind == ast.KindPropertyAccessExpression && node.Name().Text() == "m" {
							chain := ""
							for p := node.Parent; p != nil && p.Kind != ast.KindSourceFile; p = p.Parent {
								chain += p.Kind.String() + ">"
								if len(chain) > 200 {
									break
								}
							}
							fmt.Printf("PROBE %s|%s|%s presenceConsumer=%v truthinessOnly=%v chain=%s\n", name, form.name, extension, presenceConsumer(node), truthinessOnly(node), chain)
						}
						node.ForEachChild(visit)
						return false
					}
					file.AsNode().ForEachChild(visit)
				}
			}
		}
	}
}
