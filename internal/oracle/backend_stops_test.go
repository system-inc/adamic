package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

func TestRuledWeakBackendOutcomes(t *testing.T) {
	t.Parallel()
	if len(ruledBackendDivergences) != 2 {
		t.Fatal("only the weak-after-free and named-stack fixtures are ruled to diverge")
	}
	entry := ruledBackendDivergences[0]
	path, err := filepath.Abs(filepath.Join(repository, entry.fixture))
	if err != nil {
		t.Fatal(err)
	}
	source := onSliceNode(t, path)
	if difference := disagreement(entry.javascript, source); difference != "" {
		t.Fatalf("source Node: %s: exit %d stdout %q stderr %q", difference, source.exitCode, source.stdout, source.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	backend := onSliceJavaScript(t, program)
	native, _ := nativelyUncached(t, program)
	if difference := backendDisagreement(entry.fixture, backend, native); difference != "" {
		t.Fatal(difference)
	}
	requireExactBackendOutcomes(t, entry.fixture, backend, native)
}

func requireExactBackendOutcomes(t *testing.T, fixture string, backend, native run) {
	t.Helper()
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
				t.Fatalf("listed %s %s mutation was waived", side, field)
			}
		}
	}
}

// This positive control is also run with a mutated JavaScript emitter. The native
// output stays intact, so only the unlisted backend comparison can reject it.
func TestUnlistedBackendSourceJavaScriptAgreement(t *testing.T) {
	t.Parallel()
	assertUnlistedBackendSourceAgreement(t, false)
}

func TestUnlistedBackendSourceNativeAgreement(t *testing.T) {
	t.Parallel()
	assertUnlistedBackendSourceAgreement(t, true)
}

func assertUnlistedBackendSourceAgreement(t *testing.T, nativeBackend bool) {
	t.Helper()
	fixture := "dedication/dedication.a"
	path := filepath.Join(repository, fixture)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onSliceNode(t, path)
	var actual run
	if nativeBackend {
		actual, _ = nativelyUncached(t, program)
	} else {
		actual = onSliceJavaScript(t, program)
	}
	if difference := backendDisagreement(fixture, truth, actual); difference != "" {
		t.Fatalf("unlisted backend divergence: %s; exit %d stdout %q stderr %q", difference, actual.exitCode, actual.stdout, actual.stderr)
	}
}

func terminalStackProgram(t *testing.T) *ir.Program {
	t.Helper()
	path := filepath.Join(repository, "internal/oracle/testdata/d96d304_try_stack.a")
	source := onSliceNode(t, path)
	wantSource := run{stdout: []byte("caught the overflow\nafter\n"), exitCode: 0}
	if difference := disagreement(wantSource, source); difference != "" {
		t.Fatalf("original Node: %s", difference)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	return program
}

func TestTerminalStackStopJavaScript(t *testing.T) {
	t.Parallel()
	program := terminalStackProgram(t)
	entry := ruledBackendDivergences[1]
	backend := onSliceJavaScript(t, program)
	if difference := backendDisagreement(entry.fixture, backend, entry.native); difference != "" {
		t.Fatalf("terminal stack stop: %s", difference)
	}
	requireExactBackendOutcomes(t, entry.fixture, backend, entry.native)
}

func TestTerminalStackStopNative(t *testing.T) {
	t.Parallel()
	program := terminalStackProgram(t)
	entry := ruledBackendDivergences[1]
	actual, _ := nativelyUncached(t, program)
	if difference := backendDisagreement(entry.fixture, entry.javascript, actual); difference != "" {
		t.Fatalf("native stack stop: %s", difference)
	}
	requireExactBackendOutcomes(t, entry.fixture, entry.javascript, actual)
}

func TestTerminalStackStopMarkedJavaScript(t *testing.T) {
	t.Parallel()
	program := terminalStackProgram(t)
	code := javascript.JavaScriptWith(program, javascript.Options{Mark: func(at *ir.Statement, part int) string {
		if _, isTry := (*at).(ir.Try); isTry && part == 1 {
			return `console.log("catch instrumentation ran")`
		}
		return ""
	}})
	generated := filepath.Join(t.TempDir(), "marked.mjs")
	if err := os.WriteFile(generated, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(ruledBackendDivergences[1].javascript, onSliceNode(t, generated)); difference != "" {
		t.Fatalf("catch instrumentation preceded terminal guard: %s", difference)
	}
}

// V8 RangeErrors with other messages must remain catchable. This changes only
// the generated control's ordinary Error constructor, outside Adamic lowering;
// native's unrelated terminal code-point check is not an oracle exception.
func TestNonOverflowRangeErrorCatch(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/backend_stop_catches.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := javascript.JavaScript(program)
	original := `throw new Error("ordinary");`
	if strings.Count(code, original) != 1 {
		t.Fatal("ordinary throw control missing")
	}
	code = strings.Replace(code, original, `throw new RangeError("ordinary");`, 1)
	generated := filepath.Join(t.TempDir(), "range-error.mjs")
	if err := os.WriteFile(generated, []byte(code), 0644); err != nil {
		t.Fatal(err)
	}
	want := onSliceNode(t, path)
	got := onSliceNode(t, generated)
	if difference := disagreement(want, got); difference != "" {
		t.Fatalf("non-overflow RangeError was terminal: %s; exit %d stdout %q stderr %q", difference, got.exitCode, got.stdout, got.stderr)
	}
}
