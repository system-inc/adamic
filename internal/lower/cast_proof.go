package lower

import (
	"reflect"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The proof contains checker types only, so refusals precede representation lowering of unknown
// operands. Concrete generic instantiations are proved again when they are lowered.
type castProof struct {
	structuralView bool
	interfaceView  bool
	field          string
	allowed        []*checker.Type
	classes        []*checker.Type
	construction   bool
	nodeArray      bool
	speculation    bool
	// brandMint is a cast that adds only phantom brand members (brand_mint.go): erased, and listed.
	brandMint *brandMint
	// overloadDefault is an `x as unknown as T` pair every overload proves (overload_default_cast.go):
	// erased, lowered as x.
	overloadDefault bool
}

// castRepair is the fix a refused cast names, by file kind. In .a a cast to a union member is refused too
// (keptCheckDowncastRefusal), so .a narrows by the tag; .ts keeps the checked member cast.
func (lowering *lowering) castRepair(node *ast.Node) string {
	if strings.HasSuffix(lowering.program.FileName(ast.GetSourceFileOfNode(node)), ".a") {
		return "use a proven upcast, narrow a discriminated union by its tag (`if (shape.kind !== 'Circle') panic('why')`), downcast along nominal class ancestry, or move the file to `.ts`, where a member cast is checked (adamic/no-unchecked-cast)"
	}
	return "use a proven upcast, cast a discriminated object union with unique literal or enum tags to members or a sub-union, or downcast along nominal class ancestry (adamic/no-unchecked-cast)"
}

func (lowering *lowering) castAssignable(source, target *checker.Type) bool {
	return lowering.castAssignableAt(nil, source, target)
}

// castAssignableAt is castAssignable for the value at operand: its widening is judged as at any slot,
// so a fresh literal ([]) held by nothing else relates by its own elements (freshOrWidened). A nil
// operand is a type with no value behind it, judged by widened alone. One list of rules for both, so a
// rule added to one can't miss the other (Lower's should on #yth65gz).
func (lowering *lowering) castAssignableAt(operand *ast.Node, source, target *checker.Type) bool {
	if !lowering.classAssignable(source, target) || !lowering.enumAssignable(source, target) {
		return false
	}
	if operand != nil {
		if lowering.freshOrWidened(operand, source, target) != nil {
			return false
		}
	} else if lowering.widened(source, target, map[[2]*checker.Type]bool{}) != nil {
		return false
	}
	// Two walks guard an optional field a view hides. A value at a cast (an operand) is judged by
	// optionalRelationFailure with the operand as its origin, which follows containers and judges an
	// object literal written there exactly: a missing optional field is absent, unless it spreads. The
	// castOptionalFields walk, over every member, misjudges such a value (measured on Golem and Cloud,
	// #yth65gz): it refused a T[] seen as ReadonlyArray<T> by the array's lib members, and a { a: 1 }
	// missing an optional b. So it's skipped for an array seen as an array and for an object literal,
	// where optionalRelationFailure is the judge. A type with no value (castAssignable) keeps both.
	if !(operand != nil && (lowering.arrayLikeType(source) && lowering.arrayLikeType(target) ||
		ast.SkipParentheses(operand).Kind == ast.KindObjectLiteralExpression)) {
		if !lowering.castOptionalFields(source, target, map[[2]*checker.Type]bool{}) {
			return false
		}
	}
	return lowering.optionalRelationFailure(source, target, operand, map[[2]*checker.Type]bool{}) == ""
}

// arrayLikeType is whether a type is an array, mutable or readonly, by the checker's own test. A cast is
// judged on the brand, so a phantom extension isn't one here, though lowering holds it as its base
// (unbranded.go): this asks the checker directly, and keeps isLibraryType's unbranding off the brand.
func (lowering *lowering) arrayLikeType(proven *checker.Type) bool {
	if base, _ := lowering.phantomExtension(proven); base != nil {
		return false
	}
	return lowering.checker.IsArrayType(proven) || lowering.isLibraryType(proven, "ReadonlyArray")
}

// A structural view can hide an optional field of an incompatible type. Assignability permits
// adding that field back, but a cast cannot prove what the hidden field holds. Follow nested
// readonly views and callable results too; nominal class construction already proves its fields.
func (lowering *lowering) castOptionalFields(source, target *checker.Type, seen map[[2]*checker.Type]bool) bool {
	pair := [2]*checker.Type{source, target}
	if source == target || seen[pair] || isClassInstance(target) {
		return true
	}
	seen[pair] = true
	defer delete(seen, pair)
	if source.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range source.Types() {
			if !lowering.castOptionalFields(member, target, seen) {
				return false
			}
		}
		return true
	}
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range target.Types() {
			if lowering.checker.IsTypeAssignableTo(source, member) &&
				lowering.castOptionalFields(source, member, seen) {
				return true
			}
		}
		return false
	}
	if !lowering.structured(source) || !lowering.structured(target) {
		return true
	}
	from, to := lowering.checker.GetSignaturesOfType(
		source,
		checker.SignatureKindCall,
	), lowering.checker.GetSignaturesOfType(target, checker.SignatureKindCall)
	if len(from) == 1 && len(to) == 1 {
		if !lowering.castOptionalFields(
			lowering.checker.GetReturnTypeOfSignature(from[0]),
			lowering.checker.GetReturnTypeOfSignature(to[0]),
			seen,
		) {
			return false
		}
		for index, parameter := range to[0].Parameters() {
			if index < len(from[0].Parameters()) && !lowering.castOptionalFields(
				lowering.checker.GetTypeOfSymbol(parameter),
				lowering.checker.GetTypeOfSymbol(from[0].Parameters()[index]),
				seen,
			) {
				return false
			}
		}
		return true
	}
	if source.Flags()&checker.TypeFlagsObject != 0 &&
		target.Flags()&checker.TypeFlagsObject != 0 &&
		source.ObjectFlags()&checker.ObjectFlagsReference != 0 &&
		target.ObjectFlags()&checker.ObjectFlagsReference != 0 {
		fromArguments, toArguments := lowering.checker.GetTypeArguments(
			source,
		), lowering.checker.GetTypeArguments(target)
		for index, argument := range toArguments {
			if index < len(fromArguments) && !lowering.castOptionalFields(fromArguments[index], argument, seen) {
				return false
			}
		}
	}
	for _, property := range lowering.checker.GetPropertiesOfType(target) {
		inside := lowering.checker.GetPropertyOfType(source, property.Name)
		if inside == nil {
			if property.Flags&ast.SymbolFlagsOptional != 0 {
				return false
			}
			continue
		}
		if !lowering.castOptionalFields(
			lowering.checker.GetTypeOfSymbol(inside),
			lowering.checker.GetTypeOfSymbol(property),
			seen,
		) {
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

// castRefusal refuses a cast the runtime can't check, and records a checked one in
// l.refusal.checkedCasts for stringWideningRefusal (3100).
func (lowering *lowering) castRefusal(node *ast.Node) error {
	if node.Kind == ast.KindAsExpression && lowering.processPath(node.AsAsExpression().Expression) == "" {
		proof, castProofError := lowering.castProof(node)
		if castProofError != nil {
			return castProofError
		}
		if proof.interfaceView || len(proof.allowed) > 0 || len(proof.classes) > 0 {
			if lowering.refusal.checkedCasts == nil {
				lowering.refusal.checkedCasts = map[*ast.Node]bool{}
			}
			lowering.refusal.checkedCasts[node] = true
		}
	}
	return nil
}

func init() { refusalRules.register(1900, "castRefusal", (*lowering).castRefusal) }

func (lowering *lowering) castProof(node *ast.Node) (castProof, error) {
	as := node.AsAsExpression()
	if as.Type.Kind == ast.KindTypeReference &&
		ast.IsIdentifier(as.Type.AsTypeReferenceNode().TypeName) &&
		as.Type.AsTypeReferenceNode().TypeName.Text() == "const" {
		return castProof{}, nil
	}
	source := lowering.concrete(lowering.checker.GetTypeAtLocation(as.Expression))
	target := lowering.concrete(lowering.checker.GetTypeAtLocation(node))
	if lowering.phantomCast(source, target) {
		return castProof{}, nil
	}
	operand, from := lowering.mintOperand(as)
	if mint := lowering.brandMintOf(operand, from, target); mint != nil {
		return castProof{brandMint: mint}, lowering.brandMintRefusal(node, target, mint)
	}
	refused := &Refused{
		Where: lowering.program.Where(node),
		What:  "a cast the runtime can't check",
		Fix:   lowering.castRepair(node),
	}
	inner := ast.SkipParentheses(as.Expression)
	if inner.Kind == ast.KindAsExpression &&
		lowering.checker.GetTypeAtLocation(inner).Flags()&checker.TypeFlagsUnknown != 0 {
		if lowering.overloadDefaultPair(node) {
			return castProof{overloadDefault: true}, nil
		}
		return castProof{}, refused
	}
	if source.Flags()&checker.TypeFlagsAny != 0 || target.Flags()&checker.TypeFlagsAny != 0 {
		return castProof{}, refused
	}
	if lowering.captureFreeSplit(as.Expression, target) {
		return castProof{}, nil
	}
	if lowering.speculativeCast(node) {
		return castProof{speculation: true}, nil
	}
	if lowering.nodeArrayCast(node) {
		return castProof{nodeArray: true}, nil
	}
	if lowering.constructionCast(node) {
		return castProof{construction: true}, nil
	}
	members, targets := castMembers(source), castMembers(target)
	allClasses := true
	for _, member := range append(append([]*checker.Type{}, members...), targets...) {
		allClasses = allClasses && isClassInstance(member)
	}
	var upcastFailure error
	// A structural assignability result is only an upcast candidate. Nominal
	// downcasts may also look assignable, but still need their runtime identity check.
	if lowering.checker.IsTypeAssignableTo(source, target) ||
		lowering.checker.IsTypeAssignableTo(lowering.checker.GetWidenedType(source), target) {
		provenRelationError := lowering.provenRelation(node, as.Expression, lowering.checker.GetTypeAtLocation(node))
		if provenRelationError == nil {
			return castProof{}, nil
		}
		rawSource, rawTarget := lowering.checker.GetTypeAtLocation(
			as.Expression,
		), lowering.checker.GetTypeAtLocation(node)
		if !allClasses ||
			lowering.optionalRelationFailure(rawSource, rawTarget, as.Expression, map[[2]*checker.Type]bool{}) != "" ||
			lowering.freshOrWidened(as.Expression, rawSource, rawTarget) != nil {
			if reason, ok := provenRelationError.(*Refused); ok {
				reason.Fix += "; " + lowering.castRepair(node)
			}
			return castProof{}, provenRelationError
		}
		if reason, ok := provenRelationError.(*Refused); ok {
			reason.Fix += "; " + lowering.castRepair(node)
		}
		upcastFailure = provenRelationError
	}

	if !allClasses && lowering.checker.IsTypeAssignableTo(source, target) {
		if refuseWideningError := lowering.refuseWidening(node); refuseWideningError != nil {
			if reason, ok := refuseWideningError.(*Refused); ok {
				reason.Fix += "; " + lowering.castRepair(node)
			}
			return castProof{}, refuseWideningError
		}
	}
	if allClasses {
		for _, wanted := range targets {
			related := false
			for _, member := range members {
				if lowering.castAssignable(wanted, member) &&
					lowering.nominalAncestor(wanted, member, map[[2]*checker.Type]bool{}) {
					if !lowering.classCastArguments(member, wanted) {
						return castProof{}, refused
					}
					related = true
				}
				// An erased identity cannot distinguish incompatible arguments of the same
				// generic class, even if another source member provides a valid ancestor.
				if lowering.classView(member, lowering.classNodeFor(wanted)) != nil &&
					!lowering.nominalAncestor(member, wanted, map[[2]*checker.Type]bool{}) {
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

		// A proof's view is no view: neither its origin nor its contract's getters outlive it.
		origins, getters := len(lowering.result.ViewOrigins), len(lowering.contractGetters)
		defer func() {
			lowering.result.ViewOrigins = lowering.result.ViewOrigins[:origins]
			lowering.contractGetters = lowering.contractGetters[:getters]
		}()
		if checked, interfaceCastError := lowering.interfaceCast(
			node,
			ir.Read{Of: ir.Object},
			source,
			target,
		); checked != nil || interfaceCastError != nil {
			proof := castProof{interfaceView: checked != nil}
			// The tag the view tests, which .a's refusal names in the narrowing it offers.
			if cast, isCast := checked.(ir.CheckedCast); isCast {
				proof.field = cast.Field
			}
			return proof, interfaceCastError
		}
		from, fromKnown := lowering.representation(source)
		to, toKnown := lowering.representation(target)
		if source.Flags()&checker.TypeFlagsObject != 0 &&
			target.Flags()&checker.TypeFlagsObject != 0 &&
			fromKnown &&
			toKnown &&
			from == ir.Object &&
			to == ir.Object &&
			!isClassInstance(target) &&
			lowering.checker.IsTypeAssignableTo(target, source) &&
			lowering.widened(target, source, map[[2]*checker.Type]bool{}) == nil &&
			len(lowering.checker.GetIndexInfosOfType(target)) == 0 {
			return castProof{structuralView: true}, nil
		}

		return castProof{}, refused
	}
	if !lowering.castUnionWrites(source) {
		return castProof{}, refused
	}
	// Checking kind cannot prove a refinement of a member's payload or callback signature.
	for _, wanted := range targets {
		found := false
		for _, member := range members {
			if lowering.castAssignable(member, wanted) && lowering.castAssignable(wanted, member) {
				found = true
			}
		}
		if !found {
			return castProof{}, refused
		}
	}
	for _, property := range lowering.checker.GetPropertiesOfType(members[0]) {
		proof := castProof{field: property.Name}
		literals := []*checker.Type{}
		valid := true
		for _, member := range members {
			literal := lowering.fieldLiteral(member, proof.field)
			if literal == nil {
				valid = false
				break
			}
			// Enum aliases compare by runtime value, including aliases from different enums.
			for _, previous := range literals {
				mask := checker.TypeFlagsStringLiteral | checker.TypeFlagsNumberLiteral | checker.TypeFlagsBooleanLiteral
				if previous.Flags()&mask != literal.Flags()&mask ||
					reflect.DeepEqual(previous.AsLiteralType().Value(), literal.AsLiteralType().Value()) {
					valid = false
				}
			}
			literals = append(literals, literal)
			if lowering.castAssignable(member, target) {
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
func (lowering *lowering) castUnionWrites(source *checker.Type) bool {
	for _, property := range lowering.checker.GetPropertiesOfType(source) {
		if lowering.checker.IsReadonlySymbol(property) || property.Flags&ast.SymbolFlagsMethod != 0 {
			continue
		}
		written := lowering.checker.GetTypeOfSymbol(property)
		for _, member := range source.Types() {
			inside := lowering.checker.GetPropertyOfType(member, property.Name)
			if inside == nil || !lowering.castAssignable(written, lowering.checker.GetTypeOfSymbol(inside)) {
				return false
			}
		}
	}
	return true
}

// An erased identity proves no type arguments. Every target parameter must survive directly in
// the source ancestor's arguments, so that nominal invariance already fixes it. A Box<T> whose
// base is a nongeneric Animal is uncheckable; Box<number> and Box<string> have the same tag there.
func (lowering *lowering) classCastArguments(source, target *checker.Type) bool {
	declaration := lowering.classNodeFor(target)
	if declaration == nil || len(declaration.TypeParameters()) == 0 {
		return true
	}
	view := lowering.classView(lowering.checker.GetTypeAtLocation(declaration.Name()), lowering.classNodeFor(source))
	if view == nil {
		return false
	}
	arguments := lowering.checker.GetTypeArguments(view)
	for _, parameter := range declaration.TypeParameters() {
		proven := lowering.checker.GetTypeAtLocation(parameter.Name())
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
func (lowering *lowering) checkedClassCast(
	node *ast.Node,
	value ir.Expression,
	targets []*checker.Type,
	message string,
) (ir.Expression, error) {
	identities := []int{}
	for _, target := range targets {
		declaration := lowering.classNodeFor(target)
		if declaration == nil || declaration.Kind != ast.KindClassDeclaration {
			return nil, &Refused{
				Where: lowering.program.Where(node),
				What:  "a cast without a declared class identity",
				Fix:   lowering.castRepair(node),
			}
		}
		if len(declaration.TypeParameters()) > 0 {
			identities = append(identities, lowering.classIdentity(declaration))
		} else {
			instance, instantiateError := lowering.instantiate(declaration, target, node)
			if instantiateError != nil {
				return nil, instantiateError
			}
			identities = append(identities, instance.class)
		}
	}
	function := len(lowering.result.Functions)
	local := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "cast_value", Type: ir.Object, Function: function},
	)
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
	lowering.result.Functions = append(
		lowering.result.Functions,
		ir.Function{Name: "checked_class_cast", Parameters: []int{local}, Returns: ir.Object, Body: []ir.Statement{
			ir.If{Condition: condition, Then: []ir.Statement{ir.Return{Value: read}}},
			ir.Panic{Message: ir.StringConstant{Index: lowering.constant(message)}},
		}},
	)
	return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.Object}, nil
}
