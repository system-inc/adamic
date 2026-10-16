package estree

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

// The helper execs the original command in the same process. RLIMIT_CPU counts
// that child's user and system CPU, including startup CPU but excluding time
// waiting for a core. A separate pipe leaves stdout/stderr assertions unchanged.
// All callers have whole-second budgets. The hard limit is one second beyond
// the soft limit, so even a child handling SIGXCPU cannot evade it.
// Not parallel: process-wide environment, RLIMIT_CPU and syscall.Exec
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
	path, err := toolchain.Path(argv[0])
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

func runWithCPUBudget(t *testing.T, argv []string, stdout, stderr io.Writer, budget time.Duration) (error, bool, time.Duration) {
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
	seconds := estreeBudgetSeconds(budget)
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
	return err, exceeded, cpu
}

// estreeParserBudget is the CPU a refusal's parsing may take beyond its program's own fixed cost.
const estreeParserBudget = 2 * time.Second

// estreeCalibrationCeiling refuses a calibration that is itself absurd: a program that needs more than this to accept a
// one-line input is broken or starved, and granting it a budget that large would bound nothing.
const estreeCalibrationCeiling = 10 * time.Second

// estreeCalibrationSource is the accepted input a program's fixed cost is measured on. Every program here loads its whole
// port at startup, so a one-line input loads everything a refusal loads.
const estreeCalibrationSource = "const calibration = 1;\n"

type estreeRefusalCalibration struct {
	once          sync.Once
	fixed, budget time.Duration
	err           error
}

var estreeRefusalCalibrations sync.Map

// estreeRefusalBudget is the CPU a refusal of argv (a program, then its input path) may use: the program's own fixed
// cost plus estreeParserBudget. The fixed cost is measured once per process, by the same command on an accepted input:
// Node's startup and its stripping and loading of the TypeScript port took nearly all of a flat 2 s budget, CPU summed
// over its JIT and GC threads, and on a loaded box went past it before the parser did any work (#f34y1bv: Adam, Oct 14,
// TestBoundedPortParser_034 and _035). Load inflates the fixed cost and the refusal alike, so what the check bounds is
// the parser's own work, which is what it is for.
func estreeRefusalBudget(t *testing.T, argv []string) time.Duration {
	t.Helper()
	program := argv[:len(argv)-1]
	value, _ := estreeRefusalCalibrations.LoadOrStore(strings.Join(program, "\x00"), &estreeRefusalCalibration{})
	calibration := value.(*estreeRefusalCalibration)
	calibration.once.Do(func() {
		directory, err := os.MkdirTemp("", "estree-calibration-")
		if err != nil {
			calibration.err = err
			return
		}
		defer os.RemoveAll(directory)
		input := filepath.Join(directory, "accepted.ts")
		if err := os.WriteFile(input, []byte(estreeCalibrationSource), 0644); err != nil {
			calibration.err = err
			return
		}
		var stdout, stderr bytes.Buffer
		err, exceeded, cpu := runWithCPUBudget(t, append(append([]string(nil), program...), input), &stdout, &stderr, estreeCalibrationCeiling)
		switch {
		case exceeded:
			calibration.err = fmt.Errorf("%v: calibration used %s of CPU on a one-line input, over the %s ceiling", program, cpu, estreeCalibrationCeiling)
		case err != nil || stderr.Len() != 0 || stdout.Len() == 0:
			calibration.err = fmt.Errorf("%v: calibration input was not accepted: %v stdout=%d stderr=%s", program, err, stdout.Len(), &stderr)
		default:
			calibration.fixed, calibration.budget = cpu, cpu+estreeParserBudget
		}
	})
	if calibration.err != nil {
		t.Fatal(calibration.err)
	}
	t.Logf("%v: fixed cost %s, refusal budget %s", program, calibration.fixed, calibration.budget)
	return calibration.budget
}

// estreeBudgetSeconds is a budget as RLIMIT_CPU takes it, whole seconds rounded up; the check itself compares exact CPU.
func estreeBudgetSeconds(budget time.Duration) uint64 {
	return uint64((budget + time.Second - 1) / time.Second)
}
