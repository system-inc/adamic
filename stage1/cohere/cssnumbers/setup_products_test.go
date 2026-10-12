package cssnumbers

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
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

// The oracle answers the batch and raw units compare against are three products: every batch's Go answers, every
// batch's Prettier answers and every raw case's Go answers. Each answer is still its own oracle run on the input its
// unit ran the oracle on before, kept as one file per batch or raw case, so a unit compares the same bytes, and a
// build runs them GOMAXPROCS at a time. A unit and the product's TestProduct_ fetch it through the same function.
type numbersAnswerCase struct {
	file  string
	input []byte
}

// A process keys and looks up each answer product once: the key hashes every case's input, and every batch or raw
// unit reads the same product.
type numbersAnswerProduct struct {
	once      sync.Once
	directory string
}

var numbersGoBatchAnswers, numbersPrettierBatchAnswers, numbersGoRawAnswers numbersAnswerProduct

func numbersBatchAnswerFile(unit numbersUnit) string {
	return fmt.Sprintf("batch-%06d-%06d.txt", unit.lo, unit.hi-1)
}

func numbersRawAnswerFile(index int, preference string) string {
	return fmt.Sprintf("raw-%06d-%s.txt", index, preference)
}

func numbersBatchCases(corpus numbersCorpus) []numbersAnswerCase {
	var cases []numbersAnswerCase
	for _, unit := range numbersUnits(corpus) {
		if unit.kind == "batch" {
			cases = append(cases, numbersAnswerCase{numbersBatchAnswerFile(unit), numbersInput(corpus.texts[unit.lo:unit.hi])})
		}
	}
	return cases
}

func numbersRawCases(corpus numbersCorpus) []numbersAnswerCase {
	var cases []numbersAnswerCase
	for index, text := range corpus.raw {
		for _, preference := range []string{"d", "s"} {
			cases = append(cases, numbersAnswerCase{numbersRawAnswerFile(index, preference), []byte(preference + numbersEncode.Replace(text) + "\n")})
		}
	}
	return cases
}

// numbersAnswerProductInputs keys an answer product by its oracle, every case's file name and input, and this file,
// which holds the recipe.
func numbersAnswerProductInputs(name, identity, toolchain string, files []string, cases []numbersAnswerCase) numbersInputs {
	hash := sha256.New()
	for _, each := range cases {
		fmt.Fprintf(hash, "%s %d\n", each.file, len(each.input))
		hash.Write(each.input)
	}
	return numbersInputs{Name: name, Files: append([]string{"setup_products_test.go"}, files...), Flags: []string{identity, fmt.Sprintf("cases-sha256=%x", hash.Sum(nil))}, Toolchain: toolchain}
}

func numbersAnswerDirectory(t *testing.T, product *numbersAnswerProduct, prepare func() (numbersInputs, []numbersAnswerCase), command func(t *testing.T, input string) *exec.Cmd) string {
	t.Helper()
	product.once.Do(func() {
		inputs, cases := prepare()
		product.directory = numbersProduct(t, inputs, func(dir string) error {
			pending := make(chan numbersAnswerCase)
			failures := make(chan error, len(cases))
			var workers sync.WaitGroup
			for range runtime.GOMAXPROCS(0) {
				workers.Add(1)
				go func() {
					defer workers.Done()
					for each := range pending {
						if err := numbersAnswerRun(t, dir, each, command); err != nil {
							failures <- err
						}
					}
				}()
			}
			for _, each := range cases {
				pending <- each
			}
			close(pending)
			workers.Wait()
			close(failures)
			return <-failures
		})
	})
	if product.directory == "" {
		t.Fatal("answer product did not complete")
	}
	return product.directory
}

// numbersAnswerRun runs the oracle on one case and keeps its stdout; the input file leaves with the run.
func numbersAnswerRun(t *testing.T, dir string, each numbersAnswerCase, oracleCommand func(t *testing.T, input string) *exec.Cmd) error {
	input := filepath.Join(dir, each.file+".input")
	if err := os.WriteFile(input, each.input, 0644); err != nil {
		return err
	}
	defer os.Remove(input)
	command := oracleCommand(t, input)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := childguard.Run(command, childguard.Options{}); err != nil || stderr.Len() != 0 {
		return fmt.Errorf("oracle on %s: %v; %s", each.file, err, &stderr)
	}
	return os.WriteFile(filepath.Join(dir, each.file), stdout.Bytes(), 0644)
}

func numbersGoBatchAnswersDirectory(t *testing.T, oracle, oracleIdentity string, corpus numbersCorpus) string {
	t.Helper()
	return numbersAnswerDirectory(t, &numbersGoBatchAnswers, func() (numbersInputs, []numbersAnswerCase) {
		cases := numbersBatchCases(corpus)
		return numbersAnswerProductInputs("Go batch oracle answers", oracleIdentity, runtime.Version(), nil, cases), cases
	}, func(t *testing.T, input string) *exec.Cmd { return numbersCommand(t, oracle, input) })
}

func numbersPrettierBatchAnswersDirectory(t *testing.T, library, libraryIdentity string, corpus numbersCorpus) string {
	t.Helper()
	script, err := filepath.Abs("testdata/library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	return numbersAnswerDirectory(t, &numbersPrettierBatchAnswers, func() (numbersInputs, []numbersAnswerCase) {
		cases := numbersBatchCases(corpus)
		return numbersAnswerProductInputs("Prettier batch oracle answers", libraryIdentity, buildcache.Tool("node", "--version"), []string{"testdata/library.mjs"}, cases), cases
	}, func(t *testing.T, input string) *exec.Cmd { return numbersCommand(t, "node", script, library, input) })
}

func numbersGoRawAnswersDirectory(t *testing.T, oracle, oracleIdentity string, corpus numbersCorpus) string {
	t.Helper()
	return numbersAnswerDirectory(t, &numbersGoRawAnswers, func() (numbersInputs, []numbersAnswerCase) {
		cases := numbersRawCases(corpus)
		return numbersAnswerProductInputs("Go raw oracle answers", oracleIdentity, runtime.Version(), nil, cases), cases
	}, func(t *testing.T, input string) *exec.Cmd { return numbersCommand(t, oracle, input) })
}

func numbersAnswer(t *testing.T, directory, file string) run {
	t.Helper()
	answer, err := os.ReadFile(filepath.Join(directory, file))
	if err != nil {
		t.Fatal(err)
	}
	return run{stdout: answer}
}

func numbersBatchGoAnswers(t *testing.T, oracle, oracleIdentity string, corpus numbersCorpus, unit numbersUnit) run {
	return numbersAnswer(t, numbersGoBatchAnswersDirectory(t, oracle, oracleIdentity, corpus), numbersBatchAnswerFile(unit))
}

func numbersBatchPrettierAnswers(t *testing.T, library, libraryIdentity string, corpus numbersCorpus, unit numbersUnit) run {
	return numbersAnswer(t, numbersPrettierBatchAnswersDirectory(t, library, libraryIdentity, corpus), numbersBatchAnswerFile(unit))
}

func numbersRawGoAnswer(t *testing.T, oracle, oracleIdentity string, corpus numbersCorpus, index int, preference string) run {
	return numbersAnswer(t, numbersGoRawAnswersDirectory(t, oracle, oracleIdentity, corpus), numbersRawAnswerFile(index, preference))
}

// What the answer products' keys need, read once per process: the corpus, and the oracles by content. Reading it
// builds nothing new, since the Go oracle is its own product.
var numbersAnswerInputs struct {
	once            sync.Once
	corpus          numbersCorpus
	oracle          string
	oracleIdentity  string
	library         string
	libraryIdentity string
}

func prepareNumbersAnswerInputs(t *testing.T) {
	t.Helper()
	numbersAnswerInputs.once.Do(func() {
		corpus := enumerateNumbers(t)
		if units := numbersUnits(corpus); len(units) != testCSSNumbersShards {
			t.Fatalf("enumerated %d shards, declared %d", len(units), testCSSNumbersShards)
		}
		oracle := numbersPreparedOracle(t)
		library := os.Getenv("ADAMIC_CSSNUMBERS_LIBRARY")
		if library != "" {
			numbersAnswerInputs.library, numbersAnswerInputs.libraryIdentity = library, numbersOracleIdentity(t, library)
		}
		numbersAnswerInputs.corpus = corpus
		numbersAnswerInputs.oracle, numbersAnswerInputs.oracleIdentity = oracle, numbersOracleIdentity(t, oracle)
	})
	if numbersAnswerInputs.oracleIdentity == "" {
		t.Fatal("answer inputs did not complete")
	}
}

// A one-byte edit to one batch's or one raw case's input re-keys the product holding its answer, by name key (what a
// runner finds it by) and by declared key (what this machine alone builds it by).
func TestCSSNumbersAnswerProductsRekey(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	corpus := enumerateNumbers(t)
	edited := numbersCorpus{texts: append([]string{}, corpus.texts...), raw: append([]string{}, corpus.raw...)}
	edited.texts[700] += "x"
	edited.raw[174] += "x"
	for _, product := range []struct {
		name  string
		files []string
		cases func(numbersCorpus) []numbersAnswerCase
	}{
		{"Go batch oracle answers", nil, numbersBatchCases},
		{"Prettier batch oracle answers", []string{"testdata/library.mjs"}, numbersBatchCases},
		{"Go raw oracle answers", nil, numbersRawCases},
	} {
		keys := func(corpus numbersCorpus) (string, string) {
			inputs := numbersCacheInputs(t, numbersAnswerProductInputs(product.name, "oracle-content-sha256=0", "toolchain", product.files, product.cases(corpus)))
			declared, err := buildcache.Key(root, inputs)
			if err != nil {
				t.Fatal(err)
			}
			return buildcache.NameKey(root, inputs), declared
		}
		nameKey, declared := keys(corpus)
		editedNameKey, editedDeclared := keys(edited)
		if nameKey == editedNameKey || declared == editedDeclared {
			t.Fatalf("%s: a one-byte input edit kept name key %s or key %s", product.name, nameKey[:12], declared[:12])
		}
		t.Logf("%s: name key %s to %s, key %s to %s", product.name, nameKey[:12], editedNameKey[:12], declared[:12], editedDeclared[:12])
	}
}

func TestProduct_CSSNumbersGoOracle(t *testing.T) { t.Parallel(); numbersPreparedOracle(t) }
func TestProduct_CSSNumbersGoAnswers(t *testing.T) {
	t.Parallel()
	corpus := enumerateNumbers(t)
	oracle := numbersPreparedOracle(t)
	clean(t, "Go", numbersSetupOracleAnswers(t, "Go", oracle, nil, numbersInput(corpus.texts), numbersOracleIdentity(t, oracle)))
}
func TestProduct_CSSNumbersGoBatchAnswers(t *testing.T) {
	t.Parallel()
	prepareNumbersAnswerInputs(t)
	numbersGoBatchAnswersDirectory(t, numbersAnswerInputs.oracle, numbersAnswerInputs.oracleIdentity, numbersAnswerInputs.corpus)
}

// Without the external library no unit reads Prettier answers, so there is nothing to build.
func TestProduct_CSSNumbersPrettierBatchAnswers(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_CSSNUMBERS_LIBRARY") == "" {
		t.Log("external library not checked: set ADAMIC_CSSNUMBERS_LIBRARY")
		return
	}
	prepareNumbersAnswerInputs(t)
	numbersPrettierBatchAnswersDirectory(t, numbersAnswerInputs.library, numbersAnswerInputs.libraryIdentity, numbersAnswerInputs.corpus)
}
func TestProduct_CSSNumbersGoRawAnswers(t *testing.T) {
	t.Parallel()
	prepareNumbersAnswerInputs(t)
	numbersGoRawAnswersDirectory(t, numbersAnswerInputs.oracle, numbersAnswerInputs.oracleIdentity, numbersAnswerInputs.corpus)
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
