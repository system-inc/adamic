//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleModuleAlias() rule.Rule { return nexus.ImportRequireModuleAlias }
func oracleModuleAliasOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var options nexus.ImportRequireModuleAliasOptions
	if err := rule.UnmarshalOptions([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
