//go:build lintoracle

package main

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/system-inc/cohere/internal/edit"
	"github.com/system-inc/cohere/internal/lint/rule"
)

func main() {
	source := "debugger; debugger;"
	result, err := edit.FixText("fixture.ts", source, func(_ string, text string) ([]edit.Proposal, error) {
		return []edit.Proposal{{RuleName: "no-debugger", Fix: rule.ReplaceRange(core.NewTextRange(0, 9), text[0:9])}, {RuleName: "no-debugger", Fix: rule.ReplaceRange(core.NewTextRange(10, 19), text[10:19])}}, nil
	}, 10)
	if err != nil {
		panic(err)
	}
	for _, r := range result.Rejected {
		fmt.Printf("rejected %s %d %d %s %s\n", r.Proposal.RuleName, r.Proposal.Fix.Range.Pos(), r.Proposal.Fix.Range.End(), r.ConflictsWith, r.Reason)
	}
	fmt.Printf("fixed\t%s\n", result.Text)
}
