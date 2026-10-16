package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"enums_open.a", "enums_open_members.a", "enums_open_never.a", "enums_open_never_if.a", "enums_open_never_field.a", "enums_open_never_update.a", "enums_open_never_implicit.a", "enums_open_never_return.a", "enums_open_never_index.a"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"internal/oracle/testdata/" + name, true, name != "enums_open.a" && name != "enums_open_members.a",
		})
	}
}

// Each path a numeric enum's never value reaches, pinned to its exact output on both backends: a test of its own, since
// the six as subtests of one were 25.4 s of its own on Cloud, each a check, a lowering and a sanitized build.
func TestNumericEnumNeverPathsPinned_if(t *testing.T) {
	t.Parallel()
	numericEnumNeverPathPinned(t, "if")
}

func TestNumericEnumNeverPathsPinned_field(t *testing.T) {
	t.Parallel()
	numericEnumNeverPathPinned(t, "field")
}

func TestNumericEnumNeverPathsPinned_update(t *testing.T) {
	t.Parallel()
	numericEnumNeverPathPinned(t, "update")
}

func TestNumericEnumNeverPathsPinned_implicit(t *testing.T) {
	t.Parallel()
	numericEnumNeverPathPinned(t, "implicit")
}

func TestNumericEnumNeverPathsPinned_return(t *testing.T) {
	t.Parallel()
	numericEnumNeverPathPinned(t, "return")
}

func TestNumericEnumNeverPathsPinned_index(t *testing.T) {
	t.Parallel()
	numericEnumNeverPathPinned(t, "index")
}

// numericEnumNeverPathPinned holds enums_open_never_<name>.a to its output on the native and JavaScript backends.
func numericEnumNeverPathPinned(t *testing.T, name string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enums_open_never_"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	stdout := "before\n"
	if name == "implicit" {
		stdout += "lookup\n"
	}
	want := run{stdout: []byte(stdout), stderr: []byte("adamic: panic: unreachable value 42 for numeric enum SyntaxKind\n"), exitCode: 70}
	native, _ := natively(t, program)
	for backend, result := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Errorf("%s: %s; %+v", backend, difference, result)
		}
	}
}

func TestNumericEnumNeverPinned(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enums_open_never.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: unreachable value 42 for numeric enum SyntaxKind\n"), exitCode: 70}
	native, _ := natively(t, program)
	for name, result := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Errorf("%s: %s; exit %d, stdout %q, stderr %q", name, difference, result.exitCode, result.stdout, result.stderr)
		}
	}
	if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != "before\nunreachable\nafter\n" {
		t.Fatalf("independent Node must run on: %+v", node)
	}
}
