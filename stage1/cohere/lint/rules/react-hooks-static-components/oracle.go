//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	reactRules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleReactHooksStaticComponents() rule.Rule                 { return reactRules.StaticComponents }
func oracleReactHooksStaticComponentsOptions(fields []string) any { return nil }
