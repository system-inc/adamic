package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/native_runtime"
)

// jsonObjectsRuntimeBinary builds a harness in testdata/json-objects-runtime/ against runtime/json_objects.c, its
// entry renamed so the archive's copy stays out, with one replacement made in the runtime's source for a mutant.
func jsonObjectsRuntimeBinary(t *testing.T, harnessName, old, replacement string) string {
	t.Helper()
	runtimeSource, err := os.ReadFile("../native_runtime/c/json_objects.c")
	if err != nil {
		t.Fatal(err)
	}
	source := string(runtimeSource)
	if old != "" {
		if !strings.Contains(source, old) {
			t.Fatal("json_objects.c mutant anchor missing")
		}
		source = strings.Replace(source, old, replacement, 1)
	}
	source = strings.ReplaceAll(source, "adamic_json_stringify_objects(", "probe_adamic_json_stringify_objects(")
	harness, err := os.ReadFile(filepath.Join("testdata", "json-objects-runtime", harnessName))
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "objects")
	if err := Build(source+"\n"+string(harness), binary, native_runtime.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return binary
}

// jsonObjectsRuntimeRun runs the harness, with an argument for the stop's case.
func jsonObjectsRuntimeRun(t *testing.T, binary string, arguments ...string) (string, string, int) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, arguments...)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatal(err)
	}
	return stdout.String(), stderr.String(), code
}

// jsonObjectsNullPointers is what Node prints for each byte a null pointer's write can leave that says which it
// was: null for null's, and the key left out for undefined's and for a string's, a union's, a weak link's and a
// record's, each boxed where it could be null, so its null pointer is undefined.
const jsonObjectsNullPointers = "12 {\"target\":null}\n13 {}\n3 {}\n10 {}\n11 {}\n14 {}\n"

const jsonObjectsNullPointerStop = "NotYet: JSON.stringify a reference field holding null or undefined that the allocation doesn't record which"

// A null pointer in an object slot reads by its byte, and one whose byte says only object stops.
// Nullable writes now leave a semantic tag; a cast can hide that contract from the writer. This
// harness pins the backstop for an allocation whose last write did not record its nullish meaning.
func TestJSONObjectsNullPointers(t *testing.T) {
	t.Parallel()
	binary := jsonObjectsRuntimeBinary(t, "harness.c", "", "")
	stdout, stderr, code := jsonObjectsRuntimeRun(t, binary)
	if code != 0 || stdout != jsonObjectsNullPointers {
		t.Fatalf("exit %d, stdout %q, want %q\n%s", code, stdout, jsonObjectsNullPointers, stderr)
	}
	stdout, stderr, code = jsonObjectsRuntimeRun(t, binary, "stop")
	if code != 70 || stdout != "" || !strings.Contains(stderr, jsonObjectsNullPointerStop) {
		t.Fatalf("want exit 70 with %q: exit %d, stdout %q, stderr %q", jsonObjectsNullPointerStop, code, stdout, stderr)
	}
}

// Mutants: the stop answered as undefined, so the Object byte's case prints {}; and a string's null pointer read
// as null, so its case prints null where Node leaves the key out.
func TestJSONObjectsNullPointerMutants(t *testing.T) {
	t.Parallel()
	stop := jsonObjectsRuntimeBinary(t, "harness.c", `stop("NotYet: JSON.stringify a reference field holding null or undefined that the allocation doesn't record which");`, "return &json_undefined;")
	if stdout, stderr, code := jsonObjectsRuntimeRun(t, stop, "stop"); code != 0 || stdout != "4 {}\n" {
		t.Fatalf("the stop mutant: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	text := jsonObjectsRuntimeBinary(t, "harness.c", "case adamic_ir_String: case adamic_ir_Union:", "case adamic_ir_String: return &json_null;\n\t\tcase adamic_ir_Union:")
	if stdout, stderr, code := jsonObjectsRuntimeRun(t, text); code != 0 || stdout == jsonObjectsNullPointers || !strings.Contains(stdout, "3 {\"target\":null}\n") {
		t.Fatalf("the string mutant: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}

const jsonObjectsLibraryValueStop = "NotYet: JSON.stringify a runtime library value (a Date, a Stats or another whose own fields aren't Node's)"

// JSON.stringify stops on a runtime library value by its shape's methods table, and a RegExp match's groups object,
// whose table is adamic_groups_methods, stops as a host's does (adamic_host_shape, #eksbz4c's head G). No program
// reaches this stop with groups yet: JSON.stringify of groups stops at compile time when its type is the groups
// type, and natively at "JSON union lacks proven container metadata" when it's held as unknown or object, before
// json_objects.c reads the shape. So testdata/json-objects-runtime/library.c drives the runtime directly, one object
// a run: a plain one prints, and a host's and a groups object stop. Mutant: the check left on
// &adamic_host_methods alone, where the groups object prints {"year":2024}.
func TestJSONObjectsLibraryValueStops(t *testing.T) {
	t.Parallel()
	binary := jsonObjectsRuntimeBinary(t, "library.c", "", "")
	if stdout, stderr, code := jsonObjectsRuntimeRun(t, binary); code != 0 || stdout != "{\"year\":2024}\n" {
		t.Fatalf("plain: exit %d, stdout %q, want {\"year\":2024}\n\n%s", code, stdout, stderr)
	}
	for _, kind := range []string{"host", "groups"} {
		if stdout, stderr, code := jsonObjectsRuntimeRun(t, binary, kind); code != 70 || stdout != "" || !strings.Contains(stderr, jsonObjectsLibraryValueStop) {
			t.Errorf("%s: want exit 70 with %q: exit %d, stdout %q, stderr %q", kind, jsonObjectsLibraryValueStop, code, stdout, stderr)
		}
	}
}
