package nexus

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const consistencyNoIsoStringDateCutId = "isoStringCutToDate"

// consistencyNoIsoStringDateCutText is the rule's message, whose wording lives in
// `policy/messages/consistency-no-iso-string-date-cut.json`.
var consistencyNoIsoStringDateCutText = policy.MessageOf("nexus/consistency-no-iso-string-date-cut", consistencyNoIsoStringDateCutId)

func consistencyNoIsoStringDateCutMessage() rule.Message {
	return rule.Message{Id: consistencyNoIsoStringDateCutId, Description: consistencyNoIsoStringDateCutText.Render(nil)}
}

// ConsistencyNoIsoStringDateCut reports a `toISOString()` string cut down to its date part, in favour of
// `dateIso8601(date, timeZone)`, so every calendar day in the code names the zone it is in.
//
//	invalid: const today = (): string => new Date().toISOString().slice(0, 10);
//	invalid: return date.toISOString().substring(0, 10);
//	invalid: const date = message.createdAt.toISOString().split('T')[0];
//	invalid: const [mondayDate = ''] = monday.toISOString().split('T');
//	invalid: const iso = new Date().toISOString(); const datePart = iso.slice(0, 10);
//	valid:   const today = (): string => dateIso8601(new Date(), userTimeZone());
//	valid:   return dateIso8601(date, 'UTC');
//	valid:   const isExactUtcMidnight = isoUtc.slice(11, 23) === '00:00:00.000';
//	valid:   return new Date(epochInMilliseconds).toISOString().slice(0, 16).replace('T', ' ');
//
// # Where it came from
//
// The new-rules sweep (`#tevhg3f`) counted 77 of these in ahra, built as task `#qd05ghr`. The shape is
// a day with its zone left implicit, and the implicit zone is UTC. Two kinds of real bug sit behind it
// in ahra:
//
//   - **Local arithmetic read as a UTC day.** `app/api/finance/route.ts:48`,
//     `modules/finance/FinanceApi.ts:474`, `modules/finance/FinanceAnalysisWindow.ts:45` and
//     `modules/finance/FinanceTransactionsCommandLineInterface.ts:270` step a Date back with
//     `setDate(getDate() - days)`, in local days, then slice the UTC day, so in a Utah evening every
//     one of those windows starts a day late.
//   - **Today as a UTC day.** The finance `today()` helper in
//     `modules/finance/FinanceCommandLineInterfaceShared.ts:56`, documented as the default as-of for
//     every dated write verb, stamps an evening write with tomorrow's date.
//
// Many other cuts mean the UTC day on purpose (`shiftDateByDays` anchors at `T00:00:00Z` and steps with
// `setUTCDate`), and nothing in the code can tell those from the bugs. That is why this is a ban on the
// shape rather than a guess at intent: the cut is the violation wherever it is, and the fix at a UTC
// site is the same call with `'UTC'`, which says on the line what the cut left unsaid.
//
// # What counts as the date part
//
// The string a `toISOString()` call returns, the call's symbol resolved to the method the default
// library declares on `Date`, cut by one of:
//
//   - `.slice(a, b)`, `.substring(a, b)` (in either order) or `.substr(a, length)` with non-negative
//     integer literals whose range is non-empty and ends at or before index 10, so the result lies
//     inside `YYYY-MM-DD`. `.slice(0, 16)` (date and time) and `.slice(11, 23)` (time only) are not the
//     date part; neither is a negative index, which counts from an end that moves with the year.
//   - `.split('T')[0]`, the split taking exactly that one argument.
//   - `const [day] = <iso>.split('T')`, the first element bound and not a rest.
//
// The ISO string may be held in a `const` first (`const iso = date.toISOString(); iso.slice(0, 10)`),
// declared in the same file: a `const` cannot change between the two, and another file's binding is a
// walk this rule does not make. A `let` is not followed.
//
// A `toISOString` that is not the default library's (a local class, moment, dayjs) is not reported:
// what its string holds is that library's business. A string cut after anything else touched it
// (`.replace('T', ' ').slice(0, 19)`) is not reported either; those are timestamps with their time.
//
// # The definition is a use
//
// `dateIso8601` in Structure's `FormatTime.ts` cuts the ISO string itself on its no-zone branch, and
// this rule reports it like any other site, with no path or name in the rule. The fix there is to give
// that branch the zone too (`'UTC'` through the same `Intl.DateTimeFormat` the zoned branch uses).
//
// # No fix
//
// The zone is the author's call, and choosing `'UTC'` or `userTimeZone()` for them is exactly the guess
// the rule exists to stop.
var ConsistencyNoIsoStringDateCut = rule.Rule{
	Name: "nexus/consistency-no-iso-string-date-cut",

	// Symbol identity is what tells the default library's `Date.prototype.toISOString` from a local
	// class or a date library wearing the same name, and one `const` from a shadowing one.
	NeedsTypeChecker: true,

	// Compiler options and the default library, through type_checking's default-library test.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	// The rule follows a `const` only in this file, and of an imported member reads only the interface
	// it is declared on, which no function body holds, so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil || ctx.Program == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				consistencyNoIsoStringDateCutCheckRange(ctx, node)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				consistencyNoIsoStringDateCutCheckSplitIndex(ctx, node)
			},
			ast.KindVariableDeclaration: func(node *ast.Node) {
				consistencyNoIsoStringDateCutCheckDestructuredSplit(ctx, node)
			},
		}
	},
}

// consistencyNoIsoStringDateCutDatePartEnd is the length of `YYYY-MM-DD`, where an ISO string's date
// part ends for every year from 0000 to 9999.
const consistencyNoIsoStringDateCutDatePartEnd = 10

// consistencyNoIsoStringDateCutCheckRange reports `<iso>.slice|substring|substr(a, b)` whose range lies
// inside the date part.
func consistencyNoIsoStringDateCutCheckRange(ctx rule.Context, call *ast.Node) {
	method, receiver, arguments := consistencyNoIsoStringDateCutMethodCall(call)
	if receiver == nil || len(arguments) != 2 {
		return
	}
	first, firstOk := consistencyNoIsoStringDateCutIndex(arguments[0])
	second, secondOk := consistencyNoIsoStringDateCutIndex(arguments[1])
	if !firstOk || !secondOk {
		return
	}
	var start, end int
	switch method {
	case "slice":
		start, end = first, second
	case "substring":
		start, end = min(first, second), max(first, second)
	case "substr":
		start, end = first, first+second
	default:
		return
	}
	if start >= end || end > consistencyNoIsoStringDateCutDatePartEnd {
		return
	}
	if consistencyNoIsoStringDateCutIsIsoString(ctx, receiver) {
		ctx.ReportNode(call, consistencyNoIsoStringDateCutMessage())
	}
}

// consistencyNoIsoStringDateCutCheckSplitIndex reports `<iso>.split('T')[0]`.
func consistencyNoIsoStringDateCutCheckSplitIndex(ctx rule.Context, node *ast.Node) {
	access := node.AsElementAccessExpression()
	if index, ok := consistencyNoIsoStringDateCutIndex(access.ArgumentExpression); !ok || index != 0 {
		return
	}
	if consistencyNoIsoStringDateCutIsSplitOnT(ctx, ast.SkipParentheses(access.Expression)) {
		ctx.ReportNode(node, consistencyNoIsoStringDateCutMessage())
	}
}

// consistencyNoIsoStringDateCutCheckDestructuredSplit reports `const [day] = <iso>.split('T')`, at the
// split.
func consistencyNoIsoStringDateCutCheckDestructuredSplit(ctx rule.Context, node *ast.Node) {
	name := node.Name()
	initializer := node.AsVariableDeclaration().Initializer
	if name == nil || name.Kind != ast.KindArrayBindingPattern || initializer == nil {
		return
	}
	elements := name.Elements()
	// A hole (`[, time]`) parses as a binding element with no name, and binds nothing.
	if len(elements) == 0 || elements[0].Kind != ast.KindBindingElement || elements[0].Name() == nil ||
		elements[0].AsBindingElement().DotDotDotToken != nil {
		return
	}
	split := ast.SkipParentheses(initializer)
	if consistencyNoIsoStringDateCutIsSplitOnT(ctx, split) {
		ctx.ReportNode(split, consistencyNoIsoStringDateCutMessage())
	}
}

// consistencyNoIsoStringDateCutMethodCall splits `receiver.method(arguments)` into its parts, the
// receiver with parentheses skipped, or returns a nil receiver for any other node.
func consistencyNoIsoStringDateCutMethodCall(call *ast.Node) (string, *ast.Node, []*ast.Node) {
	if call == nil || call.Kind != ast.KindCallExpression {
		return "", nil, nil
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return "", nil, nil
	}
	access := callee.AsPropertyAccessExpression()
	if access.Name().Kind != ast.KindIdentifier {
		return "", nil, nil
	}
	return access.Name().Text(), ast.SkipParentheses(access.Expression), call.Arguments()
}

// consistencyNoIsoStringDateCutIsSplitOnT says whether a node is `<iso>.split('T')`, with exactly that
// one argument. A limit changes what index 0 can be (`split('T', 0)` is empty), so it is left alone.
func consistencyNoIsoStringDateCutIsSplitOnT(ctx rule.Context, node *ast.Node) bool {
	method, receiver, arguments := consistencyNoIsoStringDateCutMethodCall(node)
	if method != "split" || receiver == nil || len(arguments) != 1 {
		return false
	}
	separator := arguments[0]
	if (separator.Kind != ast.KindStringLiteral && separator.Kind != ast.KindNoSubstitutionTemplateLiteral) || separator.Text() != "T" {
		return false
	}
	return consistencyNoIsoStringDateCutIsIsoString(ctx, receiver)
}

// consistencyNoIsoStringDateCutIndex reads a non-negative integer literal.
func consistencyNoIsoStringDateCutIndex(node *ast.Node) (int, bool) {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindNumericLiteral {
		return 0, false
	}
	value, err := strconv.Atoi(node.Text())
	if err != nil || value < 0 {
		return 0, false
	}
	return value, true
}

// consistencyNoIsoStringDateCutIsIsoString says whether an expression is the default library's
// `<date>.toISOString()`, called directly or held in a `const` of this file.
func consistencyNoIsoStringDateCutIsIsoString(ctx rule.Context, expression *ast.Node) bool {
	if expression.Kind == ast.KindIdentifier {
		initializer := consistencyNoIsoStringDateCutConstInitializer(ctx, expression)
		if initializer == nil {
			return false
		}
		expression = ast.SkipParentheses(initializer)
	}
	method, receiver, arguments := consistencyNoIsoStringDateCutMethodCall(expression)
	if method != "toISOString" || receiver == nil || len(arguments) != 0 {
		return false
	}
	name := ast.SkipParentheses(expression.AsCallExpression().Expression).AsPropertyAccessExpression().Name()
	return consistencyNoIsoStringDateCutIsDateMember(ctx, name)
}

// consistencyNoIsoStringDateCutConstInitializer is the initializer of the `const` an identifier names,
// when that `const` is the symbol's only declaration, binds a plain identifier, and is in this file.
func consistencyNoIsoStringDateCutConstInitializer(ctx rule.Context, identifier *ast.Node) *ast.Node {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if ast.GetSourceFileOfNode(declaration) != ctx.SourceFile || declaration.Kind != ast.KindVariableDeclaration ||
		declaration.Name() == nil || declaration.Name().Kind != ast.KindIdentifier {
		return nil
	}
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList || list.Flags&ast.NodeFlagsBlockScoped != ast.NodeFlagsConst {
		return nil
	}
	return declaration.AsVariableDeclaration().Initializer
}

// consistencyNoIsoStringDateCutIsDateMember says whether a member name resolves to a method the
// default library declares on `Date`.
func consistencyNoIsoStringDateCutIsDateMember(ctx rule.Context, name *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(name)
	if symbol == nil || !type_checking.IsSymbolFromDefaultLibrary(ctx.Program, symbol) {
		return false
	}
	for _, declaration := range symbol.Declarations {
		parent := declaration.Parent
		if parent != nil && parent.Kind == ast.KindInterfaceDeclaration && parent.Name() != nil && parent.Name().Text() == "Date" {
			return true
		}
	}
	return false
}
