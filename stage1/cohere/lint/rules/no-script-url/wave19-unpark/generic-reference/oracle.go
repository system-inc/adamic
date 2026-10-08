//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleWave19GenericConstructors() rule.Rule { return rules.ConsistentGenericConstructors }
func oracleWave19GenericConstructorsOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return rules.DefaultConsistentGenericConstructorsSettings()
	}
	var options rules.ConsistentGenericConstructorsOptions
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
