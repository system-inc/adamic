package differential

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

/*
 * Controls on the instrument, before it measures anything real.
 *
 * Every test here is aimed at one specific way a differential can lie, and each was shown to fail by
 * mutating the harness before it was trusted. A harness that has never reported a mismatch has not
 * been shown to detect one, and a harness that has never reported 100% has not been shown to agree.
 */

// trimmingOracle stands in for Prettier: it rewrites trailing whitespace, and refuses files named BROKEN.
type trimmingOracle struct{ calls *atomic.Int64 }

func (trimmingOracle) Handles(string) bool { return true }

func (oracle trimmingOracle) Format(fileName string, text string) (string, error) {
	if oracle.calls != nil {
		oracle.calls.Add(1)
	}
	if strings.Contains(filepath.Base(fileName), "BROKEN") {
		return "", errors.New("SyntaxError: unexpected token")
	}
	lines := strings.Split(text, "\n")
	for index, line := range lines {
		lines[index] = strings.TrimRight(line, " \t")
	}
	return strings.Join(lines, "\n"), nil
}

// unchanged returns its input. It is the candidate that flatters an overall score.
type unchanged struct{}

func (unchanged) Handles(string) bool                          { return true }
func (unchanged) Format(_ string, text string) (string, error) { return text, nil }

// refuser handles nothing, which is what a native printer looks like before it exists.
type refuser struct{}

func (refuser) Handles(string) bool                   { return false }
func (refuser) Format(string, string) (string, error) { return "", errors.New("unreachable") }

func factory(formatter Formatter) NewFormatter {
	return func() (Formatter, error) { return formatter, nil }
}

// corpus writes files and returns their paths. Three are already formatted, two are not, one is BROKEN.
func corpus(t *testing.T) (string, []string) {
	t.Helper()
	root := t.TempDir()
	contents := map[string]string{
		"clean-a.ts":   "const a = 1;\n",
		"clean-b.md":   "# Title\n",
		"clean-c.json": "{}\n",
		"messy-d.ts":   "const d = 1;   \n",
		"messy-e.md":   "text\t\n",
		"BROKEN-f.ts":  "const {{{\n",
	}
	paths := make([]string, 0, len(contents))
	for name, text := range contents {
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	return root, paths
}

// TestOracleAgainstItselfReadsComplete proves the harness can report full agreement.
//
// The rewrite subset must be non-empty, or a 100% there is a perfect score over nothing.
func TestOracleAgainstItselfReadsComplete(t *testing.T) {
	t.Parallel()
	root, files := corpus(t)
	report, err := Compare(root, files, factory(trimmingOracle{}), factory(trimmingOracle{}), 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	total := report.Total()
	if total.Percent() != 100 {
		t.Fatalf("oracle against itself reads %.2f%%, want 100\n%s", total.Percent(), report.Summary())
	}
	if total.RewriteTotal != 2 || total.RewritePercent() != 100 {
		t.Fatalf("rewrites %d at %.2f%%, want 2 at 100%%", total.RewriteTotal, total.RewritePercent())
	}
}

// TestUnchangedCandidateIsCaughtOnRewrites is the flattery control.
//
// A candidate with no formatting ability at all scores 3 of 5 overall on this corpus, because three
// files are already formatted. On the rewrite subset it must score exactly zero, and the two files it
// failed must be the two the oracle rewrites.
func TestUnchangedCandidateIsCaughtOnRewrites(t *testing.T) {
	t.Parallel()
	root, files := corpus(t)
	report, err := Compare(root, files, factory(trimmingOracle{}), factory(unchanged{}), 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	total := report.Total()
	if total.Identical != 3 || total.Different != 2 {
		t.Fatalf("identical %d different %d, want 3 and 2\n%s", total.Identical, total.Different, report.Summary())
	}
	if total.RewriteTotal != 2 || total.RewriteIdentical != 0 {
		t.Fatalf("rewrite subset %d/%d, want 0/2: an unchanged candidate must score nothing where formatting happens",
			total.RewriteIdentical, total.RewriteTotal)
	}
	for _, result := range report.Files {
		if result.Outcome == Different && !strings.Contains(filepath.Base(result.Path), "messy") {
			t.Errorf("%s reported different, but the oracle leaves it alone", result.Path)
		}
	}
}

// TestRefusalNeverCountsAsMatch pins the starting point: no native printer reads 0%, not 100% of nothing.
func TestRefusalNeverCountsAsMatch(t *testing.T) {
	t.Parallel()
	root, files := corpus(t)
	report, err := Compare(root, files, factory(trimmingOracle{}), factory(refuser{}), 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	total := report.Total()
	if total.Identical != 0 || total.CandidateRefused != 5 || total.Percent() != 0 {
		t.Fatalf("refusing candidate: identical %d refused %d at %.2f%%, want 0, 5, 0%%",
			total.Identical, total.CandidateRefused, total.Percent())
	}
}

// TestOracleFailureLeavesTheDenominatorButIsNamed holds both halves.
//
// A file Prettier cannot parse has no right answer, so it must not count against the candidate. It must
// also not vanish: the summary names it, because a silently shrinking denominator is how a corpus
// measures less than it claims.
func TestOracleFailureLeavesTheDenominatorButIsNamed(t *testing.T) {
	t.Parallel()
	root, files := corpus(t)
	report, err := Compare(root, files, factory(trimmingOracle{}), factory(trimmingOracle{}), 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	total := report.Total()
	if total.OracleFailed != 1 || total.Measured() != 5 {
		t.Fatalf("oracle failed %d measured %d, want 1 and 5", total.OracleFailed, total.Measured())
	}
	if !strings.Contains(report.Summary(), "BROKEN-f.ts") {
		t.Fatalf("the oracle failure is not named in the summary:\n%s", report.Summary())
	}
}

// TestCacheServesTheSameAnswerAndKeysOnIdentity proves the cache saves work and cannot serve a stale oracle.
func TestCacheServesTheSameAnswerAndKeysOnIdentity(t *testing.T) {
	t.Parallel()
	root, files := corpus(t)
	directory := t.TempDir()
	calls := &atomic.Int64{}
	oracle := factory(trimmingOracle{calls: calls})

	first, err := Compare(root, files, oracle, factory(unchanged{}), 2, &OracleCache{Directory: directory, Identity: "bundles-1"})
	if err != nil {
		t.Fatal(err)
	}
	afterFirst := calls.Load()

	second, err := Compare(root, files, oracle, factory(unchanged{}), 2, &OracleCache{Directory: directory, Identity: "bundles-1"})
	if err != nil {
		t.Fatal(err)
	}
	if served := calls.Load() - afterFirst; served != 1 {
		t.Fatalf("second run called the oracle %d times, want 1 (only the uncacheable failure)", served)
	}
	if first.Total() != second.Total() {
		t.Fatalf("cached run tallied %+v, uncached %+v", second.Total(), first.Total())
	}

	beforeThird := calls.Load()
	if _, err := Compare(root, files, oracle, factory(unchanged{}), 2, &OracleCache{Directory: directory, Identity: "bundles-2"}); err != nil {
		t.Fatal(err)
	}
	if missed := calls.Load() - beforeThird; missed != int64(len(files)) {
		t.Fatalf("a different oracle identity reused %d cached answers, want every file recomputed", int64(len(files))-missed)
	}
}

// TestFirstDifferenceQuotesWhitespace pins why both lines are quoted: a trailing space must be visible.
func TestFirstDifferenceQuotesWhitespace(t *testing.T) {
	t.Parallel()
	got := FirstDifference("a\nb\n", "a\nb \n")
	if !strings.Contains(got, `"b"`) || !strings.Contains(got, `"b "`) || !strings.Contains(got, "line 2") {
		t.Fatalf("FirstDifference = %q", got)
	}
}
