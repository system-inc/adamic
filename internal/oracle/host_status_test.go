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

var updateHostStatus = flag.Bool(
	"update-host-status",
	false,
	"remeasure host stage0 after recorded/current Node and native/JavaScript agreement",
)
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
func hostStatus(test *testing.T, index int, wasiOnly bool) {
	test.Helper()
	hostStatusFile(test, index, wasiOnly, filepath.Join(repository, "stage3/fixtures/host/status.json"))
}

// The same check accepts a status input for isolated mutant probes.
func hostStatusFile(test *testing.T, index int, wasiOnly bool, statusPath string) {
	test.Helper()
	if wasiOnly && os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		test.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	hostStatusMutex.Lock()
	data, err := os.ReadFile(statusPath)
	hostStatusMutex.Unlock()
	if err != nil {
		test.Fatal(err)
	}
	var rows []json.RawMessage
	if decodeError := json.Unmarshal(data, &rows); decodeError != nil {
		test.Fatal(decodeError)
	}
	if len(rows) != 25 {
		test.Fatalf("want 25 original host fixtures, got %d", len(rows))
	}
	var row hostMeasurement
	if decodeError := json.Unmarshal(rows[index], &row); decodeError != nil {
		test.Fatal(decodeError)
	}
	path, err := filepath.Abs(filepath.Join(repository, "stage3/fixtures/host", row.File))
	if err != nil {
		test.Fatal(err)
	}
	truth := run{stdout: []byte(row.Node.Stdout), stderr: []byte(row.Node.Stderr), exitCode: row.Node.Exit}
	nodeChecked := false
	test.Run("node", func(test *testing.T) {
		if why := hostconfiguration.Skip(index+1, row.Configuration); why != "" {
			test.Skipf(
				"host status row %d: the node step needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR",
				index+1,
			)
		}
		nodeBinary := hostconfiguration.Require(test, index+1, row.Configuration)
		actualNode := hostNodeWith(test, inputRun{directory: repository}, nodeBinary, path)
		if difference := disagreement(truth, actualNode); difference != "" {
			test.Fatalf("recorded Node %s: %s", row.File, difference)
		}
		nodeChecked = true
	})
	program, err := load.Load([]string{path})
	observed := hostStage{Outcome: "Compiles"}
	var code, script string
	if err == nil {
		loweredProgram, lowerError := lower.Lower(context.Background(), program)
		if lowerError == nil {
			code = native.C(loweredProgram)
			script = javascript.JavaScript(loweredProgram)
		}
		err = lowerError
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
			test.Fatal(err)
		}
		root, rootError := filepath.Abs(repository)
		if rootError != nil {
			test.Fatal(rootError)
		}
		observed.What = strings.ReplaceAll(err.Error(), root+string(filepath.Separator), "")
	} else {
		directory := test.TempDir()
		if !wasiOnly {
			test.Run("native", func(test *testing.T) {
				binary := filepath.Join(directory, "program")
				if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
					test.Fatal(buildError)
				}
				actual := executeInput(
					test,
					inputRun{directory: repository},
					[]string{"ASAN_OPTIONS=detect_leaks=1"},
					binary,
				)
				if difference := disagreement(truth, actual); difference != "" {
					test.Fatalf("%s native: %s", row.File, difference)
				}
			})
			test.Run("javascript", func(test *testing.T) {
				js := filepath.Join(directory, "program.mjs")
				if writeError := os.WriteFile(js, []byte(script), 0600); writeError != nil {
					test.Fatal(writeError)
				}
				if why := hostconfiguration.Skip(index+1, row.Configuration); why != "" {
					test.Skipf(
						"host status row %d: the javascript step needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR",
						index+1,
					)
				}
				nodeBinary := hostconfiguration.Require(test, index+1, row.Configuration)
				actual := hostNodeWith(test, inputRun{directory: repository}, nodeBinary, js)
				if difference := disagreement(truth, actual); difference != "" {
					test.Fatalf("%s JavaScript: %s", row.File, difference)
				}
			})
		}
		if wasiOnly {
			wasm := filepath.Join(directory, "program.wasm")
			if buildError := native.Build(code, wasm, native.Options{Target: "wasm32-wasi"}); buildError != nil {
				var refused *native.TargetRefused
				if !errors.As(buildError, &refused) {
					test.Fatal(buildError)
				}
				test.Logf("%s WASI refusal: %v", row.File, refused)
			} else {
				runner, runnerError := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
				if runnerError != nil {
					test.Fatal(runnerError)
				}
				test.Run("wasi", func(test *testing.T) {
					if why := hostconfiguration.Skip(index+1, row.Configuration); why != "" {
						test.Skipf(
							"host status row %d: the wasi step needs node-without-inspector; set ADAMIC_HOST_NODE_WITHOUT_INSPECTOR",
							index+1,
						)
					}
					nodeBinary := hostconfiguration.Require(test, index+1, row.Configuration)
					actual := executeInput(
						test,
						inputRun{directory: repository},
						nil,
						nodeBinary,
						"--disable-warning=ExperimentalWarning",
						runner,
						wasm,
					)
					if reason := wasiRefusalReason(actual); reason != "" {
						test.Logf("%s WASI refusal: %s", row.File, reason)
					} else if difference := disagreement(truth, actual); difference != "" {
						test.Fatalf("%s WASI: %s", row.File, difference)
					}
				})
			}
		}
	}
	if nodeChecked {
		test.Logf("%s: current Node agrees; stage=%s diagnostic=%q", row.File, observed.Outcome, observed.What)
	} else {
		test.Logf("%s: reference Node not verified; stage=%s diagnostic=%q", row.File, observed.Outcome, observed.What)
	}
	if observed == row.Stage {
		return
	}
	if !*updateHostStatus || wasiOnly || !nodeChecked {
		test.Fatalf("host status changed: got %+v; recorded %+v", observed, row.Stage)
	}
	// Replace only this stage0 field. Keep original fixture provenance and Node bytes.
	hostStatusMutex.Lock()
	defer hostStatusMutex.Unlock()
	data, err = os.ReadFile(statusPath)
	if err != nil {
		test.Fatal(err)
	}
	if decodeError := json.Unmarshal(data, &rows); decodeError != nil {
		test.Fatal(decodeError)
	}
	var fields map[string]json.RawMessage
	if decodeError := json.Unmarshal(rows[index], &fields); decodeError != nil {
		test.Fatal(decodeError)
	}
	replacement, err := json.Marshal(observed)
	if err != nil {
		test.Fatal(err)
	}
	raw := rows[index]
	marker := bytes.Index(raw, []byte(`"stage0":`))
	if marker < 0 {
		test.Fatal("missing stage0 field")
	}
	offset := bytes.Index(raw[marker:], fields["stage0"])
	if offset < 0 {
		test.Fatal("cannot locate stage0 field")
	}
	position := marker + offset
	changed := append(
		append(append([]byte{}, raw[:position]...), replacement...),
		raw[position+len(fields["stage0"]):]...,
	)
	data = bytes.Replace(data, raw, changed, 1)
	if writeError := os.WriteFile(statusPath, data, 0644); writeError != nil {
		test.Fatal(writeError)
	}
}

func TestHostStatus_000(test *testing.T) { test.Parallel(); hostStatus(test, 0, false) }

func TestHostStatus_001(test *testing.T) { test.Parallel(); hostStatus(test, 1, false) }

func TestHostStatus_002(test *testing.T) { test.Parallel(); hostStatus(test, 2, false) }

func TestHostStatus_003(test *testing.T) { test.Parallel(); hostStatus(test, 3, false) }

func TestHostStatus_004(test *testing.T) { test.Parallel(); hostStatus(test, 4, false) }

func TestHostStatus_005(test *testing.T) { test.Parallel(); hostStatus(test, 5, false) }

func TestHostStatus_006(test *testing.T) { test.Parallel(); hostStatus(test, 6, false) }

func TestHostStatus_007(test *testing.T) { test.Parallel(); hostStatus(test, 7, false) }

func TestHostStatus_008(test *testing.T) { test.Parallel(); hostStatus(test, 8, false) }

func TestHostStatus_009(test *testing.T) { test.Parallel(); hostStatus(test, 9, false) }

func TestHostStatus_010(test *testing.T) { test.Parallel(); hostStatus(test, 10, false) }

func TestHostStatus_011(test *testing.T) { test.Parallel(); hostStatus(test, 11, false) }

func TestHostStatus_012(test *testing.T) { test.Parallel(); hostStatus(test, 12, false) }

func TestHostStatus_013(test *testing.T) { test.Parallel(); hostStatus(test, 13, false) }

func TestHostStatus_014(test *testing.T) { test.Parallel(); hostStatus(test, 14, false) }

func TestHostStatus_015(test *testing.T) { test.Parallel(); hostStatus(test, 15, false) }

func TestHostStatus_016(test *testing.T) { test.Parallel(); hostStatus(test, 16, false) }

func TestHostStatus_017(test *testing.T) { test.Parallel(); hostStatus(test, 17, false) }

func TestHostStatus_018(test *testing.T) { test.Parallel(); hostStatus(test, 18, false) }

func TestHostStatus_019(test *testing.T) { test.Parallel(); hostStatus(test, 19, false) }

func TestHostStatus_020(test *testing.T) { test.Parallel(); hostStatus(test, 20, false) }

func TestHostStatus_021(test *testing.T) { test.Parallel(); hostStatus(test, 21, false) }

func TestHostStatus_022(test *testing.T) { test.Parallel(); hostStatus(test, 22, false) }

func TestHostStatus_023(test *testing.T) { test.Parallel(); hostStatus(test, 23, false) }

func TestHostStatus_024(test *testing.T) { test.Parallel(); hostStatus(test, 24, false) }

func TestHostStatusWasi_000(test *testing.T) { test.Parallel(); hostStatus(test, 0, true) }

func TestHostStatusWasi_001(test *testing.T) { test.Parallel(); hostStatus(test, 1, true) }

func TestHostStatusWasi_002(test *testing.T) { test.Parallel(); hostStatus(test, 2, true) }

func TestHostStatusWasi_003(test *testing.T) { test.Parallel(); hostStatus(test, 3, true) }

func TestHostStatusWasi_004(test *testing.T) { test.Parallel(); hostStatus(test, 4, true) }

func TestHostStatusWasi_005(test *testing.T) { test.Parallel(); hostStatus(test, 5, true) }

func TestHostStatusWasi_006(test *testing.T) { test.Parallel(); hostStatus(test, 6, true) }

func TestHostStatusWasi_007(test *testing.T) { test.Parallel(); hostStatus(test, 7, true) }

func TestHostStatusWasi_008(test *testing.T) { test.Parallel(); hostStatus(test, 8, true) }

func TestHostStatusWasi_009(test *testing.T) { test.Parallel(); hostStatus(test, 9, true) }

func TestHostStatusWasi_010(test *testing.T) { test.Parallel(); hostStatus(test, 10, true) }

func TestHostStatusWasi_011(test *testing.T) { test.Parallel(); hostStatus(test, 11, true) }

func TestHostStatusWasi_012(test *testing.T) { test.Parallel(); hostStatus(test, 12, true) }

func TestHostStatusWasi_013(test *testing.T) { test.Parallel(); hostStatus(test, 13, true) }

func TestHostStatusWasi_014(test *testing.T) { test.Parallel(); hostStatus(test, 14, true) }

func TestHostStatusWasi_015(test *testing.T) { test.Parallel(); hostStatus(test, 15, true) }

func TestHostStatusWasi_016(test *testing.T) { test.Parallel(); hostStatus(test, 16, true) }

func TestHostStatusWasi_017(test *testing.T) { test.Parallel(); hostStatus(test, 17, true) }

func TestHostStatusWasi_018(test *testing.T) { test.Parallel(); hostStatus(test, 18, true) }

func TestHostStatusWasi_019(test *testing.T) { test.Parallel(); hostStatus(test, 19, true) }

func TestHostStatusWasi_020(test *testing.T) { test.Parallel(); hostStatus(test, 20, true) }

func TestHostStatusWasi_021(test *testing.T) { test.Parallel(); hostStatus(test, 21, true) }

func TestHostStatusWasi_022(test *testing.T) { test.Parallel(); hostStatus(test, 22, true) }

func TestHostStatusWasi_023(test *testing.T) { test.Parallel(); hostStatus(test, 23, true) }

func TestHostStatusWasi_024(test *testing.T) { test.Parallel(); hostStatus(test, 24, true) }

// Both the source oracle and generated JavaScript use the row's named Node build.
func hostNodeWith(test *testing.T, how inputRun, binary, path string) run {
	test.Helper()
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return executeInput(test, how, nil, binary, "--disable-warning=ExperimentalWarning", runner, path)
}
