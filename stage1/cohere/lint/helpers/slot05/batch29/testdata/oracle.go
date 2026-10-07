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
	Value          string                 `json:"value"`
	ValuePresent   bool                   `json:"valuePresent"`
	Fraction       string                 `json:"fraction"`
	Theme          []string               `json:"theme"`
	ThemePresent   bool                   `json:"themePresent"`
	Negative       bool                   `json:"negative"`
	Fractions      bool                   `json:"fractions"`
	Default        string                 `json:"default"`
	DefaultPresent bool                   `json:"defaultPresent"`
	Kind           string                 `json:"kind"`
	Suffix         string                 `json:"suffix"`
	Names          []string               `json:"names"`
	ParseOK        bool                   `json:"parseOK"`
	Parsed         string                 `json:"parsed"`
	Normal         collapse.AdamicHandler `json:"normal"`
	NoSuffix       collapse.AdamicHandler `json:"nosuffix"`
	None           collapse.AdamicHandler `json:"none"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse." + map[string]string{"functional": "FrameworkFunctionalUtility.Description", "multi": "FrameworkMultiDeclarationUtility.Description", "font": "isFontStretchPercentage"}[mode]

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
	// Preserve every captured string; project numeric percentage candidates without verdicts.
	originals := []string{}
	for value := range sources {
		originals = append(originals, value)
	}
	for _, value := range originals {
		for _, part := range strings.FieldsFunc(value, func(r rune) bool {
			return !(r >= '0' && r <= '9' || r == '.' || r == '+' || r == '-' || r == 'e' || r == 'E' || r == '%')
		}) {
			sources[part] = true
			sources[part+"%"] = true
		}
	}
	for _, value := range []string{"", "%", "50", "50x", "50%", "49%", "200%", "201%", "50.0%", "50.25%", "-50%", "+50%", "5e1%", "0x1.9p5%", "NaN%", "Inf%", "+Inf%", "-Inf%", "Infinity%", " 50%", "50 %", "bad%", "5_0%", "_50%", "50_%", "5__0%", "0x1.9_p5%", "49.99999999999999%", "49.999999999999999%", "50.00000000000001%", "199.99999999999997%", "200.00000000000003%", "50%%", "1e999%", "猫😀%", "a\x00b%", "a\nb%"} {
		sources[value] = true
	}
	for i := -4; i <= 804; i++ {
		s := strconv.FormatFloat(float64(i)/4, 'f', -1, 64)
		sources[s+"%"] = true
	}
	for i := 0; i < 128; i++ {
		sources["50"+string(rune(i))+"%"] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	rows := []Row{}
	var want strings.Builder
	kinds := []string{"", "PositiveInteger", "Opacity", "StrictPositiveInteger", "SpacingMultiplier", "FontStretchPercentage", "GridRepeat", "Fraction", "unknown"}
	add := func(row Row) {
		if row.Theme == nil {
			row.Theme = []string{}
		}
		if row.Names == nil {
			row.Names = []string{}
		}
		var value *collapse.ParsedValue
		if row.ValuePresent {
			value = &collapse.ParsedValue{Value: row.Value, Fraction: row.Fraction}
		}
		row.Normal = collapse.AdamicHandle(row.Kind, row.Suffix, value)
		row.NoSuffix = collapse.AdamicHandle(row.Kind, "", value)
		row.None = collapse.AdamicHandle("", row.Suffix, value)
		rows = append(rows, row)
		theme := append([]string{}, row.Theme...)
		if !row.ThemePresent {
			theme = nil
		}
		description := func() *collapse.FunctionalUtilityDescription {
			if mode == "functional" {
				return collapse.AdamicFunctional(theme, row.Negative, row.DefaultPresent, row.Default, row.Kind, row.Suffix, row.Names)
			}
			return collapse.AdamicMulti(theme, row.Negative, row.Fractions, row.DefaultPresent, row.Default, row.Kind, row.Suffix)
		}
		d := description()
		second := description()
		fmt.Fprintf(&want, "%t|%s|%t|%t|%t|%s\n", d.ThemeKeys != nil, strings.Join(d.ThemeKeys, "~"), d.SupportsNegative, d.SupportsFractions, d.DefaultValuePresent, d.DefaultValue)
		fmt.Fprintf(&want, "%t|%t|%t|%t|%t\n", d.HandleBareValue != nil, d.StaticValueNames != nil, d.HandleNegativeBareValue != nil, d.AcceptsModifierOnArbitrary, d.Arms != nil)
		keys := []string{}
		for name := range d.StaticValueNames {
			keys = append(keys, name)
		}
		sort.Strings(keys)
		printed := []string{}
		for _, key := range keys {
			printed = append(printed, fmt.Sprintf("%s:%t", key, d.StaticValueNames[key]))
		}
		fmt.Fprintln(&want, strings.Join(printed, "~"))
		if d.HandleBareValue != nil {
			s, ok := d.HandleBareValue(value)
			fmt.Fprintf(&want, "%t|%s\n", ok, s)
		} else {
			fmt.Fprintln(&want, "none")
		}
		if d.StaticValueNames != nil {
			d.StaticValueNames["__owned_freshness_probe__"] = true
		}
		fmt.Fprintln(&want, second.StaticValueNames["__owned_freshness_probe__"])
		if len(theme) > 0 {
			theme[0] = "__owned_alias_probe__"
		}
		if len(d.ThemeKeys) > 0 {
			fmt.Fprintln(&want, d.ThemeKeys[0])
		} else {
			fmt.Fprintln(&want, "none")
		}
	}
	for _, value := range ordered {
		if mode == "font" {
			prefix := value
			if len(prefix) > 0 {
				prefix = prefix[:len(prefix)-1]
			}
			parsed, ok := collapse.AdamicParse(prefix)
			if parsed == "+Inf" {
				parsed = "Infinity"
			}
			if parsed == "-Inf" {
				parsed = "-Infinity"
			}
			rows = append(rows, Row{Value: value, ParseOK: ok, Parsed: parsed, Theme: []string{}, Names: []string{}})
			answer := collapse.AdamicFont(value)
			fmt.Fprintln(&want, answer)
			continue
		}
		for variant := 0; variant < 9; variant++ {
			row := Row{Value: "1", ValuePresent: variant != 0, Fraction: "2/3", Theme: []string{value, "--x", "--x"}, ThemePresent: true, Negative: variant&1 != 0, Fractions: variant&2 != 0, Default: value, DefaultPresent: variant&4 != 0, Kind: kinds[variant], Suffix: "px", Names: []string{value, "auto", value}}
			if variant == 0 {
				row.Theme = []string{}
				row.ThemePresent = false
				row.Names = []string{}
			}
			if variant == 1 {
				row.Theme = []string{}
				row.Names = []string{}
			}
			if variant == 5 {
				row.Value = "50%"
				row.Suffix = "%"
			}
			if variant == 6 {
				row.Suffix = ""
			}
			add(row)
		}
	}
	// Include every real framework table row, isolated from the table's mutable slices.
	if mode == "functional" {
		roots := []string{}
		for root := range collapse.FrameworkFunctionalUtilities {
			roots = append(roots, root)
		}
		sort.Strings(roots)
		for _, root := range roots {
			u := collapse.FrameworkFunctionalUtilities[root]
			names := []string{}
			for _, v := range u.StaticValues {
				names = append(names, v.Name)
			}
			for _, value := range []string{"1", "0", "50%", "bad"} {
				add(Row{Value: value, ValuePresent: true, Fraction: "2/3", Theme: append([]string{}, u.ThemeKeys...), ThemePresent: u.ThemeKeys != nil, Negative: u.SupportsNegative, Default: u.DefaultValue, DefaultPresent: u.DefaultValuePresent, Kind: string(u.BareValue), Suffix: u.BareValueSuffix, Names: names})
			}
		}
	}
	if mode == "multi" {
		roots := []string{}
		for root := range collapse.FrameworkMultiDeclarationUtilities {
			roots = append(roots, root)
		}
		sort.Strings(roots)
		for _, root := range roots {
			u := collapse.FrameworkMultiDeclarationUtilities[root]
			for _, value := range []string{"1", "0", "50%", "bad"} {
				add(Row{Value: value, ValuePresent: true, Fraction: "2/3", Theme: append([]string{}, u.ThemeKeys...), ThemePresent: u.ThemeKeys != nil, Negative: u.SupportsNegative, Fractions: u.SupportsFractions, Default: u.DefaultValue, DefaultPresent: u.DefaultValuePresent, Kind: string(u.BareValue), Suffix: u.BareValueSuffix, Names: []string{}})
			}
		}
	}
	data, e = json.Marshal(map[string]any{"mode": mode, "rows": rows})
	must(e)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, e = json.MarshalIndent(map[string]any{"consumers": counts, "files": files, "captured_strings": captured, "sources": len(ordered), "rows": len(rows), "boundary": "all captured strings; percentage quarter increments -1..201 and ASCII suffix bytes; description flag/kind scenarios and every live framework table row; theme aliases and fresh static-name maps; actual Go handler/float dependencies"}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
