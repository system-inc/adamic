package nexus

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessRequireResponseStatusCheckFile = "/repository/source/CorrectnessRequireResponseStatusCheck.ts"

const correctnessRequireResponseStatusCheckNetworkServiceFile = "/repository/libraries/structure/source/services/network/NetworkService.ts"

// correctnessRequireResponseStatusCheckNetworkService models Structure's
// `source/services/network/NetworkService.ts` on 2026-10-02, trimmed to the two request methods and
// the exported instance. Like the real one, it hands back whatever Response fetch resolved with.
var correctnessRequireResponseStatusCheckNetworkService = strings.Join([]string{
	`/// <reference lib="dom" />`,
	"export class NetworkService {",
	"    request(input: RequestInfo | URL, options?: RequestInit): Promise<Response> {",
	"        return fetch(input, options);",
	"    }",
	"    baseApiRequest(path: string, options?: RequestInit): Promise<Response> {",
	"        return this.request(`https://api.example.com${path}`, options);",
	"    }",
	"}",
	"export const networkService = new NetworkService();",
}, "\n")

// correctnessRequireResponseStatusCheckPrelude declares every name the cases use, so the checker
// resolves each one and no verdict rests on an unresolved identifier. The DOM library supplies the
// platform's `fetch` and `Response`, exactly as lib.dom does in the trees.
var correctnessRequireResponseStatusCheckPrelude = strings.Join([]string{
	`/// <reference lib="dom" />`,
	"import { networkService } from '../libraries/structure/source/services/network/NetworkService';",
	"interface TaskInterface { id: string; title: string }",
	"declare const url: string;",
	"declare const urls: string[];",
	"declare const flag: boolean;",
	"declare const results: unknown[];",
	"declare function use(value: unknown): void;",
	"declare function parseResponse(response: Response): Promise<unknown>;",
	"declare function pipe(stream: ReadableStream<Uint8Array>): Promise<void>;",
	"declare const r2: { signedS3Request(path: string): Promise<Response> };",
	"",
}, "\n")

func correctnessRequireResponseStatusCheckSource(lines ...string) string {
	return correctnessRequireResponseStatusCheckPrelude + strings.Join(lines, "\n") + "\n"
}

func correctnessRequireResponseStatusCheckRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessRequireResponseStatusCheck, map[string]string{
		correctnessRequireResponseStatusCheckFile:               sourceText,
		correctnessRequireResponseStatusCheckNetworkServiceFile: correctnessRequireResponseStatusCheckNetworkService,
	}, correctnessRequireResponseStatusCheckFile)
}

// correctnessRequireResponseStatusCheckFinding is one finding as its message id and the text it
// points at.
type correctnessRequireResponseStatusCheckFinding struct {
	id   string
	span string
}

// correctnessRequireResponseStatusCheckExpect asserts the findings in source order. The harness
// trims the subject file before writing it, so spans are read from the program's own text.
func correctnessRequireResponseStatusCheckExpect(t *testing.T, result rule_testing.Result, want []correctnessRequireResponseStatusCheckFinding) {
	t.Helper()
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
	var got []correctnessRequireResponseStatusCheckFinding
	for _, diagnostic := range diagnostics {
		got = append(got, correctnessRequireResponseStatusCheckFinding{diagnostic.Message.Id, text[diagnostic.Range.Pos():diagnostic.Range.End()]})
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
	if len(got) != len(want) {
		t.Fatalf("expected %d findings %v, got %d %v", len(want), want, len(got), got)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("finding %d is %v, want %v (all: %v)", index, got[index], want[index], got)
		}
	}
}

func correctnessRequireResponseStatusCheckBody(span string) correctnessRequireResponseStatusCheckFinding {
	return correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckBodyId, span}
}

func correctnessRequireResponseStatusCheckDiscard(span string) correctnessRequireResponseStatusCheckFinding {
	return correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckDiscardId, span}
}

// correctnessRequireResponseStatusCheckTaskInbox models `app/(os-layout)/_hooks/useTaskInboxRequest.ts`
// in ahra on 2026-10-02, the request function the inbox's read request caches. `check` is the line
// the fix adds between the request and the parse; empty is the file as it stands.
func correctnessRequireResponseStatusCheckTaskInbox(check string) string {
	return correctnessRequireResponseStatusCheckSource(
		"export function useTaskInboxRequest() {",
		"    return {",
		"        cacheKey: 'tasks-inbox',",
		"        request: async function(): Promise<TaskInterface[]> {",
		"            const response = await networkService.request('/api/tasks?inbox=true');",
		check,
		"            const data = (await response.json()) as { tasks: TaskInterface[] };",
		"            return data.tasks;",
		"        },",
		"    };",
		"}",
	)
}

// correctnessRequireResponseStatusCheckFinanceReview models the two write actions at the top of
// `app/(os-layout)/finance/_components/FinanceReviewView.tsx` in ahra on 2026-10-02. `fixed` keeps
// each Response and throws when it is not ok.
func correctnessRequireResponseStatusCheckFinanceReview(fixed bool) string {
	request := func(path string, body string) []string {
		call := []string{
			"    await networkService.request('" + path + "', {",
			"        method: 'POST',",
			"        body: " + body + ",",
			"    });",
		}
		if !fixed {
			return call
		}
		call[0] = "    const response = await networkService.request('" + path + "', {"
		return append(call,
			"    if(!response.ok) {",
			"        throw new Error(`"+path+" failed with ${response.status}`);",
			"    }",
		)
	}
	lines := []string{"async function postCategorize(variables: { transactionId: string; categoryId: string | null }): Promise<void> {"}
	lines = append(lines, request("/api/finance/categorize", "JSON.stringify(variables)")...)
	lines = append(lines, "}", "async function postRule(variables: { targetCategoryId: string; matchValue: string; entityId: string }): Promise<void> {")
	lines = append(lines, request("/api/finance/rules", "JSON.stringify({ ...variables, matchField: 'Merchant', source: 'Learned' })")...)
	lines = append(lines, "}", "export const actions = { postCategorize, postRule };")
	return correctnessRequireResponseStatusCheckSource(lines...)
}

// The real inbox request, before the fix: the parse is reported, and nothing else is.
func TestCorrectnessRequireResponseStatusCheckFiresOnTaskInbox(t *testing.T) {
	t.Parallel()

	result := correctnessRequireResponseStatusCheckRun(t, correctnessRequireResponseStatusCheckTaskInbox(""))
	correctnessRequireResponseStatusCheckExpect(t, result, []correctnessRequireResponseStatusCheckFinding{
		correctnessRequireResponseStatusCheckBody("response.json()"),
	})
}

// After the fix: a failed request throws before the parse, and nothing fires.
func TestCorrectnessRequireResponseStatusCheckStaysSilentOnFixedTaskInbox(t *testing.T) {
	t.Parallel()

	result := correctnessRequireResponseStatusCheckRun(t, correctnessRequireResponseStatusCheckTaskInbox(
		"            if(!response.ok) { throw new Error(`Inbox request failed with ${response.status}`); }",
	))
	rule_testing.ExpectClean(t, result)
}

// The real finance actions, before the fix: both writes drop their Response.
func TestCorrectnessRequireResponseStatusCheckFiresOnFinanceReview(t *testing.T) {
	t.Parallel()

	result := correctnessRequireResponseStatusCheckRun(t, correctnessRequireResponseStatusCheckFinanceReview(false))
	correctnessRequireResponseStatusCheckExpect(t, result, []correctnessRequireResponseStatusCheckFinding{
		correctnessRequireResponseStatusCheckDiscard("await networkService.request('/api/finance/categorize', {\n        method: 'POST',\n        body: JSON.stringify(variables),\n    });"),
		correctnessRequireResponseStatusCheckDiscard("await networkService.request('/api/finance/rules', {\n        method: 'POST',\n        body: JSON.stringify({ ...variables, matchField: 'Merchant', source: 'Learned' }),\n    });"),
	})
}

func TestCorrectnessRequireResponseStatusCheckStaysSilentOnFixedFinanceReview(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessRequireResponseStatusCheckRun(t, correctnessRequireResponseStatusCheckFinanceReview(true)))
}

func TestCorrectnessRequireResponseStatusCheckFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []correctnessRequireResponseStatusCheckFinding
	}{
		{"the platform fetch, parsed and returned", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    return response.json();",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckBody("response.json()")}},
		{"fetch reached through globalThis", []string{
			"export async function load() {",
			"    const response = await globalThis.fetch(url);",
			"    return await response.text();",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckBody("response.text()")}},
		{"the base API request", []string{
			"export async function load() {",
			"    const response = await networkService.baseApiRequest('/tasks');",
			"    return (await response.json()) as TaskInterface[];",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckBody("response.json()")}},
		{"a body read inline, on a Response nothing else can see", []string{
			"export async function load() {",
			"    return (await fetch(url)).json();",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckBody("(await fetch(url)).json()")}},
		{"a then callback that parses", []string{
			"export function load() {",
			"    return fetch(url).then(function(response) {",
			"        return response.json();",
			"    });",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckBody("response.json()")}},
		{"a body field tested in place of the status", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    const data = (await response.json()) as { error?: string };",
			"    if(data.error) {",
			"        use(response.status);",
			"        throw new Error(data.error);",
			"    }",
			"    return data;",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckBody("response.json()")}},
		{"a check on one branch only", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    if(flag) {",
			"        if(!response.ok) throw new Error('failed');",
			"    }",
			"    return response.json();",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckBody("response.json()")}},
		{"a catch that returns normally", []string{
			"export async function load() {",
			"    try {",
			"        const response = await fetch(url);",
			"        return await response.json();",
			"    }",
			"    catch {",
			"        return null;",
			"    }",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckBody("response.json()")}},
		{"the body streamed", []string{
			"export async function download() {",
			"    const response = await fetch(url);",
			"    if(!response.body) throw new Error('no body');",
			"    await pipe(response.body);",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{
			correctnessRequireResponseStatusCheckBody("response.body"),
			correctnessRequireResponseStatusCheckBody("response.body"),
		}},
		{"a loop that never leaves, so only the next turn ends the Response", []string{
			"export async function poll() {",
			"    while(true) {",
			"        const response = await fetch(url);",
			"        results.push(await response.json());",
			"    }",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckBody("response.json()")}},
		{"a Response awaited as a statement", []string{
			"export async function ping() {",
			"    await fetch(url, { method: 'POST' });",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckDiscard("await fetch(url, { method: 'POST' });")}},
		{"a Response kept and never referenced", []string{
			"export async function ping() {",
			"    const response = await fetch(url, { method: 'POST' });",
			"}",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckDiscard("await fetch(url, { method: 'POST' })")}},
		{"a module-level fetch", []string{
			"const response = await fetch(url);",
			"export const configuration = await response.json();",
		}, []correctnessRequireResponseStatusCheckFinding{correctnessRequireResponseStatusCheckBody("response.json()")}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := correctnessRequireResponseStatusCheckRun(t, correctnessRequireResponseStatusCheckSource(testCase.lines...))
			correctnessRequireResponseStatusCheckExpect(t, result, testCase.want)
		})
	}
}

func TestCorrectnessRequireResponseStatusCheckStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"checked before the parse", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    if(!response.ok) throw new Error(`failed with ${response.status}`);",
			"    return response.json();",
			"}",
		}},
		{"parsed, then checked before the data is used", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    const data = (await response.json()) as { message: string };",
			"    if(!response.ok) throw new Error(data.message);",
			"    return data;",
			"}",
		}},
		{"a status read on every path", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    if(response.status === 204) return null;",
			"    return response.json();",
			"}",
		}},
		{"checked in the ternary that parses", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    return response.ok ? response.json() : null;",
			"}",
		}},
		{"the status short-circuits ahead of the stream", []string{
			"export async function download() {",
			"    const response = await fetch(url);",
			"    if(!response.ok || !response.body) throw new Error('failed');",
			"    await pipe(response.body);",
			"}",
		}},
		{"read to be thrown, never trusted", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    const text = await response.text();",
			"    throw new Error(text);",
			"}",
		}},
		{"handed to a helper", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    return parseResponse(response);",
			"}",
		}},
		{"returned to the caller", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    return response;",
			"}",
		}},
		{"stored in an object", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    const text = await response.text();",
			"    return { response, text };",
			"}",
		}},
		{"a closure reads it, and the graph cannot say when", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    const describe = () => response.status;",
			"    use(describe);",
			"    return response.json();",
			"}",
		}},
		{"a then callback that checks", []string{
			"export function load() {",
			"    return fetch(url).then(function(response) {",
			"        if(!response.ok) throw new Error('failed');",
			"        return response.json();",
			"    });",
			"}",
		}},
		{"the status read inline", []string{
			"export async function ping() {",
			"    return (await fetch(url)).ok;",
			"}",
		}},
		{"headers only, read and kept", []string{
			"export async function etag() {",
			"    const response = await fetch(url, { method: 'HEAD' });",
			"    return response.headers.get('etag');",
			"}",
		}},
		{"a local function named fetch", []string{
			"function fetch(input: string): Promise<Response> {",
			"    return globalThis.fetch(input).then(function(response) {",
			"        if(!response.ok) throw new Error('failed');",
			"        return response;",
			"    });",
			"}",
			"export async function load() {",
			"    const response = await fetch(url);",
			"    return response.json();",
			"}",
		}},
		{"a wrapper that already throws on failure", []string{
			"export async function load() {",
			"    const response = await r2.signedS3Request('/bucket/key');",
			"    return response.text();",
			"}",
		}},
		{"a destructured Response", []string{
			"export async function ping() {",
			"    const { ok } = await fetch(url);",
			"    return ok;",
			"}",
		}},
		{"an unawaited fetch statement, a floating promise and not this rule's", []string{
			"export function ping() {",
			"    void fetch(url);",
			"    fetch(url).catch(use);",
			"}",
		}},
		{"a request method on another NetworkService class", []string{
			"class NetworkService {",
			"    request(input: string): Promise<Response> { return globalThis.fetch(input); }",
			"}",
			"const otherService = new NetworkService();",
			"export async function load() {",
			"    const response = await otherService.request(url);",
			"    return response.json();",
			"}",
		}},
		{"the parse method handed on uncalled", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    const read = response.json;",
			"    use(read);",
			"}",
		}},
		{"a clone handed on", []string{
			"export async function load() {",
			"    const response = await fetch(url);",
			"    use(response.clone());",
			"    return response.text();",
			"}",
		}},
		{"each fetch in a loop checked", []string{
			"export async function loadAll() {",
			"    for(const each of urls) {",
			"        const response = await fetch(each);",
			"        if(!response.ok) continue;",
			"        results.push(await response.json());",
			"    }",
			"}",
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rule_testing.ExpectClean(t, correctnessRequireResponseStatusCheckRun(t, correctnessRequireResponseStatusCheckSource(testCase.lines...)))
		})
	}
}

// A `fetch` that is ambient but not global, a method of a typed client declared in a namespace, is
// some other function, which may well throw on an error status. Only the global one is the platform's.
func TestCorrectnessRequireResponseStatusCheckIgnoresANamespacedFetch(t *testing.T) {
	t.Parallel()

	declarations := strings.Join([]string{
		`/// <reference lib="dom" />`,
		"declare namespace strictClient {",
		"    function fetch(input: string): Promise<Response>;",
		"}",
	}, "\n")
	subject := strings.Join([]string{
		`/// <reference lib="dom" />`,
		"export async function load(url: string) {",
		"    const response = await strictClient.fetch(url);",
		"    return response.json();",
		"}",
	}, "\n")
	rule_testing.ExpectClean(t, rule_testing.RunTypedFiles(t, CorrectnessRequireResponseStatusCheck, map[string]string{
		"/repository/types/strict-client.d.ts": declarations,
		"/repository/source/Load.ts":           subject,
	}, "/repository/source/Load.ts"))
}

// `fetch` as `@types/node` declares it: inside `declare global` in a module, not in a global script.
// Typed this way the platform's fetch must still be recognized, and lib.dom is left out so the node
// declaration is the only one.
func TestCorrectnessRequireResponseStatusCheckRecognizesTheNodeDeclaration(t *testing.T) {
	t.Parallel()

	globals := strings.Join([]string{
		"export {};",
		"declare global {",
		"    interface Response { readonly ok: boolean; readonly status: number; json(): Promise<unknown>; text(): Promise<string> }",
		"    function fetch(input: string, init?: { method?: string }): Promise<Response>;",
		"}",
	}, "\n")
	subject := strings.Join([]string{
		"export async function load(url: string) {",
		"    const response = await fetch(url);",
		"    return response.json();",
		"}",
		"export async function loadChecked(url: string) {",
		"    const response = await fetch(url);",
		"    if(!response.ok) throw new Error('failed');",
		"    return response.json();",
		"}",
	}, "\n")
	result := rule_testing.RunTypedFiles(t, CorrectnessRequireResponseStatusCheck, map[string]string{
		"/repository/types/node/web-globals/fetch.d.ts": globals,
		"/repository/source/Load.ts":                    subject,
	}, "/repository/source/Load.ts")
	correctnessRequireResponseStatusCheckExpect(t, result, []correctnessRequireResponseStatusCheckFinding{
		correctnessRequireResponseStatusCheckBody("response.json()"),
	})
}
