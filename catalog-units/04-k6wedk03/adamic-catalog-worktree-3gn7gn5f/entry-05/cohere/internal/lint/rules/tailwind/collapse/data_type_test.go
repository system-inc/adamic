package tailwind

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// The fixture is what the shipped Tailwind 4.3.3 engine answered for every shape in the corpus,
// captured by internal/lint/rules/tailwind/tools/generate_data_type and checked in next to this test.
//
// Measured rather than transcribed, because reading the source and asking the engine are different
// claims and they disagree in ways a careful reader would not predict. Two examples that this suite
// would have gotten wrong if the expectations had been hand-written from infer-data-type.ts:
// `serif` answers family-name rather than generic-name, and `1 2 3` answers line-width rather than
// vector. Both because an earlier type in the list also matches, which is the first-match-wins
// contract stated as data.

type dataTypeCorpus struct {
	TailwindVersion string `json:"tailwindVersion"`
	// Chunk records which bundled file the predicates were read out of, so a fixture regenerated
	// against a different Tailwind build is visible in the diff.
	Chunk                     string         `json:"chunk"`
	IsLengthResolved          bool           `json:"isLengthResolved"`
	IsPositiveIntegerResolved bool           `json:"isPositiveIntegerResolved"`
	SegmentResolved           bool           `json:"segmentResolved"`
	Types                     []DataType     `json:"types"`
	ReversedTypes             []DataType     `json:"reversedTypes"`
	Cases                     []dataTypeCase `json:"cases"`
}

// dataTypeCase is one value and everything the engine said about it.
type dataTypeCase struct {
	Value string `json:"value"`
	// All is the answer with the full type list in declaration order: what a caller passing
	// everything sees.
	All DataType `json:"all"`
	// Reversed is the answer with that list reversed. It exists to catch a port that iterates its
	// own fixed order rather than the caller's, which would agree with All on every value and
	// disagree here on every value matching more than one type.
	Reversed DataType `json:"reversed"`
	// PerType isolates each predicate, so a wrong check cannot hide behind an earlier type that
	// also matched.
	PerType map[DataType]DataType `json:"perType"`
	// IsLength and IsPositiveInteger are the two exported helpers, asked directly.
	IsLength          bool `json:"isLength"`
	IsPositiveInteger bool `json:"isPositiveInteger"`
	// SegmentComma and SegmentSpace are the engine's own splitter, recorded so that segment can be
	// tested directly rather than only through the answers it feeds.
	SegmentComma []string `json:"segmentComma"`
	SegmentSpace []string `json:"segmentSpace"`
}

func loadDataTypeCorpus(t *testing.T) dataTypeCorpus {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "datatype_fixtures.json"))
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var corpus dataTypeCorpus
	if err := json.Unmarshal(raw, &corpus); err != nil {
		t.Fatalf("parse fixtures: %v", err)
	}
	// A corpus that quietly shrank is the failure this approach exists to prevent, and it would
	// otherwise read as a passing run. The task this port was written for set the bar explicitly:
	// a suite that ran 12 cases has not tested this.
	if len(corpus.Cases) < 400 {
		t.Fatalf("fixture holds only %d cases; expected several hundred shapes", len(corpus.Cases))
	}
	if !corpus.SegmentResolved {
		t.Fatal("fixture was generated without resolving the engine's segment; its segment columns are not measurements")
	}
	if !corpus.IsLengthResolved || !corpus.IsPositiveIntegerResolved {
		t.Fatalf("fixture was generated without resolving both helper predicates; its isLength/isPositiveInteger columns are not measurements")
	}
	if len(corpus.Types) != 17 {
		t.Fatalf("fixture carries %d types; the engine declares 17", len(corpus.Types))
	}
	return corpus
}

// TestInferDataTypeMatchesEngine is the whole point of this package's test suite: every answer, for
// every shape, compared against what the shipped engine said.
func TestInferDataTypeMatchesEngine(t *testing.T) {
	t.Parallel()
	corpus := loadDataTypeCorpus(t)

	comparisons := 0
	for _, testCase := range corpus.Cases {
		if got := InferDataType(testCase.Value, corpus.Types); got != testCase.All {
			t.Errorf("InferDataType(%q, allTypes) = %q, engine says %q", testCase.Value, got, testCase.All)
		}
		comparisons++

		if got := InferDataType(testCase.Value, corpus.ReversedTypes); got != testCase.Reversed {
			t.Errorf("InferDataType(%q, reversedTypes) = %q, engine says %q", testCase.Value, got, testCase.Reversed)
		}
		comparisons++

		for _, dataType := range corpus.Types {
			want := testCase.PerType[dataType]
			if got := InferDataType(testCase.Value, []DataType{dataType}); got != want {
				t.Errorf("InferDataType(%q, [%q]) = %q, engine says %q", testCase.Value, dataType, got, want)
			}
			comparisons++
		}
	}

	t.Logf("compared %d answers across %d distinct value shapes against Tailwind %s (%s)",
		comparisons, len(corpus.Cases), corpus.TailwindVersion, corpus.Chunk)
}

// TestExportedPredicatesMatchEngine covers IsLength and IsPositiveInteger directly, since callers
// reach for them without going through a type list.
func TestExportedPredicatesMatchEngine(t *testing.T) {
	t.Parallel()
	corpus := loadDataTypeCorpus(t)

	for _, testCase := range corpus.Cases {
		if got := IsLength(testCase.Value); got != testCase.IsLength {
			t.Errorf("IsLength(%q) = %v, engine says %v", testCase.Value, got, testCase.IsLength)
		}
		if got := IsPositiveInteger(testCase.Value); got != testCase.IsPositiveInteger {
			t.Errorf("IsPositiveInteger(%q) = %v, engine says %v", testCase.Value, got, testCase.IsPositiveInteger)
		}
	}
}

// TestVarShortCircuitBeatsEveryCheck states the first of the two load-bearing behaviors as its own
// test rather than trusting it to fall out of the corpus.
//
// A `var()` value returns no type even when a check would plainly match it. `var(--a)` contains no
// math function and is not a color, so most checks decline anyway; the ones that matter are the
// permissive ones. `family-name` accepts nearly anything, and without the short-circuit every
// `var()` value in the codebase would read as a font family.
func TestVarShortCircuitBeatsEveryCheck(t *testing.T) {
	t.Parallel()
	permissive := []DataType{DataTypeFamilyName, DataTypeLineWidth, DataTypePosition, DataTypeBackgroundSize}

	for _, value := range []string{"var(--a)", "var(--a, red)", "var(--a, 10px)", "var("} {
		if got := InferDataType(value, AllDataTypes()); got != "" {
			t.Errorf("InferDataType(%q, all) = %q, want no type", value, got)
		}
		for _, dataType := range permissive {
			if got := InferDataType(value, []DataType{dataType}); got != "" {
				t.Errorf("InferDataType(%q, [%q]) = %q, want no type", value, dataType, got)
			}
		}
	}

	// The short-circuit is on the prefix, not on containment, so a value merely holding a var() is
	// still typed normally.
	if got := InferDataType("10px var(--a)", []DataType{DataTypePosition}); got != DataTypePosition {
		t.Errorf("InferDataType(\"10px var(--a)\", [position]) = %q, want position", got)
	}
}

// TestFirstMatchWinsIsTheCallersOrder states the second load-bearing behavior.
//
// Each value below matches at least two types, and the answer follows whichever the caller listed
// first. A port with its own fixed precedence would pass one column of this table and fail the
// other.
func TestFirstMatchWinsIsTheCallersOrder(t *testing.T) {
	t.Parallel()
	cases := []struct {
		value  string
		first  DataType
		second DataType
	}{
		{"1 2 3", DataTypeLineWidth, DataTypeVector},
		{"serif", DataTypeFamilyName, DataTypeGenericName},
		{"medium", DataTypeLineWidth, DataTypeAbsoluteSize},
		{"10px", DataTypeLength, DataTypeLineWidth},
		{"50%", DataTypePercentage, DataTypePosition},
		{"12", DataTypeNumber, DataTypeInteger},
		{"cover", DataTypeBackgroundSize, DataTypeFamilyName},
		{"calc(1px + 2px)", DataTypeLength, DataTypeNumber},
		{"url(a.png)", DataTypeURL, DataTypeImage},
	}

	for _, testCase := range cases {
		forward := []DataType{testCase.first, testCase.second}
		backward := []DataType{testCase.second, testCase.first}

		if got := InferDataType(testCase.value, forward); got != testCase.first {
			t.Errorf("InferDataType(%q, [%q %q]) = %q, want %q",
				testCase.value, testCase.first, testCase.second, got, testCase.first)
		}
		if got := InferDataType(testCase.value, backward); got != testCase.second {
			t.Errorf("InferDataType(%q, [%q %q]) = %q, want %q",
				testCase.value, testCase.second, testCase.first, got, testCase.second)
		}
	}
}

// TestUnknownTypeMatchesNothing covers the `checks[type]?.(value)` fallthrough: a type the table
// does not know declines rather than panicking. The generated descriptor tables carry type names
// from Tailwind's source, so a version that adds one must degrade, not crash.
func TestUnknownTypeMatchesNothing(t *testing.T) {
	t.Parallel()
	if got := InferDataType("10px", []DataType{"future-type"}); got != "" {
		t.Errorf("InferDataType with unknown type = %q, want no type", got)
	}
	if got := InferDataType("10px", []DataType{"future-type", DataTypeLength}); got != DataTypeLength {
		t.Errorf("unknown type should be skipped, got %q", got)
	}
}

// TestAllDataTypesIsTheEnginesOrder guards the convenience list against drift, since first-match-wins
// makes its order behavior rather than presentation.
func TestAllDataTypesIsTheEnginesOrder(t *testing.T) {
	t.Parallel()
	corpus := loadDataTypeCorpus(t)
	all := AllDataTypes()

	if len(all) != len(corpus.Types) {
		t.Fatalf("AllDataTypes has %d entries, engine declares %d", len(all), len(corpus.Types))
	}
	for index, dataType := range corpus.Types {
		if all[index] != dataType {
			t.Errorf("AllDataTypes()[%d] = %q, engine declares %q", index, all[index], dataType)
		}
	}

	// A caller mutating the returned slice must not change what the next caller sees.
	all[0] = "mutated"
	if AllDataTypes()[0] != DataTypeColor {
		t.Error("AllDataTypes returns shared state; a caller mutated it")
	}
}

// TestSegmentMatchesEngine compares the splitter directly against the engine's own.
//
// This test exists because of a measured gap rather than for completeness. Breaking bracket nesting
// in segment — making `[` stop pushing a closer, so `[a,b]` splits into two parts — changes the
// split of 15 corpus values and changes the inferred type of none of them, because the bracket
// fragments are nonsense to every predicate whether they arrive as one part or two. A mutation that
// real code would notice slipped through a suite that only ever asked segment questions through
// InferDataType.
//
// The type answers are what production consumes, so they stay the primary check. But segment is
// about to have other callers in this port — the candidate parser and the value parser both need
// it — and a splitter verified only by the one consumer that cannot see its mistakes is verified
// for that consumer alone.
func TestSegmentMatchesEngine(t *testing.T) {
	t.Parallel()
	corpus := loadDataTypeCorpus(t)

	comparisons := 0
	for _, testCase := range corpus.Cases {
		for _, separator := range []struct {
			character byte
			want      []string
		}{
			{',', testCase.SegmentComma},
			{' ', testCase.SegmentSpace},
		} {
			got := segment(testCase.Value, separator.character)
			if len(got) != len(separator.want) {
				t.Errorf("segment(%q, %q) = %q (%d parts), engine says %q (%d parts)",
					testCase.Value, string(separator.character), got, len(got), separator.want, len(separator.want))
				continue
			}
			for index := range got {
				if got[index] != separator.want[index] {
					t.Errorf("segment(%q, %q)[%d] = %q, engine says %q",
						testCase.Value, string(separator.character), index, got[index], separator.want[index])
				}
			}
			comparisons++
		}
	}

	t.Logf("compared %d segment splits across %d distinct value shapes", comparisons, len(corpus.Cases))
}

// TestImageVarSkipIsUnreachableThroughInferDataType records a measured limit of the differential
// suite, and covers the branch the suite cannot reach.
//
// isImage skips `var(` parts without counting them, and a mutation that counts them instead survives
// every one of the 9,386 differential comparisons. That is not a gap in the corpus; it is provably
// unobservable through InferDataType. Counting a var() part can only raise `count`, and `count > 0`
// is already satisfied whenever any real image part is present, so the two spellings differ only on
// a value whose parts are all `var(...)`. Such a value necessarily starts with `var(` and is
// short-circuited before isImage ever runs.
//
// Which leaves the branch correct, unreachable from the public entry point, and untested unless the
// predicate is called directly. It is called directly here, because "no caller can see it today" is
// a fact about today's callers.
func TestImageVarSkipIsUnreachableThroughInferDataType(t *testing.T) {
	t.Parallel()
	// Directly: a value of nothing but var() parts is not an image, because the skips never count.
	for _, value := range []string{"var(--a)", "var(--a),var(--b)", "var(--a),var(--b),var(--c)"} {
		if isImage(value) {
			t.Errorf("isImage(%q) = true, want false: var() parts are skipped, not counted", value)
		}
		// And through the public entry point the same value is short-circuited, for a different
		// reason, which is why the two must both be checked.
		if got := InferDataType(value, []DataType{DataTypeImage}); got != "" {
			t.Errorf("InferDataType(%q, [image]) = %q, want no type", value, got)
		}
	}

	// A real image part alongside a skipped var() part still counts, with no leading space, since
	// segment does not trim and ` var(` fails the prefix test.
	for _, value := range []string{"url(a.png),var(--b)", "linear-gradient(red,blue),var(--x)"} {
		if !isImage(value) {
			t.Errorf("isImage(%q) = false, want true", value)
		}
	}
	// With the space, the part is not var-prefixed, so it falls through to the reject.
	if isImage("url(a.png), var(--b)") {
		t.Error(`isImage("url(a.png), var(--b)") = true, want false: " var(" is not a var() part`)
	}
}
