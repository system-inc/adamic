package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageGlobalNonLexicalBinding = rule.Message{
	Id: "globalNonLexicalBinding",
	Description: "This declaration lands in the global scope, where every other script on the page " +
		"can see it and any of them can replace it. Wrap it in an immediately invoked function for " +
		"a local, or assign it as a property of the global object if it is meant to be shared.",
}

var messageGlobalLexicalBinding = rule.Message{
	Id: "globalLexicalBinding",
	Description: "This declaration lands in the global lexical scope, which is shared across every " +
		"script on the page, so a second script declaring the same name is a redeclaration error " +
		"rather than a shadow. Wrap it in a block or in an immediately invoked function.",
}

var messageGlobalVariableLeak = rule.Message{
	Id: "globalVariableLeak",
	Description: "This assigns to a name that was never declared, which creates a global at run " +
		"time rather than the local the code reads as. Declare the variable if it is meant to be " +
		"local, which is almost always what a bare assignment like this intended.",
}

// NoImplicitGlobalsSettings is the decoded option surface.
//
// Upstream's schema is one object with a single `lexicalBindings` boolean, defaulting to FALSE, so
// the zero value happens to be right here and the usual default-inversion hazard does not apply.
// It is still decoded through a pointer so that an explicit `false` and an absent key stay
// distinguishable, which keeps the shape correct if the default ever moves.
type NoImplicitGlobalsSettings struct {
	// LexicalBindings turns on the `let`, `const` and `class` half of the rule.
	LexicalBindings bool
}

// DefaultNoImplicitGlobalsSettings is upstream's `lexicalBindings: false`.
func DefaultNoImplicitGlobalsSettings() NoImplicitGlobalsSettings {
	return NoImplicitGlobalsSettings{LexicalBindings: false}
}

type noImplicitGlobalsWireShape struct {
	LexicalBindings *bool `json:"lexicalBindings"`
}

// DecodeNoImplicitGlobalsOptions reads the option object off the config.
func DecodeNoImplicitGlobalsOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return DefaultNoImplicitGlobalsSettings(), nil
	}

	var wire noImplicitGlobalsWireShape
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return DefaultNoImplicitGlobalsSettings(), err
	}

	settings := DefaultNoImplicitGlobalsSettings()
	if wire.LexicalBindings != nil {
		settings.LexicalBindings = *wire.LexicalBindings
	}
	return settings, nil
}

// NoImplicitGlobals flags a declaration that lands in the global scope, and an assignment that
// creates a global because nothing declared the name.
//
//	valid:   (function() { var foo = 1; })();
//	valid:   { let a = 1; }
//	valid:   var foo = 1;                    in a module, which has no global scope
//	invalid: var foo = 1;                    in a script
//	invalid: function foo() {}               in a script
//	invalid: foo = 1;                        in sloppy code, since nothing declared foo
//
// # It only fires in a SCRIPT, and that is the whole reason it measures zero here
//
// A module has its own scope, so nothing declared at its top level is global and the first two
// judgments cannot apply. Measured against the installed rule: every one of `var a = 1`,
// `function f(){}`, `let a = 1`, `const a = 1`, `class C {}`, `a = 1` and `for (b in c) {}` is
// CLEAN under `sourceType: "module"` and reports under `"script"`.
//
// This tree is essentially all modules. A scan of `app`, `modules` and the structure libraries found
// 3174 files carrying a top level import or export against 2 that do not, which is why the audit
// measured zero violations. That zero is a real property of the tree rather than a blind rule, and
// the fixtures below prove the rule can fire: the harness builds a file with no import or export as
// a SCRIPT, verified with two module controls.
//
// # Three of the five messages are ported and two are declined, for a reason already settled here
//
// Upstream also reports `assignmentToReadonlyGlobal` and `redeclarationOfReadonlyGlobal`, which
// require knowing that a name is a read-only global. Both of its sources are eslint configuration
// surfaces rather than anything in the tree: the `globals` config key, and the `/*global foo:readonly*/`
// comment directive. Under typescript-eslint's parser the TypeScript lib's types are read-only globals
// as well, which `@typescript-eslint/no-redeclare`'s `builtinGlobals` reads since #e1zk9s0; these two
// messages were not revisited there.
//
// The decline is pinned by `TestNoImplicitGlobalsDeclinesReadonlyGlobals` rather than left as
// silence, and it costs 3 of upstream's failing cases outright plus every case carrying a directive.
//
// The `/*exported foo*/` directive is declined on the same grounds and costs 42 more cases. Of
// upstream's 245 cases, 110 are expressible here: 75 clean and 35 reporting.
//
// # The leak half asks the checker, and that is the port rather than an addition
//
// Upstream reads `scope.implicit.variables`, which is eslint-scope's record of names assigned but
// never declared. We ask resolution the same question: an assignment target that resolves to no
// symbol was never declared. Probed with a control before being built on -- `foo = 1` answers no
// symbol, `var foo; foo = 1` answers a symbol with one declaration.
//
// This half is not gated on the file being a script, since a leak escapes from inside a function as
// well: `window.foo = function() { bar = 1; }` is one of upstream's failing cases. It is gated on the
// code being sloppy, because in strict code an assignment to an undeclared name throws instead of
// creating a global (see liesInStrictModeCode). A module is strict throughout, and so is every
// TypeScript file under `alwaysStrict`, so in a TypeScript program the leak half fires only in a
// sloppy JavaScript script.
var NoImplicitGlobals = rule.Rule{
	Name: "no-implicit-globals",

	// The leak half asks resolution whether an assignment target was ever declared, which is a
	// question about the whole program rather than about this file's syntax.
	NeedsTypeChecker: true,
	// Asks only whether a name resolves to any symbol, never reading a declaration, and no function
	// body can declare a name another file sees, so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,
	// `alwaysStrict` decides whether a TypeScript script can leak, so the cached verdict has to
	// follow the tsconfig.
	ProgramReads: rule.ReadsCompilerOptions,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, ok := rule.OptionsAs[NoImplicitGlobalsSettings](options)
		if !ok {
			settings = DefaultNoImplicitGlobalsSettings()
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				file := node.AsSourceFile()
				if file == nil {
					return
				}
				// A module has no global scope, so the declaration halves do not apply to it. A
				// script's top-level var or function is a global whether or not the script is
				// strict, which is why this half reads module detection and nothing else (#rbvd7sv).
				if file.ExternalModuleIndicator == nil {
					reportGlobalDeclarations(ctx, node, settings.LexicalBindings)
				}
				// The leak half is the only one that asks the checker, so it is guarded separately
				// rather than at the top, and the declaration halves still run in the untyped
				// harness, which parses a script with no checker.
				if ctx.TypeChecker != nil {
					reportGlobalVariableLeaks(ctx, node)
				}
			},
		}
	},
}

// reportGlobalDeclarations reports every top-level declaration that lands in the global scope.
//
// Only the file's own statements are considered, deliberately: a declaration inside a function or a
// block is not global, which is exactly what upstream tells the reader to do about one. So this does
// not descend, and the absence of a recursive walk here is the rule rather than an omission.
func reportGlobalDeclarations(ctx rule.Context, node *ast.Node, lexicalBindings bool) {
	for _, statement := range node.AsSourceFile().Statements.Nodes {
		switch statement.Kind {
		case ast.KindFunctionDeclaration:
			// A function EXPRESSION assigned to a global var is reported through the var rather
			// than here, which is why upstream's `var foo = function foo() {};` is one finding and
			// not two: the inner name is scoped to the expression itself.
			name := statement.Name()
			if name != nil {
				ctx.ReportNode(statement, messageGlobalNonLexicalBinding)
			}

		case ast.KindClassDeclaration:
			if lexicalBindings && statement.Name() != nil {
				ctx.ReportNode(statement, messageGlobalLexicalBinding)
			}

		case ast.KindVariableStatement:
			list := statement.AsVariableStatement().DeclarationList
			if list == nil {
				break
			}
			declarationList := list.AsVariableDeclarationList()
			if declarationList == nil {
				break
			}
			// `var` is the non-lexical binding; `let` and `const` are lexical and gated on the
			// option. The distinction is not cosmetic: a global `var` becomes a property of the
			// global object while a global `let` does not, which is why upstream words the two
			// messages differently and offers different advice.
			isLexical := declarationList.Flags&(ast.NodeFlagsLet|ast.NodeFlagsConst) != 0
			if isLexical && !lexicalBindings {
				break
			}
			message := messageGlobalNonLexicalBinding
			if isLexical {
				message = messageGlobalLexicalBinding
			}
			for _, declaration := range declarationList.Declarations.Nodes {
				// A destructuring pattern binds several names and upstream reports each one, so
				// `const [a, b, ...c] = [];` is three findings rather than one.
				//
				// A plain identifier reports on the whole DECLARATOR rather than on the name,
				// because upstream reports `def.node`, which is the VariableDeclarator. Measured
				// against the installed rule: `var foo = 1;` reports columns 5 to 12, which is
				// `foo = 1` and not `foo`. Every message-id fixture is satisfied by either reading,
				// so the span assertion is the only thing that records which one was ported.
				if name := declaration.Name(); name != nil && name.Kind == ast.KindIdentifier {
					ctx.ReportNode(declaration, message)
					continue
				}
				reportEachBoundName(ctx, declaration.Name(), message)
			}
		}
	}
}

// reportEachBoundName reports once per name a binding introduces.
//
// A plain identifier is one name. A pattern is as many as it binds, counting a rest element and
// recursing through a nested pattern, which is what makes upstream's
// `let { a, foo: b, bar: { c } } = {};` three findings.
func reportEachBoundName(ctx rule.Context, name *ast.Node, message rule.Message) {
	if name == nil {
		return
	}
	switch name.Kind {
	case ast.KindIdentifier:
		ctx.ReportNode(name, message)

	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		for _, element := range name.AsBindingPattern().Elements.Nodes {
			if element.Kind != ast.KindBindingElement {
				continue
			}
			reportEachBoundName(ctx, element.AsBindingElement().Name(), message)
		}
	}
}

// reportGlobalVariableLeaks reports an assignment to a name nothing ever declared.
//
// This is upstream's `scope.implicit.variables`, asked of resolution instead: a target resolving to
// no symbol was never declared anywhere, so assigning to it in sloppy code creates a global at run
// time. It is not gated on the file being a script, because a leak escapes from a function too, and
// upstream's `window.foo = function() { bar = 1; }` is the case that says so.
//
// The three assignment forms are upstream's own set: an assignment expression, and the left side of
// a `for...in` or `for...of`.
func reportGlobalVariableLeaks(ctx rule.Context, node *ast.Node) {
	var visit func(*ast.Node) bool
	visit = func(current *ast.Node) bool {
		switch current.Kind {
		case ast.KindBinaryExpression:
			binary := current.AsBinaryExpression()
			if binary != nil && binary.OperatorToken != nil &&
				binary.OperatorToken.Kind == ast.KindEqualsToken &&
				!liesInStrictModeCode(ctx, current) {
				reportIfTargetIsUndeclared(ctx, binary.Left)
			}

		case ast.KindForInStatement, ast.KindForOfStatement:
			// Only a bare target leaks. A `for (var x in y)` declares, and that declaration is
			// judged by the declaration half if it is global.
			//
			// This kind test is EQUIVALENT rather than load bearing, and it is kept because it
			// states the intent at the point a reader looks for it. A declaration list matches no
			// arm of reportIfTargetIsUndeclared's switch, so dropping the test reaches the same
			// silence by falling through instead of by declining. Measured over seven for-in and
			// for-of shapes including var, let, const, a bare target, a bare pattern, a declared
			// pattern and an initialized declaration: byte identical findings under both spellings.
			initializer := current.AsForInOrOfStatement().Initializer
			if initializer != nil && initializer.Kind != ast.KindVariableDeclarationList &&
				!liesInStrictModeCode(ctx, current) {
				reportIfTargetIsUndeclared(ctx, initializer)
			}
		}

		current.ForEachChild(visit)
		return false
	}
	node.ForEachChild(visit)
}

// reportIfTargetIsUndeclared reports an assignment target that resolves to nothing.
//
// A destructuring target binds several names and each one is asked separately, which is what makes
// upstream's `[foo, bar] = [];` two findings.
func reportIfTargetIsUndeclared(ctx rule.Context, target *ast.Node) {
	if target == nil {
		return
	}
	switch target.Kind {
	case ast.KindIdentifier:
		// No symbol means no declaration anywhere, which is exactly the implicit global. A symbol
		// carrying declarations means the name was declared and this is an ordinary write.
		if symbol := ctx.TypeChecker.GetSymbolAtLocation(target); symbol == nil {
			ctx.ReportNode(target, messageGlobalVariableLeak)
		}

	case ast.KindArrayLiteralExpression:
		for _, element := range target.AsArrayLiteralExpression().Elements.Nodes {
			reportIfTargetIsUndeclared(ctx, element)
		}

	case ast.KindObjectLiteralExpression:
		for _, property := range target.AsObjectLiteralExpression().Properties.Nodes {
			switch property.Kind {
			case ast.KindShorthandPropertyAssignment:
				reportIfTargetIsUndeclared(ctx, property.Name())
			case ast.KindPropertyAssignment:
				reportIfTargetIsUndeclared(ctx, property.AsPropertyAssignment().Initializer)
			}
		}

	case ast.KindSpreadElement:
		reportIfTargetIsUndeclared(ctx, target.AsSpreadElement().Expression)
	}
}

// liesInStrictModeCode says whether an assignment sits somewhere strict mode is in force.
//
// In strict mode an assignment to an undeclared name is a ReferenceError at run time rather than a
// new global, so there is no leak to report and upstream stays silent. This is not defensive coding
// and it is not visible from upstream's rule body at all: it falls out of eslint-scope refusing to
// record an implicit global in a strict scope. Measured against the installed rule with a control:
//
//	foo = 1;                                     globalVariableLeak
//	'use strict';foo = 1;                        clean
//	(function() {'use strict'; foo = 1; })();    clean
//	{ class Foo { constructor() { bar = 1; } } } clean, a class body is always strict
//
// Four sources of strictness are checked. A `use strict` directive at the top of the file or of any
// enclosing function body. A class, whose body is strict by specification with no directive written
// anywhere: the source a port is most likely to miss, and it costs a false positive on ordinary code
// rather than a missed finding. A module, which is strict throughout. And the compiler: under
// `alwaysStrict` every TypeScript file is bound and emitted as strict, which is ESLint's
// `ecmaFeatures.impliedStrict`, where eslint-scope likewise records no implicit global (#hks3djf).
//
// The last two settled #jjfa7qb's 34 leak rows. The port this replaced followed oxc, which reports a
// leak in a module, and ran ESLint's corpus as modules where ESLint reports nothing. Under the authority
// rule, and on first principles, an assignment that throws leaks nothing.
func liesInStrictModeCode(ctx rule.Context, node *ast.Node) bool {
	for ancestor := node; ancestor != nil; ancestor = ancestor.Parent {
		// A class body is strict by specification, with no directive to find.
		if ast.IsClassLike(ancestor) {
			return true
		}

		var statements []*ast.Node
		switch {
		case ast.IsSourceFile(ancestor):
			file := ancestor.AsSourceFile()
			if file.ExternalModuleIndicator != nil || type_checking.CompilerImpliesStrict(ctx, file) {
				return true
			}
			statements = file.Statements.Nodes
		case ast.IsBlock(ancestor) && ancestor.Parent != nil &&
			ast.IsFunctionLikeDeclaration(ancestor.Parent):
			statements = ancestor.AsBlock().Statements.Nodes
		default:
			continue
		}
		if hasUseStrictDirective(statements) {
			return true
		}
	}
	return false
}

// hasUseStrictDirective says whether a directive prologue opens with `use strict`.
//
// A directive is a bare string expression statement at the very top of a program or a function body,
// and the prologue ends at the first statement that is not one, which is why this stops rather than
// scanning the whole list.
func hasUseStrictDirective(statements []*ast.Node) bool {
	for _, statement := range statements {
		if !ast.IsExpressionStatement(statement) {
			return false
		}
		expression := statement.AsExpressionStatement().Expression
		if expression == nil || expression.Kind != ast.KindStringLiteral {
			return false
		}
		if expression.Text() == "use strict" {
			return true
		}
	}
	return false
}
