//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoDupeClassMembers() rule.Rule                 { return rules.NoDupeClassMembers }
func oracleNoDupeClassMembersOptions(fields []string) any { return nil }
