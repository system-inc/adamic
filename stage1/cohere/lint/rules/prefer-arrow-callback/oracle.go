//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oraclePreferArrowCallback() rule.Rule { return rules.PreferArrowCallback }
func oraclePreferArrowCallbackOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" {
		return rules.DefaultPreferArrowCallbackSettings()
	}
	var wire struct {
		AllowNamedFunctions bool
		AllowUnboundThis    *bool
	}
	if err := json.Unmarshal([]byte(fields[5]), &wire); err != nil {
		panic(err)
	}
	options := rules.DefaultPreferArrowCallbackSettings()
	options.AllowNamedFunctions = wire.AllowNamedFunctions
	if wire.AllowUnboundThis != nil {
		options.AllowUnboundThis = *wire.AllowUnboundThis
	}
	return options
}
