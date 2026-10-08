//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oraclePreferFind() rule.Rule { return rules.PreferFind }

// PreferFind has no options type; preserve supplied JSON for the oracle's guard.
func oraclePreferFindOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var options any
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
