//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNoCssTags() rule.Rule                 { return rules.NoCssTags }
func oracleNoCssTagsOptions(fields []string) any { return nil }
