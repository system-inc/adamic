//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleJsxFragments() rule.Rule { return rules.JsxFragments }
func oracleJsxFragmentsOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" {
		return rules.DefaultJsxFragmentsOptions()
	}
	var decoded struct{ Mode rules.JsxFragmentsMode }
	if fields[5][0] == '{' {
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
		return rules.JsxFragmentsOptions{Mode: decoded.Mode}
	}
	options, err := rules.DecodeJsxFragmentsOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return options
}
