package buildcache

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// resolvedLookedUp is nameKey's product as a new process's lookup answers it, holding nothing an earlier lookup or map
// read: the test changes the map and the cache between lookups, which a resolved unit's runner never does.
func resolvedLookedUp(t *testing.T, r *rig, nameKey string) found {
	t.Helper()
	forgetHeldForLife()
	resolvedMaps.Clear()
	looked, err := productFor(r.root, r.cache, nameKey, 0, nil)
	if err != nil {
		t.Fatalf("name key %s: %v", nameKey[:12], err)
	}
	return looked
}

// resolvedMapAt lays a map as Loom's runner writes it, its content given whole, and points ADAMIC_BUILD_RESOLVED at it.
func resolvedMapAt(t *testing.T, content string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "resolved.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv(resolvedVariable, path)
}

// resolvedNaming is a version 1 map naming each name key's product key, its sum a stand-in buildcache never reads.
func resolvedNaming(products map[string]string) string {
	var entries []string
	for nameKey, key := range products {
		entries = append(entries, fmt.Sprintf(`%q:{"key":%q,"sum":%q}`, nameKey, key, strings.Repeat("5", 64)))
	}
	return `{"version":1,"products":{` + strings.Join(entries, ",") + `}}`
}

// A resolved lookup finds exactly the product the index would pick, from the map alone: with every read set gone it
// still answers, and a map naming another laid product answers that one, so the map decides and never the index. A name
// key the map lacks, a key it names that isn't laid, and a map that doesn't read are each a miss with its reason, even
// when the read sets would have found the product. Mutants: the map consulted but the read sets answering a name it
// holds; a miss falling back to the read sets; the laid check dropped.
// Not parallel: newRig, and t.Setenv for ADAMIC_BUILD_RESOLVED.
func TestAResolvedLookupFindsTheProductTheIndexWouldPick(t *testing.T) {
	r := newRig(t)
	Product(t, port, building("port"))
	notes := Inputs{Name: "notes"}
	Product(t, notes, building("notes"))
	built := r.built(t)
	r.window(built["port"].Build, func() { r.open(r.pid, "source/a.c") })
	r.window(built["notes"].Build, func() { r.open(r.pid, "notes.txt") })
	if settlement := r.settle(t); len(settlement.Settled) != 2 {
		t.Fatalf("settled %q, refused %q", settlement.Settled, settlement.Refused)
	}
	portNameKey, notesNameKey := built["port"].NameKey, built["notes"].NameKey
	for _, nameKey := range []string{portNameKey, notesNameKey} {
		if _, err := os.Stat(readsPath(r.cache, nameKey)); err != nil {
			t.Fatalf("the settled build recorded no read sets: %v", err)
		}
	}

	// The index's answers, with the variable unset, are what a resolved lookup must give.
	t.Setenv(resolvedVariable, "")
	portIndexed, notesIndexed := resolvedLookedUp(t, r, portNameKey), resolvedLookedUp(t, r, notesNameKey)
	if portIndexed.directory == "" || notesIndexed.directory == "" || portIndexed.key == notesIndexed.key {
		t.Fatalf("the index found port %+v and notes %+v; want two laid products", portIndexed, notesIndexed)
	}
	if portIndexed.directory != filepath.Join(r.cache, portIndexed.key) {
		t.Fatalf("the index found port at %s, not under its key %s", portIndexed.directory, portIndexed.key)
	}

	resolvedMapAt(t, resolvedNaming(map[string]string{portNameKey: portIndexed.key, notesNameKey: notesIndexed.key}))
	for _, want := range []struct {
		nameKey string
		found   found
	}{{portNameKey, portIndexed}, {notesNameKey, notesIndexed}} {
		if looked := resolvedLookedUp(t, r, want.nameKey); looked.key != want.found.key || looked.directory != want.found.directory {
			t.Fatalf("the map found %s at %q, %q (%s); the index found %s at %s", want.nameKey[:12], looked.key, looked.directory, looked.why, want.found.key, want.found.directory)
		}
	}

	// A map naming port's name key to notes's product answers notes's: the index would have said port's.
	resolvedMapAt(t, resolvedNaming(map[string]string{portNameKey: notesIndexed.key}))
	if looked := resolvedLookedUp(t, r, portNameKey); looked.key != notesIndexed.key || looked.directory != notesIndexed.directory {
		t.Fatalf("a map naming notes's product for port found %q, %q (%s); want %s", looked.key, looked.directory, looked.why, notesIndexed.key)
	}

	// A name key the map lacks misses, though its read sets are still there and the index would find it.
	resolvedMapAt(t, resolvedNaming(map[string]string{notesNameKey: notesIndexed.key}))
	if looked := resolvedLookedUp(t, r, portNameKey); looked.directory != "" || !strings.Contains(looked.why, "the resolved map names no product") {
		t.Fatalf("a name key the map lacks found %q (%s); want a miss naming the map", looked.directory, looked.why)
	}

	// A key the map names that isn't laid in the cache misses rather than handing out a path to nothing.
	unlaid := strings.Repeat("ab", 32)
	resolvedMapAt(t, resolvedNaming(map[string]string{portNameKey: unlaid}))
	if looked := resolvedLookedUp(t, r, portNameKey); looked.directory != "" || !strings.Contains(looked.why, "isn't laid") {
		t.Fatalf("a map naming an unlaid key found %q (%s); want a miss", looked.directory, looked.why)
	}

	// A map that doesn't read misses every lookup, even a name key it names well, and never reads the index instead.
	for name, content := range map[string]string{
		"an unknown field": `{"version":1,"products":{"` + portNameKey + `":{"key":"` + portIndexed.key + `","sum":"","size":1}}}`,
		"version 2":        strings.Replace(resolvedNaming(map[string]string{portNameKey: portIndexed.key}), `"version":1`, `"version":2`, 1),
		"a key not 64 hex": resolvedNaming(map[string]string{portNameKey: portIndexed.key, notesNameKey: "../" + notesIndexed.key}),
	} {
		resolvedMapAt(t, content)
		for _, nameKey := range []string{portNameKey, notesNameKey} {
			if looked := resolvedLookedUp(t, r, nameKey); looked.directory != "" || !strings.Contains(looked.why, "the resolved map doesn't read") {
				t.Errorf("%s: %s found %q (%s); want a miss saying the map doesn't read", name, nameKey[:12], looked.directory, looked.why)
			}
		}
	}

	// With the read sets gone, the map alone still answers: the lookup never needed them. Only the files this test's
	// builds recorded are removed.
	for _, nameKey := range []string{portNameKey, notesNameKey} {
		if err := os.Remove(readsPath(r.cache, nameKey)); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv(resolvedVariable, "")
	if looked := resolvedLookedUp(t, r, portNameKey); looked.directory != "" {
		t.Fatalf("the index found port at %s with its read sets removed", looked.directory)
	}
	resolvedMapAt(t, resolvedNaming(map[string]string{portNameKey: portIndexed.key, notesNameKey: notesIndexed.key}))
	for _, want := range []struct {
		nameKey string
		found   found
	}{{portNameKey, portIndexed}, {notesNameKey, notesIndexed}} {
		if looked := resolvedLookedUp(t, r, want.nameKey); looked.key != want.found.key || looked.directory != want.found.directory {
			t.Fatalf("with no read sets the map found %s at %q, %q (%s); want %s at %s", want.nameKey[:12], looked.key, looked.directory, looked.why, want.found.key, want.found.directory)
		}
	}
}
