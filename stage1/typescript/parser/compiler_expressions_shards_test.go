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
	_ "github.com/system-inc/adamic/internal/ccjobserver"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testCompilerExpressionsAgreeShards = 16

func compilerExpressionsPartition(test *testing.T, paths []string) [][]string {
	test.Helper()
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
				test.Fatalf("case %q occurs %d times", path, seen[path])
			}
			count++
		}
	}
	if count != len(paths) {
		test.Fatalf("union %d != enumeration %d", count, len(paths))
	}
	return shards
}

func compilerExpressionsLower(test *testing.T, absolute string) string {
	test.Helper()
	files := []string{
		"stage1/typescript/parser",
		"stage1/typescript/scanner",
		"internal/load",
		"internal/lower",
		"internal/ir",
		"internal/native",
		"cohere",
		"go.mod",
	}
	tools := []string{buildcache.Tool("clang", "--version"), buildcache.Tool("go", "version")}
	lowered := buildcache.Product(test, buildcache.Inputs{
		Name:      "compiler-expressions-lowered",
		Files:     files,
		Flags:     []string{"native.C"},
		Toolchain: tools,
	}, func(dir string) error {
		program, loadError := load.Load([]string{filepath.Join(absolute, "main.ts")})
		if loadError != nil {
			return loadError
		}
		ir, lowerError := lower.Lower(context.Background(), program)
		if lowerError != nil {
			return lowerError
		}
		return os.WriteFile(filepath.Join(dir, "main.c"), []byte(native.C(ir)), 0644)
	})
	return lowered
}

func compilerExpressionsNative(test *testing.T, absolute string) string {
	test.Helper()
	lowered := compilerExpressionsLower(test, absolute)
	files := []string{
		"stage1/typescript/parser",
		"stage1/typescript/scanner",
		"internal/load",
		"internal/lower",
		"internal/ir",
		"internal/native",
		"cohere",
		"go.mod",
	}
	tools := []string{buildcache.Tool("clang", "--version"), buildcache.Tool("go", "version")}
	flags := append(
		native.Flags(native.Options{Sanitize: true}),
		"ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT"),
	)
	product := buildcache.Product(test, buildcache.Inputs{
		Name:      "compiler-expressions-native",
		Files:     files,
		Flags:     flags,
		Toolchain: tools,
	}, func(dir string) error {
		source, readError := os.ReadFile(filepath.Join(lowered, "main.c"))
		if readError != nil {
			return readError
		}
		return native.Build(string(source), filepath.Join(dir, "parser"), native.Options{Sanitize: true})
	})
	return filepath.Join(product, "parser")
}

var compilerExpressionsProducts struct {
	once                     sync.Once
	oracle, binary, absolute string
}

func compilerExpressionsSetup(test *testing.T) (string, string, string) {
	test.Helper()
	compilerExpressionsProducts.once.Do(func() {
		started := time.Now()
		// Setup has no deadline; Loom still bounds the complete unit.
		// Preserve the pinned corpus prerequisite even for a setup-only invocation.
		compilerManifest(test)
		absolute, absolutePathError := filepath.Abs(".")
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		compilerExpressionsProducts.oracle = wholeMutantBuildOracleProduct(test)
		compilerExpressionsProducts.absolute = absolute
		compilerExpressionsProducts.binary = compilerExpressionsNative(test, absolute)
		test.Logf("TestCompilerExpressionsAgree_Setup: %.3fs", time.Since(started).Seconds())
	})
	return compilerExpressionsProducts.oracle, compilerExpressionsProducts.binary, compilerExpressionsProducts.absolute
}

func compilerExpressionsEnumeration(test *testing.T) ([][]string, string, int) {
	test.Helper()
	path, files := compilerManifest(test)
	data, readError := os.ReadFile(path)
	if readError != nil {
		test.Fatal(readError)
	}
	paths := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(paths) != files || files == 0 {
		test.Fatal("enumeration differs or is empty")
	}
	shards := compilerExpressionsPartition(test, paths)
	if len(shards) != testCompilerExpressionsAgreeShards {
		test.Fatal("shard enumeration differs")
	}
	return shards, paths[0], files
}

// TestCompilerExpressionsAgree_* and TestWholeCompilerAgrees_* both run compilerExpressionsPartition's shards, so one
// union holds both: each family has a top-level test per shard, every compiler file is in exactly one shard, and a
// planted disagreement is caught by exactly one.
func TestCompilerShardsHoldEveryFileOnce(test *testing.T) {
	test.Parallel()
	for _, functions := range [][]func(*testing.T){
		{
			TestCompilerExpressionsAgree_000,
			TestCompilerExpressionsAgree_001,
			TestCompilerExpressionsAgree_002,
			TestCompilerExpressionsAgree_003,
			TestCompilerExpressionsAgree_004,
			TestCompilerExpressionsAgree_005,
			TestCompilerExpressionsAgree_006,
			TestCompilerExpressionsAgree_007,
			TestCompilerExpressionsAgree_008,
			TestCompilerExpressionsAgree_009,
			TestCompilerExpressionsAgree_010,
			TestCompilerExpressionsAgree_011,
			TestCompilerExpressionsAgree_012,
			TestCompilerExpressionsAgree_013,
			TestCompilerExpressionsAgree_014,
			TestCompilerExpressionsAgree_015,
		},
		{
			TestWholeCompilerAgrees_000,
			TestWholeCompilerAgrees_001,
			TestWholeCompilerAgrees_002,
			TestWholeCompilerAgrees_003,
			TestWholeCompilerAgrees_004,
			TestWholeCompilerAgrees_005,
			TestWholeCompilerAgrees_006,
			TestWholeCompilerAgrees_007,
			TestWholeCompilerAgrees_008,
			TestWholeCompilerAgrees_009,
			TestWholeCompilerAgrees_010,
			TestWholeCompilerAgrees_011,
			TestWholeCompilerAgrees_012,
			TestWholeCompilerAgrees_013,
			TestWholeCompilerAgrees_014,
			TestWholeCompilerAgrees_015,
		},
	} {
		if len(functions) != testCompilerExpressionsAgreeShards {
			test.Fatal("top-level shard enumeration differs")
		}
	}
	shards, planted, files := compilerExpressionsEnumeration(test)
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
		test.Fatalf("union=%d/%d planted catches=%d", count, files, catches)
	}
	test.Logf(
		"union=%d each exactly once; planted disagreement %s caught by shard-%03d",
		count,
		filepath.Base(planted),
		caught,
	)
}

func compilerExpressionsShard(test *testing.T, index int) {
	test.Helper()
	// A selected shard prepares its products itself before its own deadline.
	oracle, binary, absolute := compilerExpressionsSetup(test)
	started := time.Now()
	timer := time.AfterFunc(90*time.Second, func() { panic(fmt.Sprintf("cooked: shard-%03d exceeded 90s", index)) })
	defer timer.Stop()
	shards, planted, _ := compilerExpressionsEnumeration(test)
	paths := shards[index]
	manifest := filepath.Join(test.TempDir(), "compiler.txt")
	if writeError := os.WriteFile(manifest, []byte(strings.Join(paths, "\n")+"\n"), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	want := execute(test, "", oracle, "--manifest", manifest)
	for caseIndex, path := range paths {
		if path != planted {
			continue
		}
		marker := []byte(fmt.Sprintf("case %d\n", caseIndex))
		offset := bytes.Index(want.output, marker) + len(marker)
		if offset < len(marker) || offset >= len(want.output) {
			test.Fatal("planted case absent from oracle output")
		}
		mutant := append([]byte(nil), want.output...)
		mutant[offset] ^= 1
		if difference(mutant, want.output) == "" {
			test.Fatal("planted disagreement survived")
		}
		test.Logf("planted disagreement caught by shard-%03d for %s", index, filepath.Base(planted))
	}
	got := node(test, absolute, manifest, false)
	if diff := difference(got.output, want.output); diff != "" {
		test.Fatalf("Node: %s", diff)
	}
	got = execute(test, "", binary, "--manifest", manifest)
	if diff := difference(got.output, want.output); diff != "" {
		test.Fatalf("native: %s", diff)
	}
	test.Logf(
		"shard-%03d: %d whole compiler files, %d identical expression tree bytes; %.3fs cooked=false",
		index,
		len(paths),
		len(want.output),
		time.Since(started).Seconds(),
	)
}

func TestCompilerExpressionsAgree_000(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 0)
}

func TestCompilerExpressionsAgree_001(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 1)
}

func TestCompilerExpressionsAgree_002(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 2)
}

func TestCompilerExpressionsAgree_003(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 3)
}

func TestCompilerExpressionsAgree_004(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 4)
}

func TestCompilerExpressionsAgree_005(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 5)
}

func TestCompilerExpressionsAgree_006(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 6)
}

func TestCompilerExpressionsAgree_007(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 7)
}

func TestCompilerExpressionsAgree_008(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 8)
}

func TestCompilerExpressionsAgree_009(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 9)
}

func TestCompilerExpressionsAgree_010(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 10)
}

func TestCompilerExpressionsAgree_011(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 11)
}

func TestCompilerExpressionsAgree_012(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 12)
}

func TestCompilerExpressionsAgree_013(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 13)
}

func TestCompilerExpressionsAgree_014(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 14)
}

func TestCompilerExpressionsAgree_015(test *testing.T) {
	test.Parallel()
	compilerExpressionsShard(test, 15)
}
