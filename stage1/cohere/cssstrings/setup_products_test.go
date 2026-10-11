package cssstrings

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
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

// The oracle answers a batch unit reads. The unit and its TestProduct_ in answer_products_test.go both call these,
// so they key one product and a runner reading the cache never builds one.
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

// stringsRawGoAnswer is one raw case's Go answer, read from the one product that holds every raw answer, which
// TestProduct_CSSStringsGoRawAnswers builds through the same stringsRawAnswersDirectory.
func stringsRawGoAnswer(t *testing.T, oracle, oracleIdentity string, corpus stringsCorpus, index int, preference string) run {
	t.Helper()
	answer, err := os.ReadFile(filepath.Join(stringsRawAnswersDirectory(t, oracle, oracleIdentity, corpus), stringsRawAnswerName(index, preference)))
	if err != nil {
		t.Fatal(err)
	}
	return run{stdout: answer}
}

func stringsRawAnswerName(index int, preference string) string {
	return fmt.Sprintf("raw-%06d-%s.txt", index, preference)
}

func stringsRawCase(text, preference string) []byte {
	return []byte(preference + stringsEncode.Replace(text) + "\n")
}

// Every raw unit reads the raw answers product, and a lookup in read mode values its whole read set, so a process
// keys and looks it up once.
var stringsRawAnswers struct {
	once      sync.Once
	directory string
}

// stringsRawAnswersDirectory is the product holding every raw case's Go answer, one file per raw index and
// preference. Each answer is still its own single-case oracle run on the case's own input, as a raw unit's was
// when it ran the oracle itself; the build runs them GOMAXPROCS at a time.
func stringsRawAnswersDirectory(t *testing.T, oracle, oracleIdentity string, corpus stringsCorpus) string {
	t.Helper()
	stringsRawAnswers.once.Do(func() {
		var cases []byte
		for _, text := range corpus.raw {
			for _, preference := range []string{"d", "s"} {
				cases = append(cases, stringsRawCase(text, preference)...)
			}
		}
		stringsRawAnswers.directory = stringsProduct(t, stringsInputs{Name: "Go raw oracle answers", Flags: []string{oracleIdentity, fmt.Sprintf("raw-cases-sha256=%x", sha256.Sum256(cases))}, Toolchain: runtime.Version()}, func(dir string) error {
			type rawCase struct {
				index      int
				preference string
			}
			pending := make(chan rawCase)
			failures := make(chan error, 2*len(corpus.raw))
			var workers sync.WaitGroup
			for range runtime.GOMAXPROCS(0) {
				workers.Add(1)
				go func() {
					defer workers.Done()
					for each := range pending {
						if err := stringsRawAnswerRun(t, oracle, dir, each.index, each.preference, corpus.raw[each.index]); err != nil {
							failures <- err
						}
					}
				}()
			}
			for index := range corpus.raw {
				for _, preference := range []string{"d", "s"} {
					pending <- rawCase{index, preference}
				}
			}
			close(pending)
			workers.Wait()
			close(failures)
			return <-failures
		})
	})
	if stringsRawAnswers.directory == "" {
		t.Fatal("raw answers did not complete")
	}
	return stringsRawAnswers.directory
}

// stringsRawAnswerRun runs the oracle on one raw case and keeps its stdout; the input file leaves with the run.
func stringsRawAnswerRun(t *testing.T, oracle, dir string, index int, preference, text string) error {
	name := stringsRawAnswerName(index, preference)
	input := filepath.Join(dir, name+".input")
	if err := os.WriteFile(input, stringsRawCase(text, preference), 0644); err != nil {
		return err
	}
	defer os.Remove(input)
	command := stringsCommand(t, oracle, input)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := childguard.Run(command, childguard.Options{}); err != nil || stderr.Len() != 0 {
		return fmt.Errorf("Go oracle on raw %d %s: %v; %s", index, preference, err, &stderr)
	}
	return os.WriteFile(filepath.Join(dir, name), stdout.Bytes(), 0644)
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

// stringsGoAnswersProduct builds the Go answers batch shard ordinal reads.
func stringsGoAnswersProduct(t *testing.T, ordinal int) {
	t.Helper()
	prepareStringsAnswerInputs(t)
	unit := stringsAnswerInputs.units[ordinal]
	if unit.kind != "batch" {
		t.Fatalf("%s is a %s unit, with no batch Go answers", unit.name, unit.kind)
	}
	clean(t, "Go", stringsBatchGoAnswers(t, stringsAnswerInputs.oracle, stringsAnswerInputs.oracleIdentity, stringsAnswerInputs.corpus, unit))
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
func TestProduct_CSSStringsGoRawAnswers(t *testing.T) {
	t.Parallel()
	prepareStringsAnswerInputs(t)
	stringsRawAnswersDirectory(t, stringsAnswerInputs.oracle, stringsAnswerInputs.oracleIdentity, stringsAnswerInputs.corpus)
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
