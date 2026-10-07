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
	next "github.com/system-inc/cohere/internal/lint/ecmascript/nextjs"
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
	pathCases := []map[string]any{}
	pathOutputs := []string{}
	addPath := func(file string) {
		base, parent := next.AdamicSplitPath(file)
		bytes := []int{}
		for _, b := range []byte(file) {
			bytes = append(bytes, int(b))
		}
		pathCases = append(pathCases, map[string]any{"op": "path", "path": file, "base": base, "parent": parent, "bytes": bytes})
		pathOutputs = append(pathOutputs, fmt.Sprintf("%d/%d", bit(next.IsDocumentFile(file)), next.AdamicLastSeparator(file)))
	}
	for {
		var row struct{ Rule, File, Source string }
		err = d.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		addPath(row.File)
		addPath(path.Clean("/" + row.File))
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
	pathLiveCount := len(pathCases)
	z.Close()
	f.Close()
	for _, source := range []string{
		"function f(){try {a()}catch(e){b()}}",
		"function f(){try {a()}catch{b()}finally{c()}}",
		"function f(){try {return a()}finally {b()}}",
		"function f(){try {throw 1}finally{b()}}",
		"function f(){try {try {return a()}finally{b()}}finally{c()}}",
		"function f(){try {try {throw 1}finally{b()}}catch(e){c()}finally{d()}}",
		"function f(){try {if(x)return 1;throw 2;}finally{if(y)return 3;}}",
		"function f(){try {return 1}finally {throw 2}}",
		"function f(){try{try{return 1}finally{b()}}catch(e){c()}}",
		"function f(){try {a()}catch({x=def(),[key()]:z}){throw x;}finally{c()}}",
		"function f(){try {try {throw 1}finally{b()}}finally{c()}}",
	} {
		cfg.AdamicControls(observe("control.ts", source))
	}

	for _, p := range []string{"", "_document", "_document.", "_document.tsx", "_documentation.tsx", "pages/_document/index.tsx", "pages/_documentation/indexical", "pages/_document/Index.tsx", "pages/_document/indexical", "components/_document.tsx", "index", "_document/index", "x/index", "a/b\\c", "a\\b/c", "a/", "a\\", "a//b", "é/文\\😀/x", "😀\\x", "\xff/\xfe\\x", "\x00/\x00"} {
		addPath(p)
	}
	all := []any{}
	for _, c := range cfg.AdamicCases {
		all = append(all, c)
	}
	for _, c := range pathCases {
		all = append(all, c)
	}
	data, err := json.Marshal(all)
	if err != nil {
		panic(err)
	}
	if err = os.WriteFile(os.Args[2], data, 0644); err != nil {
		panic(err)
	}
	stats, e := json.Marshal(map[string]any{"live": live, "live_calls": liveCount, "control_calls": len(cfg.AdamicCases) - liveCount, "cfg_calls": len(cfg.AdamicCases), "path_live_calls": pathLiveCount, "path_control_calls": len(pathCases) - pathLiveCount, "total_lines": len(all)})
	if e != nil {
		panic(e)
	}
	if e = os.WriteFile(os.Args[2]+".stats.json", stats, 0644); e != nil {
		panic(e)
	}
	cfg.AdamicOutputs = append(cfg.AdamicOutputs, pathOutputs...)
	for _, s := range cfg.AdamicOutputs {
		fmt.Println(s)
	}
}

func bit(v bool) int {
	if v {
		return 1
	}
	return 0
}
