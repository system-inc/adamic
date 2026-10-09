package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
)

func TestRuledBackendOutcomes(t *testing.T) {
	t.Parallel()
	if len(ruledBackendDivergences) != 2 {
		t.Fatal("only the weak-after-free and named-stack fixtures are ruled to diverge")
	}
	entry := ruledBackendDivergences[0]
	path, err := filepath.Abs(filepath.Join(repository, entry.fixture))
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if difference := disagreement(entry.javascript, source); difference != "" {
		t.Fatalf("source Node: %s: exit %d stdout %q stderr %q", difference, source.exitCode, source.stdout, source.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	backend := onJavaScriptBackend(t, program)
	native, _ := natively(t, program)
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
func TestUnlistedBackendAgreement(t *testing.T) {
	t.Parallel()
	fixture := "dedication/dedication.a"
	program, err := lowered(t, filepath.Join(repository, fixture))
	if err != nil {
		t.Fatal(err)
	}
	backend := onJavaScriptBackend(t, program)
	native, _ := natively(t, program)
	if difference := backendDisagreement(fixture, backend, native); difference != "" {
		t.Fatalf("unlisted backend divergence: %s; JavaScript exit %d stdout %q stderr %q; native exit %d stdout %q stderr %q", difference, backend.exitCode, backend.stdout, backend.stderr, native.exitCode, native.stdout, native.stderr)
	}
}

func TestTerminalStackStop(t *testing.T) {
	t.Parallel()
	fixture := "internal/oracle/testdata/d96d304_try_stack.a"
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		t.Fatal(err)
	}
	// Original Node catches V8's overflow; the inserted guard must be terminal.
	source := onNode(t, path)
	wantSource := run{stdout: []byte("caught the overflow\nafter\n"), exitCode: 0}
	if difference := disagreement(wantSource, source); difference != "" {
		t.Fatalf("original Node: %s", difference)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{stderr: []byte("adamic: panic: RangeError: Maximum call stack size exceeded\n"), exitCode: 70}
	backend := onJavaScriptBackend(t, program)
	if difference := disagreement(want, backend); difference != "" {
		t.Fatalf("terminal stack stop: %s; JavaScript exit %d stdout %q stderr %q", difference, backend.exitCode, backend.stdout, backend.stderr)
	}
	wantNative := run{stderr: []byte("adamic: panic: RangeError: Maximum call stack size exceeded in depth\n"), exitCode: 70}
	native, _ := natively(t, program)
	if difference := disagreement(wantNative, native); difference != "" {
		t.Fatalf("native stack stop: %s", difference)
	}
	if difference := backendDisagreement(fixture, backend, native); difference != "" {
		t.Fatal(difference)
	}
	requireExactBackendOutcomes(t, fixture, backend, native)
	// Catch instrumentation must not run before the terminal guard either.
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
	if difference := disagreement(want, onNode(t, generated)); difference != "" {
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
	want := onNode(t, path)
	got := onNode(t, generated)
	if difference := disagreement(want, got); difference != "" {
		t.Fatalf("non-overflow RangeError was terminal: %s; exit %d stdout %q stderr %q", difference, got.exitCode, got.stdout, got.stderr)
	}
}
