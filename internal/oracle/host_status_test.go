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

	"github.com/system-inc/adamic/internal/hostconfiguration"
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
	Configuration *hostconfiguration.Configuration `json:"configuration,omitempty"`
	File          string                           `json:"file"`
	Node          struct {
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
	truth := run{stdout: []byte(row.Node.Stdout), stderr: []byte(row.Node.Stderr), exitCode: row.Node.Exit}
	nodeChecked := false
	t.Run("node", func(t *testing.T) {
		if why := hostconfiguration.Skip(index+1, row.Configuration); why != "" {
			t.Skipf("host status row %d: the node step needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR", index+1)
		}
		nodeBinary := hostconfiguration.Require(t, index+1, row.Configuration)
		actualNode := hostNodeWith(t, inputRun{directory: repository}, nodeBinary, path)
		if difference := disagreement(truth, actualNode); difference != "" {
			t.Fatalf("recorded Node %s: %s", row.File, difference)
		}
		nodeChecked = true
	})
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
			t.Run("native", func(t *testing.T) {
				binary := filepath.Join(directory, "program")
				if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
					t.Fatal(err)
				}
				actual := executeInput(t, inputRun{directory: repository}, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
				if difference := disagreement(truth, actual); difference != "" {
					t.Fatalf("%s native: %s", row.File, difference)
				}
			})
			t.Run("javascript", func(t *testing.T) {
				js := filepath.Join(directory, "program.mjs")
				if err := os.WriteFile(js, []byte(script), 0600); err != nil {
					t.Fatal(err)
				}
				if why := hostconfiguration.Skip(index+1, row.Configuration); why != "" {
					t.Skipf("host status row %d: the javascript step needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR", index+1)
				}
				nodeBinary := hostconfiguration.Require(t, index+1, row.Configuration)
				actual := hostNodeWith(t, inputRun{directory: repository}, nodeBinary, js)
				if difference := disagreement(truth, actual); difference != "" {
					t.Fatalf("%s JavaScript: %s", row.File, difference)
				}
			})
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
				t.Run("wasi", func(t *testing.T) {
					if why := hostconfiguration.Skip(index+1, row.Configuration); why != "" {
						t.Skipf("host status row %d: the wasi step needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR", index+1)
					}
					nodeBinary := hostconfiguration.Require(t, index+1, row.Configuration)
					actual := executeInput(t, inputRun{directory: repository}, nil, nodeBinary, "--disable-warning=ExperimentalWarning", runner, wasm)
					if reason := wasiRefusalReason(actual); reason != "" {
						t.Logf("%s WASI refusal: %s", row.File, reason)
					} else if difference := disagreement(truth, actual); difference != "" {
						t.Fatalf("%s WASI: %s", row.File, difference)
					}
				})
			}
		}
	}
	if nodeChecked {
		t.Logf("%s: current Node agrees; stage=%s diagnostic=%q", row.File, observed.Outcome, observed.What)
	} else {
		t.Logf("%s: reference Node not verified; stage=%s diagnostic=%q", row.File, observed.Outcome, observed.What)
	}
	if observed == row.Stage {
		return
	}
	if !*updateHostStatus || wasiOnly || !nodeChecked {
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

func TestHostStatus_000(t *testing.T) { t.Parallel(); hostStatus(t, 0, false) }

func TestHostStatus_001(t *testing.T) { t.Parallel(); hostStatus(t, 1, false) }

func TestHostStatus_002(t *testing.T) { t.Parallel(); hostStatus(t, 2, false) }

func TestHostStatus_003(t *testing.T) { t.Parallel(); hostStatus(t, 3, false) }

func TestHostStatus_004(t *testing.T) { t.Parallel(); hostStatus(t, 4, false) }

func TestHostStatus_005(t *testing.T) { t.Parallel(); hostStatus(t, 5, false) }

func TestHostStatus_006(t *testing.T) { t.Parallel(); hostStatus(t, 6, false) }

func TestHostStatus_007(t *testing.T) { t.Parallel(); hostStatus(t, 7, false) }

func TestHostStatus_008(t *testing.T) { t.Parallel(); hostStatus(t, 8, false) }

func TestHostStatus_009(t *testing.T) { t.Parallel(); hostStatus(t, 9, false) }

func TestHostStatus_010(t *testing.T) { t.Parallel(); hostStatus(t, 10, false) }

func TestHostStatus_011(t *testing.T) { t.Parallel(); hostStatus(t, 11, false) }

func TestHostStatus_012(t *testing.T) { t.Parallel(); hostStatus(t, 12, false) }

func TestHostStatus_013(t *testing.T) { t.Parallel(); hostStatus(t, 13, false) }

func TestHostStatus_014(t *testing.T) { t.Parallel(); hostStatus(t, 14, false) }

func TestHostStatus_015(t *testing.T) { t.Parallel(); hostStatus(t, 15, false) }

func TestHostStatus_016(t *testing.T) { t.Parallel(); hostStatus(t, 16, false) }

func TestHostStatus_017(t *testing.T) { t.Parallel(); hostStatus(t, 17, false) }

func TestHostStatus_018(t *testing.T) { t.Parallel(); hostStatus(t, 18, false) }

func TestHostStatus_019(t *testing.T) { t.Parallel(); hostStatus(t, 19, false) }

func TestHostStatus_020(t *testing.T) { t.Parallel(); hostStatus(t, 20, false) }

func TestHostStatus_021(t *testing.T) { t.Parallel(); hostStatus(t, 21, false) }

func TestHostStatus_022(t *testing.T) { t.Parallel(); hostStatus(t, 22, false) }

func TestHostStatus_023(t *testing.T) { t.Parallel(); hostStatus(t, 23, false) }

func TestHostStatus_024(t *testing.T) { t.Parallel(); hostStatus(t, 24, false) }

func TestHostStatusWASI_000(t *testing.T) { t.Parallel(); hostStatus(t, 0, true) }

func TestHostStatusWASI_001(t *testing.T) { t.Parallel(); hostStatus(t, 1, true) }

func TestHostStatusWASI_002(t *testing.T) { t.Parallel(); hostStatus(t, 2, true) }

func TestHostStatusWASI_003(t *testing.T) { t.Parallel(); hostStatus(t, 3, true) }

func TestHostStatusWASI_004(t *testing.T) { t.Parallel(); hostStatus(t, 4, true) }

func TestHostStatusWASI_005(t *testing.T) { t.Parallel(); hostStatus(t, 5, true) }

func TestHostStatusWASI_006(t *testing.T) { t.Parallel(); hostStatus(t, 6, true) }

func TestHostStatusWASI_007(t *testing.T) { t.Parallel(); hostStatus(t, 7, true) }

func TestHostStatusWASI_008(t *testing.T) { t.Parallel(); hostStatus(t, 8, true) }

func TestHostStatusWASI_009(t *testing.T) { t.Parallel(); hostStatus(t, 9, true) }

func TestHostStatusWASI_010(t *testing.T) { t.Parallel(); hostStatus(t, 10, true) }

func TestHostStatusWASI_011(t *testing.T) { t.Parallel(); hostStatus(t, 11, true) }

func TestHostStatusWASI_012(t *testing.T) { t.Parallel(); hostStatus(t, 12, true) }

func TestHostStatusWASI_013(t *testing.T) { t.Parallel(); hostStatus(t, 13, true) }

func TestHostStatusWASI_014(t *testing.T) { t.Parallel(); hostStatus(t, 14, true) }

func TestHostStatusWASI_015(t *testing.T) { t.Parallel(); hostStatus(t, 15, true) }

func TestHostStatusWASI_016(t *testing.T) { t.Parallel(); hostStatus(t, 16, true) }

func TestHostStatusWASI_017(t *testing.T) { t.Parallel(); hostStatus(t, 17, true) }

func TestHostStatusWASI_018(t *testing.T) { t.Parallel(); hostStatus(t, 18, true) }

func TestHostStatusWASI_019(t *testing.T) { t.Parallel(); hostStatus(t, 19, true) }

func TestHostStatusWASI_020(t *testing.T) { t.Parallel(); hostStatus(t, 20, true) }

func TestHostStatusWASI_021(t *testing.T) { t.Parallel(); hostStatus(t, 21, true) }

func TestHostStatusWASI_022(t *testing.T) { t.Parallel(); hostStatus(t, 22, true) }

func TestHostStatusWASI_023(t *testing.T) { t.Parallel(); hostStatus(t, 23, true) }

func TestHostStatusWASI_024(t *testing.T) { t.Parallel(); hostStatus(t, 24, true) }

// Both the source oracle and generated JavaScript use the row's named Node build.
func hostNodeWith(t *testing.T, how inputRun, binary, path string) run {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	return executeInput(t, how, nil, binary, "--disable-warning=ExperimentalWarning", runner, path)
}
