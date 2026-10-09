package program

import (
	"context"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/incremental"
)

// IncrementalSession is one warm-capable pass over a graph: check, then persist what the next
// run needs to know which files it can skip.
//
// It exists as a type rather than as two methods on Graph because of the single most
// expensive thing to get wrong here. The build info must be emitted from the SAME incremental
// program that performed the checking. That program's snapshot records which files it
// checked, and the emit writes those records; a build info emitted from a freshly-wrapped
// program has an empty snapshot, so it marks every file as "unchanged but never checked",
// which on load means "check it".
//
// Measured, because the difference does not look like anything from the call site:
//
//	emitted from a fresh wrapper     warm semantic pass 712ms, nothing skipped
//	emitted from the checker         warm semantic pass   4ms
//
// The first version of this file had two independent methods and produced the 712ms. It was
// correct, it was slower than not caching at all, and every test passed. Holding the program
// across both calls is what makes the type worth its weight.
//
// THE RISK, stated where the code is rather than in a task nobody will read: a cache that
// fails to invalidate reports zero findings forever, and that is indistinguishable from a
// clean tree on every measurement except a planted violation. "Cold equals warm" does not test
// it, because a permanently-stale cache passes that too. The test that matters is
// warm-after-a-change, in incremental_test.go.
type IncrementalSession struct {
	graph   *Graph
	program *incremental.Program
}

// NewIncrementalSession prepares a warm-capable pass. Returns nil when the graph cannot
// support one, which is a reason to run the full pass rather than an error: no compiler host,
// no config, or a tsconfig without `incremental` all mean the caller should use
// AllDiagnostics.
func (g *Graph) NewIncrementalSession() *IncrementalSession {
	if g.CompilerHost == nil || g.Config == nil {
		return nil
	}
	if !g.Config.CompilerOptions().IsIncremental() {
		return nil
	}

	// The host must be the one this graph's program was built with. Build info paths resolve
	// relative to it, so a fresh host with a different current directory resolves them
	// elsewhere and reports every file as changed. Silently, and in the direction that looks
	// like a cold cache rather than like a bug.
	host := incremental.CreateHost(g.CompilerHost)

	// A nil previous program is a cold start, not a failure: no build info on disk, a version
	// mismatch, or a non-incremental one all return nil and all mean check everything.
	reader := incremental.NewBuildInfoReader(g.CompilerHost)
	previous := incremental.ReadBuildInfoProgram(g.Config, reader, g.CompilerHost)

	incrementalProgram := incremental.NewProgram(g.Program, previous, host)
	if incrementalProgram == nil {
		return nil
	}
	return &IncrementalSession{graph: g, program: incrementalProgram}
}

// Diagnostics returns the same findings AllDiagnostics returns, re-checking only the files
// whose contents changed since the last Write.
//
// The skip is not a replay of stored diagnostics. The build info stores content hashes; on
// load, upstream seeds a "no diagnostics" entry for every file whose hash still matches, and
// the checker visits only the files missing from that map.
//
// The program is still built either way. A build info can skip checking the graph, never
// constructing it, so the caller has already paid the graph phase before calling this.
func (s *IncrementalSession) Diagnostics(ctx context.Context) []*ast.Diagnostic {
	return s.DiagnosticParts(ctx).All()
}

// DiagnosticParts is Diagnostics with its parts kept apart. See TypeDiagnosticParts.
func (s *IncrementalSession) DiagnosticParts(ctx context.Context) TypeDiagnosticParts {
	// Syntactic first, and bail on failure, for the same reason AllDiagnostics does: a file
	// that does not parse produces cascading nonsense from the later phases.
	syntactic := s.graph.Program.GetSyntacticDiagnostics(ctx, nil)
	if len(syntactic) > 0 {
		return TypeDiagnosticParts{Syntactic: syntactic}
	}
	return TypeDiagnosticParts{Bind: s.graph.Program.GetBindDiagnostics(ctx, nil), Semantic: s.program.GetSemanticDiagnostics(ctx, nil)}
}

// Write persists what the next run needs to tell changed files from unchanged ones.
//
// It costs about 146ms on ahra warm, and the cost is CPU rather than disk. Measured by
// instrumenting inside upstream's emitBuildInfo:
//
//	snapshotToBuildInfo   104ms   walking the snapshot
//	json.Marshal            9ms   3.5 MB of JSON
//	the write itself       0.4ms
//
// Writing 3.5 MB on this machine is 0.5ms, ten trials. So "the cache rewrites an unchanged file
// every run" is true and is not worth fixing: skipping the write when the bytes match was built,
// measured, and reverted, because reading the old file to compare costs more than the write it
// avoids (blocked A/B, six runs each: 362ms median unchanged against 371ms with the skip).
//
// Upstream's own already-up-to-date check cannot fire for us and that is not a bug to fix here:
// buildInfoEmitPending is initialized true for any incremental config with no in-memory
// predecessor (programtosnapshot.go:69) and set true again unconditionally after semantic
// diagnostics (program.go:319). Both are correct for the watch process it was written for. cohere
// is one-shot, so a fresh snapshot every run means the flag is structurally always true.
//
// The remaining cost lives in rebuilding state a resident process would still hold -- the 204ms
// parse in NewIncrementalSession and the 104ms snapshot walk here are the same expense paid at
// both ends. Neither is reachable without deciding whether cohere stays one-shot. See #hfv0ae3.
//
// Call it after Diagnostics, on the same session. Calling it before, or on a different
// session, produces a build info that records nothing as checked and a warm run that skips
// nothing, which costs more than not caching at all.
//
// Cohere does not otherwise emit, and this does not change that: under NoEmit the emit path
// writes the build info and no JavaScript.
//
// Returns any diagnostics from the write itself. A failed write must not pass silently: the
// next run would be cold while this one reported success, and the symptom is a saving that
// never appears rather than an error anyone can see.
func (s *IncrementalSession) Write(ctx context.Context) []*ast.Diagnostic {
	result := incremental.EmitBuildInfo(ctx, s.program)
	if result == nil {
		return nil
	}
	return result.Diagnostics
}

// IncrementalDiagnostics is the one-shot form: check warm and persist for next time, falling
// back to the full pass when the graph cannot support a session.
//
// Prefer this at a call site that just wants the findings. Use NewIncrementalSession directly
// when the caller needs to decide whether to write, for example when it is about to report a
// failure and would rather not persist state from a run it is discarding.
func (g *Graph) IncrementalDiagnostics(ctx context.Context) []*ast.Diagnostic {
	session := g.NewIncrementalSession()
	if session == nil {
		return g.AllDiagnostics(ctx)
	}
	diagnostics := session.Diagnostics(ctx)
	session.Write(ctx)
	return diagnostics
}
