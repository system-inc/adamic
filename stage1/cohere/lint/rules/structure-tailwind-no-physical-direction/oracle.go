//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/tailwind"
)

func oraclePhysicalDirection() rule.Rule                 { return rules.NoPhysicalDirection }
func oraclePhysicalDirectionOptions(fields []string) any { return nil }
