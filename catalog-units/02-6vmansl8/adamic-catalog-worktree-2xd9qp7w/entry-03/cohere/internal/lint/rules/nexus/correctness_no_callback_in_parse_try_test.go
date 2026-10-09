package nexus

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoCallbackInParseTryFile = "/repository/source/CorrectnessNoCallbackInParseTry.ts"

const correctnessNoCallbackInParseTryJsonFile = "/repository/libraries/structure/libraries/nexus/source/structured-text/json/Json.ts"

const correctnessNoCallbackInParseTryLookalikeFile = "/repository/source/LocalJson.ts"

const correctnessNoCallbackInParseTryModuleJsonFile = "/repository/source/ModuleJson.d.ts"

// correctnessNoCallbackInParseTryModuleJson is an interface named `JSON` with a `parse` method, in a
// declaration file that is a module rather than the global scope: not the platform's JSON.
var correctnessNoCallbackInParseTryModuleJson = strings.Join([]string{
	"export interface JSON { parse(text: string): unknown }",
	"export declare const ModuleJson: JSON;",
}, "\n")

// correctnessNoCallbackInParseTryJson models nexus `source/structured-text/json/Json.ts` on
// 2026-10-03, trimmed to the two parsers: `parseJson` returns an outcome and never throws,
// `parseJsonOrThrow` throws on malformed or misshapen text.
var correctnessNoCallbackInParseTryJson = strings.Join([]string{
	"export type GuardType<ValueType> = (value: unknown) => value is ValueType;",
	"export type JsonParseOutcomeType<ValueType = unknown> =",
	"    | { outcome: 'Parsed'; value: ValueType }",
	"    | { outcome: 'Invalid'; message: string };",
	"export function parseJson<ValueType = unknown>(text: string, context: string, isValue?: GuardType<ValueType>): JsonParseOutcomeType<ValueType> {",
	"    let parsed: unknown;",
	"    try {",
	"        parsed = JSON.parse(text);",
	"    }",
	"    catch(error) {",
	"        const message = error instanceof Error ? error.message : String(error);",
	"        return { outcome: 'Invalid', message: `${context}: not valid JSON (${message})` };",
	"    }",
	"    if(isValue === undefined) return { outcome: 'Parsed', value: parsed as ValueType };",
	"    if(!isValue(parsed)) return { outcome: 'Invalid', message: `${context}: valid JSON, but not the expected shape` };",
	"    return { outcome: 'Parsed', value: parsed };",
	"}",
	"export function parseJsonOrThrow<ValueType>(text: string, context: string, isValue: GuardType<ValueType>): ValueType {",
	"    const parseOutcome = parseJson(text, context, isValue);",
	"    if(parseOutcome.outcome === 'Invalid') throw new Error(parseOutcome.message);",
	"    return parseOutcome.value;",
	"}",
}, "\n")

// correctnessNoCallbackInParseTryLookalike is a function of the same name outside nexus, which is
// not the parser this rule reads.
var correctnessNoCallbackInParseTryLookalike = strings.Join([]string{
	"export function parseJsonOrThrow(text: string): unknown {",
	"    return text.length;",
	"}",
}, "\n")

// correctnessNoCallbackInParseTryPrelude declares every name the cases use, so the checker resolves
// each one and no verdict rests on an unresolved identifier.
var correctnessNoCallbackInParseTryPrelude = strings.Join([]string{
	"import { parseJson, parseJsonOrThrow } from '../libraries/structure/libraries/nexus/source/structured-text/json/Json';",
	"import type { GuardType } from '../libraries/structure/libraries/nexus/source/structured-text/json/Json';",
	"import { parseJsonOrThrow as parseLocalJson } from './LocalJson';",
	"import { parseJsonOrThrow as parseCheckedJson } from '../libraries/structure/libraries/nexus/source/structured-text/json/Json';",
	"import { ModuleJson } from './ModuleJson';",
	"interface SettingsInterface { theme: string }",
	"interface MessageInterface { type: string }",
	"declare const isSettings: GuardType<SettingsInterface>;",
	"declare const isMessage: GuardType<MessageInterface>;",
	"declare function use(value: unknown): void;",
	"declare function setTimeout(callback: () => void, milliseconds: number): number;",
	"",
}, "\n")

func correctnessNoCallbackInParseTrySource(lines ...string) string {
	return correctnessNoCallbackInParseTryPrelude + strings.Join(lines, "\n") + "\n"
}

func correctnessNoCallbackInParseTryRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessNoCallbackInParseTry, map[string]string{
		correctnessNoCallbackInParseTryFile:           sourceText,
		correctnessNoCallbackInParseTryJsonFile:       correctnessNoCallbackInParseTryJson,
		correctnessNoCallbackInParseTryLookalikeFile:  correctnessNoCallbackInParseTryLookalike,
		correctnessNoCallbackInParseTryModuleJsonFile: correctnessNoCallbackInParseTryModuleJson,
	}, correctnessNoCallbackInParseTryFile)
}

// correctnessNoCallbackInParseTryExpect asserts the findings in source order, by the text each one
// points at. The harness trims the subject file before writing it, so spans are read from the
// program's own text.
func correctnessNoCallbackInParseTryExpect(t *testing.T, result rule_testing.Result, want []string) {
	t.Helper()
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	wantIds := make([]string, 0, len(want))
	for range want {
		wantIds = append(wantIds, correctnessNoCallbackInParseTryId)
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	text := result.SourceFile.Text()
	for index, diagnostic := range diagnostics {
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
		if got := text[diagnostic.Range.Pos():diagnostic.Range.End()]; got != want[index] {
			t.Fatalf("finding %d is %q, want %q", index, got, want[index])
		}
	}
}

// correctnessNoCallbackInParseTryClaudeStream models `parseStreamLines` in
// `modules/claude/utilities/ClaudeUtilities.ts` in ahra. Before the fix (`4e385b61^`) one try held
// the parse and the caller's handler; after it, nexus `parseJson` decides and the handler runs
// outside any try.
func correctnessNoCallbackInParseTryClaudeStream(fixed bool) string {
	header := []string{
		"export function parseStreamLines(",
		"    lineBuffer: string,",
		"    data: string,",
		"    onMessage: (message: unknown) => void,",
		"    onUnparseable?: (line: string) => void,",
		"): string {",
		"    const lines = (lineBuffer + data).split('\\n');",
		"    const unconsumedRemainder = lines.pop() || '';",
		"",
		"    for(const line of lines) {",
		"        if(!line.trim()) continue;",
	}
	body := []string{
		"        try {",
		"            const message: unknown = JSON.parse(line);",
		"            onMessage(message);",
		"        }",
		"        catch {",
		"            onUnparseable?.(line);",
		"        }",
	}
	if fixed {
		body = []string{
			"        const lineParse = parseJson(line, 'claude stream-json line', isMessage);",
			"        if(lineParse.outcome === 'Invalid') {",
			"            onUnparseable?.(line);",
			"            continue;",
			"        }",
			"        onMessage(lineParse.value);",
		}
	}
	footer := []string{
		"    }",
		"",
		"    return unconsumedRemainder;",
		"}",
	}
	return correctnessNoCallbackInParseTrySource(append(append(header, body...), footer...)...)
}

// correctnessNoCallbackInParseTryXStream models `consumeStream` in `modules/x/XStreamApi.ts` in ahra,
// the inner line loop. Before the fix (`1382cbe9^`) the handler ran inside the parse try under an
// empty catch; after it, a `'Parsed'` outcome gates the handler.
func correctnessNoCallbackInParseTryXStream(fixed bool) string {
	body := []string{
		"            try {",
		"                const parsed = JSON.parse(line) as StreamTweetEnvelopeInterface;",
		"                onEvent(parsed);",
		"            }",
		"            catch {",
		"                // Heartbeat or partial frame; skip.",
		"            }",
	}
	if fixed {
		body = []string{
			"            // A partial frame, or a frame that is not a tweet envelope, is skipped",
			"            const lineParse = parseJson(line, 'X stream line', isStreamTweetEnvelope);",
			"            if(lineParse.outcome === 'Parsed') onEvent(lineParse.value);",
		}
	}
	lines := []string{
		"interface StreamTweetEnvelopeInterface {",
		"    data?: { id: string; text: string; author_id?: string; created_at?: string };",
		"}",
		"declare const isStreamTweetEnvelope: GuardType<StreamTweetEnvelopeInterface>;",
		"export function consumeLines(",
		"    chunks: string[],",
		"    onEvent: (event: StreamTweetEnvelopeInterface) => void,",
		"): void {",
		"    let buffer = '';",
		"    for(const chunk of chunks) {",
		"        buffer += chunk;",
		"        let newlineIndex = buffer.indexOf('\\n');",
		"        while(newlineIndex !== -1) {",
		"            const line = buffer.substring(0, newlineIndex).trim();",
		"            buffer = buffer.substring(newlineIndex + 1);",
		"            newlineIndex = buffer.indexOf('\\n');",
		"            if(line.length === 0) continue;",
	}
	lines = append(lines, body...)
	lines = append(lines, "        }", "    }", "}")
	return correctnessNoCallbackInParseTrySource(lines...)
}

// The original, `#wv045mc`: the handler inside the parse try is reported, and the fallback in the
// catch is not.
func TestCorrectnessNoCallbackInParseTryFiresOnClaudeStream(t *testing.T) {
	t.Parallel()

	result := correctnessNoCallbackInParseTryRun(t, correctnessNoCallbackInParseTryClaudeStream(false))
	correctnessNoCallbackInParseTryExpect(t, result, []string{"onMessage(message)"})
}

func TestCorrectnessNoCallbackInParseTryStaysSilentOnFixedClaudeStream(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoCallbackInParseTryRun(t, correctnessNoCallbackInParseTryClaudeStream(true)))
}

// The new copy at `XStreamApi.ts:45`, under an empty catch.
func TestCorrectnessNoCallbackInParseTryFiresOnXStream(t *testing.T) {
	t.Parallel()

	result := correctnessNoCallbackInParseTryRun(t, correctnessNoCallbackInParseTryXStream(false))
	correctnessNoCallbackInParseTryExpect(t, result, []string{"onEvent(parsed)"})
}

func TestCorrectnessNoCallbackInParseTryStaysSilentOnFixedXStream(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoCallbackInParseTryRun(t, correctnessNoCallbackInParseTryXStream(true)))
}

func TestCorrectnessNoCallbackInParseTryFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"a Node-style callback given the parse and its failure", []string{
			"export function readSettings(text: string, callback: (error: Error | null, settings?: unknown) => void): void {",
			"    try {",
			"        callback(null, JSON.parse(text));",
			"    }",
			"    catch {",
			"        callback(new Error('settings are not JSON'));",
			"    }",
			"}",
		}, []string{"callback(null, JSON.parse(text))"}},
		{"a callback that supplies the text, through nexus parseJsonOrThrow", []string{
			"export function loadSettings(read: () => string): SettingsInterface | null {",
			"    try {",
			"        return parseJsonOrThrow(read(), 'settings', isSettings);",
			"    }",
			"    catch {",
			"        return null;",
			"    }",
			"}",
		}, []string{"read()"}},
		{"nexus parseJsonOrThrow imported under another name", []string{
			"export function emit(line: string, onMessage: (message: MessageInterface) => void): void {",
			"    try {",
			"        onMessage(parseCheckedJson(line, 'line', isMessage));",
			"    }",
			"    catch {}",
			"}",
		}, []string{"onMessage(parseCheckedJson(line, 'line', isMessage))"}},
		{"a callback run before the parse", []string{
			"export function start(text: string, onStart: () => void): unknown {",
			"    try {",
			"        onStart();",
			"        return JSON.parse(text);",
			"    }",
			"    catch {",
			"        return undefined;",
			"    }",
			"}",
		}, []string{"onStart()"}},
		{"a callback destructured from a parameter", []string{
			"export function Editor({ text, onChange }: { text: string; onChange: (value: unknown) => void }): string {",
			"    try {",
			"        onChange(JSON.parse(text));",
			"        return '';",
			"    }",
			"    catch {",
			"        return 'Invalid JSON';",
			"    }",
			"}",
		}, []string{"onChange(JSON.parse(text))"}},
		{"an optional callback called with ?.", []string{
			"export function emit(text: string, onMessage?: (message: unknown) => void): void {",
			"    try {",
			"        const message: unknown = JSON.parse(text);",
			"        onMessage?.(message);",
			"    }",
			"    catch {}",
			"}",
		}, []string{"onMessage?.(message)"}},
		{"a non-null asserted, parenthesized callback", []string{
			"export function emit(text: string, onMessage: ((message: unknown) => void) | undefined): void {",
			"    try {",
			"        (onMessage!)(JSON.parse(text));",
			"    }",
			"    catch {}",
			"}",
		}, []string{"(onMessage!)(JSON.parse(text))"}},
		{"a catch binding that is never read", []string{
			"export function emit(text: string, onMessage: (message: unknown) => void): boolean {",
			"    try {",
			"        onMessage(JSON.parse(text));",
			"        return true;",
			"    }",
			"    catch(error) {",
			"        return false;",
			"    }",
			"}",
		}, []string{"onMessage(JSON.parse(text))"}},
		{"middleware handing on to next inside the body parse", []string{
			"export function parseBody(body: string, request: { body?: unknown }, next: () => void, fail: (status: number) => void): void {",
			"    try {",
			"        request.body = JSON.parse(body);",
			"        next();",
			"    }",
			"    catch {",
			"        fail(400);",
			"    }",
			"}",
		}, []string{"next()"}},
		{"a fallback in a nested try's catch still reaches the outer catch", []string{
			"export function emit(text: string, onUnparseable: (text: string) => void): unknown {",
			"    try {",
			"        const value: unknown = JSON.parse(text);",
			"        try {",
			"            use(value);",
			"        }",
			"        catch {",
			"            onUnparseable(text);",
			"        }",
			"        return value;",
			"    }",
			"    catch {",
			"        return null;",
			"    }",
			"}",
		}, []string{"onUnparseable(text)"}},
		{"a callback inside a nested try with no catch of its own", []string{
			"export function emit(text: string, onMessage: (message: unknown) => void, onDone: () => void): void {",
			"    try {",
			"        try {",
			"            onMessage(JSON.parse(text));",
			"        }",
			"        finally {",
			"            onDone();",
			"        }",
			"    }",
			"    catch {}",
			"}",
		}, []string{"onMessage(JSON.parse(text))", "onDone()"}},
		{"a local class named Promise is not the platform's", []string{
			"class Promise {",
			"    constructor(executor: (resolve: (value: unknown) => void) => void) { executor(use); }",
			"}",
			"export function load(text: string): Promise {",
			"    return new Promise(function(resolve) {",
			"        try {",
			"            resolve(JSON.parse(text));",
			"        }",
			"        catch {}",
			"    });",
			"}",
		}, []string{"resolve(JSON.parse(text))"}},
		{"two callbacks in one guarded block", []string{
			"export function emit(text: string, onMessage: (message: unknown) => void, onDone: () => void): void {",
			"    try {",
			"        onMessage(JSON.parse(text));",
			"        onDone();",
			"    }",
			"    catch {}",
			"}",
		}, []string{"onMessage(JSON.parse(text))", "onDone()"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := correctnessNoCallbackInParseTryRun(t, correctnessNoCallbackInParseTrySource(testCase.lines...))
			correctnessNoCallbackInParseTryExpect(t, result, testCase.want)
		})
	}
}

func TestCorrectnessNoCallbackInParseTryStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"the parse in the try and the callback after it", []string{
			"export function emit(line: string, onMessage: (message: unknown) => void, onUnparseable?: (line: string) => void): void {",
			"    let message: unknown;",
			"    try {",
			"        message = JSON.parse(line);",
			"    }",
			"    catch {",
			"        onUnparseable?.(line);",
			"        return;",
			"    }",
			"    onMessage(message);",
			"}",
		}},
		{"a catch that rethrows what is not a SyntaxError", []string{
			"export function emit(line: string, onMessage: (message: unknown) => void): void {",
			"    try {",
			"        onMessage(JSON.parse(line));",
			"    }",
			"    catch(error) {",
			"        if(!(error instanceof SyntaxError)) throw error;",
			"    }",
			"}",
		}},
		{"a catch that hands the error on as itself", []string{
			"export function emit(line: string, onMessage: (message: unknown) => void, onError: (error: unknown) => void): void {",
			"    try {",
			"        onMessage(JSON.parse(line));",
			"    }",
			"    catch(error) {",
			"        onError(error);",
			"    }",
			"}",
		}},
		{"a catch that reads the error through shorthand", []string{
			"export function emit(line: string, onMessage: (message: unknown) => void): unknown {",
			"    try {",
			"        onMessage(JSON.parse(line));",
			"        return null;",
			"    }",
			"    catch(error) {",
			"        return { error };",
			"    }",
			"}",
		}},
		{"a destructured catch binding", []string{
			"export function emit(line: string, onMessage: (message: unknown) => void): string {",
			"    try {",
			"        onMessage(JSON.parse(line));",
			"        return '';",
			"    }",
			"    catch({ message }) {",
			"        return 'failed';",
			"    }",
			"}",
		}},
		{"the platform Promise's resolve and reject", []string{
			"export function load(text: string): Promise<unknown> {",
			"    return new Promise(function(resolve, reject) {",
			"        try {",
			"            resolve(JSON.parse(text));",
			"        }",
			"        catch {",
			"            reject(new Error('not JSON'));",
			"        }",
			"    });",
			"}",
		}},
		{"the platform Promise's resolve, through an arrow in parentheses", []string{
			"export function load(text: string): Promise<unknown> {",
			"    return new Promise(((resolve) => {",
			"        try {",
			"            resolve(JSON.parse(text));",
			"        }",
			"        catch {}",
			"    }));",
			"}",
		}},
		{"the callback deferred in a closure", []string{
			"export function emit(line: string, onMessage: (message: unknown) => void): void {",
			"    try {",
			"        const message: unknown = JSON.parse(line);",
			"        setTimeout(() => onMessage(message), 0);",
			"    }",
			"    catch {}",
			"}",
		}},
		{"the callback inside a nested try that catches it first", []string{
			"export function emit(line: string, onMessage: (message: unknown) => void): void {",
			"    try {",
			"        const message: unknown = JSON.parse(line);",
			"        try {",
			"            onMessage(message);",
			"        }",
			"        catch(error) {",
			"            use(error);",
			"        }",
			"    }",
			"    catch {}",
			"}",
		}},
		{"a method on a parameter", []string{
			"export function emit(line: string, logger: { info(value: unknown): void }): void {",
			"    try {",
			"        logger.info(JSON.parse(line));",
			"    }",
			"    catch {}",
			"}",
		}},
		{"an any parameter", []string{
			"export function emit(line: string, handler: any): void {",
			"    try {",
			"        handler(JSON.parse(line));",
			"    }",
			"    catch {}",
			"}",
		}},
		{"a local function", []string{
			"export function emit(line: string): void {",
			"    const handle = (message: unknown) => use(message);",
			"    try {",
			"        handle(JSON.parse(line));",
			"    }",
			"    catch {}",
			"}",
		}},
		{"a local object named JSON", []string{
			"const JSON = { parse(text: string): unknown { return text; } };",
			"export function emit(line: string, onMessage: (message: unknown) => void): void {",
			"    try {",
			"        onMessage(JSON.parse(line));",
			"    }",
			"    catch {}",
			"}",
		}},
		{"a JSON interface declared in a module, not the global one", []string{
			"export function emit(line: string, onMessage: (message: unknown) => void): void {",
			"    try {",
			"        onMessage(ModuleJson.parse(line));",
			"    }",
			"    catch {}",
			"}",
		}},
		{"JSON.stringify, which is not a parse", []string{
			"export function emit(value: unknown, onText: (text: string) => void): void {",
			"    try {",
			"        onText(JSON.stringify(value));",
			"    }",
			"    catch {}",
			"}",
		}},
		{"a parseJsonOrThrow that is not nexus's", []string{
			"export function emit(line: string, onMessage: (message: unknown) => void): void {",
			"    try {",
			"        onMessage(parseLocalJson(line));",
			"    }",
			"    catch {}",
			"}",
		}},
		{"nexus parseJson, which never throws", []string{
			"export function emit(line: string, onMessage: (message: unknown) => void): void {",
			"    try {",
			"        const lineParse = parseJson(line, 'line');",
			"        if(lineParse.outcome === 'Parsed') onMessage(lineParse.value);",
			"    }",
			"    catch {}",
			"}",
		}},
		{"a try with no catch", []string{
			"export function emit(line: string, onMessage: (message: unknown) => void, onDone: () => void): void {",
			"    try {",
			"        onMessage(JSON.parse(line));",
			"    }",
			"    finally {",
			"        onDone();",
			"    }",
			"}",
		}},
		{"a callback in a try that parses nothing", []string{
			"export function emit(message: unknown, onMessage: (message: unknown) => void): void {",
			"    try {",
			"        onMessage(message);",
			"    }",
			"    catch {}",
			"}",
		}},
		{"a callback in the finally of the parse try", []string{
			"export function emit(line: string, onDone: () => void): unknown {",
			"    try {",
			"        return JSON.parse(line);",
			"    }",
			"    catch {",
			"        return null;",
			"    }",
			"    finally {",
			"        onDone();",
			"    }",
			"}",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rule_testing.ExpectClean(t, correctnessNoCallbackInParseTryRun(t, correctnessNoCallbackInParseTrySource(testCase.lines...)))
		})
	}
}
