//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUselessRename() rule.Rule { return rules.NoUselessRename }
func oracleNoUselessRenameOptions(fields []string) any {
	settings := rules.DefaultNoUselessRenameSettings()
	if fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &settings); err != nil {
			panic(err)
		}
	}
	return settings
}
