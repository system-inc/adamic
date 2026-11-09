//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleConsistencyNoAbbreviatedIdentifier(options any) rule.Bound {
	return bindOracle(nexus.ConsistencyNoAbbreviatedIdentifier, options)
}
func oracleConsistencyNoAbbreviatedIdentifierOptions(fields []string) any {
	var options nexus.ConsistencyNoAbbreviatedIdentifierOptions
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
