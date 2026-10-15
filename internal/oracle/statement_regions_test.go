package oracle

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// Updating the count golden must not hide loss of the statement-region optimization.
func TestStatementRegionsAreUsed(t *testing.T) {
	t.Parallel()
	recorded, err := os.ReadFile(countsPath)
	if err != nil {
		t.Fatal(err)
	}
	rows := make(map[string][]string)
	for _, line := range strings.Split(string(recorded), "\n") {
		fields := strings.Split(line, "|")
		if len(fields) == 9 {
			rows[strings.TrimSpace(fields[1])] = fields
		}
	}
	for _, fixture := range []string{
		"route_targets_virtual_fresh.a", // Virtual producers return fresh values consumed within a statement.
		"borrow_chain_override.a",       // An override changes a parent while a borrowed chain remains readable.
		"regions.a",                     // Recursive fresh trees become statement temporaries beside escaping values.
		"regions_throw.a",               // Throwing through a fresh-tree consumer cleans up its statement region.
		"devirtualize_factory.a",        // A factory's fresh return flows through devirtualized methods.
		"call_targets_region.a",         // Resolved call targets feed fresh values to region-safe consumers.
	} {
		t.Run(fixture, func(t *testing.T) {
			path := "internal/oracle/testdata/" + fixture
			fields, ok := rows[path]
			if !ok {
				t.Fatalf("%s missing from %s", path, countsPath)
			}
			read := func(column int) int64 {
				t.Helper()
				value, err := strconv.ParseInt(strings.TrimSpace(fields[column]), 10, 64)
				if err != nil || value < 0 {
					t.Fatalf("%s: invalid count in column %d: %q", path, column, fields[column])
				}
				return value
			}
			allocations, frees, regions := read(2), read(3), read(7)
			if regions < 1 {
				t.Errorf("%s: In regions = %d, want at least 1", path, regions)
			}
			if allocations != frees+regions {
				t.Errorf("%s: Allocations = %d, want Frees (%d) + In regions (%d)", path, allocations, frees, regions)
			}
		})
	}
}
