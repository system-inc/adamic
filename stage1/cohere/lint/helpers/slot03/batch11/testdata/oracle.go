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
	for _, s := range []string{"", "😀", "é", "calc", "calc (", "calс(", "url(calc(x))", "xcalc(", "CALC(", "(calc", "math\x00", "fangsong\n"} {
		texts[s] = true
	}
	for _, name := range []string{"serif", "sans-serif", "monospace", "cursive", "fantasy", "system-ui", "ui-serif", "ui-sans-serif", "ui-monospace", "ui-rounded", "math", "emoji", "fangsong"} {
		for _, s := range []string{name, strings.ToUpper(name), " " + name, name + " ", name + "\n", name + "\x00", "x" + name, name + "x"} {
			texts[s] = true
		}
	}
	for _, name := range []string{"calc", "min", "max", "clamp", "mod", "rem", "sin", "cos", "tan", "asin", "acos", "atan", "atan2", "pow", "sqrt", "hypot", "log", "exp", "round"} {
		for _, s := range []string{name, name + "(", strings.ToUpper(name) + "(", name + " (", name + "\n(", "x" + name + "(", "😀" + name + "(", "'" + name + "('", name + "\x00(", "url(" + name + "(x))", name + "(\n"} {
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
		want = append(want, collapse.AdamicGeneric(s), collapse.AdamicMath(s))
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
	for _, handle := range []int{-1, 0, 1} {
		got, alias, unchanged, repeated, stable := collapse.AdamicUtilityHandle(handle)
		fmt.Printf("%d/%t/%t\n%d/%t\n", got, alias, unchanged, repeated, stable)
	}
}
