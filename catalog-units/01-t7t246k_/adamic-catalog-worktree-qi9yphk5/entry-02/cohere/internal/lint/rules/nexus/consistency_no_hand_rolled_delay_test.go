package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const consistencyNoHandRolledDelayFile = "/repository/source/ConsistencyNoHandRolledDelay.ts"

// Each case is a real ahra site at HEAD on 2026-10-01, trimmed to the wait, plus the spelling variants
// the rule names. One finding per case, anchored on the `new` expression.
func TestConsistencyNoHandRolledDelayReports(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"ReplicateApi.ts:228, arrow with an expression body", []string{
			"async function poll(intervalMilliseconds: number) {",
			"    await new Promise((resolve) => setTimeout(resolve, intervalMilliseconds));",
			"}",
		}},
		{"TasksWatchCommandLineInterface.ts:307, function expression with a block", []string{
			"async function watch(intervalInMilliseconds: number) {",
			"    await new Promise(function(resolve) {",
			"        setTimeout(resolve, intervalInMilliseconds);",
			"    });",
			"}",
		}},
		{"GeminiApi.ts:253, a literal duration", []string{
			"async function wait() {",
			"    await new Promise(function(resolve) {",
			"        setTimeout(resolve, 2000);",
			"    });",
			"}",
		}},
		{"FigmaMcpLauncher.ts:17, top level with a type argument", []string{
			"await new Promise<void>(function(resolve) {",
			"    setTimeout(resolve, 500);",
			"});",
			"export {};",
		}},
		{"RainbowMatrix.ts:1188, a computed duration", []string{
			"async function frame(targetElapsed: number, actualElapsed: number) {",
			"    const sleep = Math.max(0, targetElapsed - actualElapsed);",
			"    await new Promise(function(resolve) {",
			"        setTimeout(resolve, sleep);",
			"    });",
			"}",
		}},
		{"PhiSocialUpload.ts:287, a named function expression", []string{
			"async function upload(gapInMilliseconds: number) {",
			"    await new Promise(function pause(resolve) {",
			"        setTimeout(resolve, gapInMilliseconds);",
			"    });",
			"}",
		}},
		{"IntelligenceMediaGenerationApi.ts:168, a numeric separator", []string{
			"async function poll() {",
			"    await new Promise(function(resolve) {",
			"        setTimeout(resolve, 5_000);",
			"    });",
			"}",
		}},
		// NextEventLoopTurn.ts:79. Not awaited on the spot, held in a variable and raced later, and
		// still a plain timed wait: `const turnReached = delay(turnDurationInMilliseconds);` is the
		// same promise.
		{"NextEventLoopTurn.ts:79, held in a variable", []string{
			"function nextTurn(turnDurationInMilliseconds: number) {",
			"    const turnReached = new Promise<void>(function(resolve) {",
			"        setTimeout(resolve, turnDurationInMilliseconds);",
			"    });",
			"    return turnReached;",
			"}",
		}},
		{"the callback wraps resolve in an arrow", []string{
			"async function wait() {",
			"    await new Promise((resolve) => setTimeout(() => resolve(), 500));",
			"}",
		}},
		{"the callback wraps resolve in a block-bodied function", []string{
			"async function wait() {",
			"    await new Promise(function(resolve) {",
			"        setTimeout(function() {",
			"            resolve();",
			"        }, 500);",
			"    });",
			"}",
		}},
		{"globalThis.setTimeout", []string{
			"async function wait() {",
			"    await new Promise((resolve) => globalThis.setTimeout(resolve, 500));",
			"}",
		}},
		{"window.setTimeout, with a block holding a return", []string{
			"async function wait() {",
			"    await new Promise((resolve) => { return window.setTimeout(resolve, 500); });",
			"}",
		}},
		{"an unused reject parameter", []string{
			"async function wait() {",
			"    await new Promise((resolve, reject) => setTimeout(resolve, 500));",
			"}",
		}},
		// A function with a fixed duration is a use with a name around it, not a definition.
		{"a helper with a fixed duration", []string{
			"function waitASecond() {",
			"    return new Promise((resolve) => setTimeout(resolve, 1000));",
			"}",
		}},
		// The duration is a parameter, but the promise is not the whole body.
		{"a parameter duration with other statements around it", []string{
			"function wait(milliseconds: number) {",
			"    log('waiting');",
			"    return new Promise((resolve) => setTimeout(resolve, milliseconds));",
			"}",
		}},
		// The duration is computed from a parameter rather than being one, so the function is a use
		// of a wait (`return delay(seconds * 1000);`), not the primitive.
		{"a duration computed from a parameter", []string{
			"function waitSeconds(seconds: number) {",
			"    return new Promise((resolve) => setTimeout(resolve, seconds * 1000));",
			"}",
		}},
		// The function has a parameter, and the duration names something else.
		{"a duration that is not the function's parameter", []string{
			"function waitFor(label: string) {",
			"    return new Promise((resolve) => setTimeout(resolve, defaultWait));",
			"}",
		}},
		// A conditional return of the shape is a use, braced or not.
		{"a braceless conditional return inside a would-be primitive", []string{
			"function wait(milliseconds: number) {",
			"    if(ready) return new Promise((resolve) => setTimeout(resolve, milliseconds));",
			"    return undefined;",
			"}",
		}},
		{"a braced conditional return inside a would-be primitive", []string{
			"function wait(milliseconds: number) {",
			"    if(ready) {",
			"        return new Promise((resolve) => setTimeout(resolve, milliseconds));",
			"    }",
			"    return undefined;",
			"}",
		}},
		{"a parenthesized arrow body", []string{
			"async function wait() {",
			"    await new Promise((resolve) => (setTimeout(resolve, 500)));",
			"}",
		}},
		{"a defaulted resolve, which Promise always supplies", []string{
			"async function wait() {",
			"    await new Promise((resolve = noop) => setTimeout(resolve, 500));",
			"}",
		}},
		// The duration names a parameter of an outer function, not of the one the promise is the body of.
		{"the duration is an outer function's parameter", []string{
			"function outer(milliseconds: number) {",
			"    return function inner() {",
			"        return new Promise((resolve) => setTimeout(resolve, milliseconds));",
			"    };",
			"}",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceText := strings.Join(testCase.lines, "\n") + "\n"
			result := rule_testing.Run(t, ConsistencyNoHandRolledDelay, consistencyNoHandRolledDelayFile, sourceText)
			rule_testing.ExpectFindings(t, result, consistencyNoHandRolledDelayId)
			if !strings.HasPrefix(sourceText[result.Diagnostics[0].Range.Pos():], "new Promise") {
				t.Fatalf("finding starts at %q, not at the new expression", sourceText[result.Diagnostics[0].Range.Pos():][:12])
			}
			if len(result.Diagnostics[0].Fixes) != 0 {
				t.Fatalf("expected no fix, got %d", len(result.Diagnostics[0].Fixes))
			}
		})
	}
}

// Promises that wait on a timer and do something `delay` does not, each a real site or the reason a
// condition exists.
func TestConsistencyNoHandRolledDelayStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		// Delay.ts at HEAD: the definition of the primitive. The promise is the function's single
		// return and the duration is the function's own parameter.
		{"Delay.ts, the definition of delay", []string{
			"export function delay(durationInMilliseconds: number): Promise<void> {",
			"    return new Promise(function(resolve) {",
			"        setTimeout(resolve, durationInMilliseconds);",
			"    });",
			"}",
		}},
		{"an arrow definition of the same primitive", []string{
			"export const pause = (milliseconds: number) => new Promise((resolve) => setTimeout(resolve, milliseconds));",
		}},
		// AirPlayRtspConnection.ts:131: the timer is kept so a reply can cancel it and resolve early.
		{"AirPlayRtspConnection.ts:131, the timer handle is kept", []string{
			"class Connection {",
			"    waiting: (() => void) | null = null;",
			"    async wait(remaining: number) {",
			"        await new Promise<void>((resolve) => {",
			"            const timer = setTimeout(resolve, remaining);",
			"            this.waiting = function() {",
			"                clearTimeout(timer);",
			"                resolve();",
			"            };",
			"        });",
			"    }",
			"}",
		}},
		// FrameTvWebSocketConnection.ts:154: the timer is a fallback beside the real resolution.
		{"FrameTvWebSocketConnection.ts:154, a fallback timer beside other work", []string{
			"async function close(webSocketToClose: { close(): void; onclose: (() => void) | null }) {",
			"    await new Promise<void>((resolve) => {",
			"        webSocketToClose.onclose = () => resolve();",
			"        webSocketToClose.close();",
			"        setTimeout(resolve, 2000);",
			"    });",
			"}",
		}},
		// PromiseBarrier.test.ts:315 and PromiseGroup.test.ts:184: a timed value, not a wait.
		{"PromiseBarrier.test.ts:315, resolving with a value", []string{
			"const slow = new Promise(function(resolve) {",
			"    setTimeout(function() {",
			"        resolve('slow');",
			"    }, 50);",
			"});",
		}},
		{"a value passed through setTimeout's third argument", []string{
			"const slow = new Promise((resolve) => setTimeout(resolve, 50, 'slow'));",
		}},
		{"a reject path", []string{
			"const timedOut = new Promise((resolve, reject) => setTimeout(() => reject(new Error('timeout')), 50));",
		}},
		{"a reject parameter the body uses", []string{
			"const timedOut = new Promise((resolve, reject) => setTimeout(reject, 50));",
		}},
		{"no duration", []string{
			"const nextTask = new Promise((resolve) => setTimeout(resolve));",
		}},
		{"a callback that does more than resolve", []string{
			"const done = new Promise((resolve) => setTimeout(() => { log('tick'); resolve(); }, 50));",
		}},
		{"a callback with a parameter", []string{
			"const done = new Promise((resolve) => setTimeout((value: unknown) => resolve(), 50));",
		}},
		{"a different timer", []string{
			"const done = new Promise((resolve) => setInterval(resolve, 50));",
			"const later = new Promise((resolve) => clock.setTimeout(resolve, 50));",
		}},
		{"an executor that resolves immediately", []string{
			"const done = new Promise((resolve) => resolve(undefined));",
		}},
		{"a destructured or defaulted resolve", []string{
			"const done = new Promise((...callbacks) => setTimeout(callbacks[0], 50));",
		}},
		{"a second argument to the constructor", []string{
			"const done = new Promise((resolve) => setTimeout(resolve, 50), extra);",
		}},
		{"an executor passed by reference", []string{
			"const done = new Promise(startTimer);",
		}},
		{"a third executor parameter", []string{
			"const done = new Promise((resolve, reject, extra) => setTimeout(resolve, 50));",
		}},
		{"a reject parameter handed to the duration's computation", []string{
			"const done = new Promise((resolve, reject) => setTimeout(resolve, durationFor(reject)));",
		}},
		{"the timer followed by another statement", []string{
			"const done = new Promise((resolve) => {",
			"    setTimeout(resolve, 50);",
			"    log('scheduled');",
			"});",
		}},
		{"a different timer on globalThis", []string{
			"const done = new Promise((resolve) => globalThis.setInterval(resolve, 50));",
		}},
		{"a rest parameter in the resolve position", []string{
			"const done = new Promise((...resolve) => setTimeout(resolve, 50));",
		}},
		{"an executor with no parameters", []string{
			"const done = new Promise(() => setTimeout(markDone, 50));",
		}},
		{"a destructured second parameter", []string{
			"const done = new Promise((resolve, { length }) => setTimeout(resolve, 50));",
		}},
		// In a server render `window` is undefined, the optional chain skips the call, and the promise
		// never settles, which is not what `delay` does.
		{"an optional chain to the timer", []string{
			"const done = new Promise((resolve) => window?.setTimeout(resolve, 50));",
		}},
		{"a promise that is not Promise", []string{
			"const done = new Deferred((resolve) => setTimeout(resolve, 50));",
		}},
		{"the replacement", []string{
			"import { delay } from '@nexus/source/coordination/Delay';",
			"await delay(500);",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.Run(t, ConsistencyNoHandRolledDelay, consistencyNoHandRolledDelayFile, strings.Join(testCase.lines, "\n")+"\n")
			rule_testing.ExpectClean(t, result)
		})
	}
}
