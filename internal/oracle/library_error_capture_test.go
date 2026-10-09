package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, path := range []string{"internal/oracle/testdata/library_error_capture_stack.a", "internal/oracle/testdata/library_error_stack_limit.a", "internal/oracle/testdata/library_error_capture_own.a", "internal/oracle/testdata/library_error_captured_error.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestErrorCaptureMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct{ name, fixture string }{
		{"capture does not set stack", "library_error_capture_own.a"},
		{"stack limit write ignored", "library_error_stack_limit.a"},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", mutant.fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			changed := false
			for index := range program.Functions {
				function := &program.Functions[index]
				if mutant.fixture == "library_error_capture_own.a" && function.Name == "library_Error_captureStackTrace" {
					function.Body = nil
					changed = true
				}
				if mutant.fixture == "library_error_stack_limit.a" && function.Name == "Error_setStackTraceLimit" {
					body := []ir.Statement{}
					for _, statement := range function.Body {
						if _, write := statement.(ir.Assign); write {
							changed = true
							continue
						}
						body = append(body, statement)
					}
					function.Body = body
				}
			}
			if !changed {
				t.Fatal("mutant changed no implementation")
			}
			expected := onNode(t, path)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant must run cleanly: exit %d stderr %s", actual.exitCode, actual.stderr)
			}
			if difference := disagreement(expected, actual); difference != "stdout differs" {
				t.Fatalf("Node did not catch native mutant: %s", difference)
			}
			javascript := onJavaScriptBackend(t, program)
			if javascript.exitCode != 0 || len(javascript.stderr) != 0 || disagreement(expected, javascript) != "stdout differs" {
				t.Fatalf("Node did not catch clean JavaScript mutant: %+v", javascript)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				wasi := onWASI(t, native.C(program))
				if wasi.exitCode != 0 || len(wasi.stderr) != 0 || disagreement(expected, wasi) != "stdout differs" {
					t.Fatalf("Node did not catch clean WASI mutant: %+v", wasi)
				}
			}
			if mutant.fixture == "library_error_capture_own.a" && strings.TrimSpace(string(actual.stdout)) != "false" {
				t.Fatalf("want failed own-property check, got %q", actual.stdout)
			}
			t.Log("clean exit 0, no sanitizer finding; caught only by Node stdout comparison")
		})
	}
}
