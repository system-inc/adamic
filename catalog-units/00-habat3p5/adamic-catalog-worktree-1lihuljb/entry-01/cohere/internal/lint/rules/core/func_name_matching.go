package core

import (
	"encoding/json"
	"fmt"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// FuncNameMatchingOptions configures the rule.
//
// Upstream's schema is unusual and worth stating exactly, because it decides how the decoder has to
// behave. It is an `anyOf` of two array shapes:
//
//	["always" | "never", {options}]     a direction followed by an options object
//	[{options}]                         an options object alone, direction defaulting to "always"
//
// So the FIRST element is either a string or the object. The rule registers with `DecodeOptionList`
// and is handed upstream's list as written; see DecodeFuncNameMatchingOptions, which reads both.
type FuncNameMatchingOptions struct {
	// Direction is upstream's `nameMatches`: "always" requires the names to match, "never" requires
	// them to differ. Empty means "always".
	Direction string `json:"-"`

	// ConsiderPropertyDescriptor makes `{value: function foo(){}}` inside `Object.defineProperty`
	// compare against the PROPERTY being defined rather than against the literal key `value`.
	ConsiderPropertyDescriptor bool `json:"considerPropertyDescriptor"`

	// IncludeCommonJSModuleExports stops `module.exports = function foo(){}` from being exempt.
	IncludeCommonJSModuleExports bool `json:"includeCommonJSModuleExports"`
}

// funcNameMatchingIsIdentifier answers upstream's `isIdentifier`, which routes through `esutils`.
//
// An earlier draft was an ASCII regex, documented as a deliberate narrowing on the reasoning that
// every corpus case is ASCII. That reasoning was wrong and upstream's own corpus said so:
// `var obj = { '\u1885': function foo() {} };` is a reporting case, and U+1885 is not ASCII. The
// same shape as the `no-invalid-html-attribute` whitespace gap -- a written justification for a
// limitation, surviving exactly until a fixture reached it.
//
// # Which of upstream's two checks this is, measured rather than assumed
//
// Upstream picks `isIdentifierES6` when `ecmaVersion >= 2015` and `isIdentifierES5` otherwise.
// Driving `esutils` directly on the corpus character settles which one matters here:
//
//	U+1885 MONGOLIAN LETTER ALI GALI BALUDA    ES5=false  ES6=true
//
// Unicode 8 moved it into `Other_ID_Start`, so it is an identifier start under ES6 and not under
// ES5.
//
// # The corpus carries that source TWICE, with opposite verdicts, and that is the whole answer
//
// `var obj = { '\u1885': function foo() {} };` appears as both `invalid[49]` and `valid[142]`,
// byte for byte identical. The only difference is a field a careless extractor throws away:
//
//	invalid[49]   languageOptions: {ecmaVersion: 6}                 reports
//	valid[142]    languageOptions: {ecmaVersion: 5, sourceType: script}   clean
//
// So it is a version gate rather than a contradiction, and a rule that hardcodes either answer must
// disagree with one of the two rows. `rule.Context` has no `ecmaVersion` channel, so this port picks
// ES6 deliberately: every configuration in this tree is well past 2015, an ES5-pinned project is not
// what cohere lints, and choosing ES5 would silently drop findings on modern source. The `valid[142]`
// row is recorded in the test file as a known divergence with the reason, rather than being deleted
// to make the suite green.
//
// # Go ships the tables this needs
//
// `ID_Start` is `L | Nl | Other_ID_Start` and `ID_Continue` adds `Mn | Mc | Nd | Pc |
// Other_ID_Continue`, plus the two zero-width joiners JavaScript allows. `unicode.Other_ID_Start`
// is the table that carries U+1885, and reaching for a regex instead is what hid it. Checked
// against `esutils` on seven inputs including the corpus character, all agreeing.
func funcNameMatchingIsIdentifier(name string) bool {
	if name == "" {
		return false
	}
	for index, character := range name {
		if index == 0 {
			if !funcNameMatchingIsIdentifierStart(character) {
				return false
			}
			continue
		}
		if !funcNameMatchingIsIdentifierPart(character) {
			return false
		}
	}
	return true
}

// funcNameMatchingIsIdentifierStart answers ES6's IdentifierStart.
func funcNameMatchingIsIdentifierStart(character rune) bool {
	return character == '$' || character == '_' ||
		unicode.In(character, unicode.L, unicode.Nl, unicode.Other_ID_Start)
}

// funcNameMatchingIsIdentifierPart answers ES6's IdentifierPart.
//
// The two zero-width characters are U+200C ZERO WIDTH NON-JOINER and U+200D ZERO WIDTH JOINER,
// which JavaScript admits in identifiers and Unicode's ID_Continue does not. Written as escapes
// rather than as literals: this file is scanned for non-ASCII bytes.
func funcNameMatchingIsIdentifierPart(character rune) bool {
	if funcNameMatchingIsIdentifierStart(character) {
		return true
	}
	if character == '\u200c' || character == '\u200d' {
		return true
	}
	return unicode.In(character, unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc,
		unicode.Other_ID_Continue)
}

// The four messages. Which one fires is a product of two booleans -- the configured direction, and
// whether the target is a property or a variable -- and upstream picks among them in a single
// if/else chain that this port reproduces in `funcNameMatchingMessage`.
func funcNameMatchingMessage(direction string, isProperty bool, functionName string, name string) rule.Message {
	if direction == "never" {
		if isProperty {
			return rule.Message{
				Id: "notMatchProperty",
				Description: fmt.Sprintf(
					"Function name `%s` should not match property name `%s`. This project has "+
						"configured `never`, which is the convention where a function's own name is "+
						"reserved for stack traces and the property is just where it happens to be "+
						"attached. Repeating the property name in the function makes the two look "+
						"linked when renaming one will not rename the other.",
					functionName, name,
				),
			}
		}
		return rule.Message{
			Id: "notMatchVariable",
			Description: fmt.Sprintf(
				"Function name `%s` should not match variable name `%s`. This project has configured "+
					"`never`, so the function's own name is expected to carry information the "+
					"variable does not, rather than repeating it.",
				functionName, name,
			),
		}
	}

	if isProperty {
		return rule.Message{
			Id: "matchProperty",
			Description: fmt.Sprintf(
				"Function name `%s` should match property name `%s`. A named function expression's "+
					"name is what appears in stack traces and in the debugger, so when it disagrees "+
					"with the property it is reached through, every trace names something the reader "+
					"cannot find by searching. Rename the function to match, or drop its name and let "+
					"it be anonymous.",
				functionName, name,
			),
		}
	}
	return rule.Message{
		Id: "matchVariable",
		Description: fmt.Sprintf(
			"Function name `%s` should match variable name `%s`. A named function expression's name "+
				"is what appears in stack traces and in the debugger, so when it disagrees with the "+
				"variable holding it, every trace names something the reader cannot find by "+
				"searching. Rename the function to match, or drop its name and let it be anonymous.",
			functionName, name,
		),
	}
}

// FuncNameMatching requires a named function expression's name to match what it is assigned to.
//
//	valid:   var foo = function foo() {};
//	valid:   var foo = function() {};
//	valid:   obj.foo = function foo() {};
//	valid:   module.exports = function foo() {};
//	invalid: var foo = function bar() {};
//	invalid: obj.foo = function bar() {};
//	invalid: ({ foo: function bar() {} });
//
// Ported from `func-name-matching` in ESLint, read from the clone at
// `lib/rules/func-name-matching.js`. Two options plus a direction, four messages, no fixer.
//
// The whole 193-case corpus was extracted from upstream's own tester and replayed against the rule
// through the ESLint Linter API before any code was written.
//
// # The option surface is two shapes, not one, and the decoder has to accept both
//
// `meta.schema` is an `anyOf`: the direction string may be present or absent, so `["never", {...}]`
// and `[{...}]` are both valid and mean different things. Upstream reads it as
// `typeof context.options[0] === "object" ? options[0] : options[1]` for the object and
// `typeof context.options[0] === "string" ? options[0] : "always"` for the direction. Both spellings
// appear in the corpus, which is what makes the decoder's job real rather than theoretical.
//
// # `module.exports` is exempt by default, and that exemption is the whole of one option
//
// `module.exports = function foo() {}` is CLEAN unless `includeCommonJSModuleExports` is set. The
// test covers both `module.exports` and `module["exports"]`, which upstream spells out separately
// because the second is a computed access whose property is a literal.
//
// # An ARROW function is never checked, and neither is an anonymous function
//
// The rule compares a function's OWN name against its target, so a function with no name has nothing
// to compare and an arrow function cannot have one. Every arm requires
// `node.init.type === "FunctionExpression"` and then `node.init.id`, so both exclusions come from
// the same two lines rather than from a separate guard.
var FuncNameMatching = rule.Rule{
	Name: "func-name-matching",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		resolved := DefaultFuncNameMatchingOptions()
		if given, isFuncNameMatchingOptions := rule.OptionsAs[FuncNameMatchingOptions](options); isFuncNameMatchingOptions {
			resolved = given
		}
		direction := resolved.Direction
		if direction == "" {
			direction = "always"
		}

		checker := &funcNameMatchingChecker{
			ctx:                        ctx,
			direction:                  direction,
			considerPropertyDescriptor: resolved.ConsiderPropertyDescriptor,
			includeModuleExports:       resolved.IncludeCommonJSModuleExports,
		}

		return rule.Listeners{
			ast.KindVariableDeclaration: checker.checkVariableDeclarator,
			ast.KindBinaryExpression:    checker.checkAssignment,
			ast.KindPropertyAssignment:  checker.checkProperty,
			ast.KindPropertyDeclaration: checker.checkProperty,
		}
	},
}

// funcNameMatchingChecker carries the resolved configuration across the listeners.
type funcNameMatchingChecker struct {
	ctx                        rule.Context
	direction                  string
	considerPropertyDescriptor bool
	includeModuleExports       bool
}

// shouldWarn answers upstream's `shouldWarn`: the direction decides which way the comparison runs.
func (checker *funcNameMatchingChecker) shouldWarn(left string, right string) bool {
	if checker.direction == "never" {
		return left == right
	}
	return left != right
}

// report emits the finding, choosing among the four messages the way upstream's `report` does.
func (checker *funcNameMatchingChecker) report(node *ast.Node, name string, functionName string, isProperty bool) {
	checker.ctx.ReportNode(node, funcNameMatchingMessage(checker.direction, isProperty, functionName, name))
}

// checkVariableDeclarator is upstream's `VariableDeclarator` listener.
//
// Our parser groups declarators under one `KindVariableDeclaration` per `var`, so the listener fires
// once and iterates rather than firing per declarator. Upstream reports on the DECLARATOR, so the
// finding is anchored on each declaration individually rather than on the statement.
func (checker *funcNameMatchingChecker) checkVariableDeclarator(node *ast.Node) {
	declaration := node.AsVariableDeclaration()
	if declaration == nil || declaration.Initializer == nil {
		return
	}
	name := declaration.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return
	}
	functionName, named := funcNameMatchingFunctionExpressionName(declaration.Initializer)
	if !named {
		return
	}
	if checker.shouldWarn(name.Text(), functionName) {
		checker.report(node, name.Text(), functionName, false)
	}
}

// checkAssignment is upstream's `AssignmentExpression` listener.
func (checker *funcNameMatchingChecker) checkAssignment(node *ast.Node) {
	expression := node.AsBinaryExpression()
	if expression == nil || expression.OperatorToken == nil ||
		!funcNameMatchingIsAssignmentOperator(expression.OperatorToken.Kind) {
		return
	}
	functionName, named := funcNameMatchingFunctionExpressionName(expression.Right)
	if !named {
		return
	}

	target := expression.Left
	switch target.Kind {
	case ast.KindIdentifier:
		if checker.shouldWarn(target.Text(), functionName) {
			checker.report(node, target.Text(), functionName, false)
		}

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		// `module.exports = function foo(){}` is exempt unless the option turns it on. Checked
		// before the name is read, matching upstream's guard order.
		if !checker.includeModuleExports && funcNameMatchingIsModuleExports(target) {
			return
		}
		name, readable := funcNameMatchingStaticPropertyName(target)
		if !readable || !funcNameMatchingIsIdentifier(name) {
			return
		}
		if checker.shouldWarn(name, functionName) {
			checker.report(node, name, functionName, true)
		}
	}
}

// checkProperty is upstream's `"Property, PropertyDefinition[value]"` listener.
func (checker *funcNameMatchingChecker) checkProperty(node *ast.Node) {
	value := funcNameMatchingPropertyValue(node)
	if value == nil {
		return
	}
	functionName, named := funcNameMatchingFunctionExpressionName(value)
	if !named {
		return
	}

	key := node.Name()
	if key == nil {
		return
	}

	// A COMPUTED key whose expression is a string literal is still a static name upstream:
	// `({['foo']: function bar(){}})` reports. Any other computed key -- `['b' + 'ar']`, or a
	// variable -- is not comparable and is declined. Unwrapping here rather than in the literal
	// arm keeps the identifier arm below reading only genuine identifiers.
	if key.Kind == ast.KindComputedPropertyName {
		computed := key.AsComputedPropertyName()
		if computed == nil || computed.Expression == nil ||
			computed.Expression.Kind != ast.KindStringLiteral {
			return
		}
		keyText := computed.Expression.Text()
		if !funcNameMatchingIsIdentifier(keyText) {
			return
		}
		if checker.shouldWarn(keyText, functionName) {
			checker.report(node, keyText, functionName, true)
		}
		return
	}

	if key.Kind == ast.KindIdentifier {
		propertyName := key.Text()

		// The property-descriptor arm. `{value: function foo(){}}` inside `Object.defineProperty`
		// compares against the property BEING DEFINED rather than against the literal key `value`.
		if checker.considerPropertyDescriptor && propertyName == "value" &&
			node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression {
			descriptorName, isDescriptor, readable := funcNameMatchingDescriptorTarget(node.Parent)
			if isDescriptor {
				// Inside a descriptor whose target name cannot be read -- `'b' + 'ar'`, or a
				// computed key -- upstream reports NOTHING. It does not fall back to comparing
				// against the literal key `value`, which is what a reader assumes and what an
				// earlier draft here did: each of its branches guards the report with
				// `isStringLiteral(property)` or `!node.parent.parent.computed` and simply
				// declines otherwise. Four of upstream's clean cases pin this.
				if !readable {
					return
				}
				if checker.shouldWarn(descriptorName, functionName) {
					checker.report(node, descriptorName, functionName, true)
				}
				return
			}
		}

		if checker.shouldWarn(propertyName, functionName) {
			checker.report(node, propertyName, functionName, true)
		}
		return
	}

	// A string-literal key is compared only when it spells a legal identifier, so `{'foo-bar':
	// function foo(){}}` is silent: renaming the function to match would not be writable as a name.
	if key.Kind == ast.KindStringLiteral {
		keyText := key.Text()
		if !funcNameMatchingIsIdentifier(keyText) {
			return
		}
		if checker.shouldWarn(keyText, functionName) {
			checker.report(node, keyText, functionName, true)
		}
	}
}

// funcNameMatchingPropertyValue reads the value of a property-shaped node.
func funcNameMatchingPropertyValue(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAssignment:
		if assignment := node.AsPropertyAssignment(); assignment != nil {
			return assignment.Initializer
		}
	case ast.KindPropertyDeclaration:
		if declaration := node.AsPropertyDeclaration(); declaration != nil {
			return declaration.Initializer
		}
	}
	return nil
}

// funcNameMatchingFunctionExpressionName reads a NAMED function expression's own name.
//
// Both halves are the rule's whole precondition. It must be a function EXPRESSION, which excludes an
// arrow (which cannot be named) and a function declaration (whose name is its binding); and it must
// be NAMED, since an anonymous function has nothing to compare.
func funcNameMatchingFunctionExpressionName(node *ast.Node) (string, bool) {
	if node == nil || node.Kind != ast.KindFunctionExpression {
		return "", false
	}
	expression := node.AsFunctionExpression()
	if expression == nil {
		return "", false
	}
	name := expression.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return "", false
	}
	return name.Text(), true
}

// funcNameMatchingIsModuleExports answers upstream's `isModuleExports`.
//
// Two accepted spellings, and upstream writes them out separately: `module.exports` as a property
// access, and `module["exports"]` as an element access whose argument is a string literal. Anything
// else, including `notModule.exports` and `module[exports]`, is not exempt.
func funcNameMatchingIsModuleExports(target *ast.Node) bool {
	switch target.Kind {
	case ast.KindPropertyAccessExpression:
		access := target.AsPropertyAccessExpression()
		if access == nil || access.Expression == nil || access.Name() == nil {
			return false
		}
		return access.Expression.Kind == ast.KindIdentifier &&
			access.Expression.Text() == "module" &&
			access.Name().Kind == ast.KindIdentifier &&
			access.Name().Text() == "exports"

	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		if access == nil || access.Expression == nil || access.ArgumentExpression == nil {
			return false
		}
		return access.Expression.Kind == ast.KindIdentifier &&
			access.Expression.Text() == "module" &&
			access.ArgumentExpression.Kind == ast.KindStringLiteral &&
			access.ArgumentExpression.Text() == "exports"
	}
	return false
}

// funcNameMatchingStaticPropertyName answers upstream's `getStaticPropertyName` for an assignment
// target.
//
// A computed access is readable only when its argument is a literal, which is upstream's
// `node.left.computed && node.left.property.type !== "Literal"` guard inverted.
func funcNameMatchingStaticPropertyName(target *ast.Node) (string, bool) {
	switch target.Kind {
	case ast.KindPropertyAccessExpression:
		access := target.AsPropertyAccessExpression()
		if access == nil || access.Name() == nil || access.Name().Kind != ast.KindIdentifier {
			return "", false
		}
		return access.Name().Text(), true

	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		if access == nil || access.ArgumentExpression == nil {
			return "", false
		}
		if access.ArgumentExpression.Kind != ast.KindStringLiteral {
			return "", false
		}
		return access.ArgumentExpression.Text(), true
	}
	return "", false
}

// funcNameMatchingDescriptorTarget finds the property name a descriptor object is defining.
//
// Three shapes, which upstream handles as three branches off the same `considerPropertyDescriptor`
// arm:
//
//	Object.defineProperty(o, 'foo', {value: function bar(){}})     the second ARGUMENT names it
//	Object.defineProperties(o, {foo: {value: function bar(){}}})   the grandparent KEY names it
//	Object.create(o, {foo: {value: function bar(){}}})             likewise
//
// Three returns rather than two, and the third is the one an earlier draft got wrong. They separate:
//
//	isDescriptor=false             not a descriptor at all; the caller compares against the key
//	isDescriptor=true, readable=false   a descriptor whose target name is not static, so upstream
//	                               reports NOTHING rather than falling back
//	isDescriptor=true, readable=true    the target name, to compare against
func funcNameMatchingDescriptorTarget(descriptor *ast.Node) (name string, isDescriptor bool, readable bool) {
	call := descriptor.Parent
	if call == nil {
		return "", false, false
	}

	// `Object.defineProperty(o, 'foo', {...})` and `Reflect.defineProperty(...)`: the descriptor is
	// an argument of the call, and the name is the argument before it.
	if funcNameMatchingIsPropertyCall(call, "Object", "defineProperty") ||
		funcNameMatchingIsPropertyCall(call, "Reflect", "defineProperty") {
		arguments := call.AsCallExpression().Arguments
		if arguments == nil || len(arguments.Nodes) < 2 {
			return "", true, false
		}
		second := arguments.Nodes[1]
		if second.Kind != ast.KindStringLiteral {
			return "", true, false
		}
		return second.Text(), true, true
	}

	// `Object.defineProperties(o, {foo: {...}})` and `Object.create(o, {foo: {...}})`: the descriptor
	// is the value of a property, whose key names the target. Upstream reaches it as
	// `node.parent.parent.key` with the call at `node.parent.parent.parent.parent`.
	property := descriptor.Parent
	if property == nil || property.Kind != ast.KindPropertyAssignment {
		return "", false, false
	}
	container := property.Parent
	if container == nil {
		return "", false, false
	}
	outerCall := container.Parent
	if !funcNameMatchingIsPropertyCall(outerCall, "Object", "defineProperties") &&
		!funcNameMatchingIsPropertyCall(outerCall, "Object", "create") {
		return "", false, false
	}
	// Upstream guards with `!node.parent.parent.computed`, so a computed key here is a descriptor
	// whose name cannot be read rather than a non-descriptor: `Object.defineProperties(foo,
	// {['bar']: {value: function bar(){}}})` is CLEAN.
	key := property.Name()
	if key == nil || key.Kind != ast.KindIdentifier {
		return "", true, false
	}
	return key.Text(), true, true
}

// funcNameMatchingIsPropertyCall answers upstream's `isPropertyCall`.
func funcNameMatchingIsPropertyCall(node *ast.Node, objectName string, functionName string) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil {
		return false
	}
	// The callee may be parenthesized, and the access inside it may be optional. Upstream's
	// `isSpecificMemberAccess` sees neither, because espree folds parentheses away and treats an
	// optional access as the same node kind. Both `(Object?.defineProperty)(...)` and
	// `(Object.defineProperty)(...)` are in the corpus and both report, so the unwrap is required
	// rather than defensive. A loop because `((x))` nests, and not `ast.SkipParentheses`, which
	// dereferences its argument.
	callee := call.Expression
	for callee != nil && callee.Kind == ast.KindParenthesizedExpression {
		callee = callee.AsParenthesizedExpression().Expression
	}
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if access == nil || access.Expression == nil || access.Name() == nil {
		return false
	}
	return access.Expression.Kind == ast.KindIdentifier &&
		access.Expression.Text() == objectName &&
		access.Name().Kind == ast.KindIdentifier &&
		access.Name().Text() == functionName
}

// DefaultFuncNameMatchingOptions is the unconfigured answer: direction "always", both flags off.
func DefaultFuncNameMatchingOptions() FuncNameMatchingOptions {
	return FuncNameMatchingOptions{Direction: "always"}
}

// DecodeFuncNameMatchingOptions reads upstream's option list off the config.
//
// Hand-rolled because upstream's schema is an `anyOf` of two LIST shapes rather than a single
// object, and the generic decoder has no way to express that:
//
//	["never"]                                 the direction alone
//	["never", {"considerPropertyDescriptor": true}]  both, in upstream's order
//	[{"considerPropertyDescriptor": true}]    the options object alone, direction "always"
//
// Upstream reads the object from `options[0]` when that is an object and from `options[1]`
// otherwise, which is the order below. Anything else is refused rather than defaulted: a direction
// that is neither "always" nor "never", a key the schema does not declare, a second element after an
// object, or a third element. Each was silently read as the default by the decoder this replaced,
// and the defaults are exactly what the author wrote the option to change.
func DecodeFuncNameMatchingOptions(list []byte) (any, error) {
	options := DefaultFuncNameMatchingOptions()
	elements, err := rule.OptionElements(list, 2)
	if err != nil || len(elements) == 0 {
		return options, err
	}

	objectElement := elements[0]
	var direction string
	if err := json.Unmarshal(elements[0], &direction); err == nil {
		if direction != "always" && direction != "never" {
			return options, fmt.Errorf(
				"func-name-matching: unknown direction %q, wanted always or never", direction)
		}
		options.Direction = direction
		if len(elements) < 2 {
			return options, nil
		}
		objectElement = elements[1]
	} else if len(elements) > 1 {
		return options, fmt.Errorf(
			"func-name-matching: an options object first takes no second element, so %s would "+
				"never be read", elements[1])
	}

	var object FuncNameMatchingOptions
	if err := rule.UnmarshalOptions(objectElement, &object); err != nil {
		return options, fmt.Errorf("func-name-matching options object: %w", err)
	}
	options.ConsiderPropertyDescriptor = object.ConsiderPropertyDescriptor
	options.IncludeCommonJSModuleExports = object.IncludeCommonJSModuleExports
	return options, nil
}

// funcNameMatchingIsAssignmentOperator answers whether a binary operator assigns.
//
// Upstream's listener is `AssignmentExpression`, which in ESTree covers EVERY assignment operator
// rather than just `=`. Our parser models all of them as binary expressions, so the operator has to
// be tested explicitly, and restricting it to `=` costs three of upstream's reporting cases:
// `foo &&= function bar(){}`, `obj.foo ||= ...` and `obj['foo'] ??= ...` are all in the corpus.
//
// The logical assignments are the ones a reader is most likely to leave out, because they are the
// newest and because the rule reads as being about plain assignment.
func funcNameMatchingIsAssignmentOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}
