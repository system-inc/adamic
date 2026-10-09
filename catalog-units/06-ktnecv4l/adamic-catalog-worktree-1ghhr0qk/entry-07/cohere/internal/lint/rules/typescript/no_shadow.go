package typescript

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoShadow = rule.Message{
	Id: "noShadow",
	Description: "This binding takes a name that is already bound in an enclosing scope. Every " +
		"read of the name inside this scope now resolves here instead of to the outer binding, " +
		"so code written against the outer one silently reads the inner one. Rename the inner " +
		"binding.",
}

var messageNoEnumShadow = rule.Message{
	Id: "noEnumShadow",
	Description: "This enum member takes a name that is already bound in an enclosing scope. " +
		"Enum members are added to the enum's own scope, so a later member initializer that " +
		"names it resolves to this member rather than to the outer binding. Rename the member.",
}

var messageNoShadowGlobal = rule.Message{
	Id: "noShadowGlobal",
	Description: "This binding takes the name of a global the language or the standard library " +
		"already provides. Inside this scope the name now resolves here, so code that reads it " +
		"expecting the built-in one silently gets this binding instead. Rename the binding.",
}

// NoShadowOptions are upstream's six options, under upstream's names.
//
// Hoist is one of upstream's five spellings, "all", "functions", "functions-and-types", "never" and
// "types". The config layer checks the value against upstream's schema before the decoder sees it.
type NoShadowOptions struct {
	Allow                                      []string
	BuiltinGlobals                             bool
	Hoist                                      string
	IgnoreFunctionTypeParameterNameValueShadow bool
	IgnoreOnInitialization                     bool
	IgnoreTypeValueShadow                      bool
}

// noShadowWire is the shape the config layer delivers, before defaults are applied. Pointers, because
// two of the options default to true and a plain bool cannot tell an absent key from an explicit false.
type noShadowWire struct {
	Allow                                      []string `json:"allow"`
	BuiltinGlobals                             *bool    `json:"builtinGlobals"`
	Hoist                                      *string  `json:"hoist"`
	IgnoreFunctionTypeParameterNameValueShadow *bool    `json:"ignoreFunctionTypeParameterNameValueShadow"`
	IgnoreOnInitialization                     *bool    `json:"ignoreOnInitialization"`
	IgnoreTypeValueShadow                      *bool    `json:"ignoreTypeValueShadow"`
}

// DefaultNoShadowOptions is the configuration upstream applies when the rule is written bare.
func DefaultNoShadowOptions() NoShadowOptions {
	return NoShadowOptions{
		Hoist: "functions-and-types",
		IgnoreFunctionTypeParameterNameValueShadow: true,
		IgnoreTypeValueShadow:                      true,
	}
}

// DecodeNoShadowOptions reads this rule's configuration from the config layer.
//
// Its own decoder rather than `rule.DecodeOptionsInto`, for no-redeclare's reason: two options default
// to true, and the generic helper's zero value would turn both off for every bare configuration.
func DecodeNoShadowOptions(raw []byte) (any, error) {
	options := DefaultNoShadowOptions()
	if len(raw) == 0 {
		return options, nil
	}
	var wire noShadowWire
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return options, err
	}
	options.Allow = wire.Allow
	if wire.BuiltinGlobals != nil {
		options.BuiltinGlobals = *wire.BuiltinGlobals
	}
	if wire.Hoist != nil {
		switch *wire.Hoist {
		case "all", "functions", "functions-and-types", "never", "types":
			options.Hoist = *wire.Hoist
		default:
			return options, fmt.Errorf("hoist is %q, and upstream takes all, functions, functions-and-types, never or types", *wire.Hoist)
		}
	}
	if wire.IgnoreFunctionTypeParameterNameValueShadow != nil {
		options.IgnoreFunctionTypeParameterNameValueShadow = *wire.IgnoreFunctionTypeParameterNameValueShadow
	}
	if wire.IgnoreOnInitialization != nil {
		options.IgnoreOnInitialization = *wire.IgnoreOnInitialization
	}
	if wire.IgnoreTypeValueShadow != nil {
		options.IgnoreTypeValueShadow = *wire.IgnoreTypeValueShadow
	}
	return options, nil
}

// NoShadow flags a binding whose name is already bound in an enclosing scope.
//
//	valid:   const x = 1; type x = string;
//	valid:   var a = 1; var b = function a() {};
//	valid:   { let a; } let a;
//	valid:   function foo() { var Object = 0; }
//	invalid: var a = 3; function b() { var a = 10; }
//	invalid: type T = 1; function foo<T>(arg: T) {}
//	invalid: var a = 1; var b = wrap(function a() {});
//	invalid: enum A { A, B }
//
// # Upstream's six options, each checked against upstream's own rows
//
// `allow`, `builtinGlobals`, `hoist`, `ignoreFunctionTypeParameterNameValueShadow`,
// `ignoreOnInitialization` and `ignoreTypeValueShadow`, under upstream's defaults when written bare.
// All 239 rows of upstream's two test files, and edge rows for what they leave open, were replayed
// through the installed rule (typescript-eslint 8.71.0) and are replayed here by
// TestNoShadowUpstreamCorpus; no_shadow_corpus_data_test.go says how.
//
// The checks run in upstream's order: `allow`, a `declare` in a definition file, then the search for
// the shadowed binding, then the type-and-value, function-type-parameter, static-generic and
// declaration-merging carve-outs, then the initializer tests, then the temporal dead zone.
//
//   - builtinGlobals: when no enclosing binding has the name, a global ESLint or the program's lib
//     declares is the shadowed one, reported as `noShadowGlobal`. The set is no-redeclare's
//     (builtin_globals.go). A script's top level IS the global scope, which upstream never examines,
//     so its declarations are clean and only its nested scopes can shadow a builtin; ruled on #e1zk9s0
//     for no-redeclare, and the same TypeScript notion of a script applies here. Upstream's rows that
//     configure `globals` are pinned, since cohere carries no globals configuration.
//   - hoist: which outer declarations are already in scope before they are written, below.
//   - ignoreOnInitialization: a shadow inside a function expression passed to a call that sits in
//     the outer binding's initializer, below at bindingIsInitializedThrough.
//
// Four default behaviours changed with the replay, each where cohere disagreed with the installed
// rule: a function's first declaration hoists only when it has a body, so an overload set does not;
// a bodiless function's name counts with the function-type parameter names; a binding merged from
// several declarations reports once, at its first; and the finding's span runs through the
// identifier's `?`, `!` and type annotation, as upstream's Identifier node does.
//
// # The scope graph is the binder's, not one this rule builds
//
// Upstream is written entirely against eslint-scope: walk every scope, enumerate `scope.variables`,
// and for each ask `findVariable(scope.upper, name)`. That reads as a missing substrate here and is
// not one. TypeScript's binder already populates a per-container symbol table, reachable through
// `ast.IsLocalsContainer` plus `ast.GetLocals`, and it answers the same question.
//
// Two consequences, both measured:
//
//   - The table is EMPTY under `rule_testing.Run` and populated under `RunTyped`, because the
//     binder runs with the program. So this rule declares NeedsTypeChecker for BINDING, not because
//     it asks any type question. Declared false it would receive a nil checker and an empty table,
//     and go silently, vacuously clean on every file.
//   - `ast.GetLocals` PANICS on a node that is not a locals container rather than returning nil, and
//     an enum declaration is a shape that reaches it. Every call here is guarded by
//     `ast.IsLocalsContainer` first.
//
// # Three shapes the binder answers differently from eslint-scope, and each is load bearing
//
// **Declaration merging does the work of upstream's `isTypeValueShadow` and its namespace cases.**
// In eslint-scope `const x = 1; type x = string;` is two variables in one scope and upstream needs
// an option to stop reporting it. Here the binder MERGES them into a single symbol carrying
// `0x80002` (BlockScopedVariable plus TypeAlias), so there is one entry in the table and no pair to
// compare. Measured the same way for `class Foo {} namespace Foo {}` (`0x220`) and for
// `class Foo {} interface Foo {}`. Those corpus cases come out clean without a carve-out, which is
// why no ignoreTypeValueShadow logic appears below.
//
// **A function-expression or class-expression NAME is in no table at all.** Upstream gives each its
// own `functionExpressionName` scope holding exactly that name, and ten-plus corpus cases turn on
// it. Measured here: for `(function a() {})`, `a` is absent from the expression's own locals and
// from its parent's, and is reachable only as the symbol on the name node itself. So those bindings
// are collected from AST SHAPE rather than from the scope tables, in collectExpressionNames below.
//
// **An interface declaration is NOT a locals container either, and this one is not in the probe
// package's map.** Its type parameters are therefore invisible to the walk the same way an enum's
// members are, and they come off `InterfaceDeclaration.TypeParameters` instead. A type ALIAS *is* a
// locals container and needs none of this, which is what makes the absence easy to miss: the two
// read as the same construct and behave differently. Found by driving the installed rule over
// shapes the corpus does not write, because upstream's defaults corpus contains no interface type
// parameter anywhere.
//
// **An enum declaration is NOT a locals container.** `ast.IsLocalsContainer` answers false for it,
// so the general walk cannot see its members at all. They live on the enum symbol's `Exports`
// table instead, which is why the enum arm reads a different accessor from everything else.
//
// # hoist: which outer declarations are in scope before they are written
//
// An inner binding declared textually BEFORE the outer one it shadows is reported only when the
// outer declaration hoists under the setting. `functions-and-types`, the default, hoists a function
// declaration with a body, an interface and a type alias; `functions` and `types` hoist one half;
// `never` hoists nothing, and `all` reports every such shadow. An outer `var`, `let`, `const` or class
// never hoists, so before it is written the inner binding shadows nothing yet.
//
//	{ let a; } function a() {}   reports by default, the function hoists
//	{ let a; } var a;            clean unless hoist is all
//
// A function hoists by its FIRST declaration, upstream's `defs[0]`, and the first of an overload set
// is a bodiless signature (upstream's TSDeclareFunction), which is not in either set.
var NoShadow = rule.Rule{
	Name: "@typescript-eslint/no-shadow",

	// For BINDING rather than for types. See the doc comment above: the locals tables this rule
	// reads are populated by the binder, and the binder runs with the program.
	NeedsTypeChecker: true,

	// builtinGlobals asks whether a global's declarations live in a lib file.
	ProgramReads: rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Declined once per file rather than once per node. NeedsTypeChecker governs the
		// registration path only; a Context built by hand still arrives with a nil checker, and
		// `TestNoRegisteredRuleCrashesOnAbsentOptionalNodes` builds exactly that.
		if ctx.TypeChecker == nil {
			return nil
		}
		settings, ok := rule.OptionsAs[NoShadowOptions](options)
		if !ok {
			// A bare "error" arrives as nil, whose zero value turns off both options that default to
			// true; the real defaults are what a bare configuration means.
			settings = DefaultNoShadowOptions()
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				checkFileForShadowing(ctx, node, settings)
			},
		}
	},
}

// shadowBinding is one name bound in one scope: this rule's stand-in for an eslint-scope Variable.
//
// Upstream's Variable carries `identifiers` and `defs`. The pair that matters is the same here: the
// identifier the finding is reported at, and the declaration node whose kind decides hoisting and
// whose position decides the temporal dead zone.
type shadowBinding struct {
	name string

	// identifier is where the finding points, and is the node whose position upstream compares.
	identifier *ast.Node

	// declaration is the node the binding was declared by, its FIRST declaration when the symbol
	// merges several (upstream's `defs[0]`). Its KIND answers the hoisting question and its symbol
	// answers the enum question.
	declaration *ast.Node

	// declarations is every declaration of the binding in its scope, in source order, which upstream's
	// `defs.some` and `defs.every` tests read. Just `declaration` for a binding collected from shape.
	declarations []*ast.Node

	// symbol is nil for an expression-name binding collected from AST shape, because those are not
	// in any table, and for a builtin global. Everything reading it guards first.
	symbol *ast.Symbol

	// global marks the stand-in for a builtin global under builtinGlobals, which has no identifier,
	// no declaration and no symbol: upstream's variable with no identifiers.
	global bool
}

// shadowScope is one locals container plus the bindings it holds.
//
// Deliberately not a graph. Upstream walks `scope.upper` because eslint-scope's scopes are linked;
// here the containers nest in the AST, so the enclosing scopes of a container are just its
// ancestors in this slice, which is built outermost-first.
type shadowScope struct {
	container *ast.Node
	bindings  []shadowBinding
}

// checkFileForShadowing collects every scope in the file and reports each shadowed binding.
//
// The whole rule runs inside the KindSourceFile listener rather than across several kind listeners,
// because a shadowing judgment needs the ENCLOSING scopes and there is no rule.OnExit and no
// upward walk from a listener. KindSourceFile fires before its children, so the tree is walked
// here directly and the answer is computed in one pass.
func checkFileForShadowing(ctx rule.Context, sourceFile *ast.Node, settings NoShadowOptions) {
	var scopes []shadowScope

	// The chain of containers currently open, outermost first. A binding's enclosing scopes are
	// exactly the entries below its own.
	var openScopes []*shadowScope

	var visit func(node *ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}

		opened := false

		// node.Locals() reads the table without creating it. ast.GetLocals creates a missing one,
		// which writes to a node another worker's checker may be resolving names through, a data
		// race; a container the binder gave no locals has none to shadow, and ranging nil is empty.
		if ast.IsLocalsContainer(node) {
			scope := shadowScope{container: node}
			for name, symbol := range node.Locals() {
				if symbol == nil {
					continue
				}
				// One binding per symbol, as upstream has one variable per name in a scope: an
				// overload set or a merged value and type reports once, at its first declaration.
				// Declarations was one binding each, and an overload set shadowing an outer name
				// reported once per signature.
				for _, declaration := range symbol.Declarations {
					identifier := bindingIdentifierOf(declaration)
					if identifier == nil || bindingIsAnIndexSignatureParameter(declaration) {
						continue
					}
					scope.bindings = append(scope.bindings, shadowBinding{
						name:         name,
						identifier:   identifier,
						declaration:  declaration,
						declarations: symbol.Declarations,
						symbol:       symbol,
					})
					break
				}
			}
			// A class's type parameters are not in its locals: the binder declares them as class
			// members, on the class symbol. They are bindings of the class's own scope to upstream,
			// so `class A<T> { n<T>() {} }` reports, and they come off the declaration here the way
			// an interface's do below. Found by the replay's edge rows; upstream's own rows write only
			// the static-method form, which is clean for another reason.
			if node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression {
				if typeParameterList := node.ClassLikeData().TypeParameters; typeParameterList != nil {
					for _, typeParameter := range typeParameterList.Nodes {
						identifier := bindingIdentifierOf(typeParameter)
						if identifier == nil {
							continue
						}
						scope.bindings = append(scope.bindings, shadowBinding{
							name:         identifier.Text(),
							identifier:   identifier,
							declaration:  typeParameter,
							declarations: []*ast.Node{typeParameter},
							symbol:       ctx.TypeChecker.GetSymbolAtLocation(identifier),
						})
					}
				}
			}
			scopes = append(scopes, scope)
			openScopes = append(openScopes, &scopes[len(scopes)-1])
			opened = true
		}

		// A function or class EXPRESSION name is in no locals table, so it is collected from shape.
		// It behaves as a binding in a scope of its own between the expression and its parent, which
		// is what upstream's functionExpressionName scope is, so it is recorded against the
		// expression's own scope when the expression opened one and against the enclosing scope
		// otherwise (a class expression is not a locals container).
		if expressionName := expressionSelfName(node); expressionName != nil && len(openScopes) > 0 {
			target := openScopes[len(openScopes)-1]
			target.bindings = append(target.bindings, shadowBinding{
				name:         expressionName.Text(),
				identifier:   expressionName,
				declaration:  node,
				declarations: []*ast.Node{node},
				symbol:       nil,
			})
		}

		// An enum's members are not in any locals table, because an enum declaration is not a
		// locals container. They live on the enum symbol's Exports table instead, and they belong in
		// a scope of their OWN nested inside the enum rather than in the enclosing one: the enum's
		// name is bound outside it and its members inside it, so `enum A { A }` is a member shadowing
		// the enum name. Folding the members into the enclosing scope would make the two siblings,
		// and the containment test that finds an enclosing binding would never fire.
		// An interface declaration is NOT a locals container either, so its type parameters are
		// invisible to the walk the same way an enum's members are. They belong in a scope of
		// their own covering the interface, because `type T = 1; interface I<T> {}` reports.
		//
		// This is a third gap of the same family as the enum and the expression name, and it was
		// found by a wide sweep against the installed rule rather than by the imported corpus,
		// which writes no interface type parameter anywhere. A type ALIAS is a locals container and
		// needs none of this, which is what made the absence easy to miss.
		if node.Kind == ast.KindInterfaceDeclaration {
			interfaceScope := shadowScope{container: node}
			// TypeParameters is OPTIONAL and every interface without a `<T>` has it nil. The walk
			// recovers per FILE rather than per rule, so dereferencing it would cost every rule in
			// the package its verdict on any file holding a plain interface, which is most of them.
			// Skipped rather than returned when absent: a bare return here would abandon the
			// whole subtree, losing every binding nested inside the interface's members for the
			// common case of an interface with no type parameters at all.
			if typeParameterList := node.AsInterfaceDeclaration().TypeParameters; typeParameterList != nil {
				for _, typeParameter := range typeParameterList.Nodes {
					identifier := bindingIdentifierOf(typeParameter)
					if identifier == nil {
						continue
					}
					interfaceScope.bindings = append(interfaceScope.bindings, shadowBinding{
						name:         identifier.Text(),
						identifier:   identifier,
						declaration:  typeParameter,
						declarations: []*ast.Node{typeParameter},
						symbol:       ctx.TypeChecker.GetSymbolAtLocation(identifier),
					})
				}
			}
			if len(interfaceScope.bindings) > 0 {
				scopes = append(scopes, interfaceScope)
			}
		}

		if node.Kind == ast.KindEnumDeclaration {
			enumScope := shadowScope{container: node}
			collectEnumMembers(ctx, node, &enumScope)
			if len(enumScope.bindings) > 0 {
				scopes = append(scopes, enumScope)
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})

		if opened {
			openScopes = openScopes[:len(openScopes)-1]
		}
	}
	visit(sourceFile)

	// Report. A scope's enclosing scopes are the ones that CONTAIN its container, which is decided
	// by node ranges rather than by a parent pointer walk, so a binding collected from AST shape
	// participates on the same terms as one read out of a table.
	for scopeIndex := range scopes {
		scope := &scopes[scopeIndex]
		if scopeIsInsideGlobalAugmentation(scope.container) {
			continue
		}
		for _, binding := range scope.bindings {
			reportIfShadowing(ctx, settings, scopes, scopeIndex, binding)
		}
	}
}

// reportIfShadowing finds the nearest enclosing binding of the same name and reports if one exists.
func reportIfShadowing(ctx rule.Context, settings NoShadowOptions, scopes []shadowScope, scopeIndex int, binding shadowBinding) {
	// A `this` parameter is a pseudo-parameter that cannot be shadowed. Upstream skips it by name
	// and definition type; here the parameter's name node is the `this` keyword.
	if binding.name == "this" {
		return
	}

	if slices.Contains(settings.Allow, binding.name) {
		return
	}

	if bindingIsDeclaredInADefinitionFile(ctx, binding) {
		return
	}

	// `arguments` has no identifier of its own in eslint-scope, so upstream's
	// `variable.identifiers.length === 0` skip covers it. Here a real `var arguments;` DOES have an
	// identifier, and upstream's corpus keeps `var arguments; function bar() {}` clean for the same
	// underlying reason: the implicit `arguments` it would shadow is not a declared binding.
	// Nothing declares it in a table here either, so this case is clean without a carve-out and the
	// name needs no special handling.

	shadowed, shadowedContainer, found := findEnclosingBinding(scopes, scopeIndex, binding)
	if !found {
		// No enclosing binding has the name, so under builtinGlobals the global scope is next, as
		// upstream's findVariable reaches it. Not from a script's own top level, which IS the global
		// scope and is never examined; an expression name collected there sits in a scope of its own,
		// so it still looks outward.
		if !settings.BuiltinGlobals {
			return
		}
		if container := scopes[scopeIndex].container; container.Kind == ast.KindSourceFile &&
			!ast.IsExternalModule(container.AsSourceFile()) && binding.symbol != nil {
			return
		}
		if !isBuiltinGlobal(ctx.TypeChecker, ctx.Program, binding.name) {
			return
		}
		shadowed = shadowBinding{name: binding.name, global: true}
	}

	// There is deliberately NO self-comparison guard here, and the reason is the SEARCH BOUND
	// rather than any identity test. findEnclosingBinding searches candidateIndex < scopeIndex
	// only, and no container appears twice in the walk, so a binding is never compared against a
	// binding in its own scope through that path. Measured rather than argued: neutering a guard
	// here survived the whole fixture set, and a probe over a file holding a function, a method
	// and an enum found zero duplicate containers in the walk.
	//
	// This verdict names the callers it was taken against: the two loops in findEnclosingBinding.
	// A third search path, or one that widened the bound to include the scope itself, voids it.
	//
	// Upstream needs its equivalent skips (isDuplicatedClassNameVariable,
	// isDuplicatedEnumNameVariable) because eslint-scope binds a class or enum name TWICE, once
	// outside the declaration and once inside its own scope. Our binder does not duplicate them,
	// which is why neither has a counterpart in this rule.

	// ignoreTypeValueShadow, default true: a type shadowing a value, or a value shadowing a type,
	// is not a shadow anybody can be confused by, because the two live in different declaration
	// spaces and no reference resolves to the wrong one.
	if settings.IgnoreTypeValueShadow && bindingCrossesTheTypeValueBoundary(binding, shadowed) {
		return
	}

	// ignoreFunctionTypeParameterNameValueShadow, default true: a parameter NAME inside a function
	// TYPE is documentation rather than a binding. Nothing can read it, so it shadows nothing.
	if settings.IgnoreFunctionTypeParameterNameValueShadow &&
		bindingIsAFunctionTypeParameterName(ctx, binding, shadowed) {
		return
	}

	// A static method's type parameter over its class's: the class's type parameters are not in
	// scope in a static member, which upstream's scope analyser cannot see and handles by hand.
	if bindingIsAStaticMethodGenericOverAClassGeneric(binding, shadowed) {
		return
	}

	// A type-only import of a name, augmented inside `declare module` naming that same module, is
	// declaration merging rather than shadowing: the interface being declared is the very thing the
	// import refers to.
	if bindingIsExternalDeclarationMerging(ctx, scopes[scopeIndex].container, binding, shadowed) {
		return
	}

	// One `infer` does not shadow another. See the helper for the measurement behind the narrowness.
	if bothBindingsAreInferredTypeParameters(binding, shadowed) {
		return
	}

	// `var a = function a() {};` is clean while `var a = wrap(function a() {})` reports. Upstream
	// separates them in isOnInitializer; the same test here, from AST shape.
	if expressionNameIsDirectInitializerOf(binding, shadowed) {
		return
	}

	if settings.IgnoreOnInitialization && bindingIsInitializedThrough(ctx, binding, shadowed, shadowedContainer) {
		return
	}

	// The temporal dead zone under the hoist setting. See the doc comment.
	if settings.Hoist != "all" && bindingIsInTemporalDeadZoneOf(binding, shadowed, settings.Hoist) {
		return
	}

	reportRange := noShadowReportRange(ctx, binding.identifier)

	if shadowed.global {
		ctx.ReportRange(reportRange, rule.Message{
			Id: messageNoShadowGlobal.Id,
			Description: fmt.Sprintf("%s The name '%s' is already a global variable.",
				messageNoShadowGlobal.Description, binding.name),
		})
		return
	}

	shadowedLine, shadowedColumn := noShadowLineAndColumnOf(ctx, shadowed.identifier)

	// An enum member shadowing gets its own message, because the consequence is different: the
	// member is added to the enum's scope, so a LATER member initializer naming it resolves to the
	// member rather than to the outer binding.
	if bindingIsAnEnumName(shadowed) {
		ctx.ReportRange(reportRange, rule.Message{
			Id: messageNoEnumShadow.Id,
			Description: fmt.Sprintf("%s The name '%s' is already declared in the upper scope on "+
				"line %d column %d.", messageNoEnumShadow.Description, binding.name,
				shadowedLine, shadowedColumn),
		})
		return
	}

	ctx.ReportRange(reportRange, rule.Message{
		Id: messageNoShadow.Id,
		Description: fmt.Sprintf("%s The name '%s' is already declared in the upper scope on "+
			"line %d column %d.", messageNoShadow.Description, binding.name,
			shadowedLine, shadowedColumn),
	})
}

// noShadowReportRange is where upstream's finding points: the binding's identifier, through the `?`,
// `!` and type annotation upstream's Identifier node carries.
//
// typescript-estree hangs a variable's or a parameter's annotation on its Identifier, so the node
// upstream reports covers `a?: number` and `a!: number`. A rest parameter's annotation hangs on the
// RestElement instead and a destructured name has none, so both report the name alone. Measured on the
// installed rule; the edge rows named "span:" record each shape.
func noShadowReportRange(ctx rule.Context, identifier *ast.Node) core.TextRange {
	nameRange := rule.TokenRange(ctx.SourceFile, identifier)
	end := identifier.End()
	switch declaration := identifier.Parent; declaration.Kind {
	case ast.KindVariableDeclaration:
		variable := declaration.AsVariableDeclaration()
		if variable.ExclamationToken != nil {
			end = variable.ExclamationToken.End()
		}
		if variable.Type != nil {
			end = variable.Type.End()
		}
	case ast.KindParameter:
		parameter := declaration.AsParameterDeclaration()
		if parameter.DotDotDotToken != nil {
			break
		}
		if parameter.QuestionToken != nil {
			end = parameter.QuestionToken.End()
		}
		if parameter.Type != nil {
			end = parameter.Type.End()
		}
	}
	return core.NewTextRange(nameRange.Pos(), end)
}

// findEnclosingBinding searches outward for a binding of the same name in a scope that contains
// this one.
//
// Upstream's `findVariable(scope.upper, name)` walks the scope chain from the parent upward and
// returns the first match. Here containment is decided by range, and scopes were appended in
// pre-order, so every candidate enclosing scope sits at a LOWER index and strictly contains this
// container. Searching from the nearest outward reproduces upstream's ordering.
//
// The container returned is the scope the shadowed binding was found in, which
// ignoreOnInitialization compares against; nil for an expression name found in the binding's own
// scope, which stands for a scope of its own that no container is.
func findEnclosingBinding(scopes []shadowScope, scopeIndex int, binding shadowBinding) (shadowBinding, *ast.Node, bool) {
	inner := scopes[scopeIndex].container
	for candidateIndex := scopeIndex - 1; candidateIndex >= 0; candidateIndex-- {
		candidate := &scopes[candidateIndex]
		if !nodeContains(candidate.container, inner) {
			continue
		}
		for _, outer := range candidate.bindings {
			// No identity test: candidateIndex is strictly below scopeIndex and no container
			// appears twice in the walk, so `candidate` is never the binding's own scope and
			// `outer` is never `binding`. A mutant neutering an identity test here survived every
			// fixture, which is what sent us to measure the bound instead of trusting the test.
			if outer.name != binding.name {
				continue
			}
			return outer, candidate.container, true
		}
	}

	// A binding can also be shadowed by another binding in its OWN scope when one of the two came
	// from AST shape rather than from the table, which is how a function expression's name and a
	// declaration inside it end up together. Upstream separates them into two scopes; here the
	// order they were collected in preserves the nesting, so an expression name recorded before a
	// binding in the same slice encloses it.
	own := &scopes[scopeIndex]
	for _, outer := range own.bindings {
		// This loop DOES search the binding's own scope, so unlike the one above it can be handed
		// the binding itself, and the identity test is what declines that.
		if outer.name != binding.name || outer.identifier == binding.identifier {
			continue
		}
		// The symbol test is the whole condition, and a containment test beside it would be
		// REDUNDANT rather than merely untested. An expression-name binding is recorded against
		// the innermost scope open at the point the walk reaches the expression, so any binding
		// sharing that scope with it is inside the expression by construction. Two sibling
		// function expressions each open their own container and never land in one list together.
		//
		// Measured rather than argued, because a mutant removing the containment test survived and
		// two fixtures written for it did too: the condition was instrumented to panic whenever it
		// held with containment FALSE, and across every case in this file it never fired.
		if outer.symbol == nil {
			return outer, nil, true
		}
	}
	return shadowBinding{}, nil, false
}

// bindingIdentifierOf returns the identifier a declaration binds its name with.
//
// Nil for a declaration with no readable name, which is how a default export or a binding pattern
// arrives. A caller that reported on one of those would point at the wrong text.
func bindingIdentifierOf(declaration *ast.Node) *ast.Node {
	if declaration == nil {
		return nil
	}
	name := declaration.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return nil
	}
	return name
}

// expressionSelfName returns the name of a named function or class EXPRESSION.
//
// Nil for everything else, including an anonymous expression and every declaration form, because a
// declaration's name is already in a table and collecting it twice would double every finding.
func expressionSelfName(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	if node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindClassExpression {
		return nil
	}
	name := node.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return nil
	}
	return name
}

// collectEnumMembers records an enum's members as bindings in the enclosing scope.
//
// Through the enum symbol's Exports table rather than through Locals, because an enum declaration
// is not a locals container and `ast.GetLocals` would panic on it. Measured: for `enum E { A }` the
// symbol carries `exports=[A]` and `members=[]`.
func collectEnumMembers(ctx rule.Context, enumDeclaration *ast.Node, scope *shadowScope) {
	name := enumDeclaration.Name()
	if name == nil {
		return
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil {
		return
	}
	for memberName, memberSymbol := range symbol.Exports {
		if memberSymbol == nil {
			continue
		}
		for _, declaration := range memberSymbol.Declarations {
			identifier := bindingIdentifierOf(declaration)
			if identifier == nil {
				continue
			}
			// Two `enum A` declarations MERGE onto one symbol, so the Exports table reached from
			// either one lists every member of both. Without this test each declaration collects
			// the whole merged membership and every finding is emitted once per declaration, at
			// the identical span. Upstream cannot hit this because its members come from the
			// declaration's own scope rather than from a merged symbol.
			if !nodeContains(enumDeclaration, declaration) {
				continue
			}
			scope.bindings = append(scope.bindings, shadowBinding{
				name:         memberName,
				identifier:   identifier,
				declaration:  declaration,
				declarations: []*ast.Node{declaration},
				symbol:       memberSymbol,
			})
		}
	}
}

// bindingIsAnEnumName answers whether the SHADOWED binding is an enum declaration, which picks the
// message.
//
// The enum message is chosen by what is being shadowed, not by what is doing the shadowing, and the
// distinction is the whole content of the second message: when the outer binding is an enum, a
// member initializer naming it resolves to the shadowing member rather than to the enum, which is a
// different failure from an ordinary shadow. Upstream reads this off the SHADOWED variable's
// definitions (`shadowed.defs.some(def => def.type === TSEnumName)`), and reading it off the inner
// binding instead renders every enum finding with the ordinary message.
func bindingIsAnEnumName(binding shadowBinding) bool {
	if binding.declaration != nil && binding.declaration.Kind == ast.KindEnumDeclaration {
		return true
	}
	// A merged symbol carries several declarations and upstream asks whether ANY of them is an
	// enum, so a namespace merged onto an enum still selects the enum message.
	if binding.symbol == nil {
		return false
	}
	for _, declaration := range binding.symbol.Declarations {
		if declaration != nil && declaration.Kind == ast.KindEnumDeclaration {
			return true
		}
	}
	return false
}

// expressionNameIsDirectInitializerOf reproduces upstream's isOnInitializer.
//
// `var a = function a() {};` is clean because the inner name IS the thing being assigned to the
// outer one, so no read can be confused: every reference to `a` inside the function refers to the
// same function the outer name refers to. `var a = wrap(function a() {})` reports, because the
// outer `a` is the wrapper's RESULT and the inner name is a different value.
//
// Upstream unwraps through `||` and a conditional's branches, so `var a = foo || function a() {}`
// and `var a = foo ? function a() {} : bar` are clean too. Both are in the default corpus.
//
// Parenthesized forms are unwrapped as well. Our parser keeps KindParenthesizedExpression, which
// upstream's folds away, so without this a parenthesized initializer would fail the identity test
// and report where upstream is silent. Unwrapped in a loop because parentheses nest.
func expressionNameIsDirectInitializerOf(inner shadowBinding, outer shadowBinding) bool {
	// Only an expression NAME qualifies. A function or class DECLARATION shadowing an outer binding
	// is a real shadow whatever the initializer says.
	if inner.symbol != nil || inner.declaration == nil {
		return false
	}
	if inner.declaration.Kind != ast.KindFunctionExpression &&
		inner.declaration.Kind != ast.KindClassExpression {
		return false
	}

	initializer := initializerAssignedTo(outer.identifier)
	if initializer == nil {
		return false
	}

	return unwrapToAssignedExpression(inner.declaration) == initializer
}

// initializerAssignedTo returns the expression assigned to a binding identifier.
//
// Covers a variable declarator (`var a = X`) and a default, which upstream reads as an
// AssignmentPattern and our parser keeps on the parameter (`function f(a = X)`) or on the binding
// element (`const { a = X } = o`). Nil for anything else.
func initializerAssignedTo(identifier *ast.Node) *ast.Node {
	if identifier == nil || identifier.Parent == nil {
		return nil
	}
	switch identifier.Parent.Kind {
	case ast.KindVariableDeclaration:
		return identifier.Parent.AsVariableDeclaration().Initializer
	case ast.KindParameter:
		return identifier.Parent.AsParameterDeclaration().Initializer
	case ast.KindBindingElement:
		return identifier.Parent.AsBindingElement().Initializer
	}
	return nil
}

// unwrapToAssignedExpression walks outward from an expression through the wrappers that keep it the
// value being assigned: parentheses, both sides of `&&` and `||`, and a conditional's branches.
//
// Upstream's unwrapExpression does the same for the logical and conditional forms. Parentheses are
// added because our parser keeps a node upstream's does not have.
func unwrapToAssignedExpression(node *ast.Node) *ast.Node {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
		case ast.KindBinaryExpression:
			operator := parent.AsBinaryExpression().OperatorToken
			if operator == nil {
				return current
			}
			if operator.Kind != ast.KindAmpersandAmpersandToken &&
				operator.Kind != ast.KindBarBarToken &&
				operator.Kind != ast.KindQuestionQuestionToken {
				return current
			}
		case ast.KindConditionalExpression:
			// The test position is not a value the conditional evaluates to, so a name there is
			// not the thing being assigned.
			if parent.AsConditionalExpression().Condition == current {
				return current
			}
		default:
			return current
		}
		current = parent
	}
	return current
}

// bindingIsInTemporalDeadZoneOf reproduces upstream's isInTdz, for every hoist setting but `all`,
// which never asks.
//
// An inner binding written textually BEFORE the outer one it would shadow is only a shadow if the
// outer declaration hoists to the top of its scope under the setting. An outer `var`, `let`, `const`
// or class never does, so at the inner binding's position it is not yet in scope and the inner one
// shadows nothing.
//
//	{ let a; } function a() {}    reports under the default, the function hoists
//	{ let a; } var a;             clean
//
// A builtin global has no identifier and is never in a dead zone, as upstream's has no name range.
func bindingIsInTemporalDeadZoneOf(inner shadowBinding, outer shadowBinding, hoist string) bool {
	if inner.identifier == nil || outer.identifier == nil {
		return false
	}

	// Not in the dead zone at all when the inner binding is written at or after the outer one.
	//
	// A mutant rewriting this to compare against the outer identifier's END rather than its start
	// SURVIVES, and it is equivalent rather than untested. Two distinct identifiers cannot
	// partially overlap, so the inner one's end is never strictly inside the outer one's extent and
	// the two spellings can only differ there. Measured rather than argued: the condition was
	// instrumented to panic whenever the two disagreed, and across every case in this rule's test
	// file, plus eight hand-built shapes where an inner binding sits inside an outer declaration's
	// name range, it never fired.
	//
	// Upstream's spelling is kept because it is upstream's, not because a fixture separates them.
	if inner.identifier.End() >= outer.identifier.Pos() {
		return false
	}

	if outer.declaration == nil {
		return true
	}

	// Upstream's two hoisted sets, read off the outer binding's first declaration. A bodiless
	// function declaration is upstream's TSDeclareFunction, which neither set holds, so an overload
	// set or a `declare function` does not hoist; measured on the installed rule.
	hoistsAsFunction := outer.declaration.Kind == ast.KindFunctionDeclaration && outer.declaration.Body() != nil
	hoistsAsType := outer.declaration.Kind == ast.KindInterfaceDeclaration ||
		outer.declaration.Kind == ast.KindTypeAliasDeclaration
	switch hoist {
	case "functions":
		return !hoistsAsFunction
	case "types":
		return !hoistsAsType
	case "functions-and-types":
		return !hoistsAsFunction && !hoistsAsType
	}
	// never
	return true
}

// scopeIsInsideGlobalAugmentation answers whether a container sits inside `declare global { ... }`.
//
// Upstream skips those scopes outright: a global augmentation is by construction adding to an outer
// scope, so every name in it looks like a shadow and none of them is one.
func scopeIsInsideGlobalAugmentation(container *ast.Node) bool {
	for current := container; current != nil; current = current.Parent {
		if current.Kind != ast.KindModuleDeclaration {
			continue
		}
		if ast.IsGlobalScopeAugmentation(current) {
			return true
		}
	}
	return false
}

// nodeContains answers whether outer strictly contains inner by source range.
func nodeContains(outer *ast.Node, inner *ast.Node) bool {
	if outer == nil || inner == nil {
		return false
	}
	return outer.Pos() <= inner.Pos() && inner.End() <= outer.End()
}

// bindingCrossesTheTypeValueBoundary reproduces upstream's isTypeValueShadow, which
// ignoreTypeValueShadow (default true) turns on.
//
// Under it, a binding that is purely a type shadowing one that
// is purely a value (or the reverse) is not reported. The two occupy different declaration spaces:
// a type position resolves only to types and an expression position only to values, so neither
// reference can land on the wrong one and there is nothing for a reader to get wrong.
//
//	const x = 1; { type x = string; }   clean, a type over a value
//	var a = 3; function b() { var a = 10; }   reports, both are values
//
// In the SAME scope the binder merges the two into one symbol and the question never arises, which
// is why `const x = 1; type x = string;` needs nothing here. Across scopes they are two symbols and
// this test is what separates them.
//
// Upstream reads a boolean pair, isValueVariable and isTypeVariable. Here the symbol carries
// TypeScript's own SymbolFlags, which say the same thing more precisely, so the question is asked
// of the flags directly.
func bindingCrossesTheTypeValueBoundary(inner shadowBinding, outer shadowBinding) bool {
	innerIsValue, innerIsType := bindingDeclarationSpaces(inner)
	outerIsValue, outerIsType := bindingDeclarationSpaces(outer)

	// A binding in neither space says nothing either way.
	if !innerIsValue && !innerIsType {
		return false
	}
	if !outerIsValue && !outerIsType {
		return false
	}

	// The question is whether the two OVERLAP, not whether either is pure. A class and an enum
	// occupy BOTH spaces, and the direction of the comparison is what decides them:
	//
	//	class A {} ... type A = string     clean, the inner type shadows only the type half
	//	class A {} ... const A = 1         reports, the inner value shadows the value half
	//	type A = string ... class A {}     clean upstream, measured
	//	const A = 1 ... class A {}         reports
	//
	// Measured against the installed rule at 8.67.0 rather than reasoned about, because an earlier
	// version of this function bailed out whenever either side occupied both spaces and got four
	// of those six rows wrong while passing every imported fixture. The corpus writes no
	// class-over-type pair at all.
	//
	// The rule that reproduces all six: they collide only where the INNER binding's spaces meet the
	// outer's. A pure type inner meets a class outer in the type space, and upstream is silent
	// there, so the collision test is on the VALUE space alone, which is the space a reference in
	// expression position resolves through.
	return innerIsValue != outerIsValue
}

// bindingDeclarationSpaces answers which declaration spaces a binding occupies.
//
// An expression-name binding carries no symbol, because it is collected from AST shape rather than
// from a table, and a function or class expression name is always a value.
func bindingDeclarationSpaces(binding shadowBinding) (isValue bool, isType bool) {
	if binding.symbol == nil {
		return true, false
	}
	flags := binding.symbol.Flags

	// An alias is an import, and it stands for whatever it imports, so it is treated as occupying
	// both spaces rather than neither. Upstream reaches the same answer through isTypeImport: a
	// plain import counts as a value, and only a TYPE-only one counts as a type alone.
	if flags&ast.SymbolFlagsAlias != 0 {
		if bindingIsTypeOnlyImport(binding) {
			return false, true
		}
		return true, true
	}

	return flags&ast.SymbolFlagsValue != 0, flags&ast.SymbolFlagsType != 0
}

// bindingIsTypeOnlyImport answers whether an imported binding came in through `import type` or
// through an inline `{ type Foo }` specifier.
func bindingIsTypeOnlyImport(binding shadowBinding) bool {
	declaration := binding.declaration
	if declaration == nil {
		return false
	}
	if declaration.Kind == ast.KindImportSpecifier {
		if declaration.AsImportSpecifier().IsTypeOnly {
			return true
		}
	}
	// `import type { Foo } from 'bar'` carries the marker on the CLAUSE rather than on the
	// specifier, so the specifier reads false and the clause has to be consulted.
	for current := declaration; current != nil; current = current.Parent {
		if current.Kind == ast.KindImportClause {
			return current.AsImportClause().PhaseModifier == ast.KindTypeKeyword
		}
		if current.Kind == ast.KindImportDeclaration {
			break
		}
	}
	return false
}

// bindingIsAFunctionTypeParameterName reproduces upstream's isFunctionTypeParameterNameValueShadow,
// which ignoreFunctionTypeParameterNameValueShadow (default true) turns on.
//
// A parameter name written inside a function TYPE binds nothing. It exists so the signature reads
// well and so an editor can show a label; no expression can reference it, so it cannot shadow.
//
//	type Args = 1; function foo<T extends (Args: any) => void>(arg: T) {}   clean
//
// Only over a VALUE: a name in a function type shadowing a type is a real shadow, which the
// type-and-value carve-out decides before this one.
//
// Upstream tests every definition's node against seven type-position function shapes. A parameter's
// definition node is its function, so the test here is the parameter's own parent; the shapes include
// the bodiless forms, TSDeclareFunction (an overload signature or a `declare function`) and
// TSEmptyBodyFunctionExpression (an abstract, overload or `declare class` method). A bodiless
// function's NAME has that same function as its definition node, so a name declared only by
// signatures counts too, while an overload set with its implementation does not: measured on the
// installed rule.
func bindingIsAFunctionTypeParameterName(ctx rule.Context, binding shadowBinding, shadowed shadowBinding) bool {
	if shadowed.global {
		if !builtinGlobalIsValue(ctx.TypeChecker, ctx.Program, shadowed.name) {
			return false
		}
	} else if shadowedIsValue, _ := bindingDeclarationSpaces(shadowed); !shadowedIsValue {
		return false
	}
	if len(binding.declarations) == 0 {
		return false
	}
	for _, declaration := range binding.declarations {
		var function *ast.Node
		switch declaration.Kind {
		case ast.KindParameter:
			function = declaration.Parent
		case ast.KindFunctionDeclaration:
			function = declaration
		default:
			return false
		}
		if function == nil || !isBodilessOrTypePositionFunction(function) {
			return false
		}
	}
	return true
}

// isBodilessOrTypePositionFunction reports whether a function-like node is one of upstream's seven
// allowedFunctionVariableDefTypes.
func isBodilessOrTypePositionFunction(function *ast.Node) bool {
	switch function.Kind {
	case ast.KindFunctionType,
		ast.KindConstructorType,
		ast.KindCallSignature,
		ast.KindConstructSignature,
		ast.KindMethodSignature:
		return true
	case ast.KindFunctionDeclaration, ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor:
		// An accessor in an interface is upstream's TSMethodSignature, and in a declare class a
		// TSEmptyBodyFunctionExpression; bodiless either way.
		return function.Body() == nil
	}
	return false
}

// bindingIsAnIndexSignatureParameter reports whether a declaration is the parameter of an index
// signature, which our binder gives a locals table and upstream's scope manager never declares.
func bindingIsAnIndexSignatureParameter(declaration *ast.Node) bool {
	return declaration.Kind == ast.KindParameter && declaration.Parent != nil &&
		declaration.Parent.Kind == ast.KindIndexSignature
}

// bindingIsDeclaredInADefinitionFile reproduces upstream's isDeclareInDTSFile: in a `.d.ts`, `.d.cts`
// or `.d.mts` file, a binding any of whose declarations carries `declare` itself is skipped.
//
// The four definition types upstream reads are a variable (the `declare` is on its statement), a
// class, an enum and a namespace. A declaration inside a `declare namespace` without its own `declare`
// is not one, and still reports; measured on the installed rule.
func bindingIsDeclaredInADefinitionFile(ctx rule.Context, binding shadowBinding) bool {
	fileName := strings.ToLower(ctx.SourceFile.FileName().AsString())
	if !strings.HasSuffix(fileName, ".d.ts") && !strings.HasSuffix(fileName, ".d.cts") &&
		!strings.HasSuffix(fileName, ".d.mts") {
		return false
	}
	for _, declaration := range binding.declarations {
		switch declaration.Kind {
		case ast.KindVariableDeclaration, ast.KindBindingElement:
			for current := declaration.Parent; current != nil; current = current.Parent {
				if current.Kind == ast.KindVariableStatement {
					if ast.HasSyntacticModifier(current, ast.ModifierFlagsAmbient) {
						return true
					}
					break
				}
				if current.Kind != ast.KindVariableDeclarationList && current.Kind != ast.KindObjectBindingPattern &&
					current.Kind != ast.KindArrayBindingPattern && current.Kind != ast.KindBindingElement &&
					current.Kind != ast.KindVariableDeclaration {
					break
				}
			}
		case ast.KindClassDeclaration, ast.KindEnumDeclaration, ast.KindModuleDeclaration:
			if ast.HasSyntacticModifier(declaration, ast.ModifierFlagsAmbient) {
				return true
			}
		}
	}
	return false
}

// bindingIsAStaticMethodGenericOverAClassGeneric reproduces upstream's isGenericOfAStaticMethodShadow.
//
// A class's type parameters are not in scope in a static member, so a static method's own `T` shadows
// nothing, but the class's `T` encloses it by range. Upstream handles it by hand for the same reason.
// The method may be bodiless, as upstream's TSEmptyBodyFunctionExpression is.
func bindingIsAStaticMethodGenericOverAClassGeneric(binding shadowBinding, shadowed shadowBinding) bool {
	inner := binding.declaration
	outer := shadowed.declaration
	if inner == nil || outer == nil || inner.Kind != ast.KindTypeParameter || outer.Kind != ast.KindTypeParameter {
		return false
	}
	if inner.Parent == nil || inner.Parent.Kind != ast.KindMethodDeclaration || !ast.IsStatic(inner.Parent) {
		return false
	}
	return outer.Parent != nil &&
		(outer.Parent.Kind == ast.KindClassDeclaration || outer.Parent.Kind == ast.KindClassExpression)
}

// bindingIsInitializedThrough reproduces upstream's isInitPatternNode, for ignoreOnInitialization.
//
// The inner binding lives in a function expression or arrow whose enclosing scope is the shadowed
// binding's own, the function sits inside a call, and that call ends inside the shadowed binding's
// initializer, default, or `for in`/`for of` right side. The outer binding is presumably not
// initialized yet when the callback runs, so the inner name cannot mean it:
//
//	const a = wrap(() => { let a; });   clean under the option
//	const a = new Wrap(() => { let a; });   reports, upstream finds only a call
//	const a = wrap(class { m() { let a; } });   reports, the method's outer scope is the class
//
// A method is a function expression to upstream, which nests it under a MethodDefinition.
func bindingIsInitializedThrough(ctx rule.Context, binding shadowBinding, shadowed shadowBinding, shadowedContainer *ast.Node) bool {
	if shadowed.global || shadowed.symbol == nil || shadowed.identifier == nil || shadowedContainer == nil {
		return false
	}
	function := variableScopeFunctionOf(binding)
	if function == nil || enclosingScopeContainerOf(function) != shadowedContainer {
		return false
	}

	var call *ast.Node
	for current := function.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindCallExpression {
			call = current
			break
		}
	}
	if call == nil {
		return false
	}
	location := call.End()
	spans := func(node *ast.Node) bool {
		return node != nil && rule.TokenRange(ctx.SourceFile, node).Pos() <= location && location <= node.End()
	}

	for current := shadowed.identifier; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindVariableDeclaration:
			if spans(current.AsVariableDeclaration().Initializer) {
				return true
			}
			if list := current.Parent; list != nil && list.Kind == ast.KindVariableDeclarationList && list.Parent != nil &&
				(list.Parent.Kind == ast.KindForInStatement || list.Parent.Kind == ast.KindForOfStatement) {
				return spans(list.Parent.AsForInOrOfStatement().Expression)
			}
			return false
		case ast.KindBindingElement:
			if spans(current.AsBindingElement().Initializer) {
				return true
			}
		case ast.KindParameter:
			// Upstream's AssignmentPattern, whose parent is the function, where its walk stops.
			return spans(current.AsParameterDeclaration().Initializer)
		case ast.KindArrowFunction, ast.KindCatchClause, ast.KindClassDeclaration, ast.KindClassExpression,
			ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindImportDeclaration:
			return false
		}
	}
	return false
}

// variableScopeFunctionOf returns the function expression, arrow or method whose scope upstream's
// `variable.scope.variableScope` is for this binding, or nil when that scope is not one.
//
// The walk goes outward from the binding's identifier to the nearest variable scope: a function, a
// class static block, a class field's initializer, a namespace or the file. A function or namespace
// whose NAME the identifier is binds it in the scope outside, so that one is passed over.
func variableScopeFunctionOf(binding shadowBinding) *ast.Node {
	start := binding.identifier
	if start == nil {
		return nil
	}
	previous := start
	for current := start.Parent; current != nil; previous, current = current, current.Parent {
		switch current.Kind {
		case ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindModuleDeclaration:
			if current.Name() == start {
				continue
			}
		case ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor,
			ast.KindSetAccessor, ast.KindClassStaticBlockDeclaration, ast.KindSourceFile:
		case ast.KindPropertyDeclaration:
			if current.Initializer() != previous {
				continue
			}
		default:
			continue
		}
		switch current.Kind {
		case ast.KindFunctionExpression, ast.KindArrowFunction:
			return current
		case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			if current.Body() != nil {
				return current
			}
		}
		return nil
	}
	return nil
}

// enclosingScopeContainerOf returns the scope a function sits in, the container the binder declares a
// block-scoped name at the function's position into.
//
// A function body's block holds no names of its own (the function does), so it is passed over, and
// so is a static block's body. Every other locals container is a scope, the class among them, which
// is a method's enclosing scope as upstream's class scope is.
func enclosingScopeContainerOf(function *ast.Node) *ast.Node {
	for current := function.Parent; current != nil; current = current.Parent {
		if !ast.IsLocalsContainer(current) {
			continue
		}
		if current.Kind == ast.KindBlock && current.Parent != nil &&
			(ast.IsFunctionLike(current.Parent) || current.Parent.Kind == ast.KindClassStaticBlockDeclaration) {
			continue
		}
		return current
	}
	return nil
}

// bothBindingsAreInferredTypeParameters answers whether two bindings are each an `infer` in a
// conditional type, which upstream does not treat as shadowing.
//
// Upstream scopes an `infer T` to the `extends` clause of the conditional type that introduces it,
// so two `infer T` in a chain of conditionals are siblings and never see each other. Our binder
// puts each on the conditional type it belongs to, and a chained conditional nests inside its
// parent's false branch BY RANGE, so the general containment search finds the outer one and would
// report a shadow upstream is silent about.
//
// # The narrowness is measured, not assumed
//
// The obvious repair is to exempt `infer` from shadowing altogether, and it is wrong. Driven
// against the installed rule at 8.67.0, an `infer` DOES report when what it shadows is an ordinary
// outer type:
//
//	type T = string; type A<F> = F extends (a: infer T) => any ? T : never;   REPORTS
//	type A<T> = T extends (a: infer T) => any ? T : never;                    REPORTS
//	type A<F> = F extends (a: Array<infer T>) => any ? T[] : F extends (...a: infer T) => any ? T : never;   clean
//
// So the exemption is specifically infer-over-infer, and both sides are tested. A blanket exemption
// would silence the first two, which no imported fixture would catch because upstream's default
// corpus writes only the third.
//
// A third measured row is `const T = 1;` shadowed by an `infer T`, which is clean, but that one is
// already handled upstream of here by the type-versus-value carve-out rather than by this test.
func bothBindingsAreInferredTypeParameters(inner shadowBinding, outer shadowBinding) bool {
	return bindingIsAnInferredTypeParameter(inner) && bindingIsAnInferredTypeParameter(outer)
}

// bindingIsAnInferredTypeParameter answers whether a binding was introduced by `infer`.
func bindingIsAnInferredTypeParameter(binding shadowBinding) bool {
	declaration := binding.declaration
	if declaration == nil || declaration.Kind != ast.KindTypeParameter {
		return false
	}
	return declaration.Parent != nil && declaration.Parent.Kind == ast.KindInferType
}

// bindingIsExternalDeclarationMerging reproduces upstream's isExternalDeclarationMerging.
//
// A type-only import of a name, augmented inside a `declare module` naming the SAME module, is
// declaration merging rather than shadowing: the interface or alias being declared inside the
// augmentation is the very thing the import refers to, so a reference resolving to it is correct.
//
//	import type { Foo } from 'bar'; declare module 'bar'  { interface Foo {} }   clean
//	import type { Foo } from 'bar'; declare module 'baz'  { interface Foo {} }   reports
//	interface Foo {};               declare module 'bar'  { interface Foo {} }   reports
//
// The module name is what separates the first two, and the outer being an import is what separates
// the first from the third. Both are checked, because dropping either one silently converts one of
// upstream's own reporting cases into silence.
func bindingIsExternalDeclarationMerging(ctx rule.Context, container *ast.Node,
	inner shadowBinding, outer shadowBinding) bool {
	if container == nil || container.Kind != ast.KindModuleDeclaration {
		return false
	}

	// The inner declaration has to be a type declaration, which is what can merge with an imported
	// type. A value declared in an augmentation shadows the import for real.
	if inner.declaration == nil {
		return false
	}
	switch inner.declaration.Kind {
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
	default:
		return false
	}

	// The outer binding has to be an IMPORT, not merely any binding of that name. A local
	// `interface Foo {}` augmented in `declare module 'bar'` reports, because nothing connects the
	// two, and that is one of upstream's own default invalid cases.
	//
	// It does NOT have to be a type-only import, and requiring that was wrong. Measured against the
	// installed rule at 8.67.0: a PLAIN `import { Foo } from 'bar'` augmented in `declare module
	// 'bar'` is clean, for both an interface and a type alias. Upstream reaches that answer by a
	// different route, through its type-versus-value carve-out rather than through this one, but
	// the verdict is the verdict and this is where our binder puts the question.
	if outer.symbol == nil || outer.symbol.Flags&ast.SymbolFlagsAlias == 0 {
		return false
	}

	// And the augmented module has to be the one that import came from.
	moduleName := container.Name()
	if moduleName == nil || moduleName.Kind != ast.KindStringLiteral {
		return false
	}
	importedFrom := moduleSpecifierTextOf(outer.declaration)
	return importedFrom != "" && importedFrom == moduleName.Text()
}

// moduleSpecifierTextOf returns the module a binding was imported from, or the empty string.
func moduleSpecifierTextOf(declaration *ast.Node) string {
	for current := declaration; current != nil; current = current.Parent {
		if current.Kind != ast.KindImportDeclaration {
			continue
		}
		specifier := current.AsImportDeclaration().ModuleSpecifier
		if specifier == nil || specifier.Kind != ast.KindStringLiteral {
			return ""
		}
		return specifier.Text()
	}
	return ""
}

// noShadowLineAndColumnOf renders a node's one-based line and column, which the message quotes.
func noShadowLineAndColumnOf(ctx rule.Context, node *ast.Node) (int, int) {
	line, column := scanner.GetLineAndCharacterOfPosition(ctx.SourceFile,
		rule.TokenRange(ctx.SourceFile, node).Pos())
	return line + 1, column + 1
}
