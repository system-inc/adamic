package nexus

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const performanceNoIndependentAwaitInLoopFile = "/repository/source/PerformanceNoIndependentAwaitInLoop.ts"

const performanceNoIndependentAwaitInLoopTestdata = "testdata/performance_no_independent_await_in_loop"

// performanceNoIndependentAwaitInLoopPrelude declares every name the synthetic cases use, so the
// checker resolves each one and no verdict rests on an unresolved identifier.
var performanceNoIndependentAwaitInLoopPrelude = strings.Join([]string{
	"declare const items: string[];",
	"declare const groups: string[][];",
	"declare const record: Record<string, string>;",
	"declare const stream: AsyncIterable<string>;",
	"declare const options: { retries: number };",
	"declare const console: { log(...values: unknown[]): void; warn(...values: unknown[]): void };",
	"declare const process: { stdout: { write(text: string): void; clearLine(direction: number): void } };",
	"declare const logger: { info(...values: unknown[]): void };",
	"declare const fileSystem: { appendFile(path: string, text: string): Promise<void> };",
	"declare function fetchItem(item: string): Promise<number>;",
	"declare function fetchWith(item: string, settings: { retries: number }): Promise<number>;",
	"declare function loadItems(): Promise<string[]>;",
	"declare function sleep(milliseconds: number): Promise<void>;",
	"declare function setTimeout(callback: (value?: unknown) => void, milliseconds: number): unknown;",
	"declare function reportProgress(done: number): void;",
	"declare function printSection(item: string): Promise<void>;",
	"declare function openHandle(item: string): { [Symbol.asyncDispose](): Promise<void> };",
	// Callees with bodies, for the one-level scan. Each is named so that the loop-body name list
	// alone cannot decide it, which is what makes the callee scan the thing under test.
	"async function quietOnSuccess(item: string): Promise<void> { try { await fetchItem(item); } catch (error) { console.log('failed', item, error); } }",
	"async function warnsWhenEmpty(item: string): Promise<void> { if (!item) console.error('empty'); await fetchItem(item); }",
	"async function withTimeout(item: string): Promise<void> { setTimeout(() => undefined, 1000); await fetchItem(item); }",
	"async function generateWithProgress(item: string): Promise<void> { console.log('generating ' + item); await fetchItem(item); }",
	"const announce = async (item: string): Promise<void> => { process.stdout.write(item); await fetchItem(item); };",
	"async function pacedSend(item: string): Promise<void> { await sleep(100); await fetchItem(item); }",
	"",
}, "\n")

// performanceNoIndependentAwaitInLoopSource wraps body lines in an async function after the prelude.
func performanceNoIndependentAwaitInLoopSource(bodyLines ...string) string {
	return performanceNoIndependentAwaitInLoopPrelude +
		"export async function subject(): Promise<unknown> {\n" +
		strings.Join(bodyLines, "\n") + "\n" +
		"    return undefined;\n" +
		"}\n"
}

// The real ahra shapes the rule was written against, one file each, with the finding count each
// must produce. The directory is read and compared against this table both ways, so a fixture added
// without a verdict, or a verdict whose file went missing, fails rather than going unread.
var performanceNoIndependentAwaitInLoopAhraShapes = map[string]int{
	"fires/BackupStaleness.ts":            2,
	"fires/AhraOsCommandLineInterface.ts": 1,
	"fires/FinanceConnections.ts":         1,
	"silent/StripeApi.ts":                 0,
	"silent/OpenAiAdsRegistry.ts":         0,
	"silent/GeminiApi.ts":                 0,
	"silent/FrameTvApi.ts":                0,
	"silent/KlingApi.ts":                  0,
	"silent/AsanaApi.ts":                  0,
	"silent/DataRebuild.ts":               0,
	"silent/PromiseGroupConsumer.ts":      0,
	"silent/FacetsMonthlies.ts":           0,
	"silent/AhraOsReportPdf.ts":           0,
}

func TestPerformanceNoIndependentAwaitInLoopAhraShapes(t *testing.T) {
	t.Parallel()

	var onDisk []string
	for _, directory := range []string{"fires", "silent"} {
		entries, err := os.ReadDir(filepath.Join(performanceNoIndependentAwaitInLoopTestdata, directory))
		if err != nil {
			t.Fatalf("reading %s: %v", directory, err)
		}
		for _, entry := range entries {
			onDisk = append(onDisk, directory+"/"+entry.Name())
		}
	}
	var inTable []string
	for name := range performanceNoIndependentAwaitInLoopAhraShapes {
		inTable = append(inTable, name)
	}
	sort.Strings(onDisk)
	sort.Strings(inTable)
	if strings.Join(onDisk, ",") != strings.Join(inTable, ",") {
		t.Fatalf("testdata and verdict table disagree:\n  on disk:  %v\n  in table: %v", onDisk, inTable)
	}

	for name, wantCount := range performanceNoIndependentAwaitInLoopAhraShapes {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join(performanceNoIndependentAwaitInLoopTestdata, name))
			if err != nil {
				t.Fatal(err)
			}
			result := rule_testing.RunTyped(t, PerformanceNoIndependentAwaitInLoop, "/repository/source/"+filepath.Base(name), string(source))
			if wantCount == 0 {
				rule_testing.ExpectClean(t, result)
				return
			}
			want := make([]string, wantCount)
			for index := range want {
				want[index] = "independentAwaitInLoop"
			}
			rule_testing.ExpectFindings(t, result, want...)
			for _, diagnostic := range result.Diagnostics {
				reported := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
				if !strings.HasPrefix(reported, "for(const ") || !strings.HasSuffix(reported, ")") {
					t.Errorf("finding should span the loop header, got %q", reported)
				}
			}
		})
	}
}

func TestPerformanceNoIndependentAwaitInLoopFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"a for-of pushing each awaited result", []string{
			"    const results: number[] = [];",
			"    for (const item of items) { results.push(await fetchItem(item)); }",
		}},
		{"the indexed form with i++", []string{
			"    for (let index = 0; index < items.length; index++) { await fetchItem(items[index]); }",
		}},
		{"the indexed form with ++i", []string{
			"    for (let index = 0; index < items.length; ++index) { await fetchItem(items[index]); }",
		}},
		{"the indexed form with i += 1, reading the index as a value too", []string{
			"    for (let index = 0; index < items.length; index += 1) { await fetchWith(items[index], { retries: index }); }",
		}},
		{"an element write into an outer record", []string{
			"    const summary: Record<string, number> = {};",
			"    for (const item of items) { summary[item] = await fetchItem(item); }",
		}},
		{"a Map set", []string{
			"    const values = new Map<string, number>();",
			"    for (const item of items) { values.set(item, await fetchItem(item)); }",
		}},
		{"a loop-invariant outer value read by the await", []string{
			"    for (const item of items) { await fetchWith(item, options); }",
		}},
		{"a validation throw that reads no awaited value", []string{
			"    for (const item of items) { if (!item) throw new Error('empty'); await fetchItem(item); }",
		}},
		{"a continue on an awaited value, which only skips this item", []string{
			"    const results: number[] = [];",
			"    for (const item of items) { const value = await fetchItem(item); if (value < 0) continue; results.push(value); }",
		}},
		{"a catch that records and continues", []string{
			"    const results: number[] = [];",
			"    const failures: string[] = [];",
			"    for (const item of items) {",
			"        try { results.push(await fetchItem(item)); }",
			"        catch (error) { failures.push(String(error)); continue; }",
			"    }",
		}},
		{"a break that leaves only an inner switch", []string{
			"    const results: string[] = [];",
			"    for (const item of items) {",
			"        switch (await fetchItem(item)) { case 1: break; default: results.push(item); }",
			"    }",
		}},
		{"a throw caught inside the body", []string{
			"    const failures: string[] = [];",
			"    for (const item of items) {",
			"        try { if ((await fetchItem(item)) < 0) throw new Error(item); }",
			"        catch (error) { failures.push(String(error)); }",
			"    }",
		}},
		{"a body-local let reassigned within the iteration", []string{
			"    const results: number[] = [];",
			"    for (const item of items) { let value = await fetchItem(item); value += 1; results.push(value); }",
		}},
		{"a property of the loop's own item written from the await", []string{
			"    const rows = items.map((name) => ({ name, value: 0 }));",
			"    for (const row of rows) { row.value = await fetchItem(row.name); }",
		}},
		{"an awaited Promise.all inside the body is still one wait per item", []string{
			"    for (const item of items) { await Promise.all(item.split(',').map((part) => fetchItem(part))); }",
		}},
		{"an await on a nested loop's iterable belongs to the outer loop", []string{
			"    const results: string[] = [];",
			"    for (const item of items) { for (const part of await loadItems()) { results.push(item + part); } }",
		}},
		{"an await on a value computed from the item, not the item itself", []string{
			"    const results: number[] = [];",
			"    for (const item of items) { results.push(await fetchItem(item)); }",
		}},
		{"a called function that reports failure only from its catch", []string{
			"    for (const item of items) { await quietOnSuccess(item); }",
		}},
		{"a called function whose only console call is an error report", []string{
			"    for (const item of items) { await warnsWhenEmpty(item); }",
		}},
		{"a called function whose setTimeout is a request timeout", []string{
			"    for (const item of items) { await withTimeout(item); }",
		}},
		{"a catch in the body that warns about the one item that failed", []string{
			"    const results: number[] = [];",
			"    for (const item of items) { try { results.push(await fetchItem(item)); } catch (error) { console.warn('failed', item, error); } }",
		}},
		{"a sink by prefix whose receiver the body never reads back", []string{
			"    const document = { pages: [] as string[], addPage(name: string): void { this.pages.push(name); } };",
			"    for (const item of items) { const value = await fetchItem(item); document.addPage(item + value); }",
		}},
		{"a nested for await waits inside this loop's iteration", []string{
			"    const results: string[] = [];",
			"    for (const item of items) { for await (const line of stream) { results.push(item + line); } }",
		}},
		{"an await using disposes once per iteration", []string{
			"    for (const item of items) { await using handle = openHandle(item); }",
		}},
		{"a cache on this filled while a different member of this is awaited", []string{
			"    const service = {",
			"        cache: new Map<string, number>(),",
			"        fetch(item: string): Promise<number> { return fetchItem(item); },",
			"        async warm(): Promise<void> { for (const item of items) { this.cache.set(item, await this.fetch(item)); } },",
			"    };",
			"    await service.warm();",
		}},
		{"a method named like a sink prefix but lower-case after it", []string{
			"    const book = { label: 'x', address(name: string): string { return name; } };",
			"    for (const item of items) { await fetchItem(book.address(item) + book.label); }",
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PerformanceNoIndependentAwaitInLoop, performanceNoIndependentAwaitInLoopFile,
				performanceNoIndependentAwaitInLoopSource(testCase.lines...))
			rule_testing.ExpectFindings(t, result, "independentAwaitInLoop")
		})
	}
}

func TestPerformanceNoIndependentAwaitInLoopStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		// The loop is not a walk over a collection.
		{"a while", []string{
			"    while (items.length > 0) { await fetchItem(items[0]); }",
		}},
		{"a do-while", []string{
			"    do { await fetchItem(items[0]); } while (items.length > 0);",
		}},
		{"a for await", []string{
			"    for await (const item of stream) { await fetchItem(item); }",
		}},
		{"a for-in", []string{
			"    for (const key in record) { await fetchItem(key); }",
		}},
		{"a for bounded by an attempt count", []string{
			"    for (let attempt = 0; attempt < 3; attempt++) { await fetchItem('probe'); }",
		}},
		{"a for with no header", []string{
			"    for (;;) { await fetchItem('probe'); }",
		}},
		{"an indexed for bounded with <=", []string{
			"    for (let index = 0; index <= items.length; index++) { await fetchItem(items[index] ?? ''); }",
		}},
		{"a for bounded by a property that is not a length", []string{
			"    for (let attempt = 0; attempt < options.retries; attempt++) { await fetchItem('probe'); }",
		}},
		{"an indexed for stepping by two", []string{
			"    for (let index = 0; index < items.length; index += 2) { await fetchItem(items[index]); }",
		}},

		// The loop does not itself wait.
		{"an await only on the for-of iterable", []string{
			"    const results: string[] = [];",
			"    for (const item of await loadItems()) { results.push(item); }",
		}},
		{"an await only inside a callback written in the body", []string{
			"    const tasks: Array<() => Promise<void>> = [];",
			"    for (const item of items) { tasks.push(async () => { await fetchItem(item); }); }",
		}},
		{"an await only inside a nested while, which does not report either", []string{
			"    for (const group of groups) { while (group.length > 0) { await fetchItem(group[0] ?? ''); } }",
		}},

		// Loop-carried state.
		{"the indexed form writing its index in the body", []string{
			"    for (let index = 0; index < items.length; index++) { await fetchItem(items[index]); index++; }",
		}},
		{"an outer variable reassigned from the await", []string{
			"    let cursor = '';",
			"    for (const item of items) { cursor = String(await fetchItem(item + cursor)); }",
		}},
		{"an outer counter incremented", []string{
			"    let count = 0;",
			"    for (const item of items) { await fetchItem(item); count++; }",
		}},
		{"an outer accumulator added to", []string{
			"    let total = 0;",
			"    for (const item of items) { total += await fetchItem(item); }",
		}},
		{"an outer property incremented", []string{
			"    const summary = { count: 0 };",
			"    for (const item of items) { await fetchItem(item); summary.count++; }",
		}},
		{"an outer property added to", []string{
			"    const summary = { total: 0 };",
			"    for (const item of items) { summary.total += await fetchItem(item); }",
		}},
		{"an outer binding written by destructuring", []string{
			"    let first = 0;",
			"    for (const item of items) { [first] = [await fetchItem(item)]; }",
		}},
		{"a var declared in the body", []string{
			"    for (const item of items) { var previous = await fetchItem(item); }",
		}},
		{"a collection read back in the body", []string{
			"    const results: number[] = [];",
			"    for (const item of items) { results.push(await fetchItem(item + results.length)); }",
		}},
		{"an Object.assign target read back", []string{
			"    const merged: Record<string, number> = {};",
			"    for (const item of items) { Object.assign(merged, { [item]: await fetchWith(item, { retries: Object.keys(merged).length }) }); }",
		}},
		{"a set consulted before it is filled", []string{
			"    const seen = new Set<string>();",
			"    for (const item of items) { if (seen.has(item)) continue; seen.add(item); await fetchItem(item); }",
		}},
		{"a queue that grows while it is walked", []string{
			"    const queue = [...items];",
			"    for (const item of queue) { const value = await fetchItem(item); if (value > 0) queue.push(item + value); }",
		}},
		{"a stack consumed by the body", []string{
			"    const stack = [...items];",
			"    for (const item of items) { await fetchItem(stack.pop() ?? item); }",
		}},
		{"an outer property written and read back as a cursor", []string{
			"    const cursor = { after: '' };",
			"    for (const item of items) { const value = await fetchItem(item + cursor.after); cursor.after = String(value); }",
		}},

		// An exit on an awaited value.
		{"a return on an awaited value, which is find-first", []string{
			"    for (const item of items) { if ((await fetchItem(item)) > 0) return item; }",
		}},
		{"a break on a value derived from the await", []string{
			"    for (const item of items) { const value = await fetchItem(item); const done = value > 9; if (done) break; }",
		}},
		{"a throw on a value derived from the await", []string{
			"    for (const item of items) { const value = await fetchItem(item); if (value < 0) throw new Error(item); }",
		}},
		{"a return from a bare catch whose try awaits", []string{
			"    for (const item of items) { try { await fetchItem(item); } catch { return; } }",
		}},
		{"an unconditional return of the awaited value", []string{
			"    for (const item of items) { return await fetchItem(item); }",
		}},
		{"an unconditional throw built from an awaited value", []string{
			"    for (const item of items) { const value = await fetchItem(item); throw new Error(String(value)); }",
		}},
		{"a return inside a nested while whose test reads an awaited value", []string{
			"    for (const item of items) { const value = await fetchItem(item); while (value > 0) { return item; } }",
		}},
		{"a return under a case whose label reads an awaited value", []string{
			"    for (const item of items) { const value = await fetchItem(item); switch (item) { case String(value): return item; } }",
		}},
		{"a flag set from the caught error and read after the try", []string{
			"    for (const item of items) {",
			"        let failed = false;",
			"        try { await fetchItem(item); } catch (error) { failed = error instanceof Error; }",
			"        if (failed) return;",
			"    }",
		}},
		{"an exit on a value an inner loop carries backward from an await", []string{
			"    for (const item of items) {",
			"        await fetchItem(item);",
			"        let previous = 0;",
			"        let latest = 0;",
			"        for (const part of groups[0] ?? []) { previous = latest; latest = await fetchItem(part); }",
			"        if (previous > 9) return;",
			"    }",
		}},
		{"a rethrow from a catch whose try awaits", []string{
			"    for (const item of items) { try { await fetchItem(item); } catch (error) { throw error; } }",
		}},
		{"a return inside a try that awaits first and swallows failure", []string{
			"    for (const item of items) { try { await fetchItem(item); return item; } catch { } }",
		}},
		{"a labeled continue to an outer loop on an awaited value", []string{
			"    outer: for (const group of groups) {",
			"        for (const item of group) { if ((await fetchItem(item)) > 0) continue outer; }",
			"    }",
		}},
		{"an unconditional break reached only past an awaited continue", []string{
			"    for (const item of items) { const value = await fetchItem(item); if (value < 0) { continue; } break; }",
		}},
		{"a break inside a switch on an awaited value that names this loop", []string{
			"    scan: for (const item of items) {",
			"        switch (await fetchItem(item)) { case 1: break scan; default: break; }",
			"    }",
		}},

		// An ordered side effect.
		{"console output", []string{
			"    for (const item of items) { console.log(item); await fetchItem(item); }",
		}},
		{"a write to stdout", []string{
			"    for (const item of items) { process.stdout.write(item); await fetchItem(item); }",
		}},
		{"a sleep between items", []string{
			"    for (const item of items) { await sleep(250); await fetchItem(item); }",
		}},
		{"a setTimeout delay", []string{
			"    for (const item of items) { await new Promise((resolve) => setTimeout(resolve, 100)); await fetchItem(item); }",
		}},
		{"a progress callback", []string{
			"    for (const item of items) { await fetchItem(item); reportProgress(1); }",
		}},
		{"a logger call", []string{
			"    for (const item of items) { logger.info(item); await fetchItem(item); }",
		}},
		{"an append to one file", []string{
			"    for (const item of items) { await fileSystem.appendFile('/tmp/out.txt', String(await fetchItem(item))); }",
		}},
		{"console output inside a callback the iteration runs", []string{
			"    for (const item of items) { const value = await fetchItem(item); [value].forEach((each) => console.log(each)); }",
		}},
		{"a stdout cursor move", []string{
			"    for (const item of items) { process.stdout.clearLine(0); await fetchItem(item); }",
		}},
		{"progress output from a catch in the body", []string{
			"    const results: number[] = [];",
			"    for (const item of items) { try { results.push(await fetchItem(item)); } catch (error) { console.log('skipped', item); } }",
		}},
		{"a print call", []string{
			"    for (const item of items) { await printSection(item); }",
		}},

		// The loop awaits work started before it began.
		{"awaiting the for-of item itself", []string{
			"    const pending = items.map((item) => fetchItem(item));",
			"    const results: number[] = [];",
			"    for (const promise of pending) { results.push(await promise); }",
		}},
		{"awaiting the indexed element itself", []string{
			"    const pending = items.map((item) => fetchItem(item));",
			"    const results: number[] = [];",
			"    for (let index = 0; index < pending.length; index++) { results.push(await pending[index]); }",
		}},

		// The ordered effect lives one call down.
		{"a called function that prints progress", []string{
			"    for (const item of items) { await generateWithProgress(item); }",
		}},
		{"a called arrow function that prints progress", []string{
			"    for (const item of items) { await announce(item); }",
		}},
		{"a called function that paces itself", []string{
			"    for (const item of items) { await pacedSend(item); }",
		}},

		// A collection-building prefix makes the receiver a sink.
		{"a shared document built page by page and handed to the await", []string{
			"    const document = { pages: [] as string[], addPage(name: string): void { this.pages.push(name); } };",
			"    for (const item of items) { document.addPage(item); await fetchWith(item, { retries: document.pages.length }); }",
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, PerformanceNoIndependentAwaitInLoop, performanceNoIndependentAwaitInLoopFile,
				performanceNoIndependentAwaitInLoopSource(testCase.lines...))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// A yield hands results out in order, and it needs a generator rather than the shared wrapper.
func TestPerformanceNoIndependentAwaitInLoopStaysSilentOnYield(t *testing.T) {
	t.Parallel()

	source := performanceNoIndependentAwaitInLoopPrelude +
		"export async function* subject(): AsyncGenerator<number> {\n" +
		"    for (const item of items) { yield await fetchItem(item); }\n" +
		"}\n"
	rule_testing.ExpectClean(t, rule_testing.RunTyped(t, PerformanceNoIndependentAwaitInLoop, performanceNoIndependentAwaitInLoopFile, source))

	// Control: the same loop collecting rather than yielding reports, so the silence above is the yield.
	control := performanceNoIndependentAwaitInLoopPrelude +
		"export async function* subject(): AsyncGenerator<number> {\n" +
		"    const results: number[] = [];\n" +
		"    for (const item of items) { results.push(await fetchItem(item)); }\n" +
		"    yield* results;\n" +
		"}\n"
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PerformanceNoIndependentAwaitInLoop, performanceNoIndependentAwaitInLoopFile, control),
		"independentAwaitInLoop")
}

// The callee scan follows an import to the function's source through the checker's alias
// resolution. The control imports a function from the same module that does not print, so the
// silence is the printing rather than the import.
func TestPerformanceNoIndependentAwaitInLoopFollowsAnImportedCallee(t *testing.T) {
	t.Parallel()

	printer := strings.Join([]string{
		"declare const console: { log(...values: unknown[]): void };",
		"export async function renderRow(item: string): Promise<void> { console.log(item); }",
		"export async function storeRow(item: string): Promise<number> { return item.length; }",
		"",
	}, "\n")
	subject := func(callee string) string {
		return strings.Join([]string{
			"import { " + callee + " } from './Printer';",
			"declare const items: string[];",
			"export async function subject(): Promise<void> {",
			"    for (const item of items) { await " + callee + "(item); }",
			"}",
			"",
		}, "\n")
	}

	silent := rule_testing.RunTypedFiles(t, PerformanceNoIndependentAwaitInLoop, map[string]string{
		"/repository/source/Printer.ts": printer,
		"/repository/source/Subject.ts": subject("renderRow"),
	}, "/repository/source/Subject.ts")
	rule_testing.ExpectClean(t, silent)

	control := rule_testing.RunTypedFiles(t, PerformanceNoIndependentAwaitInLoop, map[string]string{
		"/repository/source/Printer.ts": printer,
		"/repository/source/Subject.ts": subject("storeRow"),
	}, "/repository/source/Subject.ts")
	rule_testing.ExpectFindings(t, control, "independentAwaitInLoop")
}

// One finding per loop however many awaits it holds, and an await in a nested loop is reported at
// the nested loop rather than at the one around it.
func TestPerformanceNoIndependentAwaitInLoopReportsTheInnermostLoopOnce(t *testing.T) {
	t.Parallel()

	twoAwaits := performanceNoIndependentAwaitInLoopSource(
		"    const results: number[] = [];",
		"    for (const item of items) { results.push(await fetchItem(item)); results.push(await fetchItem(item + item)); }",
	)
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PerformanceNoIndependentAwaitInLoop, performanceNoIndependentAwaitInLoopFile, twoAwaits),
		"independentAwaitInLoop")

	nested := performanceNoIndependentAwaitInLoopSource(
		"    const results: number[] = [];",
		"    for (const group of groups) {",
		"        for (const item of group) { results.push(await fetchItem(item)); }",
		"    }",
	)
	result := rule_testing.RunTyped(t, PerformanceNoIndependentAwaitInLoop, performanceNoIndependentAwaitInLoopFile, nested)
	rule_testing.ExpectFindings(t, result, "independentAwaitInLoop")
	reported := result.SourceFile.Text()[result.Diagnostics[0].Range.Pos():result.Diagnostics[0].Range.End()]
	if reported != "for (const item of group)" {
		t.Fatalf("expected the inner loop's header, got %q", reported)
	}
}

// The span is the header, so a suppression goes above the loop, and the message is asserted whole.
func TestPerformanceNoIndependentAwaitInLoopSpanAndMessage(t *testing.T) {
	t.Parallel()

	source := performanceNoIndependentAwaitInLoopSource(
		"    for (let index = 0; index < items.length; index++)",
		"    {",
		"        await fetchItem(items[index]);",
		"    }",
	)
	result := rule_testing.RunTyped(t, PerformanceNoIndependentAwaitInLoop, performanceNoIndependentAwaitInLoopFile, source)
	rule_testing.ExpectFindings(t, result, "independentAwaitInLoop")
	diagnostic := result.Diagnostics[0]
	reported := result.SourceFile.Text()[diagnostic.Range.Pos():diagnostic.Range.End()]
	if reported != "for (let index = 0; index < items.length; index++)" {
		t.Fatalf("span: got %q", reported)
	}
	if diagnostic.Message.Description != messagePerformanceNoIndependentAwaitInLoop().Description {
		t.Fatalf("message: got %q", diagnostic.Message.Description)
	}
	for _, phrase := range []string{"Promise.all", "bounded", "cohere-disable-next-line nexus/performance-no-independent-await-in-loop -- "} {
		if !strings.Contains(diagnostic.Message.Description, phrase) {
			t.Errorf("message should mention %q", phrase)
		}
	}
	if strings.ContainsRune(diagnostic.Message.Description, '\u2014') {
		t.Errorf("message carries an em-dash")
	}
}

// The rule declares the checker and declines a file without one, rather than reporting from an
// unbound tree. Pinned so a later revert of the guard fails loudly rather than going vacuously
// green; the control is the same source reporting under the typed harness.
func TestPerformanceNoIndependentAwaitInLoopDeclinesWithoutAChecker(t *testing.T) {
	t.Parallel()

	if !PerformanceNoIndependentAwaitInLoop.NeedsTypeChecker {
		t.Fatal("the rule resolves bindings through the checker and must declare it")
	}
	source := performanceNoIndependentAwaitInLoopSource(
		"    for (const item of items) { await fetchItem(item); }",
	)
	rule_testing.ExpectClean(t, rule_testing.Run(t, PerformanceNoIndependentAwaitInLoop, performanceNoIndependentAwaitInLoopFile, source))
	rule_testing.ExpectFindings(t, rule_testing.RunTyped(t, PerformanceNoIndependentAwaitInLoop, performanceNoIndependentAwaitInLoopFile, source),
		"independentAwaitInLoop")
}
