//go:build lintoracle

package main

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoCaller() rule.Rule { return rules.NoCaller }

// NoCaller defines no options struct and ignores its any argument. Decode captured
// JSON nevertheless so an explicit options row never silently drops its payload.
func oracleNoCallerOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" {
		return nil
	}
	var decoded any
	if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
		panic(err)
	}
	return decoded
}
