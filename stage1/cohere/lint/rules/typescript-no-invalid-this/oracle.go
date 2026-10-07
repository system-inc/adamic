//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleThis() rule.Rule { return rules.NoInvalidThis }
func oracleThisOptions(fields []string) any {
	settings := rules.DefaultNoInvalidThisSettings()
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		var wire struct {
			CapIsConstructor *bool `json:"capIsConstructor"`
		}
		if err := json.Unmarshal([]byte(fields[5]), &wire); err != nil {
			panic(err)
		}
		if wire.CapIsConstructor != nil {
			settings.CapIsConstructor = *wire.CapIsConstructor
		}
	}
	return settings
}
