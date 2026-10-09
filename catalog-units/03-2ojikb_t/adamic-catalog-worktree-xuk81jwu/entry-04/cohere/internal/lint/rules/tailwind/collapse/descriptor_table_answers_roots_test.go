package tailwind

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"testing"
)

type wave3Fixture struct {
	RootCount int `json:"rootCount"`
	CaseCount int `json:"caseCount"`
	Answered  int `json:"answered"`
	Cases     []struct {
		Root         string `json:"root"`
		ClassName    string `json:"className"`
		FromRegistry bool   `json:"fromRegistry"`
		Reading      *struct {
			Order []int `json:"order"`
			Count int   `json:"count"`
		} `json:"reading"`
	} `json:"cases"`
}

// TestWave3RootsAreAlreadyAnsweredByTheDescriptorTable is the measurement that ended wave 3.
//
// The 26 direct `utilities.functional` registrations were scoped as the last porting wave: the color
// families, the shadows, the transforms, `flex`, `bg-linear`. Seventeen of them read differently
// depending on their value, which is what made them look like the hard case.
//
// They are the case the descriptor model was built for. A root's reading being a pure function of its
// value's resolved type is exactly what `Table.Lookup` answers, and Phase 0 measured that model on
// 37,643 registry classes before any of this port existed. So these roots needed no new table, and
// this test is the evidence rather than a claim: every one of them is compared against the engine
// through the existing lookup.
//
// Kept as a standing test rather than deleted after the measurement, because the interesting failure
// is the future one. If a wave-4 change or a Tailwind bump makes one of these roots stop being a pure
// function of its type, this is where it shows up, and the answer then is a new table rather than a
// wider descriptor.
func TestWave3RootsAreAlreadyAnsweredByTheDescriptorTable(t *testing.T) {
	t.Parallel()
	fixture := loadWave3Fixture(t, "wave3_fixtures.json")
	system := loadWave1DesignSystem(t)
	table := NewTable(system)

	var agreed, bothDeclined, tableSilent, inventedOnRegistry, inventedOnProbe int
	var disagreements []string

	for _, testCase := range fixture.Cases {
		candidates := ParseCandidate(testCase.ClassName, system)
		var reading Reading
		answered := false
		if len(candidates) > 0 {
			reading, answered = table.Lookup(&candidates[0])
		}

		switch {
		case testCase.Reading == nil && !answered:
			bothDeclined++
		case testCase.Reading == nil && answered:
			// An invented reading, split by whether anyone can write the class.
			//
			// `filter-red-500` and `transform-2xl` are shapes this fixture generates and the registry
			// does not advertise, so a table answering them costs nothing and reports nothing. A
			// registry class in the same state is a real invented reading. Folding the two together
			// buries the second under 418 of the first.
			if testCase.FromRegistry {
				inventedOnRegistry++
				disagreements = append(disagreements, testCase.ClassName+": registry class the engine declines, table read "+formatReading(reading))
			} else {
				inventedOnProbe++
			}
		case testCase.Reading != nil && !answered:
			tableSilent++
			disagreements = append(disagreements, testCase.ClassName+": engine reads "+formatReading(Reading{Order: testCase.Reading.Order, Count: testCase.Reading.Count})+", table declined")
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
	t.Logf("wave 3: %d roots, %d cases, agreed %d, both-declined %d, table-silent %d, invented on registry %d, invented on probe %d",
		fixture.RootCount, fixture.CaseCount, agreed, bothDeclined, tableSilent, inventedOnRegistry, inventedOnProbe)

	if agreed < 7000 {
		t.Errorf("only %d of %d cases agreed; a lookup that stopped answering scores perfectly on any comparison counting disagreements alone", agreed, fixture.Answered)
	}
	if inventedOnRegistry > 0 {
		t.Errorf("%d registry classes are answered by the table and declined by the engine", inventedOnRegistry)
	}
	for index, disagreement := range disagreements {
		if index >= 20 {
			t.Errorf("... and %d more", len(disagreements)-20)
			break
		}
		t.Error(disagreement)
	}
}

// TestWave3ValueDependentRootsReadByType pins the property that made a new table unnecessary.
//
// Seventeen of these roots answer differently for a color than for a length, and the descriptor model
// is what expresses that. Asserted on the pairs rather than on the count, because a count keeps
// passing while the pairs underneath it change.
func TestWave3ValueDependentRootsReadByType(t *testing.T) {
	t.Parallel()
	forEachEngineSystem(t, func(t *testing.T, system *LoadedDesignSystem) {
		table := NewTable(system)

		cases := []struct {
			colorClass  string
			otherClass  string
			description string
		}{
			{"ring-amber-500", "ring-2", "a ring color against a ring width"},
			{"text-amber-500", "text-2xl", "a text color against a font size"},
			{"stroke-amber-500", "stroke-2", "a stroke color against a stroke width"},
			{"shadow-amber-500", "shadow-lg", "a shadow color against a shadow size"},
		}

		for _, testCase := range cases {
			colorReading, colorFound := lookupClass(t, system, table, testCase.colorClass)
			otherReading, otherFound := lookupClass(t, system, table, testCase.otherClass)
			if !colorFound || !otherFound {
				t.Errorf("%s: one side did not resolve (%v, %v)", testCase.description, colorFound, otherFound)
				continue
			}
			if readingsEqual(colorReading, otherReading) {
				t.Errorf("%s: %s and %s read identically as %v#%d; the type partition is what lets one root answer both",
					testCase.description, testCase.colorClass, testCase.otherClass, colorReading.Order, colorReading.Count)
			}
		}
	})
}

func lookupClass(t *testing.T, system *LoadedDesignSystem, table *Table, className string) (Reading, bool) {
	t.Helper()
	candidates := ParseCandidate(className, system)
	if len(candidates) == 0 {
		return Reading{}, false
	}
	return table.Lookup(&candidates[0])
}

func loadWave3Fixture(t *testing.T, name string) wave3Fixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	var fixture wave3Fixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	if fixture.CaseCount == 0 || fixture.Answered == 0 {
		t.Fatalf("%s carries %d cases and %d answered; a comparison over it would pass by comparing nothing", name, fixture.CaseCount, fixture.Answered)
	}
	return fixture
}

// TestNamespacePrecedenceIsUnobservableInThisCorpus records why one mutation cannot be caught here.
//
// `bareReading` walks the theme namespaces longest-first, and the file comment calls that order part
// of the contract. Reversing it to shortest-first changes no answer, in wave 3 or in the full
// differential, and this is the measurement behind that rather than a shrug.
//
// The order is observable only when a key belongs to two namespaces AND some root consumes both. The
// second half nearly holds: `font` consumes `--font-weight` and `--font`, the only such root of the
// 84 namespaces here. The first half does not. `bold` and `medium` are keys of `--font-weight`
// alone, `sans` of `--font` alone, so the loop's membership guard admits exactly one namespace per
// key and the order it tries them in cannot matter.
//
// So the ordering is correct, unexercised, and worth keeping: a repository declaring `--font-bold`
// alongside the framework's `--font-weight-bold` would make it load-bearing immediately, and this
// test is what turns that from a silent behaviour change into a failure. It fails when the corpus
// gains an ambiguous key, which is the moment the mutation becomes possible and is owed.
func TestNamespacePrecedenceIsUnobservableInThisCorpus(t *testing.T) {
	t.Parallel()
	system := loadWave1DesignSystem(t)
	table := NewTable(system)

	ambiguous := map[string][]string{}
	for _, root := range []string{"font", "text", "shadow", "drop-shadow", "inset-shadow"} {
		descriptor, known := table.Descriptors[root]
		if !known {
			continue
		}
		consumed := []string{}
		for _, namespace := range table.Namespaces {
			if _, uses := descriptor.Absent.ByNamespace[namespace]; uses {
				consumed = append(consumed, namespace)
			}
		}
		for _, namespace := range consumed {
			for key := range table.KeysByNamespace[namespace] {
				owners := []string{}
				for _, other := range consumed {
					if table.KeysByNamespace[other][key] {
						owners = append(owners, other)
					}
				}
				if len(owners) > 1 {
					ambiguous[root+"/"+key] = owners
				}
			}
		}
	}

	if len(ambiguous) > 0 {
		t.Errorf("%d keys are now claimed by two namespaces a single root consumes, so namespace precedence is observable and the longest-first order needs a mutation covering it: %v",
			len(ambiguous), ambiguous)
	}
	t.Logf("no key is claimed by two namespaces of one root, so namespace order is unobservable in this corpus")
}
