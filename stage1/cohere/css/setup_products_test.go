package css

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

var cssParserPreparedPrograms = make([]struct {
	once    sync.Once
	program cssParserProgram
}, len(mutants)+1)

func cssParserPreparedProgram(t *testing.T, mutation int) cssParserProgram {
	t.Helper()
	state := &cssParserPreparedPrograms[mutation+1]
	state.once.Do(func() {
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
	})
	return state.program
}

func cssParserPreparedNative(t *testing.T, mutation int) string {
	t.Helper()
	return buildCSSParserBinaries(t, []cssParserProgram{cssParserPreparedProgram(t, mutation)})[0]
}

func cssParserPreparedLeaks(t *testing.T) string {
	t.Helper()
	return buildCSSParserUnsanitized(t, cssParserPreparedProgram(t, -1))
}

func TestProduct_CSSParserGoOracle(t *testing.T)      { t.Parallel(); cssParserOracle(t) }
func TestProduct_CSSParserCorpus(t *testing.T)        { t.Parallel(); prepareCSSParserTopCorpus(t) }
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

func TestProduct_CSSCompositionCorpusOracle(t *testing.T) {
	t.Parallel()
	compositionOracle(t, "css-composition-corpus-oracle", "testdata/composition_corpus_side_test.go", "internal/format/css/postcss/adamic_port_side_test.go", "./internal/format/css/postcss")
}
func TestProduct_CSSCompositionGoOracle(t *testing.T) {
	t.Parallel()
	compositionOracle(t, "css-composition-go-oracle", "testdata/compose_side_test.go", "internal/format/css/adamic_compose_side_test.go", "./internal/format/css")
}
func TestProduct_CSSCompositionLowered(t *testing.T) {
	t.Parallel()
	compositionLoweredProduct(t)
}
func TestProduct_CSSCompositionNative(t *testing.T) {
	t.Parallel()
	compositionNativeProduct(t)
}
