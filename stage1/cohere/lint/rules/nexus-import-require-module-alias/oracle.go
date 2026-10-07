//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleModuleAlias() rule.Rule { return nexus.ImportRequireModuleAlias }
func oracleModuleAliasOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var options nexus.ImportRequireModuleAliasOptions
	// The shared all-rule manifest contains other rules' option keys as well.
	// Exact rule selections still use cohere's strict decoder.
	if len(fields) > 1 && fields[1] == "all" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
		return options
	}
	if err := rule.UnmarshalOptions([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
