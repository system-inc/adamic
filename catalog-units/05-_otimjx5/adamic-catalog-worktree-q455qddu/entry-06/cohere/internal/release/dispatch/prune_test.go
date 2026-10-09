package dispatch

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

// A prune is judged by the exact set it removes. Each test seeds a cache with every kind of file the
// real one holds and asserts that set name by name, because "it removed some files" is the result a
// prune that deletes the wrong ones also produces.

var pruneNow = time.Date(2026, time.October, 2, 12, 0, 0, 0, time.UTC)

func TestPruneRemovesExactlyWhatNoRunCanReach(t *testing.T) {
	t.Parallel()

	paths, ages := seedPruneCache(t)

	plan, err := PlanPrune(paths, []string{"cohere-" + platformTag() + "-keepme0000000000"}, "", pruneNow)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Examined != len(ages)+1 {
		t.Fatalf("examined %d entries, the cache holds %d", plan.Examined, len(ages)+1)
	}

	got := []string{}
	for _, file := range plan.Remove {
		got = append(got, file.Name)
	}
	want := []string{
		// Twelve old cohere binaries, the newest eight kept. The keep-listed one is older than all of
		// them and still survives, so the four removed are the four oldest of the rest.
		"cohere-" + platformTag() + "-old00",
		"cohere-" + platformTag() + "-old01",
		"cohere-" + platformTag() + "-old02",
		"cohere-" + platformTag() + "-old03",
		// Ten old Swift engines, the newest eight kept.
		"cohere-swift-" + platformTag() + "-old00",
		"cohere-swift-" + platformTag() + "-old01",
		// A partial whose writer is long gone.
		"cohere-" + platformTag() + "-stale.partial-123",
	}
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("would remove\n  %s\nexpected\n  %s", strings.Join(got, "\n  "), strings.Join(want, "\n  "))
	}

	wantUnrecognised := []string{"a-subdirectory", "something-else"}
	if !reflect.DeepEqual(plan.Unrecognised, wantUnrecognised) {
		t.Fatalf("unrecognised %v, expected %v kept and reported", plan.Unrecognised, wantUnrecognised)
	}
}

func TestApplyPruneRemovesOnlyThePlanAndLeavesFrozenACandidate(t *testing.T) {
	t.Parallel()

	paths, ages := seedPruneCache(t)
	plan, err := PlanPrune(paths, nil, "", pruneNow)
	if err != nil {
		t.Fatal(err)
	}

	removed, err := ApplyPrune(paths, plan)
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != len(plan.Remove) {
		t.Fatalf("removed %d of the %d planned", len(removed), len(plan.Remove))
	}
	for name := range ages {
		_, statErr := os.Stat(filepath.Join(paths.BinaryDirectory(), name))
		planned := false
		for _, file := range plan.Remove {
			planned = planned || file.Name == name
		}
		if planned != os.IsNotExist(statErr) {
			t.Errorf("%s: planned=%v, but on disk the file %s", name, planned, map[bool]string{true: "is gone", false: "is still there"}[os.IsNotExist(statErr)])
		}
	}

	// A second apply of the same plan finds every file already gone, which is a concurrent prune, not
	// an error.
	if again, err := ApplyPrune(paths, plan); err != nil || len(again) != 0 {
		t.Fatalf("re-applying an applied plan: removed %d, err %v", len(again), err)
	}

	if _, err := ResolveFrozen(paths); err != nil {
		t.Fatalf("after a prune --frozen has nothing to run: %v", err)
	}
}

func TestApplyPruneRefusesANameThatIsAPath(t *testing.T) {
	t.Parallel()

	paths, _ := seedPruneCache(t)
	outside := filepath.Join(paths.CacheDirectory, "outside")
	writeFile(t, outside, "not in the binary cache\n")

	_, err := ApplyPrune(paths, PrunePlan{Remove: []PrunedFile{{Name: "../outside"}}})
	if err == nil {
		t.Fatal("a name that walks out of the binary cache was accepted")
	}
	if _, statErr := os.Stat(outside); statErr != nil {
		t.Fatalf("the file outside the binary cache was removed: %v", statErr)
	}
}

// Compiler extractions and leftover snapshots are judged by the same rule as binaries: the current one
// and anything used within the hour stay, the rest go. Ages are real (relative to now) rather than to
// pruneNow, because ApplyPrune asks the clock again at the moment of removal.
func TestPruneRemovesStaleExtractionsAndKeepsWhatABuildCanReach(t *testing.T) {
	t.Parallel()

	paths, _ := seedPruneCache(t)
	now := time.Now()
	directory := func(parent string, name string, age time.Duration) string {
		path := filepath.Join(parent, name)
		writeFile(t, filepath.Join(path, "tsc", "go.mod"), "module "+name+"\n")
		stamp := now.Add(-age)
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
		return path
	}
	// Older than everything, and kept because the build just made reads it.
	current := directory(paths.CompilerDirectory(), "currentcurrentcurrentcurrentcurrentcurren", 900*time.Hour)
	// Another pin, read by a build within the hour.
	inUse := directory(paths.CompilerDirectory(), "inuseinuseinuseinuseinuseinuseinuseinuse", 10*time.Minute)
	stale := directory(paths.CompilerDirectory(), "stalestalestalestalestalestalestalestale", 5*time.Hour)
	abandoned := directory(paths.CompilerDirectory(), "stalestalestalestalestalestalestalestale.partial-1", 5*time.Hour)
	extracting := directory(paths.CompilerDirectory(), "freshfreshfreshfreshfreshfreshfreshfresh.partial-2", time.Minute)
	killed := directory(paths.SnapshotDirectory(), "58cd5d457ebf-1999002027", 3*time.Hour)
	building := directory(paths.SnapshotDirectory(), "58cd5d457ebf-1130357668", 2*time.Minute)

	plan, err := PlanPrune(paths, nil, current, now)
	if err != nil {
		t.Fatal(err)
	}
	names := func(list []PrunedFile) []string {
		out := []string{}
		for _, entry := range list {
			if entry.Bytes == 0 {
				t.Errorf("%s planned with no size, so the log would understate what was reclaimed", entry.Name)
			}
			out = append(out, entry.Name)
		}
		sort.Strings(out)
		return out
	}
	wantCompilers := []string{filepath.Base(stale), filepath.Base(abandoned)}
	sort.Strings(wantCompilers)
	if got := names(plan.StaleCompilers); !reflect.DeepEqual(got, wantCompilers) {
		t.Fatalf("stale compilers %v, expected %v", got, wantCompilers)
	}
	if got := names(plan.StaleSnapshots); !reflect.DeepEqual(got, []string{filepath.Base(killed)}) {
		t.Fatalf("stale snapshots %v, expected only %s", got, filepath.Base(killed))
	}

	if _, err := ApplyPrune(paths, plan); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{stale, abandoned, killed} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("%s survived the prune: %v", path, err)
		}
	}
	for _, path := range []string{current, inUse, extracting, building} {
		if _, err := os.Stat(filepath.Join(path, "tsc", "go.mod")); err != nil {
			t.Errorf("%s was removed, and a build can still reach it: %v", path, err)
		}
	}

	// Without the current compiler named, nothing says which one a build reads, so none is planned.
	if unknown, err := PlanPrune(paths, nil, "", now); err != nil || len(unknown.StaleCompilers) != 0 {
		t.Fatalf("with no current compiler, planned %v (err %v)", unknown.StaleCompilers, err)
	}
}

// A build that touches an extraction between the plan and its removal keeps it. The plan is made, the
// extraction is marked used the way ensureCompiler marks it, and the apply must leave it.
func TestApplyPruneKeepsAnExtractionTouchedSinceThePlan(t *testing.T) {
	t.Parallel()

	paths, _ := seedPruneCache(t)
	now := time.Now()
	current := filepath.Join(paths.CompilerDirectory(), "current")
	writeFile(t, filepath.Join(current, "tsc", "go.mod"), "module current\n")
	raced := filepath.Join(paths.CompilerDirectory(), "raced")
	writeFile(t, filepath.Join(raced, "tsc", "go.mod"), "module raced\n")
	old := now.Add(-5 * time.Hour)
	if err := os.Chtimes(raced, old, old); err != nil {
		t.Fatal(err)
	}

	plan, err := PlanPrune(paths, nil, current, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.StaleCompilers) != 1 {
		t.Fatalf("expected the raced extraction planned, got %v", plan.StaleCompilers)
	}
	if err := os.Chtimes(raced, time.Now(), time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := ApplyPrune(paths, plan); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(raced, "tsc", "go.mod")); err != nil {
		t.Fatalf("an extraction a build marked in use after the plan was removed: %v", err)
	}
}

// The other half of the race: a build reusing an extraction marks it, so the hour counts from that
// build. Without the mark, a pin extracted yesterday is stale to the prune while a build compiles
// against it.
func TestEnsureCompilerMarksAReusedExtractionInUse(t *testing.T) {
	t.Parallel()

	paths := Paths{ModuleDirectory: t.TempDir(), CacheDirectory: t.TempDir()}
	commit := "8d550c837c90bd1805b047b7eeccc2baac2d5e7a"
	extraction := filepath.Join(paths.CompilerDirectory(), commit)
	writeFile(t, filepath.Join(extraction, "tsc", "go.mod"), "module tsc\n")
	old := time.Now().Add(-48 * time.Hour)
	if err := os.Chtimes(extraction, old, old); err != nil {
		t.Fatal(err)
	}

	got, err := ensureCompiler(paths, commit)
	if err != nil {
		t.Fatal(err)
	}
	if got != extraction {
		t.Fatalf("reused %s, expected %s", got, extraction)
	}
	information, err := os.Stat(extraction)
	if err != nil {
		t.Fatal(err)
	}
	if time.Since(information.ModTime()) > time.Minute {
		t.Fatalf("a reused extraction still reads as last used at %s, so a prune would take it mid-build", information.ModTime())
	}
}

func TestApplyPruneRefusesADirectoryNameThatIsAPath(t *testing.T) {
	t.Parallel()

	paths, _ := seedPruneCache(t)
	outside := filepath.Join(paths.CacheDirectory, "keep")
	writeFile(t, filepath.Join(outside, "file"), "not in the compiler cache\n")

	_, err := ApplyPrune(paths, PrunePlan{StaleCompilers: []PrunedFile{{Name: "../keep"}}})
	if err == nil {
		t.Fatal("a directory name that walks out of the compiler cache was accepted")
	}
	if _, statErr := os.Stat(filepath.Join(outside, "file")); statErr != nil {
		t.Fatalf("the directory outside the compiler cache was removed: %v", statErr)
	}
}

func TestAPruneThatRemovesNothingSaysHowManyItLookedAt(t *testing.T) {
	t.Parallel()

	paths := Paths{ModuleDirectory: t.TempDir(), CacheDirectory: t.TempDir()}
	if err := os.MkdirAll(paths.BinaryDirectory(), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(paths.BinaryDirectory(), "cohere-dev"), "binary\n")

	plan, err := PlanPrune(paths, nil, "", pruneNow)
	if err != nil {
		t.Fatal(err)
	}
	if err := RecordPrune(paths, plan, nil, pruneNow); err != nil {
		t.Fatal(err)
	}
	logged, err := os.ReadFile(paths.PruneLogPath())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(logged), "examined 1, removed 0 (0 bytes)") {
		t.Fatalf("the log does not say what an empty prune looked at:\n%s", logged)
	}

	// The other direction: a listing that cannot be read is an error, never an empty plan.
	if _, err := PlanPrune(Paths{CacheDirectory: filepath.Join(t.TempDir(), "absent")}, nil, "", pruneNow); err == nil {
		t.Fatal("a missing binary cache planned as if it were empty")
	}
}

// seedPruneCache fills a binary cache with every kind of entry the real one holds, and returns each
// file's age relative to pruneNow.
func seedPruneCache(t *testing.T) (Paths, map[string]time.Duration) {
	t.Helper()
	paths := Paths{ModuleDirectory: t.TempDir(), CacheDirectory: t.TempDir()}
	if err := os.MkdirAll(filepath.Join(paths.BinaryDirectory(), "a-subdirectory"), 0o755); err != nil {
		t.Fatal(err)
	}

	ages := map[string]time.Duration{}
	for index := 0; index < 12; index++ {
		// old00 is the oldest. Each is at least two hours old, outside the in-flight window.
		ages[fmt.Sprintf("cohere-%s-old%02d", platformTag(), index)] = time.Duration(100-index) * time.Hour
	}
	for index := 0; index < 10; index++ {
		ages[fmt.Sprintf("cohere-swift-%s-old%02d", platformTag(), index)] = time.Duration(100-index) * time.Hour
	}
	// Recent enough to be in use, so kept however many there are.
	ages["cohere-"+platformTag()+"-fresh0"] = 10 * time.Minute
	ages["cohere-"+platformTag()+"-fresh1"] = 20 * time.Minute
	ages["cohere-"+platformTag()+"-writing.partial-456"] = 2 * time.Minute
	ages["cohere-"+platformTag()+"-stale.partial-123"] = 5 * time.Hour
	// Named keepers, older than everything, so only their names protect them. Listed here rather than
	// read from prunedByName: a fixture built from the map under test changes with it, so dropping a
	// keeper from the map also dropped it from the fixture and the test passed. That happened.
	for _, name := range []string{"cohere-dev", "cohere-dev.hash", "cohere-dispatch", "cohere-swift-current", "cohere-swift-current.tmp"} {
		ages[name] = 500 * time.Hour
	}
	ages["something-else"] = 500 * time.Hour
	ages["cohere-"+platformTag()+"-keepme0000000000"] = 900 * time.Hour

	for name, age := range ages {
		path := filepath.Join(paths.BinaryDirectory(), name)
		if err := os.WriteFile(path, []byte(name), 0o755); err != nil {
			t.Fatal(err)
		}
		stamp := pruneNow.Add(-age)
		if err := os.Chtimes(path, stamp, stamp); err != nil {
			t.Fatal(err)
		}
	}
	return paths, ages
}

// platformTag is the "<goos>-<goarch>" the binary names carry.
func platformTag() string {
	return strings.TrimSuffix(strings.TrimPrefix(platformBinaryPrefix(), "cohere-"), "-")
}
