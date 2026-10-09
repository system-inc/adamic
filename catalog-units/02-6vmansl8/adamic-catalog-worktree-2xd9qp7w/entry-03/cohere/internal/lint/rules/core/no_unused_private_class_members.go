package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// privateClassMember is one `#name` declared in one class body, and what has been seen of it.
//
// Keyed by name rather than by declaration node, which is what makes a getter and a setter of the
// same name one entry instead of two. That is not a shortcut: `get #a` and `set #a` are the two
// halves of one accessor, so using either half uses both, and a pair with neither half referenced
// has to report once rather than twice. Upstream's map is keyed the same way and its corpus pins
// the count.
type privateClassMember struct {
	// name is the node the finding points at, which is the `#name` token including the hash.
	name *ast.Node
	// isAccessor records that this entry is a getter or a setter, for which any reference at all
	// counts as a use. Reading or writing an accessor runs a function body, and that body can do
	// anything, so a write is not a write-only use the way it is for a field.
	isAccessor bool
	// used records that a reference counting as a use has been seen.
	used bool
}

// NoUnusedPrivateClassMembers flags a `#private` class member nothing in the class ever reads.
//
//	valid:   class A { #used = 1; m() { return this.#used; } }
//	valid:   class A { #m() {} n() { return this.#m(); } }
//	valid:   class A { set #a(v) {} m() { this.#a = 1; } }
//	valid:   class A { #brand; static has(o) { return #brand in o; } }
//	invalid: class A { #unused = 1; }
//	invalid: class A { #writeOnly = 1; m() { this.#writeOnly = 2; } }
//	invalid: class A { #counter = 0; m() { this.#counter++; } }
//	invalid: class A { #unusedMethod() {} }
//
// A private member is visible only inside the class body that declares it, which is what makes this
// rule decidable at all: there is no other file to check and no dynamic access to worry about, so a
// member no line of this class body reads is dead with certainty rather than by inference. That
// certainty is the whole reason the rule can be in `correctness` instead of a style tier.
//
// # Written but never read is the finding, not merely never mentioned
//
// The interesting half is not `#unused = 1` with no other mention, which any reader spots. It is a
// field the class assigns and never reads back, because that one looks busy. `this.#writeOnly = 42`
// reads like the member is doing work while the value it stores can never be observed, so the
// assignment is as dead as the declaration and usually the residue of a refactor that removed the
// reader. Same for `this.#counter++` as a statement, which increments a number nobody will ever
// look at.
//
// So the question a reference is asked is "is this a read", and the write forms are enumerated
// rather than inferred. A compound assignment is the sharp edge: `this.#x += 1` discarded as a
// statement is write-only, but `return this.#x += 1` reads the value, so the same operator lands on
// either side of the line depending on whether the result is used. `++this.#x` splits the same way.
//
// # An accessor is used by any reference, including a pure write
//
// `set #a(v) {}` referenced only as `this.#a = 1` is *used*, unlike a field with the same shape,
// because the assignment calls the setter body and that body can have effects the class depends on.
// This is the one place where the rule cannot reason from the reference alone: it has to know what
// the member is. Getting this backwards reports every write-only setter in the tree, which is the
// normal way to write one.
//
// # The brand check is a use that looks like nothing else
//
// `#brand in obj` is the only reference to a private name that is not a property access. It parses
// as a binary expression whose left operand is the bare `PrivateIdentifier`, with no member
// expression above it, so a rule that only inspects `PropertyAccessExpression` nodes cannot see it
// and reports a live brand check as dead. It is a read.
//
// # Where this follows ESLint over oxc
//
// oxc reports four cases ESLint declares clean, all the same shape: a conditional whose result is
// discarded as a statement, such as `class Foo { #x; m() { a ? this.#x : b; } }`. oxc tracks whether
// the conditional's value reaches a value context and calls the read dead when it does not.
//
// Measured against `eslint@9` rather than assumed: all four are clean there, and they are the only
// semantic disagreement between the two implementations across the whole 87-case corpus. The other
// ten places the two differ are TypeScript syntax ESLint's default parser rejects, which says
// nothing about the rule.
//
// This follows ESLint, for a reason specific to what kind of rule this is. Every finding here is an
// accusation that code is dead, so a false positive lands on correct code and asks the reader to
// delete something that works. Reading `this.#x` in a discarded ternary still runs a getter if `#x`
// is one, and still pins the field against a later refactor, so calling it dead is a claim about
// intent rather than about reachability. When the two upstreams disagree on an unused-thing rule,
// the quieter one is the right default and the divergence is stated rather than silent.
//
// # Why the fix is a suggestion upstream and nothing here
//
// ESLint offers removal as a *suggestion*, never as an unattended fix, and oxc ships no repair at
// all. Deleting a member changes the shape of the class: the surrounding comment may describe it,
// a `declare` may be pinning a type another file depends on, and the author's answer is at least as
// often "add the reader I forgot" as "delete this". That is a choice a human makes, so nothing is
// proposed here rather than proposing a rewrite the engine would apply while the reader is not
// looking.
var NoUnusedPrivateClassMembers = rule.Rule{
	Name: "no-unused-private-class-members",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		check := func(members *ast.NodeList) {
			if members == nil {
				return
			}

			// Declared first, in source order, so the findings come out in the order a reader would
			// scan them rather than in map order.
			declared := map[string]*privateClassMember{}
			var order []string
			for _, member := range members.Nodes {
				name := privateMemberName(member)
				if name == nil {
					continue
				}
				text := name.Text()
				if existing, alreadySeen := declared[text]; alreadySeen {
					// A getter and a setter of one name share an entry. A genuine duplicate lands
					// here too, and keeping the first declaration is what makes the pair report
					// once; `no-dupe-class-members` is the rule that judges the duplication itself.
					existing.isAccessor = existing.isAccessor || ast.IsAccessor(member)
					continue
				}
				declared[text] = &privateClassMember{name: name, isAccessor: ast.IsAccessor(member)}
				order = append(order, text)
			}
			if len(declared) == 0 {
				return
			}

			// Every reference lives inside this class body, so walking it is the whole search. A
			// nested class declaring the same name shadows it, and the walk stops at that boundary
			// rather than crediting the inner class's references to the outer member.
			for _, member := range members.Nodes {
				markPrivateMemberUses(member, declared)
			}

			for _, text := range order {
				if declared[text].used {
					continue
				}
				ctx.ReportNode(declared[text].name, rule.Message{
					Id: "noUnusedPrivateClassMember",
					// `text` is the name node's own text, which already carries the leading
					// hash: `#foo`, not `foo`. Measured by the dry run against the real tree,
					// which printed `'##unusedField'` from a `'#%s'` format. The fixture asserting
					// the name appears used `strings.Contains`, and a doubled hash contains the
					// single-hash needle, so it stayed green while the message was wrong. It now
					// asserts the whole message text instead.
					Description: fmt.Sprintf(
						"'%s' is declared but nothing in this class ever reads it. A private "+
							"member is visible only inside its own class body, so a value no line "+
							"here reads back can never be observed and the code that maintains it "+
							"is dead too. Read it, or remove the member and its writes.", text),
				})
			}
		}

		return rule.Listeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				check(node.AsClassDeclaration().Members)
			},
			ast.KindClassExpression: func(node *ast.Node) {
				check(node.AsClassExpression().Members)
			},
		}
	},
}

// privateMemberName returns the `#name` node a class member declares, or nil for anything else.
//
// A member whose key is a plain identifier, a string, or a computed expression is public and out of
// scope, and a static block or an index signature names nothing. `accessor #a` and `declare #a`
// both parse as ordinary property declarations, so neither needs an arm of its own: measured with a
// probe rather than assumed, because a modifier that changed the node kind would make this silently
// skip the member and report nothing about it.
func privateMemberName(member *ast.Node) *ast.Node {
	switch member.Kind {
	case ast.KindPropertyDeclaration, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor:
		name := member.Name()
		if name != nil && name.Kind == ast.KindPrivateIdentifier {
			return name
		}
	}
	return nil
}

// markPrivateMemberUses walks one class member and marks every declared name it reads.
//
// The walk descends through everything inside the member, which is what makes a use inside a nested
// arrow, inside a static block, or inside a field initializer count the same as one in a method
// body. `this` is irrelevant to the search: a private name is lexically scoped to its class, so
// `other.#foo` inside the class reaches the same member that `this.#foo` does, and both are uses.
//
// It stops at a nested class that redeclares the same name, because from there inward `#foo` means
// the inner class's member. A walk without that boundary credits the inner class's own references
// to the outer declaration and goes silent on a genuinely dead outer member, which is the shape
// upstream carries three separate fail cases for.
func markPrivateMemberUses(node *ast.Node, declared map[string]*privateClassMember) {
	if node == nil {
		return
	}

	if node.Kind == ast.KindPrivateIdentifier {
		// `!entry.used` is a short-circuit rather than a discrimination, and the sweep says so: the
		// flag is only ever set, never cleared, so re-testing a member already marked used can only
		// set it again. The mutant dropping it survived, correctly, and there is no input that
		// distinguishes the two. Kept because most classes reference a member more than once.
		if entry, isDeclared := declared[node.Text()]; isDeclared && !entry.used {
			if isPrivateMemberRead(node, entry.isAccessor) {
				entry.used = true
			}
		}
		// A `PrivateIdentifier` is a leaf: probed across all three of its syntactic positions and it
		// reports zero children in every one, so this return is a short-circuit too. The mutant
		// dropping it survived for that reason rather than for a gap in the fixtures.
		return
	}

	// A nested class shadows the names it redeclares. Anything it does not redeclare stays visible,
	// so the recursion continues with a narrowed map rather than stopping outright.
	if node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression {
		if narrowed := withoutShadowedNames(node, declared); narrowed != nil {
			declared = narrowed
			// Another short-circuit the sweep flagged as equivalent: with nothing left to look for,
			// the walk below would descend the whole nested class and mark nothing. Stopping early
			// is faster and cannot differ.
			if len(declared) == 0 {
				return
			}
		}
	}

	node.ForEachChild(func(child *ast.Node) bool {
		markPrivateMemberUses(child, declared)
		return false
	})
}

// withoutShadowedNames returns the map minus every name a nested class redeclares, or nil when it
// redeclares none.
//
// Nil rather than a copy for the common case, because most nested classes shadow nothing and
// copying the map at every one of them would make the walk quadratic in nesting depth for no
// result.
func withoutShadowedNames(class *ast.Node, declared map[string]*privateClassMember) map[string]*privateClassMember {
	var members *ast.NodeList
	if class.Kind == ast.KindClassDeclaration {
		members = class.AsClassDeclaration().Members
	} else {
		members = class.AsClassExpression().Members
	}
	if members == nil {
		return nil
	}

	var narrowed map[string]*privateClassMember
	for _, member := range members.Nodes {
		name := privateMemberName(member)
		if name == nil {
			continue
		}
		if _, isDeclared := declared[name.Text()]; !isDeclared {
			continue
		}
		if narrowed == nil {
			narrowed = make(map[string]*privateClassMember, len(declared))
			for key, value := range declared {
				narrowed[key] = value
			}
		}
		delete(narrowed, name.Text())
	}
	return narrowed
}

// isPrivateMemberRead reports whether one occurrence of a private name counts as using the member.
//
// The declaration itself is not a use, and neither is a reference that only ever writes. Everything
// else is, which is the conservative default this rule needs: a reference form nobody enumerated
// reads as a use and the member stays unreported, rather than reading as dead and accusing working
// code. For an unused-thing rule that direction is the cheap failure, since the expensive one asks
// a reader to delete something live.
func isPrivateMemberRead(name *ast.Node, isAccessor bool) bool {
	parent := name.Parent
	if parent == nil {
		return false
	}

	// The `#name` in `#name = 1`, `#name() {}`, `get #name() {}` and `set #name(v) {}` is the
	// declaration rather than a reference to it. Without this the rule marks every member used the
	// moment it is declared and reports nothing ever.
	switch parent.Kind {
	case ast.KindPropertyDeclaration, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor:
		if parent.Name() == name {
			return false
		}
	}

	// Everything that is not a property access is a read, and there is exactly one such shape.
	//
	// Measured rather than reasoned: a `PrivateIdentifier` occupies three positions in valid
	// syntax, and a probe over every one of them found no fourth. It is the name of a declaration,
	// handled above; it is the left operand of `#brand in obj`, always as the left; or it is the
	// name of a `PropertyAccessExpression`. So the brand check is the only thing this arm answers.
	//
	// It is written as the broad test rather than as a narrow `in` arm on purpose. A narrow arm and
	// a broad default are indistinguishable given the measurement above, which the sweep confirmed
	// by leaving both mutants alive, so the broad form is the one that keeps this conservative if a
	// future syntax adds a fourth position. Reporting a member as dead is an accusation about
	// working code, and an unrecognized reference should read as a use rather than as nothing.
	//
	// The brand check is genuinely a read: the class is asking whether an object carries this
	// member, which is the member doing work. A rule inspecting only property accesses cannot see
	// it and calls a live brand check dead.
	if parent.Kind != ast.KindPropertyAccessExpression {
		return true
	}

	// Reading or writing an accessor calls a function body either way, so any reference to one is a
	// use. This is decided from what the member *is* rather than from the reference, and it is the
	// reason a write-only setter is correct code rather than a finding.
	if isAccessor {
		return true
	}

	return !isWriteOnlyPrivateAccess(parent)
}

// isWriteOnlyPrivateAccess reports whether a `this.#name` access only ever stores, never loads.
//
// The parameter is the property access, not the name, because every question below is about what
// sits above the access in the tree.
//
// `ast.IsWriteAccess` is deliberately not the predicate here, for two reasons that both point the
// same way. It answers "does this store" rather than "does this *only* store", so it reports true
// for `this.#x += 1` and for `this.#x++`, both of which read. And it is known to under-report: a
// rest element in a destructuring target reads as no write at all, measured and documented on
// `prefer_const`. Under-reporting a write is harmless for `prefer-const` and wrong here, since a
// missed write becomes a phantom read and the member goes unreported. Enumerating the write forms
// directly avoids inheriting either behavior.
func isWriteOnlyPrivateAccess(access *ast.Node) bool {
	parent := access.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.Left != access || !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			// The right side of any assignment is a read, and so is either side of a non-assignment
			// operator. `foo = this.#x` and `this.#x + 1` both land here.
			return false
		}
		if binary.OperatorToken.Kind == ast.KindEqualsToken {
			// A plain store. The old value is never loaded, whatever the result is used for:
			// `foo = (this.#x = 1)` reads the value being *stored*, not the member.
			return true
		}
		// A compound or logical assignment loads before it stores, so it is a read unless the
		// result is thrown away. `this.#x += 1;` as a statement observes nothing; `return this.#x
		// += 1` observes the loaded value. This is the one place the same operator falls on both
		// sides of the line, and upstream carries four separate cases pinning it.
		return parent.Parent != nil && parent.Parent.Kind == ast.KindExpressionStatement

	case ast.KindPrefixUnaryExpression:
		operator := parent.AsPrefixUnaryExpression().Operator
		if operator != ast.KindPlusPlusToken && operator != ast.KindMinusMinusToken {
			// `!this.#x`, `-this.#x`, `typeof this.#x` all read.
			return false
		}
		return parent.Parent != nil && parent.Parent.Kind == ast.KindExpressionStatement

	case ast.KindPostfixUnaryExpression:
		// Both fixities, and both only write-only when discarded. `return this.#x++` reads the old
		// value and `++this.#x;` as a statement reads nothing that survives, so neither fixity nor
		// the statement test alone answers.
		return parent.Parent != nil && parent.Parent.Kind == ast.KindExpressionStatement

	case ast.KindForInStatement:
		// `for (this.#x in obj)` stores each key into the member and never loads it. The other
		// position, `for (const k in this.#x)`, is the object being enumerated and is a read.
		return parent.AsForInOrOfStatement().Initializer == access

	case ast.KindForOfStatement:
		return parent.AsForInOrOfStatement().Initializer == access

	case ast.KindArrayLiteralExpression:
		// A destructuring assignment target parses as an array *literal*: `[this.#x] = bar` stores
		// into the member. A genuine array value, `foo([this.#x])`, reads it, so the climb has to
		// reach an assignment before calling this a write.
		return isDestructuringAssignmentTarget(parent)

	case ast.KindSpreadElement:
		// `[...this.#x] = bar` is the rest element `ast.IsWriteAccess` cannot see, which is why the
		// write forms are enumerated here rather than delegated to it.
		return isDestructuringAssignmentTarget(parent)

	case ast.KindPropertyAssignment:
		// `({ x: this.#x } = bar)` stores into the member, but only from the value position. The
		// key position, `({ [this.#x]: a } = bar)`, is a computed key being read to decide which
		// property to take, and reaches here with a `ComputedPropertyName` in between rather than
		// directly, so the identity test is what separates the two.
		if parent.AsPropertyAssignment().Initializer != access {
			return false
		}
		return isDestructuringAssignmentTarget(parent)
	}

	return false
}

// isDestructuringAssignmentTarget reports whether a node sits on the left of a destructuring
// assignment.
//
// The climb exists because a destructuring target and an ordinary array or object value parse to
// the same node kinds, so the node alone cannot tell `[this.#x] = bar` from `foo([this.#x])`. Only
// the assignment above it distinguishes them, and it can be several levels up: `[{ a: this.#x }] =
// bar` reaches its operator through a property assignment inside an object literal inside an array
// literal.
//
// It stops at anything that is not one of those pass-through kinds, which is what keeps a spread
// argument (`foo(...this.#x)`) and a constructed value (`const o = { ...this.#x }`) reading as
// reads: neither climb reaches an assignment operator.
func isDestructuringAssignmentTarget(node *ast.Node) bool {
	current := node
	for current != nil {
		parent := current.Parent
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression,
			ast.KindPropertyAssignment, ast.KindSpreadElement, ast.KindParenthesizedExpression:
			current = parent
			continue
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			// Only the left side is a target. `foo = [this.#x]` builds a value from a read.
			return binary.Left == current && ast.IsAssignmentOperator(binary.OperatorToken.Kind)
		default:
			return false
		}
	}
	return false
}
