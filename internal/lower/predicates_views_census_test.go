package lower

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// This audits refusal roots on the unchanged census input, not executable IR.
func TestPredicateViewCensus(t *testing.T) {
	root := os.Getenv("PREDICATE_CENSUS_ROOT")
	if root == "" {
		t.Skip("set PREDICATE_CENSUS_ROOT")
	}
	input, err := os.Open("../../stage3/refusal-predicates/coverage.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	rows, err := csv.NewReader(input).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	selected := map[string]int{}
	for i, row := range rows[1:] {
		selected[fmt.Sprintf("%s:%s:%s", row[0], row[1], row[2])] = i + 1
	}
	config, diags := tsoptions.GetParsedCommandLineOfConfigFile(tspath.RootedFilePathFromAbsolute(filepath.Join(root, "src/compiler/tsconfig.json")), nil, nil, osvfs.FS(), nil)
	if config == nil || len(diags) != 0 || len(config.Errors) != 0 {
		t.Fatal("config diagnostics", diags)
	}
	checked := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: compiler.NewCachedFSCompilerHost(cachedvfs.From(bundled.WrapFS(osvfs.FS())), bundled.LibPath(), nil, nil, nil)})
	scratch := filepath.Join(t.TempDir(), "formatter.a")
	if err := os.WriteFile(scratch, []byte("export {};"), 0600); err != nil {
		t.Fatal(err)
	}
	formatter, err := load.Load([]string{scratch})
	if err != nil {
		t.Fatal(err)
	}
	matched, admitted := 0, 0
	for _, file := range checked.GetSourceFiles() {
		if file.IsDeclarationFile || !strings.HasPrefix(string(file.FileName()), root+"/src/compiler/") {
			continue
		}
		ck, release := checked.GetTypeCheckerForFile(context.Background(), file)
		audit := &lowering{checker: ck, program: formatter, result: &ir.Program{}}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			pos := scanner.GetTokenPosOfNode(node, file, false)
			prefix := file.Text()[:pos]
			line := strings.Count(prefix, "\n") + 1
			col := pos - strings.LastIndex(prefix, "\n")
			key := fmt.Sprintf("%s:%d:%d", strings.TrimPrefix(string(file.FileName()), root+"/src/compiler/"), line, col)
			index, found := selected[key]
			if found && ((strings.Contains(rows[index][3], "argument") && node.Kind == ast.KindIdentifier) || (!strings.Contains(rows[index][3], "argument") && node.Kind == ast.KindReturnStatement)) {
				matched++
				var failure error
				if node.Kind == ast.KindIdentifier {
					failure = audit.predicateArguments(node.Parent)
				} else {
					parent := node.Parent
					for parent != nil && !ast.IsFunctionLike(parent) {
						parent = parent.Parent
					}
					if parent == nil || parent.Type() == nil {
						t.Fatalf("no predicate at %s", key)
					}
					failure = audit.predicateRefusal(parent.Type())
				}
				if failure == nil {
					admitted++
					rows[index][4] = "yes"
				} else {
					rows[index][4] = failure.Error()
					if node.Kind == ast.KindReturnStatement {
						parent := node.Parent
						for parent != nil && !ast.IsFunctionLike(parent) {
							parent = parent.Parent
						}
						annotation := parent.Type().AsTypePredicateNode()
						target := audit.concrete(ck.GetTypeFromTypeNode(annotation.Type))
						for _, parameter := range parent.Parameters() {
							if parameter.Name().Text() != annotation.ParameterName.Text() {
								continue
							}
							source := ck.GetNonNullableType(audit.concrete(ck.GetTypeAtLocation(parameter.Name())))
							keeping := audit.widened(target, source, map[[2]*checker.Type]bool{})
							t.Logf("remaining %s: structural source=%t target=%t assignable=%t writable restriction=%t", key, audit.predicateViewObject(source), audit.predicateViewObject(target), ck.IsTypeAssignableTo(target, source), keeping != nil)
							if keeping != nil {
								t.Logf("readonly field=%s source=%s target=%s", keeping.readonlyField, ck.TypeToString(keeping.source), ck.TypeToString(keeping.target))
							}
						}
					}
				}
				delete(selected, key)
			}
			node.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		release()
	}
	if matched != 122 {
		t.Fatalf("matched %d of 122; missing %v", matched, selected)
	}
	out, err := os.Create(os.Getenv("PREDICATE_CENSUS_OUTPUT"))
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	writer := csv.NewWriter(out)
	writer.WriteAll(rows)
	if err := writer.Error(); err != nil {
		t.Fatal(err)
	}
	t.Log("matched=" + strconv.Itoa(matched) + " admitted=" + strconv.Itoa(admitted) + "; refusal roots only, no production compilation claimed")
}
