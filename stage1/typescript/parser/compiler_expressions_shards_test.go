package parser

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/testgrain"
)

const testCompilerExpressionsAgreeShards = 16

func compilerExpressionsPartition(t *testing.T, paths []string) [][]string {
	t.Helper()
	identities := compilerExpressionsIdentities(paths)
	shards := make([][]string, testCompilerExpressionsAgreeShards)
	for shard, indices := range testgrain.Assign(identities, testCompilerExpressionsAgreeShards) {
		for _, i := range indices {
			shards[shard] = append(shards[shard], paths[i])
		}
	}
	return shards
}
func compilerExpressionsIdentities(paths []string) []string {
	identities := make([]string, len(paths))
	for i, path := range paths {
		identities[i] = filepath.ToSlash(filepath.Base(path))
	}
	return identities
}

func compilerExpressionsLower(t *testing.T, absolute string) string {
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
	return lowered
}

func compilerExpressionsNative(t *testing.T, absolute string) string {
	t.Helper()
	lowered := compilerExpressionsLower(t, absolute)
	files := []string{"stage1/typescript/parser", "stage1/typescript/scanner", "internal/load", "internal/lower", "internal/ir", "internal/native", "cohere", "go.mod"}
	tools := []string{buildcache.Tool("clang", "--version"), buildcache.Tool("go", "version")}
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

type compilerExpressionsProducts struct {
	oracle, binary, absolute string
	paths                    []string
}

func compilerExpressionsSetup(t *testing.T) (string, string, string) {
	products := compilerExpressionsReady(t)
	return products.oracle, products.binary, products.absolute
}
func compilerExpressionsReady(t *testing.T) compilerExpressionsProducts {
	return testgrain.Setup(t, "compiler-expressions/ready", func() (compilerExpressionsProducts, error) {
		manifest, files := compilerExpressionsManifest(t)
		data, err := os.ReadFile(manifest)
		if err != nil {
			return compilerExpressionsProducts{}, err
		}
		paths := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
		if len(paths) != files || files == 0 {
			return compilerExpressionsProducts{}, fmt.Errorf("enumeration differs or is empty")
		}
		absolute, err := filepath.Abs(".")
		if err != nil {
			return compilerExpressionsProducts{}, err
		}
		oracle := compilerExpressionsOracle(t)
		binary := compilerExpressionsNative(t, absolute)
		return compilerExpressionsProducts{oracle: oracle, binary: binary, absolute: absolute, paths: paths}, nil
	})
}

func compilerExpressionsEnumeration(t *testing.T) ([][]string, string, int) {
	products := compilerExpressionsReady(t)
	return compilerExpressionsPartition(t, products.paths), products.paths[0], len(products.paths)
}

func TestCompilerExpressionsAgreeUnion(t *testing.T) {
	t.Parallel()
	products := compilerExpressionsReady(t)
	assignments := testgrain.Assign(compilerExpressionsIdentities(products.paths), testCompilerExpressionsAgreeShards)
	testgrain.Union(t, "TestCompilerExpressionsAgree", testCompilerExpressionsAgreeShards, assignments, len(products.paths))
	caught := make(map[int]bool)
	owner := -1
	for shard, indices := range assignments {
		for _, i := range indices {
			want := []byte("case 0\nexpression\nreference\n")
			got := append([]byte(nil), want...)
			if i == 0 {
				got[len("case 0\n")] ^= 1
				owner = shard
			}
			if difference(got, want) != "" {
				caught[shard] = true
			}
		}
	}
	testgrain.CaughtByExactly(t, caught, owner)
	t.Logf("union=%d each exactly once; planted disagreement %s caught by shard-%03d", len(products.paths), filepath.Base(products.paths[0]), owner)
}

func compilerExpressionsShard(t *testing.T, index int) {
	t.Helper()
	// A selected shard prepares its products itself before its own deadline.
	oracle, binary, absolute := compilerExpressionsSetup(t)
	testgrain.Unit(t)

	shards, planted, _ := compilerExpressionsEnumeration(t)
	paths := shards[index]
	manifest := filepath.Join(t.TempDir(), "compiler.txt")
	if err := os.WriteFile(manifest, []byte(strings.Join(paths, "\n")+"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	want := compilerExpressionsExecute(t, "", oracle, "--manifest", manifest)
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
	got := compilerExpressionsNode(t, absolute, manifest, false)
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("Node: %s", diff)
	}
	got = compilerExpressionsExecute(t, "", binary, "--manifest", manifest)
	if diff := difference(got.output, want.output); diff != "" {
		t.Fatalf("native: %s", diff)
	}

	t.Logf("shard-%03d: %d whole compiler files, %d identical expression tree bytes", index, len(paths), len(want.output))
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

func compilerExpressionsExecute(t *testing.T, directory, name string, args ...string) execution {
	t.Helper()
	command := testgrain.Command(t, name, args...)
	command.Dir = directory
	output, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	command.Stdout = output
	var stderr bytes.Buffer
	command.Stderr = &stderr
	started := time.Now()
	err = command.Run()
	duration := time.Since(started)
	if err != nil || stderr.Len() != 0 {
		t.Fatalf("%s %v: %v\n%s", name, args, err, &stderr)
	}
	data, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	return execution{data, duration}
}

func compilerExpressionsNode(t *testing.T, directory, manifest string, count bool) execution {
	t.Helper()
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), "--manifest", manifest}
	if count {
		args = append(args, "--count")
	}
	return compilerExpressionsExecute(t, "", "node", args...)
}

func compilerExpressionsManifest(t *testing.T) (string, int) {
	t.Helper()
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		t.Skip("set ADAMIC_TYPESCRIPT_SOURCE to the pinned v6.0.3 checkout")
	}
	output, err := testgrain.Command(t, "git", "-C", source, "rev-parse", "HEAD").Output()
	if err != nil || strings.TrimSpace(string(output)) != compilerCommit {
		t.Fatalf("corpus pin differs: %q %v", output, err)
	}
	var manifest strings.Builder
	files := 0
	err = filepath.WalkDir(filepath.Join(source, "src/compiler"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
			manifest.WriteString(path + "\n")
			files++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "compiler.txt")
	if err := os.WriteFile(path, []byte(manifest.String()), 0644); err != nil {
		t.Fatal(err)
	}
	return path, files
}

func compilerExpressionsOracle(t *testing.T) string {
	t.Helper()
	files := append(wholeMutantBuildFiles(t), "stage1/typescript/parser/testdata/oracle.go")
	inputs := buildcache.Inputs{Name: "typescript-parser-oracle", Files: files, Flags: wholeMutantBuildFlags(), Toolchain: []string{buildcache.Tool("go", "version")}}
	directory := buildcache.Product(t, inputs, func(dir string) error {
		root, err := filepath.Abs(filepath.Join(repository, "cohere/TypeScript/tsc"))
		if err != nil {
			return err
		}
		side, err := filepath.Abs("testdata/oracle.go")
		if err != nil {
			return err
		}
		virtual := filepath.Join(root, "adamic_parser_oracle.go")
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side}})
		if err != nil {
			return err
		}
		path := filepath.Join(dir, "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			return err
		}
		command := testgrain.Command(t, "go", "build", "-overlay="+path, "-o", filepath.Join(dir, "oracle"), virtual)
		command.Dir = root
		output, err := command.CombinedOutput()
		if err != nil {
			return fmt.Errorf("oracle: %w\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(directory, "oracle")
}
