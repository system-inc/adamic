// Package react_conformance scores an implementation against React's own compiler error fixtures.
//
// # What this is
//
// React's compiler ships 1,810 expectation files under `fixtures/compiler/`. Each pairs an input
// with a `.expect.md` holding either `## Code` (the transformed output) or `## Error` (the
// diagnostics) — never both, measured at the pinned sha as 1,485 `## Code` plus 325 `## Error`,
// zero overlap and zero neither. The 325 error goldens are plain text: a message, a
// `file:line:col`, and a caret span. Nothing in them is JavaScript-shaped, so a Go implementation
// can be scored against them without emitting a byte of JavaScript.
//
// That is the entire premise. The corpus is the asset; this runner is the cheap part.
//
// # What this deliberately is not
//
// It does not plug into upstream's Jest harness, and the reason is mechanical rather than a
// preference. `compiler/scripts/test-e2e.ts` declares `type Variant = 'babel'` with
// `ALL_VARIANTS = ['babel']`. There is no backend-registration mechanism to register into and no
// `--variant` value that would reach a Go implementation. Building against that surface would mean
// maintaining a Node shim to satisfy a Babel `PluginObj` contract in order to reach a corpus that
// is already readable as text.
//
// The contract shape instead comes from upstream's own
// `eslint-plugin-react-compiler/__tests__/RustBackendComparison-test.ts`, which compares two
// backends by asserting equal `{message, line}` pairs. That is backend-agnostic by construction: no
// AST, no transform, no Rust. This runner uses the same pair and takes the corpus from the 325
// goldens rather than that test's ~21 inline cases.
//
// # Reporting zero honestly
//
// Nothing implements these rules yet, so the first honest score is 0 of 325. The failure this
// package is built against is not a low score, it is a low score that cannot be told apart from an
// empty one: a runner that finds no fixtures, or refuses them all, also reports zero, and reads
// identically in a terminal. So `Score` carries `Considered`, `Excluded`, and `Refused` alongside
// `Passed`, no path can decrement `Considered` without incrementing exactly one of the outcome
// counters, and `TestRunnerCanDetectAPass` scores a hand-written correct result to prove the
// comparison can return true at all.
//
// # Why this suite is not gated
//
// The brief allowed a build tag or `testing.Short()` if the suite turned out slow or
// network-dependent. It is neither, and both properties come from the same decision: the corpus is
// vendored, so there is no fetch, and the work is reading 650 small files and splitting strings.
// Measured at about 0.3-0.8s for the whole package, which is inside the noise of the ordinary gate.
// Gating it would cost more than it saves — a suite behind a tag is a suite that stops being run,
// and the first thing it would stop catching is a vendored corpus that quietly changed size.
//
// `internal/lint/rules/react/tools/vendor_fixtures` is the network-touching half, and it is a command run deliberately
// when the pin moves, never from a test.
package react_conformance

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Fixture is one input paired with its expectation.
type Fixture struct {
	// Name is the path relative to the fixture root, e.g. `error.assign-global.js` or
	// `rules-of-hooks/error.use-in-loop.js`. Subdirectory fixtures keep their prefix, because 103
	// of the 325 live in one and a basename-keyed map would collide them.
	Name string

	InputPath  string
	ExpectPath string

	Source  string
	Pragmas []Pragma

	Expected Expectation
}

// RequiresFlow reports whether the input needs a Flow parser.
//
// The test is upstream's, verbatim from `packages/snap/src/compiler.ts:36`: the substring `@flow`
// anywhere in the source, not a first-line pragma. Measured at the pinned sha this is 35 of the
// 325, so 290 need no Flow parser. Restricting the search to the first line instead reports 20,
// which is wrong in the direction that matters: it would hand 15 unparseable fixtures to a
// TypeScript parser and read their parse failures as rule failures.
func (f Fixture) RequiresFlow() bool {
	return strings.Contains(f.Source, "@flow")
}

// Expectation is the parsed `## Error` block.
type Expectation struct {
	// Raw is the fenced block verbatim, kept so a mismatch can be shown rather than described.
	Raw string

	// DeclaredCount is the N from `Found N errors`. It is 0 with HasDeclaredCount false for the 14
	// fixtures that carry no such header: those are parse errors (`Missing semicolon. (7:24)`) and
	// pipeline exceptions (`unexpected error`), a different diagnostic shape entirely.
	DeclaredCount    int
	HasDeclaredCount bool

	Errors []ExpectedError
}

// ExpectedError is one diagnostic from the golden.
type ExpectedError struct {
	// Heading is the severity word the golden printed: `Error`, `Compilation Skipped`, `Invariant`,
	// or `Todo`. Carried rather than discarded because a Go port that reports the right message
	// under the wrong heading has got a real thing wrong, and a comparison that never saw the
	// heading could not say so.
	Heading string

	// Message is the first line after the heading, which is the part upstream's own backend
	// comparison asserts on. The paragraphs below it are prose with react.dev URLs that change
	// without notice; asserting on them would turn documentation edits into rule regressions.
	Message string

	// File, Line, Column come from the FIRST `path:line:col` line after the message, which is the
	// primary span. A diagnostic can carry more than one: 31 of the 429 have a secondary location
	// underneath the primary, captioned separately (`This modifies `cache``) to point at the
	// contributing site. Taking the first is what upstream's `file:line:col` header means, and
	// counting them all would report 427 locations for 429 diagnostics while 31 diagnostics
	// actually have none — a number that is wrong in both directions at once and looks reasonable.
	//
	// Line and Column are 0 with HasLocation false when absent. Measured at the pinned sha: 398 of
	// the 429 carry a primary location, 31 do not.
	File   string
	Line   int
	Column int

	HasLocation bool
}

// Result is what an implementation reported for one fixture. This is the interface an
// implementation satisfies; nothing in this package produces one yet.
type Result struct {
	Errors []ReportedError
}

// ReportedError is one diagnostic an implementation produced.
type ReportedError struct {
	// Heading is the severity word, one of `Error`, `Compilation Skipped`, `Invariant`, `Todo`.
	Heading string

	Message string
	Line    int
}

// Load reads every vendored fixture pair.
//
// It returns an error rather than a short list when anything is missing. A loader that skips an
// unreadable pair and returns the rest hands the scorer a smaller corpus that scores as a clean
// run, which is the failure this package exists to make impossible.
func Load(root string) ([]Fixture, error) {
	var inputPaths []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || strings.HasSuffix(path, ".expect.md") {
			return nil
		}
		inputPaths = append(inputPaths, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(inputPaths)

	// The corpus must hold both kinds, and the count of each is asserted rather than merely being
	// nonzero.
	//
	// This replaces a guard that rejected any fixture not named `error.*`, justified in its own
	// comment as protecting a tree that "holds only error-named fixtures" -- a property that guard
	// created. The consequence was not cosmetic: a corpus of error fixtures cannot measure false
	// positives, because over-reporting produces more findings and reads as success. That limitation
	// was reasoned about for hours as a property of upstream. Upstream ships the clean fixtures; this
	// loader excluded them.
	//
	// So the replacement asserts what a correct tree contains instead of what it excludes, and it
	// asserts exact counts because a floor cannot see the failure that produced this. Collecting
	// fixtures by copying them into one directory loses any whose basenames collide, and four of the
	// clean ones do -- that collection returned 66 of 70 and errored on nothing. A check for "some
	// clean fixtures exist" passes on 66 as happily as on 70.
	var errorNamed, clean int
	for _, path := range inputPaths {
		if base := filepath.Base(path); strings.HasPrefix(base, "error.") ||
			strings.HasPrefix(base, "todo.error.") {
			errorNamed++
			continue
		}
		clean++
	}
	if errorNamed != ExpectedErrorFixtureCount || clean != ExpectedCleanFixtureCount {
		return nil, fmt.Errorf(
			"corpus holds %d error-named and %d clean fixtures, want %d and %d; re-run "+
				"`internal/lint/rules/react/tools/vendor_fixtures` at the pinned sha, and if upstream really changed, "+
				"read the diff before moving these numbers",
			errorNamed, clean, ExpectedErrorFixtureCount, ExpectedCleanFixtureCount)
	}

	fixtures := make([]Fixture, 0, len(inputPaths))
	for _, inputPath := range inputPaths {
		name, err := filepath.Rel(root, inputPath)
		if err != nil {
			return nil, err
		}
		// Slashed, because the stated divergences are keyed `new-mutability/error.x.js`, and on Windows
		// Rel answers with `\`, so every lookup missed and a known divergence scored as a new one.
		name = filepath.ToSlash(name)

		sourceBytes, err := os.ReadFile(inputPath)
		if err != nil {
			return nil, err
		}
		source := string(sourceBytes)

		expectPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".expect.md"
		expectBytes, err := os.ReadFile(expectPath)
		if err != nil {
			return nil, fmt.Errorf("%s: expectation missing; upstream pairs these 1:1 and a fixture without one cannot be scored: %w", name, err)
		}

		// The `## Error` invariant is conditional on the fixture being error-named, rather than
		// universal. Error-named fixtures and `## Error` sections are still exactly the same set,
		// which is what makes a missing section on one of those a real defect. A clean fixture has
		// no section by definition -- that absence is its expectation -- so requiring one of every
		// fixture is what made the corpus error-only, and it would reject the clean population on
		// the way in.
		base := filepath.Base(name)
		errorNamed := strings.HasPrefix(base, "error.") || strings.HasPrefix(base, "todo.error.")

		var expectation Expectation
		if errorNamed {
			expectation, err = ParseExpectation(string(expectBytes))
			if err != nil {
				return nil, fmt.Errorf("%s: %w", name, err)
			}
		} else if strings.Contains("\n"+string(expectBytes), "\n## Error") {
			// The other direction of the same set equality, and it is the one that matters for a
			// scorer: a fixture counted as clean while its golden reports errors would make every
			// finding on it read as a false positive.
			return nil, fmt.Errorf("%s: not error-named but its expectation has an `## Error` "+
				"section, so it is neither a clean fixture nor an error one and cannot be scored", name)
		}

		firstLine, _, _ := strings.Cut(source, "\n")
		fixtures = append(fixtures, Fixture{
			Name:       name,
			InputPath:  inputPath,
			ExpectPath: expectPath,
			Source:     source,
			Pragmas:    ParsePragmas(firstLine),
			Expected:   expectation,
		})
	}
	return fixtures, nil
}

// ParseExpectation reads the `## Error` block out of a `.expect.md`.
//
// An expectation with no `## Error` section is an error, not an empty result. Error-named fixtures
// and `## Error` sections are the same set at the pinned sha — verified as exact set equality in
// both directions, 325 each way, zero counterexamples — so a missing section means the corpus is
// not what this runner believes it is, and continuing would score a transform fixture as a rule
// that found nothing.
func ParseExpectation(text string) (Expectation, error) {
	_, after, found := strings.Cut(text, "\n## Error\n")
	if !found {
		return Expectation{}, fmt.Errorf("no `## Error` section; error-named fixtures always have one at sha %s", UpstreamSha)
	}

	block, found := fencedBlock(after)
	if !found {
		return Expectation{}, fmt.Errorf("`## Error` section has no fenced block")
	}

	expectation := Expectation{Raw: block}

	lines := strings.Split(block, "\n")
	for index := 0; index < len(lines); index++ {
		// Go whitespace: the layout of a vendored markdown golden, which no upstream code parses.
		line := strings.TrimSpace(lines[index])

		if count, ok := parseFoundHeader(line); ok {
			expectation.DeclaredCount = count
			expectation.HasDeclaredCount = true
			continue
		}

		heading, message, ok := cutHeading(line)
		if !ok {
			continue
		}

		// Go whitespace: the layout of a vendored markdown golden, which no upstream code parses.
		expected := ExpectedError{Heading: heading, Message: strings.TrimSpace(message)}

		// The location follows the message and its prose, and is the next line that parses as
		// `path:line:col`. Scanning forward to the next heading bounds the search so a multi-error
		// golden does not attach the second error's location to the first.
		for scan := index + 1; scan < len(lines); scan++ {
			// Go whitespace: the layout of a vendored markdown golden, which no upstream code parses.
			candidate := strings.TrimSpace(lines[scan])
			if _, _, isHeading := cutHeading(candidate); isHeading {
				break
			}
			if file, lineNumber, column, ok := parseLocation(candidate); ok {
				expected.File = file
				expected.Line = lineNumber
				expected.Column = column
				expected.HasLocation = true
				break
			}
		}

		expectation.Errors = append(expectation.Errors, expected)
	}

	return expectation, nil
}

// diagnosticHeadings are the four severity words a golden can print before a diagnostic.
//
// This list is not inferred from the corpus, and the difference matters. Taking it from the corpus
// would produce whatever spellings happen to appear in the 325, and a heading used by a rule with
// no error fixture would be missing with nothing to say so. It comes instead from upstream's
// printer, `CompilerError.ts` around line 570, where a switch over every `ErrorCategory` assigns
// exactly one of four headings and closes with `assertExhaustive`. That exhaustiveness check is the
// guarantee the list is complete: a new category cannot be added upstream without either reusing
// one of these or failing to compile.
//
// This cost a defect worth recording. The parser first matched `Error: ` only, and read 302 of the
// 429 diagnostics — a 70% read that produced a fully green parse and a runner that would have
// scored 37 fixtures as expecting nothing at all. Nothing about the output looked partial. It was
// caught by `TestDeclaredCountMatchesParsedErrors`, which compares the parser's count against the
// `Found N errors` header the goldens carry independently, and that test exists solely because two
// numbers that were not derived from each other can disagree.
var diagnosticHeadings = []string{"Error", "Compilation Skipped", "Invariant", "Todo"}

// cutHeading splits a diagnostic line into its heading and message.
//
// `Compilation Skipped` is checked before `Error` only incidentally; the headings share no prefix,
// so order does not change the result.
func cutHeading(line string) (heading string, message string, ok bool) {
	for _, candidate := range diagnosticHeadings {
		if rest, found := strings.CutPrefix(line, candidate+": "); found {
			return candidate, rest, true
		}
	}
	return "", "", false
}

// fencedBlock returns the contents of the first ``` fence in text.
func fencedBlock(text string) (string, bool) {
	_, after, found := strings.Cut(text, "```")
	if !found {
		return "", false
	}
	// Drop the info string on the opening fence, if any.
	_, after, found = strings.Cut(after, "\n")
	if !found {
		return "", false
	}
	body, _, found := strings.Cut(after, "\n```")
	if !found {
		return "", false
	}
	return body, true
}

// parseFoundHeader reads `Found 3 errors` and returns 3.
//
// Both spellings occur: `Found 1 error` and `Found 2 errors`. Matching only the plural would drop
// the 238 single-error fixtures, which is 73% of the corpus, and would still leave a suite that
// runs and reports numbers.
func parseFoundHeader(line string) (int, bool) {
	rest, ok := strings.CutPrefix(line, "Found ")
	if !ok {
		return 0, false
	}
	digits, rest, found := strings.Cut(rest, " ")
	if !found {
		return 0, false
	}
	if rest != "error" && rest != "errors" && !strings.HasPrefix(rest, "error") {
		return 0, false
	}
	count, err := strconv.Atoi(digits)
	if err != nil {
		return 0, false
	}
	return count, true
}

// parseLocation reads `error.foo.ts:3:4` into its parts.
//
// It requires the line to be exactly a location. A prose line can contain a colon-delimited run
// that looks like one, and a looser match would attach a URL fragment to a diagnostic as its
// source position.
func parseLocation(line string) (file string, lineNumber int, column int, ok bool) {
	if line == "" || strings.ContainsAny(line, " \t") {
		return "", 0, 0, false
	}
	lastColon := strings.LastIndex(line, ":")
	if lastColon < 0 {
		return "", 0, 0, false
	}
	column, err := strconv.Atoi(line[lastColon+1:])
	if err != nil {
		return "", 0, 0, false
	}
	rest := line[:lastColon]
	secondColon := strings.LastIndex(rest, ":")
	if secondColon < 0 {
		return "", 0, 0, false
	}
	lineNumber, err = strconv.Atoi(rest[secondColon+1:])
	if err != nil {
		return "", 0, 0, false
	}
	file = rest[:secondColon]
	if file == "" {
		return "", 0, 0, false
	}
	return file, lineNumber, column, true
}
