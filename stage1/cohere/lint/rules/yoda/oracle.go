//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleYoda() rule.Rule { return rules.Yoda }
func oracleYodaOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	if fields[5][0] == '[' {
		options, err := rules.DecodeYodaOptions([]byte(fields[5]))
		if err != nil {
			panic(err)
		}
		return options
	}
	var options rules.YodaSettings
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
