package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// smokeList is the gate's smoke set, one fixture per feature family and each fixture a whole gate once caught: "fixture"
// for TestNativeAgreesWithNode, "Lane fixture" for another lane.
const smokeList = "cloud/fast-gate/smoke.txt"

// Every smoke entry is a fixture its lane runs, so oracle's own units run the whole smoke set whenever their key moves:
// a TestNativeAgreesWithNode entry is one of the family's keys, and the member that owns it is named; a
// TestInputAgreesWithNode entry is one of its input fixtures. This is the coverage that retired the gate's smoke phase,
// which compiled oracle's test binary in every unit to run these 46 (148.3 s for 179 passes, Planner's measurement on
// #zw6hb47), now held where the fixtures live: a fixture dropped from its lane, or an entry naming a lane oracle doesn't
// run, is red. Mutants that each fail it: a native fixture dropped from fixtures; an input fixture dropped; the input
// fixtures unread; an unknown lane taken; a fixture its family doesn't hold taken.
func TestEverySmokeEntryRunsInAnOracleUnit(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile(filepath.Join(repository, smokeList))
	if err != nil {
		t.Fatal(err)
	}
	runs := smokeRunner(t)
	entries := 0
	for _, line := range strings.Split(string(content), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || strings.HasPrefix(fields[0], "#") {
			continue
		}
		entries++
		lane, fixture := "TestNativeAgreesWithNode", fields[0]
		if len(fields) == 2 {
			lane, fixture = fields[0], fields[1]
		}
		if test, err := runs(lane, fixture); err != nil {
			t.Error(err)
		} else {
			t.Logf("%s runs in %s", fixture, test)
		}
	}
	if entries == 0 {
		t.Fatalf("%s holds no entry", smokeList)
	}
	// What no entry names today is still refused: a lane oracle doesn't run, and a fixture its family doesn't hold.
	for _, planted := range [][2]string{{"TestWASIInputAgreesWithNode", "internal/oracle/testdata/walk.a"},
		{"TestNativeAgreesWithNode", "internal/oracle/testdata/no-such-fixture.a"}, {"TestInputAgreesWithNode", "internal/oracle/testdata/bitwise.a"}} {
		if test, err := runs(planted[0], planted[1]); err == nil {
			t.Errorf("%s %s runs in %s, want it refused", planted[0], planted[1], test)
		}
	}
}

// smokeRunner says which of oracle's tests runs a smoke entry, the lane's family member that owns the fixture, or why
// none does.
func smokeRunner(t *testing.T) func(lane, fixture string) (string, error) {
	t.Helper()
	native, paths := familyNamed(t, "TestNativeAgreesWithNode"), fixturePaths()
	owner := map[string]string{}
	for _, test := range familyTests(native, paths) {
		for _, index := range test.cases {
			owner[paths[index]] = test.name
		}
	}
	inputs := []string{}
	for _, fixture := range inputFixtures {
		inputs = append(inputs, fixture.path)
	}
	return func(lane, fixture string) (string, error) {
		switch lane {
		case "TestNativeAgreesWithNode":
			if member, found := owner[fixture]; found {
				return member, nil
			}
			return "", fmt.Errorf("%s isn't a key of TestNativeAgreesWithNode's family, so no oracle unit runs it", fixture)
		case "TestInputAgreesWithNode":
			if slices.Contains(inputs, fixture) {
				return lane, nil
			}
			return "", fmt.Errorf("%s isn't one of TestInputAgreesWithNode's fixtures, so no oracle unit runs it", fixture)
		}
		return "", fmt.Errorf("%s names the lane %s, which oracle doesn't run", fixture, lane)
	}
}
