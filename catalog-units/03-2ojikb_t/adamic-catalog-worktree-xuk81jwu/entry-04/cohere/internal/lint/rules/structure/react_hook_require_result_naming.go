package structure

import (
	"strings"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// messageHookResultNamingGenericText is the rule's `hookResultNamingGeneric` message, whose wording
// lives in `policy/messages/react-hook-require-result-naming.json`.
var messageHookResultNamingGenericText = policy.MessageOf("structure/react-hook-require-result-naming", "hookResultNamingGeneric")

// messageHookResultNamingGeneric is the finding, rendered when it is reported so the text comes
// from the current catalog.
func messageHookResultNamingGeneric() rule.Message {
	return rule.Message{Id: messageHookResultNamingGenericText.Id, Description: messageHookResultNamingGenericText.Render(nil)}
}

// messageHookResultNamingBadSuffixText is the rule's `hookResultNamingBadSuffix` message, whose
// wording lives in `policy/messages/react-hook-require-result-naming.json`.
var messageHookResultNamingBadSuffixText = policy.MessageOf("structure/react-hook-require-result-naming", "hookResultNamingBadSuffix")

// messageHookResultNamingBadSuffix is the finding, rendered when it is reported so the text comes
// from the current catalog.
func messageHookResultNamingBadSuffix() rule.Message {
	return rule.Message{Id: messageHookResultNamingBadSuffixText.Id, Description: messageHookResultNamingBadSuffixText.Render(nil)}
}

// messageHookResultNamingMismatchText is the rule's `hookResultNamingMismatch` message, whose
// wording lives in `policy/messages/react-hook-require-result-naming.json`.
var messageHookResultNamingMismatchText = policy.MessageOf("structure/react-hook-require-result-naming", "hookResultNamingMismatch")

// messageHookResultNamingMismatch is the finding, rendered when it is reported so the text comes
// from the current catalog.
func messageHookResultNamingMismatch() rule.Message {
	return rule.Message{Id: messageHookResultNamingMismatchText.Id, Description: messageHookResultNamingMismatchText.Render(nil)}
}

// emptySuffixes are the endings that describe a variable rather than its contents.
var emptySuffixes = []string{"Result", "Data", "Value", "State", "Hook"}

// ReactHookRequireResultNaming flags a hook result whose name says nothing about its origin.
//
//	valid:   const accountQuery = useAccountQuery()
//	valid:   const accountQueryError = useAccountQuery()      starts with the expected name
//	invalid: const data = useAccountQuery()
//	invalid: const accountQueryResult = useAccountQuery()
//	invalid: const thing = useAccountQuery()
//
// Only when the hook is called once in the file, which is the judgment that keeps this from being
// noise. Two calls to the same hook have to be told apart by their names, so a prefix or suffix is
// exactly right there and the rule has no business asking for the bare expected name.
//
// The acceptance test is deliberately loose: the expected name, or a name starting or ending with
// it, case-insensitively, plus a singular match for a plural hook. That breadth is the point. The
// rule is trying to catch a name that says nothing, not to enforce one spelling, and a narrow test
// would report `watchEmail` for `useWatch`, which is better than `watch`.
//
// Three message ids for three different complaints about the same name, so the message says which
// one applies rather than making a reader work out why their name was rejected.
var ReactHookRequireResultNaming = rule.Rule{
	Name:       "structure/react-hook-require-result-naming",
	NoListener: rule.NoListenerAnswersInRun,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if !FileContextFor(ctx.SourceFile.FileName().AsString()).IsReactFile {
			return nil
		}

		// Two passes over the file, because the judgment on any one result needs a fact about the
		// whole file: how many times its hook was called. A single walk cannot answer that at the
		// declaration, since a later call would change the verdict.
		invocationCounts := map[string]int{}
		var results []hookResultBinding

		sourceFile := ctx.SourceFile.AsNode()
		countHookInvocations(sourceFile, invocationCounts, &results)

		for _, result := range results {
			if invocationCounts[result.hookName] != 1 {
				continue
			}
			if isRelatedHookResultName(result.variableName, result.hookName) {
				continue
			}
			ctx.ReportNode(result.node, hookResultNamingMessage(result.variableName))
		}

		return nil
	},
}

// hookResultBinding is one `const x = useThing()` binding.
type hookResultBinding struct {
	node         *ast.Node
	variableName string
	hookName     string
}

// countHookInvocations walks the file once, counting hook calls and collecting result bindings.
//
// Both facts come from the same walk because they are found at the same nodes and a second traversal
// would double the cost for nothing.
func countHookInvocations(node *ast.Node, counts map[string]int, results *[]hookResultBinding) {
	if node == nil {
		return
	}

	// The `use` prefix on the count is redundant with the one on the binding: a name is only ever
	// looked up in this map after being read off a call that already passed the same test, so a
	// non-hook call counted here is counted under a key nobody reads. Kept because the map is named
	// for hook invocations and filling it with every call in the file would make that a lie.
	if node.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee != nil && callee.Kind == ast.KindIdentifier && strings.HasPrefix(callee.Text(), "use") {
			counts[callee.Text()]++
		}
	}

	if node.Kind == ast.KindVariableDeclaration {
		declaration := node.AsVariableDeclaration()
		name := declaration.Name()

		// The nil check comes before SkipParentheses, which dereferences its argument. A
		// declaration without an initializer is ordinary (`let x;`, `declare const y: T`), and
		// putting the check after the call panics on every one of them.
		//
		// I wrote this exact bug three rules ago, documented it, added a standing guard for it, and
		// then wrote it again here. It reached the real tree: 42 files crashed before the fix.
		var initializer *ast.Node
		if declaration.Initializer != nil {
			initializer = ast.SkipParentheses(declaration.Initializer)
		}

		if name != nil && name.Kind == ast.KindIdentifier &&
			initializer != nil && initializer.Kind == ast.KindCallExpression {
			callee := ast.SkipParentheses(initializer.AsCallExpression().Expression)
			if callee != nil && callee.Kind == ast.KindIdentifier && strings.HasPrefix(callee.Text(), "use") {
				*results = append(*results, hookResultBinding{
					node:         name,
					variableName: name.Text(),
					hookName:     callee.Text(),
				})
			}
		}
	}

	node.ForEachChild(func(child *ast.Node) bool {
		countHookInvocations(child, counts, results)
		return false
	})
}

// expectedHookResultName derives the name a hook's result would naturally take.
//
// `useAccountQuery` gives `accountQuery`. A name that is exactly `use` has nothing after the prefix
// and is returned unchanged, which the caller then compares against, so such a hook effectively
// accepts any name.
//
// The lowercasing cannot be measured by a fixture, and that is a property of the caller rather than
// a gap here: every comparison in isRelatedHookResultName is case-insensitive, so the first letter's
// case never reaches a decision. Kept because the value also appears in nothing else, and a reader
// seeing `AccountQuery` returned from a function named "expected result name" would reasonably
// think it a bug.
func expectedHookResultName(hookName string) string {
	if !strings.HasPrefix(hookName, "use") || len(hookName) <= 3 {
		return hookName
	}
	rest := []rune(hookName[3:])
	rest[0] = unicode.ToLower(rest[0])
	return string(rest)
}

// isRelatedHookResultName reports whether a name says where it came from.
//
// Deliberately loose. `watchEmail` and `emailWatch` both relate to `useWatch`, and both are better
// names than `watch`, so a test demanding the exact expected name would report improvements.
func isRelatedHookResultName(variableName string, hookName string) bool {
	expected := expectedHookResultName(hookName)
	if variableName == expected {
		return true
	}

	lowerVariable := strings.ToLower(variableName)
	lowerExpected := strings.ToLower(expected)

	if strings.HasPrefix(lowerVariable, lowerExpected) || strings.HasSuffix(lowerVariable, lowerExpected) {
		return true
	}

	// A plural hook accepts a name built on its singular: `lineSprings` relates to `useSprings`.
	if strings.HasSuffix(lowerExpected, "s") &&
		strings.Contains(lowerVariable, strings.TrimSuffix(lowerExpected, "s")) {
		return true
	}

	return false
}

// hookResultNamingMessage picks which complaint applies.
//
// Ordered rather than combined: a name can be both generic and badly suffixed, and saying one thing
// about it is more useful than saying two. The suffix check runs first because it names a concrete
// edit, where the generic message only says the name is thin.
func hookResultNamingMessage(variableName string) rule.Message {
	for _, suffix := range emptySuffixes {
		if strings.HasSuffix(variableName, suffix) {
			return messageHookResultNamingBadSuffix()
		}
	}

	lower := strings.ToLower(variableName)
	if lower == "data" || lower == "result" || len(variableName) <= 2 {
		return messageHookResultNamingGeneric()
	}

	return messageHookResultNamingMismatch()
}
