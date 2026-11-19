package lower

import (
	"path/filepath"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

// A generic whose own members instantiate it deeper (Own<T> { next: Own<Own<T>> }, Own<[T, T]>,
// Own<T[]>, Own<{ v: T }>) is a new type at every level, so any walk over it whose only stop is a set
// of the types it has seen never ends. Four walks met it one at a time (collectViewedElements on
// #k031ey0, sameRecordStorage on #fzx0j5z, the cycle finder's use, holdsVoidWalk), and a fifth would
// too, so the program is refused once, where the generic is declared, before lowering walks any type
// (#ykd89v9). This is the expansive-cycle check of ECMA-335 II.9.2, Kennedy and Pierce's finitary rule.
//
// The graph's nodes are the type parameters the program declares. A reference G<..., A_j, ...> gives
// an edge from every type parameter P written inside A_j to G's j-th parameter: instantiating the
// generic that owns P instantiates G with something built from P. An edge is plain when A_j is P
// itself, or an indexed access rooted at P (P[K], P[K][L]): G gets P's own value or a member of it,
// never anything larger. Every other occurrence is expanding: inside another reference's arguments,
// a tuple, an array, an object type, a union, an intersection, a function type, G gets a type that
// holds P and so is larger than it. Instantiation is finite exactly when no expanding edge lies on a
// cycle, so an expanding edge whose two ends are in one strongly connected component (a self-loop
// included) is a generic that grows without end, and the first one in source order is the stop. It's
// refused, not NotYet: generics compile by monomorphization, one copy per instantiation, and docs/0.1.md
// refuses a generic that would instantiate itself without bound, the rule class.go and generic.go already
// keep for classes and functions in the same words, which this keeps for every program generic.
//
// It reads only syntax and asks the checker only for symbols, never for a type, so it instantiates
// nothing itself. That's why it ends: the graph is as large as the program's type parameters, however
// deep the types the program describes would go.
//
// expansiveGenerics runs it, or takes the stop a latent census run already walked to over the same
// checker and program (expansive_generics_cache.go, #bmqm0k1).
func (l *lowering) walkExpansiveGenerics(modules []*ast.SourceFile) error {
	if counter := expansiveGenericsWalks.Load(); counter != nil {
		counter.Add(1)
	}
	graph := &expansiveGraph{l: l, walked: map[*ast.SourceFile]bool{}, successors: map[expansiveParameter][]expansiveParameter{}}
	for _, module := range modules {
		graph.enqueue(module)
	}
	// A generic a module only type-imports is declared in a file moduleOrder doesn't run, since it has
	// no body to run, so each file that declares a reference's target is walked too, after the
	// modules, in the order it was first met.
	for index := 0; index < len(graph.files); index++ {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			switch {
			case node.Kind == ast.KindTypeReference:
				graph.reference(node, node.AsTypeReferenceNode().TypeName)
			case node.Kind == ast.KindExpressionWithTypeArguments && node.Parent != nil && node.Parent.Kind == ast.KindHeritageClause:
				// interface A<T> extends B<A<T>>: a base is instantiated as surely as a member's type is.
				graph.reference(node, node.AsExpressionWithTypeArguments().Expression)
			}
			return node.ForEachChild(visit)
		}
		graph.files[index].AsNode().ForEachChild(visit)
	}
	if !slices.ContainsFunc(graph.edges, func(edge expansiveEdge) bool { return edge.expanding }) {
		return nil
	}
	component := expansiveComponents(graph.order, graph.successors)
	for _, edge := range graph.edges {
		if edge.expanding && component[edge.from] == component[edge.to] {
			return &Refused{
				Where: l.program.Where(edge.reference),
				What:  "a generic instantiated without end (polymorphic recursion): " + sourceExpression(edge.reference) + " in " + l.expansiveOwnerName(edge.reference) + " instantiates it deeper at each level",
				Fix:   "keep recursive type arguments unchanged, or write a type per instantiation",
			}
		}
	}
	return nil
}

// expansiveParameter is one type parameter the program declares: its owner and its position in the
// owner's type parameters. The owner is the symbol of a generic a reference can name (an interface,
// a class, a type alias), so the declarations of one merged interface share their parameters, and
// otherwise the declaring node (a function, a method, a function or constructor type, a mapped type,
// an infer), which nothing names.
type expansiveParameter struct {
	owner any
	index int
}

// expansiveEdge is one occurrence of from in the argument a reference passes to to's position.
type expansiveEdge struct {
	from      expansiveParameter
	to        expansiveParameter
	expanding bool
	reference *ast.Node
}

type expansiveGraph struct {
	l *lowering

	// files is every program file walked, in order: the modules first, as moduleOrder gives them.
	files  []*ast.SourceFile
	walked map[*ast.SourceFile]bool

	// edges are in source order, which makes the stop the same on every run; order is every parameter
	// in the order first met, and successors each one's edges' ends.
	edges      []expansiveEdge
	order      []expansiveParameter
	successors map[expansiveParameter][]expansiveParameter
}

// enqueue adds a file to the walk once. The library, the prelude and Node's declarations aren't the
// program's: their generics are tsc's and Adamic's, and nothing the program writes can make them grow.
func (g *expansiveGraph) enqueue(file *ast.SourceFile) {
	if file == nil || g.walked[file] || load.IsLibrary(file) || load.IsPrelude(file) || load.IsNodeLibrary(file) {
		return
	}
	g.walked[file] = true
	g.files = append(g.files, file)
}

// reference adds the edges a reference to a program-declared generic makes, one for each type
// parameter written in each of its arguments.
func (g *expansiveGraph) reference(reference, name *ast.Node) {
	arguments := reference.TypeArguments()
	if len(arguments) == 0 {
		return
	}
	declaration := g.generic(name)
	if declaration == nil {
		return
	}
	parameters := declaration.TypeParameters()
	for position, argument := range arguments {
		if position >= len(parameters) {
			break
		}
		to := expansiveParameter{owner: g.ownerKey(declaration), index: position}
		g.occurrences(argument, expansivePlain, func(from expansiveParameter, expanding bool) {
			g.edge(expansiveEdge{from: from, to: to, expanding: expanding, reference: reference})
		})
	}
}

func (g *expansiveGraph) edge(edge expansiveEdge) {
	g.meet(edge.from)
	g.meet(edge.to)
	g.edges = append(g.edges, edge)
	g.successors[edge.from] = append(g.successors[edge.from], edge.to)
}

func (g *expansiveGraph) meet(parameter expansiveParameter) {
	if _, met := g.successors[parameter]; !met {
		g.successors[parameter] = nil
		g.order = append(g.order, parameter)
	}
}

// generic is the declaration of the generic a reference's name resolves to, through import aliases,
// when the program declares it, or nil.
func (g *expansiveGraph) generic(name *ast.Node) *ast.Node {
	switch name.Kind {
	case ast.KindQualifiedName:
		name = name.AsQualifiedName().Right
	case ast.KindPropertyAccessExpression:
		name = name.Name()
	}
	if name == nil || !ast.IsIdentifier(name) {
		return nil
	}
	symbol := g.l.symbol(name)
	if symbol == nil || symbol.Flags&(ast.SymbolFlagsInterface|ast.SymbolFlagsClass|ast.SymbolFlagsTypeAlias) == 0 {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		switch declaration.Kind {
		case ast.KindInterfaceDeclaration, ast.KindClassDeclaration, ast.KindClassExpression, ast.KindTypeAliasDeclaration:
			file := ast.GetSourceFileOfNode(declaration)
			if file == nil || load.IsLibrary(file) || load.IsPrelude(file) || load.IsNodeLibrary(file) || len(declaration.TypeParameters()) == 0 {
				return nil
			}
			g.enqueue(file)
			return declaration
		}
	}
	return nil
}

// ownerKey is what expansiveParameter keys a declaring node by: the merged symbol of a generic a
// reference can name, which a reference resolves to from any of its declarations, or the node.
func (g *expansiveGraph) ownerKey(owner *ast.Node) any {
	switch owner.Kind {
	case ast.KindInterfaceDeclaration, ast.KindClassDeclaration, ast.KindClassExpression, ast.KindTypeAliasDeclaration:
		if symbol := owner.Symbol(); symbol != nil {
			return g.l.checker.GetExportSymbolOfSymbol(symbol)
		}
	}
	return owner
}

// parameter is the type parameter a type reference names, when it names one.
func (g *expansiveGraph) parameter(reference *ast.Node) (expansiveParameter, bool) {
	name := reference.AsTypeReferenceNode().TypeName
	if !ast.IsIdentifier(name) {
		return expansiveParameter{}, false
	}
	symbol := g.l.symbol(name)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsTypeParameter == 0 {
		return expansiveParameter{}, false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindTypeParameter || declaration.Parent == nil {
			continue
		}
		owner := declaration.Parent
		switch owner.Kind {
		case ast.KindMappedType, ast.KindInferType:
			// Each declares its one parameter: [K in keyof T], infer U.
			return expansiveParameter{owner: owner, index: 0}, true
		case ast.KindInterfaceDeclaration, ast.KindClassDeclaration, ast.KindClassExpression, ast.KindTypeAliasDeclaration:
		default:
			if owner.FunctionLikeData() == nil {
				// No other node declares a type parameter in a program's TypeScript.
				continue
			}
		}
		if index := slices.Index(owner.TypeParameters(), declaration); index >= 0 {
			return expansiveParameter{owner: g.ownerKey(owner), index: index}, true
		}
	}
	return expansiveParameter{}, false
}

// expansiveExposure is how what a reference's argument holds stands in it.
type expansiveExposure int

const (
	// expansivePlain is the argument itself, through parentheses and the object of an indexed access.
	expansivePlain expansiveExposure = iota
	// expansiveExpanding is inside anything the argument builds around it.
	expansiveExpanding
)

// occurrences calls each for every program type parameter written in node, as exposure says it
// stands, and whether that makes the edge expanding.
func (g *expansiveGraph) occurrences(node *ast.Node, exposure expansiveExposure, each func(parameter expansiveParameter, expanding bool)) {
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindParenthesizedType:
		g.occurrences(node.AsParenthesizedTypeNode().Type, exposure, each)
		return
	case ast.KindIndexedAccessType:
		// T[K] is one of T's members, held inside T, so it's no larger than T: the object keeps its
		// exposure. K only picks which member, and a key is drawn from what T already has, so its
		// occurrences make no edge.
		g.occurrences(node.AsIndexedAccessTypeNode().ObjectType, exposure, each)
		return
	case ast.KindTypeOperator:
		if node.AsTypeOperatorNode().Operator == ast.KindKeyOfKeyword {
			// keyof T is a union of the property names T declares, drawn from the program's finite
			// set of names, never a type that holds T, so it bounds the instantiation and makes no
			// edge.
			return
		}
	case ast.KindConditionalType:
		// A conditional's check type is only tested against its extends type; what the conditional
		// becomes is one of its branches, whose occurrences are counted where they're written, so the
		// check type's make no edge.
		conditional := node.AsConditionalTypeNode()
		g.occurrences(conditional.ExtendsType, expansiveExpanding, each)
		g.occurrences(conditional.TrueType, expansiveExpanding, each)
		g.occurrences(conditional.FalseType, expansiveExpanding, each)
		return
	case ast.KindTypeReference:
		if parameter, ok := g.parameter(node); ok {
			each(parameter, exposure == expansiveExpanding)
			return
		}
	}
	// Everything else builds a type around what it holds: another reference's arguments, a tuple, an
	// array, an object type, a union, an intersection, a function type.
	node.ForEachChild(func(child *ast.Node) bool {
		g.occurrences(child, expansiveExpanding, each)
		return false
	})
}

// expansiveComponents numbers each parameter's strongly connected component by Tarjan's algorithm:
// two parameters share a number exactly when each reaches the other.
func expansiveComponents(order []expansiveParameter, successors map[expansiveParameter][]expansiveParameter) map[expansiveParameter]int {
	index := map[expansiveParameter]int{}
	low := map[expansiveParameter]int{}
	onStack := map[expansiveParameter]bool{}
	component := map[expansiveParameter]int{}
	var stack []expansiveParameter
	visits, components := 0, 0
	var connect func(parameter expansiveParameter)
	connect = func(parameter expansiveParameter) {
		index[parameter], low[parameter] = visits, visits
		visits++
		stack = append(stack, parameter)
		onStack[parameter] = true
		for _, successor := range successors[parameter] {
			if _, visited := index[successor]; !visited {
				connect(successor)
				low[parameter] = min(low[parameter], low[successor])
			} else if onStack[successor] {
				low[parameter] = min(low[parameter], index[successor])
			}
		}
		if low[parameter] != index[parameter] {
			return
		}
		for {
			top := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[top] = false
			component[top] = components
			if top == parameter {
				break
			}
		}
		components++
	}
	for _, parameter := range order {
		if _, visited := index[parameter]; !visited {
			connect(parameter)
		}
	}
	return component
}

// expansiveOwnerName is the name of the nearest named declaration around a reference, as the stop
// names where the generic grows, or the file's name.
func (l *lowering) expansiveOwnerName(reference *ast.Node) string {
	for ancestor := reference.Parent; ancestor != nil; ancestor = ancestor.Parent {
		switch ancestor.Kind {
		case ast.KindInterfaceDeclaration, ast.KindClassDeclaration, ast.KindClassExpression, ast.KindTypeAliasDeclaration, ast.KindFunctionDeclaration, ast.KindMethodDeclaration, ast.KindMethodSignature, ast.KindVariableDeclaration:
			if name := ancestor.Name(); name != nil && ast.IsIdentifier(name) {
				return name.Text()
			}
		}
	}
	return filepath.Base(l.program.FileName(ast.GetSourceFileOfNode(reference)))
}
