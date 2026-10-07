package typeaware

import (
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func wave21ProcessFixtures(h *harness, stem, prefix string) []string {
	h.t.Helper()
	tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/nexus", stem+"_test.go"), nil, 0)
	if err != nil {
		h.t.Fatal(err)
	}
	values := map[string]goast.Expr{}
	parents := map[goast.Node]goast.Node{}
	var stack []goast.Node
	goast.Inspect(tree, func(n goast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if len(stack) > 0 {
			parents[n] = stack[len(stack)-1]
		}
		stack = append(stack, n)
		if v, ok := n.(*goast.ValueSpec); ok {
			for i, name := range v.Names {
				if i < len(v.Values) {
					values[name.Name] = v.Values[i]
				}
			}
		}
		return true
	})
	var evaluate func(goast.Expr) string
	evaluate = func(e goast.Expr) string {
		switch n := e.(type) {
		case *goast.BasicLit:
			value, err := strconv.Unquote(n.Value)
			if err != nil {
				h.t.Fatal(err)
			}
			return value
		case *goast.Ident:
			return evaluate(values[n.Name])
		case *goast.BinaryExpr:
			if n.Op == token.ADD {
				return evaluate(n.X) + evaluate(n.Y)
			}
		case *goast.CallExpr:
			if name, ok := n.Fun.(*goast.Ident); ok && name.Name == prefix+"Lines" {
				var lines []string
				for _, e := range n.Args {
					lines = append(lines, evaluate(e))
				}
				return strings.Join(lines, "\n") + "\n"
			}
			if s, ok := n.Fun.(*goast.SelectorExpr); ok && s.Sel.Name == "Join" {
				list := n.Args[0].(*goast.CompositeLit)
				var lines []string
				for _, e := range list.Elts {
					lines = append(lines, evaluate(e))
				}
				return strings.Join(lines, evaluate(n.Args[1]))
			}
		}
		h.t.Fatalf("unsupported fixture expression %T", e)
		return ""
	}
	for _, suffix := range []string{"NodeTypes", "WebConsole", "NodeTimers"} {
		if value := values[prefix+suffix]; value != nil {
			h.write(prefix+suffix+".d.ts", evaluate(value))
		}
	}
	prelude := evaluate(values[prefix+"Prelude"])
	if fixture, ok := values[prefix+"NexusFiles"].(*goast.CompositeLit); ok {
		for _, e := range fixture.Elts {
			kv := e.(*goast.KeyValueExpr)
			path := filepath.Join(h.directory, evaluate(kv.Key))
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				h.t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(evaluate(kv.Value)), 0644); err != nil {
				h.t.Fatal(err)
			}
		}
	}
	var sources []string
	goast.Inspect(tree, func(n goast.Node) bool {
		if call, ok := n.(*goast.CallExpr); ok {
			if name, ok := call.Fun.(*goast.Ident); ok && name.Name == prefix+"Source" && call.Ellipsis == token.NoPos {
				var lines []string
				for _, e := range call.Args {
					if _, ok := e.(*goast.BasicLit); !ok {
						return true
					}
					lines = append(lines, evaluate(e))
				}
				sources = append(sources, strings.Join(lines, "\n"))
			}
		}
		list, ok := n.(*goast.CompositeLit)
		if !ok {
			return true
		}
		typ, ok := list.Type.(*goast.ArrayType)
		if !ok || typ.Len != nil {
			return true
		}
		element, ok := typ.Elt.(*goast.Ident)
		if !ok || element.Name != "string" {
			return true
		}
		take := false
		switch parent := parents[n].(type) {
		case *goast.CompositeLit:
			if len(parent.Elts) > 1 && parent.Elts[1] == list {
				_, take = parent.Elts[0].(*goast.BasicLit)
			}
		case *goast.KeyValueExpr:
			if name, ok := parent.Key.(*goast.Ident); ok {
				take = name.Name == "before" || name.Name == "after"
			}
		}
		if take {
			var lines []string
			for _, e := range list.Elts {
				lines = append(lines, evaluate(e))
			}
			sources = append(sources, strings.Join(lines, "\n"))
		}
		return true
	})
	var paths []string
	if stem == "correctness_no_process_exit_after_output" {
		paths = append(paths, h.write("Help.ts", evaluate(values[prefix+"HelpModule"])), h.write("ScriptHelp.ts", evaluate(values[prefix+"Script"])))
	}
	for i, source := range sources {
		name := fmt.Sprintf("%s-%03d.a", stem, i)
		if stem == "correctness_no_discarded_outcome" {
			name = filepath.Join("repository/modules/meta", name)
			if err := os.MkdirAll(filepath.Join(h.directory, "repository/modules/meta"), 0755); err != nil {
				h.t.Fatal(err)
			}
		}
		paths = append(paths, h.write(name, prelude+source+"\nexport {};\n"))
	}
	if len(paths) == 0 {
		h.t.Fatal("no upstream fixtures imported")
	}
	h.t.Logf("%s: %d upstream fixture sources", stem, len(paths))
	return paths
}
