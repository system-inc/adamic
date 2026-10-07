//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oraclePathAlias() rule.Rule { return rules.ImportRequirePathAlias }
func oraclePathAliasOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" {
		return nil
	}
	var options rules.ImportRequirePathAliasOptions
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
