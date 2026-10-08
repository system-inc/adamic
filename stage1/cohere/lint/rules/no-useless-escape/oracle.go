//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUselessEscape() rule.Rule { return rules.NoUselessEscape }
func oracleNoUselessEscapeOptions(fields []string) any {
	var options rules.NoUselessEscapeOptions
	if len(fields) > 5 && fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
