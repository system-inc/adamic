package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	texthelper "github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"os"
	"strings"
)

type Row struct{ Name, Source string }
type Entity struct {
	Body  string `json:"body"`
	Text  string `json:"text"`
	Valid bool   `json:"valid"`
}
type Adapted struct {
	Text     string   `json:"text"`
	Entities []Entity `json:"entities"`
}

func main() {
	path := os.Args[1]
	adapt := path == "--ast"
	if adapt {
		path = os.Args[2]
	}
	data, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var rows []Row
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	output := []Adapted{}
	for _, row := range rows {
		values := []string{row.Source}
		if !strings.HasPrefix(row.Name, "control:") {
			kind := core.ScriptKindTS
			switch {
			case strings.HasSuffix(row.Name, ".tsx"):
				kind = core.ScriptKindTSX
			case strings.HasSuffix(row.Name, ".jsx"):
				kind = core.ScriptKindJSX
			case strings.HasSuffix(row.Name, ".js"):
				kind = core.ScriptKindJS
			}
			name := tspath.NormalizePath("/corpus/" + row.Name)
			file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: name, Path: tspath.Path(name)}, row.Source, kind)
			var visit func(*ast.Node)
			visit = func(n *ast.Node) {
				if n == nil {
					return
				}
				if n.Kind == ast.KindStringLiteral {
					values = append(values, n.Text())
				} else if n.Kind == ast.KindJsxText {
					values = append(values, n.AsJsxText().Text)
				}
				n.ForEachChild(func(child *ast.Node) bool { visit(child); return false })
			}
			visit(file.AsNode())
		}
		for _, value := range values {
			if adapt {
				item := Adapted{Text: value, Entities: []Entity{}}
				seen := map[string]bool{}
				for index := 0; index < len(value); index++ {
					if value[index] != '&' {
						continue
					}
					semicolon := strings.IndexByte(value[index:], ';')
					if semicolon < 2 {
						continue
					}
					body := value[index+1 : index+semicolon]
					if seen[body] {
						continue
					}
					seen[body] = true
					replacement, valid := texthelper.AdamicDecodeEntity(body)
					item.Entities = append(item.Entities, Entity{body, replacement, valid})
				}
				output = append(output, item)
			} else {
				for _, b := range []byte(texthelper.UnescapeStringLiteralText(value)) {
					fmt.Printf("%d,", b)
				}
				fmt.Println()
			}
		}
	}
	if adapt {
		if err = json.NewEncoder(os.Stdout).Encode(output); err != nil {
			panic(err)
		}
	}
}
