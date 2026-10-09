package estree

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// The helper execs the original command in the same process. RLIMIT_CPU counts
// that child's user and system CPU, including startup CPU but excluding time
// waiting for a core. A separate pipe leaves stdout/stderr assertions unchanged.
// All callers have whole-second budgets. The hard limit is one second beyond
// the soft limit, so even a child handling SIGXCPU cannot evade it.
func TestDeadlineChild(t *testing.T) {
	if os.Getenv("ADAMIC_ESTREE_DEADLINE_CHILD") != "1" {
		return
	}
	separator := 0
	for index, argument := range os.Args {
		if argument == "--" {
			separator = index
			break
		}
	}
	if separator == 0 || len(os.Args) < separator+3 {
		os.Exit(125)
	}
	seconds, err := strconv.ParseUint(os.Args[separator+1], 10, 64)
	if err != nil {
		os.Exit(125)
	}
	if err := syscall.Setrlimit(syscall.RLIMIT_CPU, &syscall.Rlimit{Cur: seconds, Max: seconds + 1}); err != nil {
		fmt.Fprintln(os.Stderr, "child CPU limit:", err)
		os.Exit(125)
	}
	argv := os.Args[separator+2:]
	path, err := exec.LookPath(argv[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(125)
	}
	ready := os.NewFile(3, "ready")
	if _, err := ready.WriteString("started\n"); err != nil {
		os.Exit(125)
	}
	ready.Close()
	os.Unsetenv("ADAMIC_ESTREE_DEADLINE_CHILD")
	if err := syscall.Exec(path, argv, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(125)
	}
}

func runWithCPUBudget(t *testing.T, argv []string, stdout, stderr io.Writer, budget time.Duration) (error, bool) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	seconds := uint64((budget + time.Second - 1) / time.Second)
	args := append([]string{"-test.run=^TestDeadlineChild$", "--", strconv.FormatUint(seconds, 10)}, argv...)
	command := exec.Command(executable, args...)
	command.Env = append(os.Environ(), "ADAMIC_ESTREE_DEADLINE_CHILD=1")
	command.ExtraFiles = []*os.File{writer}
	command.Stdout, command.Stderr = stdout, stderr
	startup := time.NewTimer(5 * time.Minute)
	defer startup.Stop()
	if err := command.Start(); err != nil {
		writer.Close()
		t.Fatalf("%v: child never started: %v", argv, err)
	}
	writer.Close()
	t.Cleanup(func() { _ = command.Process.Kill() })
	ready := make(chan string, 1)
	go func() { line, _ := bufio.NewReader(reader).ReadString('\n'); ready <- line }()
	select {
	case line := <-ready:
		if line != "started\n" {
			_ = command.Process.Kill()
			_ = command.Wait()
			t.Fatalf("%v: child never started: ready line %q", argv, line)
		}
	case <-startup.C:
		_ = command.Process.Kill()
		_ = command.Wait()
		t.Fatalf("%v: child never started within 5m", argv)
	}
	err = command.Wait()
	cpu := command.ProcessState.UserTime() + command.ProcessState.SystemTime()
	status := command.ProcessState.Sys().(syscall.WaitStatus)
	exceeded := cpu >= budget || (status.Signaled() && status.Signal() == syscall.SIGXCPU)
	if exceeded {
		t.Logf("%v: child CPU deadline exceeded: used %s, budget %s", argv, cpu, budget)
	}
	return err, exceeded
}
