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
	index, parseError := strconv.Atoi(parts[0])
	if parseError != nil {
		return 0, 0, parseError
	}
	count, parseError := strconv.Atoi(parts[1])
	if parseError != nil || count < 1 || index < 0 || index >= count {
		return 0, 0, fmt.Errorf("ADAMIC_TEST_SHARD must satisfy 0 <= i < n")
	}
	return index, count, nil
}

func bridgeFileActive(file string) bool {
	return file == "" || (file == "sample.ts") == (os.Getenv("ADAMIC_TSGO_CORPUS") == "")
}
func bridgeCaseActive(index int) bool {
	shard, count, bridgeShardError := bridgeShard()
	return bridgeShardError == nil && bridgeFileActive(bridgeCases[index].file) && index%count == shard
}

// Not parallel: bridge subprocess timing and sanitizer probes share the machine.
// Each root is independently selectable; corpus pieces also have one shard owner.
func runBridgeCase(test *testing.T, index int) {
	if _, _, bridgeShardError := bridgeShard(); bridgeShardError != nil {
		test.Fatal(bridgeShardError)
	}
	if !bridgeCaseActive(index) {
		test.Skip("piece belongs to another corpus mode or shard")
	}
	// The subprocesses get a hang guard, not a budget: a loaded gate box can be far slower than
	// the reference box, and only a stuck process should fail here.
	timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	repository, absolutePathError := filepath.Abs("../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	run := &bridgeRun{t: test, repository: repository, scratch: test.TempDir(), ctx: timeoutContext}
	unit := bridgeCases[index]
	for _, name := range bridgeCaseProducts(unit.piece) {
		run.product(name)
	}

	// A unit is held to the 60-second budget where it's measured: on the reference box (one Codex
	// instance, 4 CPUs, cold), which sets ADAMIC_UNIT_BUDGET=1. Elsewhere a loaded machine only
	// logs it, so the gate's correctness verdict never depends on its load.
	begun := time.Now()
	test.Cleanup(func() {
		if elapsed := time.Since(begun); elapsed >= 60*time.Second {
			if os.Getenv("ADAMIC_UNIT_BUDGET") == "1" {
				test.Errorf("bridge unit exceeded 60 seconds: %s", elapsed)
			} else {
				test.Logf("bridge unit took %s, over the 60-second budget measured on the reference box", elapsed)
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

func (run *bridgeRun) product(name string) string {
	run.t.Helper()
	return bridgeProduct(run.t, run.repository, name)
}
func (run *bridgeRun) sample() string {
	return filepath.Join(run.repository, "bridge/tsgo/testdata/sample.ts")
}
func (run *bridgeRun) config() string {
	return filepath.Join(run.repository, "bridge/tsgo/testdata/tsconfig.json")
}
func (run *bridgeRun) command(binary string, arguments ...string) *exec.Cmd {
	command := exec.CommandContext(run.ctx, binary, arguments...)
	command.Dir = run.repository
	return command
}
func (run *bridgeRun) run(name string, command *exec.Cmd) ([]byte, error) {
	run.t.Helper()
	output, combinedOutputError := command.CombinedOutput()
	if write := os.WriteFile(filepath.Join(run.scratch, name+".log"), output, 0o644); write != nil {
		run.t.Fatal(write)
	}
	if run.ctx.Err() != nil {
		run.t.Fatalf("bridge subprocess exceeded five minutes: %v", run.ctx.Err())
	}
	return output, combinedOutputError
}
func (run *bridgeRun) mustRun(name string, command *exec.Cmd) []byte {
	run.t.Helper()
	output, runError := run.run(name, command)
	if runError != nil {
		run.t.Fatalf("%s: %v\n%s", name, runError, output)
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
func (run *bridgeRun) arguments(file string) []string {
	run.t.Helper()
	roots := bridgeRoots(run.repository)
	var queries strings.Builder
	count := 0
	for _, path := range roots {
		if filepath.Base(path) != file {
			continue
		}
		text, readError := os.ReadFile(path)
		if readError != nil {
			run.t.Fatal(readError)
		}
		for _, position := range bridgePositions(text) {
			fmt.Fprintf(&queries, "%s\t%d\n", path, position)
			count++
		}
	}
	if count == 0 {
		run.t.Fatal("no queries in the active corpus piece")
	}
	manifest := filepath.Join(run.scratch, "queries.tsv")
	if writeError := os.WriteFile(manifest, []byte(queries.String()), 0o644); writeError != nil {
		run.t.Fatal(writeError)
	}
	run.t.Logf("queries: %s, %d positions, %d unchanged program roots", file, count, len(roots))
	return append([]string{run.config(), manifest}, roots...)
}
func (run *bridgeRun) answers(name, binary string, arguments []string, timing bool) []byte {
	run.t.Helper()
	command := run.command(binary, arguments...)
	if timing {
		command.Env = append(os.Environ(), "ADAMIC_TSGO_TIMING=1")
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	stdout, outputError := command.Output()
	if write := os.WriteFile(filepath.Join(run.scratch, name+".log"), stderr.Bytes(), 0o644); write != nil {
		run.t.Fatal(write)
	}
	if outputError != nil {
		run.t.Fatalf("%s: %v\n%s", name, outputError, stderr.String())
	}
	if timing {
		run.t.Logf("%s: %s", name, strings.TrimSpace(stderr.String()))
	} else if stderr.Len() != 0 {
		run.t.Fatalf("%s sanitizer stderr: %s", name, stderr.String())
	}
	return stdout
}
func bridgeABI(r *bridgeRun, _ string, _ int) {
	output := r.mustRun("api", r.command(r.product("api"), r.config(), r.sample(), "check"))
	r.t.Logf("C ABI: 100 queries, owned outputs, handles and argument checks; %s", output)
}
func bridgeExpectedFailure(r *bridgeRun, name, driver, mode, message string) {
	r.t.Helper()
	output, runError := r.run(name, r.command(r.product(driver), r.config(), r.sample(), mode))
	if runError == nil || !bytes.Contains(output, []byte(message)) {
		r.t.Fatalf("%s escaped: %v\n%s", name, runError, output)
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
	bridgeExpectedFailure(
		r,
		"stale-handle",
		"stale-driver",
		"check",
		"tsgo_query(handle, file, 14, &answer, &error) == TSGO_HANDLE",
	)
}
func bridgeOutputFree(r *bridgeRun, _ string, _ int) {
	bridgeExpectedFailure(r, "output-free", "leak-driver", "check", "LeakSanitizer: detected memory leaks")
}
func bridgeUnlinked(r *bridgeRun, mode string, _ int) {
	arguments := []string{mode, filepath.Join(r.repository, "bridge/tsgo/testdata/queries.a")}
	if mode == "build" {
		arguments = append(arguments, "-o", filepath.Join(r.scratch, "unlinked"))
	}
	output, runError := r.run("refusal-"+mode, r.command(r.product("stage0"), arguments...))
	if runError == nil || !bytes.Contains(output, []byte("unlinked typescript-go library call")) {
		r.t.Fatalf("unlinked %s was not refused: %v\n%s", mode, runError, output)
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
	command := r.command(
		r.product("linkage.test"),
		"-test.run=^TestTSGoRequiresLink$",
		"-test.count=1",
		"-test.v",
		"-test.timeout=5m",
	)
	command.Dir = filepath.Join(r.repository, "bridge/tsgo")
	output, runError := r.run("linkage-mutant", command)
	if runError == nil || !bytes.Contains(output, []byte("lowering accepted an unlinked checker call")) {
		r.t.Fatalf("linkage mutant escaped refusal test: %v\n%s", runError, output)
	}
	r.t.Log("linkage mutant: refusal test catches the missing guard")
}
func bridgeRegion(r *bridgeRun, _ string, _ int) {
	manifest := filepath.Join(r.scratch, "region.tsv")
	if writeError := os.WriteFile(manifest, []byte(r.sample()+"\t14\n"), 0o644); writeError != nil {
		r.t.Fatal(writeError)
	}
	arguments := []string{r.config(), manifest, r.sample()}
	r.mustRun("region-oracle", r.command(r.product("oracle"), arguments...))
	truth := r.answers("region-frame", r.product("oracle"), arguments, true)
	frame := bytes.SplitN(truth, []byte("\n"), 2)
	var kind, symbolLength, typeLength int
	if len(frame) != 2 {
		r.t.Fatal("invalid region oracle frame")
	}
	if _, scanError := fmt.Sscanf(string(frame[0]), "%d %d %d", &kind, &symbolLength, &typeLength); scanError != nil {
		r.t.Fatal(scanError)
	}
	if symbolLength+typeLength > len(frame[1]) {
		r.t.Fatal("short region oracle frame")
	}
	expected := fmt.Sprintf("%d\n", len(utf16.Encode([]rune(string(frame[1][:symbolLength+typeLength])))))
	command := r.command(r.product("healthy-region"), r.config(), r.sample())
	var report bytes.Buffer
	command.Stderr = &report
	answer, outputError := command.Output()
	if outputError != nil || string(answer) != expected || !strings.HasSuffix(report.String(), " regions 1\n") {
		r.t.Fatalf(
			"region probe did not use one region and agree with Go: %v stdout=%q stderr=%s",
			outputError,
			answer,
			report.String(),
		)
	}
	r.t.Logf(
		"region probe: Go and native answer %s; %s",
		strings.TrimSpace(expected),
		strings.TrimSpace(report.String()),
	)
}
func bridgeRegionOwnership(r *bridgeRun, _ string, _ int) {
	command := r.command(r.product("region-native"), r.config(), r.sample())
	// LeakSanitizer counts a stale copy of the leaked pointer, left on a stack or in a register, as a root, and then
	// misses the leak. Whether one survives depends on the code the native links: with its archive's cgo built by clang
	// rather than gcc (toolchain.Env), the default options missed it in 4 of 10 runs on Workshop. Without those roots,
	// only globals, the heap and thread-local storage reach the result, so the unowned heap result is found every run.
	options := "LSAN_OPTIONS=use_stacks=0:use_registers=0"
	command.Env = append(os.Environ(), options)
	// The control: without those roots LeakSanitizer also reports memory a live frame still holds at exit, so the same
	// options must find nothing in healthy-region, the owned variant of this program, sanitized too. Then the report
	// below is the unowned result's, not the options' (Markup's review).
	control := r.command(r.product("healthy-region"), r.config(), r.sample())
	control.Env = append(os.Environ(), options)
	if output, runError := r.run(
		"region-ownership-control",
		control,
	); runError != nil || bytes.Contains(output, []byte("LeakSanitizer")) {
		r.t.Fatalf("owned region variant reported under %s: %v\n%s", options, runError, output)
	}
	output, runError := r.run("region-ownership", command)
	if runError == nil || !bytes.Contains(output, []byte("LeakSanitizer: detected memory leaks")) {
		r.t.Fatalf("heap result in region variant escaped LSan: %v\n%s", runError, output)
	}
	r.t.Log("region ownership: LeakSanitizer catches the unowned heap result")
}
