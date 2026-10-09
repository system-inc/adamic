// Package reference answers what an identifier occurrence does to the binding it names.
//
// Lifted after a census read every rule in the tree at once and found four implementations of one
// decision, in four files that could not see each other. Three of them exist only to work around a
// measured gap in `ast.IsWriteAccess`, and each wrote the workaround differently.
package reference

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// WritesToBinding reports whether an identifier occurrence assigns to the binding it names.
//
// This is the structural half of every rule that anchors on a declaration and looks for writes to
// it. Symbol identity says which binding an identifier names and says nothing about whether the
// occurrence writes: `A.x = 0` and `foo(A)` both resolve to the anchor and neither reassigns it.
//
// # Why this is not `ast.IsWriteAccess`
//
// The shelf accessor is typescript-go's own answer to the question oxc asks as `is_write()`, and it
// is right about almost everything: compound assignment, both fixities of `++` and `--`,
// destructuring through array and object literals, property assignments, shorthand properties, and
// the `for (x of ...)` head. Two shapes it gets right that a hand-written climb tends to get wrong
// in the quiet direction:
//
//	const FOO = 1; ({ files = FOO } = arg1);   FOO is a default value, a read
//	for (const x of [1,2,3]) { foo(x); }       the head declares rather than assigns
//
// It has one measured gap: **a rest element in a destructuring assignment target reads as no write
// at all.** Its `accessKind` switch carries no arm for `KindSpreadElement` or
// `KindSpreadAssignment`, so both fall through to the default and are classified as reads.
//
// That gap was measured independently by three rules — `no-const-assign` on its imported corpus,
// `prefer-const` on its own, and `no-unused-private-class-members` citing both — and each wrote a
// different workaround. Measured across 57 inputs, `ast.IsWriteAccess` disagrees with all three
// workarounds on twelve of them, every one a rest element in an assignment target:
//
//	[...w] = []                        ({...w} = {})
//	[a, ...w] = []                     ({a, ...w} = o)
//	[a, b, ...[c, ...w]] = [1,2,3,4]   [, {foo: a, ...w}] = foo()
//	for ([...w] of xs) {}              for ({...w} of xs) {}
//	[[...w]] = xs                      [{...w}] = xs
//	({x: [...w]} = o)                  ({x: {...w}} = o)
//
// # Why the climb walks parentheses, which two of the three workarounds do not
//
// The same 57-input probe found the three workarounds disagreeing with **each other** on two
// inputs, and both are a parenthesis sitting directly around the identifier:
//
//	[...(a)] = xs      a real write; node assigns a = [1,2]
//	({...(a)} = o)     a real write; node assigns a the whole object
//
// `no-const-assign` and `prefer-const` both gate their climb on the *immediate* parent being a
// spread or a literal. A `KindParenthesizedExpression` is neither, so both bail on the first step
// and report no write. `no-class-assign` walks parentheses and gets both right — it had shipped the
// same defect for a plain assignment target and fixed it, recording at its own line that `(A) = 1`
// "is a write and read as a read before this line existed; neither upstream's corpus nor ours had a
// parenthesized assignment target, which is why it shipped silently."
//
// The fix did not travel to the two siblings, because nothing connected them. This function is the
// union, which is `no-class-assign`'s answer.
//
// # The shorthand property, which is two positions wearing one node kind
//
// `KindShorthandPropertyAssignment` carries the identifier in either of two slots and they mean
// opposite things:
//
//	({a} = o)           the identifier is Name()                       a WRITE
//	({ files = a } = o) the identifier is ObjectAssignmentInitializer  a READ, supplying a default
//
// `ast.IsWriteAccess` distinguishes them and answers correctly for both, so the shorthand is
// deliberately absent from the pass-through set below: climbing past it would reach the assignment
// from either slot and call both a write.
//
// The six rules that share `writesToItsIdentifier` list the shorthand as a transparent wrapper with
// no position test, so they answer true for the default-value read. Measured with node —
// `let a = "ORIGINAL"; ({ files = a } = {})` leaves `a` untouched — and it is upstream's own clean
// case for `no-const-assign`, quoted in that rule as "FOO is the default value, a read."
//
// That is a latent defect rather than a live one: all four rules route a shorthand through
// `GetShorthandAssignmentValueSymbol`, which resolves to the property rather than the anchor, so the
// symbol comparison declines before the wrong structural answer can reach a finding. Verified with
// controls on no-class-assign, no-func-assign, no-global-assign and no-import-assign; all four stay
// silent on the default-value shape today while their controls fire.
//
// # What keeps the unguarded climb honest
//
// It only ever hands the question to `ast.IsWriteAccess`. `foo(...w)` and `const xs = [...w]` climb
// to a call or a declaration nothing assigns to, so the answer comes back false and the caller stays
// quiet. Every wrapper in the pass-through set is transparent to assignment by construction: a
// destructuring assignment target parses as an array or object *literal* rather than as a pattern,
// so the wrappers below are the path from a rest element out to the assignment that decides it.
//
// # What this is NOT
//
// A rule asking whether an occurrence *only* writes, never reads, wants a different predicate.
// `x += 1` and `x++` both write and both read, and `no-unused-private-class-members` enumerates the
// write forms itself for exactly that reason: for it, a missed write becomes a phantom read and a
// member goes unreported, so it cannot inherit either this function's answer or the shelf's.
func WritesToBinding(identifier *ast.Node) bool {
	if identifier == nil {
		return false
	}

	// The name side of a property access is not a binding at all. `document.cookie = x` writes to a
	// property of `document`; `cookie` there names a member, and no global, class, or const named
	// `cookie` is touched by it. `ast.IsWriteAccess` answers true for that identifier because the
	// access as a whole is a write target, so a rule reaching this without the guard reports every
	// property write whose member name happens to collide with a binding it watches.
	//
	// Measured: `no-global-assign` produced 625 findings on our own tree against oxlint's 0, and
	// `document.cookie` alone accounted for most of them. The same shape is reachable from every
	// rule in this family, which is why the guard is here rather than in one of them.
	//
	// An element access is deliberately not excluded. `globalThis["Object"] = 1` writes through a
	// computed member, and the identifier inside the brackets is a read of a string, not a name
	// position, so it never reaches here as a write in the first place.
	if identifier.Parent != nil &&
		identifier.Parent.Kind == ast.KindPropertyAccessExpression &&
		identifier.Parent.AsPropertyAccessExpression().Name() == identifier {
		return false
	}

	current := identifier
	for current.Parent != nil {
		switch current.Parent.Kind {
		// The wrappers a rest element reaches its assignment through. A nested rest interleaves
		// them — `[a, b, ...[c, ...d]] = xs` puts `d` under spread, array, spread, array before
		// reaching the assignment — so both the spreads and the literals have to be walked or
		// neither is enough. Measured: a spread-only climb passed twenty-two of upstream's
		// twenty-four failing cases for `no-const-assign` and went silent on two.
		//
		// A property assignment passes through only from the value side: `({k: [...w]} = o)` writes
		// through `w` while `({[w]: k} = o)` names it as a computed key and writes nothing.
		//
		// Parentheses are here because they can sit anywhere in the chain, including directly
		// around the identifier, which is the shape two of the three lifted implementations miss.
		case ast.KindSpreadElement,
			ast.KindSpreadAssignment,
			ast.KindArrayLiteralExpression,
			ast.KindObjectLiteralExpression,
			ast.KindParenthesizedExpression:
			current = current.Parent

		case ast.KindPropertyAssignment:
			if current.Parent.AsPropertyAssignment().Initializer != current {
				return false
			}
			current = current.Parent

		default:
			return ast.IsWriteAccess(current)
		}
	}
	return ast.IsWriteAccess(current)
}

// IsUpdateOperator reports whether an operator writes back to its operand.
//
// Only `++` and `--`. A `-x`, `!x` or `typeof x` reads the binding and leaves it alone, so treating
// every unary as a write would report on code that never assigns.
//
// Lifted because two packages declared it separately under two names, byte-identical in logic, and
// neither could see the other.
func IsUpdateOperator(operator ast.Kind) bool {
	return operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken
}
