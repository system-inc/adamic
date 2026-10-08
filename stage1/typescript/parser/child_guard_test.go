package parser

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Local estree-style adapter until internal/childguard lands. CPU time counts
// progress while quiet children wait for a loaded machine. Wall time is only a
// long backstop for sleeping or blocked children, never a performance target.
const parserChildBackstop = 60 * time.Minute

func parserCPUChild() {
	if len(os.Args) < 2 || os.Args[1] != "--adamic-parser-cpu-child" {
		return
	}
	if len(os.Args) < 5 {
		fmt.Fprintln(os.Stderr, "invalid CPU child request")
		os.Exit(1)
	}
	seconds, err := strconv.ParseUint(os.Args[2], 10, 64)
	if err == nil {
		err = syscall.Setrlimit(syscall.RLIMIT_CPU, &syscall.Rlimit{Cur: seconds, Max: seconds + 1})
	}
	if err == nil {
		err = syscall.Exec(os.Args[3], os.Args[4:], os.Environ())
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// Mutate an unstarted command so callers retain its directory, environment and
// output files. Exec preserves the PID and CPU limit across the actual program.
func parserRunGuard(command *exec.Cmd, budget time.Duration) error {
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	seconds := uint64((budget + time.Second - 1) / time.Second)
	if seconds == 0 {
		return fmt.Errorf("invalid CPU budget %s", budget)
	}
	original := append([]string(nil), command.Args...)
	target := command.Path
	if !strings.ContainsRune(target, '/') {
		target, err = exec.LookPath(target)
		if err != nil {
			return err
		}
	}
	command.Path = executable
	command.Args = append([]string{executable, "--adamic-parser-cpu-child", strconv.FormatUint(seconds, 10), target}, original...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := command.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	timer := time.NewTimer(parserChildBackstop)
	defer timer.Stop()
	select {
	case err = <-done:
		// A compound build may leave descendants when its own CPU guard fires.
		syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		cpu := command.ProcessState.UserTime() + command.ProcessState.SystemTime()
		status, _ := command.ProcessState.Sys().(syscall.WaitStatus)
		if status.Signal() == syscall.SIGXCPU || (err != nil && cpu >= time.Duration(seconds)*time.Second) {
			return fmt.Errorf("%s stalled: CPU budget %s exhausted (%s CPU): %w", original, time.Duration(seconds)*time.Second, cpu, err)
		}
		return err
	case <-timer.C:
		syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		<-done
		return fmt.Errorf("%s stalled: %s wall backstop exhausted", original, parserChildBackstop)
	}
}

func parserGuardOutput(command *exec.Cmd, budget time.Duration) ([]byte, error) {
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	err := parserRunGuard(command, budget)
	return output.Bytes(), err
}
