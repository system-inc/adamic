//go:build lintoracle

// Built through the cohere overlay, with the pinned Go rule unchanged.
package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoCondAssign(options any) rule.Bound { return bindOracle(rules.NoCondAssign, options) }
func oracleNoCondAssignOptions(fields []string) any {
	if fields[5] == "" || fields[5][0] != '"' {
		return nil
	}
	var options rules.NoCondAssignOptions
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
