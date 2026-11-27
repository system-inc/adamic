package oracle

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// narrowedArmMissingField is what native says where an object's shape lacks a field lowering proved is
// there (adamic_object_find, object.c).
const narrowedArmMissingField = "adamic: panic: NotYet: a field missing from this object, where a narrowing kept across a call that rewrote the property said it is there (#8nj5pn4)\n"

// A property narrowed to arm A of A | B, the narrowing kept by the checker across a call that put a B there,
// is read as A's (#8nj5pn4). In these programs a is A's alone and at one slot in every layout that holds it, so
// native reads it on the uniform road (fields.go). On main that road took slot k with no shape check: it printed
// 1 for string_tags and enum_tags, B's string read as a number, where Node and the JavaScript backend print NaN,
// and it read out_of_bounds's a past the end of B's allocation, a heap-buffer-overflow under ASan, since a is A's
// seventh field and B has two. The slot now sits behind a check of the object's shape against the literal
// layouts that hold a (uniformFieldGuard, emit_objects.go), so B's object takes the by-name road and stops
// loudly, sanitized and release alike, while the JavaScript backend, which reads by name, prints Node's NaN.
// Each honest twin writes an A, takes the slot, and agrees with Node on every backend, its leaks checked.
// Mutant M1 (testdata/narrowed_arm_read/mutants.md): the guard dropped, and out_of_bounds goes red under ASan
// while the tag rows print 1 again.
func TestNarrowedArmReadStopsOrAgrees(t *testing.T) {
	t.Parallel()
	uniform := regexp.MustCompile(`ADAMIC_UNIFORM_OBJECT\(\w+, "a"\)->slots\[\d+\]`)
	for _, probe := range []struct {
		name, node string
		stops      bool
	}{
		{"out_of_bounds.ts", "NaN\n", true},
		{"out_of_bounds_honest.ts", "61\nb\n", false},
		{"string_tags.ts", "NaN\n", true},
		{"string_tags_handed_on.ts", "NaN\n", true},
		{"string_tags_honest.ts", "6\n", false},
		{"enum_tags.ts", "NaN\n", true},
		{"enum_tags_honest.ts", "6\n", false},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/narrowed_arm_read", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if difference := disagreement(run{stdout: []byte(probe.node)}, node); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			// The read of a is the uniform road's, so a row that stops proves the guard, not a by-name read
			// that never had the hole.
			if !uniform.MatchString(native.C(program)) {
				t.Fatal("the read of a doesn't take the uniform road")
			}
			want := node
			if probe.stops {
				want = run{stderr: []byte(narrowedArmMissingField), exitCode: 70}
			}
			sanitized, binary := nativelyUncached(t, program)
			for index, got := range []run{sanitized, releasedUncached(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s on %s: exit %d, stdout %q, stderr %.2000q", difference, []string{"sanitized native", "release native"}[index], got.exitCode, got.stdout, got.stderr)
				}
			}
			if got := onJavaScriptBackend(t, program); disagreement(node, got) != "" {
				t.Fatalf("%s on the JavaScript backend: exit %d, stdout %q, stderr %q", disagreement(node, got), got.exitCode, got.stdout, got.stderr)
			}
			if !probe.stops {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}
