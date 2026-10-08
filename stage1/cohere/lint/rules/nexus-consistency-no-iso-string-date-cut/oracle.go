//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleWave30IsoStringDateCut() rule.Rule { return rules.ConsistencyNoIsoStringDateCut }
func oracleWave30IsoStringDateCutOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	// The upstream rule has no options type and ignores its options argument.
	var options any
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
