package core

import (
	"encoding/json"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// StrictMode selects which spelling of strict mode the rule asks for.
//
// The values are upstream's own spellings, because the option is upstream's bare string and the
// decoder reads it straight into this type. A config written for ESLint reads the same here.
type StrictMode string

const (
	// StrictSafe picks Global or Function by what the file is. The default.
	//
	// Upstream resolves it to Global for a CommonJS file or one allowing a top-level return, and to
	// Function otherwise. In this tree it resolves to Function for every script, because nothing
	// here is parsed as CommonJS.
	StrictSafe StrictMode = "safe"

	// StrictGlobal wants exactly one directive at the top of the file and none inside a function.
	StrictGlobal StrictMode = "global"

	// StrictFunction wants a directive at the top of each top-level function and none at the file
	// level. This is the only mode that tracks nesting, because a directive is redundant inside a
	// function whose parent is already strict.
	StrictFunction StrictMode = "function"

	// StrictNever wants no directive anywhere.
	StrictNever StrictMode = "never"
)

// DecodeStrictOptions reads this rule's one option element, upstream's bare mode string.
//
// Upstream's schema is `[{ enum: ["never", "global", "function", "safe"] }]`, so `["error", "never"]`
// is the whole configuration and the element is a string, never an object. An earlier decoder here
// read it from a `{"mode": ...}` object no ESLint version accepts, which refused upstream's own form
// and stopped a run on a config copied from a real repository. The object is refused now, with the
// error naming the shape that works.
//
// Hand-rolled so an unrecognized mode fails loudly. Every arm here is selected by string equality,
// so an unknown string left in the mode would pick a silent fifth behaviour of reporting nothing at
// all. No element at all is upstream's default, Safe.
func DecodeStrictOptions(raw []byte) (any, error) {
	if len(raw) == 0 {
		return StrictSafe, nil
	}
	var mode StrictMode
	if err := json.Unmarshal(raw, &mode); err != nil {
		return StrictSafe, fmt.Errorf(
			"strict: the option is a bare string, one of \"never\", \"global\", \"function\" or "+
				"\"safe\", got %s", raw)
	}
	switch mode {
	case StrictSafe, StrictGlobal, StrictFunction, StrictNever:
	default:
		return StrictSafe, fmt.Errorf(
			"strict: unknown mode %q, wanted one of \"never\", \"global\", \"function\" or \"safe\"", mode)
	}
	return mode, nil
}

// The messages. Upstream carries ten ids for what reads like one judgment, and the split is not
// decorative: each names a different reason the directive is wrong, and three of them are conditions
// the author cannot fix by moving the directive.
var (
	messageStrictModule = rule.Message{
		Id: "module",
		Description: "This file is a module, and every module is already strict. The directive " +
			"does nothing here, so a reader has to check the file's imports to learn that it is " +
			"redundant rather than load-bearing. Delete it.",
	}
	messageStrictImplied = rule.Message{
		Id: "implied",
		Description: "The compiler already makes this file strict: `alwaysStrict`, which `strict` " +
			"turns on, has TypeScript check it as strict code and emit the directive itself. This " +
			"one does nothing, so a reader has to check the tsconfig to learn that it is " +
			"redundant rather than load-bearing. Delete it.",
	}
	messageStrictNever = rule.Message{
		Id: "never",
		Description: "This directive asks for strict mode, which the configuration does not " +
			"permit. Delete it, or change the configuration if strict mode is what this file " +
			"wants.",
	}
	messageStrictGlobal = rule.Message{
		Id: "global",
		Description: "This file has no `use strict` directive at the top, and the configuration " +
			"asks for the global form. Without it the file runs in sloppy mode, where an " +
			"assignment to an undeclared name silently creates a global. Add the directive as " +
			"the first statement.",
	}
	messageStrictFunction = rule.Message{
		Id: "function",
		Description: "This function has no `use strict` directive, and the configuration asks " +
			"for the function form. Without it the body runs in sloppy mode, where `this` falls " +
			"back to the global object and an assignment to an undeclared name creates a global. " +
			"Add the directive as the function's first statement.",
	}
	messageStrictMultiple = rule.Message{
		Id: "multiple",
		Description: "This is a second `use strict` directive in the same prologue. Only the " +
			"first has any effect, so the rest read as though they do something and do not. " +
			"Delete the duplicates.",
	}
	messageStrictUnnecessary = rule.Message{
		Id: "unnecessary",
		Description: "This function is already inside strict-mode code, so the directive changes " +
			"nothing. A redundant directive suggests the author was unsure which parts of the " +
			"file are strict. Delete it.",
	}
	messageStrictUnnecessaryInClasses = rule.Message{
		Id: "unnecessaryInClasses",
		Description: "A class body is always strict, so a directive inside a method changes " +
			"nothing and suggests the author did not know that. Delete it.",
	}
	messageStrictNonSimpleParameterList = rule.Message{
		Id: "nonSimpleParameterList",
		Description: "A `use strict` directive is a syntax error since ES2016 in a function whose " +
			"parameter list uses a default, a rest element or destructuring. This does not run at " +
			"all. Delete the directive, or make every parameter a plain name.",
	}
)

// messageStrictWrap is built per finding, because it names the function.
func messageStrictWrapFor(description string) rule.Message {
	return rule.Message{
		Id: "wrap",
		Description: fmt.Sprintf("The configuration asks for the function form of `use strict`, "+
			"and %s cannot carry a directive because its parameter list uses a default, a rest "+
			"element or destructuring, which makes the directive a syntax error since ES2016. "+
			"Wrap it in a function that does carry one.", description),
	}
}

// Strict enforces one spelling of the `use strict` directive.
//
//	valid:   export const a = 1;                       // a module is already strict
//	valid:   'use strict'; function f() {}             // under Global, in a script
//	invalid: 'use strict'; export const a = 1;         // a module, so the directive is redundant
//	invalid: const a = 1;                              // under Global, in a script
//	invalid: 'use strict'; 'use strict';               // only the first does anything
//
// # A module collapses every mode into one
//
// Upstream's first act is `if (node.sourceType === "module") mode = "module"`, which overrides
// whatever was configured. So in a module the rule has exactly one thing to say: the directive is
// redundant, delete it. Every other mode, and the whole nesting apparatus below, applies only to a
// script.
//
// That matters more here than upstream, because this tree is essentially all modules: measured,
// 2,917 of 2,919 TypeScript files carry a top-level import or export, and the two that do not are
// still compiled as modules. So the mode option is nearly inert in practice and the rule is, in
// effect, a guard against somebody pasting a directive into a module.
//
// The script arms are ported in full anyway. They are not dead code -- a `.js` or `.cjs` file reaches
// them, and so does a TypeScript file under a tsconfig that turns `alwaysStrict` off -- and
// reproducing only the arm this tree exercises would be a port of the configuration rather than of
// the rule.
//
// # A TypeScript file is strict whatever its shape
//
// Upstream's other override is `ecmaFeatures.impliedStrict`, a parser flag saying the code will run
// strict however it is written, which selects the `implied` mode: every directive is redundant and
// nothing is asked for. The TypeScript compiler's `alwaysStrict` is that flag. Under it the checker
// binds every TypeScript file as strict code, and the emit writes the directive into any output that
// is not already ESM. `strict` turns it on, and TypeScript 6 turns `strict` on by default, so an
// unset tsconfig gets it. The transpilers that read a tsconfig (esbuild, ts-jest) honor it too.
//
// Without this, a test file with no import or export, which is a script by TypeScript's module
// detection, was asked for a directive in every top-level function. That is api's Base
// ProjectSecrets.test.ts, where ESLint was silent because its flat config parses every `.ts` file as
// a module (#hks3djf). ESLint's silence was right for a different reason. The file is strict because
// the compiler says so, not because it is a module, so here it reads `implied` rather than `module`.
//
// A JavaScript file is left on the script arms. `alwaysStrict` governs what the compiler emits, and a
// `.js` or `.cjs` file in these repositories is run as written rather than emitted, so a sloppy one
// really does run sloppy.
//
// # Three findings the author cannot fix by moving the directive
//
// `nonSimpleParameterList` is the sharpest: since ES2016 a directive inside a function with a
// default, rest or destructured parameter is a SYNTAX ERROR, so that code does not run at all.
// Upstream reports it and deliberately offers no repair, because the fix is a judgment about which
// half to change. `wrap` is its counterpart under the function form, where the function cannot be
// made strict without restructuring it.
//
// # The repair only ever deletes a whole statement
//
// Every fix this rule offers removes one directive statement, and never rebuilds a span. That is
// what makes it safe on TypeScript: a parameter annotation, a return type and a generic list all sit
// outside the statement being deleted and cannot be caught in it. Measured against the installed
// rule on annotated functions, which its own JavaScript corpus cannot contain.
//
// Five of the ten findings carry that repair and five do not, and the split is upstream's
// `shouldFix`: a directive that is merely redundant is removed, while one whose presence is a
// symptom of something else is reported and left alone.
var Strict = rule.Rule{
	Name: "strict",
	// `alwaysStrict` decides whether a TypeScript script is strict, so the cached verdict has to
	// follow the tsconfig.
	ProgramReads: rule.ReadsCompilerOptions,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		mode := StrictSafe
		if decoded, configured := rule.OptionsAs[StrictMode](options); configured {
			mode = decoded
		}
		return rule.Listeners{
			ast.KindSourceFile: func(file *ast.Node) {
				checkStrict(ctx, file, mode)
			},
		}
	},
}

// strictWalker carries the mode and the nesting state through one file.
//
// Upstream keeps two stacks, pushed and popped by enter and exit listeners: one of whether each
// enclosing function is strict, and one of whether we are inside a class body. There is no exit
// listener here, so the walk is explicit and the two stacks are ordinary recursion parameters.
type strictWalker struct {
	ctx  rule.Context
	mode StrictMode
}

// checkStrict resolves the mode for this file and walks it.
func checkStrict(ctx rule.Context, file *ast.Node, mode StrictMode) {
	source := file.AsSourceFile()

	if type_checking.CompilerImpliesStrict(ctx, source) {
		mode = strictModeImplied
	} else if mode == StrictSafe {
		// Upstream picks Global when the file is CommonJS or the parser allows a top-level
		// return, and Function otherwise.
		//
		// Only the first half is expressible. `ecmaFeatures.globalReturn` is a PARSER feature
		// with no counterpart in cohere, so the second half can never fire here -- and it is not
		// a detail: seventeen of upstream's cases turn on it, and the same source under the same
		// options is clean with it and reports twice without it. Those cases are recorded as
		// unimportable in the test rather than being approximated.
		//
		// `CommonJSModuleIndicator` is what our parser sets for a file using `require` or
		// `module.exports`. Measured: it stays nil for `module.exports = 1` in a `.ts` file, so
		// even this half is narrower here than upstream, and Safe resolves to Function for
		// effectively every script this tree can produce.
		if source.CommonJSModuleIndicator != nil {
			mode = StrictGlobal
		} else {
			mode = StrictFunction
		}
	}
	// A module overrides everything, including an explicitly configured mode. This is upstream's
	// first act inside its Program listener and it is why the option is nearly inert in this tree.
	if ast.IsExternalModule(source) {
		mode = strictModeModule
	}

	walker := strictWalker{ctx: ctx, mode: mode}
	statements := source.Statements.Nodes
	directives := strictDirectivesIn(ctx, statements)

	if mode == StrictGlobal {
		// The global form wants exactly one directive at the top. A file with statements and no
		// directive is reported once, on the span covering every statement, which is upstream's
		// `loc` spanning the first through the last.
		if len(statements) > 0 && len(directives) == 0 {
			span := rule.TokenRange(ctx.SourceFile, statements[0])
			last := statements[len(statements)-1]
			ctx.ReportRange(core.NewTextRange(span.Pos(), last.End()), messageStrictGlobal)
		}
		walker.reportAllExceptFirst(directives, messageStrictMultiple, true)
	} else {
		walker.reportAll(directives, walker.fileLevelMessage(), strictShouldFix(mode))
	}

	// The function arms. Under the function form the walk carries whether the enclosing scope is
	// already strict and whether we are inside a class body; every other mode ignores both.
	walker.walkStatements(statements, false, false, false)
}

// strictModeModule is the mode a module collapses to. Not one of the configurable four, which is
// why it is a package constant rather than a StrictMode the decoder accepts.
const strictModeModule StrictMode = "module"

// strictModeImplied is the mode a TypeScript file collapses to when the compiler makes it strict. Like
// strictModeModule it is selected by the file rather than configured, which is upstream's shape too:
// `impliedStrict` is a parser feature, and `implied` is not one of the option's four values.
const strictModeImplied StrictMode = "implied"

// fileLevelMessage is what a directive at the top of the file means under the current mode.
func (w strictWalker) fileLevelMessage() rule.Message {
	switch w.mode {
	case strictModeModule:
		return messageStrictModule
	case strictModeImplied:
		return messageStrictImplied
	case StrictNever:
		return messageStrictNever
	case StrictFunction:
		// Under the function form a file-level directive is in the wrong place, and upstream says
		// so with the same id it uses to ask for the function form.
		return messageStrictFunction
	}
	return messageStrictGlobal
}

// strictShouldFix is upstream's `shouldFix`: which findings carry a repair.
//
// A directive that is merely redundant is removed. One whose presence is a symptom of a different
// problem -- the wrong form for the configuration -- is reported and left alone, because the repair
// is a judgment about what the author meant rather than a deletion.
func strictShouldFix(mode StrictMode) bool {
	return mode == strictModeModule || mode == strictModeImplied
}

// walkStatements descends the tree, tracking strictness and class nesting.
//
// `insideAnyFunction` says whether ANY function encloses this node, strict or not, which is
// upstream's `scopes.length > 0` and is what suppresses the global-level arms. `parentIsStrict`
// says whether the nearest enclosing function runs strict, which is upstream's `scopes.at(-1)`
// and turns a directive into `unnecessary`. The two are genuinely independent: a non-strict
// enclosing function sets the first and not the second. `inClass`
// whether we are inside a class body, which is always strict. Both are only consulted under the
// function form; the other modes report every directive they find wherever it is.
func (w strictWalker) walkStatements(statements []*ast.Node, insideAnyFunction bool, parentIsStrict bool, inClass bool) {
	for _, statement := range statements {
		w.walkNode(statement, insideAnyFunction, parentIsStrict, inClass)
	}
}

// walkNode visits one node, entering a function or a class body when it meets one.
func (w strictWalker) walkNode(node *ast.Node, insideAnyFunction bool, parentIsStrict bool, inClass bool) {
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		w.enterFunction(node, insideAnyFunction, parentIsStrict, inClass)
		return
	case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor,
		ast.KindSetAccessor:
		// A class member carries a function body without being one of the three function
		// kinds upstream lists, because ESTree hangs a FunctionExpression under the member
		// while our parser folds the two into one node. Omitting this arm left every
		// directive inside a method invisible, which upstream's own `class A { foo() {
		// "use strict"; } }` case caught.
		w.enterFunction(node, insideAnyFunction, parentIsStrict, inClass)
		return
	case ast.KindClassDeclaration, ast.KindClassExpression:
		// A class body is always strict, and that fact travels down rather than the class itself
		// being checked.
		node.ForEachChild(func(child *ast.Node) bool {
			w.walkNode(child, insideAnyFunction, parentIsStrict, true)
			return false
		})
		return
	}

	node.ForEachChild(func(child *ast.Node) bool {
		w.walkNode(child, insideAnyFunction, parentIsStrict, inClass)
		return false
	})
}

// enterFunction applies the rule to one function and continues into its body.
func (w strictWalker) enterFunction(node *ast.Node, insideAnyFunction bool, parentIsStrict bool, inClass bool) {
	body := strictFunctionBody(node)
	var directives []*ast.Node
	var bodyStatements []*ast.Node
	if body != nil && body.Kind == ast.KindBlock {
		bodyStatements = body.AsBlock().Statements.Nodes
		directives = strictDirectivesIn(w.ctx, bodyStatements)
	}

	isStrict := len(directives) > 0

	if w.mode == StrictFunction {
		w.enterFunctionInFunctionMode(node, directives, insideAnyFunction, parentIsStrict, inClass)
	} else if len(directives) > 0 {
		// Every other mode reports the directives it finds, unless the parameter list makes them a
		// syntax error, in which case that is the more urgent thing to say.
		if strictHasSimpleParameterList(node) {
			w.reportAll(directives, w.fileLevelMessage(), strictShouldFix(w.mode))
		} else {
			w.ctx.ReportNode(directives[0], messageStrictNonSimpleParameterList)
			w.reportAllExceptFirst(directives, messageStrictMultiple, true)
		}
	}

	// A function body is strict if it says so or if its parent was. A class body stays strict
	// throughout, which is why `inClass` travels down unchanged.
	w.walkStatements(bodyStatements, true, parentIsStrict || isStrict, inClass)
	if body != nil && body.Kind != ast.KindBlock {
		// A concise arrow body is an expression rather than a block, and can still contain a
		// function of its own.
		w.walkNode(body, true, parentIsStrict || isStrict, inClass)
	}
	// The parameter list can hold functions too, in a default value.
	w.walkParameters(node, true, parentIsStrict || isStrict, inClass)
}

// enterFunctionInFunctionMode is the nesting-aware arm, and the only one that uses the stacks.
func (w strictWalker) enterFunctionInFunctionMode(
	node *ast.Node,
	directives []*ast.Node,
	insideAnyFunction bool,
	parentIsStrict bool,
	inClass bool,
) {
	// Upstream's `scopes.length === 0 && classScopes.length === 0`. `scopes` is pushed on
	// entering ANY function, strict or not, so this asks whether a function encloses this one
	// rather than whether a STRICT one does. Deriving it from strictness instead reported both
	// the `function` and the `wrap` message on a nested function upstream leaves alone, caught
	// by `(function() { function foo(a = 0) { } }())`.
	isParentGlobal := !insideAnyFunction && !inClass

	if len(directives) > 0 {
		switch {
		case !strictHasSimpleParameterList(node):
			// A syntax error since ES2016, and the most urgent thing to say about this function.
			w.ctx.ReportNode(directives[0], messageStrictNonSimpleParameterList)
		case parentIsStrict:
			w.ctx.ReportNodeWithFixes(directives[0], messageStrictUnnecessary,
				strictRemoveDirective(w.ctx, directives[0]))
		case inClass:
			w.ctx.ReportNodeWithFixes(directives[0], messageStrictUnnecessaryInClasses,
				strictRemoveDirective(w.ctx, directives[0]))
		}
		w.reportAllExceptFirst(directives, messageStrictMultiple, true)
		return
	}

	// No directive. Only a function whose parent is neither strict nor a class needs one, because
	// everywhere else strictness is already inherited.
	//
	// Both arms below sit behind that one gate, and putting `wrap` outside it was a defect: a
	// function with a default parameter nested inside an already-strict function got asked to be
	// wrapped when upstream says nothing, because its strictness is inherited and there is
	// nothing to fix. Caught by upstream's `(function() { function foo(a = 0) { } }())`.
	if !isParentGlobal {
		return
	}
	if strictHasSimpleParameterList(node) {
		w.ctx.ReportNode(node, messageStrictFunction)
		return
	}
	// The function cannot carry a directive at all, so upstream asks for a wrapper instead and
	// names the function so the reader can find it.
	w.ctx.ReportNode(node, messageStrictWrapFor(strictDescribeFunction(w.ctx, node)))
}

// walkParameters descends into parameter default values, which can hold functions.
func (w strictWalker) walkParameters(node *ast.Node, insideAnyFunction bool, parentIsStrict bool, inClass bool) {
	for _, parameter := range strictParametersOf(node) {
		if initializer := parameter.AsParameterDeclaration().Initializer; initializer != nil {
			w.walkNode(initializer, insideAnyFunction, parentIsStrict, inClass)
		}
	}
}

// reportAll reports every directive with one message.
func (w strictWalker) reportAll(directives []*ast.Node, message rule.Message, fix bool) {
	w.reportSlice(directives, 0, message, fix)
}

// reportAllExceptFirst reports every directive after the first, which is upstream's `multiple`.
func (w strictWalker) reportAllExceptFirst(
	directives []*ast.Node,
	message rule.Message,
	fix bool,
) {
	w.reportSlice(directives, 1, message, fix)
}

// reportSlice reports directives from an index onward.
func (w strictWalker) reportSlice(
	directives []*ast.Node,
	from int,
	message rule.Message,
	fix bool,
) {
	for index := from; index < len(directives); index++ {
		if fix {
			w.ctx.ReportNodeWithFixes(directives[index], message,
				strictRemoveDirective(w.ctx, directives[index]))
			continue
		}
		w.ctx.ReportNode(directives[index], message)
	}
}

// strictRemoveDirective builds the repair, which deletes one whole directive statement.
//
// The span is the statement's own token range, so nothing outside it can be caught in the deletion.
// That is the whole reason this fixer is safe on TypeScript: a parameter annotation, a return type
// and a generic list all live in the function's signature, which is a different node entirely.
func strictRemoveDirective(ctx rule.Context, directive *ast.Node) rule.Fix {
	return rule.RemoveRange(rule.TokenRange(ctx.SourceFile, directive))
}

// strictDirectivesIn returns the `use strict` directives at the head of a statement list.
//
// Upstream reads the DIRECTIVE PROLOGUE, which is the run of expression statements whose expression
// is a string literal, starting at the first statement. The scan stops at the first statement that
// is not one, so a directive after any real code is not a directive at all -- it is a string
// expression, and the engine ignores it.
//
// Only `use strict` counts. Another directive such as `use asm` occupies a prologue slot and is not
// this rule's business, and upstream's value comparison skips it the same way.
func strictDirectivesIn(ctx rule.Context, statements []*ast.Node) []*ast.Node {
	var directives []*ast.Node
	for _, statement := range statements {
		if statement.Kind != ast.KindExpressionStatement {
			break
		}
		expression := statement.AsExpressionStatement().Expression
		if expression == nil || expression.Kind != ast.KindStringLiteral {
			break
		}
		if expression.Text() != "use strict" {
			// A different prologue directive, such as `use asm`. Upstream writes into a SPARSE
			// array here -- `directives[i] = statement`, indexed by the statement's own position
			// -- and skips the assignment for anything that is not `use strict`. Its later
			// `forEach` and `slice` then skip the holes, so a `use strict` sitting AFTER another
			// directive is never reported at all.
			//
			// Measured against the installed rule: `'use asm'; 'use strict';` is clean and
			// `'use strict'; 'use asm';` reports. That is an upstream defect rather than a
			// judgment -- the directive is equally redundant either way round -- and it is
			// reproduced rather than corrected, because the corpus is the specification and a
			// port that quietly reported the case would disagree with the tool it replaces.
			//
			// Reproduced by STOPPING here rather than by continuing, which is what makes the
			// dense slice below equivalent to upstream's sparse one: once a hole exists, nothing
			// after it can be reported.
			break
		}
		directives = append(directives, statement)
	}
	return directives
}

// strictHasSimpleParameterList answers whether every parameter is a plain name.
//
// Since ES2016 a `use strict` directive is a syntax error in a function whose parameter list uses a
// default, a rest element or destructuring, so this is what separates a redundant directive from
// one that stops the file running. A TypeScript annotation does NOT make a parameter non-simple:
// `function f(a: string)` is still a plain name with a type on it, which is a distinction upstream's
// JavaScript corpus cannot express and which is measured here rather than assumed.
func strictHasSimpleParameterList(node *ast.Node) bool {
	for _, parameter := range strictParametersOf(node) {
		declaration := parameter.AsParameterDeclaration()
		if declaration.Initializer != nil || declaration.DotDotDotToken != nil {
			return false
		}
		if name := declaration.Name(); name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
	}
	return true
}

// strictParametersOf returns a function-like node's parameters.
func strictParametersOf(node *ast.Node) []*ast.Node {
	if parameters := node.Parameters(); parameters != nil {
		return parameters
	}
	return nil
}

// strictFunctionBody returns a function-like node's body, or nil.
func strictFunctionBody(node *ast.Node) *ast.Node {
	return node.Body()
}

// strictDescribeFunction names a function for the `wrap` message.
//
// Upstream uses `getFunctionNameWithKind`, which renders `function 'foo'` or `arrow function`. The
// same shape is produced here from the node's own name, and an anonymous function is described by
// its kind alone.
func strictDescribeFunction(ctx rule.Context, node *ast.Node) string {
	kind := "function"
	if node.Kind == ast.KindArrowFunction {
		kind = "arrow function"
	}
	if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
		return fmt.Sprintf("%s `%s`", kind, name.Text())
	}
	return "this " + kind
}
