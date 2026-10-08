//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleWave15PreferArrowCallback() rule.Rule { return rules.PreferArrowCallback }
func oracleWave15PreferArrowCallbackOptions(fields []string) any {
	if len(fields) <= 5 || fields[5] == "" {
		return rules.DefaultPreferArrowCallbackSettings()
	}
	options, err := rules.DecodePreferArrowCallbackOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
