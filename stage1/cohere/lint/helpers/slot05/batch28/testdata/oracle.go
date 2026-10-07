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
	Value       string `json:"value"`
	Kind        string `json:"kind"`
	Positive    bool   `json:"positive"`
	Opacity     bool   `json:"opacity"`
	Strict      bool   `json:"strict"`
	Spacing     bool   `json:"spacing"`
	FontStretch bool   `json:"fontStretch"`
	Quarter     bool   `json:"quarter"`
	Half        bool   `json:"half"`
	Whole       bool   `json:"whole"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse." + map[string]string{"predicate": "bareValuePredicate", "transform": "bareValueTransform", "opacity": "isValidOpacityValue"}[mode]

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
	// Project possible candidate spellings without computing their verdicts.
	originals := []string{}
	for s := range sources {
		originals = append(originals, s)
	}
	for _, s := range originals {
		for _, part := range strings.FieldsFunc(s, func(r rune) bool {
			return !(r >= '0' && r <= '9' || r == '.' || r == '%' || r == '-' || r == '+' || r == 'e' || r == 'E')
		}) {
			sources[part] = true
		}
	}
	for _, s := range []string{"", "0", "-0", "00", "01", "+1", "0.25", "0.5", "0.75", "1", "1.25", "1.50", "-0.25", "1.3", "NaN", "Inf", "+Inf", "Infinity", " 1", "1 ", "1e3", "0x1p2", "50%", "49%", "50.0%", "200%", "201%", "50.25%", "猫😀", "a\x00b", "a\nb"} {
		sources[s] = true
	}
	for i := 0; i < 128; i++ {
		sources[string(rune(i))] = true
		sources["1"+string(rune(i))+"2"] = true
	}
	for i := -8; i <= 808; i++ {
		sources[strconv.FormatFloat(float64(i)/4, 'f', -1, 64)] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	rows := []Row{}
	var want strings.Builder
	kinds := []string{"", "PositiveInteger", "Opacity", "StrictPositiveInteger", "SpacingMultiplier", "FontStretchPercentage", "GridRepeat", "Fraction", "unknown", "positiveinteger", "Opacity\x00", "猫"}
	for _, value := range ordered {
		deps := collapse.AdamicDependencies(value)
		row := Row{Value: value, Positive: deps[0], Opacity: deps[1], Strict: deps[2], Spacing: deps[3], FontStretch: deps[4], Quarter: collapse.AdamicMultiple(value, 0.25), Half: collapse.AdamicMultiple(value, 0.5), Whole: collapse.AdamicMultiple(value, 1)}
		if mode == "opacity" {
			rows = append(rows, row)
			fmt.Fprintln(&want, collapse.AdamicOpacity(value))
			continue
		}
		for _, kind := range kinds {
			row.Kind = kind
			rows = append(rows, row)
			if mode == "transform" {
				fmt.Fprintln(&want, collapse.AdamicTransform(kind, value))
			} else {
				name, answer := collapse.AdamicPredicate(kind, value)
				if name == "none" {
					fmt.Fprintln(&want, name)
				} else {
					fmt.Fprintf(&want, "%s|%t\n", name, answer)
				}
			}
		}
	}
	data, e = json.Marshal(map[string]any{"mode": mode, "rows": rows})
	must(e)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, e = json.MarshalIndent(map[string]any{"consumers": counts, "files": files, "captured_strings": captured, "sources": len(ordered), "rows": len(rows), "boundary": "all captured strings and candidate projections; ASCII bytes, Unicode/NUL/newline, quarter increments -2..202, every named kind and unknown kinds; real Go dependency answers"}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
