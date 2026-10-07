//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleRestrictedTypes() rule.Rule { return rules.NoRestrictedTypes }
func oracleRestrictedTypesOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var keys map[string]json.RawMessage
	if err := json.Unmarshal([]byte(fields[5]), &keys); err != nil {
		panic(err)
	}
	if _, captured := keys["Types"]; captured {
		var result rules.NoRestrictedTypesOptions
		if err := json.Unmarshal([]byte(fields[5]), &result); err != nil {
			panic(err)
		}
		return result
	}
	result, err := rules.DecodeNoRestrictedTypesOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return result
}
