//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUnsafeNegation(options any) rule.Bound {
	return bindOracle(rules.NoUnsafeNegation, options)
}
func oracleNoUnsafeNegationOptions(fields []string) any {
	var options rules.NoUnsafeNegationOptions
	if len(fields) > 5 && fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
