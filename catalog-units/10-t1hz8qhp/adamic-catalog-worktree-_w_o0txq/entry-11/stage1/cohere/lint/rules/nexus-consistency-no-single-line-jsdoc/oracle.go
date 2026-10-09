//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleNexusConsistencyNoSingleLineJsdoc() rule.Rule { return rules.ConsistencyNoSingleLineJsDoc }

// Upstream declares no options type and ignores options. Decode supplied JSON explicitly;
// returning the decoded value preserves that contract without dropping a captured options row.
func oracleNexusConsistencyNoSingleLineJsdocOptions(fields []string) any {
	decoded := any(struct{}{})
	if fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
	}
	return decoded
}
