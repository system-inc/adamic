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
			"internal/lower/testdata/scalar_union_views/" + name + ".a", true, name == "misfit_number" || name == "misfit_boolean",
		})
	}
}

func scalarUnionViewMisfit(t *testing.T, name, declared string) {
	t.Helper()
	path, err := filepath.Abs("../lower/testdata/scalar_union_views/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if source.exitCode != 0 || string(source.stdout) != "wrong\n" {
		t.Fatalf("Node control: %+v", source)
	}
	expected := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: viewed.value matches no member of " + declared + "; expected " + declared + ", found string\n")}
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": func() run { got, _ := nativelyUncached(t, program); return got }(), "javascript": onJavaScriptBackend(t, program)} {
		t.Logf("%s exit=%d stdout=%q stderr=%q", backend, got.exitCode, got.stdout, got.stderr)
		if difference := disagreement(expected, got); difference != "" {
			t.Errorf("%s: %s", backend, difference)
		}
	}
}

func TestScalarUnionViewMisfitNumber(t *testing.T) {
	t.Parallel()
	scalarUnionViewMisfit(t, "misfit_number", "number | undefined")
}

func TestScalarUnionViewMisfitBoolean(t *testing.T) {
	t.Parallel()
	scalarUnionViewMisfit(t, "misfit_boolean", "boolean | undefined")
}
