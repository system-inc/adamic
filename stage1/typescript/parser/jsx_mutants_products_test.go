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

	"github.com/system-inc/adamic/internal/childguard"
)

func jsxMutantsRows(test *testing.T, manifest string, count int) []string {
	test.Helper()
	data, readError := os.ReadFile(manifest)
	if readError != nil {
		test.Fatal(readError)
	}
	rows := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(rows) != count {
		test.Fatalf("enumerated %d inputs, want %d", len(rows), count)
	}
	return rows
}

func jsxMutantsUnion(test *testing.T, names, rows []string, shards [][]string) {
	test.Helper()
	expected := map[string]bool{}
	for _, name := range names {
		for _, row := range rows {
			id := name + "\t" + row
			if expected[id] {
				test.Fatalf("duplicate unsplit case %q", id)
			}
			expected[id] = true
		}
	}
	seen := map[string]bool{}
	for _, ids := range shards {
		for _, id := range ids {
			if seen[id] || !expected[id] {
				test.Fatalf("repeated or unexpected shard case %q", id)
			}
			seen[id] = true
		}
	}
	if len(seen) != len(expected) {
		test.Fatalf("shard union %d differs from enumeration %d", len(seen), len(expected))
	}
	for id := range expected {
		if !seen[id] {
			test.Fatalf("missing shard case %q", id)
		}
	}
}

func jsxMutantsSelection(test *testing.T) int {
	test.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return -1
	}
	fields := strings.Split(value, "/")
	if len(fields) != 2 {
		test.Fatalf("ADAMIC_TEST_SHARD must be i/%d", testJsxMutantsShards)
	}
	index, parseError := strconv.Atoi(fields[0])
	n, nerr := strconv.Atoi(fields[1])
	if parseError != nil || nerr != nil || n != testJsxMutantsShards || index < 0 || index >= n {
		test.Fatalf("invalid ADAMIC_TEST_SHARD %q; want i/%d", value, testJsxMutantsShards)
	}
	return index
}

// Each selected shard prepares its own products before its comparison deadline.
func jsxMutantsRun(test *testing.T, shard int) {
	test.Helper()
	selected := jsxMutantsSelection(test)
	if selected >= 0 && selected != shard {
		test.Skip("different ADAMIC_TEST_SHARD selected")
	}
	changes := jsxMutantsChanges()
	if len(changes) != testJsxMutantsShards {
		test.Fatalf("enumerated %d mutants, want %d", len(changes), testJsxMutantsShards)
	}
	setupStarted := time.Now()
	manifest, count := jsxManifest(test)
	jsxMutantsRows(test, manifest, count)
	change := changes[shard]
	prepared := jsxMutantsPrepare(test, shard)
	oracle, controlBinary, binary, mutant := prepared.Oracle, prepared.Control, prepared.Binary, prepared.Mutant
	test.Logf("setup %.2fs", time.Since(setupStarted).Seconds())
	ownStarted := time.Now()
	deadline := time.AfterFunc(60*time.Second, func() { panic("JSX mutant own work exceeded 60 seconds") })
	defer deadline.Stop()
	defer func() { test.Logf("own work %.2fs", time.Since(ownStarted).Seconds()) }()
	want := execute(test, "", oracle, "--manifest", manifest, "--whole").output
	control := execute(test, "", controlBinary, "--manifest", manifest, "--whole")
	if diff := difference(control.output, want); diff != "" {
		test.Fatal(diff)
	}
	node := wholeNode(test, mutant, manifest, false)
	compiled := execute(test, "", binary, "--manifest", manifest, "--whole")
	if os.Getenv("ADAMIC_JSX_MUTANTS_PLANTED_SURVIVOR") == change.name {
		node.output = want
	}
	for _, side := range []struct {
		name   string
		output []byte
	}{{"Node", node.output}, {"native", compiled.output}} {
		if bytes.Equal(side.output, want) {
			test.Fatalf("%s mutant survived: %s", side.name, change.name)
		}
		test.Logf("%s compiled mutant caught: %s", side.name, difference(side.output, want))
	}
	test.Logf("shard-%03d: %s, %d inputs", shard, change.name, count)
	if os.Getenv("ADAMIC_JSX_MUTANTS_PROOF_CHILD") == "" {
		jsxMutantsPlantedProof(test, shard, prepared)
	}
}

// jsxMutantsPrepared is what a shard runs: the oracle, the control and mutant binaries, and the mutant's source.
type jsxMutantsPrepared struct{ Oracle, Control, Binary, Mutant string }

// jsxMutantsPrepare fetches a shard's products. The planted-proof child is handed its parent's instead, so the child's
// fetch never runs on the parent's own-work clock.
func jsxMutantsPrepare(test *testing.T, shard int) jsxMutantsPrepared {
	test.Helper()
	if descriptor := os.Getenv(jsxMutantsPreparedEnv); descriptor != "" {
		data, readError := os.ReadFile(descriptor)
		if readError != nil {
			test.Fatal(readError)
		}
		var prepared jsxMutantsPrepared
		if decodeError := json.Unmarshal(data, &prepared); decodeError != nil {
			test.Fatal(decodeError)
		}
		return prepared
	}
	directory, _ := filepath.Abs(".")
	mutant := jsxMutantsSource(test, shard)
	return jsxMutantsPrepared{
		Oracle:  goOracle(test),
		Control: buildPort(test, directory, true),
		Binary:  buildPort(test, mutant, true),
		Mutant:  mutant,
	}
}

const jsxMutantsPreparedEnv = "ADAMIC_JSX_MUTANTS_PREPARED"

func TestJsxMutants_000(test *testing.T) { test.Parallel(); jsxMutantsRun(test, 0) }
func TestJsxMutants_001(test *testing.T) { test.Parallel(); jsxMutantsRun(test, 1) }
func TestJsxMutants_002(test *testing.T) { test.Parallel(); jsxMutantsRun(test, 2) }
func TestJsxMutants_003(test *testing.T) { test.Parallel(); jsxMutantsRun(test, 3) }
func TestJsxMutants_004(test *testing.T) { test.Parallel(); jsxMutantsRun(test, 4) }
func TestJsxMutants_005(test *testing.T) { test.Parallel(); jsxMutantsRun(test, 5) }
func TestJsxMutants_006(test *testing.T) { test.Parallel(); jsxMutantsRun(test, 6) }
func TestJsxMutants_007(test *testing.T) { test.Parallel(); jsxMutantsRun(test, 7) }
func TestJsxMutants_008(test *testing.T) { test.Parallel(); jsxMutantsRun(test, 8) }

func TestJsxMutantsUnion(test *testing.T) {
	test.Parallel()
	changes := jsxMutantsChanges()
	if len(changes) != testJsxMutantsShards {
		test.Fatalf("enumerated %d shards, want %d", len(changes), testJsxMutantsShards)
	}
	manifest, count := jsxManifest(test)
	rows := jsxMutantsRows(test, manifest, count)
	names := make([]string, len(changes))
	shards := make([][]string, len(changes))
	for index, change := range changes {
		names[index] = change.name
		for _, row := range rows {
			shards[index] = append(shards[index], change.name+"\t"+row)
		}
	}
	jsxMutantsUnion(test, names, rows, shards)
	test.Logf("union: %d mutants x %d inputs = %d unique case ids", len(changes), count, len(changes)*count)
}

// Prove each ownership leaf independently; exactly shard 004 must fail.
func jsxMutantsPlantedProof(test *testing.T, shard int, prepared jsxMutantsPrepared) {
	test.Helper()
	executable, executableError := os.Executable()
	if executableError != nil {
		test.Fatal(executableError)
	}
	encoded, encodeError := json.Marshal(prepared)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	descriptor := filepath.Join(test.TempDir(), "prepared.json")
	if writeError := os.WriteFile(descriptor, encoded, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	name := fmt.Sprintf("TestJsxMutants_%03d", shard)
	command := jsxMutantsCommand(timeoutContext, executable, "-test.run=^"+name+"$", "-test.v", "-test.timeout=90s")
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") &&
			!strings.HasPrefix(entry, "ADAMIC_JSX_MUTANTS_PLANTED_SURVIVOR=") &&
			!strings.HasPrefix(entry, "ADAMIC_JSX_MUTANTS_PROOF_CHILD=") &&
			!strings.HasPrefix(entry, jsxMutantsPreparedEnv+"=") {
			command.Env = append(command.Env, entry)
		}
	}
	command.Env = append(
		command.Env,
		"ADAMIC_JSX_MUTANTS_PLANTED_SURVIVOR=type argument comma",
		"ADAMIC_JSX_MUTANTS_PROOF_CHILD=1",
		jsxMutantsPreparedEnv+"="+descriptor,
	)
	// Read through childguard.Capture for jsxMutantsCommand's reason, stdout then stderr.
	stdout, stderr, captureError := childguard.Capture(timeoutContext, command)
	out := append(stdout, stderr...)
	if timeoutContext.Err() != nil {
		test.Fatal("planted proof exceeded 60 seconds")
	}
	if shard == 4 {
		if _, ok := captureError.(*exec.ExitError); !ok ||
			!strings.Contains(string(out), "Node mutant survived: type argument comma") ||
			!strings.Contains(string(out), "--- FAIL: "+name) {
			test.Fatalf("owner failed to catch planted survivor: %v\n%s", captureError, out)
		}
	} else if captureError != nil {
		test.Fatalf("nonowner caught planted survivor: %v\n%s", captureError, out)
	}
	test.Logf("planted survivor ownership proved for %s", name)
}

// Give spawned compilers the child's process group so cancellation cleans up
// the whole build on Linux and macOS without an external timeout executable.
// Its output is read through childguard.Capture, to its end with no timer
// deciding when: exec's own copying under a three-second WaitDelay reported
// exec.ErrWaitDelay whenever the reads hadn't reached EOF three seconds after
// the child exited, so a planted proof that had finished could fail under load,
// and the owner's exit error was never seen (#n3gk22h's shape). The proof's
// 60 s deadline ends a hung run.
func jsxMutantsCommand(callerContext context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(callerContext, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		killError := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(killError, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return killError
	}
	return command
}

func jsxMutantsSource(test *testing.T, index int) string {
	test.Helper()
	change := jsxMutantsChanges()[index]
	return copyPort(test, "jsx.ts", change.from, change.to)
}

// The oracle is TestProduct_ParserFindsOracle's and the control TestProduct_WholeMutantsNative_Control's.
func TestProduct_JsxMutantsNative_000(test *testing.T) {
	test.Parallel()
	buildPort(test, jsxMutantsSource(test, 0), true)
}
func TestProduct_JsxMutantsNative_001(test *testing.T) {
	test.Parallel()
	buildPort(test, jsxMutantsSource(test, 1), true)
}
func TestProduct_JsxMutantsNative_002(test *testing.T) {
	test.Parallel()
	buildPort(test, jsxMutantsSource(test, 2), true)
}
func TestProduct_JsxMutantsNative_003(test *testing.T) {
	test.Parallel()
	buildPort(test, jsxMutantsSource(test, 3), true)
}
func TestProduct_JsxMutantsNative_004(test *testing.T) {
	test.Parallel()
	buildPort(test, jsxMutantsSource(test, 4), true)
}
func TestProduct_JsxMutantsNative_005(test *testing.T) {
	test.Parallel()
	buildPort(test, jsxMutantsSource(test, 5), true)
}
func TestProduct_JsxMutantsNative_006(test *testing.T) {
	test.Parallel()
	buildPort(test, jsxMutantsSource(test, 6), true)
}
func TestProduct_JsxMutantsNative_007(test *testing.T) {
	test.Parallel()
	buildPort(test, jsxMutantsSource(test, 7), true)
}
func TestProduct_JsxMutantsNative_008(test *testing.T) {
	test.Parallel()
	buildPort(test, jsxMutantsSource(test, 8), true)
}
