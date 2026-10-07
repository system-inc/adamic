package main

import (
	"encoding/json"
	tsast "github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	tsparser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	cfg "github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"go/ast"
	goParser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Source struct{ Rule, File, Source string }
type Config struct {
	Inputs []Source
	Paths  []string
}

func main() {
	var config Config
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()
	if err = json.NewDecoder(file).Decode(&config); err != nil {
		panic(err)
	}
	config.Inputs = append(config.Inputs, Source{File: "control.ts", Source: `function f(){let a=1,b=2,c=3;const d=4,e=5;for(let i=0,j=1;i<2;i++){j++}}`})
	texts := []string{}
	for _, path := range config.Paths {
		file, err := goParser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			panic(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			if literal, ok := node.(*ast.BasicLit); ok && literal.Kind == token.STRING {
				value, err := strconv.Unquote(literal.Value)
				if err != nil {
					panic(err)
				}
				texts = append(texts, value)
			}
			return true
		})
	}
	seen := map[string]bool{}
	for _, input := range config.Inputs {
		texts = append(texts, input.Source)
		input.File = "/fixture" + filepath.Ext(input.File)
		identity := input.File + "\x00" + input.Source
		if seen[identity] {
			continue
		}
		seen[identity] = true
		kind := core.ScriptKindTS
		if strings.HasSuffix(input.File, ".tsx") {
			kind = core.ScriptKindTSX
		} else if strings.HasSuffix(input.File, ".jsx") {
			kind = core.ScriptKindJSX
		} else if strings.HasSuffix(input.File, ".js") {
			kind = core.ScriptKindJS
		}
		source := tsparser.ParseSourceFile(tsast.SourceFileParseOptions{FileName: input.File, Path: tspath.Path(input.File)}, input.Source, kind)
		var walk func(*tsast.Node) bool
		walk = func(node *tsast.Node) bool {
			if node == nil {
				return false
			}
			switch node.Kind {
			case tsast.KindStringLiteral, tsast.KindNoSubstitutionTemplateLiteral, tsast.KindTemplateHead, tsast.KindTemplateMiddle, tsast.KindTemplateTail:
				texts = append(texts, node.Text())
			}
			node.ForEachChild(walk)
			return false
		}
		walk(source.AsNode())
		for _, root := range cfg.IndexRoots(source.AsNode()) {
			cfg.Slot11Collect(root.Node)
		}
	}
	css, want := collapse.Slot11CSS(texts)
	graph, other := cfg.Slot11CFG()
	if err = json.NewEncoder(os.Stdout).Encode(struct {
		CSS   []collapse.Slot11Ingest
		CFG   []cfg.Slot11Case
		Want  string
		Calls int
	}{css, graph, want + other, cfg.Slot11Calls}); err != nil {
		panic(err)
	}
}
