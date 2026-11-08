//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoSelfCompare(options any) rule.Bound { return bindOracle(rules.NoSelfCompare, options) }

// Upstream accepts no configurable options. Decode supplied JSON nonetheless,
// preserving it rather than silently dropping an options-bearing manifest row.
func oracleNoSelfCompareOptions(fields []string) any {
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		var options any
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
		return options
	}
	return nil
}
