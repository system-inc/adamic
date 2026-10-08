package typeaware

import (
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Listener names follow the registry contract and stay stable when the parser adds a kind.
// Keep the production comparison and its independent wrong-kind mutation.
func TestWave14NamedListenerDeclarations(t *testing.T) {
	kinds := map[string]ast.Kind{
		"KindDeleteExpression":         ast.KindDeleteExpression,
		"KindBinaryExpression":         ast.KindBinaryExpression,
		"KindCallExpression":           ast.KindCallExpression,
		"KindTemplateExpression":       ast.KindTemplateExpression,
		"KindClassDeclaration":         ast.KindClassDeclaration,
		"KindClassExpression":          ast.KindClassExpression,
		"KindNewExpression":            ast.KindNewExpression,
		"KindLabeledStatement":         ast.KindLabeledStatement,
		"KindRegularExpressionLiteral": ast.KindRegularExpressionLiteral,
		"KindJsxExpression":            ast.KindJsxExpression,
		"KindSourceFile":               ast.KindSourceFile,
	}
	declaration := regexp.MustCompile(`readonly syntaxKinds: readonly string\[\] = \[([^\]]*)\];`)
	for _, rule := range []struct{ native, directory, production string }{
		{"no_array_delete", "typescript", "no_array_delete"},
		{"no_base_to_string", "typescript", "no_base_to_string"},
		{"no_extraneous_class", "typescript", "no_extraneous_class"},
		{"no_global_listener_target_assertion", "nexus", "correctness_no_global_listener_target_assertion"},
		{"no_mock_on_module_namespace", "nexus", "correctness_no_mock_on_module_namespace"},
		{"no_leaked_number_render", "nexus", "correctness_no_leaked_number_render"},
		{"no_label_var", "core", "no_label_var"},
		{"no_invalid_regexp", "core", "no_invalid_regexp"},
		{"no_misleading_character_class", "core", "no_misleading_character_class"},
	} {
		t.Run(rule.native, func(t *testing.T) {
			path := filepath.Join("..", "..", "..", "cohere", "internal", "lint", "rules", rule.directory, rule.production+".go")
			production, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			var expected []string
			goast.Inspect(production, func(node goast.Node) bool {
				literal, ok := node.(*goast.CompositeLit)
				if !ok {
					return true
				}
				selector, ok := literal.Type.(*goast.SelectorExpr)
				if !ok || selector.Sel.Name != "Listeners" {
					return true
				}
				for _, element := range literal.Elts {
					entry, ok := element.(*goast.KeyValueExpr)
					if !ok {
						t.Fatal("listener without key")
					}
					key, ok := entry.Key.(*goast.SelectorExpr)
					if !ok {
						t.Fatal("listener without SyntaxKind")
					}
					kind, ok := kinds[key.Sel.Name]
					if !ok {
						t.Fatal("unknown production kind", key.Sel.Name)
					}
					expected = append(expected, "'"+strings.TrimPrefix(kind.String(), "Kind")+"'")
				}
				return true
			})
			if len(expected) == 0 {
				t.Fatal("no production listeners")
			}
			source, err := os.ReadFile(rule.native + ".a")
			if err != nil {
				t.Fatal(err)
			}
			want := strings.Join(expected, ", ")
			check := func(source []byte) error {
				matches := declaration.FindAllSubmatch(source, -1)
				if len(matches) != 1 || string(matches[0][1]) != want {
					return fmt.Errorf("named listeners disagree with production Go: want [%s]", want)
				}
				return nil
			}
			if err := check(source); err != nil {
				t.Fatal(err)
			}
			mutant := []byte(strings.Replace(string(source), "= ["+want+"];", "= ['MissingSyntaxKind'];", 1))
			if err := check(mutant); err == nil {
				t.Fatal("listener declaration mutant survived")
			}
			t.Logf("Go listener kinds [%s] agree; wrong-kind declaration mutant caught", want)
		})
	}
}
