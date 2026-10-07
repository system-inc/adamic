//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoReturnAssign() rule.Rule { return rules.NoReturnAssign }

// oracleNoReturnAssignOptions decodes the rule's one option, a mode written as a bare JSON string such as
// "always". Anything else carries no mode: a null row, or an "all" row's object of other rules' options,
// leaves the rule its default, as the shared oracle's decoder did before this rule moved.
func oracleNoReturnAssignOptions(fields []string) any {
	if fields[5] == "" || fields[5][0] != '"' {
		return nil
	}
	var decoded rules.NoReturnAssignOptions
	if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
		panic(err)
	}
	return decoded
}
