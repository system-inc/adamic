//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNextNoImgElement() rule.Rule                 { return rules.NoImgElement }
func oracleNextNoImgElementOptions(fields []string) any { return nil }
