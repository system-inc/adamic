package cssstrings

import (
	"os"
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

// The oracle answers a batch or raw unit reads. The unit and its TestProduct_ in answer_products_test.go both call
// these, so they key one product and a runner reading the cache never builds one.
func stringsBatchGoAnswers(t *testing.T, oracle, oracleIdentity string, corpus stringsCorpus, unit stringsUnit) run {
	return stringsOracleAnswers(t, "Go", oracle, nil, stringsInput(corpus.texts[unit.lo:unit.hi]), oracleIdentity)
}

func stringsBatchPrettierAnswers(t *testing.T, library, libraryIdentity string, corpus stringsCorpus, unit stringsUnit) run {
	script, err := filepath.Abs("testdata/library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	return stringsOracleAnswers(t, "Prettier", "node", []string{script, library}, stringsInput(corpus.texts[unit.lo:unit.hi]), libraryIdentity)
}

func stringsRawGoAnswers(t *testing.T, oracle, oracleIdentity, text, preference string) run {
	return stringsOracleAnswers(t, "Go", oracle, nil, []byte(preference+stringsEncode.Replace(text)+"\n"), oracleIdentity)
}

// What every answer product's key needs, read once per process: the corpus and its units, and the oracles by
// content. Reading it builds nothing new, since the Go oracle is its own product.
var stringsAnswerInputs struct {
	once            sync.Once
	corpus          stringsCorpus
	units           []stringsUnit
	oracle          string
	oracleIdentity  string
	library         string
	libraryIdentity string
}

func prepareStringsAnswerInputs(t *testing.T) {
	t.Helper()
	stringsAnswerInputs.once.Do(func() {
		corpus := enumerateStrings(t)
		units := stringsUnits(corpus)
		if len(units) != testCSSStringsShards {
			t.Fatalf("enumerated %d shards, declared %d", len(units), testCSSStringsShards)
		}
		oracle := stringsPreparedOracle(t)
		library := os.Getenv("ADAMIC_CSSSTRINGS_LIBRARY")
		if library != "" {
			stringsAnswerInputs.library, stringsAnswerInputs.libraryIdentity = library, stringsOracleIdentity(t, library)
		}
		stringsAnswerInputs.corpus, stringsAnswerInputs.units = corpus, units
		stringsAnswerInputs.oracle, stringsAnswerInputs.oracleIdentity = oracle, stringsOracleIdentity(t, oracle)
	})
	if stringsAnswerInputs.oracleIdentity == "" {
		t.Fatal("answer inputs did not complete")
	}
}

// stringsGoAnswersProduct builds the Go answers shard ordinal reads: a batch unit's (no preference), or one
// preference of a raw unit's.
func stringsGoAnswersProduct(t *testing.T, ordinal int, preference string) {
	t.Helper()
	prepareStringsAnswerInputs(t)
	unit := stringsAnswerInputs.units[ordinal]
	switch {
	case unit.kind == "batch" && preference == "":
		clean(t, "Go", stringsBatchGoAnswers(t, stringsAnswerInputs.oracle, stringsAnswerInputs.oracleIdentity, stringsAnswerInputs.corpus, unit))
	case unit.kind == "raw" && (preference == "d" || preference == "s"):
		clean(t, "Go raw", stringsRawGoAnswers(t, stringsAnswerInputs.oracle, stringsAnswerInputs.oracleIdentity, stringsAnswerInputs.corpus.raw[unit.lo], preference))
	default:
		t.Fatalf("%s is a %s unit, with no Go answers for preference %q", unit.name, unit.kind, preference)
	}
}

// stringsPrettierAnswersProduct builds the Prettier answers batch shard ordinal reads. Without the external
// library the unit reads none, so there is nothing to build.
func stringsPrettierAnswersProduct(t *testing.T, ordinal int) {
	t.Helper()
	if os.Getenv("ADAMIC_CSSSTRINGS_LIBRARY") == "" {
		t.Log("external library not checked: set ADAMIC_CSSSTRINGS_LIBRARY")
		return
	}
	prepareStringsAnswerInputs(t)
	unit := stringsAnswerInputs.units[ordinal]
	if unit.kind != "batch" {
		t.Fatalf("%s is a %s unit, with no Prettier answers", unit.name, unit.kind)
	}
	clean(t, "Prettier", stringsBatchPrettierAnswers(t, stringsAnswerInputs.library, stringsAnswerInputs.libraryIdentity, stringsAnswerInputs.corpus, unit))
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
