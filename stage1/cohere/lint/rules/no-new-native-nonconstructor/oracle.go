//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleCheckerNoNewNativeNonconstructor() rule.Rule                 { return rules.NoNewNativeNonconstructor }
func oracleCheckerNoNewNativeNonconstructorOptions(fields []string) any { return nil }
