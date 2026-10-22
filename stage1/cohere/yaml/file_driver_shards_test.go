package yaml

import (
	"bytes"
	"context"
	"hash/fnv"
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

const testFileDriverShards = 8

func fileDriverPartition(inputs []string) [][]int {
	shards := make([][]int, testFileDriverShards)
	for index, input := range inputs {
		h := fnv.New64a()
		_, _ = h.Write([]byte(input))
		shard := int(h.Sum64() % testFileDriverShards)
		shards[shard] = append(shards[shard], index)
	}
	return shards
}

func fileDriverCheckUnion(t *testing.T, shards [][]int, count int) {
	t.Helper()
	if len(shards) != testFileDriverShards {
		t.Fatal("shard enumeration disagrees with constant")
	}
	seen := make([]int, count)
	for _, indices := range shards {
		for _, index := range indices {
			if index < 0 || index >= count {
				t.Fatalf("invalid case %d", index)
			}
			seen[index]++
		}
	}
	for index, visits := range seen {
		if visits != 1 {
			t.Fatalf("case %d visited %d times", index, visits)
		}
	}
}

func fileDriverSetup(t *testing.T) *fileDriverState {
	started := time.Now()
	cases, files, _ := formatCases(t)
	data, err := os.ReadFile(cases)
	if err != nil {
		t.Fatal(err)
	}
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")[:files]
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	for _, text := range []string{"", "  \t\n", "\ufeff", "\ufeffa: b\n", "|+", "|+\n", "|+ # header\n", ">+", "a: |+", "a: |+\n", "a: |+\n  b\n\n", "{a: b, c: [x,y]}\n", "---\n...\n", "key: 'x\\y'\n"} {
		inputs = append(inputs, "0\t"+escape.Replace(text))
	}
	path := filepath.Join(t.TempDir(), "driver-cases.txt")
	if err := os.WriteFile(path, []byte(strings.Join(inputs, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	expected := bytes.Split(bytes.TrimSuffix(fileDriverGoFormat(t, path), []byte("\n")), []byte("\n"))
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs("main.ts")
	if err != nil {
		t.Fatal(err)
	}
	binary := yamlFormatBuild.binary(t, "main.ts")
	emitted := yamlFormatBuild.emitted(t, "main.ts")
	runner := filepath.Join(root, "oracle/node.mjs")
	if len(expected) != len(inputs) {
		t.Fatalf("oracle returned %d answers for %d inputs", len(expected), len(inputs))
	}
	t.Logf("TestFileDriver (setup): %.3fs", time.Since(started).Seconds())
	return &fileDriverState{inputs: inputs, expected: expected, binary: binary, emitted: emitted, runner: runner, entry: entry}
}

type fileDriverState struct {
	inputs                         []string
	expected                       [][]byte
	binary, emitted, runner, entry string
}

var fileDriverShared *fileDriverState

func fileDriverGet(t *testing.T) *fileDriverState {
	t.Helper()
	if fileDriverShared == nil {
		t.Fatal("file driver setup must finish before a shard starts")
	}
	return fileDriverShared
}

func TestFileDriverUnion(t *testing.T) {
	t.Parallel()
	state := fileDriverGet(t)
	shards := fileDriverPartition(state.inputs)
	fileDriverCheckUnion(t, shards, len(state.inputs))
	// Plant one disagreement in the actual native stdout and exercise the byte oracle
	// over every enumerated slice; precisely one owning shard must detect it.
	file := filepath.Join(t.TempDir(), "input.yaml")
	if err := os.WriteFile(file, []byte(unescapeCase(state.inputs[0])), 0644); err != nil {
		t.Fatal(err)
	}
	actual := fileDriverRun(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, state.binary, file)
	wanted := []byte(unescapeCase("0\t" + strings.TrimPrefix(string(state.expected[0]), "ok\t")))
	if !bytes.Equal(actual, wanted) {
		t.Fatal("native control disagrees before planting failure")
	}
	planted := append(append([]byte(nil), actual...), '!')
	var caught []int
	for shard, indices := range shards {
		for _, index := range indices {
			output := wanted
			if index == 0 {
				output = planted
			}
			if !bytes.Equal(output, wanted) {
				caught = append(caught, shard)
			}
		}
	}
	if len(caught) != 1 {
		t.Fatalf("planted disagreement caught by %v", caught)
	}
	t.Logf("union %d cases exactly once; planted native stdout case 0 caught by TestFileDriver_%03d", len(state.inputs), caught[0])
}

func fileDriverRunShard(t *testing.T, shard int) {
	t.Helper()
	state := fileDriverGet(t)
	started := time.Now()
	shards := fileDriverPartition(state.inputs)
	fileDriverCheckUnion(t, shards, len(state.inputs))
	for _, index := range shards[shard] {
		answer := string(state.expected[index])
		if !strings.HasPrefix(answer, "ok\t") {
			t.Fatalf("driver control %d invalid: %s", index, answer)
		}
		wanted := []byte(unescapeCase("0\t" + strings.TrimPrefix(answer, "ok\t")))
		file := filepath.Join(t.TempDir(), "input.yaml")
		if err := os.WriteFile(file, []byte(unescapeCase(state.inputs[index])), 0644); err != nil {
			t.Fatal(err)
		}
		for _, side := range []struct {
			name string
			out  []byte
		}{
			{"native", fileDriverRun(t, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, state.binary, file)},
			{"Node", fileDriverRun(t, "", nil, "node", "--disable-warning=ExperimentalWarning", state.runner, state.entry, file)},
			{"emitted JavaScript", fileDriverRun(t, "", nil, "node", "--disable-warning=ExperimentalWarning", state.runner, state.emitted, file)},
		} {
			if !bytes.Equal(side.out, wanted) {
				t.Fatalf("%s file %d: %s", side.name, index, firstDifference(side.out, wanted))
			}
		}
	}
	t.Logf("%d cases; shard case time (setup excluded) %.3fs", len(shards[shard]), time.Since(started).Seconds())
}

func TestFileDriver_000(t *testing.T) {
	t.Parallel()
	fileDriverRunShard(t, 0)
}
func TestFileDriver_001(t *testing.T) {
	t.Parallel()
	fileDriverRunShard(t, 1)
}
func TestFileDriver_002(t *testing.T) {
	t.Parallel()
	fileDriverRunShard(t, 2)
}
func TestFileDriver_003(t *testing.T) {
	t.Parallel()
	fileDriverRunShard(t, 3)
}
func TestFileDriver_004(t *testing.T) {
	t.Parallel()
	fileDriverRunShard(t, 4)
}
func TestFileDriver_005(t *testing.T) {
	t.Parallel()
	fileDriverRunShard(t, 5)
}
func TestFileDriver_006(t *testing.T) {
	t.Parallel()
	fileDriverRunShard(t, 6)
}
func TestFileDriver_007(t *testing.T) {
	t.Parallel()
	fileDriverRunShard(t, 7)
}

// fileDriverRun retains the strict stderr and stdout oracle while bounding every child. Output is read through
// childguard.Capture, to its end with no timer deciding when: exec's own copying under a one-second WaitDelay reported
// exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after the child exited, so a Go oracle, native or
// Node run that had finished could fail its shard under load (#n3gk22h's shape). The run's 90 s deadline ends a hung
// one.
func fileDriverRun(t *testing.T, directory string, environment []string, name string, args ...string) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node) is a tool: the one internal/toolchain resolves, with a tool's environment plus the caller's
		// settings, not PATH's, run in directory with PWD to match, as toolchain.CommandIn runs one under a context it
		// has no form for. A program a test built (the Go oracle, the port) runs as named, with this process's
		// environment. An error resolving it fails the command as it starts.
		if path, err := toolchain.Path(name); err != nil {
			command.Err = err
		} else if absolute, err := filepath.Abs(directory); err != nil {
			command.Err = err
		} else {
			command = exec.CommandContext(ctx, path, args...)
			command.Dir = directory
			command.Env = append(toolchain.Env(environment...), "PWD="+absolute)
		}
	}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	output, errors, err := childguard.Capture(ctx, command)
	if err != nil {
		t.Fatalf("%s: %v\n%s", name, err, errors)
	}
	if len(errors) != 0 {
		t.Fatalf("%s stderr: %s", name, errors)
	}
	return output
}

func fileDriverGoFormat(t *testing.T, cases string) []byte {
	t.Helper()
	goBinary := goFormatOracle(t)
	expected := fileDriverRun(t, "", nil, goBinary, "--cases", cases)
	if artifacts := os.Getenv("ADAMIC_YAML_ARTIFACTS"); artifacts != "" {
		if err := os.MkdirAll(artifacts, 0755); err != nil {
			t.Fatal(err)
		}
		executable, err := os.ReadFile(goBinary)
		if err != nil {
			t.Fatal(err)
		}
		inputs, err := os.ReadFile(cases)
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range []struct {
			name string
			data []byte
			mode os.FileMode
		}{{"go-format", executable, 0755}, {"format-cases.txt", inputs, 0644}, {"format-expected.txt", expected, 0644}} {
			if err := os.WriteFile(filepath.Join(artifacts, file.name), file.data, file.mode); err != nil {
				t.Fatal(err)
			}
		}
	}
	return expected
}
