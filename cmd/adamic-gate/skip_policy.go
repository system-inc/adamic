package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type skipCensusSummary struct {
	Status                     string
	NotApplicable, Measurement []string
}

// Integration hook for internal/skipcensus once it lands in the integration area.
// Load its declarations, Scan and Validate current source, then call CheckLog on
// the shard streams. Append checker errors before Green is computed and collect
// its named not-applicable/measurement rows into the corresponding summary lists.
// Do not introduce a second classification table or silently bypass a landed checker.
func checkSkipCensus(root string) (skipCensusSummary, error) {
	summary := skipCensusSummary{Status: "not-landed", NotApplicable: []string{}, Measurement: []string{}}
	_, err := os.Stat(filepath.Join(root, "internal", "skipcensus", "census.go"))
	if os.IsNotExist(err) {
		return summary, nil
	}
	summary.Status = "not-wired"
	if err != nil {
		return summary, fmt.Errorf("skip census source: %w", err)
	}
	return summary, fmt.Errorf("internal/skipcensus has landed but its gate hook is not wired: use Scan, Load, Validate and CheckLog before accepting merge")
}
