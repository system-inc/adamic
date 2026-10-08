//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	reactRules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactStateInConstructor() rule.Rule { return reactRules.StateInConstructor }
func oracleReactStateInConstructorOptions(fields []string) any {
	options := reactRules.DefaultStateInConstructorOptions()
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		var mode string
		if err := json.Unmarshal([]byte(fields[5]), &mode); err != nil {
			if err = json.Unmarshal([]byte(fields[5]), &options); err != nil {
				panic(err)
			}
			return options
		}
		if mode == "never" {
			options.Mode = reactRules.StateInConstructorNever
		}
	}
	return options
}
