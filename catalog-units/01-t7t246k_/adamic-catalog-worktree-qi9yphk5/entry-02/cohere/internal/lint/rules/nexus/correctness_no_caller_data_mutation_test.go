package nexus

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoCallerDataMutationFile = "/repository/source/CorrectnessNoCallerDataMutation.ts"

// correctnessNoCallerDataMutationHostFile stands in for a host's types: a declaration file, which is
// where the DOM's, a package's and `@types/*` declarations live. The fixture program's lib is ES2022,
// so the DOM itself is not there to write to.
const correctnessNoCallerDataMutationHostFile = "/repository/source/Host.d.ts"

var correctnessNoCallerDataMutationHost = strings.Join([]string{
	"interface CanvasContextInterface { fillStyle: string; lineWidth: number }",
	"interface PortInterface { onmessage: ((data: unknown) => void) | null }",
	"interface HostElementInterface { style: { color: string }; dataset: Record<string, string>; classNames: string[] }",
	"",
}, "\n")

var correctnessNoCallerDataMutationPrelude = strings.Join([]string{
	"interface UsageInterface { inputTokens: number; outputTokens: number; dollars: number }",
	"interface DiffInterface { path: string }",
	"interface MessageInterface { createdAt: number; body: string }",
	"interface EntryInterface { status: string; position: number; child?: EntryInterface }",
	"interface HubInterface { watchersByPath: Map<string, number>; seen: Set<string>; debounceTimer: number | null }",
	"interface RequestInterface { headers: Record<string, string> }",
	"interface PanelInterface { canvas: CanvasContextInterface; label: string }",
	"declare function use(value: unknown): void;",
	"",
}, "\n")

func correctnessNoCallerDataMutationSource(lines ...string) string {
	return correctnessNoCallerDataMutationPrelude + strings.Join(lines, "\n") + "\n"
}

func correctnessNoCallerDataMutationRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessNoCallerDataMutation, map[string]string{
		correctnessNoCallerDataMutationFile:     sourceText,
		correctnessNoCallerDataMutationHostFile: correctnessNoCallerDataMutationHost,
	}, correctnessNoCallerDataMutationFile)
}

// correctnessNoCallerDataMutationExpect asserts the reported text in source order, and that no
// finding carries a fix.
func correctnessNoCallerDataMutationExpect(t *testing.T, result rule_testing.Result, want ...string) {
	t.Helper()
	// The shared assertion by message id, which proves the rule can fire; the checks below hold each
	// finding's exact text.
	if len(want) > 0 {
		wantIds := make([]string, len(want))
		for index := range wantIds {
			wantIds[index] = correctnessNoCallerDataMutationId
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
		if diagnostic.Message.Id != correctnessNoCallerDataMutationId {
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

// correctnessNoCallerDataMutationAddUsage models `addUsage` in ahra's `modules/os/AhraOsKingdom.ts` on
// 2026-10-03, one of the writes Kirk's ruling names. `fixed` returns the sum instead of writing it
// into the caller's total.
func correctnessNoCallerDataMutationAddUsage(fixed bool) string {
	if fixed {
		return correctnessNoCallerDataMutationSource(
			"declare function priceUsage(usage: UsageInterface): number;",
			"export function addUsage(total: UsageInterface, from: UsageInterface): UsageInterface {",
			"    const sum = { inputTokens: total.inputTokens + from.inputTokens, outputTokens: total.outputTokens + from.outputTokens, dollars: 0 };",
			"    return { ...sum, dollars: priceUsage(sum) };",
			"}",
		)
	}
	return correctnessNoCallerDataMutationSource(
		"declare function priceUsage(usage: UsageInterface): number;",
		"export function addUsage(into: UsageInterface, from: UsageInterface): void {",
		"    into.inputTokens += from.inputTokens;",
		"    into.outputTokens += from.outputTokens;",
		"    into.dollars = priceUsage(into);",
		"}",
	)
}

func TestCorrectnessNoCallerDataMutationFiresOnAddUsage(t *testing.T) {
	t.Parallel()

	result := correctnessNoCallerDataMutationRun(t, correctnessNoCallerDataMutationAddUsage(false))
	correctnessNoCallerDataMutationExpect(t, result, "into.inputTokens", "into.outputTokens", "into.dollars")
}

func TestCorrectnessNoCallerDataMutationStaysSilentOnFixedAddUsage(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoCallerDataMutationRun(t, correctnessNoCallerDataMutationAddUsage(true)))
}

func TestCorrectnessNoCallerDataMutationFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"an accumulator array pushed", []string{
			"export function checkFile(diffs: DiffInterface[], path: string): void {",
			"    diffs.push({ path });",
			"}",
		}, []string{"diffs.push"}},
		{"a caller's list sorted and returned", []string{
			"export function mergeOldestFirst(messages: MessageInterface[]): MessageInterface[] {",
			"    return messages.sort(function(left, right) { return left.createdAt - right.createdAt; });",
			"}",
		}, []string{"messages.sort"}},
		{"a record's key written through a property", []string{
			"export function sign(request: RequestInterface, host: string): void {",
			"    request.headers['host'] = host;",
			"}",
		}, []string{"request.headers['host']"}},
		{"a computed key into a record parameter", []string{
			"export function flatten(object: Record<string, unknown>, result: Record<string, unknown> = {}): Record<string, unknown> {",
			"    for(const [key, value] of Object.entries(object)) result[key] = value;",
			"    return result;",
			"}",
		}, []string{"result[key]"}},
		{"a key of a Partial Record, from FinanceQuickBooksEnricher's pushSample", []string{
			"type OutcomeType = 'Agrees' | 'Skip';",
			"export function pushSample(samples: Partial<Record<OutcomeType, DiffInterface[]>>, outcome: OutcomeType, diff: DiffInterface): void {",
			"    const bucket = samples[outcome] ?? [];",
			"    bucket.push(diff);",
			"    samples[outcome] = bucket;",
			"}",
		}, []string{"samples[outcome]"}},
		{"an element of an array parameter", []string{
			"export function parseInto(out: (string | null)[], offset: number): void {",
			"    out[offset] = null;",
			"}",
		}, []string{"out[offset]"}},
		{"updates and a delete", []string{
			"export function step(entry: EntryInterface, cache: Record<string, number>): void {",
			"    entry.position++;",
			"    --entry.position;",
			"    delete cache[entry.status];",
			"}",
		}, []string{"entry.position", "entry.position", "cache[entry.status]"}},
		{"a destructured parameter", []string{
			"export function settle({ entry }: { entry: EntryInterface }): void {",
			"    entry.status = 'done';",
			"}",
		}, []string{"entry.status"}},
		{"collections reached through a state hub", []string{
			"export function teardown(hub: HubInterface, path: string): void {",
			"    hub.watchersByPath.delete(path);",
			"    hub.seen.add(path);",
			"    hub.watchersByPath.clear();",
			"    hub.debounceTimer = null;",
			"}",
		}, []string{"hub.watchersByPath.delete", "hub.seen.add", "hub.watchersByPath.clear", "hub.debounceTimer"}},
		{"a Set and a Map passed down a recursion", []string{
			"export function collect(name: string, seen: Set<string>, depths: Map<string, number>): void {",
			"    seen.add(name);",
			"    depths.set(name, seen.size);",
			"}",
		}, []string{"seen.add", "depths.set"}},
		{"through a non-null assertion, in a class method", []string{
			"export class Walker {",
			"    visit(entry: EntryInterface): void {",
			"        entry.child!.status = 'seen';",
			"    }",
			"}",
		}, []string{"entry.child!.status"}},
		{"a project class's field", []string{
			"export class ConnectionClass { heartbeat: number | null = null; }",
			"export function arm(connection: ConnectionClass): void {",
			"    connection.heartbeat = 1;",
			"}",
		}, []string{"connection.heartbeat"}},
		{"a project property reached past a host object's", []string{
			"export function rename(panel: PanelInterface): void {",
			"    panel.canvas.fillStyle = 'red';",
			"    panel.label = 'Panel';",
			"}",
		}, []string{"panel.label"}},
		{"a local helper's parameter", []string{
			"export function count(rows: string[]): Record<string, number> {",
			"    const totals: Record<string, number> = {};",
			"    const note = function(bucket: Record<string, number>, key: string) { bucket[key] = (bucket[key] ?? 0) + 1; };",
			"    for(const row of rows) note(totals, row);",
			"    return totals;",
			"}",
		}, []string{"bucket[key]"}},
		{"a mutating call whose result is used", []string{
			"export function append(diffs: DiffInterface[]): number {",
			"    return diffs.push({ path: '' });",
			"}",
		}, []string{"diffs.push"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := correctnessNoCallerDataMutationRun(t, correctnessNoCallerDataMutationSource(testCase.lines...))
			correctnessNoCallerDataMutationExpect(t, result, testCase.want...)
		})
	}
}

func TestCorrectnessNoCallerDataMutationStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"a host object's property, declared in a declaration file", []string{
			"export function paint(context: CanvasContextInterface, port: PortInterface): void {",
			"    context.fillStyle = 'red';",
			"    context.lineWidth += 1;",
			"    port.onmessage = null;",
			"}",
		}},
		{"a host object reached from project data", []string{
			"export function paint(panel: PanelInterface, element: HostElementInterface): void {",
			"    panel.canvas.fillStyle = 'red';",
			"    element.style.color = 'blue';",
			"    element.dataset['state'] = 'open';",
			"    element.classNames.push('open');",
			"}",
		}},
		{"a default-library object's property", []string{
			"export function rewind(pattern: RegExp, error: Error): void {",
			"    pattern.lastIndex = 0;",
			"    error.message = 'wrapped: ' + error.message;",
			"}",
		}},
		{"a default-library typed array", []string{
			"export function write(buffer: Uint8Array, offset: number): void {",
			"    buffer[offset] = 0;",
			"    buffer.fill(0);",
			"}",
		}},
		{"a forEach callback's element", []string{
			"export function settleAll(entries: EntryInterface[]): void {",
			"    entries.forEach(function(entry) { entry.status = 'done'; });",
			"    entries.forEach((entry) => { entry.position = 0; });",
			"}",
		}},
		{"a reduce accumulator", []string{
			"export function group(diffs: DiffInterface[]): Record<string, DiffInterface[]> {",
			"    return diffs.reduce(function(groups: Record<string, DiffInterface[]>, diff) {",
			"        (groups[diff.path] ??= []).push(diff);",
			"        groups[diff.path] = groups[diff.path] ?? [];",
			"        return groups;",
			"    }, {});",
			"}",
		}},
		{"a callback handed to new", []string{
			"declare class RunnerClass { constructor(step: (entry: EntryInterface) => void) }",
			"export function run(): RunnerClass {",
			"    return new RunnerClass(function(entry) { entry.status = 'done'; });",
			"}",
		}},
		{"a function invoked on the spot", []string{
			"export function settle(): EntryInterface {",
			"    const holder = { status: '', position: 0 };",
			"    (function(entry: EntryInterface) { entry.status = 'done'; })(holder);",
			"    return holder;",
			"}",
		}},
		{"a parameter reassigned to a copy first", []string{
			"export function settle(entry: EntryInterface): EntryInterface {",
			"    entry = { ...entry };",
			"    entry.status = 'done';",
			"    return entry;",
			"}",
		}},
		{"a parameter reassigned in a default", []string{
			"export function settle(entry: EntryInterface, unused = (entry = { ...entry })): EntryInterface {",
			"    use(unused);",
			"    entry.status = 'done';",
			"    return entry;",
			"}",
		}},
		{"a project class's setter", []string{
			"export class GaugeClass {",
			"    #value = 0;",
			"    set value(next: number) { this.#value = Math.max(0, next); }",
			"    get value(): number { return this.#value; }",
			"}",
			"export function reset(gauge: GaugeClass): void {",
			"    gauge.value = 0;",
			"}",
		}},
		{"a local value, not a parameter", []string{
			"export function build(): DiffInterface[] {",
			"    const diffs: DiffInterface[] = [];",
			"    diffs.push({ path: '' });",
			"    const entry = { status: '', position: 0 };",
			"    entry.status = 'done';",
			"    return diffs;",
			"}",
		}},
		{"a copy sorted, not the caller's list", []string{
			"export function oldestFirst(messages: MessageInterface[]): MessageInterface[] {",
			"    return messages.slice().sort(function(left, right) { return left.createdAt - right.createdAt; });",
			"}",
			"export function oldestFirstSpread(messages: MessageInterface[]): MessageInterface[] {",
			"    return [...messages].sort(function(left, right) { return left.createdAt - right.createdAt; });",
			"}",
		}},
		{"reads of a parameter", []string{
			"export function describe(entry: EntryInterface, diffs: DiffInterface[]): string {",
			"    use(diffs.length);",
			"    use(diffs.includes(diffs[0]!));",
			"    return entry.status;",
			"}",
		}},
		{"a write through this", []string{
			"export class Counter {",
			"    count = 0;",
			"    add(amount: number): void { this.count += amount; }",
			"}",
		}},
		{"a write through a cast", []string{
			"export function force(entry: EntryInterface): void {",
			"    (entry as { status: string }).status = 'done';",
			"}",
		}},
		{"a parameter typed any", []string{
			"export function anything(value: any): void {",
			"    value.status = 'done';",
			"    value.items.push(1);",
			"}",
		}},
		{"a project type whose push is its own", []string{
			"interface RecorderInterface { push(value: string): void; set(key: string): void }",
			"export function record(recorder: RecorderInterface): void {",
			"    recorder.push('a');",
			"    recorder.set('b');",
			"}",
		}},
		{"a parameter shadowed by a local", []string{
			"export function settle(entry: EntryInterface): void {",
			"    use(entry);",
			"    {",
			"        const entry = { status: '', position: 0 };",
			"        entry.status = 'done';",
			"    }",
			"}",
		}},
		{"an array's length, which the default library declares", []string{
			"export function truncate(diffs: DiffInterface[]): void {",
			"    diffs.length = 0;",
			"}",
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rule_testing.ExpectClean(t, correctnessNoCallerDataMutationRun(t, correctnessNoCallerDataMutationSource(testCase.lines...)))
		})
	}
}

// correctnessNoCallerDataMutationHubs declares process state three ways, an interface, a class and a
// type alias, each with a reason, beside an untagged interface of the same shape and the near misses
// the match must not reach.
var correctnessNoCallerDataMutationHubs = strings.Join([]string{
	"/**",
	" * The live observer, one per process.",
	" * @processState every helper updates its counters",
	" */",
	"interface ObserverInterface { count: number; names: string[] }",
	"/** @processState the feed's one connection table */",
	"class FeedHub { count = 0; names: string[] = [] }",
	"/** @processState shared by every report */",
	"type ReportHubType = { count: number; names: string[] };",
	"interface PlainInterface { count: number; names: string[] }",
	"interface ExtendedObserverInterface extends ObserverInterface { label: string }",
	"",
}, "\n")

// A write through a parameter whose declared type is process state is exempt, through an interface, a
// class and a type alias, and through a collection method as well as an assignment.
func TestCorrectnessNoCallerDataMutationExemptsProcessState(t *testing.T) {
	t.Parallel()

	result := correctnessNoCallerDataMutationRun(t, correctnessNoCallerDataMutationSource(
		correctnessNoCallerDataMutationHubs,
		"export function observe(observer: ObserverInterface) { observer.count += 1; observer.names.push('x'); }",
		"export function feed(hub: FeedHub) { hub.count = 2; }",
		"export function report(hub: ReportHubType) { hub.count++; }",
	))
	rule_testing.ExpectClean(t, result)
}

// The match is exact: only the tagged type itself. Each of these is a different type, so each write
// reports, in the same file as the exempt ones (#dz42gce).
func TestCorrectnessNoCallerDataMutationProcessStateIsExact(t *testing.T) {
	t.Parallel()

	result := correctnessNoCallerDataMutationRun(t, correctnessNoCallerDataMutationSource(
		correctnessNoCallerDataMutationHubs,
		"export function plain(state: PlainInterface) { state.count = 1; }",
		"export function union(state: ObserverInterface | PlainInterface) { state.count = 1; }",
		"export function partial(state: Partial<ObserverInterface>) { state.count = 1; }",
		"export function readonly(state: Readonly<ObserverInterface>) { state.names.push('x'); }",
		"export function many(states: ObserverInterface[]) { states.push({ count: 0, names: [] }); }",
		"export function extended(state: ExtendedObserverInterface) { state.count = 1; }",
		"export function optional(state?: ObserverInterface) { if(state) { state.count = 1; } }",
	))
	correctnessNoCallerDataMutationExpect(t, result,
		"state.count", "state.count", "state.count", "state.names.push", "states.push", "state.count", "state.count")
}

// A bare tag is reported where it is written, and it exempts nothing: the write through its type
// still reports.
func TestCorrectnessNoCallerDataMutationReportsABareProcessStateTag(t *testing.T) {
	t.Parallel()

	result := correctnessNoCallerDataMutationRun(t, correctnessNoCallerDataMutationSource(
		"/** The hub. @processState */",
		"interface BareHubInterface { count: number }",
		"export function touch(hub: BareHubInterface) { hub.count = 1; }",
	))
	rule_testing.ExpectFindings(t, result, "processStateWithoutReason", correctnessNoCallerDataMutationId)
}

// correctnessNoCallerDataMutationContracts declares out-parameters by contract the ways the tag is
// written: on an abstract method a scheduler's jobs override (Base's OrmPersistedScheduledExecutable
// shape), and on an interface method a processor implements.
var correctnessNoCallerDataMutationContracts = strings.Join([]string{
	"interface JobRecordInterface { status: string; attempts: number; log: string[] }",
	"interface JobContextInterface { now: number; seen: string[] }",
	"abstract class ScheduledExecutable {",
	"    /**",
	"     * Runs one job.",
	"     * @mutates entity the scheduler persists the record after run returns",
	"     */",
	"    protected abstract run(entity: JobRecordInterface, context: JobContextInterface): void;",
	"}",
	"interface ProcessorInterface {",
	"    /** @mutates record the processor saves it afterwards */",
	"    process(record: JobRecordInterface): void;",
	"}",
	"",
}, "\n")

// A write through the tagged parameter is exempt in an override, which may rename it, in an override
// of that override, and in an implementation of a tagged interface method, through an assignment, an
// update and a collection call. Another parameter of the same method still reports (#tnn31qs).
func TestCorrectnessNoCallerDataMutationExemptsAnOutParameterByContract(t *testing.T) {
	t.Parallel()

	result := correctnessNoCallerDataMutationRun(t, correctnessNoCallerDataMutationSource(
		correctnessNoCallerDataMutationContracts,
		"export class SendEmails extends ScheduledExecutable {",
		"    protected run(job: JobRecordInterface, context: JobContextInterface): void {",
		"        job.status = 'sent';",
		"        job.log.push('sent');",
		"        context.seen.push('sent');",
		"    }",
		"}",
		"export class RetryEmails extends SendEmails {",
		"    protected override run(record: JobRecordInterface, context: JobContextInterface): void {",
		"        record.attempts++;",
		"        use(context);",
		"    }",
		"}",
		"export class Processor implements ProcessorInterface {",
		"    process(record: JobRecordInterface): void { record.status = 'done'; }",
		"}",
	))
	correctnessNoCallerDataMutationExpect(t, result, "context.seen.push")
}

// What the tag does not reach: a method that shares the tagged method's name and parameter names but
// overrides nothing, one that shares an interface method's name without implementing it, the tagged
// parameter's neighbour, a plain function, and a parameter at another position under the tagged name.
func TestCorrectnessNoCallerDataMutationOutParameterIsExact(t *testing.T) {
	t.Parallel()

	result := correctnessNoCallerDataMutationRun(t, correctnessNoCallerDataMutationSource(
		correctnessNoCallerDataMutationContracts,
		"export class LooksAlike {",
		"    run(entity: JobRecordInterface, context: JobContextInterface): void { entity.status = 'x'; use(context); }",
		"}",
		"export class Unrelated {",
		"    process(record: JobRecordInterface): void { record.status = 'x'; }",
		"}",
		"export class Swapped extends ScheduledExecutable {",
		"    protected run(context: JobRecordInterface, entity: JobContextInterface): void { use(context); entity.seen.push('x'); }",
		"}",
		"export function free(entity: JobRecordInterface): void { entity.status = 'x'; }",
	))
	correctnessNoCallerDataMutationExpect(t, result, "entity.status", "record.status", "entity.seen.push", "entity.status")
}

// The tagged method itself honors its own tag, for the named parameter only.
func TestCorrectnessNoCallerDataMutationHonorsAMethodsOwnContract(t *testing.T) {
	t.Parallel()

	result := correctnessNoCallerDataMutationRun(t, correctnessNoCallerDataMutationSource(
		"export class Totals {",
		"    /** @mutates into the caller hands its running total to be added to */",
		"    add(into: UsageInterface, from: UsageInterface): void {",
		"        into.inputTokens += from.inputTokens;",
		"        from.inputTokens = 0;",
		"    }",
		"}",
	))
	correctnessNoCallerDataMutationExpect(t, result, "from.inputTokens")
}

// A tag without a reason, and one naming no parameter of its method, are each reported where they are
// written and exempt nothing, so the override's write still reports.
func TestCorrectnessNoCallerDataMutationReportsAnUnusableMutatesTag(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name string
		tag  string
		want string
	}{
		{"no reason", "/** @mutates entity */", "mutatesWithoutReason"},
		{"a misspelled parameter", "/** @mutates entitty the scheduler persists it */", "mutatesUnknownParameter"},
		{"no parameter at all", "/** @mutates */", "mutatesUnknownParameter"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := correctnessNoCallerDataMutationRun(t, correctnessNoCallerDataMutationSource(
				"interface JobRecordInterface { status: string }",
				"abstract class Base {",
				"    "+testCase.tag,
				"    abstract run(entity: JobRecordInterface): void;",
				"}",
				"export class Child extends Base {",
				"    run(entity: JobRecordInterface): void { entity.status = 'x'; }",
				"}",
			))
			rule_testing.ExpectFindings(t, result, testCase.want, correctnessNoCallerDataMutationId)
		})
	}
}
