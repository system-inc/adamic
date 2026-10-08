//go:build lintoracle

package main

import (
	"bytes"
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleForbidDomProps() rule.Rule { return rules.ForbidDomProps }
func oracleForbidDomPropsOptions(fields []string) any {
	var raw []byte
	if len(fields) > 5 {
		raw = []byte(fields[5])
	}
	if len(fields) > 1 && (fields[1] == "all" || fields[1] == "") || bytes.Contains(raw, []byte(`"Forbid"`)) {
		var options rules.ForbidDomPropsOptions
		if len(raw) > 0 {
			if err := json.Unmarshal(raw, &options); err != nil {
				panic(err)
			}
		}
		return options
	}
	options, err := rules.DecodeForbidDomPropsOptions(raw)
	if err != nil {
		panic(err)
	}
	return options
}
