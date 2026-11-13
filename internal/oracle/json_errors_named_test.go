package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// Like the runtime-stop fixtures in library_intl_test.go, this stays outside the
// ordinary fixture and counts tables. Both backends print untouched errors, then stop
// at the name write, before Node prints its newly enumerable name.
const jsonNamedErrorsFixture = "internal/oracle/testdata/json_errors_named/named.a"

// Restoring Object.assign(new Error(message), {name}) reds type_error,
// range_error and library_type_error. The library row reaches toPrimitiveTypeError.
func TestJSONNamedErrorsJavaScript(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, jsonNamedErrorsFixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	backend := onJavaScriptBackend(t, program)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("source exit %d stderr %q; backend exit %d stderr %q", truth.exitCode, truth.stderr, backend.exitCode, backend.stderr)
	}
	want := []string{`{"marker":1}`, "{}", "{}", "{}", "{}", `{"name":"Changed"}`}
	names := []string{"object", "type_error", "range_error", "error", "library_type_error", "renamed_type_error"}
	sourceLines := strings.Split(strings.TrimSuffix(string(truth.stdout), "\n"), "\n")
	backendLines := strings.Split(strings.TrimSuffix(string(backend.stdout), "\n"), "\n")
	if len(sourceLines) != len(want) || len(backendLines) != len(want)-1 {
		t.Fatalf("want %d rows; source stdout %q, backend stdout %q", len(want), truth.stdout, backend.stdout)
	}
	for i, name := range names {
		t.Run(name, func(t *testing.T) {
			if sourceLines[i] != want[i] {
				t.Fatalf("Node printed %q, want %q", sourceLines[i], want[i])
			}
			if i < len(backendLines) && backendLines[i] != sourceLines[i] {
				t.Fatalf("Node printed %q, backend printed %q", sourceLines[i], backendLines[i])
			}
		})
	}
	stop := run{stdout: []byte(strings.Join(want[:len(want)-1], "\n") + "\n"), stderr: []byte(jsonErrorWritten), exitCode: 70}
	if difference := disagreement(stop, backend); difference != "" {
		t.Errorf("%s: source stdout %q, backend stdout %q", difference, truth.stdout, backend.stdout)
	}
}

func TestJSONNamedErrorsNativeStop(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, jsonNamedErrorsFixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("{\"marker\":1}\n{}\n{}\n{}\n{}\n"), stderr: []byte(jsonErrorWritten), exitCode: 70}
	for _, backend := range []string{"sanitized", "release"} {
		t.Run(backend, func(t *testing.T) {
			var got run
			if backend == "sanitized" {
				got, _ = natively(t, program)
			} else {
				got = released(t, program)
			}
			if difference := disagreement(want, got); difference != "" {
				t.Fatalf("%s: exit %d stdout %q stderr %q", difference, got.exitCode, got.stdout, got.stderr)
			}
		})
	}
}
