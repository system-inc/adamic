package oracle

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// The fixtures compare stdout and stderr each on its own, which can't see the order the two were
// written in, or what happens when stdout has no reader. These run the same programs where those
// show: both streams on one file, and stdout a pipe already closed. Native buffers stdout (adamic.c),
// and these are what hold its flushing to Node's writing each line at once.

// sanitized builds a fixture natively under the sanitizers, and its JavaScript beside it, returning
// the source's path, the binary and the JavaScript.
func sanitized(t *testing.T, fixture string) (string, string, string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("Lower: %v", err)
	}
	directory := t.TempDir()
	binary := filepath.Join(directory, "program")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(directory, "program.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	return path, binary, script
}

// withStreams runs a command with stdout and stderr given as files, and returns its exit code and
// what it wrote to stderr when stderr isn't one of them.
func withStreams(t *testing.T, stdout *os.File, stderr *os.File, name string, arguments ...string) int {
	t.Helper()
	command := bounded(t, name, arguments...)
	command.Stdout = stdout
	command.Stderr = stderr
	if runtime.GOOS == "linux" {
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	command.Wait()
	return command.ProcessState.ExitCode()
}

// Node writes each line at once, so with stdout and stderr on one file the lines land in the order
// the program wrote them. Native must flush stdout before every write to stderr to land the same.
func TestOneFileHoldsNodesOrder(t *testing.T) {
	t.Parallel()
	path, binary, script := sanitized(t, "internal/oracle/testdata/interleaved.a")
	runner := filepath.Join(repository, "oracle", "node.mjs")
	landed := func(name string, arguments ...string) (int, []byte) {
		file, err := os.Create(filepath.Join(t.TempDir(), "both"))
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		exitCode := withStreams(t, file, file, name, arguments...)
		contents, err := os.ReadFile(file.Name())
		if err != nil {
			t.Fatal(err)
		}
		return exitCode, contents
	}
	nodeExit, node := landed("node", "--disable-warning=ExperimentalWarning", runner, path)
	nativeExit, natively := landed(binary)
	backendExit, backend := landed("node", "--disable-warning=ExperimentalWarning", runner, script)
	if !bytes.Contains(node, []byte("out 0\nerr 0\nout 1\n")) {
		t.Fatalf("want Node's lines in the order written, got %.200q", node)
	}
	if nativeExit != nodeExit || !bytes.Equal(natively, node) {
		t.Errorf("native: exit %d, %d bytes; Node: exit %d, %d bytes; first difference at %d", nativeExit, len(natively), nodeExit, len(node), firstDifference(natively, node))
	}
	if backendExit != nodeExit || !bytes.Equal(backend, node) {
		t.Errorf("JavaScript backend: exit %d, %d bytes; Node: exit %d, %d bytes", backendExit, len(backend), nodeExit, len(node))
	}
}

// A pipe whose reader is gone: Node's write fails, the program runs on with what it writes to stdout
// dropped and stderr still written, and the oracle's runtime ends it with exit 70. Native must too,
// rather than die of SIGPIPE.
func TestClosedStdoutEndsAsOnNode(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{
		"internal/oracle/testdata/large_output.a",
		"internal/oracle/testdata/output_then_panic.a",
		"internal/oracle/testdata/interleaved.a",
	} {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			path, binary, _ := sanitized(t, fixture)
			runner := filepath.Join(repository, "oracle", "node.mjs")
			ended := func(name string, arguments ...string) (int, []byte) {
				reader, writer, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				reader.Close()
				defer writer.Close()
				stderr, err := os.Create(filepath.Join(t.TempDir(), "stderr"))
				if err != nil {
					t.Fatal(err)
				}
				defer stderr.Close()
				exitCode := withStreams(t, writer, stderr, name, arguments...)
				said, err := os.ReadFile(stderr.Name())
				if err != nil {
					t.Fatal(err)
				}
				return exitCode, said
			}
			nodeExit, nodeSaid := ended("node", "--disable-warning=ExperimentalWarning", runner, path)
			nativeExit, nativeSaid := ended(binary)
			if nodeExit != 70 {
				t.Fatalf("want Node to end with exit 70 when stdout's reader is gone, got %d, stderr %q", nodeExit, nodeSaid)
			}
			if nativeExit != nodeExit || !bytes.Equal(nativeSaid, nodeSaid) {
				t.Errorf("native: exit %d, stderr %.300q; Node: exit %d, stderr %.300q", nativeExit, nativeSaid, nodeExit, nodeSaid)
			}
		})
	}
}

// firstDifference is the first index where two byte strings differ.
func firstDifference(left []byte, right []byte) int {
	for index := 0; index < min(len(left), len(right)); index++ {
		if left[index] != right[index] {
			return index
		}
	}
	return min(len(left), len(right))
}
