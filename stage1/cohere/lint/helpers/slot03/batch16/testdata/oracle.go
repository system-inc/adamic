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
	for {
		var row struct{ File, Source string }
		err = d.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		observe(row.File, row.Source)
	}
	z.Close()
	f.Close()
	for _, source := range []string{"outer: inner: while(x){if(y)break outer;else continue inner;}", "do {continue;}while(x);", "switch(x){case 1:break;default:throw x;}", "function f(){try {return x;} finally {x();}}", "function* f(){yield x;yield* y;}", "for(;;){for(;;){break;}continue;}", "a: {break a;}", "class C {static {while(x)break;} x=()=>{return 1;};}"} {
		observe("control.ts", source)
	}
	cfg.AdamicControls(observe("control.ts", ""))
	data, err := json.Marshal(cfg.AdamicCases)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(os.Args[2], data, 0644); err != nil {
		panic(err)
	}
	for _, s := range cfg.AdamicOutputs {
		fmt.Println(s)
	}
}
