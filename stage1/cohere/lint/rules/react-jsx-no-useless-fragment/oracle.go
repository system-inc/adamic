//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleJsxNoUselessFragment() rule.Rule { return rules.JsxNoUselessFragment }
func oracleJsxNoUselessFragmentOptions(fields []string) any {
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		var decoded rules.JsxNoUselessFragmentOptions
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
		return decoded
	}
	return rules.JsxNoUselessFragmentOptions{}
}
