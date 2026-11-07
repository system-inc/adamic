// Query the actual pinned Adamic checker without lowering or producing executable code.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
	"os"
	"path/filepath"
)

type site struct {
	File                       string
	Line, Column               int
	Kind, Text, Source, Target string
}

func main() {
	if len(os.Args) != 4 && len(os.Args) != 5 {
		panic("checker CONFIG SITES OUTPUT")
	}
	var sites []site
	data, readError := os.ReadFile(os.Args[2])
	if readError != nil {
		panic(readError)
	}
	if decodeError := json.Unmarshal(data, &sites); decodeError != nil {
		panic(decodeError)
	}
	program, openError := bridge.Open(os.Args[1], nil)
	if openError != nil {
		panic(openError)
	}
	diagnostics := -1
	if len(os.Args) == 5 {
		if os.Args[4] != "--diagnostics" {
			panic("unknown option")
		}
		backgroundContext := context.Background()
		diagnostics = len(program.Compiler.GetSyntacticDiagnostics(backgroundContext, nil)) +
			len(program.Compiler.GetConfigFileParsingDiagnostics()) +
			len(program.Compiler.GetProgramDiagnostics()) +
			len(program.Compiler.GetGlobalDiagnostics(backgroundContext)) +
			len(program.Compiler.GetBindDiagnostics(backgroundContext, nil)) +
			len(program.Compiler.GetSemanticDiagnostics(backgroundContext, nil))
	}
	var rows []map[string]any
	for _, site := range sites {
		filename := filepath.Join(filepath.Dir(os.Args[1]), "src/compiler", site.File)
		file := program.Compiler.GetSourceFile(tspath.RootedFilePath(filename))
		if file == nil {
			panic(filename)
		}
		lines := scanner.GetECMALineStarts(file)
		pos := int(lines[site.Line-1]) + site.Column - 1
		var found *ast.Node
		var visit func(*ast.Node) bool
		visit = func(node *ast.Node) bool {
			start := scanner.GetTokenPosOfNode(node, file, false)
			if start == pos && fmt.Sprint(node.Kind) == site.Kind && file.Text()[start:node.End()] == site.Text {
				found = node
			}
			node.ForEachChild(visit)
			return false
		}
		visit(file.AsNode())
		if found == nil {
			panic(fmt.Sprintf("unbound %s:%d:%d", site.File, site.Line, site.Column))
		}
		c, release := program.Compiler.GetTypeCheckerForFile(context.Background(), file)
		subjectType := c.GetTypeAtLocation(found)
		var parts []map[string]any
		if subjectType.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range subjectType.Types() {
				parts = append(parts, map[string]any{
					"type":   c.TypeToString(member),
					"flags":  uint32(member.Flags()),
					"never":  member.Flags()&checker.TypeFlagsNever != 0,
					"has_id": c.GetPropertyOfType(member, "id") != nil,
				})
			}
		}
		rows = append(rows, map[string]any{
			"id":            fmt.Sprintf("%s:%d:%d", site.File, site.Line, site.Column),
			"type":          c.TypeToString(subjectType),
			"flags":         uint32(subjectType.Flags()),
			"never":         subjectType.Flags()&checker.TypeFlagsNever != 0,
			"relation":      valueTrace(c, found, targetAt(c, found), "root"),
			"members":       parts,
			"diagnostics":   diagnostics,
			"census_source": site.Source,
			"census_target": site.Target,
		})
		release()
	}
	out, createError := os.Create(os.Args[3])
	if createError != nil {
		panic(createError)
	}
	defer out.Close()
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if encodeError := encoder.Encode(rows); encodeError != nil {
		panic(encodeError)
	}
}
