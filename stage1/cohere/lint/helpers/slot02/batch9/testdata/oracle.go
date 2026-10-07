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
	controls := []string{`function f(x){try {if(x)return 1;throw x} catch(e){return 2} finally{x++} }`, `outer:while(x){if(y)continue outer;break} do{x()}while(y);for(;;){break}`, `function* g(x){yield x;return x}const f=(x)=>x?f(x-1):0;`, `for(const x of y){try{x()}catch(e){continue}};switch(x){case 1:break;default:throw x}`}
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
			cfg.Slot02Collect(root.Node)
		}
	}
	cases := cfg.Slot02Unique()
	observed := len(cases)
	// Every graph-state control is evaluated by the actual Go private method.
	for n := 0; n <= 5; n++ {
		for present := 0; present < 2; present++ {
			for length := 0; length <= n; length++ {
				if present == 0 && length != 0 {
					continue
				}
				for reachable := 0; reachable < 2; reachable++ {
					for incoming := 0; incoming < 2; incoming++ {
						for _, op := range []string{"set", "with", "link"} {
							for _, barrier := range []bool{false, true} {
								for from := -1; from <= 0; from++ {
									for to := -1; to <= 1; to++ {
										blocks := []cfg.Slot02Block{{Reachable: reachable == 1, Incoming: false, BarriersPresent: present == 1, Successors: []int{}, Barriers: []bool{}}, {Incoming: incoming == 1, Successors: []int{}, Barriers: []bool{}}}
										for i := 0; i < n; i++ {
											blocks[0].Successors = append(blocks[0].Successors, i%2)
										}
										for i := 0; i < length; i++ {
											blocks[0].Barriers = append(blocks[0].Barriers, i%2 == 0)
										}
										indices := []int{-1}
										if n > 0 {
											indices = append(indices, 0, n-1)
										}
										if !barrier && present == 0 {
											indices = append(indices, n+5)
										}
										if op != "set" {
											indices = []int{0}
										}
										for _, index := range indices {
											cases = append(cases, cfg.Slot02Case{Op: op, From: from, To: to, Index: index, Barrier: barrier, Blocks: blocks})
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
	var want strings.Builder
	for _, c := range cases {
		want.WriteString(cfg.Slot02Replay(c))
	}
	if err = json.NewEncoder(os.Stdout).Encode(struct {
		Cases                    []cfg.Slot02Case
		Want                     string
		Sources, Roots, Observed int
	}{cases, want.String(), sources, roots, observed}); err != nil {
		panic(err)
	}
}
