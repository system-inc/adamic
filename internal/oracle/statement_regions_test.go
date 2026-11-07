package oracle

import (
	"strconv"
	"strings"
	"testing"
)

// Updating the count golden must not hide loss of the statement-region optimization.
func TestStatementRegionsAreUsed(test *testing.T) {
	test.Parallel()
	recorded, recordedCountsError := recordedCounts()
	if recordedCountsError != nil {
		test.Fatal(recordedCountsError)
	}
	rows := make(map[string][]string)
	for key, row := range recorded {
		rows[key] = strings.Split(row, "|")
	}
	for _, fixture := range []string{
		"route_targets_virtual_fresh.a", // Virtual producers return fresh values consumed within a statement.
		"borrow_chain_override.a",       // An override changes a parent while a borrowed chain remains readable.
		"regions.a",                     // Recursive fresh trees become statement temporaries beside escaping values.
		"regions_throw.a",               // Throwing through a fresh-tree consumer cleans up its statement region.
		"devirtualize_factory.a",        // A factory's fresh return flows through devirtualized methods.
		"call_targets_region.a",         // Resolved call targets feed fresh values to region-safe consumers.
	} {
		test.Run(fixture, func(test *testing.T) {
			path := "internal/oracle/testdata/" + fixture
			fields, ok := rows[path]
			if !ok {
				test.Fatalf("%s missing from %s", path, countFile("runtime", path))
			}
			read := func(column int) int64 {
				test.Helper()
				value, parseError := strconv.ParseInt(strings.TrimSpace(fields[column]), 10, 64)
				if parseError != nil || value < 0 {
					test.Fatalf("%s: invalid count in column %d: %q", path, column, fields[column])
				}
				return value
			}
			allocations, frees, regions := read(2), read(3), read(7)
			if regions < 1 {
				test.Errorf("%s: In regions = %d, want at least 1", path, regions)
			}
			if allocations != frees+regions {
				test.Errorf(
					"%s: Allocations = %d, want Frees (%d) + In regions (%d)",
					path,
					allocations,
					frees,
					regions,
				)
			}
		})
	}
}
