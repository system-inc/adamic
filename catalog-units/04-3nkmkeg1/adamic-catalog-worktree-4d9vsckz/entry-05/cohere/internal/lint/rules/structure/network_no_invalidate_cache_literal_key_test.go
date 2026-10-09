package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const cacheKeyDeclarations = "declare const networkService: { cache: { invalidate(key: unknown): void } };\n" +
	"declare const UserCacheKey: string;\ndeclare const PostCacheKey: string;\ndeclare const id: string;\n"

func TestNetworkNoInvalidateCacheLiteralKeyFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
		wantIds    []string
	}{
		{
			"a string literal",
			"export function run() {\n    networkService.cache.invalidate('users');\n}\n",
			[]string{"noStringLiteralInvalidateCache"},
		},
		// A template is worse than a plain string, since interpolation makes the final key
		// invisible at the call site.
		{
			"a template literal with interpolation",
			"export function run() {\n    networkService.cache.invalidate(`users-${id}`);\n}\n",
			[]string{"noTemplateLiteralInvalidateCache"},
		},
		// A template with no substitutions is a string wearing backticks and reaches a different
		// node kind, so it needs its own case or the rule is silent on it.
		{
			"a template literal with no substitutions",
			"export function run() {\n    networkService.cache.invalidate(`users`);\n}\n",
			[]string{"noTemplateLiteralInvalidateCache"},
		},
		// Reported per element, so an array with one literal among imported keys points at the
		// literal rather than at the array.
		{
			"one literal among imported keys",
			"export function run() {\n    networkService.cache.invalidate([UserCacheKey, 'posts', PostCacheKey]);\n}\n",
			[]string{"noArrayWithStringLiteralInvalidateCache"},
		},
		{
			"two literals in one array report twice",
			"export function run() {\n    networkService.cache.invalidate(['users', 'posts']);\n}\n",
			[]string{"noArrayWithStringLiteralInvalidateCache", "noArrayWithStringLiteralInvalidateCache"},
		},
		{
			"a template inside an array",
			"export function run() {\n    networkService.cache.invalidate([`users-${id}`]);\n}\n",
			[]string{"noTemplateLiteralInvalidateCache"},
		},
		// The receiver is deliberately not judged here, unlike the sibling rule, so a cache that is
		// nobody's service still has its keys checked.
		{
			"a cache that is not a service's",
			"declare const store: { cache: { invalidate(key: unknown): void } };\nexport function run() {\n    store.cache.invalidate('users');\n}\n",
			[]string{"noStringLiteralInvalidateCache"},
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NetworkNoInvalidateCacheLiteralKey, cacheFile,
				cacheKeyDeclarations+testCase.sourceText), testCase.wantIds...)
		})
	}
}

func TestNetworkNoInvalidateCacheLiteralKeyStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{"an imported key", "export function run() {\n    networkService.cache.invalidate(UserCacheKey);\n}\n"},
		{"an array of imported keys", "export function run() {\n    networkService.cache.invalidate([UserCacheKey, PostCacheKey]);\n}\n"},
		{"an empty array", "export function run() {\n    networkService.cache.invalidate([]);\n}\n"},
		{"a call with no arguments", "export function run() {\n    networkService.cache.invalidate(undefined);\n}\n"},
		// The shape check is what these break, each in a different place.
		{
			"invalidate reached directly off a binding",
			"declare const direct: { invalidate(key: unknown): void };\nexport function run() {\n    direct.invalidate('users');\n}\n",
		},
		{
			"invalidate reached through a property that is not cache",
			"declare const holder: { store: { invalidate(key: unknown): void } };\nexport function run() {\n    holder.store.invalidate('users');\n}\n",
		},
		{
			"a cache method that is not invalidate",
			"declare const svc: { cache: { clear(key: unknown): void } };\nexport function run() {\n    svc.cache.clear('users');\n}\n",
		},
		// A literal passed somewhere else entirely is not this rule's business.
		{"a literal in an unrelated call", "declare function log(message: string): void;\nexport function run() {\n    log('users');\n}\n"},
		{"no calls at all", "export const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NetworkNoInvalidateCacheLiteralKey, cacheFile,
				cacheKeyDeclarations+testCase.sourceText))
		})
	}
}
