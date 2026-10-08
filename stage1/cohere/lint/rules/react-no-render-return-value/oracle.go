//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactNoRenderReturnValue() rule.Rule                 { return rules.NoRenderReturnValue }
func oracleReactNoRenderReturnValueOptions(fields []string) any { return nil }
