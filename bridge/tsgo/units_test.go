package tsgo_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
	"unicode/utf16"
)

type bridgeCase struct {
	name, piece, file string
	round             int
	check             func(*bridgeRun, string, int)
}

func bridgeShard() (int, int, error) {
	value := os.Getenv("ADAMIC_TEST_SHARD")
	if value == "" {
		return 0, 1, nil
	}
	parts := strings.Split(value, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("ADAMIC_TEST_SHARD must be i/n")
	}
	index, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	count, err := strconv.Atoi(parts[1])
	if err != nil || count < 1 || index < 0 || index >= count {
		return 0, 0, fmt.Errorf("ADAMIC_TEST_SHARD must satisfy 0 <= i < n")
	}
	return index, count, nil
}

func bridgeFileActive(file string) bool {
	return file == "" || (file == "sample.ts") == (os.Getenv("ADAMIC_TSGO_CORPUS") == "")
}
func bridgeCaseActive(index int) bool {
	shard, count, err := bridgeShard()
	return err == nil && bridgeFileActive(bridgeCases[index].file) && index%count == shard
}

// Not parallel: bridge subprocess timing and sanitizer probes share the machine.
// Each root is independently selectable; corpus pieces also have one shard owner.
func runBridgeCase(t *testing.T, index int) {
	if _, _, err := bridgeShard(); err != nil {
		t.Fatal(err)
	}
	if !bridgeCaseActive(index) {
		t.Skip("piece belongs to another corpus mode or shard")
	}
	// The subprocesses get a hang guard, not a budget: a loaded gate box can be far slower than
	// the reference box, and only a stuck process should fail here.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	run := &bridgeRun{t: t, repository: repository, scratch: t.TempDir(), ctx: ctx}
	unit := bridgeCases[index]
	for _, name := range bridgeCaseProducts(unit.piece) {
		run.product(name)
	}

	// A unit is held to the 60-second budget where it's measured: on the reference box (one Codex
	// instance, 4 CPUs, cold), which sets ADAMIC_UNIT_BUDGET=1. Elsewhere a loaded machine only
	// logs it, so the gate's correctness verdict never depends on its load.
	begun := time.Now()
	t.Cleanup(func() {
		if elapsed := time.Since(begun); elapsed >= 60*time.Second {
			if os.Getenv("ADAMIC_UNIT_BUDGET") == "1" {
				t.Errorf("bridge unit exceeded 60 seconds: %s", elapsed)
			} else {
				t.Logf("bridge unit took %s, over the 60-second budget measured on the reference box", elapsed)
			}
		}
	})
	unit.check(run, unit.file, unit.round)
}

type bridgeRun struct {
	t                   *testing.T
	repository, scratch string
	ctx                 context.Context
}

func (r *bridgeRun) product(name string) string {
	r.t.Helper()
	return bridgeProduct(r.t, r.repository, name)
}
func (r *bridgeRun) sample() string {
	return filepath.Join(r.repository, "bridge/tsgo/testdata/sample.ts")
}
func (r *bridgeRun) config() string {
	return filepath.Join(r.repository, "bridge/tsgo/testdata/tsconfig.json")
}
func (r *bridgeRun) command(binary string, arguments ...string) *exec.Cmd {
	command := exec.CommandContext(r.ctx, binary, arguments...)
	command.Dir = r.repository
	return command
}
func (r *bridgeRun) run(name string, command *exec.Cmd) ([]byte, error) {
	r.t.Helper()
	output, err := command.CombinedOutput()
	if write := os.WriteFile(filepath.Join(r.scratch, name+".log"), output, 0o644); write != nil {
		r.t.Fatal(write)
	}
	if r.ctx.Err() != nil {
		r.t.Fatalf("bridge subprocess exceeded five minutes: %v", r.ctx.Err())
	}
	return output, err
}
func (r *bridgeRun) mustRun(name string, command *exec.Cmd) []byte {
	r.t.Helper()
	output, err := r.run(name, command)
	if err != nil {
		r.t.Fatalf("%s: %v\n%s", name, err, output)
	}
	return output
}
func bridgeRoots(repository string) []string {
	if corpus := os.Getenv("ADAMIC_TSGO_CORPUS"); corpus != "" {
		roots := []string{}
		for _, file := range []string{"checker.ts", "parser.ts", "types.ts", "utilities.ts"} {
			roots = append(roots, filepath.Join(corpus, "src/compiler", file))
		}
		return roots
	}
	return []string{filepath.Join(repository, "bridge/tsgo/testdata/sample.ts")}
}
func bridgePositions(text []byte) []int {
	count := min(400, len(text))
	positions := make([]int, count)
	for index := range positions {
		positions[index] = index * len(text) / count
	}
	return positions
}
func (r *bridgeRun) arguments(file string) []string {
	r.t.Helper()
	roots := bridgeRoots(r.repository)
	var queries strings.Builder
	count := 0
	for _, path := range roots {
		if filepath.Base(path) != file {
			continue
		}
		text, err := os.ReadFile(path)
		if err != nil {
			r.t.Fatal(err)
		}
		for _, position := range bridgePositions(text) {
			fmt.Fprintf(&queries, "%s\t%d\n", path, position)
			count++
		}
	}
	if count == 0 {
		r.t.Fatal("no queries in the active corpus piece")
	}
	manifest := filepath.Join(r.scratch, "queries.tsv")
	if err := os.WriteFile(manifest, []byte(queries.String()), 0o644); err != nil {
		r.t.Fatal(err)
	}
	r.t.Logf("queries: %s, %d positions, %d unchanged program roots", file, count, len(roots))
	return append([]string{r.config(), manifest}, roots...)
}
func (r *bridgeRun) answers(name, binary string, arguments []string, timing bool) []byte {
	r.t.Helper()
	command := r.command(binary, arguments...)
	if timing {
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdout, err := command.Output()
	if write := os.WriteFile(filepath.Join(r.scratch, name+".log"), stderr.Bytes(), 0o644); write != nil {
		r.t.Fatal(write)
	}
	if err != nil {
		r.t.Fatalf("%s: %v\n%s", name, err, stderr.String())
	}
	if timing {
		r.t.Logf("%s: %s", name, strings.TrimSpace(stderr.String()))
	} else if stderr.Len() != 0 {
		r.t.Fatalf("%s sanitizer stderr: %s", name, stderr.String())
	}
	return stdout
}
func bridgeABI(r *bridgeRun, _ string, _ int) {
	output := r.mustRun("api", r.command(r.product("api"), r.config(), r.sample(), "check"))
	r.t.Logf("C ABI: 100 queries, owned outputs, handles and argument checks; %s", output)
}
func bridgeExpectedFailure(r *bridgeRun, name, driver, mode, message string) {
	r.t.Helper()
	output, err := r.run(name, r.command(r.product(driver), r.config(), r.sample(), mode))
	if err == nil || !bytes.Contains(output, []byte(message)) {
		r.t.Fatalf("%s escaped: %v\n%s", name, err, output)
	}
	r.t.Logf("%s: %s", name, message)
}
func bridgeInputLength(r *bridgeRun, _ string, _ int) {
	bridgeExpectedFailure(r, "input-length", "api", "length", "AddressSanitizer: heap-buffer-overflow")
}
func bridgeOutputLength(r *bridgeRun, _ string, _ int) {
	bridgeExpectedFailure(r, "output-length", "length-driver", "check", "AddressSanitizer: heap-buffer-overflow")
}
func bridgeStale(r *bridgeRun, _ string, _ int) {
	bridgeExpectedFailure(r, "stale-handle", "stale-driver", "check", "tsgo_query(handle, file, 14, &answer, &error) == TSGO_HANDLE")
}
func bridgeOutputFree(r *bridgeRun, _ string, _ int) {
	bridgeExpectedFailure(r, "output-free", "leak-driver", "check", "LeakSanitizer: detected memory leaks")
}
func bridgeUnlinked(r *bridgeRun, mode string, _ int) {
	arguments := []string{mode, filepath.Join(r.repository, "bridge/tsgo/testdata/queries.a")}
	if mode == "build" {
		arguments = append(arguments, "-o", filepath.Join(r.scratch, "unlinked"))
	}
	output, err := r.run("refusal-"+mode, r.command(r.product("stage0"), arguments...))
	if err == nil || !bytes.Contains(output, []byte("unlinked typescript-go library call")) {
		r.t.Fatalf("unlinked %s was not refused: %v\n%s", mode, err, output)
	}
	r.t.Logf("%s refuses unlinked checker calls", mode)
}
func bridgeUnlinkedBuild(r *bridgeRun, _ string, _ int)      { bridgeUnlinked(r, "build", 0) }
func bridgeUnlinkedC(r *bridgeRun, _ string, _ int)          { bridgeUnlinked(r, "c", 0) }
func bridgeUnlinkedJavaScript(r *bridgeRun, _ string, _ int) { bridgeUnlinked(r, "js", 0) }
func bridgeOracle(r *bridgeRun, file string, _ int) {
	arguments := r.arguments(file)
	truth := r.answers("go", r.product("oracle"), arguments, true)
	observed := r.answers("native-asan", r.product("native-asan"), arguments, false)
	if !bytes.Equal(truth, observed) {
		r.t.Fatalf("native oracle mismatch at byte %d", firstDifference(truth, observed))
	}
	r.t.Logf("oracle: %s, %d bytes identical under ASan/UBSan/LSan", file, len(truth))
}
func bridgeTiming(r *bridgeRun, file string, round int) {
	arguments := r.arguments(file)
	truth := r.answers("go-baseline", r.product("oracle"), arguments, true)
	observed := r.answers(fmt.Sprintf("native-round-%d", round), r.product("native"), arguments, true)
	direct := r.answers(fmt.Sprintf("go-round-%d", round), r.product("oracle"), arguments, true)
	if !bytes.Equal(truth, observed) || !bytes.Equal(truth, direct) {
		r.t.Fatal("timed answers changed")
	}
}
func bridgeWrongPosition(r *bridgeRun, file string, _ int) {
	arguments := r.arguments(file)
	truth := r.answers("go", r.product("oracle"), arguments, true)
	observed := r.answers("wrong-position", r.product("wrong-native"), arguments, false)
	if bytes.Equal(truth, observed) {
		r.t.Fatal("wrong-position type string escaped oracle")
	}
	r.t.Logf("wrong-position: oracle mismatch at byte %d", firstDifference(truth, observed))
}
func bridgeLinkage(r *bridgeRun, _ string, _ int) {
	command := r.command(r.product("linkage.test"), "-test.run=^TestTSGoRequiresLink$", "-test.count=1", "-test.v", "-test.timeout=5m")
	command.Dir = filepath.Join(r.repository, "bridge/tsgo")
	output, err := r.run("linkage-mutant", command)
	if err == nil || !bytes.Contains(output, []byte("lowering accepted an unlinked checker call")) {
		r.t.Fatalf("linkage mutant escaped refusal test: %v\n%s", err, output)
	}
	r.t.Log("linkage mutant: refusal test catches the missing guard")
}
func bridgeRegion(r *bridgeRun, _ string, _ int) {
	manifest := filepath.Join(r.scratch, "region.tsv")
	if err := os.WriteFile(manifest, []byte(r.sample()+"\t14\n"), 0o644); err != nil {
		r.t.Fatal(err)
	}
	arguments := []string{r.config(), manifest, r.sample()}
	r.mustRun("region-oracle", r.command(r.product("oracle"), arguments...))
	truth := r.answers("region-frame", r.product("oracle"), arguments, true)
	frame := bytes.SplitN(truth, []byte("\n"), 2)
	var kind, symbolLength, typeLength int
	if len(frame) != 2 {
		r.t.Fatal("invalid region oracle frame")
	}
	if _, err := fmt.Sscanf(string(frame[0]), "%d %d %d", &kind, &symbolLength, &typeLength); err != nil {
		r.t.Fatal(err)
	}
	if symbolLength+typeLength > len(frame[1]) {
		r.t.Fatal("short region oracle frame")
	}
	expected := fmt.Sprintf("%d\n", len(utf16.Encode([]rune(string(frame[1][:symbolLength+typeLength])))))
	command := r.command(r.product("healthy-region"), r.config(), r.sample())
	var report bytes.Buffer
	command.Stderr = &report
	answer, err := command.Output()
	if err != nil || string(answer) != expected || !strings.HasSuffix(report.String(), " regions 1\n") {
		r.t.Fatalf("region probe did not use one region and agree with Go: %v stdout=%q stderr=%s", err, answer, report.String())
	}
	r.t.Logf("region probe: Go and native answer %s; %s", strings.TrimSpace(expected), strings.TrimSpace(report.String()))
}
func bridgeRegionOwnership(r *bridgeRun, _ string, _ int) {
	output, err := r.run("region-ownership", r.command(r.product("region-native"), r.config(), r.sample()))
	if err == nil || !bytes.Contains(output, []byte("LeakSanitizer: detected memory leaks")) {
		r.t.Fatalf("heap result in region variant escaped LSan: %v\n%s", err, output)
	}
	r.t.Log("region ownership: LeakSanitizer catches the unowned heap result")
}
