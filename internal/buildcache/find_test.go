package buildcache

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// lookedUpAlone is a name key's answer as one product test's lookup gives it: a pass of its own, nothing held.
func lookedUpAlone(t *testing.T, r *rig, nameKey string) Finding {
	t.Helper()
	answer := Finding{NameKey: nameKey, Outcome: FindingMissing}
	looked, err := productFor(r.root, r.cache, nameKey, 0, nil)
	switch {
	case err != nil:
		answer.Why = err.Error()
	case looked.directory != "":
		file, _ := loadReads(r.cache, nameKey)
		answer.Outcome, answer.Key, answer.Name = FindingFound, looked.key, file.Name
	default:
		answer.Why = looked.why
	}
	return answer
}

// settledPair settles header, and port reading header's product and source/a.c, and returns their name keys.
func settledPair(t *testing.T, r *rig) (string, string) {
	t.Helper()
	header := Inputs{Name: "header", Files: []string{"include"}}
	headerProduct := Product(t, header, building("header"))
	Product(t, port, func(directory string) error {
		content, err := os.ReadFile(filepath.Join(headerProduct, "product"))
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "product"), append(content, " and port"...), 0o644)
	})
	built := r.built(t)
	r.window(built["header"].Build, func() {
		r.open(r.pid, "include/x.h")
		r.write(r.pid, filepath.Join(headerProduct, "product"))
	})
	r.window(built["port"].Build, func() {
		r.open(r.pid, filepath.Join(headerProduct, "product"))
		r.open(r.pid, "source/a.c")
	})
	if settlement := r.settle(t); len(settlement.Settled) != 2 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	return built["header"].NameKey, built["port"].NameKey
}

// Find answers every name key as that product's own lookup does, found or missing and why, on a tree where all is
// built and on one where a file one product read changed. Mutants: a held tree value that drops the changed entry (both
// still found); a memo shared across name keys that keeps a pass's answer past a change.
// Not parallel: newRig.
func TestFindAnswersEveryNameKeyAsItsOwnLookupDoes(t *testing.T) {
	r := newRig(t)
	headerKey, portKey := settledPair(t, r)
	notes := Inputs{Name: "notes"}
	Product(t, notes, building("notes"))
	notesBuild := r.built(t)["notes"]
	r.window(notesBuild.Build, func() { r.open(r.pid, "notes.txt") })
	if settlement := r.settle(t); len(settlement.Settled) != 1 {
		t.Fatalf("notes: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	asked := []string{portKey, headerKey, notesBuild.NameKey, strings.Repeat("0", 64)}
	compare := func(want map[string]string) {
		t.Helper()
		findings, err := Find(asked)
		if err != nil {
			t.Fatal(err)
		}
		if len(findings) != len(asked) {
			t.Fatalf("%d findings for %d name keys", len(findings), len(asked))
		}
		if settledTree.Load() != nil {
			t.Fatal("Find left the tree's values held")
		}
		for index, finding := range findings {
			alone := lookedUpAlone(t, r, asked[index])
			if finding != alone {
				t.Errorf("name key %s: Find says %+v, its own lookup %+v", asked[index][:12], finding, alone)
			}
			if finding.Outcome != want[asked[index]] {
				t.Errorf("name key %s is %s, want %s (%s)", asked[index][:12], finding.Outcome, want[asked[index]], finding.Why)
			}
		}
	}
	compare(map[string]string{portKey: FindingFound, headerKey: FindingFound, notesBuild.NameKey: FindingFound, asked[3]: FindingMissing})
	write(t, r.root, "include/x.h", "#define X 2\n")
	gitIn(t, r.root, "commit", "-qam", "x is 2")
	forgetTracked()
	compare(map[string]string{portKey: FindingMissing, headerKey: FindingMissing, notesBuild.NameKey: FindingFound, asked[3]: FindingMissing})
	if findings, _ := Find([]string{"not a name key"}); findings != nil {
		t.Fatalf("a line that isn't a name key was answered: %+v", findings)
	}
}

// Find values each tree entry once for the whole run, however many products read it, and each product once however
// many read it: 36 products in 12 layers, each reading go.mod and the 3 below, read go.mod once.
// Not parallel: newRig.
func TestFindValuesTheTreeOncePerRun(t *testing.T) {
	r := newRig(t)
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	nameKeyOf := func(layer, index int) string {
		return fmt.Sprintf("%064x", layer*16+index+1)
	}
	var asked []string
	for layer := range 12 {
		for index := range 3 {
			set := readSet{Entries: []readEntry{{Kind: "content", Path: "go.mod"}}}
			if layer > 0 {
				for below := range 3 {
					set.Entries = append(set.Entries, readEntry{Kind: "product", Path: nameKeyOf(layer-1, below)})
				}
			}
			set.Entries = sortedEntries(set.Entries)
			evaluated, err := evaluate(r.root, r.cache, nameKeyOf(layer, index), set, 0, nil)
			if err != nil {
				t.Fatal(err)
			}
			set.Key = evaluated.key
			if err := recordReads(r.cache, nameKeyOf(layer, index), fmt.Sprintf("layer %d product %d", layer, index), set); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(r.cache, evaluated.key), 0o755); err != nil {
				t.Fatal(err)
			}
			asked = append(asked, nameKeyOf(layer, index))
		}
	}
	before := treeValuesRead.Load()
	findings, err := Find(asked)
	if err != nil {
		t.Fatal(err)
	}
	for _, finding := range findings {
		if finding.Outcome != FindingFound {
			t.Fatalf("%s: %+v", finding.NameKey[:12], finding)
		}
	}
	if read := treeValuesRead.Load() - before; read != 1 {
		t.Fatalf("36 products valued the tree %d times, want once", read)
	}
}

// Find joins the store's read sets to the cache's as a lookup does, so a cache holding a product without its read
// sets finds it; a product only the store holds is missing, naming the store's key, since its test fetches it (with
// its description, and its audit).
// Not parallel: newRig.
func TestFindJoinsTheStoresReadSetsAndLeavesFetchesToTheTest(t *testing.T) {
	r := newRig(t)
	shared(t)
	t.Setenv("ADAMIC_BUILD_STORE", "traced="+os.Getenv("ADAMIC_BUILD_STORE"))
	t.Setenv("ADAMIC_BUILD_CACHE_DIR", r.cache)
	token(t)
	Product(t, port, building("port"))
	build := r.built(t)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	if settlement := r.settle(t); len(settlement.Settled) != 1 || strings.Contains(settlement.Settled[0], "publish failed") {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	file, _ := loadReads(r.cache, build.NameKey)
	key := file.Sets[0].Key
	if err := os.Remove(readsPath(r.cache, build.NameKey)); err != nil {
		t.Fatal(err)
	}
	findings, err := Find([]string{build.NameKey})
	if err != nil || findings[0] != (Finding{NameKey: build.NameKey, Outcome: FindingFound, Key: key, Name: "port"}) {
		t.Fatalf("a product whose read sets only the store holds: %+v, %v", findings, err)
	}
	if file, _ := loadReads(r.cache, build.NameKey); len(file.Sets) == 0 || file.Name != "port" {
		t.Fatalf("the store's read sets weren't joined to the cache's: %+v", file)
	}
	if err := os.RemoveAll(filepath.Join(r.cache, key)); err != nil {
		t.Fatal(err)
	}
	findings, err = Find([]string{build.NameKey})
	if err != nil || findings[0].Outcome != FindingMissing || !strings.Contains(findings[0].Why, "the store may hold it under "+key[:12]) {
		t.Fatalf("a product only the store holds: %+v, %v", findings, err)
	}
	if _, err := os.Stat(filepath.Join(r.cache, key)); err == nil {
		t.Fatal("Find fetched a product, which only its test may")
	}
}

// recordedProduct records a read set of entries for nameKey, valued on the tree as it is, and makes its product in the
// cache when made: what a settled product leaves. It returns the key the set gives.
func recordedProduct(t *testing.T, r *rig, nameKey string, entries []readEntry, made bool) string {
	t.Helper()
	set := readSet{Entries: sortedEntries(entries)}
	evaluated, err := evaluate(r.root, r.cache, nameKey, set, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	set.Key = evaluated.key
	if err := recordReads(r.cache, nameKey, "product "+nameKey[:12], set); err != nil {
		t.Fatal(err)
	}
	if made {
		if err := os.MkdirAll(filepath.Join(r.cache, evaluated.key), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return evaluated.key
}

// A read-mode process (a runner's: its source never changes under it) values each tree entry once for its life and
// each product it finds once (#szswjtc: TypeScript's readproofs re-valued a product's whole read set on every lookup,
// 8 to 27 s a hit): a second lookup of a product reads no tree value and answers the same, and another product reading
// what the first read reads none either, even after Find has held and dropped values of its own. A product it misses
// is looked up again, so one that arrives later is found. Mutants: no product held (each lookup values its read set
// again, over held tree values); read mode's held tree values kept in settledTree (Find's end drops them); a miss
// held as found is.
// Not parallel: newRig, and treeValuesRead counts the whole process's reads.
func TestAReadModeProcessValuesEachProductOnce(t *testing.T) {
	r := newRig(t)
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetReadMode()
	t.Cleanup(forgetReadMode)
	first, second, late := fmt.Sprintf("%064x", 1), fmt.Sprintf("%064x", 2), fmt.Sprintf("%064x", 3)
	recordedProduct(t, r, first, []readEntry{{Kind: "content", Path: "go.mod"}, {Kind: "content", Path: "source/a.c"}, {Kind: "listing", Path: "source"}}, true)
	recordedProduct(t, r, second, []readEntry{{Kind: "content", Path: "source/a.c"}, {Kind: "product", Path: first}}, true)
	lateKey := recordedProduct(t, r, late, []readEntry{{Kind: "content", Path: "go.mod"}}, false)
	want := lookedUpAlone(t, r, second)
	if want.Outcome != FindingFound {
		t.Fatalf("outside read mode: %+v", want)
	}
	t.Setenv("ADAMIC_BUILD_CACHE", "read")
	var products int64
	lookup := func(nameKey string) (found, int64) {
		t.Helper()
		before, productsBefore := treeValuesRead.Load(), productsLookedUp.Load()
		looked, err := productFor(r.root, r.cache, nameKey, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		products = productsLookedUp.Load() - productsBefore
		return looked, treeValuesRead.Load() - before
	}
	looked, read := lookup(second)
	if looked.key != want.Key || read != 3 || products != 2 {
		t.Fatalf("the first lookup found %q (want %q), reading %d tree values (want go.mod, a.c and source's listing) and %d products' read sets (want 2)", looked.key, want.Key, read, products)
	}
	again, read := lookup(second)
	if !reflect.DeepEqual(again, looked) || read != 0 || products != 0 {
		t.Fatalf("the second lookup answered %+v (the first %+v), reading %d tree values and %d products' read sets", again, looked, read, products)
	}
	if looked, read = lookup(first); looked.directory == "" || read != 0 || products != 0 {
		t.Fatalf("a product the first lookup read was looked up again, reading %d tree values and %d read sets: %+v", read, products, looked)
	}
	if _, err := Find(nil); err != nil {
		t.Fatal(err)
	}
	if looked, read = lookup(late); looked.directory != "" || read != 0 {
		t.Fatalf("after Find, a product reading go.mod read %d tree values: %+v", read, looked)
	}
	if err := os.MkdirAll(filepath.Join(r.cache, lateKey), 0o755); err != nil {
		t.Fatal(err)
	}
	if looked, _ = lookup(late); looked.key != lateKey {
		t.Fatalf("a product that arrived after a lookup missed it isn't found: %+v", looked)
	}
}

// Outside read mode a lookup values the tree afresh: a builder, or a test, changes it between lookups, and the second
// answers for the tree as it is then. Mutant: read mode's per-process memo on in every mode.
// Not parallel: newRig.
func TestOutsideReadModeEveryLookupValuesTheTreeAfresh(t *testing.T) {
	r := newRig(t)
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetReadMode()
	t.Cleanup(forgetReadMode)
	nameKey := fmt.Sprintf("%064x", 4)
	recordedProduct(t, r, nameKey, []readEntry{{Kind: "content", Path: "source/a.c"}}, true)
	if before := lookedUpAlone(t, r, nameKey); before.Outcome != FindingFound {
		t.Fatalf("before the change: %+v", before)
	}
	write(t, r.root, "source/a.c", "int a2;\n")
	gitIn(t, r.root, "commit", "-qam", "a is a2")
	forgetTracked()
	read := treeValuesRead.Load()
	after := lookedUpAlone(t, r, nameKey)
	if after.Outcome != FindingMissing || !strings.Contains(after.Why, "content source/a.c") {
		t.Fatalf("after source/a.c changed, the product is %+v", after)
	}
	if treeValuesRead.Load() == read {
		t.Fatal("the second lookup read no tree value")
	}
}
