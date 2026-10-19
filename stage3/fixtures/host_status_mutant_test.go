package fixtures

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Exercise the directory oracle itself against isolated status inputs, retaining
// the real tsc source, recorded Node result, and compiler invocation.
func TestHostStatusMutants(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("host/status.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []fixture
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 25 {
		t.Fatalf("got %d host fixtures, want 25", len(rows))
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	for _, index := range []int{0, 15, 17} {
		for _, mutation := range []string{"real", "wrong stage", "missing fixture"} {
			t.Run(rows[index].File+"/"+mutation, func(t *testing.T) {
				t.Parallel()
				root := t.TempDir()
				directory := filepath.Join(root, "host")
				if err := os.Mkdir(directory, 0700); err != nil {
					t.Fatal(err)
				}
				row := rows[index]
				source, err := os.ReadFile(filepath.Join("host", row.File))
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(directory, row.File), source, 0600); err != nil {
					t.Fatal(err)
				}
				switch mutation {
				case "wrong stage":
					row.Stage0 = stage0{Outcome: "Checker", What: "planted incorrect stage"}
				case "missing fixture":
					row.File = "missing-host-status-mutant.a"
				}
				encoded, err := json.Marshal([]fixture{row})
				if err != nil {
					t.Fatal(err)
				}
				if err = os.WriteFile(filepath.Join(directory, "status.json"), encoded, 0600); err != nil {
					t.Fatal(err)
				}
				result := execute(t, filepath.Join(repository, "stage3/fixtures"), nil, executable,
					"-test.run=^TestFixturesHost$", "-test.count=1", "-test.timeout=25s", "-fixtures="+root)
				if mutation == "real" {
					if result.Exit != 0 {
						t.Fatalf("real status failed: %s%s", result.Stdout, result.Stderr)
					}
					return
				}
				diagnostic := "gap changed"
				if mutation == "missing fixture" {
					diagnostic = "missing-host-status-mutant.a"
				}
				if result.Exit == 0 || !strings.Contains(result.Stdout+result.Stderr, diagnostic) {
					t.Fatalf("directory oracle failed to catch %s: exit=%d %s%s", mutation, result.Exit, result.Stdout, result.Stderr)
				}
				t.Logf("directory oracle caught %s: %s%s", mutation, result.Stdout, result.Stderr)
			})
		}
	}
}
