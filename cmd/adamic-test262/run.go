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

	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// engine owns one checkout's compiler, keyed runtime archive and observation cache. Each worker
// owns a compiler process; tests are classified and their observations reduced in serial order.
type engine struct {
	// Tests may supply their declared archive; commands select the runtime per source.
	runtimeLibrary   func(string, native.Options) (string, error)
	test262          string
	work             string
	adamic           string
	include          string
	log              io.Writer
	adapt            bool
	jobs             int
	runtimeKey       string
	cache            *resultCache
	nodeVersion      string
	context          string
	nodeContext      string
	compiler         *compilerWorker
	tools            runnerTools
	inProcess        bool
	compilerIdentity string
	profile          *runProfile
	worker           int
	fallback         *compilerFallback
	buildAdamic      adamicBuilder
}

// An adamicBuilder puts the stage 0 compiler at output: go build of ./cmd/adamic for the command, the shared keyed
// product for a test (buildcache.Adamic), so no test unit runs go build. A nil one has no compiler to give, for an
// in-process engine that never falls back.
type adamicBuilder func(output string) error

// goBuildAdamic builds ./cmd/adamic of the checkout at root.
func goBuildAdamic(root string) adamicBuilder {
	return func(output string) error {
		build := toolchain.CommandIn(root, "go", "build", "-o", output, "./cmd/adamic")
		if combined, combinedOutputError := build.CombinedOutput(); combinedOutputError != nil {
			return fmt.Errorf("building adamic: %w\n%s", combinedOutputError, combined)
		}
		return nil
	}
}

func prepareMode(
	root string,
	test262 string,
	work string,
	buildAdamic adamicBuilder,
	profile *runProfile,
	inProcess bool,
) (*engine, error) {
	done := profile.preparing("go-build")
	if createDirectoryError := os.MkdirAll(work, 0o755); createDirectoryError != nil {
		return nil, createDirectoryError
	}
	adamic := filepath.Join(work, "adamic")
	if !inProcess {
		if buildAdamic == nil {
			return nil, fmt.Errorf("a subprocess engine needs a stage 0 compiler")
		}
		if buildAdamicError := buildAdamic(adamic); buildAdamicError != nil {
			return nil, buildAdamicError
		}
	}
	done()
	include := filepath.Join(root, "internal", "native", "runtime")
	var executableError error
	done = profile.preparing("cache-context")
	identityPath := adamic
	if inProcess {
		identityPath, executableError = os.Executable()
		if executableError != nil {
			return nil, executableError
		}
	}
	compilerBytes, readError := os.ReadFile(identityPath)
	if readError != nil {
		return nil, readError
	}
	tools, resolveToolsError := resolveTools()
	if resolveToolsError != nil {
		return nil, resolveToolsError
	}
	cache, version, context, prepareCacheError := prepareCache(root, tools)
	if prepareCacheError != nil {
		return nil, prepareCacheError
	}
	sourceIdentity, loweringSourceIdentityError := loweringSourceIdentity(root)
	if loweringSourceIdentityError != nil {
		return nil, loweringSourceIdentityError
	}
	context = cacheKey(context, sourceIdentity)
	done()
	return &engine{
		profile:  profile,
		fallback: &compilerFallback{}, inProcess: inProcess, buildAdamic: buildAdamic,
		test262:          test262,
		work:             work,
		adamic:           adamic,
		compilerIdentity: cacheKey(string(compilerBytes)),
		cache:            cache, nodeVersion: version, context: context, nodeContext: cache.nodeContext,
		tools:   tools,
		include: include,
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

func (mainEngine *engine) runFilter(filter string, limit int, classifyOnly bool) (filterReport, error) {
	files, listTestsError := listTests(mainEngine.test262, filter)
	if listTestsError != nil {
		return filterReport{}, listTestsError
	}
	report := filterReport{Path: filter}
	attempted := 0
	tests := make([]classified, len(files))
	results := make([]result, len(files))
	ready := make([]chan struct{}, len(files))
	indices := make([]int, 0, len(files))
	for index, file := range files {
		source, readError := os.ReadFile(file)
		if readError != nil {
			return report, readError
		}
		relative, relativePathError := filepath.Rel(filepath.Join(mainEngine.test262, "test"), file)
		if relativePathError != nil {
			relative = file
		}
		relative = filepath.ToSlash(relative)
		test := classify(relative, string(source), mainEngine.adapt)
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
	jobs := mainEngine.jobs
	if jobs < 1 {
		jobs = 1
	}
	if jobs > len(indices) {
		jobs = len(indices)
	}
	if jobs > 1 {
		sort.SliceStable(
			indices,
			func(left, right int) bool {
				return len(tests[indices[left]].Program) > len(tests[indices[right]].Program)
			},
		)
	}
	queue := make(chan int)
	var workers sync.WaitGroup
	locals := make([]engine, jobs)
	for worker := 0; worker < jobs; worker++ {
		local := *mainEngine
		local.worker = worker
		// Preserve the serial artifact path when there is just one worker.
		if jobs > 1 {
			local.work = filepath.Join(mainEngine.work, fmt.Sprintf("worker-%d", worker))
			if createDirectoryError := os.MkdirAll(local.work, 0755); createDirectoryError != nil {
				return report, createDirectoryError
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
				one := local.attempt(tests[index])
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
		report.add(results[index])
		if (index+1)%50 == 0 || index+1 == len(files) {
			fmt.Fprintf(mainEngine.log, "%s %d/%d pass=%d fail=%d refused=%d crashed=%d skipped=%d\n",
				filter, index+1, len(files), report.Pass, report.Fail, report.Refused, report.Crashed, report.Skipped)
		}
	}
	workers.Wait()
	report.finish()
	return report, nil
}

// Stable paths are part of both execution commands, not omitted from their keys. The lock
// keeps concurrent invocations from overwriting a program while another invocation runs it.
func (mainEngine *engine) programDirectory(test classified) string {
	if mainEngine.cache == nil {
		return mainEngine.work
	}
	return filepath.Join(mainEngine.cache.directory, "programs", cacheKey(test.Program, mainEngine.nodeContext))
}

func (mainEngine *engine) attempt(test classified) result {
	profile := mainEngine.profile.begin(test.Path, mainEngine.worker)
	defer profile.finish()
	base := result{Path: test.Path, Directory: test.Directory, Adaptations: test.Adaptations}
	directory := mainEngine.programDirectory(test)
	if createDirectoryError := os.MkdirAll(directory, 0700); createDirectoryError != nil {
		// A broken observation cache must not prevent the uncached program from running.
		directory = mainEngine.work
		if createDirectoryError := os.MkdirAll(directory, 0700); createDirectoryError != nil {
			base.Kind = outcomeCrashed
			base.Reason = createDirectoryError.Error()
			return base
		}
	}
	lock, err := os.OpenFile(filepath.Join(directory, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil && directory != mainEngine.work {
		directory = mainEngine.work
		lock, err = os.OpenFile(filepath.Join(directory, ".lock"), os.O_CREATE|os.O_RDWR, 0600)
	}
	if err != nil {
		base.Kind = outcomeCrashed
		base.Reason = err.Error()
		return base
	}
	defer lock.Close()
	if flockError := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); flockError != nil {
		base.Kind = outcomeCrashed
		base.Reason = flockError.Error()
		return base
	}
	defer syscall.Flock(int(lock.Fd()), syscall.LOCK_UN)
	typescript := filepath.Join(directory, "program.a")
	module := filepath.Join(directory, "program.mts")
	if writeError := os.WriteFile(typescript, []byte(test.Program), 0o644); writeError != nil {
		return result{Path: test.Path, Directory: test.Directory, Kind: outcomeCrashed, Reason: writeError.Error()}
	}
	if writeError := os.WriteFile(module, []byte(test.Program), 0o644); writeError != nil {
		return result{Path: test.Path, Directory: test.Directory, Kind: outcomeCrashed, Reason: writeError.Error()}
	}
	cache := mainEngine.cache
	if dependentProgram(test.Program) {
		cache = nil
	}
	compilerCommand := cacheKey("adamic c", mainEngine.adamic, typescript, "2m", "16MiB")
	if mainEngine.compiler != nil {
		// The worker executes this runner, not the scratch copy of adamic used for fallback.
		// Its exact executable and protocol are already part of the runner context.
		compilerCommand = cacheKey("runner --compiler-worker", typescript, "2m", "16MiB")
	}
	compilerKey := compilerResultKey(test.Program, mainEngine.compilerIdentity, compilerCommand, mainEngine.context)
	lowered := profile.observation("compile", cache, compilerKey, func() (execution, bool) {
		var result execution
		if mainEngine.compiler != nil {
			result = mainEngine.compiler.compile(typescript)
			if result.Exit == -1 && !result.TimedOut {
				result = mainEngine.fallbackCompile(typescript)
			}
		} else {
			result = runCommandWithLimit(2*time.Minute, nil, 16<<20, mainEngine.adamic, "c", typescript)
		}
		// Refusals, compiler crashes and deadlines are always retried.
		return result, result.Exit == 0 && !result.TimedOut
	})
	kind, reason := compileClass(lowered.Stderr, lowered.Exit, lowered.TimedOut)
	if kind == "refused" {
		base.Kind = outcomeRefused
		base.Reason = reason
		return base
	}
	if kind == "crashed" || lowered.Exit != 0 {
		base.Kind = outcomeCrashed
		base.Reason = reason
		return base
	}
	cPath := filepath.Join(directory, "program.c")
	if writeError := os.WriteFile(cPath, []byte(lowered.Stdout), 0o644); writeError != nil {
		base.Kind = outcomeCrashed
		base.Reason = writeError.Error()
		return base
	}
	binary := filepath.Join(directory, "program.bin")
	options := native.Options{Sanitize: true}
	var library string
	if mainEngine.runtimeLibrary != nil {
		library, err = mainEngine.runtimeLibrary(lowered.Stdout, options)
	} else {
		library, err = native.RuntimeLibraryForSource(mainEngine.include, lowered.Stdout, options)
	}
	if err != nil {
		base.Kind, base.Reason = outcomeCrashed, err.Error()
		return base
	}
	arguments := append(native.SourceFlags(lowered.Stdout, options), "-I", filepath.Dir(library), "-o", binary, cPath)
	arguments = append(arguments, native.RuntimeLinkFlags(library)...)
	arguments = append(arguments, "-lm")
	nativeCommand := cacheKey(
		cacheKey(arguments...),
		binary,
		"15s-cpu",
		"2m-wall",
		fmt.Sprint(outputLimit),
		fmt.Sprint(nativeEnvironment),
	)
	key := nativeResultKey(
		lowered.Stdout,
		cacheKey(filepath.Base(filepath.Dir(library)), mainEngine.runtimeKey),
		nativeCommand,
		mainEngine.context,
	)
	linkFailed := false
	// Imported modules can read files or have mutable dependencies. Until their whole input
	// graph is keyed, all their observations run fresh.
	nativeRun := profile.observation("native-observation", cache, key, func() (execution, bool) {
		linked := profile.command(
			"clang",
			func() execution { return runCommand(2*time.Minute, nil, mainEngine.tools.clang, arguments...) },
		)
		if linked.Exit != 0 || linked.TimedOut {
			linkFailed = true
			return linked, false
		}
		defer os.Remove(binary)
		return profile.command(
			"native",
			func() execution { return runProgram(15*time.Second, nativeEnvironment, binary) },
		), true
	})
	// A link failure is a compiler crash rather than a native execution verdict.
	// Only successful links publish observations, so a hit always holds a real execution.
	// Link stderr has the same classification as the original runner.
	if linkFailed {
		base.Kind = outcomeCrashed
		base.Reason = "clang: " + firstLine(nativeRun.Stderr)
		return base
	}
	nodeCommand := cacheKey(
		"node",
		"--disable-warning=ExperimentalWarning",
		module,
		"15s-cpu",
		"2m-wall",
		fmt.Sprint(outputLimit),
	)
	nodeKey := nodeResultKey(
		test.Program,
		mainEngine.nodeVersion,
		fmt.Sprint(mainEngine.adapt),
		nodeCommand,
		mainEngine.nodeContext,
	)
	nodeRun := profile.observation("node", cache, nodeKey, func() (execution, bool) {
		return runProgram(
			15*time.Second,
			nil,
			mainEngine.tools.node,
			"--disable-warning=ExperimentalWarning",
			module,
		), true
	})
	decided := decide(verdictInput{
		NegativePhase: test.NegativePhase,
		NegativeType:  test.NegativeType,
		Node:          nodeRun,
		Native:        nativeRun,
	})
	base.Kind = decided.Kind
	base.Reason = decided.Reason
	if base.Kind == outcomePass && test.ConstructorAssertion {
		base.Kind = outcomeRefused
		base.Reason = "not yet: constructor-identity assertion in RegExp harness"
	}
	if base.Kind == outcomePass && test.Original != "" {
		original := mainEngine.originalRegExp(test)
		if crashedExecution(original) {
			base.Kind = outcomeCrashed
			base.Reason = "original Node: " + crashReason(verdictInput{Node: original})
		} else if test.NegativePhase == "runtime" {
			if original.Exit == 0 ||
				!errorNamed(original.Stderr, test.NegativeType) ||
				original.Stdout != nodeRun.Stdout {
				base.Kind = outcomeFail
				base.Reason = "original Node negative test disagrees with adaptation"
			}
		} else if original.Exit != 0 || original.Stdout != nodeRun.Stdout || original.Stderr != nodeRun.Stderr {
			base.Kind = outcomeFail
			base.Reason = "original Node test disagrees with adaptation"
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
	timeoutContext, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	command := exec.CommandContext(timeoutContext, name, args...)
	command.Env = childEnvironment(extra...)
	var stdout, stderr limitedBuffer
	stdout.limit = limit
	stderr.limit = limit
	captureToError := childguard.CaptureTo(timeoutContext, command, &stdout, &stderr)
	result := execution{Stdout: stdout.String(), Stderr: stderr.String()}
	if stdout.exceeded || stderr.exceeded {
		result.Exit = -1
		result.Stderr = "command output exceeded capture limit\n"
		return result
	}
	if timeoutContext.Err() == context.DeadlineExceeded {
		result.TimedOut = true
		result.Exit = -1
		return result
	}
	if captureToError == nil {
		return result
	}
	exit, isExit := captureToError.(*exec.ExitError)
	if !isExit {
		result.Exit = -1
		result.Stderr = result.Stderr + captureToError.Error()
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
		result.Exit = status.ExitStatus()
		return result
	}
	result.Exit = exit.ExitCode()
	return result
}

// limitedBuffer keeps the start of a stream and drops the rest, so a test that prints without end
// cannot fill memory before its timeout.
type limitedBuffer struct {
	buf      bytes.Buffer
	limit    int
	exceeded bool
}

func (buffer *limitedBuffer) Write(data []byte) (int, error) {
	written := len(data)
	room := buffer.limit - buffer.buf.Len()
	if written > room {
		buffer.exceeded = true
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
	info, statError := os.Stat(target)
	if statError != nil {
		return nil, fmt.Errorf("test262 filter %s: %w", filter, statError)
	}
	if !info.IsDir() {
		return []string{target}, nil
	}
	var files []string
	walkError := filepath.WalkDir(target, func(path string, entry os.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
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
	if walkError != nil {
		return nil, walkError
	}
	sort.Strings(files)
	return files, nil
}

// filterReport is one directory filter's counts. Directories are the folders that hold tests.
type filterReport struct {
	Path           string        `json:"path"`
	Pass           int           `json:"pass"`
	Fail           int           `json:"fail"`
	Refused        int           `json:"refused"`
	Crashed        int           `json:"crashed"`
	Skipped        int           `json:"skipped"`
	Unrun          int           `json:"unrun,omitempty"`
	Total          int           `json:"total"`
	Directories    []dirCount    `json:"directories"`
	RefusalReasons []reasonCount `json:"refusalReasons"`
	SkipReasons    []reasonCount `json:"skipReasons"`
	CrashReasons   []reasonCount `json:"crashReasons"`
	FailReasons    []reasonCount `json:"failReasons"`
	Passes         []string      `json:"passes,omitempty"`
	Adaptations    []reasonCount `json:"adaptations,omitempty"`
	directories    map[string]*dirCount
	refusalReasons map[string]int
	skipReasons    map[string]int
	crashReasons   map[string]int
	failReasons    map[string]int
	adaptations    map[string]int
}

type dirCount struct {
	Path    string `json:"path"`
	Pass    int    `json:"pass"`
	Fail    int    `json:"fail"`
	Refused int    `json:"refused"`
	Crashed int    `json:"crashed"`
	Skipped int    `json:"skipped"`
	Unrun   int    `json:"unrun,omitempty"`
	Total   int    `json:"total"`
}

type reasonCount struct {
	Reason string `json:"reason"`
	Count  int    `json:"count"`
}

func (report *filterReport) add(one result) {
	if report.directories == nil {
		report.directories = map[string]*dirCount{}
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
	sort.Slice(
		report.Directories,
		func(left int, right int) bool { return report.Directories[left].Path < report.Directories[right].Path },
	)
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
	sort.Slice(reasons, func(left int, right int) bool {
		if reasons[left].Count != reasons[right].Count {
			return reasons[left].Count > reasons[right].Count
		}
		return reasons[left].Reason < reasons[right].Reason
	})
	return reasons
}

// reportDocument is what --json prints.
type reportDocument struct {
	Test262 string         `json:"test262"`
	Commit  string         `json:"commit,omitempty"`
	Adapt   bool           `json:"adapt"`
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
	command := toolchain.GitCommand("-C", test262, "rev-parse", "HEAD")
	output, outputError := command.Output()
	if outputError != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

func writeJSON(writer io.Writer, document reportDocument) error {
	encoded, encodeError := json.MarshalIndent(document, "", "  ")
	if encodeError != nil {
		return encodeError
	}
	_, writeError := writer.Write(append(encoded, '\n'))
	return writeError
}

// The normal worker does not execute the backup compiler. Build it only if worker transport
// fails, once across all workers, so isolation still has the original diagnostic fallback.
type compilerFallback struct {
	once sync.Once
	err  error
}

func (mainEngine *engine) fallbackCompile(path string) execution {
	if mainEngine.fallback != nil && mainEngine.inProcess {
		mainEngine.fallback.once.Do(func() {
			if mainEngine.buildAdamic == nil {
				mainEngine.fallback.err = fmt.Errorf("this engine has no stage 0 compiler to fall back to")
				return
			}
			mainEngine.fallback.err = mainEngine.buildAdamic(mainEngine.adamic)
		})
		if mainEngine.fallback.err != nil {
			return execution{Exit: -1, Stderr: mainEngine.fallback.err.Error()}
		}
	}
	return runCommandWithLimit(2*time.Minute, nil, 16<<20, mainEngine.adamic, "c", path)
}
