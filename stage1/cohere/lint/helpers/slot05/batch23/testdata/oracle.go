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
	Input   []int               `json:"input"`
	Ops     []esregexp.AdamicOp `json:"-"`
	Subject string              `json:"subject"`
}
type corpus struct {
	Mode          string                  `json:"mode"`
	Sources       []esregexp.AdamicSource `json:"sources"`
	Rows          []row                   `json:"rows"`
	OperationSets [][]esregexp.AdamicOp   `json:"operationSets"`
	OperationIDs  []int                   `json:"operationIDs"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := "github.com/system-inc/cohere/internal/lint/ecmascript/regexp." + map[string]string{"decode": "decodeEscape", "atoms": "classAtoms", "test": "*RegExp.Test"}[mode]
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

	original := len(sources)
	for _, s := range []string{"", "\\", "\\u", "\\u{1f600}", "\\u00ff", "\\x41", "\\x", "\\c_", "\\c0", "\\cZ", "\\d\\D\\s\\S\\w\\W", "\\p{L}\\P{Greek}", "\\b\\B\\k", "\\0\\1\\8\\9", "a-z", "z-a", "\\w-z", "z-\\W", "[a]", "-", "\\q", "\\😀"} {
		sources[s] = true
	}
	for _, raw := range [][]byte{{92, 0xff}, {0xff, 45, 0xc2, 0x85}, {92, 0xc0, 0xaf}} {
		sources[string(raw)] = true
	}
	for _, raw := range [][]byte{{92, 'w', 92, 'W', 92, 'p', '{', 'L', '}'}, {92, 'b', 92, 'B', 92, 'k'}, {92, 'u', '{', '1', 'f', '6', '0', '0', '}'}, {92, 'x', '4', '1'}, {92, '0', 92, '1', 92, '8', 92, '9'}} {
		sources[string(raw)] = true
	}
	for b := 0; b <= 255; b++ {
		sources[string([]byte{92, byte(b)})] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	cs := []esregexp.AdamicSource{}
	sourceIDs := map[string]int{}
	sourceID := func(s string) int {
		id, exists := sourceIDs[s]
		if !exists {
			id = len(cs)
			sourceIDs[s] = id
			cs = append(cs, esregexp.AdamicSourceOf(s))
		}
		return id
	}
	rows := []row{}
	var want strings.Builder
	add := func(s string, i int, u, c, n bool, g int, ignore bool) {
		ops, text := esregexp.AdamicObserve(mode, s, i, u, c, n, g, ignore)
		input := []int{sourceID(s), i, 0, 0, 0, g, 0}
		if u {
			input[2] = 1
		}
		if c {
			input[3] = 1
		}
		if n {
			input[4] = 1
		}
		if ignore {
			input[6] = 1
		}
		rows = append(rows, row{input, ops, ""})
		fmt.Fprintln(&want, text)
	}
	invalidPatterns := 0
	if mode == "test" {
		id := 0
		for _, pattern := range append(ordered, "a", "^a$", "(a+)+$", "[a-z]+", "(?=a)a") {
			for _, subject := range []string{"", "a", "A", "abc", "😀", "123", pattern} {
				id++
				ops, text, ok := esregexp.AdamicTest(pattern, subject, id, 2, false, false)
				state := 2
				if !ok {
					invalidPatterns++
					state = 0
					ops, text, _ = esregexp.AdamicTest(pattern, subject, id, state, false, false)
				}
				rows = append(rows, row{[]int{0, id, state}, ops, fmt.Sprintf("%x", subject)})
				fmt.Fprintln(&want, text)
			}
		}
		for _, state := range []int{0, 1} {
			ops, text, _ := esregexp.AdamicTest("a", "a", 1, state, false, false)
			rows = append(rows, row{[]int{0, 1, state}, ops, "61"})
			fmt.Fprintln(&want, text)
		}
		ops, text, _ := esregexp.AdamicTest("a", "a", 2, 2, true, false)
		rows = append(rows, row{[]int{0, 2, 2}, ops, "61"})
		fmt.Fprintln(&want, text)
		subject := strings.Repeat("a", 256) + "!"
		ops, text, _ = esregexp.AdamicTest("(a+)+$", subject, 3, 2, false, true)
		rows = append(rows, row{[]int{0, 3, 2}, ops, fmt.Sprintf("%x", subject)})
		fmt.Fprintln(&want, text)
	} else {
		for _, s := range ordered {
			indices := []int{0}
			if mode == "decode" {
				indices = []int{len(s)}
				for i := 0; i < len(s); i++ {
					if s[i] == 92 {
						indices = append(indices, i+1)
					}
				}
			}
			for _, i := range indices {
				for _, u := range []bool{false, true} {
					for _, c := range []bool{false, true} {
						for _, n := range []bool{false, true} {
							for _, g := range []int{0, 9} {
								if mode == "atoms" {
									for _, ignore := range []bool{false, true} {
										add(s, i, u, c, n, g, ignore)
									}
								} else {
									add(s, i, u, c, n, g, false)
								}
							}
						}
					}
				}
			}
		}
	}
	sets := [][]esregexp.AdamicOp{}
	ids := []int{}
	interned := map[string]int{}
	for _, row := range rows {
		ops := row.Ops
		sort.Slice(ops, func(i, j int) bool { return ops[i].Key < ops[j].Key })
		encoded, e := json.Marshal(ops)
		must(e)
		key := string(encoded)
		id, exists := interned[key]
		if !exists {
			id = len(sets)
			interned[key] = id
			sets = append(sets, ops)
		}
		ids = append(ids, id)
	}
	data, err = json.Marshal(corpus{mode, cs, rows, sets, ids})
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	coverage := map[string]any{"consumers": counts, "consumer_strings": original, "rows": len(rows), "sources": len(cs), "unique_operation_sets": len(sets), "invalid_pattern_attempts": invalidPatterns}
	data, err = json.MarshalIndent(coverage, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
