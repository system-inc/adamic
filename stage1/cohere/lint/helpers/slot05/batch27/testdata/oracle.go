package main

import (
	"encoding/json"
	"fmt"
	tailwind "github.com/system-inc/cohere/internal/lint/rules/tailwind"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Row struct {
	Bytes []int `json:"bytes"`
	Range []int `json:"range"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := map[string]string{"positive": "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.isPositiveInteger", "strict": "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.isStrictPositiveInteger", "trim": "github.com/system-inc/cohere/internal/lint/rules/tailwind.trimDelimiters"}[mode]

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
	// Project bare numeric values from captured class source without imposing a verdict.
	originals := []string{}
	for s := range sources {
		originals = append(originals, s)
	}
	for _, s := range originals {
		for _, part := range strings.FieldsFunc(s, func(r rune) bool { return r < '0' || r > '9' }) {
			sources[part] = true
		}
	}
	for _, s := range []string{"", "0", "00", "01", "1", "-0", "-1", "+1", "1.0", "1e3", "0x10", " 1", "1 ", "9007199254740991", "9007199254740992", "9007199254740993", "1000000000000000100", "1000000000000000128"} {
		sources[s] = true
	}
	for i := 0; i < 256; i++ {
		sources[string([]byte{byte(i)})] = true
		sources["1"+string([]byte{byte(i)})+"2"] = true
	}
	for n := 0; n <= 310; n++ {
		sources["1"+strings.Repeat("0", n)] = true
		sources[strings.Repeat("9", n)] = true
	}
	for i := int64(-128); i <= 128; i++ {
		sources[strconv.FormatInt(9007199254740992+i, 10)] = true
	}
	state := uint64(0x123456789abcdef)
	for i := 0; i < 4096; i++ {
		state = state*6364136223846793005 + 1
		v := math.Float64frombits(state & 0x7fffffffffffffff)
		if math.IsInf(v, 0) || math.IsNaN(v) || v < 1 {
			continue
		}
		v = math.Trunc(v)
		canonical := strconv.FormatFloat(v, 'f', -1, 64)
		sources[canonical] = true
		if len(canonical) < 310 {
			sources[canonical+"0"] = true
		}
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	rows := []Row{}
	var want strings.Builder
	addRange := func(start, end, leading, trailing int) {
		rows = append(rows, Row{Bytes: []int{}, Range: []int{start, end, leading, trailing}})
		a, b := tailwind.AdamicTrim(start, end, leading, trailing)
		fmt.Fprintf(&want, "%d|%d\n", a, b)
	}
	if mode == "trim" {
		for _, s := range ordered {
			for _, start := range []int{0, 3, 10} {
				for _, pair := range [][2]int{{0, 0}, {1, 1}, {1, 2}, {2, 2}, {len(s) + 1, 1}} {
					addRange(start, start+len(s), pair[0], pair[1])
				}
			}
		}
		for start := -3; start <= 3; start++ {
			for end := -3; end <= 12; end++ {
				for leading := -2; leading <= 5; leading++ {
					for trailing := -2; trailing <= 5; trailing++ {
						addRange(start, end, leading, trailing)
					}
				}
			}
		}
	} else {
		for _, s := range ordered {
			v := []int{}
			for _, b := range []byte(s) {
				v = append(v, int(b))
			}
			rows = append(rows, Row{Bytes: v, Range: []int{}})
			fmt.Fprintln(&want, collapse.AdamicPositive(s, mode == "strict"))
		}
	}
	data, e = json.Marshal(map[string]any{"mode": mode, "rows": rows})
	must(e)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, e = json.MarshalIndent(map[string]any{"consumers": counts, "files": files, "captured_strings": captured, "sources": len(ordered), "rows": len(rows), "boundary": "all 256 bytes; decimal powers through overflow, 2^53 neighbors, 4096 deterministic float-bit samples; byte-length range projections and exhaustive small signed range parameters"}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
