package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const correctnessNoDiscardedOutcomeId = "outcomeDiscarded"

// correctnessNoDiscardedOutcomeText is the rule's message, whose wording lives in
// `policy/messages/correctness-no-discarded-outcome.json`.
var correctnessNoDiscardedOutcomeText = policy.MessageOf("nexus/correctness-no-discarded-outcome", correctnessNoDiscardedOutcomeId)

func correctnessNoDiscardedOutcomeMessage(typeName string) rule.Message {
	return rule.Message{
		Id:          correctnessNoDiscardedOutcomeId,
		Description: correctnessNoDiscardedOutcomeText.Render(map[string]string{"typeName": typeName}),
	}
}

// CorrectnessNoDiscardedOutcome reports a call whose result is one of Nexus's `{ outcome }` types,
// made as a statement on its own so the outcome is thrown away.
//
//	invalid: writeJsonFile(statePath, state, 'meta polling state');
//	invalid: await saveLedger(ledger); // async, resolving to JsonFileWriteOutcomeType
//	invalid: store?.save(); // JsonFileWriteOutcomeType | undefined
//	valid:   const stateWrite = writeJsonFile(statePath, state, 'meta polling state');
//	         if(stateWrite.outcome === 'Unwritable') console.warn(stateWrite.message);
//	valid:   return writeJsonFile(statePath, state, 'meta polling state');
//	valid:   void writeJsonFile(cachePath, cache, 'scratch cache');
//
// An outcome nobody reads is the boolean it replaced, minus the boolean. `consistency-no-boolean-outcome`
// moves results onto named outcomes so a failure has a name; this rule is the other half, rustc's
// `unused_must_use` and go vet's `unusedresult` for those types: the name only helps if someone reads it.
//
// # Where it came from
//
// Four state writes in ahra, found by the cross-language pass of the new-rules sweep (`#tevhg3f`)
// and built as task `#1vf0qwd`. Each saves a ledger with `writeJsonFile(...)` as a bare statement,
// in a function returning `void`:
//
//   - `modules/meta/MetaPollingApi.ts:41` (`saveState`): the poller's last-seen cursor. A write that
//     fails leaves the old cursor on disk, and the next poll replays every item as new.
//   - `modules/google/youtube/YouTubeSafety.ts:57` (`saveLedger`): the ledger that stops a repeated
//     action. A failed write forgets the action, and the safety limit it backs stops counting.
//   - `modules/pensieve/PensieveWindows.ts:43` (`saveJournalState`) and
//     `modules/pensieve/PensieveWeeklies.ts:99` (`recordPartialSource`): journal and partial state,
//     where a lost write re-journals windows already done.
//
// # Which types are outcomes
//
// Decided by declaration, never by a name or a field: the five type aliases Nexus declares as a union
// of `{ outcome: ... }` arms, each matched by its name and the file that declares it, at the top level
// of that file. The list is every such alias in Nexus on 2026-10-03, read from the source:
//
//   - `JsonParseOutcomeType` in `nexus/source/structured-text/json/Json.ts` (`parseJson`)
//   - `JsoncParseOutcomeType` in `nexus/source/structured-text/json/Jsonc.ts` (`parseJsonc`)
//   - `JsonFileReadOutcomeType` in `nexus/source/structured-text/json/JsonFile.ts` (`readJsonFile`)
//   - `JsonFileWriteOutcomeType` in `nexus/source/structured-text/json/JsonFile.ts` (`writeJsonFile`)
//   - `PackageJsonReadOutcomeType` in `nexus/source/system/PackageJson.ts` (`readPackageJson`)
//
// A new outcome type in Nexus is added to this list by hand, after reading it. A project's own
// outcome types are not here: an alias of the same name in another file, or a union of the same
// shape, is someone else's type, and this rule reports nothing for it.
//
// The test is on the arms, not on the alias, so a value still counts when the alias did not survive
// into its type: an optional call's `JsonFileWriteOutcomeType | undefined`, or a project union adding
// an arm of its own beside Nexus's. Each arm of a Nexus alias is a type literal, and an instantiated
// arm (`JsonFileReadOutcomeType<LedgerInterface>`) keeps the symbol of the literal it came from, so
// the arm leads back to its alias by declaration. A value counts when it can be every arm of one of
// the five, failure arms included. A value narrowed to some arms (`Extract<JsonFileWriteOutcomeType,
// { outcome: 'Written' }>`) is not reported: whatever it holds, nothing is left to handle.
//
// # What counts as discarded
//
// An expression statement whose expression, through parentheses, is a call, or `await` of a call,
// and whose value (awaited, for the second) is an outcome. Function calls, method calls and optional
// calls alike, so a project wrapper that returns `writeJsonFile(...)` is reported where its result
// is dropped, whether its return type is written or inferred.
//
// Declined, each a missed finding rather than a false one:
//
//   - **An unawaited call to an async producer.** `saveAsync();` drops a Promise, and the floating
//     promise rules own that.
//   - **Short-circuit and ternary statements.** `ready && writeJsonFile(...)` is already reported by
//     `no-unused-expressions`.
//   - **A callback whose result is dropped by its caller.** `items.forEach(function(item) { return
//     writeJsonFile(...); })` hands the outcome to a callee that ignores it; `strict-void-return`
//     reports a value returned where `void` is expected.
//   - **A binding never read.** `const outcome = writeJsonFile(...)` and nothing after it is an unused
//     variable, which the unused-variable rules report.
//   - **An outcome inside something else.** `await Promise.all([writeA(), writeB()])` drops an array of
//     outcomes. Following them through containers is a different analysis.
//
// # The opt-out
//
// `void writeJsonFile(...)`: a statement that is a `void` expression, which says in the code that the
// outcome is ignored on purpose. Nothing else silences it, short of handling the outcome.
//
// # No fix
//
// What a failure should do is the author's call: warn, exit, retry, or return the outcome upward.
var CorrectnessNoDiscardedOutcome = rule.Rule{
	Name: "nexus/correctness-no-discarded-outcome",

	// The type of the call is what makes it an outcome, and the declaration behind each arm is what
	// says whose outcome it is.
	NeedsTypeChecker: true,

	// The rule reads the call's type and the top-level type alias each arm is declared in, never a
	// body, so its findings key on imports' shapes.
	TypeReach: rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			return nil
		}
		return rule.Listeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				expression := ast.SkipParentheses(node.AsExpressionStatement().Expression)
				awaited := false
				if expression.Kind == ast.KindAwaitExpression {
					expression = ast.SkipParentheses(expression.AsAwaitExpression().Expression)
					awaited = true
				}
				if expression.Kind != ast.KindCallExpression {
					return
				}
				valueType := ctx.TypeChecker.GetTypeAtLocation(expression)
				if valueType != nil && awaited {
					valueType = checker.Checker_getAwaitedType(ctx.TypeChecker, valueType)
				}
				if typeName := correctnessNoDiscardedOutcomeTypeName(valueType); typeName != "" {
					ctx.ReportNode(node, correctnessNoDiscardedOutcomeMessage(typeName))
				}
			},
		}
	},
}

// correctnessNoDiscardedOutcomeDeclarations are Nexus's outcome types, by the file that declares
// each and the alias's name. The file is matched by its path from the Nexus checkout's own folder,
// so the list holds wherever a project keeps Nexus.
var correctnessNoDiscardedOutcomeDeclarations = []struct {
	fileSuffix string
	typeName   string
}{
	{"/nexus/source/structured-text/json/Json.ts", "JsonParseOutcomeType"},
	{"/nexus/source/structured-text/json/Jsonc.ts", "JsoncParseOutcomeType"},
	{"/nexus/source/structured-text/json/JsonFile.ts", "JsonFileReadOutcomeType"},
	{"/nexus/source/structured-text/json/JsonFile.ts", "JsonFileWriteOutcomeType"},
	{"/nexus/source/system/PackageJson.ts", "PackageJsonReadOutcomeType"},
}

// correctnessNoDiscardedOutcomeTypeName returns the name of the Nexus outcome type a value can be
// every arm of, or empty when there is none.
func correctnessNoDiscardedOutcomeTypeName(valueType *checker.Type) string {
	if valueType == nil {
		return ""
	}
	constituents := []*checker.Type{valueType}
	if checker.Type_flags(valueType)&checker.TypeFlagsUnion != 0 {
		constituents = valueType.Types()
	}

	// The arms seen, grouped by the union they were declared in.
	armsByUnion := map[*ast.Node]map[*ast.Node]bool{}
	for _, constituent := range constituents {
		symbol := checker.Type_symbol(constituent)
		if symbol == nil {
			continue
		}
		for _, declaration := range symbol.Declarations {
			union := correctnessNoDiscardedOutcomeArmUnion(declaration)
			if union == nil {
				continue
			}
			if armsByUnion[union] == nil {
				armsByUnion[union] = map[*ast.Node]bool{}
			}
			armsByUnion[union][declaration] = true
		}
	}

	// A value that is every arm of two of them names the first in the list, so the message is the same
	// on every run.
	complete := map[string]bool{}
	for union, arms := range armsByUnion {
		if len(arms) == len(union.AsUnionTypeNode().Types.Nodes) {
			complete[correctnessNoDiscardedOutcomeAliasName(union)] = true
		}
	}
	for _, known := range correctnessNoDiscardedOutcomeDeclarations {
		if complete[known.typeName] {
			return known.typeName
		}
	}
	return ""
}

// correctnessNoDiscardedOutcomeArmUnion returns the union node a declaration is an arm of, when the
// declaration is a type literal written directly in the union a listed Nexus alias is declared as.
func correctnessNoDiscardedOutcomeArmUnion(declaration *ast.Node) *ast.Node {
	if declaration == nil || declaration.Kind != ast.KindTypeLiteral {
		return nil
	}
	union := correctnessNoDiscardedOutcomeSkipParenthesizedTypeUp(declaration.Parent)
	if union == nil || union.Kind != ast.KindUnionType {
		return nil
	}
	if correctnessNoDiscardedOutcomeAliasName(union) == "" {
		return nil
	}
	return union
}

// correctnessNoDiscardedOutcomeAliasName returns the name of the listed Nexus alias declared as this
// union, at the top level of its file, or empty.
func correctnessNoDiscardedOutcomeAliasName(union *ast.Node) string {
	alias := correctnessNoDiscardedOutcomeSkipParenthesizedTypeUp(union.Parent)
	if alias == nil || alias.Kind != ast.KindTypeAliasDeclaration {
		return ""
	}
	declaration := alias.AsTypeAliasDeclaration()
	if declaration.Name() == nil || declaration.Type == nil || correctnessNoDiscardedOutcomeSkipParenthesizedType(declaration.Type) != union {
		return ""
	}
	if alias.Parent == nil || alias.Parent.Kind != ast.KindSourceFile {
		return ""
	}
	file := ast.GetSourceFileOfNode(alias)
	if file == nil {
		return ""
	}
	name := declaration.Name().Text()
	for _, known := range correctnessNoDiscardedOutcomeDeclarations {
		if name == known.typeName && strings.HasSuffix(file.FileName().AsString(), known.fileSuffix) {
			return name
		}
	}
	return ""
}

// correctnessNoDiscardedOutcomeSkipParenthesizedTypeUp walks up out of parenthesized type nodes.
func correctnessNoDiscardedOutcomeSkipParenthesizedTypeUp(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedType {
		node = node.Parent
	}
	return node
}

// correctnessNoDiscardedOutcomeSkipParenthesizedType walks down into parenthesized type nodes.
func correctnessNoDiscardedOutcomeSkipParenthesizedType(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedType {
		node = node.AsParenthesizedTypeNode().Type
	}
	return node
}
