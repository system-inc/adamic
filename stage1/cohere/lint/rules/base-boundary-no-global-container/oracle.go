//go:build lintoracle

package main

import (
	"encoding/json"

	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/base"
)

func oracleBaseBoundaryNoGlobalContainer() rule.Rule { return rules.BoundaryNoGlobalContainer }

func oracleBaseBoundaryNoGlobalContainerOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	// Upstream declares schema [] and no options type. Decode supplied JSON
	// rather than dropping it; the upstream listener deliberately ignores it.
	var decoded json.RawMessage
	if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
		panic(err)
	}
	return decoded
}
