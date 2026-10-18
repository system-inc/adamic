package oracle

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const hostStatusProbeInput = "ADAMIC_HOST_STATUS_MUTANT_INPUT"
const hostStatusMutantRow = 7

// A child runs the actual status checker: its Fatal must reject the planted input,
// rather than a second implementation deciding whether a row looks plausible.
func TestHostStatusMutantProbe(t *testing.T) {
	t.Parallel()
	path := os.Getenv(hostStatusProbeInput)
	if path == "" {
		t.Skip("selected by TestHostStatusMutants")
	}
	hostStatusFile(t, hostStatusMutantRow, false, path)
}

func TestHostStatusMutants(t *testing.T) {
	t.Parallel()
	real, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/host/status.json"))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(real)
	if err != nil {
		t.Fatal(err)
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 25 {
		t.Fatalf("want 25 host fixtures, got %d", len(rows))
	}
	var first map[string]json.RawMessage
	if err := json.Unmarshal(rows[hostStatusMutantRow], &first); err != nil {
		t.Fatal(err)
	}
	// This getDirectories row changes in host-6. Node runs it, then lowering
	// refuses its overload contract. Change only the stage, not Node's observation.
	var recorded hostMeasurement
	if err := json.Unmarshal(rows[hostStatusMutantRow], &recorded); err != nil {
		t.Fatal(err)
	}
	if recorded.Stage.Outcome != "Refused" {
		t.Fatalf("wrong-stage witness changed: %+v", recorded.Stage)
	}
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, mutation := range []string{"wrong-stage", "missing-row", "missing-file", "real-file"} {
		t.Run(mutation, func(t *testing.T) {
			input := real
			want := ""
			changed := append([]json.RawMessage(nil), rows...)
			fields := make(map[string]json.RawMessage, len(first))
			for key, value := range first {
				fields[key] = value
			}
			switch mutation {
			case "wrong-stage":
				fields["stage0"] = json.RawMessage(`{"outcome":"Compiles","what":""}`)
				changed[hostStatusMutantRow], err = json.Marshal(fields)
				want = "host status changed: got {Outcome:Refused"
			case "missing-row":
				changed = changed[1:]
				want = "want 25 original host fixtures, got 24"
			case "missing-file":
				name := "missing-host-status-mutant-fixture.a"
				if _, statErr := os.Stat(filepath.Join(repository, "stage3/fixtures/host", name)); !os.IsNotExist(statErr) {
					t.Fatalf("missing fixture exists or cannot be checked: %v", statErr)
				}
				fields["file"], err = json.Marshal(name)
				if err == nil {
					changed[hostStatusMutantRow], err = json.Marshal(fields)
				}
				want = "recorded Node " + name + ":"
			}
			if err != nil {
				t.Fatal(err)
			}
			if mutation != "real-file" {
				planted, marshalErr := json.Marshal(changed)
				if marshalErr != nil {
					t.Fatal(marshalErr)
				}
				input = filepath.Join(t.TempDir(), "status.json")
				if err := os.WriteFile(input, planted, 0600); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, binary, "-test.run=^TestHostStatusMutantProbe$", "-test.v", "-test.count=1", "-test.timeout=20s")
			command.Env = append(os.Environ(), hostStatusProbeInput+"="+input)
			output, runErr := command.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("status probe exceeded 20s: %v\n%s", ctx.Err(), output)
			}
			if want == "" {
				if runErr != nil || !strings.Contains(string(output), "current Node agrees; stage=Refused") {
					t.Fatalf("real file must pass the same check: %v\n%s", runErr, output)
				}
				t.Log("real status file passes after all planted failures")
				return
			}
			exit, ok := runErr.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), want) {
				t.Fatalf("want checker failure %q and exit 1: %v\n%s", want, runErr, output)
			}
			if mutation == "wrong-stage" && !strings.Contains(string(output), "current Node agrees; stage=Refused") {
				t.Fatalf("wrong stage failed before observing its real outcome:\n%s", output)
			}
			t.Logf("%s caught by the real status checker: %s", mutation, want)
		})
	}
}
