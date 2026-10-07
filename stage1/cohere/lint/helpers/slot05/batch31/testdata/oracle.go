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
	Value       string   `json:"value"`
	Keys        []string `json:"keys"`
	KeysPresent bool     `json:"keysPresent"`
	Guard       bool     `json:"guard"`
	Resolved    string   `json:"resolved"`
	Found       bool     `json:"found"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse." + map[string]string{"border": "borderSideDescription", "mask": "maskStopDescription", "resolve": "resolveArmColor"}[mode]

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

	for _, value := range []string{"", "inherit", "transparent", "current", "currentcolor", "red-500", "red", "1.5", "1_5", "INHERIT", "Transparent", "CURRENT", "inherit ", "var(--x)", "猫😀", "a\x00b", "a\nb"} {
		sources[value] = true
	}
	ordered := []string{}
	for v := range sources {
		ordered = append(ordered, v)
	}
	sort.Strings(ordered)
	rows := []Row{}
	var want strings.Builder
	print := func(d *collapse.FunctionalUtilityDescription) {
		fmt.Fprintf(&want, "%t|%s|%t|%t|%s|%t|%t|%t|%t|%t|%t|%d\n", d.ThemeKeys != nil, strings.Join(d.ThemeKeys, "~"), d.SupportsNegative, d.SupportsFractions, d.DefaultValue, d.DefaultValuePresent, d.HandleBareValue != nil, d.HandleNegativeBareValue != nil, d.StaticValueNames != nil, d.AcceptsModifierOnArbitrary, d.Arms != nil, len(d.Arms))
		for _, a := range d.Arms {
			inferred := []string{}
			for _, v := range a.InferTypes {
				inferred = append(inferred, string(v))
			}
			fmt.Fprintf(&want, "%t|%s|%t|%s|%s|%t|%t|%s|%t\n", a.ThemeKeys != nil, strings.Join(a.ThemeKeys, "~"), a.IsColor, a.BareValue, a.BareValueSuffix, a.RefusesModifier, a.InferTypes != nil, strings.Join(inferred, "~"), a.PercentagePassesThrough)
		}
	}
	for _, value := range ordered {
		if mode != "resolve" {
			call := collapse.AdamicBorder
			if mode == "mask" {
				call = collapse.AdamicMask
			}
			first, second := call(), call()
			print(first)
			first.ThemeKeys = append(first.ThemeKeys, value)
			if first.StaticValueNames != nil {
				first.StaticValueNames[value] = true
			}
			for i := range first.Arms {
				a := &first.Arms[i]
				if len(a.ThemeKeys) > 0 {
					a.ThemeKeys[0] = value
				}
				if len(a.InferTypes) > 0 {
					a.InferTypes[0] = collapse.DataType(value)
				}
			}
			print(second)
			rows = append(rows, Row{Value: value, Keys: []string{}})
			continue
		}
		for variant := 0; variant < 8; variant++ {
			keys := []string{"--border-color", "--color"}
			if variant == 0 {
				keys = nil
			}
			if variant == 1 {
				keys = []string{}
			}
			if variant == 6 {
				keys = []string{"--color", "--border-color"}
			}
			if variant == 7 {
				keys = []string{"--unknown", "--color", "--color"}
			}
			theme := collapse.NewTheme()
			if variant == 3 {
				theme.Prefix = "tw"
			}
			if variant > 1 {
				for _, keyword := range []string{"inherit", "transparent", "current", "red-500", "red", "1_5", ""} {
					primary := "primary"
					if variant == 4 {
						primary = ""
					}
					options := collapse.ThemeOptionInline
					if variant == 3 {
						options = collapse.ThemeOptionReference
					}
					if variant == 5 {
						options = collapse.ThemeOptionNone
					}
					must(theme.Add("--border-color-"+keyword, primary, options))
					must(theme.Add("--color-"+keyword, "secondary", options))
				}
			}
			dependency, found := theme.Resolve(value, true, keys, 0)
			answer, ok := collapse.AdamicResolve(value, variant%2 == 0, keys, theme)
			fmt.Fprintf(&want, "%t|%s\n", ok, answer)
			row := Row{Value: value, Keys: keys, KeysPresent: keys != nil, Guard: variant%2 == 0, Resolved: dependency, Found: found}
			if row.Keys == nil {
				row.Keys = []string{}
			}
			rows = append(rows, row)
		}
	}
	data, e = json.Marshal(map[string]any{"mode": mode, "rows": rows})
	must(e)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, e = json.MarshalIndent(map[string]any{"consumers": counts, "files": files, "captured_strings": captured, "sources": len(ordered), "rows": len(rows), "boundary": "all inventory consumer strings as value/freshness probes; descriptor zero fields, arm order/namespaces/inference and fresh arrays; keywords versus actual Go empty/inline/reference/prefixed/ordered themes"}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
