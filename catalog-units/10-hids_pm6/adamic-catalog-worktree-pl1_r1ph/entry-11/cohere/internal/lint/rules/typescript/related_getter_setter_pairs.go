package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimcore "github.com/microsoft/TypeScript/tsc/shim/core"
	shimscanner "github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"

	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
)

var messageRelatedGetterSetterPairs = rule.Message{
	Id: "mismatch",
	Description: "A getter and its setter name one property, so a caller that reads the property " +
		"and writes the value straight back has to type-check. The getter here returns a type the " +
		"setter will not accept, which makes that round trip an error and usually means one of the " +
		"two annotations is wrong. Widen the setter's parameter, or narrow what the getter returns.",
}

// RelatedGetterSetterPairs flags a getter whose return type is not assignable to the type its
// matching setter accepts.
//
//	valid:   interface E { get value(): string; set value(v: string); }
//	valid:   interface E { get value(): string; set value(v: string | undefined); }
//	valid:   interface E { get value(): string | undefined; set value(); }
//	valid:   class E { get value() { return ''; } set value(param: number) {} }
//	invalid: interface E { get value(): string | undefined; set value(v: string); }
//	invalid: type E = { get value(): number; set value(v: string) };
//	invalid: class E { get value(): boolean { return true; } set value(v: string) {} }
//
// Ported from `@typescript-eslint/related-getter-setter-pairs`, which defines the rule. The clone at
// `packages/eslint-plugin/src/rules/related-getter-setter-pairs.ts` and the installed 8.67.0 build in
// `node_modules` are byte-identical in behavior, and every claim below was measured by driving that
// installed build through the ESLint Linter API rather than read off the source.
//
// The corpus is `tests/rules/related-getter-setter-pairs.test.ts`: sixteen clean cases and seven
// reporting ones, each reporting case carrying an explicit line, column, endLine and endColumn. All
// twenty three were reproduced on the installed build before a line of this was written, which is
// what makes the invented probes below trustworthy.
//
// # The judgment
//
// Upstream keeps a stack of name-to-pair maps, pushed when a class body, an interface body or a type
// literal is entered and popped on exit. A getter is recorded only when it carries a return type
// annotation; a setter only when it declares exactly one parameter. On exit, for every name holding
// both, the getter's type must be assignable to the setter's parameter type or the getter's return
// type node is reported.
//
// Both filters are load-bearing and the corpus pins each of them: `get value() { return ”; }` with
// no annotation is clean beside a mismatching setter, and so are `set value()` and
// `set value(a: string, b: string)`. Neither is an optimization, since dropping either would report
// cases upstream leaves alone.
//
// This is written as a per-container listener rather than as a stack, because typescript-go hands us
// each container node and its members are reachable from it. That is the same decision expressed
// without the bookkeeping ESLint's one-node-at-a-time visitor forces, and it makes nesting fall out
// for free: an inner class inside an outer method gets its own pass and never sees the outer names.
// Measured against the installed build on a class holding a mismatching pair whose method body
// declares an inner class holding another, which reports twice, and on the same shape where only the
// inner mismatches, which reports once and points inside.
//
// # The container set, and the two shapes that are NOT in it
//
// `ClassBody`, `TSInterfaceBody` and `TSTypeLiteral` are the three, and both omissions were measured
// rather than assumed:
//
//	const o = { get value(): number { return 1; }, set value(v: string) {} };    clean
//	class C { get value(): number { return 1; } set value(v: string) {} }        reports
//
// An object literal's accessors are `Property` nodes under an `ObjectExpression`, which is in none of
// the three, so upstream never records them. In typescript-go they are the same `KindGetAccessor` and
// `KindSetAccessor` the class holds, distinguished only by their parent, so listening on the
// containers rather than on the accessors reproduces the gap by construction. A rule anchored on the
// accessor kind would have to grow a parent test to decline the object literal, and the natural way
// to write that rule would report it.
//
// A class expression's body is also a `ClassBody`, so `const C = class { ... }` reports. That is why
// the listener covers `KindClassExpression` beside `KindClassDeclaration`: the two are one node type
// in TSESTree and two here.
//
// # abstract accessors are silent, and this is a PARSER difference we have to reproduce by hand
//
// This is the single most surprising thing in the rule and no imported fixture can see it, because
// the corpus writes no `abstract` anywhere.
//
//	abstract class C { abstract get value(): number; abstract set value(v: string); }   clean
//	abstract class C { abstract get value(): number; set value(v: string) {} }          clean
//	abstract class C { get value(): number { return 1; } abstract set value(v: string); }  clean
//	declare class C { get value(): number; set value(v: string); }                     reports
//
// The mechanism is the selector. Upstream matches `:matches(MethodDefinition, TSMethodSignature)`,
// and typescript-eslint's parser gives an `abstract` accessor the node type
// `TSAbstractMethodDefinition`, which is neither. So an abstract accessor is never recorded on
// either side of the pair, and one abstract half silences the whole pair. A `declare class`
// accessor stays a plain `MethodDefinition` and reports, which is the control that proves the
// distinction is `abstract` rather than "has no body". Both verdicts were read off the parser
// directly as well as off the rule.
//
// typescript-go has no such node: an abstract accessor is a `KindGetAccessor` carrying an
// `abstract` modifier, so nothing about our tree reproduces the gap. It is reproduced explicitly
// below. That is fidelity to upstream's decision through a workaround we do not need, which is the
// case where fidelity wins: the gate compares against upstream, and reporting these would read as a
// defect.
//
// # Name keying, which is upstream's `getNameFromMember` and does NOT match the property shelf
//
// `internal/utilities/ecmascript/property.Name` is the obvious shelf function here and it is wrong for
// this rule in two directions, both measured. It DECLINES a bare identifier inside brackets, on the
// stated reasoning that `[a]` names whatever the variable holds; upstream keys that as the
// identifier's own name, so `get [k]()` and `set [k]()` DO pair and report. And it reads THROUGH the
// brackets for a literal, so `['lit']` and `'lit'` both answer `lit` while upstream keys the first
// as its inner literal's cooked value too. The first disagreement costs findings, the second does
// not, and only the first is visible from the corpus, which writes no computed key at all.
//
// So this reproduces `getNameFromMember` rather than reaching for the shelf. Read off the parser and
// the helper together on nine key spellings:
//
//	get plain()             plain          identifier, bare name
//	get #priv()             #priv          private identifier, hash included
//	get 'quoted'()          quoted         literal that IS a valid identifier, bare
//	get 'a-b'()             "a-b"          literal that is NOT, wrapped in double quotes
//	get 12()                "12"           numeric literal, cooked then wrapped
//	get 1e1()               "10"           cooked to the canonical rendering first
//	get [k]()               k              computed identifier, keyed by the identifier
//	get ['lit']()           lit            computed literal, keyed like a bare literal
//	get [Symbol.iterator]() Symbol.iterator  anything else, raw source text of the key
//
// The quoting is `requiresQuoting`, which is exactly "this string is not a valid identifier", so
// `scanner.IsIdentifierText` answers it inverted. `LanguageVariantStandard` rather than the JSX
// variant, because upstream calls `ts.isIdentifierPart` with no variant.
//
// The keying produces collisions that look like defects and are upstream's actual behavior, all
// measured on the installed build:
//
//	class C { get 'value'(): number {...} set value(v: string) {} }        reports
//	class C { get value(): number {...} set ["value"](v: string) {} }      reports
//	class C { get 12(): number {...} set '12'(v: string) {} }              reports
//	class C { static get value(): number {...} set value(v: string) {} }   reports
//
// The last is the one worth staring at. `static` is not part of the key at all, so a static getter
// pairs with an instance setter that can never be the same property. Reproduced rather than
// corrected.
//
// # Last write wins, which decides which of two duplicates is judged
//
// The map is written with `.set`, so a second getter under one name REPLACES the first and a second
// setter replaces the first. That is a behavioral decision hiding in a data-structure choice and it
// is invisible to any fixture asserting a count:
//
//	get value(): number; get value(): string; set value(v: string)   clean, second getter wins
//	get value(): string; get value(): number; set value(v: string)   reports, second getter wins
//	get value(): number; set value(v: number); set value(v: string)  reports, second setter wins
//	get value(): number; set value(v: string); set value(v: number)  clean, second setter wins
//
// All four measured. The natural Go spelling, writing into a map only when the key is absent, gets
// every one of them backwards.
//
// # The span, and the one place our tree differs
//
// The finding points at the getter's return type node, not at the annotation and not at the member.
// The corpus states this seven times with explicit columns, and the widest case spans
// `string | undefined` across eighteen columns while a multi-line object type spans three lines.
//
// A parenthesized return type is where the two trees part. TSESTree has no parenthesized-type node,
// so `get value(): (string)` reports on `string` with the parens excluded, measured. typescript-go
// keeps a `KindParenthesizedType` spanning `(string)`, so reporting `node.Type()` directly would
// point at a wider span than upstream on exactly that input. `ast.SkipTypeParentheses` unwraps it.
// The corpus writes no parenthesized type, so nothing imported can see this either way.
//
// # The checker
//
// Required. `isTypeAssignableTo` is not on the checker's exported surface, so it comes through the
// shim's linkname at `shim/checker/shim.go:136`, the same way `no_unsafe_unary_minus` and the
// builtins shelf reach their own upstream helpers. `GetTypeAtLocation` on the getter node answers
// the getter's return type and on a parameter node answers that parameter's type, which is what
// upstream asks for through its parser services. Probed rather than assumed, on an interface, a
// class and a type literal.
//
// The direction is getter-assignable-to-setter and it is not symmetric. `get value(): string` beside
// `set value(v: never)` reports while `get value(): never` beside `set value(v: string)` is clean,
// which is the pair that pins the direction, since `never` is assignable to everything and nothing
// but `never` is assignable to it. Both measured.
var RelatedGetterSetterPairs = rule.Rule{
	Name:             "@typescript-eslint/related-getter-setter-pairs",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A typed rule handed a nil checker goes silent rather than crashing, which is the more
		// dangerous of the two failure modes because a whole StaysSilent suite passes vacuously
		// over it. Kept at the top of the only listener, and it is also the `.TypeChecker`
		// selector the registry's declaration guard reads in this file.
		if ctx.TypeChecker == nil {
			return rule.Listeners{}
		}

		checkContainer := func(members []*ast.Node) {
			// Insertion order matters for nothing here, since every pair is judged independently,
			// but the ORDER OF REPORTS does: two mismatching pairs in one container report in the
			// order their getters were written, which a map iteration would randomize. So the
			// names are kept in a slice beside the map.
			type pair struct {
				getter *ast.Node
				setter *ast.Node
			}
			pairs := map[string]*pair{}
			var order []string
			record := func(name string, member *ast.Node, isGetter bool) {
				existing, seen := pairs[name]
				if !seen {
					existing = &pair{}
					pairs[name] = existing
					order = append(order, name)
				}
				// Assignment rather than a first-write-wins guard, because upstream's map is
				// written with `.set` and a second declaration under one name replaces the first.
				// See the last-write-wins note on the rule: all four orderings were measured and a
				// first-write-wins spelling gets every one of them backwards.
				if isGetter {
					existing.getter = member
				} else {
					existing.setter = member
				}
			}

			for _, member := range members {
				switch member.Kind {
				case ast.KindGetAccessor:
					// An abstract accessor is a different NODE TYPE in typescript-eslint's parser
					// and its selector never matches one, so upstream records neither half of a
					// pair that touches `abstract`. typescript-go gives us an ordinary accessor
					// with a modifier, so the gap has to be written. See the abstract note above;
					// `declare class` is the control and it reports.
					if isAbstractAccessor(member) {
						continue
					}
					// A getter with no return annotation is never recorded, which the corpus pins
					// twice: `get value() { return ''; }` is clean beside both a mismatching setter
					// and an empty one.
					if member.Type() == nil {
						continue
					}
					name, ok := relatedAccessorName(ctx, member)
					if !ok {
						continue
					}
					record(name, member, true)

				case ast.KindSetAccessor:
					if isAbstractAccessor(member) {
						continue
					}
					// Exactly one parameter, which the corpus pins from both sides: `set value()`
					// and `set value(a: string, b: string)` are both clean beside a mismatching
					// getter. A `this` parameter counts toward the total on both sides, measured,
					// so `set value(this: C, v: string)` is clean upstream and here.
					parameters := member.Parameters()
					if len(parameters) != 1 {
						continue
					}
					name, ok := relatedAccessorName(ctx, member)
					if !ok {
						continue
					}
					record(name, member, false)
				}
			}

			for _, name := range order {
				candidate := pairs[name]
				if candidate.getter == nil || candidate.setter == nil {
					continue
				}
				// The getter NODE, which is what upstream asks its parser services for. Asking the
				// getter's TYPE NODE instead is a mutant that survives, and it is equivalent rather
				// than uncovered: measured over sixteen annotation shapes (a keyword, `this`, a type
				// parameter, an alias, a readonly array, `typeof`, a literal type, a generic
				// instantiation, `keyof`, a parenthesized type, a union, an abstract member, an
				// interface member, and four paired with a mismatching setter) the two locations
				// answer identically.
				//
				// They can only diverge on an UNANNOTATED getter, where the node form takes the
				// setter's type by inference and the type node does not exist to ask. The
				// annotation filter above declines every one of those before this line, so the
				// divergence is behind an earlier guard rather than absent.
				getterType := ctx.TypeChecker.GetTypeAtLocation(candidate.getter)
				setterType := ctx.TypeChecker.GetTypeAtLocation(candidate.setter.Parameters()[0])
				// No nil check on either type, and its absence is measured rather than assumed. A
				// mutant deleting one survived the whole suite, which sent me to the source:
				// `GetTypeAtLocation` is `getTypeOfNode`, whose every failing path returns the
				// checker's error type rather than nil. Probed over eight malformed sources chosen
				// to make the parser recover (`get ()`, `get }`, `get []()`, an unterminated return
				// type, a numeric key that overflows) and it never answered nil.
				//
				// The callers this verdict was taken over are the four container listeners at the
				// bottom of this file, and nothing else calls into here. Adding a caller that can
				// reach this line with a node the checker has never seen voids the verdict.
				if shimchecker.Checker_isTypeAssignableTo(ctx.TypeChecker, getterType, setterType) {
					continue
				}
				// The getter's return TYPE, unwrapped past parentheses. TSESTree has no
				// parenthesized-type node, so upstream reports `string` for `(string)` and
				// reporting `member.Type()` here would span the parens instead. See the span note.
				ctx.ReportNode(ast.SkipTypeParentheses(candidate.getter.Type()), messageRelatedGetterSetterPairs)
			}
		}

		return rule.Listeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				checkContainer(node.Members())
			},
			// A class expression's body is a ClassBody too, so `const C = class { ... }` reports.
			// One node type upstream, two here.
			ast.KindClassExpression: func(node *ast.Node) {
				checkContainer(node.Members())
			},
			ast.KindInterfaceDeclaration: func(node *ast.Node) {
				checkContainer(node.Members())
			},
			ast.KindTypeLiteral: func(node *ast.Node) {
				checkContainer(node.Members())
			},
		}
	},
}

// isAbstractAccessor reports whether a class member carries the `abstract` modifier.
//
// It exists because typescript-eslint's parser encodes this as a distinct node type,
// `TSAbstractMethodDefinition`, that the rule's selector excludes, while typescript-go encodes it as
// a modifier on the ordinary accessor. Reproducing the exclusion is the whole reason this function
// is here; see the abstract note on the rule for the four inputs that pin it.
func isAbstractAccessor(member *ast.Node) bool {
	modifiers := member.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindAbstractKeyword {
			return true
		}
	}
	return false
}

// relatedAccessorName is upstream's `getNameFromMember`, which decides when two accessors are the
// same property.
//
// Reproduced rather than delegated to `property.Name`, which disagrees in both directions on
// computed keys. The nine spellings this was measured on are listed on the rule.
//
// The second return separates "no name the syntax settles" from an empty one. Every arm answers true
// today, and the signature is kept because the four callers below read it and a future arm may not.
//
// # The nil-name and nil-inner guards were here and were removed as unreachable
//
// Both looked load-bearing and neither is, and mutants deleting them survived the whole suite, which
// is what sent me to probe the parse rather than add a fixture. Measured over eight malformed
// sources written to make the parser's error recovery synthesize nodes that valid source cannot:
//
//	class C { get (): number {} }        parses as a METHOD named `get`, not an accessor at all
//	class C { get }                      the same, no accessor node is produced
//	class C { get [](): number {} }      an accessor whose computed key holds a SYNTHESIZED
//	                                     identifier, never a nil expression
//
// So no accessor node reaches here with a nil name, and no computed key reaches here with a nil
// inner expression. The callers this was taken over are the two arms of the member loop in the rule
// above and nothing else; if a caller is added that hands this a node from somewhere other than a
// container's member list, the verdict is void and both guards have to be re-argued.
func relatedAccessorName(ctx rule.Context, member *ast.Node) (string, bool) {
	name := member.Name()

	// A computed key in TSESTree is not a wrapper node: `member.key` IS the inner expression, so
	// every branch below applies to `[x]` exactly as it applies to `x`. Read off the parser rather
	// than inferred, on six computed spellings.
	if name.Kind == ast.KindComputedPropertyName {
		inner := name.AsComputedPropertyName().Expression
		return relatedKeyName(ctx, inner)
	}
	return relatedKeyName(ctx, name)
}

// relatedKeyName answers the key string for one key expression, matching `getNameFromMember`'s four
// arms.
func relatedKeyName(ctx rule.Context, key *ast.Node) (string, bool) {
	switch key.Kind {
	case ast.KindIdentifier:
		return key.Text(), true

	case ast.KindPrivateIdentifier:
		// `Text()` already carries the leading hash here, the same way TSESTree's arm prepends one
		// to a name that does not, so `#p` answers `#p` on both sides.
		//
		// A mutant doubling the hash survives, and it is equivalent rather than uncovered. The
		// output of this arm lives in a namespace nothing else can reach: enumerated over six
		// spellings of the same name, `'#p'` and `["#p"]` key as a quoted `"#p"`, a template keys
		// as its backticked source, a member access keys as its own text, and only a private
		// identifier produces a bare leading hash. So no other key can collide with this one, and
		// any prefix applied here moves both accessors together.
		return key.Text(), true

	case ast.KindStringLiteral, ast.KindNumericLiteral:
		// `${member.key.value}` on the JavaScript side, which cooks the literal: a numeric key is
		// rendered as its canonical number, so `1e1` and `10` are one key. `Text()` on our
		// numeric literal already holds that same canonical rendering, measured.
		text := key.Text()
		// `requiresQuoting` is exactly "not a valid identifier", so the wrap is its inverse.
		//
		// The PREDICATE decides real verdicts and a mutant inverting it is caught: without it,
		// `get 'a-b'()` would key as `a-b` and stop pairing with `set ["a-b"]()`, which upstream
		// reports. The WRAP CHARACTER is a different question and the answer is that nothing can
		// see it, which took an enumeration rather than an argument.
		//
		// A different character could only change a verdict by making a quoted key collide, or stop
		// colliding, with a key that went through the raw-text arm below. Both arms move together
		// for a literal, so the only exposure is a NON-literal key whose entire source text is a
		// bare double-quoted string. Enumerated over twelve shapes that reach the raw-text arm
		// (template, member access, call, non-null, optional chain, unary, binary, `as`,
		// `satisfies`, and a bracketed member access): every one of them that contains a quote
		// carries more source around it, because a bare `"a-b"` parses as a Literal and takes the
		// arm above instead. So no input distinguishes the wrap characters, and a mutant changing
		// them survives correctly rather than for want of a fixture.
		if !shimscanner.IsIdentifierText(text, shimcore.LanguageVariantStandard) {
			return `"` + text + `"`, true
		}
		return text, true
	}

	// Everything else is keyed by the key's own source text: a member expression like
	// `Symbol.iterator`, a template literal, a call. Two such keys pair when they are spelled
	// identically and not otherwise, which is upstream's behavior and not a semantic comparison.
	//
	// `Pos()` includes leading trivia here, which upstream's `range` does not, so a key written
	// after a newline would key with the whitespace attached and stop pairing with the same key
	// written inline. `TrimNodeTextRange` is the same token-start walk `ReportNode` uses.
	trimmed := type_checking.TrimNodeTextRange(ctx.SourceFile, key)
	return ctx.SourceFile.Text()[trimmed.Pos():trimmed.End()], true
}
