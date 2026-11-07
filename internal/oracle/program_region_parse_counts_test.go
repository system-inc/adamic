package oracle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

func programParseCountRelative(enabled bool) string {
	return fmt.Sprintf("internal/oracle/testdata/program_region/parse.a (Program=%t)", enabled)
}
func init() {
	for _, enabled := range []bool{false, true} {
		key := "Program region: " + programParseCountRelative(enabled)
		countFixtureSources[key] = "internal/oracle/testdata/program_region/parse.a"
		additionalFixtureCounts = append(
			additionalFixtureCounts,
			countCase{key, func(test *testing.T) (string, func() string) {
				code, err := os.ReadFile(filepath.Join(programParseLower(test, enabled), "code.c"))
				if err != nil {
					test.Fatal(err)
				}
				manifest, err := os.ReadFile("testdata/program_region/parse_manifest.json")
				if err != nil {
					test.Fatal(err)
				}
				runs := string(code) +
					strings.Join(native.Flags(native.Options{Count: true, ProgramRegion: enabled}), " ") +
					string(manifest)
				return countPortable(test, runs, ""), func() string {
					binary := filepath.Join(programParseBuild(test, enabled), "parse")
					name, args := pinnedStack(binary, programParseArguments(test)...)
					result := execute(test, name, args...)
					match := countsLine.FindSubmatch(result.stderr)
					if result.exitCode != 0 || match == nil {
						test.Fatalf("parse counts: %d %s", result.exitCode, result.stderr)
					}
					if report := leakcheck.Unbalanced(leakRun(result)); report != "" {
						test.Fatal(report)
					}
					return fmt.Sprintf(
						"| %s | %s | %s | %s | %s | %s | %s |",
						key,
						match[1],
						match[2],
						match[3],
						match[4],
						match[5],
						match[6],
					)
				}
			}},
		)
	}
}
func TestProgramRegionRecordedCountParseOn(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, programParseCountRelative(true))
}
func TestProgramRegionRecordedCountParseOff(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, programParseCountRelative(false))
}
