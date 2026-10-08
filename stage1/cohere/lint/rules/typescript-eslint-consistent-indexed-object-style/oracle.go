//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleConsistentIndexedObjectStyle() rule.Rule { return rules.ConsistentIndexedObjectStyle }
func oracleConsistentIndexedObjectStyleOptions(fields []string) any {
	options := rules.DefaultConsistentIndexedObjectStyleSettings()
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return options
	}
	if fields[5][0] == '"' {
		value, err := rules.DecodeConsistentIndexedObjectStyleOptions([]byte(fields[5]))
		if err != nil {
			panic(err)
		}
		return value
	}
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
