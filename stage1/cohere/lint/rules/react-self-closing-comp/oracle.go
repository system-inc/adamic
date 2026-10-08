//go:build lintoracle

// Use the unchanged upstream rule with its defaulted, concrete options type.
package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleSelfClosingComp() rule.Rule { return rules.SelfClosingComp }
func oracleSelfClosingCompOptions(fields []string) any {
	settings := rules.DefaultSelfClosingCompOptions()
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &settings); err != nil {
			panic(err)
		}
	}
	return settings
}
