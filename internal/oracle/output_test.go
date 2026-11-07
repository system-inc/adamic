package oracle

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// The fixtures compare stdout and stderr each on its own, which can't see the order the two were
// written in, or what happens when stdout has no reader, is stdin's prompt, or is stopped by a
// signal. These run programs where those show: both streams on one file or one pipe, stdout a pipe
// already closed, stdin answered only after the prompt, and a signal from outside. Native buffers stdout (adamic.c),
// and these are what hold its flushing to Node's writing each line at once.

// sanitized builds a fixture natively under the sanitizers, and its JavaScript beside it, returning
// the source's path, the binary and the JavaScript.
func sanitized(test *testing.T, fixture string) (string, string, string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, fixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatalf("Lower: %v", loweredError)
	}
	directory := test.TempDir()
	binary := filepath.Join(directory, "program")
	if buildError := native.Build(native.C(program), binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	script := filepath.Join(directory, "program.mjs")
	if writeError := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	return path, binary, script
}

// withStreams runs a command with stdout and stderr given as files, and returns its exit code and
// what it wrote to stderr when stderr isn't one of them.
func withStreams(test *testing.T, stdout *os.File, stderr *os.File, name string, arguments ...string) int {
	test.Helper()
	_, command := bounded(test, name, arguments...)
	command.Stdout = stdout
	command.Stderr = stderr
	if runtime.GOOS == "linux" {
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	}
	if startError := command.Start(); startError != nil {
		test.Fatal(startError)
	}
	command.Wait()
	return command.ProcessState.ExitCode()
}

// Node writes each line at once, so with stdout and stderr on one file the lines land in the order
// the program wrote them. Native must flush stdout before every write to stderr to land the same, and
// before a status is taken or a directory listed, since the file looked at may be stdout itself.
func TestOneFileHoldsNodesOrder(test *testing.T) {
	test.Parallel()
	for _, fixture := range []struct {
		path string

		// written is what Node's file must hold, so a run where the case didn't come up can't pass.
		written string
	}{
		{"internal/oracle/testdata/interleaved.a", "out 0\nerr 0\nout 1\n"},
		{"internal/oracle/testdata/status_of_stdout.a", "written before the status\nfile 26\n"},
	} {
		test.Run(fixture.path, func(test *testing.T) {
			test.Parallel()
			cacheProbe(test, fixture.path, nil, "", func() {

				path, binary, script := sanitized(test, fixture.path)
				runner := filepath.Join(repository, "oracle", "node.mjs")
				landed := func(name string, arguments ...string) (int, []byte) {
					file, createError := os.Create(filepath.Join(test.TempDir(), "both"))
					if createError != nil {
						test.Fatal(createError)
					}
					defer file.Close()
					exitCode := withStreams(test, file, file, name, arguments...)
					contents, readError := os.ReadFile(file.Name())
					if readError != nil {
						test.Fatal(readError)
					}
					rememberRun(test, run{stdout: contents, exitCode: exitCode})
					return exitCode, contents
				}
				nodeExit, node := landed("node", "--disable-warning=ExperimentalWarning", runner, path)
				nativeExit, natively := landed(binary)
				backendExit, backend := landed("node", "--disable-warning=ExperimentalWarning", runner, script)
				if !bytes.Contains(node, []byte(fixture.written)) {
					test.Fatalf("want Node's file to hold %q, got %.200q", fixture.written, node)
				}
				if nativeExit != nodeExit || !bytes.Equal(natively, node) {
					test.Errorf(
						"native: exit %d, %q; Node: exit %d, %q; first difference at %d",
						nativeExit,
						natively,
						nodeExit,
						node,
						firstDifference(natively, node),
					)
				}
				if backendExit != nodeExit || !bytes.Equal(backend, node) {
					test.Errorf(
						"JavaScript backend: exit %d, %d bytes; Node: exit %d, %d bytes",
						backendExit,
						len(backend),
						nodeExit,
						len(node),
					)
				}

			})
		})
	}
}

// A pipe whose reader is gone: Node's write fails, the program runs on with what it writes to stdout
// dropped and stderr still written, and the oracle's runtime ends it with exit 70. Native must too,
// rather than die of SIGPIPE.
func TestClosedStdoutEndsAsOnNode(test *testing.T) {
	test.Parallel()
	for _, fixture := range []string{
		"internal/oracle/testdata/host_stdout/closed.a",
		"internal/oracle/testdata/large_output.a",
		"internal/oracle/testdata/output_then_panic.a",
		"internal/oracle/testdata/interleaved.a",
	} {
		test.Run(fixture, func(test *testing.T) {
			test.Parallel()
			cacheProbe(test, fixture, nil, "", func() {

				path, binary, _ := sanitized(test, fixture)
				runner := filepath.Join(repository, "oracle", "node.mjs")
				ended := func(name string, arguments ...string) (int, []byte) {
					reader, writer, pipeError := os.Pipe()
					if pipeError != nil {
						test.Fatal(pipeError)
					}
					reader.Close()
					defer writer.Close()
					stderr, createError := os.Create(filepath.Join(test.TempDir(), "stderr"))
					if createError != nil {
						test.Fatal(createError)
					}
					defer stderr.Close()
					exitCode := withStreams(test, writer, stderr, name, arguments...)
					said, readError := os.ReadFile(stderr.Name())
					if readError != nil {
						test.Fatal(readError)
					}
					rememberRun(test, run{stderr: said, exitCode: exitCode})
					return exitCode, said
				}
				nodeExit, nodeSaid := ended("node", "--disable-warning=ExperimentalWarning", runner, path)
				nativeExit, nativeSaid := ended(binary)
				if nodeExit != 70 {
					test.Fatalf(
						"want Node to end with exit 70 when stdout's reader is gone, got %d, stderr %q",
						nodeExit,
						nodeSaid,
					)
				}
				if nativeExit != nodeExit || !bytes.Equal(nativeSaid, nodeSaid) {
					test.Errorf(
						"native: exit %d, stderr %.300q; Node: exit %d, stderr %.300q",
						nativeExit,
						nativeSaid,
						nodeExit,
						nodeSaid,
					)
				}

			})
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

// onePipe runs a command with stdout and stderr on one pipe, and returns its exit code and what came
// through it, in the order it came.
func onePipe(test *testing.T, name string, arguments ...string) (int, []byte) {
	test.Helper()
	_, command := bounded(test, name, arguments...)
	var both bytes.Buffer
	// The same writer for both: os/exec gives them one pipe.
	command.Stdout = &both
	command.Stderr = &both
	// WaitDelay: the order the two streams interleave in is what this compares, and
	// childguard.Capture reads each stream from a pipe of its own, so only exec's one shared pipe
	// keeps it; the delay bounds a pipe still held after the exit until childguard can capture both
	// streams on one pipe.
	command.WaitDelay = 5 * time.Second
	if runtime.GOOS == "linux" {
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	}
	if runError := command.Run(); runError != nil && command.ProcessState == nil {
		test.Fatal(runError)
	}
	rememberRun(test, run{stdout: both.Bytes(), exitCode: command.ProcessState.ExitCode()})
	return command.ProcessState.ExitCode(), both.Bytes()
}

// A file written may be stdout or stderr themselves, so what was printed before it has to be out
// first, as Node's is, or the file's text overtakes it.
func TestFileWritesLandInNodesOrder(test *testing.T) {
	test.Parallel()
	for _, fixture := range []string{
		"internal/oracle/testdata/write_stdout_order.a",
		"internal/oracle/testdata/write_stderr_order.a",
	} {
		test.Run(fixture, func(test *testing.T) {
			test.Parallel()
			cacheProbe(test, fixture, nil, "", func() {

				path, binary, script := sanitized(test, fixture)
				runner := filepath.Join(repository, "oracle", "node.mjs")
				nodeExit, node := onePipe(test, "node", "--disable-warning=ExperimentalWarning", runner, path)
				if nodeExit != 0 || !bytes.Equal(node, []byte("first\nsecond\nthird\n")) {
					test.Fatalf("want Node's lines in the order written, got exit %d, %q", nodeExit, node)
				}
				for _, other := range [][]string{
					{binary},
					{"node", "--disable-warning=ExperimentalWarning", runner, script},
				} {
					exitCode, landed := onePipe(test, other[0], other[1:]...)
					if exitCode != nodeExit || !bytes.Equal(landed, node) {
						test.Errorf(
							"%s: exit %d, %q; Node: exit %d, %q",
							filepath.Base(other[len(other)-1]),
							exitCode,
							landed,
							nodeExit,
							node,
						)
					}
				}

			})
		})
	}
}

// A prompt, then a read of stdin: a driver waits to see the prompt before it answers, so the prompt
// has to be out before the read waits. Each run is answered once its first line has come, or once it
// has printed nothing and sat blocked for a second, asleep with its CPU time unchanged, which is the
// failure: it's waiting for the read. A process starved by load is runnable, not asleep, so load
// can't fake it, and a real hang still ends at bounded's deadline with no prompt.
func TestAPromptComesBeforeTheRead(test *testing.T) {
	test.Parallel()
	cacheProbe(test, "internal/oracle/testdata/prompt_then_read.a", nil, "", func() {

		path, binary, script := sanitized(test, "internal/oracle/testdata/prompt_then_read.a")
		runner := filepath.Join(repository, "oracle", "node.mjs")
		converse := func(name string, arguments ...string) (bool, string) {
			_, command := bounded(test, name, arguments...)
			if runtime.GOOS == "linux" {
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
			}
			stdin, stdinPipeError := command.StdinPipe()
			if stdinPipeError != nil {
				test.Fatal(stdinPipeError)
			}
			stdout, stdoutPipeError := command.StdoutPipe()
			if stdoutPipeError != nil {
				test.Fatal(stdoutPipeError)
			}
			if startError := command.Start(); startError != nil {
				test.Fatal(startError)
			}
			reader := bufio.NewReader(stdout)
			first := make(chan string, 1)
			go func() {
				line, _ := reader.ReadString('\n')
				first <- line
			}()
			prompted, received := false, false
			var prompt string
			for idle, spent := 0, -1.0; !received && idle < 20; {
				select {
				case prompt = <-first:
					// A whole line; at an exit without one there's no prompt.
					prompted, received = strings.HasSuffix(prompt, "\n"), true
				case <-time.After(50 * time.Millisecond):
					now := cpuSeconds(test, command.Process.Pid)
					if now == spent && asleep(test, command.Process.Pid) {
						idle++
					} else {
						idle = 0
					}
					spent = now
				}
			}
			stdin.Write([]byte("yes\n"))
			stdin.Close()
			if !received {
				prompt = <-first
			}
			rest, _ := io.ReadAll(reader)
			command.Wait()
			rememberRun(test, run{stdout: []byte(prompt + string(rest)), exitCode: command.ProcessState.ExitCode()})
			return prompted, prompt + string(rest)
		}
		nodePrompted, node := converse("node", "--disable-warning=ExperimentalWarning", runner, path)
		if !nodePrompted || node != "ready\ngot yes\n" {
			test.Fatalf("want Node to prompt before reading, then answer: prompted %t, %q", nodePrompted, node)
		}
		for _, other := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
			prompted, said := converse(other[0], other[1:]...)
			if !prompted || said != node {
				test.Errorf(
					"%s: prompted before the read %t, said %q; Node said %q",
					filepath.Base(other[len(other)-1]),
					prompted,
					said,
					node,
				)
			}
		}

	})
}

// Each signal, inherited disposition and process leg is a bounded top-level test.
// Native readiness uses CPU time, since its line stays buffered until termination.
// lockedBuffer is a child's output, read while the child still writes it.
type lockedBuffer struct {
	lock   sync.Mutex
	buffer bytes.Buffer
}

func (b *lockedBuffer) Write(data []byte) (int, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.buffer.Write(data)
}

func (b *lockedBuffer) Bytes() []byte {
	b.lock.Lock()
	defer b.lock.Unlock()
	return bytes.Clone(b.buffer.Bytes())
}

func (b *lockedBuffer) String() string { return string(b.Bytes()) }

// asleep reports whether a running process is blocked, every thread waiting (state S on Linux; S or
// I, idle, from ps on macOS), rather than running or waiting to run.
func asleep(test *testing.T, pid int) bool {
	test.Helper()
	if runtime.GOOS == "linux" {
		data, readError := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if readError != nil {
			return false
		}
		fields := strings.Fields(string(data[bytes.LastIndexByte(data, ')')+1:]))
		return len(fields) != 0 && fields[0] == "S"
	}
	output, outputError := exec.Command("ps", "-o", "state=", "-p", strconv.Itoa(pid)).Output()
	if outputError != nil {
		return false
	}
	state := strings.TrimSpace(string(output))
	return strings.HasPrefix(state, "S") || strings.HasPrefix(state, "I")
}

// cpuSeconds is the user and system CPU time a running process has spent, from /proc on Linux
// (in clock ticks of 1/100 s) and ps elsewhere. A process that has ended reads as 0.
func cpuSeconds(test *testing.T, pid int) float64 {
	test.Helper()
	if runtime.GOOS == "linux" {
		data, readError := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if readError != nil {
			return 0
		}
		// Fields after the command name, which is in parentheses and may hold spaces: utime and
		// stime are the 14th and 15th fields of the line, the 12th and 13th after the name.
		fields := strings.Fields(string(data[bytes.LastIndexByte(data, ')')+1:]))
		user, _ := strconv.ParseFloat(fields[11], 64)
		system, _ := strconv.ParseFloat(fields[12], 64)
		return (user + system) / 100
	}
	output, outputError := exec.Command("ps", "-o", "time=", "-p", strconv.Itoa(pid)).Output()
	if outputError != nil {
		return 0
	}
	// macOS prints minutes:seconds.hundredths, as 0:00.52.
	minutes, seconds, _ := strings.Cut(strings.TrimSpace(string(output)), ":")
	whole, _ := strconv.ParseFloat(minutes, 64)
	part, _ := strconv.ParseFloat(seconds, 64)
	return whole*60 + part
}
