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
	"strconv"
	"strings"
	"unicode/utf16"
)

type childCase struct {
	Present  bool  `json:"present"`
	Nodes    []int `json:"nodes"`
	Capacity int   `json:"capacity"`
}
type orderCase struct {
	GroupPresent bool `json:"groupPresent"`
	GroupOrder   int  `json:"groupOrder"`
	LastOrder    int  `json:"lastOrder"`
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
	add := func(text string) { texts[text] = true }
	for _, text := range []string{"", "@", "@media", ".flex", " :where(& > *) ", "@media (width >= 40rem)", "é", "😀", "\x00", "\n", " leading ", "--a-b"} {
		add(text)
	}
	for character := 0; character < 256; character++ {
		add(string(rune(character)))
		add("x" + string(rune(character)) + "y")
	}
	pattern := regexp.MustCompile(`[[:alnum:]_@.-]+`)
	for {
		var row struct{ Rule, File, Source string }
		err = decoder.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		add(row.Source)
		for _, text := range pattern.FindAllString(row.Source, -1) {
			add(text)
		}
	}
	ordered := []string{}
	for text := range texts {
		ordered = append(ordered, text)
	}
	sort.Strings(ordered)
	children := []childCase{{false, []int{}, 0}, {true, []int{}, 0}, {true, []int{0}, 1}, {true, []int{0, -1, 1, 0}, 4}, {true, []int{}, 4}, {true, []int{0, 1}, 4}}
	orders := []orderCase{}
	for _, last := range []int{-9007199254740990, -99, -1, 0, 1, 64, 82, 9007199254740990} {
		orders = append(orders, orderCase{false, 0, last})
		for _, group := range []int{-9007199254740991, -17, 0, 1, 64, 9007199254740991} {
			orders = append(orders, orderCase{true, group, last})
		}
	}
	for _, text := range ordered {
		orders = append(orders, orderCase{false, 0, len(text)}, orderCase{true, len(text), -1})
	}
	cases, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(cases).Encode(map[string]any{"texts": ordered, "children": children, "orders": orders}); err != nil {
		panic(err)
	}
	if err = cases.Close(); err != nil {
		panic(err)
	}
	render := func(text string) string {
		out := ""
		for _, unit := range utf16.Encode([]rune(text)) {
			out += fmt.Sprintf("%d,", unit)
		}
		return out
	}
	childPointers := []*collapse.Node{collapse.Comment("a"), collapse.Comment("b")}
	ids := func(nodes []*collapse.Node) string {
		out := []string{}
		for _, node := range nodes {
			id := -1
			for index, pointer := range childPointers {
				if node == pointer {
					id = index
				}
			}
			out = append(out, strconv.Itoa(id))
		}
		return strings.Join(out, ",")
	}
	snapshot := func(node *collapse.Node) string {
		return strings.Join([]string{string(node.Kind), render(node.Selector), render(node.Name), render(node.Params), render(node.Property), render(node.Value), fmt.Sprint(node.ValuePresent), fmt.Sprint(node.Important), fmt.Sprint(node.Context != nil), fmt.Sprint(node.Nodes != nil), fmt.Sprint(len(node.Nodes)), fmt.Sprint(cap(node.Nodes)), ids(node.Nodes)}, "|")
	}
	for _, text := range ordered {
		for _, childrenCase := range children {
			for kind := 0; kind < 2; kind++ {
				var nodes []*collapse.Node
				if childrenCase.Present {
					nodes = make([]*collapse.Node, len(childrenCase.Nodes), childrenCase.Capacity)
				}
				for index, id := range childrenCase.Nodes {
					if id < 0 {
						nodes[index] = nil
					} else {
						nodes[index] = childPointers[id]
					}
				}
				construct := func() *collapse.Node {
					if kind == 0 {
						return collapse.AtRule(text, "params:"+text, nodes...)
					}
					return collapse.StyleRule(text, nodes...)
				}
				first := construct()
				fmt.Println(snapshot(first))
				first.Name = "changed"
				first.Selector = "changed"
				if len(first.Nodes) > 0 {
					first.Nodes[0] = nil
				}
				first.Nodes = first.Nodes[:0:0]
				fmt.Printf("%d:%d:%t:%s\n", len(nodes), cap(nodes), nodes != nil, ids(nodes))
				fmt.Println(snapshot(construct()))
			}
		}
	}
	for _, state := range orders {
		first, present, group, last, second := collapse.AdamicNextOrder(state.GroupPresent, state.GroupOrder, state.LastOrder)
		fmt.Println(first)
		fmt.Printf("%t:%d:%d\n", present, group, last)
		fmt.Println(second)
	}
}
