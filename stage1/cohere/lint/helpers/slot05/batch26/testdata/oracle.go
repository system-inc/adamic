package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
	reactrules "github.com/system-inc/cohere/internal/lint/rules/react"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type Row struct {
	Bytes     []int `json:"bytes"`
	Widths    []int `json:"widths"`
	Positions []int `json:"positions"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := map[string]string{"escape": "github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax.SkipPatternEscape", "class": "github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax.ClassEnd", "react": "github.com/system-inc/cohere/internal/lint/rules/react.isReactComponentBaseName"}[mode]

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
			Name         string `json:"name"`
			Dependencies []struct {
				Symbol string `json:"symbol"`
			} `json:"dependencies"`
			Tests struct {
				Files []string `json:"files"`
			} `json:"tests"`
		} `json:"rules"`
	}
	must(json.Unmarshal(data, &inventory))
	// Include all inventory consumers, even those already ported outside the blocked cohort.
	for _, r := range inventory.Rules {
		for _, d := range r.Dependencies {
			if d.Symbol == symbol {
				consumers[r.Name] = true
			}
		}
	}
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
	for _, s := range []string{"", "\\", "\\x", "\\x00", "\\x0g", "\\u1234", "\\u123g", "\\u{}", "\\u{1f600}", "\\u{]}", "\\u{", "\\c", "\\cA", "\\p{]}", "\\P{Greek}", "\\q{x]y}", "\\q{", "[]", "[^]", "[[]]", "[[a]b]", "[\\]a]", "[\\q{x]y}]", "[a", "[\\", "[😀]", "α[😀]", "\\😀", "\\α", "Component", "PureComponent", "ComponentX", "PureComponentX", "component", "pureComponent", "React.Component", " Component", "Component\x00"} {
		sources[s] = true
	}
	for b := 0; b < 256; b++ {
		for _, s := range []string{string([]byte{byte(b)}), "\\" + string([]byte{byte(b)}), "[" + string([]byte{byte(b)}) + "]", "Component" + string([]byte{byte(b)})} {
			sources[s] = true
		}
	}
	for _, s := range []string{string([]byte{'[', 255, ']'}), string([]byte{'\\', 0xc0, 0xaf, ']'}), string([]byte{'[', 0xf0, 0x9f, 0x98, 0x80, ']'})} {
		sources[s] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	rows := []Row{}
	var want strings.Builder
	calls := 0
	for _, s := range ordered {
		values := []int{}
		widths := []int{}
		positions := []int{0, len(s), len(s) + 1}
		seen := map[int]bool{0: true, len(s): true, len(s) + 1: true}
		for i, b := range []byte(s) {
			values = append(values, int(b))
			_, w := utf8.DecodeRuneInString(s[i:])
			widths = append(widths, w)
			if (b == '\\' || b == '[') && !seen[i] {
				positions = append(positions, i)
				seen[i] = true
			}
		}
		sort.Ints(positions)
		rows = append(rows, Row{values, widths, positions})
		if mode == "react" {
			fmt.Fprintln(&want, reactrules.AdamicBaseName(s))
			calls++
			continue
		}
		for _, start := range positions {
			for bits := 0; bits < 4; bits++ {
				flags := regexsyntax.RegexFlags{Unicode: bits%2 == 1, UnicodeSets: bits >= 2}
				var n int
				var ok bool
				if mode == "escape" {
					n, ok = regexsyntax.SkipPatternEscape(s, start, flags)
				} else {
					n, ok = regexsyntax.ClassEnd(s, start, flags)
				}
				fmt.Fprintf(&want, "%d|%v\n", n, ok)
				calls++
			}
		}
	}
	data, e = json.Marshal(map[string]any{"mode": mode, "rows": rows})
	must(e)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, e = json.MarshalIndent(map[string]any{"consumers": counts, "files": files, "captured_strings": captured, "sources": len(ordered), "rows": len(rows), "calls": calls, "boundary": "all 256 bytes, all four flag field combinations, every backslash/bracket in consumer strings, zero/EOF/past EOF, malformed UTF-8, nested classes and brace escapes"}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
