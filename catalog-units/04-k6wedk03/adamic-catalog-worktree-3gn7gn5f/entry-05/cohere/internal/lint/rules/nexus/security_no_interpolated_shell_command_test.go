package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const securityNoInterpolatedShellCommandFile = "/repository/source/SecurityNoInterpolatedShellCommand.ts"

const securityNoInterpolatedShellCommandNodeFile = "/repository/source/node.d.ts"

func securityNoInterpolatedShellCommandSource(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

// securityNoInterpolatedShellCommandNodeTypes is the slice of `@types/node` the rule resolves
// against, in the layout of 26.2.0 (ahra's): the functions live in `declare module
// "node:child_process"` and `"child_process"` re-exports them, `exec` carries its `__promisify__`
// overload in a namespace of the same name, and `util.promisify` returns that overload's type.
var securityNoInterpolatedShellCommandNodeTypes = securityNoInterpolatedShellCommandSource(
	`declare module "node:child_process" {`,
	"    interface ExecSyncOptions { encoding?: string; timeout?: number; stdio?: unknown; shell?: string; maxBuffer?: number }",
	"    interface SpawnOptions { stdio?: unknown; shell?: boolean | string }",
	"    interface ChildProcess { pid?: number }",
	"    function exec(command: string, callback?: (error: Error | null, stdout: string) => void): ChildProcess;",
	"    namespace exec {",
	"        function __promisify__(command: string): Promise<{ stdout: string; stderr: string }>;",
	"    }",
	"    function execSync(command: string): Uint8Array;",
	"    function execSync(command: string, options: ExecSyncOptions): string;",
	"    function execFile(file: string, args: readonly string[], options?: SpawnOptions): ChildProcess;",
	"    function execFileSync(file: string, args?: readonly string[], options?: ExecSyncOptions): string;",
	"    function spawn(command: string, options?: SpawnOptions): ChildProcess;",
	"    function spawn(command: string, args: readonly string[], options?: SpawnOptions): ChildProcess;",
	"    function spawnSync(command: string, args: readonly string[], options?: SpawnOptions): { stdout: string };",
	"}",
	`declare module "child_process" {`,
	`    export * from "node:child_process";`,
	"}",
	`declare module "database" {`,
	"    function exec(sql: string): void;",
	"}",
	`declare module "node:util" {`,
	"    interface CustomPromisifySymbol<TCustom extends Function> extends Function { __promisify__: TCustom }",
	"    function promisify<TCustom extends Function>(fn: CustomPromisifySymbol<TCustom>): TCustom;",
	"}",
)

// securityNoInterpolatedShellCommandOlderNodeTypes is the layout `@types/node` used before it
// flipped the two names: the functions in `declare module "child_process"`, re-exported by
// `"node:child_process"`.
var securityNoInterpolatedShellCommandOlderNodeTypes = securityNoInterpolatedShellCommandSource(
	`declare module "child_process" {`,
	"    function execSync(command: string): Uint8Array;",
	"}",
	`declare module "node:child_process" {`,
	`    export * from "child_process";`,
	"}",
)

func securityNoInterpolatedShellCommandRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, SecurityNoInterpolatedShellCommand, map[string]string{
		securityNoInterpolatedShellCommandFile:     sourceText,
		securityNoInterpolatedShellCommandNodeFile: securityNoInterpolatedShellCommandNodeTypes,
	}, securityNoInterpolatedShellCommandFile)
}

func securityNoInterpolatedShellCommandExpectSpans(t *testing.T, result rule_testing.Result, sourceText string, wantSpans []string) {
	t.Helper()
	wantIds := make([]string, 0, len(wantSpans))
	for range wantSpans {
		wantIds = append(wantIds, securityNoInterpolatedShellCommandId)
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	for index, diagnostic := range result.Diagnostics {
		reported := strings.TrimSpace(sourceText[diagnostic.Range.Pos():diagnostic.Range.End()])
		if reported != wantSpans[index] {
			t.Fatalf("finding %d points at\n%s\nwant\n%s", index, reported, wantSpans[index])
		}
		if diagnostic.Message.Description != securityNoInterpolatedShellCommandMessage().Description {
			t.Fatalf("message is %q", diagnostic.Message.Description)
		}
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
}

// securityNoInterpolatedShellCommandContactsApi models `modules/apple/contacts/ContactsApi.ts` in
// ahra on 2026-10-02, trimmed to the search at `:79`: the search words become `LIKE` conditions, the
// conditions become the query, and the query and the database path are spliced into a shell
// command. The `runLine` argument is the call written at the site, before or after the fix.
func securityNoInterpolatedShellCommandContactsApi(runLine string) string {
	return securityNoInterpolatedShellCommandSource(
		"import * as NodeChildProcess from 'node:child_process';",
		"declare function findAllDatabases(): string[];",
		"export function searchContacts(searchText: string): string[] {",
		"    const found: string[] = [];",
		"    const searchWords = searchText.toLowerCase().split(/\\s+/).filter(Boolean);",
		"    for(const databasePath of findAllDatabases()) {",
		"        try {",
		"            const wordConditions = searchWords",
		"                .map(",
		"                    (word) =>",
		"                        `(LOWER(r.ZFIRSTNAME) LIKE '%${word}%' OR LOWER(r.ZLASTNAME) LIKE '%${word}%')`,",
		"                )",
		"                .join(' AND ');",
		"            const query = `",
		"        SELECT r.Z_PK, r.ZFIRSTNAME, r.ZLASTNAME",
		"        FROM ZABCDRECORD r",
		"        WHERE ${wordConditions}",
		"      `;",
		"            "+runLine,
		"                encoding: 'utf-8',",
		"                timeout: 5000,",
		"            });",
		"            found.push(result);",
		"        }",
		"        catch {",
		"            continue;",
		"        }",
		"    }",
		"    return found;",
		"}",
	)
}

// The real site, before the fix: the search words reach the shell through `query`, two `const`
// bindings away from the call, and the finding is on the command.
func TestSecurityNoInterpolatedShellCommandFiresOnContactsApi(t *testing.T) {
	t.Parallel()

	sourceText := securityNoInterpolatedShellCommandContactsApi("const result = NodeChildProcess.execSync(`sqlite3 \"${databasePath}\" \"${query}\"`, {")
	result := securityNoInterpolatedShellCommandRun(t, sourceText)
	securityNoInterpolatedShellCommandExpectSpans(t, result, sourceText, []string{"`sqlite3 \"${databasePath}\" \"${query}\"`"})
}

// The same search after the fix: `sqlite3` gets the path and the query as its own arguments, no
// shell reads either, and nothing fires.
func TestSecurityNoInterpolatedShellCommandStaysSilentOnFixedContactsApi(t *testing.T) {
	t.Parallel()

	result := securityNoInterpolatedShellCommandRun(t, securityNoInterpolatedShellCommandContactsApi("const result = NodeChildProcess.execFileSync('sqlite3', [databasePath, query], {"))
	rule_testing.ExpectClean(t, result)
}

func TestSecurityNoInterpolatedShellCommandFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		lines     []string
		wantSpans []string
	}{
		{"a string spliced into a template", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function open(url: string) {",
			"    NodeChildProcess.execSync(`open \"${url}\"`);",
			"}",
		}, []string{"`open \"${url}\"`"}},
		{"a string concatenated into the command", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function open(url: string) {",
			"    NodeChildProcess.execSync('open \"' + url + '\"');",
			"}",
		}, []string{"'open \"' + url + '\"'"}},
		// `MacOsApi.ts:134`'s shape: the command is built into a `const` and passed by name.
		{"a command built into a const", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function convert(input: string) {",
			"    const command = `ffmpeg -y -i \"${input}\" out.webp`;",
			"    NodeChildProcess.execSync(command);",
			"}",
		}, []string{"command"}},
		{"a named import under another name", []string{
			"import { execSync as run } from 'node:child_process';",
			"export function open(url: string) {",
			"    run(`open ${url}`);",
			"}",
		}, []string{"`open ${url}`"}},
		{"the module under its unprefixed name", []string{
			"import { exec } from 'child_process';",
			"export function log(branch: string) {",
			"    exec(`git log ${branch}`);",
			"}",
		}, []string{"`git log ${branch}`"}},
		{"an import-equals require", []string{
			"import ChildProcess = require('node:child_process');",
			"export function log(branch: string) {",
			"    ChildProcess.execSync(`git log ${branch}`);",
			"}",
		}, []string{"`git log ${branch}`"}},
		{"a promisified exec", []string{
			"import { exec } from 'node:child_process';",
			"import { promisify } from 'node:util';",
			"const execAsync = promisify(exec);",
			"export async function log(branch: string) {",
			"    return await execAsync(`git log ${branch}`);",
			"}",
		}, []string{"`git log ${branch}`"}},
		// `iMessageApi.ts:789`: the single-quote escape is correct, and the command still reaches a
		// shell (the doc comment gives the reasoning).
		{"a hand-escaped script", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function run(pythonScript: string) {",
			"    return NodeChildProcess.execSync(`python3 -c '${pythonScript.replace(/'/g, \"'\\\\''\")}'`, { encoding: 'utf-8' });",
			"}",
		}, []string{"`python3 -c '${pythonScript.replace(/'/g, \"'\\\\''\")}'`"}},
		// `ConversationsBackup.ts:78`: JSON quoting escapes `\"` but leaves `$` and backticks live
		// inside the double quotes it adds.
		{"a JSON-quoted path", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"const compressionLevel = '-19';",
			"export function archive(destinationPath: string) {",
			"    NodeChildProcess.execSync(",
			"        `tar -cf - . ` +",
			"            `| zstd ${compressionLevel} -q -o ${JSON.stringify(destinationPath)}`,",
			"        { shell: '/bin/zsh' },",
			"    );",
			"}",
		}, []string{"`tar -cf - . ` +\n            `| zstd ${compressionLevel} -q -o ${JSON.stringify(destinationPath)}`"}},
		// `PlanetScaleApi.ts:380`: the SQL sits in a quoted heredoc and is safe there, but the
		// database and branch are spliced unquoted into the opener line.
		{"values on a heredoc's opener line", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function pscaleShell(database: string, branch: string, sql: string) {",
			"    const command = `cat <<'EOSQL' | pscale shell ${database} ${branch}\\n${sql}\\nEOSQL`;",
			"    return NodeChildProcess.execSync(command, { encoding: 'utf-8', shell: '/bin/bash' });",
			"}",
		}, []string{"command"}},
		// An unquoted delimiter leaves the body expanded: `$(...)` in the SQL runs.
		{"a value in an unquoted heredoc's body", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function query(sql: string) {",
			"    return NodeChildProcess.execSync(`cat <<EOSQL | pscale shell phi main\\n${sql}\\nEOSQL`, { encoding: 'utf-8' });",
			"}",
		}, []string{"`cat <<EOSQL | pscale shell phi main\\n${sql}\\nEOSQL`"}},
		{"a value after a quoted heredoc has closed", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function query(sql: string, outputPath: string) {",
			"    return NodeChildProcess.execSync(`cat <<'EOSQL' > /tmp/query.sql\\n${sql}\\nEOSQL\\nmv /tmp/query.sql ${outputPath}`);",
			"}",
		}, []string{"`cat <<'EOSQL' > /tmp/query.sql\\n${sql}\\nEOSQL\\nmv /tmp/query.sql ${outputPath}`"}},
		// `<<-` strips leading tabs before matching the delimiter, so the indented line closes it.
		{"a value after a tab-stripped heredoc has closed", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function query(sql: string, outputPath: string) {",
			"    return NodeChildProcess.execSync(`cat <<-'EOSQL' > /tmp/query.sql\\n${sql}\\n\\tEOSQL\\nmv /tmp/query.sql ${outputPath}`);",
			"}",
		}, []string{"`cat <<-'EOSQL' > /tmp/query.sql\\n${sql}\\n\\tEOSQL\\nmv /tmp/query.sql ${outputPath}`"}},
		// The closing line is a `const` holding the delimiter, read as its literal text, so the
		// value after it is outside the body.
		{"a value after a heredoc closed by a literal constant", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"const delimiter = 'EOSQL';",
			"export function query(sql: string, outputPath: string) {",
			"    return NodeChildProcess.execSync(`cat <<'${delimiter}' > /tmp/query.sql\\n${sql}\\n${delimiter}\\nmv /tmp/query.sql ${outputPath}`);",
			"}",
		}, []string{"`cat <<'${delimiter}' > /tmp/query.sql\\n${sql}\\n${delimiter}\\nmv /tmp/query.sql ${outputPath}`"}},
		{"a value after a heredoc closed by a literal-typed parameter", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function query(delimiter: 'EOSQL', sql: string, outputPath: string) {",
			"    return NodeChildProcess.execSync(`cat <<'${delimiter}' > /tmp/query.sql\\n${sql}\\n${delimiter}\\nmv /tmp/query.sql ${outputPath}`);",
			"}",
		}, []string{"`cat <<'${delimiter}' > /tmp/query.sql\\n${sql}\\n${delimiter}\\nmv /tmp/query.sql ${outputPath}`"}},
		// A here-string's word is parsed by the shell like any other.
		{"a value in a here-string", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function count(text: string) {",
			"    return NodeChildProcess.execSync(`wc -c <<< \"${text}\"`, { encoding: 'utf-8' });",
			"}",
		}, []string{"`wc -c <<< \"${text}\"`"}},
		{"a string branch of a conditional", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function capture(windowId: string | null) {",
			"    NodeChildProcess.execSync(`screencapture ${windowId ? `-l ${windowId}` : '-x'} out.png`);",
			"}",
		}, []string{"`screencapture ${windowId ? `-l ${windowId}` : '-x'} out.png`"}},
		{"a conditional command", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function open(url: string, background: boolean) {",
			"    NodeChildProcess.execSync(background ? `open -g ${url}` : 'true');",
			"}",
		}, []string{"background ? `open -g ${url}` : 'true'"}},
		{"a type parameter constrained to string", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function open<Url extends string>(url: Url) {",
			"    NodeChildProcess.execSync(`open ${url}`);",
			"}",
		}, []string{"`open ${url}`"}},
		{"an any", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function open(url: any) {",
			"    NodeChildProcess.execSync(`open ${url}`);",
			"}",
		}, []string{"`open ${url}`"}},
		{"a template literal type", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function open(url: `https://${string}`) {",
			"    NodeChildProcess.execSync(`open ${url}`);",
			"}",
		}, []string{"`open ${url}`"}},
		{"spawn with shell true and a built command", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function log(branch: string) {",
			"    NodeChildProcess.spawn(`git log ${branch}`, { shell: true });",
			"}",
		}, []string{"`git log ${branch}`"}},
		// Under a truthy `shell`, node joins the arguments into the line, so a bare reference in
		// the list is spliced too.
		{"spawn with shell true and a string argument", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function log(branch: string) {",
			"    NodeChildProcess.spawn('git', ['log', branch], { stdio: 'inherit', shell: true });",
			"}",
		}, []string{"branch"}},
		{"spawnSync with a named shell", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function log(branch: string) {",
			"    NodeChildProcess.spawnSync('git', [`log ${branch}`], { shell: '/bin/zsh' });",
			"}",
		}, []string{"`log ${branch}`"}},
		{"execFileSync with shell true", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function log(branch: string) {",
			"    NodeChildProcess.execFile('git', ['log', branch], { shell: true });",
			"}",
		}, []string{"branch"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceText := securityNoInterpolatedShellCommandSource(testCase.lines...)
			result := securityNoInterpolatedShellCommandRun(t, sourceText)
			securityNoInterpolatedShellCommandExpectSpans(t, result, sourceText, testCase.wantSpans)
		})
	}
}

func TestSecurityNoInterpolatedShellCommandStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"an all-literal command", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function list() {",
			"    return NodeChildProcess.execSync('ps -Ao pid=,command=', { encoding: 'utf-8' });",
			"}",
		}},
		{"a template with no substitution", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function list() {",
			"    return NodeChildProcess.execSync(`ps -Ao pid=,command=`, { encoding: 'utf-8' });",
			"}",
		}},
		// `AhraOsWatchers.ts:720`: a process id is a number, and a number has no shell syntax.
		{"a number spliced in", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function environment(candidate: { pid: number }) {",
			"    return NodeChildProcess.execSync(`ps -Eww -o command= -p ${candidate.pid}`, { encoding: 'utf-8' });",
			"}",
		}},
		{"a number concatenated", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function stop(pid: number) {",
			"    NodeChildProcess.execSync('kill ' + pid);",
			"}",
		}},
		{"a literal union", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function git(mode: 'status' | 'diff', verbose: boolean) {",
			"    NodeChildProcess.execSync(`git ${mode} ${verbose}`);",
			"}",
		}},
		{"an enum member", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"enum Mode { Status = 'status', Diff = 'diff' }",
			"export function git(mode: Mode) {",
			"    NodeChildProcess.execSync(`git ${mode}`);",
			"}",
		}},
		{"a type parameter constrained to a literal union", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function git<Mode extends 'status' | 'diff'>(mode: Mode) {",
			"    NodeChildProcess.execSync(`git ${mode}`);",
			"}",
		}},
		// `PhiAnalyticsApi.ts:262`: both paths are `const` string literals.
		{"literal constants spliced in", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function trim() {",
			"    const rawMapPath = '/tmp/phi-subscriber-map-raw.png';",
			"    const mapPath = '/tmp/phi-subscriber-map.png';",
			"    NodeChildProcess.execSync(`convert \"${rawMapPath}\" -trim +repage \"${mapPath}\"`);",
			"}",
		}},
		// The checker types a nested template as `string`; its own pieces are a literal and a
		// number, so it is judged by those.
		{"a nested template built from literals", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function list(verbose: boolean, depth: number) {",
			"    const depthFlag = `-L ${depth}`;",
			"    NodeChildProcess.execSync(`tree ${verbose ? `-a ${depthFlag}` : ''} .`);",
			"}",
		}},
		// `i18n-conversion.ts:243`: the SQL sits in the body of a heredoc whose delimiter is
		// quoted, so the shell expands nothing in it, and the opener line is all literal.
		{"a value in a quoted heredoc's body", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"const pscaleDatabase = 'connected';",
			"const pscaleBranch = 'main';",
			"export function pscaleQuery(sql: string) {",
			"    const command = `cat <<'EOSQL' | pscale shell ${pscaleDatabase} ${pscaleBranch} --replica\\n${sql}\\nEOSQL`;",
			"    return NodeChildProcess.execSync(command, { encoding: 'utf-8', shell: '/bin/bash' });",
			"}",
		}},
		{"values in a double-quoted and a backslashed heredoc's body", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function feed(first: string, second: string) {",
			"    NodeChildProcess.execSync(`cat <<\"ONE\" <<\\\\TWO\\n${first}\\nONE\\n${second}\\nTWO`);",
			"}",
		}},
		{"a value in a tab-stripped heredoc's body", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function feed(sql: string) {",
			"    NodeChildProcess.execSync(`cat <<-'EOSQL'\\n\\t${sql}\\n\\tEOSQL`);",
			"}",
		}},
		// A `<<` the rule cannot read as a delimiter declines the command rather than guess.
		{"a heredoc delimiter it cannot read", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function feed(delimiter: string, sql: string) {",
			"    NodeChildProcess.execSync(`cat <<'${delimiter}'\\n${sql}\\n${delimiter}`);",
			"}",
		}},
		{"a literal union that opens a heredoc", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function feed(opener: \"<<'EOSQL'\" | \"<<-'EOSQL'\", sql: string) {",
			"    NodeChildProcess.execSync(`pscale shell phi main ${opener}\\n${sql}\\nEOSQL`);",
			"}",
		}},
		// The branch's body runs on past the conditional, over `${sql}`, and the rule cannot know
		// which branch is taken, so it declines the whole command.
		{"a conditional branch that opens a heredoc", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function feed(fromStandardInput: boolean, batch: number, sql: string) {",
			"    NodeChildProcess.execSync(`pscale shell phi main ${fromStandardInput ? `<<'EOSQL' # batch ${batch}` : `<<-'EOSQL'`}\\n${sql}\\nEOSQL`);",
			"}",
		}},
		// `PlanetScaleApi.ts:344`'s `run(command)`: an opaque command is not built at this call.
		{"an opaque command parameter", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function run(command: string) {",
			"    return NodeChildProcess.execSync(command, { encoding: 'utf-8' });",
			"}",
		}},
		// `MacOsApi.ts:128`: a `let` may be reassigned, so it is opaque too.
		{"a command held in a let", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function capture(path: string) {",
			"    let command = 'screencapture -x';",
			"    command = `screencapture -x \"${path}\"`;",
			"    NodeChildProcess.execSync(command);",
			"}",
		}},
		{"the fix: execFileSync with an argument list", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function query(databasePath: string, query: string) {",
			"    return NodeChildProcess.execFileSync('sqlite3', [databasePath, `${query};`], { encoding: 'utf-8' });",
			"}",
		}},
		// `FigmaMcpLauncher.ts:22`: a shell, but every argument is literal.
		{"spawn with shell true and literal arguments", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function launch() {",
			"    return NodeChildProcess.spawn('npx', ['-y', 'github:hyun1202/talk-to-figma-mcp'], { stdio: 'inherit', shell: true });",
			"}",
		}},
		{"spawn with no shell", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function log(branch: string) {",
			"    NodeChildProcess.spawn(`git-${branch}`, ['log', branch], { stdio: 'inherit' });",
			"}",
		}},
		{"spawn with shell false", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function log(branch: string) {",
			"    NodeChildProcess.spawn('git', ['log', branch], { shell: false });",
			"}",
		}},
		{"spawn with an empty shell", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function log(branch: string) {",
			"    NodeChildProcess.spawn('git', ['log', branch], { shell: '' });",
			"}",
		}},
		// A `boolean` may be false at run time, so the shell is not proven on.
		{"spawn with a boolean shell", []string{
			"import * as NodeChildProcess from 'node:child_process';",
			"export function log(branch: string, useShell: boolean) {",
			"    NodeChildProcess.spawn('git', ['log', branch], { shell: useShell });",
			"}",
		}},
		{"a local function named execSync", []string{
			"function execSync(command: string): string { return command; }",
			"export function log(branch: string) {",
			"    return execSync(`git log ${branch}`);",
			"}",
		}},
		{"a function named exec in another module", []string{
			"import { exec } from 'database';",
			"export function drop(table: string) {",
			"    exec(`DROP TABLE ${table}`);",
			"}",
		}},
		{"a method named exec on another object", []string{
			"declare const database: { exec(sql: string): void };",
			"export function drop(table: string) {",
			"    database.exec(`DROP TABLE ${table}`);",
			"}",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := securityNoInterpolatedShellCommandRun(t, securityNoInterpolatedShellCommandSource(testCase.lines...))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The older `@types/node` layout declares the functions under the unprefixed name; both import
// spellings resolve there.
func TestSecurityNoInterpolatedShellCommandResolvesTheOlderModuleLayout(t *testing.T) {
	t.Parallel()

	sourceText := securityNoInterpolatedShellCommandSource(
		"import * as NodeChildProcess from 'node:child_process';",
		"import * as ChildProcess from 'child_process';",
		"export function open(url: string) {",
		"    NodeChildProcess.execSync(`open ${url}`);",
		"    ChildProcess.execSync(`open -g ${url}`);",
		"}",
	)
	result := rule_testing.RunTypedFiles(t, SecurityNoInterpolatedShellCommand, map[string]string{
		securityNoInterpolatedShellCommandFile:     sourceText,
		securityNoInterpolatedShellCommandNodeFile: securityNoInterpolatedShellCommandOlderNodeTypes,
	}, securityNoInterpolatedShellCommandFile)
	securityNoInterpolatedShellCommandExpectSpans(t, result, sourceText, []string{"`open ${url}`", "`open -g ${url}`"})
}

// The rule declares the checker and declines a file without one, rather than matching `exec` by
// name.
func TestSecurityNoInterpolatedShellCommandDeclinesWithoutAChecker(t *testing.T) {
	t.Parallel()
	if !SecurityNoInterpolatedShellCommand.NeedsTypeChecker {
		t.Fatal("the rule must declare NeedsTypeChecker")
	}
	if listeners := SecurityNoInterpolatedShellCommand.Run(rule.Context{}, nil); listeners != nil {
		t.Errorf("with no checker the rule must register no listeners, got %d", len(listeners))
	}
}
