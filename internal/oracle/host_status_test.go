package oracle

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

var updateHostStatus = flag.Bool("update-host-status", false, "remeasure host stage0 after recorded/current Node and native/JavaScript agreement")
var hostStatusMutex sync.Mutex

type hostStage struct {
	Outcome string `json:"outcome"`
	What    string `json:"what"`
}
type hostMeasurement struct {
	File string `json:"file"`
	Node struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		Exit   int    `json:"exit"`
	} `json:"node"`
	Stage hostStage `json:"stage0"`
}

// Each unchanged upstream program is its own unit. Status is an input, never
// success evidence: current Node must agree before any recorded stage is trusted.
func hostStatus(t *testing.T, index int, wasiOnly bool) {
	t.Helper()
	hostStatusFile(t, index, wasiOnly, filepath.Join(repository, "stage3/fixtures/host/status.json"))
}

// The same check accepts a status input for isolated mutant probes.
func hostStatusFile(t *testing.T, index int, wasiOnly bool, statusPath string) {
	t.Helper()
	t.Parallel()
	if wasiOnly && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	hostStatusMutex.Lock()
	data, err := os.ReadFile(statusPath)
	hostStatusMutex.Unlock()
	if err != nil {
		t.Fatal(err)
	}
	var rows []json.RawMessage
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	if len(rows) != 25 {
		t.Fatalf("want 25 original host fixtures, got %d", len(rows))
	}
	var row hostMeasurement
	if err := json.Unmarshal(rows[index], &row); err != nil {
		t.Fatal(err)
	}
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/host", row.File))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNodeWith(t, inputRun{directory: repository}, path)
	recorded := run{stdout: []byte(row.Node.Stdout), stderr: []byte(row.Node.Stderr), exitCode: row.Node.Exit}
	if difference := disagreement(recorded, truth); difference != "" {
		t.Fatalf("recorded Node %s: %s", row.File, difference)
	}
	program, err := load.Load([]string{path})
	observed := hostStage{Outcome: "Compiles"}
	var code, script string
	if err == nil {
		loweredProgram, lowerErr := lower.Lower(context.Background(), program)
		if lowerErr == nil {
			code = native.C(loweredProgram)
			script = javascript.JavaScript(loweredProgram)
		}
		err = lowerErr
	}
	if err != nil {
		var checker *load.CheckError
		var refused *lower.Refused
		var notYet *lower.NotYet
		switch {
		case errors.As(err, &checker):
			observed.Outcome = "Checker"
		case errors.As(err, &refused):
			observed.Outcome = "Refused"
		case errors.As(err, &notYet):
			observed.Outcome = "NotYet"
		default:
			t.Fatal(err)
		}
		root, rootErr := filepath.Abs(repository)
		if rootErr != nil {
			t.Fatal(rootErr)
		}
		observed.What = strings.ReplaceAll(err.Error(), root+string(filepath.Separator), "")
	} else {
		directory := t.TempDir()
		if !wasiOnly {
			binary := filepath.Join(directory, "program")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeInput(t, inputRun{directory: repository}, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
			if difference := disagreement(truth, actual); difference != "" {
				t.Fatalf("%s native: %s", row.File, difference)
			}
			js := filepath.Join(directory, "program.mjs")
			if err := os.WriteFile(js, []byte(script), 0600); err != nil {
				t.Fatal(err)
			}
			actual = onNodeWith(t, inputRun{directory: repository}, js)
			if difference := disagreement(truth, actual); difference != "" {
				t.Fatalf("%s JavaScript: %s", row.File, difference)
			}
		}
		if wasiOnly {
			wasm := filepath.Join(directory, "program.wasm")
			if err := native.Build(code, wasm, native.Options{Target: "wasm32-wasi"}); err != nil {
				var refused *native.TargetRefused
				if !errors.As(err, &refused) {
					t.Fatal(err)
				}
				t.Logf("%s WASI refusal: %v", row.File, refused)
			} else {
				runner, runnerErr := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
				if runnerErr != nil {
					t.Fatal(runnerErr)
				}
				actual := executeInput(t, inputRun{directory: repository}, nil, "node", "--disable-warning=ExperimentalWarning", runner, wasm)
				if reason := wasiRefusalReason(actual); reason != "" {
					t.Logf("%s WASI refusal: %s", row.File, reason)
				} else if difference := disagreement(truth, actual); difference != "" {
					t.Fatalf("%s WASI: %s", row.File, difference)
				}
			}
		}
	}
	t.Logf("%s: current Node agrees; stage=%s diagnostic=%q", row.File, observed.Outcome, observed.What)
	if observed == row.Stage {
		return
	}
	if !*updateHostStatus || wasiOnly {
		t.Fatalf("host status changed: got %+v; recorded %+v", observed, row.Stage)
	}
	// Replace only this stage0 field. Keep original fixture provenance and Node bytes.
	hostStatusMutex.Lock()
	defer hostStatusMutex.Unlock()
	data, err = os.ReadFile(statusPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(rows[index], &fields); err != nil {
		t.Fatal(err)
	}
	replacement, err := json.Marshal(observed)
	if err != nil {
		t.Fatal(err)
	}
	raw := rows[index]
	marker := bytes.Index(raw, []byte(`"stage0":`))
	if marker < 0 {
		t.Fatal("missing stage0 field")
	}
	offset := bytes.Index(raw[marker:], fields["stage0"])
	if offset < 0 {
		t.Fatal("cannot locate stage0 field")
	}
	position := marker + offset
	changed := append(append(append([]byte{}, raw[:position]...), replacement...), raw[position+len(fields["stage0"]):]...)
	data = bytes.Replace(data, raw, changed, 1)
	if err := os.WriteFile(statusPath, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestHostStatus_000(t *testing.T) { hostStatus(t, 0, false) }

func TestHostStatus_001(t *testing.T) { hostStatus(t, 1, false) }

func TestHostStatus_002(t *testing.T) { hostStatus(t, 2, false) }

func TestHostStatus_003(t *testing.T) { hostStatus(t, 3, false) }

func TestHostStatus_004(t *testing.T) { hostStatus(t, 4, false) }

func TestHostStatus_005(t *testing.T) { hostStatus(t, 5, false) }

func TestHostStatus_006(t *testing.T) { hostStatus(t, 6, false) }

func TestHostStatus_007(t *testing.T) { hostStatus(t, 7, false) }

func TestHostStatus_008(t *testing.T) { hostStatus(t, 8, false) }

func TestHostStatus_009(t *testing.T) { hostStatus(t, 9, false) }

func TestHostStatus_010(t *testing.T) { hostStatus(t, 10, false) }

func TestHostStatus_011(t *testing.T) { hostStatus(t, 11, false) }

func TestHostStatus_012(t *testing.T) { hostStatus(t, 12, false) }

func TestHostStatus_013(t *testing.T) { hostStatus(t, 13, false) }

func TestHostStatus_014(t *testing.T) { hostStatus(t, 14, false) }

func TestHostStatus_015(t *testing.T) { hostStatus(t, 15, false) }

func TestHostStatus_016(t *testing.T) { hostStatus(t, 16, false) }

func TestHostStatus_017(t *testing.T) { hostStatus(t, 17, false) }

func TestHostStatus_018(t *testing.T) { hostStatus(t, 18, false) }

func TestHostStatus_019(t *testing.T) { hostStatus(t, 19, false) }

func TestHostStatus_020(t *testing.T) { hostStatus(t, 20, false) }

func TestHostStatus_021(t *testing.T) { hostStatus(t, 21, false) }

func TestHostStatus_022(t *testing.T) { hostStatus(t, 22, false) }

func TestHostStatus_023(t *testing.T) { hostStatus(t, 23, false) }

func TestHostStatus_024(t *testing.T) { hostStatus(t, 24, false) }

func TestHostStatusWASI_000(t *testing.T) { hostStatus(t, 0, true) }

func TestHostStatusWASI_001(t *testing.T) { hostStatus(t, 1, true) }

func TestHostStatusWASI_002(t *testing.T) { hostStatus(t, 2, true) }

func TestHostStatusWASI_003(t *testing.T) { hostStatus(t, 3, true) }

func TestHostStatusWASI_004(t *testing.T) { hostStatus(t, 4, true) }

func TestHostStatusWASI_005(t *testing.T) { hostStatus(t, 5, true) }

func TestHostStatusWASI_006(t *testing.T) { hostStatus(t, 6, true) }

func TestHostStatusWASI_007(t *testing.T) { hostStatus(t, 7, true) }

func TestHostStatusWASI_008(t *testing.T) { hostStatus(t, 8, true) }

func TestHostStatusWASI_009(t *testing.T) { hostStatus(t, 9, true) }

func TestHostStatusWASI_010(t *testing.T) { hostStatus(t, 10, true) }

func TestHostStatusWASI_011(t *testing.T) { hostStatus(t, 11, true) }

func TestHostStatusWASI_012(t *testing.T) { hostStatus(t, 12, true) }

func TestHostStatusWASI_013(t *testing.T) { hostStatus(t, 13, true) }

func TestHostStatusWASI_014(t *testing.T) { hostStatus(t, 14, true) }

func TestHostStatusWASI_015(t *testing.T) { hostStatus(t, 15, true) }

func TestHostStatusWASI_016(t *testing.T) { hostStatus(t, 16, true) }

func TestHostStatusWASI_017(t *testing.T) { hostStatus(t, 17, true) }

func TestHostStatusWASI_018(t *testing.T) { hostStatus(t, 18, true) }

func TestHostStatusWASI_019(t *testing.T) { hostStatus(t, 19, true) }

func TestHostStatusWASI_020(t *testing.T) { hostStatus(t, 20, true) }

func TestHostStatusWASI_021(t *testing.T) { hostStatus(t, 21, true) }

func TestHostStatusWASI_022(t *testing.T) { hostStatus(t, 22, true) }

func TestHostStatusWASI_023(t *testing.T) { hostStatus(t, 23, true) }

func TestHostStatusWASI_024(t *testing.T) { hostStatus(t, 24, true) }
