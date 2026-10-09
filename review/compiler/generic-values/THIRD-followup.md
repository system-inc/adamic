Built a higher-rank diagnostic priority fix toward step 16: known free binders take precedence over an incomplete deep-type scan.
Commit: this commit, following d34d7634.
Checks: full internal/lower passes 35.006s; focused higher-rank tests pass 0.097s; Node refusal/alias-escape oracles pass 0.596s; reader guard and explicit go vet pass (logs included).
Mutant: restore the old traversal-limit priority; TestHigherRankBindersBeforeTraversalLimit fails in 0.038s because run's U and V are hidden by the traversal diagnostic.
Uncovered: finite traversal still refuses an incomplete graph with no known higher-rank signature; final census and admission evidence follow against this corrected revision.

The real census found a state-machine value whose own signature binders were already collected before traversal of its recursive AST-node parameter hit the depth limit. The diagnostic discarded those binders. Prefer the known higher-rank refusal; the new planted parameter combines U/V with an expanding Grow<readonly T[]> graph and requires both binder names and the ruled fix. The new leaf passes in 0.09s. The original finite-graph refusal still covers scans without any known higher-rank signature, preserving the growing generic class test.

Commands: go test ./internal/lower -timeout 90s; go test ./internal/lower -run '^TestHigherRank' -timeout 90s -v; go test ./internal/oracle -run '^(TestGenericValueHigherRankNodeControl|TestGenericValueAliasEscapeMutant)$' -timeout 90s -v; go test ./internal/ir -run '^TestCallTargetReaders$' -timeout 90s; go vet ./internal/lower. The same new leaf with the recorded old-source overlay is required to fail.
