package oracle

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

func programHostMeasuredBuild(t *testing.T, enabled bool) string {
	t.Helper()
	return buildcache.Product(t, buildcache.Inputs{Name: fmt.Sprint("program-host-counted-", enabled), Files: []string{"internal/oracle/program_region_measure_test.go", "stage3/fixtures/host/14_getCurrentDirectory.a", "internal/lower", "internal/native"}}, func(directory string) error {
		absolute, err := filepath.Abs("../../stage3/fixtures/host/14_getCurrentDirectory.a")
		if err != nil {
			return err
		}
		code := native.C(programRegionLowered(t, absolute, enabled))
		at := strings.LastIndex(code, "\treturn 0;")
		if at < 0 {
			return fmt.Errorf("main return missing")
		}
		code = "#include <sys/resource.h>\n#include <stdio.h>\n" + code[:at] + "\tstruct rusage graph_usage; getrusage(RUSAGE_SELF, &graph_usage); printf(\"graph_peak_rss_kib %ld\\n\", graph_usage.ru_maxrss);\n" + code[at:]
		if enabled {
			return programRegionBuild(t, code, filepath.Join(directory, "host"), native.Options{ProgramRegion: true, Count: true})
		}
		return native.Build(code, filepath.Join(directory, "host"), native.Options{Count: true})
	})
}
func TestProduct_ProgramHostOn(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux RSS witness")
	}
	programHostMeasuredBuild(t, true)
}
func TestProduct_ProgramHostOff(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux RSS witness")
	}
	programHostMeasuredBuild(t, false)
}
func programHostMeasurement(t *testing.T, sample int) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("Linux getrusage measurement")
	}
	absolute, err := filepath.Abs("../../stage3/fixtures/host/14_getCurrentDirectory.a")
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, absolute)
	for _, enabled := range []bool{false, true} {
		binary := filepath.Join(programHostMeasuredBuild(t, enabled), "host")
		name, args := pinnedStack("/bin/sh", "-c", `"$@" & child=$!; wait "$child"`, "rss", binary)
		started := time.Now()
		result := execute(t, name, args...)
		seconds := time.Since(started).Seconds()
		match := countsLine.FindSubmatch(result.stderr)
		rss := graphRSS.FindSubmatch(result.stdout)
		if result.exitCode != truth.exitCode || match == nil || rss == nil || string(graphRSS.ReplaceAll(result.stdout, nil)) != string(truth.stdout) {
			t.Fatalf("host measurement disagrees with Node: %d %s %s", result.exitCode, result.stdout, result.stderr)
		}
		if report := leakcheck.Unbalanced(leakRun(result)); report != "" {
			t.Fatal(report)
		}
		t.Logf("sample=%d Program=%t retains=%s releases=%s peak-live=%s peak-RSS-KiB=%s seconds=%.6f", sample, enabled, match[3], match[4], match[5], rss[1], seconds)
	}
}
func TestProgramRegionHostMeasurement0(t *testing.T) { t.Parallel(); programHostMeasurement(t, 0) }
func TestProgramRegionHostMeasurement1(t *testing.T) { t.Parallel(); programHostMeasurement(t, 1) }
func TestProgramRegionHostMeasurement2(t *testing.T) { t.Parallel(); programHostMeasurement(t, 2) }
func TestProgramRegionHostMeasurement3(t *testing.T) { t.Parallel(); programHostMeasurement(t, 3) }
func TestProgramRegionHostMeasurement4(t *testing.T) { t.Parallel(); programHostMeasurement(t, 4) }
