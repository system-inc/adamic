package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"p54_n", "p54_b", "p54_s", "misfit_number", "misfit_boolean"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{
			"internal/lower/testdata/scalar_union_views/" + name + ".ts",
			true,
			name == "misfit_number" || name == "misfit_boolean",
		})
	}
}

func scalarUnionViewMisfit(test *testing.T, name, declared, backend string) {
	test.Helper()
	path, err := filepath.Abs("../lower/testdata/scalar_union_views/" + name + ".ts")
	if err != nil {
		test.Fatal(err)
	}
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	source := onNode(test, path)
	if source.exitCode != 0 || string(source.stdout) != "wrong\n" {
		test.Fatalf("Node control: %+v", source)
	}
	expected := run{
		exitCode: 70,
		stderr: []byte("adamic: panic: field read failed: viewed.value matches no member of " +
			declared +
			"; expected " +
			declared +
			", found string\n"),
	}
	var got run
	switch backend {
	case "native":
		got = releasedUncached(test, program)
	case "sanitized":
		got, _ = nativelyUncached(test, program)
	case "javascript":
		got = onJavaScriptBackend(test, program)
	default:
		test.Fatalf("unknown backend %q", backend)
	}
	test.Logf("%s exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
	if difference := disagreement(expected, got); difference != "" {
		test.Errorf("%s: %s", backend, difference)
	}
}

func TestScalarUnionViewMisfitNumberNative(test *testing.T) {
	test.Parallel()
	scalarUnionViewMisfit(test, "misfit_number", "number | undefined", "native")
}

func TestScalarUnionViewMisfitNumberSanitized(test *testing.T) {
	test.Parallel()
	scalarUnionViewMisfit(test, "misfit_number", "number | undefined", "sanitized")
}

func TestScalarUnionViewMisfitNumberJavaScript(test *testing.T) {
	test.Parallel()
	scalarUnionViewMisfit(test, "misfit_number", "number | undefined", "javascript")
}

func TestScalarUnionViewMisfitBooleanNative(test *testing.T) {
	test.Parallel()
	scalarUnionViewMisfit(test, "misfit_boolean", "boolean | undefined", "native")
}

func TestScalarUnionViewMisfitBooleanSanitized(test *testing.T) {
	test.Parallel()
	scalarUnionViewMisfit(test, "misfit_boolean", "boolean | undefined", "sanitized")
}

func TestScalarUnionViewMisfitBooleanJavaScript(test *testing.T) {
	test.Parallel()
	scalarUnionViewMisfit(test, "misfit_boolean", "boolean | undefined", "javascript")
}
