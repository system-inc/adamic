//go:build cohere_planted_panic

package main

import (
	"strings"

	"github.com/system-inc/cohere/internal/lint/rule"
)

// plantedPanicMarker is the text that makes the planted rule panic on a file.
const plantedPanicMarker = "cohere-test: panic here"

// A rule that panics on any file holding plantedPanicMarker, so a test can watch a whole run survive a
// rule's crash and say so. Compiled only under the cohere_planted_panic tag, which only the binary
// TestARuleCrashFailsTheRunAndNamesItselfAsCohere builds carries: no shipped binary has it.
func init() {
	rule.Register(rule.Registration{Rule: rule.Rule{
		Name: "cohere-test/panics-on-marked-file",
		Run: func(ctx rule.Context, options any) rule.Listeners {
			if strings.Contains(ctx.SourceFile.Text(), plantedPanicMarker) {
				panic("planted: Unhandled case in Node.Text: *ast.ComputedPropertyName")
			}
			return nil
		},
	}})
}
