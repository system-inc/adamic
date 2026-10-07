package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestProcessExitPathOwnsLeakPolicy(t *testing.T) {
	t.Parallel()
	build := func(source string) (run, string) {
		path := filepath.Join(t.TempDir(), "exit.a")
		if err := os.WriteFile(path, []byte(source), 0600); err != nil {
			t.Fatal(err)
		}
		program, err := lowered(t, path)
		if err != nil {
			t.Fatal(err)
		}
		truth := onNode(t, path)
		actual, _ := nativelyUncached(t, program)
		if difference := disagreement(truth, actual); difference != "" {
			t.Fatal(difference)
		}
		if difference := disagreement(truth, onJavaScriptBackend(t, program)); difference != "" {
			t.Fatal(difference)
		}
		code := native.C(program)
		binary := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(code, binary, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		return execute(t, binary), code
	}
	exited, code := build("const end = (): never => process.exit(undefined); end();")
	if exited.exitCode != 0 || !abruptTermination(exited) || unbalanced(t, exited) != "" {
		t.Fatalf("explicit status 0 must remain an abrupt path: %+v", exited)
	}
	match := countsLine.FindSubmatch(exited.stderr)
	if countOf(t, match[1]) <= countOf(t, match[2])+countOf(t, match[6]) {
		t.Fatal("fixture no longer leaves the never-returning closure live")
	}
	// Mutate only the runtime's exit-path metadata. Output, status, counts and execution stay real.
	runtime, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/adamic.c"))
	if err != nil {
		t.Fatal(err)
	}
	anchor := `ADAMIC_COUNT_TERMINATE("process_exit");`
	if strings.Count(string(runtime), anchor) != 1 {
		t.Fatal("exit marker anchor changed")
	}
	changed := strings.Replace(string(runtime), anchor, "(void)0;", 1)
	for _, name := range []string{"adamic_start", "adamic_write_line", "adamic_write_raw", "adamic_output_flush", "adamic_panic", "adamic_unreachable", "adamic_process_exit_now"} {
		changed = strings.ReplaceAll(changed, name, "mutant_"+name)
	}
	// The normal process runtime calls the renamed exit implementation through a local wrapper.
	changed += "\n_Noreturn void mutant_exit(adamic_maybe_number code) { adamic_process_set_exit_code(code); mutant_adamic_process_exit_now(adamic_process_status()); }\n"
	code = strings.ReplaceAll(code, "adamic_process_exit(", "mutant_exit(")
	binary := filepath.Join(t.TempDir(), "exit-marker-mutant")
	if err := native.Build(changed+"\n"+code, binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	bad := execute(t, binary)
	if bad.exitCode != 0 || string(bad.stdout) != string(exited.stdout) || !strings.Contains(unbalanced(t, bad), "heap values leaked") {
		t.Fatal("metadata mutant must run cleanly and fail only the counted leak-policy check")
	}
	normal, normalCode := build("const end = (): never => process.exit(undefined); if (false) { end(); } console.log('normal');")
	if abruptTermination(normal) || unbalanced(t, normal) != "" {
		t.Fatal("unreached exit must not exempt normal completion")
	}
	release := "adamic_release(adamic_global_0_end);"
	if !strings.Contains(normalCode, release) {
		t.Fatal("normal closure-release anchor changed")
	}
	normalMutant := filepath.Join(t.TempDir(), "normal-release-mutant")
	if err := native.Build(strings.ReplaceAll(normalCode, release, "(void)adamic_global_0_end;"), normalMutant, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	leaked := execute(t, normalMutant)
	if leaked.exitCode != 0 || string(leaked.stdout) != string(normal.stdout) || abruptTermination(leaked) || !strings.Contains(unbalanced(t, leaked), "heap values leaked") {
		t.Fatal("normal-return release mutant must fail only counted leak detection")
	}
	nonzero, nonzeroCode := build("const end = (): never => process.exit(7); end();")
	if nonzero.exitCode != 7 || !abruptTermination(nonzero) || unbalanced(t, nonzero) != "" {
		t.Fatal("nonzero explicit exit must use the same policy")
	}
	if report, abrupt := leaksCountedWithPath(t, nonzeroCode); report != "" || !abrupt {
		t.Fatalf("exit path must bypass leak tools at any status: %s %t", report, abrupt)
	}
	inputPath := filepath.Join(t.TempDir(), "input-exit.a")
	if err := os.WriteFile(inputPath, []byte("const end = (): never => process.exit(0); end();"), 0600); err != nil {
		t.Fatal(err)
	}
	inputProgram, err := lowered(t, inputPath)
	if err != nil {
		t.Fatal(err)
	}
	if report := inputLeaksUncached(t, func() inputRun { return inputRun{directory: repository} }, inputProgram, "must-not-run-the-leak-tool"); report != "" {
		t.Fatal(report)
	}
	t.Log("status 0 exit abandons its live closure; normal status 0 releases it; both metadata and release mutants run cleanly and fail only counted checks")
}
