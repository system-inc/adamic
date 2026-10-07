//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	typescript "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleMethodSignatureStyle() rule.Rule { return typescript.MethodSignatureStyle }

// oracleMethodSignatureStyleOptions decodes field 5's captured options into the upstream type, and with none leaves the
// rule its defaults, as the shared oracle's decoder did before this rule moved.
func oracleMethodSignatureStyleOptions(fields []string) any {
	if fields[5] == "" {
		return nil
	}
	var decoded typescript.MethodSignatureStyleOptions
	if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
		panic(err)
	}
	return decoded
}
