//go:build lintoracle

package main

import (
	"encoding/json"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleNoInvalidHtmlAttribute() rule.Rule { return rules.NoInvalidHtmlAttribute }
func oracleNoInvalidHtmlAttributeOptions(fields []string) any {
	if len(fields) > 5 && fields[1] == "react/no-invalid-html-attribute" && fields[5] != "" && fields[5] != "null" {
		var decoded rules.NoInvalidHtmlAttributeOptions
		if err := json.Unmarshal([]byte(fields[5]), &decoded); err != nil {
			panic(err)
		}
		return decoded
	}
	return rules.NoInvalidHtmlAttributeOptions{Attributes: []string{"rel"}}
}
