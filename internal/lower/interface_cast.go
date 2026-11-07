package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A tag chooses the interface; its fields remain checked at every read. In .a the tag must tell the
// target from the source: a test every source value passes proves nothing, so .a takes the first test
// a source value can fail and refuses the cast when there is none. .ts keeps the first tag, checked.
func (lowering *lowering) interfaceCast(
	node *ast.Node,
	value ir.Expression,
	source, target *checker.Type,
) (ir.Expression, error) {
	if source.Flags()&checker.TypeFlagsObject == 0 ||
		target.Flags()&checker.TypeFlagsObject == 0 ||
		value.Type() != ir.Object ||
		isClassInstance(target) {
		return nil, nil
	}
	var vacuous *Refused
	for _, property := range lowering.checker.GetPropertiesOfType(source) {
		field := property.Name
		literal := lowering.fieldLiteral(target, field)
		if literal == nil ||
			!lowering.checker.IsTypeAssignableTo(target, source) ||
			lowering.widened(target, source, map[[2]*checker.Type]bool{}) != nil {
			continue
		}
		allowed, fieldType, constant := lowering.literalConstant(literal)
		if !constant {
			continue
		}
		if !lowering.checkedAssertionSource(node) &&
			lowering.withinLiterals(lowering.checker.GetTypeOfSymbol(property), []*checker.Type{literal}) {
			if vacuous == nil {
				vacuous = &Refused{
					Where: lowering.program.Where(node),
					What: "a downcast to " +
						lowering.checker.TypeToString(target) +
						" whose tag test is vacuous: every " +
						lowering.checker.TypeToString(source) +
						" has " +
						field +
						" " +
						lowering.checker.TypeToString(literal),
					Fix: "give " +
						lowering.checker.TypeToString(target) +
						" a tag of its own in a discriminated union, or build the value as a " +
						lowering.checker.TypeToString(target) +
						" so the compiler can prove it",
				}
			}
			continue
		}
		if _, viewError := lowering.view(node, value, source, target); viewError != nil {
			return nil, viewError
		}
		lowering.viewSources[len(lowering.viewSources)-1].cast = true
		return ir.CheckedCast{
			Value:         value,
			Field:         field,
			FieldType:     fieldType,
			Allowed:       []ir.Expression{allowed},
			CheckedFields: true,
			Message: "cast failed: this " +
				lowering.checker.TypeToString(source) +
				" is not a " +
				lowering.checker.TypeToString(target),
		}, nil
	}
	if vacuous != nil {
		return nil, vacuous
	}
	return nil, nil
}

// withinLiterals is whether every value declared's type admits is one of the allowed literals,
// compared by runtime value as a cast's tag test compares, so the test passes for every one of them: a
// union whose members all are, a literal equal to one (two enums' members sharing a value are equal
// here), an intersection with a member that is, or a type parameter whose constraint is. Any other
// type (string, undefined, an object, unknown, a template) admits a value the test can fail.
func (lowering *lowering) withinLiterals(declared *checker.Type, allowed []*checker.Type) bool {
	switch {
	case declared.Flags()&checker.TypeFlagsUnion != 0:
		for _, member := range declared.Types() {
			if !lowering.withinLiterals(member, allowed) {
				return false
			}
		}
		return true
	case declared.Flags()&checker.TypeFlagsIntersection != 0:
		for _, member := range declared.Types() {
			if lowering.withinLiterals(member, allowed) {
				return true
			}
		}
		return false
	case declared.Flags()&checker.TypeFlagsTypeParameter != 0:
		constraint := lowering.checker.GetBaseConstraintOfType(declared)
		return constraint != nil && constraint != declared && lowering.withinLiterals(constraint, allowed)
	}
	mask := checker.TypeFlagsStringLiteral | checker.TypeFlagsNumberLiteral | checker.TypeFlagsBooleanLiteral
	if declared.Flags()&mask == 0 {
		return false
	}
	for _, literal := range allowed {
		if literal.Flags()&mask == declared.Flags()&mask &&
			reflect.DeepEqual(literal.AsLiteralType().Value(), declared.AsLiteralType().Value()) {
			return true
		}
	}
	return false
}

// view registers checks by receiver type and member, including nested contracts.
// An unrelated receiver with the same member name does not become a view. Source
// is the type the value had before the view, nil when there's none to name; what
// it already guarantees is read honestly through a wider type (view_widened.go).
// Value is where the view's value enters it, unless a CheckedCast wraps it.
func (lowering *lowering) view(
	node *ast.Node,
	value ir.Expression,
	source, target *checker.Type,
) (ir.Expression, error) {
	lowering.viewSources = append(lowering.viewSources, viewSource{source: source, target: target, entry: value})
	if target.Flags()&checker.TypeFlagsUnion != 0 {
		return lowering.legacyView(node, value, target)
	}
	fields, viewSchemaError := lowering.viewSchema(node, target)
	if viewSchemaError != nil {
		return nil, viewSchemaError
	}
	if refuseViewElementReadsError := lowering.refuseViewElementReads(fields); refuseViewElementReadsError != nil {
		return nil, refuseViewElementReadsError
	}
	if lowering.result.CheckedFields == nil {
		lowering.result.CheckedFields = map[string]bool{}
	}

	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return nil, moduleOrderError
	}
	var refused error
	var getters []*ast.Node
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if refused != nil {
			return true
		}
		// An accessor slot still needs a conversion the checked read boundary can't emit. An
		// optional slot or receiver is emission's to judge, at the read (unemittableCheckedViewRead).
		if part.Kind == ast.KindPropertyAccessExpression && lowering.checkedViewReceiverField(
			part,
			lowering.checker.GetTypeAtLocation(part.AsPropertyAccessExpression().Expression),
			part.Name().Text(),
		) {
			access := part.AsPropertyAccessExpression()
			if base, _ := lowering.representation(
				lowering.checker.GetTypeAtLocation(access.Expression),
			); base == ir.Object {
				field := lowering.checker.GetSymbolAtLocation(part.Name())
				if field != nil && len(
					lowering.checker.GetSignaturesOfType(
						lowering.checker.GetTypeOfSymbol(field),
						checker.SignatureKindCall,
					),
				) == 0 && accessorSymbol(field) {
					refused = lowering.notYet(part, checkedFieldAliasAccessor)
				}
			}
		}
		if part.Kind == ast.KindPropertyAccessExpression && lowering.checkedViewReceiverField(
			part,
			lowering.checker.GetTypeAtLocation(part.AsPropertyAccessExpression().Expression),
			part.Name().Text(),
		) && ast.IsAssignmentTarget(part) {
			if symbol := lowering.checker.GetSymbolAtLocation(part.Name()); symbol != nil {
				of, _ := lowering.representation(lowering.checker.GetTypeOfSymbol(symbol))
				if of == ir.Object {
					refused = lowering.notYet(
						part,
						"writing a checked object field without its source-slot type certificate",
					)
				}
			}
		}
		// A getter named as a contract field is judged by what can reach this view, once lowering
		// is done (refuseReachedContractGetters), never by its name alone (#c7w6j0y).
		if part.Kind == ast.KindGetAccessor && part.Name() != nil && fields[part.Name().Text()] {
			getters = append(getters, part)
		}
		if refused == nil {
			part.ForEachChild(visit)
		}
		return refused != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if refused != nil {
		return nil, refused
	}
	if len(getters) != 0 {
		lowering.contractGetters = append(lowering.contractGetters, contractGetters{origin: value, getters: getters})
	}
	lowering.result.ViewOrigins = append(lowering.result.ViewOrigins, value)
	lowering.graphViewSite = node
	return value, nil
}

const checkedFieldAliasAccessor = "a checked field alias of an accessor"

const checkedFieldAliasNotScalar = "a checked field alias whose field isn't a scalar"

// checkedFieldAliasGap names the first conversion a checked field alias lacks,
// in one order at every site: an optional slot, then an accessor, then a field
// whose representation isn't a plain scalar. Empty means the alias needs none.
func checkedFieldAliasGap(access *ast.PropertyAccessExpression, field *ast.Symbol, scalar bool) string {
	switch {
	case field.Flags&ast.SymbolFlagsOptional != 0 || access.QuestionDotToken != nil:
		return "a checked field alias of an optional field or through ?."
	case accessorSymbol(field):
		return checkedFieldAliasAccessor
	case !scalar:
		return checkedFieldAliasNotScalar
	}
	return ""
}

func interfaceScalar(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if !interfaceScalar(member) {
				return false
			}
		}
		return true
	}
	return proven.Flags()&(checker.TypeFlagsString|checker.TypeFlagsStringLiteral|checker.TypeFlagsNumber|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBoolean|checker.TypeFlagsBooleanLiteral) != 0 &&
		proven.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsIntersection|checker.TypeFlagsTypeParameter) == 0
}

// A finite literal contract is checked as well as its primitive representation.
func (lowering *lowering) viewLiterals(declared *checker.Type) []ir.Expression {
	if declared.Flags()&checker.TypeFlagsUnion != 0 {
		var allowed []ir.Expression
		for _, member := range declared.Types() {
			values := lowering.viewLiterals(member)
			if len(values) == 0 {
				return nil
			}
			allowed = append(allowed, values...)
		}
		return allowed
	}
	value, _, known := lowering.literalConstant(declared)
	if known {
		return []ir.Expression{value}
	}
	return nil
}

func (lowering *lowering) callableViewContract(proven *checker.Type) bool {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range proven.Types() {
			if lowering.callableViewContract(member) {
				return true
			}
		}
	}
	return len(lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)) != 0 ||
		len(lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindConstruct)) != 0
}

func (lowering *lowering) viewSchema(node *ast.Node, target *checker.Type) (map[string]bool, error) {
	lowering.graphViewSite = node
	id, viewContractError := lowering.viewContract(node, target)
	if viewContractError != nil {
		return nil, viewContractError
	}
	fields := map[string]bool{}
	// A failed whole-object adapter can leave an unsupported descriptor without
	// its fields. Keep those names registered so demanded reads reach the lazy
	// refusal instead of silently reading unchecked storage.
	if contract := lowering.result.ViewContracts[id-1]; contract.Unsupported != "" {
		if lowering.result.CheckedFields == nil {
			lowering.result.CheckedFields = map[string]bool{}
		}
		for _, property := range lowering.checker.GetPropertiesOfType(target) {
			fields[property.Name] = true
			// Preserve the boundary under receiver-specific registration too.
			lowering.result.CheckedFields[ir.CheckedFieldKey(int(target.Id()), property.Name)] = true
		}
	}
	seen := map[ir.ViewContractID]bool{}
	var visit func(ir.ViewContractID)
	visit = func(id ir.ViewContractID) {
		if id == 0 || seen[id] {
			return
		}
		seen[id] = true
		contract := lowering.result.ViewContracts[id-1]
		for _, field := range contract.Fields {
			fields[field.Name] = true
			for typeID, contractID := range lowering.result.ViewContractTypes {
				if contractID == id {
					if lowering.result.CheckedFields == nil {
						lowering.result.CheckedFields = map[string]bool{}
					}
					lowering.result.CheckedFields[ir.CheckedFieldKey(typeID, field.Name)] = true
				}
			}
			visit(field.Contract)
		}
		for _, member := range contract.Members {
			visit(member)
		}
	}
	visit(id)
	return fields, nil
}

func (lowering *lowering) legacyView(node *ast.Node, value ir.Expression, target *checker.Type) (ir.Expression, error) {
	if lowering.callableViewContract(target) {
		return nil, &Refused{
			Where: lowering.program.Where(node),
			What:  "a checked view with a callable contract",
			Fix:   "prove the callable body rather than asserting its signature",
		}
	}
	// Diagnose unreifiable contracts before temporary backend limitations.
	for _, property := range lowering.checker.GetPropertiesOfType(target) {
		declared := lowering.checker.GetTypeOfSymbol(property)
		if lowering.callableViewContract(declared) {
			return nil, &Refused{
				Where: lowering.program.Where(node),
				What:  "a checked view with callable field " + property.Name,
				Fix:   "prove the callable body rather than asserting its signature",
			}
		}
	}
	fields := map[string]bool{}
	for _, property := range lowering.checker.GetPropertiesOfType(target) {
		declared := lowering.checker.GetTypeOfSymbol(property)
		if lowering.callableViewContract(declared) {
			return nil, &Refused{
				Where: lowering.program.Where(node),
				What:  "a checked view with callable field " + property.Name,
				Fix:   "prove the callable body rather than asserting its signature",
			}
		}
		of, known := lowering.representation(declared)
		if property.Flags&ast.SymbolFlagsOptional != 0 ||
			!interfaceScalar(declared) ||
			!known ||
			of < ir.Number ||
			of > ir.String {
			return nil, lowering.notYet(
				node,
				"checked view field "+property.Name+" of type "+lowering.checker.TypeToString(declared),
			)
		}
		fields[property.Name] = true
	}
	modules, moduleOrderError := lowering.moduleOrder(lowering.program.Files()[0])
	if moduleOrderError != nil {
		return nil, moduleOrderError
	}
	var found error
	var getters []*ast.Node
	var visit ast.Visitor
	visit = func(part *ast.Node) bool {
		if found != nil {
			return true
		}
		// As view's: judged by reach once lowering is done (refuseReachedContractGetters).
		if part.Kind == ast.KindGetAccessor && part.Name() != nil && fields[part.Name().Text()] {
			getters = append(getters, part)
		}
		if part.Kind == ast.KindObjectBindingPattern {
			for _, binding := range part.AsBindingPattern().Elements.Nodes {
				name := binding.Name()
				if binding.AsBindingElement().PropertyName != nil {
					name = binding.AsBindingElement().PropertyName
				}
				if name != nil && fields[name.Text()] {
					declared := lowering.checker.GetTypeAtLocation(binding.Name())
					of, known := lowering.representation(declared)
					if !known || of < ir.Number || of > ir.String || !interfaceScalar(declared) {
						found = lowering.notYet(
							binding,
							"a checked destructured alias requiring a representation conversion",
						)
					}
				}
			}
		}
		if part.Kind == ast.KindPropertyAccessExpression && lowering.checkedViewReceiverField(
			part,
			lowering.checker.GetTypeAtLocation(part.AsPropertyAccessExpression().Expression),
			part.Name().Text(),
		) {
			access := part.AsPropertyAccessExpression()
			if base, _ := lowering.representation(
				lowering.checker.GetTypeAtLocation(access.Expression),
			); base == ir.Object {
				field := lowering.checker.GetSymbolAtLocation(part.Name())
				if field != nil && len(
					lowering.checker.GetSignaturesOfType(
						lowering.checker.GetTypeOfSymbol(field),
						checker.SignatureKindCall,
					),
				) == 0 {
					of, known := lowering.representation(lowering.checker.GetTypeOfSymbol(field))
					if gap := checkedFieldAliasGap(
						access,
						field,
						interfaceScalar(lowering.checker.GetTypeOfSymbol(field)) &&
							known &&
							of >= ir.Number &&
							of <= ir.String,
					); gap != "" {
						if gap == checkedFieldAliasNotScalar {
							gap += " (" + lowering.checker.TypeToString(lowering.checker.GetTypeOfSymbol(field)) + ")"
						}
						found = lowering.notYet(part, gap)
					}
				}
			}
		}
		if found == nil {
			part.ForEachChild(visit)
		}
		return found != nil
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if found != nil {
		return nil, found
	}
	if refuseViewElementReadsError := lowering.refuseViewElementReads(fields); refuseViewElementReadsError != nil {
		return nil, refuseViewElementReadsError
	}
	if len(getters) != 0 {
		lowering.contractGetters = append(lowering.contractGetters, contractGetters{origin: value, getters: getters})
	}
	if lowering.result.CheckedFields == nil {
		lowering.result.CheckedFields = map[string]bool{}
	}
	for field := range fields {
		lowering.result.CheckedFields[ir.CheckedFieldKey(int(target.Id()), field)] = true
	}
	lowering.graphViewSite = node
	return value, nil
}

// Nullable receivers read the present object's contract. A generic receiver
// uses its instantiated type, or its constraint when the binder remains rigid.
// Neither operation relates unrelated objects merely because their names match.
func (lowering *lowering) viewReceiverTypeID(receiver *checker.Type) int {
	return lowering.viewTypeID(lowering.viewReceiver(receiver))
}

// viewFieldReceiverTypeID is the one place the id a field read is checked under is
// decided, for the read lowering emits and for every guard that asks whether that
// read is checked (record storage's, the alias refusals, a destructured view, a
// view's lie, a whole copy), so a guard can't ask an id the read isn't checked
// under (Lower, Nov 2). A read takes viewReadReceiverTypeID's routed id. An
// assignment target takes the raw viewReceiverTypeID its write uses, the read half
// of a compound or logical assignment included. A nil node is a read with no
// access of its own: a destructured binding, a view's source, a copied value.
// TestViewReceiverCallersAreClassified holds every caller to a role.
func (lowering *lowering) viewFieldReceiverTypeID(node *ast.Node, receiver *checker.Type) int {
	var id int
	if node != nil && ast.IsAssignmentTarget(node) {
		id = lowering.viewReceiverTypeID(receiver)
	} else {
		id = lowering.viewReadReceiverTypeID(receiver)
	}
	// Whether a record's or a merged class's field name checks it (#9jnjwrw).
	lowering.plainReceiver(id)
	return id
}

// viewReadReceiverTypeID is a field read's receiver id, viewReceiverTypeID's but
// with a branded supertype intersected only with its supertypes read as that
// supertype (brandedIntersectionReceiver). Only a read is routed: a write keeps
// its receiver's own id, since a write routed to S's checked write checked the
// slot S's view keys, which a value admitted as S needn't have, and stopped where
// Node wrote (Language's y4, Optimizer's ruling, Nov 2). Only
// viewFieldReceiverTypeID calls it.
func (lowering *lowering) viewReadReceiverTypeID(receiver *checker.Type) int {
	return lowering.viewTypeID(lowering.brandedIntersectionReceiver(lowering.viewReceiver(receiver)))
}

// viewReceiver is the type a receiver's view is keyed by, before its id.
func (lowering *lowering) viewReceiver(receiver *checker.Type) *checker.Type {
	receiver = lowering.concrete(receiver)
	if receiver.Flags()&checker.TypeFlagsTypeParameter != 0 {
		if constraint := lowering.checker.GetBaseConstraintOfType(receiver); constraint != nil {
			receiver = lowering.concrete(constraint)
		}
	}
	return lowering.checker.GetNonNullableType(receiver)
}

// viewTypeID is the id a read's view names a type by, kept with its type in viewTypes.
func (lowering *lowering) viewTypeID(named *checker.Type) int {
	if lowering.viewTypes == nil {
		lowering.viewTypes = map[int]*checker.Type{}
	}
	lowering.viewTypes[int(named.Id())] = named
	return int(named.Id())
}

// A union receiver can carry a checked member's allocation. Preserve that
// boundary for its common field without marking unrelated receiver types. Node
// is the access asked about, or nil for a read with none, and decides the id as
// its read is checked under (viewFieldReceiverTypeID).
func (lowering *lowering) checkedViewReceiverField(node *ast.Node, receiver *checker.Type, name string) bool {
	id := lowering.viewFieldReceiverTypeID(node, receiver)
	if lowering.result.CheckedField(id, name) {
		return true
	}
	receiver = lowering.checker.GetNonNullableType(lowering.concrete(receiver))
	if receiver.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range receiver.Types() {
			if lowering.checkedViewReceiverField(node, member, name) {
				lowering.result.CheckedFields[ir.CheckedFieldKey(id, name)] = true
				return true
			}
		}
	}
	return false
}
