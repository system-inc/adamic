package estree

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

type recoverySetup struct {
	start  time.Time
	builds time.Duration
}

func beginRecoverySetup(t *testing.T) *recoverySetup {
	s := &recoverySetup{start: time.Now()}
	cpu := recoveryCPU(t)
	t.Cleanup(func() { t.Logf("TEST totalCPU=%.6fs", (recoveryCPU(t) - cpu).Seconds()) })
	return s
}
func (s *recoverySetup) report(t *testing.T) {
	wall := time.Since(s.start)
	t.Logf("setup wall=%.6fs builds=%.6fs without-builds=%.6fs", wall.Seconds(), s.builds.Seconds(), (wall - s.builds).Seconds())
}
func recoveryProduct(t *testing.T, s *recoverySetup, in buildcache.Inputs, build func(string) error) string {
	t.Helper()
	start := time.Now()
	dir := buildcache.Product(t, in, func(dir string) error {
		cold := time.Now()
		err := build(dir)
		t.Logf("BUILD %s cold wall=%.6fs", in.Name, time.Since(cold).Seconds())
		return err
	})
	s.builds += time.Since(start)
	return dir
}

// Overlay Go builds stay local until an oracle without an overlay can use GoBuild.
// Their inputs are not hand-listed as a Product key.
func recoveryLocalGoBuild(t *testing.T, s *recoverySetup, build func(string) error) string {
	t.Helper()
	dir := t.TempDir()
	start := time.Now()
	err := build(dir)
	elapsed := time.Since(start)
	s.builds += elapsed
	t.Logf("BUILD Go oracle cold wall=%.6fs", elapsed.Seconds())
	if err != nil {
		t.Fatal(err)
	}
	return dir
}
func recoveryBuild(t *testing.T, path string, sanitize bool) (string, string) {
	t.Helper()
	if !sanitize {
		t.Fatal("recovery native products require sanitizers")
	}
	setup := beginRecoverySetup(t)
	defer setup.report(t)
	return recoveryPort(t, setup, path)
}
func recoveryOracle(t *testing.T, s *recoverySetup) string {
	t.Helper()
	repo := root(t)
	source, err := filepath.Abs("testdata/oracle.go")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(repo, "cohere/adamic_estree_oracle.go")
	dir := recoveryLocalGoBuild(t, s, func(dir string) error {
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "overlay.json")
		if err = os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		cmd := exec.Command("go", "build", "-overlay="+path, "-o", filepath.Join(dir, "oracle"), virtual)
		cmd.Dir = filepath.Join(repo, "cohere")
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("oracle build: %w: %s", err, output)
		}
		if len(output) != 0 {
			return fmt.Errorf("oracle build output: %s", output)
		}
		return nil
	})
	return filepath.Join(dir, "oracle")
}
func recoveryLowered(t *testing.T, s *recoverySetup, path string) string {
	t.Helper()
	inputs := recoveryCacheInputs(t, path)
	return recoveryProduct(t, s, inputs, func(dir string) error {
		program, err := load.Load([]string{path})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		if err = os.WriteFile(filepath.Join(dir, "port.c"), []byte(native.C(ir)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "port.mjs"), []byte(javascript.JavaScript(ir)), 0644)
	})
}

func recoveryPort(t *testing.T, s *recoverySetup, path string) (string, string) {
	t.Helper()
	lowered := recoveryLowered(t, s, path)
	code, err := os.ReadFile(filepath.Join(lowered, "port.c"))
	if err != nil {
		t.Fatal(err)
	}
	inputs := buildcache.Inputs{
		Name:      "estree-sanitized-native",
		Files:     []string{"internal/native/runtime", "internal/native/native.go", "internal/native/library.go", "stage1/cohere/estree/recovery_cache_test.go"},
		Flags:     append(native.Flags(native.Options{Sanitize: true}), fmt.Sprintf("source-sha256=%x", sha256.Sum256(code)), "relative-source-names", "canonical-debug-prefix=/adamic-estree"),
		Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version"), buildcache.Tool("getconf", "GNU_LIBC_VERSION")},
	}
	binary := recoveryProduct(t, s, inputs, func(dir string) error { return recoveryStableNative(root(t), dir, string(code)) })
	return filepath.Join(binary, "port"), filepath.Join(lowered, "port.mjs")
}

type recoveryCase struct {
	id     string
	source string
}

func recoveryCases(prefix string, sources []string) []recoveryCase {
	out := make([]recoveryCase, len(sources))
	for i, source := range sources {
		out[i] = recoveryCase{fmt.Sprintf("%s/%03d", prefix, i), source}
	}
	return out
}

// Partition the entire enumeration before applying ADAMIC_TEST_SHARD. Verify
// both cardinality and exact IDs, including duplicate IDs in the original list.
func partitionRecovery(t *testing.T, cases []recoveryCase, count int) [][]recoveryCase {
	t.Helper()
	if count < 1 || count > len(cases) {
		t.Fatalf("declared shard count %d for %d cases", count, len(cases))
	}
	shards := make([][]recoveryCase, count)
	want := map[string]bool{}
	for i, c := range cases {
		if want[c.id] {
			t.Fatalf("repeated case ID %s", c.id)
		}
		want[c.id] = true
		shards[i%count] = append(shards[i%count], c)
	}
	seen := map[string]bool{}
	total := 0
	for _, shard := range shards {
		if len(shard) == 0 {
			t.Fatal("empty enumerated shard")
		}
		for _, c := range shard {
			if !want[c.id] || seen[c.id] {
				t.Fatalf("missing or repeated case ID %s", c.id)
			}
			seen[c.id] = true
			total++
		}
	}
	if len(shards) != count || total != len(cases) || len(seen) != len(want) {
		t.Fatalf("union mismatch: shards=%d/%d cases=%d/%d IDs=%d/%d", len(shards), count, total, len(cases), len(seen), len(want))
	}
	for id := range want {
		if !seen[id] {
			t.Fatalf("missing case ID %s", id)
		}
	}
	t.Logf("union: %d cases, %d unique IDs, %d shards", total, len(seen), len(shards))
	return shards
}
func recoveryShardSelection(t *testing.T) (int, int) {
	t.Helper()
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return 0, 1
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	i, e1 := strconv.Atoi(parts[0])
	n, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil || n < 1 || i < 0 || i >= n {
		t.Fatalf("invalid ADAMIC_TEST_SHARD %q", value)
	}
	return i, n
}
func runRecoveryShards(t *testing.T, shards [][]recoveryCase, run func(*testing.T, int, []recoveryCase)) {
	t.Helper()
	selected, total := recoveryShardSelection(t)
	for i, cases := range shards {
		if i%total != selected {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", i), func(t *testing.T) { t.Parallel(); run(t, i, cases) })
	}
}
func recoveryComparison(want, got []byte, mutant bool) string {
	diff := firstDifference(want, got)
	if mutant {
		if diff == "" {
			return "mutant survived"
		}
		return ""
	}
	return diff
}

func recoveryCount(got []byte, needle string, want int) string {
	count := strings.Count(string(got), needle)
	if count != want {
		return fmt.Sprintf("%q count=%d want=%d", needle, count, want)
	}
	return ""
}

type recoveryProof struct {
	IDs            []string
	Count, Planted int
	Mode           string
}

// The subprocess deliberately fails. Its parent requires exactly the owning
// shard to fail, using the same dispatcher and verdict checks as the real tests.
func proveRecoveryShard(t *testing.T, cases []recoveryCase, count int, mutant bool) {
	mode := "agreement"
	if mutant {
		mode = "mutant"
	}
	proveRecoveryVerdict(t, cases, count, mode)
}
func proveRecoveryVerdict(t *testing.T, cases []recoveryCase, count int, mode string) {
	t.Helper()
	partitionRecovery(t, cases, count)
	planted := len(cases)/2 + 1
	if planted >= len(cases) {
		planted = len(cases) - 1
	}
	config := recoveryProof{Count: count, Planted: planted, Mode: mode}
	for _, c := range cases {
		config.IDs = append(config.IDs, c.id)
	}
	data, err := json.Marshal(config)
	if err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(executable, "-test.v", "-test.run=^TestRecoveryShardProofChild$")
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "ADAMIC_TEST_SHARD=") && !strings.HasPrefix(entry, "ADAMIC_ESTREE_SHARD_PROOF=") {
			command.Env = append(command.Env, entry)
		}
	}
	command.Env = append(command.Env, "ADAMIC_ESTREE_SHARD_PROOF="+string(data))
	output, err := command.CombinedOutput()
	owner := fmt.Sprintf("shard-%03d", planted%count)
	prefix := "--- FAIL: TestRecoveryShardProofChild/"
	if err == nil || strings.Count(string(output), prefix) != 1 || !strings.Contains(string(output), prefix+owner+" ") {
		t.Fatalf("planted %s must fail only %s: %v\n%s", cases[planted].id, owner, err, output)
	}
	t.Logf("planted %s %s failed only %s", mode, cases[planted].id, owner)
}
func TestRecoveryShardProofChild(t *testing.T) {
	t.Parallel()
	data := os.Getenv("ADAMIC_ESTREE_SHARD_PROOF")
	if data == "" {
		return
	}
	var config recoveryProof
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		t.Fatal(err)
	}
	var cases []recoveryCase
	for _, id := range config.IDs {
		cases = append(cases, recoveryCase{id: id})
	}
	shards := partitionRecovery(t, cases, config.Count)
	planted := cases[config.Planted].id
	runRecoveryShards(t, shards, func(t *testing.T, i int, cases []recoveryCase) {
		var want, got strings.Builder
		for _, c := range cases {
			if config.Mode == "refusal" {
				script := "printf 'ESTree unattached decorator' >&2; exit 1"
				if c.id == planted {
					script = "printf 'accepted'"
				}
				refusedBeforeDeadline(t, []string{"/bin/sh", "-c", script}, "ESTree unattached decorator")
				continue
			}
			want.WriteString("0 Program " + c.id + "\n")
			bad := c.id == planted
			if config.Mode == "mutant" {
				bad = i != config.Planted%config.Count
			}
			if bad {
				got.WriteString("0 WrongProgram " + c.id + "\n")
			} else {
				got.WriteString("0 Program " + c.id + "\n")
			}
		}
		var failure string
		if config.Mode == "control" {
			failure = recoveryCount([]byte(got.String()), "0 Program ", len(cases))
		} else if config.Mode != "refusal" {
			failure = recoveryComparison([]byte(want.String()), []byte(got.String()), config.Mode == "mutant")
		}
		if failure != "" {
			t.Fatal(failure)
		}
	})
}

func recoveryCPU(t *testing.T) time.Duration {
	t.Helper()
	var self, children syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &self); err != nil {
		t.Fatal(err)
	}
	if err := syscall.Getrusage(syscall.RUSAGE_CHILDREN, &children); err != nil {
		t.Fatal(err)
	}
	seconds := self.Utime.Sec + self.Stime.Sec + children.Utime.Sec + children.Stime.Sec
	micros := self.Utime.Usec + self.Stime.Usec + children.Utime.Usec + children.Stime.Usec
	return time.Duration(seconds)*time.Second + time.Duration(micros)*time.Microsecond
}
