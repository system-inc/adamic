package tsprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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
	options := native.Options{Sanitize: true, Split: true, Jobs: 4}
	state := &mutantNativeProducts[index]
	state.once.Do(func() {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		// What lowering and the native build read (mutantNativeBuildFiles, which TestMutantNativeInputsCoverWhatItReads
		// keeps complete): not cohere whole, and the scanner the parser imports included.
		var inputs []string
		for _, file := range mutantNativeBuildFiles(t) {
			inputs = append(inputs, filepath.Join(root, file))
		}
		state.directory = expressionBuild(t, printerBuildInputs{
			Name:      fmt.Sprintf("TS printer mutant %03d sanitized native", index),
			Files:     inputs,
			Flags:     append(native.Flags(options), "split=true", "jobs=4", fmt.Sprintf("mutation=%+v", mutations[index])),
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
			err = native.Build(source, filepath.Join(directory, "port"), options)
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
		// The corpus is the oracle's answers: the oracle is a product (mutantOracleBinary, or documentOracle for the doc
		// mutation), a deterministic build of what go compiles, so it stands in this key by its own key (mutantOracleKey),
		// and only a build of the corpus resolves the binary. A reader keys the corpus without holding an oracle of about
		// 53 MB (Tests, Oct 12). Then the gaps file, and the code that runs it and captures its output. Not cohere whole:
		// its TypeScript submodule made a populated tree hash 67,000 files per lookup, up to 58 times a cold process
		// (#4ycgm2n).
		state.directory = expressionBuild(t, mutantOracleCorpusInputs(root, index, mutantOracleKey(t, root, index)), func(directory string) error {
			var cases string
			if mutations[index].entry == "docMain.ts" {
				cases, _, _ = documentCorpus(t)
			} else {
				cases, _ = mutantOracle(t, mutations[index], mutantOracleBinary(t, mutations[index]))
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

// mutantOracleCorpusInputs are mutation index's oracle corpus inputs under root: the oracle by its product's key, the
// gaps file, and the code that runs the oracle and captures its output.
func mutantOracleCorpusInputs(root string, index int, oracle string) printerBuildInputs {
	inputs := []string{root + "/internal/childguard"}
	for _, file := range []string{"mutants_test.go", "mutant_oracle_test.go", "mutant_products_test.go", "helpers_test.go", "doc_test.go", "shards_test.go", "testdata/notyet.json", "expressions_test.go"} {
		inputs = append(inputs, root+"/stage1/cohere/tsprinter/"+file)
	}
	return printerBuildInputs{
		Name:      fmt.Sprintf("TS printer mutant %03d Go oracle corpus", index),
		Files:     inputs,
		Flags:     []string{"oracle " + oracle, fmt.Sprintf("mutation=%+v", mutations[index]), fmt.Sprintf("families=%q", mutantFamilies[mutations[index].name]), "oracle-width=80"},
		Toolchain: runtime.Version(),
	}
}

// mutantOracleKey is the key of mutation index's oracle product (documentOracle for the doc mutation, else
// mutantOracleBinary), computed without building or fetching it.
func mutantOracleKey(t *testing.T, root string, index int) string {
	t.Helper()
	// Only the doc mutation keys the doc oracle, so its package never joins another mutation's reads.
	var inputs buildcache.Inputs
	if mutations[index].entry == "docMain.ts" {
		inputs = documentOracleInputs(t)
	} else {
		inputs = mutantOracleInputs(t, mutations[index])
	}
	key, err := buildcache.Key(root, inputs)
	if err != nil {
		t.Fatal(err)
	}
	return key
}

// A mutant's oracle corpus is keyed by its oracle product's key and the code that runs the oracle, never by the
// oracle's bytes or cohere. In a stand-in tree, an edit to a file the oracle compiles moves the corpus key through the
// oracle's key, an edit to each declared file moves it, and an edit to a file neither reads holds it. No declared input
// sits under cohere or outside the repository, so where the oracle's product sits (its cache path) never reaches the
// key. And the real oracles of two mutations with different witness families key their corpora apart, each the same
// asked twice. Keying builds nothing. Mutant: the oracle's key dropped from the corpus flags.
func TestMutantOracleCorpusKeyIsItsOracle(t *testing.T) {
	t.Parallel()
	stand := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(stand, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	compiled := "cohere/internal/format/javascript/printer.go"
	declared := []string{"internal/childguard/childguard.go"}
	for _, file := range mutantOracleCorpusInputs(stand, 3, "").Files[1:] {
		declared = append(declared, filepath.ToSlash(mustRelative(t, stand, file)))
	}
	for _, name := range append([]string{compiled, "internal/unrelated.go"}, declared...) {
		write(name, name)
	}
	key := func() string {
		t.Helper()
		oracle, err := buildcache.Key(stand, buildcache.Inputs{Name: "stand-in oracle", Files: []string{compiled}, Toolchain: []string{runtime.Version()}})
		if err != nil {
			t.Fatal(err)
		}
		inputs := mutantOracleCorpusInputs(stand, 3, oracle)
		corpus := buildcache.Inputs{Name: inputs.Name, Flags: inputs.Flags, Toolchain: []string{inputs.Toolchain}}
		for _, file := range inputs.Files {
			corpus.Files = append(corpus.Files, filepath.ToSlash(mustRelative(t, stand, file)))
		}
		key, err := buildcache.Key(stand, corpus)
		if err != nil {
			t.Fatal(err)
		}
		return key
	}
	before := key()
	write("internal/unrelated.go", "package unrelated // edited")
	if key() != before {
		t.Fatal("an edit to a file neither the oracle nor the corpus reads moved the corpus key")
	}
	write(compiled, compiled+"!")
	if key() == before {
		t.Fatal("an edit to a file the oracle compiles kept the corpus key")
	}
	write(compiled, compiled)
	for _, name := range declared {
		write(name, name+"!")
		if key() == before {
			t.Errorf("a one-byte edit to %s kept the corpus key", name)
		}
		write(name, name)
	}

	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range mutantOracleCorpusInputs(root, 3, "").Files {
		relative := mustRelative(t, root, file)
		if relative == ".." || strings.HasPrefix(relative, "../") || relative == "cohere" || strings.HasPrefix(relative, "cohere/") {
			t.Fatalf("the oracle corpus declares %s, outside the repository or under cohere", file)
		}
	}
	if fmt.Sprint(mutantFamilies[mutations[3].name]) == fmt.Sprint(mutantFamilies[mutations[6].name]) {
		t.Fatalf("mutations 3 and 6 share witness families, so their oracles can't show the key moving")
	}
	three, six := mutantOracleKey(t, root, 3), mutantOracleKey(t, root, 6)
	if three == six {
		t.Fatal("two oracles filtered to different witness families keyed alike")
	}
	if again := mutantOracleKey(t, root, 3); again != three {
		t.Fatalf("one oracle keyed twice gave %s and %s", three[:12], again[:12])
	}
	t.Logf("the oracle's key moves the corpus key; mutation 3's oracle %s, 6's %s; %d declared inputs, none under cohere or outside the repository", three[:12], six[:12], len(declared))
}

func mustRelative(t *testing.T, root, path string) string {
	t.Helper()
	relative, err := filepath.Rel(root, path)
	if err != nil {
		t.Fatal(err)
	}
	return filepath.ToSlash(relative)
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
