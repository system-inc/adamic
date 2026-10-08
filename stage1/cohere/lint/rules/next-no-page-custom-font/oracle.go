//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNextNoPageCustomFont() rule.Rule                 { return rules.NoPageCustomFont }
func oracleNextNoPageCustomFontOptions(fields []string) any { return nil }
