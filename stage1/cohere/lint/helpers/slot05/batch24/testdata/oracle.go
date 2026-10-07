package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Row struct {
	Source  string `json:"source"`
	Options int    `json:"options"`
	Ops     []int  `json:"ops"`
	Bytes   []int  `json:"bytes"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := "github.com/system-inc/cohere/internal/lint/ecmascript/" + map[string]string{"rewrite": "regexp.rewrite", "hex": "regexsyntax.IsHexDigit", "all": "regexsyntax.AllHexDigits"}[mode]
	data, e := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/helpers/readiness.json"))
	must(e)
	var ready struct {
		Remaining []struct {
			Rule    string   `json:"rule"`
			Helpers []string `json:"remaining_helpers"`
		} `json:"remaining"`
	}
	must(json.Unmarshal(data, &ready))
	consumers := map[string]bool{}
	for _, r := range ready.Remaining {
		for _, h := range r.Helpers {
			if h == symbol {
				consumers[r.Rule] = true
			}
		}
	}
	if len(consumers) == 0 {
		panic("missing consumers")
	}
	data, e = os.ReadFile(filepath.Join(root, "stage1/cohere/lint/inventory/inventory.json"))
	must(e)
	var inventory struct {
		Rules []struct {
			Name  string `json:"name"`
			Tests struct {
				Files []string `json:"files"`
			} `json:"tests"`
		} `json:"rules"`
	}
	must(json.Unmarshal(data, &inventory))
	counts := map[string]int{}
	sources := map[string]bool{}
	files := map[string][]string{}
	for _, r := range inventory.Rules {
		if !consumers[r.Name] {
			continue
		}
		for _, p := range r.Tests.Files {
			files[r.Name] = append(files[r.Name], p)
			tree, e := parser.ParseFile(token.NewFileSet(), filepath.Join(root, p), nil, 0)
			must(e)
			ast.Inspect(tree, func(n ast.Node) bool {
				v, ok := n.(*ast.BasicLit)
				if ok && v.Kind == token.STRING {
					s, e := strconv.Unquote(v.Value)
					must(e)
					sources[s] = true
					counts[r.Name]++
				}
				return true
			})
		}
		if counts[r.Name] == 0 {
			panic("no strings for " + r.Name)
		}
	}
	captured := len(sources)
	for _, s := range []string{"", "a", "A", "0", "F", "g", "abcF09", "00g", "f0F", "α", "😀", "^$.", "\\w\\W\\p{L}\\P{Greek}", "\\b+", "\\B?", "\\1(a)", "(?<n>a)\\k<n>", "(?i-m:a.(?-i:b)$)^", "(?s:.)(?-s:.)", "(?m:^$)(?-m:^$)", "(?=a)+", "(?<=a)+", "(?!a){2}", "(?<!a)*", "(?i)", "[]", "[^]", "[a-z]", "[z-a]", "[\\w]", "[\\p{L}]", "[\\B]", "\\u{1f600}", "\\x41", "\\", "(?<n>", "((?i:a)b)c", "^+", "$?"} {
		sources[s] = true
	}
	for b := 0; b < 256; b++ {
		sources[string([]byte{byte(b)})] = true
		sources["a"+string([]byte{byte(b)})+"F"] = true
	}
	for _, b := range [][]byte{{0xff, 92, 'w'}, {0xc0, 0xaf, 46}, {0xf0, 0x9f, 0x98, 0x80}} {
		sources[string(b)] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	rows := []Row{}
	operationPool := []esregexp.AdamicOp{}
	operationIDs := map[string]int{}
	pool := func(ops []esregexp.AdamicOp) []int {
		ids := []int{}
		for _, op := range ops {
			encoded, e := json.Marshal(op)
			must(e)
			key := string(encoded)
			id, ok := operationIDs[key]
			if !ok {
				id = len(operationPool)
				operationIDs[key] = id
				operationPool = append(operationPool, op)
			}
			ids = append(ids, id)
		}
		return ids
	}
	var want strings.Builder
	addHex := func(b byte) {
		rows = append(rows, Row{Bytes: []int{int(b)}, Ops: []int{}})
		fmt.Fprintln(&want, regexsyntax.IsHexDigit(b))
	}
	if mode == "hex" {
		for b := 0; b < 256; b++ {
			addHex(byte(b))
		}
	}
	for _, s := range ordered {
		if mode == "rewrite" {
			for options := 0; options < 16; options++ {
				ops, result := esregexp.AdamicRewrite(s, options)
				rows = append(rows, Row{Source: hex.EncodeToString([]byte(s)), Options: options, Ops: pool(ops), Bytes: []int{}})
				fmt.Fprintln(&want, result)
			}
		} else if mode == "all" {
			b := []int{}
			for _, v := range []byte(s) {
				b = append(b, int(v))
			}
			rows = append(rows, Row{Bytes: b, Ops: []int{}})
			fmt.Fprintln(&want, regexsyntax.AllHexDigits(s))
		} else {
			for _, b := range []byte(s) {
				addHex(b)
			}
		}
	}
	data, e = json.Marshal(map[string]any{"mode": mode, "rows": rows, "operationPool": operationPool})
	must(e)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, e = json.MarshalIndent(map[string]any{"consumers": counts, "files": files, "captured_strings": captured, "sources": len(ordered), "rows": len(rows), "unique_operations": len(operationPool), "options": "all sixteen flag combinations for every rewrite source", "hex": "all 256 bytes plus every byte in captured consumer strings"}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
