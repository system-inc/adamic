// Package consistentreturn answers whether the `return` statements in one function agree about
// carrying a value, and where to point when they do not.
//
// Lifted out of `internal/lint/rules/core/consistent_return.go` because
// `@typescript-eslint/consistent-return` needs the same judgment and a rule package may not import
// another rule package: `TestRulePackagesStayLeaves` measures the cost at a 1.8s leaf rebuild
// against 8.5s for a deep one, paid by everyone on every edit.
//
// The extension rule delegates rather than re-deriving this, which is what upstream does
// (`baseRule.create(context)`), and the reason is not style: two implementations of one question
// give the two rules two chances to disagree about it. This rule in particular renders THREE
// messages whose only difference in a fixture is a function name assembled from nine modifiers, and
// computes a different report span per node kind. Re-deriving that would be re-deriving the part
// most likely to drift.
//
// # What the extension needs, and why it is two hooks rather than a filter over the output
//
// typescript-eslint's wrapper does not filter findings. It intercepts the ReturnStatement listener
// BEFORE the core sees it, and does two things a post-filter cannot express:
//
//	a bare `return` in a function typed `void` is dropped entirely, so it never sets the
//	  expectation the first return sets, and never contradicts a later one
//	under `treatUndefinedAsUnspecified`, a `return <expr>` whose expression's TYPE is undefined is
//	  handed to the core as though it were a bare return
//
// Both change what the core is asked about rather than what it answered, so both are inputs. The
// same distinction was found and written down for `no-dupe-class-members`, where filtering the
// output reported on a member whose own key was not computed.
package consistentreturn

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
)

// Settings is the option surface both rules share.
type Settings struct {
	// TreatUndefinedAsUnspecified makes `return undefined` and `return void 0` count as returning
	// nothing, so a function mixing them with a bare `return` is consistent.
	TreatUndefinedAsUnspecified bool
}

// Hooks are the type-driven decisions the extension supplies and the core rule does not have.
//
// A zero Hooks is exactly the core rule: every field nil means every question falls back to the
// syntactic answer, which is what makes this package's default behaviour byte-identical to the
// rule it was lifted from.
type Hooks struct {
	// SuppressBareReturn is asked about a `return;` with no argument, before it is recorded at all.
	// Upstream's `isReturnVoidOrThenableVoid`: a bare return in a function whose declared return
	// type is `void`, or `Promise<void>` for an async one, is not a value-channel decision and
	// upstream drops it.
	//
	// Nil means never suppress, which is the core rule.
	SuppressBareReturn func(scope *ast.Node, returnStatement *ast.Node) bool

	// TreatsArgumentAsUnspecified is asked about a `return <expr>` when TreatUndefinedAsUnspecified
	// is set, after the syntactic test has already declined. Upstream reads the ARGUMENT's type and
	// treats an exactly-undefined one as a bare return, which catches `return undef` for a
	// `declare const undef: undefined` that the syntactic test cannot see.
	//
	// Nil means the syntactic test is the whole answer, which is the core rule.
	TreatsArgumentAsUnspecified func(argument *ast.Node) bool
}

// Reporter receives one finding. Both rules render their own message text from the pieces.
type Reporter struct {
	// ReportNode reports at a whole node, which is where the two per-return findings point.
	ReportNode func(node *ast.Node, messageId string, description string)
	// ReportRange reports at a computed span, which is where the end-of-function finding points.
	ReportRange func(textRange core.TextRange, messageId string, description string)
	// TokenRange trims a node to its first token, the way rule.TokenRange does. Passed in rather
	// than imported so this package stays below `rule`.
	TokenRange func(node *ast.Node) core.TextRange
}

// Message ids, which are upstream's and are asserted by both corpora.
const (
	MessageIdMissingReturn         = "missingReturn"
	MessageIdMissingReturnValue    = "missingReturnValue"
	MessageIdUnexpectedReturnValue = "unexpectedReturnValue"
)

// Judge walks one file and reports every disagreement, descending into nested functions.
//
// `sourceFile` is the file node, which is itself a scope: upstream treats the Program as a function
// for this purpose and its corpus asserts a finding named "Program".
//
// # Findings come out in SOURCE order, which is not the order the scopes are judged in
//
// Each scope is judged as a whole, because the end-of-function question needs every return in it,
// and the scopes are walked outermost-first. Upstream instead fires a listener per return statement
// during one traversal, so its findings arrive in source order regardless of nesting.
//
// The difference is visible whenever an inner and an outer function both report. Upstream's own
// corpus for the typescript extension writes exactly one such case -- an inner `baz` reporting at
// line 6 and its enclosing `bar` at line 9 -- and it is the only input in either corpus that can see
// this: the bare core rule's 143 rows contain nested functions but none where both report, so the
// ordering was unconstrained until that case arrived.
//
// Collected and sorted rather than restructured into a per-return listener. The scope-at-a-time
// shape is what makes the end-of-function judgment expressible at all, and sorting is a smaller
// change than inverting the traversal for a difference that is only ever about output order.
func Judge(sourceFile *ast.Node, settings Settings, hooks Hooks, reporter Reporter,
	describe func(messageId string, name string) string) {
	var collected []finding
	collector := Reporter{
		ReportNode: func(node *ast.Node, messageId string, description string) {
			collected = append(collected, finding{
				position: reporter.TokenRange(node).Pos(), node: node,
				messageId: messageId, description: description,
			})
		},
		ReportRange: func(textRange core.TextRange, messageId string, description string) {
			collected = append(collected, finding{
				position: textRange.Pos(), textRange: &textRange,
				messageId: messageId, description: description,
			})
		},
		TokenRange: reporter.TokenRange,
	}

	judgeScope(sourceFile, settings, hooks, collector, describe)

	// Stable, so two findings anchored at the same position keep the order the judgment produced
	// them in. The end-of-function finding for the Program is anchored at offset 0 and would
	// otherwise race anything else reported there.
	sort.SliceStable(collected, func(left, right int) bool {
		return collected[left].position < collected[right].position
	})
	for _, item := range collected {
		if item.textRange != nil {
			reporter.ReportRange(*item.textRange, item.messageId, item.description)
			continue
		}
		reporter.ReportNode(item.node, item.messageId, item.description)
	}
}

// finding is one pending report, held until every scope has been judged so they can be emitted in
// source order.
type finding struct {
	position    int
	node        *ast.Node
	textRange   *core.TextRange
	messageId   string
	description string
}

// judgeScope judges one function-like scope and descends into those below it.
func judgeScope(scope *ast.Node, settings Settings, hooks Hooks, reporter Reporter,
	describe func(messageId string, name string) string) {
	var returns []*ast.Node
	var nested []*ast.Node

	var visit func(*ast.Node) bool
	visit = func(current *ast.Node) bool {
		if IsScope(current) {
			// A nested function owns its own returns and is judged on its own pass. This is what
			// upstream's code path stack does when it pushes a frame.
			nested = append(nested, current)
			return false
		}
		if current.Kind == ast.KindReturnStatement {
			// The extension's first hook, and it is asked HERE rather than at report time. A
			// suppressed bare return must not be recorded at all: recorded, it would be the first
			// return and would set the expectation every later return is measured against, which
			// is a different answer from dropping the finding it produces.
			if current.AsReturnStatement().Expression == nil &&
				hooks.SuppressBareReturn != nil && hooks.SuppressBareReturn(scope, current) {
				return false
			}
			returns = append(returns, current)
		}
		current.ForEachChild(visit)
		return false
	}
	// A class member's body is reached through the class, so the descent must not stop at a class
	// declaration; only a function-like node opens a new frame.
	scope.ForEachChild(visit)

	judgeReturns(scope, returns, settings, hooks, reporter, describe)

	for _, function := range nested {
		judgeScope(function, settings, hooks, reporter, describe)
	}
}

// judgeReturns runs both judgments over one function's own returns.
func judgeReturns(scope *ast.Node, returns []*ast.Node, settings Settings, hooks Hooks,
	reporter Reporter, describe func(messageId string, name string) string) {
	if len(returns) == 0 {
		return
	}

	// A constructor's `return` is a language feature rather than a value channel, so upstream skips
	// both judgments for one. Checked before anything is reported rather than only at the
	// end-of-function judgment, because upstream's ReturnStatement listener also runs inside a
	// constructor and its first return simply sets an expectation nothing later contradicts... and
	// measured, `class A { constructor() { if (a) return true; else return; } }` DOES report the
	// per-return finding upstream. So only the end-of-function judgment is exempt, which is where
	// upstream's isClassConstructor test actually sits.
	firstReturnHasValue := HasValue(returns[0], settings, hooks)
	expectationMessageId := MessageIdUnexpectedReturnValue
	if firstReturnHasValue {
		expectationMessageId = MessageIdMissingReturnValue
	}
	capitalisedName := Name(scope, true)

	for _, returnStatement := range returns[1:] {
		if HasValue(returnStatement, settings, hooks) == firstReturnHasValue {
			continue
		}
		reporter.ReportNode(returnStatement, expectationMessageId,
			describe(expectationMessageId, capitalisedName))
	}

	// The end-of-function judgment. Upstream reads `funcInfo.hasReturnValue`, which is set from the
	// FIRST return alone and never updated, so a function whose first return is bare and whose
	// later return carries a value is NOT reported here even though a value does escape it. That
	// looks like a defect and it is upstream's behaviour; its own corpus case
	// `function foo() { if (true) return; else return false; }` reports only the per-return
	// finding. Reproduced by reading the first return rather than any return.
	if !firstReturnHasValue {
		return
	}
	if isExemptFromEndJudgment(scope) {
		return
	}
	if !canRunOffEnd(scope) {
		return
	}

	reporter.ReportRange(ReportRange(scope, reporter.TokenRange), MessageIdMissingReturn,
		describe(MessageIdMissingReturn, Name(scope, false)))
}

// HasValue says whether a return statement carries a value, under the options and hooks.
//
// `treatUndefinedAsUnspecified` makes `return undefined` and `return void 0` count as returning
// nothing. Upstream tests `isSpecificId(argument, "undefined")` and `argument.operator !== "void"`,
// so it matches the bare identifier `undefined` and any `void` expression, not only `void 0`.
func HasValue(returnStatement *ast.Node, settings Settings, hooks Hooks) bool {
	argument := returnStatement.AsReturnStatement().Expression
	if argument == nil {
		return false
	}
	if !settings.TreatUndefinedAsUnspecified {
		return true
	}
	// NOT unwrapping parentheses here, because upstream does not either: its parser folds them away
	// so `return (undefined)` arrives as a bare identifier there and is treated as unspecified,
	// while ours keeps the node and treats it as a value. Recorded as a divergence rather than
	// corrected, because correcting it would be a silent improvement on a rule whose whole option
	// surface is about which spellings count as nothing, and upstream's corpus writes no
	// parenthesized form to settle the intent.
	if argument.Kind == ast.KindVoidExpression {
		return false
	}
	if argument.Kind == ast.KindIdentifier && argument.Text() == "undefined" {
		return false
	}
	// The extension's second hook. Upstream's wrapper asks the checker whether the argument's type
	// is exactly `undefined` and, if so, hands the core rule a synthesised argument-less return.
	// Reaching the same answer by returning false here rather than by fabricating an AST node,
	// because nothing downstream reads the argument again.
	if hooks.TreatsArgumentAsUnspecified != nil && hooks.TreatsArgumentAsUnspecified(argument) {
		return false
	}
	return true
}

// IsScope says whether a node owns its own returns.
//
// Every function-like construct, including a constructor and both accessors, because each is a
// separate code path frame upstream. A class declaration is deliberately absent: its members are
// reached by descending through it.
func IsScope(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindConstructor:
		return true
	}
	return false
}

// isExemptFromEndJudgment reproduces upstream's two constructor exemptions.
func isExemptFromEndJudgment(node *ast.Node) bool {
	// A class constructor. Our parser gives it its own kind where upstream reads a
	// FunctionExpression whose parent is a MethodDefinition of kind "constructor", which is why an
	// object literal's method named `constructor` is NOT caught by this and does report.
	if node.Kind == ast.KindConstructor {
		return true
	}

	// An ES5 constructor: `astUtils.isES5Constructor` is a function whose own name starts with an
	// upper case letter. Upstream tests `node.id && node.id.name[0] !== node.id.name[0].toLowerCase()`,
	// which is a case comparison rather than an ASCII range, so a name with no case distinction at
	// all -- a digit cannot start an identifier, but `_foo` and `$foo` can -- is NOT a constructor.
	// The corpus asserts `function _foo()` reports, which is the case that separates the two
	// readings.
	if node.Kind != ast.KindFunctionDeclaration && node.Kind != ast.KindFunctionExpression {
		return false
	}
	name := node.Name()
	if name == nil {
		return false
	}
	text := name.Text()
	if text == "" {
		return false
	}
	first := text[:1]
	return first != strings.ToLower(first)
}

// canRunOffEnd reports whether control can reach the end of a function body.
//
// Upstream's `isAnySegmentReachable(funcInfo.currentSegments)` at the function's exit.
// `control_flow_graph.Graph.EndReachable` is documented as answering that question, and
// `array_callback_return` already calls it for the identical purpose.
//
// A nil graph answers true, which is the conservative direction: it reports rather than going
// silent, so a shape the builder cannot model surfaces instead of disappearing.
//
// # A measured over-report, in the shelf rather than in this judgment
//
// A dry run over the ahra tree reported 141 findings where the installed rule reports 130. The 130
// agree exactly, same file, same line, same column, same message id, and there are ZERO findings
// the installed rule produces that this one misses. All 11 extra come from one shape, isolated by
// bisection down to a minimal pair:
//
//	try { const r = g(); return r; } catch (e) { throw e; } finally { k(); }   EndReachable TRUE
//	try { return 1; }                catch (e) { throw e; } finally { k(); }   EndReachable false
//	try { const r = g(); return r; } catch (e) { throw e; }                    EndReachable false
//
// So a variable declaration inside a `try` that also has a `finally` makes the graph believe
// control reaches the function's end when every path returns or throws. Removing either the
// declaration or the `finally` gives the right answer, which is what places the defect in
// `control_flow_graph` rather than here: this judgment asks one question and the shelf answers it
// wrongly for that one shape.
//
// Not repaired here, and not worked around here either. A workaround would be this judgment
// disagreeing with the shelf about reachability, which is worse than an over-report: the next rule
// to ask the same question would get the old answer and nobody would know the two had diverged.
// Every site is an `async` function with a `try`/`catch`/`finally` in this tree, and each is a real
// mixed-return function whose end genuinely cannot be reached, so the findings are false positives
// rather than corruption.
func canRunOffEnd(node *ast.Node) bool {
	graph := control_flow_graph.Build(node, control_flow_graph.Hooks[struct{}]{})
	if graph == nil {
		return true
	}
	return graph.EndReachable
}

// Name renders upstream's `getFunctionNameWithKind`, plus the Program case.
//
// Not the shelf's `arrayCallbackFunctionNameWithKind`, which answers a narrower question: it knows
// only `arrow function`, `function` and `function 'name'`, because those are the only renderings
// its corpus asserts. This rule's corpus asserts nine more -- `method 'foo'`, `getter 'foo'`,
// `setter 'foo'`, `static method 'foo'`, `private method #foo`, `static private method #foo`,
// `async function 'foo'`, `generator function 'foo'`, `async generator function 'foo'` -- so the
// two are prefixed apart rather than merged. Merging would widen every message the other rule
// renders.
//
// `capitalise` is the `upperCaseFirst` upstream applies to the per-return messages and not to
// `missingReturn`. The corpus asserts both spellings, and the Program case asserts them in
// opposite directions, so the flag is load-bearing rather than cosmetic.
//
// Every rendering below was measured by driving the installed rule rather than derived from the
// helper's source.
func Name(node *ast.Node, capitalise bool) string {
	if node.Kind == ast.KindSourceFile {
		if capitalise {
			return "Program"
		}
		return "program"
	}

	var tokens []string

	// The proposal spells `static` before a visibility word, which is why the order here is static
	// then private rather than the reverse.
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsStatic) {
		tokens = append(tokens, "static")
	}
	privateName := ""
	if name := node.Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
		tokens = append(tokens, "private")
		// `Text()` on a private identifier already carries the leading hash, and upstream renders
		// it unquoted where an ordinary name is quoted.
		privateName = name.Text()
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
		tokens = append(tokens, "async")
	}
	if IsGenerator(node) {
		tokens = append(tokens, "generator")
	}

	switch node.Kind {
	case ast.KindConstructor:
		// Upstream returns the bare word with no modifiers at all, discarding whatever it pushed.
		return capitaliseFirst("constructor", capitalise)
	case ast.KindGetAccessor:
		tokens = append(tokens, "getter")
	case ast.KindSetAccessor:
		tokens = append(tokens, "setter")
	case ast.KindMethodDeclaration:
		tokens = append(tokens, "method")
	default:
		if node.Kind == ast.KindArrowFunction {
			tokens = append(tokens, "arrow")
		}
		tokens = append(tokens, "function")
		// A function expression assigned to a property or a class field is a `method` upstream,
		// because it reads the PARENT's node type rather than the function's. Measured:
		// `var o = { foo: function() {...} }` renders "method 'foo'", not "function 'foo'".
		if node.Kind == ast.KindFunctionExpression && node.Parent != nil {
			switch node.Parent.Kind {
			case ast.KindPropertyAssignment, ast.KindPropertyDeclaration:
				tokens = tokens[:len(tokens)-1]
				tokens = append(tokens, "method")
			}
		}
	}

	if privateName != "" {
		tokens = append(tokens, privateName)
		return capitaliseFirst(strings.Join(tokens, " "), capitalise)
	}
	if name := staticName(node); name != "" {
		tokens = append(tokens, "'"+name+"'")
	}
	return capitaliseFirst(strings.Join(tokens, " "), capitalise)
}

// staticName is the name upstream quotes into the message, or "" for none.
//
// Upstream reads `getStaticPropertyName(parent)` for a member and falls back to `node.id.name`, so
// a computed key whose value the syntax does not settle renders no name at all: measured,
// `class A { [x]() { if (a) return true; } }` renders "method." with no name, while
// `class A { ['computed']() {...} }` renders "method 'computed'". That is exactly the question
// `property.Name` answers, including its refusal to read a bare identifier through brackets, so it
// is a caller rather than a reimplementation.
func staticName(node *ast.Node) string {
	// A function expression takes its parent's key when it has one, which is what makes
	// `{ foo: function() {} }` render `'foo'` despite the function being anonymous.
	if node.Kind == ast.KindFunctionExpression && node.Parent != nil {
		switch node.Parent.Kind {
		case ast.KindPropertyAssignment, ast.KindPropertyDeclaration:
			if name, ok := property.Name(node.Parent.Name(), property.Static); ok {
				return name
			}
		}
	}
	name := node.Name()
	if name == nil {
		return ""
	}
	if text, ok := property.Name(name, property.Static); ok {
		return text
	}
	return ""
}

// capitaliseFirst is upstream's `upperCaseFirst`, applied only to the per-return messages.
func capitaliseFirst(text string, capitalise bool) string {
	if !capitalise || text == "" {
		return text
	}
	return strings.ToUpper(text[:1]) + text[1:]
}

// IsGenerator says whether a function carries the `*`.
func IsGenerator(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

// Verb renders the middle of the two per-return messages.
//
// Upstream spells these as `{{name}} expected a return value.` and
// `{{name}} expected no return value.`, so the verb differs with the message rather than with the
// name and there is exactly one of each.
func Verb(messageId string) string {
	if messageId == MessageIdMissingReturnValue {
		return "expected a return value."
	}
	return "expected no return value."
}

// ReportRange is where the `missingReturn` finding points.
//
// Upstream computes a `loc` per node kind rather than reporting the whole function, and the four
// answers are distinct enough that a port pointing at the function would pass every message-id
// fixture while being wrong on all of them:
//
//	Program                 line 1 column 0, the head of the file
//	ArrowFunctionExpression the `=>` token
//	a method or an object    the key
//	  literal method
//	anything else           the function's name, or the `function` keyword when it has none
//
// The columns the corpus asserts pin all four. `f(() => { if (a) return true; })` reports at column
// 6, which is the `=>`; `f(function() {...})` reports at column 3, which is the `function` keyword;
// `f(function foo() {...})` reports at column 12, which is the name rather than the keyword; and
// `var obj = {foo() {...}}` reports at column 12, which is the key.
func ReportRange(node *ast.Node, tokenRange func(*ast.Node) core.TextRange) core.TextRange {
	if node.Kind == ast.KindSourceFile {
		// The head of the program. A zero-length range at the first character rather than the whole
		// file, so the finding points at line 1 column 1 the way upstream's `{line: 1, column: 0}`
		// does.
		//
		// A mutation removing this arm survives, and it is equivalent rather than untested:
		// `TokenRange` on a source file also begins at offset 0, so both paths report at the same
		// position. Its INVERSE, moving the range to offset 1, fails five lines, which proves the
		// fixtures can see the program span. Kept because it states the intent -- upstream reports
		// the head of the program rather than the whole file -- and because it does not depend on
		// what `TokenRange` happens to answer for a construct with no leading token.
		return core.NewTextRange(0, 0)
	}

	if node.Kind == ast.KindArrowFunction {
		// The `=>` token. Found by scanning from the end of the parameter list rather than by
		// taking a child, because the arrow is a token rather than a node here.
		if arrow := node.AsArrowFunction().EqualsGreaterThanToken; arrow != nil {
			return tokenRange(arrow)
		}
		return tokenRange(node)
	}

	// A method, an accessor or a constructor points at its key, which upstream reads as
	// `node.parent.key`. There is deliberately no arm for those kinds here, because the fallback at
	// the end of this function already reads `node.Name()` and a method's name IS its key, so an
	// explicit arm would call the same accessor and return the same range.
	//
	// This was measured rather than assumed. An arm was written first, and a mutation neutralising
	// it survived the whole fixture set including five cases asserting a method's span against a
	// modifier. Its INVERSE -- returning the whole node from that arm -- fails nine lines, which
	// proves the fixtures can see the span and that the arm was simply subsumed rather than
	// untested. Removed, with the reasoning here so it is not helpfully restored.
	//
	// A function expression assigned to a property points at the property's key, because upstream
	// reads the parent there too.
	if node.Kind == ast.KindFunctionExpression && node.Parent != nil {
		switch node.Parent.Kind {
		case ast.KindPropertyAssignment, ast.KindPropertyDeclaration:
			if name := node.Parent.Name(); name != nil {
				return tokenRange(name)
			}
		}
	}

	// The function's own name, or the `function` keyword when it is anonymous. `TokenRange` trims
	// to the first token, which for an anonymous function expression is `function` itself, and for
	// an async one is `async` -- matching upstream's `getFirstToken`, measured at column 20 for
	// `const f = async () => {...}` where the `=>` branch applies instead.
	if name := node.Name(); name != nil {
		return tokenRange(name)
	}
	return tokenRange(node)
}
