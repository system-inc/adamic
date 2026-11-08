package oracle

import (
	"github.com/system-inc/adamic/internal/native"
	"os"
	"path/filepath"
	"testing"
)

var stdoutWriteFixtures = []string{"text", "large", "order", "clear", "closed"}

func init() {
	for _, name := range stdoutWriteFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/host_stdout/" + name + ".a", true, false})
	}
}

func stdoutWriteAgreement(test *testing.T, name string) {
	test.Helper()
	var rows []int
	for index, fixture := range fixtures {
		if fixture.path == "internal/oracle/testdata/host_stdout/"+name+".a" {
			rows = append(rows, index)
		}
	}
	if len(rows) != 1 {
		test.Fatal("stdout fixture must be registered exactly once")
	}
	// Ordinary oracle execution captures a pipe, including the interleaved writes.
	for _, row := range rows {
		fixture := fixtures[row]
		path, absolutePathError := filepath.Abs(filepath.Join(repository, fixture.path))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		program, loweredError := lowered(test, path)
		if loweredError != nil {
			test.Fatal(loweredError)
		}
		truth := onNode(test, path)
		js := onJavaScriptBackend(test, program)
		native, _ := nativelyUncached(test, program)
		if d := disagreement(truth, js); d != "" {
			test.Fatal("javascript: " + d)
		}
		if d := disagreement(truth, native); d != "" {
			test.Fatal("native: " + d)
		}
	}
}
func TestStdoutWriteTextMatchesNode(test *testing.T) {
	test.Parallel()
	stdoutWriteAgreement(test, "text")
}
func TestStdoutWriteLargeMatchesNode(test *testing.T) {
	test.Parallel()
	stdoutWriteAgreement(test, "large")
}
func TestStdoutWriteOrderingMatchesNode(test *testing.T) {
	test.Parallel()
	stdoutWriteAgreement(test, "order")
}
func TestStdoutWriteClearScreenMatchesNode(test *testing.T) {
	test.Parallel()
	stdoutWriteAgreement(test, "clear")
}

// No console call may accidentally register teardown for this fixture.
func TestStdoutWriteClosedPipeMatchesNode(test *testing.T) {
	test.Parallel()
	path, binary, script := sanitized(test, "internal/oracle/testdata/host_stdout/closed.a")

	for _, backend := range []struct {
		name string
		args []string
	}{
		{"node", []string{"--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path}},
		{binary, nil},
		{
			"node",
			[]string{"--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script},
		},
	} {
		stdoutClosedPipe(test, backend.name, backend.args...)
	}
}

func stdoutClosedPipe(test *testing.T, name string, arguments ...string) {
	test.Helper()

	reader, writer, pipeError := os.Pipe()
	if pipeError != nil {
		test.Fatal(pipeError)
	}
	reader.Close()
	stderr, createError := os.Create(filepath.Join(test.TempDir(), "stderr"))
	if createError != nil {
		test.Fatal(createError)
	}
	code := withStreams(test, writer, stderr, name, arguments...)
	writer.Close()
	stderr.Close()
	said, readError := os.ReadFile(stderr.Name())
	if readError != nil {
		test.Fatal(readError)
	}
	if code != 70 || len(said) != 0 {
		test.Fatalf("%s closed stdout: exit %d stderr %q", name, code, said)
	}
}

// Called only by the existing opt-in WASI family, after its setup and gate.
func stdoutWasiClosedPipe(test *testing.T, source string) {
	test.Helper()
	wasm := filepath.Join(test.TempDir(), "closed.wasm")
	if buildError := native.Build(source, wasm, native.Options{Target: "wasm32-wasi"}); buildError != nil {
		test.Fatal(buildError)
	}
	stdoutClosedPipe(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/wasi.mjs"),
		wasm,
	)
}

func TestStdoutWriteParenthesizedReceiverMatchesNode(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "main.a")
	source := "import process from 'node:process';\n(process.stdout).write('x');\n"
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	truth := onNode(test, path)
	if truth.exitCode != 0 || string(truth.stdout) != "x" || len(truth.stderr) != 0 {
		test.Fatalf("Node witness changed: %+v", truth)
	}
	js := onJavaScriptBackend(test, program)
	actual, _ := nativelyUncached(test, program)
	if d := disagreement(truth, js); d != "" {
		test.Fatal("javascript: " + d)
	}
	if d := disagreement(truth, actual); d != "" {
		test.Fatal("native: " + d)
	}
}
