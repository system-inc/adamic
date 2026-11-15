package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// looseSlotMainLeakProbes are the stop's roads (#kaz4pk8's g01 to k18 and Language's l roads that reach the
// stop), copied from compiler/loose-slot-cycles-3 onto main. A scratch probe, never merged: it's Compiler's
// condition 2 as Adamic set it, the leak on main witnessed by the counted build, never LeakSanitizer.
var looseSlotMainLeakProbes = []string{
	"g01_field_unknown", "g02_field_object", "g06_array_of_unknown", "g11_unknown_plain_cycle", "g12_object_plain_cycle",
	"g13_self_unknown", "k01_generic_t", "k03_union_string_object", "k04_record_unknown", "k06_set_object",
	"k10_index_signature", "k11_object_array", "k14_default_parameter", "k17_captured_unknown",
	"k18_entry_then_declared", "l01_upcast_value", "l02_upcast_entry", "l03_two_hops", "l05_push_unknown",
	"l08_generic_t", "l09_three_hops",
}

// TestLooseSlotMainLeakProbe lowers each road on main and runs its counted build, one LEAK line each with
// the counts and leakcheck.Unbalanced's verdict, or a STOP line when main stops it. It fails only when a
// probe can't be measured.
func TestLooseSlotMainLeakProbe(t *testing.T) {
	t.Parallel()
	for _, name := range looseSlotMainLeakProbes {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/loose_slot_cycles", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), name+".ts")
			if err := os.WriteFile(path, source, 0o644); err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Logf("STOP %s %v", name, err)
				return
			}
			binary := filepath.Join(t.TempDir(), "counted")
			if buildErr := native.Build(native.C(program), binary, native.Options{Count: true}); buildErr != nil {
				t.Fatalf("LEAK %s lowered, the counted build failed: %v", name, buildErr)
			}
			got := execute(t, binary)
			match := leakcheck.CountsLine.FindSubmatch(got.stderr)
			if match == nil {
				t.Fatalf("LEAK %s lowered, no counts line: exit %d, stderr %q", name, got.exitCode, got.stderr)
			}
			verdict := leakcheck.Unbalanced(leakcheck.Run{Stdout: got.stdout, Stderr: got.stderr, ExitCode: got.exitCode})
			if verdict == "" {
				verdict = "balanced"
			}
			t.Logf("LEAK %s allocations %s frees %s regions %s graph types %d exit %d stdout %q: %s", name, match[1], match[2], match[6], len(program.GraphTypes), got.exitCode, got.stdout, verdict)
		})
	}
}
