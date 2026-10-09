//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	typescript "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoInferrableTypes() rule.Rule { return typescript.NoInferrableTypes }
func oracleNoInferrableTypesOptions(fields []string) any {
	var options typescript.NoInferrableTypesOptions
	if fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
