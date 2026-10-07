//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoEmptyObjectType() rule.Rule { return rules.NoEmptyObjectType }
func oracleNoEmptyObjectTypeOptions(fields []string) any {
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		value, err := rules.DecodeNoEmptyObjectTypeOptions([]byte(fields[5]))
		if err != nil {
			panic(err)
		}
		return value
	}
	return rules.DefaultNoEmptyObjectTypeOptions()
}
