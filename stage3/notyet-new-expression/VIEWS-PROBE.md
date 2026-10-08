Fetched views-integration ffe428ab and replayed all 22 roots in an isolated measurement checkout; the combined merge is blocked by conflicting assertion contracts.
Current production code is 6b0d705e91d13e99135883f07a3379089a794270; the evidence commit SHA is reported with the push.
21 exact constructor signatures reproduce; Uint16Array stops earlier at its type representation. The class-value fixture passes source Node, both backends and sanitizers in the isolated views checkout plus our constructor helper.
All six class-value mutants fail Node comparisons; restored fixture passes 0.644s with allocations/frees 72/72, retains/releases 31/114, peak 14, regions 0, graph regions/merges 0/0.
No original root is newly certified. Main branch retains the Uint16Array NotYet and the reviewed helper removal.

## Merge blocker

c41c0e062e99da37820f822968d4df1b48cdaee7 is already an ancestor of the production branch. Newest origin/codex/views-integration resolved to ffe428ab. Attempting that merge produced 19 conflicted files. More seriously, the two integrations require opposite outcomes for identical non_null_literal_statement.a text:

console.log('before');
undefined!;
console.log('after');

Our TestImpossibleNonNullFixturesAreRefused requires lower.Refused with "the non-null assertion !". Views' readiness_test.go registers literal_statement, initialized, literal_return and uninitialized_default as lowering, and requires runtime readiness behavior. Picking either policy invalidates tests from the other integration. A contract ruling was requested; it remains pending. The merge was aborted after preserving independent reviewed draft resolutions under /tmp/new-expression-views-independent-resolutions.patch and /tmp/new-expression-views-expression-draft.patch. These drafts are untested and are not applied or pushed as compiler changes. The production working tree is restored.

The bulk conflict script was rejected by automatic approval review for heuristic concatenation across shared compiler files. Small explicit patches were accepted for independent cases. No bulk script executed. Combining the integrations still requires the semantic ruling, not sandbox approval.

## Independent evidence

The detached /tmp/new-expression-views-probe checkout is exactly ffe428ab for compiler source at the time its census binary was built. Measurement tooling was selected from 9a1f14c5 in stage3/census/latent only, with the production Lower and ordinary loader disabled in the scratch Go overlay. The cohere submodule is referenced through a symlink to the existing identical pinned 7945d102 checkout; no cohere code was copied. First replay attempts failed before measurement because @types/node was absent. npm ci --ignore-scripts --no-audit --no-fund in stage3/api installed the pinned three packages, including @types/node 25.3.3, then every root was replayed again. Only the successful setup's results are reported as measurements.

The unchanged 81-file adapted corpus is /tmp/new-expression-adapted/src/tsc/tsc.ts. Exact commands, assertion exits and ordered findings are in evidence/views-probe-roots.json. Binder Symbol, checker Symbol and inspector Session now reach and reproduce their exact constructor stop, instead of being masked by __String or the outer closure. All six weak collections, all eight parenthesized constructors, SymbolLinks and the remaining ordinary allocators also reproduce. Uint16Array instead reports "a value of type Uint16Array<ArrayBuffer>" at utilities.ts:10491:11; this is not counted as lowered. This views-only compiler does not include the area's typed-array guard, which remains in our production branch. NodeLinks' body still positively reproduces "this outside a method" at checker.ts:1456:5, so ordinary constructor receiver semantics remain a separate ruling.

After building the measurement binary, our newClassValue helper and its three-line newExpression hook were added only to the isolated checkout, together with the unchanged class-value fixture and its own temporary registrar. The focused Node/native/JavaScript/release/ASan/UBSan fixture passed in 15.709s. The six original independent mutants were rerun against that fixture: eager cache, missing cache store, static allocator, constructor reread, repeated constructor and repeated argument. Five failed Node stdout comparisons; missing cache store failed Node exit comparison through the registered allocator trap. The helper was restored byte for byte, and the fixture passed again in 0.644s. Final independent counted build passed 0.205s with the unchanged six original counts plus the views table's two zero graph counters. An earlier count probe overlapped a mutant and is not used as evidence; the final count probe ran after all mutants finished.

Commands sourced /workspace/adamic-tools/env.sh and directed output to files. No full package or repository gate was run for the unmerged integration. The focused tests and counts certify only this isolated class-value probe, not the combined c41/views branch. No new production fixture or count row was added in this evidence group.

Remaining original blockers: six weak-collection sites need lifetime/ephemeron design, five ordinary constructor identifiers and seven cached ordinary constructors need receiver/initialization support, NodeLinks as any also needs a sound any policy, SymbolLinks needs the class-expression dispatcher worker, inspector needs its host profiling API adapter, and Uint16Array awaits the runtime-owned real storage integration. No stronger map substitution, empty profiler or Int32-backed Uint16 was introduced.
