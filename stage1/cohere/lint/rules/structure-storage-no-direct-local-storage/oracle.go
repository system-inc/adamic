//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleStructureStorageNoDirectLocalStorage() rule.Rule                 { return rules.StorageNoDirectLocalStorage }
func oracleStructureStorageNoDirectLocalStorageOptions(fields []string) any { return nil }
