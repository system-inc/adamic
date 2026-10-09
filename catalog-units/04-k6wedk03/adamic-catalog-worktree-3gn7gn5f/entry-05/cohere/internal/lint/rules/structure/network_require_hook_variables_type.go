package structure

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

// The rule's messages, one handle per id, whose wording lives in
// `policy/messages/network-require-hook-variables-type.json`.
var (
	networkRequireHookVariablesTypeInlineVariablesTypeText               = policy.MessageOf("structure/network-require-hook-variables-type", "inlineVariablesType")
	networkRequireHookVariablesTypeUnnecessaryVariablesDestructuringText = policy.MessageOf("structure/network-require-hook-variables-type", "unnecessaryVariablesDestructuring")
)

// messageInlineVariablesType is the finding, rendered when it is reported so the text comes from the current catalog.
func messageInlineVariablesType() rule.Message {
	return rule.Message{
		Id:          networkRequireHookVariablesTypeInlineVariablesTypeText.Id,
		Description: networkRequireHookVariablesTypeInlineVariablesTypeText.Render(nil),
	}
}

// messageUnnecessaryVariablesDestructuring is the finding, rendered when it is reported so the text comes from the current catalog.
func messageUnnecessaryVariablesDestructuring() rule.Message {
	return rule.Message{
		Id:          networkRequireHookVariablesTypeUnnecessaryVariablesDestructuringText.Id,
		Description: networkRequireHookVariablesTypeUnnecessaryVariablesDestructuringText.Render(nil),
	}
}

// NetworkRequireHookVariablesType flags a hook whose variables parameter is typed inline, or whose
// call rebuilds that parameter property by property.
//
//	valid:   function useUserRequest(variables: UserQueryVariablesType) {
//	             return networkService.useGraphQlQuery(Document, variables)
//	         }
//	invalid: function useUserRequest(variables: { identifier: string }) { ... }
//	invalid: function useUserRequest(variables: UserQueryVariablesType) {
//	             return networkService.useGraphQlQuery(Document, { identifier: variables.identifier })
//	         }
//
// Two message ids because they are two defects that happen to co-occur. The type is about what the
// caller is asked for; the rebuild is about what actually reaches the server.
//
// A parameter is judged as variables in one of two ways, and the second is the interesting one. A
// parameter named for variables is judged on its annotation directly. A parameter named anything
// else is judged only when its inline type holds properties that give it away as a variables object
// (input, id, identifier, slug, pagination). Without that second arm a rule keyed on the name alone
// would miss the case people actually write, which is a parameter named for the thing it fetches.
//
// The rebuild check is deliberately narrow: only a query, only the second argument, and only when
// every property is a same-named read off the variables parameter. `{ identifier: variables.slug }`
// is a rename rather than a rebuild and is left alone, because collapsing it to the parameter would
// change what the call sends.
var NetworkRequireHookVariablesType = rule.Rule{
	Name:       "structure/network-require-hook-variables-type",
	NoListener: rule.NoListenerAnswersInRun,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		analysis := NetworkFileAnalysisFor(ctx)
		if len(analysis.HookDeclarations) == 0 {
			return nil
		}

		for _, hook := range analysis.HookDeclarations {
			for _, call := range findNetworkServiceCalls(hook.Body) {
				checkHookVariablesPattern(ctx, hook, call)
			}
		}
		return nil
	},
}

// graphQlLikePropertyNames are the property names that mark an inline type as a variables object.
//
// A closed list rather than a heuristic, matching the original. These are the arguments nearly every
// operation in this codebase takes, and a wider test would start flagging ordinary parameter types
// that happen to carry an id.
var graphQlLikePropertyNames = map[string]bool{
	"input":      true,
	"id":         true,
	"identifier": true,
	"slug":       true,
	"pagination": true,
}

// checkHookVariablesPattern judges one hook against one of its NetworkService calls.
func checkHookVariablesPattern(ctx rule.Context, hook NetworkHookDeclaration, call NetworkServiceCall) {
	isQuery := call.MethodName == "useGraphQlQuery" || call.MethodName == "useSuspenseGraphQlQuery"
	isMutation := call.MethodName == "useGraphQlMutation"
	if !isQuery && !isMutation {
		return
	}

	for _, parameter := range hook.Parameters {
		name := strings.ToLower(parameterName(parameter))

		// The options parameter is another rule's business, and judging it here would report the
		// generated options type as a wrong variables type.
		if strings.Contains(name, "option") {
			continue
		}

		if strings.Contains(name, "variable") || name == "input" {
			if !hasGeneratedVariablesType(parameter) {
				ctx.ReportNode(parameter, messageInlineVariablesType())
			}
			continue
		}

		// A parameter named for what it fetches rather than for being variables. Judged only on the
		// shape of an inline type, since a named type here is very often a legitimate domain type
		// that has nothing to do with this document.
		if inlineTypeLooksLikeVariables(parameterTypeNode(parameter)) {
			ctx.ReportNode(parameter, messageInlineVariablesType())
		}
	}

	if isQuery {
		checkVariablesRebuild(ctx, hook, call)
	}
}

// hasGeneratedVariablesType reports whether a parameter carries a generated variables type.
//
// An unannotated parameter passes. The rule is about a hand-written type competing with the
// generated one, and no annotation at all is a different complaint that belongs to the type checker.
func hasGeneratedVariablesType(parameter *ast.Node) bool {
	typeNode := parameterTypeNode(parameter)
	if typeNode == nil {
		return true
	}

	// An inline object type is the defect this rule exists for, named or not.
	if typeNode.Kind == ast.KindTypeLiteral {
		return false
	}

	name := typeReferenceName(typeNode)
	if name == "" {
		// Anything that is not a plain named reference (a union, a mapped type, a qualified name)
		// is left alone. The original does the same, and widening here would report shapes nobody
		// has decided about.
		return true
	}
	return strings.HasSuffix(name, "QueryVariablesType") || strings.HasSuffix(name, "MutationVariablesType")
}

// inlineTypeLooksLikeVariables reports an inline object type carrying a telltale GraphQL property.
func inlineTypeLooksLikeVariables(typeNode *ast.Node) bool {
	if typeNode == nil || typeNode.Kind != ast.KindTypeLiteral {
		return false
	}

	for _, member := range typeNode.AsTypeLiteralNode().Members.Nodes {
		if member.Kind != ast.KindPropertySignature {
			continue
		}
		propertyName := member.AsPropertySignatureDeclaration().Name()
		if propertyName == nil || propertyName.Kind != ast.KindIdentifier {
			continue
		}
		if graphQlLikePropertyNames[strings.ToLower(propertyName.Text())] {
			return true
		}
	}
	return false
}

// checkVariablesRebuild flags a second argument that reassembles the variables parameter verbatim.
func checkVariablesRebuild(ctx rule.Context, hook NetworkHookDeclaration, call NetworkServiceCall) {
	if len(call.Arguments) < 2 {
		return
	}
	argument := ast.SkipParentheses(call.Arguments[1])
	if argument == nil || argument.Kind != ast.KindObjectLiteralExpression {
		return
	}
	properties := argument.AsObjectLiteralExpression().Properties.Nodes
	if len(properties) == 0 {
		return
	}

	variablesName := ""
	for _, parameter := range hook.Parameters {
		name := parameterName(parameter)
		if strings.Contains(strings.ToLower(name), "variable") {
			variablesName = name
			break
		}
	}
	if variablesName == "" {
		return
	}

	for _, property := range properties {
		if !isSameNamedReadOff(property, variablesName) {
			return
		}
	}

	ctx.ReportNode(argument, messageUnnecessaryVariablesDestructuring())
}

// isSameNamedReadOff reports `name: variables.name`, where the key and the property agree.
//
// A spread disqualifies the whole object rather than being skipped: `{ ...variables, extra: 1 }` is
// not a rebuild of the parameter, it is the parameter plus something, and collapsing it would drop
// the something.
func isSameNamedReadOff(property *ast.Node, variablesName string) bool {
	if property.Kind != ast.KindPropertyAssignment {
		return false
	}
	assignment := property.AsPropertyAssignment()

	// A shorthand property (`{ identifier }`) has no initializer, and SkipParentheses dereferences
	// its argument, so the nil check has to come first.
	if assignment.Initializer == nil {
		return false
	}
	value := ast.SkipParentheses(assignment.Initializer)
	if value == nil || value.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := value.AsPropertyAccessExpression()
	receiver := ast.SkipParentheses(access.Expression)
	if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.Text() != variablesName {
		return false
	}

	keyName := propertyKeyText(assignment.Name())
	accessedName := propertyKeyText(access.Name())
	return keyName != "" && keyName == accessedName
}

// propertyKeyText reads an identifier or string-literal key as text, and anything else as "".
//
// `property.Textual` rather than the wider set. This compares a key against the property read off
// the variables parameter, so both sides go through the same accept set and a spelling neither side
// accepts simply fails to match, which is the answer this rule wants.
func propertyKeyText(node *ast.Node) string {
	text, _ := property.Name(node, property.Textual)
	return text
}
