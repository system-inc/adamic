# Item 137: tuple length object view

The original p70 source is valid TypeScript: Node prints `2\n` and exits 0. Lowering now returns NotYet at `main.a:9:15` before emission: a tuple through an object view with a length field is unsupported; expose a scalar length field on the view and copy `pair.length` into that field. The review witness moved from pending agreement to active refusal.

The guard uses the demanded field contract and projected reaching allocations. It applies to a known real tuple under an object contract containing length, excluding fixed tuple contracts. It does not reject ordinary tuple length reads. This is the conservative authorized refusal, not implementation of tuple/object representation interoperability. Native tuples lack an object length slot, and JavaScript distinguishes array storage from object storage. A plain nested object substitution also failed the existing backend agreement; that is not presented as a workaround. Destructuring the original p70 already returns a separate located NotYet for the binding representation before this guard; that separate limitation was not changed.

The verified workaround copies tuple length into a scalar view field; Node, generated JavaScript, native release, and sanitized native agree, with no leak report. New lower leaves: TestTupleObjectViewRefused 0.09s, TestTupleObjectViewCopiedLength 0.33s, TestTupleObjectViewOrdinaryTuple 0.16s. Oracle workaround leaf 0.77s and active review refusal 0.10s.

Revert mutant p70-revert.diff removes only the new demand guard. TestTupleObjectViewRefused fails in 0.11s: expected a located NotYet, got nil. Source restored before all green checks.

Commands (all test output captured in sibling p70-*.log files):
- timeout 120 go test ./internal/lower -run '^TestTupleObjectView' -count=1 -v -timeout 90s: PASS 0.338s.
- timeout 240 go test ./internal/lower -count=1 -timeout 180s: PASS 73.577s.
- timeout 120 go test ./internal/ir -run TestCallTargetReaders -count=1 -timeout 90s: PASS 35.004s.
- timeout 120 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/lower/testdata/tuple_object_views/p70_copy.a|TestReviewProgramsRefuse/fxspptb_oct9_views_p70|TestReviewProgramsNoLooseFiles' -count=1 -v -timeout 90s: PASS 0.869s.
- timeout 300 go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 4m -args -update-counts: PASS 59.231s; counts.md refreshed.

Candidate delivery completed first: compiler/fx6-candidates at 9aff8dbf25a0dd99f8e71d0d636bf835912c4a9d. It merges receiver-keyed registration and element-key refusal, with full lower, reader guard, counts, lane, receiver and destructuring revert mutants, and element refusal revert (including native wrong-answer witness). Its evidence lives under review/compiler/fx6-candidates/ on that branch.

This advances the unit's valid-view roadmap work by replacing p70's backend compiler stop with a located, actionable refusal. Full tuple/object view support remains outstanding. No PR opened.

Lane checks: PASS 1.5s, gofmt/tools on 10 Go files, t.Parallel on 2 test packages, vet 2 packages. Initial invocation lacked the sourced toolchain and failed to locate gofmt; rerunning after source /workspace/adamic-tools/env.sh passed.
