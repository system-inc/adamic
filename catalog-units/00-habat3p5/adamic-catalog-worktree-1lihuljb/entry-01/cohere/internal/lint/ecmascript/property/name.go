// Package property answers what property a key or a member access names, when the syntax settles it
// without evaluation.
//
// Lifted after a census read every rule at once and found six implementations of this decision in
// six files that could not see each other. They were not in conflict: measured over thirteen key
// spellings they form a strict subset lattice, agreeing wherever they overlap and differing only in
// how many kinds each accepts. That is the shape a shared function serves well, because the accept
// set can be a parameter rather than a compromise.
package property

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// Kinds selects which key spellings a caller is willing to read.
//
// A caller states its set rather than taking a default, because the six lifted implementations
// disagreed about the set and every one of them had a reason. Making the widest set the default
// would silently widen five rules; making the narrowest the default would silently narrow one.
type Kinds uint8

const (
	// Named is a bare identifier key: `{ a: 1 }`, `x.a`, `class C { a() {} }`. Every caller wants
	// this one.
	Named Kinds = 1 << iota

	// Quoted is a string-literal key: `{ 'a': 1 }`, `x['a']`.
	Quoted

	// Templated is a no-substitution template key: `x[`+"`"+`a`+"`"+`]`. It can only reach a property key
	// through brackets, since `+"`"+`{ `+"`"+`a`+"`"+`: 1 }`+"`"+` is not valid JavaScript.
	Templated

	// Numeric is a numeric-literal key: `{ 42: 1 }`, `x[0]`.
	//
	// The parser normalizes these, which is worth knowing before a caller decides whether it wants
	// them: `1e1` and `0x10` answer "10" and "16", so two spellings of one number compare equal
	// without any work here. A caller that must keep a number and a string apart when their text
	// agrees wants `NameTagged` rather than this flag alone.
	Numeric

	// Private is a `#name` key. Text() carries the leading hash, so `#a` answers "#a" and can never
	// collide with the ordinary property `a`.
	Private

	// Computed reads through the brackets of a computed key, accepting whatever the flags above
	// accept for the expression inside.
	//
	// A bare identifier inside brackets is always declined, whatever the flags say, and that is the
	// one judgment this package makes rather than delegates. `{ [a]: 1 }` names whichever property
	// the variable `a` holds, which is not knowable before it runs. Reading the variable's spelling
	// as the key reported `[foo]()` and `foo()` as duplicate class members, caught by upstream's own
	// clean case rather than by inspection.
	Computed
)

// Common accept sets, named for the question their callers ask.
const (
	// Static is every spelling whose value the syntax settles: the union of the six lifted
	// implementations, minus the bare identifier inside brackets that none of them accepted.
	Static = Named | Quoted | Templated | Numeric | Private | Computed

	// Textual is the identifier and string spellings alone, which is what a caller comparing a key
	// against one fixed non-numeric name needs. Three of the lifted implementations took exactly
	// this set, and widening them would change no answer: no number and no template renders as
	// `onSuccess` or `children`.
	Textual = Named | Quoted
)

// Name returns the property a key or member-access name denotes, and whether the syntax settles it.
//
// The second return separates "not a static name" from a name that is empty, which a caller
// reporting on the name node has to be able to tell apart.
//
// # The kind switch is load-bearing rather than defensive
//
// `Node.Text()` PANICS on a `KindComputedPropertyName` rather than returning empty, so a caller that
// read the text first and filtered afterwards crashes the linter on `{[children]: 1}`. That is not
// hypothetical: a probe written before `no-children-prop` existed hit exactly that panic, and three
// of the six lifted implementations had no guard against it because their callers never met the
// shape.
func Name(node *ast.Node, accept Kinds) (string, bool) {
	if node == nil {
		return "", false
	}

	switch node.Kind {
	case ast.KindIdentifier:
		if accept&Named != 0 {
			return node.Text(), true
		}

	case ast.KindPrivateIdentifier:
		if accept&Private != 0 {
			return node.Text(), true
		}

	case ast.KindStringLiteral:
		if accept&Quoted != 0 {
			return node.Text(), true
		}

	case ast.KindNoSubstitutionTemplateLiteral:
		if accept&Templated != 0 {
			return node.Text(), true
		}

	case ast.KindNumericLiteral:
		if accept&Numeric != 0 {
			return node.Text(), true
		}

	case ast.KindComputedPropertyName:
		if accept&Computed == 0 {
			return "", false
		}
		inner := ast.SkipParentheses(node.AsComputedPropertyName().Expression)
		if inner == nil {
			return "", false
		}
		// A variable inside brackets names whatever it holds. Declined whatever the flags say; see
		// the Computed doc.
		if inner.Kind == ast.KindIdentifier || inner.Kind == ast.KindPrivateIdentifier {
			return "", false
		}
		return Name(inner, accept)
	}
	return "", false
}

// NameTagged is Name with the value's type carried alongside it, for a caller that must keep a
// number and a string apart when their text agrees.
//
// One caller needs this and the rest must not have it. `no-dupe-class-members` decides whether two
// members collide, and there `[1.0]` and `['1.0']` are different members while `10` and `1e1` are
// the same one, so neither the source spelling nor the cooked text answers alone. Every other
// caller compares against one fixed name, where a tag would only get in the way.
func NameTagged(node *ast.Node, accept Kinds) (string, bool) {
	if node == nil {
		return "", false
	}
	if node.Kind == ast.KindComputedPropertyName {
		if accept&Computed == 0 {
			return "", false
		}
		inner := ast.SkipParentheses(node.AsComputedPropertyName().Expression)
		if inner == nil {
			return "", false
		}
		if inner.Kind == ast.KindIdentifier || inner.Kind == ast.KindPrivateIdentifier {
			return "", false
		}
		return NameTagged(inner, accept)
	}

	text, ok := Name(node, accept)
	if !ok {
		return "", false
	}
	if node.Kind == ast.KindNumericLiteral {
		return "number:" + text, true
	}
	return "string:" + text, true
}

// AccessedName returns the property a member access reads, when the syntax settles it.
//
// `a.b` and `a['b']` both answer "b", which is what makes the two spellings compare as one
// reference. `a[i]` answers nothing, so a comparison involving it is false rather than optimistic:
// two reads of `a[i]` are the same reference only if `i` has not changed, and nothing here can know
// that.
func AccessedName(node *ast.Node, accept Kinds) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return Name(node.AsPropertyAccessExpression().Name(), accept)
	case ast.KindElementAccessExpression:
		// The subscript is an expression rather than a key node, so it is read directly rather than
		// through the computed arm: there are no brackets in the tree to see through.
		argument := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		if argument == nil || argument.Kind == ast.KindIdentifier ||
			argument.Kind == ast.KindPrivateIdentifier {
			return "", false
		}
		return Name(argument, accept)
	}
	return "", false
}
