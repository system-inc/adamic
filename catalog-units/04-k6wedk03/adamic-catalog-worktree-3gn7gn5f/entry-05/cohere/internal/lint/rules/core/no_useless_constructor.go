package core

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoUselessConstructor = rule.Message{
	Id: "noUselessConstructor",
	Description: "This constructor does nothing the language would not do without it. An empty " +
		"constructor on a base class, or one whose only statement forwards its own parameters to " +
		"`super`, is exactly the behaviour a class gets when it declares no constructor at all, so " +
		"the code reads as though something happens at construction time when nothing does.",
}

var messageNoUselessConstructorRemove = rule.Message{
	Id:          "removeConstructor",
	Description: "Remove the constructor, leaving the behaviour the class already had.",
}

// NoUselessConstructor flags a constructor that could be deleted with no change in behaviour.
//
//	valid:   class A { constructor() { doSomething(); } }
//	valid:   class A { private constructor() {} }
//	valid:   class A extends B { constructor(x) { super(y); } }
//	invalid: class A { constructor() {} }
//	invalid: class A extends B { constructor() { super(); } }
//	invalid: class A extends B { constructor(...args) { super(...args); } }
//
// # Two judgments, chosen by whether the class extends anything
//
// A base class constructor is useless when its body is empty. A subclass constructor is useless when
// its body is exactly one `super(...)` call that passes its own parameters straight through, because
// that is what an absent constructor does implicitly.
//
// Passing through has two accepted spellings and both are upstream's. `super(...arguments)` forwards
// everything whatever the parameters are. Otherwise the parameter list and the argument list must
// correspond one for one, each pair being either the same identifier or a rest matched by a spread
// of the same identifier. Anything else, including a reordering or a renamed argument, means the
// constructor does something.
//
// Every parameter must also be SIMPLE, an identifier or a rest. A default or a destructuring pattern
// can run arbitrary code while binding, so `constructor(x = f()) { super(x); }` is not useless even
// though it forwards.
//
// # Four things make a constructor useful regardless
//
// A parameter property (`constructor(private x: number)`) declares and assigns a field, so the
// parameter list is the body. A parameter decorator runs. And accessibility is the subtle one,
// measured against the installed eslint at 10.8.1 rather than reasoned about:
//
//	private constructor      always useful, it restricts who may construct
//	protected constructor    always useful, same reason
//	public constructor       useful ONLY on a subclass
//
// That last row is not an oversight. On a base class `public` says what was already true, so the
// constructor is still useless and upstream reports it. On a subclass it widens the visibility the
// parent declared, which is a real effect. Both directions are pinned by fixtures.
//
// # The typescript-eslint extension of this rule adds nothing, and that is measured
//
// `@typescript-eslint/no-useless-constructor` exists upstream and is not ported here. It is a
// wrapper over this same core rule that lays three filters over the member listener: skip a
// protected constructor, skip a private one, skip a public one when the class extends, and skip any
// constructor carrying a parameter property or a decorated parameter.
//
// Those are the four judgments the section above already makes, so the extension narrows nothing
// that this rule has not already narrowed. Established by running upstream's core and upstream's
// extension over identical inputs on the installed 8.67.0 build, three ways:
//
//	upstream's own 41-case corpus FOR THE EXTENSION   identical verdicts on all 41
//	13 shapes written for this comparison            identical on all 13
//	11 adversarial shapes, aimed one per filter       identical on all 11
//
// Sixty-five inputs, no divergence. This rule was then run against that same extension corpus and
// reproduced it exactly, 32 of 32 passing and 9 of 9 failing.
//
// So porting the extension would register a second NAME rather than a second check, and the name is
// not free: a rule package may not import another rule package, so a wrapper would need this file's
// body lifted into `internal/utilities` the way `no-dupe-class-members` was at `27c5ff0`. That is
// 427 lines and sixteen helpers moved for no behavioural change.
//
// Nothing asks for the name either. `TestParityAgainstInventory` walks REGISTERED rules and asks
// whether each appeared in the captured rule inventory, not the reverse, and it carried no entry for
// this rule under either spelling, so a missing `@typescript-eslint/` registration costs nothing in
// the differential. Checked with a control against a rule that does have an entry.
//
// This note exists because the audit lists the extension as an unported line item, which reads as
// work outstanding. It is not, and re-deriving that has cost more than one reader a cycle.
//
// # A suggestion, not a fix
//
// Upstream sets `hasSuggestions` and no `fixable`. Deleting a constructor is a change to what the
// class declares rather than to how it is spelled, so nothing may apply it unattended.
var NoUselessConstructor = rule.Rule{
	Name: "no-useless-constructor",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindConstructor: func(node *ast.Node) {
				constructor := node.AsConstructorDeclaration()
				// An overload signature and a constructor in a `declare class` have no body.
				// Upstream guards this explicitly, with a comment naming the parsers that produce
				// it, and ours is one of them.
				if constructor.Body == nil || constructor.Body.Kind != ast.KindBlock {
					return
				}
				if uselessConstructorHasDecoratedOrPropertyParameter(constructor) {
					return
				}
				if uselessConstructorHasUsefulAccessibility(node) {
					return
				}

				statements := constructor.Body.AsBlock().Statements.Nodes
				if uselessConstructorClassExtends(node) {
					if !uselessConstructorIsRedundantSuperCall(statements, constructor) {
						return
					}
				} else if len(statements) != 0 {
					return
				}

				ctx.ReportRangeWithSuggestions(uselessConstructorHeadRange(ctx, node),
					messageNoUselessConstructor, rule.Suggestion{
						Message: messageNoUselessConstructorRemove,
						Fixes:   []rule.Fix{uselessConstructorRemoval(ctx, node)},
					})
			},
		}
	},
}

// uselessConstructorHeadRange is the span upstream reports: the member start through the token
// before the parameter list.
//
// That covers an accessibility modifier, so `public constructor() {}` reports `public constructor`
// rather than `constructor`. Measured at columns 11 to 29 against the installed build.
//
// The end is found by scanning to the opening parenthesis rather than by asking for a token list,
// because the only thing between the member start and that parenthesis is modifiers and the keyword,
// and neither can contain one.
func uselessConstructorHeadRange(ctx rule.Context, node *ast.Node) core.TextRange {
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	text := ctx.SourceFile.Text()
	end := nodeRange.Pos()
	for end < nodeRange.End() && text[end] != '(' {
		end++
	}
	// Back off the whitespace between the keyword and the parenthesis, which upstream excludes by
	// ending at the previous token rather than at the paren.
	for end > nodeRange.Pos() && (text[end-1] == ' ' || text[end-1] == '\t' ||
		text[end-1] == '\n' || text[end-1] == '\r') {
		end--
	}
	return core.NewTextRange(nodeRange.Pos(), end)
}

// uselessConstructorRemoval deletes the whole constructor, leaving a semicolon when one is needed.
//
// The semicolon is not cosmetic and this is the part of the rule most likely to be got wrong. A
// class field written without a trailing semicolon is terminated by automatic semicolon insertion,
// and that insertion only happens because the next line cannot continue the expression. Delete a
// constructor from between them and the field can suddenly reach the member after it:
//
//	class A { foo = 'bar'    constructor() {}    [0]() {} }
//
// becomes `foo = 'bar'` followed by `[0]() {}`, which parses as `foo = 'bar'[0]()`, an index into
// the string. Four of upstream's own reporting cases assert exactly this and a first draft of this
// port failed all four, having claimed in a doc comment that the corpus wrote no case needing it.
//
// Upstream decides with two questions and both are reproduced. The token that would follow must be
// one that can continue an expression, which is `[`, `*`, `in` or `instanceof` and nothing else.
// And the member before must be one that could be continued, which a method or an
// already-semicolon-terminated field is not.
func uselessConstructorRemoval(ctx rule.Context, node *ast.Node) rule.Fix {
	nodeRange := rule.TokenRange(ctx.SourceFile, node)
	if uselessConstructorRemovalNeedsSemicolon(ctx, node) {
		return rule.ReplaceRange(nodeRange, ";")
	}
	return rule.RemoveRange(nodeRange)
}

// uselessConstructorRemovalNeedsSemicolon reports whether deleting this constructor would let the
// preceding member run into the following one.
func uselessConstructorRemovalNeedsSemicolon(ctx rule.Context, node *ast.Node) bool {
	previous, next := uselessConstructorSiblings(node)
	if previous == nil || next == nil {
		return false
	}
	if !uselessConstructorCanBeContinued(previous, ctx) {
		return false
	}
	return uselessConstructorCanContinueExpression(ctx, next)
}

// uselessConstructorSiblings returns the class members either side of a member.
func uselessConstructorSiblings(node *ast.Node) (*ast.Node, *ast.Node) {
	parent := node.Parent
	if parent == nil {
		return nil, nil
	}
	classLike := parent.ClassLikeData()
	if classLike == nil || classLike.Members == nil {
		return nil, nil
	}
	members := classLike.Members.Nodes
	for index, member := range members {
		if member != node {
			continue
		}
		var previous, next *ast.Node
		if index > 0 {
			previous = members[index-1]
		}
		if index+1 < len(members) {
			next = members[index+1]
		}
		return previous, next
	}
	return nil, nil
}

// uselessConstructorCanBeContinued reports whether a class member's text could run on.
//
// Three conditions, and the middle one was measured rather than assumed. It must be a property
// declaration, because a method's body ends in a brace and nothing continues that. It must have an
// INITIALIZER, because a bare `foo` is only a name and there is no expression to continue; upstream
// leaves no semicolon for `class A { foo <constructor> [0]() {} }` and a draft testing only for a
// property inserted one. And it must not already end in a semicolon, which terminates it.
func uselessConstructorCanBeContinued(member *ast.Node, ctx rule.Context) bool {
	if member.Kind != ast.KindPropertyDeclaration {
		return false
	}
	if member.AsPropertyDeclaration().Initializer == nil {
		return false
	}
	text := ctx.SourceFile.Text()
	end := member.End()
	for end > 0 && (text[end-1] == ' ' || text[end-1] == '\t' ||
		text[end-1] == '\n' || text[end-1] == '\r') {
		end--
	}
	return end > 0 && text[end-1] != ';'
}

// uselessConstructorCanContinueExpression is upstream's `canContinueExpressionInClassBody`.
//
// Four token values and no others: a computed key opens with `[`, a generator with `*`, and `in` and
// `instanceof` are binary operators that read as a continuation rather than as a member name.
// Measured against the installed build, a plain `bar() {}` following needs no semicolon.
func uselessConstructorCanContinueExpression(ctx rule.Context, member *ast.Node) bool {
	memberRange := rule.TokenRange(ctx.SourceFile, member)
	text := ctx.SourceFile.Text()[memberRange.Pos():memberRange.End()]
	switch {
	case strings.HasPrefix(text, "["), strings.HasPrefix(text, "*"):
		return true
	// `instanceof` is tested BEFORE `in`, because `in` is its prefix: testing `in` first answers
	// false for `instanceof` on the whole-word check and never reaches the longer arm. A draft with
	// the two the other way round failed upstream's own `instanceof` case.
	case strings.HasPrefix(text, "instanceof"):
		return uselessConstructorIsWholeWord(text, "instanceof")
	case strings.HasPrefix(text, "in"):
		return uselessConstructorIsWholeWord(text, "in")
	}
	return false
}

// uselessConstructorIsWholeWord reports whether a keyword prefix is the whole identifier rather than
// the start of a longer one, so `input` does not read as `in`.
func uselessConstructorIsWholeWord(text string, word string) bool {
	if len(text) == len(word) {
		return true
	}
	next := text[len(word)]
	return !(next == '_' || next == '$' ||
		(next >= 'a' && next <= 'z') || (next >= 'A' && next <= 'Z') ||
		(next >= '0' && next <= '9'))
}

// uselessConstructorClassExtends reports whether the enclosing class has a superclass.
//
// `HeritageClauses` is nil for a class that extends nothing, which is most of them, so the nil test
// is load-bearing rather than defensive: a probe written before this rule existed panicked on
// `class A { constructor() {} }` for exactly that reason, and a panic costs every rule the file.
//
// `implements` is a heritage clause too, so the token has to be checked rather than the presence of
// a clause. A class that only implements an interface has no superclass and its empty constructor is
// judged as a base class.
func uselessConstructorClassExtends(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	classLike := parent.ClassLikeData()
	if classLike == nil || classLike.HeritageClauses == nil {
		return false
	}
	for _, clause := range classLike.HeritageClauses.Nodes {
		if clause.AsHeritageClause().Token == ast.KindExtendsKeyword {
			return true
		}
	}
	return false
}

// uselessConstructorHasUsefulAccessibility reports whether a modifier makes the constructor matter.
//
// See the rule's doc comment for why `public` depends on the superclass and the other two do not.
func uselessConstructorHasUsefulAccessibility(node *ast.Node) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		switch modifier.Kind {
		case ast.KindPrivateKeyword, ast.KindProtectedKeyword:
			return true
		case ast.KindPublicKeyword:
			return uselessConstructorClassExtends(node)
		}
	}
	return false
}

// uselessConstructorHasDecoratedOrPropertyParameter reports whether any parameter does work.
//
// A parameter property declares and assigns a field; a parameter decorator runs. Either way the
// parameter list is doing something a missing constructor would not.
func uselessConstructorHasDecoratedOrPropertyParameter(constructor *ast.ConstructorDeclaration) bool {
	if constructor.Parameters == nil {
		return false
	}
	for _, parameter := range constructor.Parameters.Nodes {
		if len(parameter.Decorators()) > 0 {
			return true
		}
		if modifiers := parameter.Modifiers(); modifiers != nil {
			for _, modifier := range modifiers.Nodes {
				// A decorator is carried in the modifier list too, so it is skipped here rather
				// than counted twice; every other modifier on a parameter makes it a property.
				if modifier.Kind != ast.KindDecorator {
					return true
				}
			}
		}
	}
	return false
}

// uselessConstructorIsRedundantSuperCall reports whether a subclass constructor only forwards.
func uselessConstructorIsRedundantSuperCall(statements []*ast.Node,
	constructor *ast.ConstructorDeclaration) bool {
	call := uselessConstructorSoleSuperCall(statements)
	if call == nil {
		return false
	}
	parameters := []*ast.Node{}
	if constructor.Parameters != nil {
		parameters = constructor.Parameters.Nodes
	}
	// Every parameter must bind without running anything. A default or a destructuring pattern can
	// call a function while binding, so forwarding it is not the same as not existing.
	for _, parameter := range parameters {
		if !uselessConstructorIsSimpleParameter(parameter) {
			return false
		}
	}

	arguments := []*ast.Node{}
	if call.AsCallExpression().Arguments != nil {
		arguments = call.AsCallExpression().Arguments.Nodes
	}
	if uselessConstructorIsSpreadArguments(arguments) {
		return true
	}
	return uselessConstructorIsPassingThrough(parameters, arguments)
}

// uselessConstructorSoleSuperCall returns the call when the body is exactly one `super(...)`.
func uselessConstructorSoleSuperCall(statements []*ast.Node) *ast.Node {
	if len(statements) != 1 || statements[0].Kind != ast.KindExpressionStatement {
		return nil
	}
	expression := statements[0].AsExpressionStatement().Expression
	if expression == nil || expression.Kind != ast.KindCallExpression {
		return nil
	}
	callee := expression.AsCallExpression().Expression
	if callee == nil || callee.Kind != ast.KindSuperKeyword {
		return nil
	}
	return expression
}

// uselessConstructorIsSimpleParameter reports a parameter that binds without running anything.
//
// Upstream's `isSimple`: an identifier or a rest. A rest's own binding must be an identifier too,
// since `...{a}` destructures.
func uselessConstructorIsSimpleParameter(parameter *ast.Node) bool {
	declaration := parameter.AsParameterDeclaration()
	if declaration.Initializer != nil {
		return false
	}
	if declaration.Name() == nil {
		return false
	}
	return declaration.Name().Kind == ast.KindIdentifier
}

// uselessConstructorIsSpreadArguments reports the `super(...arguments)` spelling.
func uselessConstructorIsSpreadArguments(arguments []*ast.Node) bool {
	if len(arguments) != 1 || arguments[0].Kind != ast.KindSpreadElement {
		return false
	}
	inner := arguments[0].AsSpreadElement().Expression
	return inner != nil && inner.Kind == ast.KindIdentifier && inner.Text() == "arguments"
}

// uselessConstructorIsPassingThrough reports a one-for-one forwarding of every parameter.
func uselessConstructorIsPassingThrough(parameters []*ast.Node, arguments []*ast.Node) bool {
	if len(parameters) != len(arguments) {
		return false
	}
	for index, parameter := range parameters {
		if !uselessConstructorIsValidPair(parameter, arguments[index]) {
			return false
		}
	}
	return true
}

// uselessConstructorIsValidPair reports whether one parameter forwards to one argument unchanged.
//
// Two accepted shapes: the same identifier on both sides, or a rest parameter matched by a spread of
// the same identifier. A rest forwarded as a bare identifier is NOT a pass-through, because
// `super(args)` passes the array itself where `super(...args)` passes its elements.
func uselessConstructorIsValidPair(parameter *ast.Node, argument *ast.Node) bool {
	declaration := parameter.AsParameterDeclaration()
	name := declaration.Name()
	if name == nil {
		return false
	}

	if declaration.DotDotDotToken != nil {
		if argument.Kind != ast.KindSpreadElement {
			return false
		}
		inner := argument.AsSpreadElement().Expression
		return name.Kind == ast.KindIdentifier && inner != nil &&
			inner.Kind == ast.KindIdentifier && inner.Text() == name.Text()
	}

	return name.Kind == ast.KindIdentifier && argument.Kind == ast.KindIdentifier &&
		argument.Text() == name.Text()
}
