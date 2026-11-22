package oracle

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

// #a4frknq's digest probe, copied into each commit a proof job checks out and never merged: every count row's key, its
// product shard and its digest (countDigest), from keying alone, as countMeasurementInputs keys a shard. Nothing is
// built or run. A4_DIGESTS names the file it writes.
func TestPerformanceA4frknqRowDigests(t *testing.T) {
	out := os.Getenv("A4_DIGESTS")
	if out == "" {
		t.Skip("A4_DIGESTS names no file")
	}
	cases := countCases()
	digests := make([]string, len(cases))
	t.Run("rows", func(t *testing.T) {
		for index, each := range cases {
			t.Run(strconv.Itoa(index), func(t *testing.T) {
				t.Parallel()
				runs, _ := each.prepare(t)
				digests[index] = countDigest(t, each.key, runs)
			})
		}
	})
	var lines strings.Builder
	for index, each := range cases {
		fmt.Fprintf(&lines, "%s\t%d\t%s\n", each.key, familyShard(each.key, countProductShards), digests[index])
	}
	if err := os.WriteFile(out, []byte(lines.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	shared, err := countsShared()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("A4 digests %d rows, shared %s", len(cases), shared[:16])
}
