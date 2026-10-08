//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	reactRules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactPreferStatelessFunction() rule.Rule { return reactRules.PreferStatelessFunction }
func oracleReactPreferStatelessFunctionOptions(fields []string) any {
	options := reactRules.DefaultPreferStatelessFunctionOptions()
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
