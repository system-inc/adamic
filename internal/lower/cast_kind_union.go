package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// kindUnion is a cast from an open-kind interface to a union of kinded interfaces (tsc's `node as
// PropertyAccessExpression | ElementAccessExpression` of a Node, whose kind is the open SyntaxKind): the
// tag interface downcast Adamic ruled on Nov 3 (#b5w3ycg slice 4). The cast tests the tag against exactly
// the target members' literals, and every later read through the result is checked by each member's own
// view, since the tag proves the kind, not the layout. Two members may share a literal (a
// MethodDeclaration and a PrivateIdentifierMethodDeclaration): the tag passes either, and a field the
// value's own layout lacks stops at its read through the view.
type kindUnion struct {
	field   string
	members []*checker.Type
	// literals are the tag's allowed values, each once, in the members' order.
	literals []*checker.Type
}

// kindUnionCast is the judgment castProof makes of a plain interface source cast to a union target, after
// interfaceCast, which takes one member, finds nothing. The target holds two or more object members, none a
// class instance or an intersection (an intersection target is a later head), each assignable to the source
// with no slot widened, and each holding one constant literal on the same tag field, a field the source
// declares, every literal of one runtime type. The source's fields are tried in order, as interfaceCast tries
// them. In .a a test every source value passes is vacuous and refused, as interfaceCast refuses it. Nil and no
// error is a cast this judgment doesn't hold, which castProof refuses as before.
func (l *lowering) kindUnionCast(node *ast.Node, source, target *checker.Type) (*kindUnion, error) {
	if source.Flags()&checker.TypeFlagsObject == 0 || target.Flags()&checker.TypeFlagsUnion == 0 || len(target.Types()) < 2 {
		return nil, nil
	}
	if from, known := l.representation(source); !known || from != ir.Object {
		return nil, nil
	}
	members := target.Types()
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsObject == 0 || member.Flags()&checker.TypeFlagsIntersection != 0 || isClassInstance(member) {
			return nil, nil
		}
		if !l.checker.IsTypeAssignableTo(member, source) || l.widened(member, source, map[[2]*checker.Type]bool{}) != nil {
			return nil, nil
		}
	}
	var vacuous *Refused
	for _, property := range l.checker.GetPropertiesOfType(source) {
		cast := &kindUnion{field: property.Name, members: members}
		var fieldType ir.Type
		valid := true
		for _, member := range members {
			literal := l.fieldLiteral(member, cast.field)
			if literal == nil {
				valid = false
				break
			}
			_, of, constant := l.literalConstant(literal)
			if !constant || (fieldType != 0 && fieldType != of) {
				valid = false
				break
			}
			fieldType = of
			seen := false
			for _, previous := range cast.literals {
				seen = seen || previous == literal
			}
			if !seen {
				cast.literals = append(cast.literals, literal)
			}
		}
		if !valid {
			continue
		}
		if !l.checkedAssertionSource(node) && l.withinLiterals(l.checker.GetTypeOfSymbol(property), cast.literals) {
			if vacuous == nil {
				vacuous = &Refused{Where: l.program.Where(node), What: "a downcast to " + l.checker.TypeToString(target) + " whose tag test is vacuous: every " + l.checker.TypeToString(source) + "'s " + cast.field + " is one of the target's", Fix: "give each of " + l.checker.TypeToString(target) + " a tag of its own in a discriminated union, or build the value as one so the compiler can prove it"}
			}
			continue
		}
		return cast, nil
	}
	if vacuous != nil {
		return nil, vacuous
	}
	return nil, nil
}

// kindUnionValue lowers the cast kindUnionCast proved: the tag test over the target's literals, with the
// cast's own message, then each member's checked view over the same value, so a read through the result is
// checked wherever any member's view checks that field (checkedViewReceiverField's union walk). The union is
// never viewed whole: legacyView refuses every non-scalar field, which every tsc node has (predicates.go
// views a union target the same way, member by member).
func (l *lowering) kindUnionValue(node *ast.Node, value ir.Expression, source *checker.Type, proof *kindUnion, message string) (ir.Expression, error) {
	if value.Type() != ir.Object {
		return nil, l.notYet(node, "a checked kind cast from "+l.checker.TypeToString(source)+" whose value isn't a plain object")
	}
	cast := ir.CheckedCast{Value: value, Field: proof.field, CheckedFields: true, Message: message}
	for _, literal := range proof.literals {
		allowed, fieldType, _ := l.literalConstant(literal)
		cast.Allowed = append(cast.Allowed, allowed)
		cast.FieldType = fieldType
	}
	for _, member := range proof.members {
		if _, err := l.view(node, value, source, member); err != nil {
			return nil, err
		}
		l.viewSources[len(l.viewSources)-1].cast = true
	}
	if err := l.kindUnionWrites(proof.members); err != nil {
		return nil, err
	}
	return cast, nil
}

// kindUnionWrites stops a write whose receiver is the kind cast's union, or a union of two or more of its
// members, as NotYet: Adamic ruled that write on #b5w3ycg as head 2's, with a presence check at the write,
// Rule C's record and a tag that can't move, and this head keeps it waiting. The union's members are compared,
// not its identity, since an alias (AccessExpression) and the same union written out are two checker types. A
// write after narrowing to one member (`if (a.kind === X) a.f = v`, receiver X) isn't stopped: it goes on
// main's checked-write road through X's view, the road main's one-kind cast (interfaceCast) already admits, so
// this head adds no write road main doesn't have. Views are keyed by type, so a write through a value of the
// union's type that no cast made stops here too. It reads view's site index, every access of the program built
// once (view_sites.go), so it walks no module.
func (l *lowering) kindUnionWrites(members []*checker.Type) error {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	if l.viewSites == nil {
		l.viewSites = &viewSiteIndex{}
	}
	for _, part := range l.viewSites.sites(modules).accesses {
		if !ast.IsAssignmentTarget(part) && part.Parent.Kind != ast.KindDeleteExpression {
			continue
		}
		var receiver *ast.Node
		if part.Kind == ast.KindPropertyAccessExpression {
			receiver = part.AsPropertyAccessExpression().Expression
		} else {
			receiver = part.AsElementAccessExpression().Expression
		}
		held := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(receiver)))
		if held.Flags()&checker.TypeFlagsUnion == 0 || len(held.Types()) < 2 {
			continue
		}
		within := true
		for _, member := range held.Types() {
			found := false
			for _, wanted := range members {
				found = found || member == wanted
			}
			within = within && found
		}
		if within {
			return l.notYet(part, "a write through a kind-checked cast's union "+l.checker.TypeToString(held)+", which waits on #b5w3ycg head 2's ruled write")
		}
	}
	return nil
}
