package oracle

import (
	"fmt"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

var programRegionCountFixtures = []string{"cycles_graph_parent.a", "cycles_graph_relations.a", "cycles_graph_symbols.a", "cycles_weak_parent.a", "cycles_weak_relations.a", "cycles_weak_symbols.a", "million.a", "program_region_ownership.a", "program_region_map_storage.a", "optional.a", "constructor.a"}

func init() { additionalFixtureCounts = append(additionalFixtureCounts, programRegionCountRows) }
func programRegionCountRows(t *testing.T) []string {
	t.Helper()
	rows := []string{}
	for _, name := range programRegionCountFixtures {
		relative := "internal/oracle/testdata/program_region/" + name
		path, err := filepath.Abs(filepath.Join(repository, relative))
		if err != nil {
			t.Fatal(err)
		}
		p := programRegionLowered(t, path, true)
		binary := filepath.Join(t.TempDir(), "counts")
		if err := native.Build(native.C(p), binary, native.Options{ProgramRegion: true, Count: true}); err != nil {
			t.Fatal(err)
		}
		command, args := pinnedStack(binary)
		result := execute(t, command, args...)
		match := countsLine.FindSubmatch(result.stderr)
		if result.exitCode != 0 || match == nil {
			t.Fatalf("%s: counts %d %s", name, result.exitCode, result.stderr)
		}
		columns := []string{}
		for _, field := range match[1:] {
			columns = append(columns, string(field))
		}
		alloc, _ := strconv.Atoi(columns[0])
		frees, _ := strconv.Atoi(columns[1])
		members, _ := strconv.Atoi(columns[5])
		if alloc != frees+members {
			t.Fatalf("%s: unaccounted allocation: %s", name, result.stderr)
		}
		rows = append(rows, fmt.Sprintf("| Program region: %s | %s |", relative, strings.Join(columns, " | ")))
	}
	return rows
}
