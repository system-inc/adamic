//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoProto() rule.Rule { return rules.NoProto }

// Upstream has no configuration type. Preserve supplied JSON rather than silently
// dropping it: the oracle's option guard also applies to configuration-free rules.
func oracleNoProtoOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var options any
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
