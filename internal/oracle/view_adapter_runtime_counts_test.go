package oracle

import "testing"

// Header growth must not change the counted allocation/ownership contract of
// existing closure fixtures. These rows are the recorded before-change values.
func viewAdapterBaselineCounts(t *testing.T, fixture, before string) {
	t.Helper()
	after := counted(t, "internal/oracle/testdata/"+fixture, false, nil, false, false)
	if after != before {
		t.Fatalf("oracle counts changed:\nbefore %s\nafter  %s", before, after)
	}
	t.Logf("before %s\nafter  %s", before, after)
}

func TestViewAdapterBaselineCountsClosures(t *testing.T) {
	t.Parallel()
	viewAdapterBaselineCounts(t, "closures.a", "| internal/oracle/testdata/closures.a | 58 | 58 | 48 | 84 | 26 | 0 |")
}

func TestViewAdapterBaselineCountsMethods(t *testing.T) {
	t.Parallel()
	viewAdapterBaselineCounts(t, "method_closures.a", "| internal/oracle/testdata/method_closures.a | 20 | 20 | 14 | 30 | 8 | 0 |")
}

func TestViewAdapterBaselineCountsReceiverRest(t *testing.T) {
	t.Parallel()
	viewAdapterBaselineCounts(t, "closure_convention_receiver_rest.a", "| internal/oracle/testdata/closure_convention_receiver_rest.a | 11 | 11 | 12 | 23 | 7 | 0 |")
}
