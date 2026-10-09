package nexus

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoDiscardedPureResultId = "discardedPureResult"

// correctnessNoDiscardedPureResultText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-discarded-pure-result.json`.
var correctnessNoDiscardedPureResultText = policy.MessageOf("nexus/correctness-no-discarded-pure-result", correctnessNoDiscardedPureResultId)

func correctnessNoDiscardedPureResultMessage() rule.Message {
	return rule.Message{
		Id:          correctnessNoDiscardedPureResultId,
		Description: correctnessNoDiscardedPureResultText.Render(nil),
	}
}

// CorrectnessNoDiscardedPureResult reports a statement that calls a side-effect-free method of the
// default library's String or Array and drops what it returns.
//
//	invalid: name.trim();
//	invalid: path.replace('/old/', '/new/');
//	invalid: list.concat(more);
//	invalid: list.slice(1);
//	valid:   name = name.trim();
//	valid:   const trimmed = name.trim();
//	valid:   void name.trim();
//	valid:   text.replace(pattern, function(match) { matches.push(match); return match; });
//	valid:   list.push(...more);
//
// # Where it came from
//
// A guard, not a cleanup: zero sites in ahra when it was built. It is the classic refactor and
// authoring slip of treating an immutable value as if the call changed it in place, and it reads as
// correct at a glance because the mutating twins (`push`, `sort`, `splice`) are written the same way.
// Found by the cross-language and JavaScript-catalog passes of the new-rules sweep (`#tevhg3f`, after
// staticcheck SA4017, go vet `unusedresult`, sonarjs `no-ignored-return` and unicorn
// `no-unused-builtin-method-return`), built as part of `#j03vwm6`.
//
// # The shape, exactly
//
//  1. An expression statement whose expression (parentheses looked through) is a call of a member
//     access, optional or not: `receiver.method(...)` or `receiver?.method(...)`.
//  2. The member's symbol has every declaration in the default library, on the `String`, `Array` or
//     `ReadonlyArray` interface, and the method is in that interface's set below. A union receiver
//     (`string | string[]`) qualifies only when every member it merges qualifies. A class or
//     interface of the project's own with a `trim` method is never read.
//  3. No argument is a function, by type: `text.replace(pattern, (match) => { ... })` is the idiom
//     that walks the matches for the callback's side effects, and `toSorted(compare)` runs code. A
//     callback array method (`map`, `filter`, `forEach`) is not in any set: an ignored `map` with a
//     side-effecting callback is a style choice, not a lost value.
//
// The sets are the methods whose only effect is their result:
//
//   - **String**: `at`, `charAt`, `charCodeAt`, `codePointAt`, `concat`, `endsWith`, `includes`,
//     `indexOf`, `isWellFormed`, `lastIndexOf`, `match`, `matchAll`, `padEnd`, `padStart`, `replace`,
//     `replaceAll`, `search`, `slice`, `split`, `startsWith`, `substr`, `substring`,
//     `toLocaleLowerCase`, `toLocaleUpperCase`, `toLowerCase`, `toString`, `toUpperCase`,
//     `toWellFormed`, `trim`, `trimEnd`, `trimLeft`, `trimRight`, `trimStart`, `valueOf`.
//   - **Array and ReadonlyArray**: `at`, `concat`, `entries`, `flat`, `includes`, `indexOf`, `join`,
//     `keys`, `lastIndexOf`, `slice`, `toReversed`, `toSorted`, `toSpliced`, `values`, `with`.
//
// Left out on purpose: `normalize`, `repeat`, `localeCompare` and `toLocaleString`, which can throw
// on a bad argument and so could, however unlikely, be written as a validation; and every mutating
// method. `match` and `replace` with a `g` regex reset its `lastIndex` to zero, which nobody calls
// them for.
//
// # No fix
//
// Whether the result should be assigned back, kept in a new binding, or the statement deleted depends
// on what the author meant, and assigning back is not even possible for a `const`.
var CorrectnessNoDiscardedPureResult = rule.Rule{
	Name: "nexus/correctness-no-discarded-pure-result",

	// The method is identified by its declaration in the default library, and an argument's type
	// decides whether it is a callback.
	NeedsTypeChecker: true,

	// Whether a declaration is in the default library, through type_checking.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	// The rule reads types and the default library's declarations, never an imported body.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				correctnessNoDiscardedPureResultCheck(ctx, node)
			},
		}
	},
}

var correctnessNoDiscardedPureResultStringMethods = map[string]bool{
	"at": true, "charAt": true, "charCodeAt": true, "codePointAt": true, "concat": true, "endsWith": true,
	"includes": true, "indexOf": true, "isWellFormed": true, "lastIndexOf": true, "match": true,
	"matchAll": true, "padEnd": true, "padStart": true, "replace": true, "replaceAll": true, "search": true,
	"slice": true, "split": true, "startsWith": true, "substr": true, "substring": true,
	"toLocaleLowerCase": true, "toLocaleUpperCase": true, "toLowerCase": true, "toString": true,
	"toUpperCase": true, "toWellFormed": true, "trim": true, "trimEnd": true, "trimLeft": true,
	"trimRight": true, "trimStart": true, "valueOf": true,
}

var correctnessNoDiscardedPureResultArrayMethods = map[string]bool{
	"at": true, "concat": true, "entries": true, "flat": true, "includes": true, "indexOf": true, "join": true,
	"keys": true, "lastIndexOf": true, "slice": true, "toReversed": true, "toSorted": true, "toSpliced": true,
	"values": true, "with": true,
}

// correctnessNoDiscardedPureResultMethods are the side-effect-free methods by the default library
// interface that declares them.
var correctnessNoDiscardedPureResultMethods = map[string]map[string]bool{
	"String":        correctnessNoDiscardedPureResultStringMethods,
	"Array":         correctnessNoDiscardedPureResultArrayMethods,
	"ReadonlyArray": correctnessNoDiscardedPureResultArrayMethods,
}

func correctnessNoDiscardedPureResultCheck(ctx rule.Context, statement *ast.Node) {
	call := ast.SkipParentheses(statement.AsExpressionStatement().Expression)
	if call.Kind != ast.KindCallExpression {
		return
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return
	}
	name := callee.AsPropertyAccessExpression().Name()
	if name.Kind != ast.KindIdentifier || !correctnessNoDiscardedPureResultIsPure(ctx, name) {
		return
	}
	if arguments := call.AsCallExpression().Arguments; arguments != nil {
		for _, argument := range arguments.Nodes {
			if correctnessNoDiscardedPureResultIsCallable(ctx, argument) {
				return
			}
		}
	}
	ctx.ReportNode(call, correctnessNoDiscardedPureResultMessage())
}

// correctnessNoDiscardedPureResultIsPure says whether a method name resolves, through every
// declaration, to a side-effect-free method of the default library's String, Array or ReadonlyArray.
func correctnessNoDiscardedPureResultIsPure(ctx rule.Context, name *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		owner := declaration.Parent
		if declaration.Kind != ast.KindMethodSignature || owner == nil || owner.Kind != ast.KindInterfaceDeclaration ||
			owner.Name() == nil {
			return false
		}
		if !correctnessNoDiscardedPureResultMethods[owner.Name().Text()][name.Text()] {
			return false
		}
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || !type_checking.IsSourceFileDefaultLibrary(ctx.Program, file) {
			return false
		}
	}
	return true
}

// correctnessNoDiscardedPureResultIsCallable says whether an argument is a function: written as one,
// or of a type with a call signature in any part of it. A spread argument is read by its own type.
func correctnessNoDiscardedPureResultIsCallable(ctx rule.Context, argument *ast.Node) bool {
	value := ast.SkipParentheses(argument)
	if value.Kind == ast.KindArrowFunction || value.Kind == ast.KindFunctionExpression {
		return true
	}
	argumentType := ctx.TypeChecker.GetTypeAtLocation(value)
	if argumentType == nil {
		return true
	}
	for part := range type_checking.UnionTypePartsSeq(argumentType) {
		if type_checking.IsTypeFlagSet(part, checker.TypeFlagsAny|checker.TypeFlagsUnknown) {
			return true
		}
		if len(ctx.TypeChecker.GetSignaturesOfType(part, checker.SignatureKindCall)) > 0 {
			return true
		}
	}
	return false
}
