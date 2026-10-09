package yaml

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
)

const scalarsMatchGoShardCount = 8

// Hash the complete encoded case, including its streaming chunk size. Repeated
// cases keep their multiplicity, and inserting a case cannot move another owner.
func scalarsMatchGoPartition(cases []string) [][]int {
	shards := make([][]int, scalarsMatchGoShardCount)
	for index, input := range cases {
		hash := fnv.New64a()
		_, _ = hash.Write([]byte(input))
		owner := int(hash.Sum64() % scalarsMatchGoShardCount)
		shards[owner] = append(shards[owner], index)
	}
	return shards
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

// Include compiler implementations, imported TS, Go replacements and embedded
// runtime/checker data. Test evidence is not a build input.
func scalarsMatchGoInputs(t *testing.T, name string) buildcache.Inputs {
	t.Helper()
	files := []string{"go.mod", "go.work", "stage1/cohere/yaml/scalars_match_go_shards_test.go"}
	for _, directory := range []string{"internal", "stage1/cohere/yaml", "cohere"} {
		err := filepath.WalkDir(filepath.Join(repository, directory), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				switch entry.Name() {
				case ".git", "node_modules", "testdata", "audit", "performance", "evidence", "logs":
					return filepath.SkipDir
				}
				return nil
			}
			if strings.HasSuffix(path, "_test.go") {
				return nil
			}
			switch filepath.Ext(path) {
			case ".go", ".ts", ".c", ".h", ".mod", ".sum", ".gz":
			default:
				return nil
			}
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				return err
			}
			files = append(files, relative)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return buildcache.Inputs{Name: name, Files: files, Toolchain: []string{runtime.Version()}, Flags: []string{"GOTOOLCHAIN=" + os.Getenv("GOTOOLCHAIN"), "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}}
}

type scalarsMatchGoProduct struct {
	once      sync.Once
	directory string
}

var scalarsMatchGoLowered, scalarsMatchGoNative, scalarsMatchGoGo scalarsMatchGoProduct

func scalarsMatchGoLoweredProduct(t *testing.T) string {
	t.Helper()
	scalarsMatchGoLowered.once.Do(func() {
		inputs := scalarsMatchGoInputs(t, "yaml-scalars-lowered")
		inputs.Flags = append(inputs.Flags, "scalar_main.ts: lower+native.C+javascript.JavaScript")
		scalarsMatchGoLowered.directory = buildcache.Product(t, inputs, func(directory string) error {
			entry, err := filepath.Abs("scalar_main.ts")
			if err != nil {
				return err
			}
			program, err := load.Load([]string{entry})
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			lowered, err := lower.Lower(ctx, program)
			if err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(directory, "scalars.c"), []byte(native.C(lowered)), 0644); err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "scalars.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
		})
	})
	if scalarsMatchGoLowered.directory == "" {
		t.Fatal("scalar lowering did not complete")
	}
	return scalarsMatchGoLowered.directory
}

func scalarsMatchGoNativeProduct(t *testing.T) string {
	t.Helper()
	scalarsMatchGoNative.once.Do(func() {
		lowered := scalarsMatchGoLoweredProduct(t)
		options := native.Options{Sanitize: true}
		inputs := scalarsMatchGoInputs(t, "yaml-scalars-native")
		inputs.Flags = append(inputs.Flags, "scalar_main.ts: lower+native.C+javascript.JavaScript")
		inputs.Flags = append(inputs.Flags, native.Flags(options)...)
		inputs.Flags = append(inputs.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"))
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
		scalarsMatchGoNative.directory = buildcache.Product(t, inputs, func(directory string) error {
			source, err := os.ReadFile(filepath.Join(lowered, "scalars.c"))
			if err != nil {
				return err
			}
			return native.Build(string(source), filepath.Join(directory, "scalars"), options)
		})
	})
	if scalarsMatchGoNative.directory == "" {
		t.Fatal("scalar native product did not complete")
	}
	return scalarsMatchGoNative.directory
}

// Product remains the shared buildcache API until GoBuild is available on main.
// Both the gate declaration and shards call this exact overlay recipe and key.
func scalarsMatchGoGoProduct(t *testing.T) string {
	t.Helper()
	scalarsMatchGoGo.once.Do(func() {
		inputs := scalarsMatchGoInputs(t, "yaml-scalars-go")
		inputs.Files = append(inputs.Files, "stage1/cohere/yaml/testdata/scalar_go.go", "stage1/cohere/yaml/testdata/scalar_exports.txt")
		inputs.Flags = append(inputs.Flags, "go build -overlay scalar_go.go+scalar_exports.txt ./command/formatter_comparison", "CGO_ENABLED="+os.Getenv("CGO_ENABLED"))
		inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("go", "version"))
		scalarsMatchGoGo.directory = buildcache.Product(t, inputs, func(directory string) error {
			root, err := filepath.Abs(repository)
			if err != nil {
				return err
			}
			adapter, err := filepath.Abs("testdata/scalar_go.go")
			if err != nil {
				return err
			}
			numbers := filepath.Join(root, "cohere/internal/format/yaml/compose/numbers.go")
			original, err := os.ReadFile(numbers)
			if err != nil {
				return err
			}
			exports, err := os.ReadFile("testdata/scalar_exports.txt")
			if err != nil {
				return err
			}
			source := strings.Replace(string(original), "import (", "import (\n\"fmt\"\n\"strings\"\n\"github.com/system-inc/cohere/internal/format/yaml/cst\"", 1) + string(exports)
			replacement := filepath.Join(directory, "numbers.go")
			if err := os.WriteFile(replacement, []byte(source), 0644); err != nil {
				return err
			}
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{filepath.Join(root, "cohere/command/formatter_comparison/main.go"): adapter, numbers: replacement}})
			if err != nil {
				return err
			}
			path := filepath.Join(directory, "overlay.json")
			if err := os.WriteFile(path, overlay, 0644); err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, "go", "build", "-overlay", path, "-o", filepath.Join(directory, "go-scalars"), "./command/formatter_comparison")
			command.Dir = filepath.Join(root, "cohere")
			output, err := command.CombinedOutput()
			if err != nil {
				return fmt.Errorf("Go scalar build: %w\n%s", err, output)
			}
			return nil
		})
	})
	if scalarsMatchGoGo.directory == "" {
		t.Fatal("Go scalar product did not complete")
	}
	return scalarsMatchGoGo.directory
}

func TestProduct_YamlScalarsLowered(t *testing.T) { t.Parallel(); scalarsMatchGoLoweredProduct(t) }
func TestProduct_YamlScalarsNative(t *testing.T)  { t.Parallel(); scalarsMatchGoNativeProduct(t) }
func TestProduct_YamlScalarsGo(t *testing.T)      { t.Parallel(); scalarsMatchGoGoProduct(t) }

func scalarsMatchGoRun(t *testing.T, ctx context.Context, environment []string, name string, args ...string) []byte {
	t.Helper()
	command := exec.CommandContext(ctx, name, args...)
	command.Env = append(os.Environ(), environment...)
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
	oracle := scalarsMatchGoGoProduct(t)
	lowered := scalarsMatchGoLoweredProduct(t)
	compiled := scalarsMatchGoNativeProduct(t)
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
	expected := scalarsMatchGoRun(t, ctx, nil, filepath.Join(oracle, "go-scalars"), path)
	runner := filepath.Join(root, "oracle/node.mjs")
	for _, side := range []struct {
		name    string
		env     []string
		command string
		args    []string
	}{
		{"native ASan/UBSan/LSan", []string{"ASAN_OPTIONS=detect_leaks=1"}, filepath.Join(compiled, "scalars"), []string{path}},
		{"Node", nil, "node", []string{"--disable-warning=ExperimentalWarning", runner, entry, path}},
		{"emitted JavaScript", nil, "node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(lowered, "scalars.mjs"), path}},
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
	oracle := scalarsMatchGoGoProduct(t)
	compiled := scalarsMatchGoNativeProduct(t)
	cases, _ := scalarsMatchGoCases(t)
	// Execute a real case through both sides, then corrupt its actual stdout.
	path := scalarsMatchGoWriteCases(t, cases[:1])
	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Second)
	defer cancel()
	expected := scalarsMatchGoRun(t, ctx, nil, filepath.Join(oracle, "go-scalars"), path)
	actual := scalarsMatchGoRun(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=1"}, filepath.Join(compiled, "scalars"), path)
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
