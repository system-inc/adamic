package oracle

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"hash/fnv"
	"os"
	"os/exec"
	"path"
	"regexp"
	"slices"
	"strings"
	"testing"
	"time"
	"unicode"
)

// A test over every fixture is one unit to Loom however its subtests are cut, since Loom runs top-level tests, and
// one over a thousand fixtures is many minutes on one core. So each is a family of top-level tests, <name>_000 to
// <name>_<N-1>, and a case belongs to the shard the FNV-64a hash of its key names (a fixture's path), never to its
// place in a list, so registering a fixture moves no other. The wrappers are generated from this table:
// TestFamilyWrappersAreCurrent says when they aren't, and rewrites them with -update-family-wrappers.
//
// Each family is sized from its cases' one-core times (Workshop, Oct 11, warm runtime) so that its worst shard is at
// most about 10 s alone on one core, half the 20 s each test is allowed, since a loaded box stretches it. A case that
// is half that alone would make its shard twice the rest, so it is solo: its own test, <name>_<its file's name>.
var families = []family{
	// 730 s over 946 fixtures, the four solo ones 6.7 to 11.9 s each.
	{name: "TestNativeAgreesWithNode", shards: 128, runner: "runNativeAgreesShard", keys: fixturePaths, planted: "dedication/dedication.a", caught: "stdout differs", solo: []string{
		"internal/oracle/testdata/bitwise_sweep.a",
		"internal/oracle/testdata/normalize_coverage_limit.a",
		"internal/oracle/testdata/normalize_coverage_long.a",
		"internal/oracle/testdata/user_iterators.a",
	}},
	// 505 s: each fixture's wasm32 build, run on Node's WASI; at 64 the worst was 15.5 s.
	{name: "TestWASIAgreesWithNode", shards: 128, runner: "runWASIAgreesShard", keys: fixturePaths, planted: "dedication/dedication.a", caught: "stdout differs", requires: "ADAMIC_ORACLE_WASI", solo: []string{
		"internal/oracle/testdata/bitwise_sweep.a",
		"internal/oracle/testdata/normalize_coverage_limit.a",
	}},
	// 201 s: each fixture's C compiled for wasm32.
	{name: "TestWASIEmission", shards: 32, runner: "runWASIEmissionShard", keys: fixturePaths, planted: "dedication/dedication.a", caught: "planted = ;", requires: "ADAMIC_ORACLE_WASI"},
	// 56 s over 208 review programs, run one at a time within a shard as they always were; at 8 the worst was 15 s.
	{name: "TestReviewProgramsAgreeWithNode", shards: 12, runner: "runReviewAgreeShard", keys: reviewAgreePrograms, planted: "smoke.a", caught: "native: stdout differs"},
	// 236 s over about 1,030 rows of counts.md, each a counted build and run.
	{name: "TestCountsAreRecorded", shards: 32, runner: "runCountsShard", keys: countKeys, planted: "dedication/dedication.a", caught: "the counts changed"},
	// 92 s as one program of every loop, most of it clang on one huge C file.
	{name: "TestLoopCountersAgreeWithNode", shards: 8, runner: "runLoopCountersShard", keys: loopCounterKeys, planted: "let NAME = 0; NAME < values.length; NAME++\n", caught: "kept in an integer true, want false", oneProgram: true},
	// 22.7 s: 20 mutants of library_method_values.a, each built and run.
	{name: "TestLibraryMethodValueMutants", shards: 4, runner: "runLibraryMethodValueMutantsShard", keys: libraryMethodValueMutants, planted: "trim", caught: "mutant changed nothing"},
}

// A family is one test over many cases, run as shards. runner is the function each wrapper calls with its shard, keys
// lists every case's key in the order the cases run, and planted is a key the family's runner plants a mutant in when
// probed, which exactly its owner must catch, with caught in its failure. A family that runs only when requires is set
// to 1 skips otherwise. solo names the keys run each in a test of its own instead of a shard: the runner is given -1
// for the first, -2 for the second, and so on. A family whose shard builds its cases into one program, oneProgram, has
// no subtest per case.
type family struct {
	name       string
	shards     int
	runner     string
	keys       func() []string
	planted    string
	caught     string
	requires   string
	solo       []string
	oneProgram bool
}

// A familyTest is one test of a family as its wrapper runs it: its name, the shard its wrapper gives the runner, and
// the indices of the keys that shard runs.
type familyTest struct {
	name  string
	shard int
	cases []int
}

var updateFamilyWrappers = flag.Bool("update-family-wrappers", false, "rewrite "+familyWrappersPath+" from the families table")

const familyWrappersPath = "families_generated_test.go"

// familyProbeVariable, set to <family>=<key> in a child of TestFamiliesCatchAPlantedCase, narrows every shard of that
// family to the one case and has its runner plant the mutant there.
const familyProbeVariable = "ADAMIC_FAMILY_PROBE"

func familyShard(key string, shards int) int {
	hash := fnv.New64a()
	hash.Write([]byte(key))
	return int(hash.Sum64() % uint64(shards))
}

func familyNamed(t *testing.T, name string) family {
	t.Helper()
	for _, candidate := range families {
		if candidate.name == name {
			return candidate
		}
	}
	t.Fatalf("no family %s", name)
	return family{}
}

// familyProbe is the key the family's runner plants its mutant in, when this process is a probe of that family.
func familyProbe(name string) (string, bool) {
	probed, key, found := strings.Cut(os.Getenv(familyProbeVariable), "=")
	return key, found && probed == name
}

// familyCases is the indices of the keys a shard owns, in order: under a probe, only the planted key, in its owner.
// A negative shard is a solo test, -1 the family's first solo key.
func familyCases(t *testing.T, name string, shard int, keys []string) []int {
	t.Helper()
	family := familyNamed(t, name)
	planted, probing := familyProbe(name)
	var owned []int
	for index, key := range keys {
		if familyOwns(family, shard, key) && (!probing || key == planted) {
			owned = append(owned, index)
		}
	}
	return owned
}

// familyOwns says whether a shard runs a key: a solo key only in its own test, any other in the shard its hash names.
func familyOwns(family family, shard int, key string) bool {
	solo := slices.Index(family.solo, key)
	return solo < 0 && familyShard(key, family.shards) == shard || solo >= 0 && shard == -1-solo
}

// familyTests is a family's tests, as its wrappers are generated: <name>_000 to <name>_<N-1>, then each solo test,
// each with the keys familyOwns gives its shard.
func familyTests(family family, keys []string) []familyTest {
	var tests []familyTest
	for shard := 0; shard < family.shards; shard++ {
		tests = append(tests, familyTest{name: fmt.Sprintf("%s_%03d", family.name, shard), shard: shard})
	}
	for solo, key := range family.solo {
		tests = append(tests, familyTest{name: familyTestName(family, key), shard: -1 - solo})
	}
	for test := range tests {
		for index, key := range keys {
			if familyOwns(family, tests[test].shard, key) {
				tests[test].cases = append(tests[test].cases, index)
			}
		}
	}
	return tests
}

// familyTestName is the test that owns a key: its solo test, or the shard its hash names.
func familyTestName(family family, key string) string {
	if solo := slices.Index(family.solo, key); solo >= 0 {
		return family.name + "_" + familySoloName(key)
	}
	return fmt.Sprintf("%s_%03d", family.name, familyShard(key, family.shards))
}

// familySoloName is a solo key's file name, without its extension, as a Go identifier.
func familySoloName(key string) string {
	name := strings.TrimSuffix(path.Base(key), path.Ext(key))
	return strings.Map(func(character rune) rune {
		if character < 128 && (unicode.IsLetter(character) || unicode.IsDigit(character)) {
			return character
		}
		return '_'
	}, name)
}

// familyWrappers is the text of the wrappers file the families table makes.
func familyWrappers(t *testing.T) []byte {
	t.Helper()
	var text bytes.Buffer
	fmt.Fprintf(&text, "// Code generated by TestFamilyWrappersAreCurrent from the families table (families_test.go); DO NOT EDIT.\n")
	fmt.Fprintf(&text, "// After changing the table: go test ./internal/oracle -run '^TestFamilyWrappersAreCurrent$' -args -update-family-wrappers\n\n")
	fmt.Fprintf(&text, "package oracle\n\nimport \"testing\"\n")
	for _, family := range families {
		for _, test := range familyTests(family, nil) {
			fmt.Fprintf(&text, "\nfunc %s(t *testing.T) {\n\tt.Parallel()\n\t%s(t, %d)\n}\n", test.name, family.runner, test.shard)
		}
	}
	formatted, err := format.Source(text.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	return formatted
}

func TestFamilyWrappersAreCurrent(t *testing.T) {
	t.Parallel()
	want := familyWrappers(t)
	if *updateFamilyWrappers {
		if err := os.WriteFile(familyWrappersPath, want, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	got, err := os.ReadFile(familyWrappersPath)
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("%s doesn't match the families table (%v): go test ./internal/oracle -run '^TestFamilyWrappersAreCurrent$' -args -update-family-wrappers", familyWrappersPath, err)
	}
}

// A familyFault is a way a family's tests fail to run every case exactly once, in exactly one test.
type familyFault int

const (
	// A test the family names, a shard or a solo key's, isn't among its tests.
	familyTestMissing familyFault = iota + 1
	// A case two tests run, or a key registered twice, which Go would run as two subtests, the second named #01.
	familyCaseRepeated
	// A case no test runs.
	familyCaseOmitted
	// A solo key that isn't one of the cases, or whose test's name is another's, or is all digits as a shard's is.
	familySoloMisnamed
)

// A familyError is one fault in a family's tests, and what it found.
type familyError struct {
	fault   familyFault
	message string
}

func (failure familyError) Error() string { return failure.message }

func (fault familyFault) String() string {
	return [...]string{"", "a test missing", "a case repeated", "a case omitted", "a solo misnamed"}[fault]
}

// familyShardLike is a solo name a shard's could be: all digits, as 007 is in <name>_007, or nothing.
var familyShardLike = regexp.MustCompile(`^[0-9]*$`)

// familyOwnership is every fault in how tests, each with the cases it runs, run a family's keys: every key unique,
// every test the family names there, every case run by exactly one of them, and each solo key one of the cases with a
// test name of its own that no shard's has or could have.
func familyOwnership(family family, keys []string, tests []familyTest) []familyError {
	var faults []familyError
	fault := func(kind familyFault, format string, arguments ...any) {
		faults = append(faults, familyError{fault: kind, message: family.name + ": " + fmt.Sprintf(format, arguments...)})
	}
	seen := map[string]bool{}
	for _, key := range keys {
		if seen[key] {
			fault(familyCaseRepeated, "%s is registered twice", key)
		}
		seen[key] = true
	}
	names := map[string]bool{}
	for _, key := range family.solo {
		name := familySoloName(key)
		if !seen[key] || names[name] || familyShardLike.MatchString(name) {
			fault(familySoloMisnamed, "solo %s must be one of its cases, named once, with a test name of its own that no shard's could be (%s)", key, familyTestName(family, key))
		}
		names[name] = true
	}
	present := map[string]bool{}
	owners := make([][]string, len(keys))
	for _, test := range tests {
		present[test.name] = true
		for _, index := range test.cases {
			owners[index] = append(owners[index], test.name)
		}
	}
	for _, test := range familyTests(family, nil) {
		if !present[test.name] {
			fault(familyTestMissing, "no %s among its tests", test.name)
		}
	}
	for index, tests := range owners {
		switch {
		case len(tests) == 0:
			fault(familyCaseOmitted, "%s is run by no test", keys[index])
		case len(tests) > 1:
			fault(familyCaseRepeated, "%s is run by %s", keys[index], strings.Join(tests, " and "))
		}
	}
	return faults
}

// Every case of every family has one owner: each test the family names runs the cases familyOwns gives its shard, and
// together they run every key exactly once (familyOwnership). A probe left in the environment would thin every shard
// to one case, so a real run refuses it.
func TestFamiliesOwnEveryCaseOnce(t *testing.T) {
	t.Parallel()
	if value := os.Getenv(familyProbeVariable); value != "" {
		t.Fatalf("%s=%s is set outside a probe, which would run one case per family", familyProbeVariable, value)
	}
	for _, family := range families {
		keys := family.keys()
		if len(keys) == 0 {
			t.Errorf("%s has no cases", family.name)
		}
		if !slices.Contains(keys, family.planted) {
			t.Errorf("%s: its planted case %s isn't one of its cases", family.name, family.planted)
		}
		tests := familyTests(family, keys)
		for _, failure := range familyOwnership(family, keys, tests) {
			t.Error(failure)
		}
		smallest, largest := len(keys), 0
		for _, test := range tests[:family.shards] {
			smallest, largest = min(smallest, len(test.cases)), max(largest, len(test.cases))
		}
		t.Logf("%s: %d cases, each owned once, over %d shards of %d to %d and %d solo tests", family.name, len(keys), family.shards, smallest, largest, len(family.solo))
	}
}

// The ownership check has to be able to fail, as wasi_shards_test.go's union check once proved: a family's tests with
// a shard dropped, with one case run by two of them, or with one case run by none, and solo keys named as a shard's
// test could be, are each refused, for that fault.
func TestFamilyOwnershipRefusesWhatItShould(t *testing.T) {
	t.Parallel()
	planted := family{name: "TestPlanted", shards: 4, solo: []string{"heavy/solo.a"}}
	var keys []string
	for index := range 24 {
		keys = append(keys, fmt.Sprintf("case/%03d.a", index))
	}
	keys = append(keys, "heavy/solo.a")
	tests := familyTests(planted, keys)
	if faults := familyOwnership(planted, keys, tests); len(faults) != 0 {
		t.Fatalf("the planted family's own tests must own every case once: %v", faults)
	}
	// The shard the first key belongs to, which therefore runs at least one case.
	owner := familyShard(keys[0], planted.shards)
	copied := func() []familyTest {
		copies := make([]familyTest, len(tests))
		for index, test := range tests {
			copies[index] = familyTest{name: test.name, shard: test.shard, cases: slices.Clone(test.cases)}
		}
		return copies
	}
	missing := slices.Delete(copied(), owner, owner+1)
	repeated := copied()
	other := (owner + 1) % planted.shards
	repeated[other].cases = append(repeated[other].cases, 0)
	omitted := copied()
	omitted[owner].cases = slices.DeleteFunc(omitted[owner].cases, func(index int) bool { return index == 0 })
	twice := append(slices.Clone(keys), keys[0])
	// One solo test named as shard 3's is, one named as a shard that doesn't exist would be.
	shardLike := planted
	shardLike.solo = []string{"case/003.a", "case/017.a", "heavy/solo.a"}
	cases := []struct {
		name   string
		family family
		keys   []string
		tests  []familyTest
		want   []familyFault
	}{
		{"a shard missing", planted, keys, missing, []familyFault{familyTestMissing, familyCaseOmitted}},
		{"a case run twice", planted, keys, repeated, []familyFault{familyCaseRepeated}},
		{"a case run by none", planted, keys, omitted, []familyFault{familyCaseOmitted}},
		{"a key registered twice", planted, twice, familyTests(planted, twice), []familyFault{familyCaseRepeated}},
		{"solo names like a shard's", shardLike, keys, familyTests(shardLike, keys), []familyFault{familySoloMisnamed}},
	}
	for _, each := range cases {
		t.Run(each.name, func(t *testing.T) {
			faults := familyOwnership(each.family, each.keys, each.tests)
			var got []familyFault
			for _, failure := range faults {
				if !slices.Contains(got, failure.fault) {
					got = append(got, failure.fault)
				}
			}
			slices.Sort(got)
			if !slices.Equal(got, each.want) {
				t.Fatalf("faults %v, want exactly the kinds %v: %v", got, each.want, faults)
			}
			t.Logf("refused: %v", faults)
		})
	}
}

// A mutant planted in one case must fail exactly the shard that owns it, through the real wrappers: a child process runs
// every wrapper of the family with the probe set, so each shard but the owner has nothing to run, and the owner runs
// the planted case with its runner's mutant.
func TestFamiliesCatchAPlantedCase(t *testing.T) {
	t.Parallel()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, family := range families {
		t.Run(family.name, func(t *testing.T) {
			t.Parallel()
			if family.requires != "" && os.Getenv(family.requires) != "1" {
				t.Skipf("set %s=1 to run %s", family.requires, family.name)
			}
			owner := familyTestName(family, family.planted)
			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()
			command := exec.CommandContext(ctx, executable, "-test.run=^"+family.name+"_", "-test.v", "-test.parallel=4", "-test.timeout=90s")
			command.Env = append(os.Environ(), familyProbeVariable+"="+family.name+"="+family.planted)
			output, err := command.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("the probe of %s was killed at 90 s:\n%s", family.name, output)
			}
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatalf("the probe of %s passed, want its planted case caught: %v\n%s", family.name, err, output)
			}
			failed := regexp.MustCompile(`(?m)^--- FAIL: (`+regexp.QuoteMeta(family.name)+`_\w+) `).FindAllSubmatch(output, -1)
			caseFailed := family.oneProgram || strings.Contains(string(output), "--- FAIL: "+owner+"/"+family.planted+" ")
			if len(failed) != 1 || string(failed[0][1]) != owner || !caseFailed {
				t.Fatalf("want exactly %s to fail, on %s; failed: %q\n%s", owner, family.planted, failed, output)
			}
			// Failing for another reason, a build that broke or a runner that crashed, isn't catching the mutant.
			if !strings.Contains(string(output), family.caught) {
				t.Fatalf("%s failed, but not with %q, the failure its planted mutant makes:\n%s", owner, family.caught, output)
			}
			t.Logf("the mutant planted in %s failed exactly %s, with %q", family.planted, owner, family.caught)
		})
	}
}

// fixturePaths is every fixture's path, in the order fixtures runs them.
func fixturePaths() []string {
	paths := make([]string, len(fixtures))
	for index, fixture := range fixtures {
		paths[index] = fixture.path
	}
	return paths
}
