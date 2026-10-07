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
)

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	z, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(z)
	texts := map[string]bool{}
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
	z.Close()
	f.Close()
	for _, s := range []string{"", "url(", "url()", "URL(x)", " url(x)", "url(x) ", "url(x)\n", "xurl(x)", "url(x))", "url(a\x00b)", "url(😀)", "url(é)", "😀", "é", "abc", "xabc", "a\nb"} {
		texts[s] = true
	}
	for _, c := range []rune{'\n', '\r', 0x2028, 0x2029, 0x85, '\t', '\v', '\f', ' ', 0xFEFF, 0xA0, 0x1F600} {
		texts["url(a"+string(c)+"b)"] = true
		texts["url(a)"+string(c)] = true
	}
	for _, name := range []string{"xx-small", "x-small", "small", "medium", "large", "x-large", "xx-large", "xxx-large", "larger", "smaller"} {
		for _, s := range []string{name, strings.ToUpper(name), " " + name, name + " ", name + "\n", name + "\x00", "x" + name, name + "x"} {
			texts[s] = true
		}
	}
	ordered := []string{}
	for s := range texts {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	want := []bool{}
	for _, s := range ordered {
		want = append(want, collapse.AdamicURL(s), collapse.AdamicAbsolute(s), collapse.AdamicRelative(s))
	}
	out, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(out).Encode(map[string]any{"texts": ordered}); err != nil {
		panic(err)
	}
	out.Close()
	for _, v := range want {
		fmt.Println(v)
	}
}
