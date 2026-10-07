package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	tailwind "github.com/system-inc/cohere/internal/lint/rules/tailwind"
	collapse "github.com/system-inc/cohere/internal/lint/rules/tailwind/collapse"
	"io"
	"os"
	"sort"
	"unicode/utf16"
)

func units(text string) string {
	out := ""
	for _, v := range utf16.Encode([]rune(text)) {
		out += fmt.Sprintf("%d,", v)
	}
	return out
}
func tree(nodes []*collapse.Node) string {
	out := "["
	for _, n := range nodes {
		out += string(n.Kind) + ":" + units(n.Selector) + ":" + units(n.Name) + ":" + units(n.Params) + ":" + units(n.Property) + ":" + units(n.Value) + fmt.Sprintf(":%t:%t:%t:", n.Important, n.ValuePresent, n.Nodes != nil) + tree(n.Nodes) + "|"
	}
	return out + "]"
}

type parseCase struct {
	Input        string                      `json:"input"`
	Normalized   string                      `json:"normalized"`
	Dependencies collapse.AdamicDependencies `json:"dependencies"`
}
type messageCase struct {
	Rule    string `json:"rule"`
	Entry   string `json:"entry"`
	Present bool   `json:"present"`
	Error   string `json:"error"`
}

func main() {
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	z, err := gzip.NewReader(f)
	if err != nil {
		panic(err)
	}
	d := json.NewDecoder(z)
	texts := map[string]bool{}
	names := map[string]bool{"": true, "rule": true, "😀": true, "x\x00y": true, "a\nb": true}
	for {
		var row struct{ Rule, File, Source string }
		err = d.Decode(&row)
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}
		texts[row.Source] = true
		names[row.Rule] = true
		names[row.File] = true
	}
	z.Close()
	f.Close()
	fixture, err := os.ReadFile(os.Args[3])
	if err != nil {
		panic(err)
	}
	var corpus struct{ Cases, ErrorCases []struct{ Input string } }
	if err = json.Unmarshal(fixture, &corpus); err != nil {
		panic(err)
	}
	for _, c := range corpus.Cases {
		texts[c.Input] = true
	}
	for _, c := range corpus.ErrorCases {
		texts[c.Input] = true
	}
	for _, s := range []string{"\ufeff}", "\ufeff.a {x:y}", "é😀}", ".a {--x:\n  a\n b; x:\n a\n b;}", "/*!a*/.a{}/*!b*/", "/*unterminated", "(--x)", "@x", "@layer x{@tailwind utilities}", ".a{x:url(a;b);}", ".a{--x:{a;b};}", ".a{--x:'a;b';}", ".a{--x:/*a:b*/ 1;}", ".a{--x}", ".a{--x:a}", "--x:a", "--x:a;", ".a{color:'a\n}", ".a{color:'a;\r\n}", ".a{color:'a", ".a{x:a\\;b;}", ".a{.b{x:y}.c{z:q}}", ".a{x:;}", ".a{x:y!important trailing;}", ".a{a}", "(", ")", "}", ".a{", "@theme{", "@import a", "\u0085.a\u0085{x:y}", ".a\r{x:y}", ".a\r\n{x:y}"} {
		texts[s] = true
	}
	// Balanced and malformed byte-control combinations exercise both scanners and their offsets.
	for _, v := range []string{"", "x", ":", "a:b", "a;b", "a}b", "(a;b)", "[a;b]", "{a;b}", "'a;b'", "\"a:b\"", "a\\;b", "a/*x*/b", "a\r\nb", "é_😀", "\x00", "!important"} {
		for _, property := range []string{"x", "--x"} {
			for _, end := range []string{";}", "}", ";", ""} {
				texts[".a{"+property+":"+v+end] = true
			}
		}
	}
	ordered := []string{}
	for s := range texts {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	cases := []parseCase{}
	want := []string{}
	for _, s := range ordered {
		collapse.AdamicReset()
		nodes, err := collapse.ParseCSS(s)
		normal := s
		if len(s) >= 3 && s[:3] == "\ufeff" {
			normal = " " + s[3:]
		}
		cases = append(cases, parseCase{collapse.AdamicBytes(s), collapse.AdamicBytes(normal), collapse.AdamicCalls})
		if err != nil {
			e := err.(*collapse.CSSSyntaxError)
			want = append(want, fmt.Sprintf("error:%d:%s", e.Offset, units(e.Message)))
		} else {
			want = append(want, "ok:"+tree(nodes))
		}
	}
	ns := []string{}
	for s := range names {
		ns = append(ns, s)
	}
	sort.Strings(ns)
	messages := []messageCase{}
	for _, name := range ns {
		for _, entry := range []string{"", name, "theme.css", "/a/😀\n.css"} {
			for _, present := range []bool{false, true} {
				for _, text := range []string{"", name, "bad CSS: missing }"} {
					var e error
					if present {
						e = fmt.Errorf("%s", text)
					}
					messages = append(messages, messageCase{name, entry, present, text})
					want = append(want, units(tailwind.DesignSystemDeclineMessage(name, tailwind.DesignSystemResult{EntryPoint: entry, Err: e})))
				}
			}
		}
	}
	for _, handle := range []int{-1, 0, 1} {
		id, alias, unchanged := collapse.AdamicThemeHandle(handle)
		want = append(want, fmt.Sprintf("%d:%t:%t", id, alias, unchanged))
	}
	out, err := os.Create(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err = json.NewEncoder(out).Encode(map[string]any{"css": cases, "messages": messages}); err != nil {
		panic(err)
	}
	out.Close()
	for _, s := range want {
		fmt.Println(s)
	}
}
