package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// Keep this unit separate from the shared fixture list.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/4e1649c_generic_receiver_number.a",
		"internal/oracle/testdata/4e1649c_generic_receiver.a",
		"internal/oracle/testdata/4e1649c_generic_receiver_charat.a",
		"internal/oracle/testdata/4e1649c_generic_template.a",
		"internal/oracle/testdata/4e1649c_generic_pad_fill_call.a",
		"internal/oracle/testdata/4e1649c_generic_pad_fill.a",
		"internal/oracle/testdata/4e1649c_generic_conversion.a",
		"internal/oracle/testdata/4e1649c_generic_null.a",
		"internal/oracle/testdata/4e1649c_generic_receiver_order.a",
		"internal/oracle/testdata/4e1649c_raw_crlf.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, true, false})
	}
}

// A same-named writer must still be judged as a mutable view, rather than the raw intrinsic.
func TestRawNamedWriterStaysRefused(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/string_refused/4e1649c_raw_named_writer.a"))
	if err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if observed.exitCode != 0 || string(observed.stdout) != "2\nrex 4\nstray NaN\n" {
		t.Fatalf("Node must show the wider writer invalidating Aged: exit %d, stdout %q, stderr %q", observed.exitCode, observed.stdout, observed.stderr)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(err.Error(), "adamic/invariant-mutable") {
		t.Fatalf("want mutable widening refusal, got %v", err)
	}
}
