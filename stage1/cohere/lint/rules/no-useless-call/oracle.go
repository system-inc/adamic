//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUselessCall() rule.Rule { return rules.NoUselessCall }

// Upstream accepts no configurable options. Decode supplied JSON nonetheless,
// preserving it rather than silently dropping an options-bearing manifest row.
func oracleNoUselessCallOptions(fields []string) any {
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		var options any
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
		return options
	}
	return nil
}
