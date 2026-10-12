package cssstrings

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

// The oracle answers the batch and raw units compare against are three products: every batch's Go answers, every
// batch's Prettier answers and every raw case's Go answers. Each answer is still its own oracle run on the input its
// unit ran the oracle on before, kept as one file per batch or raw case, so a unit compares the same bytes, and a
// build runs them GOMAXPROCS at a time. A unit and the product's TestProduct_ fetch it through the same function.
type stringsAnswerCase struct {
	file  string
	input []byte
}

// A process keys and looks up each answer product once: the key hashes every case's input, and every batch or raw
// unit reads the same product. A failure is kept, so every later caller reports the first one rather than a bare
// "did not complete".
type stringsAnswerProduct struct {
	once      sync.Once
	directory string
	err       error
}

var stringsGoBatchAnswers, stringsPrettierBatchAnswers, stringsGoRawAnswers stringsAnswerProduct

func stringsBatchAnswerFile(unit stringsUnit) string {
	return fmt.Sprintf("batch-%06d-%06d.txt", unit.lo, unit.hi-1)
}

func stringsRawAnswerFile(index int, preference string) string {
	return fmt.Sprintf("raw-%06d-%s.txt", index, preference)
}

func stringsBatchCases(corpus stringsCorpus) []stringsAnswerCase {
	var cases []stringsAnswerCase
	for _, unit := range stringsUnits(corpus) {
		if unit.kind == "batch" {
			cases = append(cases, stringsAnswerCase{stringsBatchAnswerFile(unit), stringsInput(corpus.texts[unit.lo:unit.hi])})
		}
	}
	return cases
}

func stringsRawCases(corpus stringsCorpus) []stringsAnswerCase {
	var cases []stringsAnswerCase
	for index, text := range corpus.raw {
		for _, preference := range []string{"d", "s"} {
			cases = append(cases, stringsAnswerCase{stringsRawAnswerFile(index, preference), []byte(preference + stringsEncode.Replace(text) + "\n")})
		}
	}
	return cases
}

// stringsAnswerProductInputs keys an answer product by its oracle, every case's file name and input, and this file,
// which holds the recipe.
func stringsAnswerProductInputs(name, identity, toolchain string, files []string, cases []stringsAnswerCase) stringsInputs {
	hash := sha256.New()
	for _, each := range cases {
		fmt.Fprintf(hash, "%s %d\n", each.file, len(each.input))
		hash.Write(each.input)
	}
	return stringsInputs{Name: name, Files: append([]string{"setup_products_test.go"}, files...), Flags: []string{identity, fmt.Sprintf("cases-sha256=%x", hash.Sum(nil))}, Toolchain: toolchain}
}

func stringsAnswerDirectory(t *testing.T, product *stringsAnswerProduct, prepare func() (stringsInputs, []stringsAnswerCase), command func(t *testing.T, input string) *exec.Cmd) string {
	t.Helper()
	product.once.Do(func() {
		// Until the lookup returns, a failure is the first caller stopping in prepare or keying, with its own log.
		product.err = fmt.Errorf("answer product: %s stopped before its lookup; its log has the failure", t.Name())
		inputs, cases := prepare()
		keyed := stringsCacheInputs(t, inputs)
		product.directory, product.err = buildcache.Get(keyed, func(dir string) error {
			pending := make(chan stringsAnswerCase)
			failures := make(chan error, len(cases))
			var workers sync.WaitGroup
			for range runtime.GOMAXPROCS(0) {
				workers.Add(1)
				go func() {
					defer workers.Done()
					for each := range pending {
						if err := stringsAnswerRun(t, dir, each, command); err != nil {
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
		if product.err != nil {
			product.err = fmt.Errorf("%s: %w", keyed.Name, product.err)
		}
	})
	if product.err != nil {
		t.Fatal(product.err)
	}
	return product.directory
}

// stringsAnswerRun runs the oracle on one case and keeps its stdout; the input file leaves with the run.
func stringsAnswerRun(t *testing.T, dir string, each stringsAnswerCase, oracleCommand func(t *testing.T, input string) *exec.Cmd) error {
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

func stringsGoBatchAnswersDirectory(t *testing.T, oracle, oracleIdentity string, corpus stringsCorpus) string {
	t.Helper()
	return stringsAnswerDirectory(t, &stringsGoBatchAnswers, func() (stringsInputs, []stringsAnswerCase) {
		cases := stringsBatchCases(corpus)
		return stringsAnswerProductInputs("Go batch oracle answers", oracleIdentity, runtime.Version(), nil, cases), cases
	}, func(t *testing.T, input string) *exec.Cmd { return stringsCommand(t, oracle, input) })
}

func stringsPrettierBatchAnswersDirectory(t *testing.T, library, libraryIdentity string, corpus stringsCorpus) string {
	t.Helper()
	script, err := filepath.Abs("testdata/library.mjs")
	if err != nil {
		t.Fatal(err)
	}
	return stringsAnswerDirectory(t, &stringsPrettierBatchAnswers, func() (stringsInputs, []stringsAnswerCase) {
		cases := stringsBatchCases(corpus)
		return stringsAnswerProductInputs("Prettier batch oracle answers", libraryIdentity, buildcache.Tool("node", "--version"), []string{"testdata/library.mjs"}, cases), cases
	}, func(t *testing.T, input string) *exec.Cmd { return stringsCommand(t, "node", script, library, input) })
}

func stringsGoRawAnswersDirectory(t *testing.T, oracle, oracleIdentity string, corpus stringsCorpus) string {
	t.Helper()
	return stringsAnswerDirectory(t, &stringsGoRawAnswers, func() (stringsInputs, []stringsAnswerCase) {
		cases := stringsRawCases(corpus)
		return stringsAnswerProductInputs("Go raw oracle answers", oracleIdentity, runtime.Version(), nil, cases), cases
	}, func(t *testing.T, input string) *exec.Cmd { return stringsCommand(t, oracle, input) })
}

func stringsAnswer(t *testing.T, directory, file string) run {
	t.Helper()
	answer, err := os.ReadFile(filepath.Join(directory, file))
	if err != nil {
		t.Fatal(err)
	}
	return run{stdout: answer}
}

func stringsBatchGoAnswers(t *testing.T, oracle, oracleIdentity string, corpus stringsCorpus, unit stringsUnit) run {
	return stringsAnswer(t, stringsGoBatchAnswersDirectory(t, oracle, oracleIdentity, corpus), stringsBatchAnswerFile(unit))
}

func stringsBatchPrettierAnswers(t *testing.T, library, libraryIdentity string, corpus stringsCorpus, unit stringsUnit) run {
	return stringsAnswer(t, stringsPrettierBatchAnswersDirectory(t, library, libraryIdentity, corpus), stringsBatchAnswerFile(unit))
}

func stringsRawGoAnswer(t *testing.T, oracle, oracleIdentity string, corpus stringsCorpus, index int, preference string) run {
	return stringsAnswer(t, stringsGoRawAnswersDirectory(t, oracle, oracleIdentity, corpus), stringsRawAnswerFile(index, preference))
}

// What the answer products' keys need, read once per process: the corpus, and the oracles by content. Reading it
// builds nothing new, since the Go oracle is its own product.
var stringsAnswerInputs struct {
	once            sync.Once
	corpus          stringsCorpus
	oracle          string
	oracleIdentity  string
	library         string
	libraryIdentity string
	err             error
}

func prepareStringsAnswerInputs(t *testing.T) {
	t.Helper()
	stringsAnswerInputs.once.Do(func() {
		// The helpers below stop the first caller with its own message; later callers name where to read it.
		stringsAnswerInputs.err = fmt.Errorf("answer inputs: %s stopped reading them; its log has the failure", t.Name())
		corpus := enumerateStrings(t)
		if units := stringsUnits(corpus); len(units) != testCSSStringsShards {
			t.Fatalf("enumerated %d shards, declared %d", len(units), testCSSStringsShards)
		}
		oracle := stringsPreparedOracle(t)
		library := os.Getenv("ADAMIC_CSSSTRINGS_LIBRARY")
		if library != "" {
			stringsAnswerInputs.library, stringsAnswerInputs.libraryIdentity = library, stringsOracleIdentity(t, library)
		}
		stringsAnswerInputs.corpus = corpus
		stringsAnswerInputs.oracle, stringsAnswerInputs.oracleIdentity = oracle, stringsOracleIdentity(t, oracle)
		stringsAnswerInputs.err = nil
	})
	if stringsAnswerInputs.err != nil {
		t.Fatal(stringsAnswerInputs.err)
	}
}

// A one-byte edit to one batch's or one raw case's input re-keys the product holding its answer, by name key (what a
// runner finds it by) and by declared key (what this machine alone builds it by), and so does another oracle, so a
// changed Go oracle or Prettier library never serves stale answers.
func TestCSSStringsAnswerProductsRekey(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	corpus := enumerateStrings(t)
	edited := stringsCorpus{texts: append([]string{}, corpus.texts...), raw: append([]string{}, corpus.raw...)}
	edited.texts[700] += "x"
	edited.raw[174] += "x"
	for _, product := range []struct {
		name  string
		files []string
		cases func(stringsCorpus) []stringsAnswerCase
	}{
		{"Go batch oracle answers", nil, stringsBatchCases},
		{"Prettier batch oracle answers", []string{"testdata/library.mjs"}, stringsBatchCases},
		{"Go raw oracle answers", nil, stringsRawCases},
	} {
		keys := func(corpus stringsCorpus, identity string) (string, string) {
			inputs := stringsCacheInputs(t, stringsAnswerProductInputs(product.name, identity, "toolchain", product.files, product.cases(corpus)))
			declared, err := buildcache.Key(root, inputs)
			if err != nil {
				t.Fatal(err)
			}
			return buildcache.NameKey(root, inputs), declared
		}
		nameKey, declared := keys(corpus, "oracle-content-sha256=0")
		editedNameKey, editedDeclared := keys(edited, "oracle-content-sha256=0")
		if nameKey == editedNameKey || declared == editedDeclared {
			t.Fatalf("%s: a one-byte input edit kept name key %s or key %s", product.name, nameKey[:12], declared[:12])
		}
		oracleNameKey, oracleDeclared := keys(corpus, "oracle-content-sha256=1")
		if nameKey == oracleNameKey || declared == oracleDeclared {
			t.Fatalf("%s: another oracle kept name key %s or key %s", product.name, nameKey[:12], declared[:12])
		}
		t.Logf("%s: name key %s to %s (input) and %s (oracle), key %s to %s and %s", product.name, nameKey[:12], editedNameKey[:12], oracleNameKey[:12], declared[:12], editedDeclared[:12], oracleDeclared[:12])
	}
}

// The Go oracle's key moves with the C or C++ compiler go builds it with, and with the C compiler's report, so a
// changed compiler misses rather than reading a product another one built.
func TestCSSStringsGoOracleRekeyCompiler(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	compilers := stringsCompilers{"clang", "clang++", "clang version 18.1.8"}
	key := buildcache.NameKey(root, stringsGoProductInputs("Go oracle", compilers))
	for _, other := range []stringsCompilers{{"gcc", "clang++", compilers.report}, {"clang", "g++", compilers.report}, {"clang", "clang++", "clang version 19.1.7"}} {
		if otherKey := buildcache.NameKey(root, stringsGoProductInputs("Go oracle", other)); otherKey == key {
			t.Fatalf("compilers %+v kept the Go oracle's key %s", other, key[:12])
		}
	}
	t.Logf("Go oracle key %s moves with CC, CXX and the C compiler's report", key[:12])
}

func TestProduct_CSSStringsGoOracle(t *testing.T) { t.Parallel(); stringsPreparedOracle(t) }
func TestProduct_CSSStringsGoAnswers(t *testing.T) {
	t.Parallel()
	corpus := enumerateStrings(t)
	oracle := stringsPreparedOracle(t)
	clean(t, "Go", stringsSetupOracleAnswers(t, "Go", oracle, nil, stringsInput(corpus.texts), stringsOracleIdentity(t, oracle)))
}
func TestProduct_CSSStringsGoBatchAnswers(t *testing.T) {
	t.Parallel()
	prepareStringsAnswerInputs(t)
	stringsGoBatchAnswersDirectory(t, stringsAnswerInputs.oracle, stringsAnswerInputs.oracleIdentity, stringsAnswerInputs.corpus)
}

// Without the external library no unit reads Prettier answers, so there is nothing to build.
func TestProduct_CSSStringsPrettierBatchAnswers(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_CSSSTRINGS_LIBRARY") == "" {
		t.Log("external library not checked: set ADAMIC_CSSSTRINGS_LIBRARY")
		return
	}
	prepareStringsAnswerInputs(t)
	stringsPrettierBatchAnswersDirectory(t, stringsAnswerInputs.library, stringsAnswerInputs.libraryIdentity, stringsAnswerInputs.corpus)
}
func TestProduct_CSSStringsGoRawAnswers(t *testing.T) {
	t.Parallel()
	prepareStringsAnswerInputs(t)
	stringsGoRawAnswersDirectory(t, stringsAnswerInputs.oracle, stringsAnswerInputs.oracleIdentity, stringsAnswerInputs.corpus)
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
