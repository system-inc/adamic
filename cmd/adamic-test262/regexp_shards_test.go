package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

func regExpRunnerShard(t *testing.T, corpus, name string, want outcomeKind) {
	t.Helper()
	t.Parallel()
	e := regExpRunnerEngine(t, corpus)
	one := regExpFixture(t, corpus, name)
	got := e.attempt(one)
	if got.Kind != want {
		t.Fatalf("%s: %+v", name, got)
	}
	if want != outcomePass {
		t.Logf("%s: %s", name, got.Reason)
		if original := e.originalRegExp(one); original.Exit != 0 || crashedExecution(original) {
			t.Fatalf("refusal fixture is invalid on untouched Node: %+v", original)
		}
		return
	}
}

func TestRegExpRunnerMetadata(t *testing.T) {
	regExpRunnerShard(t, "testdata/regexp", "metadata", outcomePass)
}
func TestRegExpRunnerUtilities(t *testing.T) {
	regExpRunnerShard(t, "testdata/regexp", "utilities", outcomePass)
}
func TestRegExpRunnerConstruction(t *testing.T) {
	regExpRunnerShard(t, "testdata/regexp", "construction", outcomeRefused)
}
func TestRegExpRunnerSyntaxFlags(t *testing.T) {
	regExpRunnerShard(t, "testdata/regexp-syntax", "flags", outcomeRefused)
}
func TestRegExpRunnerSyntaxPattern(t *testing.T) {
	regExpRunnerShard(t, "testdata/regexp-syntax", "pattern", outcomeRefused)
}
func TestRegExpRunnerConstructorClone(t *testing.T) {
	regExpRunnerShard(t, "testdata/regexp-constructor-errors", "clone", outcomeRefused)
}
func TestRegExpRunnerConstructorDeclaration(t *testing.T) {
	regExpRunnerShard(t, "testdata/regexp-constructor-errors", "declaration", outcomeRefused)
}
func TestRegExpRunnerConstructorLegacy(t *testing.T) {
	regExpRunnerShard(t, "testdata/regexp-constructor-errors", "legacy", outcomeRefused)
}

func regExpRunnerWASIShard(t *testing.T, name string) {
	t.Helper()
	t.Parallel()
	e := regExpCompilerEngine(t, "testdata/regexp")
	one := regExpFixture(t, e.test262, name)
	if err := os.WriteFile(filepath.Join(e.work, "program.a"), []byte(one.Program), 0600); err != nil {
		t.Fatal(err)
	}
	regExpWASIAgrees(t, e, one)
}
func TestRegExpRunnerMetadataWASI(t *testing.T)  { regExpRunnerWASIShard(t, "metadata") }
func TestRegExpRunnerUtilitiesWASI(t *testing.T) { regExpRunnerWASIShard(t, "utilities") }

func regExpWASIAgrees(t testing.TB, e *engine, one classified) {
	t.Helper()
	options := native.Options{Target: "wasm32-wasi"}
	directory, request := regExpRunnerRuntime(t, options)
	source := filepath.Join(e.work, "program.a")
	compiled := e.compiler.compile(source)
	if compiled.Exit != 0 {
		t.Fatalf("WASI lowering: %+v", compiled)
	}
	if features := native.Features(compiled.Stdout); len(features) != 0 {
		t.Fatalf("undeclared RegExp runner runtime features: %v", features)
	}
	cPath := filepath.Join(e.work, "program.c")
	if err := os.WriteFile(cPath, []byte(compiled.Stdout), 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(e.work, "program.wasm")
	args := append(native.Flags(options), "-I", directory, cPath, "-o", binary)
	args = append(args, native.RuntimeLinkFlags(filepath.Join(directory, "runtime.a"))...)
	args = append(args, native.WASILinkFlags(options)...)
	args = append(args, "-Wl,--export=__stack_low", "-lm")
	linked := runCommand(20*time.Second, nil, request.Compiler, args...)
	if linked.Exit != 0 {
		t.Fatalf("WASI link: %+v", linked)
	}
	got := runProgram(15*time.Second, nil, e.tools.node, "--disable-warning=ExperimentalWarning", "../../internal/native/wasm/run.mjs", binary)
	original := e.originalRegExp(one)
	module := filepath.Join(e.work, "program.mts")
	if err := os.WriteFile(module, []byte(one.Program), 0600); err != nil {
		t.Fatal(err)
	}
	adapted := runProgram(15*time.Second, nil, e.tools.node, "--disable-warning=ExperimentalWarning", module)
	if original.Exit != 0 || adapted.Exit != 0 || adapted.Stdout != original.Stdout || adapted.Stderr != original.Stderr || crashedExecution(adapted) {
		t.Fatalf("adapted Node disagrees with untouched Node: %+v / %+v", adapted, original)
	}
	if got.Exit != original.Exit || got.Stdout != original.Stdout || got.Stderr != original.Stderr || crashedExecution(got) {
		t.Fatalf("WASI disagrees with untouched Node: %+v / %+v", got, original)
	}
}

func TestRegExpRunnerOriginalNodeGuard(t *testing.T) {
	t.Parallel()
	e := regExpRunnerEngine(t, "testdata/regexp")
	one := classify("built-ins/RegExp/false-pass.js", `assert.sameValue(1, 2);`, true)
	one.Program = program(`assertSameValue(1, 1);`)
	if got := e.attempt(one); got.Kind != outcomeFail || !strings.Contains(got.Reason, "original Node") {
		t.Fatalf("false pass accepted: %+v", got)
	}
}
func TestRegExpRunnerConstructorIdentityGuard(t *testing.T) {
	t.Parallel()
	e := regExpRunnerEngine(t, "testdata/regexp")
	one := classify("built-ins/RegExp/wrong-constructor.js", `assert.throws(TypeError, () => { throw new TypeError("type"); });`, true)
	one.Program = program(`assertThrows("TypeError", () => { throw new Error("type"); });`)
	if got := e.attempt(one); got.Kind != outcomeRefused || !strings.Contains(got.Reason, "constructor-identity") {
		t.Fatalf("wrong constructor accepted: %+v", got)
	}
}
func TestRegExpRunnerCaptureGuard(t *testing.T) {
	t.Parallel()
	e := regExpRunnerEngine(t, "testdata/regexp")
	one := classify("built-ins/RegExp/wrong-capture.js", "/*---\nincludes: [regExpUtils.js]\n---*/\nmatchValidator(['x'], 0, 'a')(/a/.exec('a'));", true)
	if got := e.attempt(one); got.Kind != outcomeFail {
		t.Fatalf("wrong capture was not checked: %+v", got)
	}
}
