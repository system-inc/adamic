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
	// written is the view as the source writes it, before the instantiation being lowered replaces its
	// type parameters (l.concrete), so a spread of a T is known as one (spread_prototype_members.go), and
	// mapper is that instantiation's, since a pass after lowering has no instantiation in force.
	written *checker.Type
	mapper  *typeMapper
	// held is the type the spread's copy is held as, as the source writes it, and heldInContext whether
	// that is the literal's contextual type rather than its own (spreadHeld).
	held          *checker.Type
	heldInContext bool
}

// Library storage is not a list of own enumerable properties. Until library
// layouts describe enumerability, a structural view must not copy that storage.
// The closed world includes library values returned by calls and read from
// fields, not just explicit constructors. Declaration identity also covers new
// library shapes without maintaining a list of built-in class names here.
func (l *lowering) checkLibrarySpreads(modules []*ast.SourceFile) error {
	if len(l.librarySpreads) == 0 {
		return nil
	}
	candidates := l.librarySpreadCandidates(modules)
	graph := l.flowGraph(l.result)
	for _, spread := range l.librarySpreads {
		if l.firstLinkSkips(spread.node) {
			continue
		}
		// A complete producer proof preserves ordinary record and Adamic class
		// spreads even when an unrelated compatible library object exists elsewhere.
		if graph.plainSpreadSources(spread.source, map[int]bool{}, everyClassPlain) {
			continue
		}
		if err := l.librarySpreadRefusal(spread, candidates); err != nil {
			return err
		}
	}
	return nil
}

// Follow joined assignments, arguments and returns. An unmodeled producer is
// not evidence of a plain object: it falls back to the closed-world type set.
// plainClass says whether an instance of an Adamic class, by its one-based index,
// counts as plain: every one does for the library rule (everyClassPlain), and for
// the prototype member rule (spread_prototype_members.go) only one whose chain
// holds none of the view's fields on its prototype (prototypeFree).
func (graph *allocationFlowGraph) plainSpreadSources(expression ir.Expression, active map[int]bool, plainClass func(class int) bool) bool {
	sources := func(node int) bool {
		if _, unknownSource := graph.unknownAt(node); active[node] || unknownSource || len(graph.sources[node]) == 0 {
			return false
		}
		active[node] = true
		defer delete(active, node)
		for _, source := range graph.sources[node] {
			if !graph.plainSpreadSources(source, active, plainClass) {
				return false
			}
		}
		return true
	}
	switch value := expression.(type) {
	case ir.ObjectLiteral:
		// A spread result has already been checked at its own source. An accessor
		// literal's synthetic class holds its accessors as its own, which a spread copies.
		return value.Class == 0 || graph.program.Classes[value.Class-1].Literal || plainClass(value.Class)
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
		return graph.plainSpreadSources(value.WhenTrue, active, plainClass) && graph.plainSpreadSources(value.WhenNot, active, plainClass)
	case ir.Coalesce:
		return graph.plainSpreadSources(value.Value, active, plainClass) && graph.plainSpreadSources(value.Fallback, active, plainClass)
	case ir.Box:
		return graph.plainSpreadSources(value.Value, active, plainClass)
	case ir.Narrow:
		return graph.plainSpreadSources(value.Value, active, plainClass)
	case ir.Unwrap:
		return graph.plainSpreadSources(value.Value, active, plainClass)
	case ir.Defined:
		return graph.plainSpreadSources(value.Value, active, plainClass)
	case ir.CheckedCast:
		return graph.plainSpreadSources(value.Value, active, plainClass)
	}
	return false
}

// everyClassPlain counts every Adamic class instance as a plain spread source, as the library rule
// does: its concern is library storage, which no Adamic class holds.
func everyClassPlain(int) bool {
	return true
}

// prototypeFree is the prototype member rule's plainClass: a class is plain when neither it nor any
// base holds a getter, a setter or a method under a name in fields, so a spread of its instance
// copies every field the view declares. A static member is on the constructor, not the prototype,
// and the class lowering never lists one in Methods.
func (graph *allocationFlowGraph) prototypeFree(fields map[string]bool) func(class int) bool {
	return func(class int) bool {
		for holder := class; holder != 0; holder = graph.program.Classes[holder-1].Base {
			for _, accessor := range graph.program.Classes[holder-1].Accessors {
				if fields[accessor.Name] {
					return false
				}
			}
			for _, method := range graph.program.Classes[holder-1].Methods {
				if method >= 0 && fields[graph.program.Functions[method].MethodName] {
					return false
				}
			}
		}
		return true
	}
}

func (l *lowering) librarySpreadCandidates(modules []*ast.SourceFile) map[*checker.Type]string {
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
		if len(l.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0 {
			return
		}
		if libraryDeclared(symbol) {
			candidates[proven] = symbol.Name
		}
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsExpression(node) {
			note(l.checker.GetTypeAtLocation(node))
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return candidates
}

// libraryDeclared reports whether a symbol is declared by the library, the prelude or Node's types,
// whose values' storage isn't a list of own enumerable properties.
func libraryDeclared(symbol *ast.Symbol) bool {
	for _, declaration := range symbol.Declarations {
		source := ast.GetSourceFileOfNode(declaration)
		if load.IsLibrary(source) || load.IsPrelude(source) || (source != nil && strings.Contains(strings.ReplaceAll(source.FileName().AsString(), "\\", "/"), "/@types/node/")) {
			return true
		}
	}
	return false
}

func (l *lowering) librarySpreadRefusal(spread librarySpread, candidates map[*checker.Type]string) error {
	names := []string{}
	for candidate, name := range candidates {
		if l.checker.IsTypeAssignableTo(candidate, spread.view) {
			names = append(names, name)
		}
	}
	if len(names) != 0 {
		sort.Strings(names)
		return &Refused{Where: l.program.Where(spread.node), What: "an object spread through a view that may hold library class " + names[0] + " (its internal fields are not own enumerable properties)", Fix: "copy the fields you mean by name"}
	}
	return nil
}
