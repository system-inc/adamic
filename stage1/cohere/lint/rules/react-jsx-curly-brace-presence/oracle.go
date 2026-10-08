//go:build lintoracle

package main

import (
	"bytes"
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleJsxCurlyBracePresence() rule.Rule { return rules.JsxCurlyBracePresence }
func oracleJsxCurlyBracePresenceOptions(fields []string) any {
	var raw []byte
	if len(fields) > 5 {
		raw = []byte(fields[5])
	}
	if bytes.Contains(raw, []byte(`"Props"`)) {
		options := rules.DefaultJsxCurlyBracePresenceOptions()
		if err := json.Unmarshal(raw, &options); err != nil {
			panic(err)
		}
		return options
	}
	options, err := rules.DecodeJsxCurlyBracePresenceOptions(raw)
	if err != nil {
		panic(err)
	}
	return options
}
