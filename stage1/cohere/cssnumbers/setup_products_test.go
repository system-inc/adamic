package cssnumbers

import (
	"path/filepath"
	"sync"
	"testing"
)

var numbersPreparedPrograms = make([]struct {
	once    sync.Once
	program numbersProgram
}, len(numbersMutations)+1)

func numbersPreparedProgram(t *testing.T, mutation int) numbersProgram {
	t.Helper()
	state := &numbersPreparedPrograms[mutation+1]
	state.once.Do(func() {
		main, err := filepath.Abs("main.ts")
		if err != nil {
			t.Fatal(err)
		}
		name := "port"
		if mutation >= 0 {
			change := numbersMutations[mutation]
			name = change.name
			main = prepareNumbersMutant(t, change)
		}
		state.program = buildNumbersProgram(t, name, main)
	})
	return state.program
}

func numbersPreparedNative(t *testing.T, mutation int, sanitize bool) string {
	t.Helper()
	name := "native-fast"
	if sanitize {
		name = "sanitized"
	}
	if mutation >= 0 {
		name = numbersMutations[mutation].name + "-sanitized"
	}
	return buildNumbersNative(t, name, numbersPreparedProgram(t, mutation), sanitize)
}

func TestProduct_CSSNumbersGoOracle(t *testing.T) { t.Parallel(); numbersPreparedOracle(t) }
func TestProduct_CSSNumbersGoAnswers(t *testing.T) {
	t.Parallel()
	corpus := enumerateNumbers(t)
	oracle := numbersPreparedOracle(t)
	clean(t, "Go", numbersSetupOracleAnswers(t, "Go", oracle, nil, numbersInput(corpus.texts), numbersOracleIdentity(t, oracle)))
}
func TestProduct_CSSNumbersPortLowered(t *testing.T)    { t.Parallel(); numbersPreparedProgram(t, -1) }
func TestProduct_CSSNumbersPortNative(t *testing.T)     { t.Parallel(); numbersPreparedNative(t, -1, true) }
func TestProduct_CSSNumbersMutant0Lowered(t *testing.T) { t.Parallel(); numbersPreparedProgram(t, 0) }
func TestProduct_CSSNumbersMutant0Native(t *testing.T) {
	t.Parallel()
	numbersPreparedNative(t, 0, true)
}
func TestProduct_CSSNumbersMutant1Lowered(t *testing.T) { t.Parallel(); numbersPreparedProgram(t, 1) }
func TestProduct_CSSNumbersMutant1Native(t *testing.T) {
	t.Parallel()
	numbersPreparedNative(t, 1, true)
}
func TestProduct_CSSNumbersMutant2Lowered(t *testing.T) { t.Parallel(); numbersPreparedProgram(t, 2) }
func TestProduct_CSSNumbersMutant2Native(t *testing.T) {
	t.Parallel()
	numbersPreparedNative(t, 2, true)
}
func TestProduct_CSSNumbersFastNative(t *testing.T) {
	t.Parallel()
	numbersPreparedNative(t, -1, false)
}
