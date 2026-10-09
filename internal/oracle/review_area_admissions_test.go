package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestReviewAreaAdmissions(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, stdout string }{
		{"fxspptb_57f2d04_with_frozen", "9|2\n3\n"},
		{"fxspptb_903f25b_try_assign", "caught\n1\n"},
		{"fxspptb_9984394_lib_dispatch", "1.50\ncaught\n"},
	} {
		name := test.name
		t.Run(name, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(reviewRoot, "agree", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if want.exitCode != 0 || string(want.stdout) != test.stdout || len(want.stderr) != 0 {
				t.Fatalf("unexpected Node truth: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
			}
			t.Logf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			actual, binary := natively(t, program)
			for backend, got := range map[string]run{"native": actual, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				t.Logf("%s: exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s", backend, difference)
				}
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if actual := onWASI(t, native.C(program)); disagreement(want, actual) != "" {
					t.Fatalf("wasm32: exit=%d stdout=%q stderr=%q", actual.exitCode, actual.stdout, actual.stderr)
				}
			}
			if report := leaks(t, program, binary); report != "" {
				t.Error(report)
			}
		})
	}
}

// A frozen assignment must throw TypeError, and virtual toFixed must throw RangeError.
func TestReviewAreaAdmissionErrorClasses(t *testing.T) {
	t.Parallel()
	for _, test := range []struct{ name, body, stdout string }{
		{"fxspptb_903f25b_try_assign", "if (error instanceof Error) { console.log(error.name); }", "TypeError\ncaught\n1\n"},
		{"fxspptb_9984394_lib_dispatch", "if (error instanceof Error) { return error.name; }", "1.50\nRangeError\n"},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join(reviewRoot, "agree", test.name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			source := strings.Replace(string(data), "} catch {", "} catch (error) { "+test.body, 1)
			if source == string(data) {
				t.Fatal("error class probe changed no catch")
			}
			path := filepath.Join(t.TempDir(), "class.a")
			if err := os.WriteFile(path, []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			want := onNode(t, path)
			if want.exitCode != 0 || len(want.stderr) != 0 || string(want.stdout) != test.stdout {
				t.Fatalf("Node: exit=%d stdout=%q stderr=%q", want.exitCode, want.stdout, want.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			actual, binary := natively(t, program)
			if difference := disagreement(want, actual); difference != "" {
				t.Fatalf("native: %s exit=%d stdout=%q stderr=%q", difference, actual.exitCode, actual.stdout, actual.stderr)
			}
			if report := leaks(t, program, binary); report != "" {
				t.Error(report)
			}
			if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
				if actual := onWASI(t, native.C(program)); disagreement(want, actual) != "" {
					t.Fatalf("wasm32: exit=%d stdout=%q stderr=%q", actual.exitCode, actual.stdout, actual.stderr)
				}
			}
			t.Logf("Node, native and wasm32 error class: %q", test.stdout)
		})
	}
}

// Keeping the object sealed but not frozen must be caught by Node, not a build failure.
func TestReviewAreaAdmissionsCatchFreezeMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(reviewRoot, "agree", "fxspptb_903f25b_try_assign.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	old := ", true, true, 0)"
	if strings.Count(code, old) != 1 {
		t.Fatal("freeze mutant must change exactly one integrity call")
	}
	code = strings.Replace(code, old, ", true, false, 0)", 1)
	for _, target := range []string{"native", "wasm32-wasi"} {
		t.Run(target, func(t *testing.T) {
			options := native.Options{Sanitize: true}
			if target == "wasm32-wasi" {
				if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
					t.Skip("set ADAMIC_ORACLE_WASI=1")
				}
				options = native.Options{Target: target}
			}
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, options); err != nil {
				t.Fatal(err)
			}
			var actual run
			if target == "native" {
				actual = executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			} else {
				actual = execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "wasi.mjs"), binary)
			}
			if actual.exitCode != 0 || len(actual.stderr) != 0 {
				t.Fatalf("mutant must run cleanly: exit=%d stderr=%q", actual.exitCode, actual.stderr)
			}
			if difference := disagreement(want, actual); difference != "stdout differs" {
				t.Fatalf("Node must catch stdout: %q", difference)
			}
			t.Logf("Node caught freeze mutant: Node=%q mutant=%q, both exit 0", want.stdout, actual.stdout)
		})
	}
}
