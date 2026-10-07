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

type ignored struct {
	Key       string `json:"key"`
	Namespace string `json:"namespace"`
	Ignored   string `json:"ignored"`
}
type sample struct {
	Mode       string                    `json:"mode"`
	Source     string                    `json:"source"`
	State      collapse.AdamicThemeState `json:"state"`
	Candidate  string                    `json:"candidate"`
	Present    string                    `json:"present"`
	Namespaces []string                  `json:"namespaces"`
	Ignored    []ignored                 `json:"ignored"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	names := map[string]string{"key": "resolveKey", "value": "ResolveValue"}
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.*Theme." + names[mode]
	if mode == "new" {
		symbol = "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse.NewTheme"
	}
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
	snapshots := 0
	printState := func(t *collapse.Theme) {
		state := collapse.AdamicThemeSnapshot(t)
		fmt.Fprintln(&expected, state.Prefix)
		fmt.Fprintln(&expected, state.Dead)
		fmt.Fprintln(&expected, len(state.Order))
		for _, k := range state.Order {
			fmt.Fprintln(&expected, k)
		}
		fmt.Fprintln(&expected, len(state.Entries))
		for _, v := range state.Entries {
			fmt.Fprintln(&expected, v.Key)
			fmt.Fprintln(&expected, v.Value)
			fmt.Fprintln(&expected, v.Options)
		}
		snapshots++
	}
	ordered = append(ordered, "")
	appendLookup := func(t *collapse.Theme, candidate string, present bool, namespaces []string) {
		state := collapse.AdamicThemeSnapshot(t)
		predictions := []ignored{}
		for _, namespace := range namespaces {
			// Include absent lookup keys: a semantic mutant must not die from missing oracle data.
			for _, key := range []string{namespace, namespace + "-" + candidate, namespace + "-" + strings.ReplaceAll(candidate, ".", "_")} {
				predictions = append(predictions, ignored{key, namespace, strconv.FormatBool(collapse.AdamicIgnoredKey(key, namespace))})
			}
		}
		corpus = append(corpus, sample{Mode: mode, State: state, Candidate: candidate, Present: strconv.FormatBool(present), Namespaces: namespaces, Ignored: predictions})
		var value string
		var found bool
		if mode == "key" {
			value, found = collapse.AdamicResolveKey(t, candidate, present, namespaces)
		} else {
			value, found = t.ResolveValue(candidate, present, namespaces)
		}
		fmt.Fprintln(&expected, found)
		fmt.Fprintln(&expected, value)
	}
	for _, source := range ordered {
		if mode == "new" {
			corpus = append(corpus, sample{Mode: mode, Source: source})
			first, second := collapse.NewTheme(), collapse.NewTheme()
			printState(first)
			printState(second)
			first.Prefix = "tw"
			must(first.Add("--a", source+" stored", 17))
			printState(first)
			printState(second)
			continue
		}
		candidate := "seed." + source + ".end"
		for _, exact := range []bool{true, false} {
			t := collapse.NewTheme()
			t.Prefix = "tw"
			must(t.Add("--first", "namespace first", 31))
			must(t.Add("--second", "", 4))
			if exact {
				must(t.Add("--first-"+candidate, source, 31))
			}
			must(t.Add("--first-"+strings.ReplaceAll(candidate, ".", "_"), source+" fallback", 2))
			must(t.Add("--second-"+candidate, source+" second", 0))
			for _, namespaces := range [][]string{{"--missing", "--first", "--second"}, {"--second", "--first"}, {"--first", "--first"}, {}} {
				for _, present := range []bool{true, false} {
					appendLookup(t, candidate, present, namespaces)
				}
			}
		}
	}
	if mode != "new" {
		t := collapse.NewTheme()
		for _, v := range []collapse.AdamicThemeEntry{{Key: "--font", Value: "font"}, {Key: "--font-weight-bold", Value: "bold", Options: 31}, {Key: "--font-weightless", Value: "kept"}, {Key: "--font-size-sm", Value: "sm"}, {Key: "--first-a_b_c", Value: "fallback"}, {Key: "--first-a.b.c", Value: "exact", Options: 31}, {Key: "--first-", Value: ""}, {Key: "--first-bold", Value: "other"}, {Key: "--first", Value: "namespace"}} {
			must(t.Add(v.Key, v.Value, collapse.ThemeOptions(v.Options)))
		}
		for _, candidate := range []string{"weight-bold", "weightless", "size-sm", "a.b.c", "", "missing", "bold", "é.😀", "\x00"} {
			for _, namespaces := range [][]string{{"--font", "--font-weight", "--first"}, {"--first", "--font"}, {"--font-weight", "--font"}, {"--absent"}, {}} {
				for _, present := range []bool{true, false} {
					appendLookup(t, candidate, present, namespaces)
				}
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
	fmt.Printf("%d consumers, %d distinct test-file literals, %d cases, %d constructor snapshots\n", len(consumers), len(sources), len(corpus), snapshots)
}
