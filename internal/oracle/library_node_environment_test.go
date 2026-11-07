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
func processEnvironmentRun(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	ctx, command := bounded(test, name, arguments...)
	command.Env = append([]string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, environment...)
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	var exited *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exited) {
		test.Fatal(captureError)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
}

func processEnvironmentBackends(test *testing.T, fixture string, environment []string, expected run) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join("testdata/process_env", fixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	for _, mode := range []struct {
		name    string
		options native.Options
	}{
		{"native", native.Options{}}, {"ASan", native.Options{Sanitize: true}},
	} {
		binary := filepath.Join(test.TempDir(), mode.name)
		if buildError := native.Build(native.C(program), binary, mode.options); buildError != nil {
			test.Fatal(buildError)
		}
		got := processEnvironmentRun(test, environment, binary)
		if difference := disagreement(expected, got); difference != "" {
			test.Fatalf(
				"%s: %s; got stdout=%q stderr=%q exit=%d",
				mode.name,
				difference,
				got.stdout,
				got.stderr,
				got.exitCode,
			)
		}
	}
	js := filepath.Join(test.TempDir(), "program.mjs")
	if writeError := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	got := processEnvironmentRun(
		test,
		environment,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		js,
	)
	if difference := disagreement(expected, got); difference != "" {
		test.Fatalf("JavaScript: %s; got stdout=%q stderr=%q exit=%d", difference, got.stdout, got.stderr, got.exitCode)
	}
}

func TestProcessEnvironmentReadsMatchNode(test *testing.T) {
	test.Parallel()
	env := processEnvironmentFixtureEnvironment("reads.a")
	path, _ := filepath.Abs("testdata/process_env/reads.a")
	truth := processEnvironmentRun(
		test,
		env,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		path,
	)
	want := []byte(
		"set\nundefined\n\nhéllo 世界 🌍\nset\nset\n\n\nset\ntrue\nset\nundefined\nkey evaluated once\nset\nlocal\n",
	)
	if truth.exitCode != 0 || !bytes.Equal(truth.stdout, want) || len(truth.stderr) != 0 {
		test.Fatalf("Node witness changed: %+v", truth)
	}
	processEnvironmentBackends(test, "reads.a", env, truth)
}

func TestProcessEnvironmentOwnPrototypeEntry(test *testing.T) {
	test.Parallel()
	env := processEnvironmentFixtureEnvironment("prototype.a")
	path, _ := filepath.Abs("testdata/process_env/prototype.a")
	truth := processEnvironmentRun(
		test,
		env,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		path,
	)
	if truth.exitCode != 0 ||
		string(truth.stdout) != "own environment string\nown environment string\n" ||
		len(truth.stderr) != 0 {
		test.Fatalf("Node own entry: %+v", truth)
	}
	processEnvironmentBackends(test, "prototype.a", env, truth)
}

func TestProcessEnvironmentInheritedKeyPanics(test *testing.T) {
	test.Parallel()
	// Node returns a function, outside the declared string | undefined boundary.
	// The ruled compiler check panics, identically on both backends, rather than lie.
	truth := processEnvironmentRun(test, nil, "node", "-e", "console.log(typeof process.env['toString'])")
	if truth.exitCode != 0 || string(truth.stdout) != "function\n" || len(truth.stderr) != 0 {
		test.Fatalf("Node prototype witness: %+v", truth)
	}
	expected := run{stderr: []byte("adamic: panic: node:process.Process.env inherited key: toString\n"), exitCode: 70}
	processEnvironmentBackends(test, "prototype.a", nil, expected)
}

func TestProcessEnvironmentNULFixtureMatchesNode(test *testing.T) {
	test.Parallel()
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
		test.Run(one.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs("testdata/process_env/nul.a")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			truth := processEnvironmentRun(
				test,
				one.environment,
				"node",
				"--disable-warning=ExperimentalWarning",
				filepath.Join(repository, "oracle/node.mjs"),
				path,
			)
			if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != one.output {
				test.Fatalf("Node NUL-name witness changed: %+v", truth)
			}
			test.Logf("Node stdout: %q", truth.stdout)
			processEnvironmentBackends(test, "nul.a", one.environment, truth)
		})
	}
}

func processEnvironmentMutationAgreement(test *testing.T, fixture, want string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join("testdata/process_env", fixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := processEnvironmentRun(
		test,
		nil,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		path,
	)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || string(truth.stdout) != want {
		test.Fatalf("Node mutation witness changed: %+v", truth)
	}
	processEnvironmentBackends(test, fixture, nil, truth)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	binary := filepath.Join(test.TempDir(), "counted")
	if buildError := native.Build(native.C(program), binary, native.Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	got := processEnvironmentRun(test, nil, binary)
	if got.exitCode != truth.exitCode ||
		!bytes.Equal(got.stdout, truth.stdout) ||
		countsLine.FindSubmatch(got.stderr) == nil {
		test.Fatalf("counted mutation disagrees: %+v", got)
	}
}

func TestProcessEnvironmentAdmittedUsesMatchNode(test *testing.T) {
	test.Parallel()
	processEnvironmentMutationAgreement(
		test,
		"mutations.a",
		"first\n\nhéllo 世界 🌍\ntrue\nmissing\ntrue\nown environment string\ntrue\n",
	)
}
func TestProcessEnvironmentMutationValuesMatchNode(test *testing.T) {
	test.Parallel()
	processEnvironmentMutationAgreement(test, "mutation_values.a", "a\na\ntrue\nmissing\n")
}
func TestProcessEnvironmentSpecialNamesMatchNode(test *testing.T) {
	test.Parallel()
	processEnvironmentMutationAgreement(
		test,
		"mutation_special_names.a",
		"1\nmissing\nmissing\ntrue\nmissing\nmissing\n1\nx\nx\ntrue\nmissing\nmissing\n3\na\na\ntrue\nmissing\nmissing\n1\nmissing\nmissing\ntrue\nmissing\nmissing\n",
	)
}
