//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleForDirection() rule.Rule                 { return rules.ForDirection }
func oracleForDirectionOptions(fields []string) any { return nil }
