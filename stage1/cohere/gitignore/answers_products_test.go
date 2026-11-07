package gitignore

import (
	"sync"
	"testing"
)

func TestProduct_GitignoreFindsOracle(test *testing.T) { test.Parallel(); portOracleProduct(test) }
func TestProduct_GitignoreCorrect(test *testing.T)     { test.Parallel(); portProducts(test, nil) }
func TestProduct_GitignoreMutant_000(test *testing.T) {
	test.Parallel()
	portProducts(test, &mutants[0])
}
func TestProduct_GitignoreMutant_001(test *testing.T) {
	test.Parallel()
	portProducts(test, &mutants[1])
}
func TestProduct_GitignoreMutant_002(test *testing.T) {
	test.Parallel()
	portProducts(test, &mutants[2])
}
func TestProduct_GitignoreMutant_003(test *testing.T) {
	test.Parallel()
	portProducts(test, &mutants[3])
}
func TestProduct_GitignoreMutant_004(test *testing.T) {
	test.Parallel()
	portProducts(test, &mutants[4])
}
func TestProduct_GitignoreMutant_005(test *testing.T) {
	test.Parallel()
	portProducts(test, &mutants[5])
}
func TestProduct_GitignoreMutant_006(test *testing.T) {
	test.Parallel()
	portProducts(test, &mutants[6])
}
func TestProduct_GitignoreMutant_007(test *testing.T) {
	test.Parallel()
	portProducts(test, &mutants[7])
}
func TestProduct_GitignoreMutant_008(test *testing.T) {
	test.Parallel()
	portProducts(test, &mutants[8])
}
func TestProduct_GitignoreMutant_009(test *testing.T) {
	test.Parallel()
	portProducts(test, &mutants[9])
}
func TestProduct_GitignoreMutant_010(test *testing.T) {
	test.Parallel()
	portProducts(test, &mutants[10])
}

// The build phase and isolated units share these exact recipes and keys.
var portProductStates sync.Map

type portProductState struct {
	once sync.Once
	path string
}

func portCachedProduct(test *testing.T, key string, build func() string) string {
	test.Helper()
	value, _ := portProductStates.LoadOrStore(key, &portProductState{})
	state := value.(*portProductState)
	state.once.Do(func() { state.path = build() })
	if state.path == "" {
		test.Fatalf("product %s preparation failed", key)
	}
	return state.path
}
func portProducts(test *testing.T, applied *mutant) string {
	test.Helper()
	name := "correct"
	if applied != nil {
		name = applied.name
	}
	return portCachedProduct(test, "port/"+name, func() string { return portBuildProduct(test, applied) })
}
func portOracleProduct(test *testing.T) string {
	test.Helper()
	return portCachedProduct(test, "oracle", func() string { return portBuildOracleProduct(test) })
}
