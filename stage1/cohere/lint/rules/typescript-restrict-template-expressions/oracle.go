//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleWave05RestrictTemplateExpressions() rule.Rule { return rules.RestrictTemplateExpressions }
func oracleWave05RestrictTemplateExpressionsOptions(fields []string) any {
	var data []byte
	if len(fields) > 5 {
		data = []byte(fields[5])
	}
	var captured map[string]json.RawMessage
	if len(data) > 0 {
		if err := json.Unmarshal(data, &captured); err != nil {
			panic(err)
		}
		if _, ok := captured["AllowInline"]; ok {
			var options rules.RestrictTemplateExpressionsOptions
			if err := json.Unmarshal(data, &options); err != nil {
				panic(err)
			}
			return options
		}
	}
	options, err := rules.DecodeRestrictTemplateExpressionsOptions(data)
	if err != nil {
		panic(err)
	}
	return options
}
