package css

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

var cssParserPreparedPrograms = make([]struct {
	once    sync.Once
	program cssParserProgram
	failed  string
}, len(mutants)+1)

// cssPlantedSetupFailure is how TestCSSSetupFailuresNameTheFirst plants a failure in a shared setup's Once:
// ADAMIC_CSS_PLANTED_SETUP_FAILURE names the setup ("program" or "composition"), and its first caller fails.
func cssPlantedSetupFailure(t *testing.T, setup string) {
	t.Helper()
	if os.Getenv("ADAMIC_CSS_PLANTED_SETUP_FAILURE") == setup {
		t.Fatalf("planted: the %s setup fails", setup)
	}
}

// cssParserPreparedProgram lowers the parser, or one of its mutants, once per process. A lowering that fails inside the
// Once fails its caller; every later caller then fails naming the test that holds the real error, rather than taking
// the zero program on to look up a nameless product (#qbbmqfs).
func cssParserPreparedProgram(t *testing.T, mutation int) cssParserProgram {
	t.Helper()
	state := &cssParserPreparedPrograms[mutation+1]
	state.once.Do(func() {
		state.failed = t.Name()
		cssPlantedSetupFailure(t, "program")
		main, err := filepath.Abs("main.ts")
		if err != nil {
			t.Fatal(err)
		}
		name, directory := "parser", filepath.Dir(main)
		if mutation >= 0 {
			name = fmt.Sprintf("mutant-%d", mutation)
			directory = cssParserTopPortDirectory(t, &mutants[mutation])
		}
		state.program = buildCSSParserProgram(t, name, directory)
		state.failed = ""
	})
	if state.failed != "" {
		t.Fatalf("the prepared css parser program %d failed earlier in this process, in %s (see that test's log)", mutation, state.failed)
	}
	return state.program
}

func cssParserPreparedNative(t *testing.T, mutation int) string {
	t.Helper()
	return buildCSSParserBinaries(t, []cssParserProgram{cssParserPreparedProgram(t, mutation)})[0]
}

// prepareCSSParserCompilerProducts needs no oracle, even on a cold tree. The
// runtime comparison setup fetches the corpus independently after these products.
func prepareCSSParserCompilerProducts(t *testing.T) {
	t.Helper()
	for mutation := -1; mutation < len(mutants); mutation++ {
		cssParserPreparedNative(t, mutation)
	}
	if runtime.GOOS == "darwin" {
		cssParserPreparedLeaks(t)
	}
}

func cssParserPreparedLeaks(t *testing.T) string {
	t.Helper()
	return buildCSSParserUnsanitized(t, cssParserPreparedProgram(t, -1))
}

// Each Go oracle's key moves with the C or C++ compiler go builds it with, and with the C compiler's report, so a
// changed compiler misses rather than reading a product another one built.
func TestCSSGoOraclesRekeyCompiler(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	compilers := cssCompilers{"clang", "clang++", "clang version 18.1.8"}
	for _, oracle := range []struct {
		name   string
		inputs func(cssCompilers) buildcache.Inputs
	}{
		{"parser oracle", func(compilers cssCompilers) buildcache.Inputs {
			return cssParserGoProductInputs("Go parser oracle", compilers)
		}},
		{"composition corpus oracle", func(compilers cssCompilers) buildcache.Inputs {
			return compositionOracleInputs("css-composition-corpus-oracle", "stage1/cohere/css/testdata/composition_corpus_side_test.go", "internal/format/css/postcss/adamic_port_side_test.go", "./internal/format/css/postcss", compilers)
		}},
		{"composition Go oracle", func(compilers cssCompilers) buildcache.Inputs {
			return compositionOracleInputs("css-composition-go-oracle", "stage1/cohere/css/testdata/compose_side_test.go", "internal/format/css/adamic_compose_side_test.go", "./internal/format/css", compilers)
		}},
	} {
		key := buildcache.NameKey(root, oracle.inputs(compilers))
		for _, other := range []cssCompilers{{"gcc", "clang++", compilers.report}, {"clang", "g++", compilers.report}, {"clang", "clang++", "clang version 19.1.7"}} {
			if otherKey := buildcache.NameKey(root, oracle.inputs(other)); otherKey == key {
				t.Fatalf("%s: compilers %+v kept its key %s", oracle.name, other, key[:12])
			}
		}
		t.Logf("%s key %s moves with CC, CXX and the C compiler's report", oracle.name, key[:12])
	}
}

// The compilers the Go oracles' keys name are read from toolchain.Env, the environment go runs under, not from the
// caller's: with CC and CXX unset that is clang and clang++, which a reader of os.Getenv would key as empty while go ran
// clang, and a CC the caller sets moves the key. Not parallel: it sets CC and CXX.
func TestCSSGoOraclesKeyTheCompilersEnvGivesGo(t *testing.T) {
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CC", "")
	t.Setenv("CXX", "")
	defaults := cssGoCompilers()
	inputs := cssParserGoProductInputs("Go parser oracle", defaults)
	if !slices.Contains(inputs.Flags, "CC=clang") || !slices.Contains(inputs.Flags, "CXX=clang++") || !slices.Contains(inputs.Toolchain, buildcache.Tool("clang", "--version")) {
		t.Fatalf("with CC and CXX unset, the parser oracle's key names %+v, want clang, clang++ and clang's report", defaults)
	}
	t.Setenv("CC", "cc")
	moved := cssGoCompilers()
	if moved.c != "cc" || moved.report == defaults.report {
		t.Fatalf("CC=cc read as %+v", moved)
	}
	if buildcache.NameKey(root, cssParserGoProductInputs("Go parser oracle", moved)) == buildcache.NameKey(root, inputs) {
		t.Fatal("CC=cc kept the parser oracle's key")
	}
	t.Logf("unset: %s and %s; CC=cc moves the key", defaults.c, defaults.cPlusPlus)
}

func TestProduct_CSSParserParserLowered(t *testing.T) { t.Parallel(); cssParserPreparedProgram(t, -1) }
func TestProduct_CSSParserParserNative(t *testing.T)  { t.Parallel(); cssParserPreparedNative(t, -1) }
func TestProduct_CSSParserCustomPropertyMutantLowered(t *testing.T) {
	t.Parallel()
	cssParserPreparedProgram(t, 0)
}
func TestProduct_CSSParserCustomPropertyMutantNative(t *testing.T) {
	t.Parallel()
	cssParserPreparedNative(t, 0)
}
func TestProduct_CSSParserCommentMutantLowered(t *testing.T) {
	t.Parallel()
	cssParserPreparedProgram(t, 1)
}
func TestProduct_CSSParserCommentMutantNative(t *testing.T) {
	t.Parallel()
	cssParserPreparedNative(t, 1)
}
func TestProduct_CSSParserClosingBraceMutantLowered(t *testing.T) {
	t.Parallel()
	cssParserPreparedProgram(t, 2)
}
func TestProduct_CSSParserClosingBraceMutantNative(t *testing.T) {
	t.Parallel()
	cssParserPreparedNative(t, 2)
}
func TestProduct_CSSParserDarwinLeaks(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin-only product")
	}
	cssParserPreparedLeaks(t)
}

func TestProduct_CSSCompositionLowered(t *testing.T) {
	t.Parallel()
	compositionLoweredProduct(t)
}
func TestProduct_CSSCompositionNative(t *testing.T) {
	t.Parallel()
	compositionNativeProduct(t)
}
