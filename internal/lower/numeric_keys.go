package lower

import (
	"reflect"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A computed key that's a number names the field its canonical number string names, ECMAScript's
// NumberToString, by Adamic's ruling (A) of Nov 24 23:48:03Z on #d1chbm7. Node enumerates an object's
// integer-like keys first, in ascending order, wherever they were written, and Adamic's fixed objects keep
// declaration order, so an object with number-named fields is safe only where nothing can enumerate it:
// Object.keys, for...in, JSON, a spread, console.log of the object, or the object passed anywhere that could
// do one of those. So a numeric key is named only where every one of these holds, and anywhere else keeps
// its stop, "a computed field name":
//   - the property is a plain `[key]: value`, its key, parentheses aside, a numeric literal or a name or
//     property access naming a numeric enum member, read from the checker's number literal type for it
//     (numericKey). A shorthand, a method or an accessor never comes here, and `[-1]` or `[-0]` is a minus
//     applied to a literal, not a literal;
//   - the number is an array index, an integer from 0 to 2^32 - 2, spelled by its decimal digits as
//     literalKeyName spells a finite key's read, so the literal's field and the lookup's are one name. A fraction
//     or a negative number keeps the stop, since only canonical integer keys are needed (the ruling's (2)), and so
//     does an integer past the last index, which Node orders among the string keys (Language, #d1chbm7);
//   - no two keys of the literal, numeric or named, spell one string (distinctFieldKeys): an enum alias such
//     as C = A beside A would make two fields where Node keeps only the later (the ruling's (1)). tsc refuses
//     each such pair first (TS1117), so this holds the ruling's fact where lowering stands, not tsc's;
//   - the literal is the whole initializer of a const that declares one name and isn't exported by a
//     modifier (numericKeyBinding), so every way to the object is a reference to that name;
//   - every reference to that name in every module, by the checker's symbol, is a keyed lookup by a number,
//     table[k] with every member of k's type number-like, only read: never assigned, updated, destructured
//     into, deleted or called through, since a call through it hands the function the object as its this
//     (numericKeyedLookup). A name in a type, typeof table among them, hands no value on and so is no
//     reference, as closedKindJudgeUse reads one. This is ruling 17's never-written proof, one fact proved
//     once, and its answer is held per binding (numericKeyTableClosed).

// numericFieldName is the field a numeric computed key names in literal, its canonical number string, where
// ruling (A) admits it (the comment above), and false where the key keeps its stop.
func (l *lowering) numericFieldName(literal, property *ast.Node) (string, bool) {
	if property.Kind != ast.KindPropertyAssignment {
		return "", false
	}
	name, named := l.numericKey(property.Name())
	if !named {
		return "", false
	}
	keys, spelled := l.literalFieldKeys(literal)
	if !spelled || !distinctFieldKeys(keys) {
		return "", false
	}
	declaration := numericKeyBinding(literal)
	if declaration == nil || !l.numericKeyTableClosed(declaration) {
		return "", false
	}
	return name, true
}

// numericKeyRead is the read of the enum member property's computed key names, where numericFieldName named its
// field by it, lowered so it runs before the field's value as Node evaluates the key: an ordinary enum's member is
// read through the enum object, and that read is where its enum's readiness is judged (rulings.md's "A known read
// before a regular enum's initialization is NotYet" and the computed-key entry after it). The field's name comes
// from the checker's constant, so without this read a key met before its enum is initialized named a field where
// Node throws TypeError (#kgngvxj). It's nil for a numeric literal and for a const enum's member, which Node's emit
// inlines as a constant too; a key the checker reads as an enum member through any other declaration is read, so an
// unknown one fails closed into the ordinary road.
func (l *lowering) numericKeyRead(property *ast.Node) (ir.Expression, error) {
	expression := ast.SkipParentheses(property.Name().AsComputedPropertyName().Expression)
	if expression.Kind == ast.KindNumericLiteral {
		return nil, nil
	}
	if member := l.enumMember(expression); member != nil && member.Parent != nil && member.Parent.Kind == ast.KindEnumDeclaration && ast.HasSyntacticModifier(member.Parent, ast.ModifierFlagsConst) {
		return nil, nil
	}
	return l.expression(expression)
}

// numericKey is the name a computed key written as a numeric literal or a numeric enum member gives its
// field, read from the checker's literal type, or false for any other key, and for a fraction, a negative
// number or one past the last array index, 2^32 - 2: Node orders an integer key past it among the string keys, by
// insertion, so only array indices are taken (Language, #d1chbm7).
func (l *lowering) numericKey(name *ast.Node) (string, bool) {
	if name == nil || name.Kind != ast.KindComputedPropertyName {
		return "", false
	}
	expression := ast.SkipParentheses(name.AsComputedPropertyName().Expression)
	switch expression.Kind {
	case ast.KindNumericLiteral:
		// Its number is its literal type's, read below with an enum member's.
	case ast.KindIdentifier, ast.KindPropertyAccessExpression:
		named := expression
		if expression.Kind == ast.KindPropertyAccessExpression {
			named = expression.Name()
		}
		if member := l.symbol(named); member == nil || member.Flags&ast.SymbolFlagsEnumMember == 0 {
			return "", false
		}
	default:
		return "", false
	}
	key := l.checker.GetTypeAtLocation(expression)
	if key.Flags()&checker.TypeFlagsNumberLiteral == 0 {
		return "", false
	}
	// literalKeyName spells a negative integer too, which a read may name; a key of one keeps its stop here.
	if value := reflect.ValueOf(key.AsLiteralType().Value()); value.Kind() != reflect.Float64 || value.Float() < 0 || value.Float() > 4294967294 {
		return "", false
	}
	return literalKeyName(key)
}

// literalFieldKeys is the key each of literal's properties spells, as Node names its field, or false where
// one can't be spelled here: a spread, whose fields are its source's, and a computed key that's neither a
// constant string, Symbol.iterator nor a key numericKey names.
func (l *lowering) literalFieldKeys(literal *ast.Node) ([]string, bool) {
	keys := []string{}
	for _, property := range literal.AsObjectLiteralExpression().Properties.Nodes {
		if key, named := l.numericKey(property.Name()); named {
			keys = append(keys, key)
			continue
		}
		if key, named := l.methodName(property); named {
			keys = append(keys, key)
			continue
		}
		if name := property.Name(); name != nil && name.Kind == ast.KindNumericLiteral {
			// The scanner keeps a numeric name's text as its canonical string (recordLiteral).
			keys = append(keys, name.Text())
			continue
		}
		return nil, false
	}
	return keys, true
}

// distinctFieldKeys reports whether no key is spelled twice.
func distinctFieldKeys(keys []string) bool {
	seen := map[string]bool{}
	for _, key := range keys {
		if seen[key] {
			return false
		}
		seen[key] = true
	}
	return true
}

// numericKeyBinding is the declaration literal is the initializer of, parentheses aside, where that's one
// name a const declares and no export modifier exports; nil anywhere else, a let, a var, a destructuring,
// a field and an argument among them. An export by a declaration, export { table }, names the binding, so
// numericKeyTableClosed's walk meets it as a reference.
func numericKeyBinding(literal *ast.Node) *ast.Node {
	initializer := literal
	for initializer.Parent != nil && initializer.Parent.Kind == ast.KindParenthesizedExpression {
		initializer = initializer.Parent
	}
	declaration := initializer.Parent
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration || declaration.Initializer() != initializer || !ast.IsIdentifier(declaration.Name()) {
		return nil
	}
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList || list.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst {
		return nil
	}
	if ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 {
		return nil
	}
	return declaration
}

// numericKeyTableClosed reports whether every reference to declaration's name, in every module of the
// program, is a keyed lookup by a number (numericKeyedLookup). It's held per binding symbol
// (numericKeyTables), so a literal with many numeric keys walks the program once: the answer reads only the
// program and the checker, neither of which lowering rewrites.
func (l *lowering) numericKeyTableClosed(declaration *ast.Node) bool {
	symbol := l.valueSymbol(declaration.Name())
	if symbol == nil {
		return false
	}
	if closed, held := l.numericKeyTables[symbol]; held {
		return closed
	}
	if l.numericKeyTables == nil {
		l.numericKeyTables = map[*ast.Symbol]bool{}
	}
	l.numericKeyTableWalks++
	closed := l.numericKeyReferencesKeyed(declaration, symbol)
	l.numericKeyTables[symbol] = closed
	return closed
}

// numericKeyReferencesKeyed is numericKeyTableClosed's walk: false at the first reference to symbol that
// isn't a keyed lookup by a number.
func (l *lowering) numericKeyReferencesKeyed(declaration *ast.Node, symbol *ast.Symbol) bool {
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		// A program whose modules can't be ordered is one the walk can't read whole, so nothing is proven.
		return false
	}
	keyed := true
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if !keyed {
			return true
		}
		// tsc counts typeof table's name as an expression, not part of a type node, though nothing runs it.
		if ast.IsIdentifier(node) && node != declaration.Name() && !ast.IsPartOfTypeNode(node) && !inTypeQuery(node) && l.numericKeyReference(node, symbol) && !l.numericKeyedLookup(node) {
			keyed = false
			return true
		}
		node.ForEachChild(visit)
		return !keyed
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return keyed
}

// numericKeyReference reports whether node, an identifier, names symbol: read as a value, through a
// shorthand property, or as an export specifier's local.
func (l *lowering) numericKeyReference(node *ast.Node, symbol *ast.Symbol) bool {
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindExportSpecifier && l.checker.GetExportSpecifierLocalTargetSymbol(parent) == symbol {
		return true
	}
	return l.valueSymbol(node) == symbol
}

// numericKeyedLookup reports whether reference, with the parentheses, as, <T> and ! around it, is the
// object of an element access whose key's type is number-like in every member, and that access only reads:
// it's no assignment's, update's or destructuring's target, no delete's operand, and no call's, new's or
// tagged template's callee.
func (l *lowering) numericKeyedLookup(reference *ast.Node) bool {
	object := numericKeyOuter(reference)
	access := object.Parent
	if access == nil || access.Kind != ast.KindElementAccessExpression || access.AsElementAccessExpression().Expression != object {
		return false
	}
	for _, member := range flattenedMembers(l.checker.GetTypeAtLocation(access.AsElementAccessExpression().ArgumentExpression)) {
		if member.Flags()&checker.TypeFlagsNumberLike == 0 {
			return false
		}
	}
	read := numericKeyOuter(access)
	if ast.IsAssignmentTarget(read) {
		return false
	}
	switch parent := read.Parent; parent.Kind {
	case ast.KindDeleteExpression:
		return false
	case ast.KindCallExpression:
		return parent.AsCallExpression().Expression != read
	case ast.KindNewExpression:
		return parent.AsNewExpression().Expression != read
	case ast.KindTaggedTemplateExpression:
		return parent.AsTaggedTemplateExpression().Tag != read
	}
	return true
}

// numericKeyOuter is node with the parentheses, as, <T> and ! written around it. ast.IsAssignmentTarget
// looks through parentheses and ! but not through as or <T>, and TypeScript lets `(table[k] as T) = v` write.
func numericKeyOuter(node *ast.Node) *ast.Node {
	for node.Parent != nil {
		switch node.Parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression:
			node = node.Parent
		default:
			return node
		}
	}
	return node
}
