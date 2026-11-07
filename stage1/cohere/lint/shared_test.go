package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/testguard"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

// What every test in one run can share, made once (#axg2xys): the unmodified port's sanitized and release
// builds and its emitted JavaScript, the Go oracle over this package's rules, and the upstream capture.
// Each used to be rebuilt by every test that asked, which is most of the package's time and grows with
// every rule. A mutant's copy of the port is never shared: only a directory that is this package itself is.
//
// What a shared helper returns is read-only. A test that changes an artifact copies it first.
//
// sharedDirectory lives for the run; TestMain makes and removes it, so nothing a test's TempDir owns is
// ever handed to another test.
var sharedDirectory string

type sharedValue struct {
	once   sync.Once
	path   string
	rows   []string
	report string
	err    error
}

var sharedValues sync.Map

// shared is the run's one value for key. Its maker never calls t.Fatal: a goroutine that exits inside a
// sync.Once marks it done with nothing in it, and every later test would read an empty answer.
func shared(key string, make func(value *sharedValue)) *sharedValue {
	stored, _ := sharedValues.LoadOrStore(key, &sharedValue{})
	value := stored.(*sharedValue)
	value.once.Do(func() { make(value) })
	return value
}

var packageDirectory = func() string {
	directory, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		panic(absolutePathError)
	}
	return directory
}()

// isPackage is whether directory is this package, the unmodified port, rather than a test's copy of it.
func isPackage(directory string) bool {
	absolute, absolutePathError := filepath.Abs(directory)
	return absolutePathError == nil && absolute == packageDirectory
}

// sharedPath is a fresh path under the run's shared directory.
func sharedPath(name string) (string, error) {
	directory, temporaryDirectoryError := os.MkdirTemp(sharedDirectory, "")
	if temporaryDirectoryError != nil {
		return "", temporaryDirectoryError
	}
	return filepath.Join(directory, name), nil
}

// run is execute without a test, for a tool: output goes to a file, never a pipe, and anything on standard error
// is a failure, as execute requires. The tool is the one internal/toolchain resolves, under its environment plus
// environment, in directory with PWD to match (the go command trusts PWD over the real working directory).
func run(directory string, environment []string, name string, args ...string) ([]byte, error) {
	command := toolchain.CommandIn(directory, name, args...)
	command.Env = append(command.Env, environment...)
	output, temporaryFileError := os.CreateTemp(sharedDirectory, "stdout-")
	if temporaryFileError != nil {
		return nil, temporaryFileError
	}
	defer os.Remove(output.Name())
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if runError := testguard.Run(
		command,
		testguard.Budget,
		testguard.Ceiling,
	); runError != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		return nil, fmt.Errorf("%s %v: %v\n%s", name, args, runError, &stderr)
	}
	return os.ReadFile(output.Name())
}

func buildPortTo(directory string, sanitize bool, binary string) error {
	if _, generateError := registry.Generate(directory); generateError != nil {
		return generateError
	}
	program, loadError := load.Load([]string{filepath.Join(directory, "main.ts")})
	if loadError != nil {
		return loadError
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		return lowerError
	}
	return native.Build(native.C(lowered), binary, native.Options{Sanitize: sanitize})
}

func emitJavaScriptTo(directory, module string) error {
	if _, generateError := registry.Generate(directory); generateError != nil {
		return generateError
	}
	program, loadError := load.Load([]string{filepath.Join(directory, "main.ts")})
	if loadError != nil {
		return loadError
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		return lowerError
	}
	return os.WriteFile(module, []byte(javascript.JavaScript(lowered)), 0644)
}

// lintGoSettings is every go setting a go run under toolchain.Env builds with, as go reports it there: its CC is the
// compiler a cgo build runs, which the caller's CC isn't when the caller sets none (toolchain names clang then). A
// product that builds with toolchain's go keys it.
func lintGoSettings() string {
	return buildcache.Tool("go", append([]string{"env"}, toolchain.GoSettings...)...)
}

// lintCommand is a child for a test's own deadline, run in directory (the caller's when empty) and killed with its
// whole process group when ctx ends. A bare name is a tool, the program internal/toolchain resolves under its
// environment with PWD set to directory, never PATH's with this process's; a path is a program a test built, which
// runs as named with this process's environment. A tool toolchain can't resolve fails as the command starts, with
// toolchain's error.
//
// No WaitDelay: one closed exec's reads of a buffered stream a second after the child exited and reported
// exec.ErrWaitDelay, so a run that had finished but was slow to drain under load failed by chance (#n3gk22h's shape).
// A caller that reads a pipe reads it through childguard.Capture. One that writes stdout to a file, the package's rule
// for Node's writev path, keeps exec's copy of stderr, read to its end: node, the port and the Go oracle start no
// process that could hold it, and ctx's group kill or the unit's ceiling ends a wait on one that did.
func lintCommand(ctx context.Context, directory, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	if !strings.ContainsRune(name, filepath.Separator) {
		if path, pathError := toolchain.Path(name); pathError != nil {
			command.Err = pathError
		} else {
			command = exec.CommandContext(ctx, path, args...)
			command.Env = toolchain.Env()
			if directory != "" {
				absolute, absolutePathError := filepath.Abs(directory)
				if absolutePathError != nil {
					command.Err = absolutePathError
				}
				command.Env = append(command.Env, "PWD="+absolute)
			}
		}
	}
	command.Dir = directory
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if killError == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

// captureUpstream runs cohere's own tests for sourceRoot's rules under a capture overlay and writes each
// unique asserted case as a file under directory, returning one manifest row per case. Every Run is
// captured, including tests that assert repair fields directly; the overlay changes no rule. The capture's
// destination is passed in the subprocess's environment, so no test's process environment changes.
func captureUpstream(sourceRoot, directory string) ([]string, error) {
	root, absolutePathError := filepath.Abs(filepath.Join(repository, "cohere"))
	if absolutePathError != nil {
		return nil, absolutePathError
	}
	harness := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, readError := os.ReadFile(harness)
	if readError != nil {
		return nil, readError
	}
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	if strings.Count(string(data), original) != 1 {
		return nil, fmt.Errorf("capture overlay anchor changed")
	}
	side := filepath.Join(directory, "rule_testing.go")
	if writeError := os.WriteFile(
		side,
		[]byte(strings.Replace(string(data), original, replacement, 1)),
		0644,
	); writeError != nil {
		return nil, writeError
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if writeError := os.WriteFile(overlayPath, overlay, 0644); writeError != nil {
		return nil, writeError
	}
	capture := filepath.Join(directory, "capture")
	environment := []string{"COHERE_DOCS_CAPTURE=" + capture}
	descriptors, generateError := registry.Generate(sourceRoot)
	if generateError != nil {
		return nil, generateError
	}
	discovered := map[string]bool{}
	packages := map[string][]string{}
	for _, d := range descriptors {
		discovered[d.Name] = true
		packages[d.UpstreamPackage] = append(packages[d.UpstreamPackage], d.UpstreamTest)
	}
	var names []string
	for name := range packages {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, runError := run(
			root,
			environment,
			"go",
			"test",
			"-overlay="+overlayPath,
			"./internal/lint/rules/"+name,
			"-run",
			"^("+strings.Join(packages[name], "|")+")",
			"-count=1",
			"-timeout=0",
		); runError != nil {
			return nil, runError
		}
	}
	files, globError := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	if globError != nil {
		return nil, globError
	}
	type record struct {
		Rule, File, Source, Outcome, FixedSource string
		Options                                  json.RawMessage
	}
	unique := map[string]record{}
	// The raw capture is the upstream tests' own record, a file per process named by its pid, its lines in the
	// order parallel tests wrote them: never a product's bytes, which must be the same build to build (landable-15's
	// tree, Oct 11: lint-rules-agree-capture held capture-<pid>.jsonl). The rows and cases made from it are sorted.
	defer os.RemoveAll(capture)
	for _, path := range files {
		data, readError := os.ReadFile(path)
		if readError != nil {
			return nil, readError
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row record
			if decodeError := json.Unmarshal(line, &row); decodeError != nil {
				return nil, decodeError
			}
			if !discovered[row.Rule] {
				continue
			}
			key := fmt.Sprintf("%s\t%s\t%+v\t%s", row.Rule, row.File, row.Options, row.Source)
			unique[key] = row
		}
	}
	var keys []string
	for key := range unique {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var rows []string
	for i, key := range keys {
		row := unique[key]
		// The case keeps its file name's directories, not only its base name: a rule that judges a
		// path (a utils folder, a page directory) reads them, and Go's capture recorded them.
		name := filepath.Clean(strings.TrimLeft(strings.ReplaceAll(row.File, "\\", "/"), "/"))
		if name == "." || name == "" || strings.HasPrefix(name, "..") {
			name = filepath.Base(name)
		}
		if name == "." || name == "" || name == ".." {
			name = "source.ts"
		}
		caseDirectory := filepath.Join(directory, fmt.Sprintf("case-%03d", i))
		path := filepath.Join(caseDirectory, name)
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0755); createDirectoryError != nil {
			return nil, createDirectoryError
		}
		if writeError := os.WriteFile(path, []byte(row.Source), 0644); writeError != nil {
			return nil, writeError
		}
		var legacy struct {
			Mode, Null      string
			AllowEmptyCatch bool
		}
		if len(row.Options) > 0 && row.Options[0] == '{' {
			if decodeError := json.Unmarshal(row.Options, &legacy); decodeError != nil {
				return nil, decodeError
			}
		}
		mode := ""
		if row.Rule == "@typescript-eslint/method-signature-style" {
			switch row.Source {
			case "type T = { m: => void };":
				mode = "recovery"
			case "interface I", "interface I { m(a: string): void;", "interface I { m<(a: string): void; }", "interface I { m<T(a: T): T; }":
				mode = "unsupported-recovery"
			}
		}
		if row.Rule == "no-div-regex" && (row.Source == "var a = /;" ||
			row.Source == "var a = /" ||
			row.Source == "var a = [/];" ||
			row.Source == "if (/) {}" ||
			row.Source == "var a = /=") {
			mode = "recovery"
		}
		rows = append(rows, fmt.Sprintf(
			"%s\t%s\t%s\t%s\t%t\t%s\t%s",
			path,
			row.Rule,
			legacy.Mode,
			legacy.Null,
			legacy.AllowEmptyCatch,
			string(row.Options),
			mode,
		))
	}
	if len(rows) < 150 {
		return nil, fmt.Errorf("capture unexpectedly small: %d cases", len(rows))
	}
	return rows, nil
}
