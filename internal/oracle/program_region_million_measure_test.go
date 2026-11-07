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

func programMillionMeasured(test *testing.T, enabled, sanitized bool) string {
	test.Helper()
	return buildcache.Product(test, buildcache.Inputs{
		Name: fmt.Sprint("program-million-measured-", enabled, "-", sanitized),
		Files: []string{
			"internal/oracle/program_region_million_measure_test.go",
			"internal/oracle/testdata/program_region/million.a",
			"internal/lower",
			"internal/native",
		},
	}, func(directory string) error {
		path, absolutePathError := filepath.Abs("testdata/program_region/million.a")
		if absolutePathError != nil {
			return absolutePathError
		}
		code := native.C(programRegionLowered(test, path, enabled))
		at := strings.LastIndex(code, "\treturn adamic_host_exit_status();")
		if at < 0 {
			return fmt.Errorf("main return missing")
		}
		code = "#include <sys/resource.h>\n#include <stdio.h>\n" +
			code[:at] +
			"\tstruct rusage graph_usage; getrusage(RUSAGE_SELF, &graph_usage); printf(\"graph_peak_rss_kib %ld\\n\", graph_usage.ru_maxrss);\n" +
			code[at:]
		options := native.Options{Count: !sanitized, Sanitize: sanitized, ProgramRegion: enabled}
		return programRegionBuild(test, code, filepath.Join(directory, "million"), options)
	})
}

func programMillionMeasurement(test *testing.T, enabled, sanitized bool) {
	test.Helper()
	if runtime.GOOS != "linux" {
		test.Skip("Linux RSS witness")
	}
	path, absolutePathError := filepath.Abs("testdata/program_region/million.a")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	truth := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle/node.mjs"),
		path,
	)
	binary := filepath.Join(programMillionMeasured(test, enabled, sanitized), "million")
	result := executeWith(test, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
	rss := graphRSS.FindSubmatch(result.stdout)
	if rss == nil {
		test.Fatalf("missing peak RSS: %s", result.stdout)
	}
	actual := result
	actual.stdout = graphRSS.ReplaceAll(actual.stdout, nil)
	actual.stderr = regexp.MustCompile(`(?m)^adamic: (counts|graph region|graph counts):[^\n]*\n`).ReplaceAll(
		actual.stderr,
		nil,
	)
	if difference := disagreement(truth, actual); difference != "" {
		test.Fatalf("Program=%t sanitized=%t: %s: %s", enabled, sanitized, difference, result.stderr)
	}
	if !sanitized {
		if report := leakcheck.Unbalanced(leakRun(result)); report != "" {
			test.Fatal(report)
		}
	}
	test.Logf("Program=%t sanitized=%t peak-RSS-KiB=%s %s", enabled, sanitized, rss[1], result.stderr)
}
func TestProduct_ProgramMillionMeasuredOff(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux RSS witness")
	}
	programMillionMeasured(test, false, false)
}
func TestProgramRegionMillionMeasurementsOff(test *testing.T) {
	test.Parallel()
	programMillionMeasurement(test, false, false)
}
func TestProduct_ProgramMillionMeasuredSanitizedOff(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux RSS witness")
	}
	programMillionMeasured(test, false, true)
}
func TestProgramRegionMillionMeasurementsSanitizedOff(test *testing.T) {
	test.Parallel()
	programMillionMeasurement(test, false, true)
}
func TestProduct_ProgramMillionMeasuredOn(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux RSS witness")
	}
	programMillionMeasured(test, true, false)
}
func TestProgramRegionMillionMeasurementsOn(test *testing.T) {
	test.Parallel()
	programMillionMeasurement(test, true, false)
}
func TestProduct_ProgramMillionMeasuredSanitizedOn(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux RSS witness")
	}
	programMillionMeasured(test, true, true)
}
func TestProgramRegionMillionMeasurementsSanitizedOn(test *testing.T) {
	test.Parallel()
	programMillionMeasurement(test, true, true)
}
