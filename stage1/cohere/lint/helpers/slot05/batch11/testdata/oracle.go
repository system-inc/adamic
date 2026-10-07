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

type sample struct {
	Mode       string                `json:"mode"`
	Source     string                `json:"source"`
	Kind       int                   `json:"kind"`
	Lo         int                   `json:"lo"`
	Hi         int                   `json:"hi"`
	LoText     string                `json:"loText"`
	HiText     string                `json:"hiText"`
	Unicode    bool                  `json:"unicode"`
	IgnoreCase bool                  `json:"ignoreCase"`
	Multiline  bool                  `json:"multiline"`
	DotAll     bool                  `json:"dotAll"`
	Atoms      []esregexp.AdamicAtom `json:"atoms"`
	Result     string                `json:"result"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	root, output, symbol := os.Args[1], os.Args[2], os.Args[3]
	mode := symbol
	symbol = "github.com/system-inc/cohere/internal/lint/ecmascript/regexp." + map[string]string{"atom": "classAtom.write", "word": "wordCharacters", "expand": "expandsOnUppercase"}[mode]

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

	ordered := []string{"", "\x00", "set text", "line\ntext", "é😀", "\\d", "[]"}
	for source := range sources {
		if !utf8.ValidString(source) {
			panic("invalid UTF-8 outside adapter")
		}
		ordered = append(ordered, source)
	}
	sort.Strings(ordered)
	runes := map[rune]bool{math.MinInt32: true, -1: true, 0: true, 0xd800: true, 0xdfff: true, 0x10ffff: true, 0x110000: true, math.MaxInt32: true}
	for _, source := range ordered {
		for _, r := range source {
			runes[r] = true
		}
	}
	for _, r := range []rune{0x1f80, 0x1f87, 0x1f90, 0x1f97, 0x1fa0, 0x1fa7, 0x1fb3, 0x1fc3, 0x1ff3} {
		for delta := rune(-1); delta <= 1; delta++ {
			runes[r+delta] = true
		}
	}
	values := []int{}
	for r := range runes {
		values = append(values, int(r))
	}
	sort.Ints(values)
	corpus := []sample{}
	var expected strings.Builder
	verdicts := 0
	if mode == "expand" {
		for _, r := range values {
			corpus = append(corpus, sample{Mode: mode, Lo: r})
			fmt.Fprintln(&expected, esregexp.AdamicExpandsOnUppercase(rune(r)))
			verdicts++
		}
		corpus = append(corpus, sample{Mode: "sweep"})
		for r := rune(0); r <= 0x10ffff; r++ {
			fmt.Fprintln(&expected, esregexp.AdamicExpandsOnUppercase(r))
			verdicts++
		}
	} else if mode == "atom" {
		observe := func(kind, lo, hi int, text string) {
			result, trace := esregexp.AdamicClassAtomWrite(kind, rune(lo), rune(hi), text)
			corpus = append(corpus, sample{Mode: mode, Kind: kind, Lo: lo, Hi: hi, Source: text, LoText: esregexp.AdamicLiteralRune(rune(lo)), HiText: esregexp.AdamicLiteralRune(rune(hi))})
			fmt.Fprintf(&expected, "%s\n%s\n", result, trace)
			verdicts++
		}
		for j, r := range values {
			for kind := 0; kind < 256; kind++ {
				observe(kind, r, values[(j+1)%len(values)], ordered[j%len(ordered)])
			}
		}
		for _, text := range ordered {
			observe(2, 65, 90, text)
		}
	} else if mode == "word" {
		for _, u := range []bool{false, true} {
			for _, i := range []bool{false, true} {
				for _, m := range []bool{false, true} {
					for _, d := range []bool{false, true} {
						result, trace, atoms := esregexp.AdamicWordCharacters(u, i, m, d)
						corpus = append(corpus, sample{Mode: mode, Unicode: u, IgnoreCase: i, Multiline: m, DotAll: d, Atoms: atoms, Result: result})
						fmt.Fprintf(&expected, "%s\n%s\n", result, trace)
						verdicts++
					}
				}
			}
		}
	} else {
		panic("unknown mode")
	}
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d captured strings, %d Go queries\n", len(consumers), len(ordered), verdicts)
}
