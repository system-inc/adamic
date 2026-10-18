//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoExtraBind() rule.Rule { return rules.NoExtraBind }

// The upstream rule has an empty options schema and ignores options.
func oracleNoExtraBindOptions(fields []string) any {
	if len(fields) > 5 && fields[5] != "" && fields[5] != "null" {
		var options map[string]json.RawMessage
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
		return options
	}
	return nil
}
