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
func makeReadOnly(t *testing.T, directory string) {
	t.Helper()
	var directories []string
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err == nil && entry.IsDir() {
			directories = append(directories, path)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		for _, each := range directories {
			os.Chmod(each, 0o755)
		}
	})
	for _, each := range directories {
		if err := os.Chmod(each, 0o555); err != nil {
			t.Fatal(err)
		}
	}
}

// cacheListing is every path under directory with its mode, size and modification time: what a write anywhere in it
// changes, even where the file system lets the write through (a test running as root).
func cacheListing(t *testing.T, directory string) map[string]string {
	t.Helper()
	listing := map[string]string{}
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		listing[path] = fmt.Sprintf("%s %d %d", info.Mode(), info.Size(), info.ModTime().UnixNano())
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return listing
}

// exists fails the test unless path is there.
func exists(t *testing.T, path, what string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("%s isn't at %s: %v", what, path, err)
	}
}

// With ADAMIC_BUILD_OUTPUT_DIR set, a build cache this process can't write is only read: every product it already holds
// is found there and never built (a settled one by its read sets, one built for this machine alone, and its resolved
// copy is made beside it in the output directory), a product only the store holds is fetched into the output directory,
// and every product built, pointed to, digested, valued or settled lands in the output directory, at the place it would
// have in the cache. Nothing in the cache changes, a mode, a size, a time, and nothing in the output directory is a
// link. Mutants, each caught: any one write site (filled, pointLocal, rerecord's read sets, productDigest, systemValue's
// machine value, Resolved's copy, a fetch, Settle's placement, its read sets and its .inputs) left writing in the cache.
// Not parallel: newRig, and the output directory is a variable of the process (t.Setenv).
func TestAnOutputDirectoryTakesEveryWriteAndTheCacheIsOnlyRead(t *testing.T) {
	r := newRig(t)
	refuseBuild := func(string) error {
		t.Error("built a product the read-only cache holds")
		return errors.New("built a product the read-only cache holds")
	}
	// The cache as Loom ships it: a settled product with its read sets, whose set stored a key it no longer gives, so a
	// lookup records it again; a product built for this machine alone, whose text names the tree; a read set naming a
	// product only the store holds.
	Product(t, port, building("built"))
	build := r.built(t)["port"]
	r.window(build.Build, func() { r.open(r.pid, "source/a.c") })
	settlement := r.settle(t)
	if len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	file, err := loadReads(r.cache, build.NameKey)
	if err != nil || len(file.Sets) != 1 {
		t.Fatalf("port's read sets: %+v, %v", file, err)
	}
	settledKey := file.Sets[0].Key
	settled := filepath.Join(r.cache, settledKey)
	if err = withReads(r.cache, build.NameKey, func(file *readsFile) error {
		file.Sets[0].Key = strings.Repeat("0", 64)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	forgetHeldForLife()
	local := Inputs{Name: "local", Files: []string{"source"}}
	built := Product(t, local, building("<repository>/source/a.c\n"))
	if !under(built, filepath.Join(r.cache, "local")) {
		t.Fatalf("a product for this machine alone is at %s, not in the cache's local directory", built)
	}
	stored := Inputs{Name: "stored", Files: []string{"source"}}
	storedNameKey := NameKey(r.root, stored)
	storedSet := readSet{Entries: []readEntry{{Kind: "content", Path: "source/b.c"}}}
	evaluated, err := evaluate(r.root, r.cache, storedNameKey, storedSet, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	storedSet.Key = evaluated.key
	if err = recordReads(r.cache, storedNameKey, stored.Name, storedSet); err != nil {
		t.Fatal(err)
	}

	makeReadOnly(t, r.cache)
	before := cacheListing(t, r.cache)
	scratch, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(scratch, "output")
	t.Setenv(location.OutputDirectoryVariable, output)
	forgetHeldForLife()

	// Found, never built: by its read sets, and for this machine alone.
	if found, err := Get(port, refuseBuild); err != nil || found != settled {
		t.Fatalf("the settled product was found at %q (%v), want %s", found, err, settled)
	}
	exists(t, readsPath(output, build.NameKey), "the set that found port under another key, recorded again")
	if again, err := loadReads(output, build.NameKey); err != nil || len(again.Sets) == 0 || again.Sets[0].Key != settledKey {
		t.Errorf("port's read sets read %+v, %v; want the set that found it first, under %s", again, err, settledKey[:12])
	}
	if found, err := Get(local, refuseBuild); err != nil || found != built {
		t.Fatalf("the local product was found at %q (%v), want %s", found, err, built)
	}
	// Found before anything is made or locked: its hit leaves no trace in the output directory either.
	if _, err := os.Stat(filepath.Join(output, "local")); err == nil {
		t.Error("a hit in the cache made the output directory's local directory")
	}
	// Fetched into the output directory.
	store := &fakeStore{objects: map[string][]byte{}}
	server := httptest.NewServer(store)
	t.Cleanup(server.Close)
	store.put(t, storedSet.Key, map[string]string{"product": "stored"})
	t.Setenv("ADAMIC_BUILD_STORE", server.URL)
	t.Setenv("ADAMIC_BUILD_AUDIT", "0")
	if fetched, err := Get(stored, refuseBuild); err != nil || fetched != filepath.Join(output, storedSet.Key) {
		t.Fatalf("the stored product was fetched to %q (%v), want %s", fetched, err, filepath.Join(output, storedSet.Key))
	}
	t.Setenv("ADAMIC_BUILD_STORE", "off")
	// Built, pointed to, resolved, digested and valued in the output directory.
	fresh := Inputs{Name: "fresh", Files: []string{"source"}}
	made := Product(t, fresh, building("fresh"))
	if !under(made, filepath.Join(output, "local")) {
		t.Fatalf("a new product is at %s, not in the output directory's local directory", made)
	}
	exists(t, localPointer(output, NameKey(r.root, fresh)), "the new product's local pointer")
	copied, err := Resolved(built)
	if err != nil {
		t.Fatalf("resolving a product the read-only cache holds: %v", err)
	}
	if !under(copied, filepath.Join(output, "local")) {
		t.Fatalf("the resolved copy is at %s, not beside the product's place in the output directory", copied)
	}
	if content, err := os.ReadFile(filepath.Join(copied, "product")); err != nil || string(content) != r.root+"/source/a.c\n" {
		t.Fatalf("the resolved copy reads %q, %v", content, err)
	}
	if _, err := productDigest(settled); err != nil {
		t.Fatal(err)
	}
	exists(t, filepath.Join(output, settledKey+".digest"), "the settled product's digest")
	machineFile := filepath.Join(scratch, "machine-file")
	write(t, scratch, "machine-file", "a machine file\n")
	systemValue(machineFile)
	if values, err := os.ReadDir(filepath.Join(output, "machine")); err != nil || len(values) != 1 {
		t.Errorf("the machine file's value kept in the output directory: %d, %v", len(values), err)
	}

	// Settled in the output directory: a traced build reading the source and the product only the cache holds.
	r.retrace(t, "output")
	forgetHeldForLife()
	if found := Product(t, port, refuseBuild); found != settled {
		t.Fatalf("under the trace the settled product was found at %s, want %s", found, settled)
	}
	consumer := Inputs{Name: "consumer", Files: []string{"source"}}
	pending := Product(t, consumer, building("consumer"))
	if !under(pending, filepath.Join(output, "pending")) {
		t.Fatalf("a traced build is pending at %s, not in the output directory", pending)
	}
	consuming := r.built(t)["consumer"]
	r.window(consuming.Build, func() {
		r.open(r.pid, "source/b.c")
		r.open(r.pid, filepath.Join(settled, "product"))
	})
	settlement = r.settle(t)
	if len(settlement.Settled) != 1 || len(settlement.Refused) != 0 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
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
		t.Errorf("the run's reports don't say it found port in the cache under %s: %+v", settledKey[:12], settlement.Reports)
	}
	if consumerKey == "" {
		t.Fatalf("no settled report for consumer: %+v", settlement.Reports)
	}
	exists(t, filepath.Join(output, consumerKey, "product"), "the settled product")
	exists(t, filepath.Join(output, consumerKey+".inputs"), "the settled product's inputs")
	exists(t, readsPath(output, consuming.NameKey), "the settled product's read sets")
	if entries := r.entries(t, consuming.NameKey); !strings.Contains(strings.Join(entries, "\n"), "product "+build.NameKey) {
		t.Errorf("consumer's read set doesn't read port, which only the cache holds: %q", entries)
	}
	if _, err := os.Stat(filepath.Join(r.cache, consumerKey)); err == nil {
		t.Error("the settled product is in the read-only cache")
	}

	after := cacheListing(t, r.cache)
	for path, was := range before {
		if now, ok := after[path]; !ok || now != was {
			t.Errorf("%s changed in the read-only cache: %q, now %q", path, was, now)
		}
	}
	for path := range after {
		if _, ok := before[path]; !ok {
			t.Errorf("%s appeared in the read-only cache", path)
		}
	}
	filepath.WalkDir(output, func(path string, entry fs.DirEntry, err error) error {
		if err == nil && entry.Type()&fs.ModeSymlink != 0 {
			t.Errorf("the output directory holds a link: %s", path)
		}
		return err
	})
}

// A key and a product's bytes never say where they were written: the same inputs key alike, Key and NameKey, with
// ADAMIC_BUILD_OUTPUT_DIR unset, set to one directory and set to another, and inputs naming the directory each run
// writes to key as if each named the cache; a product built there names itself and the cache as one built in the cache
// does, and reads back to where it is; a product whose text names the output directory raw is refused, as one naming
// the cache is. Mutants: the output directory left out of portable's places, Relative's places or relocatable's, or
// taken as an ADAMIC_ location.
// Not parallel: newRig, and the output directory is a variable of the process (t.Setenv).
func TestTheOutputDirectoryKeysAndRelocatesAsTheCache(t *testing.T) {
	r := newRig(t)
	t.Setenv("ADAMIC_BUILD_TRACE", "")
	same := Inputs{Name: "keyed", Files: []string{"source"}, Flags: []string{"--cache=" + r.cache}, Toolchain: []string{"clang 20.1.8"}}
	type run struct {
		key, nameKey, writing, bytes string
	}
	runs := map[string]run{}
	for _, mode := range []string{"unset", "A", "B"} {
		t.Setenv("ADAMIC_BUILD_CACHE_DIR", r.cache)
		t.Setenv(location.OutputDirectoryVariable, "")
		if mode != "unset" {
			scratch, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			t.Setenv(location.OutputDirectoryVariable, filepath.Join(scratch, "output-"+mode))
		}
		output, err := location.Output()
		if err != nil {
			t.Fatal(err)
		}
		writing := same
		writing.Name = "writing"
		writing.Flags = []string{"--products=" + output, "--into=" + filepath.Join(output, "local")}
		var one run
		one.key, err = Key(r.root, same)
		if err != nil {
			t.Fatal(err)
		}
		one.nameKey = NameKey(r.root, same)
		writingKey, err := Key(r.root, writing)
		if err != nil {
			t.Fatal(err)
		}
		one.writing = writingKey + " " + NameKey(r.root, writing)
		// A product of its own name in each run, so each builds where that run writes.
		relocated := Inputs{Name: "relocated " + mode, Files: []string{"source"}}
		directory := Product(t, relocated, func(directory string) error {
			return os.WriteFile(filepath.Join(directory, "names"), []byte(Relative(filepath.Join(directory, "main.ts"))+"\n"+Relative(output)+"\n"), 0o644)
		})
		if mode != "unset" && !under(directory, output) {
			t.Fatalf("run %s built its product at %s, not where it writes", mode, directory)
		}
		content, err := os.ReadFile(filepath.Join(directory, "names"))
		if err != nil {
			t.Fatal(err)
		}
		nameKey := NameKey(r.root, relocated)
		one.bytes = strings.ReplaceAll(string(content), nameKey, "<name key>")
		if back := Absolute("<build cache>/" + nameKey + "/main.ts"); back != filepath.Join(directory, "main.ts") {
			t.Errorf("run %s reads its product's name back as %s, want %s", mode, back, filepath.Join(directory, "main.ts"))
		}
		if mode != "unset" {
			raw := Inputs{Name: "raw " + mode, Files: []string{"source"}}
			_, err := Get(raw, func(directory string) error {
				return os.WriteFile(filepath.Join(directory, "names"), []byte(filepath.Join(output, "local")+"\n"), 0o644)
			})
			if err == nil || !strings.Contains(err.Error(), "holds a path of the machine that built it") {
				t.Errorf("run %s kept a product naming its output directory raw: %v", mode, err)
			}
		}
		runs[mode] = one
	}
	for _, mode := range []string{"A", "B"} {
		if runs[mode] != runs["unset"] {
			t.Errorf("with the output directory at run %s's, the inputs key and the product reads %+v; unset, %+v", mode, runs[mode], runs["unset"])
		}
	}
	if want := "<build cache>/<name key>/main.ts\n<build cache>\n"; runs["unset"].bytes != want {
		t.Errorf("a product built in the cache reads %q, want %q", runs["unset"].bytes, want)
	}
}
