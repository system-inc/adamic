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
	"strings"
	"unicode/utf16"
)

type registration struct {
	Name  string `json:"name"`
	Order int    `json:"order"`
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
	for _, text := range []string{"", "a", "z", "az", "tailwind", "a\n", "a\r\n", " a", "a ", "A", "é", "😀", "@", "@min", "@max", "min", "max", "a@", "@@", "\x00", "sm", "lg", "SM", "Sm", " sm", "sm ", "sm\x00", "-", "--", "---", "----", "--a-", "--a--b", "--é-😀", "--a\x00-b"} {
		add(text)
	}
	for character := 0; character < 256; character++ {
		text := string(rune(character))
		add(text)
		add("a" + text)
		add(text + "z")
		add("a" + text + "z")
		add("--" + text + "-")
		add("--a-" + text + "-b")
	}
	pattern := regexp.MustCompile(`[[:alnum:]_@-]+`)
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
	rows := [][]registration{{}, {{"SM", 1}, {"lg", 2}}, {{"sm", 0}}, {{"sm", -17}, {"sm", 23}}, {{"lg", 999}, {"sm", 42}}, {{"sm", 17}, {"sm", -99}}}
	original := []registration{}
	without := []registration{}
	for _, item := range collapse.FrameworkVariantRegistrations {
		row := registration{item.Name, item.Order}
		original = append(original, row)
		add(item.Name)
		if item.Name != "sm" {
			without = append(without, row)
		}
	}
	ordered := []string{}
	for text := range texts {
		ordered = append(ordered, text)
	}
	sort.Strings(ordered)
	rows = append(rows, original, without)
	segments := [][]string{{}, {""}, {"", ""}, {"a", "", "b", ""}, {"é", "😀"}, {"a-b", "c"}, {"\x00", "\n"}}
	for _, text := range ordered {
		segments = append(segments, []string{text}, []string{"", text, ""})
	}
	for index, text := range ordered {
		rows = append(rows, []registration{{text, index}, {"sm", -11}})
	}
	cases, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(cases).Encode(map[string]any{"texts": ordered, "registrations": rows, "segments": segments}); err != nil {
		panic(err)
	}
	if err = cases.Close(); err != nil {
		panic(err)
	}
	render := func(text string) string {
		output := ""
		for _, unit := range utf16.Encode([]rune(text)) {
			output += fmt.Sprintf("%d,", unit)
		}
		return output
	}
	renderSegments := func(segments []string) string {
		output := []string{}
		for _, segment := range segments {
			output = append(output, render(segment))
		}
		return fmt.Sprintf("%d:%s", len(segments), strings.Join(output, "|"))
	}
	for _, text := range ordered {
		parts := collapse.AdamicSplitThemeKey(text)
		fmt.Println(renderSegments(parts))
		fmt.Println(render(collapse.AdamicJoinSegments(parts)))
		if len(parts) > 0 {
			parts[0] = "changed"
		}
		fmt.Println(renderSegments(collapse.AdamicSplitThemeKey(text)))
	}
	for _, parts := range segments {
		fmt.Println(render(collapse.AdamicJoinSegments(parts)))
	}
	for _, items := range rows {
		list := []collapse.FrameworkVariantRegistration{}
		for _, item := range items {
			list = append(list, collapse.FrameworkVariantRegistration{Name: item.Name, Order: item.Order})
		}
		order, found := collapse.AdamicBreakpointGroupOrder(list)
		fmt.Printf("%d:%t\n", order, found)
	}
}
