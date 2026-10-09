package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const localStorageFile = "/repository/source/components/Panel.tsx"
const localStorageServiceFile = "/repository/source/services/local-storage/LocalStorageService.ts"
const localStorageInternalFile = "/repository/source/services/local-storage/internal/LocalStorageServiceUtilities.ts"

func TestStorageNoDirectLocalStorageFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantId     string
	}{
		{
			"a direct method call",
			"export function run() { return localStorage.getItem('key'); }\n",
			"directLocalStorage",
		},
		{
			"a direct assignment",
			"export function run() { localStorage.setItem('key', 'value'); }\n",
			"directLocalStorage",
		},
		{
			"the identifier held in a variable",
			"export function run() { const storage = localStorage; return storage; }\n",
			"directLocalStorage",
		},
		{
			"the identifier passed as an argument",
			"declare function persist(storage: unknown): void;\nexport function run() { persist(localStorage); }\n",
			"directLocalStorage",
		},
		{
			"the identifier as an object value",
			"export function run() { return { storage: localStorage }; }\n",
			"directLocalStorage",
		},
		{
			"a window method call",
			"export function run() { return window.localStorage.getItem('key'); }\n",
			"windowLocalStorage",
		},
		// The case the original cannot see, and the reason this rule diverges from the gate.
		// `window.localStorage` passed as a value falls between the original's two branches: its
		// window branch needs an outer member access and its identifier branch skips anything whose
		// parent is one. There is exactly one occurrence in the tree, and it is a real violation.
		{
			"window.localStorage passed as a value",
			"declare function persist(options: { storage: unknown }): void;\n" +
				"export function run() { persist({ storage: window.localStorage }); }\n",
			"windowLocalStorage",
		},
		{
			"window.localStorage held in a variable",
			"export function run() { const storage = window.localStorage; return storage; }\n",
			"windowLocalStorage",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, StorageNoDirectLocalStorage, localStorageFile,
				testCase.sourceText), testCase.wantId)
		})
	}
}

func TestStorageNoDirectLocalStorageStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		fileName   string
		sourceText string
	}{
		// The shape the rule asks for.
		{
			"the service",
			localStorageFile,
			"declare const localStorageService: { get(key: string): unknown };\n" +
				"export function run() { return localStorageService.get('key'); }\n",
		},
		// The exemptions, which are what make the rule satisfiable: the service has to reach the
		// real thing or the rule forbids what it recommends.
		{
			"the service implementation",
			localStorageServiceFile,
			"export function run() { return localStorage.getItem('key'); }\n",
		},
		{
			"the service internal utilities",
			localStorageInternalFile,
			"export function run() { return window.localStorage.getItem('key'); }\n",
		},
		// A property named localStorage is a name rather than a use, in each of the places a name
		// can appear.
		{
			"a property key named localStorage",
			localStorageFile,
			"export function run() { return { localStorage: 1 }; }\n",
		},
		{
			"a type member named localStorage",
			localStorageFile,
			"export interface StorageInterface { localStorage: unknown }\n",
		},
		{
			"a parameter named localStorage",
			localStorageFile,
			"export function run(localStorage: unknown) { return 1; }\n",
		},
		// A localStorage property on something that is not window is somebody's own field.
		{
			"a localStorage property on another object",
			localStorageFile,
			"declare const host: { localStorage: unknown };\nexport function run() { return host.localStorage; }\n",
		},
		{
			"a file with no storage at all",
			localStorageFile,
			"export const value = 1;\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, StorageNoDirectLocalStorage, testCase.fileName,
				testCase.sourceText))
		})
	}
}
