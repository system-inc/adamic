package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/system-inc/cohere/internal/lint/ecmascript/regexsyntax"
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
	Bytes       []int `json:"bytes"`
	Unicode     bool  `json:"unicode"`
	UnicodeSets bool  `json:"unicodeSets"`
}

func must(e error) {
	if e != nil {
		panic(e)
	}
}
func main() {
	root, output, mode := os.Args[1], os.Args[2], os.Args[3]
	symbol := "github.com/system-inc/cohere/internal/lint/ecmascript/" + map[string]string{"flags": "regexsyntax.ParseRegexFlags", "uv": "regexsyntax.RegexFlags.UV", "pattern": "regexsyntax.PatternAndFlags"}[mode]
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
	for _, s := range []string{"", "/", "//", "/a", "/a/", "/a\\/b/g", "/😀/uv", "u", "v", "uv", "vu", "gu", "gv", "U", "V", "gimsuy", "gimsvy", "x/uv", "//u/v", string([]byte{255, 'u', 128, 'v'}), string([]byte{'/', 255, '/', 128, 'u'})} {
		sources[s] = true
	}
	for b := 0; b < 256; b++ {
		for _, s := range []string{string([]byte{byte(b)}), "g" + string([]byte{byte(b)}) + "y", "/" + string([]byte{byte(b)}) + "/uv", "/u/" + string([]byte{byte(b)}), "/" + string([]byte{byte(b)})} {
			sources[s] = true
		}
	}
	for _, x := range []byte{'u', 'v', '/'} {
		for _, y := range []byte{'u', 'v', '/', 'a', 0, 255} {
			sources[string([]byte{x, y})] = true
		}
	}
	// Replay direct raw consumer strings as well as the literal suffix each consumer uses.
	originals := []string{}
	for s := range sources {
		originals = append(originals, s)
	}
	for _, s := range originals {
		_, f := regexsyntax.PatternAndFlags(s)
		sources[f] = true
	}
	ordered := []string{}
	for s := range sources {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	rows := []Row{}
	var want strings.Builder
	for _, s := range ordered {
		values := []int{}
		for _, b := range []byte(s) {
			values = append(values, int(b))
		}
		f := regexsyntax.ParseRegexFlags(s)
		rows = append(rows, Row{Bytes: values, Unicode: f.Unicode, UnicodeSets: f.UnicodeSets})
		switch mode {
		case "flags":
			fmt.Fprintf(&want, "%v|%v\n", f.Unicode, f.UnicodeSets)
		case "uv":
			fmt.Fprintln(&want, f.UV())
		case "pattern":
			p, f := regexsyntax.PatternAndFlags(s)
			fmt.Fprintf(&want, "%s|%s\n", hex.EncodeToString([]byte(p)), hex.EncodeToString([]byte(f)))
		default:
			panic("bad mode")
		}
	}
	if mode == "uv" {
		for _, u := range []bool{false, true} {
			for _, v := range []bool{false, true} {
				rows = append(rows, Row{Bytes: []int{}, Unicode: u, UnicodeSets: v})
				fmt.Fprintln(&want, (regexsyntax.RegexFlags{Unicode: u, UnicodeSets: v}).UV())
			}
		}
	}
	data, e = json.Marshal(map[string]any{"mode": mode, "rows": rows})
	must(e)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(want.String()), 0644))
	data, e = json.MarshalIndent(map[string]any{"consumers": counts, "files": files, "captured_strings": captured, "sources": len(ordered), "rows": len(rows), "boundary": "all 256 bytes alone, in flags, between slashes, after slash; malformed UTF-8; direct flags and last-slash suffixes; UV complete truth table"}, "", "  ")
	must(e)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Println(string(data))
}
