//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
	"strings"
)

func oracleWave29IdDenylist() rule.Rule { return rules.IdDenylist }
func oracleWave29IdDenylistOptions(fields []string) any {
	var settings rules.IdDenylistSettings
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return settings
	}
	raw := strings.TrimSpace(fields[5])
	if strings.HasPrefix(raw, "[") {
		value, err := rules.DecodeIdDenylistOptions([]byte(raw))
		if err != nil {
			panic(err)
		}
		return value
	}
	if err := json.Unmarshal([]byte(raw), &settings); err != nil {
		panic(err)
	}
	return settings
}
