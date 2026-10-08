package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: updating counts writes the lane's shared measured table.
// The original oracles own the behavior and mutants; these rows measure the
// same complete declaration witnesses without duplicating their certificates.
func TestCheckedViewObjectPrimitiveOverlapCounts(t *testing.T) {
	declarations, manifest := intersectionOriginalInputs(t)
	for _, test := range []struct {
		name, source string
		ids          []int
	}{
		{"bindable-static-good", "true\n", []int{9476, 9474}},
		{"bindable-access-good", "true\n", []int{9485, 9475}},
		{"jsdoc-parent-good", "true\n", []int{7642}},
		{"jsdoc-root-good", "true false\n", []int{36241}},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, _ := intersectionOriginalProgram(t, declarations, test.name, test.source)
			names := []string{"JSDoc", "Node"}
			if strings.HasPrefix(test.name, "bindable-") {
				names = []string{"Identifier", "Node", "Symbol", "PropertyAccessEntityNameExpression", "ElementAccessExpression"}
			}
			requireIntersectionOriginalComplete(t, program, manifest, names, test.ids...)
			objectPrimitiveOriginalCount(t, program, "lane7/"+test.name+".a")
		})
	}
	t.Run("jsdoc-union-parent-good", func(t *testing.T) {
		input, err := os.ReadFile("../../stage3/interface-downcasts/lane7/original/jsdoc-parent-good.a")
		if err != nil {
			t.Fatal(err)
		}
		source := strings.Replace(string(input), "JSDoc, SyntaxKind", "JSDoc, JSDocTypeLiteral, SyntaxKind", 1)
		source = strings.Replace(source, "jsDoc: JSDoc)", "jsDoc: JSDoc | JSDocTypeLiteral)", 1)
		source = strings.Replace(source, "kind: 243", "kind: 308", 1)
		source = strings.Replace(source, "parent.end > parent.pos", "parent.end > 0", 1)
		source = strings.Replace(source, "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
		path := filepath.Join(t.TempDir(), "jsdoc-union-parent-good.a")
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		if diff := disagreement(run{stdout: []byte("true\n")}, onNode(t, path)); diff != "" {
			t.Fatal("Node: " + diff)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		requireIntersectionOriginalComplete(t, program, manifest, []string{"Node", "JSDoc", "JSDocTypeLiteral"}, 7644)
		objectPrimitiveOriginalCount(t, program, "lane7/jsdoc-union-parent-good.a")
	})
}
