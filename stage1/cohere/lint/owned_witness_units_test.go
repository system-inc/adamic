package lint

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testOwnedWitnessesShards = 16

type ownedWitnessCase struct {
	id, row    string
	descriptor registry.Descriptor
}

func ownedWitnessCases(test *testing.T, directory string) []ownedWitnessCase {
	test.Helper()
	var cases []ownedWitnessCase
	for _, d := range prepareRegistry(test, directory) {
		paths, err := registry.Witnesses(filepath.Join(directory, "rules", d.Slug))
		if err != nil {
			test.Fatal(err)
		}
		rows := ownedWitnessRows(test, directory, d)
		if len(paths) != len(rows) {
			test.Fatalf("%s: %d paths, %d rows", d.Slug, len(paths), len(rows))
		}
		for index, row := range rows {
			relative, err := filepath.Rel(directory, paths[index])
			if err != nil {
				test.Fatal(err)
			}
			cases = append(cases, ownedWitnessCase{filepath.ToSlash(relative), row, d})
		}
	}
	return cases
}

// Validate the entire plan before applying a box selector: every original identity
// must occur exactly once, including on boxes which execute only part of the plan.
func planOwnedWitnesses(cases []ownedWitnessCase, count int) ([][]ownedWitnessCase, error) {
	if count < 1 || len(cases) == 0 {
		return nil, fmt.Errorf("%d witnesses cannot fill %d shards", len(cases), count)
	}
	original := make(map[string]bool, len(cases))
	plan := make([][]ownedWitnessCase, count)
	for _, witness := range cases {
		if original[witness.id] {
			return nil, fmt.Errorf("repeated unsplit witness %s", witness.id)
		}
		original[witness.id] = true
		shard := ownedWitnessShard(witness.id, count)
		plan[shard] = append(plan[shard], witness)
	}
	if err := checkOwnedWitnessUnion(cases, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// The repository corpus is live. Hash its repository-relative path and mode,
// so adding or reordering witnesses never moves an existing case.
func ownedWitnessShard(id string, count int) int {
	h := fnv.New64a()
	_, _ = h.Write([]byte("stage1/cohere/lint/" + id + "\x00all"))
	return int(h.Sum64() % uint64(count))
}

func TestOwnedWitnessesAssignmentStable(test *testing.T) {
	test.Parallel()
	cases := ownedWitnessCases(test, ".")
	original, err := planOwnedWitnesses(cases, testOwnedWitnessesShards)
	if err != nil {
		test.Fatal(err)
	}
	assignments := make(map[string]int)
	for shard, group := range original {
		for _, witness := range group {
			assignments[witness.id] = shard
		}
	}
	grown := append([]ownedWitnessCase{{id: "rules/new-rule/witness/new.ts"}}, cases...)
	for i, j := 0, len(grown)-1; i < j; i, j = i+1, j-1 {
		grown[i], grown[j] = grown[j], grown[i]
	}
	plan, err := planOwnedWitnesses(grown, testOwnedWitnessesShards)
	if err != nil {
		test.Fatal(err)
	}
	for shard, group := range plan {
		for _, witness := range group {
			if previous, exists := assignments[witness.id]; exists && previous != shard {
				test.Fatalf("%s moved from shard-%03d to shard-%03d", witness.id, previous, shard)
			}
		}
	}
	if _, err := planOwnedWitnesses(nil, testOwnedWitnessesShards); err == nil {
		test.Fatal("empty repository corpus survived")
	}
}

func checkOwnedWitnessUnion(cases []ownedWitnessCase, plan [][]ownedWitnessCase) error {
	want := make(map[string]bool, len(cases))
	for _, witness := range cases {
		if want[witness.id] {
			return fmt.Errorf("repeated unsplit witness %s", witness.id)
		}
		want[witness.id] = true
	}
	seen := make(map[string]bool, len(cases))
	total := 0
	for _, group := range plan {
		for _, witness := range group {
			total++
			if !want[witness.id] {
				return fmt.Errorf("unexpected witness %s", witness.id)
			}
			if seen[witness.id] {
				return fmt.Errorf("repeated witness %s", witness.id)
			}
			seen[witness.id] = true
		}
	}
	if total != len(cases) {
		return fmt.Errorf("union has %d witnesses, want %d", total, len(cases))
	}
	for id := range want {
		if !seen[id] {
			return fmt.Errorf("missing witness %s", id)
		}
	}
	return nil
}

// The repository corpus is live; stable path + mode hashing keeps assignments
// fixed as it grows. ADAMIC_TEST_SHARD=i/n selects shard indices modulo n;
// unset runs all. Each top-level shard is independently discoverable by go test -list.
func ownedWitnessUnit(test *testing.T, shard int) {
	test.Helper()
	if text := os.Getenv("ADAMIC_TEST_SHARD"); text != "" {
		parts := strings.Split(text, "/")
		if len(parts) != 2 {
			test.Fatal("ADAMIC_TEST_SHARD must be i/n")
		}
		i, e1 := strconv.Atoi(parts[0])
		n, e2 := strconv.Atoi(parts[1])
		if e1 != nil || e2 != nil || n < 1 || i < 0 || i >= n {
			test.Fatal("invalid ADAMIC_TEST_SHARD")
		}
		if shard%n != i {
			test.Skip("assigned to another box")
		}
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		test.Fatal(err)
	}
	cases := ownedWitnessCases(test, directory)
	plan, err := planOwnedWitnesses(cases, testOwnedWitnessesShards)
	if err != nil {
		test.Fatal(err)
	}
	if os.Getenv("ADAMIC_OWNED_WITNESS_PROBE") == "1" {
		for _, witness := range plan[shard] {
			want := []byte("case 0\n")
			got := want
			if witness.id == cases[0].id {
				got = []byte("case 0\nplanted disagreement\n")
			}
			if err := ownedWitnessAgreement(witness.id, "native", got, want); err != nil {
				test.Fatal(err)
			}
		}
		return
	}
	// A shard selected alone prepares the shared inputs itself, before its own deadline starts.
	ownedWitnessSetup(test)
	oracle, binary, module := ownedWitnessShared.oracle, ownedWitnessShared.binary, ownedWitnessShared.module
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	started := time.Now()
	defer func() {
		test.Logf("case wall=%s; shard=%03d witnesses=%d", time.Since(started), shard, len(plan[shard]))
	}()
	var rows []string
	for _, witness := range plan[shard] {
		path := strings.SplitN(witness.row, "\t", 2)[0]
		pair := ownedWitnessRecoveryRows(test, ctx, oracle, []string{witness.row, path + "\tall"})
		answer := ownedWitnessExecute(test, ctx, oracle, "--manifest", manifest(test, pair[:1]), "--count")
		if string(answer.output) == "0\n" {
			test.Fatalf("%s witness reports no findings", witness.descriptor.Name)
		}
		rows = append(rows, pair...)
	}
	if len(rows) == 0 {
		return
	}
	path := manifest(test, rows)
	want := ownedWitnessExecute(test, ctx, oracle, "--manifest", path)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", ownedWitnessExecute(
			test,
			ctx,
			"node",
			"--disable-warning=ExperimentalWarning",
			filepath.Join(repository, "oracle/node.mjs"),
			filepath.Join(directory, "main.ts"),
			"--manifest",
			path,
		)},
		{"emitted JavaScript", ownedWitnessExecute(
			test,
			ctx,
			"node",
			"--disable-warning=ExperimentalWarning",
			filepath.Join(repository, "oracle/node.mjs"),
			module,
			"--manifest",
			path,
		)},
		{"native", ownedWitnessExecute(test, ctx, binary, "--manifest", path)},
	} {
		if err := ownedWitnessAgreement(
			fmt.Sprintf("shard-%03d", shard),
			side.name,
			side.run.output,
			want.output,
		); err != nil {
			test.Fatal(err)
		}
	}
}

var ownedWitnessShared struct {
	sync.Once
	oracle, binary, module string
}

func ownedWitnessSetup(test *testing.T) {
	test.Helper()
	ownedWitnessShared.Do(func() {
		// No deadline of its own: setup is the build phase's to make fast, and Loom's 90 s kill still covers
		// the unit (@system_adamic's ruling). A shard's clock starts only after this returns.
		started := time.Now()
		directory, err := filepath.Abs(".")
		if err != nil {
			test.Fatal(err)
		}
		oracle := lintGoOracleProduct(test)
		binary, module := ownedWitnessProducts(test, directory)
		ownedWitnessShared.oracle, ownedWitnessShared.binary, ownedWitnessShared.module = oracle, binary, module
		test.Logf("shared setup=%s", time.Since(started))
	})
	if ownedWitnessShared.oracle == "" {
		test.Fatal("owned witness setup failed earlier in this process")
	}
}

func ownedWitnessExecute(test *testing.T, ctx context.Context, name string, args ...string) execution {
	test.Helper()
	command := lintCommand(ctx, "", name, args...)
	output, err := os.CreateTemp(test.TempDir(), "stdout-")
	if err != nil {
		test.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = command.Run()
	if ctx.Err() != nil {
		test.Fatalf("cooked: shard exceeded 90s: %v", ctx.Err())
	}
	if err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		test.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		test.Fatal(err)
	}
	return execution{data, time.Since(started)}
}

func ownedWitnessRecoveryRows(test *testing.T, ctx context.Context, oracle string, rows []string) []string {
	test.Helper()
	if len(rows) == 0 {
		return rows
	}
	answer := ownedWitnessExecute(test, ctx, oracle, "--manifest", manifest(test, rows), "--diagnostics")
	flags := strings.Fields(string(answer.output))
	if len(flags) != len(rows) {
		test.Fatalf("diagnostics answered %d rows of %d", len(flags), len(rows))
	}
	result := append([]string(nil), rows...)
	for index, row := range rows {
		if flags[index] != "1" {
			continue
		}
		fields := strings.Split(row, "\t")
		for len(fields) < 7 {
			fields = append(fields, "")
		}
		if fields[6] == "" {
			fields[6] = "recovery"
		}
		result[index] = strings.Join(fields, "\t")
	}
	return result
}

func ownedWitnessAgreement(id, side string, got, want []byte) error {
	if diff := difference(got, want); diff != "" {
		return fmt.Errorf("%s %s: %s", id, side, diff)
	}
	return nil
}

func TestOwnedWitnessesUnion(test *testing.T) {
	test.Parallel()
	cases := ownedWitnessCases(test, ".")
	plan, err := planOwnedWitnesses(cases, testOwnedWitnessesShards)
	if err != nil {
		test.Fatal(err)
	}
	if len(plan) != testOwnedWitnessesShards {
		test.Fatalf("enumerated %d shards, want %d", len(plan), testOwnedWitnessesShards)
	}
	missing := append([][]ownedWitnessCase(nil), plan...)
	for shard, group := range missing {
		if len(group) > 0 {
			missing[shard] = group[1:]
			break
		}
	}
	if checkOwnedWitnessUnion(cases, missing) == nil {
		test.Fatal("missing witness survived")
	}
	repeated := append([][]ownedWitnessCase(nil), plan...)
	repeated[0] = append(append([]ownedWitnessCase(nil), repeated[0]...), cases[0])
	if checkOwnedWitnessUnion(cases, repeated) == nil {
		test.Fatal("repeated witness survived")
	}
	test.Logf("live union=%d witnesses/%d manifest cases", len(cases), 2*len(cases))
}

func TestOwnedWitnessesPlantedDisagreement(test *testing.T) {
	test.Parallel()
	binary, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cases := ownedWitnessCases(test, ".")
	want := fmt.Sprintf("TestOwnedWitnesses_%03d", ownedWitnessShard(cases[0].id, testOwnedWitnessesShards))
	command := exec.CommandContext(ctx, binary, "-test.run=^"+want+"$", "-test.timeout=90s", "-test.v")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.Env = append(os.Environ(), "ADAMIC_OWNED_WITNESS_PROBE=1", "ADAMIC_TEST_SHARD=")
	// Read through childguard.Capture, to its end with no WaitDelay deciding when: exec's copying under a one-second
	// WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after the child exited, an
	// error that reads as the child's failure here whatever it exited with, beside output cut short (#n3gk22h's shape).
	stdout, stderr, err := childguard.Capture(ctx, command)
	output := append(stdout, stderr...)
	if err == nil {
		test.Fatal("planted disagreement survived")
	}
	failures := 0
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "--- FAIL: TestOwnedWitnesses_") {
			failures++
			if !strings.Contains(line, want+" ") {
				test.Fatalf("wrong shard caught disagreement: %s", line)
			}
		}
	}
	if failures != 1 || !bytes.Contains(output, []byte("planted disagreement")) {
		test.Fatalf("want exactly %s to fail: %s", want, output)
	}
	test.Logf("planted native disagreement caught only by %s", want)
}

func TestOwnedWitnesses_000(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 0)
}
func TestOwnedWitnesses_001(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 1)
}
func TestOwnedWitnesses_002(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 2)
}
func TestOwnedWitnesses_003(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 3)
}
func TestOwnedWitnesses_004(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 4)
}
func TestOwnedWitnesses_005(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 5)
}
func TestOwnedWitnesses_006(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 6)
}
func TestOwnedWitnesses_007(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 7)
}
func TestOwnedWitnesses_008(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 8)
}
func TestOwnedWitnesses_009(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 9)
}
func TestOwnedWitnesses_010(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 10)
}
func TestOwnedWitnesses_011(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 11)
}
func TestOwnedWitnesses_012(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 12)
}
func TestOwnedWitnesses_013(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 13)
}
func TestOwnedWitnesses_014(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 14)
}
func TestOwnedWitnesses_015(test *testing.T) {
	test.Parallel()
	ownedWitnessUnit(test, 15)
}
