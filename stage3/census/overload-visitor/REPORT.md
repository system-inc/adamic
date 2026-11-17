# Overload visitor — #ev14ch9

Base: cf468df0747075e50d326a97ff03e8530d7a1a88. Upstream: v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.

Initial checkpoint, before any adaptation build. Seven sets, one assigned latent site each. Counts below are assigned sites, not measured acceptance results. Checker-resolved caller inventory and all proof runs are pending.

| Set | Sites | Candidate road | State |
|---|---:|---|---|
| visitNode | 1 | (i) implementation node accepts undefined; then judge result | investigation |
| visitNodes | 1 | (i) correlate implementation callback input with node element; then judge result | investigation |
| visitArray | 1 | (i) correlate implementation callback input with node element; then judge result | investigation |
| visitArrayWorker | 1 | (i) correlate implementation callback input with node element; internal overload removal only if callers permit | investigation |
| visitParameterList | 1 | (i) annotate implementation nodesVisitor as NodesVisitor | investigation |
| visitEachChild overload 2 | 1 | (i) annotate implementation nodesVisitor as NodesVisitor | investigation |
| getParseTreeNode overload 2 | 1 | (iii) likely stop: generic result versus Node from original chain; public overload changes forbidden | investigation |

No owner is already pinned whole: searched every owners.json, targets.json, and adapt.cjs for visitorPublic.ts, utilitiesPublic.ts and the owner names. Existing matches own imports, other declarations or localized rewrites. No adaptation has been built or landed. No public overload has been edited.

Setup: fresh detached worktree created at the exact base; recursive shallow submodules initialized; stage3/api npm ci passed; independent upstream shallow tag clone confirmed the specified commit. The starting shared checkout was older and was left unchanged except for fetching the required base and worktree metadata.

Proof status for every set: scan-all NOT RUN; full latent owner/callers NOT RUN; restoring mutant NOT RUN; stock 79-root diagnostics NOT RUN; adamic types NOT RUN; apply NOT RUN; public dts NOT RUN; TestProduct_Stage3AdaptedTree NOT RUN; whole lane NOT RUN. Landing branches and SHAs: none.
