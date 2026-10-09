//go:build !windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

// returnEarly runs the engine as a child and exits with its exit code as soon as the engine writes it to
// the verdict descriptor, leaving the child to finish alone. See execute.
func returnEarly(binaryPath string, arguments []string) error {
	verdictReader, verdictWriter, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("opening the verdict pipe: %w", err)
	}
	command := exec.Command(binaryPath, arguments...)
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	command.ExtraFiles = []*os.File{verdictWriter}
	command.Env = append(os.Environ(), "COHERE_VERDICT_FD=3")

	// A signal sent to this process alone reaches the engine too, as it would have with exec. One from the
	// terminal reaches both already, and a second copy changes nothing.
	signals := make(chan os.Signal, 4)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)
	if err := command.Start(); err != nil {
		return fmt.Errorf("running %s: %w", binaryPath, err)
	}
	verdictWriter.Close()
	go func() {
		for received := range signals {
			command.Process.Signal(received)
		}
	}()

	verdict := make([]byte, 1)
	if read, _ := verdictReader.Read(verdict); read == 1 {
		os.Exit(int(verdict[0]))
	}
	// No verdict: the engine ended, or closed the descriptor, without answering.
	command.Wait()
	signal.Stop(signals)
	if status, isWait := command.ProcessState.Sys().(syscall.WaitStatus); isWait && status.Signaled() {
		// Killed by a signal, so this process dies of the same one, and the caller's shell reports it as it
		// would have reported the engine's death under exec.
		signal.Reset(status.Signal())
		syscall.Kill(os.Getpid(), status.Signal())
	}
	os.Exit(command.ProcessState.ExitCode())
	return nil
}
