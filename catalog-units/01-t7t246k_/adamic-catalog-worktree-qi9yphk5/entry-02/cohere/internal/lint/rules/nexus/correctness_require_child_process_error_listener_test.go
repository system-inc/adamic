package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessRequireChildProcessErrorListenerFile = "/repository/source/CorrectnessRequireChildProcessErrorListener.ts"

const correctnessRequireChildProcessErrorListenerNodeFile = "/repository/source/node.d.ts"

func correctnessRequireChildProcessErrorListenerSource(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

// correctnessRequireChildProcessErrorListenerNodeTypes is the slice of `@types/node` the rule
// resolves against, in the layout of 26.2.0 (ahra's): the functions live in `declare module
// "node:child_process"` and `"child_process"` re-exports them, `ChildProcess` takes its typed `on`
// from an event map, and `node:events` exports `once`.
var correctnessRequireChildProcessErrorListenerNodeTypes = correctnessRequireChildProcessErrorListenerSource(
	`declare module "node:events" {`,
	"    class EventEmitter {",
	"        on(eventName: string | symbol, listener: (...args: any[]) => void): this;",
	"        once(eventName: string | symbol, listener: (...args: any[]) => void): this;",
	"        addListener(eventName: string | symbol, listener: (...args: any[]) => void): this;",
	"        prependListener(eventName: string | symbol, listener: (...args: any[]) => void): this;",
	"        prependOnceListener(eventName: string | symbol, listener: (...args: any[]) => void): this;",
	"        off(eventName: string | symbol, listener: (...args: any[]) => void): this;",
	"        removeListener(eventName: string | symbol, listener: (...args: any[]) => void): this;",
	"        removeAllListeners(eventName?: string | symbol): this;",
	"        setMaxListeners(count: number): this;",
	"        emit(eventName: string | symbol, ...args: any[]): boolean;",
	"    }",
	"    function once(emitter: EventEmitter, eventName: string): Promise<any[]>;",
	"}",
	`declare module "node:child_process" {`,
	`    import { EventEmitter } from "node:events";`,
	"    interface Readable extends EventEmitter { on(eventName: 'data', listener: (chunk: Uint8Array) => void): this }",
	"    interface ChildProcessEventMap {",
	"        close: [code: number | null, signal: string | null];",
	"        error: [error: Error];",
	"        exit: [code: number | null, signal: string | null];",
	"        spawn: [];",
	"    }",
	"    class ChildProcess extends EventEmitter {",
	"        stdout: Readable;",
	"        stderr: Readable;",
	"        pid?: number;",
	"        kill(signal?: string): boolean;",
	"        unref(): void;",
	"        on<E extends keyof ChildProcessEventMap>(eventName: E, listener: (...args: ChildProcessEventMap[E]) => void): this;",
	"        on(eventName: string | symbol, listener: (...args: any[]) => void): this;",
	"    }",
	"    interface SpawnOptions { stdio?: unknown; detached?: boolean; shell?: boolean | string; cwd?: string }",
	"    function spawn(command: string, options?: SpawnOptions): ChildProcess;",
	"    function spawn(command: string, args: readonly string[], options?: SpawnOptions): ChildProcess;",
	"    function fork(modulePath: string, args?: readonly string[], options?: SpawnOptions): ChildProcess;",
	"    function exec(command: string, callback?: (error: Error | null, stdout: string) => void): ChildProcess;",
	"    function execFile(file: string, args?: readonly string[], options?: SpawnOptions): ChildProcess;",
	"    function spawnSync(command: string, args: readonly string[]): { stdout: string };",
	"}",
	`declare module "child_process" {`,
	`    export * from "node:child_process";`,
	"}",
)

// correctnessRequireChildProcessErrorListenerPrelude imports the module the way ahra does and
// declares the names the cases use, so the checker resolves each one.
var correctnessRequireChildProcessErrorListenerPrelude = correctnessRequireChildProcessErrorListenerSource(
	"import * as NodeChildProcess from 'node:child_process';",
	"import * as NodeEvents from 'node:events';",
	"declare const path: string;",
	"declare const flag: boolean;",
	"declare const commandArguments: string[];",
	"declare const children: NodeChildProcess.ChildProcess[];",
	"declare function ignore(): void;",
	"declare function done(): void;",
	"declare function wire(child: NodeChildProcess.ChildProcess): void;",
	"",
)

func correctnessRequireChildProcessErrorListenerRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessRequireChildProcessErrorListener, map[string]string{
		correctnessRequireChildProcessErrorListenerFile:     sourceText,
		correctnessRequireChildProcessErrorListenerNodeFile: correctnessRequireChildProcessErrorListenerNodeTypes,
	}, correctnessRequireChildProcessErrorListenerFile)
}

// correctnessRequireChildProcessErrorListenerExpect asserts one finding per wanted span, each on
// the producing call, in source order. The harness trims the subject file before writing it, so
// spans are read from the program's own text.
func correctnessRequireChildProcessErrorListenerExpect(t *testing.T, result rule_testing.Result, wantSpans []string) {
	t.Helper()
	wantIds := make([]string, 0, len(wantSpans))
	for range wantSpans {
		wantIds = append(wantIds, correctnessRequireChildProcessErrorListenerId)
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	text := result.SourceFile.Text()
	for index, diagnostic := range result.Diagnostics {
		reported := text[diagnostic.Range.Pos():diagnostic.Range.End()]
		if reported != wantSpans[index] {
			t.Fatalf("finding %d points at\n%s\nwant\n%s", index, reported, wantSpans[index])
		}
		if diagnostic.Message.Description != correctnessRequireChildProcessErrorListenerMessage().Description {
			t.Fatalf("message is %q", diagnostic.Message.Description)
		}
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
}

// correctnessRequireChildProcessErrorListenerMonitors models `prCheckStream` in
// `modules/os/sensation/AhraOsMonitors.ts` in ahra on 2026-10-03, trimmed to the child's uses: the
// spawn at `:140`, the stderr drain, the stdout reader, the `'exit'` handler and the kill in the
// `finally`. `fix` is the line the fix adds after the spawn; empty is the file as it stands.
func correctnessRequireChildProcessErrorListenerMonitors(fix string) string {
	return correctnessRequireChildProcessErrorListenerPrelude + correctnessRequireChildProcessErrorListenerSource(
		"declare function noOperation(): void;",
		"export async function* prCheckStream(pullRequest: string): AsyncIterable<string> {",
		"    const child = NodeChildProcess.spawn(",
		"        'gh',",
		"        ['pr', 'checks', pullRequest, '--watch', '--json', 'name,bucket,state,workflow'],",
		"        { stdio: ['ignore', 'pipe', 'pipe'] },",
		"    );",
		fix,
		"    child.stderr.on('data', noOperation);",
		"    const queue: string[] = [];",
		"    let childExited = false;",
		"    child.stdout.on('data', function(chunk: Uint8Array) {",
		"        queue.push(String(chunk));",
		"    });",
		"    child.on('exit', function() {",
		"        childExited = true;",
		"    });",
		"    try {",
		"        while(!childExited) {",
		"            const event = queue.shift();",
		"            if(event !== undefined) yield event;",
		"            await new Promise<void>(function(resolve) { setTimeout(resolve, 10); });",
		"        }",
		"    }",
		"    finally {",
		"        if(!childExited) {",
		"            child.kill('SIGTERM');",
		"        }",
		"    }",
		"}",
	)
}

const correctnessRequireChildProcessErrorListenerMonitorsSpan = "NodeChildProcess.spawn(\n" +
	"        'gh',\n" +
	"        ['pr', 'checks', pullRequest, '--watch', '--json', 'name,bucket,state,workflow'],\n" +
	"        { stdio: ['ignore', 'pipe', 'pipe'] },\n" +
	"    )"

// The real site, before the fix: four uses of the child and none of them `'error'`.
func TestCorrectnessRequireChildProcessErrorListenerFiresOnMonitors(t *testing.T) {
	t.Parallel()

	result := correctnessRequireChildProcessErrorListenerRun(t, correctnessRequireChildProcessErrorListenerMonitors(""))
	correctnessRequireChildProcessErrorListenerExpect(t, result, []string{correctnessRequireChildProcessErrorListenerMonitorsSpan})
}

// The real site, after the fix: a failed spawn ends the stream instead of the daemon.
func TestCorrectnessRequireChildProcessErrorListenerStaysSilentOnFixedMonitors(t *testing.T) {
	t.Parallel()

	result := correctnessRequireChildProcessErrorListenerRun(t, correctnessRequireChildProcessErrorListenerMonitors(
		"    child.on('error', function() { childExited = true; });",
	))
	correctnessRequireChildProcessErrorListenerExpect(t, result, nil)
}

// correctnessRequireChildProcessErrorListenerPresence models `getRecentConversationContext` in
// `modules/presence/PresenceApi.ts` in ahra on 2026-10-03: the module comes from
// `process.getBuiltinModule` cast to the module's type, so the call resolves through the cast.
// `fix` is the line the fix adds; empty is the file as it stands.
func correctnessRequireChildProcessErrorListenerPresence(fix string) string {
	return correctnessRequireChildProcessErrorListenerPrelude + correctnessRequireChildProcessErrorListenerSource(
		"declare function getBuiltinModule(id: string): unknown;",
		"export function getRecentConversationContext(): Promise<string> {",
		"    return new Promise(function recentConvoPromise(resolve) {",
		"        const nodeChildProcess = getBuiltinModule('node:child_process') as typeof NodeChildProcess;",
		"        const childProcess = nodeChildProcess.spawn('node', [path, 'pensieve', '--recent', '25k'], {",
		"            cwd: path,",
		"        });",
		"        let output = '';",
		"        childProcess.stdout?.on('data', function handleStdout(data: Uint8Array) {",
		"            output += String(data);",
		"        });",
		"        childProcess.on('close', function handleClose() {",
		"            resolve(output.trim() || 'No recent conversation available');",
		"        });",
		fix,
		"        setTimeout(function timeout() {",
		"            childProcess.kill();",
		"            resolve('Recent conversation unavailable');",
		"        }, 30000);",
		"    });",
		"}",
	)
}

func TestCorrectnessRequireChildProcessErrorListenerFiresOnPresence(t *testing.T) {
	t.Parallel()

	result := correctnessRequireChildProcessErrorListenerRun(t, correctnessRequireChildProcessErrorListenerPresence(""))
	correctnessRequireChildProcessErrorListenerExpect(t, result, []string{
		"nodeChildProcess.spawn('node', [path, 'pensieve', '--recent', '25k'], {\n            cwd: path,\n        })",
	})
}

func TestCorrectnessRequireChildProcessErrorListenerStaysSilentOnFixedPresence(t *testing.T) {
	t.Parallel()

	result := correctnessRequireChildProcessErrorListenerRun(t, correctnessRequireChildProcessErrorListenerPresence(
		"        childProcess.on('error', function handleError() { resolve('Recent conversation unavailable'); });",
	))
	correctnessRequireChildProcessErrorListenerExpect(t, result, nil)
}

func TestCorrectnessRequireChildProcessErrorListenerFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		// `modules/midjourney/MidjourneyApi.ts:543` and its eight siblings in ahra.
		{"a detached open, unreferenced", []string{
			"NodeChildProcess.spawn('open', [path], { detached: true, stdio: 'ignore' }).unref();",
		}, []string{"NodeChildProcess.spawn('open', [path], { detached: true, stdio: 'ignore' })"}},
		{"a call on its own", []string{
			"NodeChildProcess.spawn('afplay', [path]);",
		}, []string{"NodeChildProcess.spawn('afplay', [path])"}},
		{"stored and never referenced", []string{
			"const player = NodeChildProcess.spawn('afplay', [path], { stdio: 'ignore' });",
		}, []string{"NodeChildProcess.spawn('afplay', [path], { stdio: 'ignore' })"}},
		// `modules/os/lifecycle/AhraOsBootCommandLineInterface.ts:218`: the promise around the
		// child resolves on `'exit'`, and a missing `claude` never reaches it.
		{"awaited through a promise that listens only for exit", []string{
			"export async function launch(): Promise<number> {",
			"    return new Promise<number>(function(resolve) {",
			"        const child = NodeChildProcess.spawn('claude', commandArguments, { stdio: 'inherit' });",
			"        child.on('exit', function(code) {",
			"            resolve(code ?? 0);",
			"        });",
			"    });",
			"}",
		}, []string{"NodeChildProcess.spawn('claude', commandArguments, { stdio: 'inherit' })"}},
		{"a chain of other listeners", []string{
			"NodeChildProcess.spawn('ollama', ['serve']).on('close', done).once('exit', done);",
		}, []string{"NodeChildProcess.spawn('ollama', ['serve'])"}},
		{"a listener on a stream of the child, not on the child", []string{
			"NodeChildProcess.spawn('tail', ['-f', path]).stdout.on('error', ignore);",
		}, []string{"NodeChildProcess.spawn('tail', ['-f', path])"}},
		// `modules/figma/FigmaMcpLauncher.ts:11`.
		{"fork", []string{
			"const socketServer = NodeChildProcess.fork(path, [], { stdio: 'ignore' });",
			"socketServer.unref();",
		}, []string{"NodeChildProcess.fork(path, [], { stdio: 'ignore' })"}},
		{"a named import from the bare module name", []string{
			"import { spawn as start } from 'child_process';",
			"start('ollama', ['serve'], { stdio: 'inherit' });",
		}, []string{"start('ollama', ['serve'], { stdio: 'inherit' })"}},
		{"an event variable whose literal types exclude error", []string{
			"declare const eventName: 'close' | 'exit';",
			"const child = NodeChildProcess.spawn('gh', commandArguments);",
			"child.on(eventName, done);",
		}, []string{"NodeChildProcess.spawn('gh', commandArguments)"}},
		{"reads that let the child go nowhere", []string{
			"const child = NodeChildProcess.spawn('gh', commandArguments);",
			"if(child && !child.pid && child !== null && child instanceof NodeChildProcess.ChildProcess) child.kill();",
			"void child;",
			"export type ChildType = typeof child;",
			"console.log(typeof child === 'object', child == null);",
		}, []string{"NodeChildProcess.spawn('gh', commandArguments)"}},
		{"a removal chain, still the child", []string{
			"NodeChildProcess.spawn('gh', commandArguments).removeAllListeners('close').setMaxListeners(2);",
		}, []string{"NodeChildProcess.spawn('gh', commandArguments)"}},
		{"a let assigned twice, neither handled", []string{
			"let child: NodeChildProcess.ChildProcess;",
			"if(flag) child = NodeChildProcess.spawn('a');",
			"else child = NodeChildProcess.spawn('b');",
			"child.on('close', done);",
		}, []string{"NodeChildProcess.spawn('a')", "NodeChildProcess.spawn('b')"}},
		{"a listener on another variable of the same name", []string{
			"export function first(): void {",
			"    const child = NodeChildProcess.spawn('gh', commandArguments);",
			"    child.on('close', done);",
			"}",
			"export function second(): void {",
			"    const child = NodeChildProcess.spawn('gh', commandArguments);",
			"    child.on('error', ignore);",
			"}",
		}, []string{"NodeChildProcess.spawn('gh', commandArguments)"}},
		{"through parentheses, a non-null assertion and a cast", []string{
			"((NodeChildProcess.spawn('gh', commandArguments))! as NodeChildProcess.ChildProcess).unref();",
		}, []string{"NodeChildProcess.spawn('gh', commandArguments)"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sourceText := correctnessRequireChildProcessErrorListenerPrelude + correctnessRequireChildProcessErrorListenerSource(testCase.lines...)
			result := correctnessRequireChildProcessErrorListenerRun(t, sourceText)
			correctnessRequireChildProcessErrorListenerExpect(t, result, testCase.want)
		})
	}
}

func TestCorrectnessRequireChildProcessErrorListenerStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"a detached open with an error listener", []string{
			"NodeChildProcess.spawn('open', [path], { detached: true, stdio: 'ignore' }).on('error', ignore).unref();",
		}},
		// `modules/os/lifecycle/AhraOsBootCommandLineInterface.ts:62`.
		{"stored, listened to, and returned", []string{
			"export function playBootSound(): NodeChildProcess.ChildProcess | null {",
			"    const player = NodeChildProcess.spawn('afplay', [path], { stdio: 'ignore' });",
			"    player.on('error', function() {});",
			"    return player;",
			"}",
		}},
		{"once, addListener, prependListener and prependOnceListener", []string{
			"NodeChildProcess.spawn('a').once('error', ignore);",
			"NodeChildProcess.spawn('b').addListener('error', ignore);",
			"NodeChildProcess.spawn('c').prependListener('error', ignore);",
			"NodeChildProcess.spawn('d').prependOnceListener('error', ignore);",
		}},
		{"error at the end of a chain", []string{
			"NodeChildProcess.spawn('ollama', ['serve']).on('close', done).on('error', ignore);",
		}},
		{"a template literal event and an element access", []string{
			"NodeChildProcess.spawn('a').on(`error`, ignore);",
			"NodeChildProcess.spawn('b')['on']('error', ignore);",
		}},
		{"an event typed as any string", []string{
			"declare const eventName: string;",
			"NodeChildProcess.spawn('gh', commandArguments).on(eventName, ignore);",
		}},
		{"an event union that includes error", []string{
			"declare const eventName: 'error' | 'exit';",
			"NodeChildProcess.spawn('gh', commandArguments).on(eventName, ignore);",
		}},
		{"a listener attached inside a closure", []string{
			"const child = NodeChildProcess.spawn('gh', commandArguments);",
			"child.on('spawn', function() {",
			"    child.on('error', ignore);",
			"});",
		}},
		{"returned to the caller", []string{
			"export function start(): NodeChildProcess.ChildProcess {",
			"    return NodeChildProcess.spawn('claude', commandArguments);",
			"}",
			"export const startArrow = () => NodeChildProcess.spawn('claude', commandArguments);",
		}},
		{"handed to a function", []string{
			"export async function run(): Promise<void> {",
			"    const child = NodeChildProcess.spawn('tar', commandArguments);",
			"    await NodeEvents.once(child, 'exit');",
			"    wire(NodeChildProcess.spawn('zstd', commandArguments));",
			"}",
		}},
		{"stored in an object, an array and a property", []string{
			"const shorthand = NodeChildProcess.spawn('a');",
			"export const holder = { shorthand, named: NodeChildProcess.spawn('b') };",
			"children.push(NodeChildProcess.spawn('c'));",
			"const pushed = NodeChildProcess.spawn('d');",
			"children.push(pushed);",
			"export class Session {",
			"    process: NodeChildProcess.ChildProcess | null = null;",
			"    field = NodeChildProcess.spawn('e');",
			"    start(): void { this.process = NodeChildProcess.spawn('f'); }",
			"}",
		}},
		{"copied into another variable", []string{
			"const child = NodeChildProcess.spawn('gh', commandArguments);",
			"const alias = child;",
			"alias.on('close', done);",
		}},
		{"exported, by modifier and by specifier", []string{
			"export const exported = NodeChildProcess.spawn('a');",
			"const listed = NodeChildProcess.spawn('b');",
			"export { listed };",
		}},
		{"a conditional and a logical expression around the call", []string{
			"const chosen = flag ? NodeChildProcess.spawn('a') : NodeChildProcess.spawn('b');",
			"const fallback = flag && NodeChildProcess.spawn('c');",
		}},
		{"destructured", []string{
			"const { stdout } = NodeChildProcess.spawn('gh', commandArguments);",
		}},
		{"a listener method handed on uncalled", []string{
			"const child = NodeChildProcess.spawn('gh', commandArguments);",
			"const register = child.on.bind(child);",
		}},
		{"a let assigned twice, handled once", []string{
			"let child: NodeChildProcess.ChildProcess;",
			"if(flag) child = NodeChildProcess.spawn('a');",
			"else child = NodeChildProcess.spawn('b');",
			"child.on('error', ignore);",
		}},
		{"exec and execFile, which attach their own error listener", []string{
			"NodeChildProcess.exec('ls');",
			"NodeChildProcess.execFile('gh', commandArguments);",
			"const child = NodeChildProcess.execFile('gh', commandArguments);",
			"child.on('close', done);",
		}},
		{"spawnSync, which returns no emitter", []string{
			"NodeChildProcess.spawnSync('gh', commandArguments);",
		}},
		{"a local function named spawn", []string{
			"function spawn(command: string): { unref(): void } { return { unref() { console.log(command); } }; }",
			"spawn('open').unref();",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sourceText := correctnessRequireChildProcessErrorListenerPrelude + correctnessRequireChildProcessErrorListenerSource(testCase.lines...)
			result := correctnessRequireChildProcessErrorListenerRun(t, sourceText)
			correctnessRequireChildProcessErrorListenerExpect(t, result, nil)
		})
	}
}

// With no checker the rule registers nothing: a name match alone would report every `spawn`.
func TestCorrectnessRequireChildProcessErrorListenerDeclinesWithoutAChecker(t *testing.T) {
	t.Parallel()

	sourceText := correctnessRequireChildProcessErrorListenerPrelude + "NodeChildProcess.spawn('afplay', [path]);\n"
	result := rule_testing.Run(t, CorrectnessRequireChildProcessErrorListener, correctnessRequireChildProcessErrorListenerFile, sourceText)
	rule_testing.ExpectClean(t, result)
}
