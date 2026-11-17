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

// overloadKindProbes are #0kkpw2e's scout probes, in testdata/overload_kind_probe: x1, x2 and x3 copy
// tsc's injectClassNamedEvaluationHelperBlockIfMissing, transformNamedEvaluation and getParseTreeNode,
// and r1, r2 and r3 are the three roads an Extract<U, Pick<T, "kind">> judgment must measure (an open
// enum kind with a shared value, a kind that is a union of literals, a getter-backed kind). A throwaway
// scout: it fails on nothing it observes, and logs one ROW per probe, the stop's words or "lowers", and
// for a probe that lowers, Node on the source, the JavaScript backend and native, each run once.
var overloadKindProbes = []string{
	"x1_inject",
	"x2_transform",
	"x3_parse_tree_node",
	"r1_open_enum_shared_value",
	"r2_kind_union",
	"r3_getter_kind",
}

func TestOverloadKindProbe(t *testing.T) {
	t.Parallel()
	for _, name := range overloadKindProbes {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("testdata", "overload_kind_probe", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			path := writeAgreementSource(t, "main.ts", string(source))
			program, err := overloadKindProbeLower(path)
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
			observe("native", overloadKindProbeRun(t, binary))
		})
	}
}

// overloadKindProbeLower loads and lowers path, a load failure, a refusal and a panic each an error to
// log, so one probe's stop never hides another's row.
func overloadKindProbeLower(path string) (program *ir.Program, err error) {
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

// overloadKindProbeRun runs the sanitized native binary, as overloadCheckStops does.
func overloadKindProbeRun(t *testing.T, binary string) nodeObservation {
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
