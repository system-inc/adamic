//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleNexusConsistencyNoShouting() rule.Rule { return rules.ConsistencyNoShouting }

func oracleNexusConsistencyNoShoutingOptions(fields []string) any {
	var decoded rules.ConsistencyNoShoutingOptions
	if fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
	}
	return decoded
}
