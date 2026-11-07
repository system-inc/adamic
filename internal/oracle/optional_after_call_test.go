package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"79", "193", "variants", "required", "propagation"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/optional_after_call_" + name + ".a", true, false})
	}
	// Main now lowers this computed optional index; retain the original witness unchanged.
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/optional_after_call_gaps/computed.a", true, false})
}

// Each fixture has its own top-level test so the focused gate stays bounded.
func TestOptionalAfterCall79(test *testing.T)  { test.Parallel(); optionalAfterCall(test, "79") }
func TestOptionalAfterCall193(test *testing.T) { test.Parallel(); optionalAfterCall(test, "193") }
func TestOptionalAfterCallVariants(test *testing.T) {
	test.Parallel()
	optionalAfterCall(test, "variants")
}
func TestOptionalAfterCallRequired(test *testing.T) {
	test.Parallel()
	optionalAfterCall(test, "required")
}
func TestOptionalAfterCallPropagation(test *testing.T) {
	test.Parallel()
	optionalAfterCall(test, "propagation")
}

func optionalAfterCall(test *testing.T, name string) {
	test.Helper()
	fixture := "internal/oracle/testdata/optional_after_call_" + name + ".a"
	if name == "computed" {
		fixture = "internal/oracle/testdata/optional_after_call_gaps/computed.a"
	}
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		test.Fatal(err)
	}
	want := onNode(test, path)
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	actual, sanitized := natively(test, program)
	for backend, got := range map[string]run{
		"JavaScript":       onJavaScriptBackend(test, program),
		"sanitized native": actual,
		"release native":   released(test, program),
	} {
		if difference := disagreement(want, got); difference != "" {
			test.Errorf(
				"%s: %s; Node exit=%d stdout=%q stderr=%q; backend exit=%d stdout=%q stderr=%q",
				backend,
				difference,
				want.exitCode,
				want.stdout,
				want.stderr,
				got.exitCode,
				got.stdout,
				got.stderr,
			)
		}
	}
	if want.exitCode == 0 {
		if leaked := leaks(test, program, sanitized); leaked != "" {
			test.Errorf("leaks: %s", leaked)
		}
	}
}

func TestOptionalAfterCallComputed(test *testing.T) {
	test.Parallel()
	optionalAfterCall(test, "computed")
}
