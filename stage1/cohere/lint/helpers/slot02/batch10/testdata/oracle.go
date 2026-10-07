package main

import (
	"encoding/json"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	parser "github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	cfg "github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"os"
	"path/filepath"
	"strings"
)

type Source struct{ Rule, File, Source string }

func main() {
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	var inputs []Source
	if err = json.Unmarshal(data, &inputs); err != nil {
		panic(err)
	}
	controls := []string{`function f(x){x++;return x;x=1}`, `for(let i=0;i<2;i++){if(i)continue;break};while(x){x--};do{x()}while(y);`, `function* g(x){yield x;return x}const f=(x)=>x?f(x-1):0;`, `for(const x of y){try{x()}catch(e){continue}};switch(x){case 1:break;default:throw x}`}
	for _, s := range controls {
		inputs = append(inputs, Source{File: "control.ts", Source: s})
	}
	seen := map[string]bool{}
	sources, roots := 0, 0
	for _, in := range inputs {
		file := "/fixture" + filepath.Ext(in.File)
		key := file + "\x00" + in.Source
		if seen[key] {
			continue
		}
		seen[key] = true
		sources++
		kind := core.ScriptKindTS
		if strings.HasSuffix(file, ".tsx") {
			kind = core.ScriptKindTSX
		} else if strings.HasSuffix(file, ".jsx") {
			kind = core.ScriptKindJSX
		} else if strings.HasSuffix(file, ".js") {
			kind = core.ScriptKindJS
		}
		sf := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: file, Path: tspath.Path(file)}, in.Source, kind)
		for _, root := range cfg.IndexRoots(sf.AsNode()) {
			roots++
			cfg.Slot10Collect(root.Node)
		}
	}
	observed := len(cfg.Slot10Cases)
	cases := []cfg.Slot10Case{}
	for _, c := range cfg.Slot10Cases {
		for _, present := range []bool{false, true} {
			for mode := 0; mode < 2; mode++ {
				for builder := 0; builder < 2; builder++ {
					v := c
					v.Present = present
					v.Mode = mode
					v.Builder = builder
					cases = append(cases, v)
				}
			}
		}
	}
	for _, op := range []string{"read", "write", "loop"} {
		for node := -1; node < len(cfg.Slot10Nodes); node++ {
			for _, reachable := range []bool{false, true} {
				for _, present := range []bool{false, true} {
					for mode := 0; mode < 2; mode++ {
						for builder := 0; builder < 2; builder++ {
							cases = append(cases, cfg.Slot10Case{Op: op, Node: node, Reachable: reachable, Present: present, Mode: mode, Builder: builder})
						}
					}
				}
			}
		}
	}
	var want strings.Builder
	for _, c := range cases {
		want.WriteString(cfg.Slot10Replay(c))
	}
	if err = json.NewEncoder(os.Stdout).Encode(struct {
		Cases                                  []cfg.Slot10Case
		Want                                   string
		Sources, Roots, Observed, Calls, Nodes int
	}{cases, want.String(), sources, roots, observed, cfg.Slot10Calls, len(cfg.Slot10Nodes)}); err != nil {
		panic(err)
	}
}
