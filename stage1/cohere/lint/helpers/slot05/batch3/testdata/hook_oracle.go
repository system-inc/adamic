package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/system-inc/cohere/internal/lint/ecmascript/react"
	goast "go/ast"
	goparser "go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type sample struct {
	Mode  string `json:"mode"`
	Left  string `json:"left"`
	Right string `json:"right"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	root, output := os.Args[1], os.Args[2]
	symbol := "github.com/system-inc/cohere/internal/lint/ecmascript/react.IsHookName"
	data, err := os.ReadFile(filepath.Join(root, "stage1/cohere/lint/helpers/readiness.json"))
	must(err)
	var readiness struct {
		Remaining []struct {
			Rule    string   `json:"rule"`
			Helpers []string `json:"remaining_helpers"`
		} `json:"remaining"`
	}
	must(json.Unmarshal(data, &readiness))
	data, err = os.ReadFile(filepath.Join(root, "stage1/cohere/lint/inventory/inventory.json"))
	must(err)
	var inventory struct {
		Rules []struct {
			Name  string `json:"name"`
			Tests struct {
				Files []string `json:"files"`
			} `json:"tests"`
		} `json:"rules"`
	}
	must(json.Unmarshal(data, &inventory))
	consumers := map[string]bool{}
	for _, r := range readiness.Remaining {
		for _, h := range r.Helpers {
			if h == symbol {
				consumers[r.Rule] = true
			}
		}
	}
	if len(consumers) == 0 {
		panic("no consumers")
	}
	sources := map[string]bool{}
	counts := map[string]int{}
	for _, r := range inventory.Rules {
		if !consumers[r.Name] {
			continue
		}
		for _, path := range r.Tests.Files {
			file, err := goparser.ParseFile(token.NewFileSet(), filepath.Join(root, path), nil, 0)
			must(err)
			goast.Inspect(file, func(n goast.Node) bool {
				lit, ok := n.(*goast.BasicLit)
				if !ok || lit.Kind != token.STRING {
					return true
				}
				source, err := strconv.Unquote(lit.Value)
				must(err)
				if len(source) > 0 {
					sources[source] = true
					counts[r.Name]++
				}
				return true
			})
		}
		if counts[r.Name] == 0 {
			panic("no source strings for " + r.Name)
		}
	}

	if len(consumers) != 7 {
		panic("consumer count drift")
	}

	texts := map[string]bool{"": true, "use": true, "UseState": true, "used": true, "use2Things": true, "useÉ": true, "useΩ": true, "use😀": true, "use\U00010400": true, "use\x00": true, "useǅ": true}
	for source := range sources {
		texts[source] = true
		f := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: tspath.RootedFilePathFromAbsolute("/probe.tsx"), PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute("/probe.tsx"))}, source, core.ScriptKindTSX)
		var visit func(*ast.Node)
		visit = func(n *ast.Node) {
			if n == nil {
				return
			}
			switch n.Kind {
			case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
				texts[n.Text()] = true
			}
			n.ForEachChild(func(c *ast.Node) bool { visit(c); return false })
		}
		visit(f.AsNode())
	}
	ordered := []string{}
	for s := range texts {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	corpus := []sample{}
	var expected strings.Builder

	for _, name := range ordered {
		corpus = append(corpus, sample{Mode: "hook", Left: name})
		fmt.Fprintln(&expected, react.IsHookName(name))
	}
	corpus = append(corpus, sample{Mode: "hookSweep"})
	scalarCount := 0
	for scalar := 0; scalar <= 0x10ffff; scalar++ {
		if scalar >= 0xd800 && scalar <= 0xdfff {
			continue
		}
		scalarCount++
		fmt.Fprintln(&expected, react.IsHookName("use"+string(rune(scalar))))
	}
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d distinct test-file literals, %d source/parsed/control names, %d Unicode scalars, %d Go verdicts\n", len(consumers), len(sources), len(texts), scalarCount, len(texts)+scalarCount)
}
