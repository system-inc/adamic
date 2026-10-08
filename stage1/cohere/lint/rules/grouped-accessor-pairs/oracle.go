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
	// The all row is unconfigured for directory-owned rules. Its legacy option bag
	// configures the baseline adapters; named rows carry this rule's settings.
	if len(fields) > 1 && fields[1] == "all" {
		return nil
	}
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
