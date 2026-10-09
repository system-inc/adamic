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
		if fn, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "TestDeepMutants_") {
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
		Files:     []string{"internal", "cohere", "stage1/typescript", "go.mod", "go.sum", "go.work", "stage1/cohere/estree/deep_mutants_split_test.go", "stage1/cohere/estree/estree_test.go"},
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

func deepMutantsOracle(t *testing.T) string {
	t.Helper()
	// GoBuild is not on this base. Preserve the existing overlay build command.
	inputs := buildcache.Inputs{Name: "deep-mutants-go-oracle-v1", Files: []string{"cohere", "stage1/cohere/estree/testdata/oracle.go", "go.mod", "go.sum"}, Flags: []string{"overlay", "root=" + root(t)}, Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH}}
	product := buildcache.Product(t, inputs, func(dir string) error {
		data, err := os.ReadFile(goOracle(t))
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "oracle"), data, 0755)
	})
	return filepath.Join(product, "oracle")
}

func deepMutantsShard(t *testing.T, shard int) {
	t.Helper()
	deepMutantsProof(t)
	started := time.Now()
	list := manifest(t, []string{deepMutantsInput})
	want := execute(t, "", deepMutantsOracle(t), "--manifest", list)
	t.Logf("TestDeepMutants (setup): %.3fs", time.Since(started).Seconds())
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
