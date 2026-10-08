//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	reactRules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactNoTypos() rule.Rule                 { return reactRules.NoTypos }
func oracleReactNoTyposOptions(fields []string) any { return nil }
