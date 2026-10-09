package javascript

import (
	"regexp"
	"strconv"
	"unicode"

	"github.com/system-inc/cohere/internal/format/printing"
)

// print/return-statement.js, variable-declaration.js, sequence-expression.js, await-expression.js,
// rest-element.js, property.js and key.js.

// printReturnOrThrowArgument is upstream's printReturnOrThrowArgument. experimentalTernaries is never
// set in our configurations, so its branch is not taken.
func printReturnOrThrowArgument(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	argumentDoc := print(nil, nil)
	if returnArgumentHasLeadingComment(current, options) {
		return concatIn(path, "(", indentIn(path, concatIn(path, hardline, argumentDoc)), hardline, ")")
	}
	if isBinaryish(current) {
		return groupIn(path, concatIn(path, ifBreak("(", ""), indentIn(path, concatIn(path, softline, argumentDoc)), softline, ifBreak(")", "")))
	}
	return argumentDoc
}

// printReturnOrThrowStatement is upstream's printReturnOrThrowStatement, exported upstream as
// printReturnStatement and printThrowStatement.
func printReturnOrThrowStatement(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	keyword := "return"
	if current.Is("ThrowStatement") {
		keyword = "throw"
	}
	var argument Doc = emptyDoc
	if current.Child("argument") != nil {
		argument = concatIn(path, " ", call(path, func(path *Path) Doc {
			return printReturnOrThrowArgument(path, options, print)
		}, "argument"))
	}
	return concatIn(path, keyword, argument, printSemicolon(options))
}

func printReturnStatement(path *Path, options *Options, print PrintFunc) Doc {
	return printReturnOrThrowStatement(path, options, print)
}

func printThrowStatement(path *Path, options *Options, print PrintFunc) Doc {
	return printReturnOrThrowStatement(path, options, print)
}

// printVariableDeclaration is upstream's printVariableDeclaration.
func printVariableDeclaration(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	printed := printAll(path, print, "declarations")
	key := keyOf(path)
	parent := parentOf(path)
	isForXStatementInitializer := key == "init" && parent.Is("ForStatement") ||
		key == "left" && parent.Is("ForInStatement", "ForOfStatement")

	declarations := current.List("declarations")
	hasValue := false
	for _, declarator := range declarations {
		if declarator.Child("init") != nil {
			hasValue = true
			break
		}
	}

	var firstVariable Doc
	if len(printed) == 1 && !hasAnyComment(declarations[0]) {
		firstVariable = printed[0]
	} else if len(printed) > 0 {
		firstVariable = indentIn(path, printed[0])
	}

	var first Doc = emptyDoc
	if firstVariable != nil {
		first = concatIn(path, " ", firstVariable)
	}
	rest := make([]any, 0, len(printed))
	for _, declaration := range printed[min(1, len(printed)):] {
		separator := line
		if hasValue && !isForXStatementInitializer {
			separator = hardline
		}
		rest = append(rest, concatIn(path, ",", separator, declaration))
	}
	var semicolon Doc = emptyDoc
	if !isForXStatementInitializer {
		semicolon = printSemicolon(options)
	}
	return groupIn(path, concatIn(path, printDeclareToken(path),
		current.String("kind"),
		first,
		indentIn(path, concatIn(path, rest...)),
		semicolon,
	))
}

// shouldIndentSequenceExpression is upstream's shouldIndentSequenceExpression.
func shouldIndentSequenceExpression(path *Path, options *Options) bool {
	key := keyOf(path)
	parent := parentOf(path)
	if key == "argument" && isReturnOrThrowStatement(parent) && needsParentheses(path, options) {
		return true
	}
	return key == "body" && parent.Is("ArrowFunctionExpression")
}

// printSequenceExpression is upstream's printSequenceExpression.
func printSequenceExpression(path *Path, options *Options, print PrintFunc) Doc {
	parent := parentOf(path)
	if parent.Is("ExpressionStatement", "ForStatement") {
		var parts []any
		each(path, func(path *Path, _ int) {
			if path.IsFirst() {
				parts = append(parts, print(nil, nil))
			} else {
				parts = append(parts, ",", indentIn(path, concatIn(path, line, print(nil, nil))))
			}
		}, "expressions")
		return groupIn(path, concatIn(path, parts...))
	}
	parts := join(concatIn(path, ",", line), printAll(path, print, "expressions"))
	if shouldIndentSequenceExpression(path, options) {
		return groupIn(path, ifBreak(concatIn(path, indentIn(path, concatIn(path, softline, parts)), softline), parts))
	}
	return groupIn(path, parts)
}

// printAwaitExpression is upstream's printAwaitExpression.
func printAwaitExpression(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	parts := []any{"await"}
	if current.Child("argument") != nil {
		parts = append(parts, " ", print("argument", nil))
		parent := parentOf(path)
		if isCallExpression(parent) && parent.Child("callee") == current ||
			isMemberExpression(parent) && parent.Child("object") == current {
			parts = []any{indentIn(path, concatIn(path, append([]any{softline}, parts...)...)), softline}
			// avoid printing `await (await` on one line
			parentAwaitOrBlock, _ := path.FindAncestor(func(node Node) bool {
				return node.Is("AwaitExpression", "BlockStatement")
			})
			if !parentAwaitOrBlock.Is("AwaitExpression") ||
				!startsWithNoLookaheadToken(parentAwaitOrBlock.Child("argument"), func(leftmost Node) bool {
					return leftmost == current
				}) {
				return groupIn(path, concatIn(path, parts...))
			}
		}
	}
	return concatIn(path, parts...)
}

// printRestOrSpreadElement is upstream's printRestOrSpreadElement, exported upstream as
// printRestElement and printSpreadElement.
func printRestOrSpreadElement(path *Path, print PrintFunc) Doc {
	return concatIn(path, "...", print("argument", nil), printTypeAnnotationProperty(path, print))
}

func printRestElement(path *Path, print PrintFunc) Doc { return printRestOrSpreadElement(path, print) }
func printSpreadElement(path *Path, print PrintFunc) Doc {
	return printRestOrSpreadElement(path, print)
}

// printProperty is upstream's printProperty: Property and ImportAttribute.
func printProperty(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	if current.Bool("shorthand") {
		return print("value", nil)
	}
	return printAssignment(path, options, print, printKey(path, options, print), ":", "value")
}

// print/key.js. Our parsers are "typescript", "babel" and the JSON ones, and quoteProps is always "as-needed".

func isTsEnumMember(node Node) bool { return node.Is("TSEnumMember") }

func getKeyProperty(node Node) string {
	if isTsEnumMember(node) {
		return "id"
	}
	return "key"
}

func getKey(node Node) Node { return node.Child(getKeyProperty(node)) }

func isComputedKey(node Node) bool { return !isTsEnumMember(node) && node.Bool("computed") }

var simpleNumber = regexp.MustCompile(`^(?:\d+|\d+\.\d+)$`)

// isSimpleNumber is upstream's isSimpleNumber.
func isSimpleNumber(numberString string) bool { return simpleNumber.MatchString(numberString) }

// isTypeScript is upstream's isTypeScript(options): of its parsers, ours is "typescript".
func isTypeScript(options *Options) bool { return settingsOf(options).Parser == "typescript" }

// isKeySafeToQuote is upstream's isKeySafeToQuote.
func isKeySafeToQuote(node Node, options *Options) bool {
	key := getKey(node)
	if key.Is("Identifier") {
		return true
	}
	if !isNumericLiteral(key) {
		return false
	}
	// Quoting number keys is safe in JS and Flow, but not in TypeScript (as
	// mentioned in `isKeySafeToUnquote`).
	if isTypeScript(options) {
		return false
	}
	printedNumber := printNumber(getRaw(key))
	// Avoid converting 999999999999999999999 to 1e+21, 0.99999999999999999 to 1 and 1.0 to 1.
	return javaScriptNumberString(key.Get("value").(float64)) == printedNumber && isSimpleNumber(printedNumber)
}

// isKeySafeToUnquote is upstream's isKeySafeToUnquote. Of its parsers, ours are "typescript", "babel"
// and the JSON ones.
func isKeySafeToUnquote(node Node, options *Options) bool {
	switch settingsOf(options).Parser {
	case "json", "jsonc":
		return false
	}
	key := getKey(node)
	if !isStringLiteral(key) {
		return false
	}
	value := key.String("value")
	printed := printString(getRaw(key), options)
	if printed[1:len(printed)-1] != value {
		return false
	}
	if node.Is("TSMethodSignature") && value == "new" {
		return false
	}
	// With --strictPropertyInitialization, TypeScript treats quoted property names differently.
	if !(isTypeScript(options) && node.Is("PropertyDefinition")) && isEs5IdentifierName(value) {
		return true
	}
	// Safe to unquote as number
	// Note: It's also not safe for TypeScript as mentioned above
	// https://github.com/prettier/prettier/pull/8508
	if settingsOf(options).Parser == "babel" && !node.Is("ImportAttribute") && isSimpleNumber(value) {
		number, _ := strconv.ParseFloat(value, 64)
		return javaScriptNumberString(number) == value
	}
	return false
}

// shouldQuoteKey is upstream's shouldQuoteKey. quoteProps is always "as-needed", so only the JSON
// parsers quote.
func shouldQuoteKey(path *Path, options *Options) bool {
	switch settingsOf(options).Parser {
	case "json", "jsonc":
		return isKeySafeToQuote(node(path), options)
	}
	return false
}

// shouldUnquoteKey is upstream's shouldUnquoteKey with quoteProps "as-needed".
func shouldUnquoteKey(path *Path, options *Options) bool {
	return isKeySafeToUnquote(node(path), options)
}

var leadingDigit = regexp.MustCompile(`^\d`)

// printKey is upstream's printKey.
func printKey(path *Path, options *Options, print PrintFunc) Doc {
	current := node(path)
	property := getKeyProperty(current)
	if isComputedKey(current) {
		return concatIn(path, "[", print(property, nil), "]")
	}
	if shouldQuoteKey(path, options) {
		// a -> "a"
		// 1 -> "1"
		// 1.5 -> "1.5"
		key := getKey(current)
		name := key.String("name")
		if !key.Is("Identifier") {
			name = javaScriptNumberString(key.Get("value").(float64))
		}
		printed := printString(jsonStringify(name), options)
		return call(path, func(path *Path) Doc { return printing.PrintComments(path, concatIn(path, printed), options, nil) }, property)
	}
	if shouldUnquoteKey(path, options) {
		value := getKey(current).String("value")
		printed := value
		if leadingDigit.MatchString(value) {
			printed = printNumber(value)
		}
		return call(path, func(path *Path) Doc { return printing.PrintComments(path, concatIn(path, printed), options, nil) }, property)
	}
	return print(property, nil)
}

// isEs5IdentifierName is the is-es5-identifier-name package upstream uses: an IdentifierName under
// ES5's grammar. ES5 draws its letters from Unicode 3.0; this uses Go's current Unicode tables, which
// can differ only for characters added since, none of which our keys use.
func isEs5IdentifierName(value string) bool {
	if value == "" {
		return false
	}
	for index, character := range value {
		isStart := character == '$' || character == '_' || unicode.IsLetter(character) || unicode.Is(unicode.Nl, character)
		if index == 0 {
			if !isStart {
				return false
			}
			continue
		}
		if !isStart && !unicode.IsDigit(character) && !unicode.Is(unicode.Mn, character) &&
			!unicode.Is(unicode.Mc, character) && !unicode.Is(unicode.Pc, character) {
			return false
		}
	}
	return true
}
