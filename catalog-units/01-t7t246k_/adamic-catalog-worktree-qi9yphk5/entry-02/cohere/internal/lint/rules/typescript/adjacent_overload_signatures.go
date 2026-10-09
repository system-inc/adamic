package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimcore "github.com/microsoft/TypeScript/tsc/shim/core"
	shimscanner "github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageAdjacentOverloadSignatures is upstream's `adjacentSignature`, whose text names the member.
//
// `rule.Message` has no interpolation layer, so upstream's `{{name}}` becomes concatenation here.
// The name is worth carrying rather than dropping: a container holding two interleaved overload sets
// produces two findings under one id, and the name is the only thing in the text that separates
// them. Upstream prefixes `static ` for a static member and this reproduces that, because a static
// and an instance member of the same name are different methods to this rule and a message that
// could not tell them apart would be describing the wrong one half the time.
func messageAdjacentOverloadSignatures(name string) rule.Message {
	return rule.Message{
		Id: "adjacentSignature",
		Description: "This signature for " + name + " is separated from the others by unrelated " +
			"members, so a reader scrolling the declaration sees an overload set that looks " +
			"complete and is not. Move every signature for " + name + " together, in the order " +
			"the compiler resolves them, so the whole set can be read at once.",
	}
}

// AdjacentOverloadSignatures requires that a member's overload signatures be written consecutively.
//
//	valid:   interface I { foo(): void; foo(a: number): void; bar(): void; }
//	valid:   class C { static foo(): void {} bar(): void {} foo(): void {} }   static is a different member
//	valid:   interface I { foo(): void; }                                       one signature is a set of one
//	invalid: interface I { foo(): void; bar(): void; foo(a: number): void; }
//	invalid: class C { foo(): void {} bar(): void {} foo(a: number): void {} }
//	invalid: declare function foo(): void; declare function bar(): void; declare function foo(a: number): void;
//
// Ported from `@typescript-eslint/adjacent-overload-signatures`, reading the clone at
// `packages/eslint-plugin/src/rules/adjacent-overload-signatures.ts` together with the
// `getNameFromMember` helper it delegates to, and measuring every verdict against the installed
// 8.67.0 build driven through the ESLint 10.8.1 Linter API.
//
// # The algorithm, which is not "compare adjacent pairs"
//
// The name suggests a pairwise walk and the implementation is not one. Each container keeps a list
// of every member name SEEN SO FAR and the name of the IMMEDIATELY PRECEDING member. A member
// reports when its name is already in the seen list and it does not match the preceding member,
// which is exactly "this set was closed and has been reopened". A name is appended to the seen list
// only the first time it appears.
//
// Two consequences a pairwise reading gets wrong, both measured:
//
//	foo, bar, foo, foo      ONE finding, on the third member. The fourth matches its predecessor,
//	                        so the set is contiguous again from there and only the reopening is
//	                        reported.
//	foo, bar, foo, bar      TWO findings, one per reopening.
//
// A member that is not a method at all (a property, a statement, an import) sets the preceding name
// to nothing rather than being skipped, which is what makes a property between two signatures
// separate them. Reproducing that assignment is load-bearing, and a port that merely skipped such a
// member would be silent on upstream's own `foo, name: string, foo` case.
//
// # Six containers, and typescript-go spells three of them differently
//
// Upstream listens on `BlockStatement`, `ClassBody`, `Program`, `TSInterfaceBody`, `TSModuleBlock`
// and `TSTypeLiteral`. Here a class's and an interface's members hang off the declaration rather
// than off a separate body node, so `KindClassDeclaration`, `KindClassExpression`,
// `KindInterfaceDeclaration` and `KindTypeLiteral` are read through `Members()` while
// `KindSourceFile`, `KindBlock` and `KindModuleBlock` are read through `Statements()`. That is eight
// listeners for upstream's six because a class expression is a separate kind here and a class body
// is not; the containers reached are the same set.
//
// `Members()` PANICS on a kind with no member list, and `Statements()` likewise. The listener map is
// what keeps each call on a kind that has one, so the split above is a crash guard rather than a
// tidy grouping. Probed: `Members()` on a `KindSourceFile` panics with
// `Unhandled case in Node.MemberList`, and the walk recovers per file rather than per rule, so one
// such panic would cost every rule that file.
//
// # There is no export wrapper here, which removes upstream's recursion and keeps its span
//
// Upstream unwraps `ExportNamedDeclaration` and `ExportDefaultDeclaration` to reach the declaration
// inside, and returns nothing for an export with no declaration such as `export { a }`. typescript-go
// has no such wrapper: `export function foo() {}` is a `KindFunctionDeclaration` carrying an
// `KindExportKeyword` modifier, and `export { a }` is a `KindExportDeclaration` that is not a
// function at all and falls through the switch below. So the recursion has nothing to unwrap and is
// simply absent, which is a parser difference rather than a narrowing.
//
// The span survives that difference and it was measured rather than assumed. Upstream reports the
// wrapper, so its finding covers the `export` keyword: `export function foo(a: number): void {}` is
// reported whole, and `export default function foo() {}` likewise. Here the modifier is part of the
// same node, and `ReportNode` anchors on the node's first token, so the reported text is identical.
//
// # A name is not a string, and the four ways it is built
//
// `getNameFromMember` returns a name AND a kind, and two members are the same only when both agree.
// The kind exists so that a key needing quotes cannot collide with one that does not. Read off the
// installed build on eight spellings:
//
//	foo(): void            foo           identifier, bare
//	'foo'(): void          foo           literal that IS a valid identifier, keyed identically
//	'a-b'(): void          "a-b"         literal that is NOT, wrapped in double quotes
//	12(): void             "12"          numeric literal, cooked, then wrapped because it needs it
//	#p(): void             #p            private identifier, hash included
//	[k](): void            k             computed identifier, keyed by the identifier itself
//	['lit'](): void        lit           computed literal, keyed like a bare literal
//	[Symbol.iterator]()    Symbol.iterator   anything else, the key's own source text
//
// So `'foo'` and `foo` ARE the same member and report, while `'a-b'` and a hypothetical bare `a-b`
// cannot both exist. The kind is carried anyway because it is upstream's, and because a mutant
// dropping it is caught by the quoted cases.
//
// # `static` is part of the identity, it is THREE-valued, and a sibling rule ignores it entirely
//
// A static and an instance member of the same name are different methods, so
// `static foo, bar, foo` is CLEAN. Worth stating because `related-getter-setter-pairs` in this same
// package makes the opposite choice: there `static` is not part of the key at all, so a static
// getter pairs with an instance setter. Both are upstream's own behavior for their own rule, and
// the two must not be reconciled.
//
// The three-valued part is the trap, and a Go port written from the obvious reading gets four
// verdicts wrong. Upstream's field is `boolean | undefined` compared with `===`: the member arms
// spread `static: member.static`, which is `false` even on an interface member where static is not
// sayable, while the function-declaration, call-signature and construct-signature arms build their
// object with no such key at all. So `undefined !== false` keeps a construct signature apart from a
// member spelled `new`, which nothing else would, because `new` carries no call-signature flag.
// The field's own note carries the five measured inputs, three clean and two controls.
//
// # TSESTree folds four member shapes into one, and one of them is an EXCLUSION
//
// `MethodDefinition` covers a method, a getter, a setter and a constructor there; here those are
// four kinds. All four are methods to this rule, a getter and a setter share one name with each
// other and with a plain method of that name, and a constructor is named `constructor`. Every one
// measured on the installed build; the table is at the arm that handles them.
//
// `TSAbstractMethodDefinition` is a FIFTH type there and upstream's switch does not name it, so an
// abstract member is not a method at all: it falls through and clears the preceding name the way a
// property does. Here `abstract` is a modifier on the ordinary member, so the exclusion has to be
// written rather than inherited, and it was measured with a concrete control so that "silent" could
// be told apart from "never reached".
//
// # A call signature and a construct signature are named, not skipped, and the flag is INERT
//
// `(): void` is the name `call` with a call-signature flag set, and `new (): void` is the name `new`
// with the flag CLEAR. The flag reads as the thing keeping a call signature from colliding with a
// member literally named `call`, and the asymmetry looks like upstream forgetting to protect `new`
// the same way. Both readings are wrong here, and only a survivor made that visible.
//
// Mutants clearing the flag on the call arm and setting it on the construct arm BOTH survive, and
// both are correct survivals. The staticness field already separates these shapes from any method:
// a call signature and a construct signature are unset, and every method and accessor is instance,
// so the two can never be equal whatever the flag says. For the flag to decide anything, a call
// signature would have to share a container with another unset-staticness member of the name
// `call`, and the only other unset shape is a function declaration. A call signature is sayable
// only in an interface body or a type literal; a function declaration only in a program, a block,
// or a module block. **They can never be siblings**, so no input exists.
//
// Measured rather than argued, in both parsers, because a grammar claim is exactly the kind that is
// true of the grammar and false of the parser. The ESLint parser rejects all seven crossings
// outright: `(): void;` at the top level, in a block, and in a module block; `function foo(): void;`
// in an interface and in a type literal; `new (): void;` at the top level and in a module block.
// Ours recovers instead of rejecting, and the recovery does not produce the shape either. A stray
// `(): void;` becomes an `ExpressionStatement` holding an `ArrowFunction`, never a
// `KindCallSignature`, and a `function` written inside an interface closes the interface early
// rather than joining its members.
//
// The flag is reproduced anyway because it is upstream's, but it is recorded as inert rather than
// left to read as load-bearing. The verdict names the containers this rule listens on: adding a
// listener for a container that could hold both shapes would void it.
var AdjacentOverloadSignatures = rule.Rule{
	Name: "@typescript-eslint/adjacent-overload-signatures",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		checkMembers := func(node *ast.Node) {
			reportAdjacentOverloadSignatures(ctx, node.Members())
		}
		checkStatements := func(node *ast.Node) {
			reportAdjacentOverloadSignatures(ctx, node.Statements())
		}

		return rule.Listeners{
			ast.KindClassDeclaration:     checkMembers,
			ast.KindClassExpression:      checkMembers,
			ast.KindInterfaceDeclaration: checkMembers,
			ast.KindTypeLiteral:          checkMembers,
			ast.KindSourceFile:           checkStatements,
			ast.KindBlock:                checkStatements,
			ast.KindModuleBlock:          checkStatements,
		}
	},
}

// adjacentOverloadMethod is upstream's `Method`: what makes two members the same overload set.
//
// All four fields participate in the comparison. `name` alone is not enough, and each of the other
// three has a corpus case or a measured probe behind it: `nameKind` separates a quoted key from a
// bare one, `isStatic` separates a static member from an instance one, and `isCallSignature`
// separates `(): void` from a member named `call`.
type adjacentOverloadMethod struct {
	name     string
	nameKind adjacentOverloadNameKind

	// staticness is THREE-valued, and that is not over-engineering. Upstream's `static` field is
	// `boolean | undefined`, and its comparison is `===`, so `undefined` does not equal `false`.
	// Only the member arms set it at all: a function declaration, a call signature and a construct
	// signature each build their object without the key, leaving it undefined, while a method
	// spreads `static: member.static`, which TSESTree fills with `false` on an interface member
	// that cannot be static at all.
	//
	// A Go `bool` collapses `undefined` onto `false` and gets four verdicts wrong. Measured on the
	// installed build, all four clean upstream and all four reported by the two-valued version:
	//
	//	interface I { new (): void; foo(): void; "new"(): void; }   construct signature vs method
	//	interface I { "new"(): void; foo(): void; new(): void; }    the same pair, other order
	//	interface I { (): void; foo(): void; call(): void; }        call signature vs method
	//
	// with the controls that separate the quirk from a name difference:
	//
	//	interface I { new (): void; foo(): void; new (a: number): void; }   REPORTS, two signatures
	//	interface I { "new"(): void; foo(): void; ["new"](): void; }        REPORTS, two methods
	//
	// The call signature is additionally protected by its own flag, so that pair would be clean
	// either way; the construct signature has no such flag and this field is the only thing keeping
	// it apart from a member spelled `new`.
	staticness adjacentOverloadStaticness

	isCallSignature bool
}

// adjacentOverloadStaticness is upstream's `boolean | undefined` for the `static` field.
type adjacentOverloadStaticness string

const (
	// adjacentOverloadStaticnessUnset is upstream's `undefined`, for a member shape whose object
	// literal omits the key: a function declaration, a call signature, a construct signature.
	//
	// Written out at all three sites rather than left to the Go zero value, and the reason is a
	// measurement rather than style. It was originally implicit, and a mutant rewriting the
	// INSTANCE arm to this value then SURVIVED the whole suite: the mutation moved `Instance` onto
	// a constant nothing assigned, which was still distinct from the empty string the three arms
	// were actually producing, so the two values that upstream conflates stayed apart by accident.
	// The rule was right and the sweep could not see it, which is the shape a later reader would
	// have had to rediscover. With the constant assigned, the same mutant is caught.
	adjacentOverloadStaticnessUnset adjacentOverloadStaticness = "Unset"

	// adjacentOverloadStaticnessInstance is upstream's `false`, which TSESTree writes on every
	// method and accessor including one in an interface, where `static` is not sayable.
	adjacentOverloadStaticnessInstance adjacentOverloadStaticness = "Instance"

	// adjacentOverloadStaticnessStatic is upstream's `true`.
	adjacentOverloadStaticnessStatic adjacentOverloadStaticness = "Static"
)

// adjacentOverloadNameKind is upstream's `MemberNameType`.
//
// It is carried rather than folded into the name because upstream carries it, and it is measured
// to be INERT today. A mutant making the quoted arm answer `Normal` survives the whole suite, and
// that is a correct survival rather than a fixture gap: for the field to decide anything, two
// members would have to key to the same STRING under different kinds, and none can.
//
// Enumerated rather than argued, over fifteen key spellings covering all four arms
// (`foo`, `"foo"`, `"a-b"`, `12`, `"12"`, `[foo]`, `["foo"]`, `["a-b"]`, `[a.b]`, a template, a
// call, `[""]`, a non-null assertion, a parenthesized identifier, and an `as` expression): no name
// string is reachable with more than one kind. The quoted arm wraps in double quotes, which a bare
// identifier and a private name cannot contain, and the expression arm cannot produce a bare
// quoted string because such a key parses as a Literal and takes the quoted arm instead.
//
// So this is redundant with the current wrap characters rather than with the rule, which is exactly
// why it is reproduced: the argument for dropping it rests on a detail either side could change,
// and the next reader should find the measurement rather than repeat it.
type adjacentOverloadNameKind string

const (
	adjacentOverloadNameNormal     adjacentOverloadNameKind = "Normal"
	adjacentOverloadNamePrivate    adjacentOverloadNameKind = "Private"
	adjacentOverloadNameQuoted     adjacentOverloadNameKind = "Quoted"
	adjacentOverloadNameExpression adjacentOverloadNameKind = "Expression"
)

// reportAdjacentOverloadSignatures is upstream's `checkBodyForOverloadMethods`.
//
// The two pieces of state are upstream's and their difference is the whole algorithm. `seen` is
// every distinct method met so far in this container and only grows. `previous` is the method
// immediately before the current one, and it is cleared by any member that is not a method at all.
func reportAdjacentOverloadSignatures(ctx rule.Context, members []*ast.Node) {
	var seen []adjacentOverloadMethod
	var previous *adjacentOverloadMethod

	for _, member := range members {
		method, isMethod := adjacentOverloadMethodOf(ctx, member)
		if !isMethod {
			// Upstream's `lastMethod = null; return;`. Clearing rather than skipping is what makes
			// a property between two signatures separate them, and it decides several of upstream's
			// own reporting cases.
			previous = nil
			continue
		}

		alreadySeen := false
		for _, candidate := range seen {
			if candidate == method {
				alreadySeen = true
				break
			}
		}

		switch {
		case alreadySeen && (previous == nil || *previous != method):
			ctx.ReportNode(member, messageAdjacentOverloadSignatures(
				adjacentOverloadDisplayName(method)))
		case !alreadySeen:
			seen = append(seen, method)
		}

		// Assigned on every iteration that reached here, including the reporting one. A reporting
		// member becomes the predecessor of the next, which is what makes `foo, bar, foo, foo`
		// report once rather than twice.
		current := method
		previous = &current
	}
}

// adjacentOverloadDisplayName is upstream's static-prefixing template: a static member's name is
// rendered with a leading `static ` and every other member's is rendered bare.
func adjacentOverloadDisplayName(method adjacentOverloadMethod) string {
	// Upstream's test is truthiness, so both `false` and `undefined` take the bare branch and only
	// a genuinely static member is prefixed.
	if method.staticness == adjacentOverloadStaticnessStatic {
		return "static " + method.name
	}
	return method.name
}

// adjacentOverloadMethodOf is upstream's `getMemberMethod`, returning false for a member the rule
// does not judge.
//
// Upstream's export-unwrapping arms have no counterpart here and their absence is explained on the
// rule: typescript-go carries `export` as a modifier on the declaration rather than as a wrapping
// node, so there is nothing to unwrap, and `export { a }` is a kind this switch does not name.
func adjacentOverloadMethodOf(ctx rule.Context, member *ast.Node) (adjacentOverloadMethod, bool) {
	switch member.Kind {
	// Upstream's `TSDeclareFunction` and `FunctionDeclaration` are one kind here: a body-less
	// overload signature and an implementation are both `KindFunctionDeclaration`, distinguished by
	// whether `Body` is set, which this rule never asks.
	case ast.KindFunctionDeclaration:
		name := member.Name()
		if name == nil {
			// `export default function () {}` has no name. Upstream returns null for the same
			// reason, on `member.id?.name ?? null`.
			return adjacentOverloadMethod{}, false
		}
		return adjacentOverloadMethod{
			name:     name.Text(),
			nameKind: adjacentOverloadNameNormal,
			// Explicitly Unset rather than left to the Go zero value, matching upstream's object
			// literal, which has no `static` key on this arm. See the field's own note, and the
			// note on the constant for why writing it out is not ceremony.
			staticness: adjacentOverloadStaticnessUnset,
		}, true

	// Upstream's `TSMethodSignature` and `MethodDefinition`. TSESTree folds four things into
	// `MethodDefinition` that typescript-go keeps apart, so this arm is four kinds where upstream
	// has one, and every one of them was measured rather than inferred from the fold:
	//
	//	class C { get x() {...} f() {} get x() {...} }        reports, so an accessor is a method
	//	class C { get x() {...} f() {} set x(v) {} }          reports, so get and set share a name
	//	class C { x() {} f() {} get x() {...} }               reports, so a method and an accessor do too
	//	class C { constructor(); f() {} constructor(a) {} }   reports as `constructor`
	//
	// A constructor has no key node here, so its name is written out. Upstream reads it off
	// `member.key`, which TSESTree fills with an identifier spelled `constructor`, and that spelling
	// is load-bearing rather than incidental: measured, a member keyed `["constructor"]` collides
	// with a real constructor and reports.
	case ast.KindMethodSignature, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		// An abstract member is excluded, and this is the one arm where an exclusion rather than an
		// inclusion is being ported. TSESTree gives it a distinct `TSAbstractMethodDefinition` type
		// that upstream's switch does not name, so it is not a method to this rule at all.
		// typescript-go carries `abstract` as a modifier on the ordinary member, so the exclusion
		// has to be written. Measured with a control, because "silent" alone would not have
		// distinguished an exclusion from a rule that simply did not reach the shape:
		//
		//	abstract class C { abstract foo(): void; bar() {} abstract foo(a: number): void; }
		//	                                                        CLEAN, so abstract is excluded
		//	abstract class C { foo() {} abstract bar(): void; foo(a: number) {} }
		//	                                                        REPORTS, so it clears the
		//	                                                        predecessor like any non-method
		//	abstract class C { foo() {} bar() {} foo(a: number) {} }
		//	                                                        REPORTS, the concrete control
		//
		// The same exclusion is reproduced by `related-getter-setter-pairs` in this package, for the
		// same reason and against the same parser difference.
		if adjacentOverloadHasModifier(member, ast.KindAbstractKeyword) {
			return adjacentOverloadMethod{}, false
		}

		if member.Kind == ast.KindConstructor {
			return adjacentOverloadMethod{
				name:       "constructor",
				nameKind:   adjacentOverloadNameNormal,
				staticness: adjacentOverloadStaticnessOf(member),
			}, true
		}

		name, nameKind, named := adjacentOverloadNameOf(ctx, member.Name())
		if !named {
			return adjacentOverloadMethod{}, false
		}
		return adjacentOverloadMethod{
			name:       name,
			nameKind:   nameKind,
			staticness: adjacentOverloadStaticnessOf(member),
		}, true

	case ast.KindCallSignature:
		return adjacentOverloadMethod{
			name:            "call",
			nameKind:        adjacentOverloadNameNormal,
			isCallSignature: true,
			staticness:      adjacentOverloadStaticnessUnset,
		}, true

	case ast.KindConstructSignature:
		// `new` with the call-signature flag CLEAR, which is upstream's asymmetry rather than an
		// oversight here. See the rule's note.
		return adjacentOverloadMethod{
			name:     "new",
			nameKind: adjacentOverloadNameNormal,
			// This arm is where the unset value matters most: `new` carries no call-signature flag,
			// so staticness is the only thing separating a construct signature from a member spelled
			// `new`. See the field's note.
			staticness: adjacentOverloadStaticnessUnset,
		}, true
	}

	return adjacentOverloadMethod{}, false
}

// adjacentOverloadStaticnessOf is upstream's `member.static` for the member arms.
//
// TSESTree writes `false` rather than `undefined` on every method and accessor, including one in an
// interface where `static` cannot be written, so this never answers unset. The unset value belongs
// to the arms that build their object without the key at all; see the field's own note.
func adjacentOverloadStaticnessOf(member *ast.Node) adjacentOverloadStaticness {
	if adjacentOverloadHasModifier(member, ast.KindStaticKeyword) {
		return adjacentOverloadStaticnessStatic
	}
	return adjacentOverloadStaticnessInstance
}

// adjacentOverloadHasModifier reports whether a member carries one modifier keyword.
//
// Two callers, asking two different questions that TSESTree answers two different ways.
// `static` is a boolean field on a `MethodDefinition` there and a modifier here; `abstract` is a
// distinct NODE TYPE there and a modifier here. Both become the same question against this parser,
// which is why one helper serves both, and both are reproductions of a parser difference rather
// than of a decision.
//
// A `KindMethodSignature` can carry neither, so this answers false for every interface member
// without needing to ask the kind.
func adjacentOverloadHasModifier(member *ast.Node, keyword ast.Kind) bool {
	modifiers := member.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == keyword {
			return true
		}
	}
	return false
}

// adjacentOverloadNameOf is upstream's `getNameFromMember`, over a member's key node.
//
// The eight spellings it answers are tabled on the rule, each read off the installed build rather
// than off this code. `related-getter-setter-pairs` in this package reproduces the same helper for
// its own rule and the two agree; they are not shared because the helper is small and because
// sharing would invite a later caller to "fix" one rule's behavior into the other's, and the two
// rules disagree about `static` on purpose.
func adjacentOverloadNameOf(ctx rule.Context, key *ast.Node) (string, adjacentOverloadNameKind, bool) {
	if key == nil {
		return "", "", false
	}

	// A computed key is read THROUGH the brackets, because TSESTree's `member.key` for a computed
	// member is the inner expression rather than the bracket node. Measured: `[k]()` keys as `k` and
	// `['lit']()` keys as `lit`, so a computed key can collide with a plain one.
	if key.Kind == ast.KindComputedPropertyName {
		key = key.AsComputedPropertyName().Expression
		if key == nil {
			return "", "", false
		}
	}

	switch key.Kind {
	case ast.KindIdentifier:
		return key.Text(), adjacentOverloadNameNormal, true

	case ast.KindPrivateIdentifier:
		// `Text()` already carries the leading hash here, the same way TSESTree's arm prepends one
		// to a name that does not, so `#p` answers `#p` on both sides.
		return key.Text(), adjacentOverloadNamePrivate, true

	case ast.KindStringLiteral, ast.KindNumericLiteral:
		// `${member.key.value}` on the JavaScript side, which cooks the literal: a numeric key is
		// rendered as its canonical number, so `1e1` and `10` are one key. `Text()` on our numeric
		// literal already holds that same canonical rendering.
		text := key.Text()
		// `requiresQuoting` is exactly "this string is not a valid identifier", so
		// `IsIdentifierText` answers it inverted. `LanguageVariantStandard` rather than the JSX
		// variant, because upstream calls `ts.isIdentifierPart` with no variant.
		if !shimscanner.IsIdentifierText(text, shimcore.LanguageVariantStandard) {
			return `"` + text + `"`, adjacentOverloadNameQuoted, true
		}
		return text, adjacentOverloadNameNormal, true
	}

	// Everything else is keyed by the key's own source text: a member access like `Symbol.iterator`,
	// a template literal, a call. Two such keys match when they are spelled identically and not
	// otherwise, which is a text comparison rather than a semantic one, and is upstream's.
	//
	// `Pos()` includes leading trivia, which upstream's `range` does not, so a key written after a
	// newline would carry the whitespace and stop matching the same key written inline.
	// `TrimNodeTextRange` is the same token-start walk `ReportNode` uses.
	trimmed := type_checking.TrimNodeTextRange(ctx.SourceFile, key)
	return ctx.SourceFile.Text()[trimmed.Pos():trimmed.End()], adjacentOverloadNameExpression, true
}
