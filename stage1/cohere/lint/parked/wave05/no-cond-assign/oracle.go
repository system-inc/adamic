//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoCondAssign() rule.Rule { return rules.NoCondAssign }
func oracleNoCondAssignOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var mode rules.NoCondAssignOptions
	if err := json.Unmarshal([]byte(fields[5]), &mode); err != nil {
		panic(err)
	}
	return mode
}
