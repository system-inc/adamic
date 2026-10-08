//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
	"strings"
)

func oracleJsxFragments() rule.Rule { return rules.JsxFragments }
func oracleJsxFragmentsOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	var options rules.JsxFragmentsOptions
	if strings.HasPrefix(fields[5], "{") {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
		return options
	}
	optionsValue, err := rules.DecodeJsxFragmentsOptions([]byte(fields[5]))
	if err != nil {
		panic(err)
	}
	return optionsValue
}
