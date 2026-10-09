package registry

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

/*
 * ESLint 10.8.1's own test rows for every core rule cohere registers, replayed through cohere's rule
 * (#jjfa7qb). Each rule's authority is whoever wrote it, not whichever reimplementation we ported from,
 * and these rows are what ESLint's authors believe each rule does.
 *
 * The rows are vendored in testdata/eslint-corpus/eslint-10.8.1, one file per rule, each with ESLint's
 * verdict recorded when it was extracted (internal/lint/tools/eslint_corpus/extract.mjs, which nothing
 * runs). Every row runs as a module named Case.ts in both engines. A row agrees when cohere reports
 * findings at the same places ESLint does: the same start and end offsets, not only the same text.
 *
 * Every row that does not agree is listed in known-gaps.json under its rule, with the way it disagrees,
 * and the list is held exactly: an unlisted disagreement fails, a listed one that now disagrees
 * differently fails, and a listed one that now agrees fails until it is taken off. The list only shrinks
 * on purpose. Kirk's resolution order decides each one: first principles, and ESLint when that is
 * unclear. A divergence kept on first principles is recorded at the rule's line as well as here.
 *
 * COHERE_ESLINT_CORPUS_WRITE_GAPS=1 rewrites the list from what the run found, for a wave that settles
 * rows. Read the diff it leaves: a row leaving the list is the point, a row joining it is a regression.
 */

const (
	eslintCorpusDirectory = "testdata/eslint-corpus/eslint-10.8.1"
	eslintCorpusGapsFile  = "testdata/eslint-corpus/known-gaps.json"

	// The whole of ESLint 10.8.1's test rows for the 168 core rules cohere registers. Asserted so a
	// re-extract that lost rows fails here rather than reading as agreement.
	eslintCorpusRowCount = 18588
)

// eslintCorpusRow is one vendored row: the code, its options, and ESLint's verdict as message id and offsets
type eslintCorpusRow struct {
	Code        string            `json:"code"`
	Options     []json.RawMessage `json:"options"`
	UsesGlobals bool              `json:"usesGlobals"`
	Skipped     string            `json:"skipped"`
	ESLint      [][3]any          `json:"eslint"`
}

// key names a row by its code and options, so the gaps list survives a re-extract that reorders rows
func (row eslintCorpusRow) key() string {
	options, _ := json.Marshal(row.Options)
	sum := sha256.Sum256([]byte(row.Code + "\x00" + string(options)))
	return hex.EncodeToString(sum[:6])
}

// decodeCorpusOptions turns a row's ESLint options into the rule's decoded options, as the config layer would
func decodeCorpusOptions(registration rule.Registration, options []json.RawMessage) (any, string) {
	switch {
	case registration.DecodeOptionList != nil:
		var list []byte
		if len(options) > 0 {
			list, _ = json.Marshal(options)
		}
		decoded, err := registration.DecodeOptionList(list)
		if err != nil {
			return nil, "options refused: " + err.Error()
		}
		return decoded, ""
	case registration.Decode != nil:
		if len(options) > 1 {
			return nil, "more option elements than cohere's decoder takes"
		}
		if len(options) == 0 {
			return nil, ""
		}
		decoded, err := registration.Decode(options[0])
		if err != nil {
			return nil, "options refused: " + err.Error()
		}
		return decoded, ""
	case registration.DecodeAt != nil:
		return nil, "options anchored on disk, which a corpus row cannot name"
	default:
		if len(options) > 0 {
			return nil, "cohere's rule takes no options"
		}
		return nil, ""
	}
}

// utf16ToByteOffsets maps each UTF-16 offset in code to its byte offset, one entry past the end included.
//
// ESLint's offsets are JavaScript string indices, which count UTF-16 code units, and Go slices bytes. On
// ASCII the two agree, so the difference hid until a row had a non-ASCII character before its finding:
// then the expected slice was the wrong text, and 122 rows of the unicode rules read as span gaps that
// were not there. A character outside the Basic Multilingual Plane is two units and four bytes; a lone
// surrogate, which several rows hold on purpose, decodes to U+FFFD, one unit either way.
func utf16ToByteOffsets(code string) []int {
	offsets := make([]int, 0, len(code)+1)
	for index, character := range code {
		offsets = append(offsets, index)
		if character > 0xFFFF {
			offsets = append(offsets, index)
		}
	}
	return append(offsets, len(code))
}

// corpusVerdict replays one row through the rule and says how cohere's answer compares with ESLint's
func corpusVerdict(t *testing.T, registration rule.Registration, row eslintCorpusRow) string {
	decoded, problem := decodeCorpusOptions(registration, row.Options)
	if problem != "" {
		return "options"
	}

	var result rule_testing.Result
	ran := false
	t.Run(row.key(), func(t *testing.T) {
		result = rule_testing.RunTypedVerbatimWithOptions(t, registration.Rule, "Case.ts", row.Code, decoded)
		ran = true
	})
	if !ran {
		return "harness"
	}

	// The row is written verbatim, so cohere's offsets are into the row's own bytes, with one
	// exception: the compiler's reader drops a leading byte order mark, so the file cohere sees is the
	// row without it. ESLint's offsets are moved into that same text. ESLint counts columns after the
	// mark, and the extractor counted line starts with it, so a finding on the first line is already an
	// offset into the text without the mark, and one on a later line is one unit past it.
	//
	// The places compare as offsets and not as text. Comparing the text alone read a finding on the
	// wrong occurrence of the same text as agreement: ESLint reports prefer-const at the write in
	// `let x; x = 0;` and the port at the declaration, both slices are `x`, and the row agreed even
	// across two lines (#jjfa7qb). TestCorpusVerdictComparesPlaces plants that shape.
	code := row.Code
	firstLineEnd := -1
	if strings.HasPrefix(code, "\ufeff") {
		code = strings.TrimPrefix(code, "\ufeff")
		firstLineEnd = utf16LineBreakEnd(row.Code)
	}
	if result.SourceFile.Text() != code {
		return "harness"
	}
	byteOffset := utf16ToByteOffsets(code)
	expected := make([]string, 0, len(row.ESLint))
	for _, finding := range row.ESLint {
		start, end := int(finding[1].(float64)), int(finding[2].(float64))
		if firstLineEnd >= 0 {
			if start >= firstLineEnd {
				start--
			}
			if end >= firstLineEnd {
				end--
			}
		}
		expected = append(expected, corpusPlace(byteOffset[start], byteOffset[end]))
	}
	reported := make([]string, 0, len(result.Diagnostics))
	for _, diagnostic := range result.Diagnostics {
		start, end := diagnostic.Range.Pos(), diagnostic.Range.End()
		if start < 0 || end > len(code) || start > end {
			reported = append(reported, fmt.Sprintf("<range %d-%d>", start, end))
			continue
		}
		reported = append(reported, corpusPlace(start, end))
	}
	sort.Strings(expected)
	sort.Strings(reported)

	switch {
	case strings.Join(reported, "\x00") == strings.Join(expected, "\x00"):
		return "agree"
	case row.UsesGlobals:
		// ESLint's globals are configuration cohere has no counterpart for: it asks the checker what a name is
		return "globals"
	case len(reported) == len(expected):
		return "span"
	case len(reported) > len(expected):
		return "extra"
	default:
		return "missing"
	}
}

// utf16LineBreakEnd is the UTF-16 offset just past the first line break, as the extractor counts line
// starts, or one past the end when there is none.
func utf16LineBreakEnd(code string) int {
	units := 0
	runes := []rune(code)
	for index, character := range runes {
		switch character {
		case '\r':
			if index+1 < len(runes) && runes[index+1] == '\n' {
				return units + 2
			}
			return units + 1
		case '\n', '\u2028', '\u2029':
			return units + 1
		}
		units++
		if character > 0xFFFF {
			units++
		}
	}
	return units + 1
}

// corpusPlace names a finding by where it is, as byte offsets into the row's code.
func corpusPlace(start int, end int) string {
	return fmt.Sprintf("%d-%d", start, end)
}

func TestCohereAgreesWithESLintsCoreCorpus(t *testing.T) {
	t.Parallel()

	registrations := map[string]rule.Registration{}
	coreRules := []string{}
	for _, registration := range rule.Registered() {
		registrations[registration.Rule.Name] = registration
		if !strings.Contains(registration.Rule.Name, "/") {
			coreRules = append(coreRules, registration.Rule.Name)
		}
	}

	gapsText, err := os.ReadFile(eslintCorpusGapsFile)
	if err != nil {
		t.Fatal(err)
	}
	knownGaps := map[string]map[string]string{}
	if err := json.Unmarshal(gapsText, &knownGaps); err != nil {
		t.Fatalf("reading %s: %v", eslintCorpusGapsFile, err)
	}

	corpus := map[string][]eslintCorpusRow{}
	rowCount := 0
	for _, name := range coreRules {
		text, readError := os.ReadFile(filepath.Join(eslintCorpusDirectory, name+".json"))
		if readError != nil {
			t.Errorf("%s has no vendored ESLint rows: re-extract with internal/lint/tools/eslint_corpus", name)
			continue
		}
		var rows []eslintCorpusRow
		if err := json.Unmarshal(text, &rows); err != nil {
			t.Fatalf("reading %s's rows: %v", name, err)
		}
		corpus[name] = rows
		rowCount += len(rows)
	}
	if rowCount != eslintCorpusRowCount {
		t.Fatalf("%d vendored rows, and ESLint 10.8.1 has %d for these rules: a re-extract lost or gained rows",
			rowCount, eslintCorpusRowCount)
	}
	for name := range knownGaps {
		if _, isCore := corpus[name]; !isCore {
			t.Errorf("known-gaps.json lists %s, which has no vendored rows", name)
		}
	}

	var mutex sync.Mutex
	found := map[string]map[string]string{}
	// Not parallel: this group holds its parallel per-rule subtests until they finish, so the test can
	// read found after it returns.
	t.Run("rules", func(t *testing.T) {
		for _, name := range coreRules {
			rows, hasRows := corpus[name]
			if !hasRows {
				continue
			}
			t.Run(name, func(t *testing.T) {
				t.Parallel()
				gaps := map[string]string{}
				for _, row := range rows {
					if row.Skipped != "" {
						continue
					}
					if verdict := corpusVerdict(t, registrations[name], row); verdict != "agree" {
						gaps[row.key()] = verdict
					}
				}
				mutex.Lock()
				found[name] = gaps
				mutex.Unlock()

				listed := knownGaps[name]
				for key, verdict := range gaps {
					switch listedVerdict, isListed := listed[key]; {
					case !isListed:
						t.Errorf("row %s disagrees with ESLint (%s) and is not a known gap", key, verdict)
					case listedVerdict != verdict:
						t.Errorf("row %s is a known %s gap and now disagrees as %s", key, listedVerdict, verdict)
					}
				}
				for key := range listed {
					if _, stillDisagrees := gaps[key]; !stillDisagrees {
						t.Errorf("row %s now agrees with ESLint: take it off known-gaps.json", key)
					}
				}
			})
		}
	})

	if os.Getenv("COHERE_ESLINT_CORPUS_WRITE_GAPS") == "1" {
		for name, gaps := range found {
			if len(gaps) == 0 {
				delete(found, name)
			}
		}
		encoded, _ := json.MarshalIndent(found, "", "    ")
		if err := os.WriteFile(eslintCorpusGapsFile, append(encoded, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// ESLint counts UTF-16 units: é is one unit and two bytes, 😀 is two units and four bytes.
func TestUTF16OffsetsBecomeByteOffsets(t *testing.T) {
	t.Parallel()

	code := "é😀x"
	offsets := utf16ToByteOffsets(code)
	// Units: é at 0, 😀 at 1 and 2, x at 3, the end at 4
	want := []int{0, 2, 2, 6, 7}
	if fmt.Sprint(offsets) != fmt.Sprint(want) {
		t.Fatalf("offsets %v, want %v", offsets, want)
	}
	if got := code[offsets[3]:offsets[4]]; got != "x" {
		t.Fatalf("unit 3 slices to %q, want x", got)
	}
}

// TestCorpusVerdictComparesPlaces plants the shape a text-only comparison read as agreement, and its
// twin.
//
// The probe reports one occurrence of `x` in `let x; x = 0;`. A row whose ESLint finding is the other
// `x` has the same text and a different place, and must not agree: it is a span gap. Its twin names
// the same `x` and must agree. Behind a leading blank line, which the corpus no longer trims, both
// still hold. Behind a byte order mark, which the compiler's reader drops, ESLint's offset on the first
// line is already past the mark and one on a later line is not, so both lines are planted.
func TestCorpusVerdictComparesPlaces(t *testing.T) {
	t.Parallel()

	probe := func(occurrence int) rule.Registration {
		return rule.Registration{Rule: rule.Rule{
			Name: "corpus-place-probe",
			Run: func(ctx rule.Context, options any) rule.Listeners {
				seen := 0
				return rule.Listeners{
					ast.KindIdentifier: func(node *ast.Node) {
						if node.Text() != "x" {
							return
						}
						seen++
						if seen == occurrence {
							ctx.ReportNode(node, rule.Message{Id: "probe", Description: "One x."})
						}
					},
				}
			},
		}}
	}

	cases := []struct {
		name       string
		code       string
		occurrence int
		start      float64
		want       string
	}{
		{"same text, another place", "let x; x = 0;", 1, 7, "span"},
		{"same text, the same place", "let x; x = 0;", 1, 4, "agree"},
		{"behind a blank line, another place", "\nlet x; x = 0;", 1, 8, "span"},
		{"behind a blank line, the same place", "\nlet x; x = 0;", 1, 5, "agree"},
		{"behind a mark, the first line", "\ufefflet x; x = 0;", 1, 4, "agree"},
		{"behind a mark, the first line, another place", "\ufefflet x; x = 0;", 1, 7, "span"},
		{"behind a mark, a later line", "\ufefflet x;\nx = 0;", 2, 8, "agree"},
		{"behind a mark, a later line, another place", "\ufefflet x;\nx = 0;", 1, 8, "span"},
	}
	for _, testCase := range cases {
		row := eslintCorpusRow{Code: testCase.code, ESLint: [][3]any{{"probe", testCase.start, testCase.start + 1}}}
		if verdict := corpusVerdict(t, probe(testCase.occurrence), row); verdict != testCase.want {
			t.Errorf("%s: the verdict is %q, want %q", testCase.name, verdict, testCase.want)
		}
	}
}
