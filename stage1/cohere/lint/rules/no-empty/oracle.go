//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoEmpty() rule.Rule { return rules.NoEmpty }
func oracleNoEmptyOptions(fields []string) any {
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		options := rules.NoEmptyOptions{AllowEmptyCatch: fields[4] == "true"}
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
		return options
	}

	return rules.NoEmptyOptions{AllowEmptyCatch: fields[4] == "true"}
}
