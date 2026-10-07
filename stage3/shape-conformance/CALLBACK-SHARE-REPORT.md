Built: finite callback allocation flow through aliases, parameters, returns and joined implementation identities.
Commits: diagnostic handoff b8316acd; current main merged through 030119b4 (origin/main ce0750f2); feature commit is recorded in branch history.
Commands: lower/IR packages, IR target-reader guard, vet, native/Node fixtures, 43 unchanged graph counts and census audits pass.
Mutants: six callback guards, six census guards, host/body forgeries and native wrong-shape/readiness erasures caught.
Limits: no tsc cast proven free; opaque callback protocols, dynamic keys, constructors, whole-array contracts and one-lookup-per-cast emission remain unproven.

Every census number below is measured on a checker-rejected program.

| Outcome | Tagged | Untagged | Total |
|---|---:|---:|---:|
| Conforms and ready (free) | 0 | 0 | 0 |
| Conforms but not proven ready | 0 | 0 | 0 |
| Conforms-if | 0 | 0 | 0 |
| Unknown: host metadata | 0 | 3 | 3 |
| Unknown: unsupported flow | 313 | 281 | 594 |
| Unknown: diagnosed body/dependency | 1445 | 894 | 2339 |
| Total | 1758 | 1178 | 2936 |

Relative to e7f42029, exactly 335 unsupported sites move to diagnosed dependencies (292 tagged, 43 untagged), and two untagged unsupported sites move to host metadata. All other primary outcomes stay unchanged. The 315 checker diagnostics and all adapted source hashes are unchanged. Allocation schemas increase from 1779 to 2510 because arguments and callback bodies previously hidden behind opaque result boundaries are now represented. Own-function diagnoses remain 44 (9 tagged, 35 untagged); diagnosed dependencies rise from 1960 to 2295 (1436 tagged, 859 untagged). This is newly visible provenance, not new checker errors or newly free casts.

The largest of the requested patterns in the 931-site unsupported bucket was a definite function-value escape/callback boundary: 432 sites (376 tagged, 56 untagged). The separate indirect/callback/intrinsic result boundary affected 227 sites; it is a mixed boundary rather than proof every site uses a callback. Their union was 470 sites. Dynamic keys affected 53, constructors 39 and whole-array contracts 5. All pattern counts are distinct sites per row and overlap.

| Observed unsupported pattern | Before tagged | Before untagged | After tagged | After untagged |
|---|---:|---:|---:|---:|
| constructors | 0 | 39 | 0 | 36 |
| dynamic keys | 25 | 28 | 44 | 29 |
| function-value escape callback boundary | 376 | 56 | 88 | 36 |
| indirect/callback/intrinsic result boundary | 188 | 39 | 0 | 0 |
| unbounded callback identity | 0 | 0 | 5 | 19 |
| whole-array contracts | 0 | 5 | 0 | 5 |

The old blanket callback-escape boundary falls from 432 to 124 unsupported sites. Disappearing boundaries do not imply free casts: newly tracked incoming values commonly expose another diagnosed, host or unsupported producer. Dynamic-key obligations increase as deeper producers become visible. Exact per-site transitions and overlapping patterns are in callback/flow-patterns.json.

Named query helpers: prepareShapeCallbacks solves finite function identities over the existing graph sources and adds argument edges in a separate proof-only index; shapeFunctionTargets follows aliases, parameters, returns and joins; shapeClosureTargets uses the central IR ClosureTargetsWithFlow API; shapeFollow supplies result producers to allocation queries. The central ClosureValue accessor supplies provenance operands. ClosureTargetsWithFlow honors Direct sibling code targets separately from closure environments. Existing graph.follow, graph.sources and graph.unknownParameters remain unchanged for graph-region ownership. All 43 allocation/count snapshots pass unchanged. No allocation numbering system was added.

An unknown callable arm cannot disappear from a joined target set. Omitted parameters, a callable handed to an opaque operation, stored callable values and escaped factories returning another callable retain unknown. Opaque implementation frontiers carry raw host/diagnostic provenance into affected parameters. Identity-solving and frontier work budgets fail conservatively. Fully bounded callbacks can use the existing eraser, provided every reaching record field is certified and initialized and the existing mutation guards permit it. Generic instantiation, intrinsic callback protocols, callable properties and arbitrary signatures remain unknown. This proves identities and incoming values; it does not implement or trust a callable view signature.

Lane 5 was inspected at 6b80e39b and was not merged or copied. Its signature contracts are not needed for this finite-identity query. After the coordination ruling, all future shared lane dependencies will come through codex/views-integration. This work needs no additional lane code and is not blocked on lazy admission. Current main was merged once as 030119b4, with its signal oracle rerun.

Native fixtures proven-callback.a and nonconforming-callback.a exercise a named alias passed through a callback parameter. The proven fixture has zero checked casts and reads, pinned in IR and held to source Node (hello), native under sanitizers, release C and the JavaScript backend. The mixed fixture joins both object shapes, retains the view and exits 70 on ready: expected boolean, found number. Erasing its checks finishes with wrong output, caught by semantic assertions using valid release C. Existing readiness-erasure mutants also remain caught.

Six valid Go mutants ignore sibling code targets, drop a reaching function, ignore an unknown callable arm, ignore opaque escapes, ignore omitted arguments and ignore a returned factory escape; semantic query witnesses catch each. Six valid measurement mutants ignore whole-array certificates, exhaust the worklist budget, drop diagnostic provenance, ignore field types, ignore readiness and analyze diagnosed bodies; independent controls catch each. Forged host/body certification is rejected. Source mutant restores and the restored query tests pass.

The 26 measurement controls contain four free, one readiness-only, two conforms-if and nineteen unknown casts. Closed aliases and returned factories are free; a mixed callback requires the ready number-to-boolean guard; host, generic and diagnosed callback sinks stay unknown. Four earlier property/store controls now retain unknown because opaque callback calls are explicitly represented and the conservative mutation guard applies across the program. Their separate native immutable property fixtures still prove zero checks. This additional conservatism is not presented as conformance.

Full lower/IR tests pass in 32.997s/28.881s. The final IR suite, including the target-reader guard, passes in 19.091s. The final uncached native eraser/callback/count oracle passes in 10.802s; all 43 graph-region counts are unchanged and two callback count rows were added to the lane-owned test. Main's signal oracle passes in 12.827s; vet passes. No full repository gate was run. All test output went to logs. The first callback corpus run was killed with exit 137 and its partial output was discarded. Per-row compaction plus GOMEMLIMIT=3GiB and GOGC=50 completed the final audited run. These resource settings do not change the proof rules.

Count rows for the integrator follow; only the integrator should regenerate shared counts.md.

| Fixture | Allocations | Frees | Retains | Releases | Live peak | In regions | Graph regions | Graph merges |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| proven-callback.a | 2 | 2 | 6 | 8 | 2 | 0 | 0 | 0 |
| nonconforming-callback.a | 4 | 1 | 9 | 8 | 4 | 0 | 0 | 0 |

The nonconforming fixture exits on the expected panic; its counts record the three objects still owned at process termination. The successful fixture allocates and frees both objects. Runtime representation, class certificates, nested contracts, post-store readiness and dedicated one-shape-lookup-per-cast emission remain outside this change.

Host metadata obligations overlap the diagnosed bucket. The three primary host sites and every overlapping host value, target type and cast location are in callback/latent-host-values.json. Names and distinct-site counts are in callback/latent-share-summary.json; full producer details remain in callback/latent-share.json.gz. No host-sourced value is certified free.

The diagnostic handoff is refreshed against this frontier below. It still covers exactly 315 diagnostics in 79 code/file groups. “Cleared alone” remains a counterfactual over current recorded diagnosis obligations, not a claim that rechecking would emit a free cast. Locations are the unchanged adapted source at e7f42029; the graph query is this callback checkpoint.


Every number below is measured on a checker-rejected program.

The table covers all 315 diagnostics, grouped by code and file. Affected sites count casts whose own function or a dependency carries that group; these columns overlap across rows. Clearing the group unblocks those recorded dependencies. “Cleared alone” counts casts whose entire recorded diagnostic frontier belongs to this group. It is a counterfactual over current edges: rechecking edited source may expose other dependencies, and host/unsupported/readiness barriers remain. It is not an observed free-cast count.

| Code | File | Diagnostics | Own function sites | Dependency sites | Affected sites | Cleared alone |
|---|---|---:|---:|---:|---:|---:|
| TS2412 | `src/compiler/factory/nodeFactory.ts` | 2 | 0 | 2187 | 2187 | 280 |
| TS2345 | `src/compiler/checker.ts` | 30 | 16 | 1927 | 1943 | 10 |
| TS2345 | `src/compiler/transformers/generators.ts` | 11 | 0 | 1934 | 1934 | 0 |
| TS2532 | `src/compiler/transformers/generators.ts` | 5 | 0 | 1934 | 1934 | 0 |
| TS18048 | `src/compiler/checker.ts` | 27 | 7 | 1920 | 1927 | 1 |
| TS18048 | `src/compiler/transformers/es2017.ts` | 6 | 0 | 1926 | 1926 | 4 |
| TS2769 | `src/compiler/checker.ts` | 5 | 9 | 1914 | 1923 | 2 |
| TS2322 | `src/compiler/transformers/es2017.ts` | 1 | 0 | 1921 | 1921 | 0 |
| TS2532 | `src/compiler/transformers/es2017.ts` | 1 | 0 | 1921 | 1921 | 0 |
| TS2345 | `src/compiler/transformers/module/esnextAnd2015.ts` | 4 | 4 | 1917 | 1921 | 4 |
| TS18048 | `src/compiler/transformers/classFields.ts` | 2 | 0 | 1920 | 1920 | 0 |
| TS2345 | `src/compiler/transformers/classFields.ts` | 4 | 0 | 1920 | 1920 | 0 |
| TS2538 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 1920 | 1920 | 0 |
| TS2375 | `src/compiler/checker.ts` | 3 | 7 | 1912 | 1919 | 6 |
| TS2345 | `src/compiler/transformers/es2015.ts` | 1 | 0 | 1919 | 1919 | 0 |
| TS2532 | `src/compiler/transformers/es2015.ts` | 4 | 0 | 1919 | 1919 | 0 |
| TS18048 | `src/compiler/transformers/generators.ts` | 7 | 0 | 1919 | 1919 | 0 |
| TS2339 | `src/compiler/transformers/generators.ts` | 3 | 0 | 1919 | 1919 | 0 |
| TS2379 | `src/compiler/checker.ts` | 5 | 8 | 1910 | 1918 | 4 |
| TS2322 | `src/compiler/checker.ts` | 11 | 7 | 1910 | 1917 | 3 |
| TS2532 | `src/compiler/transformers/module/esnextAnd2015.ts` | 1 | 0 | 1917 | 1917 | 0 |
| TS2412 | `src/compiler/checker.ts` | 2 | 8 | 1907 | 1915 | 2 |
| TS2532 | `src/compiler/checker.ts` | 1 | 8 | 1907 | 1915 | 2 |
| TS2488 | `src/compiler/checker.ts` | 2 | 7 | 1907 | 1914 | 1 |
| TS2556 | `src/compiler/checker.ts` | 1 | 6 | 1907 | 1913 | 0 |
| TS2722 | `src/compiler/checker.ts` | 1 | 6 | 1907 | 1913 | 0 |
| TS2345 | `src/compiler/builder.ts` | 12 | 5 | 1901 | 1906 | 9 |
| TS2322 | `src/compiler/commandLineParser.ts` | 4 | 0 | 1906 | 1906 | 0 |
| TS2345 | `src/compiler/tsbuildPublic.ts` | 1 | 4 | 1900 | 1904 | 7 |
| TS2345 | `src/compiler/transformers/esnext.ts` | 6 | 0 | 1901 | 1901 | 3 |
| TS2412 | `src/compiler/tsbuildPublic.ts` | 10 | 2 | 1898 | 1900 | 3 |
| TS2412 | `src/compiler/watchPublic.ts` | 8 | 0 | 1900 | 1900 | 3 |
| TS2532 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 1899 | 1899 | 0 |
| TS2345 | `src/compiler/program.ts` | 4 | 0 | 1898 | 1898 | 0 |
| TS2375 | `src/compiler/program.ts` | 3 | 0 | 1898 | 1898 | 0 |
| TS2345 | `src/compiler/transformers/declarations.ts` | 1 | 1 | 1897 | 1898 | 1 |
| TS2345 | `src/compiler/transformers/jsx.ts` | 2 | 1 | 1897 | 1898 | 1 |
| TS2488 | `src/compiler/builder.ts` | 2 | 0 | 1897 | 1897 | 0 |
| TS18046 | `src/compiler/commandLineParser.ts` | 1 | 0 | 1897 | 1897 | 0 |
| TS2345 | `src/compiler/core.ts` | 1 | 0 | 1897 | 1897 | 0 |
| TS18046 | `src/compiler/program.ts` | 2 | 0 | 1897 | 1897 | 0 |
| TS18048 | `src/compiler/transformers/esDecorators.ts` | 2 | 0 | 1897 | 1897 | 0 |
| TS2538 | `src/compiler/transformers/esDecorators.ts` | 1 | 0 | 1897 | 1897 | 0 |
| TS18046 | `src/compiler/transformers/jsx.ts` | 2 | 0 | 1897 | 1897 | 0 |
| TS2488 | `src/compiler/transformers/jsx.ts` | 1 | 0 | 1897 | 1897 | 0 |
| TS2322 | `src/compiler/transformers/generators.ts` | 1 | 0 | 1626 | 1626 | 0 |
| TS2375 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 1093 | 1093 | 0 |
| TS2322 | `src/compiler/transformer.ts` | 4 | 0 | 1069 | 1069 | 0 |
| TS2532 | `src/compiler/transformer.ts` | 2 | 0 | 1065 | 1065 | 0 |
| TS2412 | `src/compiler/resolutionCache.ts` | 1 | 0 | 938 | 938 | 0 |
| TS2591 | `src/compiler/sys.ts` | 44 | 0 | 937 | 937 | 0 |
| TS18046 | `src/compiler/sys.ts` | 1 | 0 | 931 | 931 | 0 |
| TS18048 | `src/compiler/tsbuildPublic.ts` | 3 | 0 | 918 | 918 | 0 |
| TS2488 | `src/compiler/program.ts` | 1 | 0 | 801 | 801 | 0 |
| TS2322 | `src/compiler/moduleNameResolver.ts` | 4 | 0 | 796 | 796 | 0 |
| TS18048 | `src/compiler/resolutionCache.ts` | 2 | 0 | 736 | 736 | 0 |
| TS2322 | `src/compiler/resolutionCache.ts` | 2 | 0 | 736 | 736 | 0 |
| TS2345 | `src/compiler/resolutionCache.ts` | 3 | 0 | 736 | 736 | 0 |
| TS2375 | `src/compiler/resolutionCache.ts` | 1 | 0 | 736 | 736 | 0 |
| TS2379 | `src/compiler/resolutionCache.ts` | 3 | 0 | 736 | 736 | 0 |
| TS2375 | `src/compiler/builder.ts` | 2 | 0 | 544 | 544 | 0 |
| TS2769 | `src/compiler/builder.ts` | 1 | 0 | 544 | 544 | 0 |
| TS2345 | `src/compiler/commandLineParser.ts` | 3 | 1 | 470 | 471 | 2 |
| TS2375 | `src/compiler/commandLineParser.ts` | 1 | 0 | 469 | 469 | 0 |
| TS2345 | `src/compiler/moduleNameResolver.ts` | 3 | 0 | 130 | 130 | 0 |
| TS2375 | `src/compiler/moduleNameResolver.ts` | 3 | 0 | 130 | 130 | 0 |
| TS2345 | `src/compiler/sourcemap.ts` | 1 | 0 | 130 | 130 | 0 |
| TS2304 | `src/compiler/sys.ts` | 6 | 0 | 130 | 130 | 0 |
| TS2307 | `src/compiler/sys.ts` | 1 | 0 | 130 | 130 | 0 |
| TS2322 | `src/compiler/sys.ts` | 1 | 0 | 130 | 130 | 0 |
| TS2345 | `src/compiler/sys.ts` | 1 | 0 | 130 | 130 | 0 |
| TS7006 | `src/compiler/sys.ts` | 1 | 0 | 130 | 130 | 0 |
| TS7031 | `src/compiler/sys.ts` | 1 | 0 | 130 | 130 | 0 |
| TS2591 | `src/compiler/tracing.ts` | 4 | 0 | 129 | 129 | 0 |
| TS2420 | `src/compiler/checker.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2740 | `src/compiler/core.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2591 | `src/compiler/performanceCore.ts` | 2 | 0 | 0 | 0 | 0 |
| TS2322 | `src/compiler/transformers/esDecorators.ts` | 2 | 0 | 0 | 0 | 0 |
| TS2379 | `src/compiler/watchPublic.ts` | 2 | 0 | 0 | 0 | 0 |

Diagnostic count check: 315; diagnosed casts: 2339. Exact adapted line/column locations and messages for every diagnostic are in callback/diagnostic-actions.json. Locations refer to the source hashes at e7f42029, not a newer adaptation tree.


Commands for reproduction:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
python3 stage3/shape-conformance/latent/make-overlay.py /tmp/shape-callback-overlay
go build -buildvcs=false -overlay=/tmp/shape-callback-overlay/overlay.json -o /tmp/shape-callback-census ./stage3/shape-conformance/latent/tool
GOMEMLIMIT=3GiB GOGC=50 /tmp/shape-callback-census /tmp/shape-conformance-merged-adapted /tmp/shape-merged-mapped.json /tmp/shape-callback-result-final.json > stage3/shape-conformance/logs/callback-census-final.log 2>&1
python3 stage3/shape-conformance/latent/audit.py /tmp/shape-callback-result-final.json /tmp/shape-merged-mapped.json /tmp/shape-conformance-merged-adapted > stage3/shape-conformance/logs/callback-census-audit.log 2>&1
go test ./internal/lower ./internal/ir -count=1 > stage3/shape-conformance/logs/callback-packages-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^(TestShapeCallbackCountRows|TestCheckedViewShapeErasure|TestShapeGraphCountSnapshot)$' -count=1 -v > stage3/shape-conformance/logs/callback-oracle-final.log 2>&1
python3 stage3/shape-conformance/callback-mutants.py > stage3/shape-conformance/logs/callback-mutants-final.log 2>&1
python3 stage3/shape-conformance/latent/mutants.py /tmp/shape-callback-latent-mutants > stage3/shape-conformance/logs/callback-latent-mutants.log 2>&1
```

The adapter emits only zero-Direct CallClosure nodes; the final sibling-target API addition was verified by production unit/IR guards and its mutant and does not change these measured source edges. Initial setup timing and nproc=5 are recorded in DIAGNOSTIC-ACTIONS-REPORT.md. The original diagnostic-only handoff remains at b8316acd; it is not overwritten by this later measurement.
