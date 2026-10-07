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
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rules/structure"
	"os"
	"sort"
	"strings"
	"unicode/utf16"
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
	texts := map[string]bool{}
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
	points := map[int]bool{-2147483648: true, -1: true, 0x110000: true, 2147483647: true}
	for scan.Scan() {
		var row struct{ File, Source string }
		must(json.Unmarshal(scan.Bytes(), &row))
		texts[row.Source] = true
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
		parsed := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: row.File, Path: tspath.Path(row.File)}, row.Source, kind)
		var walk func(*ast.Node) bool
		walk = func(n *ast.Node) bool {
			if n == nil {
				return false
			}
			switch n.Kind {
			case ast.KindJsxText:
				texts[n.AsJsxText().Text] = true
			case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
				texts[n.Text()] = true
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
		for _, point := range row.Source {
			points[int(point)] = true
		}
	}
	must(scan.Err())
	list := []int{}
	for point := range points {
		list = append(list, point)
	}
	sort.Ints(list)
	for _, value := range []string{"", "plain é😀", "&amp;", "&amp;amp;", "&&amp;", "&;", "&", "&amp", "&#;", "&#x;", "&#X41;", "&#0;", "&#xD800;", "&#x10FFFF;", "&#1114112;", "&#99999999999999999999999999999999999999;", "&#xFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF;", "&unknown;", "&UPPER;", "&foo&amp;", "&é;", "&#-1;", "&a b;", "\x00&copy;\n", "&; &amp;;&lt;&gt;&quot;&apos;", "&#x1F600;", "& #65;", "&#65&#66;", "&#x41g;", "&!;", "&a&b;", "&amp;;"} {
		texts[value] = true
	}
	for _, name := range text.AdamicEntityNames() {
		texts["&"+name+";"] = true
		texts["a &"+name+"; é😀"] = true
	}
	textList := []string{}
	bodies := map[string]bool{}
	for value := range texts {
		textList = append(textList, value)
		for i := 0; i < len(value); i++ {
			if value[i] == '&' {
				semi := strings.IndexByte(value[i:], ';')
				if semi >= 2 {
					bodies[value[i+1:i+semi]] = true
				}
			}
		}
	}
	sort.Strings(textList)
	bodyList := []string{}
	for body := range bodies {
		bodyList = append(bodyList, body)
	}
	sort.Strings(bodyList)
	entities := []map[string]any{}
	for _, body := range bodyList {
		decoded, ok := text.AdamicDecodeEntity(body)
		entities = append(entities, map[string]any{"body": body, "text": decoded, "ok": fmt.Sprint(ok)})
	}

	data, e := json.Marshal(map[string]any{"points": list, "texts": textList, "entities": entities, "parameters": parameterRows})
	must(e)
	must(os.WriteFile(os.Args[2], data, 0644))
	for _, point := range list {
		fmt.Println(text.AdamicHexValue(rune(point)))
	}
	for point := 0; point <= 0x10ffff; point++ {
		fmt.Println(text.AdamicHexValue(rune(point)))
	}
	render := func(value string) string {
		out := ""
		for _, unit := range utf16.Encode([]rune(value)) {
			out += fmt.Sprintf("%d,", unit)
		}
		return out
	}
	for _, value := range textList {
		decoded, calls := text.AdamicUnescapeObservation(value)
		rendered := []string{}
		for _, body := range calls {
			rendered = append(rendered, render(body))
		}
		fmt.Println(render(decoded) + ":" + strings.Join(rendered, "|"))
	}

	for _, line := range parameterWant {
		fmt.Println(line)
	}
}
