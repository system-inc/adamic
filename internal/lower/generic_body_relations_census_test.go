package lower

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/cachedvfs"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
)

// Inventory uses the production relation rule without running other refusals or
// lowering. This inventory grants no admission, including to checker-rejected files.
func TestGenericBodyRelationsCensus(t *testing.T) {
	configPath := os.Getenv("GENERIC_BODY_CONFIG")
	if configPath == "" {
		t.Skip("set GENERIC_BODY_CONFIG and GENERIC_BODY_OUTPUT for the relation inventory")
	}
	configPath, err := filepath.Abs(configPath)
	if err != nil {
		t.Fatal(err)
	}
	directory := filepath.Dir(configPath)
	config, diagnostics := tsoptions.GetParsedCommandLineOfConfigFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(configPath)), nil, nil, osvfs.FS(), nil)
	if config == nil || len(diagnostics) > 0 || len(config.Errors) > 0 {
		t.Fatalf("config errors: %v %v", diagnostics, config)
	}
	host := compiler.NewCachedFSCompilerHost(cachedvfs.From(bundled.WrapFS(osvfs.FS())), bundled.LibPath(), nil, nil, nil)
	program := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: host})
	type site struct {
		File              string `json:"file"`
		Line              int    `json:"line"`
		Column            int    `json:"column"`
		Relation          string `json:"relation"`
		Source            string `json:"source"`
		Slot              string `json:"slot"`
		Text              string `json:"text"`
		CheckerAssignable bool   `json:"checker_assignable"`
	}
	sites := []site{}
	files := 0
	for _, file := range program.GetSourceFiles() {
		if file.IsDeclarationFile || !strings.HasPrefix(file.FileName().AsString(), directory+"/") {
			continue
		}
		files++
		checked, release := program.GetTypeCheckerForFile(context.Background(), file)
		l := &lowering{checker: checked}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			relation := l.genericRelation(node)
			if relation != nil && l.genericDependent(relation.target, map[*checker.Type]bool{}) {
				source := l.genericOwnType(relation.source)
				if !l.genericBodyAllows(source, relation.target) {
					position := scanner.GetTokenPosOfNode(node, file, false)
					line, column := scanner.GetLineAndCharacterOfPosition(file, position)
					sites = append(sites, site{strings.TrimPrefix(file.FileName().AsString(), directory+"/"), line + 1, column + 1, relation.kind, checked.TypeToString(source), checked.TypeToString(relation.target), file.Text()[position:node.End()], checked.IsTypeAssignableTo(checked.GetTypeAtLocation(relation.source), relation.target)})
				}
			}
			node.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
		release()
	}
	data, err := json.MarshalIndent(struct {
		Files int    `json:"files"`
		Count int    `json:"count"`
		Sites []site `json:"sites"`
	}{files, len(sites), sites}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("GENERIC_BODY_OUTPUT"), append(data, '\n'), 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d files, %d generic body refusals", files, len(sites))
}
