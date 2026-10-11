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
	{name: "TestNativeAgreesWithNode", shards: 128, runner: "runNativeAgreesShard", keys: fixturePaths, planted: "dedication/dedication.a", solo: []string{
		"internal/oracle/testdata/bitwise_sweep.a",
		"internal/oracle/testdata/normalize_coverage_limit.a",
		"internal/oracle/testdata/normalize_coverage_long.a",
		"internal/oracle/testdata/user_iterators.a",
	}},
	// 505 s: each fixture's wasm32 build, run on Node's WASI; at 64 the worst was 15.5 s.
	{name: "TestWASIAgreesWithNode", shards: 128, runner: "runWASIAgreesShard", keys: fixturePaths, planted: "dedication/dedication.a", requires: "ADAMIC_ORACLE_WASI", solo: []string{
		"internal/oracle/testdata/bitwise_sweep.a",
		"internal/oracle/testdata/normalize_coverage_limit.a",
	}},
	// 201 s: each fixture's C compiled for wasm32.
	{name: "TestWASIEmission", shards: 32, runner: "runWASIEmissionShard", keys: fixturePaths, planted: "dedication/dedication.a", requires: "ADAMIC_ORACLE_WASI"},
	// 56 s over 208 review programs, run one at a time within a shard as they always were; at 8 the worst was 15 s.
	{name: "TestReviewProgramsAgreeWithNode", shards: 12, runner: "runReviewAgreeShard", keys: reviewAgreePrograms, planted: "smoke.a"},
	// 236 s over about 1,030 rows of counts.md, each a counted build and run.
	{name: "TestCountsAreRecorded", shards: 32, runner: "runCountsShard", keys: countKeys, planted: "dedication/dedication.a"},
	// 92 s as one program of every loop, most of it clang on one huge C file.
	{name: "TestLoopCountersAgreeWithNode", shards: 8, runner: "runLoopCountersShard", keys: loopCounterKeys, planted: "let NAME = 0; NAME < values.length; NAME++\n", oneProgram: true},
	// 22.7 s: 20 mutants of library_method_values.a, each built and run.
	{name: "TestLibraryMethodValueMutants", shards: 4, runner: "runLibraryMethodValueMutantsShard", keys: libraryMethodValueMutants, planted: "trim"},
}

// A family is one test over many cases, run as shards. runner is the function each wrapper calls with its shard, keys
// lists every case's key in the order the cases run, and planted is a key the family's runner plants a mutant in when
// probed, which exactly its owner must catch. A family that runs only when requires is set to 1 skips otherwise. solo
// names the keys run each in a test of its own instead of a shard: the runner is given -1 for the first, -2 for the
// second, and so on. A family whose shard builds its cases into one program, oneProgram, has no subtest per case.
type family struct {
	name       string
	shards     int
	runner     string
	keys       func() []string
	planted    string
	requires   string
	solo       []string
	oneProgram bool
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
		solo := slices.Index(family.solo, key)
		owns := solo < 0 && familyShard(key, family.shards) == shard || solo >= 0 && shard == -1-solo
		if owns && (!probing || key == planted) {
			owned = append(owned, index)
		}
	}
	return owned
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
		for shard := 0; shard < family.shards; shard++ {
			fmt.Fprintf(&text, "\nfunc %s_%03d(t *testing.T) {\n\tt.Parallel()\n\t%s(t, %d)\n}\n", family.name, shard, family.runner, shard)
		}
		for solo, key := range family.solo {
			fmt.Fprintf(&text, "\nfunc %s(t *testing.T) {\n\tt.Parallel()\n\t%s(t, %d)\n}\n", familyTestName(family, key), family.runner, -1-solo)
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

// Every case of every family has one owner: its key is unique, since Go would run a repeated subtest name as #01 and
// the two would share a shard; the hash keeps every key inside the family's shards; and a solo key is one of the
// family's keys, named once, its test's name taken by no other. A probe left in the environment would thin every
// shard to one case, so a real run refuses it.
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
		seen := map[string]bool{}
		sizes := make([]int, family.shards)
		for _, key := range keys {
			if seen[key] {
				t.Errorf("%s: %s is registered twice", family.name, key)
			}
			seen[key] = true
			if !slices.Contains(family.solo, key) {
				sizes[familyShard(key, family.shards)]++
			}
		}
		if !seen[family.planted] {
			t.Errorf("%s: its planted case %s isn't one of its cases", family.name, family.planted)
		}
		names := map[string]bool{}
		for _, key := range family.solo {
			name := familyTestName(family, key)
			if !seen[key] || names[name] {
				t.Errorf("%s: solo %s must be one of its cases, named once, with a test name of its own (%s)", family.name, key, name)
			}
			names[name] = true
		}
		smallest, largest := len(keys), 0
		for _, size := range sizes {
			smallest, largest = min(smallest, size), max(largest, size)
		}
		t.Logf("%s: %d cases, each owned once, over %d shards of %d to %d and %d solo tests", family.name, len(keys), family.shards, smallest, largest, len(family.solo))
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
			t.Logf("the mutant planted in %s failed exactly %s", family.planted, owner)
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
