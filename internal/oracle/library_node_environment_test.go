package oracle

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// These host fixtures have their own explicit-environment harness: no inherited
// box variables, including for the missing key and the Object.prototype witness.
func processEnvironmentRun(t *testing.T, environment []string, name string, arguments ...string) run {
	t.Helper()
	ctx, command := bounded(t, name, arguments...)
	command.Env = append([]string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, environment...)
	stdout, stderr, err := childguard.Capture(ctx, command)
	var exited *exec.ExitError
	if err != nil && !errors.As(err, &exited) {
		t.Fatal(err)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
}

func processEnvironmentBackends(t *testing.T, fixture string, environment []string, expected run) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("testdata/process_env", fixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := processEnvironmentRun(t, environment, binary)
	if difference := disagreement(expected, got); difference != "" {
		t.Fatalf("native: %s; got stdout=%q stderr=%q exit=%d", difference, got.stdout, got.stderr, got.exitCode)
	}
	js := filepath.Join(t.TempDir(), "program.mjs")
	if err := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	got = processEnvironmentRun(t, environment, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), js)
	if difference := disagreement(expected, got); difference != "" {
		t.Fatalf("JavaScript: %s; got stdout=%q stderr=%q exit=%d", difference, got.stdout, got.stderr, got.exitCode)
	}
}

func TestProcessEnvironmentReadsMatchNode(t *testing.T) {
	t.Parallel()
	env := processEnvironmentFixtureEnvironment("reads.a")
	path, _ := filepath.Abs("testdata/process_env/reads.a")
	truth := processEnvironmentRun(t, env, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	want := []byte("set\nundefined\n\nhéllo 世界 🌍\nset\nset\n\n\nset\ntrue\nset\nundefined\nkey evaluated once\nset\nlocal\n")
	if truth.exitCode != 0 || !bytes.Equal(truth.stdout, want) || len(truth.stderr) != 0 {
		t.Fatalf("Node witness changed: %+v", truth)
	}
	processEnvironmentBackends(t, "reads.a", env, truth)
}

func TestProcessEnvironmentOwnPrototypeEntry(t *testing.T) {
	t.Parallel()
	env := processEnvironmentFixtureEnvironment("prototype.a")
	path, _ := filepath.Abs("testdata/process_env/prototype.a")
	truth := processEnvironmentRun(t, env, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	if truth.exitCode != 0 || string(truth.stdout) != "own environment string\nown environment string\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node own entry: %+v", truth)
	}
	processEnvironmentBackends(t, "prototype.a", env, truth)
}

func TestProcessEnvironmentInheritedKeyPanics(t *testing.T) {
	t.Parallel()
	// Node returns a function, outside the declared string | undefined boundary.
	// The ruled compiler check panics, identically on both backends, rather than lie.
	truth := processEnvironmentRun(t, nil, "node", "-e", "console.log(typeof process.env['toString'])")
	if truth.exitCode != 0 || string(truth.stdout) != "function\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node prototype witness: %+v", truth)
	}
	expected := run{stderr: []byte("adamic: panic: node:process.Process.env inherited key: toString\n"), exitCode: 70}
	processEnvironmentBackends(t, "prototype.a", nil, expected)
}

func TestProcessEnvironmentNULFixtureMatchesNode(t *testing.T) {
	t.Parallel()
	for _, one := range []struct {
		name        string
		environment []string
		output      string
	}{
		{"nul", processEnvironmentFixtureEnvironment("nul.a"), "undefined\nset\nundefined\nset\n"},
		// A non-NUL runtime separator changes only the last two reads. This
		// neighbor catches folding the runtime name to the literal NUL case.
		{"runtime_neighbor", []string{"A=set", "ADAMIC_ENV_NUL_CODE=65"}, "undefined\nset\nundefined\nundefined\n"},
	} {
		t.Run(one.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs("testdata/process_env/nul.a")
			if err != nil {
				t.Fatal(err)
			}
			truth := processEnvironmentRun(t, one.environment, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != one.output {
				t.Fatalf("Node NUL-name witness changed: %+v", truth)
			}
			t.Logf("Node stdout: %q", truth.stdout)
			processEnvironmentBackends(t, "nul.a", one.environment, truth)
		})
	}
}
