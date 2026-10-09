package lower

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A callable type is a promise, not the implementation behind a value. Follow
// closed, immutable bindings and factory returns separately from that type.
// Different instances of the same declaration retain their own closure here.
func (l *lowering) overloadValueTarget(node *ast.Node, active map[*ast.Node]bool) *ast.Node {
	if node == nil {
		return nil
	}
	node = ast.SkipParentheses(node)
	if len(l.checker.GetSignaturesOfType(l.concrete(l.checker.GetTypeAtLocation(node)), checker.SignatureKindCall)) == 0 {
		return nil
	}
	if active[node] {
		return nil
	}
	active[node] = true
	defer delete(active, node)
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol == nil {
			return nil
		}
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil && len(declaration.TypeParameters()) != 0 {
				return nil // Generic function values have a separate specialization owner.
			}
		}
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() == nil {
				return l.censusImplementation(declaration)
			}
		}
		if len(symbol.Declarations) != 1 {
			return nil
		}
		declaration := symbol.Declarations[0]
		switch declaration.Kind {
		case ast.KindVariableDeclaration:
			if !l.overloadValueBinding(declaration) {
				return nil
			}
			return l.overloadValueTarget(declaration.AsVariableDeclaration().Initializer, active)
		case ast.KindParameter:
			owner := declaration.Parent
			if owner == nil || owner.Kind != ast.KindFunctionDeclaration || owner.Name() == nil || owner.Body() == nil {
				return nil
			}
			position := -1
			for i, parameter := range owner.Parameters() {
				if parameter == declaration {
					position = i
				}
			}
			var target *ast.Node
			valid, calls := position >= 0, 0
			l.overloadSymbolUses(l.symbol(owner.Name()), func(use *ast.Node) bool {
				parent := overloadValueParent(use)
				if parent == nil || parent.Kind != ast.KindCallExpression || ast.SkipParentheses(parent.AsCallExpression().Expression) != use {
					valid = false
					return false
				}
				arguments := parent.AsCallExpression().Arguments.Nodes
				if position < 0 || position >= len(arguments) {
					valid = false
					return false
				}
				known := l.overloadValueTarget(arguments[position], active)
				if known == nil || target != nil && known != target {
					valid = false
				} else {
					target = known
				}
				calls++
				return false
			})
			if valid && calls > 0 {
				return target
			}
		}
		return nil
	}
	if node.Kind != ast.KindCallExpression {
		return nil
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if !ast.IsIdentifier(callee) {
		return nil
	}
	symbol := l.symbol(callee)
	if symbol == nil {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration || declaration.Body() == nil || len(declaration.TypeParameters()) != 0 {
			continue
		}
		if declaration.Flags&ast.NodeFlagsHasImplicitReturn != 0 {
			return nil
		}
		var target *ast.Node
		valid, returns := true, 0
		var scan ast.Visitor
		scan = func(item *ast.Node) bool {
			if ast.IsFunctionLike(item) {
				return false
			}
			if item.Kind == ast.KindReturnStatement {
				known := l.overloadValueTarget(item.AsReturnStatement().Expression, active)
				if known == nil || target != nil && known != target {
					valid = false
				} else {
					target = known
				}
				returns++
				return false
			}
			item.ForEachChild(scan)
			return false
		}
		declaration.Body().ForEachChild(scan)
		if valid && returns > 0 {
			return target
		}
	}
	return nil
}

func overloadValueParent(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node.Parent
}

// Ignore declaration names and type syntax; neither invokes the value.
func (l *lowering) overloadSymbolUses(symbol *ast.Symbol, visit ast.Visitor) {
	if symbol == nil {
		return
	}
	var scan ast.Visitor
	scan = func(node *ast.Node) bool {
		if ast.IsIdentifier(node) && l.symbol(node) == symbol && node.Parent != nil && node.Parent.Name() != node && !ast.IsPartOfTypeNode(node) {
			if visit(node) {
				return true
			}
		}
		node.ForEachChild(scan)
		return false
	}
	for _, file := range l.program.Files() {
		file.AsNode().ForEachChild(scan)
	}
}

// The closure may escape only along paths whose eventual calls are visible.
// Passing it to an opaque consumer, writing it into mutable storage, or mixing
// implementations keeps a stop instead of silently losing a result boundary.
func (l *lowering) overloadValueUse(node, implementation *ast.Node, active map[*ast.Node]bool) bool {
	if active[node] {
		return false
	}
	active[node] = true
	defer delete(active, node)
	value := node
	for value.Parent != nil && value.Parent.Kind == ast.KindParenthesizedExpression {
		value = value.Parent
	}
	parent := value.Parent
	if parent == nil {
		return false
	}
	uses := func(symbol *ast.Symbol) bool {
		valid := true
		l.overloadSymbolUses(symbol, func(use *ast.Node) bool {
			if !l.overloadValueUse(use, implementation, active) {
				valid = false
			}
			return false
		})
		return valid
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		declaration := parent.AsVariableDeclaration()
		return declaration.Initializer == value && ast.IsIdentifier(parent.Name()) && l.overloadValueBinding(parent) && uses(l.symbol(parent.Name()))
	case ast.KindReturnStatement:
		owner := parent.Parent
		for owner != nil && !ast.IsFunctionLike(owner) {
			owner = owner.Parent
		}
		if owner == nil || owner.Kind != ast.KindFunctionDeclaration || owner.Name() == nil {
			return false
		}
		valid := true
		l.overloadSymbolUses(l.symbol(owner.Name()), func(use *ast.Node) bool {
			caller := overloadValueParent(use)
			if caller == nil || caller.Kind != ast.KindCallExpression || ast.SkipParentheses(caller.AsCallExpression().Expression) != use || l.overloadValueTarget(caller, map[*ast.Node]bool{}) != implementation || !l.overloadValueUse(caller, implementation, active) {
				valid = false
			}
			return false
		})
		return valid
	case ast.KindCallExpression:
		call := parent.AsCallExpression()
		if ast.SkipParentheses(call.Expression) == ast.SkipParentheses(value) {
			_, _, known := l.overloadValueSignature(call, implementation)
			return known && l.overloadValueTarget(call.Expression, map[*ast.Node]bool{}) == implementation
		}
		callee := ast.SkipParentheses(call.Expression)
		if !ast.IsIdentifier(callee) {
			return false
		}
		symbol := l.symbol(callee)
		if symbol == nil {
			return false
		}
		for _, declaration := range symbol.Declarations {
			if declaration.Kind != ast.KindFunctionDeclaration || declaration.Body() == nil || len(declaration.TypeParameters()) != 0 {
				continue
			}
			for i, argument := range call.Arguments.Nodes {
				if argument == value && i < len(declaration.Parameters()) {
					parameter := declaration.Parameters()[i]
					return ast.IsIdentifier(parameter.Name()) && l.overloadValueTarget(parameter.Name(), map[*ast.Node]bool{}) == implementation && uses(l.symbol(parameter.Name()))
				}
			}
		}
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken || binary.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken {
			return l.overloadValueTarget(binary.Left, map[*ast.Node]bool{}) == implementation && l.overloadValueTarget(binary.Right, map[*ast.Node]bool{}) == implementation
		}
	case ast.KindExpressionStatement:
		return true
	}
	return false
}

// A single-signature variable or parameter can promise one overload just as a
// resolved call through the full set does. Match that promise to the public set;
// do not replace it with the implementation's wider annotation.
func (l *lowering) overloadValueSignature(call *ast.CallExpression, implementation *ast.Node) (*ast.Node, int, bool) {
	resolved := l.checker.GetResolvedSignature(call.AsNode())
	if resolved == nil || implementation == nil || len(implementation.TypeParameters()) != 0 {
		return nil, 0, false
	}
	ordinal := 0
	for _, declaration := range l.symbol(implementation.Name()).Declarations {
		if declaration.Kind != ast.KindFunctionDeclaration || declaration.Body() != nil {
			continue
		}
		ordinal++
		if resolved.Declaration() == declaration {
			return declaration, ordinal, true
		}
		signature := l.checker.GetSignatureFromDeclaration(declaration)
		if len(signature.TypeParameters()) != 0 || len(signature.Parameters()) != len(resolved.Parameters()) {
			continue
		}
		matches := true
		for i, parameter := range signature.Parameters() {
			from, to := l.checker.GetTypeOfSymbol(parameter), l.checker.GetTypeOfSymbol(resolved.Parameters()[i])
			matches = matches && l.censusRelated(from, to) && l.censusRelated(to, from)
		}
		from, to := l.checker.GetReturnTypeOfSignature(signature), l.checker.GetReturnTypeOfSignature(resolved)
		if matches && l.censusRelated(from, to) && l.censusRelated(to, from) {
			return declaration, ordinal, true
		}
	}
	return nil, 0, false
}

// Resolve the implementation's ABI while keeping the value and its environment.
// In particular the call's result ABI must not come from a narrow static type.
func (l *lowering) callOverloadValue(call *ast.CallExpression, closure ir.Expression, arguments []ir.Expression, spread []bool, implementation *ast.Node) (ir.Expression, error) {
	overload, ordinal, known := l.overloadValueSignature(call, implementation)
	if !known {
		return nil, l.notYet(call.AsNode(), "an overloaded value whose static promise does not match a known overload")
	}
	function := -1
	if local, known := l.locals[l.symbol(implementation.Name())]; known {
		function = l.result.Locals[local].NestedFunction - 1
	}
	for _, record := range l.closureRecords {
		declaration := record.node
		if declaration != implementation && declaration != nil && ast.IsIdentifier(declaration) {
			declaration = l.censusImplementationFromName(declaration)
		}
		if declaration == implementation && l.result.Functions[record.function].Closure {
			if local, known := l.locals[l.symbol(implementation.Name())]; known && l.result.Locals[local].NestedFunction > 0 {
				continue // Use this lexical instantiation, not another frame's ABI.
			}
			if function >= 0 && function != record.function {
				return nil, l.notYet(call.AsNode(), "an overloaded value with multiple implementation ABIs")
			}
			function = record.function
		}
	}
	if function < 0 {
		return nil, l.notYet(call.AsNode(), "an overloaded value without a known closure implementation")
	}
	if len(spread) > 0 {
		for _, expanded := range spread {
			if expanded {
				return nil, l.notYet(call.AsNode(), "a spread into an overloaded value")
			}
		}
	}
	if l.result.Functions[function].RestElement != 0 || l.result.Functions[function].ReadsArguments || overloadReadsArguments(implementation) {
		return nil, l.notYet(call.AsNode(), "an overloaded value with rest or observable argument count")
	}
	resolved := l.checker.GetResolvedSignature(call.AsNode())
	produced := l.overloadImplementationType(implementation, resolved, l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation)))
	promised := l.concrete(l.checker.GetReturnTypeOfSignature(resolved))
	invoked := ir.CallClosure{Closure: closure, Direct: function + 1, Arguments: arguments, Returns: l.result.Functions[function].Returns}
	if !l.censusRelated(produced, promised) {
		if strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(call.AsNode())), ".a") && !l.censusProveOverloadResult(implementation, overload) {
			return nil, &Refused{Where: l.program.Where(call.AsNode()), What: fmt.Sprintf("overload %d of %s result at %s requires a per-overload proof", ordinal, implementation.Name().Text(), l.overloadResultPath(produced, promised, "result", map[[2]*checker.Type]bool{})), Fix: "prove the result for every admitted overload argument"}
		}
		return l.overloadSpecialization(call, invoked, implementation, overload, ordinal, produced, promised)
	}
	for i, argument := range arguments {
		if i < len(l.result.Functions[function].Parameters) {
			invoked.Arguments[i] = fit(argument, l.result.Locals[l.result.Functions[function].Parameters[i]].Type)
		}
	}
	of, known := l.representation(promised)
	if !known {
		return nil, l.notYet(call.AsNode(), "an overloaded value result without a representation")
	}
	return fit(invoked, of), nil
}

func (l *lowering) censusImplementationFromName(name *ast.Node) *ast.Node {
	symbol := l.symbol(name)
	if symbol != nil {
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindFunctionDeclaration && declaration.Body() != nil {
				return declaration
			}
		}
	}
	return nil
}

// A wrapper cannot pad a call whose original argument count is observable.
func overloadReadsArguments(implementation *ast.Node) bool {
	found := false
	var scan ast.Visitor
	scan = func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			return false
		}
		if ast.IsIdentifier(node) && node.Text() == "arguments" {
			found = true
		}
		node.ForEachChild(scan)
		return false
	}
	implementation.Body().ForEachChild(scan)
	return found
}

// Adamic values expose only promises with per-overload proofs, even if no call
// currently observes that promise. The TypeScript hatch does not apply to .a.
func (l *lowering) overloadValueProof(node, implementation *ast.Node) error {
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		return nil
	}
	produced := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(implementation))
	ordinal := 0
	for _, overload := range l.symbol(implementation.Name()).Declarations {
		if overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil {
			continue
		}
		ordinal++
		promised := l.checker.GetReturnTypeOfSignature(l.checker.GetSignatureFromDeclaration(overload))
		if !l.censusRelated(produced, promised) && !l.censusProveOverloadResult(implementation, overload) {
			return &Refused{Where: l.program.Where(node), What: fmt.Sprintf("an indirect value of an overload requires a per-overload proof: overload %d of %s result %s cannot be served by implementation result %s at %s", ordinal, implementation.Name().Text(), l.checker.TypeToString(promised), l.checker.TypeToString(produced), l.overloadResultPath(produced, promised, "result", map[[2]*checker.Type]bool{})), Fix: "prove every return for each exposed overload"}
		}
	}
	return nil
}

// A let or legacy var with no writes has the same closed target as a const.
// Inspect captured writes too; its annotation alone establishes no target.
func (l *lowering) overloadValueBinding(declaration *ast.Node) bool {
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration || !ast.IsIdentifier(declaration.Name()) {
		return false
	}
	unchanged := true
	l.overloadSymbolUses(l.symbol(declaration.Name()), func(use *ast.Node) bool {
		if ast.IsAssignmentTarget(use) {
			unchanged = false
		}
		return false
	})
	return unchanged
}
