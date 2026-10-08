package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/finite_property_index.a", true, false})
}

func TestFinitePropertyInvalidKeyChecked(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/lower/testdata/finite_property_key_invalid.a"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	if want.exitCode == 0 || !strings.Contains(string(want.stderr), "TypeError") {
		t.Fatalf("Node must reject the missing field: %+v", want)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	native, _ := natively(t, program)
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program)} {
		if got.exitCode == 0 || !strings.Contains(string(got.stderr), "computed field key outside its proven union") {
			t.Errorf("%s must check the stale key: exit %d stdout %q stderr %q", name, got.exitCode, got.stdout, got.stderr)
		}
	}
}
