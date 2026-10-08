//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	reactRules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactNoThisInSfc() rule.Rule { return reactRules.NoThisInSfc }

// Upstream declares no options and ignores the options argument.
func oracleReactNoThisInSfcOptions(fields []string) any { return nil }
