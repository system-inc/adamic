//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoEmptyCharacterClass() rule.Rule { return rules.NoEmptyCharacterClass }

// The upstream rule has no options type and never reads options. Decode supplied
// JSON instead of silently dropping it; retain a non-nil value for the guard.
func oracleNoEmptyCharacterClassOptions(fields []string) any {
	var options any = struct{}{}
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
