//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactNoDeprecated() rule.Rule                 { return rules.NoDeprecated }
func oracleReactNoDeprecatedOptions(fields []string) any { return nil }
