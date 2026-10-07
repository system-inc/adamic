//go:build lintoracle

package main

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleNexusConsistencyNoUtilsFolder() rule.Rule { return nexus.ConsistencyNoUtilsFolder }

// Upstream accepts any and ignores options; it declares no options struct.
// Preserve supplied JSON so the harness guard can distinguish it from dropped data.
func oracleNexusConsistencyNoUtilsFolderOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var options any
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
