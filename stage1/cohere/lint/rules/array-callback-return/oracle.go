//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleArrayCallback() rule.Rule { return rules.ArrayCallbackReturn }
func oracleArrayCallbackOptions(fields []string) any {
	// The all row is unconfigured for directory-owned rules. Its legacy option bag
	// configures the baseline adapters; named rows carry this rule's settings.
	if len(fields) > 1 && fields[1] == "all" {
		return nil
	}
	var options rules.ArrayCallbackReturnOptions
	if len(fields) > 5 && fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
