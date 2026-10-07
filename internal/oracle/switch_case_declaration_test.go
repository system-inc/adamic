package oracle

import (
	"path/filepath"
	"testing"
)

func init() {
	for _, name := range []string{"scanner.a", "shared.a", "neighbors.a", "dead_zone.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/switch_case_declarations/" + name, true, false})
	}
}

func TestSwitchCaseDeadZoneStop(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/switch_case_declarations/dead_zone.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 70 || string(truth.stdout) != "value1\nbefore\n" || string(truth.stderr) != "adamic: panic: ReferenceError: Cannot access 'value' before initialization\n" {
		t.Fatalf("Node dead-zone stop: %+v", truth)
	}
}
