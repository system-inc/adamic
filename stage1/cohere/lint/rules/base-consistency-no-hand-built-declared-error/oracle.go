//go:build lintoracle

package main

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/rule"
	base "github.com/system-inc/cohere/internal/lint/rules/base"
)

func oracleConsistencyNoHandBuiltDeclaredError() rule.Rule {
	return base.ConsistencyNoHandBuiltDeclaredError
}

func oracleConsistencyNoHandBuiltDeclaredErrorOptions(fields []string) any {
	// Upstream declares no options schema or options type; Run accepts any and ignores it.
	// Retain supplied JSON rather than returning nil and dropping an option-bearing row.
	if len(fields) <= 5 || fields[1] != "base/consistency-no-hand-built-declared-error" || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
		panic(err)
	}
	return decoded
}
