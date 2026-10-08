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
	data, err := os.ReadFile(os.Args[2])
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, &sites); err != nil {
		panic(err)
	}
	p, err := bridge.Open(os.Args[1], nil)
	if err != nil {
		panic(err)
	}
	diagnostics := -1
	if len(os.Args) == 5 {
		if os.Args[4] != "--diagnostics" {
			panic("unknown option")
		}
		ctx := context.Background()
		diagnostics = len(p.Compiler.GetSyntacticDiagnostics(ctx, nil)) + len(p.Compiler.GetConfigFileParsingDiagnostics()) + len(p.Compiler.GetProgramDiagnostics()) + len(p.Compiler.GetGlobalDiagnostics(ctx)) + len(p.Compiler.GetBindDiagnostics(ctx, nil)) + len(p.Compiler.GetSemanticDiagnostics(ctx, nil))
	}
	var rows []map[string]any
	for _, r := range sites {
		filename := filepath.Join(filepath.Dir(os.Args[1]), "src/compiler", r.File)
		f := p.Compiler.GetSourceFile(tspath.RootedFilePath(filename))
		if f == nil {
			panic(filename)
		}
		lines := scanner.GetECMALineStarts(f)
		pos := int(lines[r.Line-1]) + r.Column - 1
		var found *ast.Node
		var visit func(*ast.Node) bool
		visit = func(n *ast.Node) bool {
			start := scanner.GetTokenPosOfNode(n, f, false)
			if start == pos && fmt.Sprint(n.Kind) == r.Kind && f.Text()[start:n.End()] == r.Text {
				found = n
			}
			n.ForEachChild(visit)
			return false
		}
		visit(f.AsNode())
		if found == nil {
			panic(fmt.Sprintf("unbound %s:%d:%d", r.File, r.Line, r.Column))
		}
		c, release := p.Compiler.GetTypeCheckerForFile(context.Background(), f)
		t := c.GetTypeAtLocation(found)
		var parts []map[string]any
		if t.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range t.Types() {
				parts = append(parts, map[string]any{"type": c.TypeToString(member), "flags": uint32(member.Flags()), "never": member.Flags()&checker.TypeFlagsNever != 0, "has_id": c.GetPropertyOfType(member, "id") != nil})
			}
		}
		rows = append(rows, map[string]any{"id": fmt.Sprintf("%s:%d:%d", r.File, r.Line, r.Column), "type": c.TypeToString(t), "flags": uint32(t.Flags()), "never": t.Flags()&checker.TypeFlagsNever != 0, "relation": valueTrace(c, found, targetAt(c, found), "root"), "members": parts, "diagnostics": diagnostics, "census_source": r.Source, "census_target": r.Target})
		release()
	}
	out, err := os.Create(os.Args[3])
	if err != nil {
		panic(err)
	}
	defer out.Close()
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(rows); err != nil {
		panic(err)
	}
}
