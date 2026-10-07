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

type pair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
type sample struct {
	Mode     string                    `json:"mode"`
	State    collapse.AdamicThemeState `json:"state"`
	Prefixes []pair                    `json:"prefixes"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output := os.Args[1], os.Args[2]
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*Theme.Entries"
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
	states := []collapse.AdamicThemeState{{Order: []string{}, Entries: []collapse.AdamicThemeEntry{}}, {Order: []string{"", "--missing", "--a", "--b", "--a"}, Entries: []collapse.AdamicThemeEntry{{Key: "", Value: "skip", Options: 31}, {Key: "--a", Value: "A", Options: 17}, {Key: "--b", Value: "B", Options: 4}}}}
	for _, source := range ordered {
		t := collapse.NewTheme()
		must(t.Add("--seed-a", source, 0))
		must(t.Add("--seed-b", "B", 5))
		must(t.Add("--seed-c", "C", 2))
		must(t.Add("--seed-a", source+" changed", 0))
		must(t.Add("--seed-b", "initial", 0))
		must(t.Add("--seed-b", source+" readded", 0))
		states = append(states, collapse.AdamicThemeSnapshot(t))
	}

	corpus := []sample{}
	var expected strings.Builder
	entryCount := 0
	for _, state := range states {
		for _, prefix := range []string{"", "tw", "é"} {
			theme := collapse.AdamicThemeRestore(state)
			theme.Prefix = prefix
			row := sample{Mode: "entries", State: state, Prefixes: []pair{}}
			for _, entry := range state.Entries {
				if entry.Key != "" {
					row.Prefixes = append(row.Prefixes, pair{entry.Key, theme.PrefixKey(entry.Key)})
				}
			}
			corpus = append(corpus, row)
			first := theme.Entries()
			entryCount += len(first)
			fmt.Fprintln(&expected, len(first))
			for _, entry := range first {
				fmt.Fprintln(&expected, entry.Key)
				fmt.Fprintln(&expected, entry.Value)
				fmt.Fprintln(&expected, entry.Options)
			}
			if len(first) > 0 {
				first[0] = collapse.ThemeEntry{Key: "tampered", Value: "tampered", Options: 999}
			}
			first = first[:0]
			second := theme.Entries()
			entryCount += len(second)
			fmt.Fprintln(&expected, len(second))
			for _, entry := range second {
				fmt.Fprintln(&expected, entry.Key)
				fmt.Fprintln(&expected, entry.Value)
				fmt.Fprintln(&expected, entry.Options)
			}
		}
	}
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d distinct test-file literals, %d state/prefix cases with fresh reads, %d Go entries\n", len(consumers), len(sources), len(corpus), entryCount)
}
