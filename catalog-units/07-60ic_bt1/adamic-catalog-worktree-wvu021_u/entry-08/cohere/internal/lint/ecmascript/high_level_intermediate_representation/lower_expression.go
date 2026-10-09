// Expression lowering.
//
// Every expression lowers to a Place: the value it produces. An expression that branches mid
// evaluation - `&&`, `?:`, an optional chain - ends the current block and produces its value in the
// fallthrough, which is why these return a Place rather than an InstructionValue.
package high_level_intermediate_representation

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/static_single_assignment"
)

// lowerExpressionToPlace lowers an expression and returns the value it produces.
func (b *builder) lowerExpressionToPlace(node *ast.Node) Place {
	if node == nil {
		return b.emit(&Primitive{Value: nil}, nil)
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return b.lowerIdentifier(node)

	case ast.KindNumericLiteral:
		return b.emit(&Primitive{Value: node.Text()}, node)
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return b.emit(&Primitive{Value: node.Text()}, node)
	case ast.KindTrueKeyword:
		return b.emit(&Primitive{Value: true}, node)
	case ast.KindFalseKeyword:
		return b.emit(&Primitive{Value: false}, node)
	case ast.KindNullKeyword:
		return b.emit(&Primitive{Value: nil}, node)
	case ast.KindBigIntLiteral:
		return b.emit(&Primitive{Value: node.Text()}, node)

	case ast.KindRegularExpressionLiteral:
		pattern, flags := splitRegExp(node.Text())
		return b.emit(&RegExpLiteral{Pattern: pattern, Flags: flags}, node)

	case ast.KindThisKeyword:
		return b.emit(&LoadGlobal{Name: "this", BindingKind: GlobalBindingKindGlobal}, node)

	case ast.KindParenthesizedExpression:
		return b.lowerExpressionToPlace(node.AsParenthesizedExpression().Expression)

	case ast.KindAsExpression:
		expression := node.AsAsExpression()
		value := b.lowerExpressionToPlace(expression.Expression)
		return b.emit(&TypeCastExpression{Value: value, Node: expression.Type}, node)

	case ast.KindSatisfiesExpression:
		expression := node.AsSatisfiesExpression()
		value := b.lowerExpressionToPlace(expression.Expression)
		return b.emit(&TypeCastExpression{Value: value, Node: expression.Type}, node)

	case ast.KindNonNullExpression:
		return b.lowerExpressionToPlace(node.AsNonNullExpression().Expression)

	case ast.KindTypeAssertionExpression:
		expression := node.AsTypeAssertion()
		value := b.lowerExpressionToPlace(expression.Expression)
		return b.emit(&TypeCastExpression{Value: value, Node: expression.Type}, node)

	case ast.KindPropertyAccessExpression:
		return b.lowerPropertyAccess(node)

	case ast.KindElementAccessExpression:
		return b.lowerElementAccess(node)

	case ast.KindCallExpression:
		return b.lowerCallExpression(node)

	case ast.KindNewExpression:
		return b.lowerNewExpression(node)

	case ast.KindBinaryExpression:
		return b.lowerBinaryExpression(node)

	case ast.KindPrefixUnaryExpression:
		return b.lowerPrefixUnary(node)

	case ast.KindPostfixUnaryExpression:
		return b.lowerPostfixUnary(node)

	case ast.KindTypeOfExpression:
		value := b.lowerExpressionToPlace(node.AsTypeOfExpression().Expression)
		return b.emit(&UnaryExpression{Operator: "typeof", Value: value}, node)

	case ast.KindVoidExpression:
		value := b.lowerExpressionToPlace(node.AsVoidExpression().Expression)
		return b.emit(&UnaryExpression{Operator: "void", Value: value}, node)

	case ast.KindDeleteExpression:
		return b.lowerDeleteExpression(node)

	case ast.KindAwaitExpression:
		value := b.lowerExpressionToPlace(node.AsAwaitExpression().Expression)
		return b.emit(&Await{Value: value}, node)

	case ast.KindConditionalExpression:
		return b.lowerConditionalExpression(node)

	case ast.KindObjectLiteralExpression:
		return b.lowerObjectLiteral(node)

	case ast.KindArrayLiteralExpression:
		return b.lowerArrayLiteral(node)

	case ast.KindTemplateExpression:
		return b.lowerTemplateExpression(node)

	case ast.KindTaggedTemplateExpression:
		return b.lowerTaggedTemplate(node)

	case ast.KindFunctionExpression, ast.KindArrowFunction:
		return b.lowerFunctionExpression(node)

	case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
		return b.lowerJsxElement(node)

	case ast.KindJsxFragment:
		return b.lowerJsxFragment(node)

	case ast.KindJsxExpression:
		inner := node.AsJsxExpression().Expression
		if inner == nil {
			return b.emit(&Primitive{Value: nil}, node)
		}
		return b.lowerExpressionToPlace(inner)

	case ast.KindSpreadElement:
		return b.lowerExpressionToPlace(node.AsSpreadElement().Expression)

	case ast.KindMetaProperty:
		return b.emit(&MetaProperty{Meta: "import", Property: "meta"}, node)

	default:
		return b.emit(&UnsupportedNode{Node: node, Reason: node.Kind.String()}, node)
	}
}

// lowerIdentifier resolves a name to a value.
//
// Resolution is by SYMBOL, via `symbolOf`, rather than by a scope tree this package maintains. That
// is the payoff of having a resident checker: shadowing, hoisting, and closure capture are already
// answered correctly, including through imports, and re-deriving them here would be a second
// implementation that can disagree with the first.
//
// # What happens without a checker, stated because it is silent
//
// This resolves nothing when `builder.typeChecker` is nil, and the failure produces a well-formed
// graph rather than an error: every reference takes the `LoadGlobal` path below. Nothing downstream
// can tell that apart from a program made entirely of globals. See `Lower` for the measurement.
func (b *builder) lowerIdentifier(node *ast.Node) Place {
	name := node.Text()
	if name == "undefined" {
		return b.emit(&Primitive{Value: nil}, node)
	}

	symbol := b.symbolOf(node)
	if symbol != nil {
		if identifier, ok := b.identifiers[symbol]; ok {
			place := Place{Identifier: identifier, Range: rangeOf(node)}
			return b.emit(&LoadLocal{Place: place}, node)
		}
		// Not declared here, but declared by an enclosing function: a capture. This is the case
		// that used to fall through to LoadGlobal, making a closed-over variable indistinguishable
		// from a true global and stopping every analysis at the function boundary.
		if place, ok := b.captureOf(symbol); ok {
			place.Range = rangeOf(node)
			return b.emit(&LoadContext{Place: place}, node)
		}
	}

	// Unknown to every enclosing function: a true global, an import, or a module-scope binding.
	// Retain the checker's declaration provenance rather than conflating those three cases.
	return b.emit(moduleGlobalLoad(node, symbol), node)
}

func (b *builder) lowerPropertyAccess(node *ast.Node) Place {
	if isOptionalChainLink(node) {
		return b.lowerOptionalChain(node, nil)
	}
	expression := node.AsPropertyAccessExpression()
	object := b.lowerExpressionToPlace(expression.Expression)
	optional := expression.QuestionDotToken != nil
	name := ""
	if expression.Name() != nil {
		name = expression.Name().Text()
	}
	return b.emit(&PropertyLoad{Object: object, Property: name, Optional: optional}, node)
}

func (b *builder) lowerElementAccess(node *ast.Node) Place {
	expression := node.AsElementAccessExpression()
	object := b.lowerExpressionToPlace(expression.Expression)
	property := b.lowerExpressionToPlace(expression.ArgumentExpression)
	optional := expression.QuestionDotToken != nil
	return b.emit(&ComputedLoad{Object: object, Property: property, Optional: optional}, node)
}

// calleeModuleOrigin names the React export a callee is, through imports and re-exports in other files,
// which rule.ExportNameIn reads within its imports' shapes.
func (b *builder) calleeModuleOrigin(expression *ast.Node) ModuleExportOrigin {
	if export := rule.ExportNameIn(b.typeChecker, expression, "react"); export != "" {
		return ModuleExportOrigin{Module: "react", Export: export}
	}
	return ModuleExportOrigin{}
}

// lowerCallExpression lowers a call.
//
// A call whose callee is a member access becomes a MethodCall, keeping the receiver, so an effect
// pass can attribute mutation to the object. Lowering it as a property load plus a plain call would
// lose the relationship between the loaded function and the object it came from.
func (b *builder) lowerCallExpression(node *ast.Node) Place {
	expression := node.AsCallExpression()
	optional := expression.QuestionDotToken != nil

	callee := expression.Expression
	origin := b.calleeModuleOrigin(callee)
	if callee != nil && callee.Kind == ast.KindPropertyAccessExpression {
		access := callee.AsPropertyAccessExpression()
		receiver := b.lowerExpressionToPlace(access.Expression)
		name := ""
		if access.Name() != nil {
			name = access.Name().Text()
		}
		property := b.emit(&Primitive{Value: name}, access.Name())
		args := b.lowerArguments(expression.Arguments)
		return b.emit(&MethodCall{
			CalleeOrigin: origin,
			Receiver:     receiver,
			Property:     property,
			Args:         args,
			Optional:     optional,
		}, node)
	}
	if callee != nil && callee.Kind == ast.KindElementAccessExpression {
		access := callee.AsElementAccessExpression()
		receiver := b.lowerExpressionToPlace(access.Expression)
		property := b.lowerExpressionToPlace(access.ArgumentExpression)
		args := b.lowerArguments(expression.Arguments)
		return b.emit(&MethodCall{
			CalleeOrigin: origin,
			Receiver:     receiver,
			Property:     property,
			Args:         args,
			Optional:     optional,
		}, node)
	}

	calleePlace := b.lowerExpressionToPlace(callee)
	args := b.lowerArguments(expression.Arguments)
	return b.emit(&CallExpression{Callee: calleePlace, Args: args, Optional: optional, CalleeOrigin: origin}, node)
}

func (b *builder) lowerNewExpression(node *ast.Node) Place {
	expression := node.AsNewExpression()
	callee := b.lowerExpressionToPlace(expression.Expression)
	args := b.lowerArguments(expression.Arguments)
	return b.emit(&NewExpression{Callee: callee, Args: args}, node)
}

func (b *builder) lowerArguments(list *ast.NodeList) []Argument {
	if list == nil {
		return nil
	}
	args := make([]Argument, 0, len(list.Nodes))
	for _, argument := range list.Nodes {
		if argument.Kind == ast.KindSpreadElement {
			value := b.lowerExpressionToPlace(argument.AsSpreadElement().Expression)
			args = append(args, Argument{Place: value, Spread: true})
			continue
		}
		args = append(args, Argument{Place: b.lowerExpressionToPlace(argument)})
	}
	return args
}

// lowerBinaryExpression lowers a binary operator, an assignment, or a logical operator.
//
// The three are one AST node in TypeScript and three different things here: `&&` is control flow,
// `=` is a store, and `+` is an instruction.
func (b *builder) lowerBinaryExpression(node *ast.Node) Place {
	expression := node.AsBinaryExpression()
	operator := expression.OperatorToken.Kind

	switch operator {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		return b.lowerLogicalExpression(node)

	case ast.KindEqualsToken:
		value := b.lowerExpressionToPlace(expression.Right)
		return b.lowerAssignmentTarget(expression.Left, value, InstructionKindReassign)

	case ast.KindCommaToken:
		b.lowerExpressionToPlace(expression.Left)
		return b.lowerExpressionToPlace(expression.Right)
	}

	if isCompoundAssignment(operator) {
		return b.lowerCompoundAssignment(node)
	}

	left := b.lowerExpressionToPlace(expression.Left)
	right := b.lowerExpressionToPlace(expression.Right)
	return b.emit(&BinaryExpression{
		Left:     left,
		Operator: operatorText(operator),
		Right:    right,
	}, node)
}

// lowerLogicalExpression lowers `&&`, `||`, and `??`.
//
// These short-circuit, so the right operand is not evaluated on every path. That is control flow,
// and it lowers to a Logical terminal with the two operands in separate blocks. The value is the
// result temporary, stored to on both paths.
func (b *builder) lowerLogicalExpression(node *ast.Node) Place {
	expression := node.AsBinaryExpression()
	result := b.newTemporary(node)

	testBlock := b.reserve(BlockKindValue)
	fallthroughBlock := b.reserve(BlockKindBlock)

	b.terminateAndEnter(&Logical{
		Operator:    operatorText(expression.OperatorToken.Kind),
		Test:        testBlock.Id,
		Fallthrough: fallthroughBlock.Id,
	}, testBlock)

	// The arms write DISTINCT identifiers under the result's declaration, so construction sees two
	// definitions of one binding and merges them. See `newTemporaryUnder`.
	shared := b.function.Identifiers[result.Identifier].Declaration

	left := b.lowerExpressionToPlace(expression.Left)

	// Both arms get a value block, including the short-circuiting one.
	//
	// Sending that path straight to the fallthrough leaves it with no write, so the join has one
	// incoming value and no phi is minted. Measured before this: `a && b` merged only because its
	// short circuit fell through writeless while the other arm wrote, and `a || b` and `a ?? b`
	// never merged at all, since they swap the arms and hand the fallthrough to the branch.
	//
	// Upstream gives both a block: on `const x = a ?? []` its graph is a `branch` value block plus
	// two `goto` value blocks feeding a join carrying one phi. This reproduces that for every
	// operator.
	shortCircuitBlock := b.reserve(BlockKindValue)
	rightBlock := b.reserve(BlockKindValue)
	consequent, alternate := rightBlock, shortCircuitBlock
	if expression.OperatorToken.Kind == ast.KindBarBarToken ||
		expression.OperatorToken.Kind == ast.KindQuestionQuestionToken {
		// `a || b` and `a ?? b` evaluate the right operand when the test is falsy or nullish.
		consequent, alternate = shortCircuitBlock, rightBlock
	}
	b.terminateWith(&Branch{
		Test:        left,
		Consequent:  consequent.Id,
		Alternate:   alternate.Id,
		Fallthrough: fallthroughBlock.Id,
	})

	b.enter(shortCircuitBlock)
	b.emitTo(b.newTemporaryUnder(expression.Left, shared), &LoadLocal{Place: left},
		expression.Left)
	b.gotoBlock(fallthroughBlock.Id, GotoVariantBreak)

	b.enter(rightBlock)
	right := b.lowerExpressionToPlace(expression.Right)
	b.emitTo(b.newTemporaryUnder(expression.Right, shared), &LoadLocal{Place: right},
		expression.Right)
	b.gotoBlock(fallthroughBlock.Id, GotoVariantBreak)

	b.enter(fallthroughBlock)
	return result
}

// lowerConditionalExpression lowers `a ? b : c`.
func (b *builder) lowerConditionalExpression(node *ast.Node) Place {
	expression := node.AsConditionalExpression()
	result := b.newTemporary(node)

	testBlock := b.reserve(BlockKindValue)
	fallthroughBlock := b.reserve(BlockKindBlock)

	b.terminateAndEnter(&Ternary{
		Test:        testBlock.Id,
		Fallthrough: fallthroughBlock.Id,
	}, testBlock)

	test := b.lowerExpressionToPlace(expression.Condition)

	consequent := b.reserve(BlockKindValue)
	alternate := b.reserve(BlockKindValue)
	b.terminateWith(&Branch{
		Test:        test,
		Consequent:  consequent.Id,
		Alternate:   alternate.Id,
		Fallthrough: fallthroughBlock.Id,
	})

	b.enter(consequent)
	whenTrue := b.lowerExpressionToPlace(expression.WhenTrue)
	b.emitTo(result, &LoadLocal{Place: whenTrue}, expression.WhenTrue)
	b.gotoBlock(fallthroughBlock.Id, GotoVariantBreak)

	b.enter(alternate)
	whenFalse := b.lowerExpressionToPlace(expression.WhenFalse)
	b.emitTo(result, &LoadLocal{Place: whenFalse}, expression.WhenFalse)
	b.gotoBlock(fallthroughBlock.Id, GotoVariantBreak)

	b.enter(fallthroughBlock)
	return result
}

// lowerAssignmentTarget stores a value into an assignment target and returns the value.
//
// Assignment is an expression: `a = (b = 1)` needs the inner assignment to produce a value.
func (b *builder) lowerAssignmentTarget(target *ast.Node, value Place, kind InstructionKind) Place {
	if target == nil {
		return value
	}
	switch target.Kind {
	case ast.KindIdentifier:
		symbol := b.symbolOf(target)
		if symbol != nil {
			if _, known := b.identifiers[symbol]; known {
				place := b.bind(target.Text(), symbol, target)
				// A binding this function declared is still a CONTEXT binding when a closure inside
				// it reads the value and something reassigns it, which is upstream's second rule at
				// `FindContextIdentifiers.ts:108`. The write is in the outer function either way, so
				// the writer's position cannot decide it -- see `context_identifiers.go`.
				//
				// It matters because the two lower to different effects: `StoreContext` emits a
				// `Mutate` on the binding, widening its mutable range past the memo call that
				// captured it, and without that widening the scope closes before the marker and the
				// rule stays silent where upstream reports a dependency that may be mutated later.
				if b.contextual[symbol] {
					if b.function.ContextDeclarations == nil {
						b.function.ContextDeclarations = map[static_single_assignment.DeclarationId]bool{}
					}
					if identifier := b.function.Identifiers[place.Identifier]; identifier != nil {
						b.function.ContextDeclarations[identifier.Declaration] = true
					}
					b.emit(&StoreContext{LValue: place, Value: value, Kind: kind}, target)
					return value
				}
				b.emit(&StoreLocal{LValue: place, Value: value, Kind: kind}, target)
				return value
			}
			// A write to a binding an enclosing function declared. Emitting StoreGlobal here was a
			// real over-report: `let n = 0; const f = () => { n = 1; };` claimed a global write.
			if place, ok := b.captureOf(symbol); ok {
				place.Range = rangeOf(target)
				b.emit(&StoreContext{LValue: place, Value: value, Kind: kind}, target)
				return value
			}
		}
		// Not a binding this function declared and unknown to every enclosing one: a global.
		b.emit(&StoreGlobal{Name: target.Text(), Value: value}, target)
		return value

	case ast.KindPropertyAccessExpression:
		expression := target.AsPropertyAccessExpression()
		object := b.lowerExpressionToPlace(expression.Expression)
		name := ""
		if expression.Name() != nil {
			name = expression.Name().Text()
		}
		b.emit(&PropertyStore{Object: object, Property: name, Value: value}, target)
		return value

	case ast.KindElementAccessExpression:
		expression := target.AsElementAccessExpression()
		object := b.lowerExpressionToPlace(expression.Expression)
		property := b.lowerExpressionToPlace(expression.ArgumentExpression)
		b.emit(&ComputedStore{Object: object, Property: property, Value: value}, target)
		return value

	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		// Destructuring assignment: `({a} = o)` and `[a] = xs`.
		pattern := b.lowerPattern(target)
		b.emit(&Destructure{LValue: pattern, Value: value, Kind: kind, Pattern: pattern}, target)
		return value

	case ast.KindParenthesizedExpression:
		return b.lowerAssignmentTarget(target.AsParenthesizedExpression().Expression, value, kind)
	}

	b.emit(&UnsupportedNode{Node: target, Reason: "assignment target"}, target)
	return value
}

func (b *builder) lowerCompoundAssignment(node *ast.Node) Place {
	expression := node.AsBinaryExpression()
	current := b.lowerExpressionToPlace(expression.Left)
	right := b.lowerExpressionToPlace(expression.Right)
	updated := b.emit(&BinaryExpression{
		Left:     current,
		Operator: compoundOperatorText(expression.OperatorToken.Kind),
		Right:    right,
	}, node)
	return b.lowerAssignmentTarget(expression.Left, updated, InstructionKindReassign)
}

func (b *builder) lowerPrefixUnary(node *ast.Node) Place {
	expression := node.AsPrefixUnaryExpression()
	operator := expression.Operator

	if operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken {
		current := b.lowerExpressionToPlace(expression.Operand)
		result := b.emit(&PrefixUpdate{
			LValue:    current,
			Operation: operatorText(operator),
			Value:     current,
		}, node)
		b.lowerAssignmentTarget(expression.Operand, result, InstructionKindReassign)
		return result
	}

	value := b.lowerExpressionToPlace(expression.Operand)
	return b.emit(&UnaryExpression{Operator: operatorText(operator), Value: value}, node)
}

func (b *builder) lowerPostfixUnary(node *ast.Node) Place {
	expression := node.AsPostfixUnaryExpression()
	current := b.lowerExpressionToPlace(expression.Operand)
	result := b.emit(&PostfixUpdate{
		LValue:    current,
		Operation: operatorText(expression.Operator),
		Value:     current,
	}, node)
	b.lowerAssignmentTarget(expression.Operand, result, InstructionKindReassign)
	return result
}

func (b *builder) lowerDeleteExpression(node *ast.Node) Place {
	target := node.AsDeleteExpression().Expression
	if target == nil {
		return b.emit(&Primitive{Value: true}, node)
	}
	switch target.Kind {
	case ast.KindPropertyAccessExpression:
		expression := target.AsPropertyAccessExpression()
		object := b.lowerExpressionToPlace(expression.Expression)
		name := ""
		if expression.Name() != nil {
			name = expression.Name().Text()
		}
		return b.emit(&PropertyDelete{Object: object, Property: name}, node)
	case ast.KindElementAccessExpression:
		expression := target.AsElementAccessExpression()
		object := b.lowerExpressionToPlace(expression.Expression)
		property := b.lowerExpressionToPlace(expression.ArgumentExpression)
		return b.emit(&ComputedDelete{Object: object, Property: property}, node)
	}
	b.lowerExpressionToPlace(target)
	return b.emit(&Primitive{Value: true}, node)
}

func (b *builder) lowerObjectLiteral(node *ast.Node) Place {
	expression := node.AsObjectLiteralExpression()
	properties := make([]ObjectProperty, 0, len(expression.Properties.Nodes))

	for _, property := range expression.Properties.Nodes {
		switch property.Kind {
		case ast.KindPropertyAssignment:
			assignment := property.AsPropertyAssignment()
			value := b.lowerExpressionToPlace(assignment.Initializer)
			properties = append(properties, b.objectPropertyKey(assignment.Name(), value))

		case ast.KindShorthandPropertyAssignment:
			assignment := property.AsShorthandPropertyAssignment()
			name := assignment.Name()
			value := b.lowerExpressionToPlace(name)
			properties = append(properties, ObjectProperty{Key: name.Text(), Value: value})

		case ast.KindSpreadAssignment:
			value := b.lowerExpressionToPlace(property.AsSpreadAssignment().Expression)
			properties = append(properties, ObjectProperty{Value: value, Spread: true})

		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			// An object method closes over the enclosing function exactly like any other nested
			// function. `ObjectMethod` carries no Captures field, so the pairing is not recorded
			// here, but the nested function's own `Context` is populated by the lowering and its
			// body reads captured names as LoadContext rather than LoadGlobal.
			nested, _ := b.lowerNestedFunction(property)
			if nested == nil {
				continue
			}
			id := FunctionId(len(b.function.Functions))
			b.function.Functions = append(b.function.Functions, nested)
			// A method's name can be computed, `async *[Symbol.asyncIterator]() {}`, and a computed
			// name has no text: reading it as text panicked every rule that lowers the function, on
			// TanStack Query's own tests (#zx5xvtg). The key goes through objectPropertyKey as a
			// property assignment's does, which lowers a computed name's expression after the value,
			// as React Compiler's BuildHIR lowers the method and then its key.
			key := ""
			if name := property.Name(); name != nil && name.Kind != ast.KindComputedPropertyName {
				key = name.Text()
			}
			value := b.emit(&ObjectMethod{Key: key, Function: id}, property)
			properties = append(properties, b.objectPropertyKey(property.Name(), value))

		default:
			value := b.emit(&UnsupportedNode{Node: property, Reason: "object member"}, property)
			properties = append(properties, ObjectProperty{Value: value})
		}
	}

	return b.emit(&ObjectExpression{Properties: properties}, node)
}

func (b *builder) objectPropertyKey(name *ast.Node, value Place) ObjectProperty {
	if name == nil {
		return ObjectProperty{Value: value}
	}
	if name.Kind == ast.KindComputedPropertyName {
		key := b.lowerExpressionToPlace(name.AsComputedPropertyName().Expression)
		return ObjectProperty{ComputedKey: &key, Value: value}
	}
	return ObjectProperty{Key: name.Text(), Value: value}
}

func (b *builder) lowerArrayLiteral(node *ast.Node) Place {
	expression := node.AsArrayLiteralExpression()
	elements := make([]ArrayElement, 0, len(expression.Elements.Nodes))

	for _, element := range expression.Elements.Nodes {
		switch element.Kind {
		case ast.KindOmittedExpression:
			elements = append(elements, ArrayElement{Hole: true})
		case ast.KindSpreadElement:
			value := b.lowerExpressionToPlace(element.AsSpreadElement().Expression)
			elements = append(elements, ArrayElement{Place: value, Spread: true})
		default:
			elements = append(elements, ArrayElement{Place: b.lowerExpressionToPlace(element)})
		}
	}

	return b.emit(&ArrayExpression{Elements: elements}, node)
}

func (b *builder) lowerTemplateExpression(node *ast.Node) Place {
	expression := node.AsTemplateExpression()
	quasis := []string{}
	if expression.Head != nil {
		quasis = append(quasis, expression.Head.Text())
	}
	subexprs := make([]Place, 0, len(expression.TemplateSpans.Nodes))
	for _, span := range expression.TemplateSpans.Nodes {
		templateSpan := span.AsTemplateSpan()
		subexprs = append(subexprs, b.lowerExpressionToPlace(templateSpan.Expression))
		if templateSpan.Literal != nil {
			quasis = append(quasis, templateSpan.Literal.Text())
		}
	}
	return b.emit(&TemplateLiteral{Quasis: quasis, Subexprs: subexprs}, node)
}

func (b *builder) lowerTaggedTemplate(node *ast.Node) Place {
	expression := node.AsTaggedTemplateExpression()
	tag := b.lowerExpressionToPlace(expression.Tag)

	quasis := []string{}
	var subexprs []Place
	if expression.Template != nil {
		if expression.Template.Kind == ast.KindTemplateExpression {
			template := expression.Template.AsTemplateExpression()
			if template.Head != nil {
				quasis = append(quasis, template.Head.Text())
			}
			for _, span := range template.TemplateSpans.Nodes {
				templateSpan := span.AsTemplateSpan()
				subexprs = append(subexprs, b.lowerExpressionToPlace(templateSpan.Expression))
				if templateSpan.Literal != nil {
					quasis = append(quasis, templateSpan.Literal.Text())
				}
			}
		} else {
			quasis = append(quasis, expression.Template.Text())
		}
	}

	return b.emit(&TaggedTemplateExpression{Tag: tag, Quasis: quasis, Subexprs: subexprs}, node)
}

// lowerFunctionExpression lowers a nested function and records what it captures.
func (b *builder) lowerFunctionExpression(node *ast.Node) Place {
	nested, captures := b.lowerNestedFunction(node)
	if nested == nil {
		return b.emit(&UnsupportedNode{Node: node, Reason: "function expression"}, node)
	}
	id := FunctionId(len(b.function.Functions))
	b.function.Functions = append(b.function.Functions, nested)
	return b.emit(&FunctionExpression{Function: id, Captures: captures}, node)
}

// ---------------------------------------------------------------------------
// JSX
// ---------------------------------------------------------------------------

func (b *builder) lowerJsxElement(node *ast.Node) Place {
	var opening *ast.Node
	var children []*ast.Node

	if node.Kind == ast.KindJsxSelfClosingElement {
		opening = node
	} else {
		element := node.AsJsxElement()
		opening = element.OpeningElement
		if element.Children != nil {
			children = element.Children.Nodes
		}
	}

	tagName, attributes := jsxOpeningParts(opening)
	tag := b.lowerJsxTag(tagName)
	props := b.lowerJsxAttributes(attributes)
	childPlaces := b.lowerJsxChildren(children)

	return b.emit(&JsxExpression{Tag: tag, Props: props, Children: childPlaces}, node)
}

func (b *builder) lowerJsxFragment(node *ast.Node) Place {
	fragment := node.AsJsxFragment()
	var children []*ast.Node
	if fragment.Children != nil {
		children = fragment.Children.Nodes
	}
	return b.emit(&JsxFragment{Children: b.lowerJsxChildren(children)}, node)
}

// lowerJsxTag resolves an element name.
//
// A lowercase name is a host element and stays a string; anything else is a value reference, which
// is what makes `<Foo />` a use of `Foo`.
func (b *builder) lowerJsxTag(name *ast.Node) JsxTag {
	if name == nil {
		return JsxTag{Name: "unknown"}
	}
	switch name.Kind {
	case ast.KindIdentifier:
		text := name.Text()
		if text != "" && text[0] >= 'a' && text[0] <= 'z' {
			return JsxTag{Name: text}
		}
		place := b.lowerIdentifier(name)
		return JsxTag{Place: &place}
	case ast.KindPropertyAccessExpression:
		place := b.lowerPropertyAccess(name)
		return JsxTag{Place: &place}
	}
	return JsxTag{Name: "unknown"}
}

func (b *builder) lowerJsxAttributes(attributes *ast.Node) []JsxAttribute {
	if attributes == nil {
		return nil
	}
	list := attributes.AsJsxAttributes()
	if list == nil || list.Properties == nil {
		return nil
	}

	props := make([]JsxAttribute, 0, len(list.Properties.Nodes))
	for _, property := range list.Properties.Nodes {
		switch property.Kind {
		case ast.KindJsxAttribute:
			attribute := property.AsJsxAttribute()
			name := ""
			if attribute.Name() != nil {
				name = attribute.Name().Text()
			}
			if attribute.Initializer == nil {
				// A bare attribute is `true`.
				value := b.emit(&Primitive{Value: true}, property)
				props = append(props, JsxAttribute{Name: name, Value: value})
				continue
			}
			props = append(props, JsxAttribute{
				Name:  name,
				Value: b.lowerExpressionToPlace(attribute.Initializer),
			})

		case ast.KindJsxSpreadAttribute:
			value := b.lowerExpressionToPlace(property.AsJsxSpreadAttribute().Expression)
			props = append(props, JsxAttribute{Value: value, Spread: true})
		}
	}
	return props
}

func (b *builder) lowerJsxChildren(children []*ast.Node) []Place {
	var places []Place
	for _, child := range children {
		switch child.Kind {
		case ast.KindJsxText:
			// `Node.Text()` panics on a JsxText node, so the field is read directly. The parser has
			// already answered the whitespace question in `ContainsOnlyTriviaWhiteSpaces`, which is
			// the JSX rule rather than a plain `TrimSpace`: whitespace containing a newline is not
			// rendered, whitespace on one line is.
			text := child.AsJsxText()
			if text.ContainsOnlyTriviaWhiteSpaces {
				continue
			}
			places = append(places, b.emit(&JsxText{Value: text.Text}, child))
		case ast.KindJsxExpression:
			inner := child.AsJsxExpression().Expression
			if inner == nil {
				continue
			}
			places = append(places, b.lowerExpressionToPlace(inner))
		default:
			places = append(places, b.lowerExpressionToPlace(child))
		}
	}
	return places
}

func jsxOpeningParts(opening *ast.Node) (*ast.Node, *ast.Node) {
	if opening == nil {
		return nil, nil
	}
	switch opening.Kind {
	case ast.KindJsxSelfClosingElement:
		element := opening.AsJsxSelfClosingElement()
		return element.TagName, element.Attributes
	case ast.KindJsxOpeningElement:
		element := opening.AsJsxOpeningElement()
		return element.TagName, element.Attributes
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Patterns
// ---------------------------------------------------------------------------

// lowerPattern converts a binding pattern or a destructuring assignment target to a Pattern.
//
// TypeScript represents `const {a} = o` and `({a} = o)` with different node kinds - a binding
// pattern and an object literal - so both are handled here and produce the same structure.
func (b *builder) lowerPattern(node *ast.Node) Pattern {
	if node == nil {
		return &PlacePattern{Place: b.newTemporary(nil)}
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return &PlacePattern{Place: b.bind(node.Text(), b.symbolOf(node), node)}

	case ast.KindObjectBindingPattern:
		return b.lowerObjectBindingPattern(node)

	case ast.KindArrayBindingPattern:
		return b.lowerArrayBindingPattern(node)

	case ast.KindObjectLiteralExpression:
		return b.lowerObjectAssignmentPattern(node)

	case ast.KindArrayLiteralExpression:
		return b.lowerArrayAssignmentPattern(node)

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		// `({a: o.x} = v)` targets a property rather than a binding.
		return &PlacePattern{Place: b.lowerExpressionToPlace(node)}
	}

	return &PlacePattern{Place: b.newTemporary(node)}
}

func (b *builder) lowerObjectBindingPattern(node *ast.Node) Pattern {
	pattern := node.AsBindingPattern()
	result := &ObjectPattern{}

	for _, element := range pattern.Elements.Nodes {
		binding := element.AsBindingElement()
		if binding.DotDotDotToken != nil {
			name := binding.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				place := b.bind(name.Text(), b.symbolOf(name), name)
				result.Rest = &place
			}
			continue
		}

		property := ObjectPatternProperty{}
		if binding.PropertyName != nil {
			if binding.PropertyName.Kind == ast.KindComputedPropertyName {
				key := b.lowerExpressionToPlace(binding.PropertyName.AsComputedPropertyName().Expression)
				property.ComputedKey = &key
			} else {
				property.Key = binding.PropertyName.Text()
			}
		} else if name := binding.Name(); name != nil && name.Kind == ast.KindIdentifier {
			property.Key = name.Text()
		}

		if binding.Initializer != nil {
			value := b.lowerExpressionToPlace(binding.Initializer)
			property.Default = &value
		}
		property.Value = b.lowerPattern(binding.Name())
		result.Properties = append(result.Properties, property)
	}

	return result
}

func (b *builder) lowerArrayBindingPattern(node *ast.Node) Pattern {
	pattern := node.AsBindingPattern()
	result := &ArrayPattern{}

	for _, element := range pattern.Elements.Nodes {
		// A hole in a BINDING pattern - the gap in `const [a, , b] = xs` - arrives as a
		// BindingElement with no name, NOT as an OmittedExpression. (An assignment pattern, the
		// `[a, , b] = xs` form, does use OmittedExpression; the two spellings are handled in their
		// own functions.) Measured against the parser rather than assumed: the first version of this
		// checked for OmittedExpression here and silently dropped every hole, which shifts every
		// later element left by one and is invisible in the printed form.
		if element.Kind == ast.KindOmittedExpression {
			result.Elements = append(result.Elements, ArrayPatternElement{})
			continue
		}
		binding := element.AsBindingElement()
		if binding.Name() == nil {
			result.Elements = append(result.Elements, ArrayPatternElement{})
			continue
		}
		if binding.DotDotDotToken != nil {
			name := binding.Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				place := b.bind(name.Text(), b.symbolOf(name), name)
				result.Rest = &place
			}
			continue
		}
		item := ArrayPatternElement{}
		if binding.Initializer != nil {
			value := b.lowerExpressionToPlace(binding.Initializer)
			item.Default = &value
		}
		item.Value = b.lowerPattern(binding.Name())
		result.Elements = append(result.Elements, item)
	}

	return result
}

func (b *builder) lowerObjectAssignmentPattern(node *ast.Node) Pattern {
	expression := node.AsObjectLiteralExpression()
	result := &ObjectPattern{}

	for _, property := range expression.Properties.Nodes {
		switch property.Kind {
		case ast.KindPropertyAssignment:
			assignment := property.AsPropertyAssignment()
			item := ObjectPatternProperty{Value: b.lowerPattern(assignment.Initializer)}
			if name := assignment.Name(); name != nil {
				if name.Kind == ast.KindComputedPropertyName {
					key := b.lowerExpressionToPlace(name.AsComputedPropertyName().Expression)
					item.ComputedKey = &key
				} else {
					item.Key = name.Text()
				}
			}
			result.Properties = append(result.Properties, item)

		case ast.KindShorthandPropertyAssignment:
			assignment := property.AsShorthandPropertyAssignment()
			name := assignment.Name()
			item := ObjectPatternProperty{Key: name.Text(), Value: b.lowerPattern(name)}
			if assignment.ObjectAssignmentInitializer != nil {
				value := b.lowerExpressionToPlace(assignment.ObjectAssignmentInitializer)
				item.Default = &value
			}
			result.Properties = append(result.Properties, item)

		case ast.KindSpreadAssignment:
			target := property.AsSpreadAssignment().Expression
			if target != nil && target.Kind == ast.KindIdentifier {
				place := b.bind(target.Text(), b.symbolOf(target), target)
				result.Rest = &place
			}
		}
	}

	return result
}

func (b *builder) lowerArrayAssignmentPattern(node *ast.Node) Pattern {
	expression := node.AsArrayLiteralExpression()
	result := &ArrayPattern{}

	for _, element := range expression.Elements.Nodes {
		switch element.Kind {
		case ast.KindOmittedExpression:
			result.Elements = append(result.Elements, ArrayPatternElement{})
		case ast.KindSpreadElement:
			target := element.AsSpreadElement().Expression
			if target != nil && target.Kind == ast.KindIdentifier {
				place := b.bind(target.Text(), b.symbolOf(target), target)
				result.Rest = &place
			}
		case ast.KindBinaryExpression:
			// `[a = 1] = xs`.
			binary := element.AsBinaryExpression()
			item := ArrayPatternElement{Value: b.lowerPattern(binary.Left)}
			value := b.lowerExpressionToPlace(binary.Right)
			item.Default = &value
			result.Elements = append(result.Elements, item)
		default:
			result.Elements = append(result.Elements, ArrayPatternElement{Value: b.lowerPattern(element)})
		}
	}

	return result
}

// ---------------------------------------------------------------------------
// Operator text
// ---------------------------------------------------------------------------

func isCompoundAssignment(kind ast.Kind) bool {
	switch kind {
	case ast.KindPlusEqualsToken, ast.KindMinusEqualsToken, ast.KindAsteriskEqualsToken,
		ast.KindSlashEqualsToken, ast.KindPercentEqualsToken, ast.KindAsteriskAsteriskEqualsToken,
		ast.KindLessThanLessThanEqualsToken, ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken, ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken, ast.KindCaretEqualsToken, ast.KindAmpersandAmpersandEqualsToken,
		ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

func compoundOperatorText(kind ast.Kind) string {
	text := operatorText(kind)
	return strings.TrimSuffix(text, "=")
}

func operatorText(kind ast.Kind) string {
	switch kind {
	case ast.KindPlusToken:
		return "+"
	case ast.KindMinusToken:
		return "-"
	case ast.KindAsteriskToken:
		return "*"
	case ast.KindSlashToken:
		return "/"
	case ast.KindPercentToken:
		return "%"
	case ast.KindAsteriskAsteriskToken:
		return "**"
	case ast.KindEqualsEqualsToken:
		return "=="
	case ast.KindEqualsEqualsEqualsToken:
		return "==="
	case ast.KindExclamationEqualsToken:
		return "!="
	case ast.KindExclamationEqualsEqualsToken:
		return "!=="
	case ast.KindLessThanToken:
		return "<"
	case ast.KindLessThanEqualsToken:
		return "<="
	case ast.KindGreaterThanToken:
		return ">"
	case ast.KindGreaterThanEqualsToken:
		return ">="
	case ast.KindLessThanLessThanToken:
		return "<<"
	case ast.KindGreaterThanGreaterThanToken:
		return ">>"
	case ast.KindGreaterThanGreaterThanGreaterThanToken:
		return ">>>"
	case ast.KindAmpersandToken:
		return "&"
	case ast.KindBarToken:
		return "|"
	case ast.KindCaretToken:
		return "^"
	case ast.KindAmpersandAmpersandToken:
		return "&&"
	case ast.KindBarBarToken:
		return "||"
	case ast.KindQuestionQuestionToken:
		return "??"
	case ast.KindInKeyword:
		return "in"
	case ast.KindInstanceOfKeyword:
		return "instanceof"
	case ast.KindExclamationToken:
		return "!"
	case ast.KindTildeToken:
		return "~"
	case ast.KindPlusPlusToken:
		return "++"
	case ast.KindMinusMinusToken:
		return "--"
	case ast.KindPlusEqualsToken:
		return "+="
	case ast.KindMinusEqualsToken:
		return "-="
	case ast.KindAsteriskEqualsToken:
		return "*="
	case ast.KindSlashEqualsToken:
		return "/="
	case ast.KindPercentEqualsToken:
		return "%="
	case ast.KindAsteriskAsteriskEqualsToken:
		return "**="
	case ast.KindLessThanLessThanEqualsToken:
		return "<<="
	case ast.KindGreaterThanGreaterThanEqualsToken:
		return ">>="
	case ast.KindGreaterThanGreaterThanGreaterThanEqualsToken:
		return ">>>="
	case ast.KindAmpersandEqualsToken:
		return "&="
	case ast.KindBarEqualsToken:
		return "|="
	case ast.KindCaretEqualsToken:
		return "^="
	case ast.KindAmpersandAmpersandEqualsToken:
		return "&&="
	case ast.KindBarBarEqualsToken:
		return "||="
	case ast.KindQuestionQuestionEqualsToken:
		return "??="
	}
	return kind.String()
}

func splitRegExp(text string) (string, string) {
	if len(text) < 2 || text[0] != '/' {
		return text, ""
	}
	last := strings.LastIndex(text, "/")
	if last <= 0 {
		return text, ""
	}
	return text[1:last], text[last+1:]
}
