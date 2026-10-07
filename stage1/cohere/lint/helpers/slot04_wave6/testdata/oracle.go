package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"os"
	"strings"
)

type Row struct {
	Name, Source             string
	Bytes, Prefix, Key, Root []int
	Keys                     [][]int
}
type Case struct {
	Prefix, Key, Root []int
	Keys              [][]int
}

func ints(s string) []int {
	out := []int{}
	for _, b := range []byte(s) {
		out = append(out, int(b))
	}
	return out
}
func text(v []int) string {
	out := []byte{}
	for _, b := range v {
		if b < 0 || b > 255 {
			panic("byte")
		}
		out = append(out, byte(b))
	}
	return string(out)
}
func byteText(s string) string {
	pieces := []string{}
	for _, b := range []byte(s) {
		pieces = append(pieces, fmt.Sprint(b))
	}
	return strings.Join(pieces, ",")
}
func observeCase(c Case) {
	prefix, key, root := text(c.Prefix), text(c.Key), text(c.Root)
	before, after := collapse.AdamicPrefixSequence(prefix, string([]byte{137}))
	fmt.Printf("prefix %s\n", byteText(before))
	if prefix == "" || len(key) >= 2 {
		fmt.Printf("key %s\n", byteText(collapse.AdamicPrefixKey(prefix, key)))
	} else {
		fmt.Println("key rejected")
	}
	fmt.Printf("changed %s\n", byteText(after))
	keys := []string{}
	for _, k := range c.Keys {
		keys = append(keys, text(k))
	}
	initial, deleted, added := collapse.AdamicHasSequence(keys, root)
	fmt.Printf("has %t\ndeleted %t\nadded %t\n", initial, deleted, added)
}
func makeCase(input []byte) Case {
	value := ints(string(input))
	return Case{value, value, value, [][]int{value}}
}
func observe(input []byte) { observeCase(makeCase(input)) }
func main() {
	argument := os.Args[1]
	if argument == "--bad0" || argument == "--bad1" {
		key := ""
		if argument == "--bad1" {
			key = "a"
		}
		fmt.Println(collapse.AdamicPrefixKey("tw", key))
		return
	}
	if argument == "--full" {
		observe(nil)
		for first := 0; first <= 255; first++ {
			observe([]byte{byte(first)})
			for last := 0; last <= 255; last++ {
				observe([]byte{byte(first), byte(last)})
			}
		}
		return
	}
	adapt := argument == "--cases"
	if adapt {
		argument = os.Args[2]
	}
	data, err := os.ReadFile(argument)
	if err != nil {
		panic(err)
	}
	var rows []Row
	if err = json.Unmarshal(data, &rows); err != nil {
		panic(err)
	}
	values := [][]byte{}
	cases := []Case{}
	for _, row := range rows {
		if row.Name == "control:call" {
			cases = append(cases, Case{row.Prefix, row.Key, row.Root, row.Keys})
			continue
		}
		if strings.HasPrefix(row.Name, "control:") {
			bytes := []byte{}
			for _, value := range row.Bytes {
				if value < 0 || value > 255 {
					panic("invalid byte control")
				}
				bytes = append(bytes, byte(value))
			}
			values = append(values, bytes)
			continue
		}
		values = append(values, []byte(row.Source))
		kind := core.ScriptKindTS
		switch {
		case strings.HasSuffix(row.Name, ".tsx"):
			kind = core.ScriptKindTSX
		case strings.HasSuffix(row.Name, ".jsx"):
			kind = core.ScriptKindJSX
		case strings.HasSuffix(row.Name, ".js"):
			kind = core.ScriptKindJS
		}
		path := tspath.NormalizePath("/corpus/" + row.Name)
		file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: path, Path: tspath.Path(path)}, row.Source, kind)
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			if n.Kind == ast.KindStringLiteral || n.Kind == ast.KindNoSubstitutionTemplateLiteral {
				values = append(values, []byte(n.Text()))
			}
			n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
		}
		visit(file.AsNode())
	}
	for _, value := range values {
		cases = append(cases, makeCase(value))
	}
	if adapt {
		if err = json.NewEncoder(os.Stdout).Encode(cases); err != nil {
			panic(err)
		}
		return
	}
	for _, c := range cases {
		observeCase(c)
	}
}
