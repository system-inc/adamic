//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleConsistentThis() rule.Rule { return rules.ConsistentThis }
func oracleConsistentThisOptions(fields []string) any {
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		var decoded rules.ConsistentThisSettings
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
		return decoded
	}
	return nil
}
