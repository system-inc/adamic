//go:build lintoracle

package main

import (
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/lint/rule"
	rules "github.com/system-inc/cohere/internal/lint/rules/core"
	"strings"
)

func oracleNoDebugger() rule.Rule {
	subject := rules.NoDebugger
	run := subject.Run
	subject.Run = func(ctx rule.Context, options any) rule.Listeners {
		report := ctx.Report
		ctx.Report = func(d rule.Diagnostic) {
			d.Fixes = nil
			if strings.HasPrefix(ctx.SourceFile.Text(), "/*😀*/") {
				start, end := d.Range.Pos(), d.Range.End()
				d.Fixes = []rule.Fix{
					{Range: core.NewTextRange(start, start+1), Text: "d"},
					{Range: core.NewTextRange(start, start+8), Text: ""},
					{Range: core.NewTextRange(2, 6), Text: "ok"},
					{Range: core.NewTextRange(end, end), Text: ""},
				}
			}
			report(d)
		}
		return run(ctx, options)
	}
	return subject
}
func oracleNoDebuggerOptions(fields []string) any { return nil }
