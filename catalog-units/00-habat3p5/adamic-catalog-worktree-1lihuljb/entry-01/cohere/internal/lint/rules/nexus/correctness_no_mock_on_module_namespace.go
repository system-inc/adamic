package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoMockOnModuleNamespaceId = "mockOnModuleNamespace"

// correctnessNoMockOnModuleNamespaceText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-mock-on-module-namespace.json`.
var correctnessNoMockOnModuleNamespaceText = policy.MessageOf("nexus/correctness-no-mock-on-module-namespace", correctnessNoMockOnModuleNamespaceId)

func correctnessNoMockOnModuleNamespaceMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoMockOnModuleNamespaceId,
		Description: correctnessNoMockOnModuleNamespaceText.Render(nil),
	}
}

// CorrectnessNoMockOnModuleNamespace reports node:test's `mock.method`, `mock.getter`, `mock.setter`
// or `mock.property` called with an `import * as` namespace object as the object to patch, in a file
// that runs as an ES module.
//
//	invalid: import * as NodeChildProcess from 'node:child_process'; test.mock.method(NodeChildProcess, 'execFileSync', fake);
//	invalid: import * as Store from './Store.ts'; t.mock.method(Store, 'load');
//	invalid: import * as NodeOs from 'node:os'; mock.property(NodeOs, 'EOL', '\r\n');
//	valid:   const childProcess = NodeModule.createRequire(import.meta.url)('node:child_process') as typeof NodeChildProcess; test.mock.method(childProcess, 'execFileSync', fake);
//	valid:   import * as NodeChildProcess from 'node:child_process'; test.mock.method(NodeChildProcess.default, 'execFileSync', fake);
//	valid:   import NodeChildProcess from 'node:child_process'; test.mock.method(NodeChildProcess, 'execFileSync', fake);
//	valid:   import * as Nonce from './Nonce'; jest.spyOn(Nonce, 'createNonce');
//
// # Where it came from
//
// `modules/pensieve/PensieveTransportSecurity.test.ts` in ahra, before `d75fe890`: both transport
// security tests called `test.mock.method(NodeChildProcess, 'execFileSync', ...)` on
// `import * as NodeChildProcess from 'node:child_process'`, and both threw before reaching an
// assertion. The fix mocks the CommonJS exports object behind the namespace, reached through
// `createRequire`, and calls `syncBuiltinESMExports` to push the mock into the live binding. Found by
// the own-history pass of the new-rules sweep (`#tevhg3f`, item C), built in task `#j03vwm6`.
//
// # Why it always throws
//
// A module namespace exotic object's properties are writable data properties that are not
// configurable, and its [[DefineOwnProperty]] refuses any descriptor that changes the value. node's
// `MockTracker.method` (read from `lib/internal/test_runner/mock/mock.js`, v24) finds the property's
// descriptor and installs the mock with `ObjectDefineProperty(object, methodName, { value: mock, ...
// })`, which throws `TypeError: Cannot redefine property`. `getter` and `setter` delegate to
// `method` and read `descriptor.get` or `descriptor.set`, which a namespace's data property does not
// have, so they throw `must be a method` first. `property` replaces the property with an accessor,
// which the namespace refuses the same way. There is no argument, option or export that gets past
// it, which is what makes the report exact rather than likely.
//
// `nexus/import-require-node-namespace` makes `import * as NodeX from 'node:x'` the house spelling for
// every builtin, so the namespace is the first object a test author reaches for.
//
// # What it matches, by declaration
//
//   - **The mock call**: the call's resolved signature is a method named `method`, `getter`, `setter`
//     or `property` declared on `MockTracker` inside `declare module "node:test"` (or `"test"`), the
//     declaration `@types/node` gives `mock`, `test.mock` and a test context's `t.mock`. A `mock`
//     from anywhere else, jest's `spyOn` among them, is not read: under jest's CommonJS transform an
//     `import * as` is a plain copied object and spying on it works.
//   - **The namespace**: the first argument, through parentheses, `!`, `as` and `satisfies` (which
//     change the type and never the object), is an identifier whose symbol is declared by a
//     namespace import, `import * as Name`, that is not `import type`. A default import, a named
//     import, `import Name = require(...)`, a member of the namespace (`Name.default`, the CommonJS
//     exports of a builtin) and a value merely typed `typeof Name` are all other objects, and the
//     last is exactly the fix. A namespace copied into a local first, or reached through
//     `await import(...)`, is a missed finding: following values is not what this rule claims.
//
// # Only where the file runs as an ES module
//
// The namespace is sealed only when the import stays an import at runtime. `.cts` and `.cjs` files
// always run as CommonJS, where a compiler without interop helpers turns `import * as` into a plain
// `require` whose exports are writable, so they are never reported. Other files are reported when
// the program emits ES modules: `.mts` and `.mjs` always, and the rest when the `module` option is
// an ES version or `preserve`. Under `node16`/`nodenext` a `.ts` file's format comes from the
// nearest `package.json`, which a rule cannot read without leaving the findings cache, so those are
// declined, a missed finding in a tree configured that way. ahra sets `"module": "esnext"`.
//
// # No fix
//
// The repair depends on what the module is: a builtin's CommonJS exports through `createRequire`, a
// seam the code under test takes as a parameter, or `mock.module`. A fixer would have to pick one.
var CorrectnessNoMockOnModuleNamespace = rule.Rule{
	Name: "nexus/correctness-no-mock-on-module-namespace",

	// The mock call is told from any other `method` by its resolved declaration, and the namespace
	// from a value of the same type by the declaration its symbol points at.
	NeedsTypeChecker: true,

	// The `module` option decides whether the file runs as an ES module.
	ProgramReads: rule.ReadsCompilerOptions,

	// The rule reads `node:test`'s declaration of MockTracker, never a body, and the namespace import
	// it resolves the argument to is declared in the file itself.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil || ctx.Program == nil {
			// Telling node:test's MockTracker from any other `method`, and a namespace import from a
			// value of the same type, are questions for the checker.
			return nil
		}
		if !correctnessNoMockOnModuleNamespaceRunsAsModule(ctx.SourceFile.FileName().AsString(), ctx.Program.Options()) {
			return nil
		}
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				callee := ast.SkipParentheses(call.Expression)
				if callee.Kind != ast.KindPropertyAccessExpression {
					return
				}
				name := callee.AsPropertyAccessExpression().Name()
				if name.Kind != ast.KindIdentifier || !correctnessNoMockOnModuleNamespaceMockMethods[name.Text()] {
					return
				}
				target := correctnessNoMockOnModuleNamespaceUnwrap(call.Arguments.Nodes[0])
				if target.Kind != ast.KindIdentifier || !correctnessNoMockOnModuleNamespaceIsNamespace(ctx, target) {
					return
				}
				if !correctnessNoMockOnModuleNamespaceIsMockTracker(ctx, node) {
					return
				}
				ctx.ReportNode(node, correctnessNoMockOnModuleNamespaceMessage())
			},
		}
	},
}

// correctnessNoMockOnModuleNamespaceMockMethods are the MockTracker methods that redefine a property
// of the object they are given.
var correctnessNoMockOnModuleNamespaceMockMethods = map[string]bool{
	"method": true, "getter": true, "setter": true, "property": true,
}

// correctnessNoMockOnModuleNamespaceRunsAsModule says whether a file runs as an ES module: always for
// `.mts` and `.mjs`, never for `.cts` and `.cjs`, and for the rest when the program emits ES modules.
func correctnessNoMockOnModuleNamespaceRunsAsModule(fileName string, options *core.CompilerOptions) bool {
	switch {
	case strings.HasSuffix(fileName, ".mts"), strings.HasSuffix(fileName, ".mjs"):
		return true
	case strings.HasSuffix(fileName, ".cts"), strings.HasSuffix(fileName, ".cjs"):
		return false
	}
	if options == nil {
		return false
	}
	kind := options.GetEmitModuleKind()
	return kind.IsNonNodeESM() || kind == core.ModuleKindPreserve
}

// correctnessNoMockOnModuleNamespaceUnwrap looks through the wrappers that change an expression's type
// and never its value.
func correctnessNoMockOnModuleNamespaceUnwrap(expression *ast.Node) *ast.Node {
	for {
		switch expression.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression, ast.KindAsExpression,
			ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression:
			expression = expression.Expression()
			continue
		}
		return expression
	}
}

// correctnessNoMockOnModuleNamespaceIsNamespace says whether an identifier names the binding of an
// `import * as` declaration in this file that is not type-only.
func correctnessNoMockOnModuleNamespaceIsNamespace(ctx rule.Context, identifier *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsAlias == 0 {
		return false
	}
	declarations := rule.DeclarationsIn(ctx.SourceFile, symbol)
	if len(declarations) != 1 {
		return false
	}
	declaration := declarations[0]
	return declaration.Kind == ast.KindNamespaceImport && !ast.IsTypeOnlyImportOrExportDeclaration(declaration)
}

// correctnessNoMockOnModuleNamespaceIsMockTracker says whether a call resolves to a method declared on
// `MockTracker` in `declare module "node:test"` (or `"test"`), at any depth of namespaces inside it.
func correctnessNoMockOnModuleNamespaceIsMockTracker(ctx rule.Context, call *ast.Node) bool {
	signature := ctx.TypeChecker.GetResolvedSignature(call)
	if signature == nil {
		return false
	}
	declaration := signature.Declaration()
	if declaration == nil || (declaration.Kind != ast.KindMethodSignature && declaration.Kind != ast.KindMethodDeclaration) {
		return false
	}
	owner := declaration.Parent
	if owner == nil || (owner.Kind != ast.KindInterfaceDeclaration && owner.Kind != ast.KindClassDeclaration) ||
		owner.Name() == nil || owner.Name().Text() != "MockTracker" {
		return false
	}
	for current := owner.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindSourceFile {
			return false
		}
		if current.Kind != ast.KindModuleDeclaration || current.Name() == nil {
			continue
		}
		if current.Name().Kind == ast.KindStringLiteral {
			switch current.Name().Text() {
			case "node:test", "test":
				return true
			}
			return false
		}
	}
	return false
}
