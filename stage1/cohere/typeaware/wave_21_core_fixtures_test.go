package typeaware

import (
	"fmt"
	goast "go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

func wave21CoreFixtures(h *harness) (defaults, allowVoid, jsx []string) {
	for _, stem := range []string{"no_obj_calls", "no_obj_calls_corpus", "no_object_constructor", "no_promise_executor_return"} {
		tree, e := parser.ParseFile(token.NewFileSet(), filepath.Join(h.repository, "cohere/internal/lint/rules/core", stem+"_test.go"), nil, 0)
		if e != nil {
			h.t.Fatal(e)
		}
		index := 0
		save := func(text, options string) {
			name := fmt.Sprintf("%s-%03d.a", stem, index)
			index++
			if strings.Contains(text, "<foo") {
				name = strings.TrimSuffix(name, ".a") + ".tsx"
			}
			path := h.write(name, strings.TrimSpace(text)+"\n")
			if strings.Contains(text, "<foo") || strings.Contains(text, "yield:") {
				jsx = append(jsx, path)
			} else if strings.Contains(options, "\"allowVoid\": true") || strings.Contains(options, "\"allowVoid\":true") {
				allowVoid = append(allowVoid, path)
			} else {
				defaults = append(defaults, path)
			}
		}
		literal := func(e goast.Expr) (string, bool) {
			l, ok := e.(*goast.BasicLit)
			if !ok || l.Kind != token.STRING {
				return "", false
			}
			s, err := strconv.Unquote(l.Value)
			if err != nil {
				h.t.Fatal(err)
			}
			return s, true
		}
		goast.Inspect(tree, func(n goast.Node) bool {
			list, ok := n.(*goast.CompositeLit)
			if !ok {
				return true
			}
			array, ok := list.Type.(*goast.ArrayType)
			if !ok {
				return true
			}
			if structure, ok := array.Elt.(*goast.StructType); ok {
				names := []string{}
				for _, field := range structure.Fields.List {
					for _, name := range field.Names {
						names = append(names, name.Name)
					}
				}
				source, option := -1, -1
				for i, name := range names {
					if name == "sourceText" || name == "code" || name == "source" {
						source = i
					}
					if name == "options" {
						option = i
					}
				}
				if source >= 0 {
					for _, item := range list.Elts {
						row, ok := item.(*goast.CompositeLit)
						if !ok || len(row.Elts) <= source {
							continue
						}
						text, ok := literal(row.Elts[source])
						if !ok {
							continue
						}
						settings := ""
						if option >= 0 && len(row.Elts) > option {
							settings, _ = literal(row.Elts[option])
						}
						save(text, settings)
					}
					return false
				}
			}
			// These standalone string tables exercise callable globals and constructor discrimination.
			if id, ok := array.Elt.(*goast.Ident); ok && id.Name == "string" {
				for _, item := range list.Elts {
					text, ok := literal(item)
					if ok && (strings.Contains(text, ";") || strings.Contains(text, "Promise(")) {
						save(text, "")
					}
				}
				return false
			}
			return true
		})
		h.t.Logf("%s: %d extracted upstream fixture rows", stem, index)
	}
	// Independent positive controls and declaration/pattern/written-global boundaries.
	for i, text := range []string{"Math();", "let j=JSON;j();", "Object();", "new Promise(r=>1);", "with(obj) Object();", "(<any>JSON)();", "Math=1;Math();", "[...Math]=[];Math();", "({k:[...JSON]}=x);JSON();", "({[JSON]:x}=obj);JSON();", "const {JSON:j}=globalThis;j();", "const {['JS'+'ON']:j}=globalThis;j();", "/* 世界 🌍 */\r\nconst result=Object();new Promise(r=> /* 雪 */ (1));", "(Object)();"} {
		defaults = append(defaults, h.write(fmt.Sprintf("core-boundary-%03d.a", i), text+"\n"))
	}
	return
}
