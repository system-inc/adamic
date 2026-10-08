package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Compiler workers amortize startup without giving up the original process isolation: a hung
// checker can still be killed after two minutes of CPU time (wall time on Darwin).
// Each request owns a fresh load.Program.
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
	command     *exec.Cmd
	input       io.WriteCloser
	output      io.ReadCloser
	encoder     *json.Encoder
	decoder     *json.Decoder
	cpuLimit    time.Duration
	wallTimeout time.Duration
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
		executable, err := os.Executable()
		if err != nil {
			return execution{Exit: -1, Stderr: err.Error()}
		}
		command := exec.Command(executable, "--compiler-worker")
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
		worker.command, worker.input, worker.output = command, input, output
		worker.encoder, worker.decoder = json.NewEncoder(input), json.NewDecoder(output)
	}
	budget := worker.cpuLimit
	if budget == 0 {
		budget = 2 * time.Minute
	}
	wall := worker.wallTimeout
	if wall == 0 {
		wall = 10 * time.Minute
	}
	// Darwin has no /proc accounting; retain the original wall deadline there.
	if runtime.GOOS != "linux" && worker.wallTimeout == 0 {
		wall = budget
	}
	pid := worker.command.Process.Pid
	startCPU, err := workerCPUTime(pid)
	if err != nil {
		worker.close()
		return execution{Exit: -1, Stderr: err.Error()}
	}
	encoder, decoder := worker.encoder, worker.decoder
	done := make(chan execution, 1)
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
	timer := time.NewTimer(wall)
	defer timer.Stop()
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	stop := func() execution {
		worker.close()
		<-done
		return execution{Exit: -1, TimedOut: true}
	}
	for {
		select {
		case result := <-done:
			cpu, err := workerCPUTime(pid)
			if err == nil && cpu-startCPU > budget {
				result = execution{Exit: -1, TimedOut: true}
			} else if err != nil && result.Exit != -1 {
				result = execution{Exit: -1, Stderr: err.Error()}
			}
			if result.Exit == -1 {
				worker.close()
			}
			return result
		case <-ticker.C:
			cpu, err := workerCPUTime(pid)
			if err != nil {
				worker.close()
				<-done
				return execution{Exit: -1, Stderr: err.Error()}
			}
			if cpu-startCPU > budget {
				return stop()
			}
		case <-timer.C:
			return stop()
		}
	}
}

var cpuClockTicks = sync.OnceValues(func() (int64, error) {
	output, err := exec.Command("getconf", "CLK_TCK").Output()
	if err != nil {
		return 0, err
	}
	ticks, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err == nil && ticks <= 0 {
		err = fmt.Errorf("invalid CLK_TCK: %d", ticks)
	}
	return ticks, err
})

// /proc stat fields 14 and 15 account for all threads in the worker, excluding
// children. The command name in parentheses may itself contain spaces or ')'.
func workerCPUTime(pid int) (time.Duration, error) {
	if runtime.GOOS != "linux" {
		return 0, nil
	}
	ticks, err := cpuClockTicks()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0, err
	}
	end := strings.LastIndexByte(string(data), ')')
	if end < 0 {
		return 0, fmt.Errorf("invalid worker stat")
	}
	fields := strings.Fields(string(data[end+1:]))
	if len(fields) < 13 {
		return 0, fmt.Errorf("short worker stat")
	}
	user, err := strconv.ParseInt(fields[11], 10, 64)
	if err != nil {
		return 0, err
	}
	system, err := strconv.ParseInt(fields[12], 10, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(user+system) * time.Second / time.Duration(ticks), nil
}
