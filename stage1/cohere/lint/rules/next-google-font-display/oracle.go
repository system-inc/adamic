//go:build lintoracle
package main
import (
 "github.com/system-inc/cohere/internal/lint/rule"
 rules "github.com/system-inc/cohere/internal/lint/rules/next"
)
func oracleGoogleFontDisplay() rule.Rule {return rules.GoogleFontDisplay}
func oracleGoogleFontDisplayOptions(fields []string) any {return nil}
