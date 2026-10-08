//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave16UseIsnan() rule.Rule { return rules.UseIsNaN }
func oracleWave16UseIsnanOptions(fields []string) any {
	settings := rules.UseIsNaNOptions{}
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &settings); err != nil {
			panic(err)
		}
	}
	return settings
}
