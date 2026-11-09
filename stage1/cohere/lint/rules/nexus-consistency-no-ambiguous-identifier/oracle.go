//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleNexusConsistencyNoAmbiguousIdentifier(options any) rule.Bound {
	return bindOracle(nexus.ConsistencyNoAmbiguousIdentifier, options)
}

func oracleNexusConsistencyNoAmbiguousIdentifierOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" {
		return nil
	}
	// Upstream defines no options type and ignores options. Preserve the decoded
	// payload, including unknown fields, rather than silently discarding it.
	var options any
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
