//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleConstructorSuper() rule.Rule                 { return core.ConstructorSuper }
func oracleConstructorSuperOptions(fields []string) any { return nil }
