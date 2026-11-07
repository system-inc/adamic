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
	if decodeError := json.Unmarshal(data, &rows); decodeError != nil {
		test.Fatal(decodeError)
	}
	if len(rows) != 25 {
		test.Fatalf("want 25 host fixtures, got %d", len(rows))
	}
	var first map[string]json.RawMessage
	if decodeError := json.Unmarshal(rows[hostStatusMutantRow], &first); decodeError != nil {
		test.Fatal(decodeError)
	}
	// This getDirectories row changes in host-6. Node runs it, then lowering
	// refuses its overload contract. Change only the stage, not Node's observation.
	var recorded hostMeasurement
	if decodeError := json.Unmarshal(rows[hostStatusMutantRow], &recorded); decodeError != nil {
		test.Fatal(decodeError)
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
				if _, statError := os.Stat(
					filepath.Join(repository, "stage3/fixtures/host", name),
				); !os.IsNotExist(statError) {
					test.Fatalf("missing fixture exists or cannot be checked: %v", statError)
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
				planted, marshalError := json.Marshal(changed)
				if marshalError != nil {
					test.Fatal(marshalError)
				}
				input = filepath.Join(test.TempDir(), "status.json")
				if writeError := os.WriteFile(input, planted, 0600); writeError != nil {
					test.Fatal(writeError)
				}
			}
			timeoutContext, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			command := exec.CommandContext(
				timeoutContext,
				binary,
				"-test.run=^TestOracleHostStatusMutants$",
				"-test.v",
				"-test.count=1",
				"-test.timeout=20s",
			)
			command.Env = append(os.Environ(), hostStatusProbeInput+"="+input)
			output, runError := command.CombinedOutput()
			if timeoutContext.Err() != nil {
				test.Fatalf("status probe exceeded 20s: %v\n%s", timeoutContext.Err(), output)
			}
			if want == "" {
				if runError != nil || !strings.Contains(string(output), "stage=Refused") {
					test.Fatalf("real file must pass the same check: %v\n%s", runError, output)
				}
				test.Log("real status file passes after all planted failures")
				return
			}
			exit, ok := runError.(*exec.ExitError)
			if !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), want) {
				test.Fatalf("want checker failure %q and exit 1: %v\n%s", want, runError, output)
			}
			if mutation == "wrong-stage" && !strings.Contains(string(output), "stage=Refused") {
				test.Fatalf("wrong stage failed before observing its real outcome:\n%s", output)
			}
			test.Logf("%s caught by the real status checker: %s", mutation, want)
		})
	}
}
