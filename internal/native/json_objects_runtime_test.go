package native

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// jsonObjectsRuntimeBinary builds testdata/json-objects-runtime/harness.c against runtime/json_objects.c, its entry
// renamed so the archive's copy stays out, with one replacement made in the runtime's source for a mutant.
func jsonObjectsRuntimeBinary(t *testing.T, old, replacement string) string {
	t.Helper()
	runtimeSource, err := os.ReadFile("runtime/json_objects.c")
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
	harness, err := os.ReadFile("testdata/json-objects-runtime/harness.c")
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "objects")
	if err := Build(source+"\n"+string(harness), binary, Options{Sanitize: true}); err != nil {
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
	binary := jsonObjectsRuntimeBinary(t, "", "")
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
	stop := jsonObjectsRuntimeBinary(t, `stop("NotYet: JSON.stringify a reference field holding null or undefined that the allocation doesn't record which");`, "return &json_undefined;")
	if stdout, stderr, code := jsonObjectsRuntimeRun(t, stop, "stop"); code != 0 || stdout != "4 {}\n" {
		t.Fatalf("the stop mutant: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	text := jsonObjectsRuntimeBinary(t, "case adamic_ir_String: case adamic_ir_Union:", "case adamic_ir_String: return &json_null;\n\t\tcase adamic_ir_Union:")
	if stdout, stderr, code := jsonObjectsRuntimeRun(t, text); code != 0 || stdout == jsonObjectsNullPointers || !strings.Contains(stdout, "3 {\"target\":null}\n") {
		t.Fatalf("the string mutant: exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
}
