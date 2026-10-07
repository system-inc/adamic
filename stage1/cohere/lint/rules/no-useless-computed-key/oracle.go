//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleComputedKey() rule.Rule { return rules.NoUselessComputedKey }
func oracleComputedKeyOptions(fields []string) any {
	var value rules.NoUselessComputedKeySettings = rules.DefaultNoUselessComputedKeySettings()
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &value); err != nil {
			panic(err)
		}
	}
	return value
}
