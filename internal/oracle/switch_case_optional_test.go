package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"switch_case_optional_mapping.a", "switch_case_optional_template.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + name, true, false})
	}
}

func TestSwitchCaseOptionalSources(t *testing.T) {
	t.Parallel()
	for name, want := range map[string]string{
		"switch_case_optional_mapping.a":  "HELLO\nhello\nHELLo\nhELLo\n",
		"switch_case_optional_template.a": "MIXED,TEXT\nmixed,text\nMiXeD,tEXT\ntrue\nfalse\nmissing\n",
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
