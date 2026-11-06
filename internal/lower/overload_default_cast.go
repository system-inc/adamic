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
func (l *lowering) overloadDefaultPair(node *ast.Node) bool {
	if !strings.HasSuffix(l.program.FileName(ast.GetSourceFileOfNode(node)), ".ts") {
		return false
	}
	as := node.AsAsExpression()
	inner := ast.SkipParentheses(as.Expression)
	if inner.Kind != ast.KindAsExpression || l.checker.GetTypeAtLocation(inner).Flags()&checker.TypeFlagsUnknown == 0 {
		// A lone hop, or any other cast, isn't a pair.
		return false
	}
	operand := inner.AsAsExpression().Expression
	implementation := ast.GetContainingFunction(node)
	if implementation == nil || implementation.Kind != ast.KindFunctionDeclaration || implementation.Body() == nil || implementation.Name() == nil {
		return false
	}
	absent, ok := l.overloadDefaultAbsent(node, implementation)
	if !ok || len(absent) == 0 {
		return false
	}
	symbol := l.symbol(implementation.Name())
	if symbol == nil {
		return false
	}
	overloads := 0
	for _, overload := range symbol.Declarations {
		if overload.Kind != ast.KindFunctionDeclaration || overload.Body() != nil {
			continue
		}
		overloads++
		if !l.overloadAdmitsAbsence(overload, absent) {
			continue
		}
		mapper, mapped := l.overloadDefaultMapper(implementation, overload)
		if !mapped {
			return false
		}
		from, target := l.checker.GetTypeAtLocation(operand), l.checker.GetTypeAtLocation(node)
		if mapper != nil {
			from, target = instantiateType(l.checker, from, mapper), instantiateType(l.checker, target, mapper)
		}
		if !l.castAssignable(from, target) {
			return false
		}
	}
	return overloads > 0
}

// overloadDefaultAbsent is the index of every implementation parameter whose absence reaching node
// proves: node is the right operand of a `??` (through parentheses, and up through every `??` it's the
// right operand of), and each left operand, through nested `??`, is one of implementation's
// parameters, optional, with no default, not a rest, and never assigned in its body. Anything else
// fails.
func (l *lowering) overloadDefaultAbsent(node *ast.Node, implementation *ast.Node) (map[int]bool, bool) {
	parameters := implementation.Parameters()
	absent := map[int]bool{}
	var leaves func(*ast.Node) bool
	leaves = func(left *ast.Node) bool {
		left = ast.SkipParentheses(left)
		if left.Kind == ast.KindBinaryExpression && left.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionToken {
			return leaves(left.AsBinaryExpression().Left) && leaves(left.AsBinaryExpression().Right)
		}
		if !ast.IsIdentifier(left) {
			return false
		}
		symbol := l.symbol(left)
		for index, parameter := range parameters {
			declaration := parameter.AsParameterDeclaration()
			if l.symbol(parameter.Name()) != symbol {
				continue
			}
			if declaration.QuestionToken == nil || declaration.Initializer != nil || declaration.DotDotDotToken != nil || l.predicateOverloadAssigned(implementation, symbol) {
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
func (l *lowering) overloadAdmitsAbsence(overload *ast.Node, absent map[int]bool) bool {
	parameters := overload.Parameters()
	for index := range absent {
		if index >= len(parameters) {
			continue
		}
		declaration := parameters[index].AsParameterDeclaration()
		if declaration.QuestionToken != nil || declaration.DotDotDotToken != nil {
			continue
		}
		declared := l.checker.GetTypeAtLocation(parameters[index])
		if declared.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsVoid) != 0 || l.censusHasUndefined(declared) || l.checker.GetNonNullableType(declared) != declared {
			continue
		}
		return false
	}
	return true
}

// overloadDefaultMapper is what the implementation's type parameters stand for under overload: its own
// binders by position when the counts match (censusOverload's alpha-renaming), and otherwise what the
// overload's parameters infer for them, failing when any has no witness.
func (l *lowering) overloadDefaultMapper(implementation, overload *ast.Node) (*typeMapper, bool) {
	served, declared := implementation.TypeParameters(), overload.TypeParameters()
	if len(served) == 0 {
		return nil, true
	}
	sources, targets := []*checker.Type{}, []*checker.Type{}
	if len(served) == len(declared) {
		for index, parameter := range served {
			sources = append(sources, l.checker.GetTypeAtLocation(parameter.Name()))
			targets = append(targets, l.checker.GetTypeAtLocation(declared[index].Name()))
		}
		return newTypeMapper(sources, targets), true
	}
	inferred := map[*checker.Type]*checker.Type{}
	for index, parameter := range implementation.Parameters() {
		if index < len(overload.Parameters()) {
			l.inferTypes(l.checker.GetTypeAtLocation(parameter), l.checker.GetTypeAtLocation(overload.Parameters()[index]), inferred)
		}
	}
	for _, parameter := range served {
		source := l.checker.GetTypeAtLocation(parameter.Name())
		target, found := inferred[source]
		if !found || target == nil {
			return nil, false
		}
		sources = append(sources, source)
		targets = append(targets, target)
	}
	return newTypeMapper(sources, targets), true
}
