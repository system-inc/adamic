//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoSparseArrays() rule.Rule                 { return rules.NoSparseArrays }
func oracleNoSparseArraysOptions(fields []string) any { return nil }
