package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"io"
	"os"
	"regexp"
	"sort"
	"unicode/utf16"
)

type node struct {
	Kind     string `json:"kind"`
	Value    string `json:"value"`
	Children []int  `json:"children"`
}
type arena struct {
	Nodes []node `json:"nodes"`
	Roots []int  `json:"roots"`
}

func flatten(values []collapse.ValueNode) arena {
	a := arena{Nodes: []node{}, Roots: []int{}}
	var walk func([]collapse.ValueNode) []int
	walk = func(values []collapse.ValueNode) []int {
		ids := []int{}
		for _, v := range values {
			i := len(a.Nodes)
			ids = append(ids, i)
			a.Nodes = append(a.Nodes, node{string(v.Kind), v.Value, []int{}})
			children := walk(v.Nodes)
			a.Nodes[i].Children = children
		}
		return ids
	}
	a.Roots = walk(values)
	return a
}
func units(text string) string {
	out := ""
	for _, v := range utf16.Encode([]rune(text)) {
		out += fmt.Sprintf("%d,", v)
	}
	return out
}
func snapshot(a arena) string {
	out := ""
	for _, n := range a.Nodes {
		out += n.Kind + ":" + units(n.Value) + "|"
	}
	return out
}

type decodeCase struct {
	Input string `json:"input"`
	Ast   arena  `json:"ast"`
	Css   string `json:"css"`
	Math  string `json:"math"`
}
type treeCase struct {
	Ast arena `json:"ast"`
}
type breakpointCase struct {
	Name         string   `json:"name"`
	Present      bool     `json:"present"`
	GroupPresent bool     `json:"groupPresent"`
	Order        int      `json:"order"`
	Keys         []string `json:"keys"`
}

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()
	zip, err := gzip.NewReader(file)
	if err != nil {
		panic(err)
	}
	defer zip.Close()
	decoder := json.NewDecoder(zip)
	texts := map[string]bool{}
	for _, s := range []string{"", "a_b", `a\_b`, "😀_a", "url(a_b)", "x_url(a_b)", "var(--my_var,_a_b)", "theme(--my_var,_a_b)", "calc(1px+2px)", "calc(var(--a_b)+1px)", "fn(var(--a_b,_c_d),url(a_b))", "URL(a_b)", "var( _a_b)", "var((a_b),c_d)", "fn(a_b)", "unclosed(a_b"} {
		texts[s] = true
	}
	pattern := regexp.MustCompile(`[[:alnum:]_@.\-]+`)
	for {
		var row struct{ Rule, File, Source string }
		err = decoder.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		texts[row.Source] = true
		for _, s := range pattern.FindAllString(row.Source, -1) {
			texts[s] = true
		}
	}
	ordered := []string{}
	for s := range texts {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	cases := []decodeCase{}
	expected := []string{}
	for _, s := range ordered {
		original, decoded, css, math, result := collapse.AdamicDecode(s)
		cases = append(cases, decodeCase{s, flatten(original), css, math})
		expected = append(expected, snapshot(flatten(decoded)), units(result))
	}
	trees := []treeCase{}
	for _, kind := range []string{"function", "word", "separator", "unknown"} {
		for _, name := range []string{"url", "a_url", "var", "a_var", "theme", "a_theme", "URL", "foo_bar"} {
			for _, firstKind := range []string{"word", "separator", "function"} {
				v := []collapse.ValueNode{{Kind: collapse.ValueNodeKind(kind), Value: name, Nodes: []collapse.ValueNode{{Kind: collapse.ValueNodeKind(firstKind), Value: "--a_b", Nodes: []collapse.ValueNode{{Kind: "word", Value: `c\_d_e`}}}, {Kind: "word", Value: "_f_g"}, {Kind: "function", Value: "foo_bar", Nodes: []collapse.ValueNode{{Kind: "word", Value: "h_i"}}}}}}
				trees = append(trees, treeCase{flatten(v)})
				collapse.AdamicRecursive(v)
				expected = append(expected, snapshot(flatten(v)))
			}
		}
	}
	breaks := []breakpointCase{}
	for _, name := range ordered {
		for _, present := range []bool{false, true} {
			for _, group := range []bool{false, true} {
				for _, order := range []int{-17, 0, 64, 999} {
					theme := collapse.NewTheme()
					theme.Add("--breakpoint-"+name, "50rem", 0)
					theme.Add("--breakpoint-existing", "40rem", 0)
					theme.Add("--container-ignored", "24rem", 0)
					keys := theme.KeysInNamespaces([]string{"--breakpoint"})
					breaks = append(breaks, breakpointCase{name, present, group, order, keys})
					registry := collapse.NewVariantRegistry()
					registry.Register("existing", "functional")
					registry.Register("sm", "compound")
					var passed *collapse.Theme
					if present {
						passed = theme
					}
					collapse.AdamicBreakpoint(registry, passed, order, group)
					out := ""
					for _, r := range registry.Registrations() {
						out += units(r.Name) + ":" + fmt.Sprint(r.Order) + ":" + string(r.Kind) + "|"
					}
					expected = append(expected, out)
				}
			}
		}
	}
	out, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(out).Encode(map[string]any{"decode": cases, "trees": trees, "breakpoints": breaks}); err != nil {
		panic(err)
	}
	if err = out.Close(); err != nil {
		panic(err)
	}
	for _, line := range expected {
		fmt.Println(line)
	}
}
