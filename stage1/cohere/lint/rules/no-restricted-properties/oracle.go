//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoRestrictedProperties() rule.Rule { return rules.NoRestrictedProperties }
func oracleNoRestrictedPropertiesOptions(fields []string) any {
	var options rules.NoRestrictedPropertiesOptions
	if len(fields) > 5 && fields[5] != "" {
		if err := json.Unmarshal([]byte(fields[5]), &options); err != nil {
			panic(err)
		}
	}
	return options
}
