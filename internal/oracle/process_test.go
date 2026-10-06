package oracle

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func processStreams(t *testing.T, stdout, stderr bool, environment map[string]string, command ...string) run {
	t.Helper()
	encoded, err := json.Marshal(environment)
	if err != nil {
		t.Fatal(err)
	}
	arguments := []string{filepath.Join(repository, "internal/oracle/testdata/process_tty.py"), fmt.Sprint(stdout), fmt.Sprint(stderr), string(encoded)}
	result := execute(t, "python3", append(arguments, command...)...)
	if result.exitCode != 0 {
		t.Fatalf("terminal driver: %s", result.stderr)
	}
	var observed struct {
		Stdout, Stderr []byte
		ExitCode       int
	}
	if err := json.Unmarshal(result.stdout, &observed); err != nil {
		t.Fatal(err)
	}
	return run{stdout: observed.Stdout, stderr: observed.Stderr, exitCode: observed.ExitCode}
}

func TestProcessTerminalsAndEnvironment(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/process_observations.a")
	runner := filepath.Join(repository, "oracle/node.mjs")
	for _, stdout := range []bool{false, true} {
		for _, stderr := range []bool{false, true} {
			for _, environment := range []map[string]string{
				{}, {"NO_COLOR": "", "ADAMIC_PROCESS_TEST": ""},
				{"NO_COLOR": "1", "ADAMIC_PROCESS_TEST": "hé世界"},
				{"FORCE_COLOR": "0", "ADAMIC_PROCESS_TEST": "line\nnext"},
				{"FORCE_COLOR": "3", "ADAMIC_PROCESS_TEST": "value"},
			} {
				t.Run(fmt.Sprintf("stdout=%t/stderr=%t/env=%v", stdout, stderr, environment), func(t *testing.T) {
					truth := processStreams(t, stdout, stderr, environment, "node", "--disable-warning=ExperimentalWarning", runner, path)
					if truth.exitCode != 0 || !bytes.HasPrefix(truth.stdout, []byte(fmt.Sprintf("stdout %s, stderr %s", ttyText(stdout), ttyText(stderr)))) {
						t.Fatalf("Node observation: %+v", truth)
					}
					for _, command := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
						got := processStreams(t, stdout, stderr, environment, command...)
						if difference := disagreement(truth, got); difference != "" {
							t.Errorf("%s: %s; Node %q/%q, actual %q/%q", command[0], difference, truth.stdout, truth.stderr, got.stdout, got.stderr)
						}
					}
				})
			}
		}
	}
}

func ttyText(terminal bool) string {
	if terminal {
		return "true"
	}
	return "undefined"
}

func TestProcessExitArguments(t *testing.T) {
	t.Parallel()
	for _, argument := range []string{"", "undefined", "0", "-1", "256", "257", "4294967297", "9007199254740991"} {
		t.Run("argument="+argument, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte("process.exitCode = 7; console.log('exit argument'); process.exit("+argument+");"), 0o644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			got, _ := natively(t, program)
			if difference := disagreement(truth, got); difference != "" {
				t.Fatalf("%s: Node %d native %d", difference, truth.exitCode, got.exitCode)
			}
			if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
				t.Fatal(difference)
			}
		})
	}
}

func TestProcessObjectsDoNotEscape(t *testing.T) {
	t.Parallel()
	for _, expression := range []string{"process", "process.env", "process.stdout", "process.stderr", "process.exit"} {
		t.Run(expression, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "main.ts")
			if err := os.WriteFile(path, []byte("const held = "+expression+";"), 0o644); err != nil {
				t.Fatal(err)
			}
			_, err := lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) {
				t.Fatalf("want explicit NotYet rather than an uninitialized object, got %v", err)
			}
		})
	}
}

// Every mutant builds cleanly and reaches the ordinary Node comparison, never a sanitizer failure.
func TestProcessMutants(t *testing.T) {
	t.Parallel()
	for _, one := range []struct{ name, fixture, before, after, helper, difference string }{
		{"normal status", "process_exit_code.a", "return adamic_process_status();", "return 0;", "", "exit codes differ"},
		{"immediate status", "process_exit.a", "adamic_process_exit(", "mutant_exit(", "static void mutant_exit(adamic_maybe_number code) { (void)code; adamic_process_exit((adamic_maybe_number){true, 0}); }", "exit codes differ"},
		{"missing environment", "process_observations.a", "adamic_process_environment(", "mutant_env(", "static adamic_string *mutant_env(const adamic_string *key) { (void)key; return adamic_retain(&adamic_string_empty); }", "stdout differs"},
		{"pipe tty", "process_observations.a", "adamic_process_is_tty(", "mutant_tty(", "static adamic_maybe_boolean mutant_tty(enum adamic_stream stream) { (void)stream; return (adamic_maybe_boolean){true, true}; }", "stdout differs"},
		{"range validation", "process_bad_code.a", "adamic_process_set_exit_code(", "mutant_set_code(", "static void mutant_set_code(adamic_maybe_number code) { if (!code.present || (isfinite(code.number) && trunc(code.number) == code.number)) adamic_process_set_exit_code(code); }", "stdout differs"},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", one.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := native.C(program)
			mutant := strings.ReplaceAll(original, one.before, one.after)
			if mutant == original {
				t.Fatal("mutant changed nothing")
			}
			if one.helper != "" {
				mutant = "#include \"adamic.h\"\n" + one.helper + "\n" + mutant
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(mutant, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
			if bytes.Contains(got.stderr, []byte("Sanitizer")) || bytes.Contains(got.stderr, []byte("runtime error:")) {
				t.Fatalf("mutant failed outside comparison: %s", got.stderr)
			}
			if difference := disagreement(onNode(t, path), got); difference != one.difference {
				t.Fatalf("caught by %q, want %q: exit %d stdout %q stderr %q", difference, one.difference, got.exitCode, got.stdout, got.stderr)
			}
			t.Logf("compiled and ran cleanly; caught only by Node %s", one.difference)
		})
	}
}
