package lower

// Throwaway probe for #xqq507s, never committed: lowers each program in $OR_PROBES the way main
// does, then recomputes the cycle finder's seeds on the lowered result and prints them.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestZzOrProbeSeeds(t *testing.T) {
	directory := os.Getenv("OR_PROBES")
	if directory == "" {
		t.Skip("OR_PROBES unset")
	}
	paths, _ := filepath.Glob(filepath.Join(directory, "*"))
	sort.Strings(paths)
	for _, path := range paths {
		if ext := filepath.Ext(path); ext != ".a" && ext != ".ts" {
			continue
		}
		out := &strings.Builder{}
		fmt.Fprintf(out, "\n######## %s\n", filepath.Base(path))
		probeOne(t, path, out)
		fmt.Print(out.String())
	}
}

func probeOne(t *testing.T, path string, out *strings.Builder) {
	program, err := load.Load([]string{path})
	if err != nil {
		fmt.Fprintf(out, "LOAD ERROR: %v\n", err)
		return
	}
	// As main runs it (cmd/adamic/main.go:215, ProgramRegion off without ADAMIC_PROGRAM_REGION).
	_, lowerErr := Lower(context.Background(), program)
	if lowerErr != nil {
		fmt.Fprintf(out, "LOWER (as main): %T\n  %v\n", lowerErr, lowerErr)
	} else {
		fmt.Fprintf(out, "LOWER (as main): ok\n")
	}
	entry := program.Files()[0]
	checked, release := program.Checker(context.Background(), entry)
	defer release()
	built, err := lowerChecked(program, checked, entry, nil, false, false)
	if err != nil {
		fmt.Fprintf(out, "internal lowering: %v (no seed dump)\n", err)
		return
	}
	l := built
	result := l.result
	fmt.Fprintf(out, "GraphTypes: %d\n", len(result.GraphTypes))
	for local, declared := range result.Locals {
		if !declared.Captured {
			continue
		}
		owner := "top"
		if declared.Function >= 0 {
			owner = result.Functions[declared.Function].Name
		}
		typeName := "<nil>"
		if proven := l.localTypes[local]; proven != nil {
			typeName = l.checker.TypeToString(proven)
		}
		fmt.Fprintf(out, "captured local %d %q in %s: type %s global=%t graphCell=%t closedFrameInput=%t\n", local, declared.Name, owner, typeName, declared.Global, declared.GraphCell, l.closedFrameInput(local))
	}
	for index, function := range result.Functions {
		if function.GraphClosure || len(function.Environment) != 0 || function.RestElement != 0 {
			names := []string{}
			for _, local := range function.Environment {
				names = append(names, result.Locals[local].Name)
			}
			fmt.Fprintf(out, "function %d %q closure=%t rest=%v env=%v graphClosure=%t\n", index, function.Name, function.Closure, function.RestElement, names, function.GraphClosure)
		}
	}
	for _, closure := range l.closureRecords {
		fmt.Fprintf(out, "closureRecord fn %d %q type %s\n", closure.function, result.Functions[closure.function].Name, l.checker.TypeToString(closure.proven))
	}
	// Recompute the seeds as graphTypes (graph_types.go:86-121) chooses them.
	finder := &cycleFinder{l: l, where: map[*checker.Type]*ast.Node{}}
	modules, _ := l.moduleOrder(entry)
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsPartOfTypeNode(node) {
				return false
			}
			finder.libraryIteratorMade(node)
			switch {
			case node.Kind == ast.KindObjectLiteralExpression:
				finder.shape(l.checker.GetTypeAtLocation(node), node)
			case node.Kind == ast.KindArrayLiteralExpression || node.Kind == ast.KindCallExpression || node.Kind == ast.KindNewExpression:
				finder.made(l.checker.GetTypeAtLocation(node), node)
			case ast.IsIdentifier(node) || node.Kind == ast.KindThisKeyword || node.Kind == ast.KindPropertyAccessExpression || node.Kind == ast.KindElementAccessExpression:
				if ast.IsIdentifier(node) && !ast.IsDeclarationName(node) && !ast.IsExpressionNode(node) {
					break
				}
				finder.use(l.checker.GetTypeAtLocation(node), node)
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	for _, proven := range l.instantiated {
		finder.use(proven, l.classNodeFor(proven))
	}
	for _, proven := range finder.seen {
		if err := finder.slotsOf(proven); err != nil {
			fmt.Fprintf(out, "SLOT SEED %s: %v\n", l.checker.TypeToString(proven), err)
		}
	}
	for local, declared := range result.Locals {
		proven := l.localTypes[local]
		if !declared.Captured || declared.Global || proven == nil {
			continue
		}
		reaches := !finder.weak(proven) && finder.reaches(proven, cycleNode{cell: local + 1})
		fmt.Fprintf(out, "CELL %q type %s reachesOwnCell=%t closedFrameInput=%t seed=%t\n", declared.Name, l.checker.TypeToString(proven), reaches, l.closedFrameInput(local), reaches && !l.closedFrameInput(local))
	}
	kinds := map[fresh.WriteKind]string{fresh.WriteField: "field", fresh.WriteElement: "element", fresh.WriteMapEntry: "mapEntry", fresh.WriteSetElement: "setElement", fresh.WriteUnknown: "unknown"}
	for _, write := range fresh.ProveWrites(result) {
		where := "unrecorded"
		holder := "?"
		if write.Site != 0 {
			where = l.program.Where(l.writeSites[write.Site-1].node)
			if h := l.writeSites[write.Site-1].holder; h != nil {
				holder = l.checker.TypeToString(h)
			}
		}
		fmt.Fprintf(out, "WRITE %s in %s at %s holder %s name=%q proven=%t why=%q\n", kinds[write.Kind], l.functionName(write.Function), where, holder, write.Name, write.Proven, write.Why)
	}
	// How the call to or is represented: the IR of every call into a rest function.
	dump := func(label string, body []ir.Statement) {
		walk(body, func(node any) bool {
			switch value := node.(type) {
			case ir.Call:
				if value.Function >= 0 && value.Function < len(result.Functions) && result.Functions[value.Function].RestElement != 0 {
					fmt.Fprintf(out, "CALL in %s to %q: args=%d spread=%v restPacked=%t\n", label, result.Functions[value.Function].Name, len(value.Arguments), value.Spread, value.RestPacked)
					for _, argument := range value.Arguments {
						fmt.Fprintf(out, "    arg %T %+v\n", argument, argument)
					}
				}
			case ir.ArrayLiteral:
				fmt.Fprintf(out, "ARRAY LITERAL in %s: %d elements\n", label, len(value.Elements))
			}
			return true
		})
	}
	dump("main", result.Main)
	for _, function := range result.Functions {
		dump(function.Name, function.Body)
	}
}
