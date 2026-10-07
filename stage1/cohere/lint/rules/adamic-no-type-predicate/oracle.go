//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	adamic "github.com/system-inc/cohere/internal/lint/rules/adamic"
)

func oracleNoTypePredicate() rule.Rule                 { return adamic.NoTypePredicate }
func oracleNoTypePredicateOptions(fields []string) any { return nil }
