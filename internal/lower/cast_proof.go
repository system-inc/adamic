package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The proof contains checker types only, so refusals precede representation lowering of unknown
// operands. Concrete generic instantiations are proved again when they are lowered.
type castProof struct {
	view    bool
	field   string
	allowed []*checker.Type
	classes []*checker.Type
}

const castRepair = "use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"

func (l *lowering) castAssignable(source, target *checker.Type) bool {
	return l.classAssignable(source, target) && l.enumAssignable(source, target) && l.widened(source, target, map[[2]*checker.Type]bool{}) == nil && l.castOptionalFields(source, target, map[[2]*checker.Type]bool{}) && l.optionalRelationFailure(source, target, nil, map[[2]*checker.Type]bool{}) == ""
}

// A structural view can hide an optional field of an incompatible type. Assignability permits
// adding that field back, but a cast cannot prove what the hidden field holds. Follow nested
// readonly views and callable results too; nominal class construction already proves its fields.
func (l *lowering) castOptionalFields(source, target *checker.Type, seen map[[2]*checker.Type]bool) bool {
	pair := [2]*checker.Type{source, target}
	if source == target || seen[pair] || isClassInstance(target) {
		return true
	}
	seen[pair] = true
	defer delete(seen, pair)
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if !l.castOptionalFields(member, target, seen) {
				return false
			}
		}
		return true
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if l.checker.IsTypeAssignableTo(source, member) && l.castOptionalFields(source, member, seen) {
				return true
			}
		}
		return false
	}
	if !l.structured(source) || !l.structured(target) {
		return true
	}
	from, to := l.checker.GetSignaturesOfType(source, checker.SignatureKindCall), l.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(from) == 1 && len(to) == 1 {
		if !l.castOptionalFields(l.checker.GetReturnTypeOfSignature(from[0]), l.checker.GetReturnTypeOfSignature(to[0]), seen) {
			return false
		}
		for index, parameter := range to[0].Parameters() {
			if index < len(from[0].Parameters()) && !l.castOptionalFields(l.checker.GetTypeOfSymbol(parameter), l.checker.GetTypeOfSymbol(from[0].Parameters()[index]), seen) {
				return false
			}
		}
		return true
	}
	if source.Flags()&checker.TypeFlagsObject != 0 && target.Flags()&checker.TypeFlagsObject != 0 && source.ObjectFlags()&checker.ObjectFlagsReference != 0 && target.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		fromArguments, toArguments := l.checker.GetTypeArguments(source), l.checker.GetTypeArguments(target)
		for index, argument := range toArguments {
			if index < len(fromArguments) && !l.castOptionalFields(fromArguments[index], argument, seen) {
				return false
			}
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(target) {
		inside := l.checker.GetPropertyOfType(source, property.Name)
		if inside == nil {
			if property.Flags&ast.SymbolFlagsOptional != 0 {
				return false
			}
			continue
		}
		if !l.castOptionalFields(l.checker.GetTypeOfSymbol(inside), l.checker.GetTypeOfSymbol(property), seen) {
			return false
		}
	}
	return true
}

func castMembers(proven *checker.Type) []*checker.Type {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		return proven.Types()
	}
	return []*checker.Type{proven}
}

func (l *lowering) castProof(node *ast.Node) (castProof, error) {
	as := node.AsAsExpression()
	if as.Type.Kind == ast.KindTypeReference && as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
		return castProof{}, nil
	}
	source := l.concrete(l.checker.GetTypeAtLocation(as.Expression))
	target := l.concrete(l.checker.GetTypeAtLocation(node))
	refused := &Refused{Where: l.program.Where(node), What: "a cast the runtime can't check", Fix: castRepair}
	inner := ast.SkipParentheses(as.Expression)
	if inner.Kind == ast.KindAsExpression && l.checker.GetTypeAtLocation(inner).Flags()&checker.TypeFlagsUnknown != 0 {
		return castProof{}, refused
	}
	if source.Flags()&checker.TypeFlagsAny != 0 || target.Flags()&checker.TypeFlagsAny != 0 {
		return castProof{}, refused
	}
	members, targets := castMembers(source), castMembers(target)
	allClasses := true
	for _, member := range append(append([]*checker.Type{}, members...), targets...) {
		allClasses = allClasses && isClassInstance(member)
	}
	var upcastFailure error
	// A structural assignability result is only an upcast candidate. Nominal
	// downcasts may also look assignable, but still need their runtime identity check.
	if l.checker.IsTypeAssignableTo(source, target) || l.checker.IsTypeAssignableTo(l.checker.GetWidenedType(source), target) {
		err := l.provenRelation(node, as.Expression, l.checker.GetTypeAtLocation(node))
		if err == nil {
			return castProof{}, nil
		}
		rawSource, rawTarget := l.checker.GetTypeAtLocation(as.Expression), l.checker.GetTypeAtLocation(node)
		if !allClasses || l.optionalRelationFailure(rawSource, rawTarget, as.Expression, map[[2]*checker.Type]bool{}) != "" || l.freshOrWidened(as.Expression, rawSource, rawTarget) != nil {
			if reason, ok := err.(*Refused); ok {
				reason.Fix += "; " + castRepair
			}
			return castProof{}, err
		}
		if reason, ok := err.(*Refused); ok {
			reason.Fix += "; " + castRepair
		}
		upcastFailure = err
	}

	if !allClasses && l.checker.IsTypeAssignableTo(source, target) {
		if err := l.refuseWidening(node); err != nil {
			if reason, ok := err.(*Refused); ok {
				reason.Fix += "; " + castRepair
			}
			return castProof{}, err
		}
	}
	if allClasses {
		for _, wanted := range targets {
			related := false
			for _, member := range members {
				if l.castAssignable(wanted, member) && l.nominalAncestor(wanted, member, map[[2]*checker.Type]bool{}) {
					if !l.classCastArguments(member, wanted) {
						return castProof{}, refused
					}
					related = true
				}
				// An erased identity cannot distinguish incompatible arguments of the same
				// generic class, even if another source member provides a valid ancestor.
				if l.classView(member, l.classNodeFor(wanted)) != nil && !l.nominalAncestor(member, wanted, map[[2]*checker.Type]bool{}) {
					return castProof{}, refused
				}
			}
			if !related {
				if upcastFailure != nil {
					return castProof{}, upcastFailure
				}
				return castProof{}, refused
			}
		}
		return castProof{classes: targets}, nil
	}
	if source.Flags()&checker.TypeFlagsUnion == 0 {
		// Eligibility routes to view(), which must certify the complete field
		// contract and all read paths. It is not an upcast proof.
		if source.Flags()&checker.TypeFlagsObject != 0 && target.Flags()&checker.TypeFlagsObject != 0 && !isClassInstance(target) && l.checker.IsTypeAssignableTo(target, source) {
			return castProof{view: true}, nil
		}
		return castProof{}, refused
	}
	if !l.castUnionWrites(source) {
		return castProof{}, refused
	}
	// Checking kind cannot prove a refinement of a member's payload or callback signature.
	for _, wanted := range targets {
		found := false
		for _, member := range members {
			if l.castAssignable(member, wanted) && l.castAssignable(wanted, member) {
				found = true
			}
		}
		if !found {
			return castProof{}, refused
		}
	}
	for _, property := range l.checker.GetPropertiesOfType(members[0]) {
		proof := castProof{field: property.Name}
		literals := []*checker.Type{}
		valid := true
		for _, member := range members {
			literal := l.fieldLiteral(member, proof.field)
			if literal == nil {
				valid = false
				break
			}
			// Enum aliases compare by runtime value, including aliases from different enums.
			for _, previous := range literals {
				mask := checker.TypeFlagsStringLiteral | checker.TypeFlagsNumberLiteral | checker.TypeFlagsBooleanLiteral
				if previous.Flags()&mask != literal.Flags()&mask || reflect.DeepEqual(previous.AsLiteralType().Value(), literal.AsLiteralType().Value()) {
					valid = false
				}
			}
			literals = append(literals, literal)
			if l.castAssignable(member, target) {
				proof.allowed = append(proof.allowed, literal)
			}
		}
		if valid && len(proof.allowed) > 0 {
			return proof, nil
		}
	}
	return castProof{}, refused
}

// Writing through a union view must not change a member's tag or narrow payload promise. A common
// mutable slot is safe only if everything the union can write fits every member that owns it.
func (l *lowering) castUnionWrites(source *checker.Type) bool {
	for _, property := range l.checker.GetPropertiesOfType(source) {
		if l.checker.IsReadonlySymbol(property) || property.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		written := l.checker.GetTypeOfSymbol(property)
		for _, member := range source.Types() {
			inside := l.checker.GetPropertyOfType(member, property.Name)
			if inside == nil || !l.castAssignable(written, l.checker.GetTypeOfSymbol(inside)) {
				return false
			}
		}
	}
	return true
}

// An erased identity proves no type arguments. Every target parameter must survive directly in
// the source ancestor's arguments, so that nominal invariance already fixes it. A Box<T> whose
// base is a nongeneric Animal is uncheckable; Box<number> and Box<string> have the same tag there.
func (l *lowering) classCastArguments(source, target *checker.Type) bool {
	declaration := l.classNodeFor(target)
	if declaration == nil || len(declaration.TypeParameters()) == 0 {
		return true
	}
	view := l.classView(l.checker.GetTypeAtLocation(declaration.Name()), l.classNodeFor(source))
	if view == nil {
		return false
	}
	arguments := l.checker.GetTypeArguments(view)
	for _, parameter := range declaration.TypeParameters() {
		proven := l.checker.GetTypeAtLocation(parameter.Name())
		found := false
		for _, argument := range arguments {
			if argument == proven {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// Ordinary IR exposes the check to both backends and ownership analyses. The operand is one call
// argument; the helper only rereads its parameter, never the source expression.
func (l *lowering) checkedClassCast(node *ast.Node, value ir.Expression, targets []*checker.Type, message string) (ir.Expression, error) {
	identities := []int{}
	for _, target := range targets {
		declaration := l.classNodeFor(target)
		if declaration == nil || declaration.Kind != ast.KindClassDeclaration {
			return nil, &Refused{Where: l.program.Where(node), What: "a cast without a declared class identity", Fix: castRepair}
		}
		if len(declaration.TypeParameters()) > 0 {
			identities = append(identities, l.classIdentity(declaration))
		} else {
			instance, err := l.instantiate(declaration, target, node)
			if err != nil {
				return nil, err
			}
			identities = append(identities, instance.class)
		}
	}
	function := len(l.result.Functions)
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "cast_value", Type: ir.Object, Function: function})
	read := ir.Read{Local: local, Of: ir.Object}
	var condition ir.Expression
	for _, identity := range identities {
		test := ir.InstanceOf{Value: read, Class: identity}
		if condition == nil {
			condition = test
		} else {
			condition = ir.Binary{Operator: ir.Or, Left: condition, Right: test}
		}
	}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "checked_class_cast", Parameters: []int{local}, Returns: ir.Object, Body: []ir.Statement{
		ir.If{Condition: condition, Then: []ir.Statement{ir.Return{Value: read}}},
		ir.Panic{Message: ir.StringConstant{Index: l.constant(message)}},
	}})
	return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.Object}, nil
}
