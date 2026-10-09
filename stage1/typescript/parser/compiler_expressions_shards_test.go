package parser

import (
	"bytes"
	"context"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testCompilerExpressionsAgreeShards = 16

func compilerExpressionsPartition(t *testing.T, paths []string) [][]string {
	t.Helper()
	shards := make([][]string, testCompilerExpressionsAgreeShards)
	seen := make(map[string]int)
	for _, path := range paths {
		if path == "" {
			continue
		}
		// Compiler filenames are stable even when the checkout moves.
		key := filepath.ToSlash(filepath.Base(path))
		hash := fnv.New32a()
		hash.Write([]byte(key))
		index := int(hash.Sum32() % uint32(len(shards)))
		shards[index] = append(shards[index], path)
		seen[path]++
	}
	count := 0
	for _, shard := range shards {
		for _, path := range shard {
			if seen[path] != 1 {
				t.Fatalf("case %q occurs %d times", path, seen[path])
			}
			count++
		}
	}
	if count != len(paths) {
		t.Fatalf("union %d != enumeration %d", count, len(paths))
	}
	return shards
}

func compilerExpressionsNative(t *testing.T, absolute string) string {
	t.Helper()
	files := []string{"stage1/typescript/parser", "stage1/typescript/scanner", "internal/load", "internal/lower", "internal/ir", "internal/native", "cohere", "go.mod"}
	tools := []string{buildcache.Tool("clang", "--version"), buildcache.Tool("go", "version")}
	lowered := buildcache.Product(t, buildcache.Inputs{Name: "compiler-expressions-lowered", Files: files, Flags: []string{"native.C"}, Toolchain: tools}, func(dir string) error {
		program, err := load.Load([]string{filepath.Join(absolute, "main.ts")})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(ir)), 0644)
	})
	flags := append(native.Flags(native.Options{Sanitize: true}), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"))
	product := buildcache.Product(t, buildcache.Inputs{Name: "compiler-expressions-native", Files: files, Flags: flags, Toolchain: tools}, func(dir string) error {
		source, err := os.ReadFile(filepath.Join(lowered, "main.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(dir, "parser"), native.Options{Sanitize: true})
	})
	return filepath.Join(product, "parser")
}

var compilerExpressionsProducts struct {
	once                     sync.Once
	oracle, binary, absolute string
}

func compilerExpressionsSetup(t *testing.T) (string, string, string) {
	t.Helper()
	compilerExpressionsProducts.once.Do(func() {
		started := time.Now()
		timer := time.AfterFunc(90*time.Second, func() { panic("cooked: TestCompilerExpressionsAgree_Setup exceeded 90s") })
		defer timer.Stop()
		// Preserve the pinned corpus prerequisite even for a setup-only invocation.
		compilerManifest(t)
		absolute, err := filepath.Abs(".")
		if err != nil {
			t.Fatal(err)
		}
		product := buildcache.Product(t, buildcache.Inputs{Name: "compiler-expressions-go-oracle", Files: []string{"cohere", "stage1/typescript/parser/testdata/oracle.go", "go.mod", "go.work"}, Toolchain: []string{buildcache.Tool("go", "version")}, Flags: []string{"go build -overlay"}}, func(dir string) error {
			data, err := os.ReadFile(goOracle(t))
			if err != nil {
				return err
			}
			return os.WriteFile(filepath.Join(dir, "oracle"), data, 0755)
		})
		compilerExpressionsProducts.oracle = filepath.Join(product, "oracle")
		compilerExpressionsProducts.absolute = absolute
		compilerExpressionsProducts.binary = compilerExpressionsNative(t, absolute)
		t.Logf("TestCompilerExpressionsAgree_Setup: %.3fs", time.Since(started).Seconds())
	})
	return compilerExpressionsProducts.oracle, compilerExpressionsProducts.binary, compilerExpressionsProducts.absolute
}

func compilerExpressionsEnumeration(t *testing.T) ([][]string, string, int) {
	t.Helper()
	path, files := compilerManifest(t)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	paths := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(paths) != files || files == 0 {
		t.Fatal("enumeration differs or is empty")
	}
	shards := compilerExpressionsPartition(t, paths)
	if len(shards) != testCompilerExpressionsAgreeShards {
		t.Fatal("shard enumeration differs")
	}
	return shards, paths[0], files
}

func TestCompilerExpressionsAgreeUnion(t *testing.T) {
	t.Parallel()
	functions := []func(*testing.T){TestCompilerExpressionsAgree_000, TestCompilerExpressionsAgree_001, TestCompilerExpressionsAgree_002, TestCompilerExpressionsAgree_003, TestCompilerExpressionsAgree_004, TestCompilerExpressionsAgree_005, TestCompilerExpressionsAgree_006, TestCompilerExpressionsAgree_007, TestCompilerExpressionsAgree_008, TestCompilerExpressionsAgree_009, TestCompilerExpressionsAgree_010, TestCompilerExpressionsAgree_011, TestCompilerExpressionsAgree_012, TestCompilerExpressionsAgree_013, TestCompilerExpressionsAgree_014, TestCompilerExpressionsAgree_015}
	if len(functions) != testCompilerExpressionsAgreeShards {
		t.Fatal("top-level shard enumeration differs")
	}
	shards, planted, files := compilerExpressionsEnumeration(t)
	count, catches, caught := 0, 0, -1
	for index, shard := range shards {
		for _, path := range shard {
			count++
			want := []byte("case 0\nexpression\nreference\n")
			got := append([]byte(nil), want...)
			if path == planted {
				got[len("case 0\n")] ^= 1
			}
			if difference(got, want) != "" {
				catches++
				caught = index
			}
		}
	}
	if count != files || catches != 1 {
		t.Fatalf("union=%d/%d planted catches=%d", count, files, catches)
	}
	t.Logf("union=%d each exactly once; planted disagreement %s caught by shard-%03d", count, filepath.Base(planted), caught)
}

func compilerExpressionsShard(t *testing.T, index int) {
	t.Helper()
	// Setup (including waiting for another builder) has its own 90s budget.
	// A filtered shard invocation may need to fetch the shared products first.
	oracle, binary, absolute := compilerExpressionsSetup(t)
	started := time.Now()
	timer := time.AfterFunc(90*time.Second, func() { panic(fmt.Sprintf("cooked: shard-%03d exceeded 90s", index)) })
	defer timer.Stop()
	shards, planted, _ := compilerExpressionsEnumeration(t)
	paths := shards[index]
	manifest := filepath.Join(t.TempDir(), "compiler.txt")
	if err := os.WriteFile(manifest, []byte(strings.Join(paths, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	want := execute(t, "", oracle, "--manifest", manifest)
	for caseIndex, path := range paths {
		if path != planted {
			continue
		}
		marker := []byte(fmt.Sprintf("case %d\n", caseIndex))
		offset := bytes.Index(want.output, marker) + len(marker)
		if offset < len(marker) || offset >= len(want.output) {
			t.Fatal("planted case absent from oracle output")
		}
		mutant := append([]byte(nil), want.output...)
		mutant[offset] ^= 1
		if difference(mutant, want.output) == "" {
			t.Fatal("planted disagreement survived")
		}
		t.Logf("planted disagreement caught by shard-%03d for %s", index, filepath.Base(planted))
	}
	got := node(t, absolute, manifest, false)
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("Node: %s", diff)
	}
	got = execute(t, "", binary, "--manifest", manifest)
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("native: %s", diff)
	}
	t.Logf("shard-%03d: %d whole compiler files, %d identical expression tree bytes; %.3fs cooked=false", index, len(paths), len(want.output), time.Since(started).Seconds())
}

func TestCompilerExpressionsAgree_000(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 0)
}

func TestCompilerExpressionsAgree_001(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 1)
}

func TestCompilerExpressionsAgree_002(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 2)
}

func TestCompilerExpressionsAgree_003(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 3)
}

func TestCompilerExpressionsAgree_004(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 4)
}

func TestCompilerExpressionsAgree_005(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 5)
}

func TestCompilerExpressionsAgree_006(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 6)
}

func TestCompilerExpressionsAgree_007(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 7)
}

func TestCompilerExpressionsAgree_008(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 8)
}

func TestCompilerExpressionsAgree_009(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 9)
}

func TestCompilerExpressionsAgree_010(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 10)
}

func TestCompilerExpressionsAgree_011(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 11)
}

func TestCompilerExpressionsAgree_012(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 12)
}

func TestCompilerExpressionsAgree_013(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 13)
}

func TestCompilerExpressionsAgree_014(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 14)
}

func TestCompilerExpressionsAgree_015(t *testing.T) {
	t.Parallel()
	compilerExpressionsShard(t, 15)
}
