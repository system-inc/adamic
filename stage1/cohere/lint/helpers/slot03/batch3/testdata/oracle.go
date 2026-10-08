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
	"github.com/system-inc/cohere/internal/lint/rules/structure"
	"os"
	"strings"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	f, e := os.Open(os.Args[1])
	must(e)
	defer f.Close()
	z, e := gzip.NewReader(f)
	must(e)
	defer z.Close()
	scan := bufio.NewScanner(z)
	scan.Buffer(make([]byte, 4096), 16<<20)
	parameterRows := []map[string]any{}
	parameterWant := []string{}
	nodeIDs := map[*ast.Node]int{}
	nextID := 0
	nodeID := func(n *ast.Node) int {
		if n == nil {
			return -1
		}
		if id, ok := nodeIDs[n]; ok {
			return id
		}
		id := nextID
		nextID++
		nodeIDs[n] = id
		return id
	}
	renderNodes := func(nodes []*ast.Node) string {
		ids := []string{}
		for _, node := range nodes {
			ids = append(ids, fmt.Sprint(nodeID(node)))
		}
		return strings.Join(ids, ",")
	}
	observeParameters := func(parameters *ast.NodeList, input []*ast.Node) {
		ids := []int{}
		for _, n := range input {
			ids = append(ids, nodeID(n))
		}
		parameterRows = append(parameterRows, map[string]any{"present": fmt.Sprint(parameters != nil), "nodes": ids})
		result := structure.AdamicParameterNodes(parameters)
		parameterWant = append(parameterWant, renderNodes(result))
		var first *ast.Node
		if len(input) > 0 {
			first = input[0]
		}
		if len(result) > 0 {
			result[0] = nil
		}
		parameterWant = append(parameterWant, renderNodes(input), renderNodes(structure.AdamicParameterNodes(parameters)))
		if len(input) > 0 {
			input[0] = first
		}
	}
	observeParameters(nil, []*ast.Node{{Kind: ast.KindParameter}, {Kind: ast.KindParameter}})
	observeParameters(&ast.NodeList{}, nil)
	observeParameters(&ast.NodeList{Nodes: []*ast.Node{}}, []*ast.Node{})
	controlA, controlB := &ast.Node{Kind: ast.KindParameter}, &ast.Node{Kind: ast.KindParameter}
	controlList := &ast.NodeList{Nodes: []*ast.Node{controlA, nil, controlB, controlA}}
	observeParameters(controlList, controlList.Nodes)
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
		row.File = tspath.NormalizePath(row.File)
		if !tspath.IsRootedDiskPath(row.File) {
			row.File = "/" + row.File
		}
		parsed := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute(row.File), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute(row.File))}, row.Source, kind)
		var walk func(*ast.Node) bool
		walk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			if data := n.FunctionLikeData(); data != nil {
				var input []*ast.Node
				if data.Parameters != nil {
					input = data.Parameters.Nodes
				}
				observeParameters(data.Parameters, input)
			}
			n.ForEachChild(walk)
			return false
		}
		walk(parsed.AsNode())

	}
	must(scan.Err())
	data, e := json.Marshal(map[string]any{"parameters": parameterRows})
	must(e)
	must(os.WriteFile(os.Args[2], data, 0644))

	for _, line := range parameterWant {
		fmt.Println(line)
	}
}
