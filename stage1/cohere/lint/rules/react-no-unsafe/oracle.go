//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleUnsafe() rule.Rule { return rules.NoUnsafe }
func oracleUnsafeOptions(fields []string) any {
	var settings rules.NoUnsafeOptions
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &settings); err != nil {
			panic(err)
		}
	}
	return settings
}
