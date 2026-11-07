package buildcache

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// lookedUpAlone is a name key's answer as one product test's lookup gives it: a pass of its own, nothing held.
func lookedUpAlone(test *testing.T, r *rig, nameKey string) Finding {
	test.Helper()
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
func settledPair(test *testing.T, r *rig) (string, string) {
	test.Helper()
	header := Inputs{Name: "header", Files: []string{"include"}}
	headerProduct := Product(test, header, building("header"))
	Product(test, port, func(directory string) error {
		content, err := os.ReadFile(filepath.Join(headerProduct, "product"))
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "product"), append(content, " and port"...), 0o644)
	})
	built := r.built(test)
	r.window(built["header"].Build, func() {
		r.open(r.pid, "include/x.h")
		r.write(r.pid, filepath.Join(headerProduct, "product"))
	})
	r.window(built["port"].Build, func() {
		r.open(r.pid, filepath.Join(headerProduct, "product"))
		r.open(r.pid, "source/a.c")
	})
	if settlement := r.settle(test); len(settlement.Settled) != 2 {
		test.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	return built["header"].NameKey, built["port"].NameKey
}

// Find answers every name key as that product's own lookup does, found or missing and why, on a tree where all is
// built and on one where a file one product read changed. Mutants: a held tree value that drops the changed entry (both
// still found); a memo shared across name keys that keeps a pass's answer past a change.
// Not parallel: newRig.
func TestFindAnswersEveryNameKeyAsItsOwnLookupDoes(test *testing.T) {
	r := newRig(test)
	headerKey, portKey := settledPair(test, r)
	notes := Inputs{Name: "notes"}
	Product(test, notes, building("notes"))
	notesBuild := r.built(test)["notes"]
	r.window(notesBuild.Build, func() { r.open(r.pid, "notes.txt") })
	if settlement := r.settle(test); len(settlement.Settled) != 1 {
		test.Fatalf("notes: settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	test.Setenv("ADAMIC_BUILD_TRACE", "")
	asked := []string{portKey, headerKey, notesBuild.NameKey, strings.Repeat("0", 64)}
	compare := func(want map[string]string) {
		test.Helper()
		findings, err := Find(asked)
		if err != nil {
			test.Fatal(err)
		}
		if len(findings) != len(asked) {
			test.Fatalf("%d findings for %d name keys", len(findings), len(asked))
		}
		if settledTree.Load() != nil {
			test.Fatal("Find left the tree's values held")
		}
		for index, finding := range findings {
			alone := lookedUpAlone(test, r, asked[index])
			if finding != alone {
				test.Errorf("name key %s: Find says %+v, its own lookup %+v", asked[index][:12], finding, alone)
			}
			if finding.Outcome != want[asked[index]] {
				test.Errorf(
					"name key %s is %s, want %s (%s)",
					asked[index][:12],
					finding.Outcome,
					want[asked[index]],
					finding.Why,
				)
			}
		}
	}
	compare(map[string]string{
		portKey:            FindingFound,
		headerKey:          FindingFound,
		notesBuild.NameKey: FindingFound,
		asked[3]:           FindingMissing,
	})
	write(test, r.root, "include/x.h", "#define X 2\n")
	gitIn(test, r.root, "commit", "-qam", "x is 2")
	forgetTracked()
	compare(map[string]string{
		portKey:            FindingMissing,
		headerKey:          FindingMissing,
		notesBuild.NameKey: FindingFound,
		asked[3]:           FindingMissing,
	})
	if findings, _ := Find([]string{"not a name key"}); findings != nil {
		test.Fatalf("a line that isn't a name key was answered: %+v", findings)
	}
}

// Find values each tree entry once for the whole run, however many products read it, and each product once however
// many read it: 36 products in 12 layers, each reading go.mod and the 3 below, read go.mod once.
// Not parallel: newRig.
func TestFindValuesTheTreeOncePerRun(test *testing.T) {
	r := newRig(test)
	test.Setenv("ADAMIC_BUILD_TRACE", "")
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
				test.Fatal(err)
			}
			set.Key = evaluated.key
			if err := recordReads(
				r.cache,
				nameKeyOf(layer, index),
				fmt.Sprintf("layer %d product %d", layer, index),
				set,
			); err != nil {
				test.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(r.cache, evaluated.key), 0o755); err != nil {
				test.Fatal(err)
			}
			asked = append(asked, nameKeyOf(layer, index))
		}
	}
	before := treeValuesRead.Load()
	findings, err := Find(asked)
	if err != nil {
		test.Fatal(err)
	}
	for _, finding := range findings {
		if finding.Outcome != FindingFound {
			test.Fatalf("%s: %+v", finding.NameKey[:12], finding)
		}
	}
	if read := treeValuesRead.Load() - before; read != 1 {
		test.Fatalf("36 products valued the tree %d times, want once", read)
	}
}

// Find finds a product by the cache's read sets alone: a product only the store holds is missing, naming the store's
// key, since its test fetches it (with its description, and its audit); a name key whose read sets the cache doesn't
// hold is missing too, and the store is never asked for them.
// Not parallel: newRig.
func TestFindReadsOnlyTheCachesReadSetsAndLeavesFetchesToTheTest(test *testing.T) {
	r := newRig(test)
	store, _ := shared(test)
	test.Setenv("ADAMIC_BUILD_STORE", "traced="+os.Getenv("ADAMIC_BUILD_STORE"))
	test.Setenv("ADAMIC_BUILD_CACHE_DIR", r.cache)
	Product(test, port, building("port"))
	build := r.built(test)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	if settlement := r.settle(test); len(settlement.Settled) != 1 {
		test.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	test.Setenv("ADAMIC_BUILD_TRACE", "")
	file, _ := loadReads(r.cache, build.NameKey)
	key := file.Sets[0].Key
	findings, err := Find([]string{build.NameKey})
	if err != nil || findings[0] != (Finding{NameKey: build.NameKey, Outcome: FindingFound, Key: key, Name: "port"}) {
		test.Fatalf("a product the cache holds: %+v, %v", findings, err)
	}
	if err := os.RemoveAll(filepath.Join(r.cache, key)); err != nil {
		test.Fatal(err)
	}
	findings, err = Find([]string{build.NameKey})
	if err != nil ||
		findings[0].Outcome != FindingMissing ||
		!strings.Contains(findings[0].Why, "the store may hold it under "+key[:12]) {
		test.Fatalf("a product only the store holds: %+v, %v", findings, err)
	}
	if _, err := os.Stat(filepath.Join(r.cache, key)); err == nil {
		test.Fatal("Find fetched a product, which only its test may")
	}
	if err := os.Remove(readsPath(r.cache, build.NameKey)); err != nil {
		test.Fatal(err)
	}
	findings, err = Find([]string{build.NameKey})
	if err != nil || findings[0].Outcome != FindingMissing {
		test.Fatalf("a product whose read sets the cache doesn't hold: %+v, %v", findings, err)
	}
	if len(store.reads)+len(store.writes) != 0 {
		test.Fatalf("Find reached the store: read %q, wrote %q", store.reads, store.writes)
	}
}

// recordedProduct records a read set of entries for nameKey, valued on the tree as it is, and makes its product in the
// cache when made: what a settled product leaves. It returns the key the set gives.
func recordedProduct(test *testing.T, r *rig, nameKey string, entries []readEntry, made bool) string {
	test.Helper()
	set := readSet{Entries: sortedEntries(entries)}
	evaluated, err := evaluate(r.root, r.cache, nameKey, set, 0, nil)
	if err != nil {
		test.Fatal(err)
	}
	set.Key = evaluated.key
	if err := recordReads(r.cache, nameKey, "product "+nameKey[:12], set); err != nil {
		test.Fatal(err)
	}
	if made {
		if err := os.MkdirAll(filepath.Join(r.cache, evaluated.key), 0o755); err != nil {
			test.Fatal(err)
		}
	}
	return evaluated.key
}

// A process whose tree can't change under it values each tree entry once for its life and each product it finds once:
// a read-mode one (a runner's: its source never changes under it; #szswjtc: TypeScript's readproofs re-valued a
// product's whole read set on every lookup, 8 to 27 s a hit), and a traced one (its run's builds are all refused if the
// tree changes; #47ykkmr: css's PostCSS answers, 62 to 182 s a hit under strace). A second lookup of a product reads no
// tree value and answers the same, and another product reading what the first read reads none either, even after Find
// has held and dropped values of its own. A product it misses is looked up again, so one that arrives later is found.
// Mutants: no product held (each lookup values its read set again, over held tree values); the held tree values kept in
// settledTree (Find's end drops them); a miss held as found is; a traced process holding nothing.
// Not parallel: newRig, and treeValuesRead counts the whole process's reads.
func TestAProcessWhoseTreeCannotChangeValuesEachProductOnce(test *testing.T) {
	for _, mode := range []string{"read", "traced", "unpacked"} {
		test.Run(mode, func(test *testing.T) { processValuesEachProductOnce(test, mode) })
	}
}

func processValuesEachProductOnce(test *testing.T, mode string) {
	r := newRig(test)
	test.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetHeldForLife()
	test.Cleanup(forgetHeldForLife)
	first, second, late := fmt.Sprintf("%064x", 1), fmt.Sprintf("%064x", 2), fmt.Sprintf("%064x", 3)
	recordedProduct(test, r, first, []readEntry{
		{Kind: "content", Path: "go.mod"},
		{Kind: "content", Path: "source/a.c"},
		{Kind: "listing", Path: "source"},
	}, true)
	recordedProduct(
		test,
		r,
		second,
		[]readEntry{{Kind: "content", Path: "source/a.c"}, {Kind: "product", Path: first}},
		true,
	)
	lateKey := recordedProduct(test, r, late, []readEntry{{Kind: "content", Path: "go.mod"}}, false)
	want := lookedUpAlone(test, r, second)
	if want.Outcome != FindingFound {
		test.Fatalf("outside read mode and a trace: %+v", want)
	}
	root := r.root
	switch mode {
	case "read":
		test.Setenv("ADAMIC_BUILD_CACHE", "read")
	case "traced":
		test.Setenv("ADAMIC_BUILD_TRACE", test.TempDir())
	case "unpacked":
		// Neither read mode nor a trace: Loom's key check, untraced over the source a runner unpacks.
		root = unpackedRig(test, r)
	}
	var products int64
	lookup := func(nameKey string) (found, int64) {
		test.Helper()
		before, productsBefore := treeValuesRead.Load(), productsLookedUp.Load()
		looked, err := productFor(root, r.cache, nameKey, 0, nil)
		if err != nil {
			test.Fatal(err)
		}
		products = productsLookedUp.Load() - productsBefore
		return looked, treeValuesRead.Load() - before
	}
	looked, read := lookup(second)
	if looked.key != want.Key || read != 3 || products != 2 {
		test.Fatalf(
			"the first lookup found %q (want %q), reading %d tree values (want go.mod, a.c and source's listing) and %d products' read sets (want 2)",
			looked.key,
			want.Key,
			read,
			products,
		)
	}
	again, read := lookup(second)
	if !reflect.DeepEqual(again, looked) || read != 0 || products != 0 {
		test.Fatalf(
			"the second lookup answered %+v (the first %+v), reading %d tree values and %d products' read sets",
			again,
			looked,
			read,
			products,
		)
	}
	if looked, read = lookup(first); looked.directory == "" || read != 0 || products != 0 {
		test.Fatalf(
			"a product the first lookup read was looked up again, reading %d tree values and %d read sets: %+v",
			read,
			products,
			looked,
		)
	}
	if _, err := Find(nil); err != nil {
		test.Fatal(err)
	}
	if looked, read = lookup(late); looked.directory != "" || read != 0 {
		test.Fatalf("after Find, a product reading go.mod read %d tree values: %+v", read, looked)
	}
	if err := os.MkdirAll(filepath.Join(r.cache, lateKey), 0o755); err != nil {
		test.Fatal(err)
	}
	if looked, _ = lookup(late); looked.key != lateKey {
		test.Fatalf("a product that arrived after a lookup missed it isn't found: %+v", looked)
	}
}

// unpackedRig is r's tree as Loom unpacks it for a runner: its tracked files, with no .git, and git's listing in a
// .tracked manifest.
func unpackedRig(test *testing.T, r *rig) string {
	test.Helper()
	git := func(arguments ...string) string {
		test.Helper()
		output, err := toolchain.GitCommand(append([]string{"-C", r.root}, arguments...)...).Output()
		if err != nil {
			test.Fatalf("git %v: %v", arguments, err)
		}
		return string(output)
	}
	root := test.TempDir()
	for _, name := range strings.Split(strings.TrimSpace(git("ls-files")), "\n") {
		content, err := os.ReadFile(filepath.Join(r.root, name))
		if err != nil {
			test.Fatal(err)
		}
		write(test, root, name, string(content))
	}
	write(test, root, ".tracked/files", git("ls-tree", "-r", "-z", "--full-tree", "HEAD"))
	return root
}

// Outside read mode and a trace a lookup values the tree afresh: a builder, or a test, changes it between lookups, and
// the second answers for the tree as it is then. Mutant: the per-process memo on in every mode.
// Not parallel: newRig.
func TestOutsideReadModeAndATraceEveryLookupValuesTheTreeAfresh(test *testing.T) {
	r := newRig(test)
	test.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetHeldForLife()
	test.Cleanup(forgetHeldForLife)
	nameKey := fmt.Sprintf("%064x", 4)
	recordedProduct(test, r, nameKey, []readEntry{{Kind: "content", Path: "source/a.c"}}, true)
	if before := lookedUpAlone(test, r, nameKey); before.Outcome != FindingFound {
		test.Fatalf("before the change: %+v", before)
	}
	write(test, r.root, "source/a.c", "int a2;\n")
	gitIn(test, r.root, "commit", "-qam", "a is a2")
	forgetTracked()
	read := treeValuesRead.Load()
	after := lookedUpAlone(test, r, nameKey)
	if after.Outcome != FindingMissing || !strings.Contains(after.Why, "content source/a.c") {
		test.Fatalf("after source/a.c changed, the product is %+v", after)
	}
	if treeValuesRead.Load() == read {
		test.Fatal("the second lookup read no tree value")
	}
}
