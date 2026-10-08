//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	reactRules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleRequireRenderReturn() rule.Rule { return reactRules.RequireRenderReturn }

// The upstream rule has no options or decoder.
func oracleRequireRenderReturnOptions(fields []string) any { return nil }
