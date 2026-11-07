package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

func regExpRunnerShard(test *testing.T, corpus, name string, want outcomeKind) {
	test.Helper()
	e := regExpRunnerEngine(test, corpus)
	one := regExpFixture(test, corpus, name)
	got := e.attempt(one)
	if got.Kind != want {
		test.Fatalf("%s: %+v", name, got)
	}
	if want != outcomePass {
		test.Logf("%s: %s", name, got.Reason)
		if original := e.originalRegExp(one); original.Exit != 0 || crashedExecution(original) {
			test.Fatalf("refusal fixture is invalid on untouched Node: %+v", original)
		}
		return
	}
}

func TestRegExpRunnerMetadata(test *testing.T) {
	test.Parallel()
	regExpRunnerShard(test, "testdata/regexp", "metadata", outcomePass)
}
func TestRegExpRunnerUtilities(test *testing.T) {
	test.Parallel()
	regExpRunnerShard(test, "testdata/regexp", "utilities", outcomePass)
}
func TestRegExpRunnerConstruction(test *testing.T) {
	test.Parallel()
	regExpRunnerShard(test, "testdata/regexp", "construction", outcomeRefused)
}
func TestRegExpRunnerSyntaxFlags(test *testing.T) {
	test.Parallel()
	regExpRunnerShard(test, "testdata/regexp-syntax", "flags", outcomeRefused)
}
func TestRegExpRunnerSyntaxPattern(test *testing.T) {
	test.Parallel()
	regExpRunnerShard(test, "testdata/regexp-syntax", "pattern", outcomeRefused)
}
func TestRegExpRunnerConstructorClone(test *testing.T) {
	test.Parallel()
	regExpRunnerShard(test, "testdata/regexp-constructor-errors", "clone", outcomeRefused)
}
func TestRegExpRunnerConstructorDeclaration(test *testing.T) {
	test.Parallel()
	regExpRunnerShard(test, "testdata/regexp-constructor-errors", "declaration", outcomeRefused)
}
func TestRegExpRunnerConstructorLegacy(test *testing.T) {
	test.Parallel()
	regExpRunnerShard(test, "testdata/regexp-constructor-errors", "legacy", outcomeRefused)
}

func regExpRunnerWASIShard(test *testing.T, name string) {
	test.Helper()
	e := regExpCompilerEngine(test, "testdata/regexp")
	one := regExpFixture(test, e.test262, name)
	if writeError := os.WriteFile(filepath.Join(e.work, "program.a"), []byte(one.Program), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	regExpWASIAgrees(test, e, one)
}
func TestRegExpRunnerMetadataWASI(test *testing.T) {
	test.Parallel()
	regExpRunnerWASIShard(test, "metadata")
}
func TestRegExpRunnerUtilitiesWASI(test *testing.T) {
	test.Parallel()
	regExpRunnerWASIShard(test, "utilities")
}

func regExpWASIAgrees(test testing.TB, e *engine, one classified) {
	test.Helper()
	options := native.Options{Target: "wasm32-wasi"}
	directory, request := regExpRunnerRuntime(test, options)
	source := filepath.Join(e.work, "program.a")
	compiled := e.compiler.compile(source)
	if compiled.Exit != 0 {
		test.Fatalf("WASI lowering: %+v", compiled)
	}
	if features := native.Features(compiled.Stdout); len(features) != 0 {
		test.Fatalf("undeclared RegExp runner runtime features: %v", features)
	}
	cPath := filepath.Join(e.work, "program.c")
	if writeError := os.WriteFile(cPath, []byte(compiled.Stdout), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	binary := filepath.Join(e.work, "program.wasm")
	args := append(native.Flags(options), "-I", directory, cPath, "-o", binary)
	args = append(args, native.RuntimeLinkFlags(filepath.Join(directory, "runtime.a"))...)
	args = append(args, native.WASILinkFlags(options)...)
	args = append(args, "-Wl,--export=__stack_low", "-lm")
	linked := runCommand(20*time.Second, nil, request.Compiler, args...)
	if linked.Exit != 0 {
		test.Fatalf("WASI link: %+v", linked)
	}
	got := runProgram(
		15*time.Second,
		nil,
		e.tools.node,
		"--disable-warning=ExperimentalWarning",
		"../../internal/native/wasm/run.mjs",
		binary,
	)
	original := e.originalRegExp(one)
	module := filepath.Join(e.work, "program.mts")
	if writeError := os.WriteFile(module, []byte(one.Program), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	adapted := runProgram(15*time.Second, nil, e.tools.node, "--disable-warning=ExperimentalWarning", module)
	if original.Exit != 0 ||
		adapted.Exit != 0 ||
		adapted.Stdout != original.Stdout ||
		adapted.Stderr != original.Stderr ||
		crashedExecution(adapted) {
		test.Fatalf("adapted Node disagrees with untouched Node: %+v / %+v", adapted, original)
	}
	if got.Exit != original.Exit ||
		got.Stdout != original.Stdout ||
		got.Stderr != original.Stderr ||
		crashedExecution(got) {
		test.Fatalf("WASI disagrees with untouched Node: %+v / %+v", got, original)
	}
}

func TestRegExpRunnerOriginalNodeGuard(test *testing.T) {
	test.Parallel()
	e := regExpRunnerEngine(test, "testdata/regexp")
	one := classify("built-ins/RegExp/false-pass.js", `assert.sameValue(1, 2);`, true)
	one.Program = program(`assertSameValue(1, 1);`)
	if got := e.attempt(one); got.Kind != outcomeFail || !strings.Contains(got.Reason, "original Node") {
		test.Fatalf("false pass accepted: %+v", got)
	}
}
func TestRegExpRunnerConstructorIdentityGuard(test *testing.T) {
	test.Parallel()
	e := regExpRunnerEngine(test, "testdata/regexp")
	one := classify(
		"built-ins/RegExp/wrong-constructor.js",
		`assert.throws(TypeError, () => { throw new TypeError("type"); });`,
		true,
	)
	one.Program = program(`assertThrows("TypeError", () => { throw new Error("type"); });`)
	if got := e.attempt(one); got.Kind != outcomeRefused || !strings.Contains(got.Reason, "constructor-identity") {
		test.Fatalf("wrong constructor accepted: %+v", got)
	}
}
func TestRegExpRunnerCaptureGuard(test *testing.T) {
	test.Parallel()
	e := regExpRunnerEngine(test, "testdata/regexp")
	one := classify(
		"built-ins/RegExp/wrong-capture.js",
		"/*---\nincludes: [regExpUtils.js]\n---*/\nmatchValidator(['x'], 0, 'a')(/a/.exec('a'));",
		true,
	)
	if got := e.attempt(one); got.Kind != outcomeFail {
		test.Fatalf("wrong capture was not checked: %+v", got)
	}
}
