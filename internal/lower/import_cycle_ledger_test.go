package lower

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// TestOriginalCycleLedger measures binding-initialization obligations on the pinned
// original source. It does not claim that unrelated tsc features lower natively.
func TestOriginalCycleLedger(t *testing.T) {
	// Not parallel: the public rule caches one program process-wide.
	root := os.Getenv("ADAMIC_CYCLE_LEDGER_ROOT")
	if root == "" {
		t.Skip("set ADAMIC_CYCLE_LEDGER_ROOT to pristine TypeScript 6.0.3 with generated diagnostics")
	}
	pin, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil || strings.TrimSpace(string(pin)) != "050880ce59e30b356b686bd3144efe24f875ebc8" {
		t.Fatalf("original source pin: %s %v", pin, err)
	}
	entry := tspath.RootedFilePathFromAbsolute(filepath.ToSlash(filepath.Join(root, "src/tsc/tsc.ts")))
	fs := bundled.WrapFS(osvfs.FS())
	options := &core.CompilerOptions{Module: core.ModuleKindESNext, ModuleResolution: core.ModuleResolutionKindBundler, Target: core.ScriptTargetESNext, VerbatimModuleSyntax: core.TSTrue, NoEmit: core.TSTrue}
	config := tsoptions.NewParsedCommandLine(options, []tspath.RootedFilePath{entry}, nil, tspath.RootedDirectoryPathFromAbsolute(filepath.ToSlash(root)), fs.CaseSensitivity())
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: compiler.NewCachedFSCompilerHost(fs, bundled.LibPath(), nil, nil, nil), SingleThreaded: core.TSTrue})
	source := program.GetSourceFile(entry)
	if source == nil {
		t.Fatal("entry not loaded")
	}
	typeChecker, release := program.GetTypeCheckerForFile(context.Background(), source)
	defer release()
	modules, cyclic := esmModuleOrder(typeChecker, source)
	if !cyclic {
		t.Fatal("original program lost its import cycle")
	}
	loadTimeReadLock.Lock()
	findings, err := loadTimeReadRunner.RunRule(program, typeChecker, modules, loadTimeReadRule)
	loadTimeReadLock.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	for index := range findings {
		findings[index].File = strings.TrimPrefix(findings[index].File, filepath.ToSlash(root)+"/")
	}
	if len(findings) == 0 {
		t.Fatal("public rule produced no findings on the pinned firing source")
	}
	analysis := &lowering{checker: typeChecker}
	analysis.proveModuleReads(modules)
	relative := func(file *ast.SourceFile) string {
		path, err := filepath.Rel(root, file.FileName().AsString())
		if err != nil {
			t.Fatal(err)
		}
		return filepath.ToSlash(path)
	}
	location := func(file *ast.SourceFile, node *ast.Node) string {
		line, column := scanner.GetLineAndCharacterOfPosition(file, scanner.GetTokenPosOfNode(node, file, false))
		return fmt.Sprintf("%s:%d:%d", relative(file), line+1, column+1)
	}
	type site struct{ Decision, Reason, Provider, Declaration string }
	sites := map[string]site{}
	order := []string{}
	for _, file := range modules {
		order = append(order, relative(file))
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			if ast.IsIdentifier(node) {
				symbol := typeChecker.GetSymbolAtLocation(node)
				if node.Parent != nil && ast.IsShorthandPropertyAssignment(node.Parent) {
					symbol = typeChecker.GetShorthandAssignmentValueSymbol(node.Parent)
				}
				if symbol != nil {
					symbol = typeChecker.SkipAlias(symbol)
					if symbol != nil && symbol.ValueDeclaration != nil {
						declaration := symbol.ValueDeclaration
						provider := ast.GetSourceFileOfNode(declaration)
						if provider != nil && !provider.IsDeclarationFile {
							decision, reason := "checked", "not proven at this source site; retain runtime TDZ fallback"
							if analysis.provenModuleReads[node] {
								decision, reason = "proven", "provider completes before this module's direct read"
							}
							if declaration.Kind == ast.KindFunctionDeclaration && declaration.Parent != nil && declaration.Parent.Kind == ast.KindSourceFile {
								decision, reason = "proven", "module function declaration is hoisted"
							}
							sites[location(file, node)+"|"+node.Text()] = site{decision, reason, relative(provider), location(provider, declaration)}
						}
					}
				}
			}
			node.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
	}
	bucket := filepath.Join("..", "..", "stage3", "fixtures", "cycles")
	input, err := os.Open(filepath.Join(bucket, "trace.json.gz"))
	if err != nil {
		t.Fatal(err)
	}
	defer input.Close()
	compressed, err := gzip.NewReader(input)
	if err != nil {
		t.Fatal(err)
	}
	defer compressed.Close()
	traceData, err := io.ReadAll(compressed)
	if err != nil {
		t.Fatal(err)
	}
	var trace struct {
		Reads []struct{ Location, Binding, Statement, Provider, Declaration string }
	}
	if err := json.Unmarshal(traceData, &trace); err != nil {
		t.Fatal(err)
	}
	var ledger struct {
		Statements []struct {
			Statement string
			Bindings  []struct {
				Binding, Provider, Declaration string
				ReadLocations                  []string
			}
		}
	}
	data, err := os.ReadFile(filepath.Join(bucket, "ledger.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	counts := map[string]int{"refused": 0, "checked": 0, "proven": 0}
	statementCounts := map[string]int{"refused": 0, "checked": 0, "proven": 0}
	classified := []map[string]any{}
	rowDecisions := map[string]string{}
	classify := func(loc, binding, provider, declaration string) site {
		result, ok := sites[loc+"|"+binding]
		if !ok {
			t.Fatalf("source read not resolved: %s %s", loc, binding)
		}
		if result.Provider != provider || result.Declaration != declaration {
			t.Fatalf("source identity differs: %s %s: %+v; ledger %s %s", loc, binding, result, provider, declaration)
		}
		return result
	}
	for _, read := range trace.Reads {
		if !strings.HasPrefix(read.Statement, "src/compiler/") || strings.Contains(read.Statement, ".generated.ts:") {
			continue
		}
		result := classify(read.Location, read.Binding, read.Provider, read.Declaration)
		counts[result.Decision]++
		if rowDecisions[read.Statement] != "checked" {
			rowDecisions[read.Statement] = result.Decision
		}
		classified = append(classified, map[string]any{"statement": read.Statement, "location": read.Location, "binding": read.Binding, "decision": result.Decision, "reason": result.Reason})
	}
	for _, row := range ledger.Statements {
		if rowDecisions[row.Statement] == "" {
			for _, binding := range row.Bindings {
				for _, loc := range binding.ReadLocations {
					result := classify(loc, binding.Binding, binding.Provider, binding.Declaration)
					if rowDecisions[row.Statement] != "checked" {
						rowDecisions[row.Statement] = result.Decision
					}
				}
			}
		}
		decision := rowDecisions[row.Statement]
		if decision == "" {
			t.Fatalf("unclassified statement %s", row.Statement)
		}
		statementCounts[decision]++
	}
	if len(classified) != 914 || len(rowDecisions) != 58 {
		t.Fatalf("ledger totals: %d reads %d statements", len(classified), len(rowDecisions))
	}
	referenceData, err := os.ReadFile(filepath.Join(bucket, "order.json"))
	if err != nil {
		t.Fatal(err)
	}
	var reference struct{ Order []string }
	if err := json.Unmarshal(referenceData, &reference); err != nil {
		t.Fatal(err)
	}
	if strings.Join(order, "\n") != strings.Join(reference.Order, "\n") {
		t.Fatal("compiler ESM order differs from independently verified Node order")
	}
	output := map[string]any{"sourceCommit": strings.TrimSpace(string(pin)), "entry": "src/tsc/tsc.ts", "moduleOrder": order, "ruleFindings": findings, "readCounts": counts, "statementCounts": statementCounts, "statementDecisions": rowDecisions, "reads": classified, "scope": "Source binding-initialization proof/check obligations, not a full native tsc build. Statement decisions cover bounded ledger reads, not unexecuted branches."}
	encoded, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	destination := os.Getenv("ADAMIC_CYCLE_LEDGER_OUTPUT")
	if destination == "" {
		t.Fatal("set ADAMIC_CYCLE_LEDGER_OUTPUT")
	}
	if err := os.WriteFile(destination, append(encoded, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("statements=%v reads=%v findings=%d", statementCounts, counts, len(findings))
}
