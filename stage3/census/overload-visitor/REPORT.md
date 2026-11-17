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

## Resolved-call inventory

Stock TypeScript 6.0.3 parsed 39,402 checkout source files, selected 46 files with candidate call expressions by AST, and checked 603 roots (all non-declaration src sources plus candidates). Only calls whose getResolvedSignature declaration is the named function in compiler/visitorPublic.ts or compiler/utilitiesPublic.ts count. Text matches never establish caller identity. Raw rows record every call, argument, argument type, result type, resolved declaration and immediate consumer. Test and baseline files were included in discovery; none of their candidates resolved to these source declarations. Harness calls are included. The first attempt queried all call signatures and overflowed the checker stack; the retained successful retry queries only AST candidates.

| Set | Calls | Resolved declaration lines | Result consumers |
|---|---:|---|---|
| visitNode | 433 | {122: 429, 142: 4} | {'ConditionalExpression': 11, 'discarded': 5, 'passed as argument': 271, 'binary expression': 20, 'variable initializer': 86, 'ArrayLiteralExpression': 5, 'returned': 16, 'NonNullExpression': 9, 'ArrowFunction': 10} |
| visitNodes | 187 | {216: 3, 194: 184} | {'returned': 4, 'passed as argument': 105, 'ConditionalExpression': 17, 'variable initializer': 48, 'SpreadElement': 2, 'ArrowFunction': 1, 'binary expression': 9, 'PropertyAccessExpression': 1} |
| visitArray | 5 | {274: 5} | {'returned': 2, 'passed as argument': 2, 'SpreadElement': 1} |
| visitArrayWorker | 2 | {321: 2} | {'variable initializer': 1, 'returned': 1} |
| visitParameterList | 42 | {393: 40, 394: 2} | {'passed as argument': 25, 'variable initializer': 5, 'ConditionalExpression': 9, 'returned': 2, 'binary expression': 1} |
| visitEachChild | 229 | {587: 2, 585: 226, 595: 1} | {'variable initializer': 19, 'returned': 171, 'ConditionalExpression': 4, 'binary expression': 17, 'passed as argument': 11, 'ArrowFunction': 7} |
| getParseTreeNode | 86 | {817: 45, 826: 41} | {'variable initializer': 60, 'passed as argument': 16, 'binary expression': 3, 'NonNullExpression': 2, 'AsExpression': 4, 'PrefixUnaryExpression': 1} |

visitParameterList arguments: 35 omit nodesVisitor, 7 pass the existing nodesVisitor parameter. visitEachChild: 227 omit nodesVisitor, one passes nodesClone, one passes assignPositionsToNodeArray. visitArrayWorker: both calls pass visitor and the optional boolean test from visitNodes / visitArray; neither resolves to its predicate overload.

The baseline adapted-tree build and the two measurement tool builds are in progress. These inventory counts do not establish census acceptance.
