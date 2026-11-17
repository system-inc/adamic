# Overload visitor — #ev14ch9

**Final decision: stop (iii) on all seven sets. No landing branches, adaptation directories, ledger changes or landing commits.** Six JS-identical implementation trials were measured; none meets the complete census gate. The assigned sites are not reached by scan-all's first-declaration-stop probe. This is a stop report under the brief, not a claim that any set was accepted.

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

## Rule review and candidate selection

Read census_small.go, overload_predicate_result.go, census_overload_proof.go and overload_result_check.go in full at the base. Also inspected overload_callback_use.go. censusOverload compares incoming parameters contravariantly and results covariantly under rigid overload binders. A callback forwarded to another function is outside overloadCallbackUse; an indexed loop local is not the supported const for-of argument witness. A generic implementation cannot use narrowOverload; array promises cannot use it even for a nongeneric implementation. The return proof rejects returned local bindings, reached casts, parameter writes and unsupported control flow.

- visitNode: node is Node in the implementation but the overload admits undefined. Widening it is JS-identical, but cannot by itself settle the generic visitor and result contracts. The returned visitedNode local is not body-proof evidence. 429 callers use its predicate overload, so dropping that public declaration is both disallowed and used.
- visitNodes / visitArray: Visitor<TIn> is not served by Visitor<Node>; visitor is forwarded to visitArrayWorker, outside the callback-use proof. Retaining the array predicate promises while merely fixing the input parameter can expose the independently narrower result. 184 / 5 callers use the predicate overloads. Public overload edits are a stop.
- visitArrayWorker: Visitor<TIn> is not served by Visitor<Node> (with optional undefined in the worker). Both resolved callers use the broad-result overload. Candidate (i): remove both internal overload declarations, retaining the implementation and every runtime body byte. This removes unused generic promises rather than asserting an unproved result.
- visitParameterList: inferred typeof visitNodes and the overload's NodesVisitor expose a signature relation mismatch. Candidate (i): nodesVisitor: NodesVisitor = visitNodes on the implementation only. Result differs only in presence, already handled by censusNullableOverloadResult.
- visitEachChild overload 2: same inferred typeof visitNodes versus NodesVisitor mismatch. Candidate (i): nodesVisitor: NodesVisitor = visitNodes on the implementation only. Keep all public overloads.
- getParseTreeNode overload 2: Node | undefined does not serve rigid T | undefined. The body reassigns node while traversing original and may return node without nodeTest; it is outside both return proof and predicate-return proof. Merely annotating its result T would assert what the body does not establish. Candidate (iii): stop; changing the public overload to promise Node | undefined is a declaration change, and a ruling about original-node subtype preservation would be needed for a different proof road.

These are candidates, not acceptance claims. No adaptation will be pushed without the required zero-new/zero-changed census evidence.

## Failed owner-census gates (checkpoint)

The baseline full visitorPublic census completed: 41 units, 95 raw findings, no measurement-checker diagnostics. All runs use LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE at the owner file and CommonJSBundle. Raw files and full new/gone records are retained, including recovery Boundary events separately. compare-full.py normalizes tree roots, exact unchanged source lines, and matching characters on edited signature lines so line/column movement cannot count as a semantic change.

| Trial | New non-Boundary findings | Gone non-Boundary findings | Gate |
|---|---:|---:|---|
| visitNodes correlated implementation | 2 | 1 | FAIL: narrower array-result refusal and generic function-value refusal |
| visitArrayWorker remove internal overloads | 3 | 1 | FAIL: newly reached union-member checks in body |
| visitParameterList implementation annotation | 1 | 1 | FAIL: generic instantiation at default value |
| visitEachChild implementation annotation | 1 | 0 | FAIL: generic instantiation at default value; assigned overload 2 not visible in this owner's full census |

Each full-diff.json lists every new and gone finding in full. A removed overload refusal that exposes a new failure is not accepted. No adaptation directory or landing branch has been created. The remaining trial measurements and scan-all runs are still in progress. Full caller-file runs, restoring mutants, candidate declarations, product builds and lanes have not been run because no candidate has passed its owner gate.

## Final per-set disposition

Every row has one assigned site. All landing branches and landing SHAs are **none**. Evidence is on `typescript/overload-visitor-evidence` only.

| Set | Checker-resolved callers | Road | Concrete reason to stop |
|---|---:|---|---|
| visitNode | 433 | (iii), after rejected (i) | Widening implementation node removes the parameter refusal but creates a result refusal: TOut plus admitted undefined cannot be served by Node or undefined. |
| visitNodes | 187 | (iii), after rejected (i) | Correlating implementation callback input creates the array-result refusal and a generic-function-value refusal. |
| visitArray | 5 | (iii), after rejected (i) | Owner-only census defers the new generic implementation and appears to lose one site; a resolved caller produces the narrower-array-result refusal. |
| visitArrayWorker | 2 | (iii), after rejected (i) | Removing unused internal overloads exposes three new union-member check failures at original adapted lines 355, 361 and 362. |
| visitParameterList | 42 | (iii), after rejected (i) | NodesVisitor annotation removes the assigned refusal but introduces an unreadable generic instantiation at the default visitNodes value. |
| visitEachChild overload 2 | 229 total, 2 resolving to overload 2 | (iii), after rejected (i) | Same new generic-instantiation default failure. Its assigned refusal is visible in the resolved caller census, not its deferred owner census. |
| getParseTreeNode overload 2 | 86 total, 41 resolving to overload 2 | (iii) | Full utility census confirms T or undefined versus Node or undefined; the original chain can change kind and is returned by identity without a test. A truthful public overload edit would change declarations. |

No public overload was edited, so there is no public declaration diff to land. The actual-source witness uses synthetic nodes to demonstrate the implementation contract; it does not claim those synthetic values occurred in the lane.

Candidate (ii) roads were not built. The array helpers would need a proof of element predicates and original-array alias preservation beyond this return proof; copying the original array is forbidden by the brief. The worker needs a way to serve newly reached array/object union checks without a new runtime shape. The default-callback helpers could be investigated with explicit undefined dispatch instead of materializing a generic default, but this changes JavaScript and needs the full oracle/witness/mutant proofs. Even such a candidate cannot satisfy the requested scan-all removal proof while that probe stops earlier. Those are candidate roads for a ruling or a follow-up, not accepted adaptations.

## Complete finding deltas

The six `*-full-diff.json` files contain every new/gone finding in full, including Boundary events; the paired `visitArray-caller-full-diff.json` contains every caller delta. A parameter refusal replaced by a result refusal is retained as one gone and one new event, never counted as success. There are no accepted changes for which a zero-changed claim is made.

| Trial | Owner new / gone, excluding Boundary | Additional caller result |
|---|---|---|
| visitNode | 1 / 1 | not run after failed owner gate |
| visitNodes | 2 / 1 | not run after failed owner gate |
| visitArray | 0 / 1, generic implementation deferred | esnextAnd2015.ts: 1 / 1; new array-result refusal |
| visitArrayWorker | 3 / 1 | both callers are in the measured owner file |
| visitParameterList | 1 / 1 | other callers not run after failed owner gate |
| visitEachChild | 1 / 0; assigned generic overload deferred | original assigned refusal seen in paired visitArray caller census; candidate caller sweep not run |

## Scan-all visibility limit

The baseline visitorPublic root run completed its process with exit 0 in 924.98 s and recorded 5,070 events. Its traversal reports Completed=false and `production abort at lowering.declareModule`. Its first declaration stop is core.ts:261:1:

> overload 1 of filter result U[] cannot be served by implementation result readonly T[] | undefined

No assigned overload event appears. Candidate worker scan: 0 new / 0 gone events; parameter-list scan: 1 new / 0 gone; each-child scan: 1 new / 0 gone. All stop at the same earlier filter declaration. Raw streams and normalized event-count diffs are retained in `scan-diffs.json`. Zero new/gone in the worker scan does **not** prove its assigned site gone: that site was unobserved on both sides. This is an independent reason no landing is authorized by the brief's proof gate.

The redundant utility scan was explicitly stopped after 102.51 s once all six trials had failed and the full utility result was confirmed. Its child status is -15 (SIGTERM); Python's wrapper process returned 241. Both are recorded, as is the reason in utility-scan-stop.json. It is not a completed scan or a passing proof. No child failure is reported as zero.

## Proof-line status

| Required line | Status |
|---|---|
| scan-all assigned site gone, 0 new / 0 changed | **NOT PASSED**: assigned declarations censored by earlier filter stop; two measured candidates also add scan events. Utility scan stopped; other rejected-candidate scans not run. |
| Full single-file latent owner census | **RAN** for all six trials and both baseline owner files; rejects the candidates as detailed above. CommonJSBundle and both guards enabled. |
| Full latent every caller file | **NOT RUN as a complete sweep** after failed gates; paired esnextAnd2015.ts caller run completed and rejects visitArray. Worker has only same-file callers, already included. |
| Restoring mutant | **NOT RUN**; no candidate passed its gate and no adapter built. |
| Stock tsc 6.0.3, 79 roots | **BASELINE PASSED**, 0 diagnostics. Candidate comparison **NOT RUN**. First harness attempt lacked Node types and failed with 65 diagnostics; retained as stock-missing-node-types.json. Corrected harness uses pinned stage3/api @types/node via typeRoots, without editing any source tree. |
| adamic types | **NOT RUN** for rejected trials. |
| stage3/apply.sh | **BASELINE PASSED**, fresh full build. Candidate product builds **NOT RUN**. |
| Public declarations, npm ci then hereby dts | **UPSTREAM BASELINE PASSED**. Candidate builds / byte comparisons **NOT RUN**. Upstream SHA256: typescript.d.ts c967e03c8dbb4770f7e2f0b71b5d13593de679a41cc2b60266c4b69f1342a263; tsserverlibrary.d.ts cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd. |
| TestProduct_Stage3AdaptedTree | **NOT RUN**; no accepted candidate built. Scan-all measurement product test passed; it is not this product test. |
| stage3/lane whole, unfiltered | **NOT RUN**; no accepted candidate built. No lane/counter/value-equivalence claim. |
| Ledger | **NOT WRITTEN**; no landing adaptation. |
| Type-only JavaScript identity | **PASSED for all six trials**, comments kept and removed. Exact trial patches, SHA256s, script and emit byte counts retained. |

The final tracked-file audit parsed 39,469 tracked source files across ts/tsx/js/jsx/mts/cts/mjs/cjs, selected 47 AST candidate files and checked 604 roots. It reproduced every caller count and consumer/argument class of the initial inventory. Generated built outputs and node_modules are excluded by git ls-files. This audit supersedes the initial narrower extension discovery.

The evidence checkout began at the specified base and contains only report/evidence changes. The original shared checkout and all production lowering files remain unedited. No evidence is on a landing branch; none was pushed. No work was submitted to Loom.
