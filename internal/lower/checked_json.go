package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (l *lowering) jsonContract(node *ast.Node, target *checker.Type, depth int) (*ir.JSONContract, error) {
	target = l.concrete(target)
	name := l.checker.TypeToString(target)
	fail := func() (*ir.JSONContract, error) {
		return nil, l.notYet(node, "checked JSON contract "+name+" (requires a finite data structure without alias writes)")
	}
	if depth > 32 {
		return fail()
	}
	schema := &ir.JSONContract{Name: name}
	flags := target.Flags()
	switch {
	case flags == checker.TypeFlagsNumber:
		schema.Kind = "number"
	case flags == checker.TypeFlagsString:
		schema.Kind = "string"
	case flags&checker.TypeFlagsBoolean != 0 && flags&checker.TypeFlagsBooleanLiteral == 0:
		schema.Kind = "boolean"
	case flags&checker.TypeFlagsStringLiteral != 0:
		value, _, known := l.literalConstant(target)
		if !known {
			return fail()
		}
		literal, ok := value.(ir.StringConstant)
		if !ok {
			return fail()
		}
		schema.Kind = "string_literal"
		schema.LiteralText = l.result.Strings[literal.Index]
	case flags&checker.TypeFlagsNumberLiteral != 0:
		value, _, known := l.literalConstant(target)
		if !known {
			return fail()
		}
		literal, ok := value.(ir.NumberConstant)
		if !ok {
			return fail()
		}
		schema.Kind = "number_literal"
		schema.LiteralNumber = literal.Value
	case flags&checker.TypeFlagsBooleanLiteral != 0:
		value, _, known := l.literalConstant(target)
		if !known {
			return fail()
		}
		literal, ok := value.(ir.BooleanConstant)
		if !ok {
			return fail()
		}
		schema.Kind = "boolean_literal"
		schema.LiteralBoolean = literal.Value
	case flags == checker.TypeFlagsUndefined:
		schema.Kind = "undefined"
	case flags == checker.TypeFlagsNull:
		schema.Kind = "null"
	case flags&checker.TypeFlagsUnion != 0:
		if l.includesNull(target) {
			if to, known := l.representation(target); !known || to != ir.Union {
				return fail()
			}
		}
		schema.Kind = "union"
		for _, part := range target.Types() {
			child, err := l.jsonContract(node, part, depth+1)
			if err != nil {
				return nil, err
			}
			schema.Alternatives = append(schema.Alternatives, child)
		}
	case flags&checker.TypeFlagsObject != 0:
		if len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 || isClassInstance(target) {
			return fail()
		}
		if l.checker.IsArrayType(target) || l.isLibraryType(target, "ReadonlyArray") {
			args := l.typeArguments(target)
			if len(args) != 1 {
				return fail()
			}
			child, err := l.jsonContract(node, args[0], depth+1)
			if err != nil {
				return nil, err
			}
			schema.Kind = "array"
			schema.Element = child
		} else {
			if l.checker.IsArrayType(target) || checker.IsTupleType(target) || l.isLibraryType(target, "Map", "ReadonlyMap", "Set", "ReadonlySet", "Date", "RegExp") {
				return fail()
			}
			schema.Kind = "object"
			for _, field := range l.checker.GetPropertiesOfType(target) {
				if field.Name == "__proto__" || field.Name == "constructor" {
					return fail()
				}
				if accessorSymbol(field) || l.dynamicReadHazard(field.Name) || strings.ContainsRune(field.Name, 0) {
					return fail()
				}
				child, err := l.jsonContract(node, l.checker.GetTypeOfSymbol(field), depth+1)
				if err != nil {
					return nil, err
				}
				schema.Fields = append(schema.Fields, ir.JSONContractField{Name: field.Name, Optional: field.Flags&ast.SymbolFlagsOptional != 0, Contract: child})
			}
			if len(schema.Fields) == 0 {
				return fail()
			}
		}
	default:
		return fail()
	}
	return schema, nil
}

func (l *lowering) checkedJSONType(node *ast.Node, value ir.Expression, target *checker.Type) (ir.Expression, error) {
	schema, err := l.jsonContract(node, target, 0)
	if err != nil {
		return nil, err
	}
	to, known := l.representation(target)
	if !known {
		return nil, l.notYet(node, "checked JSON representation "+schema.Name)
	}
	// Mixed/null unions retain tags. Optional reference ABIs currently conflate null
	// and undefined, so they are not exposed as a structural pointer contract.
	if to != ir.Union && target.Flags()&checker.TypeFlagsUnion != 0 {
		return nil, l.notYet(node, "checked JSON optional reference representation")
	}
	// Alias guards must have been prepared before any module body is lowered.
	if !l.jsonContractPrepared(schema) {
		return nil, l.notYet(node, "checked JSON contract without a prepared dynamic binding")
	}
	l.noteJSONContract(schema)
	return ir.CheckedJSON{Value: fit(value, ir.Union), Contract: schema, Path: sourceExpression(node), Of: to}, nil
}
func (l *lowering) noteJSONContract(c *ir.JSONContract) {
	if l.result.JSONCheckedFields == nil {
		l.result.JSONCheckedFields = map[string]bool{}
	}
	if c.Kind == "array" {
		l.result.JSONCheckedArrays = true
		l.noteJSONContract(c.Element)
	}
	for _, field := range c.Fields {
		l.result.JSONCheckedFields[field.Name] = true
		l.noteJSONContract(field.Contract)
	}
	for _, part := range c.Alternatives {
		l.noteJSONContract(part)
	}
}
func (l *lowering) prepareJSONChecks() {
	if l.result.JSONChecksPrepared {
		return
	}
	l.result.JSONChecksPrepared = true
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier && node.Parent != nil && node.Parent.Name() != node && l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsAny != 0 {
			target := l.checker.GetContextualType(node, checker.ContextFlagsNone)
			if target != nil {
				if c, err := l.jsonContract(node, target, 0); err == nil {
					l.noteJSONContract(c)
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
}
func (l *lowering) jsonViewHazard(node *ast.Node) string {
	if len(l.result.JSONCheckedFields) == 0 && !l.result.JSONCheckedArrays {
		return ""
	}
	if node.Kind == ast.KindPrefixUnaryExpression || node.Kind == ast.KindPostfixUnaryExpression {
		var operand *ast.Node
		if node.Kind == ast.KindPrefixUnaryExpression {
			p := node.AsPrefixUnaryExpression()
			if p.Operator == ast.KindPlusPlusToken || p.Operator == ast.KindMinusMinusToken {
				operand = p.Operand
			}
		} else {
			operand = node.AsPostfixUnaryExpression().Operand
		}
		if operand != nil {
			operand = ast.SkipParentheses(operand)
			if operand.Kind == ast.KindElementAccessExpression || (operand.Kind == ast.KindPropertyAccessExpression && l.result.JSONCheckedFields[operand.Name().Text()]) {
				return "an update through a checked JSON view alias"
			}
		}
	}
	if node.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee.Kind == ast.KindPropertyAccessExpression {
			base := callee.AsPropertyAccessExpression().Expression
			if l.isLibraryGlobal(base, "Object") || l.isLibraryGlobal(base, "Reflect") {
				return "reflection in a checked JSON view program"
			}
		}
	}
	if node.Kind == ast.KindBinaryExpression && ast.IsAssignmentOperator(node.AsBinaryExpression().OperatorToken.Kind) {
		left := ast.SkipParentheses(node.AsBinaryExpression().Left)
		if left.Kind == ast.KindElementAccessExpression || (left.Kind == ast.KindPropertyAccessExpression && l.result.JSONCheckedFields[left.Name().Text()]) {
			return "a write through a checked JSON view alias"
		}
	}
	if l.result.JSONCheckedArrays {
		if node.Kind == ast.KindForOfStatement || node.Kind == ast.KindSpreadElement {
			return "iteration of checked JSON arrays (requires dynamic element reads)"
		}
		if node.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression {
				receiver := callee.AsPropertyAccessExpression().Expression
				if l.checker.IsArrayType(l.checker.GetTypeAtLocation(receiver)) || l.isLibraryType(l.checker.GetTypeAtLocation(receiver), "ReadonlyArray") {
					return "a method on a checked JSON array (requires dynamic element reads)"
				}
			}
		}
	}
	return ""
}

// Mixed scalar recovery results need distinct tags for both null and undefined.
// Existing nullable reference ABIs remain unchanged.
func (l *lowering) jsonRecoveryUnion(t *checker.Type) bool {
	if t.Flags()&checker.TypeFlagsUnion == 0 || !l.includesNull(t) {
		return false
	}
	kinds := map[ir.Type]bool{}
	for _, part := range t.Types() {
		flags := part.Flags()
		if flags&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
			continue
		}
		switch {
		case flags&checker.TypeFlagsNumberLike != 0:
			kinds[ir.Number] = true
		case flags&checker.TypeFlagsStringLike != 0:
			kinds[ir.String] = true
		case flags&checker.TypeFlagsBooleanLike != 0:
			kinds[ir.Boolean] = true
		default:
			return false
		}
	}
	return len(kinds) > 1
}

func (l *lowering) jsonContractPrepared(c *ir.JSONContract) bool {
	if c.Kind == "array" && (!l.result.JSONCheckedArrays || !l.jsonContractPrepared(c.Element)) {
		return false
	}
	for _, field := range c.Fields {
		if !l.result.JSONCheckedFields[field.Name] || !l.jsonContractPrepared(field.Contract) {
			return false
		}
	}
	for _, part := range c.Alternatives {
		if !l.jsonContractPrepared(part) {
			return false
		}
	}
	return true
}
