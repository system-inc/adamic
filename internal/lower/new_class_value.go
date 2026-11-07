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
func (lowering *lowering) newClassValue(node *ast.Node) (ir.Expression, bool, error) {
	created := node.AsNewExpression()
	target := lowering.concrete(lowering.checker.GetTypeAtLocation(created.Expression))
	if len(lowering.checker.GetSignaturesOfType(target, checker.SignatureKindConstruct)) != 1 {
		return nil, false, nil
	}
	declarations := []*ast.Node{}
	for _, declaration := range lowering.classes {
		if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAbstract|ast.ModifierFlagsAmbient) {
			continue
		}
		if !lowering.needsStatics(declaration) {
			// Such classes currently have no first-class constructor value.
			continue
		}
		if lowering.checker.IsTypeAssignableTo(
			lowering.checker.GetTypeOfSymbol(lowering.symbol(declaration.Name())),
			target,
		) {
			if len(declaration.TypeParameters()) != 0 {
				return nil, true, lowering.notYet(node, "a generic class constructor value")
			}
			declarations = append(declarations, declaration)
		}
	}
	if len(declarations) == 0 {
		return nil, false, nil
	}
	sort.Slice(declarations, func(i, j int) bool {
		return lowering.program.Where(declarations[i]) < lowering.program.Where(declarations[j])
	})
	type candidate struct{ identity, constructor int }
	candidates := []candidate{}
	for _, declaration := range declarations {
		signatures := lowering.checker.GetSignaturesOfType(
			lowering.checker.GetTypeOfSymbol(lowering.symbol(declaration.Name())),
			checker.SignatureKindConstruct,
		)
		if len(signatures) != 1 {
			return nil, true, lowering.notYet(node, "an overloaded class constructor value")
		}
		instance, instantiateError := lowering.instantiate(
			declaration,
			lowering.checker.GetReturnTypeOfSignature(signatures[0]),
			node,
		)
		if instantiateError != nil {
			return nil, true, instantiateError
		}
		static, staticInstanceError := lowering.staticInstance(declaration)
		if staticInstanceError != nil {
			return nil, true, staticInstanceError
		}
		candidates = append(candidates, candidate{identity: static.class, constructor: instance.constructor})
	}
	constructor, classConstructorValueError := lowering.classConstructorValue(created.Expression)
	if classConstructorValueError != nil {
		return nil, true, classConstructorValueError
	}
	if constructor.Type() != ir.Object {
		return nil, true, lowering.notYet(node, "a constructor value without class storage")
	}
	arguments := []ir.Expression{constructor}
	for _, argument := range nodesOf(created.Arguments) {
		if argument.Kind == ast.KindSpreadElement {
			return nil, true, lowering.notYet(argument, "a spread into a constructor value")
		}
		value, expressionError := lowering.expression(argument)
		if expressionError != nil {
			return nil, true, expressionError
		}
		arguments = append(arguments, value)
	}
	b := lowering.libraryArrayBuilder(arguments)
	receiver := b.read(b.parameters[0])
	for _, candidate := range candidates {
		function := lowering.result.Functions[candidate.constructor]
		if len(function.Parameters) != len(arguments)-1 {
			return nil, true, lowering.notYet(node, "a constructor value with omitted or rest arguments")
		}
		forwarded := []ir.Expression{}
		for index, parameter := range function.Parameters {
			value := fit(b.read(b.parameters[index+1]), lowering.result.Locals[parameter].Type)
			if value.Type() != lowering.result.Locals[parameter].Type {
				return nil, true, lowering.notYet(
					node,
					"a constructor value argument with a different native representation",
				)
			}
			forwarded = append(forwarded, value)
		}
		b.body = append(b.body, ir.If{
			Condition: ir.InstanceOf{Value: receiver, Class: candidate.identity, Exact: true},
			Then: []ir.Statement{
				ir.Return{Value: ir.Call{Function: candidate.constructor, Arguments: forwarded, Returns: ir.Object}},
			},
		})
	}
	// Never silently substitute a statically guessed allocator for an unknown value.
	b.body = append(b.body, ir.Panic{
		Message: ir.StringConstant{
			Index: lowering.constant("a constructor value without a registered class allocator"),
		},
	})
	return b.finish("class_value_new", ir.Undefined{Of: ir.Object}), true, nil
}

// Keep constructor-cache assignment inside the branch that initializes it. The
// existing general assignment lowering supplies checks and ownership. Global
// bindings use a direct helper; closed lexical initializers capture one cache cell.
func (lowering *lowering) classConstructorValue(node *ast.Node) (ir.Expression, error) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindBinaryExpression {
		return lowering.expression(node)
	}
	binary := node.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindQuestionQuestionToken, ast.KindBarBarToken:
		left, classConstructorValueError := lowering.classConstructorValue(binary.Left)
		if classConstructorValueError != nil {
			return nil, classConstructorValueError
		}
		right, classConstructorValueError := lowering.classConstructorValue(binary.Right)
		if classConstructorValueError != nil {
			return nil, classConstructorValueError
		}
		if left.Type() != ir.Object || right.Type() != ir.Object {
			return nil, lowering.notYet(node, "a constructor cache without class storage")
		}
		// A constructor object is truthy; its only falsy stored values are nullish.
		return ir.Coalesce{Value: left, Fallback: right, Of: ir.Object}, nil
	case ast.KindEqualsToken:
		target := ast.SkipParentheses(binary.Left)
		if !ast.IsIdentifier(target) {
			return nil, lowering.notYet(node, "a constructor-cache assignment without a known identifier binding")
		}
		local, known := lowering.local(target)
		if !known {
			return nil, lowering.notYet(node, "a constructor-cache assignment without a known identifier binding")
		}
		statements, assignmentError := lowering.assignment(node)
		if assignmentError != nil {
			return nil, assignmentError
		}
		if lowering.result.Locals[local].Type != ir.Object {
			return nil, lowering.notYet(node, "a constructor cache without class storage")
		}
		b := lowering.libraryArrayBuilder(nil)
		b.body = statements
		if !lowering.result.Locals[local].Global {
			// This helper captures exactly the cache cell. Require a closed initializer,
			// so no other caller local can be read from the generated closure.
			if len(statements) != 1 {
				return nil, lowering.notYet(node, "a lexical constructor-cache initializer with additional statements")
			}
			assignment, plain := statements[0].(ir.Assign)
			if !plain || assignment.Local != local {
				return nil, lowering.notYet(node, "a lexical constructor-cache initializer without a plain assignment")
			}
			closed := false
			switch value := assignment.Value.(type) {
			case ir.Read:
				closed = lowering.result.Locals[value.Local].Global
			case ir.Call:
				targets := lowering.result.CallTargets(value)
				closed = len(value.Arguments) == 0 && len(targets) > 0
				for _, target := range targets {
					initializer := lowering.result.Functions[target]
					closed = closed && !initializer.Closure && len(initializer.Environment) == 0
				}
			}
			if !closed {
				return nil, lowering.notYet(node, "a lexical constructor-cache initializer needing additional captures")
			}
			lowering.result.Locals[local].Captured = true
			b.finish("class_constructor_cache", b.read(local))
			function := &lowering.result.Functions[b.function]
			function.Closure = true
			function.Environment = []int{local}
			return ir.CallClosure{Closure: ir.MakeClosure{Function: b.function}, Returns: ir.Object}, nil
		}
		return b.finish("class_constructor_cache", b.read(local)), nil
	default:
		return lowering.expression(node)
	}
}
