package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"switch_case_binder_372.a", "switch_case_binder_731.a", "switch_case_scope.a", "switch_case_read_tdz.a", "switch_case_write_tdz.a", "switch_case_capture_tdz.a", "switch_case_reentry_tdz.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

func TestSwitchCaseSourceReductions(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]string{
		"switch_case_binder_372.a": "exported\nnone\n",
		"switch_case_binder_731.a": "arg1\nother\n",
	} {
		path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", name))
		if err != nil {
			t.Fatal(err)
		}
		got := onNode(t, path)
		if got.exitCode != 0 || string(got.stdout) != want || len(got.stderr) != 0 {
			t.Fatalf("%s: %+v", name, got)
		}
	}
}
