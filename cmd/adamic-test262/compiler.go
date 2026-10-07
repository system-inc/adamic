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

	"github.com/system-inc/adamic/internal/boundedrun"
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
		if err := decoder.Decode(&path); err != nil {
			os.Exit(0)
		}
		if err := encoder.Encode(recordExecution(compileInProcess(path))); err != nil {
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
	loaded, err := load.Load([]string{path})
	if err != nil {
		var checkError *load.CheckError
		var diagnostic strings.Builder
		if errors.As(err, &checkError) {
			for _, text := range checkError.Diagnostics {
				fmt.Fprintln(&diagnostic, text)
			}
		} else {
			fmt.Fprintf(&diagnostic, "adamic: %v\n", err)
		}
		return execution{Exit: 1, Stderr: diagnostic.String()}
	}
	lowered, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		return execution{Exit: 1, Stderr: "adamic: " + err.Error() + "\n"}
	}
	if native.UsesTSGo(lowered) {
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
	release func()
}

func (worker *compilerWorker) close() {
	if worker.command == nil {
		return
	}
	_ = worker.input.Close()
	_ = worker.output.Close()
	_ = boundedrun.Kill(worker.command)
	done := make(chan struct{})
	command := worker.command
	go func() { _ = command.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
	}
	if worker.release != nil {
		worker.release()
		worker.release = nil
	}
	worker.command = nil
}

func (worker *compilerWorker) compile(path string) execution {
	if worker.command == nil {
		executable, err := os.Executable()
		if err != nil {
			return execution{Exit: -1, Stderr: err.Error()}
		}
		// A worker can serve many tests; each request retains its two-minute limit.
		// Seventy minutes also bounds the worker while idle or between requests.
		command, release := boundedrun.Command(boundedrun.Shard, executable, "--compiler-worker")
		worker.release = release
		initialized := false
		defer func() {
			if !initialized {
				release()
				worker.release = nil
			}
		}()
		input, err := command.StdinPipe()
		if err != nil {
			return execution{Exit: -1, Stderr: err.Error()}
		}
		output, err := command.StdoutPipe()
		if err != nil {
			input.Close()
			return execution{Exit: -1, Stderr: err.Error()}
		}
		if err := command.Start(); err != nil {
			input.Close()
			output.Close()
			return execution{Exit: -1, Stderr: err.Error()}
		}
		initialized = true
		worker.command, worker.input, worker.output = command.Cmd, input, output
		worker.encoder, worker.decoder = json.NewEncoder(input), json.NewDecoder(output)
	}
	done := make(chan execution, 1)
	encoder, decoder := worker.encoder, worker.decoder
	go func() {
		if err := encoder.Encode(path); err != nil {
			done <- execution{Exit: -1, Stderr: err.Error()}
			return
		}
		var result recordedExecution
		if err := decoder.Decode(&result); err != nil {
			done <- execution{Exit: -1, Stderr: err.Error()}
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
		name := worker.command.Path
		worker.close()
		return execution{Exit: -1, TimedOut: true, Stderr: fmt.Sprintf("child compiler-worker (%s): deadline exceeded; process group killed\n", name)}
	}
}
