package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The library's own-property function has no receiver state. Only its explicit
// record .call idiom uses a readiness marker; other proven receivers and dense
// .apply calls retain the area's established library adapters.
func (lowering *lowering) detachedOwnMethod(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression ||
		node.Name().Text() != "hasOwnProperty" ||
		!lowering.libraryMember(node) {
		return false
	}
	receiver := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	return receiver.Kind == ast.KindPropertyAccessExpression && receiver.Name().Text() == "prototype" &&
		lowering.isLibraryGlobal(receiver.AsPropertyAccessExpression().Expression, "Object")
}

func (lowering *lowering) detachedOwnDeclaration(node *ast.Node) bool {
	if node == nil ||
		node.Kind != ast.KindVariableDeclaration ||
		!ast.IsIdentifier(node.Name()) ||
		node.Parent == nil ||
		node.Parent.Flags&ast.NodeFlagsConst == 0 ||
		node.AsVariableDeclaration().Initializer == nil {
		return false
	}
	method, known := lowering.libraryMethod(node.AsVariableDeclaration().Initializer, map[*ast.Symbol]bool{})
	return known && method.family == "Object.prototype" && method.name == "hasOwnProperty"
}

func (lowering *lowering) detachedOwnAlias(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) {
		return nil
	}
	symbol := lowering.symbol(node)
	if symbol != nil && len(symbol.Declarations) == 1 && lowering.detachedOwnDeclaration(symbol.Declarations[0]) {
		return symbol.Declarations[0]
	}
	return nil
}

func (lowering *lowering) detachedOwnCallReceiver(node *ast.Node) *ast.Node {
	if node.Kind != ast.KindCallExpression {
		return nil
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if call.QuestionDotToken != nil ||
		callee.Kind != ast.KindPropertyAccessExpression ||
		callee.Name().Text() != "call" ||
		callee.AsPropertyAccessExpression().QuestionDotToken != nil ||
		!lowering.libraryMember(callee) {
		return nil
	}
	receiver := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	if lowering.detachedOwnMethod(receiver) {
		method := receiver.AsPropertyAccessExpression()
		prototype := ast.SkipParentheses(method.Expression).AsPropertyAccessExpression()
		if method.QuestionDotToken != nil || prototype.QuestionDotToken != nil {
			return nil
		}
	}
	if lowering.detachedOwnMethod(receiver) || lowering.detachedOwnAlias(receiver) != nil {
		return receiver
	}
	return nil
}

func (lowering *lowering) detachedOwnRefusal(node *ast.Node) error {
	refused := func() error {
		return &Refused{Where: lowering.program.Where(node), What: "detached Object.prototype.hasOwnProperty",
			Fix: "bind it to a const and use only .call(target, key), or use that call directly"}
	}
	if lowering.isExpression(node) && viewSite(node) {
		contextual := lowering.slotType(node)
		at := node
		for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
			at = at.Parent
		}
		if parent := at.Parent; parent != nil && parent.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
			// These library operations already dispatch by the actual target storage.
			if lowering.detachedOwnCallReceiver(parent) != nil || (callee.Kind == ast.KindPropertyAccessExpression &&
				lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object")) {
				contextual = nil
			}
		}
		if contextual != nil &&
			contextual.Flags()&checker.TypeFlagsNonPrimitive != 0 &&
			lowering.detachedOwnOpaqueArgument(at) {
			of, known := lowering.representation(lowering.checker.GetTypeAtLocation(node))
			if lowering.detachedOwnObjectParameter(ast.SkipParentheses(node)) {
				of, known = ir.Object, true
			}
			if !known || of != ir.Object {
				return lowering.notYet(node, "a non-object representation passed through an opaque object parameter")
			}
		}
	}
	if node.Kind == ast.KindShorthandPropertyAssignment {
		symbol := lowering.checker.GetShorthandAssignmentValueSymbol(node)
		if symbol != nil && len(symbol.Declarations) == 1 && lowering.detachedOwnDeclaration(symbol.Declarations[0]) {
			return refused()
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		receiver := node.AsPropertyAccessExpression().Expression
		if lowering.detachedOwnAlias(receiver) != nil && !called(node) {
			return refused()
		}
	}
	if lowering.detachedOwnDeclaration(node) && ast.GetCombinedModifierFlags(node)&ast.ModifierFlagsExport != 0 {
		return refused()
	}
	if !lowering.detachedOwnMethod(node) && lowering.detachedOwnAlias(node) == nil {
		return nil
	}
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if at.Parent != nil &&
		lowering.detachedOwnDeclaration(at.Parent) &&
		at.Parent.AsVariableDeclaration().Initializer == at {
		return nil
	}
	if declaration := lowering.detachedOwnAlias(node); declaration != nil && node == declaration.Name() {
		return nil
	}
	if at.Parent != nil && at.Parent.Kind == ast.KindPropertyAccessExpression {
		member := at.Parent
		if member.AsPropertyAccessExpression().Expression == at &&
			member.Name().Text() == "apply" &&
			lowering.libraryMethodReadAllowed(node) {
			return nil
		}
		if member.AsPropertyAccessExpression().Expression == at && member.Name().Text() == "call" {
			call := member
			for call.Parent != nil && call.Parent.Kind == ast.KindParenthesizedExpression {
				call = call.Parent
			}
			if call.Parent != nil && lowering.detachedOwnCallReceiver(call.Parent) != nil {
				return nil
			}
		}
	}
	return refused()
}

func init() { refusalRules.register(900, "detachedOwnRefusal", (*lowering).detachedOwnRefusal) }

func (lowering *lowering) detachedOwnCall(node *ast.Node) (ir.Expression, bool, error) {
	receiver := lowering.detachedOwnCallReceiver(node)
	if receiver == nil {
		return nil, false, nil
	}
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) != 2 || args[0].Kind == ast.KindSpreadElement || args[1].Kind == ast.KindSpreadElement {
		return nil, true, &Refused{
			Where: lowering.program.Where(node),
			What:  "detached hasOwnProperty call arguments",
			Fix:   "use exactly .call(target, key)",
		}
	}
	arguments := []ir.Expression{}
	if lowering.detachedOwnAlias(receiver) != nil {
		// Keep the binding's initialization and capture checks before either argument.
		local, ok := lowering.local(receiver)
		if !ok {
			return nil, true, lowering.notYet(receiver, "a detached own-property binding before its declaration")
		}
		arguments = append(arguments, ir.Read{Local: local, Of: ir.Boolean, Checked: lowering.checked(local)})
	}
	proven := lowering.checker.GetTypeAtLocation(args[0])
	of, known := lowering.representation(proven)
	if lowering.detachedOwnObjectParameter(ast.SkipParentheses(args[0])) {
		of, known = ir.Object, true
	}
	if known && of != ir.Record && of != ir.Object {
		return nil, false, nil // Existing primitive, array and function adapters prove these descriptors.
	}
	if !known || lowering.includesUndefined(proven) || proven.Flags()&checker.TypeFlagsNull != 0 ||
		(of != ir.Record && of != ir.Object) || checker.IsTupleType(proven) {
		return nil, true, lowering.notYet(args[0], "detached hasOwnProperty on other than a present record or object")
	}
	if of == ir.Object {
		if reason := lowering.prototypeHazard(args[0], ""); reason != "" {
			return nil, true, lowering.notYet(args[0], "detached hasOwnProperty through an object view ("+reason+")")
		}
	}
	target, expressionError := lowering.expression(args[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	key, recordKeyError := lowering.recordKey(args[1], false)
	if recordKeyError != nil {
		return nil, true, recordKeyError
	}
	arguments = append(arguments, target, key)
	builder := lowering.libraryArrayBuilder(arguments)
	offset := len(arguments) - 2
	var result ir.Expression = ir.ObjectCall{
		Method: "hasOwn",
		Arguments: []ir.Expression{
			builder.read(builder.parameters[offset]),
			builder.read(builder.parameters[offset+1]),
		},
		Returns: ir.Boolean,
	}
	if of == ir.Record {
		result = ir.RecordCall{
			Method: "hasOwn",
			Arguments: []ir.Expression{
				builder.read(builder.parameters[offset]),
				builder.read(builder.parameters[offset+1]),
			},
			Returns: ir.Boolean,
		}
	}
	return builder.finish("detached_has_own", result), true, nil
}

// The parser helper's object parameter is opaque and used solely for ownership
// tests. It is an object pointer, never an erased array, record or callable:
// the up-front argument checks must still agree with that representation.
func (lowering *lowering) detachedOwnObjectParameter(node *ast.Node) bool {
	if !ast.IsIdentifier(node) || lowering.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsNonPrimitive == 0 {
		return false
	}
	symbol := lowering.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindParameter {
		return false
	}
	parameter := symbol.Declarations[0]
	function := parameter.Parent
	if function == nil || function.Body() == nil {
		return false
	}
	safe, used := true, false
	var visit ast.Visitor
	visit = func(use *ast.Node) bool {
		if use.Kind == ast.KindShorthandPropertyAssignment &&
			lowering.checker.GetShorthandAssignmentValueSymbol(use) == symbol {
			safe = false
		}
		if ast.IsIdentifier(use) && lowering.symbol(use) == symbol {
			used = true
			at := use
			for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
				at = at.Parent
			}
			parent := at.Parent
			if parent == nil ||
				parent.Kind != ast.KindCallExpression ||
				lowering.detachedOwnCallReceiver(parent) == nil ||
				len(parent.AsCallExpression().Arguments.Nodes) != 2 ||
				parent.AsCallExpression().Arguments.Nodes[0] != at {
				safe = false
			}
		}
		return use.ForEachChild(visit)
	}
	function.Body().ForEachChild(visit)
	return safe && used
}

// Library's ordinary object parameters use tagged Union storage. Restrict this
// representation check to the own-property-only helper whose parameter is a
// native Object, so unrelated library adapters keep their existing proofs.
func (lowering *lowering) detachedOwnOpaqueArgument(at *ast.Node) bool {
	call := at.Parent
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	signature := lowering.checker.GetResolvedSignature(call)
	if signature == nil {
		return false
	}
	for index, argument := range call.AsCallExpression().Arguments.Nodes {
		if argument != at || index >= len(signature.Parameters()) {
			continue
		}
		parameter := signature.Parameters()[index]
		if len(parameter.Declarations) == 1 && parameter.Declarations[0].Kind == ast.KindParameter {
			return lowering.detachedOwnObjectParameter(parameter.Declarations[0].Name())
		}
	}
	return false
}
