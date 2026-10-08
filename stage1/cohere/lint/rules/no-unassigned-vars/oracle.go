//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleCheckerUnassigned() rule.Rule                 { return rules.NoUnassignedVars }
func oracleCheckerUnassignedOptions(fields []string) any { return nil }
