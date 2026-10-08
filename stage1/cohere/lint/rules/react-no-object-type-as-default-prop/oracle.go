//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/react"
)

func oracleWave22NoObjectTypeAsDefaultProp() rule.Rule                 { return rules.NoObjectTypeAsDefaultProp }
func oracleWave22NoObjectTypeAsDefaultPropOptions(fields []string) any { return nil }
