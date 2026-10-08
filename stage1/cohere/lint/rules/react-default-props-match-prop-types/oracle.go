//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactDefaultPropsMatchPropTypes() rule.Rule { return rules.DefaultPropsMatchPropTypes }
func oracleReactDefaultPropsMatchPropTypesOptions(fields []string) any {
	settings := rules.DefaultDefaultPropsMatchPropTypesOptions()
	if len(fields) > 5 && fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &settings); err != nil {
			panic(err)
		}
	}
	return settings
}
