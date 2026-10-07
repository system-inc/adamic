package oracle

import (
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Unknown narrowing (616870d) removed the in blocker. Keep the unchanged source's
// next refusal recorded exactly: reading hasOwnProperty as an unbound method.
func TestNodeProcessErrorNarrowingBlocker(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_process_errors.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	want := path + ":9:32: Adamic 0.1 refuses a method read as a value (hasOwnProperty would lose its object, and this with it); call it in an arrow that keeps the object: (v) => its object.hasOwnProperty(v) (unbound-method)"
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.Error() != want {
		t.Fatalf("recorded blocker changed:\nwant %s\ngot %v", want, err)
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
	source := filepath.Join(directory, "internal/oracle/testdata/node_process_errors.a")
	how.arguments = paths
	truth := onNodeWith(t, how, source)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("unexpected Node result %d %q", truth.exitCode, truth.stderr)
	}
	for _, code := range []string{"ENOENT", "ENOTDIR", "ELOOP", "ENAMETOOLONG", "EACCES", "ERR_OUT_OF_RANGE"} {
		if !strings.Contains(string(truth.stdout), code+"\n") {
			t.Fatalf("Node did not exercise %s: %s", code, truth.stdout)
		}
	}
	program := &ir.Program{Source: filepath.Base(source), Strings: append(append([]string{}, paths...), "code", "message"), Locals: []ir.Local{{Name: "error", Type: ir.Object, Function: -1}}}
	caught := ir.Read{Local: 0, Of: ir.Object}
	catch := []ir.Statement{
		ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: caught, Name: "code", Of: ir.String}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: caught, Name: "message", Of: ir.String}},
		ir.WriteLine{Stream: ir.Stdout, Value: ir.BooleanToString{Value: ir.Binary{Operator: ir.And, Left: ir.HasOwn{Object: caught, Key: ir.StringConstant{Index: len(paths)}}, Right: ir.HasOwn{Object: caught, Key: ir.StringConstant{Index: len(paths) + 1}}}}},
	}
	attempt := func(operation string, argument ir.Expression) {
		program.Main = append(program.Main, ir.Try{HasCatch: true, CatchLocal: 0, Body: []ir.Statement{ir.Evaluate{Value: ir.ProcessCall{Operation: operation, Arguments: []ir.Expression{argument}, Of: ir.Object}}}, Catch: catch})
	}
	for index := range paths {
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
	for _, got := range []run{executeInput(t, how, nil, binary, paths...), onNodeWith(t, how, script)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s\nNode %q\ngot %q %q", difference, truth.stdout, got.stdout, got.stderr)
		}
	}
	if report := inputLeaks(t, func() inputRun { return how }, program, binary); report != "" {
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
	bad := executeInput(t, how, nil, mutant, paths...)
	if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
		t.Fatalf("exit error-code mutant not caught only by Node stdout: %d %q", bad.exitCode, bad.stderr)
	}
	t.Log("clean ERR_OUT_OF_RANGE mutant caught by Node stdout comparison")

	t.Log("own string code/message agree for five directory errno and exit/exitCode NaN, infinities and fractional validation")
}
