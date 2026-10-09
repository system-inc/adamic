package lint

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

const testOwnedWitnessesShards = 16

type ownedWitnessCase struct {
	id, row    string
	descriptor registry.Descriptor
}

func ownedWitnessCases(t *testing.T, directory string) []ownedWitnessCase {
	t.Helper()
	var cases []ownedWitnessCase
	for _, d := range prepareRegistry(t, directory) {
		paths, err := registry.Witnesses(filepath.Join(directory, "rules", d.Slug))
		if err != nil {
			t.Fatal(err)
		}
		rows := ownedWitnessRows(t, directory, d)
		if len(paths) != len(rows) {
			t.Fatalf("%s: %d paths, %d rows", d.Slug, len(paths), len(rows))
		}
		for index, row := range rows {
			relative, err := filepath.Rel(directory, paths[index])
			if err != nil {
				t.Fatal(err)
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

func TestOwnedWitnessesAssignmentStable(t *testing.T) {
	t.Parallel()
	cases := ownedWitnessCases(t, ".")
	original, err := planOwnedWitnesses(cases, testOwnedWitnessesShards)
	if err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}
	for shard, group := range plan {
		for _, witness := range group {
			if previous, exists := assignments[witness.id]; exists && previous != shard {
				t.Fatalf("%s moved from shard-%03d to shard-%03d", witness.id, previous, shard)
			}
		}
	}
	if _, err := planOwnedWitnesses(nil, testOwnedWitnessesShards); err == nil {
		t.Fatal("empty repository corpus survived")
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
func ownedWitnessUnit(t *testing.T, shard int) {
	t.Helper()
	if text := os.Getenv("ADAMIC_TEST_SHARD"); text != "" {
		parts := strings.Split(text, "/")
		if len(parts) != 2 {
			t.Fatal("ADAMIC_TEST_SHARD must be i/n")
		}
		i, e1 := strconv.Atoi(parts[0])
		n, e2 := strconv.Atoi(parts[1])
		if e1 != nil || e2 != nil || n < 1 || i < 0 || i >= n {
			t.Fatal("invalid ADAMIC_TEST_SHARD")
		}
		if shard%n != i {
			t.Skip("assigned to another box")
		}
	}
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	cases := ownedWitnessCases(t, directory)
	plan, err := planOwnedWitnesses(cases, testOwnedWitnessesShards)
	if err != nil {
		t.Fatal(err)
	}
	if os.Getenv("ADAMIC_OWNED_WITNESS_PROBE") == "1" {
		for _, witness := range plan[shard] {
			want := []byte("case 0\n")
			got := want
			if witness.id == cases[0].id {
				got = []byte("case 0\nplanted disagreement\n")
			}
			if err := ownedWitnessAgreement(witness.id, "native", got, want); err != nil {
				t.Fatal(err)
			}
		}
		return
	}
	// A shard selected alone prepares the shared inputs itself, before its own deadline starts.
	ownedWitnessSetup(t)
	oracle, binary, module := ownedWitnessShared.oracle, ownedWitnessShared.binary, ownedWitnessShared.module
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	started := time.Now()
	defer func() { t.Logf("case wall=%s; shard=%03d witnesses=%d", time.Since(started), shard, len(plan[shard])) }()
	var rows []string
	for _, witness := range plan[shard] {
		path := strings.SplitN(witness.row, "\t", 2)[0]
		pair := ownedWitnessRecoveryRows(t, ctx, oracle, []string{witness.row, path + "\tall"})
		answer := ownedWitnessExecute(t, ctx, oracle, "--manifest", manifest(t, pair[:1]), "--count")
		if string(answer.output) == "0\n" {
			t.Fatalf("%s witness reports no findings", witness.descriptor.Name)
		}
		rows = append(rows, pair...)
	}
	if len(rows) == 0 {
		return
	}
	path := manifest(t, rows)
	want := ownedWitnessExecute(t, ctx, oracle, "--manifest", path)
	for _, side := range []struct {
		name string
		run  execution
	}{
		{"Node", ownedWitnessExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), filepath.Join(directory, "main.ts"), "--manifest", path)},
		{"emitted JavaScript", ownedWitnessExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), module, "--manifest", path)},
		{"native", ownedWitnessExecute(t, ctx, binary, "--manifest", path)},
	} {
		if err := ownedWitnessAgreement(fmt.Sprintf("shard-%03d", shard), side.name, side.run.output, want.output); err != nil {
			t.Fatal(err)
		}
	}
}

var ownedWitnessShared struct {
	sync.Mutex
	oracle, binary, module string
}

// Not parallel: publishes the shared oracle and immutable products before parallel shards resume.
func TestOwnedWitnesses_Setup(t *testing.T) {
	if os.Getenv("ADAMIC_OWNED_WITNESS_PROBE") == "1" {
		return
	}
	ownedWitnessSetup(t)
}

func ownedWitnessSetup(t *testing.T) {
	t.Helper()
	ownedWitnessShared.Lock()
	defer ownedWitnessShared.Unlock()
	if ownedWitnessShared.oracle != "" {
		return
	}
	// No deadline of its own: setup is the build phase's to make fast, and Loom's 90 s kill still covers
	// the unit (@system_adamic's ruling). A shard's clock starts only after this returns.
	started := time.Now()
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	oracle := goOracle(t) // Overlay builds retain their existing builder by instruction.
	binary, module := ownedWitnessProducts(t, directory)
	ownedWitnessShared.oracle, ownedWitnessShared.binary, ownedWitnessShared.module = oracle, binary, module
	t.Logf("shared setup=%s", time.Since(started))
}

func ownedWitnessExecute(t *testing.T, ctx context.Context, name string, args ...string) execution {
	t.Helper()
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = command.Run()
	if ctx.Err() != nil {
		t.Fatalf("cooked: shard exceeded 90s: %v", ctx.Err())
	}
	if err != nil || len(commandDiagnostics(name, stderr.Bytes())) != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, time.Since(started)}
}

func ownedWitnessRecoveryRows(t *testing.T, ctx context.Context, oracle string, rows []string) []string {
	t.Helper()
	if len(rows) == 0 {
		return rows
	}
	answer := ownedWitnessExecute(t, ctx, oracle, "--manifest", manifest(t, rows), "--diagnostics")
	flags := strings.Fields(string(answer.output))
	if len(flags) != len(rows) {
		t.Fatalf("diagnostics answered %d rows of %d", len(flags), len(rows))
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

// Non-Go products include the port, compiler implementation and embedded runtime.
// Go oracle overlay builds retain their existing builder until GoBuild supports them.
func ownedWitnessProducts(t *testing.T, directory string) (string, string) {
	t.Helper()
	files := []string{"go.mod", "cohere/go.mod", "cohere/go.sum", "stage1/cohere/lint/owned_witness_units_test.go"}
	for _, root := range []string{"internal", "stage1/typescript", "cohere/rule_runner", "cohere/schema", "cohere/internal", "cohere/TypeScript-shim", "cohere/TypeScript/tsc"} {
		err := filepath.WalkDir(filepath.Join(repository, root), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				if entry.Name() == ".git" {
					return filepath.SkipDir
				}
				return nil
			}
			ext := filepath.Ext(path)
			if (ext == ".go" && !strings.HasSuffix(path, "_test.go")) || ext == ".c" || ext == ".h" || ext == ".ts" || ext == ".a" || entry.Name() == "go.mod" || entry.Name() == "go.sum" {
				relative, err := filepath.Rel(repository, path)
				if err != nil {
					return err
				}
				files = append(files, filepath.ToSlash(relative))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range portFiles(t) {
		files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", file)))
	}
	tools := []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}
	lowered := buildcache.Product(t, buildcache.Inputs{Name: "lint-owned-witnesses-lowered", Files: files, Toolchain: tools}, func(dir string) error {
		started := time.Now()
		program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(ir)), 0644); err != nil {
			return err
		}
		err = os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(ir)), 0644)
		t.Logf("cold build lowered=%s", time.Since(started))
		return err
	})
	flags := append(native.Flags(native.Options{Sanitize: true, Split: true, Jobs: 4}), "Split=true", "Jobs=4", "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED="+os.Getenv("ADAMIC_GATE_UNCACHED"))
	tools = append(tools, buildcache.Tool("clang", "--version"))
	binary := buildcache.Product(t, buildcache.Inputs{Name: "lint-owned-witnesses-sanitized", Files: files, Flags: flags, Toolchain: tools}, func(dir string) error {
		source, err := os.ReadFile(filepath.Join(lowered, "program.c"))
		if err != nil {
			return err
		}
		started := time.Now()
		err = native.Build(string(source), filepath.Join(dir, "native"), native.Options{Sanitize: true, Split: true, Jobs: 4})
		t.Logf("cold build sanitized=%s", time.Since(started))
		return err
	})
	return filepath.Join(binary, "native"), filepath.Join(lowered, "program.mjs")
}

func TestOwnedWitnessesUnion(t *testing.T) {
	t.Parallel()
	cases := ownedWitnessCases(t, ".")
	plan, err := planOwnedWitnesses(cases, testOwnedWitnessesShards)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan) != testOwnedWitnessesShards {
		t.Fatalf("enumerated %d shards, want %d", len(plan), testOwnedWitnessesShards)
	}
	missing := append([][]ownedWitnessCase(nil), plan...)
	for shard, group := range missing {
		if len(group) > 0 {
			missing[shard] = group[1:]
			break
		}
	}
	if checkOwnedWitnessUnion(cases, missing) == nil {
		t.Fatal("missing witness survived")
	}
	repeated := append([][]ownedWitnessCase(nil), plan...)
	repeated[0] = append(append([]ownedWitnessCase(nil), repeated[0]...), cases[0])
	if checkOwnedWitnessUnion(cases, repeated) == nil {
		t.Fatal("repeated witness survived")
	}
	t.Logf("live union=%d witnesses/%d manifest cases", len(cases), 2*len(cases))
}

func TestOwnedWitnessesPlantedDisagreement(t *testing.T) {
	t.Parallel()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "-test.run=^TestOwnedWitnesses_[0-9]{3}$", "-test.timeout=90s", "-test.v")
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	command.Env = append(os.Environ(), "ADAMIC_OWNED_WITNESS_PROBE=1", "ADAMIC_TEST_SHARD=")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("planted disagreement survived")
	}
	cases := ownedWitnessCases(t, ".")
	want := fmt.Sprintf("TestOwnedWitnesses_%03d", ownedWitnessShard(cases[0].id, testOwnedWitnessesShards))
	failures := 0
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, "--- FAIL: TestOwnedWitnesses_") {
			failures++
			if !strings.Contains(line, want+" ") {
				t.Fatalf("wrong shard caught disagreement: %s", line)
			}
		}
	}
	if failures != 1 || !bytes.Contains(output, []byte("planted disagreement")) {
		t.Fatalf("want exactly %s to fail: %s", want, output)
	}
	t.Logf("planted native disagreement caught only by %s", want)
}

func TestOwnedWitnesses_000(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 0)
}
func TestOwnedWitnesses_001(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 1)
}
func TestOwnedWitnesses_002(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 2)
}
func TestOwnedWitnesses_003(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 3)
}
func TestOwnedWitnesses_004(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 4)
}
func TestOwnedWitnesses_005(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 5)
}
func TestOwnedWitnesses_006(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 6)
}
func TestOwnedWitnesses_007(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 7)
}
func TestOwnedWitnesses_008(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 8)
}
func TestOwnedWitnesses_009(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 9)
}
func TestOwnedWitnesses_010(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 10)
}
func TestOwnedWitnesses_011(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 11)
}
func TestOwnedWitnesses_012(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 12)
}
func TestOwnedWitnesses_013(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 13)
}
func TestOwnedWitnesses_014(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 14)
}
func TestOwnedWitnesses_015(t *testing.T) {
	t.Parallel()
	ownedWitnessUnit(t, 15)
}
