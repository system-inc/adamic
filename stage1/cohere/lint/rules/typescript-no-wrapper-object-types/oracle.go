//go:build lintoracle

// Built only through the cohere overlay, so the upstream rule remains the oracle.
package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	typescript "github.com/system-inc/cohere/internal/lint/rules/typescript"
)

func oracleNoWrapperObjectTypes() rule.Rule                 { return typescript.NoWrapperObjectTypes }
func oracleNoWrapperObjectTypesOptions(fields []string) any { return nil }
