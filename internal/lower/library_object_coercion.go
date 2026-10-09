// Copyright 2006-2011 the V8 project authors. All rights reserved.
// Use of this source code is governed by a BSD-style license in THIRD_PARTY_NOTICES.md.
package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The method order, callable test and first-primitive rule follow V8's
// JSReceiver::OrdinaryToPrimitive in src/objects/js-objects.cc, Node 24.19.0.
// This adds the number hint; String keeps its existing string-hint lowering.
// Complete plain shapes exclude Symbol.toPrimitive and getters. Only zero-argument arrows are
// admitted: their lexical this needs no new implicit receiver calling convention. Mixed return
// types and the final TypeError remain refused instead of selecting an approximate answer.
func (l *lowering) libraryObjectCoercionCall(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if !l.isLibraryGlobal(callee, "Number") {
		return nil, false, nil
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 || hasSpread(node) || !l.exactObject(arguments[0], 0) {
		return nil, false, nil
	}
	receiver := arguments[0]
	literal := l.objectCoercionLiteral(receiver, 0)
	if literal == nil {
		return nil, false, nil
	}
	fields := map[string]*ast.Node{}
	for _, property := range literal.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind == ast.KindPropertyAssignment {
			fields[property.Name().Text()] = property.AsPropertyAssignment().Initializer
		}
		if property.Kind == ast.KindShorthandPropertyAssignment {
			fields[property.Name().Text()] = property.Name()
		}
	}
	order := []string{"valueOf", "toString"}
	returns := ir.Number
	type step struct {
		name      string
		result    ir.Type
		primitive bool
		fallback  bool
	}
	steps := []step{}
	complete := false
	for _, name := range order {
		initializer := fields[name]
		if initializer == nil {
			if name == "toString" {
				steps = append(steps, step{fallback: true, primitive: true, result: ir.String})
				complete = true
				break
			}
			continue // Object.prototype.valueOf returns this object, so try the next method.
		}
		member := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(receiver), name)
		signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeOfSymbol(member), checker.SignatureKindCall)
		if len(signatures) == 0 {
			continue
		}
		if l.functionPrototype(initializer, 0) != 0 || len(signatures) != 1 || len(signatures[0].Parameters()) != 0 {
			return nil, true, l.notYet(initializer, "ToPrimitive callback requiring an implicit this receiver or parameters")
		}
		result := l.checker.GetReturnTypeOfSignature(signatures[0])
		of, known := l.representation(result)
		primitive := result.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike) != 0
		object := result.Flags()&checker.TypeFlagsObject != 0
		if object && (l.checker.IsTypeAssignableTo(l.checker.GetNumberType(), result) || l.checker.IsTypeAssignableTo(l.checker.GetStringType(), result) || l.checker.IsTypeAssignableTo(l.checker.GetBooleanType(), result)) {
			return nil, true, l.notYet(initializer, "ToPrimitive callback result whose object view can hide a primitive")
		}
		if !known || (!primitive && !object) {
			return nil, true, l.notYet(initializer, "ToPrimitive callback with a mixed or unrepresented result")
		}
		steps = append(steps, step{name: name, result: of, primitive: primitive})
		if primitive {
			complete = true
			break
		}
	}
	if !complete {
		return nil, true, l.notYet(receiver, "ToPrimitive without a proven primitive result (the final TypeError is not lowered)")
	}
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	function := len(l.result.Functions)
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "receiver", Type: ir.Object, Function: function})
	read := ir.Read{Local: held, Of: ir.Object}
	body := []ir.Statement{}
	for _, step := range steps {
		var result ir.Expression = ir.StringConstant{Index: l.constant("[object Object]")}
		if !step.fallback {
			result = ir.CallClosure{Closure: ir.Property{Object: read, Name: step.name, Of: ir.Closure}, Returns: step.result}
		}
		if !step.primitive {
			body = append(body, ir.Evaluate{Value: result})
			continue
		}
		result = ir.NumberCall{Function: "convert", Arguments: []ir.Expression{result}}
		body = append(body, ir.Return{Value: result})
	}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "object_primitive", Parameters: []int{held}, Returns: returns, MayThrow: true, Body: body})
	return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: returns}, true, nil
}

func (l *lowering) objectCoercionLiteral(node *ast.Node, depth int) *ast.Node {
	if depth > 16 {
		return nil
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindObjectLiteralExpression {
		return node
	}
	if !ast.IsIdentifier(node) {
		return nil
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindVariableDeclaration {
		return nil
	}
	initializer := symbol.Declarations[0].AsVariableDeclaration().Initializer
	if initializer == nil {
		return nil
	}
	return l.objectCoercionLiteral(initializer, depth+1)
}
