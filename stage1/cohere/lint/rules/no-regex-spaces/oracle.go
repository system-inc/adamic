//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoRegexSpaces() rule.Rule                 { return rules.NoRegexSpaces }
func oracleNoRegexSpacesOptions(fields []string) any { return nil }
