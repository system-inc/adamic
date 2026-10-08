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
	// Match the permissive typed adapters used by no-underscore-dangle and
	// no-unsafe-optional-chaining for the inherited all-rule option bag.
	// Captured upstream rows also already contain decoded typed options.
	if len(fields) > 1 && (fields[1] == "all" || fields[1] == "") || bytes.Contains(raw, []byte(`"Props"`)) {
		options := rules.DefaultJsxCurlyBracePresenceOptions()
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &options); err != nil {
				panic(err)
			}
		}
		return options
	}
	options, err := rules.DecodeJsxCurlyBracePresenceOptions(raw)
	if err != nil {
		panic(err)
	}
	return options
}
