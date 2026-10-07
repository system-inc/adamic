//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleTslint() rule.Rule                 { return rules.BanTslintComment }
func oracleTslintOptions(fields []string) any { return nil }
