package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	cfg "github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"io"
	"os"
	"path"
	"strings"
)

func observe(file, source string) *ast.Node {
	file = path.Clean("/" + file)
	kind := core.ScriptKindTS
	if strings.HasSuffix(file, "x") {
		kind = core.ScriptKindTSX
	}
	f := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: file, Path: tspath.Path(file)}, source, kind)
	for _, root := range cfg.IndexRoots(f.AsNode()) {
		cfg.Build[int](root.Node, cfg.Hooks[int]{})
	}
	return f.AsNode()
}
func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	z, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	d := json.NewDecoder(z)
	live := map[string]map[string]int{}
	for {
		var row struct{ Rule, File, Source string }
		err = d.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		start := len(cfg.AdamicCases)
		observe(row.File, row.Source)
		if live[row.Rule] == nil {
			live[row.Rule] = map[string]int{}
		}
		for _, c := range cfg.AdamicCases[start:] {
			live[row.Rule][c.Op]++
		}
	}
	liveCount := len(cfg.AdamicCases)
	z.Close()
	f.Close()
	for _, source := range []string{
		"function f(a=x){const {b=y,c}=z;return a?b:c;}",
		"function f(){return a ? (b ? c() : d()) : e();}",
		"function f(){return a ? (()=>1) : (x && y);}",
		"class C { @dec [key()]: typeof x; @foo m(@bar a:typeof z):typeof y{return a;} [q:string]:typeof v; plain:number=0; }",
		"const o={ [name()](){return 1}, get x(){return 0}, set x(v){}};",
		"function f(){ const [a, b=g(), ...c]=v; ({x:y=z}=q); }",
	} {
		observe("control.ts", source)
	}

	data, err := json.Marshal(cfg.AdamicCases)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(os.Args[2], data, 0644); err != nil {
		panic(err)
	}
	stats, e := json.Marshal(map[string]any{"live": live, "live_calls": liveCount, "control_calls": len(cfg.AdamicCases) - liveCount, "total_calls": len(cfg.AdamicCases)})
	if e != nil {
		panic(e)
	}
	if e = os.WriteFile(os.Args[2]+".stats.json", stats, 0644); e != nil {
		panic(e)
	}
	for _, s := range cfg.AdamicOutputs {
		fmt.Println(s)
	}
}
