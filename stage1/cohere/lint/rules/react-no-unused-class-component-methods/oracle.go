//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	reactRules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactNoUnusedClassComponentMethods() rule.Rule {
	return reactRules.NoUnusedClassComponentMethods
}

// The upstream registration has no Decode and the rule ignores options.
func oracleReactNoUnusedClassComponentMethodsOptions(fields []string) any { return nil }
