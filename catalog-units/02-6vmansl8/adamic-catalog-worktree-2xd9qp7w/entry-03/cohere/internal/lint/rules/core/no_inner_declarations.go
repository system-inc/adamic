package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageNoInnerDeclarationsId is the id. The message is built per finding, because it carries both
// what was declared and where it should have gone, and neither is a constant.
const messageNoInnerDeclarationsId = "moveDeclToRoot"

// noInnerDeclarationsMessage renders the finding.
//
// Upstream's template is `Move {{type}} declaration to {{body}} root.` Both slots are filled from
// the same node, so passing them in the wrong order renders a sentence that still parses as English
// and says something else. The tests assert the rendered text for that reason.
func noInnerDeclarationsMessage(declarationType string, bodyDescription string) rule.Message {
	return rule.Message{
		Id: messageNoInnerDeclarationsId,
		Description: fmt.Sprintf("Move %s declaration to %s root. "+
			"A declaration written inside a block does not belong to that block: `var` reaches the "+
			"whole enclosing function regardless of the braces around it, and a hoisted function "+
			"binding has been three different things across engines and editions. Writing it at the "+
			"root says what actually happens.", declarationType, bodyDescription),
	}
}

// NoInnerDeclarationsOptions carries upstream's two positional options.
//
// The first says WHAT to check and the second says whether block scoped functions are allowed. Both
// are pointers so an absent key stays distinguishable from an explicitly configured one, which
// matters because one of the two defaults is not the zero value of its type.
type NoInnerDeclarationsOptions struct {
	// Both checks `var` declarations as well as functions. Upstream's first option is the enum
	// "functions" or "both", defaulting to "functions".
	Both *bool

	// BlockScopedFunctions allows a function declaration inside a block when the surrounding code
	// is strict and the language edition gives such a declaration block scope. Upstream's default
	// is "allow", so this defaults to TRUE and a zero-value struct would invert it.
	BlockScopedFunctions *bool
}

// DefaultNoInnerDeclarationsSettings is upstream's `defaultOptions: ["functions", {
// blockScopedFunctions: "allow" }]`.
func DefaultNoInnerDeclarationsSettings() NoInnerDeclarationsOptions {
	both := false
	blockScopedFunctions := true
	return NoInnerDeclarationsOptions{Both: &both, BlockScopedFunctions: &blockScopedFunctions}
}

// noInnerDeclarationsSecondOption is the object arm of the second positional option.
type noInnerDeclarationsSecondOption struct {
	BlockScopedFunctions *string `json:"blockScopedFunctions"`
}

// DecodeNoInnerDeclarationsOptions turns upstream's option list into options.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` for the reason the brief names: one default is
// TRUE, so a zero-value struct silently inverts the rule and every fixture built from a struct
// rather than routed through here would pass anyway.
//
// The wire shape is upstream's own positional list minus the severity: the rule registers with
// `DecodeOptionList`, so `["error", "both", {...}]` reaches this as `["both", {...}]` and
// `["error", "both"]` as `["both"]`. A key the second element's schema does not declare is refused,
// which is upstream's `additionalProperties: false`, and so is a third element.
func DecodeNoInnerDeclarationsOptions(list []byte) (any, error) {
	settings := DefaultNoInnerDeclarationsSettings()
	positional, err := rule.OptionElements(list, 2)
	if err != nil {
		return settings, err
	}
	if len(positional) > 0 {
		var first string
		if err := json.Unmarshal(positional[0], &first); err != nil {
			return DefaultNoInnerDeclarationsSettings(), err
		}
		applied, err := noInnerDeclarationsApplyFirst(settings, first)
		if err != nil {
			return DefaultNoInnerDeclarationsSettings(), err
		}
		settings = applied.(NoInnerDeclarationsOptions)
	}
	if len(positional) > 1 {
		var second noInnerDeclarationsSecondOption
		if err := rule.UnmarshalOptions(positional[1], &second); err != nil {
			return DefaultNoInnerDeclarationsSettings(),
				fmt.Errorf("no-inner-declarations element 2: %w", err)
		}
		if second.BlockScopedFunctions != nil {
			switch *second.BlockScopedFunctions {
			case "allow":
				allow := true
				settings.BlockScopedFunctions = &allow
			case "disallow":
				disallow := false
				settings.BlockScopedFunctions = &disallow
			default:
				return DefaultNoInnerDeclarationsSettings(), fmt.Errorf(
					"no-inner-declarations blockScopedFunctions takes \"allow\" or \"disallow\", got %q",
					*second.BlockScopedFunctions)
			}
		}
	}
	return settings, nil
}

// noInnerDeclarationsApplyFirst reads upstream's first enum, which is "functions" or "both".
func noInnerDeclarationsApplyFirst(settings NoInnerDeclarationsOptions, value string) (any, error) {
	switch value {
	case "functions":
		both := false
		settings.Both = &both
	case "both":
		both := true
		settings.Both = &both
	default:
		return DefaultNoInnerDeclarationsSettings(), fmt.Errorf(
			"no-inner-declarations takes \"functions\" or \"both\", got %q", value)
	}
	return settings, nil
}

// NoInnerDeclarations flags a function or `var` declaration that is not at the root of a program,
// a function body, or a class static block.
//
//	valid:   function doSomething() { }
//	valid:   function doSomething() { function somethingElse() { } }
//	valid:   if (test) { var fn = function() { }; }
//	valid:   class C { static { function foo() {} } }
//	invalid: if (foo) function f(){}
//	invalid: while (test) { var foo; }                      under "both"
//	invalid: class C { static { if (test) { var foo; } } }   under "both"
//
// # What counts as a root
//
// Three places accept a declaration: the program itself, the body block of a function, and a class
// static block. An export declaration wrapping one counts too, because the declaration is still at
// the root and the export is not a block. Everything else is a nested block, including a bare block
// with no statement attached to it.
//
// The parent test is deliberately two-step. A block is a legal home only when the block itself is a
// FUNCTION BODY, so `function f() { function g() {} }` is clean while `{ function g() {} }` is not,
// and both spell the parent `Block`. A single-step check on the parent kind would accept every bare
// block in the tree.
//
// # Where the finding says to move it, and why that is a walk
//
// The message names the nearest enclosing place a declaration is allowed, which upstream finds by
// walking up until it meets a static block or a function and calling everything else "program".
// The three answers are the three roots, and getting this wrong renders a grammatical sentence
// pointing at the wrong scope, which no message id assertion can see.
//
// # The strictness gate, which is why this rule is quiet on a modern tree
//
// A function declaration inside a block is block scoped in strict code from ES2015 onward, so
// upstream's default `blockScopedFunctions: "allow"` exempts it there. Every module is strict and
// so is every class body, so on a codebase of ES modules the function half of this rule reports
// almost nothing under its defaults. That is not this port going quiet, and the seeded probe tree
// distinguishes the two.
//
// Strictness is computed here rather than read off a scope table, since we have no scope table.
// Three sources, matching what the language says:
//
//	an external module        every module is strict, and IsExternalModule answers it
//	a class body              a class body is strict, including a static block
//	a directive prologue      a `'use strict'` at the top of the program, or at the top of any
//	                          enclosing function body
//
// Upstream reads `sourceCode.getScope(node).upper.isStrict`, which is the same question its scope
// analysis has already answered. The corpus varies `sourceType` and `ecmaVersion` case by case to
// exercise it, and those cases are reproduced by writing the directive or the export into the
// source rather than by configuring the harness, which cannot vary either.
var NoInnerDeclarations = rule.Rule{
	Name: "no-inner-declarations",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings, isSettings := rule.OptionsAs[NoInnerDeclarationsOptions](options)
		if !isSettings {
			settings = DefaultNoInnerDeclarationsSettings()
		}
		checkVariables := settings.Both != nil && *settings.Both
		// Defaults to TRUE, which is why a nil here has to mean allow rather than disallow.
		allowBlockScopedFunctions := settings.BlockScopedFunctions == nil ||
			*settings.BlockScopedFunctions

		check := func(node *ast.Node, declarationType string) {
			if noInnerDeclarationsIsAtARoot(node) {
				return
			}
			ctx.ReportNode(node, noInnerDeclarationsMessage(declarationType,
				noInnerDeclarationsAllowedBodyDescription(node)))
		}

		return rule.Listeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				// The language gives a block scoped function declaration well defined semantics in
				// strict code, so upstream's default declines to report it there.
				if allowBlockScopedFunctions && noInnerDeclarationsIsInStrictCode(ctx, node) {
					return
				}
				check(node, "function")
			},
			ast.KindVariableStatement: func(node *ast.Node) {
				if !checkVariables {
					return
				}
				// Only `var` reaches past its block. `let`, `const`, `using` and `await using` are
				// block scoped already, so the whole complaint does not apply and upstream's corpus
				// writes each of them as a passing case.
				declarationList := node.AsVariableStatement().DeclarationList
				if declarationList == nil {
					return
				}
				if declarationList.Flags&(ast.NodeFlagsLet|ast.NodeFlagsConst|ast.NodeFlagsUsing|
					ast.NodeFlagsAwaitUsing) != 0 {
					return
				}
				check(node, "variable")
			},
		}
	},
}

// noInnerDeclarationsIsAtARoot answers upstream's two acceptance sets.
//
// The block arm is the load-bearing one: a block is a legal home only when it IS a function body,
// which is why the grandparent is consulted rather than only the parent.
func noInnerDeclarationsIsAtARoot(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return true
	}

	if parent.Kind == ast.KindBlock {
		grandparent := parent.Parent
		if grandparent == nil {
			return false
		}
		switch grandparent.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindConstructor:
			// A method body is a function body. Upstream reaches the same answer through its
			// `FunctionExpression` set, because its parser gives a method a FunctionExpression
			// value; ours gives a MethodDeclaration directly, so the kinds are named here.
			return true
		case ast.KindClassStaticBlockDeclaration:
			// A parser difference rather than a rule one. Upstream's static block holds its
			// statements directly, so a declaration at its root has the static block as its parent
			// and is accepted by the parent set below. Ours wraps them in a Block, so the same
			// declaration arrives here with the static block as its GRANDPARENT and the parent set
			// never sees it.
			//
			// Without this arm `class C { static { var x; } }` reports, which is one of upstream's
			// passing cases and the only test in the whole corpus that fails. Found that way rather
			// than by reading, and the shape was confirmed by printing the tree.
			return true
		}
		return false
	}

	switch parent.Kind {
	case ast.KindSourceFile, ast.KindClassStaticBlockDeclaration, ast.KindModuleBlock:
		return true
	}
	return false
}

// noInnerDeclarationsAllowedBodyDescription names the nearest enclosing place a declaration is
// allowed, which is what the message tells the reader to move it to.
func noInnerDeclarationsAllowedBodyDescription(node *ast.Node) string {
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindClassStaticBlockDeclaration {
			return "class static block body"
		}
		if ast.IsFunctionLike(parent) {
			return "function body"
		}
	}
	return "program"
}

// noInnerDeclarationsIsInStrictCode answers whether the code around this node is strict.
//
// Upstream asks its scope analysis, which has already computed this. Recomputed here from the three
// things the language says make code strict, in the order that decides fastest.
func noInnerDeclarationsIsInStrictCode(ctx rule.Context, node *ast.Node) bool {
	// Every module is strict, whatever is written inside it.
	if ast.IsExternalModule(ctx.SourceFile) {
		return true
	}

	// A class body is strict, and so is a static block inside one. Checked by walking up rather
	// than by asking the nearest parent, because the declaration can be several blocks deep.
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindClassDeclaration, ast.KindClassExpression:
			return true
		}
		// A directive prologue at the top of an enclosing function body makes that body strict.
		if ast.IsFunctionLike(parent) {
			body := parent.Body()
			if body != nil && body.Kind == ast.KindBlock &&
				noInnerDeclarationsHasUseStrictPrologue(body.AsBlock().Statements) {
				return true
			}
		}
	}

	return noInnerDeclarationsHasUseStrictPrologue(ctx.SourceFile.Statements)
}

// noInnerDeclarationsHasUseStrictPrologue answers whether a statement list opens with a `'use
// strict'` directive.
//
// A directive prologue is a run of string-literal expression statements at the very start, so the
// scan stops at the first statement that is not one. Stopping matters: a `'use strict'` written
// after real code is an ordinary expression and does not make anything strict.
func noInnerDeclarationsHasUseStrictPrologue(statements *ast.NodeList) bool {
	if statements == nil {
		return false
	}
	for _, statement := range statements.Nodes {
		if statement.Kind != ast.KindExpressionStatement {
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
