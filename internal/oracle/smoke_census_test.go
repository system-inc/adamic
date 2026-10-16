package oracle

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// The gate's smoke list (cloud/fast-gate/smoke.txt) names fixtures every change must still see run: a plain path in
// TestNativeAgreesWithNode, or "Lane path" in that lane. Once each of the oracle's lanes is keyed by what it reads, the
// smoke phase only repeats them, so the list's promise moves here (Loom's ruling on #rmp2hdj): every entry is a
// registered case of its lane, and every case is owned by exactly one of the lane's tests (TestFamiliesOwnEveryCaseOnce),
// so each is run by an oracle unit. An entry that stops being registered is red here, so the coverage can't shrink in
// silence, as smoke's did when its lanes became families (75a273db37).
const smokeListPath = "cloud/fast-gate/smoke.txt"

// A smokeEntry is one line of the smoke list: the lane that runs a fixture, and the fixture's path.
type smokeEntry struct{ lane, path string }

// smokeLanes is each lane a smoke entry may name, with the keys its tests run.
func smokeLanes() map[string][]string {
	inputs := make([]string, len(inputFixtures))
	for index, fixture := range inputFixtures {
		inputs[index] = fixture.path
	}
	return map[string][]string{"TestNativeAgreesWithNode": fixturePaths(), "TestInputAgreesWithNode": inputs}
}

// parseSmokeList is the smoke list's entries, as run.py's smokeList reads them.
func parseSmokeList(text string) []smokeEntry {
	var entries []smokeEntry
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 2 {
			entries = append(entries, smokeEntry{fields[0], fields[1]})
		} else {
			entries = append(entries, smokeEntry{"TestNativeAgreesWithNode", strings.TrimSpace(line)})
		}
	}
	return entries
}

// smokeUnregistered is every entry whose lane isn't one of lanes, or whose path isn't one of its lane's keys.
func smokeUnregistered(entries []smokeEntry, lanes map[string][]string) []string {
	var faults []string
	for _, entry := range entries {
		keys, known := lanes[entry.lane]
		switch {
		case !known:
			faults = append(faults, entry.lane+" "+entry.path+": no such lane")
		case !slices.Contains(keys, entry.path):
			faults = append(faults, entry.lane+" "+entry.path+": not one of the lane's cases")
		}
	}
	return faults
}

func TestSmokeEntriesAreRegisteredCases(t *testing.T) {
	t.Parallel()
	text, err := os.ReadFile(filepath.Join(repository, smokeListPath))
	if err != nil {
		t.Fatal(err)
	}
	entries := parseSmokeList(string(text))
	if len(entries) == 0 {
		t.Fatalf("%s lists no entries", smokeListPath)
	}
	for _, fault := range smokeUnregistered(entries, smokeLanes()) {
		t.Errorf("%s: %s", smokeListPath, fault)
	}
	native := familyNamed(t, "TestNativeAgreesWithNode")
	for _, entry := range entries {
		owner := entry.lane
		if entry.lane == native.name {
			owner = familyTestName(native, entry.path)
		}
		t.Logf("%s runs %s", owner, entry.path)
	}
}

// The census has to be able to fail: an entry naming a lane that doesn't exist (TestWASIInputAgreesWithNode was one,
// a lane of a stack that never reached main), a path no lane registers, and a registered path in the wrong lane are
// each refused, and a list of registered entries is not.
func TestSmokeCensusRefusesWhatItShould(t *testing.T) {
	t.Parallel()
	lanes := map[string][]string{"TestNativeAgreesWithNode": {"a.a", "b.a"}, "TestInputAgreesWithNode": {"walk.a"}}
	registered := parseSmokeList("# a comment\na.a\n\nTestNativeAgreesWithNode b.a\nTestInputAgreesWithNode walk.a\n")
	if faults := smokeUnregistered(registered, lanes); len(registered) != 3 || len(faults) != 0 {
		t.Fatalf("registered entries %v: faults %v", registered, faults)
	}
	for _, each := range []struct{ name, list, fault string }{
		{"a lane that doesn't exist", "TestWASIInputAgreesWithNode walk.a\n", "no such lane"},
		{"a path no lane registers", "c.a\n", "not one of the lane's cases"},
		{"a path in the wrong lane", "TestInputAgreesWithNode a.a\n", "not one of the lane's cases"},
	} {
		faults := smokeUnregistered(parseSmokeList(each.list), lanes)
		if len(faults) != 1 || !strings.Contains(faults[0], each.fault) {
			t.Errorf("%s: faults %v, want one saying %q", each.name, faults, each.fault)
		}
	}
}
