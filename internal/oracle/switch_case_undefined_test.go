package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/switch_case_undefined.a", false, false})
}

func TestSwitchCaseUndefinedSource(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/switch_case_undefined.a"))
	if err != nil {
		t.Fatal(err)
	}
	got := onNode(t, path)
	want := "equal\nequal\ngreater\nother\nmissing\nzero\nother\nother\nmissing\nfalse\ntrue\n"
	if got.exitCode != 0 || string(got.stdout) != want || len(got.stderr) != 0 {
		t.Fatalf("%+v", got)
	}
}
