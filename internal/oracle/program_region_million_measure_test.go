package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func programMillionMeasured(t *testing.T, enabled bool) string {
	t.Helper()
	return buildcache.Product(t, buildcache.Inputs{Name: fmt.Sprint("program-million-measured-", enabled), Files: []string{"internal/oracle/program_region_million_measure_test.go", "internal/oracle/testdata/program_region/million.a", "internal/lower", "internal/native"}}, func(directory string) error {
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
		for _, sanitized := range []bool{false, true} {
			options := native.Options{Count: !sanitized, Sanitize: sanitized, ProgramRegion: enabled}
			output := filepath.Join(directory, fmt.Sprint("million-", sanitized))
			if err := programRegionBuild(t, code, output, options); err != nil {
				return err
			}
		}
		return nil
	})
}

func TestProgramRegionMillionMeasurements(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("testdata/program_region/million.a")
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	for _, enabled := range []bool{false, true} {
		directory := programMillionMeasured(t, enabled)
		for _, sanitized := range []bool{false, true} {
			binary := filepath.Join(directory, fmt.Sprint("million-", sanitized))
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
				t.Logf("Program=%t peak-RSS-KiB=%s %s", enabled, rss[1], result.stderr)
			}
		}
	}
}
