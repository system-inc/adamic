//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleMaxLines() rule.Rule { return rules.MaxLines }
func oracleMaxLinesOptions(fields []string) any {
	settings := rules.DefaultMaxLinesOptions()
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		raw := []byte(fields[5])
		var maximum int
		if json.Unmarshal(raw, &maximum) == nil {
			settings.Maximum = maximum
		} else if err := json.Unmarshal(raw, &settings); err != nil {
			panic(err)
		}
		var object struct{ Max *int }
		json.Unmarshal(raw, &object)
		if object.Max != nil {
			settings.Maximum = *object.Max
		}
	}
	return settings
}
