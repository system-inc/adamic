// Oracle-only program over the pinned Go parser and real helper functions.
package main

import (
	"encoding/json"
	"fmt"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"strconv"
	"strings"

	"github.com/system-inc/cohere/internal/lint/rules/structure"
)

func main() {
	var paths []string
	data, err := os.ReadFile(os.Args[1])
	must(err)
	must(json.Unmarshal(data, &paths))
	cases, err := os.Create(os.Args[2])
	must(err)
	defer cases.Close()
	expected, err := os.Create(os.Args[3])
	must(err)
	defer expected.Close()
	encode := json.NewEncoder(cases)
	sources := []string{"", "page.tsx", "/page.tsx", "/page.jsx", "/route.ts", "/route.jsx", "C:\\app\\page.tsx", "/page.TSX", "/source/services/network/x.ts", "/services/local-storage/LocalStorageService.ts.bak"}
	for _, prefix := range []string{"", "/", "/repo", "C:\\repo", "/é/𐐀", "/repo/../repo"} {
		for _, base := range []string{"page", "layout", "error", "global-error", "not-found", "loading", "route", "Page", "component"} {
			for _, ext := range []string{"ts", "tsx", "jsx", "js", "TSX", "tsx.bak"} {
				sources = append(sources, prefix+"/"+base+"."+ext)
			}
		}
		for _, suffix := range []string{"/source/services/network/entry.ts", "/components/navigation/Link.tsx", "/components/navigation/Link.jsx", "/components/layout/HorizontalRule.tsx", "/components/layout/HorizontalRule.jsx", "/libraries/structure/entry.ts", "/services/local-storage/LocalStorageService.ts", "/services/local-storage/internal/LocalStorageServiceUtilities.ts"} {
			for _, tail := range []string{"", ".bak", "/nested", "?query"} {
				path := prefix + suffix + tail
				sources = append(sources, path, strings.ReplaceAll(path, "/", `\`), strings.ToLower(path))
			}
		}
	}
	fixtureCount := 0
	for _, path := range paths {
		file, err := goparser.ParseFile(token.NewFileSet(), path, nil, 0)
		must(err)
		constants := map[string]goast.Expr{}
		goast.Inspect(file, func(node goast.Node) bool {
			if declaration, ok := node.(*goast.ValueSpec); ok && len(declaration.Names) == len(declaration.Values) {
				for i, name := range declaration.Names {
					constants[name.Name] = declaration.Values[i]
				}
			}
			return true
		})
		var value func(goast.Expr, int) (string, bool)
		value = func(expression goast.Expr, depth int) (string, bool) {
			if depth > 32 {
				return "", false
			}
			switch e := expression.(type) {
			case *goast.BasicLit:
				if e.Kind == token.STRING {
					s, err := strconv.Unquote(e.Value)
					return s, err == nil
				}
			case *goast.BinaryExpr:
				if e.Op == token.ADD {
					a, ok := value(e.X, depth+1)
					b, ok2 := value(e.Y, depth+1)
					return a + b, ok && ok2
				}
			case *goast.Ident:
				if e, ok := constants[e.Name]; ok {
					return value(e, depth+1)
				}
			case *goast.ParenExpr:
				return value(e.X, depth+1)
			}
			return "", false
		}
		found := 0
		seen := map[string]bool{}
		goast.Inspect(file, func(node goast.Node) bool {
			expression, ok := node.(goast.Expr)
			if !ok {
				return true
			}
			source, ok := value(expression, 0)
			if ok {
				if source != "" && !seen[source] {
					sources = append(sources, source)
					seen[source] = true
					found++
				}
				return false // Preserve complete concatenations rather than parsing their fragments.
			}
			return true
		})
		if found == 0 {
			panic("no string inputs in " + path)
		}
		fmt.Fprintf(os.Stderr, "%s: %d string inputs\n", path, found)
		fixtureCount += found
	}
	for _, path := range sources {
		must(encode.Encode(path))
		c := structure.FileContextFor(path)
		fmt.Fprintf(expected, "%t %t %t %t %t %t %t %t %t\n", c.IsReactFile, c.IsSpecialNextJsFile, c.IsNetworkServiceFile, c.IsLinkComponentFile, c.IsHorizontalRuleComponentFile, c.IsInLibrariesStructure, c.IsPageFile, c.IsLayoutFile, c.IsLocalStorageServiceFile)
	}
	fmt.Fprintf(os.Stderr, "%d fixture strings; %d path observations\n", fixtureCount, len(sources))
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
