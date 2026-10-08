//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleForeignPropTypes() rule.Rule { return rules.ForbidForeignPropTypes }
func oracleForeignPropTypesOptions(fields []string) any {
	var value rules.ForbidForeignPropTypesOptions
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &value); err != nil {
			panic(err)
		}
	}
	return value
}
