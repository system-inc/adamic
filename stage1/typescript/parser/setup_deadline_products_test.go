package parser

import (
	"path/filepath"
	"testing"
)

func compilerExpressionsProductDirectory(test *testing.T) string {
	test.Helper()
	directory, absolutePathError := filepath.Abs(".")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return directory
}

func wholeMutantsProductSource(test *testing.T, index int) string {
	test.Helper()
	if index < 0 {
		return compilerExpressionsProductDirectory(test)
	}
	mutation := wholeMutantCases()[index]
	return copyPort(test, mutation.file, mutation.from, mutation.to)
}

// Each build-phase unit invokes the recipe used by the shards.
func TestProduct_ParserFindsOracle(test *testing.T) {
	test.Parallel()
	wholeMutantBuildOracleProduct(test)
}
func TestProduct_CompilerExpressionsLower(test *testing.T) {
	test.Parallel()
	compilerExpressionsLower(test, compilerExpressionsProductDirectory(test))
}
func TestProduct_CompilerExpressionsNative(test *testing.T) {
	test.Parallel()
	compilerExpressionsNative(test, compilerExpressionsProductDirectory(test))
}

func TestProduct_WholeMutantsLower_Control(test *testing.T) {
	test.Parallel()
	wholeMutantLowerProduct(test, wholeMutantsProductSource(test, -1))
}

func TestProduct_WholeMutantsNative_Control(test *testing.T) {
	test.Parallel()
	wholeMutantBuildPortProduct(test, wholeMutantsProductSource(test, -1), true)
}

func TestProduct_WholeMutantsLower_000(test *testing.T) {
	test.Parallel()
	wholeMutantLowerProduct(test, wholeMutantsProductSource(test, 0))
}

func TestProduct_WholeMutantsNative_000(test *testing.T) {
	test.Parallel()
	wholeMutantBuildPortProduct(test, wholeMutantsProductSource(test, 0), true)
}

func TestProduct_WholeMutantsLower_001(test *testing.T) {
	test.Parallel()
	wholeMutantLowerProduct(test, wholeMutantsProductSource(test, 1))
}

func TestProduct_WholeMutantsNative_001(test *testing.T) {
	test.Parallel()
	wholeMutantBuildPortProduct(test, wholeMutantsProductSource(test, 1), true)
}

func TestProduct_WholeMutantsLower_002(test *testing.T) {
	test.Parallel()
	wholeMutantLowerProduct(test, wholeMutantsProductSource(test, 2))
}

func TestProduct_WholeMutantsNative_002(test *testing.T) {
	test.Parallel()
	wholeMutantBuildPortProduct(test, wholeMutantsProductSource(test, 2), true)
}
