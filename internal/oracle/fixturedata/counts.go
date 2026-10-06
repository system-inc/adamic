package fixturedata

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// order preserves the original table's presentation, independently per fixture.
// New fixtures omit it and sort by source path after the migrated rows. It is
// never a registration index and workers never need to allocate a shared number.
type counts struct {
	Order int    `json:"order,omitempty"`
	Row   string `json:"row"`
}

var rowPattern = regexp.MustCompile(`^\| ([^|]+) \| [0-9]+ \| [0-9]+ \| [0-9]+ \| [0-9]+ \| [0-9]+ \| [0-9]+ \|$`)

func readCounts(path, source string) (counts, error) {
	var recorded counts
	_, err := decode(path, &recorded)
	if err != nil {
		return recorded, err
	}
	match := rowPattern.FindStringSubmatch(recorded.Row)
	if match == nil || match[1] != source || recorded.Order < 0 {
		return recorded, fmt.Errorf("%s: invalid counts row for %s", path, source)
	}
	return recorded, nil
}

// CheckCounts compares every measured value even when execution was cached.
// Updating preserves presentation metadata and writes only this changed file.
func CheckCounts(path, source, row string, update bool) error {
	recorded, err := readCounts(path, source)
	if err != nil && !(update && errors.Is(err, os.ErrNotExist)) {
		return err
	}
	if err == nil && recorded.Row == row {
		return nil
	}
	if !update {
		return fmt.Errorf("%s: counts changed; run TestCountsAreRecorded with -update-counts:\nrecorded: %s\nmeasured: %s", path, recorded.Row, row)
	}
	match := rowPattern.FindStringSubmatch(row)
	if match == nil || match[1] != source {
		return fmt.Errorf("%s: invalid measured row", path)
	}
	recorded.Row = row
	contents, err := json.MarshalIndent(recorded, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(contents, '\n'), 0o644)
}

// RenderCounts reproduces the human table without compiling or running fixtures.
// Missing or malformed records fail loudly, rather than silently losing rows.
func RenderCounts(fixtures []Fixture) (string, error) {
	var records []counts
	for _, fixture := range fixtures {
		if !fixture.Lowers || fixture.Uncounted {
			continue
		}
		recorded, err := readCounts(fixture.CountsPath, fixture.Path)
		if err != nil {
			return "", err
		}
		records = append(records, recorded)
	}
	sort.Slice(records, func(i, j int) bool {
		left, right := records[i], records[j]
		if left.Order == right.Order {
			return left.Row < right.Row
		}
		if left.Order == 0 {
			return false
		}
		if right.Order == 0 {
			return true
		}
		return left.Order < right.Order
	})
	var table strings.Builder
	table.WriteString(CountsHeader)
	for _, record := range records {
		table.WriteString(record.Row)
		table.WriteByte('\n')
	}
	return table.String(), nil
}
