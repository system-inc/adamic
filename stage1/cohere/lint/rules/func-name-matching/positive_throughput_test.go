package validation

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Not parallel: report uncontended process-inclusive rates for held positive witnesses.
func TestPositiveThroughput(t *testing.T) {
	if os.Getenv("ADAMIC_LINT_BENCH") != "1" {
		t.Skip("ADAMIC_LINT_BENCH=1")
	}
	port := copyPort(t, "", "", "")
	oracle := goOracle(t, port)
	entry, _, binary := builds(t, port, false)
	all := witnesses(t)
	for _, slug := range slugs {
		name := ruleName(slug)
		var rows []string
		for _, row := range all {
			if strings.HasSuffix(row, "\t"+name) {
				rows = append(rows, row)
			}
		}
		if len(rows) == 0 {
			t.Fatal("missing positive witness", slug)
		}
		sample := make([]string, 1000)
		for i := range sample {
			sample[i] = rows[i%len(rows)]
		}
		path := manifest(t, sample)
		best := map[string]time.Duration{}
		var want []byte
		for round := 0; round < 3; round++ {
			for _, side := range []struct {
				name string
				r    observation
			}{{"Go", run(t, "", oracle, "--manifest", path, "--count")}, {"native", run(t, "", binary, "--manifest", path, "--count")}, {"Node", run(t, "", "node", "--disable-warning=ExperimentalWarning", filepath.Join(repo(t), "oracle/node.mjs"), entry, "--manifest", path, "--count")}} {
				got := clean(t, side.r)
				if want == nil {
					want = got
				}
				if !bytes.Equal(got, want) {
					t.Fatal("throughput count differs", side.name)
				}
				if best[side.name] == 0 || side.r.duration < best[side.name] {
					best[side.name] = side.r.duration
				}
			}
		}
		var count int
		if _, err := fmt.Sscan(string(want), &count); err != nil {
			t.Fatal(err)
		}
		if count == 0 {
			t.Fatal("positive benchmark has no findings", slug)
		}
		for _, side := range []string{"native", "Node", "Go"} {
			t.Logf("%s %s best-of-3 %d findings / %.6fs = %.2f findings/s", name, side, count, best[side].Seconds(), float64(count)/best[side].Seconds())
		}
	}
}
