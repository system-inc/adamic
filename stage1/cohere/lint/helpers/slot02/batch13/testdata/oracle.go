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
	"go/parser"
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
	var c Config
	f, e := os.Open(os.Args[1])
	if e != nil {
		panic(e)
	}
	defer f.Close()
	if e = json.NewDecoder(f).Decode(&c); e != nil {
		panic(e)
	}
	texts := []string{}
	for _, p := range c.Paths {
		a, e := parser.ParseFile(token.NewFileSet(), p, nil, 0)
		if e != nil {
			panic(e)
		}
		ast.Inspect(a, func(n ast.Node) bool {
			if l, ok := n.(*ast.BasicLit); ok && l.Kind == token.STRING {
				s, e := strconv.Unquote(l.Value)
				if e != nil {
					panic(e)
				}
				texts = append(texts, s)
			}
			return true
		})
	}
	c.Inputs = append(c.Inputs, Source{File: "controls.ts", Source: `function f(a,b,c){if((a && !b) || (c ?? a)){return a?b:c}else{throw a}while(a){break}do{continue}while(b);for(;;){}for(const x in a){}for(const y of b){}switch(a){case 1:break;default:debugger}try{}catch(e){}finally{}label:with(a){b;}class A{};return;} import x from 'x';export {x};export default x; +x;!(a||b);`})
	seen := map[string]bool{}
	for _, in := range c.Inputs {
		key := filepath.Ext(in.File) + "\x00" + in.Source
		if seen[key] {
			continue
		}
		seen[key] = true
		kind := core.ScriptKindTS
		if strings.HasSuffix(in.File, ".tsx") {
			kind = core.ScriptKindTSX
		} else if strings.HasSuffix(in.File, ".jsx") {
			kind = core.ScriptKindJSX
		} else if strings.HasSuffix(in.File, ".js") {
			kind = core.ScriptKindJS
		}
		p := "/fixture" + filepath.Ext(in.File)
		s := tsparser.ParseSourceFile(tsast.SourceFileParseOptions{FileName: p, Path: tspath.Path(p)}, in.Source, kind)
		cfg.Slot13Collect(s.AsNode())
		var walk func(*tsast.Node) bool
		walk = func(n *tsast.Node) bool {
			if n == nil {
				return false
			}
			switch n.Kind {
			case tsast.KindStringLiteral, tsast.KindNoSubstitutionTemplateLiteral:
				texts = append(texts, n.Text())
			}
			n.ForEachChild(walk)
			return false
		}
		walk(s.AsNode())
	}
	nodes, cases, want := cfg.Slot13CFG()
	values, other := collapse.Slot13Values(texts)
	if e = json.NewEncoder(os.Stdout).Encode(struct {
		Nodes              []cfg.Slot13Node
		Cases              []cfg.Slot13Case
		Values             []collapse.Slot13ValueCase
		Want               string
		CFGWant, ValueWant []string
	}{nodes, cases, values, want + other, cfg.Slot13CFGWant, collapse.Slot13ValueWant}); e != nil {
		panic(e)
	}
}
