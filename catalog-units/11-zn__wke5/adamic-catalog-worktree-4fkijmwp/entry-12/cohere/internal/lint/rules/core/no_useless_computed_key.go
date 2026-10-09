package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUselessComputedKeySettings is the decoded option surface.
//
// Upstream's schema is one object with a single boolean, and its `defaultOptions` sets
// `enforceForClassMembers: true`. The default is the trap here: it is TRUE, so a decoder handing
// back a zero-valued struct would silently switch off every class member while leaving object
// literals enforced, and every fixture built from a struct rather than routed through the decoder
// would pass anyway.
type NoUselessComputedKeySettings struct {
	// EnforceForClassMembers extends the rule from object literals to class members.
	EnforceForClassMembers bool
}

// DefaultNoUselessComputedKeySettings is upstream's `defaultOptions: [{enforceForClassMembers: true}]`.
func DefaultNoUselessComputedKeySettings() NoUselessComputedKeySettings {
	return NoUselessComputedKeySettings{EnforceForClassMembers: true}
}

// noUselessComputedKeyWire is the on-the-wire shape, with a pointer so an absent key stays
// distinguishable from an explicit false.
type noUselessComputedKeyWire struct {
	EnforceForClassMembers *bool `json:"enforceForClassMembers"`
}

// DecodeNoUselessComputedKeyOptions reads the option object off the config.
//
// Hand rolled rather than routed through `rule.DecodeOptionsInto` because the option defaults to
// TRUE. The generic helper yields the zero value on an absent or empty option, which here means
// `enforceForClassMembers: false` -- not the rule disabled, but the rule quietly narrowed to object
// literals. Upstream's corpus tests `enforceForClassMembers: void 0` explicitly and expects the
// class member to report, so the distinction is asserted upstream as well.
//
// The pointer field is what keeps an explicit `false` reachable. Without it an absent key and a
// written `false` decode identically and the `void 0` case above cannot be told from the
// `false` case below it.
func DecodeNoUselessComputedKeyOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return DefaultNoUselessComputedKeySettings(), nil
	}

	var wire noUselessComputedKeyWire
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return DefaultNoUselessComputedKeySettings(), err
	}
	settings := DefaultNoUselessComputedKeySettings()
	if wire.EnforceForClassMembers != nil {
		settings.EnforceForClassMembers = *wire.EnforceForClassMembers
	}
	return settings, nil
}

// messageUnnecessarilyComputedProperty is the shape of the finding. The rendered Description carries
// the key's source text, so the value below is the template rather than anything reported.
var messageUnnecessarilyComputedProperty = rule.Message{
	Id: "unnecessarilyComputedProperty",
	Description: "This property key is wrapped in brackets that compute nothing. The brackets say " +
		"the name is decided when the code runs, so a reader stops to work out what it evaluates " +
		"to, and the answer is the literal already written inside them.",
}

// NoUselessComputedKey flags a computed property key whose brackets could be dropped without
// changing what property is named.
//
//	valid:   ({ [x]: 0 })
//	valid:   ({ ['__proto__']: [] })
//	valid:   class Foo { ['constructor']() {} }
//	invalid: ({ ['x']: 0 })                        ->  ({ 'x': 0 })
//	invalid: ({ [0]: 0 })                          ->  ({ 0: 0 })
//	invalid: class Foo { ['x']() {} }              ->  class Foo { 'x'() {} }
//
// # Only a string or a number literal can lose its brackets
//
// A non-computed key may be an identifier, a number literal or a string literal, so those are the
// only values whose brackets are removable. Upstream reads `key.type !== "Literal"` and then
// `typeof value !== "number" && typeof value !== "string"`, which declines a template literal, a
// regular expression, `null`, a boolean, and -- deliberately -- a bigint. The bigint decline is
// upstream telling us what it knowingly misses, with a comment saying well-known browsers throw a
// syntax error on a bigint property name, so the rule leaves those alone "for now". Reproduced
// rather than improved on: `({ [99999999999999999n]: 0 })` is one of upstream's clean cases.
//
// A no-substitution template is a second decline worth naming, because our parser gives it its own
// kind and the shelf's `property.Name` accepts it. Upstream's ESTree calls it a `TemplateLiteral`
// rather than a `Literal`, so `({ [`+"`"+`x`+"`"+`]: 0 })` is clean. This port declines it by listing the two
// accepted kinds rather than by excluding kinds, so a kind nobody thought about is declined by
// default rather than accepted by default.
//
// # The four reserved names, and why they differ by position
//
// Dropping brackets is only safe when the resulting non-computed key means the same thing, and four
// names change meaning when they lose their brackets:
//
//	{ ['__proto__']: foo }             defines a property; { '__proto__': foo } sets the prototype
//	class C { ['constructor'] }        an instance field; class C { 'constructor' } fails to parse
//	class C { static ['constructor'] } a static field; the plain form fails to parse
//	class C { static ['prototype'] }   a runtime error; the plain form fails to PARSE
//	class C { ['constructor']() {} }   a prototype method; the plain form IS the constructor
//	class C { static ['prototype']() {} } a runtime error; the plain form fails to parse
//
// So the exempt name depends on both the member kind and whether it is static, and the `__proto__`
// exemption applies only inside an object EXPRESSION. In a destructuring pattern there is no
// prototype to set, which is why `var { ['__proto__']: a } = obj` reports while
// `({ ['__proto__']: [] })` does not. That pair is the whole reason upstream checks the parent kind
// rather than just the name.
//
// # What TypeScript adds, measured against the installed rule rather than reasoned about
//
// Three TypeScript member shapes carry a computed key and are SILENT upstream, and in every case the
// mechanism is that typescript-eslint gives them their own ESTree node type which upstream's
// `create` never listens on:
//
//	interface I { ['x']: number }      TSPropertySignature      silent
//	type T = { ['x']: number }         TSPropertySignature      silent
//	abstract class C { abstract ['x'](): void }   TSAbstractMethodDefinition   silent
//	class C { accessor ['x'] = 1 }     AccessorProperty         silent
//
// Each was measured by driving the installed rule with the typescript-eslint parser, alongside a
// control in the same file that does report, so the zero is an absence of a listener rather than a
// parse failure. Reproduced here by listing the member kinds this rule visits rather than by
// visiting every node with a computed name.
//
// An enum member cannot carry a computed key at all -- TypeScript rejects it as a parse error --
// so `KindEnumMember` is unreachable rather than declined, and it is absent from the visited set
// for the same reason.
//
// Every other TypeScript decoration around the key is reported and fixed, because none of it lives
// inside the brackets: `?`, `!`, `declare`, `readonly`, `static`, `override`, an access modifier, a
// decorator, a type annotation, type parameters, and a `satisfies` on the enclosing object all
// survive the repair untouched. That is the load-bearing fact about this fixer and it is the reason
// the repair is shipped rather than declined; see the fixer's own comment.
var NoUselessComputedKey = rule.Rule{
	Name: "no-useless-computed-key",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoUselessComputedKeySettings](options)
		if !ok {
			// A rule configured as a bare severity is handed nil options, and the zero value of
			// this struct switches off class members rather than disabling the rule.
			settings = DefaultNoUselessComputedKeySettings()
		}

		check := func(node *ast.Node) {
			checkUselessComputedKey(ctx, node, settings)
		}

		return rule.Listeners{
			// Upstream's `Property`, in both its object-literal and its destructuring-pattern
			// spellings. Our parser gives the pattern form its own kind.
			ast.KindPropertyAssignment: check,
			ast.KindBindingElement:     check,
			// Upstream's `PropertyDefinition` and `MethodDefinition`. A getter and a setter are a
			// MethodDefinition upstream with `kind` set, so both are here.
			ast.KindPropertyDeclaration: check,
			ast.KindMethodDeclaration:   check,
			ast.KindGetAccessor:         check,
			ast.KindSetAccessor:         check,
		}
	},
}

// checkUselessComputedKey judges one member and proposes the repair.
func checkUselessComputedKey(ctx rule.Context, node *ast.Node, settings NoUselessComputedKeySettings) {
	if ctx.SourceFile == nil {
		return
	}

	name := uselessComputedKeyName(node)
	if name == nil || name.Kind != ast.KindComputedPropertyName {
		return
	}

	// `abstract` and `accessor` members are silent upstream, and the mechanism is a parser
	// difference rather than a judgment. typescript-eslint promotes them to their own ESTree node
	// types -- `TSAbstractMethodDefinition`, `TSAbstractPropertyDefinition`, `AccessorProperty` --
	// which upstream's `create` never listens on, so no finding is ever produced for one. Our
	// parser keeps the ordinary kind and records the word as a modifier flag, so without this
	// decline they arrive at a listener that upstream does not have.
	//
	// Measured against the installed rule across five spellings, each with a reporting control in
	// the same file so a zero could be told from a parse failure: `abstract ['x']: number`,
	// `abstract ['x'](): void`, `abstract accessor ['x']`, `static accessor ['x']`, and
	// `abstract get ['x']()`. Only the control fired in every one.
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsAbstract) ||
		ast.HasSyntacticModifier(node, ast.ModifierFlagsAccessor) {
		return
	}

	// `enforceForClassMembers: false` narrows the rule to object literals. Upstream implements this
	// by swapping the class listeners for a no-op, which is the same decision made here at the
	// member rather than at registration, because our listener table is built once per file.
	if !settings.EnforceForClassMembers && isUselessComputedKeyClassMember(node) {
		return
	}

	key := name.AsComputedPropertyName().Expression
	// Our parser keeps parentheses and upstream's folds them away, so `({ [('x')]: 0 })` arrives
	// here wrapped in a node upstream never sees. Unwrapped in a loop rather than in one step,
	// because `((\'x\'))` nests, and written out rather than through `ast.SkipParentheses`, which
	// dereferences its argument -- a computed key's expression is optional in a recovered parse.
	//
	// The repair stays correct through the unwrap, and that is worth stating because the opposite
	// was assumed first. Upstream reports `sourceCode.getText(key)` where `key` is the FOLDED
	// literal, so its message says `'x'` and not `('x')`, and its fixer replaces the whole
	// bracket-to-bracket span, so the parentheses fall inside what is being replaced and vanish.
	// All three of upstream's parenthesized cases assert exactly that output. Declining them, which
	// this port did until the corpus said otherwise, cost three real findings.
	for key != nil && key.Kind == ast.KindParenthesizedExpression {
		key = key.AsParenthesizedExpression().Expression
	}
	if key == nil {
		return
	}
	if key.Kind != ast.KindStringLiteral && key.Kind != ast.KindNumericLiteral {
		return
	}

	if !uselessComputedKeyIsRemovable(node, key) {
		return
	}

	// The key's SOURCE text, not its cooked value. Upstream reports `sourceCode.getText(key)` and
	// its fixer writes `key.raw`, so `({ ['A']: 0 })` reports the escape as written and
	// repairs to `'A'` rather than to `'A'`. Reading `key.Text()` here would cook the escape
	// and both the message and the repair would silently change what the source says.
	keyRange := rule.TokenRange(ctx.SourceFile, key)
	keyText := ctx.SourceFile.Text()[keyRange.Pos():keyRange.End()]

	finding := rule.Message{
		Id: messageUnnecessarilyComputedProperty.Id,
		Description: fmt.Sprintf("Unnecessarily computed property [%s] found. %s",
			keyText, messageUnnecessarilyComputedProperty.Description),
	}

	fix, fixable := uselessComputedKeyRepair(ctx, name, key, keyRange, keyText)
	if !fixable {
		ctx.ReportNode(node, finding)
		return
	}
	ctx.ReportNodeWithFixes(node, finding, fix)
}

// uselessComputedKeyRepair builds the replacement for the bracketed span, or declines.
//
// # Why constructing a span is safe here, when it was not for `no-undef-init`
//
// This fixer replaces `[` through `]` and writes back the key's own source text. Everything a
// TypeScript declaration can carry sits OUTSIDE those brackets -- the `?`, the `!`, the type
// annotation, the type parameters, the modifiers, the decorators -- so none of it is inside the
// span being rewritten and none of it can be dropped. Measured against the installed rule on eleven
// TypeScript shapes: `['x']?: number`, `['x']!: number`, `declare ['x']: number`,
// `readonly ['x']: number = 1`, `protected static ['x']()`, `override ['x']()`, `@dec ['x']()`,
// `['x']<T>(a: T): T`, `['x'](a: number, b?: string): void`, `['x']: 0 as const`, and an object
// literal under `satisfies` all repair with every annotation intact.
//
// The one thing that CAN live inside the brackets besides the key is a comment, and that is the
// case upstream declines. Four of its invalid cases carry `output: null` for exactly this.
//
// So the enumeration this fixer's span requires is short and complete: the brackets, the key, and
// any trivia between them. The first two are rewritten deliberately and the third is the decline.
func uselessComputedKeyRepair(ctx rule.Context, name *ast.Node, key *ast.Node,
	keyRange core.TextRange, keyText string) (rule.Fix, bool) {

	source := ctx.SourceFile.Text()

	// The bracket span. `name.Pos()` includes leading trivia, so the opening bracket is found by
	// trimming to the token rather than by taking the node's start -- for `({ [/* c */ 'x']: 0 })`
	// the untrimmed position sits before the space, and for a key written after a line comment it
	// would sit before the comment.
	bracketRange := rule.TokenRange(ctx.SourceFile, name)
	openBracket := bracketRange.Pos()
	closeBracket := bracketRange.End()
	// Bounds before indexing, because the two lines below read `source[openBracket]` and
	// `source[closeBracket-1]` directly and a bad span would panic rather than misbehave. A panic
	// costs every rule its verdict on the whole file, not just this one.
	//
	// A mutation neutralising this survived the whole fixture set, and the reason is that no
	// `ExpectFindings` fixture can observe a panic that does not happen: the guard prevents a
	// crash rather than deciding a verdict. Its inverse was scored to prove the line is reached at
	// all -- forcing it always-true fails 79 lines -- so this is crash protection on a live path
	// rather than dead code. Kept, with the reasoning here so it is not deleted as unreachable.
	if openBracket < 0 || closeBracket > len(source) || openBracket >= closeBracket {
		return rule.Fix{}, false
	}
	if source[openBracket] != '[' || source[closeBracket-1] != ']' {
		// Not the shape this repair understands. Report without a fix rather than rewrite a span
		// whose ends are not what they are believed to be.
		return rule.Fix{}, false
	}

	// A comment anywhere between the brackets declines the repair, which is upstream's
	// `sourceCode.commentsExistBetween(leftSquareBracket, rightSquareBracket)`.
	//
	// Asked of the whole bracket interior rather than only of the two gaps flanking the key,
	// because a parenthesized key puts a third gap INSIDE the parens that neither flank covers:
	// `({ [(/* c */ 'x')]: 0 })` has `(` on the left, `)` on the right, and the comment between
	// them. Measured against the installed rule, that case reports and is left unrepaired, so
	// scanning the flanks alone would have shipped a repair upstream declines.
	//
	// The interior is either the key's own text or trivia, so a comment is present exactly when
	// removing the key's span leaves something other than whitespace and parentheses behind. That
	// is the same question upstream asks and it needs no second pass over the file.
	if uselessComputedKeyHasTrivia(source[openBracket+1:keyRange.Pos()]) ||
		uselessComputedKeyHasTrivia(source[keyRange.End():closeBracket-1]) {
		return rule.Fix{}, false
	}

	// Upstream inserts a space when the token before the bracket would otherwise fuse with the key:
	// `({ get[2]() {} })` must become `({ get 2() {} })` and not `({ get2() {} })`. It asks
	// `canTokensBeAdjacent`, and only when the bracket sits flush against the previous token.
	replacement := keyText
	if openBracket > 0 && !uselessComputedKeyTokensCanBeAdjacent(source, openBracket, key) {
		replacement = " " + replacement
	}

	return rule.ReplaceRange(core.NewTextRange(openBracket, closeBracket), replacement), true
}

// uselessComputedKeyHasTrivia says whether a stretch of source inside the brackets holds anything
// other than whitespace and parentheses, which for this span means a comment.
//
// Parentheses are accepted rather than treated as trivia because the unwrap above deliberately
// removes them: they sit inside the replaced span and disappear with it, which is what makes
// `({ [('x')]: 0 })` repair to `({ 'x': 0 })` exactly as upstream's own case asserts. Anything else
// non-whitespace here is a comment, and a comment is upstream's decline.
func uselessComputedKeyHasTrivia(between string) bool {
	for i := 0; i < len(between); i++ {
		switch between[i] {
		case ' ', '\t', '\n', '\r', '\v', '\f', '(', ')':
			continue
		}
		return true
	}
	return false
}

// uselessComputedKeyTokensCanBeAdjacent reproduces upstream's `needsSpaceBeforeKey`.
//
// Upstream's condition has two halves and both matter. The tokens must be FLUSH -- if there is
// already a space, as in `({ get [2]() {} })`, nothing is inserted and the output keeps the one
// space that was written. And they must be unable to sit next to each other, which
// `astUtils.canTokensBeAdjacent` decides by re-tokenizing the pair.
//
// Reproduced narrowly rather than by re-tokenizing, because the only token that can precede the
// opening bracket of a computed key is one of a small set: an identifier-like word (`get`, `set`,
// `async`, `static`, a modifier), `*`, `,`, `{`, `;`, `}`, `)`, or a decorator's end. Only the
// identifier-like case can fuse, and only with a key whose first character can continue an
// identifier. The corpus states all three outcomes:
//
//	({ get[2]() {} })      -> ({ get 2() {} })    a digit continues `get`, so a space is needed
//	({ get[.2]() {} })     -> ({ get.2() {} })    `.` cannot, so none is
//	({ get['foo']() {} })  -> ({ get'foo'() {} }) a quote cannot, so none is
func uselessComputedKeyTokensCanBeAdjacent(source string, openBracket int, key *ast.Node) bool {
	previous := source[openBracket-1]
	if !uselessComputedKeyIsWordCharacter(previous) {
		// Whitespace, `*`, `,`, `{`, `;`, `}` and `)` all sit against a key without fusing. This
		// arm also covers the already-spaced case, since a space is not a word character.
		return true
	}
	if key.Kind != ast.KindNumericLiteral {
		// A string literal starts with a quote, which cannot continue an identifier.
		//
		// This arm is EQUIVALENT to falling through to the numeric path below, and the mutation
		// removing it survives for that reason rather than because a fixture is missing. The
		// distinguishing input would have to be a string literal whose first source byte is a word
		// character, and no such token exists: a `KindStringLiteral` always opens with `'` or `"`,
		// so the numeric path would compute `!isWord(quote)`, which is the same `true` this
		// returns. Kept because it states the reason rather than leaving it to be re-derived, and
		// because a future accepted kind might not open with a quote.
		return true
	}
	// A numeric literal fuses only when its first character can continue an identifier. `.2` starts
	// with a dot and cannot; `2` can.
	keyStart := source[key.Pos():key.End()]
	for i := 0; i < len(keyStart); i++ {
		if keyStart[i] == ' ' || keyStart[i] == '\t' || keyStart[i] == '\n' || keyStart[i] == '\r' {
			continue
		}
		return !uselessComputedKeyIsWordCharacter(keyStart[i])
	}
	return true
}

// uselessComputedKeyIsWordCharacter says whether a byte can appear inside an identifier or a number,
// which is what decides whether two tokens fuse when written flush.
func uselessComputedKeyIsWordCharacter(b byte) bool {
	return b == '_' || b == '$' ||
		(b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

// uselessComputedKeyName returns the member's name node, for the member kinds this rule visits.
func uselessComputedKeyName(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindBindingElement:
		// A destructuring pattern's key is its PropertyName, and `Name()` is the binding it
		// introduces. Reading `Name()` here would ask about `a` in `var { ['x']: a } = obj`, which
		// is never computed, and every one of upstream's six pattern cases would go silent.
		return node.AsBindingElement().PropertyName
	}
	return node.Name()
}

// isUselessComputedKeyClassMember says whether a member belongs to a class body rather than to an
// object literal, which is what `enforceForClassMembers` switches off.
//
// Read off the member's own kind rather than off its parent, because the two families do not
// overlap: a `KindPropertyAssignment` and a `KindBindingElement` only ever appear in an object
// literal or a destructuring pattern, and a `KindPropertyDeclaration` only ever appears in a class.
// The accessor and method kinds DO appear in both, so those ask the parent.
func isUselessComputedKeyClassMember(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindPropertyAssignment, ast.KindBindingElement:
		return false
	case ast.KindPropertyDeclaration:
		return true
	}
	return node.Parent != nil && ast.IsClassLike(node.Parent)
}

// uselessComputedKeyIsRemovable reproduces the tail of upstream's `hasUselessComputedKey`: whether
// dropping the brackets around this literal would preserve what the code means.
func uselessComputedKeyIsRemovable(node *ast.Node, key *ast.Node) bool {
	value := key.Text()

	switch node.Kind {
	case ast.KindPropertyAssignment:
		// Upstream's `Property` inside an `ObjectExpression`. `{ '__proto__': foo }` sets the
		// prototype where `{ ['__proto__']: foo }` defines a property, so the brackets carry
		// meaning and stay.
		return value != "__proto__"

	case ast.KindBindingElement:
		// Upstream's `Property` inside an `ObjectPattern`, which is its `return true` arm. There is
		// no prototype to set when destructuring, so `__proto__` is removable here and this is the
		// half of the pair that makes the parent test load-bearing rather than decorative.
		return true

	case ast.KindPropertyDeclaration:
		// Upstream's `PropertyDefinition`.
		if isUselessComputedKeyStatic(node) {
			return value != "constructor" && value != "prototype"
		}
		return value != "constructor"

	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		// Upstream's `MethodDefinition`. A getter and a setter carry `kind` there rather than a
		// node type of their own, and the judgment does not consult it.
		//
		// An object literal's method reaches here too, and it takes the same non-static arm as a
		// class method: upstream's switch reads the node type, and an object method IS a
		// `Property` there rather than a `MethodDefinition`... which means it takes the Property
		// arm and is judged against `__proto__` instead. Measured against the installed rule,
		// `({ ['constructor']() {} })` REPORTS and `({ ['__proto__']() {} })` does not, so an
		// object method is a Property upstream and is routed as one below.
		if !isUselessComputedKeyClassMember(node) {
			return value != "__proto__"
		}
		if isUselessComputedKeyStatic(node) {
			return value != "prototype"
		}
		return value != "constructor"
	}
	return false
}

// isUselessComputedKeyStatic says whether a class member carries the `static` modifier.
func isUselessComputedKeyStatic(node *ast.Node) bool {
	return ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic)
}
