//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	typescriptRules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleConsistentTypeAssertions() rule.Rule { return typescriptRules.ConsistentTypeAssertions }
func oracleConsistentTypeAssertionsOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	value, err := typescriptRules.DecodeConsistentTypeAssertionsOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return value
}
