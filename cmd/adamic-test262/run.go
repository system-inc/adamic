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
	"regexp"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/system-inc/adamic/internal/native"
)

// engine owns one checkout's compiler, keyed runtime archive and observation cache. Each worker
// owns a compiler process; tests are classified and their observations reduced in serial order.
type engine struct {
	test262          string
	work             string
	adamic           string
	runtime          []string
	include          string
	flags            []string
	log              io.Writer
	adapt            bool
	jobs             int
	runtimeKey       string
	cache            *resultCache
	nodeVersion      string
	context          string
	nodeContext      string
	compiler         *compilerWorker
	inProcess        bool
	compilerIdentity string
	profile          *runProfile
	worker           int
	root             string
	fallback         *compilerFallback
	oracle           *typescriptOracle
	timeout          time.Duration
}

func (e *engine) executionTimeout() time.Duration {
	if e.timeout > 0 {
		return e.timeout
	}
	return 15 * time.Second
}

func prepare(root string, test262 string, work string) (*engine, error) {
	return prepareMode(root, test262, work, nil, false)
}

func prepareProfile(root string, test262 string, work string, profile *runProfile) (*engine, error) {
	return prepareMode(root, test262, work, profile, false)
}

func prepareMode(root string, test262 string, work string, profile *runProfile, inProcess bool) (*engine, error) {
	done := profile.preparing("go-build")
	if err := os.MkdirAll(work, 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(filepath.Join(work, "results.jsonl"), nil, 0o644); err != nil {
		return nil, err
	}
	adamic := filepath.Join(work, "adamic")
	if !inProcess {
		build := exec.Command("go", "build", "-o", adamic, "./cmd/adamic")
		build.Dir = root
		if output, err := build.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("building adamic: %w\n%s", err, output)
		}
	}
	done()
	done = profile.preparing("runtime-library")
	include := filepath.Join(root, "internal", "native", "runtime")
	// Sanitizers, as the oracle compiles: a native memory bug is a crash, not a pass. Leaks are not
	// compared to Node (the process exits either way), so leak detection stays off at run time.
	flags := native.Flags(native.Options{Sanitize: true})
	library, err := native.RuntimeLibrary(include, native.Options{Sanitize: true})
	if err != nil {
		return nil, err
	}
	done()
	done = profile.preparing("cache-context")
	identityPath := adamic
	if inProcess {
		identityPath, err = os.Executable()
		if err != nil {
			return nil, err
		}
	}
	compilerBytes, err := os.ReadFile(identityPath)
	if err != nil {
		return nil, err
	}
	cache, version, context, err := prepareCache()
	if err != nil {
		return nil, err
	}
	sourceIdentity, err := loweringSourceIdentity(root)
	if err != nil {
		return nil, err
	}
	oracle, err := startTypescript(root, work)
	if err != nil {
		return nil, err
	}
	context = cacheKey(context, sourceIdentity)
	done()
	return &engine{
		oracle:  oracle,
		profile: profile,
		root:    root, fallback: &compilerFallback{}, inProcess: inProcess,
		test262:          test262,
		work:             work,
		adamic:           adamic,
		runtime:          native.RuntimeLinkFlags(library),
		runtimeKey:       filepath.Base(filepath.Dir(library)),
		compilerIdentity: cacheKey(string(compilerBytes)),
		cache:            cache, nodeVersion: version, context: context, nodeContext: cache.nodeContext,
		include: filepath.Dir(library),
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
	tests := make([]classified, len(files))
	results := make([]result, len(files))
	diagnostics := make([]string, len(files))
	ready := make([]chan struct{}, len(files))
	indices := make([]int, 0, len(files))
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
		test := classify(relative, string(source), e.adapt)
		tests[index] = test
		ready[index] = make(chan struct{})
		if test.Skip != "" {
			results[index] = result{Path: relative, Directory: test.Directory, Kind: outcomeSkipped, Reason: test.Skip}
			close(ready[index])
		} else if classifyOnly || (limit > 0 && attempted >= limit) {
			results[index] = result{Path: relative, Directory: test.Directory, Kind: outcomeUnrun, Reason: "not run"}
			close(ready[index])
		} else {
			attempted++
			indices = append(indices, index)
		}
	}
	jobs := e.jobs
	if jobs < 1 {
		jobs = 1
	}
	if jobs > len(indices) {
		jobs = len(indices)
	}
	if jobs > 1 {
		sort.SliceStable(indices, func(i, j int) bool { return len(tests[indices[i]].Program) > len(tests[indices[j]].Program) })
	}
	queue := make(chan int)
	var workers sync.WaitGroup
	locals := make([]engine, jobs)
	for worker := 0; worker < jobs; worker++ {
		local := *e
		local.worker = worker
		// Preserve the serial artifact path when there is just one worker.
		if jobs > 1 {
			local.work = filepath.Join(e.work, fmt.Sprintf("worker-%d", worker))
			if err := os.MkdirAll(local.work, 0755); err != nil {
				return report, err
			}
		}
		locals[worker] = local
	}
	for worker := range locals {
		local := locals[worker]
		workers.Go(func() {
			if local.inProcess {
				local.compiler = &compilerWorker{}
				defer local.compiler.close()
			}
			for index := range queue {
				var diagnostic bytes.Buffer
				local.log = &diagnostic
				one := local.attempt(tests[index])
				diagnostics[index] = diagnostic.String()
				if one.Kind == outcomeCrashed {
					one.Reason = withCrashPath(one.Path, one.Reason)
				}
				results[index] = one
				close(ready[index])
			}
		})
	}
	go func() {
		for _, index := range indices {
			queue <- index
		}
		close(queue)
	}()
	for index := range files {
		<-ready[index]
		fmt.Fprint(e.log, diagnostics[index])
		report.add(results[index])
		if !classifyOnly {
			// Retain every result, not just aggregate reasons, for before/after audits.
			encoded, err := json.Marshal(results[index])
			if err != nil {
				return report, err
			}
			file, err := os.OpenFile(filepath.Join(e.work, "results.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if err != nil {
				return report, err
			}
			_, writeErr := file.Write(append(encoded, '\n'))
			closeErr := file.Close()
			if writeErr != nil {
				return report, writeErr
			}
			if closeErr != nil {
				return report, closeErr
			}
		}
		if (index+1)%50 == 0 || index+1 == len(files) {
			fmt.Fprintf(e.log, "%s %d/%d pass=%d fail=%d refused=%d not-typescript=%d crashed=%d skipped=%d\n",
				filter, index+1, len(files), report.Pass, report.Fail, report.Refused, report.NotTypescript, report.Crashed, report.Skipped)
		}
	}
	workers.Wait()
	report.finish()
	return report, nil
}

// Stable paths are part of both execution commands, not omitted from their keys. The lock
// keeps concurrent invocations from overwriting a program while another invocation runs it.
func (e *engine) programDirectory(test classified) string {
	if e.cache == nil {
		return e.work
	}
	return filepath.Join(e.cache.directory, "programs", cacheKey(test.Program, e.nodeContext))
}

func (e *engine) attempt(test classified) result {
	profile := e.profile.begin(test.Path, e.worker)
	defer profile.finish()
	base := result{Path: test.Path, Directory: test.Directory, Adaptations: test.Adaptations}
	directory := e.programDirectory(test)
	if err := os.MkdirAll(directory, 0700); err != nil {
		// A broken observation cache must not prevent the uncached program from running.
		directory = e.work
		if err := os.MkdirAll(directory, 0700); err != nil {
			base.Kind = outcomeCrashed
			base.Reason = err.Error()
			return base
		}
	}
	lock, err := os.OpenFile(filepath.Join(directory, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil && directory != e.work {
		directory = e.work
		lock, err = os.OpenFile(filepath.Join(directory, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	}
	if err != nil {
		base.Kind = outcomeCrashed
		base.Reason = err.Error()
		return base
	}
	defer lock.Close()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		base.Kind = outcomeCrashed
		base.Reason = err.Error()
		return base
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	typescript := filepath.Join(directory, "program.a")
	module := filepath.Join(directory, "program.mts")
	if err := os.WriteFile(typescript, []byte(test.Program), 0o644); err != nil {
		return result{Path: test.Path, Directory: test.Directory, Kind: outcomeCrashed, Reason: err.Error()}
	}
	if err := os.WriteFile(module, []byte(test.Program), 0o644); err != nil {
		return result{Path: test.Path, Directory: test.Directory, Kind: outcomeCrashed, Reason: err.Error()}
	}
	cache := e.cache
	if dependentProgram(test.Program) {
		cache = nil
	}
	compilerCommand := cacheKey("adamic c", e.adamic, typescript, "2m", "16MiB")
	if e.compiler != nil {
		// The worker executes this runner, not the scratch copy of adamic used for fallback.
		// Its exact executable and protocol are already part of the runner context.
		compilerCommand = cacheKey("runner --compiler-worker", typescript, "2m", "16MiB")
	}
	compilerKey := compilerResultKey(test.Program, e.compilerIdentity, compilerCommand, e.context)
	lowered := profile.observation("compile", cache, compilerKey, func() (execution, bool) {
		var result execution
		if e.compiler != nil {
			result = e.compiler.compile(typescript)
			if result.Exit == -1 && !result.TimedOut {
				result = e.fallbackCompile(typescript)
			}
		} else {
			result = runCommandWithLimit(2*time.Minute, nil, 16<<20, e.adamic, "c", typescript)
		}
		// Refusals, compiler crashes and deadlines are always retried.
		return result, result.Exit == 0 && !result.TimedOut
	})
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
	cPath := filepath.Join(directory, "program.c")
	if err := os.WriteFile(cPath, []byte(lowered.Stdout), 0o644); err != nil {
		base.Kind = outcomeCrashed
		base.Reason = err.Error()
		return base
	}
	binary := filepath.Join(directory, "program.bin")
	arguments := append(append([]string{}, e.flags...), "-I", e.include, "-o", binary, cPath)
	arguments = append(arguments, e.runtime...)
	arguments = append(arguments, "-lm")
	nativeCommand := cacheKey(cacheKey(arguments...), binary, e.executionTimeout().String()+"-cpu", "2m-wall", fmt.Sprint(outputLimit), fmt.Sprint(nativeEnvironment))
	key := nativeResultKey(lowered.Stdout, e.runtimeKey, nativeCommand, e.context)
	linkFailed := false
	// Imported modules can read files or have mutable dependencies. Until their whole input
	// graph is keyed, all their observations run fresh.
	nativeRun := profile.observation("native-observation", cache, key, func() (execution, bool) {
		linked := profile.command("clang", func() execution { return runCommand(2*time.Minute, nil, "clang", arguments...) })
		if linked.Exit != 0 || linked.TimedOut {
			linkFailed = true
			return linked, false
		}
		defer os.Remove(binary)
		return profile.command("native", func() execution { return runProgram(e.executionTimeout(), nativeEnvironment, binary) }), true
	})
	// A link failure is a compiler crash rather than a native execution verdict.
	// Only successful links publish observations, so a hit always holds a real execution.
	// Link stderr has the same classification as the original runner.
	if linkFailed {
		base.Kind = outcomeCrashed
		base.Reason = "clang: " + firstLine(nativeRun.Stderr)
		return base
	}
	nodeCommand := cacheKey("node", "--disable-warning=ExperimentalWarning", module, e.executionTimeout().String()+"-cpu", "2m-wall", fmt.Sprint(outputLimit))
	nodeKey := nodeResultKey(test.Program, e.nodeVersion, fmt.Sprint(e.adapt), nodeCommand, e.nodeContext)
	nodeRun := profile.observation("node", cache, nodeKey, func() (execution, bool) {
		return runProgram(e.executionTimeout(), nil, "node", "--disable-warning=ExperimentalWarning", module), true
	})
	decided := decide(verdictInput{
		NegativePhase: test.NegativePhase,
		NegativeType:  test.NegativeType,
		Node:          nodeRun,
		Native:        nativeRun,
	})
	base.Kind = decided.Kind
	base.Reason = decided.Reason
	// The existing generic harness checks Error ancestry, not constructor
	// identity. Independent Node success cannot prove that missing native check.
	// Do not award a RegExp pass until exact constructor checks are supported.
	if base.Kind == outcomePass && test.ConstructorAssertion {
		base.Kind = outcomeRefused
		base.Reason = "not yet: constructor-identity assertion in RegExp harness"
		return base
	}
	if base.Kind == outcomePass && test.Original != "" {
		original := e.originalRegExp(test)
		if crashedExecution(original) {
			base.Kind = outcomeCrashed
			base.Reason = "original Node test: " + crashReason(verdictInput{Node: original})
		} else if test.NegativePhase == "runtime" {
			if original.Exit == 0 || !errorNamed(original.Stderr, test.NegativeType) {
				base.Kind = outcomeFail
				base.Reason = "original Node negative expectation not observed"
			}
		} else if original.Exit != 0 || original.Stdout != nodeRun.Stdout || original.Stderr != nodeRun.Stderr {
			base.Kind = outcomeFail
			base.Reason = "original Node test disagrees with adaptation: " + firstLine(original.Stderr)
		}
	}
	return base
}

var importedProgram = regexp.MustCompile(`(?:^|[^\w$])import(?:[^\w$]|$)`)
var referencedProgram = regexp.MustCompile(`(?m)^\s*///\s*<reference\b`)

func dependentProgram(program string) bool {
	return importedProgram.MatchString(codeOnly(program)) || referencedProgram.MatchString(program)
}

var nativeEnvironment = []string{
	"ASAN_OPTIONS=detect_leaks=0:abort_on_error=1:halt_on_error=1",
	"UBSAN_OPTIONS=halt_on_error=1:abort_on_error=1",
}

const outputLimit = 256 << 10

// Test programs get a CPU budget; time waiting for a core does not spend it.
// Keep a wall-clock backstop for programs that block without consuming CPU.
func runProgram(cpuLimit time.Duration, extra []string, name string, args ...string) execution {
	seconds := int64((cpuLimit + time.Second - 1) / time.Second)
	// Set only the soft limit so the kernel delivers SIGXCPU, not SIGKILL.
	wrapper := fmt.Sprintf(`ulimit -S -t %d || exit; exec "$0" "$@"`, seconds)
	arguments := append([]string{"-c", wrapper, name}, args...)
	return runCommand(2*time.Minute, extra, "/bin/sh", arguments...)
}

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
			if status.Signal() == syscall.SIGXCPU {
				result.TimedOut = true
				result.Exit = -1
				return result
			}
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

// The normal worker does not execute the backup compiler. Build it only if worker transport
// fails, once across all workers, so isolation still has the original diagnostic fallback.
type compilerFallback struct {
	once sync.Once
	err  error
}

func (e *engine) fallbackCompile(path string) execution {
	if e.fallback != nil && e.inProcess {
		e.fallback.once.Do(func() {
			build := exec.Command("go", "build", "-o", e.adamic, "./cmd/adamic")
			build.Dir = e.root
			if output, err := build.CombinedOutput(); err != nil {
				e.fallback.err = fmt.Errorf("building adamic: %w\n%s", err, output)
			}
		})
		if e.fallback.err != nil {
			return execution{Exit: -1, Stderr: e.fallback.err.Error()}
		}
	}
	return runCommandWithLimit(2*time.Minute, nil, 16<<20, e.adamic, "c", path)
}
