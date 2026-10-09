package estree

import (
	"context"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
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
	file, err := parser.ParseFile(token.NewFileSet(), "deep_mutants_product_shards_test.go", nil, 0)
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
func deepMutantsSnapshot(t *testing.T, mutation deepMutantsMutation) (map[string][]byte, buildcache.Inputs) {
	t.Helper()
	path := mutantPort(t, mutation.file, mutation.from, mutation.to)
	files, err := filepath.Glob(filepath.Join(filepath.Dir(path), "*.ts"))
	if err != nil {
		t.Fatal(err)
	}
	snapshot := map[string][]byte{}
	inputs := buildcache.Inputs{
		Name:      "deep-mutants-lowered-v1",
		Files:     []string{"internal", "cohere", "stage1/typescript", "go.mod", "go.work", "stage1/cohere/estree/deep_mutants_product_shards_test.go", "stage1/cohere/estree/estree_test.go", "stage1/cohere/estree/loom_family_products_test.go"},
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
	return snapshot, inputs
}

func deepMutantsIRProduct(t *testing.T, mutation deepMutantsMutation) (string, buildcache.Inputs) {
	t.Helper()
	snapshot, inputs := deepMutantsSnapshot(t, mutation)
	inputs.Name = "deep-mutants-ir-v1"
	checkpoint := buildcache.Product(t, inputs, func(dir string) error {
		return estreeFamilyLowerCheckpoint(dir, snapshot)
	})
	return checkpoint, inputs
}

func deepMutantsLowered(t *testing.T, mutation deepMutantsMutation) (string, buildcache.Inputs) {
	t.Helper()
	checkpoint, inputs := deepMutantsIRProduct(t, mutation)
	inputs.Name = "deep-mutants-emitted-v1"
	lowered := buildcache.Product(t, inputs, func(dir string) error {
		program, err := estreeFamilyReadCheckpoint(checkpoint)
		if err != nil {
			return err
		}
		if err := estreeFamilyCopySource(checkpoint, dir); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, "program.c"), []byte(native.C(program)), 0644); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "port.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	return lowered, inputs
}

func deepMutantsNative(t *testing.T, mutation deepMutantsMutation) string {
	t.Helper()
	lowered, inputs := deepMutantsLowered(t, mutation)

	data, err := os.ReadFile(filepath.Join(lowered, "program.c"))
	if err != nil {
		t.Fatal(err)
	}
	inputs.Name = "deep-mutants-native-v3"
	for _, name := range []string{"ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "ADAMIC_GATE_UNCACHED"} {
		inputs.Flags = append(inputs.Flags, name+"="+os.Getenv(name))
	}
	inputs.Flags = append(inputs.Flags, native.Flags(estreeFamilyNativeOptions())...)
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("C=%x", sha256.Sum256(data)))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	product := buildcache.Product(t, inputs, func(dir string) error {
		return native.Build(string(data), filepath.Join(dir, "port"), estreeFamilyNativeOptions())
	})
	return filepath.Join(product, "port")
}

type deepMutantsReady struct {
	once           sync.Once
	source, native string
}

var deepMutantsPrepared sync.Map

func deepMutantsPrepare(t *testing.T, mutation deepMutantsMutation) *deepMutantsReady {
	t.Helper()
	value, _ := deepMutantsPrepared.LoadOrStore(mutation.name, &deepMutantsReady{})
	ready := value.(*deepMutantsReady)
	ready.once.Do(func() {
		lowered, _ := deepMutantsLowered(t, mutation)
		ready.source = filepath.Join(lowered, "source/main.ts")
		ready.native = deepMutantsNative(t, mutation)
	})
	if ready.native == "" {
		t.Fatal("deep mutant preparation failed")
	}
	return ready
}

var deepMutantsFixturePrepared struct {
	once      sync.Once
	directory string
}

func deepMutantsFixture(t *testing.T) string {
	t.Helper()
	deepMutantsFixturePrepared.once.Do(func() {
		oracle := threePortOracleProduct(t)
		inputs := buildcache.Inputs{
			Name:      "deep-mutants-oracle-output-v2",
			Files:     []string{"stage1/cohere/estree/deep_mutants_product_shards_test.go"},
			Flags:     []string{deepMutantsInput, "oracle=" + miscDigest(t, oracle)},
			Toolchain: []string{runtime.Version()},
		}
		deepMutantsFixturePrepared.directory = buildcache.Product(t, inputs, func(dir string) error {
			list := manifest(t, []string{deepMutantsInput})
			// Shared setup has no deadline of its own. Loom bounds the entire unit.
			output := threePortExecute(t, context.Background(), oracle, "--manifest", list)
			return os.WriteFile(filepath.Join(dir, "want"), output, 0644)
		})
	})
	if deepMutantsFixturePrepared.directory == "" {
		t.Fatal("deep fixture preparation failed")
	}
	return deepMutantsFixturePrepared.directory
}

func TestDeepMutants_Setup(t *testing.T) {
	t.Parallel()
	deepMutantsFixture(t)
}

func TestProduct_DeepMutantsFixture(t *testing.T) {
	t.Parallel()
	deepMutantsFixture(t)
}
func TestProduct_DeepMutantsLowered0(t *testing.T) {
	t.Parallel()
	deepMutantsLowered(t, deepMutantsEnumeration()[0])
}
func TestProduct_DeepMutantsLowered1(t *testing.T) {
	t.Parallel()
	deepMutantsLowered(t, deepMutantsEnumeration()[1])
}
func TestProduct_DeepMutantsNative1(t *testing.T) {
	t.Parallel()
	deepMutantsNative(t, deepMutantsEnumeration()[1])
}
func TestProduct_DeepMutantsLowered2(t *testing.T) {
	t.Parallel()
	deepMutantsLowered(t, deepMutantsEnumeration()[2])
}
func TestProduct_DeepMutantsNative2(t *testing.T) {
	t.Parallel()
	deepMutantsNative(t, deepMutantsEnumeration()[2])
}

func deepMutantsShard(t *testing.T, shard int) {
	t.Helper()
	deepMutantsProof(t)
	fixture := deepMutantsFixture(t)
	cases := deepMutantsEnumeration()
	products := make(map[int]*deepMutantsReady)
	for index, item := range cases {
		if deepMutantsOwner(index, len(cases)) == shard {
			products[index] = deepMutantsPrepare(t, item)
		}
	}
	// Every selected shard prepares once before starting its own 90-second clock.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	list := manifest(t, []string{deepMutantsInput})
	want, err := os.ReadFile(filepath.Join(fixture, "want"))
	if err != nil {
		t.Fatal(err)
	}
	for index, item := range cases {
		if deepMutantsOwner(index, len(cases)) != shard {
			continue
		}
		ready := products[index]
		for name, got := range map[string][]byte{
			"Node":   threePortExecute(t, ctx, "node", "--disable-warning=ExperimentalWarning", filepath.Join(root(t), "oracle/node.mjs"), ready.source, "--manifest", list),
			"native": threePortExecute(t, ctx, ready.native, "--manifest", list),
		} {
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
