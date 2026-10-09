package nexus

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoProcessExitAfterOutputFile = "/repository/source/CorrectnessNoProcessExitAfterOutput.ts"

const correctnessNoProcessExitAfterOutputNodeFile = "/repository/source/node.d.ts"

const correctnessNoProcessExitAfterOutputWebConsoleFile = "/repository/source/web-console.d.ts"

func correctnessNoProcessExitAfterOutputLines(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

// correctnessNoProcessExitAfterOutputNodeTypes is the slice of `@types/node` the rule resolves
// against, in the layout of 26.2.0 (ahra's): `process.d.ts` declares the global `process` and the
// `NodeJS.Process` interface inside `declare module "node:process" { global { ... } }` and exports it
// with `export =`, `globals.d.ts` declares the global again at the top of a script, and
// `console.d.ts` keeps its `Console` interface in a namespace of `node:console`.
var correctnessNoProcessExitAfterOutputNodeTypes = correctnessNoProcessExitAfterOutputLines(
	"declare var process: NodeJS.Process;",
	`declare module "node:process" {`,
	"    global {",
	"        var process: NodeJS.Process;",
	"        namespace NodeJS {",
	"            interface WritableStream { write(chunk: string | Uint8Array, callback?: (error?: Error | null) => void): boolean }",
	"            interface WriteStream extends WritableStream { columns: number }",
	"            interface ReadStream { isTTY?: boolean; setRawMode(mode: boolean): this; resume(): this; pause(): this }",
	"            interface Process {",
	"                stdout: WriteStream & { fd: 1 };",
	"                stderr: WriteStream & { fd: 2 };",
	"                stdin: ReadStream & { fd: 0 };",
	"                exitCode: number | string | null | undefined;",
	"                exit(code?: number | string | null): never;",
	"                on(event: string, listener: (...args: any[]) => void): this;",
	"            }",
	"        }",
	"    }",
	"    export = process;",
	"}",
	`declare module "process" {`,
	`    import process = require("node:process");`,
	"    export = process;",
	"}",
	`declare module "node:console" {`,
	"    namespace console {",
	"        interface Console {",
	"            log(...data: any[]): void;",
	"            info(...data: any[]): void;",
	"            debug(...data: any[]): void;",
	"            warn(...data: any[]): void;",
	"            error(...data: any[]): void;",
	"            trace(...data: any[]): void;",
	"            table(tabularData?: any, properties?: string[]): void;",
	"            dir(item?: any): void;",
	"            dirxml(...data: any[]): void;",
	"            group(...data: any[]): void;",
	"            time(label?: string): void;",
	"        }",
	"    }",
	"    var console: console.Console;",
	"    export = console;",
	"}",
)

// correctnessNoProcessExitAfterOutputWebConsole is `@types/node`'s `web-globals/console.d.ts`: a
// module file that puts the global `console` in `declare global`.
var correctnessNoProcessExitAfterOutputWebConsole = correctnessNoProcessExitAfterOutputLines(
	"export {};",
	`import * as console from "node:console";`,
	"declare global {",
	"    interface Console extends console.Console {}",
	"    var console: Console;",
	"}",
)

// correctnessNoProcessExitAfterOutputPrelude declares every name the cases use, so the checker
// resolves each one and no verdict rests on an unresolved identifier. lib.dom adds its own ambient
// `console`, as it does in the trees.
var correctnessNoProcessExitAfterOutputPrelude = correctnessNoProcessExitAfterOutputLines(
	`/// <reference lib="dom" />`,
	"import * as NodeProcess from 'node:process';",
	"declare const flag: boolean;",
	"declare const output: string;",
	"declare const rows: string[];",
	"declare const report: { authorized: boolean; lines: string[] };",
	"declare const result: { lines: string[]; exitCode: number; allOk: boolean; sources: { ok: boolean; label: string }[] };",
	"declare const fileConsole: Console;",
	"declare const streams: { stdout: { write(text: string): void } };",
	"declare const runner: { exit(code: number): void };",
	"declare function run(): Promise<void>;",
	"declare function use(value: unknown): void;",
	"declare function printReport(): void;",
	"declare function runLinkCommand(commandArguments: string[]): Promise<number>;",
	"",
)

func correctnessNoProcessExitAfterOutputSource(lines ...string) string {
	return correctnessNoProcessExitAfterOutputPrelude + strings.Join(lines, "\n") + "\n"
}

func correctnessNoProcessExitAfterOutputRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessNoProcessExitAfterOutput, map[string]string{
		correctnessNoProcessExitAfterOutputFile:           sourceText,
		correctnessNoProcessExitAfterOutputNodeFile:       correctnessNoProcessExitAfterOutputNodeTypes,
		correctnessNoProcessExitAfterOutputWebConsoleFile: correctnessNoProcessExitAfterOutputWebConsole,
	}, correctnessNoProcessExitAfterOutputFile)
}

// correctnessNoProcessExitAfterOutputExpect asserts the findings, in source order, by the text each
// one points at. Every finding carries the one message and no fix.
func correctnessNoProcessExitAfterOutputExpect(t *testing.T, result rule_testing.Result, want []string) {
	t.Helper()
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	wantIds := make([]string, len(want))
	for index := range want {
		wantIds[index] = correctnessNoProcessExitAfterOutputId
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	text := result.SourceFile.Text()
	var got []string
	for _, diagnostic := range diagnostics {
		got = append(got, text[diagnostic.Range.Pos():diagnostic.Range.End()])
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("findings are %q, want %q", got, want)
	}
}

// correctnessNoProcessExitAfterOutputStatement models `runStatement` in ahra's
// `modules/finance/FinanceConnectionsCommandLineInterface.ts:149` on 2026-10-03: every line of the
// statement printed in a loop, then the exit. `fixed` sets the exit code and returns.
func correctnessNoProcessExitAfterOutputStatement(fixed bool) string {
	exit := "    process.exit(result.exitCode);"
	if fixed {
		exit = "    process.exitCode = result.exitCode;\n    return;"
	}
	return correctnessNoProcessExitAfterOutputSource(
		"export function runStatement(commandArguments: string[]): void {",
		"    use(commandArguments);",
		"    for(const line of result.lines) console.log(line);",
		exit,
		"}",
	)
}

// correctnessNoProcessExitAfterOutputSync models `runSync` in the same file, `:90`: a per-source
// line to stdout or stderr, a note when there is nothing, then the exit.
func correctnessNoProcessExitAfterOutputSync(fixed bool) string {
	exit := "    process.exit(result.allOk ? 0 : 1);"
	if fixed {
		exit = "    process.exitCode = result.allOk ? 0 : 1;"
	}
	return correctnessNoProcessExitAfterOutputSource(
		"export async function runSync(commandArguments: string[]): Promise<void> {",
		"    use(commandArguments);",
		"    for(const source of result.sources) {",
		"        if(source.ok) {",
		"            console.log(`  ${source.label}: synced`);",
		"        }",
		"        else {",
		"            console.error(`RETRYABLE: ${source.label}`.trimEnd());",
		"        }",
		"    }",
		"    if(result.sources.length === 0) {",
		"        console.log('  No connected sources.');",
		"    }",
		exit,
		"}",
	)
}

// correctnessNoProcessExitAfterOutputCardBalances models `runQuickBooksCardBalances` in the same
// file, `:186`: an early report and exit when QuickBooks is not linked, then the full report and a
// second exit.
func correctnessNoProcessExitAfterOutputCardBalances(fixed bool) string {
	early, late := "        process.exit(0);", "    process.exit(0);"
	if fixed {
		early, late = "        return;", ""
	}
	return correctnessNoProcessExitAfterOutputSource(
		"export async function runQuickBooksCardBalances(): Promise<void> {",
		"    console.log('');",
		"    if(!report.authorized) {",
		"        console.log('  QuickBooks NOT linked. No balance captured, nothing written.');",
		"        console.log('');",
		early,
		"    }",
		"    for(const line of report.lines) console.log(`    ${line}`);",
		"    console.log('');",
		late,
		"}",
	)
}

// correctnessNoProcessExitAfterOutputGodword models `godwordRequireInteractiveTerminal` in ahra's
// `modules/godword/GodwordEntropy.ts:162`, which reaches the process through `import * as
// NodeProcess from 'node:process'`.
func correctnessNoProcessExitAfterOutputGodword(fixed bool) string {
	exit := "        NodeProcess.exit(1);"
	if fixed {
		exit = "        NodeProcess.exitCode = 1;\n        return;"
	}
	return correctnessNoProcessExitAfterOutputSource(
		"export function godwordRequireInteractiveTerminal(): void {",
		"    if(!NodeProcess.stdin.isTTY) {",
		"        console.error('godword needs a human at a keyboard.');",
		"        console.error('Run it in an interactive terminal.');",
		exit,
		"    }",
		"}",
	)
}

func TestCorrectnessNoProcessExitAfterOutputRealSites(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name   string
		source func(fixed bool) string
		want   []string
	}{
		{"runStatement", correctnessNoProcessExitAfterOutputStatement, []string{"process.exit(result.exitCode)"}},
		{"runSync", correctnessNoProcessExitAfterOutputSync, []string{"process.exit(result.allOk ? 0 : 1)"}},
		{"runQuickBooksCardBalances", correctnessNoProcessExitAfterOutputCardBalances, []string{"process.exit(0)", "process.exit(0)"}},
		{"godwordRequireInteractiveTerminal", correctnessNoProcessExitAfterOutputGodword, []string{"NodeProcess.exit(1)"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name+" as it stands", func(t *testing.T) {
			t.Parallel()
			correctnessNoProcessExitAfterOutputExpect(t, correctnessNoProcessExitAfterOutputRun(t, testCase.source(false)), testCase.want)
		})
		t.Run(testCase.name+" fixed", func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, correctnessNoProcessExitAfterOutputRun(t, testCase.source(true)))
		})
	}
}

func TestCorrectnessNoProcessExitAfterOutputFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"an error printed in a catch, then the exit", []string{
			"export async function main(): Promise<void> {",
			"    try {",
			"        await run();",
			"    }",
			"    catch(error) {",
			"        console.error('failed:', error);",
			"        process.exit(1);",
			"    }",
			"}",
		}, []string{"process.exit(1)"}},
		{"a write before the try, an exit in its catch", []string{
			"export async function main(): Promise<void> {",
			"    console.log('starting');",
			"    try {",
			"        await run();",
			"    }",
			"    catch {",
			"        process.exit(1);",
			"    }",
			"}",
		}, []string{"process.exit(1)"}},
		{"an exit at the top of a loop whose last turn printed", []string{
			"export function main(): void {",
			"    for(const row of rows) {",
			"        if(row === '') process.exit(0);",
			"        console.log(row);",
			"    }",
			"}",
		}, []string{"process.exit(0)"}},
		{"stdout written directly", []string{
			"export function main(): void {",
			"    process.stdout.write(output);",
			"    process.exit(0);",
			"}",
		}, []string{"process.exit(0)"}},
		{"stderr written directly", []string{
			"export function main(): void {",
			"    process.stderr.write(output);",
			"    process.exit(2);",
			"}",
		}, []string{"process.exit(2)"}},
		{"stdout written through the namespace import, then the exit (GodwordEntropy.ts:146)", []string{
			"export function onKey(data: Uint8Array): void {",
			"    if(data.length === 1 && data[0] === 0x03) {",
			"        NodeProcess.stdout.write('\\n');",
			"        NodeProcess.exit(130);",
			"    }",
			"}",
		}, []string{"NodeProcess.exit(130)"}},
		{"members imported by name", []string{
			"import { exit, stdout } from 'node:process';",
			"export function main(): void {",
			"    stdout.write(output);",
			"    exit(0);",
			"}",
		}, []string{"exit(0)"}},
		{"members of the bare process module", []string{
			"import Process = require('process');",
			"export function main(): void {",
			"    console.warn(output);",
			"    Process.exit(1);",
			"}",
		}, []string{"Process.exit(1)"}},
		{"every console method that writes", []string{
			"export function table(): void { console.table(rows); process.exit(0); }",
			"export function dir(): void { console.dir(rows); process.exit(0); }",
			"export function info(): void { console.info(output); process.exit(0); }",
			"export function debug(): void { console.debug(output); process.exit(0); }",
			"export function trace(): void { console.trace(output); process.exit(0); }",
			"export function dirxml(): void { console.dirxml(output); process.exit(0); }",
		}, []string{"process.exit(0)", "process.exit(0)", "process.exit(0)", "process.exit(0)", "process.exit(0)", "process.exit(0)"}},
		{"console reached by a quoted member", []string{
			"export function main(): void {",
			"    console['log'](output);",
			"    process.exit(0);",
			"}",
		}, []string{"process.exit(0)"}},
		{"module top level", []string{
			"console.log(output);",
			"process.exit(0);",
		}, []string{"process.exit(0)"}},
		{"a write in a try, the exit in its finally", []string{
			"export function main(): void {",
			"    try {",
			"        console.log(output);",
			"    }",
			"    finally {",
			"        process.exit(0);",
			"    }",
			"}",
		}, []string{"process.exit(0)"}},
		{"a case that prints and falls through to the exit", []string{
			"export function main(command: string): void {",
			"    switch(command) {",
			"        case 'help':",
			"            console.log('usage');",
			"        case 'quit':",
			"            process.exit(0);",
			"    }",
			"}",
		}, []string{"process.exit(0)"}},
		{"the last-resort catch on a promise", []string{
			"run().catch(function(error) {",
			"    console.error(error);",
			"    process.exit(1);",
			"});",
		}, []string{"process.exit(1)"}},
		{"a write before the try survives the catch that forgets the one inside it", []string{
			"export async function main(): Promise<void> {",
			"    console.log('starting');",
			"    try {",
			"        console.log('running');",
			"        await run();",
			"    }",
			"    catch {",
			"        process.exit(1);",
			"    }",
			"}",
		}, []string{"process.exit(1)"}},
		{"a write in an earlier try survives the catch of a later one", []string{
			"export async function main(): Promise<void> {",
			"    try {",
			"        console.log('first');",
			"        await run();",
			"    }",
			"    catch {",
			"        use(0);",
			"    }",
			"    try {",
			"        console.log('second');",
			"        await run();",
			"    }",
			"    catch {",
			"        process.exit(1);",
			"    }",
			"}",
		}, []string{"process.exit(1)"}},
		{"a write in an inner catch reaches the exit after the outer try", []string{
			"export async function main(): Promise<void> {",
			"    try {",
			"        try {",
			"            await run();",
			"        }",
			"        catch(error) {",
			"            console.error(error);",
			"        }",
			"    }",
			"    catch {",
			"        use(0);",
			"    }",
			"    process.exit(1);",
			"}",
		}, []string{"process.exit(1)"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			correctnessNoProcessExitAfterOutputExpect(t,
				correctnessNoProcessExitAfterOutputRun(t, correctnessNoProcessExitAfterOutputSource(testCase.lines...)), testCase.want)
		})
	}
}

func TestCorrectnessNoProcessExitAfterOutputStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"the exit comes before any write", []string{
			"export function main(): void {",
			"    if(!flag) process.exit(1);",
			"    console.log('ready');",
			"}",
		}},
		{"the write and the exit on different branches", []string{
			"export function main(): void {",
			"    if(flag) {",
			"        process.exit(1);",
			"    }",
			"    else {",
			"        console.log('ready');",
			"    }",
			"}",
		}},
		{"the exit in the write's own callback, the correct form", []string{
			"export function main(): void {",
			"    process.stdout.write(output, function() {",
			"        process.exit(0);",
			"    });",
			"}",
		}},
		{"the exit code set and the function left", []string{
			"export function main(): void {",
			"    console.log(output);",
			"    process.exitCode = 1;",
			"    return;",
			"}",
		}},
		{"the write in a callee (runLink, FinanceConnectionsCommandLineInterface.ts:79)", []string{
			"export async function runLink(commandArguments: string[]): Promise<void> {",
			"    process.exit(await runLinkCommand(commandArguments));",
			"}",
			"export function main(): void {",
			"    printReport();",
			"    process.exit(0);",
			"}",
		}},
		{"the write in a callback the function passes on", []string{
			"export function main(): void {",
			"    rows.forEach((row) => console.log(row));",
			"    process.exit(0);",
			"}",
		}},
		{"a write inside the try, the exit in its catch", []string{
			"export async function main(): Promise<void> {",
			"    try {",
			"        await run();",
			"        console.log('done');",
			"    }",
			"    catch {",
			"        process.exit(1);",
			"    }",
			"}",
		}},
		{"an exit in the catch binding, which runs before the catch block", []string{
			"export function main(): void {",
			"    try {",
			"        console.log(output);",
			"    }",
			"    catch({ code = process.exit(1) }) {",
			"        use(code);",
			"    }",
			"}",
		}},
		{"a write laid out after an earlier exit", []string{
			"export function main(): void {",
			"    process.exit(1);",
			"    console.log(output);",
			"    process.exit(0);",
			"}",
		}},
		{"an exit inside the write's own arguments", []string{
			"export function main(): void {",
			"    console.log(flag ? process.exit(1) : output);",
			"}",
		}},
		{"a local console", []string{
			"export function main(): void {",
			"    const console = { log(text: string): void { use(text); } };",
			"    console.log(output);",
			"    process.exit(0);",
			"}",
		}},
		{"a Console that is not the global one", []string{
			"export function main(): void {",
			"    fileConsole.log(output);",
			"    process.exit(0);",
			"}",
		}},
		{"a console method that does not write", []string{
			"export function main(): void {",
			"    console.time('run');",
			"    console.group();",
			"    process.exit(0);",
			"}",
		}},
		{"another stream's stdout", []string{
			"export function main(): void {",
			"    streams.stdout.write(output);",
			"    process.exit(0);",
			"}",
		}},
		{"another object's exit", []string{
			"export function main(): void {",
			"    console.log(output);",
			"    runner.exit(0);",
			"}",
		}},
		{"a local function named exit", []string{
			"function exit(code: number): void { use(code); }",
			"export function main(): void {",
			"    console.log(output);",
			"    exit(0);",
			"}",
		}},
		{"the exit in a callback, the write outside it", []string{
			"export function main(): void {",
			"    console.log('listening');",
			"    process.on('SIGINT', function() {",
			"        process.exit(130);",
			"    });",
			"}",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			rule_testing.ExpectClean(t, correctnessNoProcessExitAfterOutputRun(t, correctnessNoProcessExitAfterOutputSource(testCase.lines...)))
		})
	}
}

const correctnessNoProcessExitAfterOutputHelpFile = "/repository/source/Help.ts"

const correctnessNoProcessExitAfterOutputScriptFile = "/repository/source/ScriptHelp.ts"

// correctnessNoProcessExitAfterOutputHelpModule is a module the cases import a writer from, statically
// and through `import()`, as BackupCommandLineInterface.ts reaches reportFreshness.
var correctnessNoProcessExitAfterOutputHelpModule = correctnessNoProcessExitAfterOutputLines(
	"export function showModuleHelp(): void { console.log('usage'); }",
	"export function reportFreshness(): void { console.error('stale'); }",
)

// correctnessNoProcessExitAfterOutputScript is a global script, whose function no import reaches.
var correctnessNoProcessExitAfterOutputScript = correctnessNoProcessExitAfterOutputLines(
	"function showScriptHelp(): void { console.log('usage'); }",
)

// correctnessNoProcessExitAfterOutputRunWithHelpers runs the cases with the two helper files beside
// them, under a tsconfig whose module detection is `auto`. The harness's default forces every file to
// be a module, which would make the script a module too and leave its function unreachable for a
// reason other than the one the script row tests.
func correctnessNoProcessExitAfterOutputRunWithHelpers(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFilesWithSetup(t, CorrectnessNoProcessExitAfterOutput, map[string]string{
		correctnessNoProcessExitAfterOutputFile:           sourceText,
		correctnessNoProcessExitAfterOutputNodeFile:       correctnessNoProcessExitAfterOutputNodeTypes,
		correctnessNoProcessExitAfterOutputWebConsoleFile: correctnessNoProcessExitAfterOutputWebConsole,
		correctnessNoProcessExitAfterOutputHelpFile:       correctnessNoProcessExitAfterOutputHelpModule,
		correctnessNoProcessExitAfterOutputScriptFile:     correctnessNoProcessExitAfterOutputScript,
	}, correctnessNoProcessExitAfterOutputFile, func(directory string) {
		tsconfig := `{"compilerOptions": {"strict": true, "target": "ES2022", "lib": ["ES2022"], ` +
			`"moduleDetection": "auto", "types": []}, "include": ["**/*.ts"]}`
		if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(tsconfig), 0o644); err != nil {
			t.Fatalf("writing the tsconfig: %v", err)
		}
	})
}

// TestCorrectnessNoProcessExitAfterOutputReadsOneCallee covers #4pyyvfs: a call counts as a write
// when its callee's own body writes, one level down, where the body can be read and surely runs
// before the call returns. The two hazards the rule missed are modelled first: `showHelp()` then
// `exit(0)` in AhraCommandLineInterface.ts, and `reportFreshness(...)` reached through `import()` in
// BackupCommandLineInterface.ts.
func TestCorrectnessNoProcessExitAfterOutputReadsOneCallee(t *testing.T) {
	t.Parallel()

	fires := []struct {
		name  string
		lines []string
	}{
		{"a helper in this file, as showHelp", []string{
			"function showHelp(): void { console.log('usage'); }",
			"export function main(): void { showHelp(); process.exit(0); }",
		}},
		{"an arrow whose expression body writes", []string{
			"const showUsage = (): void => console.log('usage');",
			"export function main(): void { showUsage(); process.exit(0); }",
		}},
		{"a static method", []string{
			"class Help { static show(): void { process.stderr.write('usage'); } }",
			"export function main(): void { Help.show(); process.exit(1); }",
		}},
		{"an async helper awaited", []string{
			"async function report(): Promise<void> { console.log('report'); }",
			"export async function main(): Promise<void> { await report(); process.exit(0); }",
		}},
		{"a helper imported from a module", []string{
			"import { showModuleHelp } from './Help';",
			"export function main(): void { showModuleHelp(); process.exit(0); }",
		}},
		{"a helper reached through import(), as reportFreshness", []string{
			"export async function main(): Promise<void> {",
			"    const { reportFreshness } = await import('./Help');",
			"    reportFreshness();",
			"    process.exit(1);",
			"}",
		}},
	}
	for _, testCase := range fires {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			exit := "process.exit(0)"
			if strings.Contains(strings.Join(testCase.lines, "\n"), "process.exit(1)") {
				exit = "process.exit(1)"
			}
			correctnessNoProcessExitAfterOutputExpect(t,
				correctnessNoProcessExitAfterOutputRunWithHelpers(t, correctnessNoProcessExitAfterOutputSource(testCase.lines...)),
				[]string{exit})
		})
	}

	silent := []struct {
		name  string
		lines []string
		want  []string
	}{
		{"a helper that prints only through another helper, two levels down", []string{
			"function printUsage(): void { console.log('usage'); }",
			"function showHelp(): void { printUsage(); }",
			"export function main(): void { showHelp(); process.exit(0); }",
		}, nil},
		{"an async helper not awaited", []string{
			"async function report(): Promise<void> { console.log('report'); }",
			"export function main(): void { void report(); process.exit(0); }",
		}, nil},
		{"a generator, whose body runs only when iterated", []string{
			"function* lines(): Generator<string> { console.log('first'); yield 'a'; }",
			"export function main(): void { lines(); process.exit(0); }",
		}, nil},
		{"a helper declared never", []string{
			"function fail(message: string): never { console.error(message); throw new Error(message); }",
			"export function main(): void { if(flag) fail('bad'); process.exit(0); }",
		}, nil},
		// The helper's own exit is judged in the helper and reported there; the caller's exit after it
		// may never run, so the call is not read as a write.
		{"a helper that exits", []string{
			"function bail(): void { console.error('bad'); process.exit(1); }",
			"export function main(): void { bail(); process.exit(2); }",
		}, []string{"process.exit(1)"}},
		{"an overloaded helper, whose resolved declaration has no body", []string{
			"function show(value: string): void;",
			"function show(value: number): void;",
			"function show(value: unknown): void { console.log(value); }",
			"export function main(): void { show('usage'); process.exit(0); }",
		}, nil},
		{"a helper in a global script, which no import reaches", []string{
			"export function main(): void { showScriptHelp(); process.exit(0); }",
		}, nil},
		{"a declared helper, with no body to read", []string{
			"export function main(): void { printReport(); process.exit(0); }",
		}, nil},
	}
	for _, testCase := range silent {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			correctnessNoProcessExitAfterOutputExpect(t,
				correctnessNoProcessExitAfterOutputRunWithHelpers(t, correctnessNoProcessExitAfterOutputSource(testCase.lines...)),
				testCase.want)
		})
	}
}
