// Package classmembers answers which members of a class collide with each other.
//
// Lifted out of `internal/rules/core/no_dupe_class_members.go` because
// `@typescript-eslint/no-dupe-class-members` needs the same judgment and a rule package may not
// import another rule package: `TestRulePackagesStayLeaves` measures the cost at a 1.8s leaf
// rebuild against 8.5s for a deep one, paid by everyone on every edit.
//
// The extension rule delegates rather than re-deriving this, which is what upstream does
// (`baseRule.create(context)`) and the reason is not style: two implementations of one question
// give the two rules two chances to disagree about it.
package classmembers

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
)

// Key identifies one class member for duplicate detection.
//
// The name alone is not the key. `static foo()` and `foo()` are different members, and so are
// `#foo` and `foo`, so two members collide only when all three agree.
type Key struct {
	Name      string
	IsStatic  bool
	IsPrivate bool
}

// ForEachDuplicate calls report for every member that repeats one already seen in the same list.
//
// The report receives the node naming the later member and the key it collided on, which is what a
// caller needs to write a message; the caller owns the wording and the message id.
func ForEachDuplicate(members *ast.NodeList, report func(name *ast.Node, key Key)) {
	if members == nil {
		return
	}

	// Accessor kinds are tracked alongside the key so a getter and a setter of one name can
	// coexist while two of either cannot.
	seen := map[Key]ast.Kind{}

	for _, member := range members.Nodes {
		name := MemberName(member)
		if name == nil {
			continue
		}
		// An overload signature is not a member, it is a way of describing one.
		//
		// TypeScript lets a method carry several signatures and one implementation, and only the
		// implementation has a body. The signatures declare nothing on their own: they are erased,
		// they generate no code, and the implementation is what the class ends up with. Reporting
		// them says the earlier declaration is dead code that reads as live, which is the opposite
		// of true.
		//
		// The corpus this rule inherited is oxc's and is JavaScript-shaped, where a class member
		// always has a body, so no case in it could reach this. Found by running the binary over a
		// repository: `BaseSchema.ts` overloads `is` and `in` twice each, four reports on code
		// `tsc --noEmit` accepts.
		if IsOverloadSignature(member) {
			continue
		}
		key, ok := KeyOf(member, name)
		if !ok {
			continue
		}

		previousKind, duplicate := seen[key]
		if !duplicate {
			seen[key] = member.Kind
			continue
		}

		// Both accessors and of different kinds is a getter/setter pair rather than a duplicate.
		// Anything else collides, including an accessor against a method.
		bothAccessors := IsAccessorKind(member.Kind) && IsAccessorKind(previousKind)
		if bothAccessors && member.Kind != previousKind {
			continue
		}

		report(name, key)
	}
}

// IsOverloadSignature reports whether a class member declares a signature without implementing it.
//
// A method or accessor with no body, which in a class is only ever an overload signature: every
// other body-less member kind is filtered out before this runs, since `MemberName` answers for
// four kinds and a property declaration cannot carry a body at all.
//
// `abstract` members are body-less too and are deliberately included in that: two `abstract foo()`
// declarations are a genuine duplicate, and this returns true for each, so neither is reported. That
// is a real gap and it is smaller than the one it replaces, because a repeated abstract member is a
// compile error TypeScript reports itself while an overload is legal code this rule was flagging.
func IsOverloadSignature(member *ast.Node) bool {
	switch member.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return member.Body() == nil
	}
	return false
}

// IsAccessorKind reports whether a member is a getter or a setter.
func IsAccessorKind(kind ast.Kind) bool {
	return kind == ast.KindGetAccessor || kind == ast.KindSetAccessor
}

// MemberName returns the node naming a member, or nil for one this judgment does not cover.
//
// A static block and an index signature name nothing, and a member whose key is computed from an
// expression has no name known before it runs.
func MemberName(member *ast.Node) *ast.Node {
	switch member.Kind {
	case ast.KindMethodDeclaration, ast.KindPropertyDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor:
		return member.Name()
	}
	return nil
}

// KeyOf builds the identity a duplicate is judged against.
//
// The second return is false for a member whose key cannot be known statically, which is a member
// the judgment declines rather than guesses at: `[foo]()` and `foo()` are different members because
// the first depends on a variable, and the corpus pins that.
func KeyOf(member *ast.Node, name *ast.Node) (Key, bool) {
	// `NameTagged` rather than `Name`, and this is the one caller in the tree that needs the
	// tag: `[1.0]` and `['1.0']` are different members while `10` and `1e1` are the same one,
	// so neither the source spelling nor the cooked text answers alone.
	text, known := property.NameTagged(name, property.Static)
	if !known {
		return Key{}, false
	}
	return Key{
		Name:      text,
		IsStatic:  ast.HasStaticModifier(member),
		IsPrivate: name.Kind == ast.KindPrivateIdentifier,
	}, true
}
