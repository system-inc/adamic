package estree

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/estree/oracles"
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
const deepMutantsInput = oracles.DeepMutantsInput

type deepMutantsMutation struct{ name, file, from, to string }

func deepMutantsEnumeration() []deepMutantsMutation {
	return []deepMutantsMutation{
		{
			"binary-operator",
			"binaryConvert.ts",
			"node.set('operator', stringValue(operator));",
			"node.set('operator', stringValue('-'));",
		},
		{
			"postorder-alias",
			"postprocess.ts",
			"node.set(key, childValue(completed.get(value.node) ?? value.node));",
			"node.set(key, childValue(value.node));",
		},
		{
			"dump-property-order",
			"protocol.ts",
			"for(let index = node.properties.length - 1; index >= 0; index--)",
			"for(let index = 0; index < node.properties.length; index++)",
		},
	}
}

func deepMutantsOwner(index, count int) int { return index * testDeepMutantsShards / count }

// Verify the live union, actual top-level functions, and a planted disagreement
// through the same assignment and byte-comparison predicate as the execution.
func deepMutantsProof(test *testing.T) {
	test.Helper()
	cases := deepMutantsEnumeration()
	file, parseFileError := parser.ParseFile(token.NewFileSet(), "deep_mutants_product_shards_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	names := map[string]bool{}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok &&
			strings.HasPrefix(fn.Name.Name, "TestDeepMutants_") &&
			fn.Name.Name != "TestDeepMutants_Setup" {
			names[fn.Name.Name] = true
		}
	}
	if len(names) != testDeepMutantsShards {
		test.Fatalf("%d functions for %d shards", len(names), testDeepMutantsShards)
	}
	counts := make([]int, len(cases)*2)
	caught := []int{}
	for shard := 0; shard < testDeepMutantsShards; shard++ {
		if !names[fmt.Sprintf("TestDeepMutants_%03d", shard)] {
			test.Fatalf("missing shard %d", shard)
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
			test.Fatalf("case %d visited %d times", index, count)
		}
	}
	if len(caught) != 1 || caught[0] != deepMutantsOwner(1, len(cases)) {
		test.Fatalf("plant caught by %v", caught)
	}
	test.Logf(
		"union: %d mutants, %d mode checks exactly once; planted disagreement caught only by TestDeepMutants_%03d",
		len(cases),
		len(counts),
		caught[0],
	)
}

// Products are addressed by source content, never by shard. Source snapshots live
// in the product (not a shard's TempDir), so parallel users share persistent builds.
func deepMutantsSnapshot(test *testing.T, mutation deepMutantsMutation) (map[string][]byte, buildcache.Inputs) {
	test.Helper()
	path := mutantPort(test, mutation.file, mutation.from, mutation.to)
	files, globError := filepath.Glob(filepath.Join(filepath.Dir(path), "*.ts"))
	if globError != nil {
		test.Fatal(globError)
	}
	snapshot := map[string][]byte{}
	inputs := buildcache.Inputs{
		Name:      "deep-mutants-lowered-v1",
		Files:     estreeFamilyDependencyFiles(test),
		Flags:     append([]string{"root=" + root(test)}, estreeFamilyDependencyFlags(test)...),
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
	for _, file := range files {
		data, readError := os.ReadFile(file)
		if readError != nil {
			test.Fatal(readError)
		}
		name := filepath.Base(file)
		snapshot[name] = data
		// mutantPort writes this checkout's path into the imports; the key hashes them as the tree spells them.
		portable := strings.ReplaceAll(
			string(data),
			filepath.ToSlash(filepath.Join(root(test), "stage1/typescript")),
			"../../typescript",
		)
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("%s=%x", name, sha256.Sum256([]byte(portable))))
	}
	return snapshot, inputs
}

func deepMutantsIRProduct(test *testing.T, mutation deepMutantsMutation) (string, buildcache.Inputs) {
	test.Helper()
	snapshot, inputs := deepMutantsSnapshot(test, mutation)
	// v2: the IR's maps are sorted and its positions placeless, so a v1 program.gob isn't one this reads.
	inputs.Name = "deep-mutants-ir-v2"
	checkpoint := buildcache.Product(test, inputs, func(dir string) error {
		return estreeFamilyLowerCheckpoint(dir, snapshot)
	})
	return checkpoint, inputs
}

func deepMutantsLowered(test *testing.T, mutation deepMutantsMutation) (string, buildcache.Inputs) {
	test.Helper()
	checkpoint, inputs := deepMutantsIRProduct(test, mutation)
	inputs.Name = "deep-mutants-emitted-v1"
	lowered := buildcache.Product(test, inputs, func(dir string) error {
		program, estreeFamilyReadCheckpointError := estreeFamilyReadCheckpoint(checkpoint)
		if estreeFamilyReadCheckpointError != nil {
			return estreeFamilyReadCheckpointError
		}
		if estreeFamilyCopySourceError := estreeFamilyCopySource(checkpoint, dir); estreeFamilyCopySourceError != nil {
			return estreeFamilyCopySourceError
		}
		if writeError := os.WriteFile(
			filepath.Join(dir, "program.c"),
			[]byte(native.C(program)),
			0644,
		); writeError != nil {
			return writeError
		}
		return os.WriteFile(filepath.Join(dir, "port.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
	return lowered, inputs
}

func deepMutantsNative(test *testing.T, mutation deepMutantsMutation) string {
	test.Helper()
	lowered, inputs := deepMutantsLowered(test, mutation)

	data, readError := os.ReadFile(filepath.Join(lowered, "program.c"))
	if readError != nil {
		test.Fatal(readError)
	}
	inputs.Name = "deep-mutants-native-v3"
	for _, name := range []string{"ADAMIC_NATIVE_SPLIT"} {
		inputs.Flags = append(inputs.Flags, name+"="+os.Getenv(name))
	}
	inputs.Flags = append(inputs.Flags, native.Flags(estreeFamilyNativeOptions())...)
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("C=%x", sha256.Sum256(data)))
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	product := buildcache.Product(test, inputs, func(dir string) error {
		return native.Build(string(data), filepath.Join(dir, "port"), estreeFamilyNativeOptions())
	})
	return filepath.Join(product, "port")
}

type deepMutantsReady struct {
	once           sync.Once
	source, native string
}

var deepMutantsPrepared sync.Map

func deepMutantsPrepare(test *testing.T, mutation deepMutantsMutation) *deepMutantsReady {
	test.Helper()
	value, _ := deepMutantsPrepared.LoadOrStore(mutation.name, &deepMutantsReady{})
	ready := value.(*deepMutantsReady)
	ready.once.Do(func() {
		lowered, _ := deepMutantsLowered(test, mutation)
		ready.source = resolvedPort(test, lowered, "source/main.ts")
		ready.native = deepMutantsNative(test, mutation)
	})
	if ready.native == "" {
		test.Fatal("deep mutant preparation failed")
	}
	return ready
}

func deepMutantsFixture(test *testing.T) string {
	test.Helper()
	return oracles.DeepMutantsFixture(test)
}

func TestDeepMutants_Setup(test *testing.T) {
	test.Parallel()
	deepMutantsFixture(test)
}

func TestProduct_EstreeFindsDeepMutantsFixture(test *testing.T) {
	test.Parallel()
	deepMutantsFixture(test)
}
func TestProduct_DeepMutantsLowered0(test *testing.T) {
	test.Parallel()
	deepMutantsLowered(test, deepMutantsEnumeration()[0])
}
func TestProduct_DeepMutantsLowered1(test *testing.T) {
	test.Parallel()
	deepMutantsLowered(test, deepMutantsEnumeration()[1])
}
func TestProduct_DeepMutantsNative1(test *testing.T) {
	test.Parallel()
	deepMutantsNative(test, deepMutantsEnumeration()[1])
}
func TestProduct_DeepMutantsLowered2(test *testing.T) {
	test.Parallel()
	deepMutantsLowered(test, deepMutantsEnumeration()[2])
}
func TestProduct_DeepMutantsNative2(test *testing.T) {
	test.Parallel()
	deepMutantsNative(test, deepMutantsEnumeration()[2])
}

func deepMutantsShard(test *testing.T, shard int) {
	test.Helper()
	deepMutantsProof(test)
	fixture := deepMutantsFixture(test)
	cases := deepMutantsEnumeration()
	products := make(map[int]*deepMutantsReady)
	for index, item := range cases {
		if deepMutantsOwner(index, len(cases)) == shard {
			products[index] = deepMutantsPrepare(test, item)
		}
	}
	// Every selected shard prepares once before starting its own 90-second clock.
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	list := manifest(test, []string{deepMutantsInput})
	want, readError := os.ReadFile(filepath.Join(fixture, "want"))
	if readError != nil {
		test.Fatal(readError)
	}
	for index, item := range cases {
		if deepMutantsOwner(index, len(cases)) != shard {
			continue
		}
		ready := products[index]
		for name, got := range map[string][]byte{
			"Node": threePortExecute(
				test,
				ctx,
				"node",
				"--disable-warning=ExperimentalWarning",
				filepath.Join(root(test), "oracle/node.mjs"),
				ready.source,
				"--manifest",
				list,
			),
			"native": threePortExecute(test, ctx, ready.native, "--manifest", list),
		} {
			if diff := firstDifference(want, got); diff == "" {
				test.Fatal(name + " mutant survived")
			} else {
				test.Log(item.name + " " + name + ": " + diff)
			}
		}
	}
}

func TestDeepMutants_000(test *testing.T) { test.Parallel(); deepMutantsShard(test, 0) }
func TestDeepMutants_001(test *testing.T) { test.Parallel(); deepMutantsShard(test, 1) }
func TestDeepMutants_002(test *testing.T) { test.Parallel(); deepMutantsShard(test, 2) }
