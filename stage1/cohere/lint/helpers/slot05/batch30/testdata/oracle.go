package main

import (
	"encoding/json"
	"fmt"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type Row struct {
	Keys    []string `json:"keys"`
	Present bool     `json:"present"`
	Key     string   `json:"key"`
	Suffix  string   `json:"suffix"`
	Guard   bool     `json:"guard"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse." + map[string]string{"color": "colorArm", "theme": "themeArm", "width": "widthArm"}[mode]

	data, e := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/helpers/readiness.json"))
	must(e)
	var ready struct {
		Remaining []struct {
			Rule    string   `json:"rule"`
			Helpers []string `json:"remaining_helpers"`
		} `json:"remaining"`
	}
	must(json.Unmarshal(data, &ready))
	consumers := map[string]bool{}
	for _, r := range ready.Remaining {
		for _, h := range r.Helpers {
			if h == symbol {
				consumers[r.Rule] = true
			}
		}
	}
	if len(consumers) == 0 {
		panic("missing consumers")
	}
	data, e = os.ReadFile(filepath.Join(root, "stage1/cohere/lint/inventory/inventory.json"))
	must(e)
	var inventory struct {
		Rules []struct {
			Name         string `json:"name"`
			Dependencies []struct {
				Symbol string `json:"symbol"`
			} `json:"dependencies"`
			Tests struct {
				Files []string `json:"files"`
			} `json:"tests"`
		} `json:"rules"`
	}
	must(json.Unmarshal(data, &inventory))
	// Include all inventory consumers, even those already ported outside the blocked cohort.
	for _, r := range inventory.Rules {
		for _, d := range r.Dependencies {
			if d.Symbol == symbol {
				consumers[r.Name] = true
			}
		}
	}
	counts := map[string]int{}
	sources := map[string]bool{}
	files := map[string][]string{}
	for _, r := range inventory.Rules {
		if !consumers[r.Name] {
			continue
		}
		for _, p := range r.Tests.Files {
			files[r.Name] = append(files[r.Name], p)
			tree, e := parser.ParseFile(token.NewFileSet(), filepath.Join(root, p), nil, 0)
			must(e)
			ast.Inspect(tree, func(n ast.Node) bool {
				v, ok := n.(*ast.BasicLit)
				if ok && v.Kind == token.STRING {
					s, e := strconv.Unquote(v.Value)
					must(e)
					sources[s] = true
					counts[r.Name]++
				}
				return true
			})
		}
		if counts[r.Name] == 0 {
			panic("no strings for " + r.Name)
		}
	}
	captured := len(sources)

	for _, value := range []string{"", "--color", "--border-color", "--border-width", "--text", "--font", "--font-weight", "--shadow", "--outline-width", "--ring-width", "--inset-ring-width", "--stroke-width", "--text-decoration-thickness", "猫😀", "a\x00b", "a\nb", "~", "px", "ms"} {
		sources[value] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	rows := []Row{}
	var want strings.Builder
	add := func(row Row) {
		keys := append([]string{}, row.Keys...)
		if !row.Present {
			keys = nil
		}
		call := func() collapse.FunctionalUtilityArm {
			if mode == "color" {
				return collapse.AdamicColor(keys)
			}
			if mode == "theme" {
				return collapse.AdamicTheme(row.Key)
			}
			return collapse.AdamicWidth(row.Key, row.Suffix, row.Guard)
		}
		first, second := call(), call()
		inferred := []string{}
		for _, v := range first.InferTypes {
			inferred = append(inferred, string(v))
		}
		fmt.Fprintf(&want, "%t|%d|%s|%t|%s|%s|%t|%t|%s|%t\n", first.ThemeKeys != nil, len(first.ThemeKeys), strings.Join(first.ThemeKeys, "~"), first.IsColor, first.BareValue, first.BareValueSuffix, first.RefusesModifier, first.InferTypes != nil, strings.Join(inferred, "~"), first.PercentagePassesThrough)
		if len(first.ThemeKeys) > 0 {
			first.ThemeKeys[0] = "__alias_probe__"
		}
		left, right := "none", "none"
		if len(second.ThemeKeys) > 0 {
			left = second.ThemeKeys[0]
		}
		if len(keys) > 0 {
			right = keys[0]
		}
		fmt.Fprintf(&want, "%s|%s\n", left, right)
		rows = append(rows, row)
	}
	for _, value := range ordered {
		for variant := 0; variant < 6; variant++ {
			row := Row{Keys: []string{value, "--color", value}, Present: true, Key: value, Suffix: []string{"", "px", "ms", "猫😀", value, "%"}[variant], Guard: variant%2 == 0}
			if variant == 0 {
				row.Keys = []string{}
				row.Present = false
			}
			if variant == 1 {
				row.Keys = []string{}
			}
			if variant == 2 {
				row.Keys = []string{value}
			}
			add(row)
		}
	}
	data, e = json.Marshal(map[string]any{"mode": mode, "rows": rows})
	must(e)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, e = json.MarshalIndent(map[string]any{"consumers": counts, "files": files, "captured_strings": captured, "sources": len(ordered), "rows": len(rows), "boundary": "all inventory consumer strings as constructor parameters; nil and empty-present color keys, order and duplicates, flags and suffixes; color slice aliases and independently fresh single-key arrays"}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
