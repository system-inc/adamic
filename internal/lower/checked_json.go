package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (l *lowering) jsonContract(node *ast.Node, target *checker.Type, depth int) (*ir.JSONContract, error) {
	return l.jsonContractGraph(node, target, depth, map[*checker.Type]*ir.JSONContract{})
}

func (l *lowering) jsonContractGraph(node *ast.Node, target *checker.Type, depth int, seen map[*checker.Type]*ir.JSONContract) (*ir.JSONContract, error) {
	target = l.concrete(target)
	if schema := seen[target]; schema != nil {
		return &ir.JSONContract{Kind: "ref", Name: schema.Name, Reference: schema.ID}, nil
	}
	name := l.checker.TypeToString(target)
	fail := func() (*ir.JSONContract, error) {
		return nil, l.notYet(node, "checked JSON contract "+name+" (requires a finite data structure without alias writes)")
	}
	if depth > 32 {
		return fail()
	}
	schema := &ir.JSONContract{Name: name, ID: len(seen)}
	seen[target] = schema
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
		schema.Kind = "union"
		for _, part := range target.Types() {
			child, err := l.jsonContractGraph(node, part, depth+1, seen)
			if err != nil {
				return nil, err
			}
			schema.Alternatives = append(schema.Alternatives, child)
		}
	case flags&checker.TypeFlagsObject != 0:
		if len(l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)) != 0 {
			return nil, l.notYet(node, "checked JSON callable contract "+name+" (requires verified parameter and result behavior)")
		}
		if isClassInstance(target) {
			return fail()
		}
		if l.checker.IsArrayType(target) || l.isLibraryType(target, "ReadonlyArray") {
			args := l.typeArguments(target)
			if len(args) != 1 {
				return fail()
			}
			child, err := l.jsonContractGraph(node, args[0], depth+1, seen)
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
			indices := l.checker.GetIndexInfosOfType(target)
			if len(indices) > 1 {
				return fail()
			}
			for _, index := range indices {
				if index.KeyType().Flags() != checker.TypeFlagsString {
					return fail()
				}
				if index.ValueType().Flags()&checker.TypeFlagsAny != 0 {
					return nil, l.notYet(node, "checked JSON dictionary "+name+"; awaits compiler/records-maplike: any-valued own-key storage")
				}
				child, err := l.jsonContractGraph(node, index.ValueType(), depth+1, seen)
				if err != nil {
					return nil, err
				}
				schema.Element = child
			}
			for _, field := range l.checker.GetPropertiesOfType(target) {
				if l.jsonSymbolField(field) {
					return nil, l.notYet(node, "a symbol field in a checked JSON contract")
				}
				if field.Name == "__proto__" || field.Name == "constructor" {
					return fail()
				}
				if accessorSymbol(field) || l.propertyReadHazard(field.Name, true) || strings.ContainsRune(field.Name, 0) {
					return fail()
				}
				child, err := l.jsonContractGraph(node, l.checker.GetTypeOfSymbol(field), depth+1, seen)
				if err != nil {
					return nil, err
				}
				schema.Fields = append(schema.Fields, ir.JSONContractField{Name: field.Name, Optional: field.Flags&ast.SymbolFlagsOptional != 0, Contract: child})
			}
			if len(schema.Fields) == 0 && schema.Element == nil {
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
	// Prepared nullable references retain both tags. Any remaining optional ABI
	// that conflates null and undefined is not exposed as a checked contract.
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
func (l *lowering) noteJSONContract(root *ir.JSONContract) {
	if l.result.JSONCheckedFields == nil {
		l.result.JSONCheckedFields = map[string]bool{}
	}
	for _, c := range ir.JSONContractGraph(root) {
		if c.Kind == "array" {
			l.result.JSONCheckedArrays = true
		}
		if c.Kind == "object" && c.Element != nil {
			l.result.JSONCheckedDictionaries = true
		}
		if c.Kind == "null" {
			l.result.JSONTaggedNull = true
		}
		if c.Kind == "union" {
			for _, a := range c.Alternatives {
				if a.Kind == "null" {
					l.result.JSONTaggedNull = true
				}
			}
		}
		for _, field := range c.Fields {
			l.result.JSONCheckedFields[field.Name] = true
		}
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
		// A validated any parameter can return a nullable reference without an
		// unchecked use. Prepare its tagged ABI without inserting a boundary check.
		if node.Kind == ast.KindFunctionDeclaration {
			signature := l.checker.GetSignatureFromDeclaration(node)
			if signature != nil {
				target := l.checker.GetReturnTypeOfSignature(signature)
				for _, parameter := range signature.Parameters() {
					if l.checker.GetTypeOfSymbol(parameter).Flags()&checker.TypeFlagsAny != 0 && l.includesNull(target) {
						if _, err := l.jsonContract(node, target, 0); err == nil {
							l.result.JSONTaggedNull = true
						}
					}
				}
			}
		}
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
	if len(l.result.JSONCheckedFields) == 0 && !l.result.JSONCheckedArrays && !l.result.JSONCheckedDictionaries {
		return ""
	}
	if l.result.JSONCheckedDictionaries && node.Kind == ast.KindPropertyAssignment && node.Name() != nil {
		if node.Name().Kind == ast.KindComputedPropertyName {
			return "a computed key in a checked JSON dictionary (own-key metadata)"
		}
		name := node.Name().Text()
		if strings.ContainsRune(name, 0) {
			return "a NUL key in a checked JSON dictionary (native own-key metadata)"
		}
		if name == "__proto__" {
			return "a prototype setter in a checked JSON dictionary"
		}
	}
	if node.Kind == ast.KindDeleteExpression && l.result.JSONCheckedDictionaries {
		return "delete through a checked JSON dictionary alias"
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
			if operand.Kind == ast.KindElementAccessExpression || (operand.Kind == ast.KindPropertyAccessExpression && (l.result.JSONCheckedDictionaries || l.result.JSONCheckedFields[operand.Name().Text()])) {
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
		if left.Kind == ast.KindElementAccessExpression || (left.Kind == ast.KindPropertyAccessExpression && (l.result.JSONCheckedDictionaries || l.result.JSONCheckedFields[left.Name().Text()])) {
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
// Prepared reifiable nullable references also use the tagged representation.
func (l *lowering) jsonRecoveryUnion(t *checker.Type) bool {
	if t.Flags()&checker.TypeFlagsUnion == 0 || !l.includesNull(t) {
		return false
	}
	if l.result.JSONTaggedNull {
		references := true
		for _, part := range t.Types() {
			if part.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
				continue
			}
			if part.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsObject) == 0 {
				references = false
			}
		}
		if references {
			return true
		}
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

func (l *lowering) jsonContractPrepared(root *ir.JSONContract) bool {
	for _, c := range ir.JSONContractGraph(root) {
		if c.Kind == "array" && !l.result.JSONCheckedArrays {
			return false
		}
		if c.Kind == "object" && c.Element != nil && !l.result.JSONCheckedDictionaries {
			return false
		}
		for _, field := range c.Fields {
			if !l.result.JSONCheckedFields[field.Name] {
				return false
			}
		}
	}
	return true
}

// Index declarations remain erased contracts. This admits only a prepared,
// reifiable string dictionary; it does not invent MapLike<any> storage.
func (l *lowering) checkedJSONIndexSignature(node *ast.Node) bool {
	if !l.result.JSONCheckedDictionaries || node.Parent == nil {
		return false
	}
	c, err := l.jsonContract(node, l.checker.GetTypeAtLocation(node.Parent), 0)
	return err == nil && c.Kind == "object" && c.Element != nil
}

func (l *lowering) jsonScalarRecovery(t *checker.Type) bool {
	if !l.result.JSONTaggedNull || t.Flags()&checker.TypeFlagsUnion == 0 {
		return false
	}
	for _, part := range t.Types() {
		if part.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsStringLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 {
			return false
		}
	}
	return true
}

func (l *lowering) jsonSymbolField(field *ast.Symbol) bool {
	// The checker escapes well-known and synthetic symbol names with this prefix.
	if strings.HasPrefix(field.Name, "\xfe@") {
		return true
	}
	for _, declaration := range field.Declarations {
		name := declaration.Name()
		if name != nil && name.Kind == ast.KindComputedPropertyName && l.checker.GetTypeAtLocation(name.AsComputedPropertyName().Expression).Flags()&checker.TypeFlagsESSymbolLike != 0 {
			return true
		}
	}
	return false
}
