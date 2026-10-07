//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/tailwind"
)

func oracleTailwindNoPhysicalDirection() rule.Rule                 { return rules.NoPhysicalDirection }
func oracleTailwindNoPhysicalDirectionOptions(fields []string) any { return nil }
