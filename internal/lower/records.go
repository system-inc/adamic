package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"strings"
)

// recordElement recognizes resolved pure mutable string signatures, including aliases.
func (l *lowering) recordElement(t *checker.Type) *checker.Type {
	if t == nil {
		return nil
	}
	t = l.concrete(t)
	if t.Flags()&checker.TypeFlagsObject == 0 || l.checker.IsArrayType(t) || checker.IsTupleType(t) || isClassInstance(t) {
		return nil
	}
	if element := l.finitePartialRecordElement(t); element != nil {
		return element
	}
	infos := l.checker.GetIndexInfosOfType(t)
	for _, info := range infos {
		if declaration := info.Declaration(); declaration != nil {
			file := ast.GetSourceFileOfNode(declaration)
			if load.IsLibrary(file) && strings.Contains(string(file.AsSourceFile().FileName()), ".regexp.") {
				return nil
			}
		}
	}
	if len(infos) != 1 || infos[0].KeyType().Flags() != checker.TypeFlagsString || infos[0].IsReadonly() || len(l.checker.GetPropertiesOfType(t)) != 0 || len(l.checker.GetSignaturesOfType(t, checker.SignatureKindCall)) != 0 || len(l.checker.GetSignaturesOfType(t, checker.SignatureKindConstruct)) != 0 {
		return nil
	}
	return infos[0].ValueType()
}
func (l *lowering) recordLiteralElement(node *ast.Node) *checker.Type {
	if t := l.checker.GetContextualType(node, checker.ContextFlagsNone); t != nil {
		if e := l.recordElement(l.checker.GetNonNullableType(t)); e != nil {
			return e
		}
	}
	return l.recordElement(l.checker.GetTypeAtLocation(node))
}
func (l *lowering) recordSlot(node *ast.Node) (ir.Type, error) {
	t := l.recordElement(l.checker.GetTypeAtLocation(node))
	if t == nil {
		return 0, l.notYet(node, "a record without a pure mutable string index signature")
	}
	of, known := l.kept(t)
	if !known || of == ir.MaybeBoolean || of == ir.Weak {
		return 0, l.notYet(node, "a record value of type "+l.checker.TypeToString(t))
	}
	return of, nil
}
func (l *lowering) recordTarget(node *ast.Node) bool {
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindElementAccessExpression {
		return l.recordElement(l.checker.GetTypeAtLocation(n.AsElementAccessExpression().Expression)) != nil
	}
	if n.Kind == ast.KindPropertyAccessExpression {
		return l.recordElement(l.checker.GetTypeAtLocation(n.AsPropertyAccessExpression().Expression)) != nil
	}
	return false
}
func recordPrototypeName(name string) bool {
	switch name {
	case "constructor", "__defineGetter__", "__defineSetter__", "hasOwnProperty", "__lookupGetter__", "__lookupSetter__", "isPrototypeOf", "propertyIsEnumerable", "toString", "valueOf", "__proto__", "toLocaleString":
		return true
	}
	return false
}
func (l *lowering) recordKey(node *ast.Node, read bool) (ir.Expression, error) {
	n := ast.SkipParentheses(node)
	if read && (n.Kind == ast.KindStringLiteral || n.Kind == ast.KindNoSubstitutionTemplateLiteral) && recordPrototypeName(n.Text()) {
		return nil, &Refused{Where: l.program.Where(node), What: "record key " + n.Text() + " names an Object.prototype member", Fix: "use an own-property test and a supported own-key operation; inherited values are outside the record's value type"}
	}
	key, err := l.expression(node)
	if err != nil {
		return nil, err
	}
	if key.Type() == ir.Number {
		return ir.NumberToString{Value: key}, nil
	}
	if key.Type() != ir.String {
		return nil, l.notYet(node, "a record key other than string or number")
	}
	return key, nil
}
func (l *lowering) recordAccess(node *ast.Node, read bool) (ir.Expression, ir.Expression, ir.Type, error) {
	n := ast.SkipParentheses(node)
	var receiver, keyNode *ast.Node
	var name string
	if n.Kind == ast.KindElementAccessExpression {
		a := n.AsElementAccessExpression()
		if a.QuestionDotToken != nil {
			return nil, nil, 0, l.notYet(node, "optional record indexing")
		}
		receiver, keyNode = a.Expression, a.ArgumentExpression
	} else {
		a := n.AsPropertyAccessExpression()
		if a.QuestionDotToken != nil {
			return nil, nil, 0, l.notYet(node, "optional record property access")
		}
		receiver = a.Expression
		name = a.Name().Text()
	}
	of, err := l.recordSlot(receiver)
	if err != nil {
		return nil, nil, 0, err
	}
	r, err := l.expression(receiver)
	if err != nil {
		return nil, nil, 0, err
	}
	var key ir.Expression
	if keyNode != nil {
		key, err = l.recordKey(keyNode, read)
	} else {
		if read && recordPrototypeName(name) {
			return nil, nil, 0, &Refused{Where: l.program.Where(node), What: "record key " + name + " names an Object.prototype member", Fix: "use own keys"}
		}
		key = ir.StringConstant{Index: l.constant(name)}
	}
	return r, key, of, err
}
func (l *lowering) recordExpression(node *ast.Node) (ir.Expression, bool, error) {
	if l.recordTarget(node) {
		discarded := l.recordReadDiscarded(node)
		guarded := l.recordReadOwnGuarded(node)
		r, k, of, err := l.recordAccess(node, !discarded && !guarded)
		if err != nil {
			return nil, true, err
		}
		var read ir.Expression = ir.RecordCall{Method: "get", Arguments: []ir.Expression{r, k}, Element: of, Returns: ir.Maybe(of), ViewTypeID: int(l.checker.GetTypeAtLocation(node).Id()), ViewWhere: l.program.Where(node)}
		if !discarded && !guarded {
			id, contractErr := l.viewContract(node, l.concrete(l.checker.GetTypeAtLocation(node)))
			if contractErr != nil {
				return nil, true, contractErr
			}
			if _, supported := ir.DictionaryReadKinds(l.result, id); supported || l.primitiveDictionaryCandidate(l.concrete(l.checker.GetTypeAtLocation(node))) {
				checked, err := l.dictionaryRead(node, r, k)
				if err != nil {
					return nil, true, err
				}
				property := checked.(ir.Property)
				property.Of = ir.Maybe(of)
				call := read.(ir.RecordCall)
				call.DictionaryRead = &property
				read = call
			}
		}
		if discarded {
			return l.recordOwnRead(r, k, of), true, nil
		}
		if guarded {
			read = l.recordOwnRead(r, k, of)
		}
		here := l.checker.GetTypeAtLocation(node)
		if !l.includesUndefined(here) && !comparedWithUndefined(node) {
			if narrowed, known := l.representation(here); known && read.Type().IsMaybe() && !narrowed.IsMaybe() {
				read = ir.Unwrap{Value: read}
			}
			if read.Type().IsReference() && read.Type() != ir.Union {
				message := "undefined where the checker narrowed it away: a call since the narrowing put it back"
				if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression && parent.AsPropertyAccessExpression().Expression == node {
					message = "TypeError: Cannot read properties of undefined (reading '" + parent.Name().Text() + "')"
				}
				read = ir.Defined{Value: read, Message: message}
			}
		}
		return read, true, nil
	}
	if node.Kind == ast.KindDeleteExpression && l.recordTarget(node.AsDeleteExpression().Expression) {
		r, k, of, err := l.recordAccess(node.AsDeleteExpression().Expression, false)
		return ir.RecordCall{Method: "delete", Arguments: []ir.Expression{r, k}, Element: of, Returns: ir.Boolean}, true, err
	}
	if value, handled, err := l.recordScalarComparison(node); handled {
		return value, true, err
	}
	if node.Kind != ast.KindBinaryExpression {
		return nil, false, nil
	}
	b := node.AsBinaryExpression()
	if b.OperatorToken.Kind == ast.KindInKeyword && l.recordElement(l.checker.GetTypeAtLocation(b.Right)) != nil {
		k, err := l.recordKey(b.Left, true)
		if err != nil {
			return nil, true, err
		}
		r, err := l.expression(b.Right)
		return ir.RecordCall{Method: "has", Arguments: []ir.Expression{k, r}, Returns: ir.Boolean}, true, err
	}
	if (b.OperatorToken.Kind == ast.KindEqualsToken || b.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken) && l.recordTarget(b.Left) {
		r, k, of, err := l.recordAccess(b.Left, b.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken)
		if err != nil {
			return nil, true, err
		}
		v, err := l.expression(b.Right)
		if err != nil {
			return nil, true, err
		}
		v = fit(v, of)
		if v.Type() != of {
			return nil, true, l.notYet(node, "a record write with a different slot representation")
		}
		if b.OperatorToken.Kind == ast.KindQuestionQuestionEqualsToken {
			if of.IsMaybe() || of == ir.Union {
				return nil, true, l.notYet(node, "record ??= with a tagged optional value")
			}
			return ir.RecordCoalesce{Record: r, Key: k, Value: v, Element: of, Site: l.writeSite(recordReceiver(b.Left))}, true, nil
		}
		return ir.RecordCall{Method: "set", Arguments: []ir.Expression{r, k, v}, Element: of, Returns: of, Site: l.writeSite(recordReceiver(b.Left))}, true, nil
	}
	return nil, false, nil
}
func recordReceiver(node *ast.Node) *ast.Node {
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindElementAccessExpression {
		return n.AsElementAccessExpression().Expression
	}
	return n.AsPropertyAccessExpression().Expression
}
func (l *lowering) recordLiteral(node *ast.Node, t *checker.Type) (ir.Expression, error) {
	of, known := l.kept(t)
	if !known || of == ir.MaybeBoolean || of == ir.Weak {
		return nil, l.notYet(node, "record values with an unsupported slot representation")
	}
	result := ir.RecordLiteral{Element: of, Site: l.writeSite(node)}
	for i, p := range node.AsObjectLiteralExpression().Properties.Nodes {
		if p.Kind == ast.KindSpreadAssignment {
			if i != 0 {
				return nil, &Refused{Where: l.program.Where(p), What: "a spread after the first field", Fix: "spread once, first (adamic/single-spread)"}
			}
			source := p.AsSpreadAssignment().Expression
			element := l.recordElement(l.checker.GetTypeAtLocation(source))
			if element != nil {
				if !l.sameRecordStorage(element, t, map[[2]*checker.Type]bool{}) || !l.sameKeeping(element, t, map[[2]*checker.Type]bool{}) {
					return nil, l.notYet(p, "record spread values seen through different storage")
				}
				if widened := l.widened(element, t, map[[2]*checker.Type]bool{}); widened != nil {
					return nil, &Refused{Where: l.program.Where(p), What: "record spread shares values with an invariant-mutable slot that the target can widen", Fix: "keep the value's mutable fields invariant or copy the values too (adamic/invariant-mutable)"}
				}
			}
			e, err := l.recordSlot(source)
			if err != nil {
				return nil, err
			}
			if e != of {
				return nil, l.notYet(p, "record spread with a different value representation")
			}
			result.Spread, err = l.expression(source)
			if err != nil {
				return nil, err
			}
			continue
		}
		if p.Kind != ast.KindPropertyAssignment && p.Kind != ast.KindShorthandPropertyAssignment {
			return nil, l.notYet(p, "a record literal member other than a data entry")
		}
		name := p.Name()
		var key ir.Expression
		var err error
		if name.Kind == ast.KindComputedPropertyName {
			key, err = l.recordKey(name.AsComputedPropertyName().Expression, false)
		} else {
			if p.Kind == ast.KindPropertyAssignment && name.Text() == "__proto__" {
				return nil, l.notYet(p, "a prototype-setting record literal")
			}
			if name.Kind == ast.KindNumericLiteral {
				key, err = l.recordKey(name, false)
			} else {
				key = ir.StringConstant{Index: l.constant(name.Text())}
			}
		}
		if err != nil {
			return nil, err
		}
		var v ir.Expression
		if p.Kind == ast.KindPropertyAssignment {
			v, err = l.expression(p.AsPropertyAssignment().Initializer)
		} else {
			v, err = l.shorthand(p)
		}
		if err != nil {
			return nil, err
		}
		v = fit(v, of)
		if v.Type() != of {
			return nil, l.notYet(p, "a record literal value with a different representation")
		}
		result.Entries = append(result.Entries, ir.RecordEntry{Key: key, Value: v})
	}
	return result, nil
}
func (l *lowering) recordObjectCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	args := node.AsCallExpression().Arguments.Nodes
	if len(args) == 0 || l.recordElement(l.checker.GetTypeAtLocation(args[0])) == nil {
		return nil, false, nil
	}
	if name != "keys" && name != "values" && name != "entries" && name != "hasOwn" {
		return nil, true, l.notYet(node, "Object."+name+" on a record")
	}
	count := 1
	if name == "hasOwn" {
		count = 2
	}
	if len(args) != count {
		return nil, true, l.notYet(node, "record Object operation argument count")
	}
	of, err := l.recordSlot(args[0])
	if err != nil {
		return nil, true, err
	}
	r, err := l.expression(args[0])
	if err != nil {
		return nil, true, err
	}
	result := ir.RecordCall{Method: name, Arguments: []ir.Expression{r}, Element: of, Returns: ir.Array, ViewTypeID: int(l.recordElement(l.checker.GetTypeAtLocation(args[0])).Id()), ViewWhere: l.program.Where(node)}
	if name == "hasOwn" {
		k, err := l.recordKey(args[1], false)
		if err != nil {
			return nil, true, err
		}
		result.Arguments = append(result.Arguments, k)
		result.Returns = ir.Boolean
	}
	return result, true, nil
}

// Aliases cannot change the native storage kind, even when tsc accepts the structural view.
func (l *lowering) recordStorageView(node *ast.Node) error {
	if l.dictionaryReferenceConversion(node) {
		return nil
	}
	var source *ast.Node
	var target *checker.Type
	if node.Kind == ast.KindVariableDeclaration {
		source = node.AsVariableDeclaration().Initializer
		if source != nil {
			target = l.checker.GetTypeAtLocation(node.Name())
		}
	} else if node.Kind == ast.KindAsExpression {
		source = node.AsAsExpression().Expression
		target = l.checker.GetTypeAtLocation(node)
	} else if l.isExpression(node) && viewSite(node) {
		source = node
		target = l.checker.GetContextualType(node, checker.ContextFlagsNone)
	}
	if source == nil || target == nil || ast.SkipParentheses(source).Kind == ast.KindObjectLiteralExpression {
		return nil
	}
	if parent := source.Parent; parent != nil && parent.Kind == ast.KindCallExpression {
		if l.detachedOwnCallReceiver(parent) != nil {
			return nil
		}
		callee := ast.SkipParentheses(parent.AsCallExpression().Expression)
		if l.isLibraryGlobal(callee, "String") {
			return nil
		}
		if callee.Kind == ast.KindPropertyAccessExpression && (l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") || l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "JSON")) {
			return nil
		}
	}
	own := l.checker.GetTypeAtLocation(source)
	if node.Kind == ast.KindAsExpression && own.Flags()&checker.TypeFlagsObject != 0 && target.Flags()&checker.TypeFlagsObject != 0 && l.checker.IsTypeAssignableTo(target, own) {
		return nil
	}
	if !l.sameRecordStorage(own, target, map[[2]*checker.Type]bool{}) {
		return l.notYet(node, "a "+l.checker.TypeToString(own)+" seen as "+l.checker.TypeToString(target)+" (fixed objects and records have different storage; copy explicitly)")
	}
	return nil
}

// Structural views must preserve dictionary storage at every depth, including callbacks.
func (l *lowering) sameRecordStorage(from, to *checker.Type, visited map[[2]*checker.Type]bool) bool {
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return true
	}
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		visited[[2]*checker.Type{from, to}] = true
		for _, member := range l.definedMembers(from) {
			if member.Flags()&checker.TypeFlagsNull == 0 && !l.sameRecordStorage(member, to, visited) {
				return false
			}
		}
		return true
	}
	if to.Flags()&checker.TypeFlagsUnion != 0 {
		visited[[2]*checker.Type{from, to}] = true
		for _, member := range l.definedMembers(to) {
			if l.checker.IsTypeAssignableTo(from, member) && !l.sameRecordStorage(from, member, visited) {
				return false
			}
		}
		return true
	}
	from, to = l.present(from), l.present(to)
	if from == nil || to == nil || from == to || visited[[2]*checker.Type{from, to}] {
		return true
	}
	visited[[2]*checker.Type{from, to}] = true
	inside, viewed := l.recordElement(from), l.recordElement(to)
	if inside != nil || viewed != nil {
		return inside != nil && viewed != nil && l.sameRecordStorage(inside, viewed, visited)
	}
	a, b := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall), l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	if len(a) > 0 && len(b) > 0 {
		ap, bp := a[0].Parameters(), b[0].Parameters()
		for i := 0; i < len(ap) && i < len(bp); i++ {
			if !l.sameRecordStorage(l.checker.GetTypeOfSymbol(ap[i]), l.checker.GetTypeOfSymbol(bp[i]), visited) {
				return false
			}
		}
		return l.sameRecordStorage(l.checker.GetReturnTypeOfSignature(a[0]), l.checker.GetReturnTypeOfSignature(b[0]), visited)
	}
	if l.checker.IsArrayType(from) || checker.IsTupleType(from) || l.isLibraryType(from, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		a, b := l.typeArguments(from), l.typeArguments(to)
		for i := 0; i < len(a) && i < len(b); i++ {
			if !l.sameRecordStorage(a[i], b[i], visited) {
				return false
			}
		}
		return true
	}
	for _, property := range l.checker.GetPropertiesOfType(to) {
		if own := l.checker.GetPropertyOfType(from, property.Name); own != nil {
			if !l.sameRecordStorage(l.checker.GetTypeOfSymbol(own), l.checker.GetTypeOfSymbol(property), visited) {
				return false
			}
		}
	}
	return true
}

// Catch literal reads before method/builtin dispatch can bypass ordinary value lowering.
func (l *lowering) recordLiteralRead(node *ast.Node) error {
	if !l.recordTarget(node) || l.recordReadDiscarded(node) || l.recordReadOwnGuarded(node) {
		return nil
	}
	at := node
	for at.Parent != nil && at.Parent.Kind == ast.KindParenthesizedExpression {
		at = at.Parent
	}
	if parent := at.Parent; parent != nil {
		if parent.Kind == ast.KindDeleteExpression {
			return nil
		}
		if parent.Kind == ast.KindBinaryExpression {
			b := parent.AsBinaryExpression()
			if b.Left == at && b.OperatorToken.Kind == ast.KindEqualsToken {
				return nil
			}
		}
	}
	if parent := at.Parent; parent != nil && parent.Kind == ast.KindBinaryExpression {
		// The scalar comparison owns the read and proves its inherited value unobservable.
		b := parent.AsBinaryExpression()
		other := b.Right
		if b.Right == at {
			other = b.Left
		}
		t := l.checker.GetTypeAtLocation(other)
		if (b.OperatorToken.Kind == ast.KindEqualsEqualsEqualsToken || b.OperatorToken.Kind == ast.KindExclamationEqualsEqualsToken) && t.Flags()&(checker.TypeFlagsStringLike|checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike) != 0 && !l.includesUndefined(t) {
			return nil
		}
	}
	var name string
	if node.Kind == ast.KindPropertyAccessExpression {
		name = node.Name().Text()
	} else {
		key := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral {
			name = key.Text()
		}
	}
	if recordPrototypeName(name) {
		return &Refused{Where: l.program.Where(node), What: "record key " + name + " names an Object.prototype member", Fix: "use own keys"}
	}
	return nil
}
