package nexus

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoUnclearedRaceTimeoutFile = "/repository/source/CorrectnessNoUnclearedRaceTimeout.ts"

// correctnessNoUnclearedRaceTimeoutNodeTimers models `@types/node/web-globals/timers.d.ts`: the timer
// functions declared inside `declare global` in a module, with the namespace `setTimeout` merges with.
// That is how ahra's `setTimeout` resolves, and the fixture's ES2022 library has no timers of its own.
var correctnessNoUnclearedRaceTimeoutNodeTimers = strings.Join([]string{
	"export {};",
	"declare global {",
	"    namespace NodeJS { interface Timeout { unref(): this; ref(): this } }",
	"    function setTimeout<TArgs extends any[]>(callback: (...args: TArgs) => void, delay?: number, ...args: TArgs): NodeJS.Timeout;",
	"    namespace setTimeout { const __promisify__: unknown; }",
	"    function clearTimeout(timeout: NodeJS.Timeout | string | number | undefined): void;",
	"}",
}, "\n")

var correctnessNoUnclearedRaceTimeoutPrelude = strings.Join([]string{
	"declare function generateImage(prompt: string): Promise<{ path: string }>;",
	"declare function work(): Promise<string>;",
	"declare function timeoutAfter(milliseconds: number): Promise<never>;",
	"declare function use(value: unknown): void;",
	"declare const milliseconds: number;",
	"",
}, "\n")

func correctnessNoUnclearedRaceTimeoutSource(lines ...string) string {
	return correctnessNoUnclearedRaceTimeoutPrelude + strings.Join(lines, "\n") + "\n"
}

func correctnessNoUnclearedRaceTimeoutRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessNoUnclearedRaceTimeout, map[string]string{
		correctnessNoUnclearedRaceTimeoutFile:            sourceText,
		"/repository/types/node/web-globals/timers.d.ts": correctnessNoUnclearedRaceTimeoutNodeTimers,
	}, correctnessNoUnclearedRaceTimeoutFile)
}

// correctnessNoUnclearedRaceTimeoutExpect asserts the reported spans in source order, and that no
// finding carries a fix.
func correctnessNoUnclearedRaceTimeoutExpect(t *testing.T, result rule_testing.Result, want ...string) {
	t.Helper()
	// The shared assertion by message id, which proves the rule can fire; the checks below hold
	// each finding's exact text.
	if len(want) > 0 {
		wantIds := make([]string, len(want))
		for index := range wantIds {
			wantIds[index] = correctnessNoUnclearedRaceTimeoutId
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
		if diagnostic.Message.Id != correctnessNoUnclearedRaceTimeoutId {
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

// correctnessNoUnclearedRaceTimeoutPhiSocial models the per-image render bound in
// `modules/phi/social/PhiSocialGenerator.ts` in ahra on 2026-10-03, trimmed to the race. `fixed`
// keeps the handle and clears it in a `finally`.
func correctnessNoUnclearedRaceTimeoutPhiSocial(fixed bool) string {
	timer := "                setTimeout(function fireTimeout() {"
	declaration := ""
	finally := ""
	if fixed {
		timer = "                imageTimer = setTimeout(function fireTimeout() {"
		declaration = "    let imageTimer: NodeJS.Timeout | undefined;"
		finally = "    finally {\n        clearTimeout(imageTimer);\n    }"
	}
	return correctnessNoUnclearedRaceTimeoutSource(
		"export async function renderImage(prompt: string) {",
		"    const imageTimeoutInMilliseconds = 10 * 60 * 1000;",
		declaration,
		"    let imageResult;",
		"    try {",
		"        imageResult = await Promise.race([",
		"            generateImage(prompt),",
		"            new Promise<never>(function rejectAfterTimeout(_resolve, reject) {",
		timer,
		"                    reject(new Error(`image-render-timeout after ${Math.round(imageTimeoutInMilliseconds / 1000)}s`));",
		"                }, imageTimeoutInMilliseconds);",
		"            }),",
		"        ]);",
		"    }",
		"    catch(error) {",
		"        use(error);",
		"        return null;",
		"    }",
		finally,
		"    return imageResult;",
		"}",
	)
}

// The real render bound, before the fix: the ten-minute timer is reported.
func TestCorrectnessNoUnclearedRaceTimeoutFiresOnPhiSocialGenerator(t *testing.T) {
	t.Parallel()

	result := correctnessNoUnclearedRaceTimeoutRun(t, correctnessNoUnclearedRaceTimeoutPhiSocial(false))
	correctnessNoUnclearedRaceTimeoutExpect(t, result, strings.Join([]string{
		"setTimeout(function fireTimeout() {",
		"                    reject(new Error(`image-render-timeout after ${Math.round(imageTimeoutInMilliseconds / 1000)}s`));",
		"                }, imageTimeoutInMilliseconds)",
	}, "\n"))
}

// After the fix: the handle is kept and cleared whichever side wins, and nothing fires.
func TestCorrectnessNoUnclearedRaceTimeoutStaysSilentOnFixedPhiSocialGenerator(t *testing.T) {
	t.Parallel()

	rule_testing.ExpectClean(t, correctnessNoUnclearedRaceTimeoutRun(t, correctnessNoUnclearedRaceTimeoutPhiSocial(true)))
}

func TestCorrectnessNoUnclearedRaceTimeoutFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"an arrow executor whose expression body is the timer", []string{
			"export async function load() {",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => setTimeout(() => reject(new Error('timeout')), milliseconds))]);",
			"}",
		}, []string{"setTimeout(() => reject(new Error('timeout')), milliseconds)"}},
		{"the timeout promise named by a const", []string{
			"export async function load() {",
			"    const timeout = new Promise<never>(function(_resolve, reject) {",
			"        setTimeout(reject, milliseconds);",
			"    });",
			"    return Promise.race([work(), timeout]);",
			"}",
		}, []string{"setTimeout(reject, milliseconds)"}},
		{"a handle kept in a const nothing reads", []string{
			"export async function load() {",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => {",
			"        const timer = setTimeout(reject, milliseconds);",
			"    })]);",
			"}",
		}, []string{"setTimeout(reject, milliseconds)"}},
		{"a handle assigned to an outer binding nothing reads", []string{
			"export async function load() {",
			"    let timer;",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => {",
			"        timer = setTimeout(reject, milliseconds);",
			"    })]);",
			"}",
		}, []string{"setTimeout(reject, milliseconds)"}},
		{"the timer reached through globalThis", []string{
			"export async function load() {",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => { globalThis.setTimeout(reject, milliseconds); })]);",
			"}",
		}, []string{"globalThis.setTimeout(reject, milliseconds)"}},
		{"a timer dropped with void", []string{
			"export async function load() {",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => { void setTimeout(reject, milliseconds); })]);",
			"}",
		}, []string{"setTimeout(reject, milliseconds)"}},
		{"two racing timeouts", []string{
			"export async function load() {",
			"    return Promise.race([",
			"        new Promise<string>((resolve) => { setTimeout(() => resolve('slow'), milliseconds); }),",
			"        new Promise<never>((_resolve, reject) => { setTimeout(reject, milliseconds * 2); }),",
			"    ]);",
			"}",
		}, []string{"setTimeout(() => resolve('slow'), milliseconds)", "setTimeout(reject, milliseconds * 2)"}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := correctnessNoUnclearedRaceTimeoutRun(t, correctnessNoUnclearedRaceTimeoutSource(testCase.lines...))
			correctnessNoUnclearedRaceTimeoutExpect(t, result, testCase.want...)
		})
	}
}

func TestCorrectnessNoUnclearedRaceTimeoutStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"the handle cleared in a finally", []string{
			"export async function load() {",
			"    let timer: NodeJS.Timeout | undefined;",
			"    try {",
			"        return await Promise.race([work(), new Promise<never>((_resolve, reject) => { timer = setTimeout(reject, milliseconds); })]);",
			"    }",
			"    finally {",
			"        clearTimeout(timer);",
			"    }",
			"}",
		}},
		{"a const handle cleared when the work settles", []string{
			"export async function load() {",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => {",
			"        const timer = setTimeout(reject, milliseconds);",
			"        void work().finally(() => clearTimeout(timer));",
			"    })]);",
			"}",
		}},
		{"an unref'd timer that cannot hold the process", []string{
			"export async function load() {",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => { setTimeout(reject, milliseconds).unref(); })]);",
			"}",
		}},
		{"the handle stored on an object", []string{
			"export async function load(state: { timer?: NodeJS.Timeout }) {",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => { state.timer = setTimeout(reject, milliseconds); })]);",
			"}",
		}},
		{"the handle handed on", []string{
			"export async function load() {",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => { use(setTimeout(reject, milliseconds)); })]);",
			"}",
		}},
		{"a timeout built by a helper, not followed", []string{
			"export async function load() {",
			"    return Promise.race([work(), timeoutAfter(milliseconds)]);",
			"}",
		}},
		{"a timer armed in a callback nested in the executor", []string{
			"export async function load() {",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => { void work().then(() => { setTimeout(reject, milliseconds); }); })]);",
			"}",
		}},
		{"a lost timer outside any race", []string{
			"export async function load() {",
			"    return Promise.all([work(), new Promise<void>((resolve) => { setTimeout(resolve, milliseconds); })]);",
			"}",
		}},
		{"a race method that is not the Promise constructor's", []string{
			"const pool = { race(items: unknown[]) { return items; } };",
			"export function load() {",
			"    return pool.race([new Promise<never>((_resolve, reject) => { setTimeout(reject, milliseconds); })]);",
			"}",
		}},
		{"a local function named setTimeout", []string{
			"function setTimeout(callback: () => void, delay: number): number { use(callback); return delay; }",
			"export async function load() {",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => { setTimeout(() => reject(new Error('timeout')), milliseconds); })]);",
			"}",
		}},
		{"a promise class that is not the library's", []string{
			"class Promise<T> { constructor(executor: (resolve: (value: T) => void, reject: (reason: unknown) => void) => void) { use(executor); } static race(items: unknown[]) { return items; } }",
			"export function load() {",
			"    return Promise.race([work(), new Promise<never>((_resolve, reject) => { setTimeout(reject, milliseconds); })]);",
			"}",
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rule_testing.ExpectClean(t, correctnessNoUnclearedRaceTimeoutRun(t, correctnessNoUnclearedRaceTimeoutSource(testCase.lines...)))
		})
	}
}

// `setTimeout` as lib.dom declares it, a global function in a global script, is the platform's too.
func TestCorrectnessNoUnclearedRaceTimeoutRecognizesTheDomDeclaration(t *testing.T) {
	t.Parallel()

	subject := strings.Join([]string{
		`/// <reference lib="dom" />`,
		"declare function work(): Promise<string>;",
		"export async function load(milliseconds: number) {",
		"    return Promise.race([work(), new Promise<never>((_resolve, reject) => { window.setTimeout(reject, milliseconds); })]);",
		"}",
	}, "\n")
	result := rule_testing.RunTypedFiles(t, CorrectnessNoUnclearedRaceTimeout, map[string]string{
		"/repository/source/Load.ts": subject,
	}, "/repository/source/Load.ts")
	correctnessNoUnclearedRaceTimeoutExpect(t, result, "window.setTimeout(reject, milliseconds)")
}
