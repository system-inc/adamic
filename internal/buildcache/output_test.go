package buildcache

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache/location"
)

// makeReadOnly takes write permission off directory and every directory under it, as a cache Loom's tree builder hands
// a candidate's tests reads, and gives it back when the test ends so the test's cleanup can remove it.
func makeReadOnly(test *testing.T, directory string) {
	test.Helper()
	var directories []string
	walkError := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, pathError error) error {
		if pathError == nil && entry.IsDir() {
			directories = append(directories, path)
		}
		return pathError
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	test.Cleanup(func() {
		for _, each := range directories {
			os.Chmod(each, 0o755)
		}
	})
	for _, each := range directories {
		if changeModeError := os.Chmod(each, 0o555); changeModeError != nil {
			test.Fatal(changeModeError)
		}
	}
}

// cacheListing is every path under directory with its mode, size and modification time: what a write anywhere in it
// changes, even where the file system lets the write through (a test running as root).
func cacheListing(test *testing.T, directory string) map[string]string {
	test.Helper()
	listing := map[string]string{}
	walkError := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, pathError error) error {
		if pathError != nil {
			return pathError
		}
		info, infoError := entry.Info()
		if infoError != nil {
			return infoError
		}
		listing[path] = fmt.Sprintf("%s %d %d", info.Mode(), info.Size(), info.ModTime().UnixNano())
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	return listing
}

// exists fails the test unless path is there.
func exists(test *testing.T, path, what string) {
	test.Helper()
	if _, statError := os.Stat(path); statError != nil {
		test.Errorf("%s isn't at %s: %v", what, path, statError)
	}
}

// With ADAMIC_BUILD_OUTPUT_DIR set, a build cache this process can't write is only read: every product it already holds
// is found there and never built (a settled one by its read sets, one built for this machine alone, and its resolved
// copy is made beside it in the output directory), a product only the store holds is fetched into the output directory,
// and every product built, pointed to, digested, valued or settled lands in the output directory, at the place it would
// have in the cache. Nothing in the cache changes, a mode, a size, a time, and nothing in the output directory is a
// link. Mutants, each caught: any one write site (filled, pointLocal, rerecord's read sets, productDigest,
// systemValue's machine value, Resolved's copy, a fetch, Settle's placement, its read sets and its .inputs) left
// writing in the cache.
// Not parallel: newRig, and the output directory is a variable of the process (test.Setenv).
func TestAnOutputDirectoryTakesEveryWriteAndTheCacheIsOnlyRead(test *testing.T) {
	r := newRig(test)
	refuseBuild := func(string) error {
		test.Error("built a product the read-only cache holds")
		return errors.New("built a product the read-only cache holds")
	}
	// The cache as Loom ships it: a settled product with its read sets, whose set stored a key it no longer gives, so a
	// lookup records it again; a product built for this machine alone, whose text names the tree; a read set naming a
	// product only the store holds.
	Product(test, port, building("built"))
	build := r.built(test)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	settlement := r.settle(test)
	if len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		test.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, loadReadsError := loadReads(r.cache, build.NameKey)
	if loadReadsError != nil || len(file.Sets) != 1 {
		test.Fatalf("port's read sets: %+v, %v", file, loadReadsError)
	}
	settledKey := file.Sets[0].Key
	settled := filepath.Join(r.cache, settledKey)
	if withReadsError := withReads(r.cache, build.NameKey, func(file *readsFile) error {
		file.Sets[0].Key = strings.Repeat("0", 64)
		return nil
	}); withReadsError != nil {
		test.Fatal(withReadsError)
	}
	test.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetHeldForLife()
	local := Inputs{Name: "local", Files: []string{"source"}}
	built := Product(test, local, building("<repository>/source/a.c\n"))
	if !under(built, filepath.Join(r.cache, "local")) {
		test.Fatalf("a product for this machine alone is at %s, not in the cache's local directory", built)
	}
	stored := Inputs{Name: "stored", Files: []string{"source"}}
	storedNameKey := NameKey(r.root, stored)
	storedSet := readSet{Entries: []readEntry{{Kind: "content", Path: "source/b.c"}}}
	evaluated, evaluateError := evaluate(r.root, r.cache, storedNameKey, storedSet, 0, nil)
	if evaluateError != nil {
		test.Fatal(evaluateError)
	}
	storedSet.Key = evaluated.key
	if recordReadsError := recordReads(r.cache, storedNameKey, stored.Name, storedSet); recordReadsError != nil {
		test.Fatal(recordReadsError)
	}

	makeReadOnly(test, r.cache)
	before := cacheListing(test, r.cache)
	scratch, resolveSymlinksError := filepath.EvalSymlinks(test.TempDir())
	if resolveSymlinksError != nil {
		test.Fatal(resolveSymlinksError)
	}
	output := filepath.Join(scratch, "output")
	test.Setenv(location.OutputDirectoryVariable, output)
	forgetHeldForLife()

	// Found, never built: by its read sets, and for this machine alone.
	if found, getError := Get(port, refuseBuild); getError != nil || found != settled {
		test.Fatalf("the settled product was found at %q (%v), want %s", found, getError, settled)
	}
	exists(test, readsPath(output, build.NameKey), "the set that found port under another key, recorded again")
	if again, loadReadsError := loadReads(
		output,
		build.NameKey,
	); loadReadsError != nil || len(again.Sets) == 0 || again.Sets[0].Key != settledKey {
		test.Errorf(
			"port's read sets read %+v, %v; want the set that found it first, under %s",
			again,
			loadReadsError,
			settledKey[:12],
		)
	}
	if found, getError := Get(local, refuseBuild); getError != nil || found != built {
		test.Fatalf("the local product was found at %q (%v), want %s", found, getError, built)
	}
	// Found before anything is made or locked: its hit leaves no trace in the output directory either.
	if _, statError := os.Stat(filepath.Join(output, "local")); statError == nil {
		test.Error("a hit in the cache made the output directory's local directory")
	}
	// Fetched into the output directory.
	store := &fakeStore{objects: map[string][]byte{}}
	server := httptest.NewServer(store)
	test.Cleanup(server.Close)
	store.put(test, storedSet.Key, map[string]string{"product": "stored"})
	test.Setenv("ADAMIC_BUILD_STORE", server.URL)
	test.Setenv("ADAMIC_BUILD_AUDIT", "0")
	if fetched, getError := Get(
		stored,
		refuseBuild,
	); getError != nil || fetched != filepath.Join(output, storedSet.Key) {
		test.Fatalf(
			"the stored product was fetched to %q (%v), want %s",
			fetched,
			getError,
			filepath.Join(output, storedSet.Key),
		)
	}
	test.Setenv("ADAMIC_BUILD_STORE", "off")
	// Built, pointed to, resolved, digested and valued in the output directory.
	fresh := Inputs{Name: "fresh", Files: []string{"source"}}
	made := Product(test, fresh, building("fresh"))
	if !under(made, filepath.Join(output, "local")) {
		test.Fatalf("a new product is at %s, not in the output directory's local directory", made)
	}
	exists(test, localPointer(output, NameKey(r.root, fresh)), "the new product's local pointer")
	copied, resolvedError := Resolved(built)
	if resolvedError != nil {
		test.Fatalf("resolving a product the read-only cache holds: %v", resolvedError)
	}
	if !under(copied, filepath.Join(output, "local")) {
		test.Fatalf("the resolved copy is at %s, not beside the product's place in the output directory", copied)
	}
	if content, readError := os.ReadFile(
		filepath.Join(copied, "product"),
	); readError != nil || string(content) != r.root+"/source/a.c\n" {
		test.Fatalf("the resolved copy reads %q, %v", content, readError)
	}
	if _, productDigestError := productDigest(settled); productDigestError != nil {
		test.Fatal(productDigestError)
	}
	exists(test, filepath.Join(output, settledKey+".digest"), "the settled product's digest")
	machineFile := filepath.Join(scratch, "machine-file")
	write(test, scratch, "machine-file", "a machine file\n")
	systemValue(machineFile)
	if values, readDirectoryError := os.ReadDir(
		filepath.Join(output, "machine"),
	); readDirectoryError != nil || len(values) != 1 {
		test.Errorf("the machine file's value kept in the output directory: %d, %v", len(values), readDirectoryError)
	}

	// Settled in the output directory: a traced build reading the source and the product only the cache holds.
	r.retrace(test, "output")
	forgetHeldForLife()
	if found := Product(test, port, refuseBuild); found != settled {
		test.Fatalf("under the trace the settled product was found at %s, want %s", found, settled)
	}
	consumer := Inputs{Name: "consumer", Files: []string{"source"}}
	pending := Product(test, consumer, building("consumer"))
	if !under(pending, filepath.Join(output, "pending")) {
		test.Fatalf("a traced build is pending at %s, not in the output directory", pending)
	}
	consuming := r.built(test)["consumer"]
	r.window(consuming.Build, func() {
		r.open(r.pid, "source/b.c")
		r.open(r.pid, filepath.Join(settled, "product"))
	})
	settlement = r.settle(test)
	if len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		test.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	var consumerKey string
	foundInCache := false
	for _, report := range settlement.Reports {
		switch {
		case report.Outcome == ReportSettled && report.NameKey == consuming.NameKey:
			consumerKey = report.Key
		case report.Outcome == ReportFound && report.NameKey == build.NameKey && report.Key == settledKey:
			foundInCache = true
		}
	}
	if !foundInCache {
		test.Errorf(
			"the run's reports don't say it found port in the cache under %s: %+v",
			settledKey[:12],
			settlement.Reports,
		)
	}
	if consumerKey == "" {
		test.Fatalf("no settled report for consumer: %+v", settlement.Reports)
	}
	exists(test, filepath.Join(output, consumerKey, "product"), "the settled product")
	exists(test, filepath.Join(output, consumerKey+".inputs"), "the settled product's inputs")
	exists(test, readsPath(output, consuming.NameKey), "the settled product's read sets")
	if entries := r.entries(
		test,
		consuming.NameKey,
	); !strings.Contains(strings.Join(entries, "\n"), "product "+build.NameKey) {
		test.Errorf("consumer's read set doesn't read port, which only the cache holds: %q", entries)
	}
	if _, statError := os.Stat(filepath.Join(r.cache, consumerKey)); statError == nil {
		test.Error("the settled product is in the read-only cache")
	}

	after := cacheListing(test, r.cache)
	for path, was := range before {
		if now, ok := after[path]; !ok || now != was {
			test.Errorf("%s changed in the read-only cache: %q, now %q", path, was, now)
		}
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			test.Errorf("%s appeared in the read-only cache", path)
		}
	}
	filepath.WalkDir(output, func(path string, entry fs.DirEntry, pathError error) error {
		if pathError == nil && entry.Type()&fs.ModeSymlink != 0 {
			test.Errorf("the output directory holds a link: %s", path)
		}
		return pathError
	})
}

// A key and a product's bytes never say where they were written: the same inputs key alike, Key and NameKey, with
// ADAMIC_BUILD_OUTPUT_DIR unset, set to one directory and set to another, and inputs naming the directory each run
// writes to key as if each named the cache; a product built there names itself and the cache as one built in the cache
// does, and reads back to where it is; a product whose text names the output directory raw is refused, as one naming
// the cache is. Mutants: the output directory left out of portable's places, Relative's places or relocatable's, or
// taken as an ADAMIC_ location.
// Not parallel: newRig, and the output directory is a variable of the process (test.Setenv).
func TestTheOutputDirectoryKeysAndRelocatesAsTheCache(test *testing.T) {
	r := newRig(test)
	test.Setenv("ADAMIC_BUILD_TRACE", "")
	same := Inputs{
		Name:      "keyed",
		Files:     []string{"source"},
		Flags:     []string{"--cache=" + r.cache},
		Toolchain: []string{"clang 20.1.8"},
	}
	type run struct {
		key, nameKey, writing, bytes string
	}
	runs := map[string]run{}
	for _, mode := range []string{"unset", "A", "B"} {
		test.Setenv("ADAMIC_BUILD_CACHE_DIR", r.cache)
		test.Setenv(location.OutputDirectoryVariable, "")
		if mode != "unset" {
			scratch, resolveSymlinksError := filepath.EvalSymlinks(test.TempDir())
			if resolveSymlinksError != nil {
				test.Fatal(resolveSymlinksError)
			}
			test.Setenv(location.OutputDirectoryVariable, filepath.Join(scratch, "output-"+mode))
		}
		output, outputError := location.Output()
		if outputError != nil {
			test.Fatal(outputError)
		}
		writing := same
		writing.Name = "writing"
		writing.Flags = []string{"--products=" + output, "--into=" + filepath.Join(output, "local")}
		var one run
		var keyError error
		one.key, keyError = Key(r.root, same)
		if keyError != nil {
			test.Fatal(keyError)
		}
		one.nameKey = NameKey(r.root, same)
		writingKey, keyError := Key(r.root, writing)
		if keyError != nil {
			test.Fatal(keyError)
		}
		one.writing = writingKey + " " + NameKey(r.root, writing)
		// A product of its own name in each run, so each builds where that run writes.
		relocated := Inputs{Name: "relocated " + mode, Files: []string{"source"}}
		directory := Product(test, relocated, func(directory string) error {
			return os.WriteFile(
				filepath.Join(directory, "names"),
				[]byte(Relative(filepath.Join(directory, "main.ts"))+"\n"+Relative(output)+"\n"),
				0o644,
			)
		})
		if mode != "unset" && !under(directory, output) {
			test.Fatalf("run %s built its product at %s, not where it writes", mode, directory)
		}
		content, readError := os.ReadFile(filepath.Join(directory, "names"))
		if readError != nil {
			test.Fatal(readError)
		}
		nameKey := NameKey(r.root, relocated)
		one.bytes = strings.ReplaceAll(string(content), nameKey, "<name key>")
		if back := Absolute("<build cache>/" + nameKey + "/main.ts"); back != filepath.Join(directory, "main.ts") {
			test.Errorf(
				"run %s reads its product's name back as %s, want %s",
				mode,
				back,
				filepath.Join(directory, "main.ts"),
			)
		}
		if mode != "unset" {
			raw := Inputs{Name: "raw " + mode, Files: []string{"source"}}
			_, getError := Get(raw, func(directory string) error {
				return os.WriteFile(
					filepath.Join(directory, "names"),
					[]byte(filepath.Join(output, "local")+"\n"),
					0o644,
				)
			})
			if getError == nil || !strings.Contains(getError.Error(), "holds a path of the machine that built it") {
				test.Errorf("run %s kept a product naming its output directory raw: %v", mode, getError)
			}
		}
		runs[mode] = one
	}
	for _, mode := range []string{"A", "B"} {
		if runs[mode] != runs["unset"] {
			test.Errorf(
				"with the output directory at run %s's, the inputs key and the product reads %+v; unset, %+v",
				mode,
				runs[mode],
				runs["unset"],
			)
		}
	}
	if want := "<build cache>/<name key>/main.ts\n<build cache>\n"; runs["unset"].bytes != want {
		test.Errorf("a product built in the cache reads %q, want %q", runs["unset"].bytes, want)
	}
}
