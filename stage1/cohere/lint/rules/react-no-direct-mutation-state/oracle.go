//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	reactRules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactNoDirectMutationState() rule.Rule                 { return reactRules.NoDirectMutationState }
func oracleReactNoDirectMutationStateOptions(fields []string) any { return nil }
