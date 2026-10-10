package gitignore

import (
	"sync"
	"testing"
)

func TestProduct_GitignoreOracle(t *testing.T)     { t.Parallel(); portOracleProduct(t) }
func TestProduct_GitignoreCorrect(t *testing.T)    { t.Parallel(); portProducts(t, nil) }
func TestProduct_GitignoreMutant_000(t *testing.T) { t.Parallel(); portProducts(t, &mutants[0]) }
func TestProduct_GitignoreMutant_001(t *testing.T) { t.Parallel(); portProducts(t, &mutants[1]) }
func TestProduct_GitignoreMutant_002(t *testing.T) { t.Parallel(); portProducts(t, &mutants[2]) }
func TestProduct_GitignoreMutant_003(t *testing.T) { t.Parallel(); portProducts(t, &mutants[3]) }
func TestProduct_GitignoreMutant_004(t *testing.T) { t.Parallel(); portProducts(t, &mutants[4]) }
func TestProduct_GitignoreMutant_005(t *testing.T) { t.Parallel(); portProducts(t, &mutants[5]) }
func TestProduct_GitignoreMutant_006(t *testing.T) { t.Parallel(); portProducts(t, &mutants[6]) }
func TestProduct_GitignoreMutant_007(t *testing.T) { t.Parallel(); portProducts(t, &mutants[7]) }
func TestProduct_GitignoreMutant_008(t *testing.T) { t.Parallel(); portProducts(t, &mutants[8]) }
func TestProduct_GitignoreMutant_009(t *testing.T) { t.Parallel(); portProducts(t, &mutants[9]) }
func TestProduct_GitignoreMutant_010(t *testing.T) { t.Parallel(); portProducts(t, &mutants[10]) }

// The build phase and isolated units share these exact recipes and keys.
var portProductStates sync.Map

type portProductState struct {
	once sync.Once
	path string
}

func portCachedProduct(t *testing.T, key string, build func() string) string {
	t.Helper()
	value, _ := portProductStates.LoadOrStore(key, &portProductState{})
	state := value.(*portProductState)
	state.once.Do(func() { state.path = build() })
	if state.path == "" {
		t.Fatalf("product %s preparation failed", key)
	}
	return state.path
}
func portProducts(t *testing.T, applied *mutant) string {
	t.Helper()
	name := "correct"
	if applied != nil {
		name = applied.name
	}
	return portCachedProduct(t, "port/"+name, func() string { return portBuildProduct(t, applied) })
}
func portOracleProduct(t *testing.T) string {
	t.Helper()
	return portCachedProduct(t, "oracle", func() string { return portBuildOracleProduct(t) })
}
