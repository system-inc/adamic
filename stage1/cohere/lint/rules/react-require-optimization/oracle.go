//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	reactRules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactRequireOptimization() rule.Rule { return reactRules.RequireOptimization }
func oracleReactRequireOptimizationOptions(fields []string) any {
	options := reactRules.RequireOptimizationOptions{}
	if len(fields) > 5 && fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
