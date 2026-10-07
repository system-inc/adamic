package oracle

import (
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// TestNativeSlabsAgreeWithNode gives the release allocator its own sanitizer shard.
// Native observations are always fresh. The existing malloc lane still checks leaks:
// slab chunks stay reachable, so LeakSanitizer cannot account for individual slots.
func TestNativeSlabsAgreeWithNode(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("Linux is the gate of record; macOS Node fuses multiply-adds, so navigation.a differs there (#myatdyv), and this lane gets its own Linux shard")
	}
	if runtime.GOOS != "linux" {
		t.Skip("the slab lane has its own Linux gate shard")
	}
	t.Parallel()
	allocatorAgreesWithNode(t, true)
}

// This optional control measures the same work with the existing malloc allocator.
func TestNativeMallocAgreesWithNode(t *testing.T) {
	if os.Getenv("ADAMIC_SLAB_MEASURE") != "1" {
		t.Skip("set ADAMIC_SLAB_MEASURE=1 for the timing control")
	}
	t.Parallel()
	allocatorAgreesWithNode(t, false)
}

func allocatorAgreesWithNode(t *testing.T, slabs bool) {
	t.Helper()
	for _, fixture := range fixtures {
		if !fixture.lowers {
			continue
		}
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			expected := onNode(t, path)
			if fixture.checked {
				expected = onJavaScriptBackend(t, program)
			}
			binary := filepath.Join(t.TempDir(), "program")
			if err := native.Build(native.C(program), binary, native.Options{Sanitize: true, Slabs: slabs}); err != nil {
				t.Fatal(err)
			}
			actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
			if difference := disagreement(expected, actual); difference != "" {
				t.Errorf("%s\nNode: exit %d, stdout %q, stderr %q\nnative: exit %d, stdout %q, stderr %q", difference, expected.exitCode, expected.stdout, expected.stderr, actual.exitCode, actual.stdout, actual.stderr)
			}
		})
	}
}

// A real fixture's runtime-built string loses its last count before its next use.
// Both allocators must catch this; poisoning does not make early release slab-only.
func TestSlabLaneCatchesEarlyRelease(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/string_append.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	doubled, needle := doubledSelfAppend(t, source)
	source = strings.Replace(source, needle, "adamic_release("+doubled+");\n\t"+needle, 1)
	for _, slabs := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(source, binary, native.Options{Sanitize: true, Slabs: slabs}); err != nil {
			t.Fatal(err)
		}
		actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
		report := "heap-use-after-free"
		if slabs {
			report = "use-after-poison"
		}
		if actual.exitCode == 0 || !strings.Contains(string(actual.stderr), "ERROR: AddressSanitizer: "+report) {
			t.Fatalf("slabs=%v: early-release mutant escaped: exit %d\n%s", slabs, actual.exitCode, actual.stderr)
		}
		t.Logf("string_append.a slabs=%v caught %s", slabs, report)
	}
}

// This constructed mutant drops a real fixture value's count only after its address
// was recycled. Malloc's ASan quarantine prevents that path; slabs take it immediately.
func TestSlabLaneCatchesRecycledRelease(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/string_append.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	doubled, use := doubledSelfAppend(t, source)
	allocation := doubledInitialization(t, source, doubled)
	source = strings.Replace(source, allocation, "adamic_string *recycle_probe = adamic_string_allocate(7);\n\tuintptr_t recycled_address = (uintptr_t)recycle_probe;\n\tadamic_release(recycle_probe);\n\t"+allocation, 1)
	source = strings.Replace(source, use, "if ((uintptr_t)"+doubled+" == recycled_address) { adamic_release("+doubled+"); }\n\t"+use, 1)
	expected := onNode(t, path)
	for _, slabs := range []bool{false, true} {
		binary := filepath.Join(t.TempDir(), "mutant")
		if err := native.Build(source, binary, native.Options{Sanitize: true, Slabs: slabs}); err != nil {
			t.Fatal(err)
		}
		actual := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
		if slabs {
			if actual.exitCode == 0 || !strings.Contains(string(actual.stderr), "ERROR: AddressSanitizer: use-after-poison") {
				t.Fatalf("recycled-release mutant escaped: exit %d\n%s", actual.exitCode, actual.stderr)
			}
			t.Log("string_append.a: recycled-release mutant caught by slabs: use-after-poison")
		} else {
			if difference := disagreement(expected, actual); difference != "" {
				t.Fatalf("malloc control differs: %s\n%s", difference, actual.stderr)
			}
			t.Log("string_append.a: recycled-release mutant passes malloc byte for byte")
		}
	}
}

// The mutants find string_append.a's doubling() by the shape of its C, not by the emitter's numbers,
// which move whenever anything before it lowers differently.
var (
	selfAppend     = regexp.MustCompile(`(adamic_local_\d+_doubled) = adamic_string_append\((adamic_local_\d+_doubled),`)
	initialization = regexp.MustCompile(`(adamic_string \* (adamic_temporary_\d+) = adamic_string_concat\()[^\n]*\n\s*adamic_string \* (adamic_local_\d+_doubled) = (adamic_temporary_\d+);`)
)

// doubledSelfAppend finds `doubled += doubled`: the one statement that appends a local named
// doubled to itself. It returns the local and the statement's text up to its first argument.
func doubledSelfAppend(t *testing.T, source string) (string, string) {
	t.Helper()
	var sites [][]string
	for _, match := range selfAppend.FindAllStringSubmatch(source, -1) {
		if match[1] == match[2] {
			sites = append(sites, match)
		}
	}
	if len(sites) != 1 {
		t.Fatalf("mutant insertion point: want exactly one `adamic_local_N_doubled = adamic_string_append(adamic_local_N_doubled,` (doubled += doubled), found %d", len(sites))
	}
	if count := strings.Count(source, sites[0][0]); count != 1 {
		t.Fatalf("mutant insertion point: %q appears %d times", sites[0][0], count)
	}
	return sites[0][1], sites[0][0]
}

// doubledInitialization finds the string_concat temporary that builds doubled's first value: the
// declaration the very next statement declares doubled from. It returns that declaration up to its
// arguments.
func doubledInitialization(t *testing.T, source string, doubled string) string {
	t.Helper()
	var sites []string
	for _, match := range initialization.FindAllStringSubmatch(source, -1) {
		if match[3] == doubled && match[2] == match[4] {
			sites = append(sites, match[1])
		}
	}
	if len(sites) != 1 {
		t.Fatalf("mutant insertion point: want exactly one `adamic_string * adamic_temporary_N = adamic_string_concat(...);` followed by `adamic_string * %s = adamic_temporary_N;`, found %d", doubled, len(sites))
	}
	if count := strings.Count(source, sites[0]); count != 1 {
		t.Fatalf("mutant insertion point: %q appears %d times", sites[0], count)
	}
	return sites[0]
}
