package oracle

import (
	"path/filepath"
	"testing"
)

func TestUncaughtEmptyName(test *testing.T) {
	test.Parallel()
	checkUncaughtName(test, "uncaught_names.a")
}

func TestUncaughtEmptyNameAndMessage(test *testing.T) {
	test.Parallel()
	checkUncaughtName(test, "uncaught_names_empty.a")
}

func TestUncaughtCustomEmptyMessage(test *testing.T) {
	test.Parallel()
	checkUncaughtName(test, "coverage_uncaught_custom_empty.a")
}

func TestUncaughtCustomNameAndMessage(test *testing.T) {
	test.Parallel()
	checkUncaughtName(test, "coverage_uncaught_custom_message.a")
}

func TestUncaughtDefaultEmptyMessage(test *testing.T) {
	test.Parallel()
	checkUncaughtName(test, "coverage_uncaught_default_empty.a")
}

func TestUncaughtEmptyNameChangedMessage(test *testing.T) {
	test.Parallel()
	checkUncaughtName(test, "coverage_uncaught_empty_changed.a")
}

func TestUncaughtNoArgument(test *testing.T) {
	test.Parallel()
	checkUncaughtName(test, "coverage_uncaught_no_argument.a")
}

func checkUncaughtName(test *testing.T, fixture string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := onNode(test, path)
	if want.exitCode != 70 {
		test.Fatalf("Node did not throw: %+v", want)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	got, _ := natively(test, program)
	for _, result := range []run{got, released(test, program), onJavaScriptBackend(test, program)} {
		if difference := disagreement(want, result); difference != "" {
			test.Errorf("%s: got %q, Node %q", difference, result.stderr, want.stderr)
		}
	}
}
