//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoDuplicateTypeConstituents() rule.Rule { return rules.NoDuplicateTypeConstituents }
func oracleNoDuplicateTypeConstituentsOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var wire map[string]json.RawMessage
	if err := json.Unmarshal([]byte(fields[5]), &wire); err != nil {
		panic(err)
	}
	if _, present := wire["IgnoreIntersections"]; present {
		options := rules.DefaultNoDuplicateTypeConstituentsSettings()
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
		return options
	}
	if _, present := wire["IgnoreUnions"]; present {
		options := rules.DefaultNoDuplicateTypeConstituentsSettings()
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
		return options
	}
	options, err := rules.DecodeNoDuplicateTypeConstituentsOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
