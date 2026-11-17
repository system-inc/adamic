package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// A union member S whose tag field is a whole open numeric enum, as tsc's Expression is with kind:
// SyntaxKind, makes every refinement of the union a hazard (enumTagHazardOf): an S value could carry
// another member's kind with S's layout. S is closed for that field, and its kind domain K(S) proves
// the walk's narrowing where the enum couldn't, when every way a value reaches S without a hatch gives
// it a literal kind (#x2cjvy9, Adamic's ruling of Nov 17). The premise has two halves:
//
//  1. A value that reaches S with no hatch was built at a type with a literal kind in K(S), so its kind
//     names its layout.
//  2. A value that reaches S through a hatch, a cast or a predicate, is a checked view in .ts, and stays
//     one through locals, parameters, fields, returns and containers (rulings.md:18), the kind narrowing
//     included. In .a such a carry is refused unproven.
//
// With the hazard lifted, every road a lie took stopped at its read with exit 70 on both backends (job
// J-4hrkxyf). K(S) is read off the program's own files: the prelude, the library and declaration files
// build nothing that reaches S without a hatch.
//
//   - Declared types: every interface, type alias and class the program declares that the checker calls
//     assignable to S, brands and all, adds its kind's literals. One whose kind is open, the enum, a type
//     parameter or anything else that isn't literals, is abstract and adds nothing.
//   - Build sites: every object literal, spread included, new, class and Object.assign the program writes,
//     and every write of the field, whose built type is assignable to S. A literal kind adds to K(S); one
//     that isn't leaves S open, and the reason names the site. So `{ kind: K.B, ...brands }` typed S puts
//     B in K(S), and `new N(kindOf())` of a class N assignable to S keeps S open.
//
// The walk alone reads K(S) (enumTagWalkHazardOf, enumTagClosedTag). R-T80's provers ask
// enumTagHazardOf and keep the hazard.

// closedKindKey is a type, by its id, and the field closedKinds was asked about it.
type closedKindKey struct {
	id    int
	field string
}

// closedKindAnswer is closedKinds' answer: K(S) as a union of literal types, never when it's empty,
// whether S is closed, and why not when it isn't.
type closedKindAnswer struct {
	kinds  *checker.Type
	closed bool
	reason string
}

// closedKindSite is a declared type or a build site and the type it declares or builds. A build's type is
// its regular type, widened, so a fresh literal's excess properties never hide it from S; a write's is
// the type of the object it writes the field of.
type closedKindSite struct {
	node *ast.Node
	of   *checker.Type
}

// closedKindShape is how a kind's type is judged against the enum S's field holds.
type closedKindShape int

const (
	// closedKindOpen is the whole enum, a type parameter, or anything else that isn't only literals.
	closedKindOpen closedKindShape = iota
	// closedKindLiteral is one member of S's enum, or a union of them.
	closedKindLiteral
	// closedKindForeign is a literal that isn't a member of S's enum, a plain number among them: K(S)
	// holds the members the walk compares with, which never equal a plain number to the checker, so
	// one would fail open. S is open there.
	closedKindForeign
)

// closedKinds is K(of) for field, whether of is closed for it, and why not when it isn't. It's held for
// the lowering, by of's id and field.
func (l *lowering) closedKinds(of *checker.Type, field string) (*checker.Type, bool, string) {
	key := closedKindKey{id: int(of.Id()), field: field}
	if answer, held := l.closedKindAnswers[key]; held {
		return answer.kinds, answer.closed, answer.reason
	}
	answer := l.closedKindsOf(of, field)
	if l.closedKindAnswers == nil {
		l.closedKindAnswers = map[closedKindKey]closedKindAnswer{}
	}
	l.closedKindAnswers[key] = answer
	return answer.kinds, answer.closed, answer.reason
}

func (l *lowering) closedKindsOf(of *checker.Type, field string) closedKindAnswer {
	property := l.checker.GetPropertyOfType(of, field)
	if property == nil {
		return closedKindAnswer{reason: l.checker.TypeToString(of) + " has no field " + field}
	}
	tag := l.checker.GetTypeOfSymbol(property)
	if !l.openNumericEnumType(tag) {
		return closedKindAnswer{reason: l.checker.TypeToString(of) + "'s " + field + " isn't an open numeric enum, so it has no kind domain"}
	}
	kinds := []*checker.Type{}
	for _, declared := range l.closedKindDeclaredTypes() {
		if !l.checker.IsTypeAssignableTo(declared.of, of) {
			continue
		}
		kind := l.closedKindFieldType(declared.of, field)
		literals, shape := l.closedKindLiterals(kind, tag)
		switch shape {
		case closedKindLiteral:
			kinds = append(kinds, literals...)
		case closedKindForeign:
			return closedKindAnswer{reason: l.closedKindReason(of, field, "the declared type "+l.checker.TypeToString(declared.of), declared.node, kind)}
		}
	}
	for _, build := range l.closedKindBuildSites() {
		kind, construct, builds := l.closedKindBuildKind(build, field)
		if !builds || !l.closedKindReaches(build.of, of) {
			continue
		}
		if kind == nil {
			return closedKindAnswer{reason: l.closedKindReason(of, field, construct, build.node, nil)}
		}
		literals, shape := l.closedKindLiterals(kind, tag)
		if shape != closedKindLiteral {
			return closedKindAnswer{reason: l.closedKindReason(of, field, construct, build.node, kind)}
		}
		kinds = append(kinds, literals...)
	}
	return closedKindAnswer{kinds: l.checker.GetUnionType(kinds), closed: true}
}

// closedKindReason says why of is open for field: what gives a value that reaches it which kind, where.
// A nil kind is a write that isn't one value's, a compound one, an update or a destructuring.
func (l *lowering) closedKindReason(of *checker.Type, field, construct string, site *ast.Node, kind *checker.Type) string {
	given := "a " + field + " that isn't one value's"
	if kind != nil {
		given = "the " + field + " " + l.checker.TypeToString(kind)
	}
	return l.checker.TypeToString(of) + " is open: " + construct + " at " + l.program.Where(site) + " gives a value that reaches it " + given
}

// closedKindReaches is whether a value built at built reaches of with no hatch: built is assignable to of,
// or, as the object a write's target reads off can be, is a union with a member that is.
func (l *lowering) closedKindReaches(built, of *checker.Type) bool {
	if l.checker.IsTypeAssignableTo(built, of) {
		return true
	}
	if built.Flags()&checker.TypeFlagsUnion != 0 {
		for _, member := range built.Types() {
			if l.checker.IsTypeAssignableTo(member, of) {
				return true
			}
		}
	}
	return false
}

// closedKindFieldType is the type owner's field has, or nil when it has none.
func (l *lowering) closedKindFieldType(owner *checker.Type, field string) *checker.Type {
	property := l.checker.GetPropertyOfType(owner, field)
	if property == nil {
		return nil
	}
	return l.checker.GetTypeOfSymbol(property)
}

// closedKindLiterals judges a kind's type, less null and undefined, against tag, the open enum S's field
// holds, and gives its literals when it's one of tag's members or a union of them. One literal is a
// literal even when it's a one-member enum's whole type, as the probe judged it; a union naming every
// member of an enum is open (openNumericEnumType).
func (l *lowering) closedKindLiterals(kind, tag *checker.Type) ([]*checker.Type, closedKindShape) {
	if kind == nil {
		return nil, closedKindOpen
	}
	proven := l.checker.GetNonNullableType(kind)
	members := []*checker.Type{proven}
	switch {
	case proven.Flags()&checker.TypeFlagsUnion == 0 && proven.Flags()&checker.TypeFlagsLiteral != 0:
	case l.openNumericEnumType(proven):
		return nil, closedKindOpen
	case proven.Flags()&checker.TypeFlagsUnion != 0:
		members = proven.Types()
		for _, member := range members {
			if member.Flags()&checker.TypeFlagsLiteral == 0 {
				return nil, closedKindOpen
			}
		}
	default:
		return nil, closedKindOpen
	}
	identity := l.enumIdentity(l.checker.GetNonNullableType(tag))
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsEnumLiteral == 0 || member.Flags()&checker.TypeFlagsNumberLiteral == 0 || identity == nil || l.enumIdentity(member) != identity {
			return nil, closedKindForeign
		}
	}
	return members, closedKindLiteral
}

// closedKindDeclaredTypes is every interface, type alias and class the program's own files declare,
// anywhere, a namespace's and a function's included, by its symbol, found once.
func (l *lowering) closedKindDeclaredTypes() []closedKindSite {
	if l.closedKindDeclared != nil {
		return l.closedKindDeclared
	}
	l.closedKindDeclared = []closedKindSite{}
	seen := map[*ast.Symbol]bool{}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindClassDeclaration, ast.KindClassExpression:
			if symbol := l.checker.GetSymbolOfDeclaration(node); symbol != nil && !seen[symbol] {
				seen[symbol] = true
				if declared := l.checker.GetDeclaredTypeOfSymbol(symbol); declared != nil {
					l.closedKindDeclared = append(l.closedKindDeclared, closedKindSite{node: node, of: declared})
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	// neverRestProgramFile is the probe's ksetOrigin of "Program": Node's types are declaration files.
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if neverRestProgramFile(file) {
			file.AsNode().ForEachChild(visit)
		}
	}
	return l.closedKindDeclared
}

// closedKindBuildSites is every place the program's own files build a value or write a field of one,
// found once: an object literal that isn't a destructuring pattern, a new, a class declaration or
// expression, a call of the library's Object.assign, and a property or element access that is an
// assignment's target, a compound one's, an update's or a destructuring's included.
func (l *lowering) closedKindBuildSites() []closedKindSite {
	if l.closedKindBuilt != nil {
		return l.closedKindBuilt
	}
	l.closedKindBuilt = []closedKindSite{}
	built := func(node *ast.Node, of *checker.Type) {
		if of != nil {
			l.closedKindBuilt = append(l.closedKindBuilt, closedKindSite{node: node, of: l.checker.GetWidenedType(regularTypeOfObjectLiteral(l.checker, of))})
		}
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindObjectLiteralExpression:
			if !ast.IsAssignmentTarget(node) {
				built(node, l.checker.GetTypeAtLocation(node))
			}
		case ast.KindNewExpression:
			built(node, l.checker.GetTypeAtLocation(node))
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if symbol := l.checker.GetSymbolOfDeclaration(node); symbol != nil {
				built(node, l.checker.GetDeclaredTypeOfSymbol(symbol))
			}
		case ast.KindCallExpression:
			callee := ast.SkipParentheses(node.AsCallExpression().Expression)
			if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "assign" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
				built(node, l.checker.GetTypeAtLocation(node))
			}
		case ast.KindPropertyAccessExpression:
			if ast.IsAssignmentTarget(node) {
				built(node, l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression))
			}
		case ast.KindElementAccessExpression:
			if ast.IsAssignmentTarget(node) {
				built(node, l.checker.GetTypeAtLocation(node.AsElementAccessExpression().Expression))
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, file := range l.program.CompilerProgram().GetSourceFiles() {
		if neverRestProgramFile(file) {
			file.AsNode().ForEachChild(visit)
		}
	}
	return l.closedKindBuilt
}

// closedKindBuildKind is the type a build site gives field, what the site is, for a reason, and whether it
// gives field at all: a site whose built type has no such field, or a write of another field, doesn't.
// The type is nil for a write that isn't one value's.
func (l *lowering) closedKindBuildKind(site closedKindSite, field string) (*checker.Type, string, bool) {
	switch site.node.Kind {
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if !l.closedKindWritesField(site.node, field) {
			return nil, "", false
		}
		construct := "a write of " + field
		// Only a plain `=` with the target its whole left side writes one value, the right's.
		assignment := ast.GetAssignmentTarget(site.node)
		if assignment == nil || assignment.Kind != ast.KindBinaryExpression {
			return nil, construct, true
		}
		binary := assignment.AsBinaryExpression()
		if binary.OperatorToken.Kind != ast.KindEqualsToken || ast.SkipParentheses(binary.Left) != site.node {
			return nil, construct, true
		}
		return l.checker.GetTypeAtLocation(binary.Right), construct, true
	}
	if l.checker.GetPropertyOfType(site.of, field) == nil {
		return nil, "", false
	}
	switch site.node.Kind {
	case ast.KindObjectLiteralExpression:
		return l.closedKindObjectLiteral(site, field), "an object literal", true
	case ast.KindNewExpression:
		return l.closedKindFieldType(site.of, field), "a new", true
	case ast.KindCallExpression:
		return l.closedKindFieldType(site.of, field), "Object.assign", true
	}
	return l.closedKindFieldType(site.of, field), "the class " + l.checker.TypeToString(site.of), true
}

// closedKindWritesField is whether a property or element access names field: by its name, or by a key
// whose type holds field's name. A key whose type is a string but no one name can be field too.
func (l *lowering) closedKindWritesField(target *ast.Node, field string) bool {
	if target.Kind == ast.KindPropertyAccessExpression {
		return target.AsPropertyAccessExpression().Name().Text() == field
	}
	key := l.checker.GetTypeAtLocation(target.AsElementAccessExpression().ArgumentExpression)
	members := []*checker.Type{key}
	if key.Flags()&checker.TypeFlagsUnion != 0 {
		members = key.Types()
	}
	for _, member := range members {
		if member.Flags()&checker.TypeFlagsStringLiteral != 0 {
			if name, ok := member.AsLiteralType().Value().(string); ok && name == field {
				return true
			}
			continue
		}
		if member.Flags()&(checker.TypeFlagsString|checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
			return true
		}
	}
	return false
}

// closedKindObjectLiteral is the type an object literal gives field: its own last member named field's
// value, read before the literal widens it, unless a spread comes after it; otherwise, a spread's or a
// computed name's, the literal's own type's field. A shorthand member's value is its symbol's declared
// type, as shorthand lowers it (object.go), and an accessor's or a method's is the literal's field.
func (l *lowering) closedKindObjectLiteral(site closedKindSite, field string) *checker.Type {
	var own *ast.Node
	for _, property := range site.node.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind == ast.KindSpreadAssignment {
			own = nil
			continue
		}
		if name := property.Name(); name != nil && (ast.IsIdentifier(name) || name.Kind == ast.KindStringLiteral) && name.Text() == field {
			own = property
		}
	}
	if own == nil {
		return l.closedKindFieldType(site.of, field)
	}
	switch own.Kind {
	case ast.KindPropertyAssignment:
		return l.checker.GetTypeAtLocation(own.AsPropertyAssignment().Initializer)
	case ast.KindShorthandPropertyAssignment:
		symbol := l.checker.GetShorthandAssignmentValueSymbol(own)
		if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
			symbol = l.checker.GetAliasedSymbol(symbol)
		}
		if symbol == nil {
			return l.closedKindFieldType(site.of, field)
		}
		return l.checker.GetTypeOfSymbol(l.checker.GetExportSymbolOfSymbol(symbol))
	}
	return l.closedKindFieldType(site.of, field)
}
