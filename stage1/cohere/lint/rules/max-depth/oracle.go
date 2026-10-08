//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleMaxDepth() rule.Rule { return rules.MaxDepth }

func oracleMaxDepthOptions(fields []string) any {
	settings := rules.DefaultMaxDepthSettings()
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		if err := json.Unmarshal([]byte(fields[5]), &settings); err != nil {
			panic(err)
		}
	}
	return settings
}
