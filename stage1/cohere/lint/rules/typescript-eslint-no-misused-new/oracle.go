//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoMisusedNew() rule.Rule                 { return rules.NoMisusedNew }
func oracleNoMisusedNewOptions(fields []string) any { return nil }
