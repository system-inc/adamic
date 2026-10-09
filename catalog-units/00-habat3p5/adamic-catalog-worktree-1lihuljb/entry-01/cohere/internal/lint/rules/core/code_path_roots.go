package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// codePathRoots is control_flow_graph.IndexRoots for the file, computed once and shared by the rules
// that build a control-flow graph per root: no-useless-return, no-unreachable-loop and
// no-useless-assignment. Each reads from it which roots could give it anything to report, and builds
// only those.
//
// Here rather than in control_flow_graph, whose only import is the shim, so a rule can adopt the
// graph without dragging the rule package behind it.
func codePathRoots(ctx rule.Context, sourceFile *ast.Node) []control_flow_graph.RootSummary {
	return rule.Cached(ctx.FileCache, "control_flow_graph.Roots", func() []control_flow_graph.RootSummary {
		return control_flow_graph.IndexRoots(sourceFile)
	})
}

// CheckCodePathGates makes each rule that skips a root by codePathRoots build that root's graph
// anyway and panic if the graph holds anything the rule could have reported. The gates are meant to
// be exact, never a guess, and this is how the tests hold them to it: every fixture and corpus row
// the rules run on becomes a check that no skipped root could report. Set only by tests, in an init
// function, before any rule runs.
var CheckCodePathGates bool
