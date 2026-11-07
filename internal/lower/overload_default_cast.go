package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// tsc's sortAndDeduplicate falls back to a default comparer with a pair of casts (core.ts:810, after
// adaptation's `as any` became `as unknown`):
//
//	equalityComparer ?? comparer ?? compareStringsCaseSensitive as unknown as Comparer<T>
//
// The cast is true, by the overloads: only `sortAndDeduplicate(array: readonly string[])` lets a caller
// leave the comparers out, and there T is a string, so the default is the comparer T needs. Adamic ruled
// it per-overload proof, no check (#b5w3ycg, Nov 5; #25t5461). overloadDefaultPair is that proof. The
// pair `x as unknown as Target` is erased, lowered as x, when all of this holds:
//   - it's in a .ts file, inside an overload set's implementation and not in a function nested in it;
//   - it's reached only as the last operand of a `??` chain whose other operands are each one of the
//     implementation's parameters by name, optional, with no default, not a rest, and never assigned,
//     so reaching it means every one of them is absent;
//   - every overload that lets all of those be absent (a missing, optional or nullable parameter there)
//     fixes the implementation's type parameters, by alpha-renaming or by inference from its parameters,
//     so that x's type goes where Target goes by Adamic's cast relation, castAssignable. An overload that
//     requires one of them never reaches the pair.
//
// The inner `as unknown` is erased only as part of such a pair (rulings.md, brand mints); a lone hop,
// or a pair anywhere else, keeps its refusal.
func (lowering *lowering) overloadDefaultPair(node *ast.Node) bool {
	if !strings.HasSuffix(lowering.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {
		return false
	}
	as := node.AsAsExpression()
	inner := ast.SkipParentheses(as.Expression)
	if inner.Kind != ast.KindAsExpression ||
		lowering.checker.GetTypeAtLocation(inner).Flags()&checker.TypeFlagsUnknown == 0 {
		// A lone hop, or any other cast, isn't a pair.
		return false
	}
	operand := inner.AsAsExpression().Expression
	implementation := ast.GetContainingFunction(node)
	if implementation == nil ||
		implementation.Kind != ast.KindFunctionDeclaration ||
		implementation.Body() == nil ||
		implementation.Name() == nil {
		return false
	}
	absent, ok := lowering.overloadDefaultAbsent(node, implementation)
	if !ok || len(absent) == 0 {
		return false
	}
	symbol := lowering.symbol(implementation.Name())
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration && overloadThisParameter(declaration) {
			// A this parameter shifts every position the proof compares (Compiler's should).
			return false
		}
	}
	overloads := 0
	for _, overload := range symbol.Declarations {
		if overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil {
			continue
		}
		overloads++
		if !lowering.overloadAdmitsAbsence(overload, absent) {
			continue
		}
		mapper, mapped := lowering.overloadDefaultMapper(implementation, overload)
		if !mapped {
			return false
		}
		from, target := lowering.checker.GetTypeAtLocation(operand), lowering.checker.GetTypeAtLocation(node)
		if mapper != nil {
			from, target = instantiateType(
				lowering.checker,
				from,
				mapper,
			), instantiateType(lowering.checker, target, mapper)
		}
		if !lowering.castAssignable(from, target) {
			return false
		}
	}
	return overloads > 0 && lowering.onlyCalled(symbol)
}

// overloadThisParameter is whether a declaration's first parameter is a this parameter.
func overloadThisParameter(declaration *ast.Node) bool {
	parameters := declaration.Parameters()
	return len(parameters) > 0 && ast.IsIdentifier(parameters[0].Name()) && parameters[0].Name().Text() == "this"
}

// onlyCalled is whether every reference to symbol in the program's own files, through any alias, is the
// callee of a call, by name or as `namespace.name(...)`, so every call is resolved against the overloads.
// A function value doesn't have to be: tsc lets a function whose parameter is required go where that
// parameter is optional, so `const sort: (array: readonly number[], comparer?: Comparer<number>) => ... =
// sortAndDeduplicate` would reach the default at a T no overload fixed (Compiler's road on 986ad82d). The
// declarations' own names, import and export bindings and type queries move no value; any other
// reference, `.call`, `.bind` or a value read, keeps the pair refused.
func (lowering *lowering) onlyCalled(symbol *ast.Symbol) bool {
	only := true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if !only {
			return true
		}
		if node.Kind == ast.KindIdentifier {
			if referenced := lowering.checker.GetSymbolAtLocation(
				node,
			); referenced != nil && lowering.checker.SkipAlias(referenced) == symbol && !lowering.calledOrBound(node) {
				only = false
			}
			return false
		}
		node.ForEachChild(visit)
		return !only
	}
	for _, file := range lowering.program.Files() {
		if !file.IsDeclarationFile {
			file.AsNode().ForEachChild(visit)
		}
	}
	return only
}

// calledOrBound is whether a reference moves no function value: a declaration's own name, an import or
// export binding, a type query, or the callee of a call, written bare or as a property access's name.
func (lowering *lowering) calledOrBound(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindFunctionDeclaration:
		return parent.Name() == identifier
	case ast.KindImportSpecifier, ast.KindExportSpecifier, ast.KindImportClause, ast.KindNamespaceImport, ast.KindTypeQuery, ast.KindQualifiedName:
		return true
	}
	callee := identifier
	if parent.Kind == ast.KindPropertyAccessExpression {
		if parent.AsPropertyAccessExpression().Name() != identifier {
			return false
		}
		callee = parent
	}
	for callee.Parent != nil && callee.Parent.Kind == ast.KindParenthesizedExpression {
		callee = callee.Parent
	}
	return callee.Parent != nil &&
		callee.Parent.Kind == ast.KindCallExpression &&
		ast.SkipParentheses(callee.Parent.AsCallExpression().Expression) == ast.SkipParentheses(callee)
}

// overloadDefaultAbsent is the index of every implementation parameter whose absence reaching node
// proves: node is the right operand of a `??` (through parentheses, and up through every `??` it's the
// right operand of), and each left operand, through nested `??`, is one of implementation's
// parameters, optional, with no default, not a rest, and never assigned in its body. Anything else
// fails.
func (lowering *lowering) overloadDefaultAbsent(node *ast.Node, implementation *ast.Node) (map[int]bool, bool) {
	parameters := implementation.Parameters()
	absent := map[int]bool{}
	var leaves func(*ast.Node) bool
	leaves = func(left *ast.Node) bool {
		left = ast.SkipParentheses(left)
		if left.Kind == ast.KindBinaryExpression &&
			left.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionToken {
			return leaves(left.AsBinaryExpression().Left) && leaves(left.AsBinaryExpression().Right)
		}
		if !ast.IsIdentifier(left) {
			return false
		}
		symbol := lowering.symbol(left)
		for index, parameter := range parameters {
			declaration := parameter.AsParameterDeclaration()
			if lowering.symbol(parameter.Name()) != symbol {
				continue
			}
			if declaration.QuestionToken == nil ||
				declaration.Initializer != nil ||
				declaration.DotDotDotToken != nil ||
				lowering.predicateOverloadAssigned(implementation, symbol) {
				return false
			}
			absent[index] = true
			return true
		}
		return false
	}
	child := node
	for {
		parent := child.Parent
		for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
			child, parent = parent, parent.Parent
		}
		if parent == nil || parent.Kind != ast.KindBinaryExpression {
			break
		}
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindQuestionQuestionToken || binary.Right != child {
			break
		}
		if !leaves(binary.Left) {
			return nil, false
		}
		child = parent
	}
	return absent, true
}

// overloadAdmitsAbsence is whether a call through overload may leave every parameter in absent out:
// the overload has no parameter there, or it's optional, a rest, or typed to hold undefined or null.
func (lowering *lowering) overloadAdmitsAbsence(overload *ast.Node, absent map[int]bool) bool {
	parameters := overload.Parameters()
	for index := range absent {
		if index >= len(parameters) {
			continue
		}
		declaration := parameters[index].AsParameterDeclaration()
		if declaration.QuestionToken != nil || declaration.DotDotDotToken != nil {
			continue
		}
		declared := lowering.checker.GetTypeAtLocation(parameters[index])
		if declared.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsVoid) != 0 ||
			lowering.censusHasUndefined(declared) ||
			lowering.checker.GetNonNullableType(declared) != declared {
			continue
		}
		return false
	}
	return true
}

// overloadDefaultMapper is what the implementation's type parameters stand for under overload: its own
// binders by position when the counts match (censusOverload's alpha-renaming), and otherwise what the
// overload's parameters infer for them, failing when any has no witness.
func (lowering *lowering) overloadDefaultMapper(implementation, overload *ast.Node) (*typeMapper, bool) {
	served, declared := implementation.TypeParameters(), overload.TypeParameters()
	if len(served) == 0 {
		return nil, true
	}
	sources, targets := []*checker.Type{}, []*checker.Type{}
	if len(served) == len(declared) {
		for index, parameter := range served {
			sources = append(sources, lowering.checker.GetTypeAtLocation(parameter.Name()))
			targets = append(targets, lowering.checker.GetTypeAtLocation(declared[index].Name()))
		}
		return newTypeMapper(sources, targets), true
	}
	inferred := map[*checker.Type]*checker.Type{}
	for index, parameter := range implementation.Parameters() {
		if index < len(overload.Parameters()) {
			lowering.inferTypes(
				lowering.checker.GetTypeAtLocation(parameter),
				lowering.checker.GetTypeAtLocation(overload.Parameters()[index]),
				inferred,
			)
		}
	}
	for _, parameter := range served {
		source := lowering.checker.GetTypeAtLocation(parameter.Name())
		target, found := inferred[source]
		if !found || target == nil {
			return nil, false
		}
		sources = append(sources, source)
		targets = append(targets, target)
	}
	return newTypeMapper(sources, targets), true
}
