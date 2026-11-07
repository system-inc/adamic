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

// A child, this test binary run again with the planted file named, runs the actual status checker: its Fatal must
// reject the planted input, rather than a second implementation deciding whether a row looks plausible.
func TestOracleHostStatusMutants(test *testing.T) {
	test.Parallel()
	if path := os.Getenv(hostStatusProbeInput); path != "" {
		hostStatusFile(test, hostStatusMutantRow, false, path)
		return
	}
	real, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/host/status.json"))
	if err != nil {
		test.Fatal(err)
	}
	data, err := os.ReadFile(real)
	if err != nil {
		test.Fatal(err)
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(data, &rows); err != nil {
		test.Fatal(err)
	}
	if len(rows) != 25 {
		test.Fatalf("want 25 host fixtures, got %d", len(rows))
	}
	var first map[string]json.RawMessage
	if err := json.Unmarshal(rows[hostStatusMutantRow], &first); err != nil {
		test.Fatal(err)
	}
	// This getDirectories row changes in host-6. Node runs it, then lowering
	// refuses its overload contract. Change only the stage, not Node's observation.
	var recorded hostMeasurement
	if err := json.Unmarshal(rows[hostStatusMutantRow], &recorded); err != nil {
		test.Fatal(err)
	}
	if recorded.Stage.Outcome != "Refused" {
		test.Fatalf("wrong-stage witness changed: %+v", recorded.Stage)
	}
	binary, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	for _, mutation := range []string{"wrong-stage", "missing-row", "missing-file", "real-file"} {
		test.Run(mutation, func(test *testing.T) {
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
				if _, statErr := os.Stat(
					filepath.Join(repository, "stage3/fixtures/host", name),
				); !os.IsNotExist(statErr) {
					test.Fatalf("missing fixture exists or cannot be checked: %v", statErr)
				}
				fields["file"], err = json.Marshal(name)
				if err == nil {
					changed[hostStatusMutantRow], err = json.Marshal(fields)
				}
				want = "recorded Node " + name + ":"
			}
			if err != nil {
				test.Fatal(err)
			}
			if mutation != "real-file" {
				planted, marshalErr := json.Marshal(changed)
				if marshalErr != nil {
					test.Fatal(marshalErr)
				}
				input = filepath.Join(test.TempDir(), "status.json")
				if err := os.WriteFile(input, planted, 0600); err != nil {
					test.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			command := exec.CommandContext(
				ctx,
				binary,
				"-test.run=^TestOracleHostStatusMutants$",
				"-test.v",
				"-test.count=1",
				"-test.timeout=20s",
			)
			command.Env = append(os.Environ(), hostStatusProbeInput+"="+input)
			output, runErr := command.CombinedOutput()
			if ctx.Err() != nil {
				test.Fatalf("status probe exceeded 20s: %v\n%s", ctx.Err(), output)
			}
			if want == "" {
				if runErr != nil || !strings.Contains(string(output), "stage=Refused") {
					test.Fatalf("real file must pass the same check: %v\n%s", runErr, output)
				}
				test.Log("real status file passes after all planted failures")
				return
			}
			exit, ok := runErr.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), want) {
				test.Fatalf("want checker failure %q and exit 1: %v\n%s", want, runErr, output)
			}
			if mutation == "wrong-stage" && !strings.Contains(string(output), "stage=Refused") {
				test.Fatalf("wrong stage failed before observing its real outcome:\n%s", output)
			}
			test.Logf("%s caught by the real status checker: %s", mutation, want)
		})
	}
}
