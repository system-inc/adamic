package main

import (
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/imports"
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
	Mode    string `json:"mode"`
	Left    string `json:"left"`
	Missing bool   `json:"missing"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	root, output := os.Args[1], os.Args[2]
	symbol := "github.com/system-inc/cohere/internal/lint/ecmascript/imports.NormalizedFileName"
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

	paths := map[string]bool{"": true, "C:\\src\\pages\\index.tsx": true, "\\server\\share\\file.ts": true, "a\\b\\c": true, "a//b/../c": true, "é/😀\\module.a": true, "a\x00b\\c": true, "a\\": true, "/": true}
	for s := range sources {
		paths[s] = true
	}
	for _, r := range inventory.Rules {
		if consumers[r.Name] {
			for _, path := range r.Tests.Files {
				paths[path] = true
			}
		}
	}
	ordered := []string{}
	for s := range paths {
		ordered = append(ordered, s)
	}
	sort.Strings(ordered)
	corpus := []sample{{Mode: "filename", Missing: true}}
	var expected strings.Builder
	fmt.Fprintln(&expected, imports.NormalizedFileName(nil))
	for _, path := range ordered {
		f := ast.AdamicSourceFileName(path)
		corpus = append(corpus, sample{Mode: "filename", Left: f.FileName().AsString()})
		fmt.Fprintln(&expected, imports.NormalizedFileName(f))
	}
	data, err = json.Marshal(corpus)
	must(err)
	must(os.WriteFile(filepath.Join(output, "cases.json"), data, 0644))
	must(os.WriteFile(filepath.Join(output, "want.txt"), []byte(expected.String()), 0644))
	data, err = json.MarshalIndent(counts, "", "  ")
	must(err)
	must(os.WriteFile(filepath.Join(output, "coverage.json"), data, 0644))
	fmt.Printf("%d consumers, %d distinct test-file literals, %d actual Go filenames/nil, %d Go verdicts\n", len(consumers), len(sources), len(corpus), len(corpus))
}
