// Overlay-built inside the pinned typescript-go module. Research only.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"testing"
	"unsafe"

	"github.com/microsoft/TypeScript/tsc/internal/ast"
	"github.com/microsoft/TypeScript/tsc/internal/core"
	"github.com/microsoft/TypeScript/tsc/internal/parser"
)

func main() {
	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	parse := func() *ast.SourceFile {
		return parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: "/source.ts"}, string(source), core.ScriptKindTS)
	}
	runtime.KeepAlive(parse())
	runtime.GC()
	var before, allocated, retained runtime.MemStats
	runtime.ReadMemStats(&before)
	file := parse()
	runtime.ReadMemStats(&allocated)
	runtime.GC()
	runtime.ReadMemStats(&retained)
	var walk func(*ast.Node) int
	walk = func(n *ast.Node) int {
		count := 1
		n.ForEachChild(func(c *ast.Node) bool { count += walk(c); return false })
		return count
	}
	visits := walk(file.AsNode())
	walkAllocs := testing.AllocsPerRun(100, func() { walk(file.AsNode()) })
	runtime.KeepAlive(file)
	out := map[string]any{"source_bytes": len(source), "visits": visits, "node_base_bytes": unsafe.Sizeof(ast.Node{}), "node_list_bytes": unsafe.Sizeof(ast.NodeList{}), "identifier_bytes": unsafe.Sizeof(ast.Identifier{}), "parse_allocations": allocated.Mallocs - before.Mallocs, "parse_allocated_bytes": allocated.TotalAlloc - before.TotalAlloc, "heap_delta_after_gc": int64(retained.HeapAlloc) - int64(before.HeapAlloc), "walk_allocations": walkAllocs, "go": runtime.Version()}
	b, err := json.Marshal(out)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
