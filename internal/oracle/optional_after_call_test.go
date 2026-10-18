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
func TestOptionalAfterCall79(t *testing.T)       { t.Parallel(); optionalAfterCall(t, "79") }
func TestOptionalAfterCall193(t *testing.T)      { t.Parallel(); optionalAfterCall(t, "193") }
func TestOptionalAfterCallVariants(t *testing.T) { t.Parallel(); optionalAfterCall(t, "variants") }
func TestOptionalAfterCallRequired(t *testing.T) { t.Parallel(); optionalAfterCall(t, "required") }
func TestOptionalAfterCallPropagation(t *testing.T) {
	t.Parallel()
	optionalAfterCall(t, "propagation")
}

func optionalAfterCall(t *testing.T, name string) {
	t.Helper()
	fixture := "internal/oracle/testdata/optional_after_call_" + name + ".a"
	if name == "computed" {
		fixture = "internal/oracle/testdata/optional_after_call_gaps/computed.a"
	}
	path, err := filepath.Abs(filepath.Join(repository, fixture))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	actual, sanitized := natively(t, program)
	for backend, got := range map[string]run{"JavaScript": onJavaScriptBackend(t, program), "sanitized native": actual, "release native": released(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s; Node exit=%d stdout=%q stderr=%q; backend exit=%d stdout=%q stderr=%q", backend, difference, want.exitCode, want.stdout, want.stderr, got.exitCode, got.stdout, got.stderr)
		}
	}
	if want.exitCode == 0 {
		if leaked := leaks(t, program, sanitized); leaked != "" {
			t.Errorf("leaks: %s", leaked)
		}
	}
}

func TestOptionalAfterCallComputed(t *testing.T) {
	t.Parallel()
	optionalAfterCall(t, "computed")
}
