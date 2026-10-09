package lint

import (
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

const testDecodedOptionsAndMutantShards = 6

// The fixture and its six assertions are pinned, rather than a growing corpus.
func decodedOptionsCases() []string {
	return []string{"Node", "emitted JavaScript", "native", "Go count", "mutant Node", "mutant emitted JavaScript"}
}

func decodedOptionsSlice(shard int) []int {
	var indices []int
	for index := range decodedOptionsCases() {
		if index%testDecodedOptionsAndMutantShards == shard {
			indices = append(indices, index)
		}
	}
	return indices
}

func TestDecodedOptionsAndMutant_000(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 0) }
func TestDecodedOptionsAndMutant_001(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 1) }
func TestDecodedOptionsAndMutant_002(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 2) }
func TestDecodedOptionsAndMutant_003(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 3) }
func TestDecodedOptionsAndMutant_004(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 4) }
func TestDecodedOptionsAndMutant_005(t *testing.T) { t.Parallel(); decodedOptionsShard(t, 5) }

func TestDecodedOptionsAndMutantUnion(t *testing.T) {
	t.Parallel()
	// Read actual top-level registrations: a missing, repeated, or misnumbered
	// wrapper must fail even if the slice helper itself still covers every case.
	file, err := parser.ParseFile(token.NewFileSet(), "decoded_options_shards_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	registrations := map[int]int{}
	const prefix = "TestDecodedOptionsAndMutant_"
	for _, decl := range file.Decls {
		function, ok := decl.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(function.Name.Name, prefix) {
			continue
		}
		suffix := strings.TrimPrefix(function.Name.Name, prefix)
		shard, err := strconv.Atoi(suffix)
		if err != nil || shard < 0 || shard >= testDecodedOptionsAndMutantShards || suffix != fmt.Sprintf("%03d", shard) {
			t.Fatalf("invalid shard name %s", function.Name.Name)
		}
		registrations[shard]++
		calls := 0
		ast.Inspect(function.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name, ok := call.Fun.(*ast.Ident)
			if !ok || name.Name != "decodedOptionsShard" {
				return true
			}
			calls++
			if len(call.Args) != 2 {
				t.Fatalf("bad wrapper %s", function.Name.Name)
			}
			number, ok := call.Args[1].(*ast.BasicLit)
			if !ok || number.Value != strconv.Itoa(shard) {
				t.Fatalf("wrapper %s runs wrong slice", function.Name.Name)
			}
			return true
		})
		if calls != 1 {
			t.Fatalf("wrapper %s calls runner %d times", function.Name.Name, calls)
		}
	}
	if len(registrations) != testDecodedOptionsAndMutantShards {
		t.Fatalf("registered %d shards, want %d", len(registrations), testDecodedOptionsAndMutantShards)
	}
	seen := make([]int, len(decodedOptionsCases()))
	for shard := range testDecodedOptionsAndMutantShards {
		if registrations[shard] != 1 {
			t.Fatalf("shard %03d registered %d times", shard, registrations[shard])
		}
		for _, index := range decodedOptionsSlice(shard) {
			seen[index]++
		}
	}
	for index, count := range seen {
		if count != 1 {
			t.Fatalf("case %s covered %d times", decodedOptionsCases()[index], count)
		}
	}
	t.Logf("union: %d cases, each exactly once across %d top-level shards", len(seen), len(registrations))
}

// Both the live shards and the planted-failure test use this exact evaluator.
func decodedOptionsCheck(index int, got, want []byte) error {
	name := decodedOptionsCases()[index]
	if index == 3 {
		if !bytes.Equal(got, []byte("0\n")) {
			return fmt.Errorf("JSON catch option did not override the legacy default: %q", got)
		}
	} else if index >= 4 {
		if bytes.Equal(got, want) {
			return fmt.Errorf("ignored decoded-option mutant survived on %s", name)
		}
	} else if diff := difference(got, want); diff != "" {
		return fmt.Errorf("%s: %s", name, diff)
	}
	return nil
}

func TestDecodedOptionsAndMutantPlantedFailure(t *testing.T) {
	t.Parallel()
	planted := 1
	var caught []int
	for shard := range testDecodedOptionsAndMutantShards {
		failures := 0
		for _, index := range decodedOptionsSlice(shard) {
			want := []byte("case catch\n0 findings\n")
			got := append([]byte(nil), want...)
			if index == 3 {
				got = []byte("0\n")
			}
			if index >= 4 {
				got = []byte("case catch\n1 finding\n")
			}
			if err := decodedOptionsCheck(index, got, want); err != nil {
				t.Fatalf("unplanted case %s: %v", decodedOptionsCases()[index], err)
			}
			if index == planted {
				got = []byte("case catch\nplanted disagreement\n")
			}
			if decodedOptionsCheck(index, got, want) != nil {
				failures++
			}
		}
		if failures > 0 {
			caught = append(caught, shard)
		}
	}
	if len(caught) != 1 || caught[0] != planted%testDecodedOptionsAndMutantShards {
		t.Fatalf("planted disagreement caught by shards %v", caught)
	}
	t.Logf("planted case %s caught by exactly shard-%03d", decodedOptionsCases()[planted], caught[0])
}

func decodedOptionsInputs(t *testing.T, name string) buildcache.Inputs {
	t.Helper()
	prepareRegistry(t, ".")
	files := []string{"go.mod", "go.work", "stage1/cohere/lint/decoded_options_shards_test.go", "stage1/cohere/lint/shared_test.go", "stage1/cohere/lint/registry", "oracle/adamic.mjs", "cohere/TypeScript/tsc", "cohere/TypeScript-shim"}
	for _, file := range portFiles(t) {
		files = append(files, filepath.ToSlash(filepath.Join("stage1/cohere/lint", file)))
	}
	// Hash implementation sources, excluding unrelated tests and evidence trees.
	err := filepath.WalkDir(filepath.Join(repository, "internal"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() && (entry.Name() == "testdata" || entry.Name() == "performance") {
			return filepath.SkipDir
		}
		extension := filepath.Ext(path)
		if !entry.IsDir() && (extension == ".go" || extension == ".c" || extension == ".h") && !strings.HasSuffix(path, "_test.go") {
			relative, err := filepath.Rel(repository, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	return buildcache.Inputs{Name: name, Files: files, Flags: []string{"Sanitize=true", "Target=native", "Split=true", "Jobs=4", "ADAMIC_NATIVE_SPLIT=" + os.Getenv("ADAMIC_NATIVE_SPLIT"), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH}, Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")}}
}

func decodedOptionsLowered(t *testing.T, mutated bool) string {
	t.Helper()
	name := "lint-decoded-options-lowered"
	if mutated {
		name += "-ignored-allowemptycatch"
	}
	inputs := decodedOptionsInputs(t, name)
	return buildcache.Product(t, inputs, func(output string) error {
		directory := packageDirectory
		if mutated {
			directory = mutant(t, "'allowemptycatch'", "'ignored-allowemptycatch'", "main.ts")
		}
		if _, err := prepareDecodedOptionsLowering(directory, output); err != nil {
			return err
		}
		return nil
	})
}

func prepareDecodedOptionsLowering(directory, output string) (string, error) {
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		return "", err
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return "", err
	}
	source := native.C(lowered)
	if err := os.WriteFile(filepath.Join(output, "lint.c"), []byte(source), 0644); err != nil {
		return "", err
	}
	return source, os.WriteFile(filepath.Join(output, "lint.mjs"), []byte(javascript.JavaScript(lowered)), 0644)
}

func decodedOptionsNative(t *testing.T) string {
	t.Helper()
	lowered := decodedOptionsLowered(t, false)
	directory := buildcache.Product(t, decodedOptionsInputs(t, "lint-decoded-options-native-sanitized"), func(output string) error {
		source, err := os.ReadFile(filepath.Join(lowered, "lint.c"))
		if err != nil {
			return err
		}
		return native.Build(string(source), filepath.Join(output, "scanner"), native.Options{Sanitize: true, Split: true, Jobs: 4})
	})
	return filepath.Join(directory, "scanner")
}

func decodedOptionsShard(t *testing.T, shard int) {
	t.Helper()
	started := time.Now()
	fixture := filepath.Join(t.TempDir(), "catch.ts")
	if err := os.WriteFile(fixture, []byte("try { work(); } catch(e) {}\n"), 0644); err != nil {
		t.Fatal(err)
	}
	path := manifest(t, []string{fixture + "\tno-empty\t\t\tfalse\t{\"AllowEmptyCatch\":true}"})
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", path).output
	for _, index := range decodedOptionsSlice(shard) {
		var got []byte
		switch index {
		case 0:
			got = node(t, packageDirectory, path, false).output
		case 1:
			got = runJavaScript(t, filepath.Join(decodedOptionsLowered(t, false), "lint.mjs"), path, false).output
		case 2:
			got = execute(t, "", decodedOptionsNative(t), "--manifest", path).output
		case 3:
			got = execute(t, "", oracle, "--manifest", path, "--count").output
		case 4:
			got = node(t, mutant(t, "'allowemptycatch'", "'ignored-allowemptycatch'", "main.ts"), path, false).output
		case 5:
			got = runJavaScript(t, filepath.Join(decodedOptionsLowered(t, true), "lint.mjs"), path, false).output
		default:
			t.Fatalf("unhandled case %d", index)
		}
		if err := decodedOptionsCheck(index, got, want); err != nil {
			t.Fatal(err)
		}
		if index >= 4 {
			t.Logf("ignored decoded-option mutant caught on %s: %s", decodedOptionsCases()[index], difference(got, want))
		}
	}
	t.Logf("shard-%03d including setup: %s", shard, time.Since(started))
}
