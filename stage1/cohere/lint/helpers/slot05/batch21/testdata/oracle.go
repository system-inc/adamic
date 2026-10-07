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

type corpus struct {
	Mode         string                  `json:"mode"`
	Dependencies esregexp.AdamicDeps     `json:"dependencies"`
	Samples      []esregexp.AdamicSample `json:"samples"`
	Flags        []bool                  `json:"flags"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := "github.com/system-inc/cohere/internal/lint/ecmascript/regexp." + map[string]string{"join": "joinRanges", "extras": "caseExtras", "build": "buildCaseTables"}[mode]
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

	ordered := []string{}
	for source := range sources {
		ordered = append(ordered, source)
	}
	sort.Strings(ordered)
	samples := []esregexp.AdamicSample{}
	seen := map[string]bool{}
	add := func(atoms []esregexp.AdamicAtom, u bool) {
		if atoms == nil {
			atoms = []esregexp.AdamicAtom{}
		}
		s := esregexp.AdamicSample{Atoms: atoms, Unicode: u}
		data, e := json.Marshal(s)
		must(e)
		key := string(data)
		if !seen[key] {
			seen[key] = true
			samples = append(samples, s)
		}
	}
	parsed := 0
	for _, source := range ordered {
		for _, u := range []bool{false, true} {
			raw := []esregexp.AdamicAtom{}
			for _, r := range source {
				k := 0
				if r == '-' {
					k = 3
				}
				raw = append(raw, esregexp.AdamicAtom{Kind: k, Lo: int(r)})
			}
			add(raw, u)
			atoms, ok := esregexp.AdamicParse(source, u)
			if ok {
				parsed++
				add(atoms, u)
			}
		}
	}
	values := []int{-1, 0, 45, 65, 90, 97, 122, 0x17f, 0x390, 0x3b0, 0x1fd3, 0x1fe3, 0x212a, 0xfb05, 0xfb06, 0xd800, 0xffff, 0x10400, 0x10428, 0x10ffff, 0x110000}
	endpoints := []esregexp.AdamicAtom{}
	for _, r := range values {
		endpoints = append(endpoints, esregexp.AdamicAtom{Kind: 0, Lo: r, Hi: r + 1})
	}
	endpoints = append(endpoints, esregexp.AdamicAtom{Kind: 1, Lo: 65, Hi: 90}, esregexp.AdamicAtom{Kind: 2, Text: `d`}, esregexp.AdamicAtom{Kind: 3, Lo: 45}, esregexp.AdamicAtom{Kind: 7, Lo: 65})
	for _, u := range []bool{false, true} {
		add(nil, u)
		for _, a := range endpoints {
			add([]esregexp.AdamicAtom{a}, u)
			for _, b := range endpoints {
				add([]esregexp.AdamicAtom{a, {Kind: 3, Lo: 45}, b}, u)
				add([]esregexp.AdamicAtom{a, {Kind: 3, Lo: 45}, b, {Kind: 3, Lo: 45}, {Kind: 0, Lo: 122}}, u)
			}
		}
		for _, g := range esregexp.AdamicDependencies().FoldGroups {
			atoms := []esregexp.AdamicAtom{}
			for _, r := range g {
				atoms = append(atoms, esregexp.AdamicAtom{Kind: 0, Lo: int(r)})
			}
			add(atoms, u)
		}
	}
	flags := []bool{false, true, true, false}
	var want strings.Builder
	if mode == "build" {
		for _, u := range flags {
			fmt.Fprintln(&want, esregexp.AdamicBuild(u))
		}
	} else {
		for _, sample := range samples {
			if mode == "join" {
				fmt.Fprintln(&want, esregexp.AdamicJoin(sample))
			} else {
				fmt.Fprintln(&want, esregexp.AdamicExtras(sample))
			}
		}
	}
	dependencies := esregexp.AdamicDependencies()
	c := corpus{mode, dependencies, samples, flags}
	if mode == "build" {
		c.Samples = []esregexp.AdamicSample{}
	}
	data, err = json.Marshal(c)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	coverage := map[string]any{"consumers": counts, "unique_strings": len(sources), "parsed_class_bodies": parsed, "sample_rows": len(samples), "build_flags": flags, "unicode_ranges": len(dependencies.Ranges), "canonical_dependencies": len(dependencies.Canonical), "fold_dependencies": len(dependencies.Fold)}
	data, err = json.MarshalIndent(coverage, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
