package oracle

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func init() {
	for _, fixture := range []struct {
		path    string
		checked bool
	}{
		{"scout_map_strong_edges.a", false},
		{"scout_multimap_composition.a", false},
		{"scout_array_to_map.a", false}, {"scout_array_to_map_invalidated.a", true},
		{"scout_map_presence.a", false}, {"scout_map_presence_invalidated.a", true},
		{"scout_map_brands.a", false}, {"scout_map_brand_boundary.a", true},
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/" + fixture.path, true, fixture.checked})
	}

	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"stage3/map-keys/iterator_pairs.a", true, false})
	for _, path := range []string{"scout_map_objects.a", "scout_map_numbers.a", "scout_map_references.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + path, true, false})
	}
}

// A lost lookup compiles and runs cleanly; source Node alone detects the changed result.
func TestScoutMapLookupMutants(t *testing.T) {
	for _, fixture := range []string{"scout_map_objects.a", "scout_map_numbers.a", "scout_map_references.a"} {
		t.Run(fixture, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := native.C(program)
			code := strings.ReplaceAll(original, "adamic_map_get(", "scout_missing_lookup(")
			if code == original {
				t.Fatal("mutant changed no lookup")
			}
			code = insertCollectionMutant(code, `static adamic_value *scout_missing_lookup(const adamic_map *map, adamic_value key) {
    (void)map;
    (void)key;
    return NULL;
}`)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if result.exitCode != 0 || len(result.stderr) != 0 {
				t.Fatalf("mutant must run cleanly: exit %d, stderr %s", result.exitCode, result.stderr)
			}
			if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
				t.Fatalf("want Node to catch lost lookup, got %q", difference)
			}
			t.Log("lost lookup: clean exit 0, no sanitizer or leak finding, caught by Node stdout")
		})
	}
}

func TestScoutMapPairInsertionMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "stage3/map-keys/iterator_pairs.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	original := native.C(program)
	code := strings.ReplaceAll(original, "adamic_map_set(", "scout_skip_pair(")
	if code == original {
		t.Fatal("mutant changed no insertion")
	}
	// This fixture's keys and values are runtime-built strings. The skipped store must consume
	// both owned arguments, so a semantic disagreement, rather than a leak, catches the mutant.
	code = insertCollectionMutant(code, `static size_t scout_pair_count;
static void scout_skip_pair(adamic_map *map, adamic_value key, adamic_value value) {
    scout_pair_count++;
    if (scout_pair_count == 2) {
        adamic_release(key.reference);
        adamic_release(value.reference);
        return;
    }
    adamic_map_set(map, key, value);
}`)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("mutant must run cleanly: exit %d, stderr %s", result.exitCode, result.stderr)
	}
	if difference := disagreement(onNode(t, path), result); difference != "stdout differs" {
		t.Fatalf("want Node to catch dropped pair, got %q", difference)
	}
	t.Log("dropped second pair: clean exit 0, no sanitizer or leak finding, caught by Node stdout")
}

// Removing the string boundary check must yield valid code, not a clang failure.
func TestScoutMapBrandBoundaryMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_map_brand_boundary.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := onJavaScriptBackend(t, program)
	if want.exitCode != 70 || string(want.stdout) != "boundary:1\n" || string(want.stderr) != "adamic: panic: brand boundary failed: expected string for __String\n" {
		t.Fatalf("wrong boundary contract: %d %q %q", want.exitCode, want.stdout, want.stderr)
	}
	changed := false
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Name != "checked_collection_brand" {
			continue
		}
		function.Body = function.Body[1:]
		changed = true
	}
	if !changed {
		t.Fatal("mutant removed no boundary check")
	}
	mutant, binary := nativelyUncached(t, program)
	if mutant.exitCode != 0 || len(mutant.stderr) != 0 {
		t.Fatalf("mutant must finish without sanitizer failure: %d %s", mutant.exitCode, mutant.stderr)
	}
	if difference := disagreement(want, mutant); difference == "" {
		t.Fatal("missing brand check survived")
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	if actual := onJavaScriptBackend(t, program); actual.exitCode != 0 {
		t.Fatalf("JavaScript mutant did not finish: %d %s", actual.exitCode, actual.stderr)
	}
	t.Log("missing brand check: both mutant backends exit 0, native sanitizer/leaks clean; expected panic exit 70 catches omission")
}

func TestScoutMapPresenceInvalidationContract(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_map_presence_invalidated.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stdout: []byte("reading\n"), stderr: []byte("adamic: panic: collection lookup failed: map.get('key') is undefined\n")}
	actual, binary := nativelyUncached(t, program)
	if actual.exitCode == 0 {
		if report := leaks(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
	for backend, result := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Fatalf("%s invalidated proof: %s; exit %d stdout %q stderr %q", backend, difference, result.exitCode, result.stdout, result.stderr)
		}
	}
}

func TestScoutArrayRangeInvalidationContract(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_array_to_map_invalidated.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	want := run{exitCode: 70, stdout: []byte("reading\n"), stderr: []byte("adamic: panic: collection lookup failed: value is undefined\n")}
	actual, binary := nativelyUncached(t, program)
	if actual.exitCode == 0 {
		if report := leaks(t, program, binary); report != "" {
			t.Fatal(report)
		}
	}
	for backend, result := range map[string]run{"native": actual, "JavaScript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, result); difference != "" {
			t.Fatalf("%s invalidated proof: %s; exit %d stdout %q stderr %q", backend, difference, result.exitCode, result.stdout, result.stderr)
		}
	}
}

func TestScoutMultiMapEmptyBucketMutant(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_multimap_composition.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	original := native.C(program)
	code := strings.ReplaceAll(original, "adamic_map_delete(", "scout_keep_empty_bucket(")
	if code == original {
		t.Fatal("mutant changed no deletion")
	}
	code = insertCollectionMutant(code, `static bool scout_keep_empty_bucket(adamic_map *map, adamic_value key) {(void)map;(void)key;return false;}`)
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if actual.exitCode != 0 || len(actual.stderr) != 0 {
		t.Fatalf("want clean semantic mutant, got %d %s", actual.exitCode, actual.stderr)
	}
	if difference := disagreement(onNode(t, path), actual); difference != "stdout differs" {
		t.Fatalf("want original Node to catch retained empty bucket, got %q", difference)
	}
	t.Log("retained empty bucket: clean native exit 0, ASAN/UBSAN/leaks clean; source Node catches wrong size")
}

func TestScoutCollectionStrongEdgeMutants(t *testing.T) {
	for _, edge := range []string{"key", "value"} {
		t.Run(edge, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/scout_map_strong_edges.a"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			original := native.C(program)
			code := strings.ReplaceAll(original, "adamic_map_set(", "scout_drop_owned_edge(")
			if original == code {
				t.Fatal("no map insertion mutated")
			}
			code = insertCollectionMutant(code, `static void scout_drop_owned_edge(adamic_map *map, adamic_value key, adamic_value value) {adamic_map_set(map,key,value);adamic_release(`+edge+`.reference);}`)
			binary := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
			if actual.exitCode == 0 || !strings.Contains(string(actual.stderr), "heap-use-after-free") {
				t.Fatalf("want ASAN to catch lost strong %s edge, got %d %s", edge, actual.exitCode, actual.stderr)
			}
			t.Log("lost strong " + edge + " edge caught by ASAN heap-use-after-free after producer returns")
		})
	}
}
