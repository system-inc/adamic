package fixtures

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Exercise the directory oracle itself against isolated status inputs, retaining
// the real tsc source, recorded Node result, and compiler invocation.
func TestHostStatusMutants(test *testing.T) {
	test.Parallel()
	data, err := os.ReadFile("host/status.json")
	if err != nil {
		test.Fatal(err)
	}
	var rows []fixture
	if err := json.Unmarshal(data, &rows); err != nil {
		test.Fatal(err)
	}
	if len(rows) != 25 {
		test.Fatalf("got %d host fixtures, want 25", len(rows))
	}
	executable, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	repository, err := filepath.Abs("../..")
	if err != nil {
		test.Fatal(err)
	}
	for _, index := range []int{0, 15, 17} {
		for _, mutation := range []string{"real", "wrong stage", "missing fixture"} {
			test.Run(rows[index].File+"/"+mutation, func(test *testing.T) {
				test.Parallel()
				root := test.TempDir()
				directory := filepath.Join(root, "host")
				if err := os.Mkdir(directory, 0700); err != nil {
					test.Fatal(err)
				}
				row := rows[index]
				source, err := os.ReadFile(filepath.Join("host", row.File))
				if err != nil {
					test.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(directory, row.File), source, 0600); err != nil {
					test.Fatal(err)
				}
				switch mutation {
				case "wrong stage":
					row.Stage0 = stage0{Outcome: "Checker", What: "planted incorrect stage"}
				case "missing fixture":
					row.File = "missing-host-status-mutant.a"
				}
				encoded, err := json.Marshal([]fixture{row})
				if err != nil {
					test.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(directory, "status.json"), encoded, 0600); err != nil {
					test.Fatal(err)
				}
				arguments := []string{"-test.run=^TestFixturesHost$", "-test.count=1", "-fixtures=" + root}
				arguments = append(arguments, hostStatusMutantTimeout(test)...)
				result := execute(test, filepath.Join(repository, "stage3/fixtures"), nil, executable, arguments...)
				if mutation == "real" {
					if result.Exit != 0 {
						test.Fatalf("real status failed: %s%s", result.Stdout, result.Stderr)
					}
					return
				}
				diagnostic := "gap changed"
				if mutation == "missing fixture" {
					diagnostic = "missing-host-status-mutant.a"
				}
				if result.Exit == 0 || !strings.Contains(result.Stdout+result.Stderr, diagnostic) {
					test.Fatalf(
						"directory oracle failed to catch %s: exit=%d %s%s",
						mutation,
						result.Exit,
						result.Stdout,
						result.Stderr,
					)
				}
				test.Logf("directory oracle caught %s: %s%s", mutation, result.Stdout, result.Stderr)
			})
		}
	}
}

// The child's clock must not be tighter than the unit's own under load.
// Leave a small margin for the parent to report the child's result.
func hostStatusMutantTimeout(test *testing.T) []string {
	if deadline, ok := test.Deadline(); ok {
		remaining := time.Until(deadline) - 5*time.Second
		if remaining < time.Second {
			remaining = time.Second
		}
		return []string{"-test.timeout=" + remaining.String()}
	}
	return nil
}
