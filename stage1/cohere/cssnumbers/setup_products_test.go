package cssnumbers

import (
	"context"
	"fmt"
	"github.com/system-inc/adamic/stage1/cohere/cssnumbers/oracles"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/native"
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

func numbersPreparedProgram(test *testing.T, mutation int) numbersProgram {
	test.Helper()
	state := &numbersPreparedPrograms[mutation+1]
	state.once.Do(func() {
		main, absolutePathError := filepath.Abs("main.ts")
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		name := "port"
		if mutation >= 0 {
			change := numbersMutations[mutation]
			name = change.name
			main = prepareNumbersMutant(test, change)
		}
		state.program = buildNumbersProgram(test, name, main)
	})
	if state.program.dir == "" {
		test.Fatal("lowering product was not prepared")
	}
	return state.program
}

func numbersPreparedNative(test *testing.T, mutation int, sanitize bool) string {
	test.Helper()
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
	state.once.Do(
		func() { state.path = buildNumbersNative(test, name, numbersPreparedProgram(test, mutation), sanitize) },
	)
	if state.path == "" {
		test.Fatalf("native product %s was not prepared", name)
	}
	return state.path
}

// numbersPrepareUnitProducts fetches the products one unit reads and no other kind needs. A shard calls it before
// its clock starts, so shared setup holds only what every unit reads, and a mutant's builds wait for nobody else.
func numbersPrepareUnitProducts(test *testing.T, unit numbersUnit) {
	test.Helper()
	switch unit.kind {
	case "batch":
		numbersPreparedProgram(test, -1)
		numbersPreparedNative(test, -1, true)
		numbersPreparedNative(test, -1, false)
	case "raw":
		numbersPreparedNative(test, -1, true)
	case "mutant":
		numbersPreparedNative(test, unit.lo, true)
	case "throughput":
		if unit.side == "native" {
			numbersPreparedNative(test, -1, false)
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
// unit reads the same product. A failure is kept, so every later caller reports the first one rather than a bare
// "did not complete".
type numbersAnswerProduct struct {
	once      sync.Once
	directory string
	err       error
}

var numbersPrettierBatchAnswers numbersAnswerProduct

func numbersBatchAnswerFile(unit numbersUnit) string {
	return fmt.Sprintf("batch-%06d-%06d.txt", unit.lo, unit.hi-1)
}

func numbersRawAnswerFile(index int, preference string) string {
	return fmt.Sprintf("raw-%06d-%s.txt", index, preference)
}

func numbersBatchCases(corpus numbersCorpus) []numbersAnswerCase {
	var result []numbersAnswerCase
	for _, c := range oracles.BatchCases(oracles.Corpus{Texts: corpus.texts, Paths: corpus.paths, Raw: corpus.raw}) {
		result = append(result, numbersAnswerCase{c.File, c.Input})
	}
	return result
}

func numbersRawCases(corpus numbersCorpus) []numbersAnswerCase {
	var result []numbersAnswerCase
	for _, c := range oracles.RawCases(oracles.Corpus{Texts: corpus.texts, Paths: corpus.paths, Raw: corpus.raw}) {
		result = append(result, numbersAnswerCase{c.File, c.Input})
	}
	return result
}

// numbersAnswerProductInputs keys an answer product by its oracle, every case's file name and input, and this file,
// which holds the recipe.
func numbersAnswerProductInputs(
	name, identity, toolchain string,
	files []string,
	cases []numbersAnswerCase,
) numbersInputs {
	var shared []oracles.AnswerCase
	for _, c := range cases {
		shared = append(shared, oracles.AnswerCase{File: c.file, Input: c.input})
	}
	return oracles.AnswerProductInputs(name, identity, toolchain, files, shared)
}

func numbersAnswerDirectory(
	test *testing.T,
	product *numbersAnswerProduct,
	prepare func() (numbersInputs, []numbersAnswerCase),
	command func(test *testing.T, input string) (*exec.Cmd, context.Context),
) string {
	test.Helper()
	product.once.Do(func() {
		// Until the lookup returns, a failure is the first caller stopping in prepare or keying, with its own log.
		product.err = fmt.Errorf("answer product: %s stopped before its lookup; its log has the failure", test.Name())
		inputs, cases := prepare()
		keyed := numbersCacheInputs(test, inputs)
		product.directory, product.err = buildcache.Get(keyed, func(dir string) error {
			pending := make(chan numbersAnswerCase)
			failures := make(chan error, len(cases))
			var workers sync.WaitGroup
			for range runtime.GOMAXPROCS(0) {
				workers.Add(1)
				go func() {
					defer workers.Done()
					for each := range pending {
						if numbersAnswerRunError := numbersAnswerRun(
							test,
							dir,
							each,
							command,
						); numbersAnswerRunError != nil {
							failures <- numbersAnswerRunError
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

// numbersAnswerRun runs the oracle on one case and keeps its stdout; the input file leaves with the run. It reads the
// output through childguard.Capture for numbersExecute's reason: a WaitDelay could fail a case that had finished, and
// the case's 90 s deadline ends a hung run before childguard.Run's windows could.
func numbersAnswerRun(
	test *testing.T,
	dir string,
	each numbersAnswerCase,
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

func numbersGoBatchAnswersDirectory(test *testing.T, oracle, oracleIdentity string, corpus numbersCorpus) string {
	return oracles.BatchAnswers(
		test,
		oracle,
		oracleIdentity,
		oracles.BatchCases(oracles.Corpus{Texts: corpus.texts, Paths: corpus.paths, Raw: corpus.raw}),
	)
}

func numbersPrettierBatchAnswersDirectory(
	test *testing.T,
	library, libraryIdentity string,
	corpus numbersCorpus,
) string {
	test.Helper()
	script, absolutePathError := filepath.Abs("testdata/library.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return numbersAnswerDirectory(test, &numbersPrettierBatchAnswers, func() (numbersInputs, []numbersAnswerCase) {
		cases := numbersBatchCases(corpus)
		return numbersAnswerProductInputs(
			"Prettier batch oracle answers",
			libraryIdentity,
			buildcache.Tool("node", "--version"),
			[]string{"testdata/library.mjs"},
			cases,
		), cases
	}, func(test *testing.T, input string) (*exec.Cmd, context.Context) {
		return numbersCommand(test, "node", script, library, input)
	})
}

func numbersGoRawAnswersDirectory(test *testing.T, oracle, oracleIdentity string, corpus numbersCorpus) string {
	return oracles.RawAnswers(
		test,
		oracle,
		oracleIdentity,
		oracles.RawCases(oracles.Corpus{Texts: corpus.texts, Paths: corpus.paths, Raw: corpus.raw}),
	)
}

func numbersAnswer(test *testing.T, directory, file string) run {
	test.Helper()
	answer, readError := os.ReadFile(filepath.Join(directory, file))
	if readError != nil {
		test.Fatal(readError)
	}
	return run{stdout: answer}
}

func numbersBatchGoAnswers(test *testing.T, oracle, oracleIdentity string, corpus numbersCorpus, unit numbersUnit) run {
	return numbersAnswer(
		test,
		numbersGoBatchAnswersDirectory(test, oracle, oracleIdentity, corpus),
		numbersBatchAnswerFile(unit),
	)
}

func numbersBatchPrettierAnswers(
	test *testing.T,
	library, libraryIdentity string,
	corpus numbersCorpus,
	unit numbersUnit,
) run {
	return numbersAnswer(
		test,
		numbersPrettierBatchAnswersDirectory(test, library, libraryIdentity, corpus),
		numbersBatchAnswerFile(unit),
	)
}

func numbersRawGoAnswer(
	test *testing.T,
	oracle, oracleIdentity string,
	corpus numbersCorpus,
	index int,
	preference string,
) run {
	return numbersAnswer(
		test,
		numbersGoRawAnswersDirectory(test, oracle, oracleIdentity, corpus),
		numbersRawAnswerFile(index, preference),
	)
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
	err             error
}

func prepareNumbersAnswerInputs(test *testing.T) {
	test.Helper()
	numbersAnswerInputs.once.Do(func() {
		// The helpers below stop the first caller with its own message; later callers name where to read it.
		numbersAnswerInputs.err = fmt.Errorf(
			"answer inputs: %s stopped reading them; its log has the failure",
			test.Name(),
		)
		corpus := enumerateNumbers(test)
		if units := numbersUnits(corpus); len(units) != testCSSNumbersShards {
			test.Fatalf("enumerated %d shards, declared %d", len(units), testCSSNumbersShards)
		}
		oracle := numbersPreparedOracle(test)
		library := os.Getenv("ADAMIC_CSSNUMBERS_LIBRARY")
		if library != "" {
			numbersAnswerInputs.library, numbersAnswerInputs.libraryIdentity = library, numbersOracleIdentity(
				test,
				library,
			)
		}
		numbersAnswerInputs.corpus = corpus
		numbersAnswerInputs.oracle, numbersAnswerInputs.oracleIdentity = oracle, numbersOracleIdentity(test, oracle)
		numbersAnswerInputs.err = nil
	})
	if numbersAnswerInputs.err != nil {
		test.Fatal(numbersAnswerInputs.err)
	}
}

// A one-byte edit to one batch's or one raw case's input re-keys the product holding its answer, by name key (what a
// runner finds it by) and by declared key (what this machine alone builds it by), and so does another oracle, so a
// changed Go oracle or Prettier library never serves stale answers.
func TestCSSNumbersAnswerProductsRekey(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	corpus := enumerateNumbers(test)
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
		keys := func(corpus numbersCorpus, identity string) (string, string) {
			inputs := numbersCacheInputs(
				test,
				numbersAnswerProductInputs(product.name, identity, "toolchain", product.files, product.cases(corpus)),
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

// An edit to a package the compiler links must move the key of every product the compiler makes (the port's and each
// mutant's lowering, and every native binary built from one), and an edit to tsc's baselines, which nothing links,
// must keep it.
func TestCSSNumbersCompilerProductsKeyWhatTheCompilerLinks(test *testing.T) {
	test.Parallel()
	main, absolutePathError := filepath.Abs("main.ts")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	// A native product is keyed by its C's content and the toolchain's report; any C and any report stand in, since
	// only Files are checked.
	c := filepath.Join(test.TempDir(), "program.c")
	write(test, c, []byte("int main(void) { return 0; }\n"))
	products := []numbersInputs{
		numbersProgramInputs("port", main),
		numbersNativeInputs("sanitized", c, native.Options{Sanitize: true}, "toolchain"),
		numbersNativeInputs("native-fast", c, native.Options{Sanitize: false}, "toolchain"),
	}
	for _, mutation := range numbersMutations {
		products = append(
			products,
			numbersProgramInputs(mutation.name, prepareNumbersMutant(test, mutation)),
			numbersNativeInputs(mutation.name+"-sanitized", c, native.Options{Sanitize: true}, "toolchain"),
		)
	}
	covers := func(entry, path string) bool { return entry == path || strings.HasPrefix(path, entry+"/") }
	for _, product := range products {
		keyed := numbersCacheInputs(test, product)
		linked, baselines := false, false
		for _, entry := range keyed.Files {
			linked = linked || covers(entry, "cohere/TypeScript/tsc/internal/checker")
			baselines = baselines || covers(entry, "cohere/TypeScript/tsc/testdata")
		}
		if !linked {
			test.Errorf(
				"%s: no Files entry covers cohere/TypeScript/tsc/internal/checker, a package the compiler links",
				keyed.Name,
			)
		}
		if baselines {
			test.Errorf("%s: a Files entry covers cohere/TypeScript/tsc/testdata, baselines nothing links", keyed.Name)
		}
	}
	if !test.Failed() {
		test.Logf("%d compiler products key what the compiler links and no tsc baseline", len(products))
	}
}

func TestProduct_CSSNumbersFindsGoOracle(test *testing.T) {
	test.Parallel()
	numbersPreparedOracle(test)
}
func TestProduct_CSSNumbersFindsGoAnswers(test *testing.T) {
	test.Parallel()
	corpus := enumerateNumbers(test)
	oracle := numbersPreparedOracle(test)
	clean(test, "Go", numbersSetupOracleAnswers(
		test,
		"Go",
		oracle,
		nil,
		numbersInput(corpus.texts),
		numbersOracleIdentity(test, oracle),
	))
}
func TestProduct_CSSNumbersFindsGoBatchAnswers(test *testing.T) {
	test.Parallel()
	prepareNumbersAnswerInputs(test)
	numbersGoBatchAnswersDirectory(
		test,
		numbersAnswerInputs.oracle,
		numbersAnswerInputs.oracleIdentity,
		numbersAnswerInputs.corpus,
	)
}

// Without the external library no unit reads Prettier answers, so there is nothing to build.
func TestProduct_CSSNumbersPrettierBatchAnswers(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_CSSNUMBERS_LIBRARY") == "" {
		test.Log("external library not checked: set ADAMIC_CSSNUMBERS_LIBRARY")
		return
	}
	prepareNumbersAnswerInputs(test)
	numbersPrettierBatchAnswersDirectory(
		test,
		numbersAnswerInputs.library,
		numbersAnswerInputs.libraryIdentity,
		numbersAnswerInputs.corpus,
	)
}
func TestProduct_CSSNumbersFindsGoRawAnswers(test *testing.T) {
	test.Parallel()
	prepareNumbersAnswerInputs(test)
	numbersGoRawAnswersDirectory(
		test,
		numbersAnswerInputs.oracle,
		numbersAnswerInputs.oracleIdentity,
		numbersAnswerInputs.corpus,
	)
}
func TestProduct_CSSNumbersPortLowered(test *testing.T) {
	test.Parallel()
	numbersPreparedProgram(test, -1)
}
func TestProduct_CSSNumbersPortNative(test *testing.T) {
	test.Parallel()
	numbersPreparedNative(test, -1, true)
}
func TestProduct_CSSNumbersMutant0Lowered(test *testing.T) {
	test.Parallel()
	numbersPreparedProgram(test, 0)
}
func TestProduct_CSSNumbersMutant0Native(test *testing.T) {
	test.Parallel()
	numbersPreparedNative(test, 0, true)
}
func TestProduct_CSSNumbersMutant1Lowered(test *testing.T) {
	test.Parallel()
	numbersPreparedProgram(test, 1)
}
func TestProduct_CSSNumbersMutant1Native(test *testing.T) {
	test.Parallel()
	numbersPreparedNative(test, 1, true)
}
func TestProduct_CSSNumbersMutant2Lowered(test *testing.T) {
	test.Parallel()
	numbersPreparedProgram(test, 2)
}
func TestProduct_CSSNumbersMutant2Native(test *testing.T) {
	test.Parallel()
	numbersPreparedNative(test, 2, true)
}
func TestProduct_CSSNumbersFastNative(test *testing.T) {
	test.Parallel()
	numbersPreparedNative(test, -1, false)
}
