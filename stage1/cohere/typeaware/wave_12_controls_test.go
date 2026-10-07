package typeaware

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Replay the pinned Go rule's reference rows as source; production Go decides
// their default-option results anew rather than copying any fixture expectation.
func wave12ReferenceControls(t *testing.T, repository string) []string {
	t.Helper()
	var sources []string
	for _, file := range []string{"typescript/no_redeclare_corpus_data_test.go", "typescript/no_redeclare_discrimination_test.go", "nexus/correctness_no_test_on_global_regex_test.go", "nexus/correctness_no_write_only_collection_test.go"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules", file), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		goast.Inspect(tree, func(n goast.Node) bool {
			row, ok := n.(*goast.CompositeLit)
			if !ok || len(row.Elts) < 2 {
				return true
			}
			if strings.Contains(file, "nexus/") {
				lines, ok := row.Elts[1].(*goast.CompositeLit)
				if !ok {
					return true
				}
				array, ok := lines.Type.(*goast.ArrayType)
				if !ok {
					return true
				}
				element, ok := array.Elt.(*goast.Ident)
				if !ok || element.Name != "string" {
					return true
				}
				var text []string
				for _, expr := range lines.Elts {
					lit, ok := expr.(*goast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						return true
					}
					value, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Fatal(err)
					}
					text = append(text, value)
				}
				prelude := "interface RowInterface{rowid:number;id:string};declare const rows:RowInterface[];declare const lines:string[];declare function use(x:unknown):void;\n"
				sources = append(sources, prelude+strings.Join(text, "\n")+"\nexport {};\n")
				count++
				return false
			}
			at := 1
			if strings.Contains(file, "corpus_data") {
				at = 2
				if len(row.Elts) <= at {
					return true
				}
			}
			lit, ok := row.Elts[at].(*goast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(text, ";") || strings.Contains(text, "\n") {
				sources = append(sources, text)
				count++
			}
			return true
		})
		if count == 0 {
			t.Fatal("reference extraction empty", file)
		}
		t.Logf("%s: %d reference source rows", file, count)
	}
	return sources
}
