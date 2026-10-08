//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleCheckerCaughtError() rule.Rule { return rules.PreserveCaughtError }
func oracleCheckerCaughtErrorOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var options rules.PreserveCaughtErrorOptions
	if err := rule.UnmarshalOptions([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
