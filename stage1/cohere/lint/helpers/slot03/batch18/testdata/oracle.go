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
		"function f(){a && b(); a || b(); a ?? b(); a &&= b(); a ||= b(); a ??= b(); a=b(); a+=b(); a-b(); ({x:y}=z); [a,b]=z; (({x:y}))=z;}",
		"function f(){if(a) b();if(a) b();else c();if(a&&b)c();else if(d)e();else f();}",
		"function f(){if(a){return 1}else{throw 2};if(true){}else{};if(false){};}",
		"class C<T extends typeof x = typeof y> extends Base<typeof z> implements A,B { @dec [key()]: typeof value; plain=() => 1; @foo m(@bar a:typeof b){} }",
		"const C=class extends B { [key()](){}; x:number; }; class Empty {}",
		"function f(){for(;;){if(a)break;else continue;}try{if(a)throw b();else return c();}finally{d();}}",
	} {
		observe("control.ts", source)
	}
	cfg.AdamicClassControls(observe("control.ts", "class Control extends Base implements A,B { x:number; [key()](){} }"))

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
