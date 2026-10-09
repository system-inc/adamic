package structure

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// The rule's messages, one handle per id, whose wording lives in
// `policy/messages/network-require-hook-options-parameter.json`.
var (
	networkRequireHookOptionsParameterMissingOptionsParameterText          = policy.MessageOf("structure/network-require-hook-options-parameter", "missingOptionsParameter")
	networkRequireHookOptionsParameterOptionsParameterNotPassedThroughText = policy.MessageOf("structure/network-require-hook-options-parameter", "optionsParameterNotPassedThrough")
	networkRequireHookOptionsParameterIncorrectOptionsParameterTypeText    = policy.MessageOf("structure/network-require-hook-options-parameter", "incorrectOptionsParameterType")
)

// messageMissingOptionsParameter is the finding, rendered when it is reported so the text comes from the current catalog.
func messageMissingOptionsParameter() rule.Message {
	return rule.Message{
		Id:          networkRequireHookOptionsParameterMissingOptionsParameterText.Id,
		Description: networkRequireHookOptionsParameterMissingOptionsParameterText.Render(nil),
	}
}

// messageOptionsParameterNotPassedThrough is the finding, rendered when it is reported so the text comes from the current catalog.
func messageOptionsParameterNotPassedThrough() rule.Message {
	return rule.Message{
		Id:          networkRequireHookOptionsParameterOptionsParameterNotPassedThroughText.Id,
		Description: networkRequireHookOptionsParameterOptionsParameterNotPassedThroughText.Render(nil),
	}
}

// messageIncorrectOptionsParameterType is the finding, rendered when it is reported so the text comes from the current catalog.
func messageIncorrectOptionsParameterType() rule.Message {
	return rule.Message{
		Id:          networkRequireHookOptionsParameterIncorrectOptionsParameterTypeText.Id,
		Description: networkRequireHookOptionsParameterIncorrectOptionsParameterTypeText.Render(nil),
	}
}

// inferOptionsTypeNames are the two generated option types a hook's options parameter may carry.
var inferOptionsTypeNames = map[string]bool{
	"InferUseGraphQlQueryOptions":    true,
	"InferUseGraphQlMutationOptions": true,
}

// NetworkRequireHookOptionsParameter flags a NetworkService hook that does not accept, type, and
// forward an options parameter.
//
//	valid:   function useUserRequest(options?: InferUseGraphQlQueryOptions<typeof Document>) {
//	             return networkService.useGraphQlQuery(Document, undefined, options)
//	         }
//	invalid: function useUserRequest() { return networkService.useGraphQlQuery(Document) }
//	invalid: function useUserRequest(options?: SomeOtherType) { ... }
//	invalid: function useUserRequest(options?: InferUseGraphQlQueryOptions<typeof Document>) {
//	             return networkService.useGraphQlQuery(Document)
//	         }
//
// Three message ids for three genuinely different defects. Missing means the caller has no control
// at all. Not-passed-through is worse, because the signature promises control that is silently
// dropped. Wrong type is the mildest: the option reaches the call and the caller gets no help
// writing it.
//
// Judged once per NetworkService call rather than once per hook. A hook calling both a query and a
// mutation has two contracts to satisfy and they differ in where options sit, so collapsing them to
// one verdict would report the wrong position for one of them.
//
// The options-argument position is derived rather than fixed, and this is the fiddliest part of the
// rule. A mutation takes options second. A query takes variables second and options third, except
// that a query with no variables may pass options second directly. The original decides by looking
// at what the second argument actually is, and that is reproduced here rather than simplified: a
// simplification would report "options must be third" at a call site where second is correct.
var NetworkRequireHookOptionsParameter = rule.Rule{
	Name:       "structure/network-require-hook-options-parameter",
	NoListener: rule.NoListenerAnswersInRun,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		analysis := NetworkFileAnalysisFor(ctx)
		if len(analysis.HookDeclarations) == 0 {
			return nil
		}

		for _, hook := range analysis.HookDeclarations {
			for _, call := range findNetworkServiceCalls(hook.Body) {
				checkHookOptionsPattern(ctx, hook, call)
			}
		}
		return nil
	},
}

// checkHookOptionsPattern judges one hook against one of its NetworkService calls.
func checkHookOptionsPattern(ctx rule.Context, hook NetworkHookDeclaration, call NetworkServiceCall) {
	isQuery := call.MethodName == "useGraphQlQuery" || call.MethodName == "useSuspenseGraphQlQuery"
	isMutation := call.MethodName == "useGraphQlMutation"
	if !isQuery && !isMutation {
		// `graphQlRequest` is not a hook call and takes no options object, so it imposes nothing.
		return
	}

	optionsArgumentIndex := optionsArgumentIndexFor(isQuery, call.Arguments)

	// The options parameter is the last one, by convention rather than by name search. A hook
	// taking `(variables, options)` and one taking `(options)` both put it last, and a rule that
	// searched by name would accept it anywhere, which is a different convention from the one the
	// codebase actually follows.
	var lastParameter *ast.Node
	if len(hook.Parameters) > 0 {
		lastParameter = hook.Parameters[len(hook.Parameters)-1]
	}

	lastParameterName := parameterName(lastParameter)
	hasOptionsParameter := strings.Contains(strings.ToLower(lastParameterName), "option")

	if hasOptionsParameter && !hasInferOptionsType(lastParameter) {
		ctx.ReportNode(lastParameter, messageIncorrectOptionsParameterType())
	}

	if !hasOptionsParameter {
		// Reported against the function rather than the name, unlike the naming rule: what is wrong
		// is the signature, so the range should cover the thing that needs changing.
		ctx.ReportNode(hook.FunctionNode, messageMissingOptionsParameter())
		return
	}

	if !callForwardsOptions(call, optionsArgumentIndex, lastParameterName) {
		// Reported against the call rather than the parameter. The parameter is correct; the call
		// is what fails to use it.
		ctx.ReportNode(call.Node, messageOptionsParameterNotPassedThrough())
	}
}

// optionsArgumentIndexFor works out which argument position options should occupy.
//
// A mutation always takes them second. A query takes them second when the second argument is
// already an options-looking identifier, and third otherwise, which covers both the
// `(document, variables, options)` shape and the `(document, undefined, options)` one. A query
// given three or more arguments is third regardless, since the third slot exists and is where they
// go.
func optionsArgumentIndexFor(isQuery bool, arguments []*ast.Node) int {
	if !isQuery {
		return 1
	}

	index := 1
	if len(arguments) >= 2 {
		secondArgument := ast.SkipParentheses(arguments[1])
		isOptionsIdentifier := secondArgument != nil &&
			secondArgument.Kind == ast.KindIdentifier &&
			strings.Contains(strings.ToLower(secondArgument.Text()), "option")
		if !isOptionsIdentifier {
			index = 2
		}
	}
	if len(arguments) >= 3 {
		index = 2
	}
	return index
}

// hasInferOptionsType reports whether a parameter carries one of the two generated option types.
func hasInferOptionsType(parameter *ast.Node) bool {
	return inferOptionsTypeNames[typeReferenceName(parameterTypeNode(parameter))]
}

// callForwardsOptions reports whether the call passes the named parameter at the expected position.
//
// Two shapes count: the parameter passed directly, and an object literal spreading it. The spread
// case matters because a hook that adds its own defaults writes
// `{ ...options, refetchOnMount: false }`, which forwards everything the caller supplied and would
// read as a violation to a rule that only accepted a bare identifier.
func callForwardsOptions(call NetworkServiceCall, index int, name string) bool {
	if name == "" || index >= len(call.Arguments) {
		return false
	}

	argument := ast.SkipParentheses(call.Arguments[index])
	if argument == nil {
		return false
	}

	if argument.Kind == ast.KindIdentifier {
		return argument.Text() == name
	}

	if argument.Kind == ast.KindObjectLiteralExpression {
		for _, property := range argument.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind != ast.KindSpreadAssignment {
				continue
			}
			spread := ast.SkipParentheses(property.AsSpreadAssignment().Expression)
			if spread != nil && spread.Kind == ast.KindIdentifier && spread.Text() == name {
				return true
			}
		}
	}

	return false
}
