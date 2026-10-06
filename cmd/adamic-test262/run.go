package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// engine is one checkout's adamic command and its runtime, compiled once. Each test is then
// `adamic c` plus a link, the way the fuzzer avoids compiling the runtime per program.
type engine struct {
	test262 string
	work    string
	adamic  string
	runtime []string
	include string
	flags   []string
	log     io.Writer
	adapt   bool
	oracle  *typescriptOracle
}

func prepare(root string, test262 string, work string) (*engine, error) {
	if err := os.MkdirAll(work, 0o755); err != nil {
		return nil, err
	}
	adamic := filepath.Join(work, "adamic")
	build := exec.Command("go", "build", "-o", adamic, "./cmd/adamic")
	build.Dir = root
	if output, err := build.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("building adamic: %w\n%s", err, output)
	}
	include := filepath.Join(root, "internal", "native", "runtime")
	// Sanitizers, as the oracle compiles: a native memory bug is a crash, not a pass. Leaks are not
	// compared to Node (the process exits either way), so leak detection stays off at run time.
	flags := native.Flags(native.Options{Sanitize: true})
	sources, err := filepath.Glob(filepath.Join(include, "*.c"))
	if err != nil {
		return nil, err
	}
	sort.Strings(sources)
	runtimeDir := filepath.Join(work, "runtime")
	if err := os.MkdirAll(runtimeDir, 0o755); err != nil {
		return nil, err
	}
	var objects []string
	for _, source := range sources {
		object := filepath.Join(runtimeDir, strings.TrimSuffix(filepath.Base(source), ".c")+".o")
		arguments := append(append([]string{}, flags...), "-c", source, "-o", object)
		if output, err := exec.Command("clang", arguments...).CombinedOutput(); err != nil {
			return nil, fmt.Errorf("compiling %s: %w\n%s", filepath.Base(source), err, output)
		}
		objects = append(objects, object)
	}
	oracle, err := startTypescript(root, work)
	if err != nil {
		return nil, err
	}
	return &engine{
		oracle:  oracle,
		test262: test262,
		work:    work,
		adamic:  adamic,
		runtime: objects,
		include: include,
		flags:   flags,
		log:     os.Stderr,
	}, nil
}

// result is one test after it has been classified and, if attempted, run.
type result struct {
	Path        string         `json:"path"`
	Directory   string         `json:"directory"`
	Kind        outcomeKind    `json:"kind"`
	Reason      string         `json:"reason,omitempty"`
	Adaptations map[string]int `json:"adaptations,omitempty"`
}

func (e *engine) runFilter(filter string, limit int, classifyOnly bool) (filterReport, error) {
	files, err := listTests(e.test262, filter)
	if err != nil {
		return filterReport{}, err
	}
	report := filterReport{Path: filter}
	attempted := 0
	for index, file := range files {
		source, err := os.ReadFile(file)
		if err != nil {
			return report, err
		}
		relative, err := filepath.Rel(filepath.Join(e.test262, "test"), file)
		if err != nil {
			relative = file
		}
		relative = filepath.ToSlash(relative)
		classified := classify(relative, string(source), e.adapt)
		var one result
		if classified.Skip != "" {
			one = result{Path: relative, Directory: classified.Directory, Kind: outcomeSkipped, Reason: classified.Skip}
		} else if classifyOnly || (limit > 0 && attempted >= limit) {
			one = result{Path: relative, Directory: classified.Directory, Kind: outcomeUnrun, Reason: "not run"}
		} else {
			attempted++
			one = e.attempt(classified)
			if one.Kind == outcomeCrashed {
				one.Reason = withCrashPath(one.Path, one.Reason)
			}
		}
		report.add(one)
		if (index+1)%50 == 0 || index+1 == len(files) {
			fmt.Fprintf(e.log, "%s %d/%d pass=%d fail=%d refused=%d not-typescript=%d crashed=%d skipped=%d\n",
				filter, index+1, len(files), report.Pass, report.Fail, report.Refused, report.NotTypescript, report.Crashed, report.Skipped)
		}
	}
	report.finish()
	return report, nil
}

func (e *engine) attempt(test classified) result {
	base := result{Path: test.Path, Directory: test.Directory, Adaptations: test.Adaptations}
	typescript := filepath.Join(e.work, "program.ts")
	module := filepath.Join(e.work, "program.mts")
	if err := os.WriteFile(typescript, []byte(test.Program), 0o644); err != nil {
		return result{Path: test.Path, Directory: test.Directory, Kind: outcomeCrashed, Reason: err.Error()}
	}
	if err := os.WriteFile(module, []byte(test.Program), 0o644); err != nil {
		return result{Path: test.Path, Directory: test.Directory, Kind: outcomeCrashed, Reason: err.Error()}
	}
	lowered := runCommandWithLimit(2*time.Minute, nil, 16<<20, e.adamic, "c", typescript)
	kind, reason := compileClass(lowered.Stderr, lowered.Exit, lowered.TimedOut)
	if kind == "refused" {
		base.Kind = outcomeRefused
		base.Reason = reason
		if checkerCode(reason) != "" {
			codes, err := e.oracle.check(test.Program)
			if err != nil {
				base.Kind = outcomeCrashed
				base.Reason = "TypeScript oracle: " + err.Error()
			} else {
				decision := typescriptVerdict(reason, codes)
				base.Kind = decision.Kind
				base.Reason = decision.Reason
			}
		}
		return base
	}
	if kind == "crashed" || lowered.Exit != 0 {
		base.Kind = outcomeCrashed
		base.Reason = reason
		return base
	}
	cPath := filepath.Join(e.work, "program.c")
	if err := os.WriteFile(cPath, []byte(lowered.Stdout), 0o644); err != nil {
		base.Kind = outcomeCrashed
		base.Reason = err.Error()
		return base
	}
	binary := filepath.Join(e.work, "program.bin")
	arguments := append(append([]string{}, e.flags...), "-I", e.include, "-o", binary, cPath)
	arguments = append(arguments, e.runtime...)
	arguments = append(arguments, "-lm")
	linked := runCommand(2*time.Minute, nil, "clang", arguments...)
	if linked.Exit != 0 || linked.TimedOut {
		base.Kind = outcomeCrashed
		base.Reason = "clang: " + firstLine(linked.Stderr)
		return base
	}
	defer os.Remove(binary)
	nativeRun := runCommand(15*time.Second, []string{
		"ASAN_OPTIONS=detect_leaks=0:abort_on_error=1:halt_on_error=1",
		"UBSAN_OPTIONS=halt_on_error=1:abort_on_error=1",
	}, binary)
	nodeRun := runCommand(15*time.Second, nil, "node", "--disable-warning=ExperimentalWarning", module)
	decided := decide(verdictInput{
		NegativePhase: test.NegativePhase,
		NegativeType:  test.NegativeType,
		Node:          nodeRun,
		Native:        nativeRun,
	})
	base.Kind = decided.Kind
	base.Reason = decided.Reason
	return base
}

const outputLimit = 256 << 10

func runCommand(timeout time.Duration, extra []string, name string, args ...string) execution {
	return runCommandWithLimit(timeout, extra, outputLimit, name, args...)
}

// Generated C is an artifact, not program output: give it room, and never compile a truncated one.
func runCommandWithLimit(timeout time.Duration, extra []string, limit int, name string, args ...string) execution {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	command.Env = append(os.Environ(), extra...)
	command.WaitDelay = 2 * time.Second
	var stdout, stderr limitedBuffer
	stdout.limit = limit
	stderr.limit = limit
	stderr.tail = true
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	result := execution{Stdout: stdout.String(), Stderr: stderr.String()}
	if ctx.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.Exit = -1
		return result
	}
	if stdout.exceeded || stderr.exceeded {
		result.Exit = -1
		result.Stderr = "command output exceeded capture limit\n" + result.Stderr
		// Keep processing the exit status: a signal is still a crash.
		if err == nil {
			return result
		}
	}
	if err == nil {
		return result
	}
	exit, isExit := err.(*exec.ExitError)
	if !isExit {
		result.Exit = -1
		result.Stderr = result.Stderr + err.Error()
		return result
	}
	if status, ok := exit.Sys().(syscall.WaitStatus); ok {
		if status.Signaled() {
			result.Signal = status.Signal().String()
			result.Exit = -1
			return result
		}
		if stdout.exceeded || stderr.exceeded {
			return result
		}
		result.Exit = status.ExitStatus()
		return result
	}
	result.Exit = exit.ExitCode()
	return result
}

// limitedBuffer bounds captured output. Stdout keeps its start; stderr keeps its tail so
// a sanitizer report after excessive output survives until classification.
type limitedBuffer struct {
	buf      bytes.Buffer
	limit    int
	tail     bool
	exceeded bool
}

func (buffer *limitedBuffer) Write(data []byte) (int, error) {
	written := len(data)
	room := buffer.limit - buffer.buf.Len()
	if written > room {
		buffer.exceeded = true
	}
	if buffer.tail && written > room {
		if written >= buffer.limit {
			buffer.buf.Reset()
			data = data[written-buffer.limit:]
		} else {
			buffer.buf.Next(written - room)
		}
		_, _ = buffer.buf.Write(data)
		return written, nil
	}
	if room > 0 {
		if len(data) > room {
			data = data[:room]
		}
		_, _ = buffer.buf.Write(data)
	}
	return written, nil
}

func (buffer *limitedBuffer) String() string { return buffer.buf.String() }

func listTests(test262 string, filter string) ([]string, error) {
	root := filepath.Join(test262, "test")
	target := root
	if filter != "" {
		target = filepath.Join(root, filepath.FromSlash(filter))
	}
	info, err := os.Stat(target)
	if err != nil {
		return nil, fmt.Errorf("test262 filter %s: %w", filter, err)
	}
	if !info.IsDir() {
		return []string{target}, nil
	}
	var files []string
	err = filepath.WalkDir(target, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".js") || strings.Contains(name, "_FIXTURE") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	return files, nil
}

// filterReport is one directory filter's counts. Directories are the folders that hold tests.
type filterReport struct {
	Path                 string        `json:"path"`
	Pass                 int           `json:"pass"`
	Fail                 int           `json:"fail"`
	NotTypescript        int           `json:"notTypescript"`
	Refused              int           `json:"refused"`
	Crashed              int           `json:"crashed"`
	Skipped              int           `json:"skipped"`
	Unrun                int           `json:"unrun,omitempty"`
	Total                int           `json:"total"`
	Directories          []dirCount    `json:"directories"`
	NotTypescriptReasons []reasonCount `json:"notTypescriptReasons"`
	RefusalReasons       []reasonCount `json:"refusalReasons"`
	SkipReasons          []reasonCount `json:"skipReasons"`
	CrashReasons         []reasonCount `json:"crashReasons"`
	FailReasons          []reasonCount `json:"failReasons"`
	Passes               []string      `json:"passes,omitempty"`
	Adaptations          []reasonCount `json:"adaptations,omitempty"`
	directories          map[string]*dirCount
	notTypescriptReasons map[string]int
	refusalReasons       map[string]int
	skipReasons          map[string]int
	crashReasons         map[string]int
	failReasons          map[string]int
	adaptations          map[string]int
}

type dirCount struct {
	Path          string `json:"path"`
	Pass          int    `json:"pass"`
	Fail          int    `json:"fail"`
	NotTypescript int    `json:"notTypescript"`
	Refused       int    `json:"refused"`
	Crashed       int    `json:"crashed"`
	Skipped       int    `json:"skipped"`
	Unrun         int    `json:"unrun,omitempty"`
	Total         int    `json:"total"`
}

type reasonCount struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

func (report *filterReport) add(one result) {
	if report.directories == nil {
		report.directories = map[string]*dirCount{}
		report.notTypescriptReasons = map[string]int{}
		report.refusalReasons = map[string]int{}
		report.skipReasons = map[string]int{}
		report.crashReasons = map[string]int{}
		report.failReasons = map[string]int{}
		report.adaptations = map[string]int{}
	}
	directory := report.directories[one.Directory]
	if directory == nil {
		directory = &dirCount{Path: one.Directory}
		report.directories[one.Directory] = directory
	}
	directory.Total++
	report.Total++
	switch one.Kind {
	case outcomePass:
		report.Pass++
		directory.Pass++
		report.Passes = append(report.Passes, one.Path)
	case outcomeFail:
		report.Fail++
		directory.Fail++
		report.failReasons[one.Reason]++
	case outcomeNotTypescript:
		report.NotTypescript++
		directory.NotTypescript++
		report.notTypescriptReasons[reasonOr(one.Reason)]++
	case outcomeRefused:
		report.Refused++
		directory.Refused++
		report.refusalReasons[reasonOr(one.Reason)]++
	case outcomeCrashed:
		report.Crashed++
		directory.Crashed++
		report.crashReasons[reasonOr(one.Reason)]++
	case outcomeSkipped:
		report.Skipped++
		directory.Skipped++
		report.skipReasons[reasonOr(one.Reason)]++
	case outcomeUnrun:
		report.Unrun++
		directory.Unrun++
	}
	for kind, count := range one.Adaptations {
		report.adaptations[kind] += count
	}
}

func reasonOr(reason string) string {
	if reason == "" {
		return "(no reason)"
	}
	return reason
}

func (report *filterReport) finish() {
	for _, directory := range report.directories {
		report.Directories = append(report.Directories, *directory)
	}
	sort.Slice(report.Directories, func(i int, j int) bool { return report.Directories[i].Path < report.Directories[j].Path })
	report.NotTypescriptReasons = sortedReasons(report.notTypescriptReasons)
	report.RefusalReasons = sortedReasons(report.refusalReasons)
	report.SkipReasons = sortedReasons(report.skipReasons)
	report.CrashReasons = sortedReasons(report.crashReasons)
	report.FailReasons = sortedReasons(report.failReasons)
	report.Adaptations = sortedReasons(report.adaptations)
	sort.Strings(report.Passes)
}

func sortedReasons(counts map[string]int) []reasonCount {
	reasons := make([]reasonCount, 0, len(counts))
	for reason, count := range counts {
		reasons = append(reasons, reasonCount{Reason: reason, Count: count})
	}
	sort.Slice(reasons, func(i int, j int) bool {
		if reasons[i].Count != reasons[j].Count {
			return reasons[i].Count > reasons[j].Count
		}
		return reasons[i].Reason < reasons[j].Reason
	})
	return reasons
}

// reportDocument is what --json prints.
type reportDocument struct {
	Test262 string         `json:"test262"`
	Commit  string         `json:"commit,omitempty"`
	Adapt   bool           `json:"adapt"`
	Oracle  oracleStats    `json:"typescript"`
	Filters []filterReport `json:"filters"`
}

// withCrashPath puts the test file on the crash record. The reason alone was not enough to see
// which program the compiler or clang rejected.
func withCrashPath(path string, reason string) string {
	if path == "" {
		return reason
	}
	if reason == "" {
		return path
	}
	return path + ": " + reason
}

func test262Commit(test262 string) string {
	command := exec.Command("git", "-C", test262, "rev-parse", "HEAD")
	output, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func writeJSON(writer io.Writer, document reportDocument) error {
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return err
	}
	_, err = writer.Write(append(encoded, '\n'))
	return err
}
