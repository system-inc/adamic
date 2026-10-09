package nexus

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const (
	correctnessRequireBlockingStandardStreamsNexusFile       = "/repository/libraries/nexus/source/system/StandardStreams.ts"
	correctnessRequireBlockingStandardStreamsCommandLineFile = "/repository/libraries/nexus/source/command-line/CommandLineInterface.ts"
	correctnessRequireBlockingStandardStreamsNodeFile        = "/repository/types/node.d.ts"
	correctnessRequireBlockingStandardStreamsWebConsoleFile  = "/repository/types/web-console.d.ts"
	correctnessRequireBlockingStandardStreamsPreludeFile     = "/repository/types/prelude.d.ts"

	// correctnessRequireBlockingStandardStreamsSubject is where every case's entry sits, two levels
	// below the root as ahra's modules do, so its imports read `../../libraries/nexus/...`.
	correctnessRequireBlockingStandardStreamsSubject = "/repository/modules/subject/Subject.ts"

	correctnessRequireBlockingStandardStreamsImportBlock = "import { blockStandardStreams } from '../../libraries/nexus/source/system/StandardStreams';"
	correctnessRequireBlockingStandardStreamsImportRun   = "import { CommandLineInterface, runCommandLineInterface } from '../../libraries/nexus/source/command-line/CommandLineInterface';"
)

// correctnessRequireBlockingStandardStreamsNexus is Nexus's `source/system/StandardStreams.ts`, its
// body reduced to nothing the rule reads: the declaration is what it resolves to.
var correctnessRequireBlockingStandardStreamsNexus = correctnessNoProcessExitAfterOutputLines(
	"export function blockStandardStreams(): void {",
	"    Reflect.get(process.stdout, '_handle');",
	"}",
)

// correctnessRequireBlockingStandardStreamsCommandLine is Nexus's `runCommandLineInterface` as it
// stands in ahra on 2026-10-03: it blocks, then runs the command with a last-resort exit.
var correctnessRequireBlockingStandardStreamsCommandLine = correctnessNoProcessExitAfterOutputLines(
	"import { blockStandardStreams } from '../system/StandardStreams';",
	"export abstract class CommandLineInterface {",
	"    abstract run(): Promise<void>;",
	"}",
	"export function runCommandLineInterface(commandLineInterface: CommandLineInterface): void {",
	"    blockStandardStreams();",
	"    commandLineInterface.run().catch(function(error: unknown) {",
	"        console.error('Error:', error instanceof Error ? error.message : String(error));",
	"        process.exit(1);",
	"    });",
	"}",
)

// correctnessRequireBlockingStandardStreamsPrelude declares every other name the cases use, as
// globals in a declaration file the way a library's would be, so the checker resolves each one and
// none of them is code an entry could block through.
var correctnessRequireBlockingStandardStreamsPrelude = correctnessNoProcessExitAfterOutputLines(
	"declare const argumentsList: string[];",
	"declare const output: string;",
	"declare const flag: boolean;",
	"declare const results: unknown;",
	"declare const socketServer: { on(event: string, listener: (error: Error) => void): void };",
	"declare function run(): Promise<void>;",
	"declare function use(value: unknown): void;",
	"declare function onEvent(listener: () => void): void;",
	"declare function sendResults(value: unknown): void;",
)

// correctnessRequireBlockingStandardStreamsCase is one program: the subject's lines, and any other
// files beside the shared ones.
type correctnessRequireBlockingStandardStreamsCase struct {
	name    string
	shebang bool
	imports []string
	lines   []string
	others  map[string]string
}

func (testCase correctnessRequireBlockingStandardStreamsCase) files() map[string]string {
	var subject []string
	if testCase.shebang {
		subject = append(subject, "#!/usr/bin/env tsx")
	}
	subject = append(subject, testCase.imports...)
	subject = append(subject, testCase.lines...)
	files := map[string]string{
		correctnessRequireBlockingStandardStreamsSubject:         correctnessNoProcessExitAfterOutputLines(subject...),
		correctnessRequireBlockingStandardStreamsNexusFile:       correctnessRequireBlockingStandardStreamsNexus,
		correctnessRequireBlockingStandardStreamsCommandLineFile: correctnessRequireBlockingStandardStreamsCommandLine,
		correctnessRequireBlockingStandardStreamsNodeFile:        correctnessNoProcessExitAfterOutputNodeTypes,
		correctnessRequireBlockingStandardStreamsWebConsoleFile:  correctnessNoProcessExitAfterOutputWebConsole,
		correctnessRequireBlockingStandardStreamsPreludeFile:     correctnessRequireBlockingStandardStreamsPrelude,
	}
	for name, contents := range testCase.others {
		files[name] = contents
	}
	return files
}

func correctnessRequireBlockingStandardStreamsRun(t *testing.T, testCase correctnessRequireBlockingStandardStreamsCase) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessRequireBlockingStandardStreams, testCase.files(), correctnessRequireBlockingStandardStreamsSubject)
}

// correctnessRequireBlockingStandardStreamsExpect asserts the one finding an entry gets: at the text
// it points at, with a message holding every phrase in contains. An empty want asserts silence.
func correctnessRequireBlockingStandardStreamsExpect(t *testing.T, result rule_testing.Result, want string, contains ...string) {
	t.Helper()
	if want == "" {
		rule_testing.ExpectClean(t, result)
		return
	}
	rule_testing.ExpectFindings(t, result, correctnessRequireBlockingStandardStreamsId)
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	diagnostic := diagnostics[0]
	text := result.SourceFile.Text()
	if got := strings.TrimSpace(text[diagnostic.Range.Pos():diagnostic.Range.End()]); got != want {
		t.Fatalf("the finding is at %q, want %q", got, want)
	}
	if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
		t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
	}
	for _, phrase := range contains {
		if !strings.Contains(diagnostic.Message.Description, phrase) {
			t.Fatalf("the message %q does not say %q", diagnostic.Message.Description, phrase)
		}
	}
	if strings.Contains(diagnostic.Message.Description, "—") {
		t.Fatalf("the message holds an em-dash: %q", diagnostic.Message.Description)
	}
}

// correctnessRequireBlockingStandardStreamsFacets models `ahra facets`,
// `modules/facets/FacetsCommandLineInterface.ts` in ahra before and after commit c0269f0e
// (2026-10-03): a `#!` script running its own `main().catch(...)`, with usage exits after writes in
// `main` and a last-resort exit in the catch. The fix is `blockStandardStreams()` first in `main`.
func correctnessRequireBlockingStandardStreamsFacets(fixed bool) correctnessRequireBlockingStandardStreamsCase {
	testCase := correctnessRequireBlockingStandardStreamsCase{shebang: true, lines: []string{
		"function showHelp(): void {",
		"    console.log('facets - Domain-specific memory extraction from conversations');",
		"}",
		"async function main(): Promise<void> {",
		"    const commandLineArguments = argumentsList.slice(2);",
		"    if(commandLineArguments.length === 0) {",
		"        showHelp();",
		"        return;",
		"    }",
		"    const [facetName, command] = commandLineArguments;",
		"    if(!facetName || !command) {",
		"        console.error('Usage: node facets/index.js <facet> <update|build>');",
		"        process.exit(1);",
		"    }",
		"    if(command === 'update') {",
		"        console.log('Faceted Memory Extraction');",
		"        await run();",
		"        return;",
		"    }",
		"    console.error(`Unknown command: ${command}`);",
		"    process.exit(1);",
		"}",
		"main().catch(function(error: unknown) {",
		"    console.error('Error:', error);",
		"    process.exit(1);",
		"});",
	}}
	if fixed {
		testCase.imports = []string{correctnessRequireBlockingStandardStreamsImportBlock}
		testCase.lines[4] = "    blockStandardStreams();\n" + testCase.lines[4]
	}
	return testCase
}

// correctnessRequireBlockingStandardStreamsShardWorker models the forked
// `modules/data/DataConversionShardWorker.ts` before and after commit c2f95018: no `#!`, nothing
// imports it (the composer forks it by path), top-level code that prints its usage and exits. The fix
// blocks at the top of the module.
func correctnessRequireBlockingStandardStreamsShardWorker(fixed bool) correctnessRequireBlockingStandardStreamsCase {
	testCase := correctnessRequireBlockingStandardStreamsCase{lines: []string{
		"const specFilePath = argumentsList[2];",
		"if(!specFilePath) {",
		"    console.error('Usage: DataConversionShardWorker.ts <shard-spec.json>');",
		"    process.exit(2);",
		"}",
		"sendResults(results);",
		"process.exit(0);",
	}}
	if fixed {
		testCase.imports = []string{correctnessRequireBlockingStandardStreamsImportBlock}
		testCase.lines = append([]string{"blockStandardStreams();"}, testCase.lines...)
	}
	return testCase
}

// correctnessRequireBlockingStandardStreamsPhiSocialUpload models
// `modules/phi/social/scripts/PhiSocialUpload.ts` before commit c2f95018: a `#!` script whose only
// exit is the last-resort one in `main().catch(function onFatal(...))`, reached through the callback.
func correctnessRequireBlockingStandardStreamsPhiSocialUpload(fixed bool) correctnessRequireBlockingStandardStreamsCase {
	testCase := correctnessRequireBlockingStandardStreamsCase{shebang: true, lines: []string{
		"async function main() {",
		"    console.log(`uploading: ${output}`);",
		"    await run();",
		"    console.log('done');",
		"}",
		"main().catch(function onFatal(error: unknown) {",
		"    console.error('fatal:', error);",
		"    process.exit(1);",
		"});",
	}}
	if fixed {
		testCase.imports = []string{correctnessRequireBlockingStandardStreamsImportBlock}
		testCase.lines[0] = "async function main() {\n    blockStandardStreams();"
	}
	return testCase
}

// correctnessRequireBlockingStandardStreamsLintEngineParity models Structure's spawned
// `command-line/StructureLintEngineParity.ts` before and after Structure commit fed59eb3: no `#!`,
// top-level code with a usage exit, a computed `await import(...)`, and a report printed line by line
// before the final exit.
func correctnessRequireBlockingStandardStreamsLintEngineParity(fixed bool) correctnessRequireBlockingStandardStreamsCase {
	testCase := correctnessRequireBlockingStandardStreamsCase{lines: []string{
		"const [cohereBinaryPath, ...filePaths] = argumentsList.slice(2);",
		"if(cohereBinaryPath === undefined || filePaths.length === 0) {",
		"    console.error('Usage: StructureLintEngineParity.ts <cohere binary> <file> [file...]');",
		"    process.exit(1);",
		"}",
		"const projectConfigurationModule: unknown = await import(output);",
		"if(projectConfigurationModule === undefined) {",
		"    console.error('LintConfiguration.ts exports ProjectLintEngineDifferenceReasons, and it is not a record.');",
		"    process.exit(1);",
		"}",
		"for(const line of filePaths) console.log(line);",
		"process.exit(flag ? 0 : 1);",
	}}
	if fixed {
		testCase.imports = []string{correctnessRequireBlockingStandardStreamsImportBlock}
		testCase.lines = append([]string{"blockStandardStreams();"}, testCase.lines...)
	}
	return testCase
}

// correctnessRequireBlockingStandardStreamsNewMigration models `modules/os/migrations/NewMigration.ts`
// before commit c2f95018: a synchronous `main` the module calls at its end.
func correctnessRequireBlockingStandardStreamsNewMigration(fixed bool) correctnessRequireBlockingStandardStreamsCase {
	testCase := correctnessRequireBlockingStandardStreamsCase{lines: []string{
		"function main(): void {",
		"    const name = argumentsList[2];",
		"    if(!name) {",
		"        console.error('NewMigration: missing Name.');",
		"        process.exit(1);",
		"    }",
		"    console.log(`NewMigration: created ${name}`);",
		"}",
		"main();",
	}}
	if fixed {
		testCase.imports = []string{correctnessRequireBlockingStandardStreamsImportBlock}
		testCase.lines[0] = "function main(): void {\n    blockStandardStreams();"
	}
	return testCase
}

// correctnessRequireBlockingStandardStreamsStructure models Structure's `command-line/Structure.ts`
// and every ahra integration: a command line interface handed to `runCommandLineInterface` at the
// module's end. As it stands its top-level setup has no exit; guardFirst puts a usage guard that
// prints and exits above the hand-off, which is the one way such a file blocks too late.
func correctnessRequireBlockingStandardStreamsStructure(guardFirst bool) correctnessRequireBlockingStandardStreamsCase {
	guard := []string{
		"if(argumentsList.length === 0) {",
		"    console.error('Usage: s <command>');",
		"    process.exit(1);",
		"}",
	}
	lines := []string{
		"class StructureCommandLineInterface extends CommandLineInterface {",
		"    async run(): Promise<void> {",
		"        console.log(output);",
		"        process.exit(0);",
		"    }",
		"}",
	}
	if guardFirst {
		lines = append(lines, guard...)
		lines = append(lines, "runCommandLineInterface(new StructureCommandLineInterface());")
	} else {
		lines = append(lines, "runCommandLineInterface(new StructureCommandLineInterface());")
		lines = append(lines, guard...)
	}
	return correctnessRequireBlockingStandardStreamsCase{imports: []string{correctnessRequireBlockingStandardStreamsImportRun}, lines: lines}
}

func TestCorrectnessRequireBlockingStandardStreamsRealSites(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		source   func(fixed bool) correctnessRequireBlockingStandardStreamsCase
		want     string
		contains []string
	}{
		{"ahra facets", correctnessRequireBlockingStandardStreamsFacets, "process.exit(1)", []string{"it starts with `#!`", "the first of 3 such exits"}},
		{"the forked shard worker", correctnessRequireBlockingStandardStreamsShardWorker, "process.exit(2)", []string{"nothing in the program imports it"}},
		{"a main().catch script", correctnessRequireBlockingStandardStreamsPhiSocialUpload, "process.exit(1)", nil},
		{"StructureLintEngineParity", correctnessRequireBlockingStandardStreamsLintEngineParity, "process.exit(1)", nil},
		{"NewMigration", correctnessRequireBlockingStandardStreamsNewMigration, "process.exit(1)", nil},
	}
	for _, testCase := range cases {
		t.Run(testCase.name+" as it stood", func(t *testing.T) {
			t.Parallel()
			correctnessRequireBlockingStandardStreamsExpect(t, correctnessRequireBlockingStandardStreamsRun(t, testCase.source(false)), testCase.want, testCase.contains...)
		})
		t.Run(testCase.name+" fixed", func(t *testing.T) {
			t.Parallel()
			correctnessRequireBlockingStandardStreamsExpect(t, correctnessRequireBlockingStandardStreamsRun(t, testCase.source(true)), "")
		})
	}

	t.Run("a runCommandLineInterface entry as it stands", func(t *testing.T) {
		t.Parallel()
		correctnessRequireBlockingStandardStreamsExpect(t, correctnessRequireBlockingStandardStreamsRun(t, correctnessRequireBlockingStandardStreamsStructure(false)), "")
	})
	t.Run("a runCommandLineInterface entry with a guard that exits above the hand-off", func(t *testing.T) {
		t.Parallel()
		correctnessRequireBlockingStandardStreamsExpect(t, correctnessRequireBlockingStandardStreamsRun(t, correctnessRequireBlockingStandardStreamsStructure(true)), "process.exit(1)")
	})
}

func TestCorrectnessRequireBlockingStandardStreamsFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		testCase correctnessRequireBlockingStandardStreamsCase
		want     string
	}{
		{correctnessRequireBlockingStandardStreamsCase{name: "a callback the module hands on at load (FigmaMcpLauncher.ts)", lines: []string{
			"socketServer.on('error', function(error) {",
			"    console.error(`Could not start the Figma socket server: ${error.message}`);",
			"    process.exit(1);",
			"});",
		}}, "process.exit(1)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "a named const arrow handed on", lines: []string{
			"const onInterrupt = () => {",
			"    console.error('interrupted');",
			"    process.exit(130);",
			"};",
			"onEvent(onInterrupt);",
		}}, "process.exit(130)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "an immediately invoked function", lines: []string{
			"(async function() {",
			"    console.log(output);",
			"    process.exit(0);",
			"})();",
		}}, "process.exit(0)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "a function called by a function called at load", lines: []string{
			"function fail(): void {",
			"    console.error('failed');",
			"    process.exit(1);",
			"}",
			"function main(): void {",
			"    if(!flag) fail();",
			"}",
			"main();",
		}}, "process.exit(1)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "a main that writes and exits before its own block", imports: []string{correctnessRequireBlockingStandardStreamsImportBlock}, lines: []string{
			"function main(): void {",
			"    if(!flag) {",
			"        console.error('usage');",
			"        process.exit(1);",
			"    }",
			"    blockStandardStreams();",
			"    console.log(output);",
			"    process.exit(0);",
			"}",
			"main();",
		}}, "process.exit(1)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "a shebang file that other files import", shebang: true, lines: []string{
			"console.log(output);",
			"process.exit(0);",
		}, others: map[string]string{
			"/repository/modules/other/Importer.ts": "import '../subject/Subject';\n",
		}}, "process.exit(0)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "a local blockStandardStreams in a file that also imports Nexus's", imports: []string{
			"import { blockStandardStreams as nexusBlockStandardStreams } from '../../libraries/nexus/source/system/StandardStreams';",
		}, lines: []string{
			"function blockStandardStreams(): void {",
			"    void nexusBlockStandardStreams.name;",
			"}",
			"blockStandardStreams();",
			"console.error('usage');",
			"process.exit(1);",
		}}, "process.exit(1)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "another file's blockStandardStreams that is not Nexus's", imports: []string{
			"import { blockStandardStreams } from '../shared/Streams';",
		}, lines: []string{
			"blockStandardStreams();",
			"console.error('usage');",
			"process.exit(1);",
		}, others: map[string]string{
			"/repository/modules/shared/Streams.ts": "export function blockStandardStreams(): void {}\n",
		}}, "process.exit(1)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "a helper from a file that imports Nexus but never blocks", imports: []string{
			"import { describeStreams } from '../shared/Describe';",
		}, lines: []string{
			"describeStreams();",
			"console.error('usage');",
			"process.exit(1);",
		}, others: map[string]string{
			"/repository/modules/shared/Describe.ts": correctnessNoProcessExitAfterOutputLines(
				"import { blockStandardStreams } from '../../libraries/nexus/source/system/StandardStreams';",
				"export function describeStreams(): string {",
				"    return blockStandardStreams.name;",
				"}",
			),
		}}, "process.exit(1)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "an exit after an await, in a file that never blocks", lines: []string{
			"async function main(): Promise<void> {",
			"    await run();",
			"    console.error('failed');",
			"    process.exit(1);",
			"}",
			"main();",
		}}, "process.exit(1)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "a callback, in a file that never blocks", lines: []string{
			"process.on('SIGINT', function() {",
			"    console.error('interrupted');",
			"    process.exit(130);",
			"});",
		}}, "process.exit(130)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "an imported module whose function blocks, never called", imports: []string{"import { startCommand } from '../shared/StartCommand';"}, lines: []string{
			"void startCommand;",
			"console.error('usage');",
			"process.exit(1);",
		}, others: map[string]string{
			"/repository/modules/shared/StartCommand.ts": correctnessNoProcessExitAfterOutputLines(
				"import { blockStandardStreams } from '../../libraries/nexus/source/system/StandardStreams';",
				"export function startCommand(): void {",
				"    blockStandardStreams();",
				"}",
			),
		}}, "process.exit(1)"},
		{correctnessRequireBlockingStandardStreamsCase{name: "a write through process.stdout at the top level", lines: []string{
			"process.stdout.write(output);",
			"process.exit(0);",
		}}, "process.exit(0)"},
	}
	for _, testCase := range cases {
		t.Run(testCase.testCase.name, func(t *testing.T) {
			t.Parallel()
			correctnessRequireBlockingStandardStreamsExpect(t, correctnessRequireBlockingStandardStreamsRun(t, testCase.testCase), testCase.want)
		})
	}
}

func TestCorrectnessRequireBlockingStandardStreamsStaysSilent(t *testing.T) {
	t.Parallel()

	library := correctnessNoProcessExitAfterOutputLines(
		"declare function use(value: unknown): void;",
		"export function runRehydrate(username: string): void {",
		"    if(!username) {",
		"        console.error('Usage: ahra os rehydrate <username>');",
		"        process.exit(1);",
		"    }",
		"    use(username);",
		"}",
	)

	cases := []correctnessRequireBlockingStandardStreamsCase{
		{name: "a library file other files import (AhraOsLifecycleCommandLineInterface.ts)", lines: []string{
			"export function runRehydrate(username: string): void {",
			"    if(!username) {",
			"        console.error('Usage: ahra os rehydrate <username>');",
			"        process.exit(1);",
			"    }",
			"}",
		}, others: map[string]string{
			"/repository/modules/other/Importer.ts": "import { runRehydrate } from '../subject/Subject';\nrunRehydrate('');\n",
		}},
		{name: "a script without a #! that another file imports", lines: []string{
			"async function main(): Promise<void> {",
			"    console.error('failed');",
			"    process.exit(1);",
			"}",
			"main();",
		}, others: map[string]string{
			"/repository/modules/other/Importer.ts": "import '../subject/Subject';\n",
		}},
		{name: "a write and exit in an imported function, not followed", imports: []string{"import { runRehydrate } from '../shared/Lifecycle';"}, lines: []string{
			"runRehydrate(output);",
		}, others: map[string]string{
			"/repository/modules/shared/Lifecycle.ts": library,
		}},
		{name: "an exit in an exported function nothing runs at load", shebang: true, lines: []string{
			"export function runDelete(): void {",
			"    console.error('refusing');",
			"    process.exit(1);",
			"}",
		}},
		{name: "a library with a #! and no top-level work (ClaudeUsageApi.ts)", shebang: true, lines: []string{
			"export async function readUsage(): Promise<string> {",
			"    await run();",
			"    return output;",
			"}",
		}},
		{name: "the exit comes before any write", lines: []string{
			"if(!flag) process.exit(1);",
			"console.log('ready');",
		}},
		{name: "blockStandardStreams imported under another name", imports: []string{
			"import { blockStandardStreams as blockStreams } from '../../libraries/nexus/source/system/StandardStreams';",
		}, lines: []string{
			"blockStreams();",
			"console.error('usage');",
			"process.exit(1);",
		}},
		{name: "a helper in another file that blocks, called first", imports: []string{"import { startCommand } from '../shared/StartCommand';"}, lines: []string{
			"startCommand();",
			"console.error('usage');",
			"process.exit(1);",
		}, others: map[string]string{
			"/repository/modules/shared/StartCommand.ts": correctnessNoProcessExitAfterOutputLines(
				"import { blockStandardStreams } from '../../libraries/nexus/source/system/StandardStreams';",
				"export function startCommand(): void {",
				"    blockStandardStreams();",
				"}",
			),
		}},
		{name: "an imported module that blocks while it loads", imports: []string{"import '../shared/BlockOnLoad';"}, lines: []string{
			"console.error('usage');",
			"process.exit(1);",
		}, others: map[string]string{
			"/repository/modules/shared/BlockOnLoad.ts": correctnessNoProcessExitAfterOutputLines(
				"import { blockStandardStreams } from '../../libraries/nexus/source/system/StandardStreams';",
				"blockStandardStreams();",
			),
		}},
		{name: "an exit after an await, in a file that blocks later", imports: []string{correctnessRequireBlockingStandardStreamsImportBlock}, lines: []string{
			"async function main(): Promise<void> {",
			"    await run();",
			"    console.error('failed');",
			"    process.exit(1);",
			"}",
			"main();",
			"blockStandardStreams();",
		}},
		{name: "a callback, in a file that blocks after handing it on", imports: []string{correctnessRequireBlockingStandardStreamsImportBlock}, lines: []string{
			"process.on('SIGINT', function() {",
			"    console.error('interrupted');",
			"    process.exit(130);",
			"});",
			"blockStandardStreams();",
		}},
		{name: "Nexus's function handed to a call that runs it", imports: []string{correctnessRequireBlockingStandardStreamsImportBlock}, lines: []string{
			"onEvent(blockStandardStreams);",
			"console.error('usage');",
			"process.exit(1);",
		}},
		{name: "a block in the arguments of the call that starts main", imports: []string{correctnessRequireBlockingStandardStreamsImportBlock}, lines: []string{
			"function main(ready: void): void {",
			"    void ready;",
			"    console.error('usage');",
			"    process.exit(1);",
			"}",
			"main(blockStandardStreams());",
		}},
		{name: "a block in a class static block, which the graph does not lay out", imports: []string{correctnessRequireBlockingStandardStreamsImportBlock}, lines: []string{
			"class Setup {",
			"    static {",
			"        blockStandardStreams();",
			"    }",
			"}",
			"void Setup;",
			"console.error('usage');",
			"process.exit(1);",
		}},
		{name: "a call through a value the checker cannot name, in a file that can block", imports: []string{correctnessRequireBlockingStandardStreamsImportBlock}, lines: []string{
			"void blockStandardStreams;",
			"const handlers: Record<string, () => void> = {};",
			"handlers[output]?.();",
			"console.error('usage');",
			"process.exit(1);",
		}},
		{name: "a computed import() before the exit, which may load anything", lines: []string{
			"import(output).then(use);",
			"console.error('usage');",
			"process.exit(1);",
		}},
		{name: "a declared function whose body this program cannot see, in a file that can block", imports: []string{correctnessRequireBlockingStandardStreamsImportBlock}, lines: []string{
			"declare function setUpProcess(): void;",
			"void blockStandardStreams;",
			"setUpProcess();",
			"console.error('usage');",
			"process.exit(1);",
		}},
		{name: "the exit in the write's own callback, the correct form", lines: []string{
			"process.stdout.write(output, function() {",
			"    process.exit(0);",
			"});",
		}},
		{name: "the exit code set and nothing exiting", lines: []string{
			"console.log(output);",
			"process.exitCode = 1;",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			correctnessRequireBlockingStandardStreamsExpect(t, correctnessRequireBlockingStandardStreamsRun(t, testCase), "")
		})
	}
}
