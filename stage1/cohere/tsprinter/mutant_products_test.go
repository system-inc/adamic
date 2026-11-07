package tsprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/native"
)

var mutantNativeProducts [oracles.MutantsShards]oracles.MutantProductState

func mutantNativeProduct(test *testing.T, index int) string {
	test.Helper()
	options := native.Options{Sanitize: true, Split: true, Jobs: 4}
	state := &mutantNativeProducts[index]
	state.Once.Do(func() {
		root, absolutePathError := filepath.Abs(oracles.Repository)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		// What lowering and the native build read (mutantNativeBuildFiles, which TestMutantNativeInputsCoverWhatItReads
		// keeps complete): not cohere whole, and the scanner the parser imports included.
		var inputs []string
		for _, file := range mutantNativeBuildFiles(test) {
			inputs = append(inputs, filepath.Join(root, file))
		}
		state.Directory = oracles.ExpressionBuild(test, oracles.PrinterBuildInputs{
			Name:  fmt.Sprintf("TS printer mutant %03d sanitized native", index),
			Files: inputs,
			Flags: append(
				native.Flags(options),
				"split=true",
				"jobs=4",
				fmt.Sprintf("mutation=%+v", oracles.Mutations[index]),
			),
			Toolchain: runtime.Version() + "\n" + buildcache.Tool("clang", "--version"),
		}, func(directory string) error {
			path := mutatedPort(test, oracles.Mutations[index])
			started := time.Now()
			program, lowerProgramError := lowerProgram(path)
			if lowerProgramError != nil {
				return lowerProgramError
			}
			test.Logf("grain lowering: %.3fs", time.Since(started).Seconds())
			started = time.Now()
			source := native.C(program)
			test.Logf("grain C emission: %.3fs", time.Since(started).Seconds())
			started = time.Now()
			buildError := native.Build(source, filepath.Join(directory, "port"), options)
			test.Logf("grain clang runtime/compile/link: %.3fs", time.Since(started).Seconds())
			return buildError
		})
	})
	if state.Directory == "" {
		test.Fatal("native product preparation failed")
	}
	return filepath.Join(state.Directory, "port")
}

// A mutant's oracle corpus is keyed by its oracle product's key and the code that runs the oracle, never by the
// oracle's bytes or cohere. In a stand-in tree, an edit to a file the oracle compiles moves the corpus key through the
// oracle's key, an edit to each declared file moves it, and an edit to a file neither reads holds it. No declared input
// sits under cohere or outside the repository, so where the oracle's product sits (its cache path) never reaches the
// key. And the real oracles of two mutations with different witness families key their corpora apart, each the same
// asked twice. Keying builds nothing. Mutant: the oracle's key dropped from the corpus flags.
func TestMutantOracleCorpusKeyIsItsOracle(test *testing.T) {
	test.Parallel()
	stand := test.TempDir()
	write := func(name, content string) {
		test.Helper()
		path := filepath.Join(stand, filepath.FromSlash(name))
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0o755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(path, []byte(content), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	compiled := "cohere/internal/format/javascript/printer.go"
	declared := []string{"internal/childguard/childguard.go"}
	for _, file := range oracles.MutantOracleCorpusInputs(stand, 3, "").Files[1:] {
		declared = append(declared, filepath.ToSlash(mustRelative(test, stand, file)))
	}
	for _, name := range append([]string{compiled, "internal/unrelated.go"}, declared...) {
		write(name, name)
	}
	key := func() string {
		test.Helper()
		oracle, keyError := buildcache.Key(stand, buildcache.Inputs{
			Name:      "stand-in oracle",
			Files:     []string{compiled},
			Toolchain: []string{runtime.Version()},
		})
		if keyError != nil {
			test.Fatal(keyError)
		}
		inputs := oracles.MutantOracleCorpusInputs(stand, 3, oracle)
		corpus := buildcache.Inputs{Name: inputs.Name, Flags: inputs.Flags, Toolchain: []string{inputs.Toolchain}}
		for _, file := range inputs.Files {
			corpus.Files = append(corpus.Files, filepath.ToSlash(mustRelative(test, stand, file)))
		}
		key, keyError := buildcache.Key(stand, corpus)
		if keyError != nil {
			test.Fatal(keyError)
		}
		return key
	}
	before := key()
	write("internal/unrelated.go", "package unrelated // edited")
	if key() != before {
		test.Fatal("an edit to a file neither the oracle nor the corpus reads moved the corpus key")
	}
	write(compiled, compiled+"!")
	if key() == before {
		test.Fatal("an edit to a file the oracle compiles kept the corpus key")
	}
	write(compiled, compiled)
	for _, name := range declared {
		write(name, name+"!")
		if key() == before {
			test.Errorf("a one-byte edit to %s kept the corpus key", name)
		}
		write(name, name)
	}

	root, absolutePathError := filepath.Abs(oracles.Repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	for _, file := range oracles.MutantOracleCorpusInputs(root, 3, "").Files {
		relative := mustRelative(test, root, file)
		if relative == ".." ||
			strings.HasPrefix(relative, "../") ||
			relative == "cohere" ||
			strings.HasPrefix(relative, "cohere/") {
			test.Fatalf("the oracle corpus declares %s, outside the repository or under cohere", file)
		}
	}
	if fmt.Sprint(
		oracles.MutantFamilies[oracles.Mutations[3].Name],
	) == fmt.Sprint(oracles.MutantFamilies[oracles.Mutations[6].Name]) {
		test.Fatalf("mutations 3 and 6 share witness families, so their oracles can't show the key moving")
	}
	three, six := oracles.MutantOracleKey(test, root, 3), oracles.MutantOracleKey(test, root, 6)
	if three == six {
		test.Fatal("two oracles filtered to different witness families keyed alike")
	}
	if again := oracles.MutantOracleKey(test, root, 3); again != three {
		test.Fatalf("one oracle keyed twice gave %s and %s", three[:12], again[:12])
	}
	test.Logf(
		"the oracle's key moves the corpus key; mutation 3's oracle %s, 6's %s; %d declared inputs, none under cohere or outside the repository",
		three[:12],
		six[:12],
		len(declared),
	)
}

func mustRelative(test *testing.T, root, path string) string {
	test.Helper()
	relative, relativePathError := filepath.Rel(root, path)
	if relativePathError != nil {
		test.Fatal(relativePathError)
	}
	return filepath.ToSlash(relative)
}

// Each declaration owns exactly one product and uses the leaf's shared recipe.
func TestProduct_TSPrinterMutantNative_000(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 0)
}

func TestProduct_TSPrinterMutantNative_001(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 1)
}

func TestProduct_TSPrinterMutantNative_002(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 2)
}

func TestProduct_TSPrinterMutantNative_003(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 3)
}

func TestProduct_TSPrinterMutantNative_004(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 4)
}

func TestProduct_TSPrinterMutantNative_005(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 5)
}

func TestProduct_TSPrinterMutantNative_006(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 6)
}

func TestProduct_TSPrinterMutantNative_007(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 7)
}

func TestProduct_TSPrinterMutantNative_008(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 8)
}

func TestProduct_TSPrinterMutantNative_009(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 9)
}

func TestProduct_TSPrinterMutantNative_010(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 10)
}

func TestProduct_TSPrinterMutantNative_011(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 11)
}

func TestProduct_TSPrinterMutantNative_012(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 12)
}

func TestProduct_TSPrinterMutantNative_013(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 13)
}

func TestProduct_TSPrinterMutantNative_014(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 14)
}

func TestProduct_TSPrinterMutantNative_015(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 15)
}

func TestProduct_TSPrinterMutantNative_016(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 16)
}

func TestProduct_TSPrinterMutantNative_017(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 17)
}

func TestProduct_TSPrinterMutantNative_018(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 18)
}

func TestProduct_TSPrinterMutantNative_019(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 19)
}

func TestProduct_TSPrinterMutantNative_020(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 20)
}

func TestProduct_TSPrinterMutantNative_021(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 21)
}

func TestProduct_TSPrinterMutantNative_022(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 22)
}

func TestProduct_TSPrinterMutantNative_023(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 23)
}

func TestProduct_TSPrinterMutantNative_024(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 24)
}

func TestProduct_TSPrinterMutantNative_025(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 25)
}

func TestProduct_TSPrinterMutantNative_026(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 26)
}

func TestProduct_TSPrinterMutantNative_027(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 27)
}

func TestProduct_TSPrinterMutantNative_028(test *testing.T) {
	test.Parallel()
	mutantNativeProduct(test, 28)
}
