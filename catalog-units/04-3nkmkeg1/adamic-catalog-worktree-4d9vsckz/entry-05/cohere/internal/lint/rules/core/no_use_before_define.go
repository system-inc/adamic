package core

import (
	"encoding/json"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/decorators"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUseBeforeDefineOptions configures which kinds of binding the rule judges.
//
// Every field is a pointer, and that is load-bearing rather than tidy. Six of the seven options
// default to TRUE upstream, so a plain bool would put upstream's default on the wrong side of the
// zero value: a rule reached through the config layer with no options written would see false for
// all of them and switch off almost the entire check while still registering on every file. That is
// the shape the brief calls a rule that passes every fixture and lints nothing, and a pointer is
// what keeps "absent" distinguishable from "written as false".
type NoUseBeforeDefineOptions struct {
	// Functions judges references to function declarations. Defaults to TRUE.
	//
	// Separable because a function declaration is fully hoisted: calling one above its definition
	// works, so a codebase that deliberately puts helpers below their callers is not making the
	// mistake this rule exists to catch. That is why upstream ships the `"nofunc"` string as a
	// shorthand for turning this one option off and nothing else.
	Functions *bool `json:"functions"`

	// Classes judges references to class declarations. Defaults to TRUE.
	Classes *bool `json:"classes"`

	// Variables judges references to variable declarations. Defaults to TRUE.
	Variables *bool `json:"variables"`

	// AllowNamedExports stops reporting a name listed in an `export { a }` clause above its
	// declaration. Defaults to FALSE.
	//
	// The odd one out: it is the only option here whose default is off, because an export clause
	// naming a binding declared below it is a real forward reference by every other measure and
	// upstream only exempts it on request.
	AllowNamedExports *bool `json:"allowNamedExports"`

	// Enums judges references to TypeScript enum declarations. Defaults to TRUE.
	Enums *bool `json:"enums"`

	// Typedefs judges references to TypeScript type aliases and interfaces. Defaults to TRUE.
	Typedefs *bool `json:"typedefs"`

	// IgnoreTypeReferences stops reporting a name used in type position. Defaults to TRUE.
	//
	// Defaulted on because a type reference is erased before anything runs, so a type annotation
	// naming a type declared below it cannot fail at runtime the way a value reference can. The
	// forward reference is real and the consequence is not, which is why upstream separates this
	// from `typedefs`: one asks whether to judge the DECLARATION kind, the other whether to judge
	// the USE position.
	IgnoreTypeReferences *bool `json:"ignoreTypeReferences"`
}

// noUseBeforeDefineSettings is the decoded form, with every default already applied.
type noUseBeforeDefineSettings struct {
	functions            bool
	classes              bool
	variables            bool
	allowNamedExports    bool
	enums                bool
	typedefs             bool
	ignoreTypeReferences bool
}

// defaultNoUseBeforeDefineSettings mirrors upstream's `meta.defaultOptions`.
func defaultNoUseBeforeDefineSettings() noUseBeforeDefineSettings {
	return noUseBeforeDefineSettings{
		functions:            true,
		classes:              true,
		variables:            true,
		allowNamedExports:    false,
		enums:                true,
		typedefs:             true,
		ignoreTypeReferences: true,
	}
}

// resolve applies each written option over the defaults, leaving an absent key alone.
func (options NoUseBeforeDefineOptions) resolve() noUseBeforeDefineSettings {
	settings := defaultNoUseBeforeDefineSettings()
	if options.Functions != nil {
		settings.functions = *options.Functions
	}
	if options.Classes != nil {
		settings.classes = *options.Classes
	}
	if options.Variables != nil {
		settings.variables = *options.Variables
	}
	if options.AllowNamedExports != nil {
		settings.allowNamedExports = *options.AllowNamedExports
	}
	if options.Enums != nil {
		settings.enums = *options.Enums
	}
	if options.Typedefs != nil {
		settings.typedefs = *options.Typedefs
	}
	if options.IgnoreTypeReferences != nil {
		settings.ignoreTypeReferences = *options.IgnoreTypeReferences
	}
	return settings
}

// DecodeNoUseBeforeDefineOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto` for the reason above, and with one extra shape
// the generic helper could not express at all: upstream's schema is a `oneOf`, so the option may
// arrive as the bare STRING `"nofunc"` instead of an object. That spelling predates the object form
// and means exactly `{"functions": false}`, so it is normalised here rather than carried through the
// rule as a second code path. Any other string means the same as no options at all, which is what
// upstream's `parseOptions` does with it.
func DecodeNoUseBeforeDefineOptions(raw []byte) (any, error) {
	var options NoUseBeforeDefineOptions
	if len(raw) == 0 {
		return options, nil
	}

	// The string form is tried first because it is not an object and would fail the struct
	// unmarshal, which the config layer would surface as a configuration error rather than as the
	// legal spelling it is.
	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		if asString == "nofunc" {
			functionsOff := false
			options.Functions = &functionsOff
		}
		return options, nil
	}

	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}

// resolveNoUseBeforeDefineSettings turns whatever the dispatch layer handed us into settings.
//
// The nil arm is not defensive padding. A rule configured as a bare `"error"` is handed nil, and a
// type assertion on nil yields the zero value, which for this rule means six of seven options off.
// The brief records a rule that shipped that way: 3,407 registrations, zero findings, every fixture
// green, because every fixture reached the rule through the decoder and nothing tested the path the
// live config actually uses.
//
// The decoder returns the value, and rule.OptionsAs panics on any other type rather than defaulting
// in silence, which is how a mismatch hid in unified-signatures (#qmvkf83). The pointer arm this
// replaced had no caller: nothing hands this rule a pointer.
func resolveNoUseBeforeDefineSettings(options any) noUseBeforeDefineSettings {
	typed, configured := rule.OptionsAs[NoUseBeforeDefineOptions](options)
	if !configured {
		return defaultNoUseBeforeDefineSettings()
	}
	return typed.resolve()
}

// useBeforeDefineMessage renders the finding for one read, true to the kind of binding it reads.
//
// One fixed message used to tell every finding that the read "throws at runtime" or "yields
// `undefined`". Measured on ahra, 10 of the 12 findings read a function declaration, which is
// hoisted together with its body: the read works, and the message described a crash that cannot
// happen. The same message told a read inside a function written above a `const` that it throws,
// when it throws only if that function is called before the declaration runs. So the consequence is
// chosen by what was declared and by whether the read runs now or later, and the binding is named.
func useBeforeDefineMessage(identifier *ast.Node, declaration *ast.Node) rule.Message {
	name := "'" + identifier.Text() + "'"
	const repair = " Move the declaration above its first use."
	const readingOrder = "the cost is reading order: the reader meets the name before learning what it is, " +
		"and has to jump down the file to find out."

	var description string
	switch kind := useBeforeDefineDeclarationKind(declaration); {
	case kind == useBeforeDefineType || isTypeReferencePosition(identifier):
		description = name + " is used as a type above the line that declares it. A type is erased " +
			"before the program runs, so nothing fails at runtime; " + readingOrder

	case kind == useBeforeDefineFunction:
		description = name + " is read above the function declaration that defines it. A function " +
			"declaration is hoisted together with its body, so the read works at runtime; " + readingOrder

	case kind == useBeforeDefineClass || kind == useBeforeDefineVariable && useBeforeDefineIsBlockScoped(declaration):
		keyword := "`" + useBeforeDefineKeyword(declaration) + "`"
		if useBeforeDefineRunsLater(identifier, declaration) {
			description = name + " is read inside a function written above the " + keyword + " that " +
				"declares it. That is safe only while nothing calls the function before the declaration " +
				"runs: an earlier call, such as one made while the module is still loading, reaches the " +
				"binding in its temporal dead zone and throws a ReferenceError."
		} else {
			description = name + " is read above the " + keyword + " that declares it. The binding " +
				"exists from the top of its block but cannot be touched until its declaration runs, so " +
				"this read lands in the temporal dead zone and throws a ReferenceError."
		}

	case kind == useBeforeDefineVariable || kind == useBeforeDefineEnum:
		keyword := "`" + useBeforeDefineKeyword(declaration) + "`"
		if useBeforeDefineRunsLater(identifier, declaration) {
			description = name + " is read inside a function written above the " + keyword + " that " +
				"declares it. The binding is hoisted without its value, so a call made before the " +
				"declaration runs reads `undefined` silently, and the failure surfaces later and " +
				"somewhere else."
		} else {
			description = name + " is read above the " + keyword + " that declares it. The binding is " +
				"hoisted without its value, so this read silently yields `undefined`, and the failure " +
				"surfaces later and somewhere else."
		}

	default:
		description = name + " is read above the line that declares it, so the reader meets the name " +
			"before learning what it is."
	}

	return rule.Message{Id: "usedBeforeDefined", Description: description + repair}
}

// useBeforeDefineIsBlockScoped reports whether a variable's declaration list is `let`, `const` or
// a `using` form, which have a temporal dead zone, rather than `var`, which is hoisted with
// `undefined`.
func useBeforeDefineIsBlockScoped(declaration *ast.Node) bool {
	list := useBeforeDefineDeclarationList(declaration)
	return list != nil && list.Flags&ast.NodeFlagsBlockScoped != 0
}

// useBeforeDefineDeclarationList climbs from a variable declaration or a destructured element to
// the declaration list that carries the `var`, `let` or `const`.
func useBeforeDefineDeclarationList(declaration *ast.Node) *ast.Node {
	for current := declaration; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindVariableDeclarationList:
			return current
		case ast.KindVariableDeclaration, ast.KindBindingElement,
			ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
			continue
		}
		return nil
	}
	return nil
}

// useBeforeDefineKeyword is the word the source used to declare the binding.
func useBeforeDefineKeyword(declaration *ast.Node) string {
	switch declaration.Kind {
	case ast.KindClassDeclaration, ast.KindClassExpression:
		return "class"
	case ast.KindEnumDeclaration:
		return "enum"
	}
	list := useBeforeDefineDeclarationList(declaration)
	if list == nil {
		return "var"
	}
	// The flags are compared as a set, because `await using` is spelled `Const | Using`.
	switch list.Flags & ast.NodeFlagsBlockScoped {
	case ast.NodeFlagsConst:
		return "const"
	case ast.NodeFlagsLet:
		return "let"
	case ast.NodeFlagsUsing:
		return "using"
	case ast.NodeFlagsAwaitUsing:
		return "await using"
	}
	return "var"
}

// useBeforeDefineRunsLater reports whether the read sits inside a function that does not also
// contain the declaration, so it runs when that function is called rather than where it is written.
func useBeforeDefineRunsLater(identifier *ast.Node, declaration *ast.Node) bool {
	for current := identifier.Parent; current != nil; current = current.Parent {
		if !ast.IsFunctionLike(current) {
			continue
		}
		for container := declaration.Parent; container != nil; container = container.Parent {
			if container == current {
				return false
			}
		}
		return true
	}
	return false
}

// NoUseBeforeDefine flags a reference to a binding that appears above the binding's declaration.
//
//	valid:   var a = 10; alert(a);
//	valid:   function foo() { foo(); }
//	valid:   class A {} new A();
//	valid:   var a = 0, b = a;
//	valid:   class C { method() { C; } }
//	invalid: a++; var a = 19;
//	invalid: new A(); class A {};
//	invalid: {a; let a = 1}
//	invalid: var f = () => a; var a;
//
// Upstream is a thin caller over `eslint-scope`: it walks `scope.references`, reads
// `reference.resolved` for the variable each one binds to, and compares source positions. We have no
// reference index, so the reconstruction is the port. The replacement is the checker, which answers
// the same question a different way. `GetSymbolAtLocation` on an identifier gives the symbol, and
// the symbol carries its declarations, so "which binding is this" and "where is it declared" are
// both available without a scope graph.
//
// What does NOT survive that substitution is upstream's notion of an execution context, and it is
// most of the rule's subtlety rather than a detail. A position comparison alone calls
// `class C { method() { C; } }` a forward reference, because the `C` in the method body is written
// above nothing, since it is inside the very declaration it names. Upstream answers this with
// `isFromSeparateExecutionContext`: a method body runs when it is called, long after the class
// binding is initialized, so the reference is fine. A static block is the opposite, running during
// the class definition itself, so a reference there really can precede initialization. Those two
// live in the same class body and only the enclosing construct separates them, which is why this
// rule walks the parent chain rather than comparing offsets and stopping.
var NoUseBeforeDefine = rule.Rule{
	Name:             "no-use-before-define",
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		// Declared once for the file rather than once per listener. `NeedsTypeChecker` governs the
		// REGISTRATION path and says nothing about a Context built by hand, which
		// `TestNoRegisteredRuleCrashesOnAbsentOptionalNodes` does, so the declaration is not the
		// guard.
		if ctx.TypeChecker == nil {
			return nil
		}

		settings := resolveNoUseBeforeDefineSettings(options)

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				checkFileForUseBeforeDefine(ctx, node, settings)
			},
		}
	},
}

// checkFileForUseBeforeDefine walks every identifier in the file and reports the forward references.
//
// One walk rather than upstream's scope recursion, because the two are asking the same question
// through different indexes. Upstream visits every scope and reads the references it holds; we visit
// every identifier and resolve it. The set of (reference, binding) pairs considered is the same, and
// ours is the only one available without a scope graph.
func checkFileForUseBeforeDefine(ctx rule.Context, sourceFile *ast.Node, settings noUseBeforeDefineSettings) {
	cycles := newUseBeforeDefineCycles(ctx)
	var visit func(*ast.Node)
	visit = func(current *ast.Node) {
		if current == nil {
			return
		}

		if current.Kind == ast.KindIdentifier {
			checkIdentifierForUseBeforeDefine(ctx, current, settings, cycles)
		}

		current.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile)
}

// checkIdentifierForUseBeforeDefine judges one identifier occurrence.
func checkIdentifierForUseBeforeDefine(
	ctx rule.Context, identifier *ast.Node, settings noUseBeforeDefineSettings, cycles *useBeforeDefineCycles,
) {
	// A declaring name is not a reference to itself. Upstream gets this for free, because
	// eslint-scope never puts a declaration into `scope.references` in the first place; walking
	// identifiers instead means the exclusion has to be explicit, and without it every binding in
	// the file would trivially "reference" itself at its own position.
	if isDeclaringName(identifier) {
		return
	}

	// A JSX attribute NAME is not a reference to a binding. `<div key="x" />` names an attribute of
	// the element, the way `{ key: "x" }` names a property, and the checker happily resolves it to
	// whatever declaration the surrounding types associate with that attribute. Without this the
	// rule reports every attribute in the tree whose resolved declaration happens to sit later in
	// the file: 1,671 findings before the guard, with two of the three largest clusters being an
	// icon file full of `key=` and a component file full of ordinary props.
	//
	// Upstream never faces this because eslint-scope does not create a reference for a JSX attribute
	// name at all, so no imported fixture can see the difference. Measured against the installed
	// rule: `<div key="x" />` and `<div className="x" id="y" />` are both clean, while
	// `<Section />; function Section() {}` reports, which is the pair this guard has to preserve.
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindJsxAttribute &&
		identifier.Parent.Name() == identifier {
		return
	}

	// The PROPERTY NAME half of a renamed destructuring pattern is not a binding reference either.
	// In `const { sorted: sortedTypes } = f()`, `sorted` names a property of the right-hand value
	// and `sortedTypes` is the binding; only the second can be used before anything. Our parser
	// gives the property name a `PropertyName` slot on the binding element, so the two halves are
	// separable, and the checker resolves the property name to whatever declares that property,
	// which is frequently later in the file.
	//
	// `isDeclaringName` cannot cover this: it already answers true for the binding NAME of a binding
	// element, and the property name is a different slot on the same node. Measured against the
	// installed rule, which reports exactly one finding for each of these and never the property:
	//
	//	const { sorted: sortedTypes } = f(); function f() { ... }   1 finding, on sortedTypes
	//	const { a: b, c: d } = o; const o = { a: 1, c: 2 };         1 finding, on o
	//
	// Upstream's corpus writes `var {a = 0, b = a} = {}` and other defaults but never a RENAME, so
	// nothing imported can see this.
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindBindingElement &&
		identifier.Parent.AsBindingElement().PropertyName == identifier {
		return
	}

	// This file's declarations only: positions are compared below, and a declaration in another
	// file has no position here.
	declarations := rule.DeclarationsIn(ctx.SourceFile, useBeforeDefineResolveBinding(ctx, identifier))
	if len(declarations) == 0 {
		return
	}

	declaration := useBeforeDefineBindingDeclaration(declarations)
	if declaration == nil {
		return
	}

	// A declaration in another file cannot be a forward reference in this one. Upstream reaches the
	// same exclusion through `variable.defs.length === 0` for globals; ours has to compare files,
	// because the checker resolves an import to its declaration in the module it came from and that
	// declaration's position is meaningless against a position in this file.
	if ast.GetSourceFileOfNode(declaration) != ast.GetSourceFileOfNode(identifier) {
		return
	}

	declarationName := declaration.Name()
	if declarationName == nil {
		return
	}

	if !shouldCheckUseBeforeDefineReference(ctx, identifier, declaration, settings) {
		return
	}

	// The two ways a reference can be too early, and they are genuinely different questions.
	//
	// The first is position: the use ends before the declaring name does, so it is written above the
	// declaration. That is the whole rule for the ordinary cases.
	//
	// The second is initialization: a reference can sit BELOW the declaring name and still run
	// before the binding holds a value, because it is inside the initializer that produces that
	// value. `var a = a;` is the small version and `class C extends C {}` the large one. Position
	// cannot see either, since in both the use is to the right of the name.
	writtenAbove := identifier.End() < declarationName.End()
	duringInitialization := isEvaluatedDuringInitialization(identifier, declaration)

	if !writtenAbove && !duringInitialization {
		return
	}

	// A type reference is exempt from the initialization arm but not from the position arm, and the
	// asymmetry is upstream's own (`reference.identifier.parent.type !== "TSTypeReference"` sits on
	// the initialization side of its condition alone). It follows from what the two arms mean: a
	// type is erased, so it is never "evaluated during initialization" of anything, but it can still
	// be written above its declaration.
	if duringInitialization && !writtenAbove && isTypeReferencePosition(identifier) {
		return
	}

	// A forward reference inside a cycle no ordering can remove, that cannot run before its target
	// exists. See no_use_before_define_cycle.go.
	if writtenAbove && !duringInitialization && cycles.exempts(identifier, declaration) {
		return
	}

	ctx.ReportNode(identifier, useBeforeDefineMessage(identifier, declaration))
}

// useBeforeDefineBindingDeclaration picks which of a symbol's declarations the rule judges against.
//
// The EARLIEST by position, deliberately, and the question being answered is "when does this name
// become available", which is a different question from the one `no_unused_vars` asks of a merged
// symbol. A merged symbol here is ordinary rather than exotic: `interface Foo {}` above
// `function Foo() {}` is one symbol with two declarations, and so is an enum declared twice. The
// binding is usable from the first of them, so comparing against a later one would report a
// reference that is actually fine.
//
// This is the brief's index-zero trap answered rather than avoided. `Declarations[0]` is usually the
// earliest and is not guaranteed to be: a value declaration can sort ahead of a type declaration
// written above it, which the brief records as measured. Taking the minimum by position makes the
// answer independent of the ordering rather than resting on it.
func useBeforeDefineBindingDeclaration(declarations []*ast.Node) *ast.Node {
	var earliest *ast.Node
	for _, candidate := range declarations {
		if candidate == nil || candidate.Name() == nil {
			continue
		}
		if earliest == nil || candidate.Name().End() < earliest.Name().End() {
			earliest = candidate
		}
	}
	return earliest
}

// shouldCheckUseBeforeDefineReference applies the option surface to one reference.
//
// Ordered to match upstream's `shouldCheck` so a divergence is readable as a line-for-line
// difference rather than a rewrite.
func shouldCheckUseBeforeDefineReference(
	ctx rule.Context,
	identifier *ast.Node,
	declaration *ast.Node,
	settings noUseBeforeDefineSettings,
) bool {
	// A name in an `export { a }` clause, when the option allows it. Checked before anything else
	// because it is a property of the USE position rather than of the declaration, so none of the
	// per-kind options below can answer it.
	if settings.allowNamedExports && isExportSpecifierLocalName(identifier) {
		return false
	}

	kind := useBeforeDefineDeclarationKind(declaration)

	if !settings.functions && kind == useBeforeDefineFunction {
		return false
	}

	// The variables and classes options are gated on execution context, and upstream's comment says
	// why in one line: "don't skip checking the reference if it's in the same execution context,
	// because of TDZ". Turning `variables` off is a statement about hoisting across a function
	// boundary, where a `let` declared below a closure that reads it is fine at runtime. It is not a
	// licence to ignore `{ a; let a = 1; }`, which throws no matter what the option says, because
	// nothing defers that read past the declaration.
	if (!settings.variables && kind == useBeforeDefineVariable) ||
		(!settings.classes && kind == useBeforeDefineClass) {
		if isFromSeparateExecutionContext(identifier, declaration) {
			return false
		}
	}

	if !settings.enums && kind == useBeforeDefineEnum {
		return false
	}

	if !settings.typedefs && kind == useBeforeDefineType {
		return false
	}

	if settings.ignoreTypeReferences &&
		(referenceContainsTypeQuery(identifier) || isTypeReferencePosition(identifier)) {
		return false
	}

	// Upstream has a branch here restricting a namespace path to its leftmost segment, on the
	// reasoning that in `A.B.C` only `A` names a binding while `B` and `C` name members of it. That
	// branch is NOT reproduced, because `isDeclaringName` above already declines every non-leftmost
	// segment: it treats the right side of a qualified name as a property name rather than a binding
	// reference, so nothing that is not leftmost reaches this point at all.
	//
	// Measured rather than reasoned. A mutation making the leftmost test always answer true survived
	// the whole corpus, and so did one panicking on any non-leftmost segment, which is what
	// distinguishes a subsumed branch from a fixture blind spot. The probe behind that is worth
	// recording: `B` in `let x: A.B` DOES resolve to a real KindTypeAliasDeclaration later in the
	// same file, so the branch looks live from resolution alone and is dead from the earlier guard.

	// A class referenced from inside its own decorator is safe, because decorators are applied after
	// the class binding is initialized. Upstream states this as a transpilation property; it holds
	// for the runtime semantics too, since decorator expressions are evaluated in order once the
	// class has been defined.
	if isClassReferenceInsideItsOwnDecorator(identifier, declaration, kind) {
		return false
	}

	return true
}

// useBeforeDefineDeclarationCategory names what kind of thing a binding is, in the terms the option
// surface is written in.
//
// Not the AST kind, deliberately. Upstream's options are spelled against `eslint-scope`'s definition
// types (`FunctionName`, `ClassName`, `Variable`, `TSEnumName`, `Type`), which group several AST
// kinds each: an interface and a type alias are both `Type`, and a `var`, a `let` and a `const` are
// all `Variable`. Mapping to that vocabulary once here keeps the option checks above readable as
// upstream's own conditions.
type useBeforeDefineDeclarationCategory int

const (
	useBeforeDefineOther useBeforeDefineDeclarationCategory = iota
	useBeforeDefineFunction
	useBeforeDefineClass
	useBeforeDefineVariable
	useBeforeDefineEnum
	useBeforeDefineType
)

// useBeforeDefineDeclarationKind classifies a declaration into the option vocabulary above.
func useBeforeDefineDeclarationKind(declaration *ast.Node) useBeforeDefineDeclarationCategory {
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		return useBeforeDefineFunction

	case ast.KindClassDeclaration, ast.KindClassExpression:
		return useBeforeDefineClass

	case ast.KindVariableDeclaration, ast.KindBindingElement, ast.KindParameter:
		return useBeforeDefineVariable

	case ast.KindEnumDeclaration:
		return useBeforeDefineEnum

	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
		return useBeforeDefineType

	default:
		return useBeforeDefineOther
	}
}

// isExportSpecifierLocalName reports whether an identifier is the LOCAL half of an export specifier.
//
// The local half is the one that names a binding in this file. In `export { a as b }`, `a` is local
// and `b` is only a string in the module's export table, naming nothing here, so `b` is not a
// reference this rule could judge even if the option were off.
func isExportSpecifierLocalName(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil || parent.Kind != ast.KindExportSpecifier {
		return false
	}

	specifier := parent.AsExportSpecifier()

	// With `as`, the local name is `PropertyName` and `Name()` is the exported alias. Without it,
	// there is no `PropertyName` and `Name()` is the local name itself.
	if specifier.PropertyName != nil {
		return specifier.PropertyName == identifier
	}
	return specifier.Name() == identifier
}

// isTypeReferencePosition reports whether an identifier is being used as a type rather than a value.
//
// Anchored on the parent being a type node rather than on a name list, because the same identifier
// spelling appears in both positions and only the surrounding node separates them: the `Foo` in
// `let x: Foo` is a type reference and the `Foo` in `let x = Foo` is a value read.
func isTypeReferencePosition(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindTypeReference:
		return true

	// A qualified name is deliberately NOT walked through, and this is a divergence from the reading
	// that looks obviously right. `A` in `let x: A.B` is inside a type reference by any structural
	// measure, so an upward walk through the qualified name answers true and exempts it under the
	// default `ignoreTypeReferences`. Upstream does not: its check is
	// `identifier.parent.type === "TSTypeReference"`, an immediate-parent test, and `A`'s immediate
	// parent is the qualified name rather than the type reference, so the exemption never fires and
	// the qualified-name branch below decides instead.
	//
	// Measured against the installed rule rather than inferred, because the two readings differ on
	// real inputs and the corpus writes none of them:
	//
	//	let x: Foo; type Foo = 1;                       0 findings, both option settings
	//	let x: A.B; namespace A { export type B = 1 }   1 finding, both option settings
	//
	// The first is exempted and the second is not, and the only difference is whether the type
	// reference is the identifier's own parent. An upward walk reports zero on both and would have
	// shipped a silent false negative on every namespaced type reference in the tree.

	case ast.KindExpressionWithTypeArguments:
		// An `implements` clause is a type position; an `extends` clause on a CLASS is a value
		// position, because the superclass expression is evaluated. Our parser gives both the same
		// node kind and only the heritage clause's token separates them.
		heritage := parent.Parent
		if heritage != nil && heritage.Kind == ast.KindHeritageClause {
			return heritage.AsHeritageClause().Token == ast.KindImplementsKeyword
		}
		return false

	default:
		return false
	}
}

// referenceContainsTypeQuery reports whether an identifier sits inside a `typeof X` type query.
//
// A direct port of upstream's function of the same name, including its walk: from the identifier
// upward through qualified names, answering true at a type query and false at anything else. The
// upward direction is what makes `typeof A.B.C` answer true for every segment.
func referenceContainsTypeQuery(identifier *ast.Node) bool {
	for current := identifier; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindTypeQuery:
			return true
		case ast.KindQualifiedName, ast.KindIdentifier:
			continue
		default:
			return false
		}
	}
	return false
}

// isClassReferenceInsideItsOwnDecorator reports whether a class is named inside a decorator on that
// same class.
func isClassReferenceInsideItsOwnDecorator(
	identifier *ast.Node,
	declaration *ast.Node,
	kind useBeforeDefineDeclarationCategory,
) bool {
	if kind != useBeforeDefineClass {
		return false
	}

	for _, decorator := range decorators.Of(declaration) {
		if identifier.Pos() >= decorator.Pos() && identifier.End() <= decorator.End() {
			return true
		}
	}
	return false
}

// isFromSeparateExecutionContext reports whether a reference runs in a different execution context
// from the one its binding is declared in.
//
// This is the predicate that makes the class family work, and it is worth stating what it is FOR
// before what it does. A reference in a different execution context runs later, when the function
// is called rather than when the declaration is reached, so by the time it evaluates the binding it
// has been initialized whatever the source order says. That is why `var f = () => a; var a;` is a
// forward reference by position and still legal at runtime, and why the `variables` and `classes`
// options are gated on this rather than applied flatly.
//
// Upstream walks eslint-scope's `variableScope` chain. We walk the parent chain for the same
// boundaries, because a variable scope is exactly a function-like node plus the file itself. The
// substitution holds because both are asking "which function body does this sit directly inside",
// and neither the scope graph nor the parent chain has any freedom in answering it.
//
// The class static initializer exception is the part that is not a boundary walk. A static block and
// a static field initializer are implicit functions by every structural measure, and upstream says
// so outright, and they nonetheless run DURING the class definition rather than after it, so for
// this rule's purposes they belong to the enclosing context rather than to a new one. Upstream
// handles this by stepping the scope chain up past them instead of returning true; the loop below
// does the same by continuing past them instead of stopping.
func isFromSeparateExecutionContext(identifier *ast.Node, declaration *ast.Node) bool {
	declarationContext := useBeforeDefineExecutionContext(declaration)

	for current := useBeforeDefineExecutionContext(identifier); current != nil; {
		if current == declarationContext {
			return false
		}
		// A static initializer is transparent: it runs as part of whatever context encloses the
		// class, so the search continues outward rather than concluding the contexts differ.
		if isClassStaticInitializerContext(current) {
			current = useBeforeDefineExecutionContext(current.Parent)
			continue
		}
		return true
	}

	// No fallthrough return, because the loop cannot exit. `useBeforeDefineExecutionContext` walks
	// to KindSourceFile at worst and every node in a parsed file has one above it, so `current` is
	// never nil and the loop always leaves through one of the two returns above.
	//
	// Measured rather than argued: replacing this position with a panic survived all 358 corpus
	// cases, in both directions of the boolean before that. The verdict names its callers, because
	// it expires if they change. Today the only caller of this function that could supply a node
	// outside a source file is none: both call sites (`shouldCheckUseBeforeDefineReference` and
	// `isEvaluatedDuringInitialization`) pass nodes the walk reached from a KindSourceFile listener.
	// Returning false rather than panicking, deliberately. The walk recovers per FILE rather than
	// per rule, so a panic here would cost every other rule its verdict on this file, and "not a
	// separate context" is the conservative answer: it lets the position comparison decide instead
	// of exempting the reference outright.
	return false
}

// useBeforeDefineExecutionContext returns the nearest node that establishes an execution context.
//
// The list is upstream's notion of a variable scope, which is narrower than "any scope": a block is
// a scope for `let` but not an execution context, because the statements inside it run when the
// block is reached rather than when something calls it. Only function-like nodes, class static
// initializers, and the file itself defer execution.
//
// A property declaration WITH an initializer is included because a field initializer is an implicit
// function: it is compiled into the constructor (or, for a static field, into the class definition)
// and runs at that point rather than where it is written. A property declaration without one
// establishes nothing, since there is no expression to defer.
func useBeforeDefineExecutionContext(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor,
			ast.KindClassStaticBlockDeclaration,
			ast.KindModuleDeclaration,
			ast.KindSourceFile:
			return current

		case ast.KindPropertyDeclaration:
			if current.AsPropertyDeclaration().Initializer != nil {
				return current
			}
		}

		// A computed member NAME is not inside the member it names, whatever the parent chain
		// suggests. `class C { [C](){} }` evaluates the key while the class is being defined, so
		// the reference runs in the class's own enclosing context, not in the method's deferred
		// one. Walking straight up would reach KindMethodDeclaration and defer it, which is the
		// verdict for the method's BODY and the opposite of the verdict for its key.
		//
		// Sixteen corpus rows turn on this, spread across methods, accessors and property
		// declarations, static and instance alike. The static modifier does not enter into it:
		// a computed key is evaluated at class definition time either way, which is why
		// `class C { static [C](){} }` and `class C { [C](){} }` both report.
		if current.Parent != nil && current.Parent.Kind == ast.KindComputedPropertyName {
			// Skip past the member the name belongs to, and past the class body, landing in
			// whatever context encloses the class itself.
			if member := current.Parent.Parent; member != nil {
				current = member
			}
		}
	}
	return nil
}

// isClassStaticInitializerContext reports whether an execution context runs during class definition.
//
// Two shapes qualify and they are asymmetric, which is the whole point of the function. A static
// block always does. A property declaration does only when it is STATIC: a static field initializer
// runs while the class is being defined, and an instance field initializer runs later, in the
// constructor, once the class binding is long since available. `class C { field = C; }` is therefore
// clean and `class C { static field = C; }` would not be if the class binding were not already
// initialized by that point, which is why upstream tests the modifier rather than the node kind.
func isClassStaticInitializerContext(context *ast.Node) bool {
	switch context.Kind {
	case ast.KindClassStaticBlockDeclaration:
		return true

	case ast.KindPropertyDeclaration:
		return ast.HasStaticModifier(context)

	default:
		return false
	}
}

// isEvaluatedDuringInitialization reports whether a reference runs while its own binding is still
// being initialized.
//
// This is the arm position cannot see. Every case it catches has the reference written BELOW the
// declaring name and still running too early, because the reference sits inside the very expression
// whose value the binding is waiting for:
//
//	var a = a;
//	var [a = a] = list;
//	for (var a in a) {}
//	class C extends C {}
//	class C { [C]; }
//
// A direct port of upstream's function of the same name, including its two-part structure: classes
// answered by range containment, everything else by an upward walk from the declaring name that
// stops at the first construct which defers evaluation.
func isEvaluatedDuringInitialization(identifier *ast.Node, declaration *ast.Node) bool {
	// A reference from a separate execution context is deferred by definition, so it cannot be
	// running during the initialization however it is positioned. `const x = () => x;` is the case
	// this excludes, and without it the arrow body would read as inside the initializer and report.
	if isFromSeparateExecutionContext(identifier, declaration) {
		return false
	}

	location := identifier.End()

	if declaration.Kind == ast.KindClassDeclaration || declaration.Kind == ast.KindClassExpression {
		// A class binding is initialized before its static initializers run, so a reference from
		// inside one is fine even though it is inside the class's own range. That is what separates
		// `class C extends C {}`, which throws, from `class C { static foo = C; }`, which does not.
		return isInUseBeforeDefineRange(declaration, location) &&
			!isInClassStaticInitializerRange(declaration, location)
	}

	// The walk starts at the declaring name's parent and moves outward, which is upstream's
	// direction. It is looking for the construct that owns the initialization, and it stops at the
	// first one that defers evaluation instead: a function, a class, a catch clause, or an import
	// or export declaration. Past that boundary nothing is running yet.
	for current := declaration; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindVariableDeclaration:
			if isInUseBeforeDefineRange(current.AsVariableDeclaration().Initializer, location) {
				return true
			}
			// `for (var a in a) {}` and `for (var a of a) {}`: the iterated expression is evaluated
			// before the loop variable is bound, so a reference to the variable inside it runs
			// during that variable's own initialization.
			if forStatement := enclosingForInOrOfStatement(current); forStatement != nil {
				if isInUseBeforeDefineRange(forInOrOfExpression(forStatement), location) {
					return true
				}
			}
			return false

		case ast.KindBindingElement:
			// A destructuring default: `var {a = a} = obj`. The default expression is evaluated
			// during the destructuring, so a reference to the binding inside its own default runs
			// too early. The walk continues rather than returning, because a binding element nests
			// inside the variable declaration that will answer the outer question.
			if isInUseBeforeDefineRange(current.AsBindingElement().Initializer, location) {
				return true
			}

		case ast.KindParameter:
			// The same shape for a parameter default: `function f(a = a) {}`.
			if isInUseBeforeDefineRange(current.AsParameterDeclaration().Initializer, location) {
				return true
			}

		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindClassDeclaration,
			ast.KindClassExpression,
			ast.KindCatchClause,
			ast.KindImportDeclaration,
			ast.KindExportDeclaration,

			// The four below are NOT in upstream's sentinel regex, and adding them is fidelity to
			// the decision rather than to the spelling. Upstream matches
			// `(Function|Class)(Declaration|Expression)|ArrowFunctionExpression|...` against ESTree,
			// where an object literal's method, a class method, an accessor and a constructor are
			// all `FunctionExpression` nodes and are therefore all matched. Our parser gives each of
			// them its own kind, so the same regex translated literally misses every one.
			//
			// The cost of missing them is not subtle. A parameter of an object method walks past the
			// method, past the object literal, and reaches the `const` declaration the object is
			// assigned to, whose initializer contains the whole method body, so every use of every
			// such parameter reads as evaluated during that variable's initialization. Measured on
			// the real tree: `const o = { start(controller) { controller.close(); } }` reported,
			// while the byte-equivalent `const o = { start: function(c) { c.x(); } }` did not,
			// because the second spelling produces a KindFunctionExpression that was already in the
			// list. That pair is the whole argument, and no imported case can show it because
			// upstream's parser erases the distinction.
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor:
			// Reaching one of these means the walk has left the initialization behind, so nothing
			// above it can be part of it.
			return false
		}
	}

	return false
}

// isInUseBeforeDefineRange reports whether a source location falls inside a node's extent.
//
// Nil-tolerant on purpose, and it is the reason this exists rather than being written inline at each
// of its five call sites. Every one of them passes something optional: a variable initializer, a
// binding default, a parameter default, a for-in right side. Upstream's own helper opens with
// the same nil check for the same reason. `ast.SkipParentheses` is the cautionary case here: it
// dereferences its argument, and this project lost 167 files to a nil panic through exactly that
// shape.
func isInUseBeforeDefineRange(node *ast.Node, location int) bool {
	return node != nil && node.Pos() <= location && location <= node.End()
}

// isInClassStaticInitializerRange reports whether a location sits inside one of a class's static
// initializers.
//
// Range containment rather than a parent walk, because the caller has a location rather than a node:
// it is asking about a reference somewhere in the class, and the members it must check against are
// the class's own, not whatever encloses the location.
func isInClassStaticInitializerRange(classNode *ast.Node, location int) bool {
	members := classNode.Members()
	for _, member := range members {
		if member == nil {
			continue
		}

		if member.Kind == ast.KindClassStaticBlockDeclaration && isInUseBeforeDefineRange(member, location) {
			return true
		}

		if member.Kind == ast.KindPropertyDeclaration && ast.HasStaticModifier(member) {
			if isInUseBeforeDefineRange(member.AsPropertyDeclaration().Initializer, location) {
				return true
			}
		}
	}
	return false
}

// enclosingForInOrOfStatement returns the for-in or for-of statement a variable declaration is the
// binding of, or nil.
//
// The chain is declaration to declaration list to the statement, which is why this is a helper
// rather than two `.Parent` reads at the call site: a `var` in an ordinary `for` loop has the same
// two-step chain and must not answer here.
func enclosingForInOrOfStatement(declaration *ast.Node) *ast.Node {
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return nil
	}

	statement := list.Parent
	if statement == nil {
		return nil
	}

	switch statement.Kind {
	case ast.KindForInStatement, ast.KindForOfStatement:
		return statement
	default:
		return nil
	}
}

// forInOrOfExpression returns the expression a for-in or for-of statement iterates.
func forInOrOfExpression(statement *ast.Node) *ast.Node {
	switch statement.Kind {
	case ast.KindForInStatement:
		return statement.AsForInOrOfStatement().Expression
	case ast.KindForOfStatement:
		return statement.AsForInOrOfStatement().Expression
	default:
		return nil
	}
}

// useBeforeDefineResolveBinding resolves an identifier to the binding it names.
//
// `GetSymbolAtLocation` alone is not enough for one shape, and it is a shape this rule is squarely
// about. In `export { a }; const a = 1;` the checker resolves the `a` inside the export clause to
// the EXPORT SPECIFIER's own symbol rather than to the `const`, because an export clause introduces
// an alias. That alias's only declaration is the specifier itself, whose position is the reference's
// own position, so a position comparison against it is a comparison of a node with itself and always
// answers "not before". Eleven corpus rows report through exactly this shape and every one of them
// was silent until the alias was followed.
//
// `GetAliasedSymbol` is the recovery, guarded on the alias flag so it is asked only where there is
// an alias to follow.
func useBeforeDefineResolveBinding(ctx rule.Context, identifier *ast.Node) *ast.Symbol {
	// A shorthand property needs a different accessor, and getting this wrong is silent and
	// enormous. `GetSymbolAtLocation` on the `form` in `{ form }` resolves to the PROPERTY's own
	// symbol, whose only declaration is the shorthand node itself, so a position comparison against
	// it compares the reference with itself and reports every shorthand property in the tree whose
	// binding is declared on an earlier line. Measured on the real tree before this guard existed:
	// 3,243 findings, and the first three read at source were all this. With it, 6.
	//
	// The distinguishing pair is worth keeping in mind, because it is what makes the defect
	// invisible from the corpus: `{ form }` reported and `{ form: form }` did not, from the same
	// binding on the same line, and upstream writes no shorthand property anywhere in its 354 cases.
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindShorthandPropertyAssignment {
		if shorthand := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(identifier.Parent); shorthand != nil {
			return shorthand
		}
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return nil
	}

	if symbol.Flags&ast.SymbolFlagsAlias == 0 {
		return symbol
	}

	// An alias pointing into ANOTHER file is an import, and an import declares a local binding here
	// whose position is its own. Following it would compare this file's positions against a
	// different file's and answer nonsense, so the alias is kept unless it resolves to a declaration
	// in this file.
	aliased := ctx.TypeChecker.GetAliasedSymbol(symbol)
	if aliased == nil || aliased == symbol || len(rule.DeclarationsIn(ctx.SourceFile, aliased)) == 0 {
		return symbol
	}

	return aliased
}
