// Command inventory extracts regexp calls from the pinned cohere syntax trees.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Row struct {
	Rule       string  `json:"rule"`
	File       string  `json:"file"`
	Line       int     `json:"line"`
	Call       string  `json:"call"`
	Expression string  `json:"expression"`
	Pattern    *string `json:"go_pattern"`
	Feature    string  `json:"feature"`
}

func main() {
	root := os.Args[1]
	rows := []Row{}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fs := token.NewFileSet()
		f, e := parser.ParseFile(fs, path, nil, 0)
		if e != nil {
			return e
		}
		alias := ""
		for _, im := range f.Imports {
			if im.Path.Value == `"regexp"` {
				alias = "regexp"
				if im.Name != nil {
					alias = im.Name.Name
				}
			}
		}
		if alias == "" {
			return nil
		}
		defs := map[string]ast.Expr{}
		ast.Inspect(f, func(n ast.Node) bool {
			if v, ok := n.(*ast.ValueSpec); ok {
				for i, name := range v.Names {
					if i < len(v.Values) {
						defs[name.Name] = v.Values[i]
					}
				}
			}
			return true
		})
		var eval func(ast.Expr, int) (string, bool)
		eval = func(x ast.Expr, depth int) (string, bool) {
			if depth > 20 {
				return "", false
			}
			switch v := x.(type) {
			case *ast.BasicLit:
				if v.Kind == token.STRING {
					s, e := strconv.Unquote(v.Value)
					return s, e == nil
				}
			case *ast.Ident:
				if z, ok := defs[v.Name]; ok {
					return eval(z, depth+1)
				}
			case *ast.BinaryExpr:
				if v.Op == token.ADD {
					a, ok := eval(v.X, depth+1)
					b, yes := eval(v.Y, depth+1)
					return a + b, ok && yes
				}
			case *ast.ParenExpr:
				return eval(v.X, depth+1)
			}
			return "", false
		}
		ast.Inspect(f, func(n ast.Node) bool {
			c, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := c.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			id, ok := sel.X.(*ast.Ident)
			if !ok || id.Name != alias || (sel.Sel.Name != "Compile" && sel.Sel.Name != "MustCompile") || len(c.Args) != 1 {
				return true
			}
			rel, _ := filepath.Rel(root, path)
			var b bytes.Buffer
			printer.Fprint(&b, fs, c.Args[0])
			row := Row{Rule: strings.ReplaceAll(strings.TrimSuffix(filepath.Base(path), ".go"), "_", "-"), File: filepath.ToSlash(rel), Line: fs.Position(c.Pos()).Line, Call: sel.Sel.Name, Expression: b.String(), Feature: "dynamic"}
			if s, ok := eval(c.Args[0], 0); ok {
				row.Pattern = &s
				row.Feature = "plain"
				features := []string{}
				for _, p := range []string{"(?i)", `\p{`, `\z`, "(?P<", "(?s)", "(?m)"} {
					if strings.Contains(s, p) {
						features = append(features, p)
					}
				}
				if len(features) > 0 {
					row.Feature = strings.Join(features, ",")
				}
			}
			rows = append(rows, row)
			return true
		})
		return nil
	})
	if err != nil {
		panic(err)
	}
	out, _ := json.MarshalIndent(rows, "", "  ")
	fmt.Println(string(out))
}
