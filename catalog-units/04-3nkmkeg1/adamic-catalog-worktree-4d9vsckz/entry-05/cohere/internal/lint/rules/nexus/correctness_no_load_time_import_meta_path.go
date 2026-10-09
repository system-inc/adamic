package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// correctnessNoLoadTimeImportMetaPathText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-load-time-import-meta-path.json`.
var correctnessNoLoadTimeImportMetaPathText = policy.MessageOf("nexus/correctness-no-load-time-import-meta-path", "loadTimeImportMetaPath")

func messageLoadTimeImportMetaPath(name string) rule.Message {
	return rule.Message{
		Id:          "loadTimeImportMetaPath",
		Description: correctnessNoLoadTimeImportMetaPathText.Render(map[string]string{"name": name}),
	}
}

// CorrectnessNoLoadTimeImportMetaPath reports `import.meta.dirname` and `import.meta.filename` read
// while the module loads: in a top-level statement, a module-scope initializer, a top-level call's
// arguments, a static class member, a decorator, or a function invoked on the spot.
//
// Node defines both, and a Next server bundle does not. Turbopack hands the module a synthesized
// `import.meta` with only `url` and `turbopackHot`, so both read undefined there. ahra's
// RunSuites.ts joined `import.meta.dirname` at module scope; Next's server bundle reached it through
// the commit command line interface, and `NodePath.join(undefined, ...)` threw ERR_INVALID_ARG_TYPE
// while loading the instrumentation hook, which took Kirk's dev server down at startup (ahra
// 36b7fe07). A read inside a function runs only when called, by which point the caller is in Node
// or never comes, so the repair is to resolve the path when it is needed.
//
// A bare `const directory = import.meta.dirname` is reported too. It does not throw at load; it
// becomes undefined and throws later, wherever it is joined, which is a worse place to find out.
//
// One shape is exempt, exactly: the main-module guard, `import.meta.filename` compared for equality
// with `process.argv[N]` or with a call taking it (`NodePath.resolve(process.argv[1])`). In a bundle
// the comparison is against undefined and is simply false, which is what the guard should say there.
//
// No fix. Moving a path into a function means choosing the function and updating every reader of the
// constant, which is a judgment one expression cannot make.
var CorrectnessNoLoadTimeImportMetaPath = rule.Rule{
	Name: "nexus/correctness-no-load-time-import-meta-path",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindMetaProperty: func(node *ast.Node) {
				// The cheap structural gate first: `import.meta`, read for one of the two names.
				if node.AsMetaProperty().KeywordToken != ast.KindImportKeyword {
					return
				}
				access := correctnessNoLoadTimeImportMetaPathOuterParentheses(node).Parent
				if access == nil || correctnessNoLoadTimeImportMetaPathInnerParentheses(correctnessNoLoadTimeImportMetaPathAccessObject(access)) != node {
					return
				}
				name, ok := property.AccessedName(access, property.Textual|property.Templated)
				if !ok || (name != "dirname" && name != "filename") {
					return
				}
				if name == "filename" && correctnessNoLoadTimeImportMetaPathIsMainModuleGuard(access) {
					return
				}
				if !correctnessNoLoadTimeImportMetaPathRunsAtLoad(access) {
					return
				}
				ctx.ReportNode(access, messageLoadTimeImportMetaPath(name))
			},
		}
	},
}

// correctnessNoLoadTimeImportMetaPathAccessObject is the object of a property or element access, or nil for any other node.
func correctnessNoLoadTimeImportMetaPathAccessObject(node *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	}
	return nil
}

// correctnessNoLoadTimeImportMetaPathOuterParentheses climbs past parentheses wrapping node, so
// `(import.meta).dirname` reaches its access.
func correctnessNoLoadTimeImportMetaPathOuterParentheses(node *ast.Node) *ast.Node {
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node
}

// correctnessNoLoadTimeImportMetaPathInnerParentheses is ast.SkipParentheses with a nil guard, since
// correctnessNoLoadTimeImportMetaPathAccessObject answers nil for a parent that is not an access.
func correctnessNoLoadTimeImportMetaPathInnerParentheses(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipParentheses(node)
}

// correctnessNoLoadTimeImportMetaPathRunsAtLoad answers whether the expression is evaluated while
// its module loads, by climbing to the first construct that defers it.
//
// Only a function body or a parameter defers, and only when the function is not invoked on the spot.
// Everything else on the way up runs when the module or the class is defined: a static member, a
// static block, a computed member name, a decorator, a namespace body. The climb tracks which child
// it arrived from, because a method's body is deferred and the same method's computed name is not.
func correctnessNoLoadTimeImportMetaPathRunsAtLoad(expression *ast.Node) bool {
	child := expression
	for parent := expression.Parent; parent != nil; child, parent = parent, parent.Parent {
		switch parent.Kind {
		case ast.KindDecorator:
			// A decorator runs when its class is defined, whatever it decorates: a class, a member,
			// or a parameter. Resume the climb at that class, so a method's body does not count as
			// deferring a decorator on one of its parameters.
			for parent.Parent != nil && !ast.IsClassLike(parent.Parent) {
				parent = parent.Parent
			}

		case ast.KindPropertyDeclaration:
			// An instance field's initializer runs at construction. A static field, and any field's
			// computed name, runs when the class is defined.
			if !ast.HasStaticModifier(parent) && child == parent.Initializer() {
				return false
			}

		case ast.KindFunctionDeclaration, ast.KindMethodDeclaration, ast.KindConstructor,
			ast.KindGetAccessor, ast.KindSetAccessor:
			if correctnessNoLoadTimeImportMetaPathIsDeferredPart(parent, child) {
				return false
			}

		case ast.KindFunctionExpression, ast.KindArrowFunction:
			if correctnessNoLoadTimeImportMetaPathIsDeferredPart(parent, child) &&
				!correctnessNoLoadTimeImportMetaPathIsInvokedOnTheSpot(parent) {
				return false
			}
		}
	}
	return true
}

// correctnessNoLoadTimeImportMetaPathIsDeferredPart answers whether child is the part of function
// that runs when it is called: its body, or one of its parameters, whose default is evaluated per
// call.
func correctnessNoLoadTimeImportMetaPathIsDeferredPart(function *ast.Node, child *ast.Node) bool {
	if child == function.Body() {
		return true
	}
	return child.Kind == ast.KindParameter && child.Parent == function
}

// correctnessNoLoadTimeImportMetaPathIsInvokedOnTheSpot answers whether a function expression is
// the callee of a call, `(() => import.meta.dirname)()`, whose body then runs wherever the call does.
func correctnessNoLoadTimeImportMetaPathIsInvokedOnTheSpot(function *ast.Node) bool {
	outer := correctnessNoLoadTimeImportMetaPathOuterParentheses(function)
	call := outer.Parent
	return call != nil && call.Kind == ast.KindCallExpression && call.AsCallExpression().Expression == outer
}

// correctnessNoLoadTimeImportMetaPathIsMainModuleGuard answers whether `import.meta.filename` is one
// side of an equality whose other side is `process.argv[N]`, or a call taking it as an argument.
func correctnessNoLoadTimeImportMetaPathIsMainModuleGuard(access *ast.Node) bool {
	operand := correctnessNoLoadTimeImportMetaPathOuterParentheses(access)
	comparison := operand.Parent
	if comparison == nil || comparison.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := comparison.AsBinaryExpression()
	switch binary.OperatorToken.Kind {
	case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken,
		ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken:
	default:
		return false
	}
	other := binary.Left
	if other == operand {
		other = binary.Right
	}
	other = ast.SkipParentheses(other)
	if correctnessNoLoadTimeImportMetaPathIsProcessArgument(other) {
		return true
	}
	if other.Kind != ast.KindCallExpression {
		return false
	}
	for _, argument := range other.Arguments() {
		if correctnessNoLoadTimeImportMetaPathIsProcessArgument(ast.SkipParentheses(argument)) {
			return true
		}
	}
	return false
}

// correctnessNoLoadTimeImportMetaPathIsProcessArgument answers whether node is `process.argv[N]`.
func correctnessNoLoadTimeImportMetaPathIsProcessArgument(node *ast.Node) bool {
	if node.Kind != ast.KindElementAccessExpression {
		return false
	}
	arguments := ast.SkipParentheses(node.AsElementAccessExpression().Expression)
	if name, ok := property.AccessedName(arguments, property.Named); !ok || name != "argv" {
		return false
	}
	process := ast.SkipParentheses(correctnessNoLoadTimeImportMetaPathAccessObject(arguments))
	return process.Kind == ast.KindIdentifier && process.Text() == "process"
}
