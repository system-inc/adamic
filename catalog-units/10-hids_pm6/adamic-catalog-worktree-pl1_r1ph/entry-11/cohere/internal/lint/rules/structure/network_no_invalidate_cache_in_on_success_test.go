package structure

import (
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const cacheFile = "/repository/source/api/Thing.ts"

// declarations the fixtures share, so each case is only the shape under test.
const cacheDeclarations = "declare const networkService: { cache: { invalidate(key: string): void } };\n" +
	"declare const apiService: { cache: { invalidate(key: string): void } };\n" +
	"declare function useMutation(options: unknown): void;\n"

func TestNetworkNoInvalidateCacheInOnSuccessFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		{
			"an arrow handler",
			"export function run(key: string) {\n    useMutation({ onSuccess: () => networkService.cache.invalidate(key) });\n}\n",
		},
		// The method shorthand reaches a different node kind than the assignment form, so a port
		// reading only one of them misses half the call sites. Enumerated before the listener.
		{
			"the method shorthand",
			"export function run(key: string) {\n    useMutation({ onSuccess() { networkService.cache.invalidate(key); } });\n}\n",
		},
		{
			"a function expression handler",
			"export function run(key: string) {\n    useMutation({ onSuccess: function () { networkService.cache.invalidate(key); } });\n}\n",
		},
		// The receiver test is the original's and it is deliberately loose: any name containing
		// "service" in either casing counts, not just networkService.
		{
			"a differently named service",
			"export function run(key: string) {\n    useMutation({ onSuccess: () => apiService.cache.invalidate(key) });\n}\n",
		},
		{
			"nested inside the handler body",
			"export function run(key: string, flag: boolean) {\n    useMutation({ onSuccess() { if(flag) { networkService.cache.invalidate(key); } } });\n}\n",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectFindings(t, rule_testing.Run(t, NetworkNoInvalidateCacheInOnSuccess, cacheFile,
				cacheDeclarations+testCase.sourceText), "noInvalidateCacheInOnSuccess")
		})
	}
}

func TestNetworkNoInvalidateCacheInOnSuccessStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		sourceText string
	}{
		// The intended shape: the invalidation list lives next to the request.
		{"the option form", "export function run() {\n    useMutation({ invalidateOnSuccess: ['users'] });\n}\n"},
		// Three things must line up, and each of these breaks exactly one of them.
		{
			"an invalidation outside onSuccess",
			"export function run(key: string) {\n    networkService.cache.invalidate(key);\n}\n",
		},
		{
			"a different handler name",
			"export function run(key: string) {\n    useMutation({ onError: () => networkService.cache.invalidate(key) });\n}\n",
		},
		{
			"a cache that is not a service's",
			"declare const store: { cache: { invalidate(key: string): void } };\nexport function run(key: string) {\n    useMutation({ onSuccess: () => store.cache.invalidate(key) });\n}\n",
		},
		{
			"a service method that is not invalidate",
			"declare const other: { cache: { clear(key: string): void } };\nexport function run(key: string) {\n    useMutation({ onSuccess: () => other.cache.clear(key) });\n}\n",
		},
		// A direct .invalidate with no intermediate property is a different call, stopped by the
		// shape check before the name check ever runs.
		{
			"invalidate reached directly off a binding",
			"declare const networkService2: { invalidate(key: string): void };\nexport function run(key: string) {\n    useMutation({ onSuccess: () => networkService2.invalidate(key) });\n}\n",
		},
		// Two hops, but the middle one is not `cache`. This is the case that actually exercises the
		// name check: the shape matches all the way down and only the property name differs. Two
		// earlier attempts at this fixture were stopped by the shape check instead and so proved
		// nothing about the name.
		{
			"invalidate reached through a property that is not cache",
			"declare const networkService3: { store: { invalidate(key: string): void } };\nexport function run(key: string) {\n    useMutation({ onSuccess: () => networkService3.store.invalidate(key) });\n}\n",
		},
		{"an empty handler", "export function run() {\n    useMutation({ onSuccess: () => {} });\n}\n"},
		{"no mutation at all", "export const Value = 1;\n"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, rule_testing.Run(t, NetworkNoInvalidateCacheInOnSuccess, cacheFile,
				cacheDeclarations+testCase.sourceText))
		})
	}
}
