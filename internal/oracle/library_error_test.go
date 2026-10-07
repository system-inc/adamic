package oracle

import (
	"os"
	"strings"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_error_original_probe.a", true, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_error_probe.a", true, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_error_typeof.a", true, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_error_debug_fail.a", true, false}, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/library_error_stack_refused.a", false, false})
}

func TestLibraryErrorCounts(t *testing.T) {
	table, err := os.ReadFile("counts.md")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"original_probe", "probe", "typeof", "debug_fail"} {
		row := counted(t, "internal/oracle/testdata/library_error_"+name+".a", false, nil, false, false)
		t.Log(row)
		if !strings.Contains(string(table), row+"\n") {
			t.Errorf("Error slice count is not recorded: %s", row)
		}
	}
}
