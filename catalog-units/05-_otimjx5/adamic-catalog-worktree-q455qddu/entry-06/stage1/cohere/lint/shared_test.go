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

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
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
	directory, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	return directory
}()

// isPackage is whether directory is this package, the unmodified port, rather than a test's copy of it.
func isPackage(directory string) bool {
	absolute, err := filepath.Abs(directory)
	return err == nil && absolute == packageDirectory
}

// sharedPath is a fresh path under the run's shared directory.
func sharedPath(name string) (string, error) {
	directory, err := os.MkdirTemp(sharedDirectory, "")
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, name), nil
}

// run is execute without a test: output goes to a file, never a pipe, and anything on standard error is
// a failure, as execute requires.
func run(directory string, environment []string, name string, args ...string) ([]byte, error) {
	command := exec.Command(name, args...)
	command.Dir = directory
	// An explicit environment loses the PWD os/exec sets from Dir, and the go command trusts PWD over the
	// real working directory, so a stale one resolves the module through the wrong path.
	command.Env = append(os.Environ(), environment...)
	if directory != "" {
		command.Env = append(command.Env, "PWD="+directory)
	}
	output, err := os.CreateTemp(sharedDirectory, "stdout-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(output.Name())
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := testguard.Run(command, testguard.Budget, testguard.Ceiling); err != nil || stderr.Len() != 0 {
		return nil, fmt.Errorf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	return os.ReadFile(output.Name())
}

func buildPortTo(directory string, sanitize bool, binary string) error {
	if _, err := registry.Generate(directory); err != nil {
		return err
	}
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		return err
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return err
	}
	return native.Build(native.C(lowered), binary, native.Options{Sanitize: sanitize})
}

func emitJavaScriptTo(directory, module string) error {
	if _, err := registry.Generate(directory); err != nil {
		return err
	}
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		return err
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return err
	}
	return os.WriteFile(module, []byte(javascript.JavaScript(lowered)), 0644)
}

// goOracleIn builds the Go oracle over sourceRoot's rules into directory, through an overlay inside cohere
// so the upstream rules stay unmodified.
func goOracleIn(sourceRoot, directory string) (string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return "", err
	}
	side, err := filepath.Abs(filepath.Join(packageDirectory, "testdata/oracle.go"))
	if err != nil {
		return "", err
	}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return "", err
	}
	replacements := map[string]string{}
	var virtualFiles []string
	var failure error
	add := func(name, source string) {
		virtual := filepath.Join(root, "adamic_lint_"+name+".go")
		absolute, err := filepath.Abs(source)
		if err != nil {
			failure = err
			return
		}
		replacements[virtual] = absolute
		virtualFiles = append(virtualFiles, virtual)
	}
	add("oracle", side)
	add("registry", filepath.Join(sourceRoot, ".generated/registry.go"))
	for _, d := range descriptors {
		add(strings.ReplaceAll(d.Slug, "-", "_"), filepath.Join(sourceRoot, "rules", d.Slug, "oracle.go"))
	}
	if failure != nil {
		return "", failure
	}
	overlay, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		return "", err
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		return "", err
	}
	binary := filepath.Join(directory, "oracle")
	args := append([]string{"build", "-overlay=" + path, "-o", binary}, virtualFiles...)
	if _, err := run(root, nil, "go", args...); err != nil {
		return "", err
	}
	return binary, nil
}

// captureUpstream runs cohere's own tests for sourceRoot's rules under a capture overlay and writes each
// unique asserted case as a file under directory, returning one manifest row per case. Every Run is
// captured, including tests that assert repair fields directly; the overlay changes no rule. The capture's
// destination is passed in the subprocess's environment, so no test's process environment changes.
func captureUpstream(sourceRoot, directory string) ([]string, error) {
	root, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		return nil, err
	}
	harness := filepath.Join(root, "internal/lint/testing/rule_testing.go")
	data, err := os.ReadFile(harness)
	if err != nil {
		return nil, err
	}
	original := "return Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}"
	replacement := "result := Result{Diagnostics: diagnostics, SourceFile: sourceFile, capture: captured}\n RecordAssertedCase(t, result)\n return result"
	if strings.Count(string(data), original) != 1 {
		return nil, fmt.Errorf("capture overlay anchor changed")
	}
	side := filepath.Join(directory, "rule_testing.go")
	if err := os.WriteFile(side, []byte(strings.Replace(string(data), original, replacement, 1)), 0644); err != nil {
		return nil, err
	}
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{harness: side}})
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
		return nil, err
	}
	capture := filepath.Join(directory, "capture")
	environment := []string{"COHERE_DOCS_CAPTURE=" + capture}
	descriptors, err := registry.Generate(sourceRoot)
	if err != nil {
		return nil, err
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
		if _, err := run(root, environment, "go", "test", "-overlay="+overlayPath, "./internal/lint/rules/"+name, "-run", "^("+strings.Join(packages[name], "|")+")", "-count=1", "-timeout=0"); err != nil {
			return nil, err
		}
	}
	files, err := filepath.Glob(filepath.Join(capture, "*.jsonl"))
	if err != nil {
		return nil, err
	}
	type record struct {
		Rule, File, Source, Outcome, FixedSource string
		Options                                  json.RawMessage
	}
	unique := map[string]record{}
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		for _, line := range bytes.Split(data, []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var row record
			if err := json.Unmarshal(line, &row); err != nil {
				return nil, err
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
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, []byte(row.Source), 0644); err != nil {
			return nil, err
		}
		var legacy struct {
			Mode, Null      string
			AllowEmptyCatch bool
		}
		if len(row.Options) > 0 && row.Options[0] == '{' {
			if err := json.Unmarshal(row.Options, &legacy); err != nil {
				return nil, err
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
		if row.Rule == "no-div-regex" && (row.Source == "var a = /;" || row.Source == "var a = /" || row.Source == "var a = [/];" || row.Source == "if (/) {}" || row.Source == "var a = /=") {
			mode = "recovery"
		}
		rows = append(rows, fmt.Sprintf("%s\t%s\t%s\t%s\t%t\t%s\t%s", path, row.Rule, legacy.Mode, legacy.Null, legacy.AllowEmptyCatch, string(row.Options), mode))
	}
	if len(rows) < 150 {
		return nil, fmt.Errorf("capture unexpectedly small: %d cases", len(rows))
	}
	return rows, nil
}
