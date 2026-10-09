package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const networkFile = "/repository/source/api/Thing.ts"

func TestNetworkNoDirectFetchFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The three spellings that reach the same primitive. A rule catching only the bare form is
		// worked around by writing window.fetch, which nobody does deliberately and every copied
		// snippet contains.
		{"a bare call", "export async function load(url: string) {\n    return fetch(url);\n}\n"},
		{"through window", "export async function load(url: string) {\n    return window.fetch(url);\n}\n"},
		{"through globalThis", "export async function load(url: string) {\n    return globalThis.fetch(url);\n}\n"},
		{"awaited", "export async function load(url: string) {\n    const response = await fetch(url);\n    return response;\n}\n"},
		{"inside a callback", "declare function later(callback: () => void): void;\nexport function load(url: string) {\n    later(() => { void fetch(url); });\n}\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NetworkNoDirectFetch, networkFile, testCase.sourceText),
				"noDirectFetch")
		})
	}

	// NetworkService is exempt by where it lives, so a file elsewhere named after it is not.
	t.Run("in a file merely named after NetworkService", func(t *testing.T) {
		t.Parallel()
		rule_testing.ExpectFindings(t, rule_testing.Run(t, NetworkNoDirectFetch, "/repository/app/_services/MyNetworkService.ts",
			"export async function load(url: string) {\n    return fetch(url);\n}\n"), "noDirectFetch")
	})
}

func TestNetworkNoDirectFetchStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The implementation is the one place the raw primitive is allowed, and a rule that
		// forbids the thing it asks people to use is a rule people learn to disable.
		{
			"inside NetworkService.ts", "/repository/source/services/network/NetworkService.ts",
			"export async function load(url: string) {\n    return fetch(url);\n}\n",
		},
		// The module spans internal/: the device id request must use the raw primitive, because
		// NetworkService's request() waits on the very device id it is fetching.
		{
			"inside NetworkService's internal device id", "/repository/libraries/structure/source/services/network/internal/NetworkServiceDeviceId.ts",
			"export async function requestDeviceId(url: string) {\n    return fetch(url);\n}\n",
		},
		// The decision boundary is "is this the global primitive", and the way to get it wrong is
		// to match the method name. Both of these are somebody's own method.
		{
			"a method named fetch on a service", networkFile,
			"declare const networkService: { fetch(url: string): Promise<unknown> };\nexport async function load(url: string) {\n    return networkService.fetch(url);\n}\n",
		},
		{
			"a method named fetch on an arbitrary object", networkFile,
			"declare const client: { fetch(url: string): Promise<unknown> };\nexport async function load(url: string) {\n    return client.fetch(url);\n}\n",
		},
		{
			"the wrapper being used as intended", networkFile,
			"declare const networkService: { request(url: string): Promise<unknown> };\nexport async function load(url: string) {\n    return networkService.request(url);\n}\n",
		},
		{"a binding merely named fetch", networkFile, "export const fetchCount = 1;\nexport const Use = fetchCount;\n"},
		{"the word in a string", networkFile, "export const Message = 'fetch(url)';\nexport const Use = Message;\n"},
		{"no calls at all", networkFile, "export const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NetworkNoDirectFetch, testCase.fileName, testCase.sourceText))
		})
	}
}
