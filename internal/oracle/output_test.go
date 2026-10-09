package oracle

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
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
	checkChild(t, runChildWith(command, childguard.Options{KeepFiles: true}))
	return command.ProcessState.ExitCode()
}

// Node writes each line at once, so with stdout and stderr on one file the lines land in the order
// the program wrote them. Native must flush stdout before every write to stderr to land the same, and
// before a status is taken or a directory listed, since the file looked at may be stdout itself.
func TestOneFileHoldsNodesOrder(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct {
		path string

		// written is what Node's file must hold, so a run where the case didn't come up can't pass.
		written string
	}{
		{"internal/oracle/testdata/interleaved.a", "out 0\nerr 0\nout 1\n"},
		{"internal/oracle/testdata/status_of_stdout.a", "written before the status\nfile 26\n"},
	} {
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			cacheProbe(t, fixture.path, nil, "", func() {

				path, binary, script := sanitized(t, fixture.path)
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
					rememberRun(t, run{stdout: contents, exitCode: exitCode})
					return exitCode, contents
				}
				nodeExit, node := landed("node", "--disable-warning=ExperimentalWarning", runner, path)
				nativeExit, natively := landed(binary)
				backendExit, backend := landed("node", "--disable-warning=ExperimentalWarning", runner, script)
				if !bytes.Contains(node, []byte(fixture.written)) {
					t.Fatalf("want Node's file to hold %q, got %.200q", fixture.written, node)
				}
				if nativeExit != nodeExit || !bytes.Equal(natively, node) {
					t.Errorf("native: exit %d, %q; Node: exit %d, %q; first difference at %d", nativeExit, natively, nodeExit, node, firstDifference(natively, node))
				}
				if backendExit != nodeExit || !bytes.Equal(backend, node) {
					t.Errorf("JavaScript backend: exit %d, %d bytes; Node: exit %d, %d bytes", backendExit, len(backend), nodeExit, len(node))
				}

			})
		})
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
			cacheProbe(t, fixture, nil, "", func() {

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
					rememberRun(t, run{stderr: said, exitCode: exitCode})
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
func onePipe(t *testing.T, name string, arguments ...string) (int, []byte) {
	t.Helper()
	command := bounded(t, name, arguments...)
	var both bytes.Buffer
	// The same writer for both: os/exec gives them one pipe.
	command.Stdout = &both
	command.Stderr = &both
	if runtime.GOOS == "linux" {
		command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
	}
	checkChild(t, runChild(command))
	rememberRun(t, run{stdout: both.Bytes(), exitCode: command.ProcessState.ExitCode()})
	return command.ProcessState.ExitCode(), both.Bytes()
}

// A file written may be stdout or stderr themselves, so what was printed before it has to be out
// first, as Node's is, or the file's text overtakes it.
func TestFileWritesLandInNodesOrder(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{
		"internal/oracle/testdata/write_stdout_order.a",
		"internal/oracle/testdata/write_stderr_order.a",
	} {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			cacheProbe(t, fixture, nil, "", func() {

				path, binary, script := sanitized(t, fixture)
				runner := filepath.Join(repository, "oracle", "node.mjs")
				nodeExit, node := onePipe(t, "node", "--disable-warning=ExperimentalWarning", runner, path)
				if nodeExit != 0 || !bytes.Equal(node, []byte("first\nsecond\nthird\n")) {
					t.Fatalf("want Node's lines in the order written, got exit %d, %q", nodeExit, node)
				}
				for _, other := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
					exitCode, landed := onePipe(t, other[0], other[1:]...)
					if exitCode != nodeExit || !bytes.Equal(landed, node) {
						t.Errorf("%s: exit %d, %q; Node: exit %d, %q", filepath.Base(other[len(other)-1]), exitCode, landed, nodeExit, node)
					}
				}

			})
		})
	}
}

// A prompt, then a read of stdin: a driver waits to see the prompt before it answers, so the prompt
// has to be out before the read waits. Each run is answered only once its first line has come, or
// the shared guard reports a stall, which is the failure.
func TestAPromptComesBeforeTheRead(t *testing.T) {
	t.Parallel()
	cacheProbe(t, "internal/oracle/testdata/prompt_then_read.a", nil, "", func() {

		path, binary, script := sanitized(t, "internal/oracle/testdata/prompt_then_read.a")
		runner := filepath.Join(repository, "oracle", "node.mjs")
		converse := func(name string, arguments ...string) (bool, string) {
			command := bounded(t, name, arguments...)
			if runtime.GOOS == "linux" {
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
			}
			stdin, err := command.StdinPipe()
			if err != nil {
				t.Fatal(err)
			}
			var output promptOutput
			output.ready = make(chan struct{})
			command.Stdout = &output
			done := startChild(t, command, childguard.Options{})
			prompted := false
			select {
			case <-output.ready:
				prompted = true
			case err := <-done:
				checkChild(t, err)
			}
			if prompted {
				if _, err := stdin.Write([]byte("yes\n")); err != nil {
					t.Fatal(err)
				}
				stdin.Close()
				checkChild(t, <-done)
			}
			said := output.text.String()
			rememberRun(t, run{stdout: []byte(said), exitCode: command.ProcessState.ExitCode()})
			return prompted, said
		}
		nodePrompted, node := converse("node", "--disable-warning=ExperimentalWarning", runner, path)
		if !nodePrompted || node != "ready\ngot yes\n" {
			t.Fatalf("want Node to prompt before reading, then answer: prompted %t, %q", nodePrompted, node)
		}
		for _, other := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
			prompted, said := converse(other[0], other[1:]...)
			if !prompted || said != node {
				t.Errorf("%s: prompted before the read %t, said %q; Node said %q", filepath.Base(other[len(other)-1]), prompted, said, node)
			}
		}

	})
}

// A program stopped from outside by SIGTERM, SIGINT or SIGHUP: Node has written every line, and is
// killed by the signal. Native must write out what its buffer holds and be killed by the same signal,
// under the sanitizers too. No run is signaled after a fixed delay, which a loaded machine can
// outrun: Node and the JavaScript backend are signaled once their line is on the pipe, and native,
// whose line stays in its buffer until the signal, once it has spent half a second of its own CPU
// time, which only the fixture's spin after the line can spend. Load can't fake CPU time.
// Started with the three signals ignored by its parent (as under nohup, or a background job), Node
// resets them at startup and is stopped all the same, so native must be too.
func TestASignalLeavesWhatWasPrinted(t *testing.T) {
	t.Parallel()
	for _, ignored := range []bool{false, true} {
		for _, stop := range []syscall.Signal{syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP} {
			name := stop.String()
			if ignored {
				name += " inherited ignored"
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				cacheProbe(t, "internal/oracle/testdata/killed_after_output.a", nil, "", func() {

					path, binary, script := sanitized(t, "internal/oracle/testdata/killed_after_output.a")
					runner := filepath.Join(repository, "oracle", "node.mjs")
					stopped := func(name string, arguments ...string) (string, []byte) {
						onNode, label := name == "node", filepath.Base(append([]string{name}, arguments...)[len(arguments)])
						if ignored {
							// sh ignores the three, then execs the program in its place, with the same pid.
							arguments = append([]string{"-c", `trap "" TERM INT HUP; exec "$0" "$@"`, name}, arguments...)
							name = "sh"
						}
						command := bounded(t, name, arguments...)
						if runtime.GOOS == "linux" {
							command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0")
						}
						var stdout, stderr lockedBuffer
						command.Stdout, command.Stderr = &stdout, &stderr
						done := startChild(t, command, childguard.Options{})
						ready := func() bool {
							if onNode {
								return bytes.Contains(stdout.Bytes(), []byte("\n"))
							}
							return cpuSeconds(t, command.Process.Pid) >= 0.5
						}
						for !ready() {
							select {
							case err := <-done:
								checkChild(t, err)
								t.Fatalf("%s exited before reaching its spin", label)
							case <-time.After(10 * time.Millisecond):
							}
						}
						command.Process.Signal(stop)
						checkChild(t, <-done)
						status := command.ProcessState.Sys().(syscall.WaitStatus)
						ended := fmt.Sprintf("exit %d", status.ExitStatus())
						if status.Signaled() {
							ended = "killed by " + status.Signal().String()
						}
						rememberRun(t, run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()})
						return ended + ", stderr " + strconv.Quote(stderr.String()), stdout.Bytes()
					}
					nodeEnded, node := stopped("node", "--disable-warning=ExperimentalWarning", runner, path)
					if nodeEnded != "killed by "+stop.String()+`, stderr ""` || len(node) == 0 {
						t.Fatalf("want Node killed by %s after its line, got %s, stdout %q", stop, nodeEnded, node)
					}
					for _, other := range [][]string{{binary}, {"node", "--disable-warning=ExperimentalWarning", runner, script}} {
						ended, printed := stopped(other[0], other[1:]...)
						if ended != nodeEnded || !bytes.Equal(printed, node) {
							t.Errorf("%s: %s, stdout %q; Node: %s, stdout %q", filepath.Base(other[len(other)-1]), ended, printed, nodeEnded, node)
						}
					}

				})
			})
		}
	}
}

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

// cpuSeconds is the user and system CPU time a running process has spent, from /proc on Linux
// (in clock ticks of 1/100 s) and ps elsewhere. A process that has ended reads as 0.
func cpuSeconds(t *testing.T, pid int) float64 {
	t.Helper()
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			return 0
		}
		// Fields after the command name, which is in parentheses and may hold spaces: utime and
		// stime are the 14th and 15th fields of the line, the 12th and 13th after the name.
		fields := strings.Fields(string(data[bytes.LastIndexByte(data, ')')+1:]))
		user, _ := strconv.ParseFloat(fields[11], 64)
		system, _ := strconv.ParseFloat(fields[12], 64)
		return (user + system) / 100
	}
	output, err := combinedChildOutput(exec.Command("ps", "-o", "time=", "-p", strconv.Itoa(pid)))
	if err != nil {
		return 0
	}
	// macOS prints minutes:seconds.hundredths, as 0:00.52.
	minutes, seconds, _ := strings.Cut(strings.TrimSpace(string(output)), ":")
	whole, _ := strconv.ParseFloat(minutes, 64)
	part, _ := strconv.ParseFloat(seconds, 64)
	return whole*60 + part
}

// Output is read by os/exec's sole guarded reader; seeing the first newline
// releases the driver without mixing buffered readers with Wait.
type promptOutput struct {
	text  lockedBuffer
	ready chan struct{}
	once  sync.Once
}

func (p *promptOutput) Write(data []byte) (int, error) {
	n, err := p.text.Write(data)
	if bytes.Contains(p.text.Bytes(), []byte("\n")) {
		p.once.Do(func() { close(p.ready) })
	}
	return n, err
}
