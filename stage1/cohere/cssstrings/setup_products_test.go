package cssstrings

import (
	"context"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/cssstrings/oracles"
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

func stringsPreparedProgram(test *testing.T, mutation int) stringsProgram {
	test.Helper()
	state := &stringsPreparedPrograms[mutation+1]
	state.once.Do(func() {
		main, absolutePathError := filepath.Abs("main.ts")
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		name := "port"
		if mutation >= 0 {
			change := stringsMutations[mutation]
			name = change.name
			main = prepareStringsMutant(test, change)
		}
		state.program = buildStringsProgram(test, name, main)
	})
	return state.program
}

func stringsPreparedNative(test *testing.T, mutation int, sanitize bool) string {
	test.Helper()
	name := "native-fast"
	if sanitize {
		name = "sanitized"
	}
	if mutation >= 0 {
		name = stringsMutations[mutation].name + "-sanitized"
	}
	return buildStringsNative(test, name, stringsPreparedProgram(test, mutation), sanitize)
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

var stringsPrettierBatchAnswers stringsAnswerProduct

func stringsBatchAnswerFile(unit stringsUnit) string {
	return fmt.Sprintf("batch-%06d-%06d.txt", unit.lo, unit.hi-1)
}

func stringsRawAnswerFile(index int, preference string) string {
	return fmt.Sprintf("raw-%06d-%s.txt", index, preference)
}

func stringsBatchCases(corpus stringsCorpus) []stringsAnswerCase {
	var result []stringsAnswerCase
	for _, answerCase := range oracles.BatchCases(
		oracles.Corpus{Texts: corpus.texts, Paths: corpus.paths, Raw: corpus.raw},
	) {
		result = append(result, stringsAnswerCase{answerCase.File, answerCase.Input})
	}
	return result
}

func stringsRawCases(corpus stringsCorpus) []stringsAnswerCase {
	var result []stringsAnswerCase
	for _, answerCase := range oracles.RawCases(
		oracles.Corpus{Texts: corpus.texts, Paths: corpus.paths, Raw: corpus.raw},
	) {
		result = append(result, stringsAnswerCase{answerCase.File, answerCase.Input})
	}
	return result
}

// stringsAnswerProductInputs keys an answer product by its oracle, every case's file name and input, and this file,
// which holds the recipe.
func stringsAnswerProductInputs(
	name, identity, toolchain string,
	files []string,
	cases []stringsAnswerCase,
) stringsInputs {
	var shared []oracles.AnswerCase
	for _, answerCase := range cases {
		shared = append(shared, oracles.AnswerCase{File: answerCase.file, Input: answerCase.input})
	}
	return oracles.AnswerProductInputs(name, identity, toolchain, files, shared)
}

func stringsAnswerDirectory(
	test *testing.T,
	product *stringsAnswerProduct,
	prepare func() (stringsInputs, []stringsAnswerCase),
	command func(test *testing.T, input string) (*exec.Cmd, context.Context),
) string {
	test.Helper()
	product.once.Do(func() {
		// Until the lookup returns, a failure is the first caller stopping in prepare or keying, with its own log.
		product.err = fmt.Errorf("answer product: %s stopped before its lookup; its log has the failure", test.Name())
		inputs, cases := prepare()
		keyed := stringsCacheInputs(test, inputs)
		product.directory, product.err = buildcache.Get(keyed, func(dir string) error {
			pending := make(chan stringsAnswerCase)
			failures := make(chan error, len(cases))
			var workers sync.WaitGroup
			for range runtime.GOMAXPROCS(0) {
				workers.Add(1)
				go func() {
					defer workers.Done()
					for each := range pending {
						if stringsAnswerRunError := stringsAnswerRun(
							test,
							dir,
							each,
							command,
						); stringsAnswerRunError != nil {
							failures <- stringsAnswerRunError
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
		test.Fatal(product.err)
	}
	return product.directory
}

// stringsAnswerRun runs the oracle on one case and keeps its stdout; the input file leaves with the run. It reads the
// output through childguard.Capture for stringsExecute's reason: a WaitDelay could fail a case that had finished, and
// the case's 90 s deadline ends a hung run before childguard.Run's windows could.
func stringsAnswerRun(
	test *testing.T,
	dir string,
	each stringsAnswerCase,
	oracleCommand func(test *testing.T, input string) (*exec.Cmd, context.Context),
) error {
	input := filepath.Join(dir, each.file+".input")
	if writeError := os.WriteFile(input, each.input, 0644); writeError != nil {
		return writeError
	}
	defer os.Remove(input)
	command, ctx := oracleCommand(test, input)
	stdout, stderr, captureError := childguard.Capture(ctx, command)
	if captureError != nil || len(stderr) != 0 {
		return fmt.Errorf("oracle on %s: %v; %s", each.file, captureError, stderr)
	}
	return os.WriteFile(filepath.Join(dir, each.file), stdout, 0644)
}

func stringsGoBatchAnswersDirectory(test *testing.T, oracle, oracleIdentity string, corpus stringsCorpus) string {
	return oracles.BatchAnswers(
		test,
		oracle,
		oracleIdentity,
		oracles.BatchCases(oracles.Corpus{Texts: corpus.texts, Paths: corpus.paths, Raw: corpus.raw}),
	)
}

func stringsPrettierBatchAnswersDirectory(
	test *testing.T,
	library, libraryIdentity string,
	corpus stringsCorpus,
) string {
	test.Helper()
	script, absolutePathError := filepath.Abs("testdata/library.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return stringsAnswerDirectory(test, &stringsPrettierBatchAnswers, func() (stringsInputs, []stringsAnswerCase) {
		cases := stringsBatchCases(corpus)
		return stringsAnswerProductInputs(
			"Prettier batch oracle answers",
			libraryIdentity,
			buildcache.Tool("node", "--version"),
			[]string{"testdata/library.mjs"},
			cases,
		), cases
	}, func(test *testing.T, input string) (*exec.Cmd, context.Context) {
		return stringsCommand(test, "node", script, library, input)
	})
}

func stringsGoRawAnswersDirectory(test *testing.T, oracle, oracleIdentity string, corpus stringsCorpus) string {
	return oracles.RawAnswers(
		test,
		oracle,
		oracleIdentity,
		oracles.RawCases(oracles.Corpus{Texts: corpus.texts, Paths: corpus.paths, Raw: corpus.raw}),
	)
}

func stringsAnswer(test *testing.T, directory, file string) run {
	test.Helper()
	answer, readError := os.ReadFile(filepath.Join(directory, file))
	if readError != nil {
		test.Fatal(readError)
	}
	return run{stdout: answer}
}

func stringsBatchGoAnswers(test *testing.T, oracle, oracleIdentity string, corpus stringsCorpus, unit stringsUnit) run {
	return stringsAnswer(
		test,
		stringsGoBatchAnswersDirectory(test, oracle, oracleIdentity, corpus),
		stringsBatchAnswerFile(unit),
	)
}

func stringsBatchPrettierAnswers(
	test *testing.T,
	library, libraryIdentity string,
	corpus stringsCorpus,
	unit stringsUnit,
) run {
	return stringsAnswer(
		test,
		stringsPrettierBatchAnswersDirectory(test, library, libraryIdentity, corpus),
		stringsBatchAnswerFile(unit),
	)
}

func stringsRawGoAnswer(
	test *testing.T,
	oracle, oracleIdentity string,
	corpus stringsCorpus,
	index int,
	preference string,
) run {
	return stringsAnswer(
		test,
		stringsGoRawAnswersDirectory(test, oracle, oracleIdentity, corpus),
		stringsRawAnswerFile(index, preference),
	)
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

func prepareStringsAnswerInputs(test *testing.T) {
	test.Helper()
	stringsAnswerInputs.once.Do(func() {
		// The helpers below stop the first caller with its own message; later callers name where to read it.
		stringsAnswerInputs.err = fmt.Errorf(
			"answer inputs: %s stopped reading them; its log has the failure",
			test.Name(),
		)
		corpus := enumerateStrings(test)
		if units := stringsUnits(corpus); len(units) != testCssStringsShards {
			test.Fatalf("enumerated %d shards, declared %d", len(units), testCssStringsShards)
		}
		oracle := stringsPreparedOracle(test)
		library := os.Getenv("ADAMIC_CSSSTRINGS_LIBRARY")
		if library != "" {
			stringsAnswerInputs.library, stringsAnswerInputs.libraryIdentity = library, stringsOracleIdentity(
				test,
				library,
			)
		}
		stringsAnswerInputs.corpus = corpus
		stringsAnswerInputs.oracle, stringsAnswerInputs.oracleIdentity = oracle, stringsOracleIdentity(test, oracle)
		stringsAnswerInputs.err = nil
	})
	if stringsAnswerInputs.err != nil {
		test.Fatal(stringsAnswerInputs.err)
	}
}

// A one-byte edit to one batch's or one raw case's input re-keys the product holding its answer, by name key (what a
// runner finds it by) and by declared key (what this machine alone builds it by), and so does another oracle, so a
// changed Go oracle or Prettier library never serves stale answers.
func TestCssStringsAnswerProductsRekey(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	corpus := enumerateStrings(test)
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
			inputs := stringsCacheInputs(
				test,
				stringsAnswerProductInputs(product.name, identity, "toolchain", product.files, product.cases(corpus)),
			)
			declared, keyError := buildcache.Key(root, inputs)
			if keyError != nil {
				test.Fatal(keyError)
			}
			return buildcache.NameKey(root, inputs), declared
		}
		nameKey, declared := keys(corpus, "oracle-content-sha256=0")
		editedNameKey, editedDeclared := keys(edited, "oracle-content-sha256=0")
		if nameKey == editedNameKey || declared == editedDeclared {
			test.Fatalf(
				"%s: a one-byte input edit kept name key %s or key %s",
				product.name,
				nameKey[:12],
				declared[:12],
			)
		}
		oracleNameKey, oracleDeclared := keys(corpus, "oracle-content-sha256=1")
		if nameKey == oracleNameKey || declared == oracleDeclared {
			test.Fatalf("%s: another oracle kept name key %s or key %s", product.name, nameKey[:12], declared[:12])
		}
		test.Logf(
			"%s: name key %s to %s (input) and %s (oracle), key %s to %s and %s",
			product.name,
			nameKey[:12],
			editedNameKey[:12],
			oracleNameKey[:12],
			declared[:12],
			editedDeclared[:12],
			oracleDeclared[:12],
		)
	}
}

func TestProduct_CssStringsFindsGoOracle(test *testing.T) {
	test.Parallel()
	stringsPreparedOracle(test)
}
func TestProduct_CssStringsFindsGoAnswers(test *testing.T) {
	test.Parallel()
	corpus := enumerateStrings(test)
	oracle := stringsPreparedOracle(test)
	clean(test, "Go", stringsSetupOracleAnswers(
		test,
		"Go",
		oracle,
		nil,
		stringsInput(corpus.texts),
		stringsOracleIdentity(test, oracle),
	))
}
func TestProduct_CssStringsFindsGoBatchAnswers(test *testing.T) {
	test.Parallel()
	prepareStringsAnswerInputs(test)
	stringsGoBatchAnswersDirectory(
		test,
		stringsAnswerInputs.oracle,
		stringsAnswerInputs.oracleIdentity,
		stringsAnswerInputs.corpus,
	)
}

// Without the external library no unit reads Prettier answers, so there is nothing to build.
func TestProduct_CssStringsPrettierBatchAnswers(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_CSSSTRINGS_LIBRARY") == "" {
		test.Log("external library not checked: set ADAMIC_CSSSTRINGS_LIBRARY")
		return
	}
	prepareStringsAnswerInputs(test)
	stringsPrettierBatchAnswersDirectory(
		test,
		stringsAnswerInputs.library,
		stringsAnswerInputs.libraryIdentity,
		stringsAnswerInputs.corpus,
	)
}
func TestProduct_CssStringsFindsGoRawAnswers(test *testing.T) {
	test.Parallel()
	prepareStringsAnswerInputs(test)
	stringsGoRawAnswersDirectory(
		test,
		stringsAnswerInputs.oracle,
		stringsAnswerInputs.oracleIdentity,
		stringsAnswerInputs.corpus,
	)
}
func TestProduct_CssStringsPortLowered(test *testing.T) {
	test.Parallel()
	stringsPreparedProgram(test, -1)
}
func TestProduct_CssStringsPortNative(test *testing.T) {
	test.Parallel()
	stringsPreparedNative(test, -1, true)
}
func TestProduct_CssStringsMutant0Lowered(test *testing.T) {
	test.Parallel()
	stringsPreparedProgram(test, 0)
}
func TestProduct_CssStringsMutant0Native(test *testing.T) {
	test.Parallel()
	stringsPreparedNative(test, 0, true)
}
func TestProduct_CssStringsMutant1Lowered(test *testing.T) {
	test.Parallel()
	stringsPreparedProgram(test, 1)
}
func TestProduct_CssStringsMutant1Native(test *testing.T) {
	test.Parallel()
	stringsPreparedNative(test, 1, true)
}
func TestProduct_CssStringsMutant2Lowered(test *testing.T) {
	test.Parallel()
	stringsPreparedProgram(test, 2)
}
func TestProduct_CssStringsMutant2Native(test *testing.T) {
	test.Parallel()
	stringsPreparedNative(test, 2, true)
}
func TestProduct_CssStringsFastNative(test *testing.T) {
	test.Parallel()
	stringsPreparedNative(test, -1, false)
}
