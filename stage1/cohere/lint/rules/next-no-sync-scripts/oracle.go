//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/next"
)

func oracleNextNoSyncScripts() rule.Rule                 { return rules.NoSyncScripts }
func oracleNextNoSyncScriptsOptions(fields []string) any { return nil }
