//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleWave18NoAdjacentInlineElements() rule.Rule { return rules.NoAdjacentInlineElements }

func oracleWave18NoAdjacentInlineElementsOptions(fields []string) any { return nil }
