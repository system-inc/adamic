// Overlay alongside the independent expression selector in cohere's Go printer.
package javascript

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/format/estree"
	"github.com/system-inc/cohere/internal/format/formatoptions"
)

func statementSyntax(node *ast.Node) bool {
	if node == nil {
		return true
	}
	switch node.Kind {
	case ast.KindSourceFile:
		for _, item := range node.AsSourceFile().Statements.Nodes {
			if !statementSyntax(item) {
				return false
			}
		}
		return true
	case ast.KindBlock:
		valid := true
		node.ForEachChild(func(child *ast.Node) bool {
			if !statementSyntax(child) {
				valid = false
			}
			return false
		})
		return valid
	case ast.KindFunctionDeclaration:
		if node.Name() == nil || node.Type() != nil || node.TypeParameterList() != nil {
			return false
		}
		valid := true
		node.ForEachChild(func(child *ast.Node) bool {
			if child.Kind != ast.KindAsyncKeyword && child.Kind != ast.KindAsteriskToken && !statementSyntax(child) {
				valid = false
			}
			return false
		})
		return valid
	default:
		return supportedSyntax(node)
	}
}
func TestAdamicStatementCorpus(t *testing.T) {
	data, err := os.ReadFile(os.Getenv("ADAMIC_TS_STATEMENT_REQUEST"))
	if err != nil {
		t.Fatal(err)
	}
	var request struct {
		Files           []string
		Directory, Gaps string
	}
	if err = json.Unmarshal(data, &request); err != nil {
		t.Fatal(err)
	}
	cases := []portExpressionCase{}
	coverage := map[string]int{}
	rejected := map[string]int{}
	failed := map[string]string{}
	whole := 0
	add := func(label, source string) { cases = append(cases, portExpressionCase{Label: label, Source: source}) }
	for _, path := range request.Files {
		bytes, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		source := string(bytes)
		file := estree.ParseSourceFile(path, source)
		if diagnostics := file.Diagnostics(); len(diagnostics) > 0 {
			failed[path] = diagnostics[0].String()
			continue
		}
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindSourceFile || ast.IsStatement(node) {
				start := scanner.GetTokenPosOfNode(node, file, false)
				fragment := source[start:node.End()]
				if node.Kind == ast.KindSourceFile {
					start = 0
					fragment = source
				}
				if statementSyntax(node) && coreBoundaries(node, source) && !strings.Contains(fragment, "/*") && !strings.Contains(fragment, "//") && !strings.HasPrefix(fragment, "#!") && !strings.HasPrefix(fragment, "\ufeff#!") && !hasBlankLine(fragment) && !strings.Contains(fragment, "\r") {
					if _, _, _, err := estree.ParseTypeScript("statement.ts", fragment, nil); err == nil {
						add(fmt.Sprintf("%s:%d:%s", path, start, node.Kind), fragment)
						coverage[node.Kind.String()]++
						if node.Kind == ast.KindSourceFile {
							whole++
						}
						return false
					}
					rejected["standalone-context"]++
				} else {
					rejected[node.Kind.String()]++
				}
			}
			node.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
	}
	for _, first := range []string{"const x=1;", "let x,y;", "var x=1,y=2;", "f(x);", "({x:1}).value;", "(a,b);", "'use strict';", "('use strict');", ";", "debugger;", "function named(x){const y=x+1;return y;}", "async function named(){return await foo;}", "function* named(){yield* foo;}"} {
		for _, second := range []string{"f(x);", "let y=x+1;", "const y={a:1,b:2};", "{let y=1;f(y);}", ";", "'later';", "(function named(){return x;})();"} {
			for _, third := range []string{"", "f(y);", "debugger;"} {
				add("program-sequence", first+second+third)
			}
		}
	}
	for _, source := range []string{"", ";", ";;", "function named(){}", "{;}", "const veryLongIdentifierAlpha=veryLongIdentifierBeta+veryLongIdentifierGamma+veryLongIdentifierDelta;f(veryLongIdentifierAlpha);"} {
		add("program-boundary", source)
	}
	options := formatoptions.Default()
	options.PrintWidth = 80
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	var input, answers strings.Builder
	for index := range cases {
		item := &cases[index]
		item.Want, err = Format("statement.ts", item.Source, options, nil)
		if err != nil {
			t.Fatalf("case %d %s %q: %v", index, item.Label, item.Source, err)
		}
		fmt.Fprintln(&input, ">"+escape.Replace(item.Source))
		fmt.Fprintln(&answers, "ok\t"+escape.Replace(item.Want))
	}
	encoded, _ := json.Marshal(cases)
	report, _ := json.MarshalIndent(map[string]any{"files": len(request.Files), "cases": len(cases), "complete_files": whole, "coverage": coverage, "unsupported_candidates": rejected, "file_refusals": failed}, "", "  ")
	for name, data := range map[string][]byte{"cases.txt": []byte(input.String()), "answers.txt": []byte(answers.String()), "cases.json": encoded, "coverage.json": report} {
		if err = os.WriteFile(request.Directory+"/"+name, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%d files, %d parse failures, %d supported statement/program fragments, %d complete files", len(request.Files), len(failed), len(cases), whole)
}

// Pin the observed fork customizations independently of the Adamic printer.
func TestAdamicStatementUpstreamDifferences(t *testing.T) {
	data, err := os.ReadFile(os.Getenv("ADAMIC_TS_STATEMENT_UPSTREAM"))
	if err != nil {
		t.Fatal(err)
	}
	var records []struct{ Source, Go, Prettier string }
	if err = json.Unmarshal(data, &records); err != nil {
		t.Fatal(err)
	}
	options := formatoptions.Default()
	options.PrintWidth = 80
	for _, record := range records {
		got, err := Format("difference.ts", record.Source+";", options, nil)
		if err != nil || got != record.Go {
			t.Fatalf("source %q: Go %q, expected %q, error %v", record.Source, got, record.Go, err)
		}
		if record.Go == record.Prettier {
			t.Fatal("difference fixture must differ")
		}
	}
	t.Logf("%d exact Go fork differences pinned", len(records))
}
