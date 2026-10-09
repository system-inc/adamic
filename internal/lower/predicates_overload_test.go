package lower

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Node runs the original sources. Successful generated programs match it byte
// for byte; checked failures have an independent complete exit-70 contract.
func TestPredicateOverloadRuntime(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name, nodeOut, checkedOut, message string
		checked                            bool
	}{
		{"parser_every_result", "overload declarations loaded\n", "", "", false},
		{"parser_some_result", "overload declarations loaded\n", "", "", false},
		{"overload_some", "true:false\n", "", "", false},
		{"overload_some_empty", "false\n", "", "", false},
		{"overload_some_false_valid", "undefined\n", "", "", false},
		{"overload_some_false_read", "array\n", "", "overload 1 of some result: predicate array is false", true},
		{"overload_some_objects", "1\n", "", "", false},
		{"overload_some_true_only", "absent\n", "", "", false},
		{"overload_callback", "true\n", "", "", false},
		{"overload_erased", "true:false\n", "", "", false},
		{"overload_checked", "called\ntext\n", "called\n", "overload 1 of lie result: predicate value is false", true},
		{"overload_false", "called\n1\n", "called\n", "overload 1 of lie result: predicate value is false", true},
		{"overload_once", "argument\ncalled\ntrue\n", "", "", false},
		{"overload_every", "name\ntrue\n", "", "", false},
		{"overload_array_alias", "called\n1\n", "called\n", "overload 1 of corrupt result: predicate array is false", true},
		{"overload_assertion", "called\ntext\n", "called\n", "overload 1 of lie result: predicate value is false", true},
		{"overload_nominal", "called\ntrue\n", "", "", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			extension := ".a"
			if probe.name == "parser_every_result" || probe.name == "overload_every" {
				extension = ".ts"
			}
			path, err := filepath.Abs("testdata/predicates/" + probe.name + extension)
			if err != nil {
				t.Fatal(err)
			}
			run := func(command *exec.Cmd, stdout, stderr string, code int) {
				t.Helper()
				var output, errors bytes.Buffer
				command.Stdout = &output
				command.Stderr = &errors
				err := command.Run()
				got := 0
				if err != nil {
					if exit, ok := err.(*exec.ExitError); ok {
						got = exit.ExitCode()
					} else {
						t.Fatal(err)
					}
				}
				if got != code || output.String() != stdout || errors.String() != stderr {
					t.Fatalf("%s: exit %d, stdout %q, stderr %q; want exit %d, stdout %q, stderr %q", command.Path, got, output.String(), errors.String(), code, stdout, stderr)
				}
			}
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			lowerInput := lowerSource
			if extension == ".ts" {
				lowerInput = lowerTypeScriptAssertionSource
			}
			program, err := lowerInput(t, string(source))
			if err != nil {
				t.Fatal(err)
			}
			// Behavior cannot count proven, checked, or unobservable predicate sites.
			if counts, ok := map[string][3]int{
				"overload_some_empty":      {2, 0, 2},
				"overload_some_false_read": {1, 1, 1},
				"overload_some_true_only":  {1, 1, 1},
				"overload_checked":         {1, 1, 1},
				"overload_false":           {1, 1, 1},
				"overload_assertion":       {0, 1, 0},
			}[probe.name]; ok {
				got := program.PredicateChecks
				if got.Proven != counts[0] || got.Checked != counts[1] || got.Unobservable != counts[2] {
					t.Fatalf("predicate counts: %+v, want %v (proven, checked, unobservable)", got, counts)
				}
			}
			// Behavior cannot see the elimination of a redundant predicate check.
			if probe.name == "overload_erased" {
				for _, constant := range program.Strings {
					if strings.Contains(constant, "overload 1 of isNumber result:") {
						t.Fatal("body-proven predicate kept a result check")
					}
				}
			}
			stdout, stderr, code := probe.nodeOut, "", 0
			if probe.checked {
				stdout = probe.checkedOut
				stderr = "adamic: panic: " + probe.message + "\n"
				code = 70
			}
			binary := filepath.Join(t.TempDir(), "native")
			if err = native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(binary)
			run(command, stdout, stderr, code)
			var checkedFailure *nodeObservation
			if probe.checked {
				checkedFailure = &nodeObservation{[]byte(stdout), []byte(stderr), code}
			}
			compareAgreement(t, runAgreementNode(t, path), nodeObservation{[]byte(probe.nodeOut), nil, 0})
			agreeEntry(t, path, program, 0, checkedFailure, nil)
		})
	}
}

func TestIndirectPredicateOverloadIsPending(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `function lie(value: number | string): value is number; function lie(value: number | string): boolean { return true; } const alias=lie; alias("text");`)
	if err == nil || !strings.Contains(err.Error(), "indirect call of a checked predicate overload") {
		t.Fatalf("want explicit indirect-call capability gap, got %v", err)
	}
}
