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

func stdoutWriteAgreement(t *testing.T, name string) {
	t.Helper()
	var rows []int
	for i, f := range fixtures {
		if f.path == "internal/oracle/testdata/host_stdout/"+name+".a" {
			rows = append(rows, i)
		}
	}
	if len(rows) != 1 {
		t.Fatal("stdout fixture must be registered exactly once")
	}
	// Ordinary oracle execution captures a pipe, including the interleaved writes.
	for _, i := range rows {
		fixture := fixtures[i]
		path, err := filepath.Abs(filepath.Join(repository, fixture.path))
		if err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		truth := onNode(t, path)
		js := onJavaScriptBackend(t, program)
		native, _ := nativelyUncached(t, program)
		if d := disagreement(truth, js); d != "" {
			t.Fatal("javascript: " + d)
		}
		if d := disagreement(truth, native); d != "" {
			t.Fatal("native: " + d)
		}
	}
}
func TestStdoutWriteTextMatchesNode(t *testing.T)     { t.Parallel(); stdoutWriteAgreement(t, "text") }
func TestStdoutWriteLargeMatchesNode(t *testing.T)    { t.Parallel(); stdoutWriteAgreement(t, "large") }
func TestStdoutWriteOrderingMatchesNode(t *testing.T) { t.Parallel(); stdoutWriteAgreement(t, "order") }
func TestStdoutWriteClearScreenMatchesNode(t *testing.T) {
	t.Parallel()
	stdoutWriteAgreement(t, "clear")
}

// No console call may accidentally register teardown for this fixture.
func TestStdoutWriteClosedPipeMatchesNode(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/host_stdout/closed.a")

	for _, backend := range []struct {
		name string
		args []string
	}{
		{"node", []string{"--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path}},
		{binary, nil},
		{"node", []string{"--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), script}},
	} {
		stdoutClosedPipe(t, backend.name, backend.args...)
	}
}

func stdoutClosedPipe(t *testing.T, name string, arguments ...string) {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	reader.Close()
	stderr, err := os.Create(filepath.Join(t.TempDir(), "stderr"))
	if err != nil {
		t.Fatal(err)
	}
	code := withStreams(t, writer, stderr, name, arguments...)
	writer.Close()
	stderr.Close()
	said, err := os.ReadFile(stderr.Name())
	if err != nil {
		t.Fatal(err)
	}
	if code != 70 || len(said) != 0 {
		t.Fatalf("%s closed stdout: exit %d stderr %q", name, code, said)
	}
}

// Called only by the existing opt-in WASI family, after its setup and gate.
func stdoutWASIClosedPipe(t *testing.T, source string) {
	t.Helper()
	wasm := filepath.Join(t.TempDir(), "closed.wasm")
	if err := native.Build(source, wasm, native.Options{Target: "wasm32-wasi"}); err != nil {
		t.Fatal(err)
	}
	stdoutClosedPipe(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/wasi.mjs"), wasm)
}

func TestStdoutWriteParenthesizedReceiverMatchesNode(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "main.a")
	source := "import process from 'node:process';\n(process.stdout).write('x');\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "x" || len(truth.stderr) != 0 {
		t.Fatalf("Node witness changed: %+v", truth)
	}
	js := onJavaScriptBackend(t, program)
	actual, _ := nativelyUncached(t, program)
	if d := disagreement(truth, js); d != "" {
		t.Fatal("javascript: " + d)
	}
	if d := disagreement(truth, actual); d != "" {
		t.Fatal("native: " + d)
	}
}
