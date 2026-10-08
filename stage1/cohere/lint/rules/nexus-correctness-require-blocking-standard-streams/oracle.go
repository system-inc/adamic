//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleWave05RequireBlockingStandardStreams() rule.Rule {
	return rules.CorrectnessRequireBlockingStandardStreams
}
func oracleWave05RequireBlockingStandardStreamsOptions(fields []string) any {
	var options struct{}
	if len(fields) > 5 && fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
