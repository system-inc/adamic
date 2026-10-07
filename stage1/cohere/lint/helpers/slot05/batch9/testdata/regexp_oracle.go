// Executed through a Go overlay rooted in cohere. Production cohere is unchanged.
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
	"unicode/utf8"
)

type sample struct {
	Mode      string `json:"mode"`
	Source    string `json:"source"`
	Positions []int  `json:"positions"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	root, output, symbol := os.Args[1], os.Args[2], os.Args[3]
	mode := symbol
	symbol = "github.com/system-inc/cohere/internal/lint/ecmascript/regexp." + map[string]string{"decimal": "isDecimalDigit", "groups": "countGroups"}[mode]
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

	for _, source := range []string{"", "0123456789", "😀0", "é9", "Σ5", "٠١٢٣٤٥٦٧٨٩", "０１２３４５６７８９", "(a)(b)", "(?<name>x)(?<x)", "(?<=x)(?<!y)(?:a)(?=b)(?!c)", "[()]([\\)])", "\\(x\\)(?<n>a)", "\\😀(a)", "\\", "[", "](", "[[(]](", "(?<", "(?<=", "(?<!", "(?", "(??)", "[\\](]", "\x00\\(😀)"} {
		sources[source] = true
	}
	for b := 0; b < 128; b++ {
		sources[string(rune(b))] = true
		sources["\\"+string(rune(b))+"(a)"] = true
	}
	ordered := []string{}
	for source := range sources {
		if !utf8.ValidString(source) {
			panic("invalid UTF-8 source outside string adapter")
		}
		ordered = append(ordered, source)
	}
	sort.Strings(ordered)
	corpus := []sample{}
	var expected strings.Builder
	verdicts := 0
	for _, source := range ordered {
		row := sample{Mode: mode, Source: source, Positions: []int{}}
		if mode == "decimal" {
			for i := 0; i <= len(source)+2; i++ {
				row.Positions = append(row.Positions, i)
				fmt.Fprintln(&expected, esregexp.AdamicDecimalDigit(source, i))
				verdicts++
			}
		} else {
			count, named := esregexp.AdamicCountGroups(source)
			fmt.Fprintln(&expected, count)
			fmt.Fprintln(&expected, named)
			verdicts++
		}
		corpus = append(corpus, row)
	}
	sweepMode := "decimalSweep"
	if mode == "groups" {
		sweepMode = "groupSweep"
	}
	corpus = append(corpus, sample{Mode: sweepMode, Positions: []int{}})
	for scalar := 0; scalar <= 0x10ffff; scalar++ {
		if scalar >= 0xd800 && scalar <= 0xdfff {
			continue
		}
		source := string(rune(scalar))
		if mode == "decimal" {
			fmt.Fprintln(&expected, esregexp.AdamicDecimalDigit(source, 0))
			fmt.Fprintln(&expected, esregexp.AdamicDecimalDigit(source, len(source)))
			verdicts += 2
		} else {
			count, named := esregexp.AdamicCountGroups("(" + "\\" + source + "(?<named>x)(?<=x)(?:x)")
			fmt.Fprintln(&expected, count)
			fmt.Fprintln(&expected, named)
			verdicts++
		}
	}
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d strings and controls plus scalar sweep, %d Go verdicts\n", len(consumers), len(corpus)-1, verdicts)
}
