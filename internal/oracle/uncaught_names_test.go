package oracle

import (
	"path/filepath"
	"testing"
)

func TestUncaughtEmptyName(t *testing.T) {
	t.Parallel()
	checkUncaughtName(t, "uncaught_names.a")
}

func TestUncaughtEmptyNameAndMessage(t *testing.T) {
	t.Parallel()
	checkUncaughtName(t, "uncaught_names_empty.a")
}

func TestUncaughtCustomEmptyMessage(t *testing.T) {
	t.Parallel()
	checkUncaughtName(t, "coverage_uncaught_custom_empty.a")
}

func TestUncaughtCustomNameAndMessage(t *testing.T) {
	t.Parallel()
	checkUncaughtName(t, "coverage_uncaught_custom_message.a")
}

func TestUncaughtDefaultEmptyMessage(t *testing.T) {
	t.Parallel()
	checkUncaughtName(t, "coverage_uncaught_default_empty.a")
}

func TestUncaughtEmptyNameChangedMessage(t *testing.T) {
	t.Parallel()
	checkUncaughtName(t, "coverage_uncaught_empty_changed.a")
}

func TestUncaughtNoArgument(t *testing.T) {
	t.Parallel()
	checkUncaughtName(t, "coverage_uncaught_no_argument.a")
}

func checkUncaughtName(t *testing.T, fixture string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode != 70 {
		t.Fatalf("Node did not throw: %+v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := natively(t, program)
	for _, result := range []run{got, released(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Errorf("%s: got %q, Node %q", difference, result.stderr, want.stderr)
		}
	}
}
