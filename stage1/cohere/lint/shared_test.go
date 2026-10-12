package lint

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/testguard"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

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

// run is execute without a test, for a tool: output goes to a file, never a pipe, and anything on standard error
// is a failure, as execute requires. The tool is the one internal/toolchain resolves, under its environment plus
// environment, in directory with PWD to match (the go command trusts PWD over the real working directory).
func run(directory string, environment []string, name string, args ...string) ([]byte, error) {
	command := toolchain.CommandIn(directory, name, args...)
	command.Env = append(command.Env, environment...)
	output, err := os.CreateTemp(sharedDirectory, "stdout-")
	if err != nil {
		return nil, err
	}
	defer os.Remove(output.Name())
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := testguard.Run(command, testguard.Budget, testguard.Ceiling); err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
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

// goOraclePackage is the directory, inside cohere, of the main package goOracleOverlay adds: the Go oracle.
const goOraclePackage = "adamic_lint_oracle"

// goOracleOverlay writes into directory the overlay that lays the Go oracle over cohere, so the upstream rules stay
// unmodified: testdata/oracle.go, sourceRoot's generated registry and each of its rules' oracle.go, each a file of
// goOraclePackage, a main package the overlay adds to cohere. It returns the overlay's path.
func goOracleOverlay(sourceRoot, directory string) (string, error) {
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
	var failure error
	add := func(name, source string) {
		virtual := filepath.Join(root, goOraclePackage, "adamic_lint_"+name+".go")
		absolute, err := filepath.Abs(source)
		if err != nil {
			failure = err
			return
		}
		replacements[virtual] = absolute
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
	return path, nil
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
func lintCommand(ctx context.Context, directory, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	if !strings.ContainsRune(name, filepath.Separator) {
		if path, err := toolchain.Path(name); err != nil {
			command.Err = err
		} else {
			command = exec.CommandContext(ctx, path, args...)
			command.Env = toolchain.Env()
			if directory != "" {
				absolute, err := filepath.Abs(directory)
				if err != nil {
					command.Err = err
				}
				command.Env = append(command.Env, "PWD="+absolute)
			}
		}
	}
	command.Dir = directory
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	return command
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
	// The raw capture is the upstream tests' own record, a file per process named by its pid, its lines in the
	// order parallel tests wrote them: never a product's bytes, which must be the same build to build (landable-15's
	// tree, Oct 11: lint-rules-agree-capture held capture-<pid>.jsonl). The rows and cases made from it are sorted.
	defer os.RemoveAll(capture)
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

// upstreamCaptureFiles is what go test reads in cohere for its rule packages: cohere's modules (its go.work's, the
// shim's and tsc's), the packages under cohere/internal and the tsc packages they link, each by name (fourslash and
// tsc's other unlinked packages are thousands of files).
// The tests read testdata only in their own packages, never tsc's baselines.
var upstreamCaptureFiles = []string{"cohere/go.mod", "cohere/go.sum", "cohere/go.work", "cohere/internal", "cohere/policy", "cohere/static_single_assignment", "cohere/mutation_aliasing", "cohere/TypeScript-shim", "cohere/TypeScript/tsc/go.mod", "cohere/TypeScript/tsc/go.sum",
	"cohere/TypeScript/tsc/internal/ast", "cohere/TypeScript/tsc/internal/astnav", "cohere/TypeScript/tsc/internal/binder", "cohere/TypeScript/tsc/internal/bundled", "cohere/TypeScript/tsc/internal/checker", "cohere/TypeScript/tsc/internal/collections", "cohere/TypeScript/tsc/internal/compiler", "cohere/TypeScript/tsc/internal/contentmapper", "cohere/TypeScript/tsc/internal/core", "cohere/TypeScript/tsc/internal/debug", "cohere/TypeScript/tsc/internal/diagnostics", "cohere/TypeScript/tsc/internal/evaluator", "cohere/TypeScript/tsc/internal/execute", "cohere/TypeScript/tsc/internal/glob", "cohere/TypeScript/tsc/internal/ipc", "cohere/TypeScript/tsc/internal/jsnum", "cohere/TypeScript/tsc/internal/json", "cohere/TypeScript/tsc/internal/jsonrpc", "cohere/TypeScript/tsc/internal/locale", "cohere/TypeScript/tsc/internal/module", "cohere/TypeScript/tsc/internal/modulespecifiers", "cohere/TypeScript/tsc/internal/nativepath", "cohere/TypeScript/tsc/internal/nodebuilder", "cohere/TypeScript/tsc/internal/osutil", "cohere/TypeScript/tsc/internal/outputpaths", "cohere/TypeScript/tsc/internal/packagejson", "cohere/TypeScript/tsc/internal/parser", "cohere/TypeScript/tsc/internal/printer", "cohere/TypeScript/tsc/internal/pseudochecker", "cohere/TypeScript/tsc/internal/scanner", "cohere/TypeScript/tsc/internal/semver", "cohere/TypeScript/tsc/internal/sourcemap", "cohere/TypeScript/tsc/internal/spanmap", "cohere/TypeScript/tsc/internal/stringutil", "cohere/TypeScript/tsc/internal/symlinks", "cohere/TypeScript/tsc/internal/tracing", "cohere/TypeScript/tsc/internal/transformers", "cohere/TypeScript/tsc/internal/tsoptions", "cohere/TypeScript/tsc/internal/tspath", "cohere/TypeScript/tsc/internal/vfs"}

// lintFixtureOracleFiles is what go build compiles in cohere for a Go oracle whose overlay lays sources, and files of a
// fixture copy outside the repository, over cohere. The copy keeps it from being a buildcache.GoBuildIn product, and
// the compiler's list (compilerinputs.Lowering) isn't this build's: it links cohere packages no compiler package does.
// So it is GoBuildIn's key by hand: every file of every repository package go list -deps reaches from the sources'
// imports, run where the build runs, with each such module's go.mod and go.sum and cohere's go.work. A fixture file
// laid over cohere imports nothing the sources don't; the sources themselves are the caller's to key.
func lintFixtureOracleFiles(t *testing.T, sources ...string) []string {
	t.Helper()
	imports := map[string]bool{}
	for _, source := range sources {
		file, err := parser.ParseFile(token.NewFileSet(), source, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatal(err)
		}
		for _, spec := range file.Imports {
			path, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				t.Fatal(err)
			}
			imports[path] = true
		}
	}
	packages := make([]string, 0, len(imports))
	for path := range imports {
		packages = append(packages, path)
	}
	sort.Strings(packages)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	if root, err = filepath.EvalSymlinks(root); err != nil {
		t.Fatal(err)
	}
	command := lintCommand(context.Background(), filepath.Join(root, "cohere"), "go", append([]string{"list", "-deps", "-json"}, packages...)...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	listing, err := command.Output()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, &stderr)
	}
	files := map[string]bool{"cohere/go.work": true}
	add := func(directory string, names ...string) {
		for _, name := range names {
			path, err := filepath.EvalSymlinks(filepath.Join(directory, name))
			if err != nil {
				t.Fatal(err)
			}
			if relative, err := filepath.Rel(root, path); err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				files[filepath.ToSlash(relative)] = true
			}
		}
	}
	decoder := json.NewDecoder(bytes.NewReader(listing))
	for decoder.More() {
		var listed struct {
			Dir                                                              string
			Standard                                                         bool
			Module                                                           *struct{ GoMod string }
			GoFiles, CgoFiles, CFiles, HFiles, SFiles, SysoFiles, EmbedFiles []string
		}
		if err := decoder.Decode(&listed); err != nil {
			t.Fatal(err)
		}
		if listed.Standard {
			continue
		}
		for _, names := range [][]string{listed.GoFiles, listed.CgoFiles, listed.CFiles, listed.HFiles, listed.SFiles, listed.SysoFiles, listed.EmbedFiles} {
			add(listed.Dir, names...)
		}
		if listed.Module != nil && listed.Module.GoMod != "" {
			for _, name := range []string{"go.mod", "go.sum"} {
				if _, err := os.Stat(filepath.Join(filepath.Dir(listed.Module.GoMod), name)); err == nil {
					add(filepath.Dir(listed.Module.GoMod), name)
				}
			}
		}
	}
	result := make([]string, 0, len(files))
	for file := range files {
		result = append(result, file)
	}
	sort.Strings(result)
	return result
}

// upstreamCaptureFiles is held to what the captures' go test links, as compilerinputs.Lowering is to the compiler's link
// set: every repository package go list -deps -test names for cohere's rule packages, and each module file pinning one,
// sits under an entry, on linux and darwin both, and no entry holds tsc's testdata. Mutant: cohere/policy dropped (it
// names the gap).
func TestUpstreamCaptureFilesCoverTheCapture(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	// cohere's go.work picks the modules the capture builds against, though go list names no package under it.
	needed := map[string]bool{"cohere/go.work": true}
	for _, system := range []string{"linux", "darwin"} {
		command := toolchain.CommandIn(filepath.Join(root, "cohere"), "go", "list", "-deps", "-test", "-f", "{{if not .Standard}}{{.Dir}}\t{{with .Module}}{{.GoMod}}{{end}}{{end}}", "./internal/lint/rules/...")
		command.Env = append(command.Env, "GOOS="+system, "CGO_ENABLED=1")
		var stderr bytes.Buffer
		command.Stderr = &stderr
		listing, err := command.Output()
		if err != nil {
			t.Fatalf("go list for %s: %v\n%s", system, err, stderr.Bytes())
		}
		for _, line := range strings.Split(strings.TrimSpace(string(listing)), "\n") {
			directory, goMod, _ := strings.Cut(line, "\t")
			for _, path := range []string{directory, goMod} {
				if relative, err := filepath.Rel(root, path); path != "" && err == nil && !strings.HasPrefix(relative, "..") {
					needed[filepath.ToSlash(relative)] = true
				}
			}
			if goMod != "" {
				if sum := filepath.Join(filepath.Dir(goMod), "go.sum"); func() bool { _, err := os.Stat(sum); return err == nil }() {
					if relative, err := filepath.Rel(root, sum); err == nil && !strings.HasPrefix(relative, "..") {
						needed[filepath.ToSlash(relative)] = true
					}
				}
			}
		}
	}
	var missing []string
	for _, path := range sortedKeys(needed) {
		if !coveredBy(upstreamCaptureFiles, path) {
			missing = append(missing, path)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("upstreamCaptureFiles misses what the captures' go test links: %s", strings.Join(missing, ", "))
	}
	for index, dropped := range upstreamCaptureFiles {
		rest := append(append([]string(nil), upstreamCaptureFiles[:index]...), upstreamCaptureFiles[index+1:]...)
		if !slices.ContainsFunc(sortedKeys(needed), func(path string) bool { return !coveredBy(rest, path) }) {
			t.Fatalf("dropping %s from upstreamCaptureFiles leaves no gap: the captures link nothing under it", dropped)
		}
	}
	for _, entry := range upstreamCaptureFiles {
		if baselines := "cohere/TypeScript/tsc/testdata"; baselines == entry || strings.HasPrefix(baselines, entry+"/") {
			t.Fatalf("upstreamCaptureFiles' %s holds %s, which no capture reads", entry, baselines)
		}
	}
	t.Logf("%d linked directories and module files, each under one of upstreamCaptureFiles' %d entries", len(needed), len(upstreamCaptureFiles))
}

// coveredBy reports whether one of entries is path or a directory holding it.
func coveredBy(entries []string, path string) bool {
	return slices.ContainsFunc(entries, func(entry string) bool { return path == entry || strings.HasPrefix(path, entry+"/") })
}

// sortedKeys is a set's members in order.
func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// lintRuleOracleSources are the rules' oracle.go files, which a fixture oracle lays over cohere beside its own sources,
// so lintFixtureOracleFiles keys the packages they import too.
func lintRuleOracleSources(t *testing.T) []string {
	t.Helper()
	sources, err := filepath.Glob("rules/*/oracle.go")
	if err != nil {
		t.Fatal(err)
	}
	return sources
}
