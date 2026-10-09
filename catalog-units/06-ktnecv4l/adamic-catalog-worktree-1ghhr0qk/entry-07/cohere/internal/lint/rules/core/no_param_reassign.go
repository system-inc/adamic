package core

import (
	"regexp"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

func messageAssignmentToParameter(name string) rule.Message {
	return rule.Message{
		Id:          "assignmentToFunctionParam",
		Description: "Assignment to function parameter '" + name + "'.",
	}
}

func messageAssignmentToParameterProperty(name string) rule.Message {
	return rule.Message{
		Id:          "assignmentToFunctionParamProp",
		Description: "Assignment to property of function parameter '" + name + "'.",
	}
}

// NoParamReassignOptions configures whether property writes count and which names are excused.
type NoParamReassignOptions struct {
	// Props extends the rule from the parameter binding to properties reached through it. Off by
	// default, and the default is what makes the two halves separable: reassigning the binding is
	// always local and always a defect, while writing through it mutates the caller's object, which
	// is sometimes exactly the intent.
	Props bool `json:"props"`

	// IgnorePropertyModificationsFor is a list of parameter names whose property writes are excused,
	// matched literally.
	IgnorePropertyModificationsFor []string `json:"ignorePropertyModificationsFor"`

	// IgnorePropertyModificationsForRegex is the same allowance written as patterns.
	IgnorePropertyModificationsForRegex []string `json:"ignorePropertyModificationsForRegex"`
}

// NoParamReassign flags a write to a function parameter, and optionally through one.
//
//	valid:   function foo(a) { var b = a; }
//	valid:   function foo(a) { a.prop = 'value'; }
//	valid:   function foo(a) { for (b in a); }
//	invalid: function foo(bar) { bar = 13; }
//	invalid: function foo(bar) { ++bar; }
//	invalid: function foo({bar}) { bar = 13; }
//	invalid: function foo(bar) { bar.a = 0; }          with props
//
// A parameter is a local binding holding a copy of what the caller passed, so reassigning it throws
// away the argument while leaving the name in place, and every later read gets the new value while
// still reading as "the thing that was passed in". It also breaks `arguments` aliasing in sloppy
// mode, where the parameter and `arguments[0]` are the same storage. Under `props` the rule extends
// to writes THROUGH the parameter, which is a different defect: those reach the caller's object and
// mutate it, which is invisible at the call site.
//
// # Why this reaches the checker
//
// The discrimination is name resolution and almost none of it is structural. Upstream asks
// `getDeclaredVariables(functionNode)` for the parameters and then walks each one's resolved
// references, which is a find-all-references index we do not have. Four of upstream's clean cases
// write to a name spelled exactly like a parameter:
//
//	function foo(a) { (function() { var a = 12; a++; })(); }   an inner local shadows it
//	function foo() { someGlobal = 13; }                        no parameter of that name at all
//	function foo(a) { for (b in a); }                          `b` is not the parameter
//	function foo(a, z) { a.b = 0; x.y = 0; }                   `x` is not a parameter
//
// So this anchors on the parameter's own binding identifiers, which the listener already has, and
// asks the checker which declaration each candidate occurrence resolves to. A loop over the symbol's
// declarations rather than an index into it: the question is "does ANY declaration of this name's
// symbol belong to this parameter", and more declarations only mean more chances to match.
//
// # Two findings, and the second is not a write at all
//
// `bar = 13` is a write to the binding and reports as one. `bar.a = 0` is a READ of the binding whose
// surrounding expression mutates a property, so the identifier itself writes nothing and the finding
// comes from climbing outward from it. Upstream separates them the same way, and reports the property
// finding only when the reference is NOT a write, which is why `({bar} = {})` reports the binding
// finding rather than the property one.
//
// # The climb, which is upstream's algorithm rather than a simplification
//
// From the identifier, walk outward until a node that stops the search, answering along the way.
// An assignment answers true if the climb came up its left side, an update or a `delete` answers true
// outright, a `for...in` or `for...of` head answers true only if the climb came up its left. Four
// shapes answer FALSE and each is a real clean case: the arguments of a call, the computed property
// of a member access, the key of an object property, and the test of a ternary. Each is a position
// where the parameter's value is being READ to decide where some other write lands.
//
// The stop set is upstream's regular expression over node type names, which matches anything ending
// in Statement, Declaration, Function, FunctionExpression, or Program -- with `for...in` and
// `for...of` explicitly exempted so the climb can reach their heads. That set does not translate to
// node kinds directly, so it is reproduced here as a predicate over the kinds our parser produces,
// with the same three exemptions.
//
// No fix. The repair for a reassigned parameter is a new local binding, which means choosing a name
// and deciding which later reads meant the parameter; the repair for a property write is not an edit
// at all, since the point of the finding is that the mutation reaches the caller.
var NoParamReassign = rule.Rule{
	Name: "no-param-reassign",

	// See the doc above: four clean cases are textually identical to failing ones and differ only in
	// what the name resolves to.
	NeedsTypeChecker: true,
	// Reads only this file's declarations (rule.DeclarationsIn), so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		var settings NoParamReassignOptions
		if parsed, ok := rule.OptionsAs[NoParamReassignOptions](options); ok {
			settings = parsed
		}
		ignored := compileIgnoredNames(settings)

		checkFunction := func(node *ast.Node) {
			// The engine hands every rule a nil checker when the program could not be built, and
			// this rule can answer nothing without one. The failure mode without the guard is
			// silence rather than a panic, because GetSymbolAtLocation tolerates a nil receiver,
			// and a silent rule whose fixtures all pass is the worse of the two.
			if ctx.TypeChecker == nil {
				return
			}
			parameters := node.Parameters()
			if len(parameters) == 0 {
				return
			}

			// One anchor per binding identifier, not per parameter, because a destructured
			// parameter declares several names: `function foo([, {bar}])` binds `bar` alone.
			anchors := map[*ast.Node]string{}
			for _, parameter := range parameters {
				collectParameterBindings(parameter.Name(), anchors)
			}
			if len(anchors) == 0 {
				return
			}

			// The parameter list AND the body, which is one more place than the obvious reading
			// suggests. Upstream excludes only the reference that IS the declaration, through the
			// `init` flag on it, and everything else in the parameter list is a reference like any
			// other -- so `function foo(a, b = (a = 1)) { }` reports and `function foo(a, b = a) { }`
			// does not. Searching the body alone loses every write inside a default value; measured
			// against the installed rule on five such shapes after a mutant that widened the search
			// survived every fixture.
			//
			// A nested function inside the body is deliberately included, since
			// `function foo(bar) { (function() { bar = 13; })(); }` reports.
			var visit func(*ast.Node)
			visit = func(current *ast.Node) {
				if current == nil {
					return
				}
				if current.Kind == ast.KindIdentifier && !isOwnParameterBinding(current, anchors) {
					if name, isAnchored := anchoredParameterName(ctx, current, anchors); isAnchored {
						reportParameterUse(ctx, current, name, settings.Props, ignored)
					}
				}
				current.ForEachChild(func(child *ast.Node) bool {
					visit(child)
					return false
				})
			}
			for _, parameter := range parameters {
				visit(parameter)
			}
			if body := node.Body(); body != nil {
				visit(body)
			}
		}

		// Upstream registers three listeners and this registers seven, which is fidelity rather than
		// a widening. An estree tree makes a method's value a `FunctionExpression` hanging off a
		// `MethodDefinition`, so upstream's three listeners already cover every method, constructor,
		// getter and setter without naming them. Our parser gives each of those its own kind and no
		// `FunctionExpression` underneath, so registering only upstream's three loses every
		// parameter in every class body.
		//
		// Measured: a 206-input differential run reported six mismatches and all six were this,
		// including `class C { m(a) { a = 1; } }`. The corpus writes no class at all, so nothing
		// imported could see it, and the failure direction is silent -- findings lost rather than
		// added.
		return rule.Listeners{
			ast.KindFunctionDeclaration: checkFunction,
			ast.KindFunctionExpression:  checkFunction,
			ast.KindArrowFunction:       checkFunction,
			ast.KindMethodDeclaration:   checkFunction,
			ast.KindConstructor:         checkFunction,
			ast.KindGetAccessor:         checkFunction,
			ast.KindSetAccessor:         checkFunction,
		}
	},
}

// collectParameterBindings records every name a parameter binds, keyed by its declaration node.
//
// A plain parameter binds one name. A destructured one binds as many as it names, at any depth, and
// through rest elements: `function foo([, {bar}])` binds only `bar`, and `function foo({a, ...rest})`
// binds both. A default value inside the pattern is an expression rather than a binding and is not
// collected, which is what keeps `function foo({a = b})` from anchoring on `b`.
func collectParameterBindings(name *ast.Node, into map[*ast.Node]string) {
	if name == nil {
		return
	}
	switch name.Kind {
	case ast.KindIdentifier:
		// Keyed on the identifier's PARENT rather than on the identifier, because that is the node
		// the checker hands back as the declaration: `KindParameter` for a plain parameter and
		// `KindBindingElement` for a destructured one. Keying on the identifier itself compiles,
		// passes nothing, and reads as a rule that never fires -- measured, all 38 failing cases
		// went silent, which is the failure mode this whole family has: a wrong anchor costs every
		// finding and no fixture can tell it from a rule that is merely wrong.
		if name.Parent == nil {
			return
		}
		into[name.Parent] = name.Text()
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		name.ForEachChild(func(element *ast.Node) bool {
			// A binding element carries the bound name in `Name` and anything else -- a property
			// name, a default value -- in slots this deliberately does not descend into.
			if element.Kind == ast.KindBindingElement {
				collectParameterBindings(element.Name(), into)
			}
			return false
		})
	}
}

// isOwnParameterBinding reports whether an identifier IS one of the anchors rather than a use of one.
//
// This is upstream's `init` exclusion, spelled structurally. A parameter's own binding identifier
// resolves to the parameter and would otherwise be judged like any other occurrence; upstream skips
// it because the scope index flags that reference as the initialization.
//
// Two mutants disabling this guard survive every fixture, and the reason is measured rather than
// assumed: probed over 17 parameter shapes -- plain, defaulted, object and array patterns, nested
// patterns, rest, renamed, elided, a whole-pattern default, a typed default, a class method, an
// arrow -- neither `WritesToBinding` nor the property climb answers true for a parameter's own
// name, so nothing this guard declines would have been reported anyway.
//
// It is kept because deleting it would leave the rule's silence on `function foo(a) {}` resting on
// a property of a shelf helper in another package rather than on a decision made here, and because
// the search now covers the parameter LIST, where the own-binding occurrence actually appears. The
// redundancy is with an absence rather than with another branch, which is the shape a single-site
// sweep structurally cannot see. Verdict taken over this function's single caller.
func isOwnParameterBinding(identifier *ast.Node, anchors map[*ast.Node]string) bool {
	_, isAnchor := anchors[identifier.Parent]
	return isAnchor && identifier.Parent.Name() == identifier
}

// anchoredParameterName reports whether an identifier resolves to one of this function's parameters.
//
// A LOOP over the symbol's declarations rather than an index into it, because the question here is
// "does any declaration of this symbol belong to this function's parameter list". Indexing at zero
// would go silent wherever a symbol carries more than one declaration and the parameter is not the
// first, and the ordering of that list is not a promise.
//
// A mutant replacing the loop with an index at zero SURVIVES every fixture, and the reason is worth
// recording rather than acting on. Probed across overload shapes -- overloaded functions, overloaded
// class methods, overloaded interface method signatures, an ambient declaration beside a definition
// -- a merged symbol appears on the function or method NAME and never on a parameter: each overload
// signature's parameters are distinct symbols carrying one declaration each. So no input
// distinguishes the two spellings today.
//
// The loop stays because it answers the question the rule is actually asking, and because the
// equivalence is a fact about what the checker merges rather than about what this rule needs. A
// checker change that merged parameters across overloads, or an anchor set widened past parameters,
// makes the index-at-zero version wrong with nothing to notice. Verdict taken over this function's
// single caller.
//
// A shorthand property in a destructuring target resolves through a different accessor: `({bar} = {})`
// asked through the plain one answers with the PROPERTY's symbol, which reads exactly like a
// correctly declined shadow. That case is in upstream's corpus as a failing one, so the wrong
// accessor loses a finding silently.
func anchoredParameterName(ctx rule.Context, identifier *ast.Node, anchors map[*ast.Node]string) (string, bool) {
	// A cheap pre-filter before a checker call, since this runs on every identifier in every
	// function body. Symbol identity already implies the text matches, so this decides nothing.
	matchesAnyName := false
	for _, name := range anchors {
		if name == identifier.Text() {
			matchesAnyName = true
			break
		}
	}
	if !matchesAnyName {
		return "", false
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindShorthandPropertyAssignment {
		if valueSymbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(identifier.Parent); valueSymbol != nil {
			symbol = valueSymbol
		}
	}
	if symbol == nil {
		return "", false
	}
	// The anchors are this file's parameters, so only this file's declarations can be one.
	for _, declaration := range rule.DeclarationsIn(ctx.SourceFile, symbol) {
		if name, isAnchored := anchors[declaration]; isAnchored {
			return name, true
		}
	}
	return "", false
}

// reportParameterUse decides which of the two findings an occurrence earns, or neither.
//
// The order is upstream's and it is load-bearing: a write to the binding reports the binding finding
// and STOPS, so the property check never sees it. `({bar} = {})` is a write to `bar` and reports as
// one; without the ordering it would also satisfy the property climb and report twice.
func reportParameterUse(
	ctx rule.Context,
	identifier *ast.Node,
	name string,
	props bool,
	ignored ignoredParameterNames,
) {
	if reference.WritesToBinding(identifier) {
		ctx.ReportNode(identifier, messageAssignmentToParameter(name))
		return
	}
	if !props || ignored.matches(name) {
		return
	}
	if modifiesPropertyThrough(identifier) {
		ctx.ReportNode(identifier, messageAssignmentToParameterProperty(name))
	}
}

// modifiesPropertyThrough reports whether the expression around an identifier writes through it.
//
// Upstream's climb, kind for kind. It walks outward from the identifier until a node that stops the
// search, and the four FALSE arms are the whole value of the function: each one is a position where
// the parameter is read in order to decide where some OTHER write lands, and each is a clean case
// upstream shipped after somebody hit it.
//
//	bar(a.b).c = 0        the call's argument -- the write lands on the call's result
//	data[a.b] = 0         a computed member -- the write lands on `data`
//	({ [a]: v } = value)  a property key -- the write lands on `v`
//	(a ? [] : [])[0] = 1  a ternary test -- the write lands on the branch's result
func modifiesPropertyThrough(identifier *ast.Node) bool {
	node := identifier
	parent := node.Parent

	for parent != nil && (!stopsPropertyClimb(parent) || isForInOrOf(parent)) {
		switch parent.Kind {
		case ast.KindBinaryExpression:
			// Only an assignment counts, and only from its left side. `a.b === c` climbs past a
			// binary expression that assigns nothing, and upstream reaches the same answer because
			// its tree gives assignment its own node type while ours shares one with every operator.
			binary := parent.AsBinaryExpression()
			if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				return binary.Left == node
			}

		case ast.KindPrefixUnaryExpression:
			// `++a.b` and `--a.b`. A `!` or a `-` is not a write, and unlike upstream -- whose tree
			// gives updates their own node type -- our parser shares one kind with every prefix
			// operator, so the operator has to be read.
			operator := parent.AsPrefixUnaryExpression().Operator
			if operator == ast.KindPlusPlusToken || operator == ast.KindMinusMinusToken {
				return true
			}

		case ast.KindPostfixUnaryExpression:
			// `a.b++` and `a.b--`, which are always one of the two.
			return true

		case ast.KindDeleteExpression:
			return true

		case ast.KindForInStatement, ast.KindForOfStatement:
			// The head writes, the source and the body do not. `for (a.b in obj)` writes through the
			// parameter; `for (bar in a.b)` reads it, and `for (bar in baz) a.b;` merely mentions it.
			if forStatementInitializer(parent) == node {
				return true
			}
			return false

		case ast.KindCallExpression:
			if parent.AsCallExpression().Expression != node {
				return false
			}

		case ast.KindNewExpression:
			if parent.AsNewExpression().Expression != node {
				return false
			}

		case ast.KindElementAccessExpression:
			// The subscript rather than the receiver. `data[a.b] = 0` reads the parameter to build a
			// key; `a[b] = 0` writes through it.
			if parent.AsElementAccessExpression().ArgumentExpression == node {
				return false
			}

		case ast.KindPropertyAssignment:
			// A computed key reads the parameter to decide which property some other object gets.
			if parent.AsPropertyAssignment().Name() == node {
				return false
			}

		case ast.KindConditionalExpression:
			if parent.AsConditionalExpression().Condition == node {
				return false
			}
		}

		node = parent
		parent = node.Parent
	}

	return false
}

// stopsPropertyClimb reproduces upstream's stop set over the kinds our parser produces.
//
// Upstream tests the node's TYPE NAME against a regular expression matching anything ending in
// Statement, Declaration, Function, FunctionExpression, or Program. That is a string test over
// estree's type names, and the obvious translation -- `ast.IsStatement(node) || ast.IsDeclaration(node)`
// plus the function-like kinds -- is wrong here in a way that costs every property finding.
//
// `ast.IsDeclaration` answers TRUE for a `KindBinaryExpression`, because typescript-go treats an
// assignment as a declaration site: `exports.x = 1` and `this.x = 1` declare in JavaScript, and the
// predicate has to say so for the binder. So the obvious spelling stops the climb at the assignment
// node, which is the one node the climb exists to reach, and the rule goes silent on all 25 property
// cases while every binding case still passes. Measured, not reasoned: the first version of this
// function used those two predicates and 14 fixtures failed with no other symptom.
//
// So the boundary is an explicit kind set instead. It is longer and it is checkable: each entry
// either ends a statement, ends a declaration, or ends a function scope, which is what upstream's
// pattern names.
//
// The two `for` head statements are exempted by the caller rather than removed here, matching
// upstream, which lists them in the stop set and then writes a clause letting the climb through.
func stopsPropertyClimb(node *ast.Node) bool {
	switch node.Kind {
	// A function scope, which upstream's `Function` and `FunctionExpression` arms name.
	case ast.KindSourceFile,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindFunctionDeclaration,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindConstructor:
		return true

	// A statement. Every kind here ends in `Statement` in estree's naming too, which is what makes
	// the list checkable against upstream's pattern rather than merely plausible. `ForInStatement`
	// and `ForOfStatement` are in the set and the caller lets the climb through them, exactly as
	// upstream does.
	//
	// `KindBlock` is measurably redundant and is kept anyway. A block's children are statements, so
	// a climb reaching a block has already passed one and stopped -- probed over 10 shapes including
	// a class static block, a switch case, and a try/finally, and it is never the first stop. A
	// mutant removing it survives every fixture, correctly. It stays because this list is checked by
	// reading it against upstream's pattern, which names `BlockStatement`, and a set that silently
	// omits one entry for being unreachable is harder to check than one that does not.
	case ast.KindBlock,
		ast.KindVariableStatement,
		ast.KindExpressionStatement,
		ast.KindIfStatement,
		ast.KindDoStatement,
		ast.KindWhileStatement,
		ast.KindForStatement,
		ast.KindForInStatement,
		ast.KindForOfStatement,
		ast.KindContinueStatement,
		ast.KindBreakStatement,
		ast.KindReturnStatement,
		ast.KindWithStatement,
		ast.KindSwitchStatement,
		ast.KindLabeledStatement,
		ast.KindThrowStatement,
		ast.KindTryStatement,
		ast.KindDebuggerStatement,
		ast.KindEmptyStatement:
		return true

	// A declaration, which upstream's `Declaration` arm names.
	case ast.KindVariableDeclaration,
		ast.KindVariableDeclarationList,
		ast.KindClassDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration,
		ast.KindImportDeclaration,
		ast.KindExportDeclaration,
		ast.KindExportAssignment,
		ast.KindImportEqualsDeclaration:
		return true
	}
	return false
}

// isForInOrOf reports whether a node is one of the two heads the climb is allowed through.
func isForInOrOf(node *ast.Node) bool {
	return node.Kind == ast.KindForInStatement || node.Kind == ast.KindForOfStatement
}

// forStatementInitializer reads the head of a `for...in` or `for...of`.
func forStatementInitializer(node *ast.Node) *ast.Node {
	if node.Kind == ast.KindForInStatement {
		return node.AsForInOrOfStatement().Initializer
	}
	return node.AsForInOrOfStatement().Initializer
}

// ignoredParameterNames is the compiled form of the two allowance options.
type ignoredParameterNames struct {
	literal  map[string]bool
	patterns []*regexp.Regexp
}

// matches reports whether a parameter's property writes are excused.
func (ignored ignoredParameterNames) matches(name string) bool {
	if ignored.literal[name] {
		return true
	}
	for _, pattern := range ignored.patterns {
		if pattern.MatchString(name) {
			return true
		}
	}
	return false
}

// compileIgnoredNames turns the two option lists into one predicate, compiling each pattern once.
//
// A pattern that will not compile is DROPPED rather than treated as matching nothing-or-everything,
// and that is a divergence worth stating: upstream builds a `RegExp` at every call and a bad pattern
// throws, taking the whole lint run down with a message naming the rule. Dropping it means a
// misconfigured pattern silently stops excusing, which reports MORE rather than less -- the direction
// that is visible rather than the one that hides findings.
//
// The patterns are also not identical in dialect. Upstream compiles with the `u` flag, so its escapes
// are the Unicode-mode JavaScript ones; Go's regexp is RE2, which has no backreferences and no
// lookaround. A pattern using either is rejected here and accepted there, which is the same drop.
func compileIgnoredNames(options NoParamReassignOptions) ignoredParameterNames {
	compiled := ignoredParameterNames{literal: map[string]bool{}}
	for _, name := range options.IgnorePropertyModificationsFor {
		compiled.literal[name] = true
	}
	for _, source := range options.IgnorePropertyModificationsForRegex {
		if pattern, err := regexp.Compile(source); err == nil {
			compiled.patterns = append(compiled.patterns, pattern)
		}
	}
	return compiled
}

// DecodeNoParamReassignOptions reads this rule's configuration from the config layer.
//
// Hand-rolled rather than `rule.DecodeOptionsInto`, so that empty input decodes to the default
// rather than erroring into a nil the caller reads as a zero value. Every option here defaults to
// its zero value, so nothing inverts, but the two shapes should still not be told apart by whether
// the decoder happened to run.
func DecodeNoParamReassignOptions(raw []byte) (any, error) {
	var options NoParamReassignOptions
	if len(raw) == 0 {
		return options, nil
	}
	if err := rule.UnmarshalOptions(raw, &options); err != nil {
		return options, err
	}
	return options, nil
}
