//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoNewNativeNonconstructor() rule.Rule                 { return rules.NoNewNativeNonconstructor }
func oracleNoNewNativeNonconstructorOptions(fields []string) any { return nil }
