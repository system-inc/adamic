package nexus

import (
	"encoding/json"
	"os"
	"testing"
)

func TestWave30ExportBlockingControls(t *testing.T) {
	var cases []map[string]string
	for _, builder := range []func(bool) correctnessRequireBlockingStandardStreamsCase{correctnessRequireBlockingStandardStreamsFacets, correctnessRequireBlockingStandardStreamsShardWorker, correctnessRequireBlockingStandardStreamsPhiSocialUpload, correctnessRequireBlockingStandardStreamsLintEngineParity, correctnessRequireBlockingStandardStreamsNewMigration, correctnessRequireBlockingStandardStreamsStructure} {
		for _, fixed := range []bool{false, true} {
			cases = append(cases, builder(fixed).files())
		}
	}
	casesOuter := cases

	{
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
		for _, item := range cases {
			casesOutput(item.testCase.files(), &casesOuter)
		}
	}
	{
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
		for _, item := range cases {
			casesOutput(item.files(), &casesOuter)
		}
	}

	cases = casesOuter
	data, err := json.MarshalIndent(cases, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(os.Getenv("ADAMIC_WAVE_30_BLOCKING_EXPORT"), append(data, '\n'), 0600); err != nil {
		t.Fatal(err)
	}
}
func casesOutput(files map[string]string, output *[]map[string]string) {
	*output = append(*output, files)
}
