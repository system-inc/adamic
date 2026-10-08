//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave22NoImplicitGlobals() rule.Rule { return rules.NoImplicitGlobals }
func oracleWave22NoImplicitGlobalsOptions(fields []string) any {
	options := rules.DefaultNoImplicitGlobalsSettings()
	if len(fields) > 5 && fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
