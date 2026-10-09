package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/text"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// The rule's messages, one handle per id, whose wording lives in
// `policy/messages/consistency-no-multiline-arrow-function.json`.
var (
	consistencyNoMultilineArrowFunctionReactHookArrowText        = policy.MessageOf("nexus/consistency-no-multiline-arrow-function", "reactHookArrow")
	consistencyNoMultilineArrowFunctionAddEventListenerArrowText = policy.MessageOf("nexus/consistency-no-multiline-arrow-function", "addEventListenerArrow")
	consistencyNoMultilineArrowFunctionMultilineArrowText        = policy.MessageOf("nexus/consistency-no-multiline-arrow-function", "multilineArrow")
)

func messageReactHookArrow(hookName string) rule.Message {
	return rule.Message{
		Id:          "reactHookArrow",
		Description: consistencyNoMultilineArrowFunctionReactHookArrowText.Render(map[string]string{"hookName": hookName}),
	}
}

func messageAddEventListenerArrow() rule.Message {
	return rule.Message{
		Id:          "addEventListenerArrow",
		Description: consistencyNoMultilineArrowFunctionAddEventListenerArrowText.Render(nil),
	}
}

func messageMultilineArrow() rule.Message {
	return rule.Message{
		Id:          "multilineArrow",
		Description: consistencyNoMultilineArrowFunctionMultilineArrowText.Render(nil),
	}
}

// ConsistencyNoMultilineArrowFunction steers multi-line arrows to function declarations.
//
//	valid:   const add = (a, b) => a + b
//	valid:   const f = () => {}
//	valid:   element.addEventListener('click', function handleClick() { ... })
//	invalid: const compute = (input) => {
//	             return input * 2;
//	         }
//
// Three exemptions, each of which cost something to learn.
//
// An arrow using `this` is left alone, because a function declaration would rebind it. Finding that
// out requires walking into nested arrows but not into nested functions: a nested function binds its
// own `this` and says nothing about this arrow's, while a nested arrow inherits this one's, so a
// `this` inside it is this arrow's. Treating a nested arrow as its own scope let an outer arrow
// convert and rebound `this` for every nested arrow at once.
//
// A single-line arrow is not a multi-line arrow whatever shape its body has. A block body stood in
// for "spans lines" once, and the two part company at `() => {}`: a no-op callback rewritten to
// `function() {}` for no reading gain. That one was not merely noise, because an arrow has no
// prototype and a function expression does, so a fixture that existed to be non-constructible became
// constructible and its assertion flipped. The rule's name is the honest test, so this measures the
// span.
//
// A React hook or addEventListener argument reports through its own message and must not be reported
// twice, so the general case skips what the call-expression case already claimed.
//
// All three findings carry the original's fix, which rewrites the arrow in place as a function
// expression; `arrowFunctionFix` holds the conversion and the places it declines.
var ConsistencyNoMultilineArrowFunction = rule.Rule{
	Name: "nexus/consistency-no-multiline-arrow-function",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}
		sourceFile := ctx.SourceFile

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil || call.Expression == nil || call.Arguments == nil {
					return
				}
				if call.Expression.Kind != ast.KindPropertyAccessExpression {
					return
				}
				access := call.Expression.AsPropertyAccessExpression()
				if access == nil || access.Expression == nil || access.Name() == nil {
					return
				}
				objectName := ""
				if access.Expression.Kind == ast.KindIdentifier {
					objectName = access.Expression.Text()
				}
				propertyName := access.Name().Text()

				arguments := call.Arguments.Nodes

				if isReactHookCallee(objectName, propertyName) && len(arguments) > 0 &&
					arguments[0].Kind == ast.KindArrowFunction {
					reportArrow(ctx, arguments[0], messageReactHookArrow(propertyName))
					return
				}
				if propertyName == "addEventListener" && len(arguments) > 1 &&
					arguments[1].Kind == ast.KindArrowFunction {
					reportArrow(ctx, arguments[1], messageAddEventListenerArrow())
				}
			},

			ast.KindArrowFunction: func(node *ast.Node) {
				arrow := node.AsArrowFunction()
				if arrow == nil || arrow.Body == nil {
					return
				}

				// An arrow that uses `this` would change meaning as a function declaration.
				if usesThis(arrow.Body) {
					return
				}
				// An implicit return is the shape an arrow is for.
				if arrow.Body.Kind != ast.KindBlock {
					return
				}
				// One line is not multi-line, whatever the body's shape.
				if spansOneLine(sourceFile, node) {
					return
				}
				// Already claimed by the call-expression case above.
				if isClaimedByCallExpressionCase(node) {
					return
				}

				reportArrow(ctx, node, messageMultilineArrow())
			},
		}
	},
}

// isReactHookCallee reports whether a callee is React.forwardRef or a React hook.
func isReactHookCallee(objectName string, propertyName string) bool {
	if objectName != "React" {
		return false
	}
	return propertyName == "forwardRef" || strings.HasPrefix(propertyName, "use")
}

// usesThis reports whether a `this` appears in a body, counting only the ones that would rebind.
//
// A nested function declaration or expression binds its own `this`, so its contents say nothing
// about the enclosing arrow and are not searched. A nested arrow has no `this` of its own and
// inherits the enclosing one, so it is searched.
func usesThis(node *ast.Node) bool {
	found := false
	var walk func(current *ast.Node)
	walk = func(current *ast.Node) {
		if current == nil || found {
			return
		}
		if current.Kind == ast.KindThisKeyword {
			found = true
			return
		}
		if current.Kind == ast.KindFunctionDeclaration || current.Kind == ast.KindFunctionExpression ||
			current.Kind == ast.KindMethodDeclaration || current.Kind == ast.KindClassDeclaration ||
			current.Kind == ast.KindClassExpression {
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	walk(node)
	return found
}

// spansOneLine reports whether a node's own text begins and ends on the same source line.
//
// The span is measured from the token range rather than from Pos(), because Pos() sits before the
// node's leading trivia. An arrow written inside a multi-line array literal is preceded by a newline
// and indentation, so measuring from Pos() puts its start on the previous line and a genuinely
// single-line `() => {}` reads as spanning two. That is the exact case this exemption exists for,
// and it produced the one false finding this rule had on the ahra tree.
func spansOneLine(sourceFile *ast.SourceFile, node *ast.Node) bool {
	tokenRange := rule.TokenRange(sourceFile, node)
	startLine, _ := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, tokenRange.Pos())
	endLine, _ := scanner.GetECMALineAndByteOffsetOfPosition(sourceFile, tokenRange.End())
	return startLine == endLine
}

// isClaimedByCallExpressionCase reports whether an arrow is the argument the call-expression
// listener already reported, so one arrow never produces two findings.
func isClaimedByCallExpressionCase(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	if call == nil || call.Expression == nil || call.Expression.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := call.Expression.AsPropertyAccessExpression()
	if access == nil || access.Name() == nil {
		return false
	}
	objectName := ""
	if access.Expression != nil && access.Expression.Kind == ast.KindIdentifier {
		objectName = access.Expression.Text()
	}
	propertyName := access.Name().Text()
	return isReactHookCallee(objectName, propertyName) || propertyName == "addEventListener"
}

// reportArrow reports an arrow with the conversion attached when the conversion keeps its meaning.
func reportArrow(ctx rule.Context, arrow *ast.Node, message rule.Message) {
	if fix, isSafe := arrowFunctionFix(ctx, arrow); isSafe {
		ctx.ReportNodeWithFixes(arrow, message, fix)
		return
	}
	ctx.ReportNode(arrow, message)
}

// arrowFunctionFix rewrites an arrow as a function expression, the TypeScript original's fix.
//
//	async <T>(value: T): Promise<T> => { ... }   ->   async function<T>(value: T): Promise<T> { ... }
//	value => value * 2                           ->   function(value) { return value * 2; }
//
// The signature (type parameters, parameter list, return type) is copied from the source between
// the arrow's start and its own `=>` token, which the parser hands over directly, so an arrow inside
// a parameter's type cannot be mistaken for this one's. The type parameters come off the front by
// their AST span rather than by counting angle brackets, for the original's reason: a default like
// `<T extends Record<string, () => void>>` nests both. A bare single parameter gains parentheses,
// and an expression body becomes a block that returns it.
//
// # Where this differs from the original, each time in the direction of not breaking code
//
// The original strips `async ` by text and slices the type parameters at an offset measured from the
// node's start, which still counts the `async` it just removed. Both go wrong on shapes the parser
// accepts: `async(value) => {}` keeps its `async(` and becomes `async function async(value)`, and
// `async <T>(value: T) => {}` slices the parameter list mid-token. Here the signature starts after the
// async modifier's own token, so both convert cleanly.
//
// The original converts whatever it reports, and the call-expression findings never asked whether
// the arrow uses `this`, so `React.useCallback(() => { this.x })` became a function with its own
// `this`. A function binds `this`, `arguments` and `new.target` where an arrow inherits them, and
// `super` is not even legal in one, so an arrow reading any of the four, in its body or in a
// parameter default, is still reported and is not rewritten.
//
// An expression body that the parser wraps in parentheses is returned without them, matching the
// original's output: ESTree has no parenthesized node, so `getText(node.body)` never included them.
func arrowFunctionFix(ctx rule.Context, node *ast.Node) (rule.Fix, bool) {
	sourceFile := ctx.SourceFile
	if sourceFile == nil || node == nil || node.Kind != ast.KindArrowFunction {
		return rule.Fix{}, false
	}
	arrow := node.AsArrowFunction()
	if arrow == nil || arrow.Body == nil || arrow.EqualsGreaterThanToken == nil {
		return rule.Fix{}, false
	}
	if readsInheritedBinding(node) {
		return rule.Fix{}, false
	}

	sourceText := sourceFile.Text()
	signatureStart := rule.TokenRange(sourceFile, node).Pos()
	asyncKeyword := ""
	if modifiers := node.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier.Kind == ast.KindAsyncKeyword {
				asyncKeyword = "async "
				signatureStart = scanner.SkipTrivia(sourceText, modifier.End())
			}
		}
	}
	arrowTokenStart := rule.TokenRange(sourceFile, arrow.EqualsGreaterThanToken).Pos()

	generics := ""
	parametersStart := signatureStart
	if arrow.TypeParameters != nil {
		// The list's end already includes a trailing comma, so `<T,>` lands here on the `>` too;
		// the `.tsx` fix vector is what pins that.
		closingAngle := scanner.SkipTrivia(sourceText, arrow.TypeParameters.End())
		if closingAngle >= arrowTokenStart || sourceText[closingAngle] != '>' {
			return rule.Fix{}, false
		}
		generics = text.TrimWhitespace(sourceText[signatureStart : closingAngle+1])
		parametersStart = closingAngle + 1
	}
	if parametersStart > arrowTokenStart {
		return rule.Fix{}, false
	}

	signature := text.TrimWhitespace(sourceText[parametersStart:arrowTokenStart])
	if !strings.HasPrefix(signature, "(") {
		signature = "(" + signature + ")"
	}

	body := arrow.Body
	replacementBody := ""
	if body.Kind == ast.KindBlock {
		bodyRange := rule.TokenRange(sourceFile, body)
		replacementBody = sourceText[bodyRange.Pos():bodyRange.End()]
	} else {
		for body.Kind == ast.KindParenthesizedExpression {
			body = body.AsParenthesizedExpression().Expression
		}
		bodyRange := rule.TokenRange(sourceFile, body)
		replacementBody = "{ return " + sourceText[bodyRange.Pos():bodyRange.End()] + "; }"
	}

	replacement := asyncKeyword + "function" + generics + signature + " " + replacementBody
	// A statement that begins with `function` is a declaration, and a declaration needs a name, so an
	// arrow that leads an expression statement, `() => { ... };`, rewritten bare read as a nameless
	// declaration and did not parse (TS1003, two trpc test files, #kq9vtva). Parenthesized it stays
	// the expression it was.
	if arrowLeadsExpressionStatement(sourceFile, node) {
		replacement = "(" + replacement + ")"
	}
	return ctx.ReplaceNode(node, replacement), true
}

// arrowLeadsExpressionStatement reports whether the arrow's first token is its statement's first token,
// in an expression statement, where a leading `function` would start a declaration instead.
func arrowLeadsExpressionStatement(sourceFile *ast.SourceFile, node *ast.Node) bool {
	start := rule.TokenRange(sourceFile, node).Pos()
	for current := node; current.Parent != nil; current = current.Parent {
		parent := current.Parent
		if rule.TokenRange(sourceFile, parent).Pos() != start {
			return false
		}
		if parent.Kind == ast.KindExpressionStatement {
			return true
		}
	}
	return false
}

// readsInheritedBinding reports whether an arrow reads something it inherits and a function would
// bind for itself: `this`, `arguments`, `new.target`, or `super`.
//
// Searched the way `usesThis` searches, through nested arrows (which inherit the same four) and not
// into nested functions or methods (which bind their own), but over the parameters as well as the
// body, since a default value is evaluated in the same scope. It does descend into a nested class,
// whose field initializers and computed keys can read the outer `this`; that over-refuses a class
// whose own members use `this`, which costs a fix and never a wrong one.
func readsInheritedBinding(arrow *ast.Node) bool {
	found := false
	var walk func(current *ast.Node)
	walk = func(current *ast.Node) {
		if current == nil || found {
			return
		}
		switch current.Kind {
		case ast.KindThisKeyword, ast.KindSuperKeyword:
			found = true
			return
		case ast.KindMetaProperty:
			// `new.target` is bound per function; `import.meta` is the module's and survives.
			if current.AsMetaProperty().KeywordToken == ast.KindNewKeyword {
				found = true
				return
			}
		case ast.KindIdentifier:
			if current.Text() == "arguments" && !isPropertyNamePosition(current) {
				found = true
			}
			return
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindMethodDeclaration,
			ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			return
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return found
		})
	}
	arrow.ForEachChild(func(child *ast.Node) bool {
		walk(child)
		return found
	})
	return found
}

// isPropertyNamePosition reports an identifier that names a property rather than reading a binding:
// the right half of `thing.arguments`, or a key in `{ arguments: 1 }`.
func isPropertyNamePosition(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Name() == identifier
	case ast.KindPropertyAssignment, ast.KindPropertyDeclaration, ast.KindPropertySignature,
		ast.KindMethodDeclaration, ast.KindMethodSignature:
		return parent.Name() == identifier
	}
	return false
}
