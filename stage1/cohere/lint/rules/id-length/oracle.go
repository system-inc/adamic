//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleIdentifierLength() rule.Rule { return rules.IdLength }
func oracleIdentifierLengthOptions(fields []string) any {
	// The all row is unconfigured for directory-owned rules. Its legacy option bag
	// configures the baseline adapters; named rows carry this rule's settings.
	if len(fields) > 1 && fields[1] == "all" {
		return nil
	}
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var object map[string]json.RawMessage
	if err := json.Unmarshal([]byte(fields[5]), &object); err != nil {
		panic(err)
	}
	if _, ok := object["Minimum"]; !ok {
		options, err := rules.DecodeIdLengthOptions([]byte(fields[5]))
		if err != nil {
			panic(err)
		}
		return options
	}
	var options rules.IdLengthSettings
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
