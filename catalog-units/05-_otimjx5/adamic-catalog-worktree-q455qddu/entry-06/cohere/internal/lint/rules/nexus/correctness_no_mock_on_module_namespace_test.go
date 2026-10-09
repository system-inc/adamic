package nexus

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoMockOnModuleNamespaceFile = "/repository/source/CorrectnessNoMockOnModuleNamespace.test.ts"

const correctnessNoMockOnModuleNamespaceNodeFile = "/repository/source/node.d.ts"

const correctnessNoMockOnModuleNamespaceStoreFile = "/repository/source/Store.ts"

func correctnessNoMockOnModuleNamespaceSource(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

// correctnessNoMockOnModuleNamespaceNodeTypes is the slice of `@types/node` the rule resolves
// against, in the layout of 26.2.0 (ahra's): `node:test` exports `test` through `export =`, with
// `MockTracker` and `mock` inside `namespace test`. A second module declares a `MockTracker` of its
// own, which must not count.
var correctnessNoMockOnModuleNamespaceNodeTypes = correctnessNoMockOnModuleNamespaceSource(
	"interface ImportMeta { url: string }",
	`declare module "node:test" {`,
	"    function test(name?: string, fn?: (context: test.TestContext) => unknown): Promise<void>;",
	"    namespace test {",
	"        interface MockFunctionContext { restore(): void }",
	"        interface Mock { mock: MockFunctionContext }",
	"        interface MockTracker {",
	"            fn(original?: Function): Mock;",
	"            method<MockedObject extends object>(object: MockedObject, methodName: keyof MockedObject, implementation?: Function): Mock;",
	"            getter<MockedObject extends object>(object: MockedObject, methodName: keyof MockedObject, implementation?: Function): Mock;",
	"            setter<MockedObject extends object>(object: MockedObject, methodName: keyof MockedObject, implementation?: Function): Mock;",
	"            property<MockedObject extends object>(object: MockedObject, propertyName: keyof MockedObject, value?: unknown): Mock;",
	"            restoreAll(): void;",
	"        }",
	"        interface TestContext { mock: MockTracker }",
	"        const mock: MockTracker;",
	"    }",
	"    export = test;",
	"}",
	`declare module "node:child_process" {`,
	"    function execFileSync(file: string, args?: readonly string[]): string;",
	"}",
	`declare module "node:module" {`,
	"    function createRequire(path: string): (id: string) => unknown;",
	"    function syncBuiltinESMExports(): void;",
	"}",
	`declare module "node:os" {`,
	"    const EOL: string;",
	"    function hostname(): string;",
	"}",
	`declare module "node:fs" {`,
	"    interface FileSystem { readFileSync(path: string): string }",
	"    const fileSystem: FileSystem;",
	"    export default fileSystem;",
	"}",
	`declare module "mock-kit" {`,
	"    interface MockTracker { method(object: object, methodName: string, implementation?: Function): void }",
	"    const mock: MockTracker;",
	"}",
)

var correctnessNoMockOnModuleNamespaceStore = correctnessNoMockOnModuleNamespaceSource(
	"export function load(): string { return 'stored'; }",
	"export const version = 1;",
)

func correctnessNoMockOnModuleNamespaceFiles(subjectFileName string, sourceText string) map[string]string {
	return map[string]string{
		subjectFileName: sourceText,
		correctnessNoMockOnModuleNamespaceNodeFile:  correctnessNoMockOnModuleNamespaceNodeTypes,
		correctnessNoMockOnModuleNamespaceStoreFile: correctnessNoMockOnModuleNamespaceStore,
	}
}

func correctnessNoMockOnModuleNamespaceRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	return rule_testing.RunTypedFiles(t, CorrectnessNoMockOnModuleNamespace,
		correctnessNoMockOnModuleNamespaceFiles(correctnessNoMockOnModuleNamespaceFile, sourceText),
		correctnessNoMockOnModuleNamespaceFile)
}

// correctnessNoMockOnModuleNamespaceRunUnder builds the fixture under a tsconfig of its own, with the
// `module` option given and every TypeScript extension included, so a `.mts` or `.cts` subject and a
// CommonJS program can be reached.
func correctnessNoMockOnModuleNamespaceRunUnder(t *testing.T, module string, subjectFileName string, sourceText string) rule_testing.Result {
	t.Helper()
	configuration := `{
	"compilerOptions": {
		"strict": true,
		"target": "ES2022",
		"lib": ["ES2022"],
		"module": "` + module + `",
		"moduleDetection": "force",
		"types": []
	},
	"include": ["**/*.ts", "**/*.mts", "**/*.cts"]
}`
	return rule_testing.RunTypedFilesWithSetup(t, CorrectnessNoMockOnModuleNamespace,
		correctnessNoMockOnModuleNamespaceFiles(subjectFileName, sourceText), subjectFileName,
		func(directory string) {
			if err := os.WriteFile(filepath.Join(directory, "tsconfig.json"), []byte(configuration), 0o644); err != nil {
				t.Fatalf("writing the tsconfig: %v", err)
			}
		})
}

// correctnessNoMockOnModuleNamespaceExpect asserts the findings in source order, each by the first
// line of the text it points at, and that none carries a fix.
func correctnessNoMockOnModuleNamespaceExpect(t *testing.T, result rule_testing.Result, want []string) {
	t.Helper()
	// A silent case goes through the shared assertion that the rule stays quiet.
	if len(want) == 0 {
		rule_testing.ExpectClean(t, result)
		return
	}
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	wantIds := make([]string, len(want))
	for index := range want {
		wantIds[index] = correctnessNoMockOnModuleNamespaceId
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	text := result.SourceFile.Text()
	for index, diagnostic := range diagnostics {
		span := text[diagnostic.Range.Pos():diagnostic.Range.End()]
		firstLine, _, _ := strings.Cut(span, "\n")
		if firstLine != want[index] {
			t.Fatalf("finding %d starts %q, want %q", index, firstLine, want[index])
		}
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
}

// correctnessNoMockOnModuleNamespaceTransportSecurity models
// `modules/pensieve/PensieveTransportSecurity.test.ts` in ahra, trimmed to the two mocks and what
// they reach. `fixed` is the file after `d75fe890`, which mocks the CommonJS exports object behind the
// namespace; unfixed is the file before it, which mocks the namespace and throws in both tests.
func correctnessNoMockOnModuleNamespaceTransportSecurity(fixed bool) string {
	importLine := "import * as NodeChildProcess from 'node:child_process';"
	mocked := "NodeChildProcess"
	var exports []string
	if fixed {
		importLine = "import type * as NodeChildProcess from 'node:child_process';"
		mocked = "childProcessCommonJsExports"
		exports = []string{
			"const childProcessCommonJsExports = NodeModule.createRequire(import.meta.url)(",
			"    'node:child_process',",
			") as typeof NodeChildProcess;",
		}
	}
	lines := []string{
		importLine,
		"import * as NodeModule from 'node:module';",
		"import test from 'node:test';",
		"declare function searchBroadly(query: string): Promise<unknown>;",
	}
	lines = append(lines, exports...)
	for _, name := range []string{"source searches pass hostile query text as an argv value", "iMessage search keeps metacharacters in argv"} {
		lines = append(lines,
			"test('"+name+"', async function() {",
			"    const executions: string[] = [];",
			"    const executionMock = test.mock.method(",
			"        "+mocked+",",
			"        'execFileSync',",
			"        function(file: string) {",
			"            executions.push(file);",
			"            return 'No matches';",
			"        },",
			"    );",
			"    NodeModule.syncBuiltinESMExports();",
			"    try {",
			"        await searchBroadly('$(id)');",
			"    }",
			"    finally {",
			"        executionMock.mock.restore();",
			"        NodeModule.syncBuiltinESMExports();",
			"    }",
			"});",
		)
	}
	return correctnessNoMockOnModuleNamespaceSource(lines...)
}

// The real site, before the fix: both tests mock the sealed namespace.
func TestCorrectnessNoMockOnModuleNamespaceFiresOnTransportSecurity(t *testing.T) {
	t.Parallel()

	result := correctnessNoMockOnModuleNamespaceRun(t, correctnessNoMockOnModuleNamespaceTransportSecurity(false))
	correctnessNoMockOnModuleNamespaceExpect(t, result, []string{"test.mock.method(", "test.mock.method("})
}

// The real site, after the fix: the CommonJS exports object is typed `typeof NodeChildProcess` and is
// not the namespace, and the namespace import is type-only.
func TestCorrectnessNoMockOnModuleNamespaceStaysSilentOnFixedTransportSecurity(t *testing.T) {
	t.Parallel()

	result := correctnessNoMockOnModuleNamespaceRun(t, correctnessNoMockOnModuleNamespaceTransportSecurity(true))
	correctnessNoMockOnModuleNamespaceExpect(t, result, nil)
}

// Every way node:test reaches MockTracker, every method that redefines a property, a project module's
// namespace as well as a builtin's, and the wrappers that change only the type.
func TestCorrectnessNoMockOnModuleNamespaceFiresOnEveryMockOfANamespace(t *testing.T) {
	t.Parallel()

	result := correctnessNoMockOnModuleNamespaceRun(t, correctnessNoMockOnModuleNamespaceSource(
		"import * as NodeOs from 'node:os';",
		"import * as Store from './Store';",
		"import test, { mock } from 'node:test';",
		"test('a project module', function(t) { t.mock.method(Store, 'load'); });",
		"mock.getter(NodeOs, 'EOL');",
		"mock.setter(NodeOs, 'EOL');",
		"mock.property(Store, 'version', 2);",
		"test.mock.method((NodeOs), 'hostname');",
		"test.mock.method(NodeOs as typeof NodeOs, 'hostname');",
		"test.mock.method(NodeOs!, 'hostname');",
	))
	correctnessNoMockOnModuleNamespaceExpect(t, result, []string{
		"t.mock.method(Store, 'load')",
		"mock.getter(NodeOs, 'EOL')",
		"mock.setter(NodeOs, 'EOL')",
		"mock.property(Store, 'version', 2)",
		"test.mock.method((NodeOs), 'hostname')",
		"test.mock.method(NodeOs as typeof NodeOs, 'hostname')",
		"test.mock.method(NodeOs!, 'hostname')",
	})
}

// The nearest legitimate shapes: another object of the same type, a default import, a member of the
// namespace, a parameter shadowing the import, a mock that is not node:test's, and calls that patch
// nothing.
func TestCorrectnessNoMockOnModuleNamespaceStaysSilentOnOtherObjects(t *testing.T) {
	t.Parallel()

	result := correctnessNoMockOnModuleNamespaceRun(t, correctnessNoMockOnModuleNamespaceSource(
		"import * as NodeChildProcess from 'node:child_process';",
		"import * as NodeOs from 'node:os';",
		"import fileSystem from 'node:fs';",
		"import test, { mock } from 'node:test';",
		"import { mock as otherMock } from 'mock-kit';",
		"import type * as TypeOnlyOs from 'node:os';",
		"declare const childProcessCopy: typeof NodeChildProcess;",
		"declare const jest: { spyOn(object: object, methodName: string): void };",
		"interface MockTracker { method(object: object, methodName: string): void }",
		"declare const localMock: MockTracker;",
		"mock.method(childProcessCopy, 'execFileSync');",
		"mock.method({ ...NodeOs }, 'hostname');",
		"mock.method(fileSystem, 'readFileSync');",
		"mock.method(TypeOnlyOs, 'hostname');",
		"function patch(NodeChildProcess: { execFileSync(): string }) { test.mock.method(NodeChildProcess, 'execFileSync'); }",
		"jest.spyOn(NodeChildProcess, 'execFileSync');",
		"otherMock.method(NodeChildProcess, 'execFileSync');",
		"localMock.method(NodeChildProcess, 'execFileSync');",
		"mock.fn(NodeChildProcess.execFileSync);",
		"mock.restoreAll();",
		"patch({ execFileSync() { return ''; } });",
	))
	correctnessNoMockOnModuleNamespaceExpect(t, result, nil)
}

// A file that runs as an ES module whatever the program emits, and one that never does.
func TestCorrectnessNoMockOnModuleNamespaceReadsTheFileExtension(t *testing.T) {
	t.Parallel()

	source := correctnessNoMockOnModuleNamespaceSource(
		"import * as NodeOs from 'node:os';",
		"import { mock } from 'node:test';",
		"mock.method(NodeOs, 'hostname');",
	)
	module := correctnessNoMockOnModuleNamespaceRunUnder(t, "commonjs", "/repository/source/Module.test.mts", source)
	correctnessNoMockOnModuleNamespaceExpect(t, module, []string{"mock.method(NodeOs, 'hostname')"})

	commonJs := correctnessNoMockOnModuleNamespaceRunUnder(t, "esnext", "/repository/source/Script.test.cts", source)
	correctnessNoMockOnModuleNamespaceExpect(t, commonJs, nil)
}

// A `.ts` file follows the program: reported under an ES `module` and `preserve`, declined under
// CommonJS and under `nodenext`, where its format comes from a `package.json` the rule does not read.
func TestCorrectnessNoMockOnModuleNamespaceReadsTheModuleOption(t *testing.T) {
	t.Parallel()

	source := correctnessNoMockOnModuleNamespaceSource(
		"import * as NodeOs from 'node:os';",
		"import { mock } from 'node:test';",
		"mock.method(NodeOs, 'hostname');",
	)
	for _, module := range []string{"esnext", "es2020", "preserve"} {
		result := correctnessNoMockOnModuleNamespaceRunUnder(t, module, correctnessNoMockOnModuleNamespaceFile, source)
		correctnessNoMockOnModuleNamespaceExpect(t, result, []string{"mock.method(NodeOs, 'hostname')"})
	}
	for _, module := range []string{"commonjs", "nodenext"} {
		result := correctnessNoMockOnModuleNamespaceRunUnder(t, module, correctnessNoMockOnModuleNamespaceFile, source)
		correctnessNoMockOnModuleNamespaceExpect(t, result, nil)
	}
}
