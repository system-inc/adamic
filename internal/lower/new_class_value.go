package lower

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// newClassValue dispatches through a kept class constructor, rather than assuming
// that its result type identifies the runtime allocator. Ordinary function-based
// constructors retain their existing unsupported receiver semantics.
func (l *lowering) newClassValue(node *ast.Node) (ir.Expression, bool, error) {
	created := node.AsNewExpression()
	target := l.concrete(l.checker.GetTypeAtLocation(created.Expression))
	if len(l.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 1 {
		return nil, false, nil
	}
	declarations := []*ast.Node{}
	for _, declaration := range l.classes {
		if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAbstract|ast.ModifierFlagsAmbient) {
			continue
		}
		if !l.needsStatics(declaration) {
			// Such classes currently have no first-class constructor value.
			continue
		}
		if l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(l.symbol(declaration.Name())), target) {
			if len(declaration.TypeParameters()) != 0 {
				return nil, true, l.notYet(node, "a generic class constructor value")
			}
			declarations = append(declarations, declaration)
		}
	}
	if len(declarations) == 0 {
		return nil, false, nil
	}
	sort.Slice(declarations, func(i, j int) bool {
		return l.program.Where(declarations[i]) < l.program.Where(declarations[j])
	})
	type candidate struct{ identity, constructor int }
	candidates := []candidate{}
	for _, declaration := range declarations {
		signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(l.symbol(declaration.Name())), checker.SignatureKindConstruct)
		if len(signatures) != 1 {
			return nil, true, l.notYet(node, "an overloaded class constructor value")
		}
		instance, err := l.instantiate(declaration, l.checker.GetReturnTypeOfSignature(signatures[0]), node)
		if err != nil {
			return nil, true, err
		}
		static, err := l.staticInstance(declaration)
		if err != nil {
			return nil, true, err
		}
		candidates = append(candidates, candidate{identity: static.class, constructor: instance.constructor})
	}
	constructor, err := l.classConstructorValue(created.Expression)
	if err != nil {
		return nil, true, err
	}
	if constructor.Type() != ir.Object {
		return nil, true, l.notYet(node, "a constructor value without class storage")
	}
	arguments := []ir.Expression{constructor}
	for _, argument := range nodesOf(created.Arguments) {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, l.notYet(argument, "a spread into a constructor value")
		}
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, value)
	}
	b := l.libraryArrayBuilder(arguments)
	receiver := b.read(b.parameters[0])
	for _, candidate := range candidates {
		function := l.result.Functions[candidate.constructor]
		if len(function.Parameters) != len(arguments)-1 {
			return nil, true, l.notYet(node, "a constructor value with omitted or rest arguments")
		}
		forwarded := []ir.Expression{}
		for index, parameter := range function.Parameters {
			value := fit(b.read(b.parameters[index+1]), l.result.Locals[parameter].Type)
			if value.Type() != l.result.Locals[parameter].Type {
				return nil, true, l.notYet(node, "a constructor value argument with a different native representation")
			}
			forwarded = append(forwarded, value)
		}
		b.body = append(b.body, ir.If{
			Condition: ir.InstanceOf{Value: receiver, Class: candidate.identity, Exact: true},
			Then:      []ir.Statement{ir.Return{Value: ir.Call{Function: candidate.constructor, Arguments: forwarded, Returns: ir.Object}}},
		})
	}
	// Never silently substitute a statically guessed allocator for an unknown value.
	b.body = append(b.body, ir.Panic{Message: ir.StringConstant{Index: l.constant("a constructor value without a registered class allocator")}})
	return b.finish("class_value_new", ir.Undefined{Of: ir.Object}), true, nil
}

// Keep constructor-cache assignment inside the branch that initializes it. The
// existing general assignment lowering supplies checks and ownership. Global
// bindings use a direct helper; closed lexical initializers capture one cache cell.
func (l *lowering) classConstructorValue(node *ast.Node) (ir.Expression, error) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindBinaryExpression {
		return l.expression(node)
	}
	binary := node.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindQuestionQuestionToken, ast.KindBarBarToken:
		left, err := l.classConstructorValue(binary.Left)
		if err != nil {
			return nil, err
		}
		right, err := l.classConstructorValue(binary.Right)
		if err != nil {
			return nil, err
		}
		if left.Type() != ir.Object || right.Type() != ir.Object {
			return nil, l.notYet(node, "a constructor cache without class storage")
		}
		// A constructor object is truthy; its only falsy stored values are nullish.
		return ir.Coalesce{Value: left, Fallback: right, Of: ir.Object}, nil
	case ast.KindEqualsToken:
		target := ast.SkipParentheses(binary.Left)
		if !ast.IsIdentifier(target) {
			return nil, l.notYet(node, "a constructor-cache assignment without a known identifier binding")
		}
		local, known := l.local(target)
		if !known {
			return nil, l.notYet(node, "a constructor-cache assignment without a known identifier binding")
		}
		statements, err := l.assignment(node)
		if err != nil {
			return nil, err
		}
		if l.result.Locals[local].Type != ir.Object {
			return nil, l.notYet(node, "a constructor cache without class storage")
		}
		b := l.libraryArrayBuilder(nil)
		b.body = statements
		if !l.result.Locals[local].Global {
			// This helper captures exactly the cache cell. Require a closed initializer,
			// so no other caller local can be read from the generated closure.
			if len(statements) != 1 {
				return nil, l.notYet(node, "a lexical constructor-cache initializer with additional statements")
			}
			assignment, plain := statements[0].(ir.Assign)
			if !plain || assignment.Local != local {
				return nil, l.notYet(node, "a lexical constructor-cache initializer without a plain assignment")
			}
			closed := false
			switch value := assignment.Value.(type) {
			case ir.Read:
				closed = l.result.Locals[value.Local].Global
			case ir.Call:
				targets := l.result.CallTargets(value)
				closed = len(value.Arguments) == 0 && len(targets) > 0
				for _, target := range targets {
					initializer := l.result.Functions[target]
					closed = closed && !initializer.Closure && len(initializer.Environment) == 0
				}
			}
			if !closed {
				return nil, l.notYet(node, "a lexical constructor-cache initializer needing additional captures")
			}
			l.result.Locals[local].Captured = true
			b.finish("class_constructor_cache", b.read(local))
			function := &l.result.Functions[b.function]
			function.Closure = true
			function.Environment = []int{local}
			return ir.CallClosure{Closure: ir.MakeClosure{Function: b.function}, Returns: ir.Object}, nil
		}
		return b.finish("class_constructor_cache", b.read(local)), nil
	default:
		return l.expression(node)
	}
}
