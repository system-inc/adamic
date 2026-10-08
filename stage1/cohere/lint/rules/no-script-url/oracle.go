//go:build lintoracle

package main

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoScriptUrl() rule.Rule { return rules.NoScriptUrl }

// NoScriptUrl has no upstream option fields. Decode the complete JSON value anyway:
// supplied options must reach the unchanged upstream rule, not a nil adapter.
func oracleNoScriptUrlOptions(fields []string) any {
	if fields[5] == "" || fields[5] == "null" {
		return struct{}{}
	}
	var decoded any
	if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
		panic(err)
	}
	if decoded == nil {
		return struct{}{}
	}
	return decoded
}
