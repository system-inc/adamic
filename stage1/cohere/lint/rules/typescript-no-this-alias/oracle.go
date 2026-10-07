//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleThisAlias() rule.Rule { return rules.NoThisAlias }
func oracleThisAliasOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return rules.NoThisAliasOptions{}
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(fields[5]), &object); err != nil {
		panic(err)
	}
	if _, captured := object["ReportDestructuring"]; captured {
		var options rules.NoThisAliasOptions
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
		return options
	}
	options, err := rules.DecodeNoThisAliasOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
