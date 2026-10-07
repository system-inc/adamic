package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	_ "unsafe"
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
		if proven.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range proven.Types() {
				for _, property := range l.checker.GetPropertiesOfType(member) {
					if l.fieldLiteral(member, property.Name) != nil && discriminantProperty(l.checker, proven, property.Name) {
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
			if ast.IsExpression(node) || node.Kind == ast.KindTypeAliasDeclaration || node.Kind == ast.KindTypeReference || node.Kind == ast.KindUnionType {
				note(l.checker.GetTypeAtLocation(node))
			}
			return node.ForEachChild(visit)
		}
		module.AsNode().ForEachChild(visit)
	}
	return fields
}

func (l *lowering) discriminantView(fields []discriminantField, holder *checker.Type, name string) bool {
	if holder == nil {
		return false
	}
	if holder.Flags()&checker.TypeFlagsTypeParameter != 0 {
		holder = l.checker.GetBaseConstraintOfType(holder)
		if holder == nil {
			return false
		}
	}
	for _, field := range fields {
		if field.name == name && l.checker.IsTypeAssignableTo(field.member, holder) {
			return true
		}
	}
	return false
}

func (l *lowering) discriminantWriteRefusal(node *ast.Node, name string) error {
	return &Refused{Where: l.program.Where(node), What: "a write to discriminant field '" + name + "' after construction", Fix: "changing variant means building a new object"}
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
			if l.discriminantView(fields, l.checker.GetTypeAtLocation(receiver), name) {
				owner := node.Parent
				for owner != nil && !ast.IsFunctionLike(owner) {
					owner = owner.Parent
				}
				if owner != nil && owner.Kind == ast.KindConstructor && l.discriminantConstructionUpdate(node, receiver, name) {
					found = &Refused{Where: l.program.Where(node), What: "a compound write to discriminant field '" + name + "' whose result is not proven to keep its declared literal", Fix: "initialize the declared literal directly; changing variant means building a new object"}
					return true
				}
				if owner == nil || owner.Kind != ast.KindConstructor {
					found = l.discriminantWriteRefusal(node, name)
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
		for _, field := range fields {
			if l.discriminantView(fields, site.holder, field.name) {
				sites[index+1] = true
				break
			}
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
		if l.discriminantView(fields, site.holder, write.Name) {
			node := site.node
			if parent := node.Parent; parent != nil && parent.Kind == ast.KindPropertyAccessExpression {
				node = parent
			}
			return l.discriminantWriteRefusal(node, write.Name)
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
