package css

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// A printer mutant's witnesses are every corpus case where it prints something other than Go, in each mode, found
// by running the mutant over the whole corpus and comparing it with Go's answers case by case. They're derived in
// code, not declared, and built as a product keyed by the mutant's lowered port, the corpus and the Go printer, so
// the full run is paid only when one of those moves. The mutant's shard then runs it over the smallest
// cssPrinterWitnessesRun of them (#tpjban7, Loom's ruling B on #testaudit).
var cssPrinterModes = []string{"default", "narrow"}

var cssPrinterWitnessOnce [3]sync.Once
var cssPrinterWitnessSets [3][][]int

// cssPrinterWitnessInputs keys a mutant's witnesses by the bytes they're derived from: the mutant's lowered port as
// Node runs it, the corpus, and the Go printer, plus this file, which holds the derivation.
func cssPrinterWitnessInputs(test *testing.T, variant int) buildcache.Inputs {
	test.Helper()
	corpus, _ := oracles.CachedParserOutputs(test)
	return cssPrinterWitnessKey(
		test,
		variant,
		cssPrinterProduct(test, variant).source,
		corpus,
		oracles.ComparisonOracle(test, true),
	)
}

func cssPrinterWitnessKey(test *testing.T, variant int, port, corpus, oracle string) buildcache.Inputs {
	test.Helper()
	flags := []string{"variant=" + strconv.Itoa(variant), "modes=" + strings.Join(cssPrinterModes, ",")}
	for _, input := range []struct{ name, path string }{{"port", port}, {"corpus", corpus}, {"oracle", oracle}} {
		content, readError := os.ReadFile(input.path)
		if readError != nil {
			test.Fatal(readError)
		}
		flags = append(flags, fmt.Sprintf("%s-sha256=%x", input.name, sha256.Sum256(content)))
	}
	return buildcache.Inputs{
		Name:      "css-printer-mutant-witnesses-" + strconv.Itoa(variant),
		Files:     []string{"stage1/cohere/css/printer_witnesses_test.go"},
		Flags:     flags,
		Toolchain: []string{buildcache.Tool("node", "--version"), runtime.GOOS, runtime.GOARCH},
	}
}

// cssPrinterWitnesses is the product: for each mode, the corpus case ids where the mutant differs from Go.
func cssPrinterWitnesses(test *testing.T, variant int) [][]int {
	test.Helper()
	cssPrinterWitnessOnce[variant].Do(func() {
		corpus, _ := oracles.CachedParserOutputs(test)
		port := cssPrinterProduct(test, variant).source
		directory := buildcache.Product(test, cssPrinterWitnessInputs(test, variant), func(dir string) error {
			witnesses, cssPrinterDeriveWitnessesError := cssPrinterDeriveWitnesses(test, port, corpus)
			if cssPrinterDeriveWitnessesError != nil {
				return cssPrinterDeriveWitnessesError
			}
			for mode, ids := range witnesses {
				var b strings.Builder
				for _, id := range ids {
					fmt.Fprintf(&b, "%d\n", id)
				}
				if writeError := os.WriteFile(
					filepath.Join(dir, cssPrinterModes[mode]+".txt"),
					[]byte(b.String()),
					0644,
				); writeError != nil {
					return writeError
				}
			}
			return nil
		})
		var sets [][]int
		for _, mode := range cssPrinterModes {
			data, readError := os.ReadFile(filepath.Join(directory, mode+".txt"))
			if readError != nil {
				test.Fatal(readError)
			}
			ids := []int{}
			for _, field := range strings.Fields(string(data)) {
				id, parseError := strconv.Atoi(field)
				if parseError != nil {
					test.Fatal(parseError)
				}
				ids = append(ids, id)
			}
			sets = append(sets, ids)
		}
		cssPrinterWitnessSets[variant] = sets
	})
	if cssPrinterWitnessSets[variant] == nil {
		test.Fatalf(
			"printer mutant %d's witnesses unavailable: their build failed earlier in this process (see that test's log)",
			variant,
		)
	}
	return cssPrinterWitnessSets[variant]
}

// cssPrinterDeriveWitnesses runs a port over the whole corpus on Node in each mode and returns, per mode, the case ids
// whose answer differs from Go's. A port that fails to run, or prints a different set of cases, is an error.
func cssPrinterDeriveWitnesses(test *testing.T, port, corpus string) ([][]int, error) {
	test.Helper()
	oracle := oracles.ComparisonOracle(test, true)
	var witnesses [][]int
	for _, mode := range cssPrinterModes {
		want := cssPrinterAnswers(test, oracle, corpus, mode)
		got := onNode(test, port, corpus, "output", "once", mode)
		if got.exitCode != 0 || len(got.stderr) != 0 {
			return nil, fmt.Errorf(
				"%s: the port must run over the whole corpus: exit %d %s",
				mode,
				got.exitCode,
				got.stderr,
			)
		}
		ids, cssPrinterDifferingCasesError := cssPrinterDifferingCases(string(got.stdout), want)
		if cssPrinterDifferingCasesError != nil {
			return nil, fmt.Errorf("%s: %w", mode, cssPrinterDifferingCasesError)
		}
		witnesses = append(witnesses, ids)
	}
	return witnesses, nil
}

// cssPrinterDifferingCases reads two answer files, each a "case N" line and an answer line per case, and returns the
// local case numbers whose answers differ. Files that don't name the same cases in the same order are an error.
func cssPrinterDifferingCases(got, want string) ([]int, error) {
	gotLines := strings.Split(strings.TrimSuffix(got, "\n"), "\n")
	wantLines := strings.Split(strings.TrimSuffix(want, "\n"), "\n")
	if len(gotLines) != len(wantLines) || len(wantLines)%2 != 0 {
		return nil, fmt.Errorf("%d answer lines, Go has %d", len(gotLines), len(wantLines))
	}
	ids := []int{}
	for i := 0; i < len(wantLines); i += 2 {
		if gotLines[i] != wantLines[i] || wantLines[i] != fmt.Sprintf("case %d", i/2) {
			return nil, fmt.Errorf("case line %d reads %q, Go has %q", i, gotLines[i], wantLines[i])
		}
		if gotLines[i+1] != wantLines[i+1] {
			ids = append(ids, i/2)
		}
	}
	return ids, nil
}

// cssPrinterWitnessError is why a mutant's witnesses can't prove anything: a mutant that differs from Go nowhere
// in a mode survives the whole corpus there, and its shard fails rather than passing on nothing.
func cssPrinterWitnessError(name, mode string, ids []int) error {
	if len(ids) == 0 {
		return fmt.Errorf(
			"printer mutant %s differs from Go nowhere in %s mode: it survives the whole corpus, so its shard proves nothing",
			name,
			mode,
		)
	}
	return nil
}

func TestProduct_CSSPrinterSemicolonMutantWitnesses(test *testing.T) {
	test.Parallel()
	cssPrinterWitnesses(test, 0)
}
func TestProduct_CSSPrinterIndentMutantWitnesses(test *testing.T) {
	test.Parallel()
	cssPrinterWitnesses(test, 1)
}
func TestProduct_CSSPrinterWidthMutantWitnesses(test *testing.T) {
	test.Parallel()
	cssPrinterWitnesses(test, 2)
}

// The witnesses' key moves with each input they're derived from: the mutant's port, the corpus and the Go printer.
func TestCSSPrinterMutantWitnessesRekey(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	corpus, _ := oracles.CachedParserOutputs(test)
	inputs := map[string]string{
		"port":   cssPrinterProduct(test, 0).source,
		"corpus": corpus,
		"oracle": oracles.ComparisonOracle(test, true),
	}
	key := buildcache.NameKey(root, cssPrinterWitnessKey(test, 0, inputs["port"], inputs["corpus"], inputs["oracle"]))
	if other := buildcache.NameKey(
		root,
		cssPrinterWitnessKey(test, 1, inputs["port"], inputs["corpus"], inputs["oracle"]),
	); other == key {
		test.Fatal("another mutant kept the witnesses' key")
	}
	for _, name := range []string{"port", "corpus", "oracle"} {
		content, readError := os.ReadFile(inputs[name])
		if readError != nil {
			test.Fatal(readError)
		}
		edited := filepath.Join(test.TempDir(), name)
		if writeError := os.WriteFile(edited, append(content, 'x'), 0644); writeError != nil {
			test.Fatal(writeError)
		}
		moved := map[string]string{"port": inputs["port"], "corpus": inputs["corpus"], "oracle": inputs["oracle"]}
		moved[name] = edited
		if buildcache.NameKey(
			root,
			cssPrinterWitnessKey(test, 0, moved["port"], moved["corpus"], moved["oracle"]),
		) == key {
			test.Fatalf("a one-byte edit to the %s kept the witnesses' key", name)
		}
	}
	test.Logf("witnesses key %s moves with the mutant, its port, the corpus and the Go printer", key[:12])
}

// A mutant planted so it differs from Go nowhere turns its shard red. The unmutated printer stands in for it: run
// through the same derivation over the whole corpus it has no witnesses in either mode, and the shard's check refuses
// that. The derivation itself still finds a difference planted in one answer.
func TestCSSPrinterMutantThatDiffersNowhereFails(test *testing.T) {
	test.Parallel()
	corpus, _ := oracles.CachedParserOutputs(test)
	witnesses, cssPrinterDeriveWitnessesError := cssPrinterDeriveWitnesses(
		test,
		cssPrinterProduct(test, -1).source,
		corpus,
	)
	if cssPrinterDeriveWitnessesError != nil {
		test.Fatal(cssPrinterDeriveWitnessesError)
	}
	for mode, ids := range witnesses {
		if err := cssPrinterWitnessError("that differs nowhere", cssPrinterModes[mode], ids); err == nil {
			test.Fatalf(
				"%s: the unmutated printer has %d witnesses, and its shard would pass",
				cssPrinterModes[mode],
				len(ids),
			)
		} else {
			test.Logf("refused: %v", err)
		}
	}
	ids, cssPrinterDifferingCasesError := cssPrinterDifferingCases(
		"case 0\n\"a\"\ncase 1\n\"c\"\n",
		"case 0\n\"a\"\ncase 1\n\"b\"\n",
	)
	if cssPrinterDifferingCasesError != nil || len(ids) != 1 || ids[0] != 1 {
		test.Fatalf("one changed answer read as %v, %v", ids, cssPrinterDifferingCasesError)
	}
	if _, cssPrinterDifferingCasesError := cssPrinterDifferingCases(
		"case 0\n\"a\"\n",
		"case 0\n\"a\"\ncase 1\n\"b\"\n",
	); cssPrinterDifferingCasesError == nil {
		test.Fatal("a missing case read as no difference")
	}
}

// The rule that picks a shard's witnesses: the smallest by bytes, ties by case id, at most cssPrinterWitnessesRun,
// back in corpus order, and nothing outside the corpus.
func TestCSSPrinterChosenWitnessesAreTheSmallest(test *testing.T) {
	test.Parallel()
	var lines []string
	var ids []int
	for id := range cssPrinterWitnessesRun + 3 {
		lines = append(lines, strings.Repeat("x", 100-id))
		ids = append(ids, id)
	}
	lines[5], lines[9] = "a", "b"
	chosen := cssPrinterChosenWitnesses(lines, append(ids, -1, len(lines)))
	if len(chosen) != cssPrinterWitnessesRun || !slices.IsSorted(chosen) {
		test.Fatalf("chose %d witnesses %v, want %d in corpus order", len(chosen), chosen, cssPrinterWitnessesRun)
	}
	for _, dropped := range []int{0, 1, 2} {
		if slices.Contains(chosen, dropped) {
			test.Fatalf("kept witness %d, one of the three largest", dropped)
		}
	}
	if !slices.Contains(chosen, 5) || !slices.Contains(chosen, 9) {
		test.Fatalf("dropped a one-byte witness: %v", chosen)
	}
	tied := cssPrinterChosenWitnesses(
		slices.Repeat([]string{"x"}, cssPrinterWitnessesRun+1),
		ids[:cssPrinterWitnessesRun+1],
	)
	if tied[len(tied)-1] != cssPrinterWitnessesRun-1 {
		test.Fatalf("equal sizes broke ties as %v, want the lowest ids", tied)
	}
}
