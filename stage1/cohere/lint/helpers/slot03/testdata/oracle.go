package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rules/tailwind"
	"os"
	"sort"
	"strings"
)

func main() {
	f, err := os.Open(os.Args[1])
	must(err)
	defer f.Close()
	z, err := gzip.NewReader(f)
	must(err)
	defer z.Close()
	scan := bufio.NewScanner(z)
	scan.Buffer(make([]byte, 4096), 16<<20)
	names := map[string]bool{}
	points := map[int]bool{}
	for scan.Scan() {
		var row struct{ File, Source string }
		must(json.Unmarshal(scan.Bytes(), &row))
		kind := core.ScriptKindTS
		if strings.HasSuffix(row.File, ".tsx") {
			kind = core.ScriptKindTSX
		}
		if strings.HasSuffix(row.File, ".jsx") {
			kind = core.ScriptKindJSX
		}
		if !strings.HasPrefix(row.File, "/") {
			row.File = "/" + row.File
		}
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: row.File, Path: tspath.Path(row.File)}, row.Source, kind)
		var walk func(*ast.Node) bool
		walk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			switch n.Kind {
			case ast.KindIdentifier, ast.KindStringLiteral, ast.KindPrivateIdentifier:
				names[n.Text()] = true
			}
			n.ForEachChild(walk)
			return false
		}
		walk(file.AsNode())
		for _, point := range row.Source {
			points[int(point)] = true
		}
	}
	must(scan.Err())
	for _, name := range []string{"Component", "PureComponent", "component", "pureComponent", "COMPONENT", "Purecomponent", "ComponentX", "XComponent", " Component", "Component ", "Component\x00", "PureComponent\n", "React.Component", "React.PureComponent", "", "ΩComponent", "Ｃomponent", "组件", "😀"} {
		names[name] = true
	}
	list := []string{}
	for name := range names {
		list = append(list, name)
	}
	sort.Strings(list)
	values := []int{-2147483648, -1, 0xd800, 0xdfff, 0x110000, 2147483647}
	for point := range points {
		values = append(values, point)
	}
	sort.Ints(values)
	data, err := json.Marshal(map[string]any{"names": list, "spaces": values})
	must(err)
	must(os.WriteFile(os.Args[2], data, 0644))
	for _, name := range list {
		fmt.Println(react.AdamicComponentBaseName(name))
	}
	for _, point := range values {
		fmt.Println(tailwind.AdamicIsSpace(rune(point)))
	}
	for point := 0; point <= 0x10ffff; point++ {
		fmt.Println(tailwind.AdamicIsSpace(rune(point)))
	}
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}
