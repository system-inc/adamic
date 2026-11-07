package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

func requireExactBackendOutcomes(test *testing.T, fixture string, backend, native run) {
	test.Helper()
	// The exception remains exact, including bytes printed before the terminal stop.
	for _, side := range []string{"native", "JavaScript"} {
		for _, field := range []string{"exit", "stdout", "stderr"} {
			left, right := backend, native
			mutated := &right
			if side == "JavaScript" {
				mutated = &left
			}
			switch field {
			case "exit":
				mutated.exitCode++
			case "stdout":
				mutated.stdout = append(append([]byte{}, mutated.stdout...), '!')
			case "stderr":
				mutated.stderr = append(append([]byte{}, mutated.stderr...), '!')
			}
			if backendDisagreement(fixture, left, right) == "" {
				test.Fatalf("listed %s %s mutation was waived", side, field)
			}
		}
	}
}

// This positive control is also run with a mutated JavaScript emitter. The native
// output stays intact, so only the unlisted backend comparison can reject it.
func TestUnlistedBackendSourceJavaScriptAgreement(test *testing.T) {
	test.Parallel()
	assertUnlistedBackendSourceAgreement(test, false)
}

func TestUnlistedBackendSourceNativeAgreement(test *testing.T) {
	test.Parallel()
	assertUnlistedBackendSourceAgreement(test, true)
}

func assertUnlistedBackendSourceAgreement(test *testing.T, nativeBackend bool) {
	test.Helper()
	fixture := "dedication/dedication.a"
	path := filepath.Join(repository, fixture)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	truth := onSliceNode(test, path)
	var actual run
	if nativeBackend {
		actual, _ = nativelyUncached(test, program)
	} else {
		actual = onSliceJavaScript(test, program)
	}
	if difference := backendDisagreement(fixture, truth, actual); difference != "" {
		test.Fatalf(
			"unlisted backend divergence: %s; exit %d stdout %q stderr %q",
			difference,
			actual.exitCode,
			actual.stdout,
			actual.stderr,
		)
	}
}

func terminalStackProgram(test *testing.T) *ir.Program {
	test.Helper()
	// The weak-after-free stop is pinned at its weak_miscompiles path by weak_differential_test.go.
	// The rest are installed-layout's and no-javascript-plugins' two fixtures each, which pin the
	// source on Node too. node-without-inspector adds caught and uncaught fixtures.
	if len(ruledBackendDivergences) != 7 ||
		ruledBackendDivergences[0].fixture != "internal/oracle/testdata/d96d304_try_stack.a" {
		test.Fatal("only the named-stack fixture is ruled to diverge between backends, and it's the first row")
	}
	path := filepath.Join(repository, "internal/oracle/testdata/d96d304_try_stack.a")
	source := onSliceNode(test, path)
	wantSource := run{stdout: []byte("caught the overflow\nafter\n"), exitCode: 0}
	if difference := disagreement(wantSource, source); difference != "" {
		test.Fatalf("original Node: %s", difference)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	return program
}

func TestTerminalStackStopJavaScript(test *testing.T) {
	test.Parallel()
	program := terminalStackProgram(test)
	entry := ruledBackendDivergences[0]
	backend := onSliceJavaScript(test, program)
	if difference := backendDisagreement(entry.fixture, backend, entry.native); difference != "" {
		test.Fatalf("terminal stack stop: %s", difference)
	}
	requireExactBackendOutcomes(test, entry.fixture, backend, entry.native)
}

func TestTerminalStackStopNative(test *testing.T) {
	test.Parallel()
	program := terminalStackProgram(test)
	entry := ruledBackendDivergences[0]
	actual, _ := nativelyUncached(test, program)
	if difference := backendDisagreement(entry.fixture, entry.javascript, actual); difference != "" {
		test.Fatalf("native stack stop: %s", difference)
	}
	requireExactBackendOutcomes(test, entry.fixture, entry.javascript, actual)
}

func TestTerminalStackStopMarkedJavaScript(test *testing.T) {
	test.Parallel()
	program := terminalStackProgram(test)
	code := javascript.JavaScriptWith(program, javascript.Options{Mark: func(at *ir.Statement, part int) string {
		if _, isTry := (*at).(ir.Try); isTry && part == 1 {
			return `console.log("catch instrumentation ran")`
		}
		return ""
	}})
	generated := filepath.Join(test.TempDir(), "marked.mjs")
	if writeError := os.WriteFile(generated, []byte(code), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	if difference := disagreement(
		ruledBackendDivergences[0].javascript,
		onSliceNode(test, generated),
	); difference != "" {
		test.Fatalf("catch instrumentation preceded terminal guard: %s", difference)
	}
}

// V8 RangeErrors with other messages must remain catchable. This changes only
// the generated control's ordinary Error constructor, outside Adamic lowering;
// native's unrelated terminal code-point check is not an oracle exception.
func TestNonOverflowRangeErrorCatch(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/backend_stop_catches.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	code := javascript.JavaScript(program)
	original := `throw new Error("ordinary");`
	if strings.Count(code, original) != 1 {
		test.Fatal("ordinary throw control missing")
	}
	code = strings.Replace(code, original, `throw new RangeError("ordinary");`, 1)
	generated := filepath.Join(test.TempDir(), "range-error.mjs")
	if writeError := os.WriteFile(generated, []byte(code), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	want := onSliceNode(test, path)
	got := onSliceNode(test, generated)
	if difference := disagreement(want, got); difference != "" {
		test.Fatalf(
			"non-overflow RangeError was terminal: %s; exit %d stdout %q stderr %q",
			difference,
			got.exitCode,
			got.stdout,
			got.stderr,
		)
	}
}
