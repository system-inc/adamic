//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUndefInit() rule.Rule                 { return rules.NoUndefInit }
func oracleNoUndefInitOptions(fields []string) any { return nil }
