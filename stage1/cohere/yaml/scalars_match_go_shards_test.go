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

	"github.com/system-inc/adamic/internal/toolchain"
)

const scalarsMatchGoShardCount = 8

// Hash the complete encoded case, including its streaming chunk size. Repeated
// cases keep their multiplicity, and inserting a case cannot move another owner.
func scalarsMatchGoPartition(cases []string) [][]int {
	return yamlShards(cases, scalarsMatchGoShardCount)
}

func scalarsMatchGoCases(t *testing.T) ([]string, int) {
	t.Helper()
	path, _, count := scalarCases(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	cases := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(cases) != count {
		t.Fatalf("enumerated %d cases; original generator counted %d", len(cases), count)
	}
	return cases, count
}

func scalarsMatchGoWriteCases(t *testing.T, cases []string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "cases.txt")
	if err := os.WriteFile(path, []byte(strings.Join(cases, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func scalarsMatchGoRun(t *testing.T, ctx context.Context, environment []string, name string, args ...string) []byte {
	t.Helper()
	command := exec.CommandContext(ctx, name, args...)
	command.Env = append(os.Environ(), environment...)
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node) is a tool: the one internal/toolchain resolves, with a tool's environment plus the caller's
		// settings, not PATH's, as toolchain.Command runs one under a context it has no form for. A program a test built
		// (the Go oracle, the port) runs as named, with this process's environment. Path's error fails the command as it
		// starts.
		if path, err := toolchain.Path(name); err != nil {
			command.Err = err
		} else {
			command = exec.CommandContext(ctx, path, args...)
			command.Env = toolchain.Env(environment...)
		}
	}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = time.Second
	var output, errors bytes.Buffer
	command.Stdout, command.Stderr = &output, &errors
	if err := command.Run(); err != nil {
		t.Fatalf("%s: %v\n%s", name, err, errors.Bytes())
	}
	if errors.Len() != 0 {
		t.Fatalf("%s stderr: %s", name, errors.Bytes())
	}
	return output.Bytes()
}

func scalarsMatchGoDisagrees(actual, expected []byte) bool { return !bytes.Equal(actual, expected) }

func scalarsMatchGoRunShard(t *testing.T, shard int) {
	t.Helper()
	setup := time.Now()
	// Every leaf fetches its own products, once per process, before its work clock.
	oracle := goScalarsOracle(t)
	emitted := yamlScalarsBuild.emitted(t, "scalar_main.ts")
	compiled := yamlScalarsBuild.binary(t, "scalar_main.ts")
	all, _ := scalarsMatchGoCases(t)
	indices := scalarsMatchGoPartition(all)[shard]
	cases := make([]string, len(indices))
	for i, index := range indices {
		cases[i] = all[index]
	}
	if len(cases) == 0 {
		t.Fatal("empty shard")
	}
	path := scalarsMatchGoWriteCases(t, cases)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs("scalar_main.ts")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("setup %.3fs; %d/%d cases", time.Since(setup).Seconds(), len(cases), len(all))
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	expected := scalarsMatchGoRun(t, ctx, nil, oracle, path)
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
		actual := scalarsMatchGoRun(t, ctx, side.env, side.command, side.args...)
		if side.name == "native ASan/UBSan/LSan" && os.Getenv("ADAMIC_YAML_SCALARS_PLANT") == "1" {
			for _, index := range indices {
				if index == 0 {
					actual = append(actual, '!')
				}
			}
		}
		if scalarsMatchGoDisagrees(actual, expected) {
			t.Fatalf("%s: %s", side.name, firstDifference(actual, expected))
		}
	}
	library := os.Getenv("ADAMIC_YAML_LIBRARY")
	if library == "" {
		t.Skip("set ADAMIC_YAML_LIBRARY to yaml@2.9.0 and prettier@3.9.6 for the independent library oracle")
	}
	external := scalarsMatchGoRun(t, ctx, nil, "node", "testdata/scalar_library.mjs", library, path)
	if scalarsMatchGoDisagrees(external, expected) {
		t.Fatalf("yaml@2.9.0: %s", firstDifference(external, expected))
	}
	t.Logf("own work %.3fs; %d answer bytes identical on all five sides", time.Since(started).Seconds(), len(expected))
}

func TestScalarsMatchGoUnion(t *testing.T) {
	t.Parallel()
	cases, count := scalarsMatchGoCases(t)
	shards := scalarsMatchGoPartition(cases)
	visits := make([]int, count)
	for _, indices := range shards {
		for _, index := range indices {
			visits[index]++
		}
	}
	for index, n := range visits {
		if n != 1 {
			t.Fatalf("case %d visited %d times", index, n)
		}
	}
	t.Logf("counted union: all %d original cases visited exactly once", count)
}

func TestScalarsMatchGoPlantedFailure(t *testing.T) {
	t.Parallel()
	oracle := goScalarsOracle(t)
	compiled := yamlScalarsBuild.binary(t, "scalar_main.ts")
	cases, _ := scalarsMatchGoCases(t)
	// Execute a real case through both sides, then corrupt its actual stdout.
	path := scalarsMatchGoWriteCases(t, cases[:1])
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	expected := scalarsMatchGoRun(t, ctx, nil, oracle, path)
	actual := scalarsMatchGoRun(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=1"}, compiled, path)
	if scalarsMatchGoDisagrees(actual, expected) {
		t.Fatal("native control disagrees before planting")
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
		t.Fatalf("planted stdout disagreement caught by %v; want one owner", caught)
	}
	t.Logf("planted real native stdout disagreement caught by TestScalarsMatchGo_%03d only", caught[0])
}

func TestScalarsMatchGo_000(t *testing.T) { t.Parallel(); scalarsMatchGoRunShard(t, 0) }
func TestScalarsMatchGo_001(t *testing.T) { t.Parallel(); scalarsMatchGoRunShard(t, 1) }
func TestScalarsMatchGo_002(t *testing.T) { t.Parallel(); scalarsMatchGoRunShard(t, 2) }
func TestScalarsMatchGo_003(t *testing.T) { t.Parallel(); scalarsMatchGoRunShard(t, 3) }
func TestScalarsMatchGo_004(t *testing.T) { t.Parallel(); scalarsMatchGoRunShard(t, 4) }
func TestScalarsMatchGo_005(t *testing.T) { t.Parallel(); scalarsMatchGoRunShard(t, 5) }
func TestScalarsMatchGo_006(t *testing.T) { t.Parallel(); scalarsMatchGoRunShard(t, 6) }
func TestScalarsMatchGo_007(t *testing.T) { t.Parallel(); scalarsMatchGoRunShard(t, 7) }
