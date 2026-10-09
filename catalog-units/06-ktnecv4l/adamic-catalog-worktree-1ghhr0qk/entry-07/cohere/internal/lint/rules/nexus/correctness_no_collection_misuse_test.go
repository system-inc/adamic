package nexus

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoCollectionMisuseFile = "/repository/source/CorrectnessNoCollectionMisuse.ts"

// correctnessNoCollectionMisusePrelude declares every name the cases use, so the checker resolves
// each one and no verdict rests on an unresolved identifier.
var correctnessNoCollectionMisusePrelude = strings.Join([]string{
	"type Role = 'Admin' | 'Member';",
	"declare const roles: Role[];",
	"declare const role: Role;",
	"declare const pair: readonly ['first', 'second'];",
	"declare const sparse: (number | undefined)[];",
	"declare const index: number;",
	"declare const wideKey: string;",
	"declare const items: string[];",
	"declare const optionalItems: string[] | undefined;",
	"declare const text: string;",
	"declare const bytes: Uint8Array;",
	"declare const cache: Map<string, number>;",
	"declare const lookup: ReadonlyMap<string, number>;",
	"declare const seen: Set<string>;",
	"declare const weak: WeakMap<object, number>;",
	"declare const record: Record<string, number>;",
	"declare const lengthy: { length: number };",
	"declare const arrayLike: ArrayLike<number>;",
	"declare const count: number;",
	"declare function use(value: unknown): void;",
	"",
}, "\n")

func correctnessNoCollectionMisuseSource(lines ...string) string {
	return correctnessNoCollectionMisusePrelude + strings.Join(lines, "\n") + "\n"
}

func correctnessNoCollectionMisuseRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessNoCollectionMisuse, map[string]string{
		correctnessNoCollectionMisuseFile: sourceText,
	}, correctnessNoCollectionMisuseFile)
}

// correctnessNoCollectionMisuseFinding is one finding as its message id and the text it points at.
type correctnessNoCollectionMisuseFinding struct {
	id   string
	span string
}

// correctnessNoCollectionMisuseExpect asserts the findings in source order, and that none carries a
// fix.
func correctnessNoCollectionMisuseExpect(t *testing.T, result rule_testing.Result, want []correctnessNoCollectionMisuseFinding) {
	t.Helper()
	// A silent case goes through the shared assertion that the rule stays quiet.
	if len(want) == 0 {
		rule_testing.ExpectClean(t, result)
		return
	}
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	sorted := result
	sorted.Diagnostics = diagnostics
	wantIds := make([]string, 0, len(want))
	for _, finding := range want {
		wantIds = append(wantIds, finding.id)
	}
	rule_testing.ExpectFindings(t, sorted, wantIds...)
	text := result.SourceFile.Text()
	for index, diagnostic := range diagnostics {
		got := correctnessNoCollectionMisuseFinding{diagnostic.Message.Id, text[diagnostic.Range.Pos():diagnostic.Range.End()]}
		if got != want[index] {
			t.Fatalf("finding %d is %v, want %v", index, got, want[index])
		}
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
}

func correctnessNoCollectionMisuseIn(span string) correctnessNoCollectionMisuseFinding {
	return correctnessNoCollectionMisuseFinding{correctnessNoCollectionMisuseInOnArrayId, span}
}

func correctnessNoCollectionMisuseSize(span string) correctnessNoCollectionMisuseFinding {
	return correctnessNoCollectionMisuseFinding{correctnessNoCollectionMisuseSizeComparisonId, span}
}

func correctnessNoCollectionMisuseBracket(span string) correctnessNoCollectionMisuseFinding {
	return correctnessNoCollectionMisuseFinding{correctnessNoCollectionMisuseBracketAccessId, span}
}

func correctnessNoCollectionMisuseObject(span string) correctnessNoCollectionMisuseFinding {
	return correctnessNoCollectionMisuseFinding{correctnessNoCollectionMisuseObjectMethodId, span}
}

// The research's shape for `in`, a role looked for in a list of roles, beside its fix.
func TestCorrectnessNoCollectionMisuseInOnArray(t *testing.T) {
	t.Parallel()

	result := correctnessNoCollectionMisuseRun(t, correctnessNoCollectionMisuseSource(
		"if('Admin' in roles) use(1);",
		"if(roles.includes('Admin')) use(1);",
		"use(role in roles);",
		"use(roles.includes(role));",
		"use('third' in pair);",
		"use(('Admin') in (roles));",
	))
	correctnessNoCollectionMisuseExpect(t, result, []correctnessNoCollectionMisuseFinding{
		correctnessNoCollectionMisuseIn("'Admin' in roles"),
		correctnessNoCollectionMisuseIn("role in roles"),
		correctnessNoCollectionMisuseIn("'third' in pair"),
		correctnessNoCollectionMisuseIn("('Admin') in (roles)"),
	})
}

// `in` that asks for a real index or member, a left side that may hold an index, and a right side
// that is not purely an array.
func TestCorrectnessNoCollectionMisuseInStaysSilentOnIndexesAndMembers(t *testing.T) {
	t.Parallel()

	result := correctnessNoCollectionMisuseRun(t, correctnessNoCollectionMisuseSource(
		"use(0 in sparse);",
		"use(index in sparse);",
		"use('0' in roles);",
		"use('1' in pair);",
		"use('length' in roles);",
		"use('map' in roles);",
		"use(wideKey in roles);",
		"use('Admin' in record);",
		"use('extra' in (roles as Role[] & { extra: number }));",
		"use('Admin' in (roles as Role[] | Record<string, number>));",
		"use(Symbol.iterator in roles);",
	))
	correctnessNoCollectionMisuseExpect(t, result, nil)
}

// Every comparison whose answer cannot change, on every sized receiver, either way round.
func TestCorrectnessNoCollectionMisuseImpossibleSizeComparison(t *testing.T) {
	t.Parallel()

	result := correctnessNoCollectionMisuseRun(t, correctnessNoCollectionMisuseSource(
		"if(items.length < 0) use(1);",
		"use(0 > items.length);",
		"use(text.length === -1);",
		"use(bytes.length <= -1);",
		"use(cache.size !== -1);",
		"use(seen.size >= 0);",
		"use(lookup.size > -1);",
		"use(pair.length == -1);",
		"use(-1 != (items.length));",
		"function generic<Values extends string[]>(values: Values) { use(values.length < 0); }",
	))
	correctnessNoCollectionMisuseExpect(t, result, []correctnessNoCollectionMisuseFinding{
		correctnessNoCollectionMisuseSize("items.length < 0"),
		correctnessNoCollectionMisuseSize("0 > items.length"),
		correctnessNoCollectionMisuseSize("text.length === -1"),
		correctnessNoCollectionMisuseSize("bytes.length <= -1"),
		correctnessNoCollectionMisuseSize("cache.size !== -1"),
		correctnessNoCollectionMisuseSize("seen.size >= 0"),
		correctnessNoCollectionMisuseSize("lookup.size > -1"),
		correctnessNoCollectionMisuseSize("pair.length == -1"),
		correctnessNoCollectionMisuseSize("-1 != (items.length)"),
		correctnessNoCollectionMisuseSize("values.length < 0"),
	})
}

// Comparisons that can go either way, a length that may be undefined, and lengths that are not a
// built-in's.
func TestCorrectnessNoCollectionMisuseSizeComparisonStaysSilentWhenItCanChange(t *testing.T) {
	t.Parallel()

	result := correctnessNoCollectionMisuseRun(t, correctnessNoCollectionMisuseSource(
		"use(items.length === 0);",
		"use(items.length > 0);",
		"use(items.length >= 1);",
		"use(items.length <= 0);",
		"use(cache.size < 1);",
		"use(items.length !== count);",
		"use(optionalItems?.length < 0);",
		"use(lengthy.length < 0);",
		"use(arrayLike.length < 0);",
		"use(use.length < 0);",
		"use(count < 0);",
	))
	correctnessNoCollectionMisuseExpect(t, result, nil)
}

// Brackets on a Map or Set with keys that can name no member, and Object's listing methods on one.
func TestCorrectnessNoCollectionMisuseOnMapsAndSets(t *testing.T) {
	t.Parallel()

	result := correctnessNoCollectionMisuseRun(t, correctnessNoCollectionMisuseSource(
		"cache['total'] = 1;",
		"use(cache.get('total'));",
		"use(seen[0]);",
		"use(lookup[index]);",
		"use(Object.keys(cache));",
		"use(Array.from(cache.keys()));",
		"use(Object.entries(lookup));",
		"use(Object.values(seen));",
		"use(Object.getOwnPropertyNames(weak));",
	))
	correctnessNoCollectionMisuseExpect(t, result, []correctnessNoCollectionMisuseFinding{
		correctnessNoCollectionMisuseBracket("cache['total']"),
		correctnessNoCollectionMisuseBracket("seen[0]"),
		correctnessNoCollectionMisuseBracket("lookup[index]"),
		correctnessNoCollectionMisuseObject("Object.keys(cache)"),
		correctnessNoCollectionMisuseObject("Object.entries(lookup)"),
		correctnessNoCollectionMisuseObject("Object.values(seen)"),
		correctnessNoCollectionMisuseObject("Object.getOwnPropertyNames(weak)"),
	})
}

// Keys that name real members or may, plain records, a subclass with fields of its own, a project
// class named Map, and a local `Object`.
func TestCorrectnessNoCollectionMisuseOnMapsAndSetsStaysSilentOnMembersAndLookalikes(t *testing.T) {
	t.Parallel()

	result := correctnessNoCollectionMisuseRun(t, correctnessNoCollectionMisuseSource(
		"use(cache['size']);",
		"use(cache['get']);",
		"use(cache[Symbol.iterator]);",
		"use(cache[wideKey]);",
		"use(record['total']);",
		"use(Object.keys(record));",
		"use(Object.entries(items));",
		"use(Object.freeze(cache));",
		"class Registry extends Map<string, number> { label = 'registry'; }",
		"declare const registry: Registry;",
		"use(Object.keys(registry));",
		"use(registry[0]);",
		"function shadowedMap() {",
		"    class Map { size = -1; [key: string]: unknown; }",
		"    const local = new Map();",
		"    use(local.size < 0);",
		"    use(local['total']);",
		"    use(Object.keys(local));",
		"}",
		"function objectParameter(Object: ObjectConstructor) { use(Object.keys(cache)); }",
		"function shadowedObject() {",
		"    const Object = { keys(value: unknown): string[] { use(value); return []; } };",
		"    use(Object.keys(cache));",
		"}",
	))
	correctnessNoCollectionMisuseExpect(t, result, nil)
}
