package lower

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

type librarySpread struct {
	node   *ast.Node
	source ir.Expression
	view   *checker.Type
}

// Library storage is not a list of own enumerable properties. Until library
// layouts describe enumerability, a structural view must not copy that storage.
// The closed world includes library values returned by calls and read from
// fields, not just explicit constructors. Declaration identity also covers new
// library shapes without maintaining a list of built-in class names here.
func (lowering *lowering) checkLibrarySpreads(modules []*ast.SourceFile) error {
	if len(lowering.librarySpreads) == 0 {
		return nil
	}
	candidates := lowering.librarySpreadCandidates(modules)
	graph := newAllocationFlowGraph(lowering.result)
	for _, spread := range lowering.librarySpreads {
		// A complete producer proof preserves ordinary record and Adamic class
		// spreads even when an unrelated compatible library object exists elsewhere.
		if graph.plainSpreadSources(spread.source, map[int]bool{}) {
			continue
		}
		if librarySpreadRefusalError := lowering.librarySpreadRefusal(
			spread,
			candidates,
		); librarySpreadRefusalError != nil {
			return librarySpreadRefusalError
		}
	}
	return nil
}

// Follow joined assignments, arguments and returns. An unmodeled producer is
// not evidence of a plain object: it falls back to the closed-world type set.
func (graph *allocationFlowGraph) plainSpreadSources(expression ir.Expression, active map[int]bool) bool {
	sources := func(node int) bool {
		if active[node] || graph.unknownParameters[node] || len(graph.sources[node]) == 0 {
			return false
		}
		active[node] = true
		defer delete(active, node)
		for _, source := range graph.sources[node] {
			if !graph.plainSpreadSources(source, active) {
				return false
			}
		}
		return true
	}
	switch value := expression.(type) {
	case ir.ObjectLiteral:
		// A spread result has already been checked at its own source.
		return true
	case ir.Undefined, ir.Null:
		return true
	case ir.Read:
		return sources(value.Local + 1)
	case ir.Call:
		targets := graph.program.CallTargets(value)
		if len(targets) == 0 {
			return false
		}
		for _, target := range targets {
			if !sources(graph.resultNode(target)) {
				return false
			}
		}
		return true
	case ir.CallClosure:
		targets := graph.program.ClosureTargets(value)
		if targets.Unknown || len(targets.Functions) == 0 {
			return false
		}
		for _, target := range targets.Functions {
			if !sources(graph.resultNode(target)) {
				return false
			}
		}
		return true
	case ir.Conditional:
		return graph.plainSpreadSources(value.WhenTrue, active) && graph.plainSpreadSources(value.WhenNot, active)
	case ir.Coalesce:
		return graph.plainSpreadSources(value.Value, active) && graph.plainSpreadSources(value.Fallback, active)
	case ir.Box:
		return graph.plainSpreadSources(value.Value, active)
	case ir.Narrow:
		return graph.plainSpreadSources(value.Value, active)
	case ir.Unwrap:
		return graph.plainSpreadSources(value.Value, active)
	case ir.Defined:
		return graph.plainSpreadSources(value.Value, active)
	case ir.CheckedCast:
		return graph.plainSpreadSources(value.Value, active)
	}
	return false
}

func (lowering *lowering) librarySpreadCandidates(modules []*ast.SourceFile) map[*checker.Type]string {
	candidates := map[*checker.Type]string{}
	var note func(*checker.Type)
	note = func(proven *checker.Type) {
		if proven == nil {
			return
		}
		if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
			for _, member := range proven.Types() {
				note(member)
			}
			return
		}
		symbol := proven.Symbol()
		if proven.Flags()&checker.TypeFlagsObject == 0 || symbol == nil {
			return
		}
		// Constructor functions have no instance storage. Their uses have separate guards.
		if len(lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 {
			return
		}
		for _, declaration := range symbol.Declarations {
			source := ast.GetSourceFileOfNode(declaration)
			if load.IsLibrary(source) || load.IsPrelude(source) || (source != nil &&
				strings.Contains(strings.ReplaceAll(source.FileName().AsString(), "\\", "/"), "/@types/node/")) {
				candidates[proven] = symbol.Name
				return
			}
		}
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsExpression(node) {
			note(lowering.checker.GetTypeAtLocation(node))
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return candidates
}

func (lowering *lowering) librarySpreadRefusal(spread librarySpread, candidates map[*checker.Type]string) error {
	names := []string{}
	for candidate, name := range candidates {
		if lowering.checker.IsTypeAssignableTo(candidate, spread.view) {
			names = append(names, name)
		}
	}
	if len(names) != 0 {
		sort.Strings(names)
		return &Refused{
			Where: lowering.program.Where(spread.node),
			What: "an object spread through a view that may hold library class " +
				names[0] +
				" (its internal fields are not own enumerable properties)",
			Fix: "copy the fields you mean by name",
		}
	}
	return nil
}
