package tsprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

// A leaf and its product declaration share both the recipe and its key. The
// once belongs to one build product, never to a top-level setup test. Products
// contain immutable outputs; temporary Node sources belong to the caller.
type mutantProductState struct {
	once      sync.Once
	directory string
}

var mutantNativeProducts [testMutantsShards]mutantProductState
var mutantOracleProducts [testMutantsShards]mutantProductState

func mutantNativeProduct(t *testing.T, index int) string {
	t.Helper()
	state := &mutantNativeProducts[index]
	state.once.Do(func() {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		inputs := expressionInputFiles(t, root+"/internal", root+"/bridge/tsgo", root+"/cohere", root+"/stage1/typescript/parser")
		inputs = append(inputs, root+"/go.mod", root+"/go.sum", root+"/go.work", root+"/stage1/cohere/tsprinter/mutants_test.go", root+"/stage1/cohere/tsprinter/mutant_products_test.go", root+"/stage1/cohere/tsprinter/helpers_test.go", root+"/stage1/cohere/tsprinter/expressions_test.go", root+"/stage1/cohere/tsprinter/expression_units_test.go")
		files, err := filepath.Glob("*.ts")
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			inputs = append(inputs, filepath.Join(root, "stage1/cohere/tsprinter", file))
		}
		state.directory = expressionBuild(t, printerBuildInputs{
			Name:      fmt.Sprintf("TS printer mutant %03d sanitized native", index),
			Files:     inputs,
			Flags:     append(native.Flags(native.Options{Sanitize: true}), fmt.Sprintf("mutation=%+v", mutations[index])),
			Toolchain: runtime.Version() + "\n" + buildcache.Tool("clang", "--version"),
		}, func(directory string) error {
			path := mutatedPort(t, mutations[index])
			started := time.Now()
			program, err := lowerProgram(path)
			if err != nil {
				return err
			}
			t.Logf("grain lowering: %.3fs", time.Since(started).Seconds())
			started = time.Now()
			source := native.C(program)
			t.Logf("grain C emission: %.3fs", time.Since(started).Seconds())
			started = time.Now()
			err = native.Build(source, filepath.Join(directory, "port"), native.Options{Sanitize: true})
			t.Logf("grain clang runtime/compile/link: %.3fs", time.Since(started).Seconds())
			return err
		})
	})
	if state.directory == "" {
		t.Fatal("native product preparation failed")
	}
	return filepath.Join(state.directory, "port")
}

func mutantOracleProduct(t *testing.T, index int) (string, string) {
	t.Helper()
	state := &mutantOracleProducts[index]
	state.once.Do(func() {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		inputs := expressionInputFiles(t, root+"/cohere", root+"/internal/childguard")
		for _, file := range []string{"mutants_test.go", "mutant_oracle_test.go", "mutant_units_test.go", "mutant_products_test.go", "helpers_test.go", "doc_test.go", "shards_test.go", "testdata/doc_side_test.go", "testdata/expressions_side_test.go", "testdata/statements_side_test.go", "testdata/notyet.json", "expressions_test.go", "expression_units_test.go"} {
			inputs = append(inputs, root+"/stage1/cohere/tsprinter/"+file)
		}
		inputs = append(inputs, root+"/go.mod", root+"/go.sum", root+"/go.work")
		state.directory = expressionBuild(t, printerBuildInputs{
			Name:      fmt.Sprintf("TS printer mutant %03d Go oracle corpus", index),
			Files:     inputs,
			Flags:     []string{fmt.Sprintf("mutation=%+v", mutations[index]), fmt.Sprintf("families=%q", mutantFamilies[mutations[index].name]), "oracle-width=80", "Go environment=" + buildcache.Tool("go", "env", "GOFLAGS", "CGO_ENABLED", "GOEXPERIMENT", "GOTOOLCHAIN")},
			Toolchain: runtime.Version(),
		}, func(directory string) error {
			var cases string
			if mutations[index].entry == "docMain.ts" {
				cases, _, _ = documentCorpus(t)
			} else {
				cases, _ = mutantOracle(t, mutations[index])
			}
			// Requests and overlays contain private paths and are never published.
			names := []string{filepath.Base(cases), "answers.txt", "cases.json"}
			if mutations[index].entry == "docMain.ts" {
				names = []string{"docs.txt", "answers.txt", "docs.json"}
			}
			if mutations[index].name == "hashbang loses its refusal" {
				names = append(names, "gaps.txt", "gap-answers.txt")
			}
			for _, name := range names {
				data, err := os.ReadFile(filepath.Join(filepath.Dir(cases), name))
				if err != nil {
					return err
				}
				if err := os.WriteFile(filepath.Join(directory, name), data, 0644); err != nil {
					return err
				}
			}
			return nil
		})
	})
	if state.directory == "" {
		t.Fatal("oracle product preparation failed")
	}
	name := "cases.txt"
	if mutations[index].entry == "docMain.ts" {
		name = "docs.txt"
	}
	answers, err := os.ReadFile(filepath.Join(state.directory, "answers.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(state.directory, name), string(answers)
}

// Each declaration owns exactly one product and uses the leaf's shared recipe.
func TestProduct_TSPrinterMutantNative_000(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 0) }
func TestProduct_TSPrinterMutantOracle_000(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 0) }
func TestProduct_TSPrinterMutantNative_001(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 1) }
func TestProduct_TSPrinterMutantOracle_001(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 1) }
func TestProduct_TSPrinterMutantNative_002(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 2) }
func TestProduct_TSPrinterMutantOracle_002(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 2) }
func TestProduct_TSPrinterMutantNative_003(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 3) }
func TestProduct_TSPrinterMutantOracle_003(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 3) }
func TestProduct_TSPrinterMutantNative_004(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 4) }
func TestProduct_TSPrinterMutantOracle_004(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 4) }
func TestProduct_TSPrinterMutantNative_005(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 5) }
func TestProduct_TSPrinterMutantOracle_005(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 5) }
func TestProduct_TSPrinterMutantNative_006(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 6) }
func TestProduct_TSPrinterMutantOracle_006(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 6) }
func TestProduct_TSPrinterMutantNative_007(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 7) }
func TestProduct_TSPrinterMutantOracle_007(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 7) }
func TestProduct_TSPrinterMutantNative_008(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 8) }
func TestProduct_TSPrinterMutantOracle_008(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 8) }
func TestProduct_TSPrinterMutantNative_009(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 9) }
func TestProduct_TSPrinterMutantOracle_009(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 9) }
func TestProduct_TSPrinterMutantNative_010(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 10) }
func TestProduct_TSPrinterMutantOracle_010(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 10) }
func TestProduct_TSPrinterMutantNative_011(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 11) }
func TestProduct_TSPrinterMutantOracle_011(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 11) }
func TestProduct_TSPrinterMutantNative_012(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 12) }
func TestProduct_TSPrinterMutantOracle_012(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 12) }
func TestProduct_TSPrinterMutantNative_013(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 13) }
func TestProduct_TSPrinterMutantOracle_013(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 13) }
func TestProduct_TSPrinterMutantNative_014(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 14) }
func TestProduct_TSPrinterMutantOracle_014(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 14) }
func TestProduct_TSPrinterMutantNative_015(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 15) }
func TestProduct_TSPrinterMutantOracle_015(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 15) }
func TestProduct_TSPrinterMutantNative_016(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 16) }
func TestProduct_TSPrinterMutantOracle_016(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 16) }
func TestProduct_TSPrinterMutantNative_017(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 17) }
func TestProduct_TSPrinterMutantOracle_017(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 17) }
func TestProduct_TSPrinterMutantNative_018(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 18) }
func TestProduct_TSPrinterMutantOracle_018(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 18) }
func TestProduct_TSPrinterMutantNative_019(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 19) }
func TestProduct_TSPrinterMutantOracle_019(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 19) }
func TestProduct_TSPrinterMutantNative_020(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 20) }
func TestProduct_TSPrinterMutantOracle_020(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 20) }
func TestProduct_TSPrinterMutantNative_021(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 21) }
func TestProduct_TSPrinterMutantOracle_021(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 21) }
func TestProduct_TSPrinterMutantNative_022(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 22) }
func TestProduct_TSPrinterMutantOracle_022(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 22) }
func TestProduct_TSPrinterMutantNative_023(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 23) }
func TestProduct_TSPrinterMutantOracle_023(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 23) }
func TestProduct_TSPrinterMutantNative_024(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 24) }
func TestProduct_TSPrinterMutantOracle_024(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 24) }
func TestProduct_TSPrinterMutantNative_025(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 25) }
func TestProduct_TSPrinterMutantOracle_025(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 25) }
func TestProduct_TSPrinterMutantNative_026(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 26) }
func TestProduct_TSPrinterMutantOracle_026(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 26) }
func TestProduct_TSPrinterMutantNative_027(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 27) }
func TestProduct_TSPrinterMutantOracle_027(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 27) }
func TestProduct_TSPrinterMutantNative_028(t *testing.T) { t.Parallel(); mutantNativeProduct(t, 28) }
func TestProduct_TSPrinterMutantOracle_028(t *testing.T) { t.Parallel(); mutantOracleProduct(t, 28) }
