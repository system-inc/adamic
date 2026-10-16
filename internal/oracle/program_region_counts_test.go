package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"testing"
)

var programRegionCountFixtures = []string{"cycles_graph_parent.a", "cycles_graph_relations.a", "cycles_graph_symbols.a", "cycles_weak_parent.a", "cycles_weak_relations.a", "cycles_weak_symbols.a", "million.a", "program_region_ownership.a", "program_region_map_storage.a", "constructor.a", "tsc_node_membership.a"}

func init() {
	for _, name := range programRegionCountFixtures {
		programRegionCountCase("internal/oracle/testdata/program_region/" + name)
	}
	programRegionCountCase("stage3/fixtures/host/14_getCurrentDirectory.a")
}
func programRegionCountCase(relative string) {
	key := "Program region: " + relative
	additionalFixtureCounts = append(additionalFixtureCounts, countCase{key, func(t *testing.T) (string, func() string) {
		absolute, err := filepath.Abs(filepath.Join(repository, relative))
		if err != nil {
			t.Fatal(err)
		}
		code := native.C(programRegionLowered(t, absolute, true))
		return countPortable(t, code, absolute), func() string {
			binary := filepath.Join(t.TempDir(), "program-counted")
			if err := programRegionBuild(t, code, binary, native.Options{ProgramRegion: true, Count: true}); err != nil {
				t.Fatal(err)
			}
			name, args := pinnedStack(binary)
			result := execute(t, name, args...)
			if report := leakcheck.Unbalanced(leakRun(result)); report != "" {
				t.Fatal(report)
			}
			match := countsLine.FindSubmatch(result.stderr)
			if result.exitCode != 0 || match == nil {
				t.Fatalf("Program counts: %d %s", result.exitCode, result.stderr)
			}
			return fmt.Sprintf("| %s | %s | %s | %s | %s | %s | %s |", key, match[1], match[2], match[3], match[4], match[5], match[6])
		}
	}})
}

func programRegionRecordedCount(t *testing.T, relative string) {
	t.Helper()
	key := "Program region: " + relative
	for _, fixture := range countCases() {
		if fixture.key != key {
			continue
		}
		_, measure := fixture.prepare(t)
		row := measure()
		if *updateCounts {
			measuredCounts.Store(key, row)
		} else {
			recorded, err := recordedCounts()
			if err != nil {
				t.Fatal(err)
			}
			if row != recorded[key] {
				t.Fatalf("Program count changed: %s; recorded %s", row, recorded[key])
			}
		}
		t.Log(row)
		return
	}
	t.Fatalf("count fixture missing: %s", key)
}
func TestProgramRegionRecordedCount0(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "internal/oracle/testdata/program_region/cycles_graph_parent.a")
}
func TestProgramRegionRecordedCount1(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "internal/oracle/testdata/program_region/cycles_graph_relations.a")
}
func TestProgramRegionRecordedCount2(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "internal/oracle/testdata/program_region/cycles_graph_symbols.a")
}
func TestProgramRegionRecordedCount3(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "internal/oracle/testdata/program_region/cycles_weak_parent.a")
}
func TestProgramRegionRecordedCount4(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "internal/oracle/testdata/program_region/cycles_weak_relations.a")
}
func TestProgramRegionRecordedCount5(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "internal/oracle/testdata/program_region/cycles_weak_symbols.a")
}
func TestProgramRegionRecordedCount6(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "internal/oracle/testdata/program_region/million.a")
}
func TestProgramRegionRecordedCount7(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "internal/oracle/testdata/program_region/program_region_ownership.a")
}
func TestProgramRegionRecordedCount8(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "internal/oracle/testdata/program_region/program_region_map_storage.a")
}
func TestProgramRegionRecordedCount9(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "internal/oracle/testdata/program_region/constructor.a")
}
func TestProgramRegionRecordedCount10(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "internal/oracle/testdata/program_region/tsc_node_membership.a")
}
func TestProgramRegionRecordedCountHost(t *testing.T) {
	t.Parallel()
	programRegionRecordedCount(t, "stage3/fixtures/host/14_getCurrentDirectory.a")
}
