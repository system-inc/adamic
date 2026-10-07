//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
	"strings"
)

func oracleGroupedAccessors() rule.Rule { return rules.GroupedAccessorPairs }
func oracleGroupedAccessorsOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	if strings.HasPrefix(fields[5], "[") {
		options, err := rules.DecodeGroupedAccessorPairsOptions([]byte(fields[5]))
		if err != nil {
			panic(err)
		}
		return options
	}
	var options rules.GroupedAccessorPairsOptions
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
