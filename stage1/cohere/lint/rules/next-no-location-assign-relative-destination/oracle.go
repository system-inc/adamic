//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNextNoLocationAssignRelativeDestination() rule.Rule {
	return rules.NoLocationAssignRelativeDestination
}
func oracleNextNoLocationAssignRelativeDestinationOptions(fields []string) any { return nil }
