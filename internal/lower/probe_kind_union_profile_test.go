package lower

import (
	"context"
	"fmt"
	"os"
	"runtime/pprof"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
)

// A probe for #715zjph, never for landing: lowering time over casts and writes, and a CPU profile of the costly shape.

func probeProfileSource(casts, writes int) string {
	var source strings.Builder
	source.WriteString(kindUnionWriteModel)
	for index := range casts {
		fmt.Fprintf(&source, "function read%d(node: Node): number { return (node as PropertyAccessExpression | ElementAccessExpression).flags + %d; }\n", index, index)
	}
	for index := range writes {
		fmt.Fprintf(&source, "built.flags = %d;\n", index)
	}
	source.WriteString("let total = 0;\n")
	for index := range casts {
		fmt.Fprintf(&source, "total += read%d(node);\n", index)
	}
	source.WriteString("console.log(String(total));\n")
	return source.String()
}

func probeLowerTimes(t *testing.T, path string, runs int) []time.Duration {
	var times []time.Duration
	for range runs {
		loaded, err := load.Load([]string{path})
		if err != nil {
			t.Fatalf("Load: %v", err)
		}
		start := time.Now()
		if _, err := Lower(context.Background(), loaded); err != nil {
			t.Fatalf("Lower: %v", err)
		}
		times = append(times, time.Since(start))
	}
	slices.Sort(times)
	return times
}

func TestProbeKindUnionProfile(t *testing.T) {
	for _, shape := range []struct{ casts, writes int }{{200, 2000}, {200, 0}, {0, 2000}, {100, 2000}, {200, 1000}, {400, 2000}} {
		path := writeAgreementSource(t, "main.ts", probeProfileSource(shape.casts, shape.writes))
		times := probeLowerTimes(t, path, 3)
		t.Logf("TIMING casts=%d writes=%d: min %s, median %s", shape.casts, shape.writes, times[0].Round(time.Millisecond), times[1].Round(time.Millisecond))
	}
	out := os.Getenv("PROBE_PROFILE")
	if out == "" {
		return
	}
	path := writeAgreementSource(t, "main.ts", probeProfileSource(200, 2000))
	file, err := os.Create(out)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	if err := pprof.StartCPUProfile(file); err != nil {
		t.Fatal(err)
	}
	probeLowerTimes(t, path, 3)
	pprof.StopCPUProfile()
}
