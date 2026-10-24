package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

func programMillionMeasured(t *testing.T, enabled, sanitized bool) string {
	t.Helper()
	return buildcache.Product(t, buildcache.Inputs{Name: fmt.Sprint("program-million-measured-", enabled, "-", sanitized), Files: []string{"internal/oracle/program_region_million_measure_test.go", "internal/oracle/testdata/program_region/million.a", "internal/lower", "internal/native"}}, func(directory string) error {
		path, err := filepath.Abs("testdata/program_region/million.a")
		if err != nil {
			return err
		}
		code := native.C(programRegionLowered(t, path, enabled))
		at := strings.LastIndex(code, "\treturn adamic_host_exit_status();")
		if at < 0 {
			return fmt.Errorf("main return missing")
		}
		code = "#include <sys/resource.h>\n#include <stdio.h>\n" + code[:at] + "\tstruct rusage graph_usage; getrusage(RUSAGE_SELF, &graph_usage); printf(\"graph_peak_rss_kib %ld\\n\", graph_usage.ru_maxrss);\n" + code[at:]
		options := native.Options{Count: !sanitized, Sanitize: sanitized, ProgramRegion: enabled}
		return programRegionBuild(t, code, filepath.Join(directory, "million"), options)
	})
}

func programMillionMeasurement(t *testing.T, enabled, sanitized bool) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("Linux RSS witness")
	}
	path, err := filepath.Abs("testdata/program_region/million.a")
	if err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), path)
	binary := filepath.Join(programMillionMeasured(t, enabled, sanitized), "million")
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	rss := graphRSS.FindSubmatch(result.stdout)
	if rss == nil {
		t.Fatalf("missing peak RSS: %s", result.stdout)
	}
	actual := result
	actual.stdout = graphRSS.ReplaceAll(actual.stdout, nil)
	actual.stderr = regexp.MustCompile(`(?m)^adamic: (counts|graph region|graph counts):[^\n]*\n`).ReplaceAll(actual.stderr, nil)
	if difference := disagreement(truth, actual); difference != "" {
		t.Fatalf("Program=%t sanitized=%t: %s: %s", enabled, sanitized, difference, result.stderr)
	}
	if !sanitized {
		if report := leakcheck.Unbalanced(leakRun(result)); report != "" {
			t.Fatal(report)
		}
	}
	t.Logf("Program=%t sanitized=%t peak-RSS-KiB=%s %s", enabled, sanitized, rss[1], result.stderr)
}
func TestProduct_ProgramMillionMeasuredOff(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux RSS witness")
	}
	programMillionMeasured(t, false, false)
}
func TestProgramRegionMillionMeasurementsOff(t *testing.T) {
	t.Parallel()
	programMillionMeasurement(t, false, false)
}
func TestProduct_ProgramMillionMeasuredSanitizedOff(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux RSS witness")
	}
	programMillionMeasured(t, false, true)
}
func TestProgramRegionMillionMeasurementsSanitizedOff(t *testing.T) {
	t.Parallel()
	programMillionMeasurement(t, false, true)
}
func TestProduct_ProgramMillionMeasuredOn(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux RSS witness")
	}
	programMillionMeasured(t, true, false)
}
func TestProgramRegionMillionMeasurementsOn(t *testing.T) {
	t.Parallel()
	programMillionMeasurement(t, true, false)
}
func TestProduct_ProgramMillionMeasuredSanitizedOn(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux RSS witness")
	}
	programMillionMeasured(t, true, true)
}
func TestProgramRegionMillionMeasurementsSanitizedOn(t *testing.T) {
	t.Parallel()
	programMillionMeasurement(t, true, true)
}
