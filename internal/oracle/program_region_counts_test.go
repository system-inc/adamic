package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"testing"
)

var programRegionCountFixtures = []string{
	"cycles_graph_parent.a",
	"cycles_graph_relations.a",
	"cycles_graph_symbols.a",
	"cycles_weak_parent.a",
	"cycles_weak_relations.a",
	"cycles_weak_symbols.a",
	"million.a",
	"program_region_ownership.a",
	"program_region_map_storage.a",
	"constructor.a",
	"tsc_node_membership.a",
}

func init() {
	for _, name := range programRegionCountFixtures {
		programRegionCountCase("internal/oracle/testdata/program_region/" + name)
	}
	programRegionCountCase("stage3/fixtures/host/14_getCurrentDirectory.a")
}
func programRegionCountCase(relative string) {
	key := "Program region: " + relative
	countFixtureSources[key] = relative
	additionalFixtureCounts = append(
		additionalFixtureCounts,
		countCase{key, func(test *testing.T) (string, func() string) {
			absolute, err := filepath.Abs(filepath.Join(repository, relative))
			if err != nil {
				test.Fatal(err)
			}
			code := native.C(programRegionLowered(test, absolute, true))
			return countPortable(test, code, absolute), func() string {
				binary := filepath.Join(test.TempDir(), "program-counted")
				if err := programRegionBuild(
					test,
					code,
					binary,
					native.Options{ProgramRegion: true, Count: true},
				); err != nil {
					test.Fatal(err)
				}
				name, args := pinnedStack(binary)
				result := execute(test, name, args...)
				if report := leakcheck.Unbalanced(leakRun(result)); report != "" {
					test.Fatal(report)
				}
				match := countsLine.FindSubmatch(result.stderr)
				if result.exitCode != 0 || match == nil {
					test.Fatalf("Program counts: %d %s", result.exitCode, result.stderr)
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

func programRegionRecordedCount(test *testing.T, relative string) {
	test.Helper()
	key := "Program region: " + relative
	for _, fixture := range countCases() {
		if fixture.key != key {
			continue
		}
		_, measure := fixture.prepare(test)
		row := measure()
		if *updateCounts {
			measuredCounts.Store(key, row)
		} else {
			recorded, err := recordedCounts()
			if err != nil {
				test.Fatal(err)
			}
			if row != recorded[key] {
				test.Fatalf("Program count changed: %s; recorded %s", row, recorded[key])
			}
		}
		test.Log(row)
		return
	}
	test.Fatalf("count fixture missing: %s", key)
}
func TestProgramRegionRecordedCount0(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "internal/oracle/testdata/program_region/cycles_graph_parent.a")
}
func TestProgramRegionRecordedCount1(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "internal/oracle/testdata/program_region/cycles_graph_relations.a")
}
func TestProgramRegionRecordedCount2(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "internal/oracle/testdata/program_region/cycles_graph_symbols.a")
}
func TestProgramRegionRecordedCount3(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "internal/oracle/testdata/program_region/cycles_weak_parent.a")
}
func TestProgramRegionRecordedCount4(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "internal/oracle/testdata/program_region/cycles_weak_relations.a")
}
func TestProgramRegionRecordedCount5(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "internal/oracle/testdata/program_region/cycles_weak_symbols.a")
}
func TestProgramRegionRecordedCount6(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "internal/oracle/testdata/program_region/million.a")
}
func TestProgramRegionRecordedCount7(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "internal/oracle/testdata/program_region/program_region_ownership.a")
}
func TestProgramRegionRecordedCount8(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "internal/oracle/testdata/program_region/program_region_map_storage.a")
}
func TestProgramRegionRecordedCount9(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "internal/oracle/testdata/program_region/constructor.a")
}
func TestProgramRegionRecordedCount10(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "internal/oracle/testdata/program_region/tsc_node_membership.a")
}
func TestProgramRegionRecordedCountHost(test *testing.T) {
	test.Parallel()
	programRegionRecordedCount(test, "stage3/fixtures/host/14_getCurrentDirectory.a")
}
