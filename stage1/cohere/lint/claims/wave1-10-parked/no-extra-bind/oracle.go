//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoExtraBind() rule.Rule                 { return rules.NoExtraBind }
func oracleNoExtraBindOptions(fields []string) any { return nil }
