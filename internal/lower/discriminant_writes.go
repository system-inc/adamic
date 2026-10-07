package lower

import (
	"sort"
	"strconv"
	"strings"
	_ "unsafe"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
)

// Use the exact property test used by checker narrowing, through the pinned checker as in
// instantiate.go. A field name alone, or a literal field outside a union, is not a discriminant.
//
//go:linkname discriminantProperty github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).isDiscriminantProperty
func discriminantProperty(c *checker.Checker, union *checker.Type, name string) bool

//go:linkname discriminantPropertyName github.com/microsoft/TypeScript/tsc/internal/checker.(*Checker).getAccessedPropertyName
func discriminantPropertyName(c *checker.Checker, access *ast.Node) (string, bool)

type discriminantField struct {
	member *checker.Type
	name   string
}

func (l *lowering) discriminantFields(modules []*ast.SourceFile) []discriminantField {
	var fields []discriminantField
	seen := map[*checker.Type]bool{}
	var note func(*checker.Type)
	note = func(proven *checker.Type) {
		if proven == nil || seen[proven] {
			return
		}
		seen[proven] = true
		if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
			note(l.checker.GetBaseConstraintOfType(proven))
		}
		for _, signature := range l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall) {
			note(l.checker.GetReturnTypeOfSignature(signature))
		}

		if proven.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range proven.Types() {
				for _, property := range l.checker.GetPropertiesOfType(member) {
					if discriminantProperty(l.checker, proven, property.Name) {
						fields = append(fields, discriminantField{member, property.Name})
					}
				}
				note(member)
			}
		}
	}
	for _, module := range modules {
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if ast.IsExpression(node) || ast.IsTypeNode(node) || ast.IsFunctionLike(node) || node.Kind == ast.KindVariableDeclaration || node.Kind == ast.KindParameter || node.Kind == ast.KindPropertyDeclaration || node.Kind == ast.KindPropertySignature {
				note(l.checker.GetTypeAtLocation(node))
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	return fields
}

// A union view denotes all its members. A broader interface can hide any union member
// assignable to it. A member view retains its own declared property domain, including a
// domain wider than one literal. Narrowing and assignability come from the same checker.
func (l *lowering) discriminantMembers(fields []discriminantField, holder *checker.Type, name string) []*checker.Type {
	if holder == nil {
		return nil
	}
	if holder.Flags()&checker.TypeFlagsTypeParameter != 0 {
		holder = l.checker.GetBaseConstraintOfType(holder)
	}
	if holder == nil {
		return nil
	}
	if holder.Flags()&checker.TypeFlagsUnion != 0 && discriminantProperty(l.checker, holder, name) {
		return holder.Types()
	}
	var members []*checker.Type
	seen := map[*checker.Type]bool{}
	for _, field := range fields {
		if field.name == name && l.checker.IsTypeAssignableTo(field.member, holder) && !seen[field.member] {
			members = append(members, field.member)
			seen[field.member] = true
		}
	}
	if len(members) > 0 && !seen[holder] {
		members = append(members, holder)
	}
	return members
}

func (l *lowering) unsafeDiscriminantWrite(fields []discriminantField, holder, value *checker.Type, name string) bool {
	for _, member := range l.discriminantMembers(fields, holder, name) {
		declared := l.checker.GetTypeOfPropertyOfType(member, name)
		if value == nil || declared == nil || !l.checker.IsTypeAssignableTo(value, declared) {
			return true
		}
	}
	return false
}

func (l *lowering) discriminantWriteRefusal(node *ast.Node, name string, holder, value *checker.Type, fields []discriminantField) error {
	variant := "an unproven variant"
	destinations := map[string]bool{}
	for _, member := range l.discriminantMembers(fields, holder, name) {
		declared := l.checker.GetTypeOfPropertyOfType(member, name)
		if declared != nil && value != nil && l.checker.IsTypeAssignableTo(declared, value) {
			destinations[l.checker.TypeToString(declared)] = true
		}
	}
	var names []string
	for name := range destinations {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) > 0 {
		variant = strings.Join(names, " or ")
	} else if value != nil {
		variant = l.checker.TypeToString(value)
	}

	return &Refused{Where: l.program.Where(node), What: "a write to discriminant field '" + name + "' that could move the object to variant " + variant, Fix: "changing variant means building a new object"}
}

// Obtain the value written, rather than the property's contextual type. Destructuring
// follows the source property's domain down to each target; updates use their result type.
func (l *lowering) discriminantWrittenType(target *ast.Node) *checker.Type {
	var path []string
	node := target
	for node.Parent != nil {
		parent := node.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				return nil
			}
			var value *checker.Type
			if binary.OperatorToken.Kind == ast.KindEqualsToken {
				value = l.checker.GetTypeAtLocation(binary.Right)
			} else {
				value = l.checker.GetTypeAtLocation(parent)
			}
			for index := len(path) - 1; index >= 0 && value != nil; index-- {
				property := l.checker.GetTypeOfPropertyOfType(value, path[index])
				if property == nil {
					property = l.checker.GetIndexTypeOfType(value, l.checker.GetNumberType())
				}
				value = property
			}
			return value
		case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
			return l.checker.GetTypeAtLocation(parent)
		case ast.KindPropertyAssignment:
			path = append(path, parent.Name().Text())
		case ast.KindObjectLiteralExpression:
		case ast.KindArrayLiteralExpression:
			for index, element := range parent.AsArrayLiteralExpression().Elements.Nodes {
				if element == node {
					path = append(path, strconv.Itoa(index))
					break
				}
			}
		default:
			return nil
		}
		node = parent
	}
	return nil
}

// Refuse writes through every view before lowering, including assignment-pattern leaves and
// compound updates. Constructor writes are provisional until the existing fresh-holder proof runs.
func (l *lowering) refuseDiscriminantWrites(module *ast.SourceFile) error {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	fields := l.discriminantFields(modules)
	var found error
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if found != nil {
			return true
		}
		receiver, name := l.discriminantTarget(node)
		if receiver != nil && ast.IsAssignmentTarget(node) {
			value := l.discriminantWrittenType(node)
			if l.unsafeDiscriminantWrite(fields, l.checker.GetTypeAtLocation(receiver), value, name) {
				owner := node.Parent
				for owner != nil && !ast.IsFunctionLike(owner) {
					owner = owner.Parent
				}
				if owner != nil && owner.Kind == ast.KindConstructor && l.discriminantConstructionUpdate(node, receiver, name) {
					found = &Refused{Where: l.program.Where(node), What: "a compound write to discriminant field '" + name + "' whose result is not proven to keep its declared literal", Fix: "initialize the declared literal directly; changing variant means building a new object"}
					return true
				}
				if (owner == nil || owner.Kind != ast.KindConstructor) && !l.discriminantLocalCandidate(receiver) {
					found = l.discriminantWriteRefusal(node, name, l.checker.GetTypeAtLocation(receiver), value, fields)
					return true
				}
			}
		}
		return node.ForEachChild(visit)
	}
	module.AsNode().ForEachChild(visit)
	return found
}

// Construction ends when the holder escapes, not when its constructor returns. Ask the same
// abstract heap and call summaries used by fresh.ProveWrites about the holder alone: a fresh
// scalar value is never evidence that its holder is fresh.
func (l *lowering) checkDiscriminantConstruction(modules []*ast.SourceFile) error {
	fields := l.discriminantFields(modules)
	sites := map[int]bool{}
	for index, site := range l.writeSites {
		target := site.node.Parent
		if target == nil {
			continue
		}
		_, name := l.discriminantTarget(target)
		if name != "" && ast.IsAssignmentTarget(target) && l.unsafeDiscriminantWrite(fields, site.holder, l.discriminantWrittenType(target), name) {
			sites[index+1] = true
		}
	}
	if len(sites) == 0 {
		return nil
	}
	for _, write := range fresh.ProveFreshHolders(l.result, sites) {
		if write.Kind != fresh.WriteField || !sites[write.Site] || write.Proven {
			continue
		}
		site := l.writeSites[write.Site-1]
		node := site.node.Parent
		if node != nil && l.unsafeDiscriminantWrite(fields, site.holder, l.discriminantWrittenType(node), write.Name) {
			return l.discriminantWriteRefusal(node, write.Name, site.holder, l.discriminantWrittenType(node), fields)
		}
	}
	return nil
}

// Resolve bracket keys exactly as narrowing does, including constant names with literal types.
func (l *lowering) discriminantTarget(node *ast.Node) (*ast.Node, string) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression, node.Name().Text()
	case ast.KindElementAccessExpression:
		if name, known := discriminantPropertyName(l.checker, node); known {
			return node.AsElementAccessExpression().Expression, name
		}
	}
	return nil, ""
}

// TypeScript permits ++ and compound assignment on a literal-typed slot even when the result
// is the wider primitive. Freshness does not prove that result is the declared tag: a constructor
// returning the wrong tag would immediately make a union view lie. Require the literal itself.
func (l *lowering) discriminantConstructionUpdate(target, receiver *ast.Node, name string) bool {
	parent := target.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if parent == nil {
		return false
	}
	update := parent.Kind == ast.KindPrefixUnaryExpression || parent.Kind == ast.KindPostfixUnaryExpression
	if parent.Kind == ast.KindBinaryExpression {
		update = ast.IsAssignmentOperator(parent.AsBinaryExpression().OperatorToken.Kind) && parent.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken
	}
	if !update {
		return false
	}
	field := l.checker.GetPropertyOfType(l.checker.GetTypeAtLocation(receiver), name)
	if field == nil {
		return false
	}
	declared := l.checker.GetTypeOfSymbol(field)
	return !l.checker.IsTypeAssignableTo(l.checker.GetTypeAtLocation(parent), declared)
}

// This is only a route to the existing heap proof, never evidence of freshness.
// An initialized local may still be under construction outside a constructor;
// publication, aliases and calls are decided by ProveFreshHolders below.
func (l *lowering) discriminantLocalCandidate(receiver *ast.Node) bool {
	receiver = ast.SkipParentheses(receiver)
	if receiver.Kind != ast.KindIdentifier {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(receiver)
	if symbol == nil || symbol.ValueDeclaration == nil || symbol.ValueDeclaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	return symbol.ValueDeclaration.AsVariableDeclaration().Initializer != nil
}
