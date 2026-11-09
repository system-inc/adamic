//go:build lintoracle

package main

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleConsistencyNoMultilineArrowFunction(options any) rule.Bound {
	return bindOracle(rules.ConsistencyNoMultilineArrowFunction, options)
}

// Upstream accepts any and reads no options. Decode the actual JSON value rather than
// returning nil for an option-bearing row; there is no upstream concrete options struct.
func oracleConsistencyNoMultilineArrowFunctionOptions(fields []string) any {
	var decoded any
	if fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
	}
	return decoded
}
