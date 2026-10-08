// check runs independent source programs against Node, sanitizers and counted builds.
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var evidence string

func execute(env []string, name string, args ...string) leakcheck.Run {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), env...)
	var out, err bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &err
	failure := cmd.Run()
	code := 0
	if failure != nil {
		code = -1
		if cmd.ProcessState != nil {
			code = cmd.ProcessState.ExitCode()
		}
	}
	return leakcheck.Run{Stdout: out.Bytes(), Stderr: err.Bytes(), ExitCode: code}
}
func save(name string, run leakcheck.Run) {
	text := fmt.Sprintf("exit %d\nstdout:\n%sstderr:\n%s", run.ExitCode, run.Stdout, run.Stderr)
	if err := os.WriteFile(filepath.Join(evidence, name+".log"), []byte(text), 0644); err != nil {
		panic(err)
	}
}
func agrees(want, got leakcheck.Run, count bool) bool {
	if count {
		got.Stderr = leakcheck.CountsLine.ReplaceAll(got.Stderr, nil)
		got.Stderr = regexp.MustCompile(`(?m)^adamic: graph region: live \d+ bytes \d+ reachable \d+ bytes \d+ unreachable \d+ bytes \d+ metadata \d+\n`).ReplaceAll(got.Stderr, nil)
	}
	return want.ExitCode == got.ExitCode && bytes.Equal(want.Stdout, got.Stdout) && bytes.Equal(want.Stderr, got.Stderr)
}
func check(path string) (passed bool) {
	name := strings.TrimSuffix(filepath.Base(path), ".a")
	defer func() {
		if err := recover(); err != nil {
			passed = false
			fmt.Printf("FAIL %s panic: %v\n", name, err)
		}
	}()
	absolute, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	want := execute(nil, "node", "--disable-warning=ExperimentalWarning", "oracle/node.mjs", absolute)
	save(name+"-node", want)
	if want.ExitCode != 0 || len(want.Stderr) != 0 {
		fmt.Printf("FAIL %s source Node: exit=%d stderr=%s\n", name, want.ExitCode, want.Stderr)
		return
	}
	program, err := load.Load([]string{absolute})
	if err != nil {
		fmt.Printf("FAIL %s load: %v\n", name, err)
		return
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		fmt.Printf("REFUSED %s: %v\n", name, err)
		return
	}
	code := native.C(lowered)
	if err := os.WriteFile(filepath.Join(evidence, name+".c"), []byte(code), 0644); err != nil {
		panic(err)
	}
	lanes := []struct {
		name    string
		options native.Options
	}{{"sanitize", native.Options{Sanitize: true}}, {"count", native.Options{Count: true}}}
	if strings.HasPrefix(name, "parallel_") {
		lanes = append(lanes, struct {
			name    string
			options native.Options
		}{"tsan", native.Options{ThreadSanitize: true}})
	}
	passed = true
	for _, lane := range lanes {
		binary := filepath.Join(evidence, name+"-"+lane.name)
		if err := native.Build(code, binary, lane.options); err != nil {
			fmt.Printf("FAIL %s/%s build: %v\n", name, lane.name, err)
			passed = false
			continue
		}
		threads := []string{"4"}
		if strings.HasPrefix(name, "parallel_") {
			threads = []string{"4", "16"}
		}
		for _, thread := range threads {
			env := []string{"ADAMIC_THREADS=" + thread, "ASAN_OPTIONS=detect_leaks=0", "TSAN_OPTIONS=halt_on_error=1:history_size=4:report_atomic_races=1", "ADAMIC_TSAN_PERTURB=1"}
			runs := 1
			if lane.name == "tsan" {
				runs = 3
			}
			for attempt := 0; attempt < runs; attempt++ {
				got := execute(env, binary)
				label := fmt.Sprintf("%s-%s-%s-%d", name, lane.name, thread, attempt+1)
				save(label, got)
				match := agrees(want, got, lane.options.Count)
				balance := ""
				if lane.options.Count {
					balance = leakcheck.Unbalanced(got)
				}
				if !match || balance != "" {
					fmt.Printf("FAIL %s match=%t counts=%s exit=%d stderr=%s\n", label, match, balance, got.ExitCode, got.Stderr)
					passed = false
				} else {
					fmt.Printf("PASS %s %s", label, got.Stderr)
					if len(got.Stderr) == 0 {
						fmt.Println()
					}
				}
			}
			if lane.options.Sanitize && want.ExitCode == 0 {
				report, err := leakcheck.Check(leakcheck.Program{C: code, Sanitized: binary, Counted: filepath.Join(evidence, name+"-leaks-count"), Execute: func(extra []string, name string, args ...string) leakcheck.Run {
					r := execute(append(env, extra...), name, args...)
					save(fmt.Sprintf("%s-leaks-%s", filepath.Base(binary), thread), r)
					return r
				}})
				if err != nil || report != "" {
					fmt.Printf("FAIL %s/leaks/%s: %v %s\n", name, thread, err, report)
					passed = false
				} else {
					fmt.Printf("PASS %s/leaks/%s\n", name, thread)
				}
			}
		}
	}
	return
}
func main() {
	if len(os.Args) < 3 {
		panic("usage: check <evidence directory> <program.a>...")
	}
	evidence, _ = filepath.Abs(os.Args[1])
	if err := os.MkdirAll(evidence, 0755); err != nil {
		panic(err)
	}
	if os.Args[2] == "--mutants" {
		prove()
		return
	}
	if os.Args[2] == "--refusal" {
		proveRefusal()
		return
	}
	failures := 0
	for _, path := range os.Args[2:] {
		if !check(path) {
			failures++
		}
	}
	fmt.Printf("SUMMARY programs=%d failed-or-refused=%d\n", len(os.Args)-2, failures)
	if failures > 0 {
		os.Exit(1)
	}
}
