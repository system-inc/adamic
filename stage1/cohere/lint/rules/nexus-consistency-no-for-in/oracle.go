//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleNexusConsistencyNoForIn() rule.Rule { return rules.ConsistencyNoForIn }

// Upstream explicitly has no options type or decoder and ignores options.
// Decode the payload supplied by all-rule rows, preserving that upstream behavior
// while never returning nil for a row carrying a non-null options object.
func oracleNexusConsistencyNoForInOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var options json.RawMessage
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
