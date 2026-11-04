package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// recordViewPins are the record view programs in testdata/records_refused whose first line names the stop each must make,
// "// a-check: refused <text>". a-check doesn't judge a program inside a package's testdata, so this test does: each
// lowers to an error holding its header's text, never to a program (#6c4swvt).
var recordViewPins = []string{
	"records_pick_readonly_field", "records_pick_hidden_contract", "records_pick_values", "records_pick_of_pick", "records_readonly_view",
	"records_union_member_out", "records_union_member_out_storage", "records_union_member_write", "records_union_member_optional_write",
	"records_literal_fixed_branch",
}

func TestRecordViewPinsStopWithTheirHeaders(t *testing.T) {
	t.Parallel()
	for _, name := range recordViewPins {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/records_refused", name+".a"))
			if err != nil {
				t.Fatal(err)
			}
			source, err := os.ReadFile(absolute)
			if err != nil {
				t.Fatal(err)
			}
			header, _, _ := strings.Cut(string(source), "\n")
			want, pinned := strings.CutPrefix(header, "// a-check: refused ")
			if !pinned || want == "" {
				t.Fatalf("first line %q names no stop", header)
			}
			_, err = lowered(t, absolute)
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("got %v, want an error holding %q", err, want)
			}
		})
	}
}
