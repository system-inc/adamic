//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/tailwind"
)

func oraclePhysical() rule.Rule                 { return rules.NoPhysicalDirection }
func oraclePhysicalOptions(fields []string) any { return nil }
