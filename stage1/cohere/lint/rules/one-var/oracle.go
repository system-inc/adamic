//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleOneVar() rule.Rule { return rules.OneVar }
func oracleOneVarOptions(fields []string) any {
	if fields[5] == "" || fields[5] == "null" {
		return rules.DefaultOneVarSettings()
	}
	var decoded rules.OneVarSettings
	if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
		panic(err)
	}
	return decoded
}
