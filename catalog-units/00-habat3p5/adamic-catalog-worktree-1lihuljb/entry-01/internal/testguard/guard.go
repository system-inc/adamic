// Package testguard runs test commands with a child CPU hang guard.
// It uses POSIX CPU limits and process groups, like the native test tooling.
package testguard

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const marker = "ADAMIC_TEST_CPU_CHILD"
const Budget = 10 * time.Minute
const Ceiling = 60 * time.Minute

// Re-exec the importing test binary, then replace it with the original child.
// RLIMIT_CPU counts this process's own CPU, never time waiting for a core.
// The hard limit prevents a child handling SIGXCPU from evading the guard.
func init() {
	if os.Getenv(marker) != "1" {
		return
	}
	if len(os.Args) < 4 || os.Args[1] != "--adamic-cpu-child" {
		os.Exit(125)
	}
	seconds, err := strconv.ParseUint(os.Args[2], 10, 64)
	if err != nil {
		os.Exit(125)
	}
	if err = syscall.Setrlimit(syscall.RLIMIT_CPU, &syscall.Rlimit{Cur: seconds, Max: seconds + 1}); err != nil {
		fmt.Fprintln(os.Stderr, "child CPU guard setup:", err)
		os.Exit(125)
	}
	os.Unsetenv(marker)
	argv := os.Args[3:]
	if err = syscall.Exec(argv[0], argv, os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, "child CPU guard exec:", err)
		os.Exit(125)
	}
}

// Run preserves the command's streams, directory and environment. Ordinary exit
// errors retain their type. Only guard failures are wrapped with a named reason.
func Run(command *exec.Cmd, budget, ceiling time.Duration) error {
	if command.Err != nil {
		return command.Err
	}
	label := strings.Join(command.Args, " ")
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	argv := append([]string{command.Path}, command.Args[1:]...)
	seconds := uint64((budget + time.Second - 1) / time.Second)
	command.Path = executable
	command.Args = append([]string{executable, "--adamic-cpu-child", strconv.FormatUint(seconds, 10)}, argv...)
	if command.Env == nil {
		command.Env = command.Environ()
	}
	command.Env = append(command.Env, marker+"=1")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err = command.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	timer := time.NewTimer(ceiling)
	defer timer.Stop()
	select {
	case err = <-done:
		// Clean up descendants even if the group leader exited first.
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
	case <-timer.C:
		_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		<-done
		return fmt.Errorf("%s: wall backstop exceeded: ceiling %s", label, ceiling)
	}
	if err != nil && command.ProcessState != nil {
		cpu := command.ProcessState.UserTime() + command.ProcessState.SystemTime()
		status := command.ProcessState.Sys().(syscall.WaitStatus)
		if status.Signaled() && (status.Signal() == syscall.SIGXCPU || (status.Signal() == syscall.SIGKILL && cpu >= budget)) {
			return fmt.Errorf("%s: child CPU hang guard exceeded: CPU %s, budget %s: %w", label, cpu, budget, err)
		}
	}
	return err
}
