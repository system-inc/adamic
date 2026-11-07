package css

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// At 16 shards each ran 30 to 46 s in css#10/17 at load 45 to 51 (Eve and Adam, Oct 23): the Go oracle, the
// sanitized build, Node, the JavaScript backend and the leak run over a sixteenth of the corpus, one after another.
const testCompositionMatchesGoShards = 48

func compositionShard(key string) int {
	sum := sha256.Sum256([]byte(key))
	return int(binary.LittleEndian.Uint64(sum[:8]) % testCompositionMatchesGoShards)
}

func compositionPartition(test *testing.T, lines []string) [][]string {
	test.Helper()
	shards := make([][]string, testCompositionMatchesGoShards)
	seen := make([]int, len(lines))
	// Occurrences remain separate cases, including duplicate corpus entries.
	indices := make([][]int, testCompositionMatchesGoShards)
	for index, line := range lines {
		shard := compositionShard(line)
		shards[shard] = append(shards[shard], line)
		indices[shard] = append(indices[shard], index)
	}
	for _, shard := range indices {
		for _, index := range shard {
			seen[index]++
		}
	}
	for index, count := range seen {
		if count != 1 {
			test.Fatalf("case %d occurs %d times", index, count)
		}
	}
	if len(shards) != testCompositionMatchesGoShards {
		test.Fatal("shard enumeration differs from constant")
	}
	test.Logf("union: %d live cases, each exactly once across %d shards", len(lines), len(shards))
	return shards
}

type compositionProducts struct {
	lines                              []string
	shards                             [][]string
	source, product, sanitized, oracle string
}

var compositionOnce sync.Once
var compositionBuilt compositionProducts
var compositionFailed string

// compositionSetup builds the composition's products once per process. A build that fails inside the Once fails its
// caller; every later caller then fails naming the test that holds the real error, where each shard and the union read
// only "composition setup did not complete" before (#qbbmqfs).
func compositionSetup(test *testing.T) compositionProducts {
	test.Helper()
	compositionOnce.Do(func() {
		compositionFailed = test.Name()
		cssPlantedSetupFailure(test, "composition")
		compositionBuilt = buildCompositionProducts(test)
		compositionFailed = ""
	})
	if compositionFailed != "" {
		test.Fatalf(
			"the composition setup failed earlier in this process, in %s (see that test's log)",
			compositionFailed,
		)
	}
	return compositionBuilt
}
func buildCompositionProducts(test *testing.T) compositionProducts {
	started := time.Now()
	// Shared setup has no deadline; the shard starts its clock after setup returns.
	cases := compositionCases(test)
	data, readError := os.ReadFile(cases)
	if readError != nil {
		test.Fatal(readError)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	shards := compositionPartition(test, lines)
	oracle := oracles.CompositionOracle(
		test,
		"css-composition-go-oracle",
		"testdata/compose_side_test.go",
		"internal/format/css/adamic_compose_side_test.go",
		"./internal/format/css",
	)
	source, product := compositionLoweredProduct(test)
	sanitized := compositionNativeProduct(test)

	test.Logf("TestCompositionMatchesGo (setup): %.3fs", time.Since(started).Seconds())
	return compositionProducts{
		lines:     lines,
		shards:    shards,
		source:    source,
		product:   product,
		sanitized: sanitized,
		oracle:    oracle,
	}
}

func compositionRunShard(test *testing.T, index int) {
	// Prepare shared products once per process before starting this shard's clock.
	state := compositionSetup(test)
	started := time.Now()
	timer := time.AfterFunc(90*time.Second, func() { panic("cooked: composition shard exceeded 90s") })
	defer timer.Stop()
	shard := state.shards[index]
	dir := test.TempDir()
	cases := filepath.Join(dir, "cases.txt")
	answers := filepath.Join(dir, "answers.txt")
	if writeError := os.WriteFile(cases, []byte(strings.Join(shard, "\n")+"\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	request, _ := json.Marshal(map[string]string{"Cases": cases, "Answers": answers})
	requestPath := filepath.Join(dir, "request.json")
	if writeError := os.WriteFile(requestPath, request, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	result := compositionExecute(
		test,
		[]string{"ADAMIC_PORT_REQUEST=" + requestPath},
		state.oracle,
		"-test.run=^TestAdamicCompositionCases$",
		"-test.timeout=90s",
		"-test.v",
	)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		test.Fatalf("Go composition oracle: %d %s %s", result.exitCode, result.stdout, result.stderr)
	}
	data, readError := os.ReadFile(answers)
	if readError != nil {
		test.Fatal(readError)
	}
	if strings.Count(string(data), "\n") != len(shard)*2 {
		test.Fatal("Go answer count differs from live cases")
	}
	for _, side := range []struct {
		name   string
		result run
	}{
		{
			"native ASan/UBSan",
			compositionExecute(test, []string{"ASAN_OPTIONS=detect_leaks=0"}, state.sanitized, cases),
		},
		{"Node", compositionNode(test, state.source, cases)},
		{"JavaScript backend", compositionNode(test, filepath.Join(state.product, "program.mjs"), cases)},
	} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
			test.Fatalf("%s: %d %s", side.name, side.result.exitCode, side.result.stderr)
		}
		if difference := firstDifference(string(side.result.stdout), string(data)); difference != "" {
			test.Fatalf("%s: %s", side.name, difference)
		}
	}
	if report := compositionLeaks(test, state, cases); report != "" {
		test.Fatal(report)
	}
	test.Logf(
		"shard-%03d: %.3fs cooked=false cases=%d (excludes shared setup)",
		index,
		time.Since(started).Seconds(),
		len(shard),
	)
}

// Compiler products only: corpus enumeration and the Go oracle belong to the
// runtime comparison setup, so this unit needs no oracle-package ordering.
func TestCompositionMatchesGo_Setup(test *testing.T) {
	test.Parallel()
	compositionNativeProduct(test)
}

func TestCompositionMatchesGoUnion(test *testing.T) {
	test.Parallel()
	state := compositionSetup(test)
	started := time.Now()
	timer := time.AfterFunc(90*time.Second, func() { panic("cooked: composition union exceeded 90s") })
	defer timer.Stop()
	lines := state.lines
	if len(compositionShardFunctions) != testCompositionMatchesGoShards {
		test.Fatal("top-level shard enumeration differs from constant")
	}
	shards := compositionPartition(test, lines)
	// A single occurrence has a deliberately wrong answer. Run the same byte
	// comparator used by agreement on each partition, retaining duplicate cases.
	caught := 0
	owner := compositionShard(lines[0])
	offset := 0
	for shardIndex, shard := range shards {
		var got, want strings.Builder
		for j := range shard {
			fmt.Fprintf(&got, "case %d\nanswer\n", j)
			answer := "answer"
			if shardIndex == owner && j == 0 {
				answer = "planted disagreement"
			}
			fmt.Fprintf(&want, "case %d\n%s\n", j, answer)
			offset++
		}
		if firstDifference(got.String(), want.String()) != "" {
			caught++
			test.Logf("planted failure: one case disagrees; shard-%03d caught it", shardIndex)
		}
	}
	if caught != 1 || offset != len(lines) {
		test.Fatalf("planted failure caught by %d shards, union %d/%d", caught, offset, len(lines))
	}
	test.Logf("union leaf: %.3fs cooked=false", time.Since(started).Seconds())
}

// Mutant-must-fail remains a whole-corpus condition, as before sharding: each parser mutant, composed and run on Node
// over every live case, must differ from the unmutated composition's run. One test per mutant, each with its own
// baseline run beside the mutant's: the union ran all four whole-corpus Node runs, 31.0 s alone on one core (Home).
func compositionCatchesMutant(test *testing.T, index int) {
	test.Helper()
	state := compositionSetup(test)
	started := time.Now()
	timer := time.AfterFunc(90*time.Second, func() { panic("cooked: composition mutant exceeded 90s") })
	defer timer.Stop()
	cases := filepath.Join(test.TempDir(), "cases.txt")
	if writeError := os.WriteFile(cases, []byte(strings.Join(state.lines, "\n")+"\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	source, _ := filepath.Abs("compose_main.ts")
	mutation := mutants[index]
	mutated := portDirectory(test, &mutation)
	// The baseline and the mutant are two independent whole-corpus Node runs, so they run at once: 45 to 48 s one after
	// the other in css#10/17 at load 45 to 51 (Eve and Adam, Oct 23), each run about half of it.
	runs := [2]run{}
	failures := [2]error{}
	var group sync.WaitGroup
	for index, path := range []string{source, filepath.Join(mutated, "compose_main.ts")} {
		group.Add(1)
		go func() {
			defer group.Done()
			runs[index], failures[index] = compositionNodeRun(test, path, cases)
		}()
	}
	group.Wait()
	for _, err := range failures {
		if err != nil {
			test.Fatal(err)
		}
	}
	baseline, result := runs[0], runs[1]
	if baseline.exitCode != 0 || len(baseline.stderr) != 0 {
		test.Fatalf("Node baseline: %d %s", baseline.exitCode, baseline.stderr)
	}
	if keep := os.Getenv("ADAMIC_CSS_KEEP_COMPOSED"); keep != "" && index == 0 {
		if writeError := os.WriteFile(keep, baseline.stdout, 0644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if result.exitCode != 0 || len(result.stderr) != 0 {
		test.Fatalf("composed mutant must terminate: %s", result.stderr)
	}
	difference := firstDifference(string(result.stdout), string(baseline.stdout))
	if difference == "" {
		test.Fatalf("composition comparison missed mutant %s", mutation.name)
	}
	test.Logf(
		"Node composition caught %s: %s (%.3fs cooked=false)",
		mutation.name,
		difference,
		time.Since(started).Seconds(),
	)
}

func TestCompositionMatchesGoCatchesMutant_0(test *testing.T) {
	test.Parallel()
	compositionCatchesMutant(test, 0)
}
func TestCompositionMatchesGoCatchesMutant_1(test *testing.T) {
	test.Parallel()
	compositionCatchesMutant(test, 1)
}
func TestCompositionMatchesGoCatchesMutant_2(test *testing.T) {
	test.Parallel()
	compositionCatchesMutant(test, 2)
}

// Every parser mutant has its catching test.
func TestCompositionMatchesGoCatchesEveryMutant(test *testing.T) {
	test.Parallel()
	if len(mutants) != 3 {
		test.Fatalf(
			"%d parser mutants, three TestCompositionMatchesGoCatchesMutant_ tests: add one per mutant",
			len(mutants),
		)
	}
}

// Enumerate the original live corpus without its unrelated raw-parser benchmark.
func compositionCases(test *testing.T) string {
	test.Helper()
	dir := test.TempDir()
	cases := filepath.Join(dir, "cases.txt")
	repo, _ := filepath.Abs(repository)
	request, _ := json.Marshal(map[string]any{
		"cases":      cases,
		"repository": repo,
		"roots":      oracles.CorpusRoots(),
		"fixtures":   os.Getenv("ADAMIC_CSS_FIXTURES"),
	})
	requestPath := filepath.Join(dir, "request.json")
	if writeError := os.WriteFile(requestPath, request, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	oracle := oracles.CompositionOracle(
		test,
		"css-composition-corpus-oracle",
		"testdata/composition_corpus_side_test.go",
		"internal/format/css/postcss/adamic_port_side_test.go",
		"./internal/format/css/postcss",
	)
	command, ctx := compositionSetupCommand(
		test,
		oracle,
		"-test.timeout=0",
		"-test.v",
		"-test.run=^TestAdamicPortCases$",
	)
	command.Dir = filepath.Join(repo, "cohere", "internal", "format", "css", "postcss")
	command.Env = append(os.Environ(), "ADAMIC_PORT_REQUEST="+requestPath)
	// Read through childguard.Capture for compositionExecute's reason: a WaitDelay could fail a run that had finished.
	// The run has no deadline of its own, and the unit's own ceiling bounds it.
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	output := append(stdout, stderr...)
	if captureError != nil {
		test.Fatalf("composition corpus: %v\n%s", captureError, output)
	}
	test.Logf("%s", output)
	return cases
}

func TestCompositionMatchesGo_000(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 0)
}

func TestCompositionMatchesGo_001(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 1)
}

func TestCompositionMatchesGo_002(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 2)
}

func TestCompositionMatchesGo_003(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 3)
}

func TestCompositionMatchesGo_004(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 4)
}

func TestCompositionMatchesGo_005(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 5)
}

func TestCompositionMatchesGo_006(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 6)
}

func TestCompositionMatchesGo_007(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 7)
}

func TestCompositionMatchesGo_008(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 8)
}

func TestCompositionMatchesGo_009(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 9)
}

func TestCompositionMatchesGo_010(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 10)
}

func TestCompositionMatchesGo_011(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 11)
}

func TestCompositionMatchesGo_012(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 12)
}

func TestCompositionMatchesGo_013(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 13)
}

func TestCompositionMatchesGo_014(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 14)
}

func TestCompositionMatchesGo_015(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 15)
}

func TestCompositionMatchesGo_016(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 16)
}

func TestCompositionMatchesGo_017(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 17)
}

func TestCompositionMatchesGo_018(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 18)
}

func TestCompositionMatchesGo_019(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 19)
}

func TestCompositionMatchesGo_020(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 20)
}

func TestCompositionMatchesGo_021(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 21)
}

func TestCompositionMatchesGo_022(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 22)
}

func TestCompositionMatchesGo_023(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 23)
}

func TestCompositionMatchesGo_024(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 24)
}

func TestCompositionMatchesGo_025(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 25)
}

func TestCompositionMatchesGo_026(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 26)
}

func TestCompositionMatchesGo_027(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 27)
}

func TestCompositionMatchesGo_028(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 28)
}

func TestCompositionMatchesGo_029(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 29)
}

func TestCompositionMatchesGo_030(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 30)
}

func TestCompositionMatchesGo_031(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 31)
}

func TestCompositionMatchesGo_032(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 32)
}

func TestCompositionMatchesGo_033(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 33)
}

func TestCompositionMatchesGo_034(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 34)
}

func TestCompositionMatchesGo_035(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 35)
}

func TestCompositionMatchesGo_036(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 36)
}

func TestCompositionMatchesGo_037(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 37)
}

func TestCompositionMatchesGo_038(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 38)
}

func TestCompositionMatchesGo_039(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 39)
}

func TestCompositionMatchesGo_040(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 40)
}

func TestCompositionMatchesGo_041(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 41)
}

func TestCompositionMatchesGo_042(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 42)
}

func TestCompositionMatchesGo_043(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 43)
}

func TestCompositionMatchesGo_044(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 44)
}

func TestCompositionMatchesGo_045(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 45)
}

func TestCompositionMatchesGo_046(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 46)
}

func TestCompositionMatchesGo_047(test *testing.T) {
	test.Parallel()
	compositionRunShard(test, 47)
}

var compositionShardFunctions = [...]func(*testing.T){
	TestCompositionMatchesGo_000,
	TestCompositionMatchesGo_001,
	TestCompositionMatchesGo_002,
	TestCompositionMatchesGo_003,
	TestCompositionMatchesGo_004,
	TestCompositionMatchesGo_005,
	TestCompositionMatchesGo_006,
	TestCompositionMatchesGo_007,
	TestCompositionMatchesGo_008,
	TestCompositionMatchesGo_009,
	TestCompositionMatchesGo_010,
	TestCompositionMatchesGo_011,
	TestCompositionMatchesGo_012,
	TestCompositionMatchesGo_013,
	TestCompositionMatchesGo_014,
	TestCompositionMatchesGo_015,
	TestCompositionMatchesGo_016,
	TestCompositionMatchesGo_017,
	TestCompositionMatchesGo_018,
	TestCompositionMatchesGo_019,
	TestCompositionMatchesGo_020,
	TestCompositionMatchesGo_021,
	TestCompositionMatchesGo_022,
	TestCompositionMatchesGo_023,
	TestCompositionMatchesGo_024,
	TestCompositionMatchesGo_025,
	TestCompositionMatchesGo_026,
	TestCompositionMatchesGo_027,
	TestCompositionMatchesGo_028,
	TestCompositionMatchesGo_029,
	TestCompositionMatchesGo_030,
	TestCompositionMatchesGo_031,
	TestCompositionMatchesGo_032,
	TestCompositionMatchesGo_033,
	TestCompositionMatchesGo_034,
	TestCompositionMatchesGo_035,
	TestCompositionMatchesGo_036,
	TestCompositionMatchesGo_037,
	TestCompositionMatchesGo_038,
	TestCompositionMatchesGo_039,
	TestCompositionMatchesGo_040,
	TestCompositionMatchesGo_041,
	TestCompositionMatchesGo_042,
	TestCompositionMatchesGo_043,
	TestCompositionMatchesGo_044,
	TestCompositionMatchesGo_045,
	TestCompositionMatchesGo_046,
	TestCompositionMatchesGo_047,
}

// Bound the whole process group, including Go's compiler children, using only
// the declared command's executable rather than an external timeout utility.
// It returns the command's context with it, for childguard.Capture to end the reads by.
func compositionCommand(test *testing.T, name string, arguments ...string) (*exec.Cmd, context.Context) {
	test.Helper()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	test.Cleanup(cancel)
	return compositionCommandContext(timeoutContext, name, arguments...), timeoutContext
}

// compositionExecute reads a run's output through childguard.Capture, to its end with no timer deciding when: exec's
// own copying under a one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second
// after the child exited, so a run that had finished could fail its shard under load (#n3gk22h's shape). It drops
// childguard.Run's stall guard and loses nothing by it: the guard's windows (four minutes of silence, thirty before the
// first byte) are longer than the command's 90 s deadline, which ends a hung run sooner.
func compositionExecute(test *testing.T, environment []string, name string, arguments ...string) run {
	test.Helper()
	result, compositionRunError := compositionRun(test, environment, name, arguments...)
	if compositionRunError != nil {
		test.Fatal(compositionRunError)
	}
	return result
}

// compositionRun is compositionExecute returning its failure to start or read the command instead of failing the
// test, so a caller can run commands on goroutines of its own and fail on the test's.
func compositionRun(test *testing.T, environment []string, name string, arguments ...string) (run, error) {
	command, ctx := compositionCommand(test, name, arguments...)
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	var exit *exec.ExitError
	if captureError != nil && !errors.As(captureError, &exit) {
		return run{}, fmt.Errorf("running %s: %v", name, captureError)
	}
	return run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}, nil
}

func compositionNode(test *testing.T, path string, arguments ...string) run {
	test.Helper()
	result, compositionNodeRunError := compositionNodeRun(test, path, arguments...)
	if compositionNodeRunError != nil {
		test.Fatal(compositionNodeRunError)
	}
	return result
}

// compositionNodeRun is compositionNode for a goroutine: it returns its failure rather than failing the test.
func compositionNodeRun(test *testing.T, path string, arguments ...string) (run, error) {
	runner, absolutePathError := filepath.Abs(filepath.Join(repository, "oracle", "node.mjs"))
	if absolutePathError != nil {
		return run{}, absolutePathError
	}
	return compositionRun(
		test,
		nil,
		"node",
		append([]string{"--disable-warning=ExperimentalWarning", runner, path}, arguments...)...,
	)
}

func compositionLeaks(test *testing.T, state compositionProducts, cases string) string {
	test.Helper()
	switch runtime.GOOS {
	case "linux":
		report := compositionExecute(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, state.sanitized, cases)
		if report.exitCode != 0 {
			return fmt.Sprintf("exit %d\n%s", report.exitCode, report.stderr)
		}
	case "darwin":
		source, readError := os.ReadFile(filepath.Join(state.product, "program.c"))
		if readError != nil {
			test.Fatal(readError)
		}
		binary := filepath.Join(test.TempDir(), "port")
		if buildError := native.Build(string(source), binary, native.Options{}); buildError != nil {
			test.Fatal(buildError)
		}
		report := compositionExecute(test, nil, "leaks", "--atExit", "--", binary, cases)
		if report.exitCode != 0 {
			return string(report.stdout)
		}
	default:
		test.Fatalf("no leak check for %s", runtime.GOOS)
	}
	return ""
}

func compositionInputs() buildcache.Inputs {
	return buildcache.Inputs{Name: "css-composition-lowered", Files: []string{
		"stage1/cohere/css",
		"stage1/cohere/selector",
		"stage1/cohere/values",
		"stage1/cohere/mediaquery",
		"stage1/cohere/cssstrings",
		"stage1/cohere/cssnumbers",
		"internal",
		"cohere",
		"go.mod",
		"go.work",
	}, Toolchain: []string{runtime.Version()}, Flags: []string{"C and JavaScript"}}
}

func compositionLoweredProduct(test *testing.T) (string, string) {
	test.Helper()
	source, _ := filepath.Abs("compose_main.ts")
	inputs := compositionInputs()
	product := buildcache.Product(test, inputs, func(dir string) error {
		program := lowered(test, source)
		if writeError := os.WriteFile(
			filepath.Join(dir, "program.c"),
			[]byte(native.C(program)),
			0644,
		); writeError != nil {
			return writeError
		}
		return os.WriteFile(filepath.Join(dir, "program.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	return source, product
}

func compositionNativeProduct(test *testing.T) string {
	test.Helper()
	_, product := compositionLoweredProduct(test)
	content, readError := os.ReadFile(filepath.Join(product, "program.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	inputs := compositionInputs()
	inputs.Name = "css-composition-native"
	inputs.Flags = append(
		native.Flags(native.Options{Sanitize: true}),
		"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
	)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	nativeProduct := buildcache.Product(test, inputs, func(dir string) error {
		return native.Build(string(content), filepath.Join(dir, "port"), native.Options{Sanitize: true})
	})
	return filepath.Join(nativeProduct, "port")

}
