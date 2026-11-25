package lower

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Adamic's scout ruling 1 applies rulings.md:76 at both ends of a never
// parameter. A direct caller traps while evaluating the argument, so nothing
// reaches the Union slot. A read of that slot uses the same never-read trap.
// The dispatch supports are the contravariant-override refusal at
// class_inheritance.go:193 and :194 and the refusal of bivariant method
// parameter views at invariance.go:1743.
func neverParameter(parameter *ast.Node, proven *checker.Type) bool {
	if parameter == nil || parameter.Kind != ast.KindParameter || proven == nil || proven.Flags()&checker.TypeFlagsNever == 0 {
		return false
	}
	declared := parameter.AsParameterDeclaration()
	return ast.IsIdentifier(parameter.Name()) && declared.Initializer == nil && declared.QuestionToken == nil && declared.DotDotDotToken == nil
}

// directOnlyNeverFunction leaves the original signature stop in place if the
// function escapes. Reuse the same whole-program index that tracks never-rest
// markers. Each reference must be a direct call or a proven frame marker.
// The index does not see entry exports (modules.go:257 to :263). Only the
// Worker emits them (internal/worker/worker.go:330), and its bridge calls
// only handle, which must have the Worker handler signature.
func (l *lowering) directOnlyNeverFunction(declaration *ast.Node) bool {
	if declaration.Kind != ast.KindFunctionDeclaration || declaration.Name() == nil || len(declaration.TypeParameters()) != 0 {
		return false
	}
	symbol := l.symbol(declaration.Name())
	if symbol == nil || len(symbol.Declarations) != 1 || l.neverRestImplementation(symbol) != declaration {
		return false
	}
	if declaration.Parent.Kind == ast.KindModuleBlock && ast.HasSyntacticModifier(declaration, ast.ModifierFlagsExport) && l.namespaceHasValue(declaration.Parent.Parent) {
		return false
	}
	for _, reference := range l.neverRestState().references[symbol] {
		use := reference
		if parent := reference.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.Name() == reference {
			if !l.namespaceMember(parent) {
				return false
			}
			use = parent
		}
		for use.Parent != nil && use.Parent.Kind == ast.KindParenthesizedExpression {
			use = use.Parent
		}
		call := use.Parent
		if call == nil || call.Kind != ast.KindCallExpression || call.AsCallExpression().Expression != use || call.Flags&ast.NodeFlagsOptionalChain != 0 {
			if !l.neverFrameMarkerReference(use) {
				return false
			}
		}
	}
	return true
}

// Each parameter records which carriers pass the mark to it. The nil key
// is the library endpoint, and a nil predecessor is the original reference.
type neverFrameRoads map[*ast.Symbol][]*ast.Symbol

func (l *lowering) neverFrameMarkerReference(reference *ast.Node) bool {
	roads := make(neverFrameRoads)
	return l.neverFrameMarker(reference, nil, roads) && roads.reachEnd()
}

// The ruling's "ends at" requires every visited parameter to reach the
// endpoint. Walk back from that call so a closed recursion can borrow a
// path out of its cycle. An unused parameter (callback) and a cycle without
// an endpoint (f_cycle_only) cannot appear in the backwards walk.
func (roads neverFrameRoads) reachEnd() bool {
	if _, found := roads[nil]; !found {
		return false
	}
	reached := map[*ast.Symbol]bool{nil: true}
	pending := []*ast.Symbol{nil}
	for len(pending) > 0 {
		parameter := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		for _, predecessor := range roads[parameter] {
			if !reached[predecessor] {
				reached[predecessor] = true
				pending = append(pending, predecessor)
			}
		}
	}
	return len(reached) == len(roads)
}

// Adamic's ruling 1 road (c), Nov 24 22:21Z, allows only frame markers.
// neverRestCaptureStackTrace uses isLibraryGlobal and the library member's
// checker identity (c01, f_alias, f_namespace_alias, f_destructured, f_call
// and f_apply). TestPrepareStackTraceStaysStopped pins captureStackTrace
// replacement (c02) and prepareStackTrace writes. Strict module frames reveal
// no function through getFunction (c03). TestArgumentsLengthRefusalFixtures'
// indexing row and f_arguments hold access to arguments beyond .length.
func (l *lowering) neverFrameMarker(node *ast.Node, from *ast.Symbol, roads neverFrameRoads) bool {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			return l.neverFrameMarker(parent, from, roads)
		}
	case ast.KindCallExpression:
		call := parent.AsCallExpression()
		if parent.Flags&ast.NodeFlagsOptionalChain != 0 || call.Arguments == nil {
			return false
		}
		position := slices.Index(call.Arguments.Nodes, node)
		if position < 0 {
			return false
		}
		for _, argument := range call.Arguments.Nodes[:position] {
			if argument.Kind == ast.KindSpreadElement {
				return false
			}
		}
		if l.neverRestCaptureStackTrace(parent) {
			if position != 1 {
				return false
			}
			roads[nil] = append(roads[nil], from)
			return true
		}
		return l.neverFrameParameter(call, position, from, roads)
	}
	return false
}

func (l *lowering) neverFrameParameter(call *ast.CallExpression, position int, from *ast.Symbol, roads neverFrameRoads) bool {
	callee := ast.SkipParentheses(call.Expression)
	if !ast.IsIdentifier(callee) && (callee.Kind != ast.KindPropertyAccessExpression || !l.namespaceMember(callee)) {
		return false
	}
	symbol := l.valueSymbol(callee)
	if symbol == nil {
		return false
	}
	var implementation *ast.Node
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() != nil {
			if implementation != nil {
				return false
			}
			implementation = declaration
		}
	}
	if implementation == nil || !neverRestProgramFile(ast.GetSourceFileOfNode(implementation)) || neverRestThisParameter(implementation) {
		return false
	}
	parameters := implementation.Parameters()
	if position >= len(parameters) {
		return false
	}
	parameter := parameters[position].AsParameterDeclaration()
	if !ast.IsIdentifier(parameter.Name()) || parameter.DotDotDotToken != nil || parameter.Initializer != nil {
		return false
	}
	binding := l.valueSymbol(parameter.Name())
	if binding == nil {
		return false
	}
	// A recursive edge assumes the parameter being proved. Any off-list use
	// along the walk rejects the root, so no provisional result is cached.
	predecessors, visited := roads[binding]
	roads[binding] = append(predecessors, from)
	return visited || l.neverFrameParameterUses(implementation, parameter.Name(), binding, roads)
}

func (l *lowering) neverFrameParameterUses(implementation, name *ast.Node, binding *ast.Symbol, roads neverFrameRoads) bool {
	var offRoad func(*ast.Node, bool) bool
	offRoad = func(node *ast.Node, captured bool) bool {
		if ast.IsPartOfTypeNode(node) {
			return false
		}
		if ast.IsIdentifier(node) {
			if l.builtinArguments(node) {
				return l.argumentsRefusal(node) != nil
			}
			return node != name && l.valueSymbol(node) == binding &&
				(captured || !l.neverFrameMarker(node, binding, roads))
		}
		inner := captured || ast.IsFunctionLike(node) || ast.IsClassLike(node)
		return node.ForEachChild(func(child *ast.Node) bool { return offRoad(child, inner) })
	}
	return !implementation.ForEachChild(func(child *ast.Node) bool { return offRoad(child, false) })
}

// Under ruling 1 a write would give the slot a value no caller handed it.
// var redeclaration is refused at locals.go:14, as TestKindUnionWriteVarRedeclarationIsRefused
// pins in cast_kind_union_writes_test.go:762. Arguments that could alias the slot
// stop at arguments_length.go:69 (arguments other than a read of arguments.length).
func (l *lowering) neverParameterWritten(declaration, parameter *ast.Node) bool {
	binding := l.valueSymbol(parameter.Name())
	var writes ast.Visitor
	writes = func(node *ast.Node) bool {
		return ast.IsIdentifier(node) && writtenTarget(node) && l.valueSymbol(node) == binding ||
			node.ForEachChild(writes)
	}
	return declaration.ForEachChild(writes)
}

func (l *lowering) heldNeverParameter(local int) bool {
	name := l.localNodes[local]
	return name != nil && l.result.Locals[local].Type == ir.Union && neverParameter(name.Parent, l.localTypes[local])
}

func (l *lowering) checkNeverArguments(call *ast.CallExpression, function int) error {
	const condition = "a call whose argument for a never parameter isn't a rulings.md:76 never read"
	for position, local := range l.result.Functions[function].Parameters {
		if !l.heldNeverParameter(local) {
			continue
		}
		for _, argument := range call.Arguments.Nodes {
			if argument.Kind == ast.KindSpreadElement {
				return l.notYet(argument, condition)
			}
		}
		if position >= len(call.Arguments.Nodes) {
			return l.notYet(call.AsNode(), condition)
		}
		argument := call.Arguments.Nodes[position]
		making, err := l.neverRead(argument)
		if err != nil || making == nil {
			return l.notYet(argument, condition)
		}
	}
	return nil
}
