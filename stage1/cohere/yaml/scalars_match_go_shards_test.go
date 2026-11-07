package yaml

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
)

const scalarsMatchGoShardCount = 8

// Hash the complete encoded case, including its streaming chunk size. Repeated
// cases keep their multiplicity, and inserting a case cannot move another owner.
func scalarsMatchGoPartition(cases []string) [][]int {
	return yamlShards(cases, scalarsMatchGoShardCount)
}

func scalarsMatchGoCases(test *testing.T) ([]string, int) {
	test.Helper()
	path, _, count := scalarCases(test)
	data, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	cases := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(cases) != count {
		test.Fatalf("enumerated %d cases; original generator counted %d", len(cases), count)
	}
	return cases, count
}

func scalarsMatchGoWriteCases(test *testing.T, cases []string) string {
	test.Helper()
	path := filepath.Join(test.TempDir(), "cases.txt")
	if writeError := os.WriteFile(path, []byte(strings.Join(cases, "\n")+"\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	return path
}

// scalarsMatchGoRun reads a run's output through childguard.Capture for fileDriverRun's reason: a one-second WaitDelay
// could fail a Go oracle, native or Node run that had finished. The caller's 55 s deadline ends a hung one.
func scalarsMatchGoRun(
	test *testing.T,
	callerContext context.Context,
	environment []string,
	name string,
	args ...string,
) []byte {
	test.Helper()
	command := exec.CommandContext(callerContext, name, args...)
	command.Env = append(os.Environ(), environment...)
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node) is a tool: the one internal/toolchain resolves, with a tool's environment plus the
		// caller's settings, not PATH's, as toolchain.Command runs one under a context it has no form for. A program a
		// test built (the Go oracle, the port) runs as named, with this process's environment. Path's error fails the
		// command as it starts.
		if path, pathError := toolchain.Path(name); pathError != nil {
			command.Err = pathError
		} else {
			command = exec.CommandContext(callerContext, path, args...)
			command.Env = toolchain.Env(environment...)
		}
	}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	output, errors, captureError := childguard.Capture(callerContext, command)
	if captureError != nil {
		test.Fatalf("%s: %v\n%s", name, captureError, errors)
	}
	if len(errors) != 0 {
		test.Fatalf("%s stderr: %s", name, errors)
	}
	return output
}

func scalarsMatchGoDisagrees(actual, expected []byte) bool { return !bytes.Equal(actual, expected) }

func scalarsMatchGoRunShard(test *testing.T, shard int) {
	test.Helper()
	setup := time.Now()
	// Every leaf fetches its own products, once per process, before its work clock.
	oracle := goScalarsOracle(test)
	emitted := yamlScalarsBuild.emitted(test, "scalar_main.ts")
	compiled := yamlScalarsBuild.binary(test, "scalar_main.ts")
	all, _ := scalarsMatchGoCases(test)
	indices := scalarsMatchGoPartition(all)[shard]
	cases := make([]string, len(indices))
	for indexIndex, index := range indices {
		cases[indexIndex] = all[index]
	}
	if len(cases) == 0 {
		test.Fatal("empty shard")
	}
	path := scalarsMatchGoWriteCases(test, cases)
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	entry, absolutePathError := filepath.Abs("scalar_main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	test.Logf("setup %.3fs; %d/%d cases", time.Since(setup).Seconds(), len(cases), len(all))
	started := time.Now()
	timeoutContext, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	expected := scalarsMatchGoRun(test, timeoutContext, nil, oracle, path)
	runner := filepath.Join(root, "oracle/node.mjs")
	for _, side := range []struct {
		name    string
		env     []string
		command string
		args    []string
	}{
		{"native ASan/UBSan/LSan", []string{"ASAN_OPTIONS=detect_leaks=1"}, compiled, []string{path}},
		{"Node", nil, "node", []string{"--disable-warning=ExperimentalWarning", runner, entry, path}},
		{"emitted JavaScript", nil, "node", []string{"--disable-warning=ExperimentalWarning", runner, emitted, path}},
	} {
		actual := scalarsMatchGoRun(test, timeoutContext, side.env, side.command, side.args...)
		if side.name == "native ASan/UBSan/LSan" && os.Getenv("ADAMIC_YAML_SCALARS_PLANT") == "1" {
			for _, index := range indices {
				if index == 0 {
					actual = append(actual, '!')
				}
			}
		}
		if scalarsMatchGoDisagrees(actual, expected) {
			test.Fatalf("%s: %s", side.name, firstDifference(actual, expected))
		}
	}
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		test.Skip("set ADAMIC_YAML_LIBRARY to yaml@2.9.0 and prettier@3.9.6 for the independent library oracle")
	}
	external := scalarsMatchGoRun(test, timeoutContext, nil, "node", "testdata/scalar_library.mjs", library, path)
	if scalarsMatchGoDisagrees(external, expected) {
		test.Fatalf("yaml@2.9.0: %s", firstDifference(external, expected))
	}
	test.Logf(
		"own work %.3fs; %d answer bytes identical on all five sides",
		time.Since(started).Seconds(),
		len(expected),
	)
}

func TestScalarsMatchGoUnion(test *testing.T) {
	test.Parallel()
	cases, count := scalarsMatchGoCases(test)
	shards := scalarsMatchGoPartition(cases)
	visits := make([]int, count)
	for _, indices := range shards {
		for _, index := range indices {
			visits[index]++
		}
	}
	for index, visit := range visits {
		if visit != 1 {
			test.Fatalf("case %d visited %d times", index, visit)
		}
	}
	test.Logf("counted union: all %d original cases visited exactly once", count)
}

func TestScalarsMatchGoPlantedFailure(test *testing.T) {
	test.Parallel()
	oracle := goScalarsOracle(test)
	compiled := yamlScalarsBuild.binary(test, "scalar_main.ts")
	cases, _ := scalarsMatchGoCases(test)
	// Execute a real case through both sides, then corrupt its actual stdout.
	path := scalarsMatchGoWriteCases(test, cases[:1])
	timeoutContext, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	expected := scalarsMatchGoRun(test, timeoutContext, nil, oracle, path)
	actual := scalarsMatchGoRun(test, timeoutContext, []string{"ASAN_OPTIONS=detect_leaks=1"}, compiled, path)
	if scalarsMatchGoDisagrees(actual, expected) {
		test.Fatal("native control disagrees before planting")
	}
	actual = append(actual, '!')
	var caught []int
	for shard, indices := range scalarsMatchGoPartition(cases) {
		for _, index := range indices {
			output := expected
			if index == 0 {
				output = actual
			}
			if scalarsMatchGoDisagrees(output, expected) {
				caught = append(caught, shard)
			}
		}
	}
	if len(caught) != 1 {
		test.Fatalf("planted stdout disagreement caught by %v; want one owner", caught)
	}
	test.Logf("planted real native stdout disagreement caught by TestScalarsMatchGo_%03d only", caught[0])
}

func TestScalarsMatchGo_000(test *testing.T) { test.Parallel(); scalarsMatchGoRunShard(test, 0) }
func TestScalarsMatchGo_001(test *testing.T) { test.Parallel(); scalarsMatchGoRunShard(test, 1) }
func TestScalarsMatchGo_002(test *testing.T) { test.Parallel(); scalarsMatchGoRunShard(test, 2) }
func TestScalarsMatchGo_003(test *testing.T) { test.Parallel(); scalarsMatchGoRunShard(test, 3) }
func TestScalarsMatchGo_004(test *testing.T) { test.Parallel(); scalarsMatchGoRunShard(test, 4) }
func TestScalarsMatchGo_005(test *testing.T) { test.Parallel(); scalarsMatchGoRunShard(test, 5) }
func TestScalarsMatchGo_006(test *testing.T) { test.Parallel(); scalarsMatchGoRunShard(test, 6) }
func TestScalarsMatchGo_007(test *testing.T) { test.Parallel(); scalarsMatchGoRunShard(test, 7) }
