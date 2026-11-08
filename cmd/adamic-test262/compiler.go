package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Compiler workers amortize startup without giving up the original process isolation: a hung
// checker can still be killed at the two-minute deadline. Each request owns a fresh load.Program.
func init() {
	if len(os.Args) != 2 || os.Args[1] != "--compiler-worker" {
		return
	}
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	for {
		var path string
		if decodeError := decoder.Decode(&path); decodeError != nil {
			os.Exit(0)
		}
		if encodeError := encoder.Encode(recordExecution(compileInProcess(path))); encodeError != nil {
			os.Exit(1)
		}
	}
}

func compileInProcess(path string) (result execution) {
	defer func() {
		if value := recover(); value != nil {
			result = execution{Exit: 2, Stderr: fmt.Sprintf("panic: %v\n", value)}
		}
		if len(result.Stdout) > 16<<20 || len(result.Stderr) > 16<<20 {
			result = execution{Exit: -1, Stderr: "command output exceeded capture limit\n"}
		}
	}()
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		var checkError *load.CheckError
		var diagnostic strings.Builder
		if errors.As(loadError, &checkError) {
			for _, text := range checkError.Diagnostics {
				fmt.Fprintln(&diagnostic, text)
			}
		} else {
			fmt.Fprintf(&diagnostic, "adamic: %v\n", loadError)
		}
		return execution{Exit: 1, Stderr: diagnostic.String()}
	}
	lowered, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		return execution{Exit: 1, Stderr: "adamic: " + lowerError.Error() + "\n"}
	}
	if native.UsesTsGo(lowered) {
		return execution{Exit: 1, Stderr: "adamic: tsgo requires a native build with --tsgo <archive>\n"}
	}
	code := native.C(lowered)
	if len(code) > 16<<20 {
		return execution{Exit: -1, Stderr: "command output exceeded capture limit\n"}
	}
	return execution{Stdout: code}
}

type compilerWorker struct {
	command *exec.Cmd
	input   io.WriteCloser
	output  io.ReadCloser
	encoder *json.Encoder
	decoder *json.Decoder
	timeout time.Duration
}

func (worker *compilerWorker) close() {
	if worker.command == nil {
		return
	}
	_ = worker.input.Close()
	_ = worker.output.Close()
	_ = worker.command.Process.Kill()
	_ = worker.command.Wait()
	worker.command = nil
}

func (worker *compilerWorker) compile(path string) execution {
	if worker.command == nil {
		executable, executableError := os.Executable()
		if executableError != nil {
			return execution{Exit: -1, Stderr: executableError.Error()}
		}
		command := exec.Command(executable, "--compiler-worker")
		command.Env = childEnvironment()
		input, stdinPipeError := command.StdinPipe()
		if stdinPipeError != nil {
			return execution{Exit: -1, Stderr: stdinPipeError.Error()}
		}
		output, stdoutPipeError := command.StdoutPipe()
		if stdoutPipeError != nil {
			input.Close()
			return execution{Exit: -1, Stderr: stdoutPipeError.Error()}
		}
		if startError := command.Start(); startError != nil {
			input.Close()
			output.Close()
			return execution{Exit: -1, Stderr: startError.Error()}
		}
		worker.command, worker.input, worker.output = command, input, output
		worker.encoder, worker.decoder = json.NewEncoder(input), json.NewDecoder(output)
	}
	done := make(chan execution, 1)
	go func() {
		if encodeError := worker.encoder.Encode(path); encodeError != nil {
			done <- execution{Exit: -1, Stderr: encodeError.Error()}
			return
		}
		var result recordedExecution
		if decodeError := worker.decoder.Decode(&result); decodeError != nil {
			done <- execution{Exit: -1, Stderr: decodeError.Error()}
			return
		}
		done <- result.execution()
	}()
	timeout := worker.timeout
	if timeout == 0 {
		timeout = 2 * time.Minute
	}
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case result := <-done:
		if result.Exit == -1 {
			worker.close()
		}
		return result
	case <-timer.C:
		worker.close()
		<-done
		return execution{Exit: -1, TimedOut: true}
	}
}
