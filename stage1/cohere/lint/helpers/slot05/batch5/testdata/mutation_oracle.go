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

type sample struct {
	Mode  string                    `json:"mode"`
	State collapse.AdamicThemeState `json:"state"`
	Key   string                    `json:"key"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	names := map[string]string{"clear": "clearAll", "compact": "compactKeyOrder", "delete": "delete"}
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*Theme." + names[mode]
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
	states := []collapse.AdamicThemeState{}
	for _, dead := range []int{-1, 0, 1, 2, 3, 7} {
		states = append(states, collapse.AdamicThemeState{Prefix: "é", Dead: dead, Order: []string{"", "--missing", "--a", "--a", "--b", ""}, Entries: []collapse.AdamicThemeEntry{{Key: "--a", Value: "A", Options: 17}, {Key: "--b", Value: "B", Options: 4}}})
	}
	states = append(states, collapse.AdamicThemeState{Order: []string{}, Entries: []collapse.AdamicThemeEntry{}}, collapse.AdamicThemeState{Dead: 1, Order: []string{"--stale", ""}, Entries: []collapse.AdamicThemeEntry{}}, collapse.AdamicThemeState{Dead: 2, Order: []string{"--ghost", ""}, Entries: []collapse.AdamicThemeEntry{{Key: "--orphan", Value: "orphan", Options: 31}}}, collapse.AdamicThemeState{Dead: 2, Order: []string{"", ""}, Entries: []collapse.AdamicThemeEntry{{Key: "", Value: "empty key", Options: 0}}})
	for _, source := range ordered {
		t := collapse.NewTheme()
		t.Prefix = "tw"
		must(t.Add("--a", source, 17))
		must(t.Add("--b", "B", 4))
		must(t.Add("--c", "C", 2))
		must(t.Add("--é", "accent", 8))
		must(t.Add("--😀", "supplementary", 16))
		must(t.Add("--\ue000", "private", 31))
		must(t.Add("--a", source+" changed", 17))
		states = append(states, collapse.AdamicThemeSnapshot(t))
		must(t.Add("--b", "initial", 0))
		states = append(states, collapse.AdamicThemeSnapshot(t))
		must(t.Add("--b", source+" readded", 0))
		states = append(states, collapse.AdamicThemeSnapshot(t))
	}
	corpus := []sample{}
	var expected strings.Builder
	printOrder := func(order []string) {
		fmt.Fprintln(&expected, len(order))
		for _, key := range order {
			fmt.Fprintln(&expected, key)
		}
	}
	printValues := func(entries []collapse.AdamicThemeEntry) {
		fmt.Fprintln(&expected, len(entries))
		for _, v := range entries {
			fmt.Fprintln(&expected, v.Key)
			fmt.Fprintln(&expected, v.Value)
			fmt.Fprintln(&expected, v.Options)
		}
	}
	snapshots := 0
	for _, state := range states {
		keys := []string{""}
		if mode == "delete" {
			keys = []string{"--absent", "--a", "--b", "", "--😀", "--orphan"}
		}
		for _, key := range keys {
			corpus = append(corpus, sample{mode, state, key})
			trace := collapse.AdamicMutationTrace(collapse.AdamicThemeRestore(state), mode, key)
			for i, snapshot := range trace {
				if i != 1 {
					fmt.Fprintln(&expected, snapshot.Prefix)
					fmt.Fprintln(&expected, snapshot.Dead)
				}
				printOrder(snapshot.Order)
				printValues(snapshot.Entries)
				snapshots++
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
	fmt.Printf("%d consumers, %d distinct test-file literals, %d mutation cases, %d Go store/alias snapshots\n", len(consumers), len(sources), len(corpus), snapshots)
}
