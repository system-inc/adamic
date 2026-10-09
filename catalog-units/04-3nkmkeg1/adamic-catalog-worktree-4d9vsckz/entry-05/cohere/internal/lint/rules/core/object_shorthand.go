package core

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// ObjectShorthandMode selects which shorthand the rule requires.
type ObjectShorthandMode string

const (
	// ObjectShorthandAlways requires shorthand for both methods and properties. Upstream's default.
	ObjectShorthandAlways ObjectShorthandMode = "Always"
	// ObjectShorthandMethods requires shorthand for methods only.
	ObjectShorthandMethods ObjectShorthandMode = "Methods"
	// ObjectShorthandProperties requires shorthand for properties only.
	ObjectShorthandProperties ObjectShorthandMode = "Properties"
	// ObjectShorthandNever requires longform for both.
	ObjectShorthandNever ObjectShorthandMode = "Never"
	// ObjectShorthandConsistent requires each object literal to pick one style.
	ObjectShorthandConsistent ObjectShorthandMode = "Consistent"
	// ObjectShorthandConsistentAsNeeded is Consistent plus a requirement that the chosen style be
	// shorthand whenever every property could be.
	ObjectShorthandConsistentAsNeeded ObjectShorthandMode = "ConsistentAsNeeded"
)

// ObjectShorthandSettings is the decoded option surface.
type ObjectShorthandSettings struct {
	// Mode is the first tuple element. Upstream's default is "always", so the zero value of this
	// type is deliberately not a valid mode.
	Mode ObjectShorthandMode
	// AvoidQuotes keeps a quoted key in longform rather than converting it to a method.
	AvoidQuotes bool
	// IgnoreConstructors exempts a method whose name looks like a constructor.
	IgnoreConstructors bool
	// MethodsIgnorePattern exempts a method whose name matches, or nil for none.
	MethodsIgnorePattern *regexp.Regexp
	// MethodsIgnorePatternText is the pattern as written, kept only for the decoder's own test.
	MethodsIgnorePatternText string
	// AvoidExplicitReturnArrows converts an arrow with a block body into a method.
	AvoidExplicitReturnArrows bool
}

// DefaultObjectShorthandSettings is upstream's `defaultOptions: ["always"]`.
//
// Spelled out because the zero value of the struct has an EMPTY mode, which matches none of the six
// and would make the rule silently report nothing.
func DefaultObjectShorthandSettings() ObjectShorthandSettings {
	return ObjectShorthandSettings{Mode: ObjectShorthandAlways}
}

// DecodeObjectShorthandOptions reads the mode and flags off the config.
//
// # Three list shapes, and which flags are legal depends on the mode
//
// Upstream's schema is an `anyOf` over three lists, and the rule registers with `DecodeOptionList`
// so it is handed upstream's list as written:
//
//	["always" | "methods" | "properties" | "never" | "consistent" | "consistent-as-needed"]
//	["always" | "methods" | "properties", {"avoidQuotes"}]
//	["always" | "methods", {"ignoreConstructors", "methodsIgnorePattern", "avoidQuotes",
//	                        "avoidExplicitReturnArrows"}]
//
// That pairing is enforced here. The rule body reads `context.options[1] || {}` whatever the mode,
// but upstream's schema refuses a flag beside a mode that cannot use it before the rule runs, so a
// config doing it is a config error there, and accepting it here would be a flag read and never
// honoured. A key outside the four is refused for the same reason.
//
// The config layer used to keep only the first element after the severity, so this decoder accepted
// a nested `["error", ["always", {...}]]` to reach the flags. That spelling is refused now: its first
// element is not a string.
//
// A mode that is not one of the six is an error rather than a silent fallback to "always". A typo
// in a mode name would otherwise produce a rule that looks configured and enforces the default.
func DecodeObjectShorthandOptions(list []byte) (any, error) {
	settings := DefaultObjectShorthandSettings()
	elements, err := rule.OptionElements(list, 2)
	if err != nil || len(elements) == 0 {
		return settings, err
	}

	var modeText string
	if err := json.Unmarshal(elements[0], &modeText); err != nil {
		return settings, fmt.Errorf("object-shorthand element 1 takes a mode string, got %s", elements[0])
	}
	mode, ok := objectShorthandModeFor(modeText)
	if !ok {
		return settings, fmt.Errorf("object-shorthand mode %q is not one of "+
			"always, methods, properties, never, consistent, consistent-as-needed", modeText)
	}
	settings.Mode = mode
	if len(elements) < 2 {
		return settings, nil
	}

	var wire struct {
		AvoidQuotes               *bool   `json:"avoidQuotes"`
		IgnoreConstructors        *bool   `json:"ignoreConstructors"`
		MethodsIgnorePattern      *string `json:"methodsIgnorePattern"`
		AvoidExplicitReturnArrows *bool   `json:"avoidExplicitReturnArrows"`
	}
	if err := rule.UnmarshalOptions(elements[1], &wire); err != nil {
		return settings, fmt.Errorf("object-shorthand element 2: %w", err)
	}

	// Upstream's schema: no flags beside never or the consistent modes, only avoidQuotes beside
	// properties, all four beside always and methods.
	switch mode {
	case ObjectShorthandAlways, ObjectShorthandMethods:
	case ObjectShorthandProperties:
		if wire.IgnoreConstructors != nil || wire.MethodsIgnorePattern != nil ||
			wire.AvoidExplicitReturnArrows != nil {
			return settings, fmt.Errorf("object-shorthand %q reads only avoidQuotes, so %s would "+
				"never be read in full", modeText, elements[1])
		}
	default:
		return settings, fmt.Errorf("object-shorthand %q takes no second element, so %s would "+
			"never be read", modeText, elements[1])
	}

	if wire.AvoidQuotes != nil {
		settings.AvoidQuotes = *wire.AvoidQuotes
	}
	if wire.IgnoreConstructors != nil {
		settings.IgnoreConstructors = *wire.IgnoreConstructors
	}
	if wire.AvoidExplicitReturnArrows != nil {
		settings.AvoidExplicitReturnArrows = *wire.AvoidExplicitReturnArrows
	}
	if wire.MethodsIgnorePattern != nil && *wire.MethodsIgnorePattern != "" {
		compiled, err := regexp.Compile(*wire.MethodsIgnorePattern)
		if err != nil {
			return settings, fmt.Errorf(
				"object-shorthand methodsIgnorePattern %q does not compile: %w",
				*wire.MethodsIgnorePattern, err)
		}
		settings.MethodsIgnorePattern = compiled
		settings.MethodsIgnorePatternText = *wire.MethodsIgnorePattern
	}
	return settings, nil
}

// objectShorthandModeFor maps upstream's spelling onto the mode.
func objectShorthandModeFor(text string) (ObjectShorthandMode, bool) {
	switch text {
	case "always":
		return ObjectShorthandAlways, true
	case "methods":
		return ObjectShorthandMethods, true
	case "properties":
		return ObjectShorthandProperties, true
	case "never":
		return ObjectShorthandNever, true
	case "consistent":
		return ObjectShorthandConsistent, true
	case "consistent-as-needed":
		return ObjectShorthandConsistentAsNeeded, true
	}
	return "", false
}

// objectShorthandSettingsFrom recovers the settings from whatever the config layer handed over.
func objectShorthandSettingsFrom(options any) ObjectShorthandSettings {
	if settings, ok := rule.OptionsAs[ObjectShorthandSettings](options); ok && settings.Mode != "" {
		return settings
	}
	return DefaultObjectShorthandSettings()
}

var messageObjectShorthandExpectedPropertyShorthand = rule.Message{
	Id:          "expectedPropertyShorthand",
	Description: "A property whose key and value are the same name reads as a duplication.",
}

var messageObjectShorthandExpectedPropertyLongform = rule.Message{
	Id:          "expectedPropertyLongform",
	Description: "This project writes object properties in longform, so the key and value are both explicit.",
}

var messageObjectShorthandExpectedMethodShorthand = rule.Message{
	Id: "expectedMethodShorthand",
	Description: "Method shorthand says the value is a method rather than a function that happens " +
		"to be stored on the object.",
}

var messageObjectShorthandExpectedMethodLongform = rule.Message{
	Id:          "expectedMethodLongform",
	Description: "This project writes object methods in longform, as an explicit function expression.",
}

var messageObjectShorthandExpectedLiteralMethodLongform = rule.Message{
	Id: "expectedLiteralMethodLongform",
	Description: "A quoted key is kept in longform here, so the quoting is not doing two jobs at " +
		"once.",
}

var messageObjectShorthandExpectedAllPropertiesShorthanded = rule.Message{
	Id: "expectedAllPropertiesShorthanded",
	Description: "This object mixes shorthand and longform. Every property here can be shorthand, " +
		"so writing some of them longform makes the difference look meaningful when it is not.",
}

var messageObjectShorthandUnexpectedMix = rule.Message{
	Id: "unexpectedMix",
	Description: "This object mixes shorthand and longform properties, which makes the difference " +
		"look meaningful when it is not. Pick one and use it throughout.",
}

// ObjectShorthand reports an object property written in the style the configuration does not want.
//
//	valid:   var o = { a() {} };            (always, the default)
//	valid:   var o = { a: a };              (never)
//	valid:   var o = { a: a, b: b };        (consistent: all longform)
//	invalid: var o = { a: function() {} };  (always, fixed to `{ a() {} }`)
//	invalid: var o = { a };                 (never, fixed to `{ a: a }`)
//	invalid: var o = { a, b: b };           (consistent, mixed)
//
// # ESTree's two boolean flags are three node kinds here, which makes the dispatch simpler
//
// Upstream reads `node.method` and `node.shorthand` off one `Property` type, plus `node.kind` for
// accessors. Our parser gives each shape its own kind, probed:
//
//	{ a: function(){} }   KindPropertyAssignment with a FunctionExpression value
//	{ a() {} }            KindMethodDeclaration          -- upstream's `method: true`
//	{ a }                 KindShorthandPropertyAssignment -- upstream's `shorthand: true`
//	{ get a() {} }        KindGetAccessor                 -- upstream's `kind: "get"`
//
// Upstream's `node.parent.type === "ObjectPattern"` guard needs no equivalent: a destructuring
// pattern produces binding nodes rather than object-literal members, so the walk never reaches one.
//
// # Six modes, and `consistent` is a different question from the rest
//
// The four simple modes judge each property alone. `consistent` and `consistent-as-needed` judge
// the OBJECT: every property must use the same style, and under `consistent-as-needed` that style
// must be shorthand when every property could be. Those two report on the object literal rather
// than on a property, which is why they are a separate arm and a separate message id.
//
// # No fixer for the consistency modes
//
// Upstream attaches a fix only to the per-property modes. Making an object consistent means
// choosing which style to convert everything to, and only the author knows which. Reported without
// a repair, which is the subset that can be shown correct.
var ObjectShorthand = rule.Rule{
	Name: "object-shorthand",

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := objectShorthandSettingsFrom(options)

		return rule.Listeners{
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				switch settings.Mode {
				case ObjectShorthandConsistent:
					checkObjectShorthandConsistency(ctx, node, false)
				case ObjectShorthandConsistentAsNeeded:
					checkObjectShorthandConsistency(ctx, node, true)
				default:
					literal := node.AsObjectLiteralExpression()
					if literal == nil || literal.Properties == nil {
						return
					}
					for _, member := range literal.Properties.Nodes {
						checkObjectShorthandMember(ctx, member, settings)
					}
				}
			},
		}
	},
}

// checkObjectShorthandMember is upstream's `Property:exit` for the four per-property modes.
func checkObjectShorthandMember(ctx rule.Context, member *ast.Node, settings ObjectShorthandSettings) {
	// Upstream needs two guards here that this port does not, and they are recorded rather than
	// carried as dead code.
	//
	// It returns early on `node.kind === "get" || "set"` and on a spread, because in ESTree all
	// four shapes are one `Property` type and would otherwise reach the shorthand tests. Our parser
	// gives each its own kind, so an accessor or a spread simply matches none of the switch arms
	// below and falls through. Verified rather than assumed: deleting the equivalent guards changes
	// no verdict on `{ get a() {} }` under any mode, so a mutation removing them survives -- as
	// genuine equivalence, not as a fixture blind spot. `TestObjectShorthandIgnoresAccessors` pins
	// the behaviour either way.

	applyToMethods := settings.Mode == ObjectShorthandMethods || settings.Mode == ObjectShorthandAlways
	applyToProperties := settings.Mode == ObjectShorthandProperties || settings.Mode == ObjectShorthandAlways
	applyNever := settings.Mode == ObjectShorthandNever

	switch member.Kind {
	// Already concise: upstream's `isConciseProperty` branch.
	case ast.KindMethodDeclaration:
		if applyNever || (settings.AvoidQuotes && objectShorthandKeyIsStringLiteral(member)) {
			message := messageObjectShorthandExpectedLiteralMethodLongform
			if applyNever {
				message = messageObjectShorthandExpectedMethodLongform
			}
			if fix, ok := objectShorthandMethodToLongform(ctx, member); ok {
				ctx.ReportNodeWithFixes(member, message, fix...)
				return
			}
			ctx.ReportNode(member, message)
		}

	case ast.KindShorthandPropertyAssignment:
		if applyNever {
			assignment := member.AsShorthandPropertyAssignment()
			if assignment == nil || assignment.Name() == nil {
				return
			}
			name := assignment.Name()
			ctx.ReportNodeWithFixes(member, messageObjectShorthandExpectedPropertyLongform,
				rule.ReplaceRange(core.NewTextRange(name.End(), name.End()), ": "+name.Text()))
		}

	// Longform: upstream's else branch.
	case ast.KindPropertyAssignment:
		assignment := member.AsPropertyAssignment()
		if assignment == nil || assignment.Initializer == nil || assignment.Name() == nil {
			return
		}
		value := assignment.Initializer
		// A parenthesized value is still the value. Upstream's parser folds the parenthesis away
		// and sees the function directly, so `({ a: (function(){}) })` reports; ours keeps the node
		// and has to unwrap it, in a loop since `((f))` nests.
		for value.Kind == ast.KindParenthesizedExpression {
			inner := value.AsParenthesizedExpression()
			if inner == nil || inner.Expression == nil {
				break
			}
			value = inner.Expression
		}

		isFunctionValue := value.Kind == ast.KindFunctionExpression || value.Kind == ast.KindArrowFunction
		if applyToMethods && isFunctionValue && objectShorthandFunctionIsAnonymous(value) {
			if settings.IgnoreConstructors && assignment.Name().Kind == ast.KindIdentifier &&
				objectShorthandIsConstructorName(assignment.Name().Text()) {
				return
			}
			if settings.MethodsIgnorePattern != nil {
				if name, ok := property.Name(assignment.Name(), property.Static); ok &&
					settings.MethodsIgnorePattern.MatchString(name) {
					return
				}
			}
			if settings.AvoidQuotes && objectShorthandKeyIsStringLiteral(member) {
				return
			}
			// An arrow converts only when it has a BLOCK body and `avoidExplicitReturnArrows` is
			// on. Upstream's condition is
			// `ArrowFunctionExpression && body.type === "BlockStatement" && AVOID_EXPLICIT_RETURN_ARROWS`,
			// so an expression-bodied arrow such as `{ x: () => foo }` is NEVER reported: turning
			// it into a method would have to invent a `return`, which changes what the source says
			// rather than how it says it. A first draft had this backwards and reported three
			// corpus cases upstream leaves clean.
			if value.Kind == ast.KindArrowFunction {
				body := value.Body()
				if body == nil || body.Kind != ast.KindBlock || !settings.AvoidExplicitReturnArrows {
					return
				}
				// An arrow that uses `this`, `super`, `new.target` or `arguments` cannot become a
				// method, because a method binds its own.
				if objectShorthandUsesLexicalIdentifier(value) {
					return
				}
			}
			if fix, ok := objectShorthandFunctionToShorthand(ctx, member, assignment, value); ok {
				ctx.ReportNodeWithFixes(member, messageObjectShorthandExpectedMethodShorthand, fix)
				return
			}
			ctx.ReportNode(member, messageObjectShorthandExpectedMethodShorthand)
			return
		}

		// A parenthesized value is still the value: `{a: (a /* c */)}` reports, because upstream's
		// parser folds the parenthesis away and sees the identifier directly. Ours keeps it, so it
		// has to be unwrapped here -- in a loop, since `((a))` nests.
		unwrapped := value
		for unwrapped.Kind == ast.KindParenthesizedExpression {
			inner := unwrapped.AsParenthesizedExpression()
			if inner == nil || inner.Expression == nil {
				break
			}
			unwrapped = inner.Expression
		}
		if applyToProperties && unwrapped.Kind == ast.KindIdentifier {
			// Upstream has TWO property-shorthand branches, and the second is easy to miss: a
			// QUOTED key counts too, when its string value equals the identifier. `{'x': x}` is
			// reportable and `{'x': y}` is not.
			key := assignment.Name()
			quoted := key.Kind == ast.KindStringLiteral
			matches := (key.Kind == ast.KindIdentifier || quoted) && key.Text() == unwrapped.Text()
			if !matches {
				return
			}
			// The quoted branch alone honours `avoidQuotes`, since removing the quotes is part of
			// the conversion there.
			if quoted && settings.AvoidQuotes {
				return
			}
			// A JSDoc comment carrying `@type` is documentation attached to this property, and the
			// shorthand form has nowhere to put it. Both of upstream's branches skip it.
			if objectShorthandHasTypeJsDoc(ctx, member) {
				return
			}
			// Upstream's fixer declines whenever ANY comment sits inside the property, because the
			// replacement writes just the name and would delete it.
			if objectShorthandHasCommentInside(ctx, member) {
				ctx.ReportNode(member, messageObjectShorthandExpectedPropertyShorthand)
				return
			}
			ctx.ReportNodeWithFixes(member, messageObjectShorthandExpectedPropertyShorthand,
				ctx.ReplaceNode(member, unwrapped.Text()))
		}
	}
}

// checkObjectShorthandConsistency is upstream's `checkConsistency`.
//
// Every property in one object literal must use the same style. Under `checkRedundancy` the chosen
// style must additionally be shorthand whenever every property COULD be shorthand, which is what
// separates `consistent-as-needed` from `consistent`.
func checkObjectShorthandConsistency(ctx rule.Context, node *ast.Node, checkRedundancy bool) {
	literal := node.AsObjectLiteralExpression()
	if literal == nil || literal.Properties == nil {
		return
	}

	shorthand, longform := 0, 0
	allCouldBeShorthand := true
	for _, member := range literal.Properties.Nodes {
		switch member.Kind {
		case ast.KindGetAccessor, ast.KindSetAccessor, ast.KindSpreadAssignment:
			// Upstream's `canHaveShorthand` filter drops accessors and spreads before counting, and
			// unlike the per-property guards above this one is NOT redundant in spirit -- an
			// accessor reaching the shorthand tally would make `{ get a() {}, b }` look mixed and
			// report `unexpectedMix`, where upstream is clean.
			//
			// It is redundant in FACT here for the same parser reason: neither kind matches the
			// counting arms below either, so a mutation deleting this case also survives. Kept
			// because it states the intent at the place a reader looks for it, and because a future
			// arm added to this switch would need it.
			continue
		case ast.KindMethodDeclaration, ast.KindShorthandPropertyAssignment:
			shorthand++
		case ast.KindPropertyAssignment:
			longform++
			assignment := member.AsPropertyAssignment()
			if assignment == nil || assignment.Initializer == nil || assignment.Name() == nil {
				allCouldBeShorthand = false
				continue
			}
			value := assignment.Initializer
			// A longform property can be shorthand only when it is an anonymous function, or when
			// its key and value are the same name.
			isFunctionValue := value.Kind == ast.KindFunctionExpression
			sameName := assignment.Name().Kind == ast.KindIdentifier &&
				value.Kind == ast.KindIdentifier && assignment.Name().Text() == value.Text()
			if !(isFunctionValue && objectShorthandFunctionIsAnonymous(value)) && !sameName {
				allCouldBeShorthand = false
			}
		}
	}

	if shorthand == 0 && longform == 0 {
		return
	}
	if shorthand > 0 && longform > 0 {
		ctx.ReportNode(node, messageObjectShorthandUnexpectedMix)
		return
	}
	if checkRedundancy && longform > 0 && allCouldBeShorthand {
		ctx.ReportNode(node, messageObjectShorthandExpectedAllPropertiesShorthanded)
	}
}

// objectShorthandHasTypeJsDoc answers upstream's `@type` JSDoc exemption.
//
// A block comment starting with `*` and containing `@type` is a type annotation for this property,
// and the shorthand form has nowhere to carry it.
func objectShorthandHasTypeJsDoc(ctx rule.Context, member *ast.Node) bool {
	text := ctx.SourceFile.Text()
	span := text[objectShorthandMemberStart(ctx, member):member.End()]
	for index := 0; index+3 < len(span); index++ {
		if span[index] == '/' && span[index+1] == '*' && span[index+2] == '*' {
			closing := strings.Index(span[index:], "*/")
			if closing < 0 {
				closing = len(span) - index
			}
			if strings.Contains(span[index:index+closing], "@type") {
				return true
			}
		}
	}
	return false
}

// objectShorthandHasCommentInside answers upstream's `getCommentsInside(node).length > 0`, which is
// what makes its fixer decline rather than delete a comment.
func objectShorthandHasCommentInside(ctx rule.Context, member *ast.Node) bool {
	text := ctx.SourceFile.Text()
	span := text[objectShorthandMemberStart(ctx, member):member.End()]
	for index := 0; index+1 < len(span); index++ {
		if span[index] == '/' && (span[index+1] == '/' || span[index+1] == '*') {
			return true
		}
	}
	return false
}

// objectShorthandKeyIsStringLiteral answers upstream's `isStringLiteral(node.key)`.
func objectShorthandKeyIsStringLiteral(member *ast.Node) bool {
	name := member.Name()
	if name == nil {
		return false
	}
	// A computed key is its own node here, where upstream's `node.key` is the expression inside the
	// brackets directly. So `{['a']: function(){}}` has a KindComputedPropertyName whose expression
	// is the string, and `isStringLiteral(node.key)` upstream sees that string. Reading only the
	// outer node misses it, which cost one corpus case under `avoidQuotes`.
	if name.Kind == ast.KindComputedPropertyName {
		if computed := name.AsComputedPropertyName(); computed != nil && computed.Expression != nil {
			return computed.Expression.Kind == ast.KindStringLiteral
		}
		return false
	}
	return name.Kind == ast.KindStringLiteral
}

// objectShorthandFunctionIsAnonymous answers upstream's `!node.value.id`.
//
// A NAMED function expression cannot become a method, because the method form has nowhere to put
// the name and dropping it would remove a binding the body may use.
func objectShorthandFunctionIsAnonymous(value *ast.Node) bool {
	if value.Kind != ast.KindFunctionExpression {
		return true
	}
	expression := value.AsFunctionExpression()
	return expression == nil || expression.Name() == nil
}

// objectShorthandIsConstructorName answers upstream's `isConstructor`.
//
// NOT simply "starts with a capital". Upstream skips a leading run of `_`, `$` and digits with
// `CTOR_PREFIX_REGEX = /[^_$0-9]/u` and tests the first character AFTER it, so
// `__ConstructorFunction` and `_0ConstructorFunction` are both constructors. A name made entirely
// of those characters -- `_`, `$$`, `_8` -- matches nothing and is not a constructor.
//
// The corpus pins this with eight cases and a first draft testing only the first rune failed all
// of them.
func objectShorthandIsConstructorName(name string) bool {
	for _, character := range name {
		if character == '_' || character == '$' || (character >= '0' && character <= '9') {
			continue
		}
		return unicode.IsUpper(character)
	}
	return false
}

// objectShorthandUsesLexicalIdentifier answers whether an arrow body depends on the enclosing
// lexical bindings a method would shadow.
//
// Upstream tracks this with a scope stack and its `reportLexicalIdentifier`; here it is a walk that
// stops at any nested function, since those establish their own bindings.
func objectShorthandUsesLexicalIdentifier(value *ast.Node) bool {
	found := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		switch node.Kind {
		case ast.KindThisKeyword, ast.KindSuperKeyword:
			found = true
			return
		case ast.KindMetaProperty:
			if meta := node.AsMetaProperty(); meta != nil && meta.KeywordToken == ast.KindNewKeyword {
				found = true
			}
			return
		case ast.KindIdentifier:
			if node.Text() == "arguments" {
				found = true
			}
			return
		// A nested function establishes its own `this` and `arguments`, so what it contains does
		// not constrain the arrow. So does a CLASS: the `super()` inside a nested class's
		// constructor belongs to that class, and upstream's scope stack stops there for the same
		// reason. Missing the class arm made a corpus case go silent, since its arrow body
		// contained `class Foo extends Bar { constructor() { super(); } }`.
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindMethodDeclaration,
			ast.KindClassDeclaration, ast.KindClassExpression:
			return
		}
		node.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	if value.Kind == ast.KindArrowFunction {
		if body := value.Body(); body != nil {
			walk(body)
		}
	}
	return found
}

// objectShorthandMethodToLongform is upstream's `makeFunctionLongform`, which rewrites
// `{ a() {} }` into `{ a: function() {} }`.
//
// The pieces move rather than the text being rebuilt: the modifiers and the key stay where they
// are, and `: function` is inserted after the key with any `*` or `async` carried across.
func objectShorthandMethodToLongform(ctx rule.Context, member *ast.Node) ([]rule.Fix, bool) {
	method := member.AsMethodDeclaration()
	if method == nil || member.Name() == nil {
		return nil, false
	}
	name := member.Name()

	isAsync := false
	if modifiers := member.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier.Kind == ast.KindAsyncKeyword {
				isAsync = true
			}
		}
	}
	isGenerator := method.AsteriskToken != nil

	// Remove the modifiers and the `*`, which move into the inserted function expression.
	fixes := make([]rule.Fix, 0, 3)
	keyStart := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, name.Pos()).Pos()
	memberStart := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, member.Pos()).Pos()
	if keyStart > memberStart {
		fixes = append(fixes, rule.RemoveRange(core.NewTextRange(memberStart, keyStart)))
	}

	inserted := ": function"
	if isAsync {
		inserted = ": async function"
	}
	if isGenerator {
		inserted += "*"
	}
	fixes = append(fixes, rule.ReplaceRange(
		core.NewTextRange(name.End(), name.End()), inserted))
	return fixes, true
}

// objectShorthandFunctionToShorthand is upstream's `makeFunctionShorthand`, which rewrites
// `{ a: function() {} }` into `{ a() {} }`.
//
// The replacement runs from the KEY's first token through the end of the property, and is built as
// the key text plus whatever follows the `function` keyword (or the `*` after it, for a generator),
// with `async ` and `*` carried across as a prefix. Slicing rather than reconstructing is what keeps
// the parameter list, the return type and the body exactly as the author wrote them.
//
// Declines when a comment sits between the key and the value, because the removed span would take
// it: `{ key: /* c */ () => {} }` is reported and not repaired.
func objectShorthandFunctionToShorthand(ctx rule.Context, member *ast.Node,
	assignment *ast.PropertyAssignment, value *ast.Node) (rule.Fix, bool) {

	sourceText := ctx.SourceFile.Text()
	key := assignment.Name()
	if key == nil {
		return rule.Fix{}, false
	}
	keyStart := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, key.Pos()).Pos()
	keyText := sourceText[keyStart:key.End()]

	// A comment between the key and the value would be inside the replaced span, so upstream
	// declines: `{ f: /* c */ function(){} }` is reported and not repaired.
	//
	// The scan runs to the value's first TOKEN rather than to `Pos()`, because `Pos()` begins at
	// the previous token's end and so starts inside the comment rather than after the colon. It
	// also has to use the ORIGINAL initializer rather than the paren-unwrapped value, since
	// unwrapping moves the start past a parenthesis the comment may sit behind.
	valueStart := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, assignment.Initializer.Pos()).Pos()
	if objectShorthandHasCommentBetween(sourceText, key.End(), valueStart) {
		return rule.Fix{}, false
	}

	prefix := ""
	if modifiers := value.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier.Kind == ast.KindAsyncKeyword {
				prefix += "async "
			}
		}
	}
	generator := false
	if value.Kind == ast.KindFunctionExpression && value.AsFunctionExpression().AsteriskToken != nil {
		generator = true
		prefix += "*"
	}

	// An arrow drops its `=>`, so its conversion is assembled from two slices rather than one:
	// the parameter list as written, then the body. Slicing from the arrow's start would keep the
	// arrow token and produce `x() => { ... }`, which is not a method.
	if value.Kind == ast.KindArrowFunction {
		body := value.Body()
		if body == nil || body.Kind != ast.KindBlock {
			return rule.Fix{}, false
		}
		parameterStart := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, value.Pos()).Pos()
		bodyStart := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, body.Pos()).Pos()
		arrow := value.AsArrowFunction()
		if arrow == nil || arrow.EqualsGreaterThanToken == nil {
			return rule.Fix{}, false
		}
		// The slice runs to the arrow token's START and the body slice from the body's start, so
		// the whitespace between them is whatever sat AFTER the arrow rather than before it. That
		// is what upstream produces: `x: () => { ... }` becomes `x() { ... }`, keeping the space
		// the author wrote before the brace.
		// The parameter list as written, taken from the arrow's own parameter nodes rather than
		// from a text scan: the slice starts at the arrow's first token, which for
		// `async (bar = 1) => {}` is `async`, so testing whether the SLICE starts with `(` reads
		// the modifier and wrongly adds a second pair.
		// Start at the TYPE PARAMETER list when there is one, and at the parameter list otherwise.
		// Two things make this more than a scan for `(`. The arrow's first token is `async` for
		// `async (bar) => {}`, which `prefix` already carries, so slicing from it writes the
		// modifier twice. And a generic arrow `<T>(): void => {}` keeps its `<T>`, which sits
		// BEFORE the parenthesis: slicing from `(` silently drops the type parameters, which cost
		// one corpus case whose four findings all lost their generics.
		parameterStart = objectShorthandSignatureStart(ctx, value, parameterStart,
			arrow.EqualsGreaterThanToken.Pos())
		parameters := strings.TrimRight(sourceText[parameterStart:arrow.EqualsGreaterThanToken.Pos()], " \t")
		if declared := value.Parameters(); len(declared) == 1 {
			// A single parameter may be written without parentheses -- `foo => {}` -- and a
			// method's list always has them. The test is whether a `(` appears before the
			// parameter itself, not anywhere in the slice.
			parameterStartOffset := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, declared[0].Pos()).Pos()
			hasParenthesis := false
			for index := parameterStart; index < parameterStartOffset; index++ {
				if sourceText[index] == '(' {
					hasParenthesis = true
					break
				}
			}
			if !hasParenthesis {
				parameters = text.TrimWhitespace(sourceText[parameterStartOffset:arrow.EqualsGreaterThanToken.Pos()])
				parameters = "(" + text.TrimWhitespace(parameters) + ")"
			}
		}
		separator := sourceText[arrow.EqualsGreaterThanToken.End():bodyStart]
		return rule.ReplaceRange(core.NewTextRange(keyStart, member.End()),
			prefix+keyText+parameters+separator+sourceText[bodyStart:value.End()]), true
	}

	// Everything from after the `function` keyword onward, which survives unchanged.
	tail := objectShorthandParameterListStart(ctx, value, generator)
	if tail < 0 {
		return rule.Fix{}, false
	}
	return rule.ReplaceRange(core.NewTextRange(keyStart, member.End()),
		prefix+keyText+sourceText[tail:value.End()]), true
}

// objectShorthandParameterListStart finds the offset the converted method keeps from.
//
// For a function expression that is the `(` of its parameter list, which upstream reaches as the
// token after `function` (or after the `*` for a generator). For an arrow it is the arrow's own
// start, since an arrow has no keyword to drop.
func objectShorthandParameterListStart(ctx rule.Context, value *ast.Node, generator bool) int {
	text := ctx.SourceFile.Text()
	limit := value.End()
	if body := value.Body(); body != nil {
		limit = body.Pos()
	}
	if limit > len(text) {
		limit = len(text)
	}
	// Upstream slices from the token after `function`, or after the `*` for a generator, NOT from
	// the parenthesis. The difference is the whitespace the author wrote between them:
	// `foo: async function () {}` becomes `async foo () {}`, keeping the space, where slicing from
	// `(` would produce `async foo() {}`. An arrow has no keyword to skip, so its whole text is
	// kept from its own start.
	if value.Kind == ast.KindArrowFunction {
		return scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, value.Pos()).Pos()
	}
	for index := value.Pos(); index+8 <= limit && index+8 <= len(text); index++ {
		if text[index:index+8] == "function" {
			after := index + 8
			if generator {
				for after < limit && text[after] != '*' {
					after++
				}
				if after < limit {
					after++
				}
			}
			// Immediately after the keyword, which is upstream's slice and is right for every
			// shape at once. It keeps the whitespace the author wrote -- `foo: async function () {}`
			// becomes `async foo () {}` with the space -- and it keeps a generic's type parameters,
			// since `<T>` sits between the keyword and the parameter list and is therefore already
			// inside the slice. Measured against the installed rule:
			//
			//	({ key: function <T>(a: T): T {...} })   ->  ({ key <T>(a: T): T {...} })
			//
			// An earlier draft moved this offset forward to the `<` or the `(`, which was both
			// unnecessary and wrong: it dropped the space.
			return after
		}
	}
	return -1
}

// objectShorthandSignatureStart finds where a converted method's signature begins: the `<` of a
// type parameter list when there is one, and the `(` of the parameter list otherwise.
//
// Upstream needs no equivalent because its corpus's generic cases come through a fixture parser and
// its slice happens to start early enough. Here the offset has to be computed, and getting it wrong
// drops the generics without changing any message id.
func objectShorthandSignatureStart(ctx rule.Context, value *ast.Node, from int, to int) int {
	text := ctx.SourceFile.Text()
	if to > len(text) {
		to = len(text)
	}
	if typeParameters := value.TypeParameterList(); typeParameters != nil && len(typeParameters.Nodes) > 0 {
		for index := from; index < to; index++ {
			if text[index] == '<' {
				return index
			}
		}
	}
	for index := from; index < to; index++ {
		if text[index] == '(' {
			return index
		}
	}
	return from
}

// objectShorthandHasCommentBetween answers upstream's `commentsExistBetween`.
func objectShorthandHasCommentBetween(text string, start int, end int) bool {
	if start < 0 || end > len(text) || start >= end {
		return false
	}
	for index := start; index+1 < end; index++ {
		if text[index] == '/' && (text[index+1] == '/' || text[index+1] == '*') {
			return true
		}
	}
	return false
}

// objectShorthandMemberStart is the member's first TOKEN, which is upstream's `node.range[0]`.
//
// `Pos()` begins at the previous token's end, so it includes the leading trivia -- the whitespace
// AND any comment sitting between the previous property and this one. Upstream's
// `getCommentsInside` looks only within the node's own range, so a comment before the property is
// not inside it. Scanning from `Pos()` made a comment between two properties look like a comment
// inside the second, and the fixer declined a repair upstream performs.
func objectShorthandMemberStart(ctx rule.Context, member *ast.Node) int {
	return scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, member.Pos()).Pos()
}
