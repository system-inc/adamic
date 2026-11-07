package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{
		"enums_open.a",
		"enums_open_members.a",
		"enums_open_never.a",
		"enums_open_never_if.a",
		"enums_open_never_field.a",
		"enums_open_never_update.a",
		"enums_open_never_implicit.a",
		"enums_open_never_return.a",
		"enums_open_never_index.a",
	} {
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
func TestNumericEnumNeverPathsPinned_if(test *testing.T) {
	test.Parallel()
	numericEnumNeverPathPinned(test, "if")
}

func TestNumericEnumNeverPathsPinned_field(test *testing.T) {
	test.Parallel()
	numericEnumNeverPathPinned(test, "field")
}

func TestNumericEnumNeverPathsPinned_update(test *testing.T) {
	test.Parallel()
	numericEnumNeverPathPinned(test, "update")
}

func TestNumericEnumNeverPathsPinned_implicit(test *testing.T) {
	test.Parallel()
	numericEnumNeverPathPinned(test, "implicit")
}

func TestNumericEnumNeverPathsPinned_return(test *testing.T) {
	test.Parallel()
	numericEnumNeverPathPinned(test, "return")
}

func TestNumericEnumNeverPathsPinned_index(test *testing.T) {
	test.Parallel()
	numericEnumNeverPathPinned(test, "index")
}

// numericEnumNeverPathPinned holds enums_open_never_<name>.a to its output on the native and JavaScript backends.
func numericEnumNeverPathPinned(test *testing.T, name string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/enums_open_never_"+name+".a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	stdout := "before\n"
	if name == "implicit" {
		stdout += "lookup\n"
	}
	want := run{
		stdout:   []byte(stdout),
		stderr:   []byte("adamic: panic: unreachable value 42 for numeric enum SyntaxKind\n"),
		exitCode: 70,
	}
	native, _ := natively(test, program)
	for backend, result := range map[string]run{"native": native, "javascript": onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, result); difference != "" {
			test.Errorf("%s: %s; %+v", backend, difference, result)
		}
	}
}

func TestNumericEnumNeverPinned(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/enums_open_never.a"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	want := run{
		stdout:   []byte("before\n"),
		stderr:   []byte("adamic: panic: unreachable value 42 for numeric enum SyntaxKind\n"),
		exitCode: 70,
	}
	native, _ := natively(test, program)
	for name, result := range map[string]run{"native": native, "javascript": onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, result); difference != "" {
			test.Errorf(
				"%s: %s; exit %d, stdout %q, stderr %q",
				name,
				difference,
				result.exitCode,
				result.stdout,
				result.stderr,
			)
		}
	}
	if node := onNode(test, path); node.exitCode != 0 || string(node.stdout) != "before\nunreachable\nafter\n" {
		test.Fatalf("independent Node must run on: %+v", node)
	}
}
