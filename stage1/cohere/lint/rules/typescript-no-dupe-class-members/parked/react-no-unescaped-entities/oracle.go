//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleNoUnescapedEntities() rule.Rule { return rules.NoUnescapedEntities }
func oracleNoUnescapedEntitiesOptions(fields []string) any {
	if len(fields) > 5 && fields[1] == "react/no-unescaped-entities" && fields[5] != "" && fields[5] != "null" {
		decoded, err := rules.DecodeNoUnescapedEntitiesOptions([]byte(fields[5]))
		if err != nil {
			panic(err)
		}
		return decoded
	}
	return rules.NoUnescapedEntitiesOptions{}
}
