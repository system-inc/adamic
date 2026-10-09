package estree

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"go/ast"
	"go/parser"
	"go/token"
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
)

const testThreePortMutantsShards = 12
const threePortCasesPerShard = 20

type threePortMutation struct{ name, file, from, to, witness string }

func threePortMutations() []threePortMutation {
	return []threePortMutation{
		{"member-computed", "convert.ts", "boolValue(node.kind === 'ElementAccessExpression')", "boolValue(node.kind === 'PropertyAccessExpression')", "a.b; a[b];"},
		{"logical-rebalance", "postprocess.ts", "completed.set(id, this.rebalance(id));", "completed.set(id, id);", "a && (b && (c && d));"},
		{"merged-jsdoc-value", "postprocess.ts", "*//*", "*/ /*", "x; /**\n * one\n *//**\n * two\n */ y;"},
	}
}

type threePortCase struct {
	mutant, index int
	witness       bool
}
type threePortShard struct {
	name  string
	cases []threePortCase
}

func threePortShards(t *testing.T) []threePortShard {
	file, err := parser.ParseFile(token.NewFileSet(), "three_port_mutants_split_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, decl := range file.Decls {
		if f, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(f.Name.Name, "TestThreePortMutants_") && len(f.Name.Name) == len("TestThreePortMutants_000") {
			names[f.Name.Name] = true
		}
	}
	if len(names) != testThreePortMutantsShards {
		t.Fatalf("top-level enumeration: %d, declared %d", len(names), testThreePortMutantsShards)
	}
	for i := 0; i < testThreePortMutantsShards; i++ {
		if !names[fmt.Sprintf("TestThreePortMutants_%03d", i)] {
			t.Fatalf("missing shard %03d", i)
		}
	}
	return threePortShardPlan(t, generated(), threePortMutations(), testThreePortMutantsShards)
}

func threePortShardPlan(t *testing.T, cases []string, mutations []threePortMutation, declared int) []threePortShard {
	t.Helper()
	var shards []threePortShard
	var unsplit []threePortCase
	for m, mutation := range mutations {
		witnesses := 0
		var pairs []threePortCase
		for i, text := range cases {
			pair := threePortCase{m, i, strings.Contains(text, mutation.witness)}
			if pair.witness {
				witnesses++
			}
			pairs = append(pairs, pair)
			unsplit = append(unsplit, pair)
		}
		if witnesses != 1 {
			t.Fatalf("%s: expected one witness, got %d", mutation.name, witnesses)
		}
		for start := 0; start < len(pairs); start += threePortCasesPerShard {
			end := min(start+threePortCasesPerShard, len(pairs))
			shards = append(shards, threePortShard{fmt.Sprintf("shard-%03d", len(shards)), pairs[start:end]})
		}
	}
	if len(shards) != declared {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), declared)
	}
	want := make(map[[2]int]bool)
	for _, pair := range unsplit {
		id := [2]int{pair.mutant, pair.index}
		if want[id] {
			t.Fatalf("duplicate unsplit case %v", id)
		}
		want[id] = true
	}
	seen := make(map[[2]int]bool)
	count := 0
	for _, shard := range shards {
		for _, pair := range shard.cases {
			id := [2]int{pair.mutant, pair.index}
			if !want[id] || seen[id] {
				t.Fatalf("%s: extra or repeated case %v", shard.name, id)
			}
			seen[id] = true
			count++
		}
	}
	if count != len(unsplit) || len(seen) != len(want) {
		t.Fatalf("shard union: %d/%d cases, %d/%d ids", count, len(unsplit), len(seen), len(want))
	}
	for id := range want {
		if !seen[id] {
			t.Fatalf("shard union missing %v", id)
		}
	}
	t.Logf("shard union: %d pairs, %d distinct ids, %d mutants x %d cases", count, len(seen), len(mutations), len(cases))
	return shards
}

// Canonical output ends every whole-file record with one escaped "stripped"
// line. Do not divide a file's AST, comments, or stripped-source checks.
func threePortCanonicalRecords(t *testing.T, data []byte, count int) [][]byte {
	t.Helper()
	var records [][]byte
	start, offset := 0, 0
	for _, line := range bytes.SplitAfter(data, []byte("\n")) {
		offset += len(line)
		if bytes.HasPrefix(line, []byte("stripped ")) {
			records = append(records, data[start:offset])
			start = offset
		}
	}
	if start != len(data) || len(records) != count {
		t.Fatalf("canonical records: %d, want %d; trailing bytes %d", len(records), count, len(data)-start)
	}
	return records
}

func checkThreePortCase(pair threePortCase, want, source, native []byte) error {
	sourceDifference := firstDifference(want, source)
	nativeDifference := firstDifference(want, native)
	if pair.witness {
		if sourceDifference == "" {
			return fmt.Errorf("case %03d: source Node mutant survived", pair.index)
		}
		if nativeDifference == "" {
			return fmt.Errorf("case %03d: sanitized native mutant survived", pair.index)
		}
	}
	if (sourceDifference == "") != (nativeDifference == "") {
		return fmt.Errorf("case %03d: mutant ports differ in agreement with Go", pair.index)
	}
	if diff := firstDifference(source, native); diff != "" {
		return fmt.Errorf("case %03d: mutant ports disagree: %s", pair.index, diff)
	}
	return nil
}

// Products are addressed by live source content and built once across shard processes.
func threePortProduct(t *testing.T, mutation threePortMutation) (string, string) {
	t.Helper()
	inputs := buildcache.Inputs{Name: "three-port-mutant-lowered-" + mutation.name,
		Files:     []string{"stage1/cohere/estree", "stage1/typescript", "internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "go.mod", "go.work"},
		Flags:     []string{mutation.file, mutation.from, mutation.to, "repository=" + root(t)},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH}}
	lowered := buildcache.Product(t, inputs, func(dir string) error {
		source := filepath.Join(dir, "source")
		if err := os.Mkdir(source, 0755); err != nil {
			return err
		}
		files, err := filepath.Glob("*.ts")
		if err != nil {
			return err
		}
		for _, name := range files {
			data, err := os.ReadFile(name)
			if err != nil {
				return err
			}
			text := string(data)
			if name == mutation.file {
				if strings.Count(text, mutation.from) != 1 {
					return fmt.Errorf("mutant anchor count changed")
				}
				text = strings.Replace(text, mutation.from, mutation.to, 1)
			}
			text = strings.ReplaceAll(text, "'../../typescript/", "'"+filepath.ToSlash(filepath.Join(root(t), "stage1/typescript"))+"/")
			if err := os.WriteFile(filepath.Join(source, name), []byte(text), 0644); err != nil {
				return err
			}
		}
		program, err := load.Load([]string{filepath.Join(source, "main.ts")})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(ir)), 0644)
	})
	inputs.Name = "three-port-mutant-native-" + mutation.name
	inputs.Flags = append(inputs.Flags, "split=true")
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true, Split: true})...)
	inputs.Flags = append(inputs.Flags, "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS="+os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED="+os.Getenv("ADAMIC_GATE_UNCACHED"))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	binary := buildcache.Product(t, inputs, func(dir string) error {
		c, err := os.ReadFile(filepath.Join(lowered, "program.c"))
		if err != nil {
			return err
		}
		return native.Build(string(c), filepath.Join(dir, "port"), native.Options{Sanitize: true, Split: true})
	})
	return filepath.Join(lowered, "source", "main.ts"), filepath.Join(binary, "port")
}

type threePortPreparedPort struct{ Source, Native string }
type threePortPrepared struct {
	Oracle string
	Ports  []threePortPreparedPort
}

// Assigned by the serial setup test before parallel leaves are released. A
// separately selected leaf reads the published setup product without building.
var threePortPreparedForRun *threePortPrepared

func threePortSetupInputs(t *testing.T) buildcache.Inputs {
	t.Helper()
	return buildcache.Inputs{Name: "three-port-mutants-setup-v1",
		Files:     []string{"stage1/cohere/estree", "stage1/typescript", "internal", "cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "cohere/go.mod", "cohere/go.sum", "go.mod", "go.work"},
		Flags:     []string{"repository=" + root(t), "sanitize=true", "split=true", "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "ADAMIC_NATIVE_JOBS=" + os.Getenv("ADAMIC_NATIVE_JOBS"), "ADAMIC_GATE_UNCACHED=" + os.Getenv("ADAMIC_GATE_UNCACHED"), "GOTOOLCHAIN=" + os.Getenv("GOTOOLCHAIN"), "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED")},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("go", "version"), buildcache.Tool("clang", "--version")}}
}

func threePortReadPrepared(t *testing.T, path string) *threePortPrepared {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var prepared threePortPrepared
	if err := json.Unmarshal(data, &prepared); err != nil {
		t.Fatal(err)
	}
	if len(prepared.Ports) != len(threePortMutations()) {
		t.Fatal("incomplete shared setup")
	}
	paths := []string{prepared.Oracle}
	for _, port := range prepared.Ports {
		paths = append(paths, port.Source, port.Native)
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("shared setup product missing: %v", err)
		}
	}
	return &prepared
}

func threePortReady(t *testing.T) *threePortPrepared {
	t.Helper()
	if threePortPreparedForRun != nil {
		return threePortPreparedForRun
	}
	directory := buildcache.Product(t, threePortSetupInputs(t), func(string) error {
		return fmt.Errorf("run TestThreePortMutants_Setup before shards; a leaf never prepares shared state")
	})
	return threePortReadPrepared(t, filepath.Join(directory, "ready.json"))
}

func threePortCommand(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if err == syscall.ESRCH {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = 2 * time.Second
	return command
}

func threePortExecute(t *testing.T, ctx context.Context, name string, args ...string) []byte {
	t.Helper()
	command := threePortCommand(ctx, name, args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if ctx.Err() != nil {
		t.Fatalf("cooked: %s exceeded leaf's 90s deadline", t.Name())
	}
	if err != nil || stderr.Len() != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	return output
}

// Not parallel: publishes immutable shared products before parallel shard tests run
func TestThreePortMutants_Setup(t *testing.T) {
	if report := os.Getenv("ADAMIC_THREE_PORT_SETUP_CHILD"); report != "" {
		directory := buildcache.Product(t, threePortSetupInputs(t), func(directory string) error {
			prepared := threePortPrepared{Ports: make([]threePortPreparedPort, len(threePortMutations()))}
			var workers sync.WaitGroup
			for m, mutation := range threePortMutations() {
				workers.Add(1)
				go func(m int, mutation threePortMutation) {
					defer workers.Done()
					source, binary := threePortProduct(t, mutation)
					prepared.Ports[m] = threePortPreparedPort{source, binary}
				}(m, mutation)
			}
			// GoBuild is not on this base. Preserve the overlay build but publish its
			// binary through buildcache, within the setup child's process-group deadline.
			oracleInputs := threePortSetupInputs(t)
			oracleInputs.Name = "three-port-mutants-go-oracle-v1"
			oracleDirectory := buildcache.Product(t, oracleInputs, func(dir string) error {
				binary := goOracle(t)
				data, err := os.ReadFile(binary)
				if err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(dir, "oracle"), data, 0755)
			})
			prepared.Oracle = filepath.Join(oracleDirectory, "oracle")
			workers.Wait()
			if t.Failed() {
				return fmt.Errorf("shared product build failed")
			}
			data, err := json.Marshal(prepared)
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(directory, "ready.json"), data, 0644)
		})
		data, err := os.ReadFile(filepath.Join(directory, "ready.json"))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(report, data, 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(t.TempDir(), "ready.json")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := threePortCommand(ctx, executable, "-test.run=^TestThreePortMutants_Setup$", "-test.timeout=90s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_THREE_PORT_SETUP_CHILD="+report)
	output, err := command.CombinedOutput()
	t.Logf("shared setup child:\n%s", output)
	if ctx.Err() != nil {
		t.Fatal("cooked: TestThreePortMutants_Setup exceeded 90s deadline")
	}
	if err != nil {
		t.Fatalf("shared setup: %v", err)
	}
	threePortPreparedForRun = threePortReadPrepared(t, report)
}

func TestThreePortMutants_SetupRequired(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := threePortCommand(ctx, executable, "-test.run=^TestThreePortMutants_000$", "-test.timeout=90s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_BUILD_CACHE=", "ADAMIC_BUILD_CACHE_DIR="+t.TempDir(), "ADAMIC_THREE_PORT_PROOF=", "ADAMIC_THREE_PORT_SETUP_CHILD=")
	output, err := command.CombinedOutput()
	exit, ok := err.(*exec.ExitError)
	if ctx.Err() != nil || !ok || exit.ExitCode() != 1 || !strings.Contains(string(output), "a leaf never prepares shared state") {
		t.Fatalf("cold leaf must refuse setup instead of building: %v\n%s", err, output)
	}
	t.Log("cold standalone shard refused unprepared shared state without building")
}

func runThreePortShard(t *testing.T, index int) {
	t.Helper()
	shards := threePortShards(t)
	shard := shards[index]
	if planted := os.Getenv("ADAMIC_THREE_PORT_PROOF"); planted != "" {
		m, err := strconv.Atoi(planted)
		if err != nil {
			t.Fatal(err)
		}
		for _, pair := range shard.cases {
			want, source, native := []byte("oracle"), []byte("oracle"), []byte("oracle")
			if pair.witness {
				source, native = []byte("mutant"), []byte("mutant")
			}
			if pair.mutant == m && pair.witness {
				native = want
			}
			if err := checkThreePortCase(pair, want, source, native); err != nil {
				t.Fatal(err)
			}
		}
		return
	}
	// Readiness is resolved before the leaf deadline. This path cannot build.
	prepared := threePortReady(t)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	m := shard.cases[0].mutant
	mutation := threePortMutations()[m]
	oracle := prepared.Oracle
	sourcePath, binary := prepared.Ports[m].Source, prepared.Ports[m].Native
	texts := make([]string, len(shard.cases))
	for i, pair := range shard.cases {
		texts[i] = generated()[pair.index]
	}
	list := manifest(t, texts)
	want := threePortCanonicalRecords(t, threePortExecute(t, ctx, oracle, "--manifest", list), len(texts))
	source := threePortCanonicalRecords(t, threePortExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), sourcePath, "--manifest", list), len(texts))
	native := threePortCanonicalRecords(t, threePortExecute(t, ctx, binary, "--manifest", list), len(texts))
	for i, pair := range shard.cases {
		if err := checkThreePortCase(pair, want[i], source[i], native[i]); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("%s: %d cases", mutation.name, len(texts))
}

func TestThreePortMutants_000(t *testing.T) { t.Parallel(); runThreePortShard(t, 0) }
func TestThreePortMutants_001(t *testing.T) { t.Parallel(); runThreePortShard(t, 1) }
func TestThreePortMutants_002(t *testing.T) { t.Parallel(); runThreePortShard(t, 2) }
func TestThreePortMutants_003(t *testing.T) { t.Parallel(); runThreePortShard(t, 3) }
func TestThreePortMutants_004(t *testing.T) { t.Parallel(); runThreePortShard(t, 4) }
func TestThreePortMutants_005(t *testing.T) { t.Parallel(); runThreePortShard(t, 5) }
func TestThreePortMutants_006(t *testing.T) { t.Parallel(); runThreePortShard(t, 6) }
func TestThreePortMutants_007(t *testing.T) { t.Parallel(); runThreePortShard(t, 7) }
func TestThreePortMutants_008(t *testing.T) { t.Parallel(); runThreePortShard(t, 8) }
func TestThreePortMutants_009(t *testing.T) { t.Parallel(); runThreePortShard(t, 9) }
func TestThreePortMutants_010(t *testing.T) { t.Parallel(); runThreePortShard(t, 10) }
func TestThreePortMutants_011(t *testing.T) { t.Parallel(); runThreePortShard(t, 11) }

// Plant a surviving native witness and execute each actual top-level shard
// independently. This proof performs no builds and never measures the whole test.
func TestThreePortMutants_ShardProof(t *testing.T) {
	t.Parallel()
	shards := threePortShards(t)
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for m, mutation := range threePortMutations() {
		failed, caught, expected, caseID := 0, -1, -1, -1
		for s, shard := range shards {
			for _, pair := range shard.cases {
				if pair.mutant == m && pair.witness {
					expected = s
					caseID = pair.index
				}
			}
		}
		for s := range shards {
			name := fmt.Sprintf("TestThreePortMutants_%03d", s)
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			command := threePortCommand(ctx, executable, "-test.run=^"+name+"$", "-test.timeout=90s", "-test.v")
			command.Env = append(os.Environ(), fmt.Sprintf("ADAMIC_THREE_PORT_PROOF=%d", m))
			output, err := command.CombinedOutput()
			deadlineErr := ctx.Err()
			cancel()
			if deadlineErr != nil {
				t.Fatalf("proof child exceeded 90s: %v", deadlineErr)
			}
			if err != nil {
				exit, ok := err.(*exec.ExitError)
				if !ok || exit.ExitCode() != 1 {
					t.Fatalf("proof child: %v: %s", err, output)
				}
				failed++
				caught = s
				if s != expected || !strings.Contains(string(output), fmt.Sprintf("case %03d: sanitized native mutant survived", caseID)) {
					t.Fatalf("unexpected failure in %s: %s", name, output)
				}
			}
		}
		if failed != 1 || caught != expected {
			t.Fatalf("planted %s: %d failing shards, caught %d want %d", mutation.name, failed, caught, expected)
		}
		t.Logf("planted surviving native %s case %03d caught by TestThreePortMutants_%03d; exactly one shard", mutation.name, caseID, caught)
	}
}
