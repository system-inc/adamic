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

type query struct {
	Index   int  `json:"index"`
	Size    int  `json:"size"`
	Unicode bool `json:"unicode"`
	InClass bool `json:"inClass"`
	Groups  int  `json:"groups"`
	Named   bool `json:"named"`
}
type sample struct {
	Mode    string  `json:"mode"`
	Source  string  `json:"source"`
	Queries []query `json:"queries"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	root, output, symbol := os.Args[1], os.Args[2], os.Args[3]
	mode := symbol
	symbol = "github.com/system-inc/cohere/internal/lint/ecmascript/regexp." + map[string]string{"control": "decodeControlEscape", "reference": "namedBackreference", "opener": "namedGroupOpener"}[mode]

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

	for _, source := range []string{"\\k<>", "\\k<name>tail", "\\k<é>tail", "\\k<😀>tail", "\\k<line\nname>", "\\k<\x00>", "\\k<a>b>", "\\k<a\\>b>", "\\k<missing", "\\K<n>", "abc>tail", "(?<>", "(?<name>tail", "(?<é>tail", "(?<😀>tail", "(?<line\nname>", "(?<\x00>", "(?<a>b>", "(?<a\\>b>", "(?<missing", "(?<=n>tail", "(?<!n>tail", "(x?<n>", "(?<\\!>"} {
		sources[source] = true
	}
	if mode != "control" {
		all := map[string]bool{"": true}
		for source := range sources {
			if !utf8.ValidString(source) {
				panic("invalid UTF-8 source outside adapter")
			}
			all[source] = true
			for offset := range source {
				all[source[offset:]] = true
			}
		}
		ordered := []string{}
		for source := range all {
			ordered = append(ordered, source)
		}
		sort.Strings(ordered)
		corpus := []sample{}
		var expected strings.Builder
		verdicts := 0
		observe := func(source string) {
			var text string
			var width int
			var found bool
			if mode == "reference" {
				text, width, found = esregexp.AdamicNamedBackreference(source)
			} else {
				text, width, found = esregexp.AdamicNamedGroupOpener(source)
			}
			fmt.Fprintf(&expected, "%t:%d\n%s\n", found, width, text)
			verdicts++
		}
		for _, source := range ordered {
			corpus = append(corpus, sample{Mode: mode, Source: source, Queries: []query{}})
			observe(source)
		}
		corpus = append(corpus, sample{Mode: mode + "Sweep", Queries: []query{}})
		for scalar := 0; scalar <= 0x10ffff; scalar++ {
			if scalar >= 0xd800 && scalar <= 0xdfff {
				continue
			}
			prefix := "\\k<"
			if mode == "opener" {
				prefix = "(?<"
			}
			observe(prefix + string(rune(scalar)) + ">tail")
		}
		data, err = json.Marshal(corpus)
		must(err)
		must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
		must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
		data, err = json.MarshalIndent(counts, "", "  ")
		must(err)
		must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
		fmt.Printf("%d consumers, %d source suffixes and controls plus scalar sweep, %d Go verdicts\n", len(consumers), len(corpus)-1, verdicts)
		return
	}
	for _, source := range []string{"", "c", "cA", "ca", "cZ", "cz", "c0", "c9", "c_", "c!", "c😀", "cK", "😀cA", "éca", "c\x00", "\\c9", "\\c_"} {
		sources[source] = true
	}
	for scalar := 0; scalar < 128; scalar++ {
		sources["c"+string(rune(scalar))] = true
	}
	for _, scalar := range []rune{0x80, 0x100, 0x7ff, 0x800, 0xd7ff, 0xe000, 0xffff, 0x10000, 0x10ffff} {
		sources["c"+string(scalar)] = true
	}
	ordered := []string{}
	for source := range sources {
		if !utf8.ValidString(source) {
			panic("invalid UTF-8 source outside adapter")
		}
		ordered = append(ordered, source)
	}
	sort.Strings(ordered)
	corpus := []sample{}
	var expected strings.Builder
	verdicts := 0
	for _, source := range ordered {
		row := sample{Mode: mode, Source: source, Queries: []query{}}
		sizes := []int{1}
		if len(source) <= 6 {
			sizes = []int{0, 1, 2, 4}
		}
		for index := 0; index <= len(source)+1; index++ {
			for _, size := range sizes {
				for _, u := range []bool{false, true} {
					for _, c := range []bool{false, true} {
						q := query{index, size, u, c, index % 7, index%2 == 0}
						row.Queries = append(row.Queries, q)
						kind, set, r, width, negated, message := esregexp.AdamicControlEscape(source, index, size, u, c, q.Groups, q.Named)
						fmt.Fprintf(&expected, "%d:%d:%d:%d:%t\n%s\n", kind, set, r, width, negated, message)
						verdicts++
					}
				}
			}
		}
		corpus = append(corpus, row)
	}
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d strings and controls, %d Go verdicts\n", len(consumers), len(corpus), verdicts)
}
