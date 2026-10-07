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
func (l *lowering) enumAssignable(from, to *checker.Type) bool {
	if target := enumObjectSymbol(to); target != nil && from.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsNever) == 0 {
		return enumObjectSymbol(from) == target
	}
	if to.Flags()&checker.TypeFlagsNumberLiteral != 0 && l.openNumericEnumType(from) && !l.openNumericEnumType(to) {
		return false
	}
	if to.Flags()&checker.TypeFlagsEnumLike == 0 {
		return true
	}
	if from.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range from.Types() {
			if !l.enumAssignable(member, to) {
				return false
			}
		}
		return true
	}
	if l.numericEnum(l.enumIdentity(to)) {
		if !l.openNumericEnumType(to) {
			return !l.openNumericEnumType(from) && from.Flags()&checker.TypeFlagsNumberLiteral != 0 && l.checker.IsTypeAssignableTo(from, to)
		}
		return from.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsNever) != 0
	}
	return from.Flags()&checker.TypeFlagsEnumLike != 0 && l.enumIdentity(from) == l.enumIdentity(to) && l.checker.IsTypeAssignableTo(from, to)
}

// The whole declaration is open; a member-specific literal is a smaller promise.
func (l *lowering) openNumericEnumType(proven *checker.Type) bool {
	proven = l.checker.GetNonNullableType(proven)
	identity := l.enumIdentity(proven)
	if !l.numericEnum(identity) {
		return false
	}
	whole := l.checker.GetTypeAtLocation(identity.ValueDeclaration.Name())
	if proven == whole {
		return true
	}
	if proven.Flags()&checker.TypeFlagsUnion == 0 || whole.Flags()&checker.TypeFlagsUnion == 0 || len(proven.Types()) != len(whole.Types()) {
		return false
	}
	for _, member := range whole.Types() {
		found := false
		for _, part := range proven.Types() {
			if part == member {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (l *lowering) numericEnum(symbol *ast.Symbol) bool {
	if symbol == nil || symbol.ValueDeclaration == nil || symbol.ValueDeclaration.Kind != ast.KindEnumDeclaration {
		return false
	}
	for _, member := range symbol.ValueDeclaration.AsEnumDeclaration().Members.Nodes {
		constant := l.checker.GetConstantValue(member)
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

func (l *lowering) enumObjectView(node *ast.Node) error {
	if !l.isExpression(node) || node.Kind == ast.KindParenthesizedExpression {
		return nil
	}
	target := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	source := l.checker.GetTypeAtLocation(node)
	if node.Kind == ast.KindAsExpression {
		target = l.checker.GetTypeAtLocation(node)
		source = l.checker.GetTypeAtLocation(node.AsAsExpression().Expression)
	}
	if target == nil {
		return nil
	}
	target = l.checker.GetNonNullableType(target)
	if enumObjectSymbol(target) != nil && !l.enumAssignable(source, target) {
		return &Refused{Where: l.program.Where(node), What: "a structural object used as " + l.checker.TypeToString(target) + "; reverse properties and the complete enum shape are unproven", Fix: "use the enum's runtime object or its typeof alias; give an ordinary object an explicit interface"}
	}
	return nil
}

func (l *lowering) enumIdentity(proven *checker.Type) *ast.Symbol {
	symbol := proven.Symbol()
	if symbol != nil && symbol.Flags&ast.SymbolFlagsEnumMember != 0 && symbol.ValueDeclaration != nil {
		return l.symbol(symbol.ValueDeclaration.Parent.Name())
	}
	if symbol != nil && symbol.Flags&ast.SymbolFlagsEnum != 0 {
		return symbol
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 && len(proven.Types()) > 0 {
		return l.enumIdentity(proven.Types()[0])
	}
	return nil
}

func (l *lowering) enumMember(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression && node.Kind != ast.KindElementAccessExpression && node.Kind != ast.KindIdentifier {
		return nil
	}
	symbol := l.symbol(node)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsEnumMember != 0 {
		return symbol.ValueDeclaration
	}
	return nil
}

func (l *lowering) enumObject(node *ast.Node) *ast.Node {
	proven := l.checker.GetTypeAtLocation(node)
	if proven.Flags()&checker.TypeFlagsObject == 0 {
		return nil
	}
	symbol := proven.Symbol()
	if symbol == nil || symbol.Flags&ast.SymbolFlagsEnum == 0 {
		return nil
	}
	return symbol.ValueDeclaration
}

func (l *lowering) enumConstant(member *ast.Node) (ir.Expression, error) {
	constant := l.checker.GetConstantValue(member)
	if constant != nil && reflect.TypeOf(constant).Kind() == reflect.Float64 {
		value := reflect.ValueOf(constant).Float()
		if math.IsNaN(value) || math.IsInf(value, 0) {
			return nil, l.notYet(member, "a non-finite enum member")
		}
		return ir.NumberConstant{Value: value}, nil
	}
	if value, ok := constant.(string); ok {
		return ir.StringConstant{Index: l.constant(value)}, nil
	}
	return nil, l.notYet(member, "a computed enum member; use a constant number or string initializer")
}

// Build the same forward and reverse assignments as JavaScript, with the last numeric alias
// winning. A string member adds only its forward property. No runtime operation can mutate an
// enum's member slots through a structural writable view (the invariance pass checks those).
func (l *lowering) enumFields(node *ast.Node) ([]ir.Field, error) {
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
			return nil, &Refused{Where: l.program.Where(member), What: "an enum member name that changes the prototype or contains NUL", Fix: "rename the member; enum properties must be ordinary fixed slots"}
		}
		value, err := l.enumConstant(member)
		if err != nil {
			return nil, err
		}
		if value.Type() == ir.Number && (member.Name().Text() == "NaN" || member.Name().Text() == "Infinity" || member.Name().Text() == "-Infinity") {
			return nil, &Refused{Where: l.program.Where(member), What: "a numeric enum member named like a non-finite number; numeric indexing could read a number where tsc promises a string", Fix: "rename the member so numeric reverse mapping stays string-valued"}
		}
		set(member.Name().Text(), value)
		if number, numeric := value.(ir.NumberConstant); numeric {
			key := strconv.FormatFloat(number.Value, 'f', -1, 64)
			if printed, ok := l.checker.GetConstantValue(member).(interface{ String() string }); ok {
				key = printed.String()
			}
			if number.Value == 0 {
				key = "0"
			}
			set(key, ir.StringConstant{Index: l.constant(member.Name().Text())})
		}
	}
	return fields, nil
}

// In a declaration the enum name denotes its member type; in value position it denotes the
// runtime object. Declare the latter explicitly instead of using the name node's number type.
func (l *lowering) enumLocal(node *ast.Node) (int, error) {
	local, err := l.declareLocal(node.Name())
	if err != nil {
		return 0, err
	}
	l.result.Locals[local].Type = ir.Object
	l.noteLocal(local, l.checker.GetTypeOfSymbol(l.symbol(node.Name())), node.Name())
	return local, nil
}

func (l *lowering) enumDeclaration(node *ast.Node) ([]ir.Statement, error) {
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsConst) {
		return nil, nil
	}
	fields, err := l.enumFields(node)
	if err != nil {
		return nil, err
	}
	local, err := l.enumLocal(node)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Declare{Local: local, Value: ir.ObjectLiteral{Fields: fields}}}, nil
}

func (l *lowering) enumExpression(node *ast.Node) (ir.Expression, bool, error) {
	member := l.enumMember(node)
	if member != nil && node.Kind == ast.KindElementAccessExpression && !ast.HasSyntacticModifier(member.Parent, ast.ModifierFlagsConst) {
		index := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if index.Kind != ast.KindStringLiteral && index.Kind != ast.KindNoSubstitutionTemplateLiteral {
			member = nil
		}
	}
	if member != nil {
		declaration := member.Parent
		if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsConst) {
			value, err := l.enumConstant(member)
			return value, true, err
		}
		var receiver *ast.Node
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			receiver = node.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			receiver = node.AsElementAccessExpression().Expression
		default:
			return nil, true, l.notYet(node, "an unqualified enum member outside its initializer")
		}
		object, err := l.expression(receiver)
		if err != nil {
			return nil, true, err
		}
		of, err := l.typeOf(member.Name())
		optional := false
		if node.Kind == ast.KindPropertyAccessExpression {
			optional = node.AsPropertyAccessExpression().QuestionDotToken != nil
		}
		if node.Kind == ast.KindElementAccessExpression {
			optional = node.AsElementAccessExpression().QuestionDotToken != nil
		}
		if !optional && node.Flags&ast.NodeFlagsOptionalChain != 0 {
			return nil, true, l.notYet(node, "an optional enum chain longer than one step")
		}
		return ir.Property{Object: object, Name: member.Name().Text(), Of: of, Optional: optional}, true, err
	}
	if node.Kind != ast.KindElementAccessExpression {
		return nil, false, nil
	}
	access := node.AsElementAccessExpression()
	declaration := l.enumObject(access.Expression)
	if declaration == nil {
		return nil, false, nil
	}
	fields, err := l.enumFields(declaration)
	if err != nil {
		return nil, true, err
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	key, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, true, err
	}
	if key.Type() != ir.Number && key.Type() != ir.String {
		return nil, true, l.notYet(node, "an enum index other than number or string")
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	// A helper evaluates receiver and key exactly once, including a key with side effects. It
	// uses ordinary IR so flow, ownership, regions and reuse all see its reads and returns.
	index := len(l.result.Functions)
	objectLocal := len(l.result.Locals)
	keyLocal := objectLocal + 1
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "enum_object", Type: ir.Object, Function: index}, ir.Local{Name: "enum_key", Type: key.Type(), Function: index})
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
		function.Body = append(function.Body, ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: tested, Right: ir.StringConstant{Index: l.constant(field.Name)}}, Then: []ir.Statement{ir.Return{Value: fit(ir.Property{Object: readObject, Name: field.Name, Of: field.Value.Type()}, of)}}})
	}
	if !of.IsReference() && !of.IsMaybe() {
		keys := l.checker.GetTypeAtLocation(access.ArgumentExpression)
		members := []*checker.Type{keys}
		if keys.Flags()&checker.TypeFlagsUnion != 0 {
			members = keys.Types()
		}
		for _, member := range members {
			if member.Flags()&checker.TypeFlagsStringLiteral == 0 {
				return nil, true, l.notYet(node, "an enum lookup without a proven member name")
			}
			name, ok := member.AsLiteralType().Value().(string)
			found := false
			for _, field := range fields {
				if ok && field.Name == name && field.Value.Type() == of {
					found = true
				}
			}
			if !found {
				return nil, true, l.notYet(node, "an enum lookup without a proven member name")
			}
		}
		function.Body = append(function.Body, ir.Panic{Message: ir.StringConstant{Index: l.constant("enum member name outside its proven union")}})
	} else {
		function.Body = append(function.Body, ir.Return{Value: ir.Undefined{Of: of}})
	}
	l.result.Functions = append(l.result.Functions, function)
	return ir.Call{Function: index, Arguments: []ir.Expression{object, key}, Returns: of}, true, nil
}

func (l *lowering) enumRefusal(node *ast.Node) error {
	if l.isExpression(node) {
		if symbol := l.flagValueSymbol(node); symbol != nil {
			declared := l.checker.GetTypeOfSymbol(symbol)
			observed := l.checker.GetTypeAtLocation(node)
			objectMembers := 0
			if declared.Flags()&checker.TypeFlagsUnion != 0 {
				for _, member := range declared.Types() {
					if member.Flags()&checker.TypeFlagsObject != 0 {
						objectMembers++
					}
				}
			}
			if objectMembers > 1 && declared.Flags()&checker.TypeFlagsUnion != 0 && observed != declared && observed.Flags()&checker.TypeFlagsObject != 0 {
				for _, member := range declared.Types() {
					for _, field := range l.checker.GetPropertiesOfType(member) {
						tag := l.checker.GetTypeOfSymbol(field)
						if l.openNumericEnumType(tag) {
							return &Refused{Where: l.program.Where(node), What: "an object refinement using an open numeric enum as a literal tag", Fix: "use a member-specific tag from a multi-member enum, a string enum, or a plain literal tag (adamic/enum-tag)"}
						}
					}
				}
			}
		}
	}

	if err := l.enumObjectView(node); err != nil {
		return err
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind == ast.KindPropertyAccessExpression && len(call.Arguments.Nodes) > 0 {
			access := callee.AsPropertyAccessExpression()
			if access.Name().Text() == "assign" && l.isLibraryGlobal(access.Expression, "Object") && l.enumObject(call.Arguments.Nodes[0]) != nil {
				return &Refused{Where: l.program.Where(node), What: "a write into an enum runtime object; named constants and reverse lookup must stay consistent", Fix: "copy the enum into an ordinary object before writing it (adamic/enum-object)"}
			}
		}
	}
	if node.Kind == ast.KindEnumDeclaration {
		if node.Parent.Kind != ast.KindSourceFile && node.Parent.Kind != ast.KindModuleBlock {
			return l.notYet(node, "an enum inside a function or block; declare it at module scope")
		}
		symbol := l.symbol(node.Name())
		if symbol != nil && len(symbol.Declarations) != 1 {
			return l.notYet(node, "merged enum declarations; put the members in one declaration")
		}
		if ast.HasSyntacticModifier(node, ast.ModifierFlagsAmbient) {
			return l.notYet(node, "an ambient enum without a runtime definition")
		}
		// Signed masks, including 1 << 31, are ordinary numeric enum constants.
		_, err := l.enumFields(node)
		return err
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
	if updated != nil && l.checker.GetTypeAtLocation(updated).Flags()&checker.TypeFlagsEnumLike != 0 && !l.numericEnum(l.enumIdentity(l.checker.GetTypeAtLocation(updated))) {
		if l.flagUpdate(node, updated) {
			return nil
		}
		if l.flagEnum(l.enumIdentity(l.checker.GetTypeAtLocation(updated))) {
			return l.flagWriteRefusal(node, l.checker.GetTypeAtLocation(updated))
		}
		return &Refused{Where: l.program.Where(node), What: "arithmetic assigned back into an enum; the result need not be one of its members", Fix: "assign a declared member, or keep arithmetic results in a number (adamic/enum-members)"}
	}
	return nil
}

// Compare runtime values, not spellings: numeric aliases denote the same union member. A default
// handles all remaining values. A void switch still needs coverage, even when tsc accepts it.
func (l *lowering) enumSwitch(node *ast.Node) error {
	statement := node.AsSwitchStatement()
	proven := l.checker.GetTypeAtLocation(statement.Expression)
	if proven.Flags()&checker.TypeFlagsEnumLike == 0 || l.numericEnum(l.enumIdentity(proven)) {
		return nil
	}
	covered := map[any]bool{}
	for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
		if clause.Kind == ast.KindDefaultClause {
			return nil
		}
		test := clause.AsCaseOrDefaultClause().Expression
		if member := l.enumMember(test); member != nil {
			covered[l.checker.GetConstantValue(member)] = true
		} else if value := l.checker.GetTypeAtLocation(test); value.Flags()&checker.TypeFlagsLiteral != 0 {
			covered[value.AsLiteralType().Value()] = true
		}
	}
	if l.flagEnum(l.enumIdentity(proven)) {
		return &Refused{Where: l.program.Where(node), What: "a flag-enum switch without a default", Fix: "add a default for combinations and zero (adamic/enum-flags)"}
	}
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsLiteral == 0 {
			return l.notYet(node, "a switch on a computed enum member")
		}
		if !covered[member.AsLiteralType().Value()] {
			return &Refused{Where: l.program.Where(node), What: "a non-exhaustive enum switch; missing " + l.checker.TypeToString(member), Fix: "add the missing member's case or a default"}
		}
	}
	return nil
}

// A numeric enum default is unreachable only with a member origin proof and full coverage.
// This permits the never-default idiom without inventing a machine representation for never.
func (l *lowering) enumDefaultUnreachable(node *ast.Node) bool {
	statement := node.AsSwitchStatement()
	proven := l.checker.GetTypeAtLocation(statement.Expression)
	return l.enumSwitchCovered(node) && (!l.numericEnum(l.enumIdentity(proven)) || l.enumMemberOrigin(statement.Expression, l.enumIdentity(proven), map[*ast.Node]bool{}))
}

func (l *lowering) enumSwitchCovered(node *ast.Node) bool {
	statement := node.AsSwitchStatement()
	proven := l.checker.GetTypeAtLocation(statement.Expression)
	if proven.Flags()&checker.TypeFlagsEnumLike == 0 {
		return false
	}
	covered := map[any]bool{}
	for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
		if clause.Kind == ast.KindDefaultClause {
			continue
		}
		test := clause.AsCaseOrDefaultClause().Expression
		if member := l.enumMember(test); member != nil {
			covered[l.checker.GetConstantValue(member)] = true
		} else if value := l.checker.GetTypeAtLocation(test); value.Flags()&checker.TypeFlagsLiteral != 0 {
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
