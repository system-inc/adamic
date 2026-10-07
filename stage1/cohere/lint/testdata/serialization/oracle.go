//go:build lintoracle

package main

import (
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
)

func oracleNoDebugger() rule.Rule {
	subject := rules.NoDebugger
	run := subject.Run
	subject.Run = func(ctx rule.Context, options any) rule.Listeners {
		report := ctx.Report
		ctx.Report = func(d rule.Diagnostic) {
			start := d.Range.Pos()
			d.Fixes = nil
			d.Suggestions = []rule.Suggestion{
				{Message: rule.Message{Id: "first", Description: "first | :\n😀"}, Fixes: []rule.Fix{
					{Range: core.NewTextRange(start, start+1), Text: "a|:\t\n😀\\"},
					{Range: core.NewTextRange(start+1, start+2), Text: ""},
				}},
				{Message: rule.Message{Id: "second", Description: "second"}, Fixes: []rule.Fix{
					{Range: core.NewTextRange(start+2, start+2), Text: "?."},
				}},
				{Message: rule.Message{Id: "empty", Description: "no edits"}},
			}
			report(d)
		}
		return run(ctx, options)
	}
	return subject
}
func oracleNoDebuggerOptions(fields []string) any { return nil }
