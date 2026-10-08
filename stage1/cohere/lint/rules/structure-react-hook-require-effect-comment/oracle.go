//go:build lintoracle

package main

import (
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/rules/structure"
)

func oracleReactHookRequireEffectComment() rule.Rule                 { return structure.ReactHookRequireEffectComment }
func oracleReactHookRequireEffectCommentOptions(fields []string) any { return nil }
