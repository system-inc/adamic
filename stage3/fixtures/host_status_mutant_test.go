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
				arguments := []string{"-test.run=^TestFixturesHost$", "-test.count=1", "-fixtures=" + root}
				arguments = append(arguments, childTestTimeout(t)...)
				result := execute(t, filepath.Join(repository, "stage3/fixtures"), nil, executable, arguments...)
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

// childTestTimeout is a child test binary's -test.timeout, taken from the parent's own deadline less
// a margin to report the child's result, so the child dies only when the parent would. A fixed clock
// is a guess at the slowest box: TestHostConfigurationFixtureReaders' 30s child outran it on an
// 8-core box with the package running beside it (#q89xaef).
func childTestTimeout(t *testing.T) []string {
	if deadline, ok := t.Deadline(); ok {
		remaining := time.Until(deadline) - 5*time.Second
		if remaining < time.Second {
			remaining = time.Second
		}
		return []string{"-test.timeout=" + remaining.String()}
	}
	return nil
}

// A child's clock is the parent's, less the margin: a fixed one turns a loaded box's slow child red
// for the box's reason (#q89xaef). Under go test's default 10m deadline, 30s is far outside the band.
func TestChildTestTimeoutFollowsTheParentsDeadline(t *testing.T) {
	t.Parallel()
	deadline, ok := t.Deadline()
	if !ok {
		t.Skip("no -test.timeout deadline to follow")
	}
	arguments := childTestTimeout(t)
	if len(arguments) != 1 || !strings.HasPrefix(arguments[0], "-test.timeout=") {
		t.Fatalf("want one -test.timeout flag, got %q", arguments)
	}
	child, err := time.ParseDuration(strings.TrimPrefix(arguments[0], "-test.timeout="))
	if err != nil {
		t.Fatal(err)
	}
	want := time.Until(deadline) - 5*time.Second
	if want < time.Second {
		want = time.Second
	}
	if child > want+time.Second || child < want-10*time.Second {
		t.Fatalf("child timeout %v must follow the parent's deadline (want about %v)", child, want)
	}
}
