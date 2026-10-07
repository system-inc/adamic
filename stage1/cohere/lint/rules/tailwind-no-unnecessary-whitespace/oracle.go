//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/tailwind"
)

func oracleNoUnnecessaryWhitespace() rule.Rule { return rules.NoUnnecessaryWhitespace }
func oracleNoUnnecessaryWhitespaceOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var settings rules.NoUnnecessaryWhitespaceOptions
	if err := json.Unmarshal([]byte(fields[5]), &settings); err != nil {
		panic(err)
	}
	return settings
}
