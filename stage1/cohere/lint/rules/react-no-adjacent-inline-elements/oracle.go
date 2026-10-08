//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleWave22NoAdjacentInlineElements() rule.Rule                 { return rules.NoAdjacentInlineElements }
func oracleWave22NoAdjacentInlineElementsOptions(fields []string) any { return nil }
