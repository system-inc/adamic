//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleWave22NoUnusedPrivateClassMembers() rule.Rule                 { return rules.NoUnusedPrivateClassMembers }
func oracleWave22NoUnusedPrivateClassMembersOptions(fields []string) any { return nil }
