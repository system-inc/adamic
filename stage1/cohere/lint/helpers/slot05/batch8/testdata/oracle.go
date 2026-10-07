package main

import (
	"encoding/json"
	"fmt"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	names := map[string]string{"candidate": "ParseCandidate", "variant": "ParseVariant", "loader": "LoadDesignSystem"}
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse." + names[mode]
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

	if len(consumers) != 6 {
		panic("consumer count drift")
	}

	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	var expected strings.Builder
	if mode == "loader" {
		corpus := []collapse.AdamicLoaderCase{}
		for _, source := range ordered {
			for v := 0; v < 3; v++ {
				path := collapse.AdamicLoaderFileFixture(output, source, v)
				for framework := 0; framework < 3; framework++ {
					row := collapse.AdamicLoaderCase{Mode: mode, EntryPoint: path, ResolverPresent: true, FrameworkPresent: framework != 0, Framework: []collapse.AdamicFramework{}}
					if framework == 2 {
						row.Framework = []collapse.AdamicFramework{{Name: "custom", Kind: "functional"}, {Name: "hover", Kind: "static"}, {Name: "custom", Kind: "compound"}}
					}
					expected.WriteString(collapse.AdamicLoad(&row))
					corpus = append(corpus, row)
				}
			}
		}
		path := collapse.AdamicLoaderFileFixture(output, "", 0)
		for _, row := range []collapse.AdamicLoaderCase{{Mode: mode}, {Mode: mode, EntryPoint: path}, {Mode: mode, EntryPoint: path, PackageRoot: output}, {Mode: mode, EntryPoint: filepath.Join(output, "missing.css"), ResolverPresent: true}} {
			expected.WriteString(collapse.AdamicLoad(&row))
			corpus = append(corpus, row)
		}
		data, err = json.Marshal(corpus)
		must(err)
		must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
		fmt.Printf("%d consumers, %d distinct literals, %d loader cases\n", len(consumers), len(sources), len(corpus))
	} else {
		inputs := map[string]bool{}
		for _, source := range ordered {
			inputs[source] = true
			for _, field := range strings.Fields(source) {
				inputs[field] = true
			}
		}
		controls := []string{"", "dual", "dual-2", "border-b", "sm:hover:flex", "tw:sm:hover:flex", "!flex!", "flex!", "!flex", "bg-red-500/50/", "bg-red-500/50/50", "bg-red-500/[]", "[color:red]", "[Color:red]", "[color:]", "bg-(color:--a)", "bg-(--a)", "bg-[:red]", "bg-[red]", "w-1/2", "w-1/[2]", "bg-a[foo", "not-group-hover/name", "has-hover/name", "in-hover/name", "group-hover/name", "group-[@media(print)]", "[p]", "[&_p]", "[>p]", "[@media(print)]", "[@media(print)_&]", "hover/", "hover-x", "aria-[foo]", "aria-(--x)", "aria-x]", "aria-[]", "aria-😀", "unknown-hover", "custom"}
		for _, input := range controls {
			inputs[input] = true
		}
		data, err = os.ReadFile(filepath.Join(root, "cohere/internal/lint/rules/tailwind/collapse/testdata/candidate_fixtures.json"))
		must(err)
		var fixture struct {
			Cases []struct {
				Input string `json:"input"`
			} `json:"cases"`
		}
		must(json.Unmarshal(data, &fixture))
		for _, c := range fixture.Cases {
			inputs[c.Input] = true
		}
		if mode == "variant" {
			held := []string{}
			for input := range inputs {
				held = append(held, input)
			}
			for _, input := range held {
				for _, segment := range collapse.AdamicVariantInputs(input) {
					inputs[segment] = true
				}
			}
			for _, input := range collapse.AdamicVariantControls() {
				inputs[input] = true
			}
		}
		system := collapse.AdamicParserSystem(filepath.Join(output, "parser-theme.css"))
		corpus := []collapse.AdamicParserCase{}
		orderedInputs := []string{}
		for input := range inputs {
			orderedInputs = append(orderedInputs, input)
		}
		sort.Strings(orderedInputs)
		for _, input := range orderedInputs {
			for _, prefix := range []string{"", "tw"} {
				row := collapse.AdamicParserCase{Mode: mode, Input: input, Prefix: prefix}
				expected.WriteString(collapse.AdamicParse(&row, system))
				corpus = append(corpus, row)
			}
		}
		data, err = json.Marshal(corpus)
		must(err)
		must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
		fmt.Printf("%d consumers, %d distinct literals, %d parser cases\n", len(consumers), len(sources), len(corpus))
	}
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
}
