//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleWave14NoBaseToString() rule.Rule { return rules.NoBaseToString }
func oracleWave14NoBaseToStringOptions(fields []string) any {
	options := rules.DefaultNoBaseToStringSettings()
	if len(fields) >= 6 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
