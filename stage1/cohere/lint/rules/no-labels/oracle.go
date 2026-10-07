//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoLabels() rule.Rule { return rules.NoLabels }

// oracleNoLabelsOptions decodes field 5's captured options into the upstream type, and with none gives
// the type's zero value, as the shared oracle's decoder did before this rule moved.
func oracleNoLabelsOptions(fields []string) any {
	var decoded rules.NoLabelsOptions
	if fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
	}
	return decoded
}
