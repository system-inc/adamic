package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoGlobalListenerTargetAssertionId = "globalListenerTargetAssertion"

// correctnessNoGlobalListenerTargetAssertionText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-global-listener-target-assertion.json`.
var correctnessNoGlobalListenerTargetAssertionText = policy.MessageOf("nexus/correctness-no-global-listener-target-assertion", correctnessNoGlobalListenerTargetAssertionId)

func correctnessNoGlobalListenerTargetAssertionMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoGlobalListenerTargetAssertionId,
		Description: correctnessNoGlobalListenerTargetAssertionText.Render(nil),
	}
}

// CorrectnessNoGlobalListenerTargetAssertion reports an assertion of `event.target` to a specific
// element type inside a handler registered with `document.addEventListener` or
// `window.addEventListener`.
//
//	invalid: document.addEventListener('keydown', function(event) { const target = event.target as HTMLTextAreaElement; use(target.selectionStart); });
//	invalid: window.addEventListener('click', (event) => { (event.target as HTMLInputElement).select(); });
//	invalid: function handleKeyDown(event: KeyboardEvent) { (<HTMLInputElement>event.target).value; } document.addEventListener('keydown', handleKeyDown);
//	valid:   document.addEventListener('keydown', function(event) { if(!(event.target instanceof HTMLTextAreaElement)) return; use(event.target.selectionStart); });
//	valid:   document.addEventListener('click', function(event) { const target = event.target as HTMLElement; target.closest('[data-menu]'); });
//	valid:   textarea.addEventListener('keydown', function(event) { (event.target as HTMLTextAreaElement).selectionStart; });
//	valid:   document.addEventListener('click', function(event) { (event.currentTarget as Document).title; });
//
// # Where it came from
//
// Structure's `source/components/code/Code.tsx`, caught uncommitted by the own-history pass of the
// new-rules sweep (`#tevhg3f`, item 2), built in task `#j03vwm6`. The editor registered a `keydown`
// listener on `document` and read `event.target as HTMLTextAreaElement` to insert four spaces on Tab.
// Tab on any other control was swallowed, and with no caret to read on that control,
// `code.substring(undefined)` duplicated the editor's whole contents. The fix moved the handler onto
// the textarea as a React `onKeyDown` reading `event.currentTarget`.
//
// # What it matches, by declaration
//
//   - **The listener**: a call of `addEventListener` whose every declaration is in the default
//     library (lib.dom's), on the identifier `document` or `window` resolving to the global the
//     default library declares. A local `document`, a parameter named `window`, and an element's
//     own `addEventListener` are all other receivers.
//   - **The handler**: the second argument, written inline as a function or an arrow, or an
//     identifier naming a function declaration or a `const` function in this file. Its first
//     parameter is the event, followed by symbol, so a nested function's own `event` is not it.
//   - **The read**: `<event>.target`, where `target` is the default library's (`Event.target`),
//     anywhere in the handler's body, nested callbacks included, since they see the same event.
//   - **The assertion**: `as T` or `<T>` around that read, looking out through parentheses, `!` and
//     any other assertion (`event.target as unknown as HTMLInputElement`), where every non-nullish
//     member of `T` is a default-library interface that extends `Element` and is narrower than
//     `Element`, `HTMLElement`, `SVGElement` or `MathMLElement`.
//
// # What it leaves alone, and why
//
//   - **`as HTMLElement` and the other general types.** They name nothing a particular control
//     has, and the reads that follow them in practice are general too: `.closest()`, `.dataset`,
//     `.tagName`. The sweep counted three in ahra (`TasksCenter.tsx`, `TaskDetailFocusOverlay.tsx`,
//     `SourcesAccordion.tsx`), all benign. The lie that hurts is the one that promises
//     `selectionStart` or `value`.
//   - **An assertion the checker already agrees with.** When the type of `event.target` at that
//     point, narrowed by an `instanceof` guard before it, is assignable to `T`, the assertion
//     restates what is known and claims nothing.
//
// A guard the checker cannot see, `tagName === 'TEXTAREA'` or `event.target === textareaReference`
// before the assertion, still reports. The assertion is still unchecked by the type system, and
// `instanceof HTMLTextAreaElement` says the same thing and narrows, so the repair is the same one
// line. A listener on any other node (an element, `document.body`, a `ref.current`) is not read:
// its target is constrained to that node's subtree, which a rule cannot weigh. An assertion to a
// project's own interface (a custom element extending `HTMLElement`) is not read either, since only
// the default library's element types are known to promise members a general element lacks; that
// is a missed finding, never a false one.
//
// # No fix
//
// Whether to guard with `instanceof`, return early, or move the listener onto the element is the
// author's call, and an inserted guard would change which events the handler acts on.
var CorrectnessNoGlobalListenerTargetAssertion = rule.Rule{
	Name: "nexus/correctness-no-global-listener-target-assertion",

	// `document`, `window`, `addEventListener`, `Event.target` and the element types are each told
	// from a local of the same name by the declaration the checker resolves.
	NeedsTypeChecker: true,

	// Whether a declaration lives in the default library is read off the compiler options' lib list.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	// Left on the default, Contents: the dispatch scan cannot prove the handler resolution stops short of
	// an imported body, so a shape-keyed replay is not claimed (TestRulesClaimShapesOnlyWhereTheScanAllowsIt).

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil || ctx.Program == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				handler := correctnessNoGlobalListenerTargetAssertionHandler(ctx, node)
				if handler == nil {
					return
				}
				parameters := handler.Parameters()
				if len(parameters) == 0 {
					return
				}
				name := parameters[0].Name()
				if name == nil || name.Kind != ast.KindIdentifier || name.Text() == "this" {
					return
				}
				event := ctx.TypeChecker.GetSymbolAtLocation(name)
				body := handler.Body()
				if event == nil || body == nil {
					return
				}
				var visit func(current *ast.Node) bool
				visit = func(current *ast.Node) bool {
					if current.Kind == ast.KindPropertyAccessExpression {
						if assertion := correctnessNoGlobalListenerTargetAssertionAt(ctx, current, event); assertion != nil {
							ctx.ReportNode(assertion, correctnessNoGlobalListenerTargetAssertionMessage())
						}
					}
					current.ForEachChild(visit)
					return false
				}
				visit(body)
			},
		}
	},
}

// correctnessNoGlobalListenerTargetAssertionGeneralTypes are the element types that name no
// particular control.
var correctnessNoGlobalListenerTargetAssertionGeneralTypes = map[string]bool{
	"Element": true, "HTMLElement": true, "SVGElement": true, "MathMLElement": true,
}

// correctnessNoGlobalListenerTargetAssertionHandler is the function a call registers as a listener
// on the global `document` or `window`, or nil when the call is anything else or the handler is not a
// function this file declares.
func correctnessNoGlobalListenerTargetAssertionHandler(ctx rule.Context, call *ast.Node) *ast.Node {
	expression := call.AsCallExpression()
	callee := ast.SkipParentheses(expression.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || expression.Arguments == nil || len(expression.Arguments.Nodes) < 2 {
		return nil
	}
	access := callee.AsPropertyAccessExpression()
	if access.Name().Kind != ast.KindIdentifier || access.Name().Text() != "addEventListener" {
		return nil
	}
	receiver := ast.SkipParentheses(access.Expression)
	if receiver.Kind != ast.KindIdentifier || (receiver.Text() != "document" && receiver.Text() != "window") {
		return nil
	}
	if !correctnessNoGlobalListenerTargetAssertionIsGlobalVariable(ctx, ctx.TypeChecker.GetSymbolAtLocation(receiver)) ||
		!correctnessNoGlobalListenerTargetAssertionFromDefaultLibrary(ctx, ctx.TypeChecker.GetSymbolAtLocation(access.Name())) {
		return nil
	}
	handler := ast.SkipParentheses(expression.Arguments.Nodes[1])
	switch handler.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return handler
	case ast.KindIdentifier:
	default:
		return nil
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(handler)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsAlias != 0 {
		return nil
	}
	declarations := rule.DeclarationsIn(ctx.SourceFile, symbol)
	if len(declarations) != 1 || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := declarations[0]
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		return declaration
	case ast.KindVariableDeclaration:
		list := declaration.Parent
		initializer := declaration.AsVariableDeclaration().Initializer
		if list == nil || list.Kind != ast.KindVariableDeclarationList || list.Flags&ast.NodeFlagsConst == 0 || initializer == nil {
			return nil
		}
		initializer = ast.SkipParentheses(initializer)
		if initializer.Kind == ast.KindArrowFunction || initializer.Kind == ast.KindFunctionExpression {
			return initializer
		}
	}
	return nil
}

// correctnessNoGlobalListenerTargetAssertionIsGlobalVariable says whether a symbol is a variable
// declared only in the default library: lib.dom's `declare var document` and `declare var window`.
func correctnessNoGlobalListenerTargetAssertionIsGlobalVariable(ctx rule.Context, symbol *ast.Symbol) bool {
	if !correctnessNoGlobalListenerTargetAssertionFromDefaultLibrary(ctx, symbol) {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindVariableDeclaration {
			return false
		}
	}
	return true
}

// correctnessNoGlobalListenerTargetAssertionFromDefaultLibrary says whether a symbol has declarations
// and every one of them is in a default library file. Every rather than any, so a project
// declaration merged into a library one makes the symbol something this rule has not read.
func correctnessNoGlobalListenerTargetAssertionFromDefaultLibrary(ctx rule.Context, symbol *ast.Symbol) bool {
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || !type_checking.IsSourceFileDefaultLibrary(ctx.Program, file) {
			return false
		}
	}
	return true
}

// correctnessNoGlobalListenerTargetAssertionAt is the assertion to report around one property access,
// when the access is the event's own `target` and an assertion around it names a specific element
// type the checker does not already know the target to be.
func correctnessNoGlobalListenerTargetAssertionAt(ctx rule.Context, access *ast.Node, event *ast.Symbol) *ast.Node {
	property := access.AsPropertyAccessExpression()
	if property.Name().Kind != ast.KindIdentifier || property.Name().Text() != "target" {
		return nil
	}
	object := ast.SkipParentheses(property.Expression)
	if object.Kind != ast.KindIdentifier || object.Text() != event.Name || ctx.TypeChecker.GetSymbolAtLocation(object) != event {
		return nil
	}
	if !correctnessNoGlobalListenerTargetAssertionFromDefaultLibrary(ctx, ctx.TypeChecker.GetSymbolAtLocation(property.Name())) {
		return nil
	}
	known := ctx.TypeChecker.GetTypeAtLocation(access)
	for outer := access.Parent; outer != nil; outer = outer.Parent {
		switch outer.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
			continue
		case ast.KindAsExpression, ast.KindTypeAssertionExpression:
			typeNode := outer.Type()
			if typeNode == nil {
				return nil
			}
			asserted := ctx.TypeChecker.GetTypeFromTypeNode(typeNode)
			if asserted == nil || !correctnessNoGlobalListenerTargetAssertionIsSpecific(ctx, asserted) {
				continue
			}
			if known != nil && ctx.TypeChecker.IsTypeAssignableTo(ctx.TypeChecker.GetNonNullableType(known), ctx.TypeChecker.GetNonNullableType(asserted)) {
				return nil
			}
			return outer
		}
		return nil
	}
	return nil
}

// correctnessNoGlobalListenerTargetAssertionIsSpecific says whether every non-nullish member of a type
// is a specific element type, with at least one such member.
func correctnessNoGlobalListenerTargetAssertionIsSpecific(ctx rule.Context, asserted *checker.Type) bool {
	specific := false
	for part := range type_checking.UnionTypePartsSeq(asserted) {
		if type_checking.IsTypeFlagSet(part, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			continue
		}
		symbol := checker.Type_symbol(part)
		if symbol == nil || symbol.Flags&ast.SymbolFlagsInterface == 0 ||
			correctnessNoGlobalListenerTargetAssertionGeneralTypes[symbol.Name] ||
			!correctnessNoGlobalListenerTargetAssertionFromDefaultLibrary(ctx, symbol) ||
			!correctnessNoGlobalListenerTargetAssertionExtendsElement(ctx, symbol) {
			return false
		}
		specific = true
	}
	return specific
}

// correctnessNoGlobalListenerTargetAssertionExtendsElement says whether a default-library interface
// has the default library's `Element` among its bases, at any depth.
func correctnessNoGlobalListenerTargetAssertionExtendsElement(ctx rule.Context, symbol *ast.Symbol) bool {
	seen := map[*ast.Symbol]bool{symbol: true}
	queue := []*ast.Symbol{symbol}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		declared := checker.Checker_getDeclaredTypeOfSymbol(ctx.TypeChecker, current)
		if declared == nil || checker.Type_objectFlags(declared)&(checker.ObjectFlagsInterface|checker.ObjectFlagsClass) == 0 {
			continue
		}
		for _, base := range checker.Checker_getBaseTypes(ctx.TypeChecker, declared) {
			baseSymbol := checker.Type_symbol(base)
			if baseSymbol == nil || seen[baseSymbol] {
				continue
			}
			if baseSymbol.Name == "Element" && correctnessNoGlobalListenerTargetAssertionFromDefaultLibrary(ctx, baseSymbol) {
				return true
			}
			seen[baseSymbol] = true
			queue = append(queue, baseSymbol)
		}
	}
	return false
}
