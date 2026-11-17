package lower

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
)

// overloadErasurePrelude holds #5fws6q8 step 1's two overload sets, each with a last overload that narrows
// the result and omits the implementation's y. S1, pick, is accepted at its declaration by censusReturnProof
// (under the narrow overload y is undefined, so the if's branch is unreachable and x is a number), so its
// calls are erased, never checked. S2, choose, returns through a conditional the proof can't read, so each
// call resolving to its narrow overload is checked (overload_result_check.go). The narrow overload is last,
// since .call, .apply and .bind infer from an overload set's last signature.
const overloadErasurePrelude = `function pick(x: number, y: number): string;
function pick(x: number): number;
function pick(x: number, y?: number): number | string {
    if (y !== undefined) {
        return "s";
    }
    return x;
}
function choose(x: number, y: number): string;
function choose(x: number): number;
function choose(x: number, y?: number): number | string {
    return y !== undefined ? "s" : x;
}
`

// overloadErasureRoads are the entry roads, F standing for the set's name. Every road but the control and
// the plain rows is written so a lowering that keeps the narrow overload's promise disagrees with Node: the
// value is called with y, or the tuple typed [number] holds a second element at run time, so Node prints
// string where the types say number. The plain rows take the same road with nothing extra, to show where
// the road itself stops if the pushed tuple stops first.
var overloadErasureRoads = []struct{ name, body string }{
	{"control", "console.log(String(F(1)));\nconsole.log(F(1, 2));\n"},
	{"alias", "const g: (x: number, y: number) => number = F;\nconsole.log(typeof g(1, 2));\n"},
	{"callback", "function apply(h: (x: number, y: number) => number): number {\n    return h(1, 2);\n}\nconsole.log(typeof apply(F));\n"},
	{"field", "const o: { readonly h: (x: number, y: number) => number } = { h: F };\nconsole.log(typeof o.h(1, 2));\n"},
	{"call", "const args: [number] = [1];\nargs.push(2);\nconsole.log(typeof F.call(undefined, ...args));\n"},
	{"apply", "const args: [number] = [1];\nargs.push(2);\nconsole.log(typeof F.apply(undefined, args));\n"},
	{"bind", "const args: [number] = [1];\nargs.push(2);\nconst b = F.bind(undefined);\nconsole.log(typeof b(...args));\n"},
	{"spread", "const args: [number] = [1];\nargs.push(2);\nconsole.log(typeof F(...args));\n"},
	{"call_plain", "console.log(typeof F.call(undefined, 1));\n"},
	{"apply_plain", "console.log(typeof F.apply(undefined, [1]));\n"},
	{"bind_plain", "const b = F.bind(undefined);\nconsole.log(typeof b(1));\n"},
	{"spread_plain", "const one: [number] = [1];\nconsole.log(typeof F(...one));\n"},
}

// TestOverloadErasureProbe is #5fws6q8 step 1's throwaway probe: is an overload proof's erasure keyed per
// call or per declaration, that is, does any road reach an erased set's implementation with an argument its
// narrow overload omits? It fails on nothing it observes, and logs one ROW per set and road: the stop's
// words, or "lowers" and then Node on the source, the JavaScript backend and sanitized native, each once.
func TestOverloadErasureProbe(t *testing.T) {
	t.Parallel()
	for _, set := range []string{"pick", "choose"} {
		for _, road := range overloadErasureRoads {
			name := map[string]string{"pick": "S1", "choose": "S2"}[set] + "_" + road.name
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				path := writeAgreementSource(t, "main.ts", overloadErasurePrelude+strings.ReplaceAll(road.body, "F", set))
				program, err := overloadErasureProbeLower(path)
				if err != nil {
					t.Logf("ROW %s %s", name, strings.ReplaceAll(err.Error(), "\n", " | "))
					return
				}
				t.Logf("ROW %s lowers", name)
				observe := func(label string, observation nodeObservation) {
					t.Logf("ROW %s %s exit=%d stdout=%q stderr=%q", name, label, observation.code, observation.stdout, strings.SplitN(string(observation.stderr), "\n", 2)[0])
				}
				observe("node", runAgreementNode(t, path))
				generated := filepath.Join(t.TempDir(), "generated.mjs")
				if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
					t.Fatal(err)
				}
				observe("javascript", runAgreementNode(t, generated))
				binary := filepath.Join(t.TempDir(), "native")
				if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
					t.Logf("ROW %s native build: %s", name, strings.ReplaceAll(err.Error(), "\n", " | "))
					return
				}
				observe("native", overloadErasureProbeRun(t, binary))
				counts := program.OverloadResultChecks
				t.Logf("ROW %s counts checked=%d proven=%d erased=%d", name, counts.Checked, counts.Proven, counts.Erased)
			})
		}
	}
}

// overloadErasureProbeLower loads and lowers path, a load failure, a refusal and a panic each an error to
// log, so one row's stop never hides another's.
func overloadErasureProbeLower(path string) (program *ir.Program, err error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			program, err = nil, fmt.Errorf("panic: %v", recovered)
		}
	}()
	checked, err := load.Load([]string{path})
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	return Lower(context.Background(), checked)
}

// overloadErasureProbeRun runs the sanitized native binary, as overloadCheckStops does.
func overloadErasureProbeRun(t *testing.T, binary string) nodeObservation {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	code := 0
	if err := command.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatalf("native %s: %v", binary, err)
		}
		code = exit.ExitCode()
	}
	return nodeObservation{stdout.Bytes(), stderr.Bytes(), code}
}
