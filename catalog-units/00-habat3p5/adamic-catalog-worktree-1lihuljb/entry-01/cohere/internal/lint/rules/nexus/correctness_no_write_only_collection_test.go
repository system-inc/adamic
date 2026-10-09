package nexus

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoWriteOnlyCollectionFile = "/repository/source/CorrectnessNoWriteOnlyCollection.ts"

var correctnessNoWriteOnlyCollectionPrelude = strings.Join([]string{
	"interface RowInterface { rowid: number; id: string }",
	"declare const rows: RowInterface[];",
	"declare function use(value: unknown): void;",
	"",
}, "\n")

func correctnessNoWriteOnlyCollectionSource(lines ...string) string {
	return correctnessNoWriteOnlyCollectionPrelude + strings.Join(lines, "\n") + "\n"
}

func correctnessNoWriteOnlyCollectionRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessNoWriteOnlyCollection, map[string]string{
		correctnessNoWriteOnlyCollectionFile: sourceText,
	}, correctnessNoWriteOnlyCollectionFile)
}

// correctnessNoWriteOnlyCollectionExpect asserts the reported names in source order, and that no
// finding carries a fix.
func correctnessNoWriteOnlyCollectionExpect(t *testing.T, result rule_testing.Result, want ...string) {
	t.Helper()
	// The shared assertion by message id, which proves the rule can fire; the checks below hold
	// each finding's exact text.
	if len(want) > 0 {
		wantIds := make([]string, len(want))
		for index := range wantIds {
			wantIds[index] = correctnessNoWriteOnlyCollectionId
		}
		rule_testing.ExpectFindings(t, result, wantIds...)
	}
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	text := result.SourceFile.Text()
	var got []string
	for _, diagnostic := range diagnostics {
		if diagnostic.Message.Id != correctnessNoWriteOnlyCollectionId {
			t.Fatalf("unexpected message id %q", diagnostic.Message.Id)
		}
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
		got = append(got, text[diagnostic.Range.Pos():diagnostic.Range.End()])
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d findings %q, got %d %q", len(want), want, len(got), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("finding %d is %q, want %q", index, got[index], want[index])
		}
	}
}

// correctnessNoWriteOnlyCollectionHybridSearch models the rank fusion in
// `AhraOsProfileCommitsStore.searchProfileCommitsHybrid` in ahra on 2026-10-03, trimmed to the maps.
// `fixed` deletes the two maps nothing reads, which is the whole fix: the ranks the result carries
// already come from `hitData`.
func correctnessNoWriteOnlyCollectionHybridSearch(fixed bool) string {
	keep := func(line string) string {
		if fixed {
			return ""
		}
		return line
	}
	return correctnessNoWriteOnlyCollectionSource(
		"interface HitInterface { rowid: number; rrf_score: number; bm25_rank: number | null; vec_rank: number | null }",
		"export function searchHybrid(bmRows: { rowid: number }[], vecRows: { rowid: number }[], limit: number): HitInterface[] {",
		"    const rankFusionConstant = 60;",
		"    const scores = new Map<number, number>();",
		keep("    const bmRanks = new Map<number, number>();"),
		keep("    const vecRanks = new Map<number, number>();"),
		"    const hitData = new Map<number, HitInterface>();",
		"    bmRows.forEach(function(row, index) {",
		"        const rank = index;",
		"        scores.set(row.rowid, (scores.get(row.rowid) ?? 0) + 1 / (rankFusionConstant + rank + 1));",
		keep("        bmRanks.set(row.rowid, rank);"),
		"        hitData.set(row.rowid, { rowid: row.rowid, rrf_score: 0, bm25_rank: rank, vec_rank: null });",
		"    });",
		"    vecRows.forEach(function(row, index) {",
		"        const rank = index;",
		"        scores.set(row.rowid, (scores.get(row.rowid) ?? 0) + 1 / (rankFusionConstant + rank + 1));",
		keep("        vecRanks.set(row.rowid, rank);"),
		"        const existing = hitData.get(row.rowid);",
		"        if(existing) {",
		"            existing.vec_rank = rank;",
		"            return;",
		"        }",
		"        hitData.set(row.rowid, { rowid: row.rowid, rrf_score: 0, bm25_rank: null, vec_rank: rank });",
		"    });",
		"    const ranked: HitInterface[] = [];",
		"    for(const [rowid, score] of scores) {",
		"        const hit = hitData.get(rowid);",
		"        if(!hit) continue;",
		"        hit.rrf_score = score;",
		"        ranked.push(hit);",
		"    }",
		"    ranked.sort((left, right) => right.rrf_score - left.rrf_score);",
		"    return ranked.slice(0, limit);",
		"}",
	)
}

// The real rank fusion: the two rank maps are reported, and the three collections that are read are not.
func TestCorrectnessNoWriteOnlyCollectionFiresOnProfileCommitsStore(t *testing.T) {
	t.Parallel()

	result := correctnessNoWriteOnlyCollectionRun(t, correctnessNoWriteOnlyCollectionHybridSearch(false))
	correctnessNoWriteOnlyCollectionExpect(t, result, "bmRanks", "vecRanks")
}

func TestCorrectnessNoWriteOnlyCollectionStaysSilentOnFixedProfileCommitsStore(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoWriteOnlyCollectionRun(t, correctnessNoWriteOnlyCollectionHybridSearch(true)))
}

func TestCorrectnessNoWriteOnlyCollectionFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"an array pushed in a loop", []string{
			"export function collect() {",
			"    const seen: string[] = [];",
			"    for(const row of rows) {",
			"        seen.push(row.id);",
			"    }",
			"}",
		}, []string{"seen"}},
		{"a Set added to", []string{
			"export function collect() {",
			"    const seen = new Set<string>();",
			"    for(const row of rows) seen.add(row.id);",
			"    return rows.length;",
			"}",
		}, []string{"seen"}},
		{"an array written through elements", []string{
			"export function collect() {",
			"    const byIndex = [] as string[];",
			"    rows.forEach(function(row, index) { byIndex[index] = row.id; });",
			"}",
		}, []string{"byIndex"}},
		{"a Map set, deleted and cleared", []string{
			"export function collect() {",
			"    const pending = new Map<number, string>();",
			"    for(const row of rows) {",
			"        pending.set(row.rowid, row.id);",
			"        pending.delete(row.rowid - 1);",
			"    }",
			"    pending.clear();",
			"}",
		}, []string{"pending"}},
		{"an array spliced, shifted and popped", []string{
			"export function collect() {",
			"    const queue = ['a', 'b', 'c'];",
			"    queue.unshift('z');",
			"    queue.splice(1, 1);",
			"    queue.shift();",
			"    queue.pop();",
			"}",
		}, []string{"queue"}},
		{"in a property initializer's arrow", []string{
			"export class Collector {",
			"    collect = () => {",
			"        const seen = new Set<string>();",
			"        rows.forEach(function(row) { seen.add(row.id); });",
			"    };",
			"}",
		}, []string{"seen"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := correctnessNoWriteOnlyCollectionRun(t, correctnessNoWriteOnlyCollectionSource(testCase.lines...))
			correctnessNoWriteOnlyCollectionExpect(t, result, testCase.want...)
		})
	}
}

func TestCorrectnessNoWriteOnlyCollectionStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"a Set that guards with has", []string{
			"export function unique() {",
			"    const seen = new Set<string>();",
			"    for(const row of rows) {",
			"        if(seen.has(row.id)) continue;",
			"        seen.add(row.id);",
			"        use(row);",
			"    }",
			"}",
		}},
		{"returned", []string{
			"export function collect() {",
			"    const ids: string[] = [];",
			"    for(const row of rows) ids.push(row.id);",
			"    return ids;",
			"}",
		}},
		{"stored by a shorthand property", []string{
			"export function collect() {",
			"    const ids: string[] = [];",
			"    for(const row of rows) ids.push(row.id);",
			"    return { ids };",
			"}",
		}},
		{"its length read", []string{
			"export function count() {",
			"    const ids: string[] = [];",
			"    for(const row of rows) ids.push(row.id);",
			"    return ids.length;",
			"}",
		}},
		{"spread", []string{
			"export function collect() {",
			"    const ids: string[] = [];",
			"    ids.push('a');",
			"    use([...ids]);",
			"}",
		}},
		{"a mutating call whose result is used", []string{
			"export function collect() {",
			"    const ids: string[] = [];",
			"    return ids.push('a');",
			"}",
		}},
		{"a write in an arrow's expression body, declined", []string{
			"export function collect() {",
			"    const ids: string[] = [];",
			"    rows.forEach((row) => ids.push(row.id));",
			"}",
		}},
		{"a compound element write, which reads", []string{
			"export function count() {",
			"    const counts: number[] = [0];",
			"    counts[0] += 1;",
			"}",
		}},
		{"a module-level collection, declined", []string{
			"const registry = new Map<string, number>();",
			"export function register(name: string) {",
			"    registry.set(name, 1);",
			"}",
		}},
		{"a let", []string{
			"export function collect() {",
			"    let ids: string[] = [];",
			"    ids.push('a');",
			"}",
		}},
		{"never referenced, which no-unused-vars owns", []string{
			"export function collect() {",
			"    const ids: string[] = [];",
			"}",
		}},
		{"a local Map class that is not the library's", []string{
			"class Map<K, V> { set(key: K, value: V) { use([key, value]); } }",
			"export function collect() {",
			"    const ranks = new Map<number, number>();",
			"    ranks.set(1, 2);",
			"}",
		}},
		{"a Map subclass that registers itself, constructed under the library's name", []string{
			"const everyRegistry: unknown[] = [];",
			"class Map<K, V> extends globalThis.Map<K, V> { constructor() { super(); everyRegistry.push(this); } }",
			"export function collect() {",
			"    const ranks = new Map<number, number>();",
			"    ranks.set(1, 2);",
			"    return everyRegistry.length;",
			"}",
		}},
		{"a typed recorder whose push is not Array's", []string{
			"interface RecorderInterface { push(value: string): void }",
			"export function collect() {",
			"    const recorder: RecorderInterface = [];",
			"    recorder.push('a');",
			"}",
		}},
		{"iterated", []string{
			"export function collect() {",
			"    const ids: string[] = [];",
			"    ids.push('a');",
			"    for(const id of ids) use(id);",
			"}",
		}},
		{"read in a closure", []string{
			"export function collect() {",
			"    const ids: string[] = [];",
			"    ids.push('a');",
			"    return () => ids.join(',');",
			"}",
		}},
		{"a shadowing binding read, the outer only written", []string{
			"export function collect() {",
			"    const ids: string[] = [];",
			"    ids.push('a');",
			"    {",
			"        const ids = ['b'];",
			"        use(ids);",
			"    }",
			"    use(ids.length);",
			"}",
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rule_testing.ExpectClean(t, correctnessNoWriteOnlyCollectionRun(t, correctnessNoWriteOnlyCollectionSource(testCase.lines...)))
		})
	}
}

// Symbol identity, not the name: the outer collection is only written, and a shadowing inner binding
// of the same name is read. The outer is reported and the inner is not.
func TestCorrectnessNoWriteOnlyCollectionResolvesShadowing(t *testing.T) {
	t.Parallel()

	result := correctnessNoWriteOnlyCollectionRun(t, correctnessNoWriteOnlyCollectionSource(
		"export function collect() {",
		"    const ids: string[] = [];",
		"    ids.push('a');",
		"    {",
		"        const ids = ['b'];",
		"        ids.push('c');",
		"        use(ids);",
		"    }",
		"}",
	))
	correctnessNoWriteOnlyCollectionExpect(t, result, "ids")
	outer := strings.Index(result.SourceFile.Text(), "const ids: string[]") + len("const ")
	if position := result.Diagnostics[0].Range.Pos(); position != outer {
		t.Fatalf("reported at %d, want the outer binding at %d", position, outer)
	}
}
