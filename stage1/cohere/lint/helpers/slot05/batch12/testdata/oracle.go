// Executed through a Go overlay rooted in cohere. Production cohere is unchanged.
package main

import (
	"encoding/json"
	"fmt"
	esregexp "github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

type corpus struct {
	Mode  string               `json:"mode"`
	Upper esregexp.AdamicTable `json:"upper"`
	Fold  esregexp.AdamicTable `json:"fold"`
	Runes []int                `json:"runes"`
	Flags []bool               `json:"flags"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	root, output, symbol := os.Args[1], os.Args[2], os.Args[3]
	mode := symbol
	symbol = "github.com/system-inc/cohere/internal/lint/ecmascript/regexp." + map[string]string{"tables": "caseTables", "equivalents": "CaseEquivalents", "groups": "CaseEquivalenceGroups"}[mode]

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

	runes := map[int]bool{math.MinInt32: true, -1: true, 0: true, 0xd800: true, 0xdfff: true, 0x10ffff: true, 0x110000: true, math.MaxInt32: true}
	for source := range sources {
		if !utf8.ValidString(source) {
			panic("invalid UTF-8 outside adapter")
		}
		for _, r := range source {
			runes[int(r)] = true
		}
	}
	for _, r := range []int{0x39f, 0x390, 0x1fd3, 0x3b0, 0x1fe3, 0xfb05, 0xfb06, 0x212a, 0x17f, 0x10400, 0x10428} {
		runes[r] = true
	}
	values := []int{}
	for r := range runes {
		values = append(values, r)
	}
	sort.Ints(values)
	upper, fold := esregexp.AdamicTables(false), esregexp.AdamicTables(true)
	flags := []bool{false, true, true, false, false, true, false, true}
	c := corpus{mode, upper, fold, values, flags}
	var expected strings.Builder
	queries := 0
	if mode == "equivalents" {
		for _, u := range []bool{false, true} {
			for _, r := range values {
				fmt.Fprintln(&expected, esregexp.AdamicObserve(mode, rune(r), u))
				queries++
			}
			for r := rune(0); r <= 0x10ffff; r++ {
				fmt.Fprintln(&expected, esregexp.AdamicObserve(mode, r, u))
				queries++
			}
		}
	} else {
		for _, u := range flags {
			fmt.Fprintln(&expected, esregexp.AdamicObserve(mode, 0, u))
			queries++
		}
	}
	data, err = json.Marshal(c)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d captured strings, %d signed-rune controls, %d upper groups, %d fold groups, %d Go queries\n", len(consumers), len(sources), len(values), len(upper.Groups), len(fold.Groups), queries)
}
