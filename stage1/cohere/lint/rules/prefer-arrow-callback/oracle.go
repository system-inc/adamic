//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oraclePreferArrowCallback() rule.Rule { return rules.PreferArrowCallback }
func oraclePreferArrowCallbackOptions(fields []string) any {
	options := rules.DefaultPreferArrowCallbackSettings()
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
