package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

// The general .a ruling also moves existing intentional stale reads to exact
// compile-time refusals. This is a refusal census, not the admission witness list.
var step21WritingRefusals = []struct{ path, position, variable, call string }{
	{"047cb0d_n_arrayindex.a", "7:17", "list", "drop()"},
	{"e4eec87_f2b_union_narrow_number.a", "11:32", "shared", "toText()"},
	{"e4eec87_f2_union_narrow_call.a", "11:34", "shared", "toNumber()"},
	{"regexp_null_narrowed.a", "3:47", "result", "lose()"},
	{"narrowed_writes.a", "15:2", "chain", "drop()"},
	{"narrowed_union_valid.a", "11:24", "shared", "keep()"},
	{"narrowed_reads.a", "15:17", "chain", "drop()"},
	{"narrowed_numbers.a", "11:17", "count", "drop()"},
	{"narrowed_methods.a", "19:17", "box", "drop()"},
	{"review/agree/fxspptb_e4eec87_f2b_union_narrow_number.a", "12:32", "shared", "toText()"},
	{"review/agree/fxspptb_e4eec87_f2_union_narrow_call.a", "12:34", "shared", "toNumber()"},
	{"review/agree/fxspptb_9984394_defined_in_try.a", "14:19", "current", "clear()"},
	{"review/agree/fxspptb_047cb0d_n_try.a", "12:18", "chain", "drop()"},
	{"review/agree/fxspptb_047cb0d_n_arrayindex.a", "7:17", "list", "drop()"},
}

func init() {
	kept := fixtures[:0]
	for _, fixture := range fixtures {
		refused := false
		for _, row := range step21WritingRefusals {
			if fixture.path == "internal/oracle/testdata/"+row.path {
				refused = true
			}
		}
		if !refused {
			kept = append(kept, fixture)
		}
	}
	fixtures = kept
}

func TestStep21ExistingWritingCallRefusals(t *testing.T) {
	t.Parallel()
	for _, row := range step21WritingRefusals {
		t.Run(row.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", row.path))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			want := path + ":" + row.position + ": Adamic 0.1 refuses a narrowed read of " + row.variable + " after " + row.call + " can write it; narrow again after the call"
			if err == nil || err.Error() != want {
				t.Fatalf("writing-call refusal: got %v, want %s", err, want)
			}
		})
	}
}

// Existing IR stop mutation proofs remain useful in TypeScript compatibility
// mode. Keep repository .a sources intact and refused; use an ephemeral .ts copy.
func step21TypeScriptCopy(t *testing.T, path string) string {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	copy := filepath.Join(t.TempDir(), "writing-call.ts")
	if err := os.WriteFile(copy, source, 0o644); err != nil {
		t.Fatal(err)
	}
	return copy
}
