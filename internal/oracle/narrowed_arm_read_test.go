package oracle

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// A property narrowed to arm A of A | B, the narrowing kept by the checker across a call that put a B there,
// is read as A's (#8nj5pn4). In these programs a is A's alone and at one slot in every layout that holds it. On
// main before Emit's guard the uniform road (fields.go) took slot k with no shape check: it printed 1 for
// string_tags and enum_tags, B's string read as a number, where Node and the JavaScript backend print NaN, and it
// read out_of_bounds's a past the end of B's allocation, a heap-buffer-overflow under ASan, since a is A's seventh
// field and B has two. Emit's guard (uniformFieldGuard, emit_objects.go) made native stop there loudly while the
// JavaScript backend, which reads by name, printed Node's NaN. A narrowing kept across a call that may rewrite the
// property is now a claim the walk can't prove whatever the field's representation (Adamic's ruling on #p1p7k3g),
// so lowering reads it through a checked view of the declared type (narrowedArmView), and the read of a is the
// view's checked read, adamic_object_view in native's C, never the uniform road. Each lie stops loudly at that
// read with the view's own words, sanitized native, release native and the JavaScript backend alike, exit 70. Each
// honest twin writes an A, passes the check and agrees with Node on every backend, its leaks checked. Each lie's .a
// twin is refused at compile time (adamic/narrowed-arm), since .a has no checked view to read it through. Mutant M1
// (testdata/narrowed_arm_read/mutants.md), Emit's guard dropped, no longer reaches these reads; the superseded
// filter restored (narrowed_arm_property.mutants.md) takes the view away and the JavaScript backend prints NaN
// where the lies must stop.
func TestNarrowedArmReadStopsOrAgrees(t *testing.T) {
	t.Parallel()
	checked := regexp.MustCompile(`adamic_object_view\([^,]+, "a", `)
	for _, probe := range []struct {
		name, node, stop string
	}{
		{"out_of_bounds.ts", "NaN\n", "adamic: panic: field read failed: m.n.a is not initialized; expected number, found missing\n"},
		{"out_of_bounds_honest.ts", "61\nb\n", ""},
		{"string_tags.ts", "NaN\n", "adamic: panic: field read failed: m.n.a is not initialized; expected number, found missing\n"},
		{"string_tags_handed_on.ts", "NaN\n", "adamic: panic: field read failed: x.a is not initialized; expected number, found missing\n"},
		{"string_tags_honest.ts", "6\n", ""},
		{"enum_tags.ts", "NaN\n", "adamic: panic: field read failed: m.n.a is not initialized; expected number, found missing\n"},
		{"enum_tags_honest.ts", "6\n", ""},
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
			// The read of a is the checked view's, so a row that stops proves the view, not the uniform road's
			// guard beneath it.
			if !checked.MatchString(native.C(program)) {
				t.Fatal("the read of a isn't the checked view's")
			}
			want := node
			if probe.stop != "" {
				want = run{stderr: []byte(probe.stop), exitCode: 70}
			}
			sanitized, binary := nativelyUncached(t, program)
			var failures []string
			for index, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					failures = append(failures, fmt.Sprintf("%s on %s: exit %d, stdout %q, stderr %.2000q", difference, []string{"sanitized native", "release native", "the JavaScript backend"}[index], got.exitCode, got.stdout, got.stderr))
				}
			}
			if len(failures) > 0 {
				t.Fatal(strings.Join(failures, "; "))
			}
			if probe.stop == "" {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

// Each lie's .a twin, the same program in a file Adamic owns, is refused at compile time at the read of n (Adamic's
// ruling on #p1p7k3g), since .a reads no narrowing it can't prove through a checked view.
func TestNarrowedArmReadInAIsRefused(t *testing.T) {
	t.Parallel()
	const refused = ": Adamic 0.1 refuses a read of property n narrowed to A where a call may have rewritten it: the checker keeps the narrowing across the call but can't prove it (adamic/narrowed-arm); test the property again after the call, or read it into a const before the call and narrow that"
	for _, probe := range []struct{ name, where string }{
		{"out_of_bounds.a", "out_of_bounds.a:21:17"},
		{"string_tags.a", "string_tags.a:15:17"},
		{"string_tags_handed_on.a", "string_tags_handed_on.a:18:14"},
		{"enum_tags.a", "enum_tags.a:19:17"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/narrowed_arm_read", probe.name))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			if err == nil || !strings.HasSuffix(err.Error(), "/"+probe.where+refused) {
				t.Fatalf("want a refusal ending %q, got %v", "/"+probe.where+refused, err)
			}
		})
	}
}
