package lower

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"errors"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// The census is a type-relation inventory, not lowering or permission to compile unchecked code.
// It uses the production rule with upstream's project options and reports every site, even when
// other Adamic policies or the checker would stop compilation earlier.
func TestOptionalWideningCensus(t *testing.T) {
	configPath := os.Getenv("OPTIONAL_WIDENING_CONFIG")
	if configPath == "" {
		t.Skip("set OPTIONAL_WIDENING_CONFIG and OPTIONAL_WIDENING_OUTPUT to inventory a project")
	}
	configPath, err := filepath.Abs(configPath)
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Dir(configPath)
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(configPath)), nil, nil, osvfs.FS(), nil)
	if config == nil || len(diagnostics) != 0 || len(config.Errors) != 0 {
		t.Fatalf("config diagnostics: %v %v", diagnostics, config)
	}
	fs := cachedvfs.From(bundled.WrapFS(osvfs.FS()))
	host := compiler.NewCachedFSCompilerHost(fs, bundled.LibPath(), nil, nil, nil)
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host})
	if program == nil {
		t.Fatal("no checker program")
	}
	output, err := os.Create(os.Getenv("OPTIONAL_WIDENING_OUTPUT"))
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	encoder := json.NewEncoder(output)
	var formatter *load.Program
	admission := os.Getenv("OPTIONAL_WIDENING_ADMISSION") != ""
	if admission {
		path := filepath.Join(t.TempDir(), "formatter.a")
		if err := os.WriteFile(path, []byte("export {};"), 0600); err != nil {
			t.Fatal(err)
		}
		formatter, err = load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
	}
	disposition := map[string]int{}
	sites, files := 0, 0
	for _, file := range program.GetSourceFiles() {
		if file.IsDeclarationFile || !strings.HasPrefix(file.FileName().AsString(), directory+"/") {
			continue
		}
		files++
		checker, release := program.GetTypeCheckerForFile(context.Background(), file)
		l := &lowering{checker: checker}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if found := l.optionalAtSite(node); found != nil {
				position := scanner.GetTokenPosOfNode(node, file, false)
				line, column := scanner.GetLineAndCharacterOfPosition(file, position)
				end := node.End()
				if end < position || end > len(file.Text()) {
					t.Fatalf("invalid source range")
				}
				row := struct {
					File                                 string
					Line, Column                         int
					Kind, Property, Source, Target, Text string
					Disposition, Diagnostic              string
				}{
					strings.TrimPrefix(file.FileName().AsString(), directory+"/"), line + 1, column + 1, node.Kind.String(), found.property, checker.TypeToString(found.source), checker.TypeToString(found.target), file.Text()[position:end], "", "",
				}
				if admission {
					audit := &lowering{checker: checker, program: formatter, result: &ir.Program{}}
					_, err := audit.viewSchema(node, found.target)
					row.Disposition = "ContractReady"
					if err != nil {
						row.Diagnostic = err.Error()
						var refused *Refused
						if errors.As(err, &refused) {
							row.Disposition = "Refused"
						} else {
							row.Disposition = "NotYet"
						}
					}
					disposition[row.Disposition]++
				}
				if err := encoder.Encode(row); err != nil {
					t.Fatal(err)
				}
				sites++
			}
			node.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		release()
	}
	fmt.Fprintf(output, "{\"summary\":{\"files\":%d,\"sites\":%d}}\n", files, sites)
	t.Logf("%d files, %d optional-widening relation sites; target schema dispositions %v", files, sites, disposition)
}
