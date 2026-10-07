package oracle

import (
	"os"
	"path/filepath"
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
	needle := "adamic_local_17_doubled = adamic_string_append(adamic_local_17_doubled,"
	if strings.Count(source, needle) != 1 {
		t.Fatal("mutant insertion point changed")
	}
	source = strings.Replace(source, needle, "adamic_release(adamic_local_17_doubled);\n\t"+needle, 1)
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
	allocation := "adamic_string * adamic_temporary_27 = adamic_string_concat("
	use := "adamic_local_17_doubled = adamic_string_append(adamic_local_17_doubled,"
	if strings.Count(source, allocation) != 1 || strings.Count(source, use) != 1 {
		t.Fatal("mutant insertion points changed")
	}
	source = strings.Replace(source, allocation, "adamic_string *recycle_probe = adamic_string_allocate(7);\n\tuintptr_t recycled_address = (uintptr_t)recycle_probe;\n\tadamic_release(recycle_probe);\n\t"+allocation, 1)
	source = strings.Replace(source, use, "if ((uintptr_t)adamic_local_17_doubled == recycled_address) { adamic_release(adamic_local_17_doubled); }\n\t"+use, 1)
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
