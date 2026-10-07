package main

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexp"
	"os"
	"sort"
	"strings"
	"unicode/utf16"
)

type Source struct{ Rule, File, Source, Options string }
type Context struct {
	Unicode, InClass bool
	Groups           int
	Named            bool
}
type Corpus struct {
	Runes, Sizes           []int
	Contexts               []Context
	Want                   string
	Sources, OptionStrings int
}

func units(s string) string {
	var b strings.Builder
	for _, v := range utf16.Encode([]rune(s)) {
		fmt.Fprintf(&b, "%d,", v)
	}
	return b.String()
}
func result(r regexp.AdamicSlot02Batch8Result) string {
	return fmt.Sprintf("escape:%d:%d:%d:%d:%t:%s:%s\n", r.Kind, r.Set, r.Rune, r.Width, r.Negated, units(r.Error), units(r.Cause))
}
func main() {
	data, e := os.ReadFile(os.Args[1])
	if e != nil {
		panic(e)
	}
	var inputs []Source
	if e = json.Unmarshal(data, &inputs); e != nil {
		panic(e)
	}
	c := Corpus{Sizes: []int{-3, 0, 1, 2, 4, 19}, Sources: len(inputs)}
	runes := map[int]bool{}
	texts := map[string]bool{}
	for i := 0; i < 256; i++ {
		runes[i] = true
	}
	for _, r := range []int{-2147483648, -1, 0xD7FF, 0xD800, 0xDBFF, 0xDC00, 0xDFFF, 0xE000, 0xFFFF, 0x10000, 0x10400, 0x10428, 0x10FFFF, 0x110000, 2147483647} {
		runes[r] = true
	}
	var options func(any)
	options = func(v any) {
		switch x := v.(type) {
		case string:
			texts[x] = true
			c.OptionStrings++
		case []any:
			for _, e := range x {
				options(e)
			}
		case map[string]any:
			for _, e := range x {
				options(e)
			}
		}
	}
	for _, s := range inputs {
		texts[s.Source] = true
		var value any
		if e = json.Unmarshal([]byte(s.Options), &value); e != nil {
			panic(e)
		}
		options(value)
	}
	for s := range texts {
		for _, r := range s {
			runes[int(r)] = true
		}
	}
	for r := range runes {
		c.Runes = append(c.Runes, r)
	}
	sort.Ints(c.Runes)
	for _, u := range []bool{false, true} {
		for _, cl := range []bool{false, true} {
			for _, extra := range []bool{false, true} {
				groups := 0
				if extra {
					groups = 7
				}
				c.Contexts = append(c.Contexts, Context{u, cl, groups, extra})
			}
		}
	}
	var want strings.Builder
	for _, r := range c.Runes {
		fmt.Fprintf(&want, "class:%s\n", units(regexp.EscapeClassRune(rune(r))))
	}
	for _, ctx := range c.Contexts {
		for _, r := range c.Runes {
			for _, size := range c.Sizes {
				want.WriteString(result(regexp.AdamicSlot02Batch8Identity(r, size, ctx.Unicode, ctx.InClass, ctx.Groups, ctx.Named)))
			}
		}
	}
	for _, r := range c.Runes {
		fmt.Fprintf(&want, "literal:%s\n", units(regexp.AdamicSlot02Batch8Literal(r)))
	}
	c.Want = want.String()
	if e = json.NewEncoder(os.Stdout).Encode(c); e != nil {
		panic(e)
	}
}
