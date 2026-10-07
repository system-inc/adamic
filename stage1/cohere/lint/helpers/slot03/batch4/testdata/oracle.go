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
)

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
	pairs := [][3]int{{13, 10, -2}, {13, 10, 0}, {13, 0, 99}, {32, 13, -1}, {0, 10, 0}}
	namespaces := append(collapse.AdamicIgnoredNamespaces(), "", "--unknown", "--FONT", "--font-weight")
	keys := map[[2]string]bool{}
	add := func(key string) {
		for _, namespace := range namespaces {
			keys[[2]string{key, namespace}] = true
		}
	}
	for _, namespace := range namespaces {
		add(namespace)
		for _, key := range collapse.AdamicIgnoredKeys(namespace) {
			for _, suffix := range []string{"", "-", "-bold", "less", "_bold", "--nested", "é", "-😀"} {
				add(key + suffix)
			}
			add(strings.ToUpper(key))
		}
	}
	pattern := regexp.MustCompile(`--[[:alnum:]_-]+`)
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
		for _, key := range pattern.FindAllString(row.Source, -1) {
			add(key)
		}
		for index := -1; index < len(row.Source); index++ {
			peek := func(position int) byte {
				if position < 0 || position >= len(row.Source) {
					return 0
				}
				return row.Source[position]
			}
			pairs = append(pairs, [3]int{int(peek(index + 1)), int(peek(index + 2)), index})
		}
	}
	ordered := [][2]string{}
	for key := range keys {
		ordered = append(ordered, key)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i][0] == ordered[j][0] {
			return ordered[i][1] < ordered[j][1]
		}
		return ordered[i][0] < ordered[j][0]
	})
	cases, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(cases).Encode(map[string]any{"pairs": pairs, "keys": ordered}); err != nil {
		panic(err)
	}
	if err = cases.Close(); err != nil {
		panic(err)
	}
	for character := 0; character < 256; character++ {
		fmt.Println(collapse.AdamicEscapeTerminator(byte(character)))
	}
	observe := func(first, second, index int) {
		calls := []string{}
		result := collapse.AdamicFollowedByWhitespace("unused", index, func(position int) byte {
			calls = append(calls, strconv.Itoa(position))
			if position == index+1 {
				return byte(first)
			}
			return byte(second)
		})
		fmt.Printf("%t:%s\n", result, strings.Join(calls, ","))
	}
	for first := 0; first < 256; first++ {
		for second := 0; second < 256; second++ {
			observe(first, second, 9)
		}
	}
	for _, pair := range pairs {
		observe(pair[0], pair[1], pair[2])
	}
	for _, key := range ordered {
		fmt.Println(collapse.AdamicIgnoredThemeKey(key[0], key[1]))
	}
}
