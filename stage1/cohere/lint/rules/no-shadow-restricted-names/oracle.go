//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave29NoShadowRestrictedNames() rule.Rule { return rules.NoShadowRestrictedNames }
func oracleWave29NoShadowRestrictedNamesOptions(fields []string) any {
	var options rules.NoShadowRestrictedNamesOptions
	if len(fields) >= 6 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
