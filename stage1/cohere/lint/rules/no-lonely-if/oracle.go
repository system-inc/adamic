//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoLonelyIf() rule.Rule                 { return rules.NoLonelyIf }
func oracleNoLonelyIfOptions(fields []string) any { return nil }
