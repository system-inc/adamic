package wave12next

import (
	goast "go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Replay the pinned Go rule's reference rows as source; production Go decides
// their default-option results anew rather than copying any fixture expectation.
func referenceControls(t *testing.T, repository string) []string {
	t.Helper()
	var sources []string
	for _, file := range []string{"nexus/correctness_no_process_exit_after_output_test.go", "nexus/correctness_no_uncleared_race_timeout_test.go", "nexus/correctness_require_blocking_standard_streams_test.go"} {
		tree, err := parser.ParseFile(token.NewFileSet(), filepath.Join(repository, "cohere/internal/lint/rules", file), nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		count := 0
		goast.Inspect(tree, func(n goast.Node) bool {
			row, ok := n.(*goast.CompositeLit)
			if !ok {
				return true
			}
			if strings.Contains(file, "nexus/") {
				var lines *goast.CompositeLit
				var imports []string
				shebang := false
				if len(row.Elts) > 1 {
					lines, _ = row.Elts[1].(*goast.CompositeLit)
				}
				for _, element := range row.Elts {
					field, ok := element.(*goast.KeyValueExpr)
					if !ok {
						continue
					}
					key, ok := field.Key.(*goast.Ident)
					if !ok {
						continue
					}
					if key.Name == "lines" {
						lines, _ = field.Value.(*goast.CompositeLit)
					}
					if key.Name == "shebang" {
						value, _ := field.Value.(*goast.Ident)
						shebang = value != nil && value.Name == "true"
					}
					if key.Name == "imports" {
						if list, ok := field.Value.(*goast.CompositeLit); ok {
							for _, item := range list.Elts {
								if name, ok := item.(*goast.Ident); ok {
									if name.Name == "correctnessRequireBlockingStandardStreamsImportBlock" {
										imports = append(imports, "import {blockStandardStreams} from '../../libraries/nexus/source/system/StandardStreams';")
									}
									if name.Name == "correctnessRequireBlockingStandardStreamsImportRun" {
										imports = append(imports, "import {CommandLineInterface,runCommandLineInterface} from '../../libraries/nexus/source/command-line/CommandLineInterface';")
									}
								} else if literal, ok := item.(*goast.BasicLit); ok {
									value, _ := strconv.Unquote(literal.Value)
									imports = append(imports, value)
								}
							}
						}
					}
				}
				if lines == nil {
					return true
				}
				array, ok := lines.Type.(*goast.ArrayType)
				if !ok {
					return true
				}
				element, ok := array.Elt.(*goast.Ident)
				if !ok || element.Name != "string" {
					return true
				}
				var text []string
				for _, expr := range lines.Elts {
					lit, ok := expr.(*goast.BasicLit)
					if !ok || lit.Kind != token.STRING {
						return true
					}
					value, err := strconv.Unquote(lit.Value)
					if err != nil {
						t.Fatal(err)
					}
					text = append(text, value)
				}
				prelude := "declare const rows:string[];declare const lines:string[];declare const output:string;declare const flag:boolean;declare const milliseconds:number;declare function use(x:unknown):void;declare function work():Promise<string>;declare function run():Promise<void>;declare function timeoutAfter(n:number):Promise<never>;\n"
				prefix := ""
				if shebang {
					prefix = "#!/usr/bin/env tsx\n"
				}
				sources = append(sources, prefix+strings.Join(imports, "\n")+"\n"+prelude+strings.Join(text, "\n")+"\nexport {};\n")
				count++
				return false
			}
			at := 1
			if strings.Contains(file, "corpus_data") {
				at = 2
				if len(row.Elts) <= at {
					return true
				}
			}
			lit, ok := row.Elts[at].(*goast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				return true
			}
			text, err := strconv.Unquote(lit.Value)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(text, ";") || strings.Contains(text, "\n") {
				sources = append(sources, text)
				count++
			}
			return true
		})
		if count == 0 {
			t.Fatal("reference extraction empty", file)
		}
		t.Logf("%s: %d reference source rows", file, count)
	}
	return sources
}
