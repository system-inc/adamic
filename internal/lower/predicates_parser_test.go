package lower

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"strings"
	"testing"
)

func TestParserCallbackArguments(t *testing.T) {
	source, err := os.ReadFile("testdata/predicates/generic_callback_parameter_read.a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lowerSource(t, string(source)); err != nil {
		t.Fatal(err)
	}
	// Checker inference inherits the overload claim. Only the independent
	// argument-body proof can keep this callback from reaching the callee.
	lying := strings.Replace(string(source), "function isDefined(value: string | undefined): value is string {\n    return value !== undefined;\n}", "function isDefined(value: string | undefined): value is string;\nfunction isDefined(value: string | undefined): boolean { return true; }", 1)
	lying = strings.Replace(lying, `"word", isDefined`, `"word", (value: string | undefined) => isDefined(value)`, 1)
	if _, err := lowerSource(t, lying); err == nil || !strings.Contains(err.Error(), "unproven predicate argument for parameter test") {
		t.Fatalf("want independently refused generic callback, got %v", err)
	}
}

func TestParserCallbackLie(t *testing.T) {
	source, err := os.ReadFile("testdata/predicates/callback_parameter_lie.a")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := lowerSource(t, string(source)); err == nil || !strings.Contains(err.Error(), "unproven predicate argument for parameter test") {
		t.Fatalf("want callback body proof refusal, got %v", err)
	}
}

func TestParserAssertionProofErasure(t *testing.T) {
	source, err := os.ReadFile("testdata/predicates/parser_assert_defined.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range []struct{ name, before, after string }{
		{"erase assertion body", "if (value === undefined || value === null) { fail(message); }", ""},
		{"erase undefined exclusion", "value === undefined || value === null", "value === null"},
		{"erase null exclusion", "value === undefined || value === null", "value === undefined"},
		{"returning helper", "function fail(message?: string): never { throw new Error(message); }", "function fail(message?: string): void { return; }"},
		{"annotation only helper", "function fail(message?: string): never { throw new Error(message); }", "function fail(message?: string): never; function fail(message?: string): void { return; }"},
		{"recursive helper", "throw new Error(message);", "fail(message);"},
		{"early normal return", "if (value === undefined", "return; if (value === undefined"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			if strings.Count(string(source), probe.before) != 1 {
				t.Fatal("mutation did not find its unique source")
			}
			mutant := strings.Replace(string(source), probe.before, probe.after, 1)
			if _, err := lowerSource(t, mutant); err == nil || !strings.Contains(err.Error(), "normal return has not narrowed value to NonNullable<T>") {
				t.Fatalf("want assertion normal-return proof refusal, got %v", err)
			}
		})
	}
}

// Separate subtests make a missing null check fail independently in each backend.
func TestParserAssertionBackends(t *testing.T) {
	source, err := os.ReadFile("testdata/predicates/assert_defined_read.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	for _, backend := range []string{"native", "javascript"} {
		t.Run(backend, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "program")
			var command *exec.Cmd
			if backend == "native" {
				if err := native.Build(native.C(program), path, native.Options{Sanitize: true}); err != nil {
					t.Fatal(err)
				}
				command = exec.Command(path)
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1")
			} else {
				path += ".mjs"
				if err := os.WriteFile(path, []byte(javascript.JavaScript(program)), 0644); err != nil {
					t.Fatal(err)
				}
				command = exec.Command("node", "--disable-warning=ExperimentalWarning", runner, path)
			}
			output, err := command.CombinedOutput()
			want := "WORD\nmissing\nword\nno match\n1\nno number\nevaluated\nfalse\nnull evaluated\nfalse\n"
			if err != nil || string(output) != want {
				t.Fatalf("%s output %q, error %v; want %q", backend, output, err, want)
			}
		})
	}
}

// The universal closure slot still needs adapters for scalar/boxed callable views.
func TestMixedUnionCallbackABIIsPending(t *testing.T) {
	t.Parallel()
	for _, fixture := range []string{"mixed_union_callback", "mixed_union_callback_variance"} {
		t.Run(fixture, func(t *testing.T) {
			source, err := os.ReadFile("testdata/predicates/" + fixture + ".a")
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSource(t, string(source))
			if err == nil || !strings.Contains(err.Error(), "function value") || !strings.Contains(err.Error(), "union of differently held members") {
				t.Fatalf("want the mixed union callable ABI refusal, got %v", err)
			}
		})
	}
}
