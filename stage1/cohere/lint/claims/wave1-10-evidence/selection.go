package main
import (
 "github.com/system-inc/cohere/internal/lint/rule"
 "github.com/system-inc/cohere/internal/lint/rules/core"
 "github.com/system-inc/cohere/internal/lint/rules/react"
)
type registeredRule struct { subject rule.Rule; options func([]string) any }
func registeredRules() []registeredRule { return []registeredRule{
 {core.SortVars, func([]string) any { return core.DefaultSortVarsSettings() }},
 {react.SelfClosingComp, func([]string) any { return react.DefaultSelfClosingCompOptions() }},
} }
