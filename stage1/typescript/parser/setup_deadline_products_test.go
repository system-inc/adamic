package parser

import (
	"path/filepath"
	"testing"
)

func compilerExpressionsProductDirectory(t *testing.T) string {
	t.Helper()
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	return directory
}

func wholeMutantsProductSource(t *testing.T, index int) string {
	t.Helper()
	if index < 0 {
		return compilerExpressionsProductDirectory(t)
	}
	mutation := wholeMutantCases()[index]
	return copyPort(t, mutation.file, mutation.from, mutation.to)
}

// Each build-phase unit invokes the recipe used by the shards.
func TestProduct_ParserOracle(t *testing.T) {
	t.Parallel()
	wholeMutantBuildOracleProduct(t)
}
func TestProduct_CompilerExpressionsLower(t *testing.T) {
	t.Parallel()
	compilerExpressionsLower(t, compilerExpressionsProductDirectory(t))
}
func TestProduct_CompilerExpressionsNative(t *testing.T) {
	t.Parallel()
	compilerExpressionsNative(t, compilerExpressionsProductDirectory(t))
}

func TestProduct_WholeMutantsLower_Control(t *testing.T) {
	t.Parallel()
	wholeMutantLowerProduct(t, wholeMutantsProductSource(t, -1))
}

func TestProduct_WholeMutantsNative_Control(t *testing.T) {
	t.Parallel()
	wholeMutantBuildPortProduct(t, wholeMutantsProductSource(t, -1), true)
}

func TestProduct_WholeMutantsLower_000(t *testing.T) {
	t.Parallel()
	wholeMutantLowerProduct(t, wholeMutantsProductSource(t, 0))
}

func TestProduct_WholeMutantsNative_000(t *testing.T) {
	t.Parallel()
	wholeMutantBuildPortProduct(t, wholeMutantsProductSource(t, 0), true)
}

func TestProduct_WholeMutantsLower_001(t *testing.T) {
	t.Parallel()
	wholeMutantLowerProduct(t, wholeMutantsProductSource(t, 1))
}

func TestProduct_WholeMutantsNative_001(t *testing.T) {
	t.Parallel()
	wholeMutantBuildPortProduct(t, wholeMutantsProductSource(t, 1), true)
}

func TestProduct_WholeMutantsLower_002(t *testing.T) {
	t.Parallel()
	wholeMutantLowerProduct(t, wholeMutantsProductSource(t, 2))
}

func TestProduct_WholeMutantsNative_002(t *testing.T) {
	t.Parallel()
	wholeMutantBuildPortProduct(t, wholeMutantsProductSource(t, 2), true)
}
