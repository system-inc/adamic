//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleBooleanOutcome() rule.Rule { return rules.ConsistencyNoBooleanOutcome }
func oracleBooleanOutcomeOptions(fields []string) any {
	var options rules.ConsistencyNoBooleanOutcomeOptions
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
