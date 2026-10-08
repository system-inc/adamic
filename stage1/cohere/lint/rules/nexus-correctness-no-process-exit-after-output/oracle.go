//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/nexus"
)

func oracleWave05NoProcessExitAfterOutput() rule.Rule {
	return rules.CorrectnessNoProcessExitAfterOutput
}
func oracleWave05NoProcessExitAfterOutputOptions(fields []string) any {
	var options struct{}
	if len(fields) > 5 && fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
