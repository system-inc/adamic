package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"sort"
	"strings"
	"testing"
)

// A TestProduct_ function declares a product, and a declaration in one package cannot be called from another, so a
// product two packages share is declared once in each (TestProduct_Adamic lives in cmd/adamic-test262, internal/fuzz
// and internal/oracle). The ruling of Oct 14 keeps those copies and asks that they agree: every declaration of one
// TestProduct_ name must have the same body, printed through go/format after dropping a leading Parallel() call, so
// a copy whose builder or inputs drift is refused instead of quietly producing a second product under one name.

type productCensusDeclaration struct {
	File string // path from the repository root
	Name string
	Body string // statements after a leading Parallel() call, printed through go/format
}

func TestEveryDuplicatedProductDeclarationAgrees(t *testing.T) {
	t.Parallel()
	// Reads: stage1/cohere/typeaware/volume_test.go
	// Reads: internal/native/normalize_test.go
	// Reads: internal/native/string_index_test.go
	// Reads: internal/oracle/input_test.go
	// Reads: internal/oracle/oracle_test.go
	// Reads: stage1/cohere/css/print_test.go
	// Reads: stage1/cohere/lint/lint_test.go
	// Reads: *_test.go
	root := parallelCheckRoot(t)
	var declarations []productCensusDeclaration
	var printErr error
	err := parallelCheckEachTest(root, func(rel string, fn *ast.FuncDecl, param string, positions *token.FileSet, lines []string) {
		declaration, ok, err := productCensusDeclare(rel, fn, param)
		if err != nil && printErr == nil {
			printErr = err
		}
		if ok {
			declarations = append(declarations, declaration)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	if printErr != nil {
		t.Fatal(printErr)
	}
	counts := make(map[string]int)
	for _, declaration := range declarations {
		counts[declaration.Name]++
	}
	var duplicated []string
	for name, count := range counts {
		if count > 1 {
			duplicated = append(duplicated, fmt.Sprintf("%s (%d)", name, count))
		}
	}
	sort.Strings(duplicated)
	t.Logf("checked %d product declarations; declared in more than one place: %s", len(declarations), strings.Join(duplicated, ", "))
	for _, failure := range productCensusDisagreements(declarations) {
		t.Error(failure)
	}
}

func TestProductDeclarationCensusRefusesWhatItShould(t *testing.T) {
	t.Parallel()
	const adamic = "{\n\tt.Parallel()\n\tbuildcache.Adamic(t)\n}"
	cases := []struct {
		name    string
		sources map[string]string // file -> body of its TestProduct_Adamic
		refused []string          // files a failure must name
	}{
		{"same body in two packages", map[string]string{"a/a_test.go": adamic, "b/b_test.go": adamic}, nil},
		{"different arguments", map[string]string{"a/a_test.go": adamic, "b/b_test.go": "{\n\tt.Parallel()\n\tbuildcache.Adamic(t, \"-race\")\n}"}, []string{"a/a_test.go", "b/b_test.go"}},
		{"different builder", map[string]string{"a/a_test.go": adamic, "b/b_test.go": "{\n\tt.Parallel()\n\tbuildcache.GoBuild(t, \"adamic\", \"./cmd/adamic\", []string{\"-race\"})\n}"}, []string{"a/a_test.go", "b/b_test.go"}},
		{"extra statement", map[string]string{"a/a_test.go": adamic, "b/b_test.go": adamic, "c/c_test.go": "{\n\tt.Parallel()\n\tbuildcache.Adamic(t)\n\tt.Setenv(\"GOFLAGS\", \"-race\")\n}"}, []string{"c/c_test.go"}},
		{"Parallel in one and not the other", map[string]string{"a/a_test.go": adamic, "b/b_test.go": "{ buildcache.Adamic(t) }"}, nil},
		{"declared in one package only", map[string]string{"a/a_test.go": "{ buildcache.GoBuild(t, \"adamic\", \"./cmd/adamic\", nil) }"}, nil},
	}
	for _, c := range cases {
		var declarations []productCensusDeclaration
		for file, body := range c.sources {
			source := "package p\n\nimport \"testing\"\n\nfunc TestProduct_Adamic(t *testing.T) " + body + "\n"
			parsed, err := parser.ParseFile(token.NewFileSet(), file, source, 0)
			if err != nil {
				t.Fatalf("%s: %v", c.name, err)
			}
			fn := parsed.Decls[1].(*ast.FuncDecl)
			param, ok := parallelCheckTestParameter(fn, map[string]bool{"testing": true})
			if !ok {
				t.Fatalf("%s: %s is not a test", c.name, file)
			}
			declaration, ok, err := productCensusDeclare(file, fn, param)
			if err != nil || !ok {
				t.Fatalf("%s: %s declares no product: %v", c.name, file, err)
			}
			declarations = append(declarations, declaration)
		}
		failures := productCensusDisagreements(declarations)
		if len(c.refused) == 0 {
			if len(failures) != 0 {
				t.Errorf("%s: refused, want passed: %q", c.name, failures)
			}
			continue
		}
		if len(failures) != 1 {
			t.Errorf("%s: %d failures, want 1: %q", c.name, len(failures), failures)
			continue
		}
		for _, file := range c.refused {
			if !strings.Contains(failures[0], file) {
				t.Errorf("%s: failure does not name %s: %s", c.name, file, failures[0])
			}
		}
	}
}

// productCensusDeclare reads fn as a product declaration if its name starts with TestProduct_. The body is printed
// against an empty file set, so the printer sees no line positions and layout differences fall away with comments.
func productCensusDeclare(rel string, fn *ast.FuncDecl, param string) (productCensusDeclaration, bool, error) {
	if !strings.HasPrefix(fn.Name.Name, "TestProduct_") {
		return productCensusDeclaration{}, false, nil
	}
	statements := fn.Body.List
	if parallelCheckFirstStatement(fn, param) {
		statements = statements[1:]
	}
	var body bytes.Buffer
	if err := format.Node(&body, token.NewFileSet(), statements); err != nil {
		return productCensusDeclaration{}, false, fmt.Errorf("%s: print %s: %w", rel, fn.Name.Name, err)
	}
	return productCensusDeclaration{File: rel, Name: fn.Name.Name, Body: body.String()}, true, nil
}

// productCensusDisagreements returns one failure per product name whose declarations differ, sorted by name. When one
// body is the strict majority, the failure names each declaration that departs from it; otherwise it names them all.
func productCensusDisagreements(declarations []productCensusDeclaration) []string {
	byName := make(map[string][]productCensusDeclaration)
	for _, declaration := range declarations {
		byName[declaration.Name] = append(byName[declaration.Name], declaration)
	}
	names := make([]string, 0, len(byName))
	for name := range byName {
		names = append(names, name)
	}
	sort.Strings(names)
	indent := func(body string) string {
		return "\t\t" + strings.ReplaceAll(body, "\n", "\n\t\t")
	}
	var failures []string
	for _, name := range names {
		group := byName[name]
		sort.Slice(group, func(i, j int) bool { return group[i].File < group[j].File })
		counts := make(map[string]int)
		for _, declaration := range group {
			counts[declaration.Body]++
		}
		if len(counts) == 1 {
			continue
		}
		majority, best, tied := "", 0, false
		for body, count := range counts {
			switch {
			case count > best:
				majority, best, tied = body, count, false
			case count == best:
				tied = true
			}
		}
		var report strings.Builder
		if tied {
			fmt.Fprintf(&report, "%s is declared as %d different products; every declaration must have the same body:", name, len(counts))
			for _, declaration := range group {
				fmt.Fprintf(&report, "\n\t%s declares\n%s", declaration.File, indent(declaration.Body))
			}
			failures = append(failures, report.String())
			continue
		}
		var agreeing []string
		for _, declaration := range group {
			if declaration.Body == majority {
				agreeing = append(agreeing, declaration.File)
			}
		}
		fmt.Fprintf(&report, "%s is declared as more than one product; every declaration must have the same body:\n\t%s declare\n%s", name, strings.Join(agreeing, ", "), indent(majority))
		for _, declaration := range group {
			if declaration.Body != majority {
				fmt.Fprintf(&report, "\n\tbut %s declares\n%s", declaration.File, indent(declaration.Body))
			}
		}
		failures = append(failures, report.String())
	}
	return failures
}
