package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/system-inc/adamic/internal/skipcensus"
)

const skipCensusTable = "internal/skipcensus/testdata/skips.json"

type skipCensusSummary struct {
	Status                                                        string
	NotApplicable, Measurement, OptInLane, RequiredInput, Unknown []string
}

func censusDeclarations(root string) ([]skipcensus.Row, bool, error) {
	if _, err := os.Stat(filepath.Join(root, "internal/skipcensus/census.go")); os.IsNotExist(err) {
		return nil, false, nil
	} else if err != nil {
		return nil, true, err
	}
	actual, err := skipcensus.Scan(root)
	if err != nil {
		return nil, true, err
	}
	file, err := os.Open(filepath.Join(root, skipCensusTable))
	if err != nil {
		return nil, true, err
	}
	rows, err := skipcensus.Load(file)
	file.Close()
	if err != nil {
		return nil, true, err
	}
	if err := skipcensus.Validate(actual, rows); err != nil {
		return nil, true, fmt.Errorf("skip census table does not match tree: %w", err)
	}
	return rows, true, nil
}

// The concatenated shard streams are exactly the merged test.jsonl stream.
// Validate source first, then let the census classify every terminal skip.
func checkSkipCensus(root string, logs ...string) (skipCensusSummary, error) {
	summary := skipCensusSummary{Status: "not-landed", NotApplicable: []string{}, Measurement: []string{}, OptInLane: []string{}, RequiredInput: []string{}, Unknown: []string{}}
	rows, landed, err := censusDeclarations(root)
	if !landed {
		return summary, err
	}
	summary.Status = "checked"
	if err != nil {
		summary.Status = "invalid"
		return summary, err
	}
	readers := []io.Reader{}
	for _, path := range logs {
		file, err := os.Open(path)
		if err != nil {
			return summary, err
		}
		defer file.Close()
		readers = append(readers, file)
	}
	var report bytes.Buffer
	checkerErr := skipcensus.CheckLog(io.MultiReader(readers...), &report, rows)
	var problems []string
	for _, line := range strings.Split(report.String(), "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			continue
		}
		name := fields[1] + "::" + fields[2]
		switch fields[0] {
		case "not-applicable":
			summary.NotApplicable = append(summary.NotApplicable, name)
		case "measurement":
			summary.Measurement = append(summary.Measurement, name)
		case "opt-in-lane":
			summary.OptInLane = append(summary.OptInLane, name)
		case "unknown":
			summary.Unknown = append(summary.Unknown, name)
			problems = append(problems, "unclassified skip "+name)
		case "required-input":
			summary.RequiredInput = append(summary.RequiredInput, name)
			provision := ""
			if len(fields) > 3 {
				for _, row := range rows {
					directory := filepath.ToSlash(filepath.Dir(row.File))
					if row.ID == fields[3] && strings.TrimPrefix(fields[1], "github.com/system-inc/adamic/") == directory {
						provision = row.Provides
						break
					}
				}
			}
			problems = append(problems, "required-input skip "+name+"; input: "+provision)
		}
	}
	for _, names := range [][]string{summary.NotApplicable, summary.Measurement, summary.OptInLane, summary.RequiredInput, summary.Unknown} {
		sort.Strings(names)
	}
	if checkerErr != nil {
		problems = append(problems, checkerErr.Error())
	}
	if len(problems) > 0 {
		return summary, fmt.Errorf("%s", strings.Join(problems, "\n"))
	}
	return summary, nil
}
