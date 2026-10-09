package cssstrings

import (
	"path/filepath"
	"sync"
	"testing"
)

var stringsPreparedPrograms = make([]struct {
	once    sync.Once
	program stringsProgram
}, len(stringsMutations)+1)

func stringsPreparedProgram(t *testing.T, mutation int) stringsProgram {
	t.Helper()
	state := &stringsPreparedPrograms[mutation+1]
	state.once.Do(func() {
		main, err := filepath.Abs("main.ts")
		if err != nil {
			t.Fatal(err)
		}
		name := "port"
		if mutation >= 0 {
			change := stringsMutations[mutation]
			name = change.name
			main = prepareStringsMutant(t, change)
		}
		state.program = buildStringsProgram(t, name, main)
	})
	return state.program
}

func stringsPreparedNative(t *testing.T, mutation int, sanitize bool) string {
	t.Helper()
	name := "native-fast"
	if sanitize {
		name = "sanitized"
	}
	if mutation >= 0 {
		name = stringsMutations[mutation].name + "-sanitized"
	}
	return buildStringsNative(t, name, stringsPreparedProgram(t, mutation), sanitize)
}

func TestProduct_CSSStringsGoOracle(t *testing.T) { t.Parallel(); stringsPreparedOracle(t) }
func TestProduct_CSSStringsGoAnswers(t *testing.T) {
	t.Parallel()
	corpus := enumerateStrings(t)
	oracle := stringsPreparedOracle(t)
	clean(t, "Go", stringsSetupOracleAnswers(t, "Go", oracle, nil, stringsInput(corpus.texts), stringsOracleIdentity(t, oracle)))
}
func TestProduct_CSSStringsPortLowered(t *testing.T)    { t.Parallel(); stringsPreparedProgram(t, -1) }
func TestProduct_CSSStringsPortNative(t *testing.T)     { t.Parallel(); stringsPreparedNative(t, -1, true) }
func TestProduct_CSSStringsMutant0Lowered(t *testing.T) { t.Parallel(); stringsPreparedProgram(t, 0) }
func TestProduct_CSSStringsMutant0Native(t *testing.T) {
	t.Parallel()
	stringsPreparedNative(t, 0, true)
}
func TestProduct_CSSStringsMutant1Lowered(t *testing.T) { t.Parallel(); stringsPreparedProgram(t, 1) }
func TestProduct_CSSStringsMutant1Native(t *testing.T) {
	t.Parallel()
	stringsPreparedNative(t, 1, true)
}
func TestProduct_CSSStringsMutant2Lowered(t *testing.T) { t.Parallel(); stringsPreparedProgram(t, 2) }
func TestProduct_CSSStringsMutant2Native(t *testing.T) {
	t.Parallel()
	stringsPreparedNative(t, 2, true)
}
func TestProduct_CSSStringsFastNative(t *testing.T) {
	t.Parallel()
	stringsPreparedNative(t, -1, false)
}
