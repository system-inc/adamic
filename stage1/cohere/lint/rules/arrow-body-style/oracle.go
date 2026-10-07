//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
	"strings"
)

func oracleArrowBody() rule.Rule { return rules.ArrowBodyStyle }
func oracleArrowBodyOptions(fields []string) any {
	if len(fields) < 6 || fields[5] == "" || fields[5] == "null" {
		return nil
	}
	if strings.HasPrefix(fields[5], "[") {
		options, err := rules.DecodeArrowBodyStyleOptions([]byte(fields[5]))
		if err != nil {
			panic(err)
		}
		return options
	}
	var options rules.ArrowBodyStyleOptions
	if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
		panic(err)
	}
	return options
}
