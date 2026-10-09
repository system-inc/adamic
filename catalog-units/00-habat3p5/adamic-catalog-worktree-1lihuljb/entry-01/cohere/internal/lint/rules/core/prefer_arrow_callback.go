package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// PreferArrowCallbackSettings is the decoded option surface.
type PreferArrowCallbackSettings struct {
	// AllowNamedFunctions exempts a function expression that carries its own name.
	// Upstream's default is false.
	AllowNamedFunctions bool
	// AllowUnboundThis permits a callback that uses `this` without being `.bind(this)`-ed.
	// Upstream's default is TRUE, which is why this type exists separately from its wire shape.
	AllowUnboundThis bool
}

// DefaultPreferArrowCallbackSettings is upstream's
// `defaultOptions: [{ allowNamedFunctions: false, allowUnboundThis: true }]`, spelled out because
// the zero value of the struct is not it: AllowUnboundThis would be false, which is a stricter
// rule reporting every `this`-using callback in the tree.
func DefaultPreferArrowCallbackSettings() PreferArrowCallbackSettings {
	return PreferArrowCallbackSettings{AllowUnboundThis: true}
}

// preferArrowCallbackRawOptions is the wire shape.
//
// AllowUnboundThis is a pointer because its default is TRUE, so an absent key and a written `false`
// mean opposite things and a plain bool cannot tell them apart.
type preferArrowCallbackRawOptions struct {
	AllowNamedFunctions bool  `json:"allowNamedFunctions"`
	AllowUnboundThis    *bool `json:"allowUnboundThis"`
}

// DecodePreferArrowCallbackOptions reads the two flags off the config.
//
// Hand rolled because one default is true: a rule configured as a bare "error" is handed nil, the
// generic decoder errors on empty input, and the resulting zero value would say
// `allowUnboundThis: false`. That is not a weaker configuration, it is a different rule.
func DecodePreferArrowCallbackOptions(raw []byte) (any, error) {
	settings := DefaultPreferArrowCallbackSettings()
	if len(raw) == 0 {
		return settings, nil
	}
	var wire preferArrowCallbackRawOptions
	if err := rule.UnmarshalOptions(raw, &wire); err != nil {
		return settings, err
	}
	settings.AllowNamedFunctions = wire.AllowNamedFunctions
	if wire.AllowUnboundThis != nil {
		settings.AllowUnboundThis = *wire.AllowUnboundThis
	}
	return settings, nil
}

// preferArrowCallbackSettingsFrom recovers the settings from whatever the config layer handed over.
func preferArrowCallbackSettingsFrom(options any) PreferArrowCallbackSettings {
	if settings, ok := rule.OptionsAs[PreferArrowCallbackSettings](options); ok {
		return settings
	}
	return DefaultPreferArrowCallbackSettings()
}

var messagePreferArrowCallbackPreferArrowCallback = rule.Message{
	Id: "preferArrowCallback",
	Description: "An arrow function inherits `this` from where it is written instead of from how it " +
		"is called, which is what a callback almost always wants. Writing it as a function " +
		"expression leaves the binding up to the caller.",
}

// PreferArrowCallback reports a function expression passed as a callback.
//
//	valid:   foo(a => a);
//	valid:   foo(function bar() { bar(); });     (it references its own name)
//	valid:   foo(function() { arguments; });     (an arrow has no `arguments`)
//	valid:   foo(function() { this.x; });        (allowUnboundThis, on by default)
//	valid:   foo(function*() {});                (an arrow cannot be a generator)
//	invalid: foo(function() {});                 (fixed to `foo(() => {})`)
//	invalid: foo(function() { this.x; }.bind(this));  (fixed, and the .bind is removed)
//
// # The judgment needs resolution, and a text comparison gets two shapes wrong
//
// Upstream asks two scope questions through eslint-scope: does the function reference its OWN name,
// and does it reference `arguments`. Both matter because an arrow can express neither -- it has no
// binding for itself and none for `arguments` -- so converting such a function changes meaning.
//
// Neither question is answerable by comparing text, and both were probed before this rule was
// written. A self-reference has to be RESOLVED, because a nested declaration or a `const` of the
// same name shadows it:
//
//	foo(function bar() { bar(); });                     1 resolved self-reference, clean upstream
//	foo(function bar() { function bar() {} bar(); });   0, so it REPORTS
//	foo(function bar() { const bar = 1; bar; });        0, so it REPORTS
//
// Measured against the checker, `GetSymbolAtLocation` gives exactly those three answers, and they
// match upstream's verdicts on the same inputs.
//
// `arguments` needs the opposite kind of care: it belongs to the nearest enclosing FUNCTION, so a
// reference inside a nested function is not this one's.
//
//	foo(function() { arguments; });                       owned by the subject, clean
//	foo(function() { function inner() { arguments; } });  owned by `inner`, so it REPORTS
//
// That one is answered by tracking the owner during the walk rather than by the checker, since
// `arguments` has no declaration to resolve to.
//
// # `this`, `super` and `new.target` are tracked per function, and an arrow does not reset them
//
// Upstream pushes a frame on every FunctionDeclaration and FunctionExpression and records whether
// each was seen. Arrow functions are deliberately absent from that set, because an arrow inherits
// all three, so a `this` inside an arrow inside the callback still belongs to the callback.
//
// # The fixer has five declines and every one is asserted by an `output: null` in the corpus
//
// This is the rule where a textually correct repair can be semantically wrong, so the declines are
// the interesting half rather than an afterthought:
//
//	uses `this` without .bind(this)   converting would rebind `this` to the enclosing scope
//	duplicate parameter names         legal in a sloppy function, a SyntaxError in an arrow
//	a parameter literally named this  a TypeScript `this` parameter has no arrow spelling
//	async with a line break before (  removing `function` would join two lines wrongly
//	.bind(this) on a non-member       the shape the fixer would have to unwrap is not there
//
// # No `super` or `new.target` fixer distinction
//
// A callback containing either is never reported at all, so the fixer never sees one. That is
// upstream's structure and it is reproduced: the gate is in the report condition rather than in the
// repair.
var PreferArrowCallback = rule.Rule{
	Name: "prefer-arrow-callback",

	// Telling a genuine self-reference from a shadowed one is name resolution. See the judgment
	// section above for the three measured outcomes.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		settings := preferArrowCallbackSettingsFrom(options)
		if ctx.TypeChecker == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				// One walk rather than per-kind listeners, because the rule needs to know which
				// enclosing function each `this`, `super`, `new.target` and `arguments` belongs to,
				// and the listener walk is pre-order with no exit hook.
				walker := &preferArrowCallbackWalker{ctx: ctx, settings: settings}
				walker.walk(node, nil)
			},
		}
	},
}

// preferArrowCallbackFrame records what one function body was seen to contain.
type preferArrowCallbackFrame struct {
	node *ast.Node
	// usesThis, usesSuper and usesNewTarget are upstream's `scopeInfo` booleans. An arrow does not
	// open a frame, so a `this` inside one lands on the enclosing function's frame, which is what
	// makes `foo(function() { () => this; })` count as using `this`.
	usesThis      bool
	usesSuper     bool
	usesNewTarget bool
	// usesArguments is upstream's `getVariableOfArguments` check, answered by ownership rather than
	// by resolution because `arguments` has no declaration.
	usesArguments bool
}

// preferArrowCallbackWalker carries the frame stack across the recursive walk.
type preferArrowCallbackWalker struct {
	ctx      rule.Context
	settings PreferArrowCallbackSettings
	frames   []*preferArrowCallbackFrame
}

func (w *preferArrowCallbackWalker) walk(node *ast.Node, owner *preferArrowCallbackFrame) {
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindThisKeyword:
		if owner != nil {
			owner.usesThis = true
		}
	case ast.KindSuperKeyword:
		if owner != nil {
			owner.usesSuper = true
		}
	case ast.KindMetaProperty:
		// Upstream's `checkMetaProperty(node, "new", "target")`. `import.meta` is the other meta
		// property and it does not count.
		if owner != nil && node.AsMetaProperty() != nil &&
			node.AsMetaProperty().KeywordToken == ast.KindNewKeyword {
			owner.usesNewTarget = true
		}
	case ast.KindIdentifier:
		// `arguments` belongs to the nearest enclosing function, which is exactly what `owner` is.
		if owner != nil && node.Text() == "arguments" && !preferArrowCallbackIsDeclarationName(node) &&
			!preferArrowCallbackResolvesToADeclaration(w.ctx, node) {
			owner.usesArguments = true
		}
	}

	// A function declaration or expression opens a frame; an arrow deliberately does not, because
	// it inherits `this`, `super`, `new.target` and `arguments` from the function around it.
	nextOwner := owner
	var opened *preferArrowCallbackFrame
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindConstructor, ast.KindClassStaticBlockDeclaration:
		opened = &preferArrowCallbackFrame{node: node}
		w.frames = append(w.frames, opened)
		nextOwner = opened
	}

	node.ForEachChild(func(child *ast.Node) bool {
		w.walk(child, nextOwner)
		return false
	})

	if opened != nil {
		w.frames = w.frames[:len(w.frames)-1]
		if node.Kind == ast.KindFunctionExpression {
			w.judge(node, opened)
		}
	}
}

// preferArrowCallbackResolvesToADeclaration answers whether this `arguments` reference resolves to
// a binding somebody wrote, rather than to the implicit one every function has.
//
// Upstream's `getVariableOfArguments` returns the scope variable only when
// `variable.identifiers.length === 0`, which is its way of saying nobody declared the name. So
// `foo(function(arguments) { arguments; })` REPORTS: the parameter shadows the implicit binding,
// the function does not depend on it, and converting to an arrow is safe. Measured against the
// installed rule, which reports it.
func preferArrowCallbackResolvesToADeclaration(ctx rule.Context, node *ast.Node) bool {
	if ctx.TypeChecker == nil {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	return symbol != nil && len(symbol.Declarations) > 0
}

// preferArrowCallbackIsDeclarationName answers whether this identifier is a name being declared
// rather than a reference, so `const arguments = 1` does not count as using the implicit binding.
//
// Upstream reaches the same place differently: `getVariableOfArguments` returns the scope's
// `arguments` variable only when `variable.identifiers.length === 0`, which is its way of saying
// "nobody declared this name themselves".
func preferArrowCallbackIsDeclarationName(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		declaration := parent.AsVariableDeclaration()
		return declaration != nil && declaration.Name() == node
	case ast.KindParameter:
		declaration := parent.AsParameterDeclaration()
		return declaration != nil && declaration.Name() == node
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindClassDeclaration:
		return parent.Name() == node
	case ast.KindBindingElement:
		element := parent.AsBindingElement()
		return element != nil && element.Name() == node
	}
	return false
}

// judge is upstream's `FunctionExpression:exit`.
func (w *preferArrowCallbackWalker) judge(node *ast.Node, frame *preferArrowCallbackFrame) {
	expression := node.AsFunctionExpression()
	if expression == nil {
		return
	}

	if w.settings.AllowNamedFunctions && expression.Name() != nil {
		return
	}
	// An arrow cannot be a generator, so there is nothing to convert to.
	if expression.AsteriskToken != nil {
		return
	}
	// A function that calls itself by name needs that binding, which an arrow does not provide.
	if preferArrowCallbackReferencesOwnName(w.ctx, node, expression) {
		return
	}
	// An arrow has no `arguments` binding either.
	if frame.usesArguments {
		return
	}

	callback := preferArrowCallbackInfoFor(node)
	if !callback.isCallback {
		return
	}
	if w.settings.AllowUnboundThis && frame.usesThis && !callback.isLexicalThis {
		return
	}
	if frame.usesSuper || frame.usesNewTarget {
		return
	}

	// The repair is withheld in five cases; see the rule's doc comment. Reporting without a fix is
	// the subset that can be shown correct, which is what upstream does through a `fix` returning
	// null.
	if fix, ok := preferArrowCallbackFix(w.ctx, node, expression, frame, callback); ok {
		w.ctx.ReportNodeWithFixes(node, messagePreferArrowCallbackPreferArrowCallback, fix...)
		return
	}
	w.ctx.ReportNode(node, messagePreferArrowCallbackPreferArrowCallback)
}

// preferArrowCallbackReferencesOwnName is upstream's
// `isFunctionName(nameVar) && nameVar.references.length > 0`.
//
// Resolution rather than text: see the rule's doc comment for the three measured shapes, two of
// which a text comparison gets wrong.
func preferArrowCallbackReferencesOwnName(ctx rule.Context, node *ast.Node,
	expression *ast.FunctionExpression) bool {

	name := expression.Name()
	if name == nil {
		return false
	}
	declared := ctx.TypeChecker.GetSymbolAtLocation(name)
	if declared == nil {
		return false
	}

	found := false
	var walk func(*ast.Node)
	walk = func(inner *ast.Node) {
		if inner == nil || found {
			return
		}
		if inner.Kind == ast.KindIdentifier && inner != name && inner.Text() == name.Text() {
			if ctx.TypeChecker.GetSymbolAtLocation(inner) == declared {
				found = true
				return
			}
		}
		inner.ForEachChild(func(child *ast.Node) bool { walk(child); return false })
	}
	walk(node)
	return found
}

// preferArrowCallbackInfo is upstream's `getCallbackInfo` return.
type preferArrowCallbackInfo struct {
	// isCallback says the function is being passed as an argument rather than called.
	isCallback bool
	// isLexicalThis says the function is `.bind(this)`-ed, so converting it to an arrow preserves
	// the binding rather than changing it.
	isLexicalThis bool
	// bindMember is the `.bind(this)` member access the fixer has to remove, when there is one.
	bindMember *ast.Node
	// bindCall is the call that member belongs to.
	bindCall *ast.Node
}

// preferArrowCallbackInfoFor is upstream's `getCallbackInfo`, walking outward from the function to
// find whether it is an argument, looking through the shapes that do not change that answer.
func preferArrowCallbackInfoFor(node *ast.Node) preferArrowCallbackInfo {
	info := preferArrowCallbackInfo{}
	current := node
	bound := false

	for current != nil {
		parent := current.Parent
		if parent == nil {
			return info
		}

		switch parent.Kind {
		// Upstream's LogicalExpression, ChainExpression and ConditionalExpression: none of these
		// changes whether the function ends up an argument, so the walk looks through them. A
		// parenthesis is the same kind of pass-through and is ours alone, since upstream's parser
		// folds it away.
		case ast.KindConditionalExpression, ast.KindParenthesizedExpression:
			current = parent

		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary == nil || !preferArrowCallbackIsLogicalOperator(binary.OperatorToken.Kind) {
				return info
			}
			current = parent

		case ast.KindPropertyAccessExpression:
			access := parent.AsPropertyAccessExpression()
			if access == nil || access.Expression != current ||
				access.Name() == nil || access.Name().Kind != ast.KindIdentifier ||
				access.Name().Text() != "bind" {
				return info
			}
			// The call may sit behind a parenthesis: `(fn?.bind)(this)` puts a
			// KindParenthesizedExpression between the member and its call, where upstream's parser
			// has none. Upstream reaches the same place through its ChainExpression arm, which is
			// what an optional chain is in ESTree. Looking through the wrapper here keeps the walk
			// finding the call; the FIXER separately declines this shape, because its own test is
			// about adjacency rather than about reachability.
			callee := parent
			for callee.Parent != nil && callee.Parent.Kind == ast.KindParenthesizedExpression {
				callee = callee.Parent
			}
			call := callee.Parent
			if call == nil || call.Kind != ast.KindCallExpression ||
				call.AsCallExpression().Expression != callee {
				return info
			}
			if !bound {
				// Only the FIRST `.bind()` decides isLexicalThis, which is upstream's comment.
				bound = true
				arguments := call.AsCallExpression().Arguments
				info.isLexicalThis = arguments != nil && len(arguments.Nodes) == 1 &&
					arguments.Nodes[0].Kind == ast.KindThisKeyword
				info.bindMember = parent
				info.bindCall = call
			}
			current = call

		// Being the CALLEE is not being a callback; being anything else in the call is.
		//
		// The two kinds are separate arms rather than one, because the accessor has to match the
		// kind: `AsCallExpression()` on a NewExpression is an interface conversion that PANICS
		// rather than returning nil. A first draft called it before the kind test and crashed
		// **71 files** on the real tree, which no fixture saw -- upstream's corpus has no
		// `new Foo(function() {})` case, and the walk recovers per FILE, so one panic silently
		// costs every rule its verdict on that file.
		case ast.KindCallExpression:
			call := parent.AsCallExpression()
			info.isCallback = call == nil || call.Expression != current
			return info

		case ast.KindNewExpression:
			construct := parent.AsNewExpression()
			info.isCallback = construct == nil || construct.Expression != current
			return info

		default:
			return info
		}
	}
	return info
}

// preferArrowCallbackIsLogicalOperator answers upstream's LogicalExpression arm.
func preferArrowCallbackIsLogicalOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		return true
	}
	return false
}

// preferArrowCallbackFix builds the repair, or declines.
//
// The declines are the interesting half. Upstream's fixer returns null in five cases and yields
// nothing in two more inside the `.bind(this)` branch, and every one of them is a case where the
// rewritten source would be valid and mean something else, or would delete a comment. All seven are
// asserted by an `output: null` in the corpus.
func preferArrowCallbackFix(ctx rule.Context, node *ast.Node, expression *ast.FunctionExpression,
	frame *preferArrowCallbackFrame, callback preferArrowCallbackInfo) ([]rule.Fix, bool) {

	// Decline 1: a callback that uses `this` without being bound to it. An arrow would take `this`
	// from the enclosing scope, which is a different value.
	if !callback.isLexicalThis && frame.usesThis {
		return nil, false
	}
	// Decline 2: duplicate parameter names are legal in a sloppy-mode function and a SyntaxError in
	// an arrow, so the repair would not parse.
	if preferArrowCallbackHasDuplicateParameters(node) {
		return nil, false
	}
	// Decline 3: a TypeScript `this` parameter has no arrow spelling.
	if parameters := node.Parameters(); len(parameters) > 0 {
		if name := parameters[0].Name(); name != nil && name.Kind == ast.KindIdentifier &&
			name.Text() == "this" {
			return nil, false
		}
	}

	text := ctx.SourceFile.Text()
	functionKeyword := preferArrowCallbackFunctionKeywordRange(ctx, node)
	if functionKeyword.Pos() < 0 {
		return nil, false
	}
	parameterListStart := preferArrowCallbackParameterListStart(ctx, node)
	if parameterListStart < 0 || parameterListStart < functionKeyword.End() {
		return nil, false
	}

	// Decline 4: `async` followed by a line break before the parameter list. Removing `function`
	// would join two lines into something the author did not write.
	if preferArrowCallbackIsAsync(node) {
		between := text[functionKeyword.End():parameterListStart]
		for _, character := range between {
			if character == '\n' {
				return nil, false
			}
		}
	}

	fixes := make([]rule.Fix, 0, 3)

	if callback.isLexicalThis {
		// Decline 5: the shape the fixer would unwrap is not a member access.
		//
		// Upstream tests `node.parent.type !== "MemberExpression"` -- the function's IMMEDIATE
		// parent, not the member the callback walk found. The two differ whenever anything sits
		// between them, and upstream declines in exactly those cases:
		//
		//	foo((bar || function() {}).bind(this))              reports, NO fix
		//	foo((function() { return this; }?.bind)(this));     reports, NO fix
		//	foo(function() {}.bind(this))                       reports and fixes
		//
		// Our parser keeps `KindParenthesizedExpression` as a real node, so a parenthesised form
		// puts one between the function and the member and this test catches it the same way. The
		// callback walk deliberately looks THROUGH those nodes to decide `isCallback`, so the walk
		// and the fixer ask different questions and only the fixer's is about adjacency.
		if callback.bindMember == nil || callback.bindCall == nil {
			return nil, false
		}
		// The function must be the member's object AND the member must be the call's callee, with
		// nothing in between on either side. Upstream gets both for free from its parser: it tests
		// `node.parent.type !== "MemberExpression"` and then reads `memberNode.parent` as the call,
		// so any intervening node breaks one or the other. Ours keeps parentheses as real nodes, so
		// both adjacencies have to be asserted explicitly.
		if node.Parent != callback.bindMember || callback.bindMember.Parent != callback.bindCall {
			return nil, false
		}
		member := callback.bindMember.AsPropertyAccessExpression()
		if member == nil {
			return nil, false
		}
		// The span from just after the bound expression through the end of the `.bind(...)` call.
		removalStart := member.Expression.End()
		removalEnd := callback.bindCall.End()
		if removalEnd <= removalStart {
			return nil, false
		}
		// Decline 6: a comment inside the `.bind(this)` would be deleted by the removal.
		if preferArrowCallbackHasCommentBetween(text, removalStart, removalEnd) {
			return nil, false
		}
		fixes = append(fixes, rule.RemoveRange(core.NewTextRange(removalStart, removalEnd)))
	}

	// Remove `function` and any name, preserving a comment between them if there is one.
	if preferArrowCallbackHasCommentBetween(text, functionKeyword.End(), parameterListStart) {
		// Upstream removes only the tokens rather than the span, so the comment survives.
		fixes = append(fixes, rule.RemoveRange(functionKeyword))
		if expression.Name() != nil {
			fixes = append(fixes, ctx.RemoveNode(expression.Name()))
		}
	} else {
		fixes = append(fixes, rule.RemoveRange(
			core.NewTextRange(functionKeyword.Pos(), parameterListStart)))
	}

	// Insert ` =>` before the body.
	body := node.Body()
	if body == nil {
		return nil, false
	}
	arrowInsertionPoint := preferArrowCallbackTokenBeforeBodyEnd(text, body.Pos())
	fixes = append(fixes, rule.ReplaceRange(
		core.NewTextRange(arrowInsertionPoint, arrowInsertionPoint), " =>"))

	// Wrap the result in parentheses unless it is already unambiguous.
	//
	// An arrow binds looser than most operators, so `foo(nativeCb || function() {})` becomes
	// `foo(nativeCb || (() => {}))` rather than `foo(nativeCb || () => {})`, which does not parse
	// the same way. Upstream skips the wrap when the parent is a call or a conditional, since an
	// argument position and a conditional branch are already delimited, and when either the node or
	// the replaced span is parenthesised.
	//
	// The replaced span is the `.bind(this)` CALL when there was one, not the function, because
	// that whole expression is what the arrow replaces.
	replaced := node
	if callback.isLexicalThis && callback.bindCall != nil {
		replaced = callback.bindCall
	}
	if preferArrowCallbackNeedsParentheses(text, replaced, node) {
		// The TOKEN start, not `Pos()`: `Pos()` begins at the previous token's end and so includes
		// the leading whitespace, which would place the `(` before the space and produce
		// `nativeCb ||( () => {})`. Same trap sort-vars has for its declarator spans.
		open := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, replaced.Pos()).Pos()
		fixes = append(fixes,
			rule.ReplaceRange(core.NewTextRange(open, open), "("),
			rule.ReplaceRange(core.NewTextRange(replaced.End(), replaced.End()), ")"))
	}

	return fixes, true
}

// preferArrowCallbackNeedsParentheses answers upstream's final guard.
func preferArrowCallbackNeedsParentheses(text string, replaced *ast.Node, node *ast.Node) bool {
	parent := replaced.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindConditionalExpression:
		return false
	}
	if preferArrowCallbackIsParenthesised(text, replaced) ||
		preferArrowCallbackIsParenthesised(text, node) {
		return false
	}
	return true
}

// preferArrowCallbackIsParenthesised is upstream's `astUtils.isParenthesised`, which asks whether
// the tokens immediately around the node are a matching pair of parentheses.
//
// Our parser keeps `KindParenthesizedExpression` as a real node, so the common case is answered by
// the parent's kind. The textual scan covers the rest: a parenthesis that belongs to an enclosing
// construct rather than to a wrapper node.
func preferArrowCallbackIsParenthesised(text string, node *ast.Node) bool {
	if node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		return true
	}
	before := node.Pos() - 1
	for before >= 0 && (text[before] == ' ' || text[before] == '\t' ||
		text[before] == '\n' || text[before] == '\r') {
		before--
	}
	after := node.End()
	for after < len(text) && (text[after] == ' ' || text[after] == '\t' ||
		text[after] == '\n' || text[after] == '\r') {
		after++
	}
	return before >= 0 && after < len(text) && text[before] == '(' && text[after] == ')'
}

// preferArrowCallbackHasDuplicateParameters is upstream's `hasDuplicateParams`.
//
// It applies only when EVERY parameter is a plain identifier, because a destructuring pattern
// cannot collide the way two identifiers can.
func preferArrowCallbackHasDuplicateParameters(node *ast.Node) bool {
	seen := map[string]bool{}
	for _, parameter := range node.Parameters() {
		name := parameter.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		if seen[name.Text()] {
			return true
		}
		seen[name.Text()] = true
	}
	return false
}

// preferArrowCallbackIsAsync answers whether the function carries the async modifier.
func preferArrowCallbackIsAsync(node *ast.Node) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == ast.KindAsyncKeyword {
			return true
		}
	}
	return false
}

// preferArrowCallbackFunctionKeywordRange finds the `function` keyword, which is upstream's
// `getFirstToken(node, node.async ? 1 : 0)`.
func preferArrowCallbackFunctionKeywordRange(ctx rule.Context, node *ast.Node) core.TextRange {
	text := ctx.SourceFile.Text()
	limit := node.End()
	for index := node.Pos(); index+8 <= limit && index+8 <= len(text); index++ {
		if text[index:index+8] == "function" {
			return core.NewTextRange(index, index+8)
		}
	}
	return core.NewTextRange(-1, -1)
}

// preferArrowCallbackParameterListStart finds the `(` that opens the parameter list, which is
// upstream's `getTokenAfter(functionToken, isOpeningParenToken)`.
func preferArrowCallbackParameterListStart(ctx rule.Context, node *ast.Node) int {
	text := ctx.SourceFile.Text()
	body := node.Body()
	limit := node.End()
	if body != nil {
		limit = body.Pos()
	}
	if limit > len(text) {
		limit = len(text)
	}
	for index := node.Pos(); index < limit; index++ {
		if text[index] == '(' {
			return index
		}
	}
	return -1
}

// preferArrowCallbackTokenBeforeBodyEnd finds the end of the token before the body, which is where
// upstream inserts ` =>`.
//
// Scanning backward from the body's start over whitespace lands on the closing parenthesis of the
// parameter list, or on a return type annotation's last character.
func preferArrowCallbackTokenBeforeBodyEnd(text string, bodyPos int) int {
	index := bodyPos
	for index > 0 && (text[index-1] == ' ' || text[index-1] == '\t' ||
		text[index-1] == '\n' || text[index-1] == '\r') {
		index--
	}
	return index
}

// preferArrowCallbackHasCommentBetween answers upstream's `commentsExistBetween`.
//
// A textual scan rather than a token walk, because the question is only whether a comment sits in a
// span the fixer is about to delete. It has to skip string and template literals so a `//` inside
// one is not mistaken for a comment.
func preferArrowCallbackHasCommentBetween(text string, start int, end int) bool {
	if start < 0 || end > len(text) || start >= end {
		return false
	}
	for index := start; index < end-1; index++ {
		switch text[index] {
		case '\'', '"', '`':
			quote := text[index]
			index++
			for index < end && text[index] != quote {
				if text[index] == '\\' {
					index++
				}
				index++
			}
		case '/':
			if text[index+1] == '/' || text[index+1] == '*' {
				return true
			}
		}
	}
	return false
}
