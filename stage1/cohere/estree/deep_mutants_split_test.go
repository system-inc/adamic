package estree

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

const testDeepMutantsShards = 3

// This pinned corpus retains the original compound input and both execution modes.
const deepMutantsInput = "type T=(A); a+b+c; const t=tag`a${b}c`; a&&(b&&c);"

type deepMutantsMutation struct{ name, file, from, to string }

func deepMutantsEnumeration() []deepMutantsMutation {
	return []deepMutantsMutation{
		{"binary-operator", "binaryConvert.ts", "node.set('operator', stringValue(operator));", "node.set('operator', stringValue('-'));"},
		{"postorder-alias", "postprocess.ts", "node.set(key, childValue(completed.get(value.node) ?? value.node));", "node.set(key, childValue(value.node));"},
		{"dump-property-order", "protocol.ts", "for(let index = node.properties.length - 1; index >= 0; index--)", "for(let index = 0; index < node.properties.length; index++)"},
	}
}

func deepMutantsOwner(index, count int) int { return index * testDeepMutantsShards / count }

// Verify the live union, actual top-level functions, and a planted disagreement
// through the same assignment and byte-comparison predicate as the execution.
func deepMutantsProof(t *testing.T) {
	t.Helper()
	cases := deepMutantsEnumeration()
	file, err := parser.ParseFile(token.NewFileSet(), "deep_mutants_split_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	names := map[string]bool{}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "TestDeepMutants_") && fn.Name.Name != "TestDeepMutants_Setup" {
			names[fn.Name.Name] = true
		}
	}
	if len(names) != testDeepMutantsShards {
		t.Fatalf("%d functions for %d shards", len(names), testDeepMutantsShards)
	}
	counts := make([]int, len(cases)*2)
	caught := []int{}
	for shard := 0; shard < testDeepMutantsShards; shard++ {
		if !names[fmt.Sprintf("TestDeepMutants_%03d", shard)] {
			t.Fatalf("missing shard %d", shard)
		}
		for index := range cases {
			if deepMutantsOwner(index, len(cases)) != shard {
				continue
			}
			for mode := 0; mode < 2; mode++ {
				counts[index*2+mode]++
			}
			want, got := []byte("oracle"), []byte("oracle")
			if index == 1 {
				got = []byte("planted disagreement")
			}
			if firstDifference(want, got) != "" {
				caught = append(caught, shard)
			}
		}
	}
	for index, count := range counts {
		if count != 1 {
			t.Fatalf("case %d visited %d times", index, count)
		}
	}
	if len(caught) != 1 || caught[0] != deepMutantsOwner(1, len(cases)) {
		t.Fatalf("plant caught by %v", caught)
	}
	t.Logf("union: %d mutants, %d mode checks exactly once; planted disagreement caught only by TestDeepMutants_%03d", len(cases), len(counts), caught[0])
}

// Products are addressed by source content, never by shard. Source snapshots live
// in the product (not a shard's TempDir), so parallel users share persistent builds.
func deepMutantsProducts(t *testing.T, mutation deepMutantsMutation) (string, string) {
	t.Helper()
	path := mutantPort(t, mutation.file, mutation.from, mutation.to)
	files, err := filepath.Glob(filepath.Join(filepath.Dir(path), "*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := map[string][]byte{}
	inputs := buildcache.Inputs{
		Name:      "deep-mutants-lowered-v1",
		Files:     []string{"internal", "cohere", "stage1/typescript", "go.mod", "go.work", "stage1/cohere/estree/deep_mutants_split_test.go", "stage1/cohere/estree/estree_test.go"},
		Flags:     []string{"root=" + root(t)},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		name := filepath.Base(file)
		snapshot[name] = data
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("%s=%x", name, sha256.Sum256(data)))
	}
	lowered := buildcache.Product(t, inputs, func(dir string) error {
		source := filepath.Join(dir, "source")
		if err := os.Mkdir(source, 0755); err != nil {
			return err
		}
		for name, data := range snapshot {
			if err := os.WriteFile(filepath.Join(source, name), data, 0644); err != nil {
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
		if err := os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(ir)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "port.mjs"), []byte(javascript.JavaScript(ir)), 0644)
	})
	data, err := os.ReadFile(filepath.Join(lowered, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "deep-mutants-native-v1"
	for _, name := range []string{"ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED"} {
		inputs.Flags = append(inputs.Flags, name+"="+os.Getenv(name))
	}
	inputs.Flags = append(inputs.Flags, native.Flags(native.Options{Sanitize: true})...)
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("C=%x", sha256.Sum256(data)))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	product := buildcache.Product(t, inputs, func(dir string) error {
		return native.Build(string(data), filepath.Join(dir, "port"), native.Options{Sanitize: true})
	})
	return filepath.Join(lowered, "source/main.ts"), filepath.Join(product, "port")
}

func deepMutantsOracleInputs(t *testing.T) buildcache.Inputs {
	t.Helper()
	return buildcache.Inputs{Name: "deep-mutants-go-oracle-v1", Files: []string{"cohere", "stage1/cohere/estree/testdata/oracle.go", "stage1/cohere/estree/estree_test.go", "go.mod"}, Flags: []string{"overlay", "root=" + root(t)}, Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH}}
}

func deepMutantsOracle(t *testing.T) string {
	t.Helper()
	// GoBuild is not on this base. Preserve the existing overlay build command.
	product := buildcache.Product(t, deepMutantsOracleInputs(t), func(dir string) error {
		data, err := os.ReadFile(goOracle(t))
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "oracle"), data, 0755)
	})
	return filepath.Join(product, "oracle")
}

// Only the serial setup test writes this, before parallel tests resume.
var deepMutantsFixtureDirectory string

func deepMutantsFixture(t *testing.T, prepare bool) string {
	t.Helper()
	if path := os.Getenv("ADAMIC_DEEP_MUTANTS_FIXTURE"); path != "" {
		return path
	}
	if deepMutantsFixtureDirectory != "" {
		return deepMutantsFixtureDirectory
	}
	inputs := deepMutantsOracleInputs(t)
	inputs.Name = "deep-mutants-oracle-output-v1"
	inputs.Files = append(inputs.Files, "stage1/cohere/estree/deep_mutants_split_test.go")
	inputs.Flags = append(inputs.Flags, deepMutantsInput)
	return buildcache.Product(t, inputs, func(dir string) error {
		if !prepare {
			return fmt.Errorf("shared fixture is not ready: run TestDeepMutants_Setup first")
		}
		list := manifest(t, []string{deepMutantsInput})
		want := execute(t, "", deepMutantsOracle(t), "--manifest", list)
		return os.WriteFile(filepath.Join(dir, "want"), want, 0644)
	})
}

// The deadline starts only after the fixture is ready. The worker process owns
// its descendants, including Go/clang builders, so cancellation kills them too.
func deepMutantsBounded(t *testing.T, fixture string) []byte {
	t.Helper()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary, "-test.run=^"+t.Name()+"$", "-test.timeout=90s", "-test.v")
	command.Env = append(os.Environ(), "ADAMIC_DEEP_MUTANTS_WORKER="+t.Name(), "ADAMIC_DEEP_MUTANTS_FIXTURE="+fixture)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	command.WaitDelay = time.Second
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("%s worker: %v (deadline: %v)\n%s", t.Name(), err, ctx.Err(), output)
	}
	t.Logf("%s", output)
	return output
}

// Not parallel: publishes the shared oracle fixture before parallel shards resume.
func TestDeepMutants_Setup(t *testing.T) {
	if os.Getenv("ADAMIC_DEEP_MUTANTS_WORKER") == t.Name() {
		directory := deepMutantsFixture(t, true)
		fmt.Fprintln(os.Stdout, "DEEP_MUTANTS_FIXTURE="+directory)
		return
	}
	output := deepMutantsBounded(t, "")
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "DEEP_MUTANTS_FIXTURE=") {
			deepMutantsFixtureDirectory = strings.TrimPrefix(line, "DEEP_MUTANTS_FIXTURE=")
			return
		}
	}
	t.Fatal("setup worker did not publish its fixture")
}

func deepMutantsShard(t *testing.T, shard int) {
	t.Helper()
	deepMutantsProof(t)
	fixture := deepMutantsFixture(t, false)
	if os.Getenv("ADAMIC_DEEP_MUTANTS_WORKER") != t.Name() {
		deepMutantsBounded(t, fixture)
		return
	}
	list := manifest(t, []string{deepMutantsInput})
	want, err := os.ReadFile(filepath.Join(fixture, "want"))
	if err != nil {
		t.Fatal(err)
	}
	cases := deepMutantsEnumeration()
	for index, item := range cases {
		if deepMutantsOwner(index, len(cases)) != shard {
			continue
		}
		main, binary := deepMutantsProducts(t, item)
		for name, got := range map[string][]byte{"Node": onNode(t, main, "--manifest", list), "native": execute(t, "", binary, "--manifest", list)} {
			if diff := firstDifference(want, got); diff == "" {
				t.Fatal(name + " mutant survived")
			} else {
				t.Log(item.name + " " + name + ": " + diff)
			}
		}
	}
}

func TestDeepMutants_000(t *testing.T) { t.Parallel(); deepMutantsShard(t, 0) }
func TestDeepMutants_001(t *testing.T) { t.Parallel(); deepMutantsShard(t, 1) }
func TestDeepMutants_002(t *testing.T) { t.Parallel(); deepMutantsShard(t, 2) }
