//go:build lintoracle

// Built only through the cohere overlay; the unchanged Go rule is the oracle.
package main

import (
	"encoding/json"
	"strings"

	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactNoWillUpdateSetState() rule.Rule { return rules.NoWillUpdateSetState }
func oracleReactNoWillUpdateSetStateOptions(fields []string) any {
	options := rules.NoMethodSetStateOptions{}
	if len(fields) <= 5 || strings.TrimSpace(fields[5]) == "" {
		return options
	}
	raw := []byte(fields[5])
	if strings.HasPrefix(strings.TrimSpace(fields[5]), "\"") {
		decoded, err := rules.DecodeNoMethodSetStateOptions(raw)
		if err != nil {
			panic(err)
		}
		return decoded
	}
	if err := json.Unmarshal(raw, &options); err != nil {
		panic(err)
	}
	return options
}
