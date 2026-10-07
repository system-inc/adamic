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
		cfg.Build[int](root.Node, cfg.AdamicHooks())
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
		"function* f(){a;undefined;arguments;this; (a);a+b;a?b:c;++a;--a;a++;a--;!a;-a;a.b;a[b];a();new C();tag`hi`;yield a;yield;const C=class {};const g=()=>a;const h=function(){};}",
		"function f(){let {x:y=def(),[key()]:z,...rest}=source;let [a,,b=def(),...tail]=source;({x:y,[key()]:z,a=def(),...rest}=source);[a,,b=def(),...tail]=source;((a))=b;}",
		"function f(x){switch(x){};switch(x){default:a()};switch(x){case a():b();case c():break;default:d()};switch(x){default:a();case b():c();case d():e()};}",
		"function f(x){outer:while(x){switch(x){case a():if(b)break;case c():continue outer;default:break outer;}}}",
		"function f(x){switch(a&&b){case c&&d:if(e)return;case f():throw 1;case g():break;case h():i()}}",
	} {
		observe("control.ts", source)
	}
	cfg.AdamicControls(observe("control.ts", "function f(){let {[key()]:x=def(),...rest}=source;let [a,,b=def(),...tail]=source;({x:y,[key()]:z,a=def(),...rest}=source);[a,,b=def(),...tail]=source;a=b;((a));a!;a as T;a satisfies T;<T>a;switch(x){case 1:a();default:b()}}"))

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
