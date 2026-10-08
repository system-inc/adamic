//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoOctal() rule.Rule                 { return rules.NoOctal }
func oracleNoOctalOptions(fields []string) any { return nil }
