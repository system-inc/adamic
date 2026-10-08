//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUselessConcat() rule.Rule { return rules.NoUselessConcat }

// This upstream rule has no options type or schema entries and ignores its options argument.
// Decode explicit JSON instead of dropping it, including on all-rule rows with other rules' bags.
func oracleNoUselessConcatOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var options any
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
