//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleNexusNoStutteringName(options any) rule.Bound {
	return bindOracle(nexus.ConsistencyNoStutteringName, options)
}
func oracleNexusNoStutteringNameOptions(fields []string) any {
	var options nexus.ConsistencyNoStutteringNameOptions
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
