//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoUselessBackreference() rule.Rule                 { return rules.NoUselessBackreference }
func oracleNoUselessBackreferenceOptions(fields []string) any { return nil }
