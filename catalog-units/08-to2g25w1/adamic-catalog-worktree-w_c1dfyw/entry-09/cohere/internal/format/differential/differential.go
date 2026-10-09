// Package differential measures a candidate formatter against the Prettier fork, byte for byte, over
// real corpora.
//
// It exists because cohere is replacing Prettier with a native printer, and "byte-identical on our
// code" is a claim no amount of reading either implementation can establish. The acceptance test is
// this package's number reaching 100% on every corpus we own. Until then the number is the progress
// report, so the instrument has to be right before anything is measured with it.
//
// # The oracle is goja(file), not file
//
// The predecessor of this harness (internal/format/comparison) treated every tree file as a fixed point
// of Prettier, so any change a candidate proposed was a divergence. That no longer holds. The projects
// run Prettier with prettier-plugin-tailwindcss, and the embedded engine runs core Prettier without
// it, so the tree is a fixed point of a different formatter than the one being matched. Class order is
// cohere's lint fixer's job, not the format phase's. So the oracle output is computed, never assumed.
//
// # Two numbers, because one of them flatters
//
// The corpora are almost entirely already formatted. A candidate that returns its input unchanged
// matches every file the oracle leaves alone and would report a high overall score while having no
// line-breaking engine at all. So every tally carries the subset the oracle actually rewrites, and the
// identity rate on that subset is reported separately. That subset is where formatting ability shows;
// the rest only proves a candidate does not break formatted code.
//
// # What never counts as a match
//
// A candidate that refuses a file, or errors on it, has not produced the oracle's bytes. Refusal is a
// separate outcome so it is visible, and it sits in the denominator. A file the oracle itself cannot
// format (a genuine parse error in the tree) has no right answer to match, so it leaves the denominator
// but is counted and named, never dropped silently.
package differential

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/system-inc/cohere/internal/types/sourcename"
)

// Formatter is either side of the comparison.
//
// It is the shape prettier.Engine already has, so the embedded Prettier is an oracle with no adapter,
// and a native printer only has to say which files it handles and format them.
type Formatter interface {
	Handles(fileName string) bool
	Format(fileName string, text string) (string, error)
}

// NewFormatter builds one Formatter for one worker.
//
// A factory rather than a value because the embedded engine is a goja runtime and goja runtimes are not
// safe for concurrent use. Each worker owns its own instance, so parallelism never shares one.
type NewFormatter func() (Formatter, error)

// Outcome is what happened to one file.
type Outcome string

const (
	// Identical means the candidate produced the oracle's bytes exactly.
	Identical Outcome = "Identical"

	// Different means the candidate produced bytes, and they were not the oracle's.
	Different Outcome = "Different"

	// CandidateRefused means the candidate did not handle the file or errored on it. Counts against.
	CandidateRefused Outcome = "CandidateRefused"

	// OracleFailed means Prettier itself could not format the file. Excluded from the denominator,
	// because there is no correct output to match, and named in the report.
	OracleFailed Outcome = "OracleFailed"
)

// FileResult is one file's comparison.
type FileResult struct {
	Path     string
	Language string
	Outcome  Outcome

	// OracleRewrites is whether the oracle's output differs from the file on disk. True files are the
	// subset that tests formatting rather than preservation.
	OracleRewrites bool

	// Detail is the first difference for Different, the error for CandidateRefused and OracleFailed.
	Detail string
}

// Tally counts outcomes for one language.
type Tally struct {
	Identical        int
	Different        int
	CandidateRefused int
	OracleFailed     int

	// RewriteIdentical and RewriteTotal are the same count restricted to files the oracle rewrites.
	RewriteIdentical int
	RewriteTotal     int
}

// Measured is the denominator: every file with a correct answer to match.
func (tally Tally) Measured() int {
	return tally.Identical + tally.Different + tally.CandidateRefused
}

// Percent is the identity rate over measured files. A tally with nothing measured reports zero rather
// than dividing by zero or claiming a perfect score over an empty set.
func (tally Tally) Percent() float64 {
	if tally.Measured() == 0 {
		return 0
	}
	return 100 * float64(tally.Identical) / float64(tally.Measured())
}

// RewritePercent is the identity rate over files the oracle rewrites.
func (tally Tally) RewritePercent() float64 {
	if tally.RewriteTotal == 0 {
		return 0
	}
	return 100 * float64(tally.RewriteIdentical) / float64(tally.RewriteTotal)
}

// Report is one corpus measured.
type Report struct {
	// Root is the corpus that was walked. A percentage means nothing without the root it was taken
	// against, which is how a measurement silently covers half a tree.
	Root string

	ByLanguage map[string]*Tally
	Files      []FileResult
}

// Total sums every language.
func (report Report) Total() Tally {
	total := Tally{}
	for _, tally := range report.ByLanguage {
		total.Identical += tally.Identical
		total.Different += tally.Different
		total.CandidateRefused += tally.CandidateRefused
		total.OracleFailed += tally.OracleFailed
		total.RewriteIdentical += tally.RewriteIdentical
		total.RewriteTotal += tally.RewriteTotal
	}
	return total
}

// Language buckets a file by what parses it, which is what the port is organized by. An Adamic `.a` file is
// parsed as TypeScript, so it buckets with `.ts` (#6mhafvb).
func Language(fileName string) string {
	switch strings.ToLower(filepath.Ext(sourcename.TreatedAs(fileName))) {
	case ".ts", ".mts", ".cts":
		return "ts"
	case ".tsx":
		return "tsx"
	case ".js", ".mjs", ".cjs", ".jsx":
		return "js"
	case ".json", ".jsonc", ".json5":
		return "json"
	case ".md", ".markdown":
		return "md"
	case ".graphql", ".gql":
		return "graphql"
	case ".css", ".scss", ".less":
		return "css"
	case ".yaml", ".yml":
		return "yaml"
	default:
		return "other"
	}
}

// Compare formats every file with both sides and tallies the result.
//
// The oracle side goes through the cache when one is given, because the oracle is deterministic in its
// bundle digest, its options and the file's bytes, and it is by far the slower side. Re-reading the
// progress number should cost the candidate's time, not Prettier's.
func Compare(root string, files []string, newOracle NewFormatter, newCandidate NewFormatter, workers int, cache *OracleCache) (Report, error) {
	if workers < 1 {
		workers = 1
	}

	results := make([]FileResult, len(files))
	indexes := make(chan int)
	errs := make(chan error, workers)
	var group sync.WaitGroup

	for worker := 0; worker < workers; worker++ {
		group.Add(1)
		go func() {
			defer group.Done()
			oracle, err := newOracle()
			if err != nil {
				errs <- fmt.Errorf("building the oracle: %w", err)
				return
			}
			candidate, err := newCandidate()
			if err != nil {
				errs <- fmt.Errorf("building the candidate: %w", err)
				return
			}
			for index := range indexes {
				results[index] = compareOne(files[index], oracle, candidate, cache)
			}
		}()
	}

	for index := range files {
		select {
		case err := <-errs:
			close(indexes)
			group.Wait()
			return Report{}, err
		case indexes <- index:
		}
	}
	close(indexes)
	group.Wait()
	select {
	case err := <-errs:
		return Report{}, err
	default:
	}

	report := Report{Root: root, ByLanguage: map[string]*Tally{}, Files: results}
	for _, result := range results {
		tally := report.ByLanguage[result.Language]
		if tally == nil {
			tally = &Tally{}
			report.ByLanguage[result.Language] = tally
		}
		switch result.Outcome {
		case Identical:
			tally.Identical++
		case Different:
			tally.Different++
		case CandidateRefused:
			tally.CandidateRefused++
		case OracleFailed:
			tally.OracleFailed++
			continue
		}
		if result.OracleRewrites {
			tally.RewriteTotal++
			if result.Outcome == Identical {
				tally.RewriteIdentical++
			}
		}
	}
	return report, nil
}

// compareOne measures a single file.
func compareOne(path string, oracle Formatter, candidate Formatter, cache *OracleCache) FileResult {
	result := FileResult{Path: path, Language: Language(path)}

	source, err := os.ReadFile(path)
	if err != nil {
		result.Outcome = OracleFailed
		result.Detail = "reading the file: " + err.Error()
		return result
	}
	text := string(source)

	expected, err := cache.format(oracle, path, text)
	if err != nil {
		result.Outcome = OracleFailed
		result.Detail = err.Error()
		return result
	}
	result.OracleRewrites = expected != text

	if !candidate.Handles(path) {
		result.Outcome = CandidateRefused
		result.Detail = "the candidate does not handle this file type"
		return result
	}
	actual, err := candidate.Format(path, text)
	if err != nil {
		result.Outcome = CandidateRefused
		result.Detail = err.Error()
		return result
	}

	if actual == expected {
		result.Outcome = Identical
		return result
	}
	result.Outcome = Different
	result.Detail = FirstDifference(expected, actual)
	return result
}

// FirstDifference names the first line where two outputs diverge.
//
// A line, not a byte offset, because the reader is going to open the file. Both lines are quoted so a
// whitespace-only difference is visible rather than looking like two identical strings.
func FirstDifference(expected string, actual string) string {
	expectedLines := strings.Split(expected, "\n")
	actualLines := strings.Split(actual, "\n")
	for line := 0; line < len(expectedLines) || line < len(actualLines); line++ {
		var want, got string
		if line < len(expectedLines) {
			want = expectedLines[line]
		}
		if line < len(actualLines) {
			got = actualLines[line]
		}
		if want != got || line >= len(expectedLines) || line >= len(actualLines) {
			return fmt.Sprintf("line %d: want %q, got %q", line+1, truncate(want), truncate(got))
		}
	}
	return "identical by line, different by bytes (a trailing newline or a carriage return)"
}

func truncate(line string) string {
	const limit = 120
	if len(line) <= limit {
		return line
	}
	return line[:limit] + "…"
}

// Summary renders a report for a terminal: per language, then the total, then failures by name.
func (report Report) Summary() string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "corpus %s\n", report.Root)

	languages := make([]string, 0, len(report.ByLanguage))
	for language := range report.ByLanguage {
		languages = append(languages, language)
	}
	sort.Strings(languages)

	row := func(name string, tally Tally) {
		fmt.Fprintf(&builder, "  %-8s %6.2f%%  %5d/%-5d identical  rewrites %6.2f%% (%d/%d)  refused %d  oracle failed %d\n",
			name, tally.Percent(), tally.Identical, tally.Measured(),
			tally.RewritePercent(), tally.RewriteIdentical, tally.RewriteTotal,
			tally.CandidateRefused, tally.OracleFailed)
	}
	for _, language := range languages {
		row(language, *report.ByLanguage[language])
	}
	row("total", report.Total())

	for _, result := range report.Files {
		if result.Outcome == OracleFailed {
			fmt.Fprintf(&builder, "  oracle failed: %s: %s\n", result.Path, firstLine(result.Detail))
		}
	}
	return builder.String()
}

func firstLine(text string) string {
	scanner := bufio.NewScanner(strings.NewReader(text))
	if scanner.Scan() {
		return scanner.Text()
	}
	return text
}

// OracleCache stores oracle output on disk, keyed by everything the output depends on.
//
// The key is the bundle digest, a description of the options, the file name (the parser is chosen by
// it) and the file's bytes. Leave any of those out and a stale answer is served for a changed input,
// which would make the harness report a match against an oracle that no longer exists. A nil cache
// formats every time.
type OracleCache struct {
	Directory string

	// Identity is the bundle digest plus the options, everything about the oracle that is not the file.
	Identity string
}

func (cache *OracleCache) format(oracle Formatter, path string, text string) (string, error) {
	if cache == nil {
		return oracle.Format(path, text)
	}

	hash := sha256.New()
	fmt.Fprintf(hash, "%s\x00%s\x00%d\x00", cache.Identity, filepath.Base(path), len(text))
	hash.Write([]byte(text))
	key := hex.EncodeToString(hash.Sum(nil))
	entry := filepath.Join(cache.Directory, key[:2], key)

	if cached, err := os.ReadFile(entry); err == nil {
		return string(cached), nil
	}

	formatted, err := oracle.Format(path, text)
	if err != nil {
		return "", err
	}
	// A unique temporary per writer, then rename. Two workers can hold the same file (a duplicated
	// fixture, a copied config) and would otherwise write one shared temporary at the same instant.
	// The cache is an optimization, so any failure here is swallowed and the answer still returned.
	if err := os.MkdirAll(filepath.Dir(entry), 0o755); err == nil {
		if temporary, err := os.CreateTemp(filepath.Dir(entry), key+".*.tmp"); err == nil {
			_, writeError := temporary.WriteString(formatted)
			closeError := temporary.Close()
			if writeError == nil && closeError == nil {
				_ = os.Rename(temporary.Name(), entry)
			} else {
				_ = os.Remove(temporary.Name())
			}
		}
	}
	return formatted, nil
}
