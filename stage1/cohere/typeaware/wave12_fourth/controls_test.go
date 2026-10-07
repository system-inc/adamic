package wave12fourth

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func referenceControls(t *testing.T, repository string) []string {
	t.Helper()
	var sources []string
	for _, file := range []string{"core/prefer_regex_literals_test.go", "core/prefer_regex_literals_corpus_test.go", "core/prefer_rest_params_test.go", "react/exhaustive_deps_test.go"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules", file), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		count, jsx := 0, 0
		goast.Inspect(tree, func(n goast.Node) bool {
			row, ok := n.(*goast.CompositeLit)
			if !ok {
				return true
			}
			var literal *goast.BasicLit
			for _, elt := range row.Elts {
				field, ok := elt.(*goast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := field.Key.(*goast.Ident)
				if ok && (key.Name == "source" || key.Name == "sourceText") {
					literal, _ = field.Value.(*goast.BasicLit)
				}
			}
			if literal == nil && len(row.Elts) > 1 {
				at := 1
				if strings.Contains(file, "corpus_test") {
					at = 2
				} else if strings.Contains(file, "react/") {
					at = 0
					if len(row.Elts) == 4 {
						at = 1
						if option, ok := row.Elts[1].(*goast.BasicLit); ok {
							decoded, _ := strconv.Unquote(option.Value)
							if strings.HasPrefix(decoded, `{"`) {
								at = 2
							}
						}
					}
				}
				if at < len(row.Elts) {
					literal, _ = row.Elts[at].(*goast.BasicLit)
				}
			}
			if literal == nil || literal.Kind != token.STRING {
				return true
			}
			value, err := strconv.Unquote(literal.Value)
			if err != nil {
				t.Fatal(err)
			}
			if !(strings.Contains(value, "RegExp") || strings.Contains(value, "arguments") || strings.Contains(value, "useEffect") || strings.Contains(value, "useMemo") || strings.Contains(value, "useCallback") || strings.Contains(value, "useImperativeHandle") || strings.Contains(value, "useLayoutEffect")) {
				return true
			}
			if strings.Contains(value, "</") || strings.Contains(value, "/>") || strings.Contains(value, "<>") {
				jsx++
			}
			sources = append(sources, value)
			count++
			return false
		})
		if count == 0 {
			t.Fatal("empty extraction", file)
		}
		t.Logf("%s: %d literal source rows, %d JSX rows included with shared parser", file, count, jsx)
	}
	return sources
}
