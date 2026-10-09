package tailwind

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type wave2bFixture struct {
	RootCount int `json:"rootCount"`
	CaseCount int `json:"caseCount"`
	Answered  int `json:"answered"`
	Rejected  int `json:"rejected"`
	Cases     []struct {
		Root      string `json:"root"`
		ClassName string `json:"className"`
		Reading   *struct {
			Order []int `json:"order"`
			Count int   `json:"count"`
		} `json:"reading"`
	} `json:"cases"`
}

// TestFrameworkMultiDeclarationUtilitiesMatchTheEngine is the wave-2b acceptance test.
//
// Scored as a lattice for the same reason wave 1 is: mutual silence is counted apart from agreement,
// because a table that stopped answering scores perfectly on any comparison that folds the two.
func TestFrameworkMultiDeclarationUtilitiesMatchTheEngine(t *testing.T) {
	t.Parallel()
	system := loadWave1DesignSystem(t)
	for _, name := range []string{"wave2b_fixtures.json", "wave2c_fixtures.json", "wave4_fixtures.json", "wave5_fixtures.json"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			runMultiDeclarationComparison(t, loadWave2bFixture(t, name), system)
		})
	}
}

func runMultiDeclarationComparison(t *testing.T, fixture wave2bFixture, system *LoadedDesignSystem) {

	var agreed, bothDeclined, tableSilent, engineSilent int
	var disagreements, silentNames []string

	for _, testCase := range fixture.Cases {
		utility, known := FrameworkMultiDeclarationUtilities[testCase.Root]
		if !known {
			t.Fatalf("%s is in the fixture and not in the table", testCase.Root)
		}

		candidates := ParseCandidate(testCase.ClassName, system)
		var reading Reading
		produced := false
		if len(candidates) > 0 {
			parsedRoot := candidates[0].Root
			negative := false
			if strings.HasPrefix(parsedRoot, "-") && utility.SupportsNegative {
				parsedRoot = parsedRoot[1:]
				negative = true
			}
			if parsedRoot == testCase.Root {
				reading, produced = utility.ReadingFor(testCase.Root, &candidates[0], system.Theme(), negative)
			}
		}

		switch {
		case testCase.Reading == nil && !produced:
			bothDeclined++
		case testCase.Reading == nil && produced:
			engineSilent++
			disagreements = append(disagreements, testCase.ClassName+": engine declined, table read "+formatReading(reading))
		case testCase.Reading != nil && !produced:
			tableSilent++
			silentNames = append(silentNames, testCase.ClassName)
		default:
			expected := Reading{Order: testCase.Reading.Order, Count: testCase.Reading.Count}
			if readingsEqual(expected, reading) {
				agreed++
				continue
			}
			disagreements = append(disagreements, testCase.ClassName+": engine "+formatReading(expected)+", table "+formatReading(reading))
		}
	}

	sort.Strings(disagreements)
	sort.Strings(silentNames)
	t.Logf("wave 2b: %d roots, %d cases, agreed %d, both-declined %d, table-silent %d, engine-silent %d, disagreed %d",
		fixture.RootCount, fixture.CaseCount, agreed, bothDeclined, tableSilent, engineSilent, len(disagreements))
	// A silence is only allowed when the static table answers the class instead. `content-none` is
	// the case: upstream registers it as its own `staticUtility` and `ParseCandidate` reads it as a
	// static, so the functional path is right to decline and M1's table is what answers it. Anything
	// else is a gap in this table wearing the same shape as a correct decline.
	for _, className := range silentNames {
		if _, found := FrameworkStaticDeclarations[className]; !found {
			t.Errorf("%s: the engine reads it, this table declines, and no static registration covers it", className)
		}
	}
	if len(silentNames) > 0 {
		t.Logf("table-silent, all covered by the static table: %v", silentNames)
	}

	if agreed < 200 {
		t.Errorf("only %d cases agreed of %d the engine answered; a table that stopped answering scores perfectly on any comparison counting only disagreements", agreed, fixture.Answered)
	}
	if bothDeclined < 100 {
		t.Errorf("only %d cases were declined by both sides; the rejection branches are under-exercised", bothDeclined)
	}
	for _, disagreement := range disagreements {
		t.Error(disagreement)
	}
}

// TestFrameworkMultiDeclarationWrapperContributesNothing pins the measurement that made this a table.
//
// Every one of these roots calls a properties wrapper that returns `atRoot([...])` holding a dozen
// `@property` blocks. PropertySort descends into rules and at-rules and not into at-root, so those
// are structurally invisible. `grayscale` reading two declarations rather than fifteen is the proof,
// and it is asserted because the whole table's counts rest on it.
func TestFrameworkMultiDeclarationWrapperContributesNothing(t *testing.T) {
	t.Parallel()
	grayscale, known := FrameworkMultiDeclarationUtilities["grayscale"]
	if !known {
		t.Fatal("grayscale is not in the table")
	}
	if grayscale.Reading.Count != 2 {
		t.Errorf("grayscale reads %d declarations; the engine reads 2, so the at-root wrapper is being counted", grayscale.Reading.Count)
	}

	// The at-root walk itself, so a change to PropertySort fails here rather than silently changing
	// every count in this table.
	sorted := PropertySort([]*Node{
		AtRoot(Declaration("--tw-probe-one", "1"), Declaration("--tw-probe-two", "2")),
		Declaration("filter", "none"),
	})
	if sorted.Count != 1 {
		t.Errorf("a tree of one at-root wrapper and one declaration counted %d; at-root contributes nothing, so it must be 1", sorted.Count)
	}
}

// TestFrameworkMultiDeclarationLiteralReadingsBeatTheTheme pins the one root with two readings.
//
// `ease-initial` reads `[]#1` where every other `ease-*` reads `[354]#2`, because `initial` is a
// literal the handler answers directly rather than a value it wraps. The literal is checked before
// resolution, so a repository declaring `--ease-initial` cannot shadow it, which is what the engine
// does.
func TestFrameworkMultiDeclarationLiteralReadingsBeatTheTheme(t *testing.T) {
	t.Parallel()
	utility, known := FrameworkMultiDeclarationUtilities["ease"]
	if !known {
		t.Fatal("ease is not in the table")
	}
	if len(utility.LiteralReadings) == 0 {
		t.Fatal("ease carries no literal readings; ease-initial reads differently from every other ease-* and needs one")
	}

	theme := NewTheme()
	theme.Add("--ease-initial", "cubic-bezier(0,0,1,1)", 0)

	candidate := ParsedCandidate{
		Kind:  ParsedCandidateKindFunctional,
		Root:  "ease",
		Value: &ParsedValue{Kind: ParsedValueKindNamed, Value: "initial"},
	}
	reading, produced := utility.ReadingFor("ease", &candidate, theme, false)
	if !produced {
		t.Fatal("ease-initial produced nothing")
	}
	if reading.Count != 1 || len(reading.Order) != 0 {
		t.Errorf("ease-initial read %v#%d against a theme declaring --ease-initial; the engine reads []#1, so the literal must be checked before the theme",
			reading.Order, reading.Count)
	}
}

func loadWave2bFixture(t *testing.T, name string) wave2bFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	var fixture wave2bFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	if fixture.CaseCount == 0 || fixture.Answered == 0 {
		t.Fatalf("%s carries %d cases and %d answered; a comparison over it would pass by comparing nothing", name, fixture.CaseCount, fixture.Answered)
	}
	return fixture
}
