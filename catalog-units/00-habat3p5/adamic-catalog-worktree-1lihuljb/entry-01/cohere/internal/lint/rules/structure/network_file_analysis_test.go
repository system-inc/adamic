package structure

import (
	"strconv"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The analysis is shared infrastructure rather than a rule, so it is tested directly. Three rules
// will depend on it, and a defect here would surface as three unrelated-looking rule bugs.
//
// A probe rule is the harness: it reports one finding per hook found, so the fixtures read the
// analysis through the same path a real rule will.
var networkAnalysisProbe = rule.Rule{
	Name: "network-analysis-probe",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				analysis := NetworkFileAnalysisFor(ctx)
				for _, hook := range analysis.HookDeclarations {
					exported := "local"
					if hook.IsExported {
						exported = "exported"
					}
					ctx.ReportNode(hook.NameNode, rule.Message{
						Id:          "hook",
						Description: hook.Name + " " + exported + " parameters=" + strconv.Itoa(len(hook.Parameters)),
					})
				}
			},
		}
	},
}

const networkImport = "import { networkService } from '@structure/source/services/network/NetworkService';\n"

func TestNetworkFileAnalysisFindsHooks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantCount  int
	}{
		{
			"a function declaration hook", networkImport +
				"export function useThingRequest() {\n    return networkService.useGraphQlQuery('q');\n}\n", 1,
		},
		{
			"an arrow hook", networkImport +
				"export const useThingRequest = () => networkService.useGraphQlQuery('q');\n", 1,
		},
		{
			"a function expression hook", networkImport +
				"export const useThingRequest = function () {\n    return networkService.useGraphQlMutation('m');\n};\n", 1,
		},
		// A call inside a callback still counts: the hook is what the caller sees, so the walk does
		// not stop at nested function boundaries. This is the opposite of the JSX walk's rule and
		// the difference is deliberate.
		{
			"a call nested inside a callback", networkImport +
				"declare function wrap(callback: () => unknown): unknown;\n" +
				"export function useThingRequest() {\n    return wrap(() => networkService.graphQlRequest('q'));\n}\n", 1,
		},
		{
			"two hooks in one file", networkImport +
				"export function useOneRequest() {\n    return networkService.useGraphQlQuery('a');\n}\n" +
				"export function useTwoRequest() {\n    return networkService.useGraphQlMutation('b');\n}\n", 2,
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, networkAnalysisProbe, networkFile, testCase.sourceText)
			if len(result.Diagnostics) != testCase.wantCount {
				t.Fatalf("want %d hooks, got %d", testCase.wantCount, len(result.Diagnostics))
			}
		})
	}
}

func TestNetworkFileAnalysisStaysEmpty(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The import gate. Without it the file cannot be calling NetworkService, and every rule
		// built on this early-exits, so a populated list with the flag false is a shape none of
		// them expect.
		{
			"no NetworkService import",
			"declare const networkService: { useGraphQlQuery(q: string): unknown };\n" +
				"export function useThingRequest() {\n    return networkService.useGraphQlQuery('q');\n}\n",
		},
		// A use-prefixed function that calls nothing on networkService is not a network hook.
		{
			"a hook that does not call NetworkService", networkImport +
				"export function useSomething() {\n    return 1;\n}\n",
		},
		// A closed method set, so another method on the same object does not qualify. The receiver
		// is the real `networkService` on purpose: an earlier version used a differently named
		// binding, so the receiver check caught it and the method set was never exercised.
		{
			"a call to a method outside the set", networkImport +
				"export function useThingRequest() {\n    return networkService.somethingElse();\n}\n",
		},
		// The receiver is matched by name, so a different object with the same method does not
		// count. This is the boundary between "calls NetworkService" and "calls something".
		{
			"the right method on the wrong object", networkImport +
				"declare const other: { useGraphQlQuery(q: string): unknown };\n" +
				"export function useThingRequest() {\n    return other.useGraphQlQuery('q');\n}\n",
		},
		// Not a hook by name, so the naming convention the dependent rules enforce does not apply.
		{
			"a non-hook function calling NetworkService", networkImport +
				"export function fetchThing() {\n    return networkService.graphQlRequest('q');\n}\n",
		},
		// Top-level only, matching the original. A hook nested inside another function is not the
		// file's exported surface.
		{
			"a hook nested inside another function", networkImport +
				"export function outer() {\n    function useInnerRequest() {\n" +
				"        return networkService.useGraphQlQuery('q');\n    }\n    return useInnerRequest;\n}\n",
		},
		{"an empty file", "export const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, networkAnalysisProbe, networkFile, testCase.sourceText))
		})
	}
}

// Export detection and parameter capture are what the dependent rules read, so they are pinned here
// rather than discovered three times.
func TestNetworkFileAnalysisRecordsShape(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		want       string
	}{
		{
			"an exported function hook with one parameter", networkImport +
				"export function useThingRequest(id: string) {\n    return networkService.useGraphQlQuery(id);\n}\n",
			"useThingRequest exported parameters=1",
		},
		{
			"a local arrow hook with two parameters", networkImport +
				"const useThingRequest = (id: string, options: object) => networkService.useGraphQlQuery(id);\n" +
				"export const Use = useThingRequest;\n",
			"useThingRequest local parameters=2",
		},
		{
			"an exported hook with no parameters", networkImport +
				"export const useThingRequest = () => networkService.useGraphQlQuery('q');\n",
			"useThingRequest exported parameters=0",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, networkAnalysisProbe, networkFile, testCase.sourceText)
			if len(result.Diagnostics) != 1 {
				t.Fatalf("want one hook, got %d", len(result.Diagnostics))
			}
			if got := result.Diagnostics[0].Message.Description; got != testCase.want {
				t.Fatalf("got %q, want %q", got, testCase.want)
			}
		})
	}
}
