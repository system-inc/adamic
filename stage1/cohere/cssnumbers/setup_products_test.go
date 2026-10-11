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

// numbersPreparedNatives holds each native product once per process: [mutation+1][0] fast, [1] sanitized.
var numbersPreparedNatives = make([][2]struct {
	once sync.Once
	path string
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
	if state.program.dir == "" {
		t.Fatal("lowering product was not prepared")
	}
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
	index := 0
	if sanitize {
		index = 1
	}
	state := &numbersPreparedNatives[mutation+1][index]
	state.once.Do(func() { state.path = buildNumbersNative(t, name, numbersPreparedProgram(t, mutation), sanitize) })
	if state.path == "" {
		t.Fatalf("native product %s was not prepared", name)
	}
	return state.path
}

// numbersPrepareUnitProducts fetches the products one unit reads and no other kind needs. A shard calls it before
// its clock starts, so shared setup holds only what every unit reads, and a mutant's builds wait for nobody else.
func numbersPrepareUnitProducts(t *testing.T, unit numbersUnit) {
	t.Helper()
	switch unit.kind {
	case "batch":
		numbersPreparedProgram(t, -1)
		numbersPreparedNative(t, -1, true)
		numbersPreparedNative(t, -1, false)
	case "raw":
		numbersPreparedNative(t, -1, true)
	case "mutant":
		numbersPreparedNative(t, unit.lo, true)
	case "throughput":
		if unit.side == "native" {
			numbersPreparedNative(t, -1, false)
		}
	}
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
