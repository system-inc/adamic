package main

import (
	"encoding/json"
	"fmt"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type row struct {
	Input         []int  `json:"input"`
	Deps          []int  `json:"deps"`
	IdentityError string `json:"identityError"`
}
type corpus struct {
	Mode    string                  `json:"mode"`
	Sources []esregexp.AdamicSource `json:"sources"`
	Rows    []row                   `json:"rows"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := "github.com/system-inc/cohere/internal/lint/ecmascript/regexp." + map[string]string{"property": "decodePropertyEscape", "numeric": "decodeNumericEscape", "class": "readClass"}[mode]
	data, err := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/helpers/readiness.json"))
	must(err)
	var readiness struct {
		Remaining []struct {
			Rule    string   `json:"rule"`
			Helpers []string `json:"remaining_helpers"`
		} `json:"remaining"`
	}
	must(json.Unmarshal(data, &readiness))
	data, err = os.ReadFile(filepath.Join(root, "stage1/cohere/lint/inventory/inventory.json"))
	must(err)
	var inventory struct {
		Rules []struct {
			Name  string `json:"name"`
			Tests struct {
				Files []string `json:"files"`
			} `json:"tests"`
		} `json:"rules"`
	}
	must(json.Unmarshal(data, &inventory))
	consumers := map[string]bool{}
	for _, r := range readiness.Remaining {
		for _, h := range r.Helpers {
			if h == symbol {
				consumers[r.Rule] = true
			}
		}
	}
	if len(consumers) == 0 {
		panic("no consumers")
	}
	sources := map[string]bool{}
	counts := map[string]int{}
	for _, r := range inventory.Rules {
		if !consumers[r.Name] {
			continue
		}
		for _, path := range r.Tests.Files {
			file, err := goparser.ParseFile(token.NewFileSet(), filepath.Join(root, path), nil, 0)
			must(err)
			goast.Inspect(file, func(n goast.Node) bool {
				lit, ok := n.(*goast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				source, err := strconv.Unquote(lit.Value)
				must(err)
				if len(source) > 0 {
					sources[source] = true
					counts[r.Name]++
				}
				return true
			})
		}
		if counts[r.Name] == 0 {
			panic("no source strings for " + r.Name)
		}
	}

	originalStrings := len(sources)
	for _, s := range []string{"", "[", "[^", "[]", "[^]", "[a]", "[a\u0085z]tail", "[\\]x]", "[\\😀]tail", "[^\\😀]tail", "[\\", "[\\\xff]", "[\xff]", "[\xc0\xaf]", "[a\\]", "p", "P", "p{", "p{}tail", "P{Script=Greek}tail", "p{\xff}", "p{a}b}", "0", "00", "01", "0777", "3777", "400", "777", "8", "9", "10", "100", "\\123a", "0a", "0😀", strings.Repeat("9", 100), "999999999999999999999999999999999999"} {
		sources[s] = true
	}
	for _, raw := range [][]byte{{'[', 0xff, ']'}, {'[', 0xc0, 0xaf, ']'}, {'[', 0x5c, 0xff, ']'}, {'p', '{', 0xff, '}'}, {'[', 0xc2, 0x85, ']'}, {'[', 0xe2, 0x80, 0xa8, ']'}} {
		sources[string(raw)] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	cs := []esregexp.AdamicSource{}
	rows := []row{}
	var want strings.Builder
	seen := map[string]bool{}
	sourceIDs := map[string]int{}
	add := func(source string, index int, u, c bool, groups int, named bool) {
		key := fmt.Sprintf("%x|%d|%t|%t|%d|%t", source, index, u, c, groups, named)
		if seen[key] {
			return
		}
		seen[key] = true
		sourceID, exists := sourceIDs[source]
		if !exists {
			sourceID = len(cs)
			sourceIDs[source] = sourceID
			cs = append(cs, esregexp.AdamicSourceOf(source))
		}
		deps, text := esregexp.AdamicObserve(mode, source, index, u, c, groups, named)
		input := []int{sourceID, index, 0, 0, groups, 0}
		if u {
			input[2] = 1
		}
		if c {
			input[3] = 1
		}
		if named {
			input[5] = 1
		}
		ie := ""
		if mode == "property" {
			ie = esregexp.AdamicIdentityError(source, index, u, c, groups, named)
		}
		rows = append(rows, row{input, deps, ie})
		fmt.Fprintln(&want, text)
	}
	if mode == "class" {
		for _, s := range ordered {
			for i := 0; i < len(s); i++ {
				if s[i] == '[' {
					add(s[i:], 0, false, false, 0, false)
				}
			}
		}
		add("", 0, false, false, 0, false)
	} else {
		for _, s := range ordered {
			for i := 0; i < len(s); i++ {
				match := s[i] == 'p' || s[i] == 'P'
				if mode == "numeric" {
					match = s[i] >= '0' && s[i] <= '9'
				}
				if !match {
					continue
				}
				for _, u := range []bool{false, true} {
					for _, c := range []bool{false, true} {
						for _, g := range []int{0, 1, 9, 100, 1 << 20} {
							add(s, i, u, c, g, false)
						}
					}
				}
			}
		}
	}
	data, err = json.Marshal(corpus{mode, cs, rows})
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	coverage := map[string]any{"consumers": counts, "consumer_strings": originalStrings, "rows": len(rows), "source_projections": len(cs)}
	data, err = json.MarshalIndent(coverage, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
