package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func jsxMutantsRows(t *testing.T, manifest string, count int) []string {
	t.Helper()
	data, err := os.ReadFile(manifest)
	if err != nil {
		t.Fatal(err)
	}
	rows := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(rows) != count {
		t.Fatalf("enumerated %d inputs, want %d", len(rows), count)
	}
	return rows
}

func jsxMutantsUnion(t *testing.T, names, rows []string, shards [][]string) {
	t.Helper()
	expected := map[string]bool{}
	for _, name := range names {
		for _, row := range rows {
			id := name + "\t" + row
			if expected[id] {
				t.Fatalf("duplicate unsplit case %q", id)
			}
			expected[id] = true
		}
	}
	seen := map[string]bool{}
	for _, ids := range shards {
		for _, id := range ids {
			if seen[id] || !expected[id] {
				t.Fatalf("repeated or unexpected shard case %q", id)
			}
			seen[id] = true
		}
	}
	if len(seen) != len(expected) {
		t.Fatalf("shard union %d differs from enumeration %d", len(seen), len(expected))
	}
	for id := range expected {
		if !seen[id] {
			t.Fatalf("missing shard case %q", id)
		}
	}
}

func jsxMutantsSelection(t *testing.T) int {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return -1
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		t.Fatalf("ADAMIC_TEST_SHARD must be i/%d", testJsxMutantsShards)
	}
	i, err := strconv.Atoi(fields[0])
	n, nerr := strconv.Atoi(fields[1])
	if err != nil || nerr != nil || n != testJsxMutantsShards || i < 0 || i >= n {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q; want i/%d", value, testJsxMutantsShards)
	}
	return i
}

// Each selected shard prepares its own products before its comparison deadline.
func jsxMutantsRun(t *testing.T, shard int) {
	t.Helper()
	selected := jsxMutantsSelection(t)
	if selected >= 0 && selected != shard {
		t.Skip("different ADAMIC_TEST_SHARD selected")
	}
	changes := jsxMutantsChanges()
	if len(changes) != testJsxMutantsShards {
		t.Fatalf("enumerated %d mutants, want %d", len(changes), testJsxMutantsShards)
	}
	setupStarted := time.Now()
	manifest, count := jsxManifest(t)
	jsxMutantsRows(t, manifest, count)
	change := changes[shard]
	prepared := jsxMutantsPrepare(t, shard)
	oracle, controlBinary, binary, mutant := prepared.Oracle, prepared.Control, prepared.Binary, prepared.Mutant
	t.Logf("setup %.2fs", time.Since(setupStarted).Seconds())
	ownStarted := time.Now()
	deadline := time.AfterFunc(60*time.Second, func() { panic("JSX mutant own work exceeded 60 seconds") })
	defer deadline.Stop()
	defer func() { t.Logf("own work %.2fs", time.Since(ownStarted).Seconds()) }()
	want := execute(t, "", oracle, "--manifest", manifest, "--whole").output
	control := execute(t, "", controlBinary, "--manifest", manifest, "--whole")
	if diff := difference(control.output, want); diff != "" {
		t.Fatal(diff)
	}
	node := wholeNode(t, mutant, manifest, false)
	compiled := execute(t, "", binary, "--manifest", manifest, "--whole")
	if os.Getenv("ADAMIC_JSX_MUTANTS_PLANTED_SURVIVOR") == change.name {
		node.output = want
	}
	for _, side := range []struct {
		name   string
		output []byte
	}{{"Node", node.output}, {"native", compiled.output}} {
		if bytes.Equal(side.output, want) {
			t.Fatalf("%s mutant survived: %s", side.name, change.name)
		}
		t.Logf("%s compiled mutant caught: %s", side.name, difference(side.output, want))
	}
	t.Logf("shard-%03d: %s, %d inputs", shard, change.name, count)
	if os.Getenv("ADAMIC_JSX_MUTANTS_PROOF_CHILD") == "" {
		jsxMutantsPlantedProof(t, shard, prepared)
	}
}

// jsxMutantsPrepared is what a shard runs: the oracle, the control and mutant binaries, and the mutant's source.
type jsxMutantsPrepared struct{ Oracle, Control, Binary, Mutant string }

// jsxMutantsPrepare fetches a shard's products. The planted-proof child is handed its parent's instead, so the child's
// fetch never runs on the parent's own-work clock.
func jsxMutantsPrepare(t *testing.T, shard int) jsxMutantsPrepared {
	t.Helper()
	if descriptor := os.Getenv(jsxMutantsPreparedEnv); descriptor != "" {
		data, err := os.ReadFile(descriptor)
		if err != nil {
			t.Fatal(err)
		}
		var prepared jsxMutantsPrepared
		if err := json.Unmarshal(data, &prepared); err != nil {
			t.Fatal(err)
		}
		return prepared
	}
	directory, _ := filepath.Abs(".")
	mutant := jsxMutantsSource(t, shard)
	return jsxMutantsPrepared{Oracle: goOracle(t), Control: buildPort(t, directory, true), Binary: buildPort(t, mutant, true), Mutant: mutant}
}

const jsxMutantsPreparedEnv = "ADAMIC_JSX_MUTANTS_PREPARED"

func TestJsxMutants_000(t *testing.T) { t.Parallel(); jsxMutantsRun(t, 0) }
func TestJsxMutants_001(t *testing.T) { t.Parallel(); jsxMutantsRun(t, 1) }
func TestJsxMutants_002(t *testing.T) { t.Parallel(); jsxMutantsRun(t, 2) }
func TestJsxMutants_003(t *testing.T) { t.Parallel(); jsxMutantsRun(t, 3) }
func TestJsxMutants_004(t *testing.T) { t.Parallel(); jsxMutantsRun(t, 4) }
func TestJsxMutants_005(t *testing.T) { t.Parallel(); jsxMutantsRun(t, 5) }
func TestJsxMutants_006(t *testing.T) { t.Parallel(); jsxMutantsRun(t, 6) }
func TestJsxMutants_007(t *testing.T) { t.Parallel(); jsxMutantsRun(t, 7) }
func TestJsxMutants_008(t *testing.T) { t.Parallel(); jsxMutantsRun(t, 8) }

func TestJsxMutantsUnion(t *testing.T) {
	t.Parallel()
	changes := jsxMutantsChanges()
	if len(changes) != testJsxMutantsShards {
		t.Fatalf("enumerated %d shards, want %d", len(changes), testJsxMutantsShards)
	}
	manifest, count := jsxManifest(t)
	rows := jsxMutantsRows(t, manifest, count)
	names := make([]string, len(changes))
	shards := make([][]string, len(changes))
	for i, change := range changes {
		names[i] = change.name
		for _, row := range rows {
			shards[i] = append(shards[i], change.name+"\t"+row)
		}
	}
	jsxMutantsUnion(t, names, rows, shards)
	t.Logf("union: %d mutants x %d inputs = %d unique case ids", len(changes), count, len(changes)*count)
}

// Prove each ownership leaf independently; exactly shard 004 must fail.
func jsxMutantsPlantedProof(t *testing.T, shard int, prepared jsxMutantsPrepared) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(prepared)
	if err != nil {
		t.Fatal(err)
	}
	descriptor := filepath.Join(t.TempDir(), "prepared.json")
	if err := os.WriteFile(descriptor, encoded, 0644); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	name := fmt.Sprintf("TestJsxMutants_%03d", shard)
	cmd := jsxMutantsCommand(ctx, executable, "-test.run=^"+name+"$", "-test.v", "-test.timeout=90s")
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(entry, "ADAMIC_JSX_MUTANTS_PLANTED_SURVIVOR=") && !strings.HasPrefix(entry, "ADAMIC_JSX_MUTANTS_PROOF_CHILD=") && !strings.HasPrefix(entry, jsxMutantsPreparedEnv+"=") {
			cmd.Env = append(cmd.Env, entry)
		}
	}
	cmd.Env = append(cmd.Env, "ADAMIC_JSX_MUTANTS_PLANTED_SURVIVOR=type argument comma", "ADAMIC_JSX_MUTANTS_PROOF_CHILD=1", jsxMutantsPreparedEnv+"="+descriptor)
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil {
		t.Fatal("planted proof exceeded 60 seconds")
	}
	if shard == 4 {
		if _, ok := err.(*exec.ExitError); !ok || !strings.Contains(string(out), "Node mutant survived: type argument comma") || !strings.Contains(string(out), "--- FAIL: "+name) {
			t.Fatalf("owner failed to catch planted survivor: %v\n%s", err, out)
		}
	} else if err != nil {
		t.Fatalf("nonowner caught planted survivor: %v\n%s", err, out)
	}
	t.Logf("planted survivor ownership proved for %s", name)
}

// Give spawned compilers the child's process group so cancellation cleans up
// the whole build on Linux and macOS without an external timeout executable.
func jsxMutantsCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = 3 * time.Second
	return cmd
}

func jsxMutantsSource(t *testing.T, index int) string {
	t.Helper()
	change := jsxMutantsChanges()[index]
	return copyPort(t, "jsx.ts", change.from, change.to)
}

// The oracle is TestProduct_ParserFindsOracle's and the control TestProduct_WholeMutantsNative_Control's.
func TestProduct_JsxMutantsNative_000(t *testing.T) {
	t.Parallel()
	buildPort(t, jsxMutantsSource(t, 0), true)
}
func TestProduct_JsxMutantsNative_001(t *testing.T) {
	t.Parallel()
	buildPort(t, jsxMutantsSource(t, 1), true)
}
func TestProduct_JsxMutantsNative_002(t *testing.T) {
	t.Parallel()
	buildPort(t, jsxMutantsSource(t, 2), true)
}
func TestProduct_JsxMutantsNative_003(t *testing.T) {
	t.Parallel()
	buildPort(t, jsxMutantsSource(t, 3), true)
}
func TestProduct_JsxMutantsNative_004(t *testing.T) {
	t.Parallel()
	buildPort(t, jsxMutantsSource(t, 4), true)
}
func TestProduct_JsxMutantsNative_005(t *testing.T) {
	t.Parallel()
	buildPort(t, jsxMutantsSource(t, 5), true)
}
func TestProduct_JsxMutantsNative_006(t *testing.T) {
	t.Parallel()
	buildPort(t, jsxMutantsSource(t, 6), true)
}
func TestProduct_JsxMutantsNative_007(t *testing.T) {
	t.Parallel()
	buildPort(t, jsxMutantsSource(t, 7), true)
}
func TestProduct_JsxMutantsNative_008(t *testing.T) {
	t.Parallel()
	buildPort(t, jsxMutantsSource(t, 8), true)
}
