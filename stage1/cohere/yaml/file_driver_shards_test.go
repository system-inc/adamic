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

func fileDriverCheckUnion(test *testing.T, shards [][]int, count int) {
	test.Helper()
	if len(shards) != testFileDriverShards {
		test.Fatal("shard enumeration disagrees with constant")
	}
	seen := make([]int, count)
	for _, indices := range shards {
		for _, index := range indices {
			if index < 0 || index >= count {
				test.Fatalf("invalid case %d", index)
			}
			seen[index]++
		}
	}
	for index, visits := range seen {
		if visits != 1 {
			test.Fatalf("case %d visited %d times", index, visits)
		}
	}
}

func fileDriverSetup(test *testing.T) *fileDriverState {
	started := time.Now()
	cases, files, _ := formatCases(test)
	data, readError := os.ReadFile(cases)
	if readError != nil {
		test.Fatal(readError)
	}
	inputs := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")[:files]
	escape := strings.NewReplacer("\\", "\\\\", "\n", "\\n", "\r", "\\r", "\t", "\\t")
	for _, text := range []string{
		"",
		"  \t\n",
		"\ufeff",
		"\ufeffa: b\n",
		"|+",
		"|+\n",
		"|+ # header\n",
		">+",
		"a: |+",
		"a: |+\n",
		"a: |+\n  b\n\n",
		"{a: b, c: [x,y]}\n",
		"---\n...\n",
		"key: 'x\\y'\n",
	} {
		inputs = append(inputs, "0\t"+escape.Replace(text))
	}
	path := filepath.Join(test.TempDir(), "driver-cases.txt")
	if writeError := os.WriteFile(path, []byte(strings.Join(inputs, "\n")+"\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	expected := bytes.Split(bytes.TrimSuffix(fileDriverGoFormat(test, path), []byte("\n")), []byte("\n"))
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	entry, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	binary := yamlFormatBuild.binary(test, "main.ts")
	emitted := yamlFormatBuild.emitted(test, "main.ts")
	runner := filepath.Join(root, "oracle/node.mjs")
	if len(expected) != len(inputs) {
		test.Fatalf("oracle returned %d answers for %d inputs", len(expected), len(inputs))
	}
	test.Logf("TestFileDriver (setup): %.3fs", time.Since(started).Seconds())
	return &fileDriverState{
		inputs:   inputs,
		expected: expected,
		binary:   binary,
		emitted:  emitted,
		runner:   runner,
		entry:    entry,
	}
}

type fileDriverState struct {
	inputs                         []string
	expected                       [][]byte
	binary, emitted, runner, entry string
}

var fileDriverShared *fileDriverState

func fileDriverGet(test *testing.T) *fileDriverState {
	test.Helper()
	if fileDriverShared == nil {
		test.Fatal("file driver setup must finish before a shard starts")
	}
	return fileDriverShared
}

func TestFileDriverUnion(test *testing.T) {
	test.Parallel()
	state := fileDriverGet(test)
	shards := fileDriverPartition(state.inputs)
	fileDriverCheckUnion(test, shards, len(state.inputs))
	// Plant one disagreement in the actual native stdout and exercise the byte oracle
	// over every enumerated slice; precisely one owning shard must detect it.
	file := filepath.Join(test.TempDir(), "input.yaml")
	if writeError := os.WriteFile(file, []byte(unescapeCase(state.inputs[0])), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	actual := fileDriverRun(test, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, state.binary, file)
	wanted := []byte(unescapeCase("0\t" + strings.TrimPrefix(string(state.expected[0]), "ok\t")))
	if !bytes.Equal(actual, wanted) {
		test.Fatal("native control disagrees before planting failure")
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
		test.Fatalf("planted disagreement caught by %v", caught)
	}
	test.Logf(
		"union %d cases exactly once; planted native stdout case 0 caught by TestFileDriver_%03d",
		len(state.inputs),
		caught[0],
	)
}

func fileDriverRunShard(test *testing.T, shard int) {
	test.Helper()
	state := fileDriverGet(test)
	started := time.Now()
	shards := fileDriverPartition(state.inputs)
	fileDriverCheckUnion(test, shards, len(state.inputs))
	for _, index := range shards[shard] {
		answer := string(state.expected[index])
		if !strings.HasPrefix(answer, "ok\t") {
			test.Fatalf("driver control %d invalid: %s", index, answer)
		}
		wanted := []byte(unescapeCase("0\t" + strings.TrimPrefix(answer, "ok\t")))
		file := filepath.Join(test.TempDir(), "input.yaml")
		if writeError := os.WriteFile(file, []byte(unescapeCase(state.inputs[index])), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		for _, side := range []struct {
			name string
			out  []byte
		}{
			{"native", fileDriverRun(test, "", []string{"ASAN_OPTIONS=detect_leaks=1"}, state.binary, file)},
			{"Node", fileDriverRun(
				test,
				"",
				nil,
				"node",
				"--disable-warning=ExperimentalWarning",
				state.runner,
				state.entry,
				file,
			)},
			{"emitted JavaScript", fileDriverRun(
				test,
				"",
				nil,
				"node",
				"--disable-warning=ExperimentalWarning",
				state.runner,
				state.emitted,
				file,
			)},
		} {
			if !bytes.Equal(side.out, wanted) {
				test.Fatalf("%s file %d: %s", side.name, index, firstDifference(side.out, wanted))
			}
		}
	}
	test.Logf("%d cases; shard case time (setup excluded) %.3fs", len(shards[shard]), time.Since(started).Seconds())
}

func TestFileDriver_000(test *testing.T) {
	test.Parallel()
	fileDriverRunShard(test, 0)
}
func TestFileDriver_001(test *testing.T) {
	test.Parallel()
	fileDriverRunShard(test, 1)
}
func TestFileDriver_002(test *testing.T) {
	test.Parallel()
	fileDriverRunShard(test, 2)
}
func TestFileDriver_003(test *testing.T) {
	test.Parallel()
	fileDriverRunShard(test, 3)
}
func TestFileDriver_004(test *testing.T) {
	test.Parallel()
	fileDriverRunShard(test, 4)
}
func TestFileDriver_005(test *testing.T) {
	test.Parallel()
	fileDriverRunShard(test, 5)
}
func TestFileDriver_006(test *testing.T) {
	test.Parallel()
	fileDriverRunShard(test, 6)
}
func TestFileDriver_007(test *testing.T) {
	test.Parallel()
	fileDriverRunShard(test, 7)
}

// fileDriverRun retains the strict stderr and stdout oracle while bounding every child. Output is read through
// childguard.Capture, to its end with no timer deciding when: exec's own copying under a one-second WaitDelay reported
// exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after the child exited, so a Go oracle, native or
// Node run that had finished could fail its shard under load (#n3gk22h's shape). The run's 90 s deadline ends a hung
// one.
func fileDriverRun(test *testing.T, directory string, environment []string, name string, args ...string) []byte {
	test.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	if !strings.ContainsRune(name, filepath.Separator) {
		// A bare name (node) is a tool: the one internal/toolchain resolves, with a tool's environment plus the
		// caller's settings, not PATH's, run in directory with PWD to match, as toolchain.CommandIn runs one under a
		// context it has no form for. A program a test built (the Go oracle, the port) runs as named, with this
		// process's environment. An error resolving it fails the command as it starts.
		if path, pathError := toolchain.Path(name); pathError != nil {
			command.Err = pathError
		} else if absolute, absolutePathError := filepath.Abs(directory); absolutePathError != nil {
			command.Err = absolutePathError
		} else {
			command = exec.CommandContext(ctx, path, args...)
			command.Dir = directory
			command.Env = append(toolchain.Env(environment...), "PWD="+absolute)
		}
	}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	output, errors, captureError := childguard.Capture(ctx, command)
	if captureError != nil {
		test.Fatalf("%s: %v\n%s", name, captureError, errors)
	}
	if len(errors) != 0 {
		test.Fatalf("%s stderr: %s", name, errors)
	}
	return output
}

func fileDriverGoFormat(test *testing.T, cases string) []byte {
	test.Helper()
	goBinary := goFormatOracle(test)
	expected := fileDriverRun(test, "", nil, goBinary, "--cases", cases)
	if artifacts := os.Getenv("ADAMIC_YAML_ARTIFACTS"); artifacts != "" {
		if createDirectoryError := os.MkdirAll(artifacts, 0755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		executable, readError := os.ReadFile(goBinary)
		if readError != nil {
			test.Fatal(readError)
		}
		inputs, readError := os.ReadFile(cases)
		if readError != nil {
			test.Fatal(readError)
		}
		for _, file := range []struct {
			name string
			data []byte
			mode os.FileMode
		}{
			{"go-format", executable, 0755},
			{"format-cases.txt", inputs, 0644},
			{"format-expected.txt", expected, 0644},
		} {
			if writeError := os.WriteFile(
				filepath.Join(artifacts, file.name),
				file.data,
				file.mode,
			); writeError != nil {
				test.Fatal(writeError)
			}
		}
	}
	return expected
}
