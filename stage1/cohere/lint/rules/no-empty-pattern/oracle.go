//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoEmptyPattern(options any) rule.Bound { return bindOracle(rules.NoEmptyPattern, options) }

func oracleNoEmptyPatternOptions(fields []string) any {
	var options rules.NoEmptyPatternOptions
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
