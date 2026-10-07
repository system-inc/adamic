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

type predicate struct {
	Key       string `json:"key"`
	Namespace string `json:"namespace"`
	Ignored   bool   `json:"ignored"`
}
type sample struct {
	Mode       string                    `json:"mode"`
	State      collapse.AdamicThemeState `json:"state"`
	Namespaces []string                  `json:"namespaces"`
	Ignored    []predicate               `json:"ignored"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output := os.Args[1], os.Args[2]
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*Theme.KeysInNamespaces"
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
	corpus := []sample{}
	var expected strings.Builder
	keyCount := 0
	for _, source := range ordered {
		theme := collapse.NewTheme()
		theme.Prefix = "tw"
		for _, key := range []string{"--font-weight-bold", "--font-size-sm", "--font-weightless", "--font-ui", "--color-a", "--color-a--opacity", "--spacing", "--text-sm--line-height", "--text-shadow-sm", "--text-red", "é--value", "--😀-x", "--input-" + source + "-item"} {
			must(theme.Add(key, source+" value", 21))
		}
		must(theme.Add("--color-a", "initial", 0))
		must(theme.Add("--color-a", "readded", 2))
		state := collapse.AdamicThemeSnapshot(theme)
		lists := [][]string{{"--font", "--color", "--input", "--text", "--spacing", "--😀", "é", "--missing"}, {"--input", "--input", "", "--color"}, {}, {"--input-" + source}}
		for _, namespaces := range lists {
			row := sample{Mode: "namespaces", State: state, Namespaces: namespaces, Ignored: []predicate{}}
			for _, namespace := range namespaces {
				for _, entry := range state.Entries {
					row.Ignored = append(row.Ignored, predicate{entry.Key, namespace, collapse.AdamicIgnoredThemeKey(entry.Key, namespace)})
				}
			}
			corpus = append(corpus, row)
			keys := theme.KeysInNamespaces(namespaces)
			keyCount += len(keys)
			fmt.Fprintln(&expected, len(keys))
			for _, key := range keys {
				fmt.Fprintln(&expected, key)
			}
		}
	}
	if len(corpus) == 0 {
		panic("no corpus")
	}
	// Test Go's byte-slice refusal separately, never convert it to an empty result.
	short := sample{Mode: "namespaces", State: collapse.AdamicThemeState{Order: []string{"-"}, Entries: []collapse.AdamicThemeEntry{{Key: "-", Value: "short", Options: 0}}}, Namespaces: []string{""}, Ignored: []predicate{}}
	if !collapse.AdamicKeysRefuse(collapse.AdamicThemeRestore(short.State), short.Namespaces) {
		panic("Go short-key refusal drift")
	}
	data, err = json.Marshal([]sample{short})
	must(err)
	must(os.WriteFile(filepath.Join(output, "refusal.json"), data, 0644))
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d distinct test-file literals, %d theme/namespace cases, %d Go returned keys and a refusal\n", len(consumers), len(sources), len(corpus), keyCount)
}
