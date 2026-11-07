package lower

import (
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Numeric enums are numbers with named constants. String enums retain declaration identity.
func (lowering *lowering) enumAssignable(from, to *checker.Type) bool {
	if target := enumObjectSymbol(
		to,
	); target != nil && from.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsNever) == 0 {
		return enumObjectSymbol(from) == target
	}
	if to.Flags()&checker.TypeFlagsNumberLiteral != 0 &&
		lowering.openNumericEnumType(from) &&
		!lowering.openNumericEnumType(to) {
		return false
	}
	if to.Flags()&checker.TypeFlagsEnumLike == 0 {
		return true
	}
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range from.Types() {
			if !lowering.enumAssignable(member, to) {
				return false
			}
		}
		return true
	}
	if lowering.numericEnum(lowering.enumIdentity(to)) {
		if !lowering.openNumericEnumType(to) {
			return !lowering.openNumericEnumType(from) &&
				from.Flags()&checker.TypeFlagsNumberLiteral != 0 &&
				lowering.checker.IsTypeAssignableTo(from, to)
		}
		return from.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsNever) != 0
	}
	return from.Flags()&checker.TypeFlagsEnumLike != 0 &&
		lowering.enumIdentity(from) == lowering.enumIdentity(to) &&
		lowering.checker.IsTypeAssignableTo(from, to)
}

// The whole declaration is open; a member-specific literal is a smaller promise. openNumericEnumType is
// whether a type, less null and undefined, is exactly one numeric enum's whole type: its own type, or a
// union naming every member and nothing else. wholeNumericEnums decides it, as it decides
// holdsOpenNumericEnum, so the open-enum tag walk and the arm filter read one definition (#4cfvp4q).
func (lowering *lowering) openNumericEnumType(proven *checker.Type) bool {
	members := flattenedMembers(lowering.checker.GetNonNullableType(proven))
	whole, impure := lowering.wholeNumericEnums(members)
	return !impure && len(whole) == 1 && whole[0].size == len(members)
}

// wholeNumericEnum is a numeric enum all of whose members a union's flattened members hold, with its
// member count.
type wholeNumericEnum struct {
	symbol *ast.Symbol
	size   int
}

// wholeNumericEnums is, over a union's flattened members, each numeric enum all of whose members are
// there: its whole type, which tsc flattens into the enum's own members, inside a wider union or not. A
// member count is the enum's whole type's, so members tsc dedupes by value count once. impure is whether
// some member is an enum-literal number of an enum that isn't purely numeric, which each caller judges
// for itself: the arm filter fails closed on it, and openNumericEnumType answers no.
func (lowering *lowering) wholeNumericEnums(members []*checker.Type) (whole []wholeNumericEnum, impure bool) {
	present := map[*ast.Symbol]int{}
	order := []*ast.Symbol{}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsEnumLiteral == 0 || member.Flags()&checker.TypeFlagsNumberLiteral == 0 {
			continue
		}
		identity := lowering.enumIdentity(member)
		if identity == nil || !lowering.numericEnum(identity) {
			impure = true
			continue
		}
		if present[identity] == 0 {
			order = append(order, identity)
		}
		present[identity]++
	}
	for _, identity := range order {
		size := len(flattenedMembers(lowering.checker.GetTypeAtLocation(identity.ValueDeclaration.Name())))
		if present[identity] >= size {
			whole = append(whole, wholeNumericEnum{symbol: identity, size: size})
		}
	}
	return whole, impure
}

// flattenedMembers is a union's members, or the type alone.
func flattenedMembers(proven *checker.Type) []*checker.Type {
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		return proven.Types()
	}
	return []*checker.Type{proven}
}

func (lowering *lowering) numericEnum(symbol *ast.Symbol) bool {
	if symbol == nil || symbol.ValueDeclaration == nil || symbol.ValueDeclaration.Kind != ast.KindEnumDeclaration {
		return false
	}
	for _, member := range symbol.ValueDeclaration.AsEnumDeclaration().Members.Nodes {
		constant := lowering.checker.GetConstantValue(member)
		if constant == nil || reflect.TypeOf(constant).Kind() != reflect.Float64 {
			return false
		}
	}
	return true
}

// typeof E names the actual enum object and its aliases. Structural copies can omit reverse
// properties or carry hidden heterogeneous fields, so they cannot inherit that exact shape proof.
func enumObjectSymbol(proven *checker.Type) *ast.Symbol {
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	if symbol := proven.Symbol(); symbol != nil && symbol.Flags&ast.SymbolFlagsEnum != 0 {
		return symbol
	}
	return nil
}

func (lowering *lowering) enumObjectView(node *ast.Node) error {
	if !lowering.isExpression(node) || node.Kind == ast.KindParenthesizedExpression {
		return nil
	}
	target := lowering.slotType(node)
	source := lowering.checker.GetTypeAtLocation(node)
	if node.Kind == ast.KindAsExpression {
		target = lowering.checker.GetTypeAtLocation(node)
		source = lowering.checker.GetTypeAtLocation(node.AsAsExpression().Expression)
	}
	if target == nil {
		return nil
	}
	target = lowering.checker.GetNonNullableType(target)
	if enumObjectSymbol(target) != nil && !lowering.enumAssignable(source, target) {
		return &Refused{
			Where: lowering.program.Where(node),
			What: "a structural object used as " +
				lowering.checker.TypeToString(target) +
				"; reverse properties and the complete enum shape are unproven",
			Fix: "use the enum's runtime object or its typeof alias; give an ordinary object an explicit interface",
		}
	}
	return nil
}

func (lowering *lowering) enumIdentity(proven *checker.Type) *ast.Symbol {
	symbol := proven.Symbol()
	if symbol != nil && symbol.Flags&ast.SymbolFlagsEnumMember != 0 && symbol.ValueDeclaration != nil {
		return lowering.symbol(symbol.ValueDeclaration.Parent.Name())
	}
	if symbol != nil && symbol.Flags&ast.SymbolFlagsEnum != 0 {
		return symbol
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 && len(proven.Types()) > 0 {
		return lowering.enumIdentity(proven.Types()[0])
	}
	return nil
}

func (lowering *lowering) enumMember(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression &&
		node.Kind != ast.KindElementAccessExpression &&
		node.Kind != ast.KindIdentifier {
		return nil
	}
	symbol := lowering.symbol(node)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsEnumMember != 0 {
		return symbol.ValueDeclaration
	}
	return nil
}

func (lowering *lowering) enumObject(node *ast.Node) *ast.Node {
	proven := lowering.checker.GetTypeAtLocation(node)
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	symbol := proven.Symbol()
	if symbol == nil || symbol.Flags&ast.SymbolFlagsEnum == 0 {
		return nil
	}
	return symbol.ValueDeclaration
}

func (lowering *lowering) enumConstant(member *ast.Node) (ir.Expression, error) {
	constant := lowering.checker.GetConstantValue(member)
	if constant != nil && reflect.TypeOf(constant).Kind() == reflect.Float64 {
		value := reflect.ValueOf(constant).Float()
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, lowering.notYet(member, "a non-finite enum member")
		}
		return ir.NumberConstant{Value: value}, nil
	}
	if value, ok := constant.(string); ok {
		return ir.StringConstant{Index: lowering.constant(value)}, nil
	}
	return nil, lowering.notYet(member, "a computed enum member; use a constant number or string initializer")
}

// Build the same forward and reverse assignments as JavaScript, with the last numeric alias
// winning. A string member adds only its forward property. No runtime operation can mutate an
// enum's member slots through a structural writable view (the invariance pass checks those).
func (lowering *lowering) enumFields(node *ast.Node) ([]ir.Field, error) {
	fields := []ir.Field{}
	positions := map[string]int{}
	set := func(name string, value ir.Expression) {
		if position, exists := positions[name]; exists {
			fields[position].Value = value
			return
		}
		positions[name] = len(fields)
		fields = append(fields, ir.Field{Name: name, Value: value})
	}
	for _, member := range node.AsEnumDeclaration().Members.Nodes {
		if member.Name().Text() == "__proto__" || strings.ContainsRune(member.Name().Text(), 0) {
			return nil, &Refused{
				Where: lowering.program.Where(member),
				What:  "an enum member name that changes the prototype or contains NUL",
				Fix:   "rename the member; enum properties must be ordinary fixed slots",
			}
		}
		value, enumConstantError := lowering.enumConstant(member)
		if enumConstantError != nil {
			return nil, enumConstantError
		}
		if value.Type() == ir.Number && (member.Name().Text() == "NaN" ||
			member.Name().Text() == "Infinity" ||
			member.Name().Text() == "-Infinity") {
			return nil, &Refused{
				Where: lowering.program.Where(member),
				What:  "a numeric enum member named like a non-finite number; numeric indexing could read a number where tsc promises a string",
				Fix:   "rename the member so numeric reverse mapping stays string-valued",
			}
		}
		set(member.Name().Text(), value)
		if number, numeric := value.(ir.NumberConstant); numeric {
			key := strconv.FormatFloat(number.Value, 'f', -1, 64)
			if printed, ok := lowering.checker.GetConstantValue(member).(interface{ String() string }); ok {
				key = printed.String()
			}
			if number.Value == 0 {
				key = "0"
			}
			set(key, ir.StringConstant{Index: lowering.constant(member.Name().Text())})
		}
	}
	return fields, nil
}

// In a declaration the enum name denotes its member type; in value position it denotes the
// runtime object. Declare the latter explicitly instead of using the name node's number type.
func (lowering *lowering) enumLocal(node *ast.Node) (int, error) {
	local, declareLocalError := lowering.declareLocal(node.Name())
	if declareLocalError != nil {
		return 0, declareLocalError
	}
	lowering.result.Locals[local].Type = ir.Object
	lowering.noteLocal(local, lowering.checker.GetTypeOfSymbol(lowering.symbol(node.Name())), node.Name())
	return local, nil
}

// enumObjectBuilt reports whether an enum declaration has a runtime object: every ordinary enum, and
// a const enum the program reaches as a value (constEnumReached).
func (lowering *lowering) enumObjectBuilt(declaration *ast.Node) (bool, error) {
	if !ast.HasSyntacticModifier(declaration, ast.ModifierFlagsConst) {
		return true, nil
	}
	return lowering.constEnumReached(declaration)
}

// constEnumReached reports whether the program reaches a const enum's object as a value. Node builds
// every const enum's object, and Lower builds one exactly where it can be observed: the same
// ObjectLiteral an ordinary enum gets, with member reads still inlined. Nothing reached keeps
// rulings.md:218, no object and no binding (Adamic's ruling on #52zja36, Nov 6, gated on reach alone,
// since 0.1's options are fixed in load.go and have no preserveConstEnums).
func (lowering *lowering) constEnumReached(declaration *ast.Node) (bool, error) {
	// One declared inside a function or a block (#xpjcdpf) is no namespace's export, and tsc refuses
	// every other value use of it, so nothing reaches it; an object there would be made each time its
	// block runs, which no enum's is.
	if declaration.Parent.Kind != ast.KindSourceFile && declaration.Parent.Kind != ast.KindModuleBlock {
		return false, nil
	}
	// A lowering with no entry, as the latent census makes one per statement, has no program to walk
	// for reach, so it fails closed to built.
	if lowering.entry == nil {
		return true, nil
	}
	if lowering.constEnumReach == nil {
		reached, reachedConstEnumsError := lowering.reachedConstEnums()
		if reachedConstEnumsError != nil {
			return false, reachedConstEnumsError
		}
		lowering.constEnumReach = reached
	}
	return lowering.constEnumReach[lowering.symbol(declaration.Name())], nil
}

// reachedConstEnums is every const enum the program reaches as a value, found in one walk of the
// modules the entry's lowering lowers (moduleOrder, as lowerCheckedWithCycleTypes orders them), so
// the walk sees every read that lowering will place. tsc refuses every direct value use of a const
// enum (TS2475), so the road that reaches one is a namespace object, like tsc's own (ts as
// ...).SyntaxKind. Reach fails closed: a read of a const enum other than a member read lowering
// inlines reaches it, and any use of a namespace other than a qualified read reaches every const enum
// it exports, unless it's the namespace cast, which reaches the one export it names.
func (lowering *lowering) reachedConstEnums() (map[*ast.Symbol]bool, error) {
	modules, moduleOrderError := lowering.moduleOrder(lowering.entry)
	if moduleOrderError != nil {
		return nil, moduleOrderError
	}
	reached := map[*ast.Symbol]bool{}
	constEnum := func(symbol *ast.Symbol) bool {
		return symbol != nil &&
			symbol.ValueDeclaration != nil &&
			symbol.ValueDeclaration.Kind == ast.KindEnumDeclaration &&
			ast.HasSyntacticModifier(symbol.ValueDeclaration, ast.ModifierFlagsConst)
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if (ast.IsIdentifier(node) || node.Kind == ast.KindPropertyAccessExpression) &&
			lowering.namespaceValueNode(node) {
			if symbol := lowering.symbol(node); constEnum(symbol) && !lowering.enumMemberReadOf(node) {
				reached[symbol] = true
			}
			if lowering.moduleNamespace(node) && !lowering.qualifiedNamespaceRead(node) {
				if access := lowering.namespaceCastAccess(node); access != nil {
					if target := lowering.namespaceExport(node, access.Name().Text()); constEnum(target) {
						reached[target] = true
					}
				} else {
					for _, exported := range lowering.namespaceExportSymbols(node) {
						if target := lowering.resolvedExport(exported); constEnum(target) {
							reached[target] = true
						}
					}
				}
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return reached, nil
}

// enumMemberReadOf reports whether node, a read of an enum, is the receiver of one of its member
// reads, E.A or E["A"], which lowering inlines for a const enum.
func (lowering *lowering) enumMemberReadOf(node *ast.Node) bool {
	parent := node.Parent
	return parent != nil &&
		(parent.Kind == ast.KindPropertyAccessExpression || parent.Kind == ast.KindElementAccessExpression) &&
		parent.Expression() == node &&
		lowering.enumMember(parent) != nil
}

func (lowering *lowering) enumDeclaration(node *ast.Node) ([]ir.Statement, error) {
	built, enumObjectBuiltError := lowering.enumObjectBuilt(node)
	if enumObjectBuiltError != nil || !built {
		return nil, enumObjectBuiltError
	}
	fields, enumFieldsError := lowering.enumFields(node)
	if enumFieldsError != nil {
		return nil, enumFieldsError
	}
	local, enumLocalError := lowering.enumLocal(node)
	if enumLocalError != nil {
		return nil, enumLocalError
	}
	return []ir.Statement{ir.Declare{Local: local, Value: ir.ObjectLiteral{Fields: fields}}}, nil
}

func (lowering *lowering) enumExpression(node *ast.Node) (ir.Expression, bool, error) {
	member := lowering.enumMember(node)
	if member != nil &&
		node.Kind == ast.KindElementAccessExpression &&
		!ast.HasSyntacticModifier(member.Parent, ast.ModifierFlagsConst) {
		index := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if index.Kind != ast.KindStringLiteral && index.Kind != ast.KindNoSubstitutionTemplateLiteral {
			member = nil
		}
	}
	if member != nil {
		declaration := member.Parent
		if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsConst) {
			value, enumConstantError := lowering.enumConstant(member)
			if enumConstantError == nil {
				value = lowering.namespaceReadyValue(node, value)
			}
			return value, true, enumConstantError
		}
		var receiver *ast.Node
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			receiver = node.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			receiver = node.AsElementAccessExpression().Expression
		default:
			return nil, true, lowering.notYet(node, "an unqualified enum member outside its initializer")
		}
		object, expressionError := lowering.expression(receiver)
		if expressionError != nil {
			return nil, true, expressionError
		}
		of, typeOfError := lowering.typeOf(member.Name())
		optional := false
		if node.Kind == ast.KindPropertyAccessExpression {
			optional = node.AsPropertyAccessExpression().QuestionDotToken != nil
		}
		if node.Kind == ast.KindElementAccessExpression {
			optional = node.AsElementAccessExpression().QuestionDotToken != nil
		}
		if !optional && node.Flags&ast.NodeFlagsOptionalChain != 0 {
			return nil, true, lowering.notYet(node, "an optional enum chain longer than one step")
		}
		return ir.Property{Object: object, Name: member.Name().Text(), Of: of, Optional: optional}, true, typeOfError
	}
	if node.Kind != ast.KindElementAccessExpression {
		return nil, false, nil
	}
	access := node.AsElementAccessExpression()
	declaration := lowering.enumObject(access.Expression)
	if declaration == nil {
		return nil, false, nil
	}
	fields, enumFieldsError := lowering.enumFields(declaration)
	if enumFieldsError != nil {
		return nil, true, enumFieldsError
	}
	object, expressionError := lowering.expression(access.Expression)
	if expressionError != nil {
		return nil, true, expressionError
	}
	key, expressionError := lowering.expression(access.ArgumentExpression)
	if expressionError != nil {
		return nil, true, expressionError
	}
	if key.Type() != ir.Number && key.Type() != ir.String {
		return nil, true, lowering.notYet(node, "an enum index other than number or string")
	}
	of, typeOfError := lowering.typeOf(node)
	if typeOfError != nil {
		return nil, true, typeOfError
	}
	// A helper evaluates receiver and key exactly once, including a key with side effects. It
	// uses ordinary IR so flow, ownership, regions and reuse all see its reads and returns.
	index := len(lowering.result.Functions)
	objectLocal := len(lowering.result.Locals)
	keyLocal := objectLocal + 1
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "enum_object", Type: ir.Object, Function: index},
		ir.Local{Name: "enum_key", Type: key.Type(), Function: index},
	)
	readObject := ir.Read{Local: objectLocal, Of: ir.Object}
	readKey := ir.Read{Local: keyLocal, Of: key.Type()}
	var tested ir.Expression = readKey
	if key.Type() == ir.Number {
		tested = ir.NumberToString{Value: readKey}
	}
	function := ir.Function{Name: "enum_lookup", Parameters: []int{objectLocal, keyLocal}, Returns: of}
	for _, field := range fields {
		if field.Value.Type() != of.Present() {
			continue
		}
		function.Body = append(function.Body, ir.If{
			Condition: ir.Binary{
				Operator: ir.Equal,
				Left:     tested,
				Right:    ir.StringConstant{Index: lowering.constant(field.Name)},
			},
			Then: []ir.Statement{
				ir.Return{Value: fit(ir.Property{Object: readObject, Name: field.Name, Of: field.Value.Type()}, of)},
			},
		})
	}
	if !of.IsReference() && !of.IsMaybe() {
		keys := lowering.checker.GetTypeAtLocation(access.ArgumentExpression)
		members := []*checker.Type{keys}
		if keys.Flags()&checker.TypeFlagsUnion != 0 {
			members = keys.Types()
		}
		for _, member := range members {
			if member.Flags()&checker.TypeFlagsStringLiteral == 0 {
				return nil, true, lowering.notYet(node, "an enum lookup without a proven member name")
			}
			name, ok := member.AsLiteralType().Value().(string)
			found := false
			for _, field := range fields {
				if ok && field.Name == name && field.Value.Type() == of {
					found = true
				}
			}
			if !found {
				return nil, true, lowering.notYet(node, "an enum lookup without a proven member name")
			}
		}
		function.Body = append(
			function.Body,
			ir.Panic{Message: ir.StringConstant{Index: lowering.constant("enum member name outside its proven union")}},
		)
	} else {
		function.Body = append(function.Body, ir.Return{Value: ir.Undefined{Of: of}})
	}
	lowering.result.Functions = append(lowering.result.Functions, function)
	return ir.Call{Function: index, Arguments: []ir.Expression{object, key}, Returns: of}, true, nil
}

func (lowering *lowering) enumRefusal(node *ast.Node) error {
	if lowering.isExpression(node) {
		if symbol := lowering.flagValueSymbol(node); symbol != nil {
			declared := lowering.checker.GetTypeOfSymbol(symbol)
			observed := lowering.checker.GetTypeAtLocation(node)
			// A refinement can land on an intersection member or a smaller union of members, so
			// every member that isn't a primitive counts, and so does any refined shape that is.
			objectMembers := 0
			if declared.Flags()&checker.TypeFlagsUnion != 0 {
				for _, member := range declared.Types() {
					if member.Flags()&checker.TypeFlagsPrimitive == 0 {
						objectMembers++
					}
				}
			}
			if objectMembers > 1 &&
				declared.Flags()&checker.TypeFlagsUnion != 0 &&
				observed != declared &&
				refinedToObject(observed) {
				// A union whose tags can lie is admitted only where every step that refined this read
				// is a comparison its incoming type proves (enum_tag_flow.go).
				hazard := lowering.enumTagHazardOf(declared, "")
				// A property's refused read is its name, which has no flow node: the walk follows the
				// access the checker narrows (#qeanycw).
				reference := node
				if parent := node.Parent; parent != nil &&
					parent.Kind == ast.KindPropertyAccessExpression &&
					parent.AsPropertyAccessExpression().Name() == node {
					reference = parent
				}
				if hazard != enumTagNoHazard && !lowering.enumTagRefinementProven(reference, declared) {
					if hazard == enumTagSharedValueHazard {
						return &Refused{
							Where: lowering.program.Where(node),
							What:  "an object refinement by tags from different numeric enums that share a value",
							Fix:   "give the members tags from one enum, or tags with distinct values (adamic/enum-tag)",
						}
					}
					return &Refused{
						Where: lowering.program.Where(node),
						What:  "an object refinement using an open numeric enum as a literal tag",
						Fix:   "use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)",
					}
				}
			}
		}
	}

	if enumObjectViewError := lowering.enumObjectView(node); enumObjectViewError != nil {
		return enumObjectViewError
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && len(call.Arguments.Nodes) > 0 {
			access := callee.AsPropertyAccessExpression()
			if access.Name().Text() == "assign" &&
				lowering.isLibraryGlobal(access.Expression, "Object") &&
				lowering.enumObject(call.Arguments.Nodes[0]) != nil {
				return &Refused{
					Where: lowering.program.Where(node),
					What:  "a write into an enum runtime object; named constants and reverse lookup must stay consistent",
					Fix:   "copy the enum into an ordinary object before writing it (adamic/enum-object)",
				}
			}
		}
	}
	if node.Kind == ast.KindEnumDeclaration {
		if ambientDeclaration(node) {
			return nil
		}
		// A const enum makes no runtime object anywhere: enumDeclaration lowers it to nothing, every
		// member read folds to the checker's constant by the member's symbol (enumConstant), and tsc
		// refuses any other value use of it (TS2475). So one inside a function or block, as debug.ts's
		// formatControlFlowGraph declares BoxCharacter and Connection, lowers as one at module scope
		// does (#xpjcdpf). A runtime enum there would need its object made each time its block runs.
		if node.Parent.Kind != ast.KindSourceFile &&
			node.Parent.Kind != ast.KindModuleBlock &&
			!ast.HasSyntacticModifier(node, ast.ModifierFlagsConst) {
			return lowering.notYet(node, "an enum inside a function or block; declare it at module scope")
		}
		symbol := lowering.symbol(node.Name())
		if symbol != nil && len(symbol.Declarations) != 1 {
			return lowering.notYet(node, "merged enum declarations; put the members in one declaration")
		}
		// Signed masks, including 1 << 31, are ordinary numeric enum constants.
		_, enumFieldsError := lowering.enumFields(node)
		return enumFieldsError
	}
	var updated *ast.Node
	switch node.Kind {
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
			updated = unary.Operand
		}
	case ast.KindPostfixUnaryExpression:
		updated = node.AsPostfixUnaryExpression().Operand
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) && binary.OperatorToken.Kind != ast.KindEqualsToken {
			updated = binary.Left
		}
	}
	if updated != nil {
		if slot := lowering.literalUpdateSlot(updated); slot != nil {
			return &Refused{
				Where: lowering.program.Where(node),
				What: "an unproven value assigned to a numeric literal or enum member slot " +
					lowering.checker.TypeToString(slot),
				Fix: "assign this slot's literal constant, or widen the slot to the whole numeric enum or number (adamic/enum-literal)",
			}
		}
	}
	if updated != nil &&
		lowering.checker.GetTypeAtLocation(updated).Flags()&checker.TypeFlagsEnumLike != 0 &&
		!lowering.numericEnum(lowering.enumIdentity(lowering.checker.GetTypeAtLocation(updated))) {
		if lowering.flagUpdate(node, updated) {
			return nil
		}
		if lowering.flagEnum(lowering.enumIdentity(lowering.checker.GetTypeAtLocation(updated))) {
			return lowering.flagWriteRefusal(node, lowering.checker.GetTypeAtLocation(updated))
		}
		return &Refused{
			Where: lowering.program.Where(node),
			What:  "arithmetic assigned back into an enum; the result need not be one of its members",
			Fix:   "assign a declared member, or keep arithmetic results in a number (adamic/enum-members)",
		}
	}
	return nil
}

func init() { refusalRules.register(1800, "enumRefusal", (*lowering).enumRefusal) }

// An open enum field is a hazard only as a tag: two members could hold the same number while the
// checker excludes one of them. A field of one type in every member, like tsc's flags: NodeFlags,
// can't tell members apart, so no refinement rests on it. Only primitive members are passed over;
// an intersection or anything else must carry the same field too.
func sharedByEveryMember(typeChecker *checker.Checker, union *checker.Type, name string, tag *checker.Type) bool {
	for _, member := range union.Types() {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		field := typeChecker.GetPropertyOfType(member, name)
		if field == nil || typeChecker.GetTypeOfSymbol(field) != tag {
			return false
		}
	}
	return true
}

// The checker never compares members of two enums, so it can exclude a member tagged Other.X
// when the value is Kind.A, though both are 0 at run time. Members whose same-named tags come
// from different enums with one value can't be told apart by that tag.
func (lowering *lowering) enumTagsShareValue(union *checker.Type, only string) bool {
	// Each field and value keeps every enum it came from, so a member tagged Kind.A | Other.X still
	// meets another member's Kind.A. Two distinct members per enum are enough to know one of them
	// differs from any later member, which keeps tsc's 300-member node unions linear.
	tags := map[string]map[any]map[*ast.Symbol][]int{}
	for index, member := range union.Types() {
		if member.Flags()&checker.TypeFlagsPrimitive != 0 {
			continue
		}
		for _, field := range lowering.checker.GetPropertiesOfType(member) {
			if only != "" && field.Name != only {
				continue
			}
			tag := lowering.checker.GetTypeOfSymbol(field)
			parts := []*checker.Type{tag}
			if tag.Flags()&checker.TypeFlagsUnion != 0 {
				parts = tag.Types()
			}
			for _, part := range parts {
				if part.Flags()&checker.TypeFlagsNumberLiteral == 0 || part.Flags()&checker.TypeFlagsEnumLiteral == 0 {
					continue
				}
				identity := lowering.enumIdentity(part)
				value := part.AsLiteralType().Value()
				if tags[field.Name] == nil {
					tags[field.Name] = map[any]map[*ast.Symbol][]int{}
				}
				if tags[field.Name][value] == nil {
					tags[field.Name][value] = map[*ast.Symbol][]int{}
				}
				for other, earlierMembers := range tags[field.Name][value] {
					for _, earlier := range earlierMembers {
						if other != identity && earlier != index {
							return true
						}
					}
				}
				members := tags[field.Name][value][identity]
				if len(members) < 2 && (len(members) == 0 || members[0] != index) {
					tags[field.Name][value][identity] = append(members, index)
				}
			}
		}
	}
	return false
}

// A refined type reaches an object when it is one, or is a union with a member that is one.
func refinedToObject(refined *checker.Type) bool {
	if refined.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range refined.Types() {
			if refinedToObject(member) {
				return true
			}
		}
		return false
	}
	return refined.Flags()&(checker.TypeFlagsPrimitive|checker.TypeFlagsNever) == 0
}

// A compound or update write computes a number, so it can't keep a member literal's promise. Numeric
// enums are open, so writes back into a whole one stay admitted; a slot that is one member, or a
// numeric literal, is refused. A field is judged in every member of its receiver's type, since the
// union's own property for kind: Kind.Text | kind: Kind.Count is the whole open enum.
func (lowering *lowering) literalUpdateSlot(updated *ast.Node) *checker.Type {
	updated = ast.SkipParentheses(updated)
	slots := []*checker.Type{}
	if updated.Kind == ast.KindPropertyAccessExpression || updated.Kind == ast.KindElementAccessExpression {
		var receiver *ast.Node
		if updated.Kind == ast.KindPropertyAccessExpression {
			receiver = updated.AsPropertyAccessExpression().Expression
		} else {
			receiver = updated.AsElementAccessExpression().Expression
		}
		field := lowering.symbol(updated)
		if field == nil {
			return nil
		}
		owner := lowering.checker.GetTypeAtLocation(receiver)
		members := []*checker.Type{owner}
		if owner.Flags()&checker.TypeFlagsUnion != 0 {
			members = owner.Types()
		}
		for _, member := range members {
			if member.Flags()&checker.TypeFlagsPrimitive != 0 {
				continue
			}
			if property := lowering.checker.GetPropertyOfType(member, field.Name); property != nil {
				slots = append(slots, lowering.checker.GetTypeOfSymbol(property))
			}
		}
	} else if symbol := lowering.symbol(updated); symbol != nil {
		slots = append(slots, lowering.checker.GetTypeOfSymbol(symbol))
	}
	for _, slot := range slots {
		slot = lowering.checker.GetNonNullableType(slot)
		if (slot.Flags()&checker.TypeFlagsNumberLiteral != 0 || lowering.numericEnum(lowering.enumIdentity(slot))) &&
			!lowering.openNumericEnumType(slot) {
			return slot
		}
	}
	return nil
}

// Compare runtime values, not spellings: numeric aliases denote the same union member. A default
// handles all remaining values. A void switch still needs coverage, even when tsc accepts it.
func (lowering *lowering) enumSwitch(node *ast.Node) error {
	statement := node.AsSwitchStatement()
	proven := lowering.checker.GetTypeAtLocation(statement.Expression)
	if proven.Flags()&checker.TypeFlagsEnumLike == 0 || lowering.numericEnum(lowering.enumIdentity(proven)) {
		return nil
	}
	covered := map[any]bool{}
	for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
		if clause.Kind == ast.KindDefaultClause {
			return nil
		}
		test := clause.AsCaseOrDefaultClause().Expression
		if member := lowering.enumMember(test); member != nil {
			covered[lowering.checker.GetConstantValue(member)] = true
		} else if value := lowering.checker.GetTypeAtLocation(test); value.Flags()&checker.TypeFlagsLiteral != 0 {
			covered[value.AsLiteralType().Value()] = true
		}
	}
	if lowering.flagEnum(lowering.enumIdentity(proven)) {
		return &Refused{
			Where: lowering.program.Where(node),
			What:  "a flag-enum switch without a default",
			Fix:   "add a default for combinations and zero (adamic/enum-flags)",
		}
	}
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsLiteral == 0 {
			return lowering.notYet(node, "a switch on a computed enum member")
		}
		if !covered[member.AsLiteralType().Value()] {
			return &Refused{
				Where: lowering.program.Where(node),
				What:  "a non-exhaustive enum switch; missing " + lowering.checker.TypeToString(member),
				Fix:   "add the missing member's case or a default",
			}
		}
	}
	return nil
}

// A numeric enum default is unreachable only with a member origin proof and full coverage.
// This permits the never-default idiom without inventing a machine representation for never.
func (lowering *lowering) enumDefaultUnreachable(node *ast.Node) bool {
	statement := node.AsSwitchStatement()
	proven := lowering.checker.GetTypeAtLocation(statement.Expression)
	return lowering.enumSwitchCovered(node) && (!lowering.numericEnum(lowering.enumIdentity(proven)) ||
		lowering.enumMemberOrigin(statement.Expression, lowering.enumIdentity(proven), map[*ast.Node]bool{}))
}

func (lowering *lowering) enumSwitchCovered(node *ast.Node) bool {
	statement := node.AsSwitchStatement()
	proven := lowering.checker.GetTypeAtLocation(statement.Expression)
	if proven.Flags()&checker.TypeFlagsEnumLike == 0 {
		return false
	}
	covered := map[any]bool{}
	for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
		if clause.Kind == ast.KindDefaultClause {
			continue
		}
		test := clause.AsCaseOrDefaultClause().Expression
		if member := lowering.enumMember(test); member != nil {
			covered[lowering.checker.GetConstantValue(member)] = true
		} else if value := lowering.checker.GetTypeAtLocation(test); value.Flags()&checker.TypeFlagsLiteral != 0 {
			covered[value.AsLiteralType().Value()] = true
		}
	}
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsLiteral == 0 || !covered[member.AsLiteralType().Value()] {
			return false
		}
	}
	return true
}
