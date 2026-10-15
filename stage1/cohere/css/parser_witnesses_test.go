package css

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/stage1/cohere/css/oracles"
)

// A parser mutant's witnesses are every corpus case where it answers other than Go, found by running the mutant over
// the whole corpus on Node and comparing it with Go's answers case by case. They're derived in code, not declared, and
// built as a product keyed by the mutant's lowered port, the corpus and Go's answers, so the full run is paid only when
// one of those moves. The mutant's group then runs it over exactly its witnesses (#tpjban7, Loom's ruling on
// #testaudit). On Oct 14 the three mutants had 268, 238 and 1,770 witnesses of 24,016 cases.
var cssParserWitnessOnce [3]sync.Once
var cssParserWitnessSets [3][]int

// cssParserWitnessKey keys a mutant's witnesses by the bytes they're derived from: the mutant's port as Node runs it
// and the corpus with Go's answers, plus this file, which holds the derivation.
func cssParserWitnessKey(t *testing.T, mutation int, port, cases, answers string) buildcache.Inputs {
	t.Helper()
	flags := []string{"mutation=" + strconv.Itoa(mutation)}
	for _, input := range []struct{ name, path string }{{"port", port}, {"cases", cases}, {"answers", answers}} {
		content, err := os.ReadFile(input.path)
		if err != nil {
			t.Fatal(err)
		}
		flags = append(flags, fmt.Sprintf("%s-sha256=%x", input.name, sha256.Sum256(content)))
	}
	return buildcache.Inputs{
		Name:      "css-parser-mutant-witnesses-" + strconv.Itoa(mutation),
		Files:     []string{"stage1/cohere/css/parser_witnesses_test.go"},
		Flags:     flags,
		Toolchain: []string{buildcache.Tool("node", "--version"), runtime.GOOS, runtime.GOARCH},
	}
}

func cssParserWitnessInputs(t *testing.T, mutation int) buildcache.Inputs {
	t.Helper()
	cases, _ := oracles.CachedParserOutputs(t)
	return cssParserWitnessKey(t, mutation, cssParserPreparedProgram(t, mutation).main, cases, filepath.Join(filepath.Dir(cases), "answers.txt"))
}

// cssParserWitnesses is the product: the corpus case indices where the mutant differs from Go.
func cssParserWitnesses(t *testing.T, mutation int) []int {
	t.Helper()
	cssParserWitnessOnce[mutation].Do(func() {
		port := cssParserPreparedProgram(t, mutation).main
		directory := buildcache.Product(t, cssParserWitnessInputs(t, mutation), func(dir string) error {
			ids, err := cssParserDeriveWitnesses(t, port)
			if err != nil {
				return err
			}
			var b strings.Builder
			for _, id := range ids {
				fmt.Fprintf(&b, "%d\n", id)
			}
			return os.WriteFile(filepath.Join(dir, "witnesses.txt"), []byte(b.String()), 0644)
		})
		data, err := os.ReadFile(filepath.Join(directory, "witnesses.txt"))
		if err != nil {
			t.Fatal(err)
		}
		ids := []int{}
		for _, field := range strings.Fields(string(data)) {
			id, err := strconv.Atoi(field)
			if err != nil {
				t.Fatal(err)
			}
			ids = append(ids, id)
		}
		cssParserWitnessSets[mutation] = ids
	})
	if cssParserWitnessSets[mutation] == nil {
		t.Fatalf("parser mutant %d's witnesses unavailable: their build failed earlier in this process (see that test's log)", mutation)
	}
	return cssParserWitnessSets[mutation]
}

// cssParserAllWitnesses reads every mutant's witness product.
func cssParserAllWitnesses(t *testing.T) [][]int {
	t.Helper()
	var witnesses [][]int
	for mutation := range mutants {
		witnesses = append(witnesses, cssParserWitnesses(t, mutation))
	}
	return witnesses
}

// cssParserDeriveWitnesses runs a port over the whole corpus on Node and returns the case indices whose answer differs
// from Go's. A port that fails to run, or answers a different set of cases, is an error.
func cssParserDeriveWitnesses(t *testing.T, port string) ([]int, error) {
	t.Helper()
	corpus := oracles.LoadParserCorpus(t)
	all := make([]int, len(corpus.Inputs))
	for i := range all {
		all[i] = i
	}
	cases, want := oracles.WriteParserSlice(t, corpus, all)
	got := cssParserOnNode(t, port, cases)
	if got.exitCode != 0 || len(got.stderr) != 0 {
		return nil, fmt.Errorf("the port must run over the whole corpus: exit %d %s", got.exitCode, got.stderr)
	}
	return cssParserDifferingCases(string(got.stdout), want)
}

// cssParserDifferingCases reads two answer files, each a "case N" line and an answer line per case, and returns the
// local case numbers whose answers differ. Files that don't name the same cases in the same order are an error.
func cssParserDifferingCases(got, want string) ([]int, error) {
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

// cssParserWitnessError is why a mutant's witnesses can't prove anything: a mutant that differs from Go nowhere
// survives the whole corpus, and the union and its group fail rather than pass on nothing.
func cssParserWitnessError(name string, ids []int) error {
	if len(ids) == 0 {
		return fmt.Errorf("parser mutant %s differs from Go nowhere: it survives the whole corpus, so its group proves nothing", name)
	}
	return nil
}

func TestProduct_CSSParserCustomPropertyMutantWitnesses(t *testing.T) {
	t.Parallel()
	cssParserWitnesses(t, 0)
}
func TestProduct_CSSParserCommentMutantWitnesses(t *testing.T) {
	t.Parallel()
	cssParserWitnesses(t, 1)
}
func TestProduct_CSSParserClosingBraceMutantWitnesses(t *testing.T) {
	t.Parallel()
	cssParserWitnesses(t, 2)
}

// The witnesses' key moves with each input they're derived from: the mutant's port, the corpus and Go's answers.
func TestCSSParserMutantWitnessesRekey(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	cases, _ := oracles.CachedParserOutputs(t)
	inputs := map[string]string{"port": cssParserPreparedProgram(t, 0).main, "cases": cases, "answers": filepath.Join(filepath.Dir(cases), "answers.txt")}
	key := buildcache.NameKey(root, cssParserWitnessKey(t, 0, inputs["port"], inputs["cases"], inputs["answers"]))
	if buildcache.NameKey(root, cssParserWitnessKey(t, 1, inputs["port"], inputs["cases"], inputs["answers"])) == key {
		t.Fatal("another mutant kept the witnesses' key")
	}
	for _, name := range []string{"port", "cases", "answers"} {
		content, err := os.ReadFile(inputs[name])
		if err != nil {
			t.Fatal(err)
		}
		edited := filepath.Join(t.TempDir(), name)
		if err := os.WriteFile(edited, append(content, 'x'), 0644); err != nil {
			t.Fatal(err)
		}
		moved := map[string]string{"port": inputs["port"], "cases": inputs["cases"], "answers": inputs["answers"]}
		moved[name] = edited
		if buildcache.NameKey(root, cssParserWitnessKey(t, 0, moved["port"], moved["cases"], moved["answers"])) == key {
			t.Fatalf("a one-byte edit to the %s kept the witnesses' key", name)
		}
	}
	t.Logf("witnesses key %s moves with the mutant, its port, the corpus and Go's answers", key[:12])
}

// A mutant planted so it differs from Go nowhere turns the test red. The unmutated parser stands in for it: run
// through the same derivation over the whole corpus it has no witnesses, and the union refuses a plan built on that.
// The derivation still finds a difference planted in one answer.
func TestCSSParserMutantThatDiffersNowhereFails(t *testing.T) {
	t.Parallel()
	ids, err := cssParserDeriveWitnesses(t, cssParserPreparedProgram(t, -1).main)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 0 {
		t.Fatalf("the unmutated parser differs from Go on %d cases", len(ids))
	}
	keys := oracles.LoadParserCorpus(t).CaseKeys
	witnesses := cssParserAllWitnesses(t)
	witnesses[0] = ids
	err = cssParserUnionError(keys, oracles.ParserWithMutants(oracles.ParserAgreementShards(keys), keys, witnesses), witnesses)
	if err == nil || !strings.Contains(err.Error(), "differs from Go nowhere") {
		t.Fatalf("a mutant with no witnesses read as %v", err)
	}
	t.Logf("refused: %v", err)
	differing, err := cssParserDifferingCases("case 0\na\ncase 1\nc\n", "case 0\na\ncase 1\nb\n")
	if err != nil || len(differing) != 1 || differing[0] != 1 {
		t.Fatalf("one changed answer read as %v, %v", differing, err)
	}
	if _, err := cssParserDifferingCases("case 0\na\n", "case 0\na\ncase 1\nb\n"); err == nil {
		t.Fatal("a missing case read as no difference")
	}
}
