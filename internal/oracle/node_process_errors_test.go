package oracle

import (
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Keep the requested real source test visible until the compiler implements unknown/in narrowing.
func TestNodeProcessErrorNarrowingBlocker(t *testing.T) {
	adapter, err := os.ReadFile(filepath.Join(repository, "stage3/adapt/47-host-errors/adapt.cjs"))
	if err != nil {
		t.Fatal(err)
	}
	helper := strings.SplitN(strings.SplitN(string(adapter), "const helper = `", 2)[1], "`;", 2)[0]
	fixture, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/node_process_errors.a"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(fixture), helper) {
		t.Fatal("process error fixture drifted from adaptation 47's exact helpers")
	}

	_, err = lowered(t, filepath.Join(repository, "internal/oracle/testdata/node_process_errors.a"))
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "in") {
		t.Fatalf("expected reported in-language blocker; got %v", err)
	}
	t.Log(err)
}

// Independent runtime proof; this deliberately does not count as source acceptance.
func TestNodeProcessErrorsRuntime(t *testing.T) {
	t.Parallel()
	shared := sharedDirectory(t)
	file, loop, denied := filepath.Join(shared, "file"), filepath.Join(shared, "loop"), filepath.Join(shared, "denied")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("loop", loop); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(denied, 0000); err != nil {
		t.Fatal(err)
	}
	paths := []string{filepath.Join(shared, "absent"), filepath.Join(file, "child"), loop, filepath.Join(shared, strings.Repeat("x", 256))}
	directory, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: directory}
	if os.Geteuid() == 0 {
		how.credential = &syscall.Credential{Uid: 65534, Gid: 65534}
	}
	paths = append(paths, denied)
	var environment []string
	codes := []string{"ENOENT", "ENOTDIR", "ELOOP", "ENAMETOOLONG", "EACCES", "ERR_OUT_OF_RANGE"}
	if runtime.GOOS == "linux" {
		shimSource := filepath.Join(shared, "directory-fault.c")
		shim := filepath.Join(shared, "directory-fault.so")
		text := `#define _GNU_SOURCE
#include <errno.h>
#include <dlfcn.h>
#include <string.h>
int chdir(const char *path) {
    if (strstr(path, "adamic-injected-ENOMEM") != NULL) { errno = ENOMEM; return -1; }
    if (strstr(path, "adamic-injected-EIO") != NULL) { errno = EIO; return -1; }
    int (*original)(const char *) = (int (*)(const char *))dlsym(RTLD_NEXT,"chdir");
    return original(path);
}
`
		if err := os.WriteFile(shimSource, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		built := execute(t, "clang", "-shared", "-fPIC", "-o", shim, shimSource, "-ldl")
		if built.exitCode != 0 {
			t.Fatalf("fault shim: %s", built.stderr)
		}
		environment = []string{"LD_PRELOAD=" + shim}
		paths = append(paths, filepath.Join(shared, "adamic-injected-ENOMEM"), filepath.Join(shared, "adamic-injected-EIO"))
		codes = append(codes, "ENOMEM", "EIO")
	}
	source := filepath.Join(directory, "internal/oracle/testdata/node_process_errors.a")
	how.arguments = paths
	runner := filepath.Join(directory, "oracle/node.mjs")
	truth := executeInput(t, how, environment, "node", append([]string{"--disable-warning=ExperimentalWarning", runner, source}, paths...)...)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("unexpected Node result %d %q", truth.exitCode, truth.stderr)
	}
	for _, code := range codes {
		if !strings.Contains(string(truth.stdout), code+"\n") {
			t.Fatalf("Node did not exercise %s: %s", code, truth.stdout)
		}
	}
	errorPaths := append(append([]string{}, paths...), "", "\x00", "absent\x00ignored", "absent\xed\xa0\x80")
	program := &ir.Program{Source: filepath.Base(source), Strings: append(append([]string{}, errorPaths...), "code", "message"), Locals: []ir.Local{{Name: "error", Type: ir.Object, Function: -1}}}
	caught := ir.Read{Local: 0, Of: ir.Object}
	catch := []ir.Statement{
		ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: caught, Name: "code", Of: ir.String}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: caught, Name: "message", Of: ir.String}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.BooleanToString{Value: ir.Binary{Operator: ir.And, Left: ir.HasOwn{Object: caught, Key: ir.StringConstant{Index: len(errorPaths)}}, Right: ir.HasOwn{Object: caught, Key: ir.StringConstant{Index: len(errorPaths) + 1}}}}},
	}
	attempt := func(operation string, argument ir.Expression) {
		program.Main = append(program.Main, ir.Try{HasCatch: true, CatchLocal: 0, Body: []ir.Statement{ir.Evaluate{Value: ir.ProcessCall{Operation: operation, Arguments: []ir.Expression{argument}, Of: ir.Object}}}, Catch: catch})
	}
	for index := range errorPaths {
		attempt("chdir", ir.StringConstant{Index: index})
	}
	for _, value := range []float64{math.NaN(), math.Inf(1), math.Inf(-1), 0.5} {
		code := ir.MaybeOf{Value: ir.NumberConstant{Value: value}, Of: ir.MaybeNumber}
		attempt("setExitCode", code)
		attempt("exit", code)
	}
	binary := filepath.Join(shared, "errors")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(shared, "errors.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0644); err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{executeInput(t, how, environment, binary, paths...), executeInput(t, how, environment, "node", "--disable-warning=ExperimentalWarning", runner, script)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s\nNode %q\ngot %q %q", difference, truth.stdout, got.stdout, got.stderr)
		}
	}
	if report := inputLeaks(t, how, program, binary); report != "" {
		t.Fatal(report)
	}
	data, err := os.ReadFile(filepath.Join(directory, "internal/native/runtime/process.c"))
	if err != nil {
		t.Fatal(err)
	}
	anchor := "ADAMIC_STRING(\"ERR_OUT_OF_RANGE\")"
	if strings.Count(string(data), anchor) != 1 {
		t.Fatal("exit error mutant anchor changed")
	}
	runtime := strings.Replace(string(data), anchor, "ADAMIC_STRING(\"ERR_INVALID_ARG_TYPE\")", 1)
	runtime = strings.ReplaceAll(runtime, "adamic_process_", "mutant_process_")
	runtime = strings.ReplaceAll(runtime, "mutant_process_exit_now", "adamic_process_exit_now")
	code := strings.ReplaceAll(native.C(program), "adamic_process_", "mutant_process_")
	mutant := filepath.Join(shared, "errors-mutant")
	if err := native.Build(runtime+"\n"+code, mutant, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	bad := executeInput(t, how, environment, mutant, paths...)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		t.Fatalf("exit error-code mutant not caught only by Node stdout: %d %q", bad.exitCode, bad.stderr)
	}
	t.Log("clean ERR_OUT_OF_RANGE mutant caught by Node stdout comparison")

	t.Log("own string code/message agree for real directory failures, Linux ENOMEM/EIO fault injection, and exit/exitCode numeric validation")
}

func TestNodeProcessCwdErrorRuntime(t *testing.T) {
	t.Parallel()
	directory, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(directory, "internal/oracle/testdata/node_process_cwd_error.a")
	driver := filepath.Join(directory, "internal/oracle/testdata/node_process_cached_cwd.py")
	observe := func(command ...string) run {
		result := execute(t, "python3", append([]string{driver}, command...)...)
		if result.exitCode != 0 {
			t.Fatalf("removed cwd driver: %s", result.stderr)
		}
		var output struct {
			Stdout, Stderr []byte
			ExitCode       int
		}
		if err := json.Unmarshal(result.stdout, &output); err != nil {
			t.Fatal(err)
		}
		return run{stdout: output.Stdout, stderr: output.Stderr, exitCode: output.ExitCode}
	}
	truth := observe("node", "--disable-warning=ExperimentalWarning", filepath.Join(directory, "oracle/node.mjs"), source)
	if truth.exitCode != 0 || string(truth.stdout) != "ready\nENOENT\nENOENT: process.cwd failed with error no such file or directory, the current working directory was likely removed without changing the working directory, uv_cwd\ntrue\n" || len(truth.stderr) != 0 {
		t.Fatalf("unexpected Node cwd error %d %q %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program := &ir.Program{Source: filepath.Base(source), Strings: []string{".", "ready\n", "/dev/stdin", "code", "message"}, Locals: []ir.Local{{Name: "error", Type: ir.Object, Function: -1}}}
	errorValue := ir.Read{Local: 0, Of: ir.Object}
	program.Main = []ir.Statement{
		ir.Evaluate{Value: ir.ProcessCall{Operation: "chdir", Of: ir.Object, Arguments: []ir.Expression{ir.StringConstant{Index: 0}}}},
		ir.Evaluate{Value: ir.ProcessCall{Operation: "stdoutWrite", Of: ir.Object, Arguments: []ir.Expression{ir.StringConstant{Index: 1}}}},
		ir.Evaluate{Value: ir.ReadTextFile{Path: ir.StringConstant{Index: 2}}},
		ir.Try{HasCatch: true, CatchLocal: 0, Body: []ir.Statement{ir.Evaluate{Value: ir.ProcessCall{Operation: "cwd", Of: ir.String}}}, Catch: []ir.Statement{
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: errorValue, Name: "code", Of: ir.String}},
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: errorValue, Name: "message", Of: ir.String}},
			ir.WriteLine{Stream: ir.Stdout, Value: ir.BooleanToString{Value: ir.Binary{Operator: ir.And, Left: ir.HasOwn{Object: errorValue, Key: ir.StringConstant{Index: 3}}, Right: ir.HasOwn{Object: errorValue, Key: ir.StringConstant{Index: 4}}}}},
		}},
	}
	binary := filepath.Join(t.TempDir(), "cwd-error")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(t.TempDir(), "cwd-error.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{observe(binary), observe("node", "--disable-warning=ExperimentalWarning", filepath.Join(directory, "oracle/node.mjs"), script)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s: %q %q", difference, got.stdout, got.stderr)
		}
	}
	runtime, err := os.ReadFile(filepath.Join(directory, "internal/native/runtime/node_process.c"))
	if err != nil {
		t.Fatal(err)
	}
	mutated := strings.Replace(string(runtime), "case ENOENT: code = \"ENOENT\";", "case ENOENT: code = \"ENOTDIR\";", 1)
	mutated = strings.ReplaceAll(mutated, "adamic_node_", "mutant_node_")
	code := strings.ReplaceAll(native.C(program), "adamic_node_", "mutant_node_")
	mutant := filepath.Join(t.TempDir(), "cwd-mutant")
	if err := native.Build(mutated+"\n"+code, mutant, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	bad := observe(mutant)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		t.Fatal("cwd error-code mutant not caught only by Node stdout")
	}
	t.Log("removed cwd own code/message match Node; clean code mutant caught by Node stdout")
}
