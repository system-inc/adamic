package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

// The component below is the shape the cache exists for: a `useMemo`, a state setter, and a JSX
// tag, so all three rules that lower functions would want it.
const cacheProbeSource = `
import {useState, useMemo, useEffect} from 'react';
export function Widget({items}: {items: number[]}) {
	const [count, setCount] = useState(0);
	let label = 'low';
	if (items.length > 1) { label = 'high'; } else { label = 'one'; }
	const total = useMemo(() => items.reduce((a, b) => a + b, 0), [items]);
	useEffect(() => { setCount(total); }, [total]);
	return <div>{label}{count}</div>;
}
`

// totalPhis counts phis across a function AND every nested function.
//
// `CollectSSAStats` reports the top-level function only, which is why the first version of the test
// below could not see the defect it exists to catch: repeated construction adds phis wherever the
// merges are, and in a real component that is often a nested arrow. A whole-tree count is the only
// one that stays true as the probe source changes.
func totalPhis(function *Function) int {
	if function == nil {
		return 0
	}
	total := 0
	for _, block := range function.Blocks {
		total += len(block.Phis)
	}
	for _, nested := range function.Functions {
		total += totalPhis(nested)
	}
	return total
}

// TestForFunctionLowersOncePerFunction is the assertion that the cache is HIT.
//
// A faster timing number cannot distinguish a cache that hits perfectly from one that never hits,
// so this counts. It asks for the same function three times, the way the three react rules do, and
// requires exactly one lowering.
//
// It runs against the real `ForFunction` and a real `rule.Context` from the typed harness rather
// than against a stand-in cache, because the property under test is that the PRODUCTION path shares
// work. Counting is possible without a hook into `Lower` because a lowering is a distinct pointer
// every time it runs: three asks returning one pointer is one lowering.
func TestForFunctionLowersOncePerFunction(t *testing.T) {
	t.Parallel()

	var results []*Function
	var directLowering *Function

	probe := rule.Rule{
		Name:             "hir-cache-probe",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						t.Fatal("the harness produced no type checker, so this test would prove nothing")
					}
					if ctx.FileCache == nil {
						t.Fatal("the harness produced no file cache, so a hit could not be observed")
					}
					forEachFunctionLike(node, func(function *ast.Node) {
						if len(results) > 0 {
							return
						}
						// Three asks, standing in for the three rules.
						for range 3 {
							results = append(results, ForFunction(ctx, function))
						}
						// An independent lowering of the same node, to confirm that a fresh one is
						// genuinely a different object and the identity check above has teeth.
						directLowering = Lower(function, ctx.TypeChecker)
					})
				},
			}
		},
	}
	rule_testing.RunTyped(t, probe, "probe.tsx", cacheProbeSource)

	if len(results) != 3 {
		t.Fatalf("want three asks recorded, got %d", len(results))
	}
	if results[0] == nil {
		t.Fatal("the probe function did not lower, so this test proves nothing")
	}
	if results[1] != results[0] || results[2] != results[0] {
		t.Errorf("each ask lowered again: %p, %p, %p — the cache is not being hit",
			results[0], results[1], results[2])
	}
	if directLowering == results[0] {
		t.Fatal("a fresh Lower returned the cached pointer, so pointer identity proves nothing here")
	}
}

// TestForFunctionConstructsExactlyOnce is the correctness half, and the reason Construct lives
// inside the cached computation.
//
// `Construct` mutates in place and is not idempotent: a second run over one graph adds phis that
// correspond to no branch. The doubly-constructed graph still passes `VerifySSA`, so nothing would
// have caught this at runtime. Asserting the phi count is stable across repeated asks is what
// pins the design.
func TestForFunctionConstructsExactlyOnce(t *testing.T) {
	t.Parallel()

	var firstPhis, lastPhis int

	probe := rule.Rule{
		Name:             "hir-cache-construct-probe",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					done := false
					forEachFunctionLike(node, func(function *ast.Node) {
						if done {
							return
						}
						done = true
						first := ForFunction(ctx, function)
						if first == nil {
							t.Fatal("the probe function did not lower")
						}
						firstPhis = totalPhis(first)
						for range 3 {
							lastPhis = totalPhis(ForFunction(ctx, function))
						}
					})
				},
			}
		},
	}
	rule_testing.RunTyped(t, probe, "probe.tsx", cacheProbeSource)

	if firstPhis == 0 {
		t.Fatal("the probe lowered to zero phis, so a repeated Construct could not be detected — " +
			"this test would pass against any implementation")
	}
	if firstPhis != lastPhis {
		t.Errorf("repeated asks changed the graph: %d phis then %d — Construct ran more than once",
			firstPhis, lastPhis)
	}
}

// TestForFunctionWithoutCacheStillLowers pins the nil-cache contract.
//
// `rule.Cached` treats a nil cache as "no caching, compute every time", so a harness that builds a
// Context by hand keeps working. That contract is load-bearing here and easy to break by reaching
// for `ctx.FileCache` directly.
func TestForFunctionWithoutCacheStillLowers(t *testing.T) {
	t.Parallel()

	var withoutCache *Function

	probe := rule.Rule{
		Name:             "hir-cache-nil-probe",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					bare := ctx
					bare.FileCache = nil
					done := false
					forEachFunctionLike(node, func(function *ast.Node) {
						if done {
							return
						}
						done = true
						withoutCache = ForFunction(bare, function)
					})
				},
			}
		},
	}
	rule_testing.RunTyped(t, probe, "probe.tsx", cacheProbeSource)

	if withoutCache == nil {
		t.Fatal("a nil cache must still compute, but nothing came back")
	}
	if len(withoutCache.Blocks) == 0 {
		t.Fatal("a nil cache returned an empty graph rather than a real lowering")
	}
}

// TestForFunctionRecordsOneFillPerFunction proves the cache participates in the timing table's
// shared-fill accounting, which is what moves the lowering cost off whichever rule asked first.
//
// This is the second independent check that a hit is real. The pointer test shows one object comes
// back; this shows the cache recorded exactly one FILL for the function, so the work ran once by
// the cache's own accounting rather than only by identity.
func TestForFunctionRecordsOneFillPerFunction(t *testing.T) {
	t.Parallel()

	var fillKeys []string
	var askedKey string

	probe := rule.Rule{
		Name:             "hir-cache-fill-probe",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					done := false
					forEachFunctionLike(node, func(function *ast.Node) {
						if done {
							return
						}
						done = true
						askedKey = cacheKeyFor(function)
						for range 3 {
							ForFunction(ctx, function)
						}
					})
					for key := range ctx.FileCache.Fills() {
						fillKeys = append(fillKeys, key)
					}
				},
			}
		},
	}
	rule_testing.RunTyped(t, probe, "probe.tsx", cacheProbeSource)

	matching := 0
	for _, key := range fillKeys {
		if key == askedKey {
			matching++
		}
	}
	if matching != 1 {
		t.Errorf("want exactly one recorded fill for %q after three asks, got %d (keys: %v)",
			askedKey, matching, fillKeys)
	}
}

// TestForFunctionWithoutManualMemoizationSeesEscapedSpellings pins the gate in front of the erasure.
//
// The gate hands a function the memo-intact graph when its span cannot name a memo hook, and a
// wrong "cannot" is silent: the erasure never runs and every rule reading the graph judges a call
// upstream has already removed. A plain substring search made that mistake for every escaped
// spelling below, which reaches `DropManualMemoization` through the cooked identifier or string
// text while the span never holds the name's letters in a row.
//
// Each case is first proven to matter: an independent erasure over a fresh lowering must recognise
// a call, or the case would assert nothing about the gate. The memo-free control is the other
// direction, since a gate that always answered "yes" would pass every escaped case and must still
// share the cached graph here.
func TestForFunctionWithoutManualMemoizationSeesEscapedSpellings(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		source     string
		recognised bool
	}{
		{"plain", "import {useMemo} from 'react';\nexport function Component({value}: {value: number}) {\n\treturn useMemo(() => value * 2, [value]);\n}\n", true},
		{"identifierFourDigitEscape", "import {use\\u004Demo} from 'react';\nexport function Component({value}: {value: number}) {\n\treturn use\\u004Demo(() => value * 2, [value]);\n}\n", true},
		{"identifierCodePointEscape", "import {use\\u{43}allback} from 'react';\nexport function Component({value}: {value: number}) {\n\treturn use\\u{43}allback(() => value, [value]);\n}\n", true},
		{"computedHexEscape", "import * as React from 'react';\nexport function Component({value}: {value: number}) {\n\treturn React['use\\x4Demo'](() => value * 2, [value]);\n}\n", true},
		{"propertyNameEscape", "import * as React from 'react';\nexport function Component({value}: {value: number}) {\n\treturn React.use\\u0043allback(() => value, [value]);\n}\n", true},
		{"memoFreeWithBackslash", "export function Component({value}: {value: number}) {\n\treturn 'a\\tb' + value;\n}\n", false},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var shared, erased *Function
			recognised := 0
			probe := rule.Rule{
				Name:             "hir-cache-escaped-memo-probe",
				NeedsTypeChecker: true,
				Run: func(ctx rule.Context, options any) rule.Listeners {
					return rule.Listeners{
						ast.KindSourceFile: func(node *ast.Node) {
							done := false
							forEachFunctionLike(node, func(function *ast.Node) {
								if done {
									return
								}
								done = true
								independent := Lower(function, ctx.TypeChecker)
								Construct(independent)
								recognised = DropManualMemoization(independent).Recognised
								shared = ForFunction(ctx, function)
								erased = ForFunctionWithoutManualMemoization(ctx, function)
							})
						},
					}
				},
			}
			rule_testing.RunTyped(t, probe, "probe.tsx", testCase.source)

			if shared == nil || erased == nil {
				t.Fatal("the probe function did not lower, so this case proves nothing")
			}
			if (recognised > 0) != testCase.recognised {
				t.Fatalf("an independent erasure recognised %d calls, so this case does not test what it names", recognised)
			}
			if testCase.recognised && erased == shared {
				t.Error("the erasure recognises a memo call here, but the gate handed back the memo-intact graph")
			}
			if !testCase.recognised && erased != shared {
				t.Error("nothing here can be erased, but the gate lowered a second graph instead of sharing the first")
			}
		})
	}
}

// TestMayHoldComponentOrHookAnswersFromTheText pins the per-file gate in front of all seven React
// Compiler rules that lower.
//
// A wrong "no" is silent in the worst way: the file is never lowered, so every rule goes quiet on
// it with no error. So each spelling that can reach a component or hook test is a "yes" row, an
// escape among them, and the "no" rows are what make the gate worth having: a plain module, and a
// file whose only backslash is an ordinary string escape.
func TestMayHoldComponentOrHookAnswersFromTheText(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name, file, source string
		want               bool
	}{
		{"a plain module", "plain.ts", "export const total = (values: number[]) => values.reduce((a, b) => a + b, 0);\n", false},
		{"a backslash that spells no hook", "escape.ts", "export const line = 'a\\tb';\n", false},
		{"a JSX-variant file with no tag and no hook", "quiet.tsx", "export const count = 1;\n", false},
		{"a hook call", "hook.ts", "import {useState} from 'react';\nexport function useCounter() { return useState(0); }\n", true},
		{"a digit after use", "digit.ts", "declare function use2Things(): number;\nexport function useBoth() { return use2Things(); }\n", true},
		{"an escaped hook name", "escaped.ts", "import {use\\u0053tate} from 'react';\nexport function use\\u0043ounter() { return use\\u0053tate(0); }\n", true},
		{"an escaped hook name in a computed member", "computed.ts", "declare const React: Record<string, (value: number) => number>;\nexport function counter() { return React['use\\x53tate'](0); }\n", true},
		{"a tag in a JSX-variant file", "tag.tsx", "export function Shown() { return <div />; }\n", true},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			answered := false
			var got bool
			probe := rule.Rule{
				Name:             "hir-may-hold-component-or-hook-probe",
				NeedsTypeChecker: true,
				Run: func(ctx rule.Context, options any) rule.Listeners {
					return rule.Listeners{
						ast.KindSourceFile: func(node *ast.Node) {
							answered = true
							got = MayHoldComponentOrHook(ctx)
						},
					}
				},
			}
			rule_testing.RunTyped(t, probe, testCase.file, testCase.source)

			if !answered {
				t.Fatal("the probe never ran, so this case proves nothing")
			}
			if got != testCase.want {
				t.Errorf("MayHoldComponentOrHook = %v, want %v", got, testCase.want)
			}
		})
	}
}
