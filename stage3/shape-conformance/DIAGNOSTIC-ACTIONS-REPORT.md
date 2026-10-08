Built: grouped all 315 diagnostics by code and file, with per-cast dependency and unblock counts.
Commits: measured input e7f420293e94746095c2f1e2d14124484e1afb63; this report checkpoint is recorded in branch history.
Commands: diagnostic-actions.py inventory/provenance assertions pass; 315 diagnostics in 79 groups cover 2004 diagnosed sites.
Mutants: fixing one of two diagnosed producers cannot clear both; the overlapping-dependency mutant is caught.
Limits: unblock counts describe current diagnosis frontiers, not successful rechecking or free casts.

Every number below is measured on a checker-rejected program.

The table covers all 315 diagnostics, grouped by code and file. Affected sites count casts whose own function or a dependency carries that group; these columns overlap across rows. Clearing the group unblocks those recorded dependencies. “Cleared alone” counts casts whose entire recorded diagnostic frontier belongs to this group. It is a counterfactual over current edges: rechecking edited source may expose other dependencies, and host/unsupported/readiness barriers remain. It is not an observed free-cast count.

| Code | File | Diagnostics | Own function sites | Dependency sites | Affected sites | Cleared alone |
|---|---|---:|---:|---:|---:|---:|
| TS2412 | `src/compiler/factory/nodeFactory.ts` | 2 | 0 | 1734 | 1734 | 110 |
| TS2345 | `src/compiler/checker.ts` | 30 | 16 | 1656 | 1672 | 11 |
| TS2345 | `src/compiler/transformers/es2015.ts` | 1 | 0 | 1668 | 1668 | 0 |
| TS2532 | `src/compiler/transformers/es2015.ts` | 4 | 0 | 1668 | 1668 | 0 |
| TS2345 | `src/compiler/transformers/generators.ts` | 11 | 0 | 1656 | 1656 | 0 |
| TS2532 | `src/compiler/transformers/generators.ts` | 5 | 0 | 1656 | 1656 | 0 |
| TS18048 | `src/compiler/checker.ts` | 27 | 7 | 1648 | 1655 | 1 |
| TS2769 | `src/compiler/checker.ts` | 5 | 9 | 1635 | 1644 | 1 |
| TS2375 | `src/compiler/checker.ts` | 3 | 7 | 1634 | 1641 | 6 |
| TS18048 | `src/compiler/transformers/generators.ts` | 7 | 0 | 1641 | 1641 | 0 |
| TS2339 | `src/compiler/transformers/generators.ts` | 3 | 0 | 1641 | 1641 | 0 |
| TS2322 | `src/compiler/checker.ts` | 11 | 7 | 1633 | 1640 | 3 |
| TS2379 | `src/compiler/checker.ts` | 5 | 8 | 1632 | 1640 | 4 |
| TS18048 | `src/compiler/transformers/classFields.ts` | 2 | 0 | 1638 | 1638 | 0 |
| TS2345 | `src/compiler/transformers/classFields.ts` | 4 | 0 | 1638 | 1638 | 0 |
| TS2538 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 1638 | 1638 | 0 |
| TS2412 | `src/compiler/checker.ts` | 2 | 8 | 1629 | 1637 | 2 |
| TS2532 | `src/compiler/checker.ts` | 1 | 8 | 1629 | 1637 | 2 |
| TS2488 | `src/compiler/checker.ts` | 2 | 7 | 1629 | 1636 | 1 |
| TS2556 | `src/compiler/checker.ts` | 1 | 6 | 1629 | 1635 | 0 |
| TS2722 | `src/compiler/checker.ts` | 1 | 6 | 1629 | 1635 | 0 |
| TS18048 | `src/compiler/transformers/esDecorators.ts` | 2 | 0 | 1635 | 1635 | 0 |
| TS2538 | `src/compiler/transformers/esDecorators.ts` | 1 | 0 | 1635 | 1635 | 0 |
| TS18048 | `src/compiler/transformers/es2017.ts` | 6 | 0 | 1634 | 1634 | 4 |
| TS2322 | `src/compiler/transformers/es2017.ts` | 1 | 0 | 1626 | 1626 | 0 |
| TS2532 | `src/compiler/transformers/es2017.ts` | 1 | 0 | 1626 | 1626 | 0 |
| TS2345 | `src/compiler/transformers/esnext.ts` | 6 | 0 | 1623 | 1623 | 8 |
| TS2345 | `src/compiler/transformers/jsx.ts` | 2 | 1 | 1619 | 1620 | 1 |
| TS18046 | `src/compiler/transformers/jsx.ts` | 2 | 0 | 1619 | 1619 | 0 |
| TS2488 | `src/compiler/transformers/jsx.ts` | 1 | 0 | 1619 | 1619 | 0 |
| TS2345 | `src/compiler/transformers/declarations.ts` | 1 | 1 | 1616 | 1617 | 2 |
| TS2532 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 1142 | 1142 | 0 |
| TS2322 | `src/compiler/transformer.ts` | 4 | 0 | 966 | 966 | 0 |
| TS2532 | `src/compiler/transformer.ts` | 2 | 0 | 961 | 961 | 0 |
| TS18046 | `src/compiler/commandLineParser.ts` | 1 | 0 | 617 | 617 | 0 |
| TS18046 | `src/compiler/program.ts` | 2 | 0 | 617 | 617 | 0 |
| TS2345 | `src/compiler/program.ts` | 4 | 0 | 617 | 617 | 0 |
| TS2375 | `src/compiler/program.ts` | 3 | 0 | 617 | 617 | 0 |
| TS2345 | `src/compiler/transformers/module/esnextAnd2015.ts` | 4 | 4 | 591 | 595 | 4 |
| TS2322 | `src/compiler/transformers/generators.ts` | 1 | 0 | 591 | 591 | 0 |
| TS2532 | `src/compiler/transformers/module/esnextAnd2015.ts` | 1 | 0 | 591 | 591 | 0 |
| TS2345 | `src/compiler/builder.ts` | 12 | 5 | 4 | 9 | 9 |
| TS2322 | `src/compiler/commandLineParser.ts` | 4 | 0 | 8 | 8 | 0 |
| TS2345 | `src/compiler/tsbuildPublic.ts` | 1 | 4 | 1 | 5 | 5 |
| TS2375 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 3 | 3 | 0 |
| TS2412 | `src/compiler/tsbuildPublic.ts` | 10 | 2 | 1 | 3 | 3 |
| TS2345 | `src/compiler/commandLineParser.ts` | 3 | 1 | 1 | 2 | 2 |
| TS2375 | `src/compiler/builder.ts` | 2 | 0 | 0 | 0 | 0 |
| TS2488 | `src/compiler/builder.ts` | 2 | 0 | 0 | 0 | 0 |
| TS2769 | `src/compiler/builder.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2420 | `src/compiler/checker.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2375 | `src/compiler/commandLineParser.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2345 | `src/compiler/core.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2740 | `src/compiler/core.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2322 | `src/compiler/moduleNameResolver.ts` | 4 | 0 | 0 | 0 | 0 |
| TS2345 | `src/compiler/moduleNameResolver.ts` | 3 | 0 | 0 | 0 | 0 |
| TS2375 | `src/compiler/moduleNameResolver.ts` | 3 | 0 | 0 | 0 | 0 |
| TS2591 | `src/compiler/performanceCore.ts` | 2 | 0 | 0 | 0 | 0 |
| TS2488 | `src/compiler/program.ts` | 1 | 0 | 0 | 0 | 0 |
| TS18048 | `src/compiler/resolutionCache.ts` | 2 | 0 | 0 | 0 | 0 |
| TS2322 | `src/compiler/resolutionCache.ts` | 2 | 0 | 0 | 0 | 0 |
| TS2345 | `src/compiler/resolutionCache.ts` | 3 | 0 | 0 | 0 | 0 |
| TS2375 | `src/compiler/resolutionCache.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2379 | `src/compiler/resolutionCache.ts` | 3 | 0 | 0 | 0 | 0 |
| TS2412 | `src/compiler/resolutionCache.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2345 | `src/compiler/sourcemap.ts` | 1 | 0 | 0 | 0 | 0 |
| TS18046 | `src/compiler/sys.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2304 | `src/compiler/sys.ts` | 6 | 0 | 0 | 0 | 0 |
| TS2307 | `src/compiler/sys.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2322 | `src/compiler/sys.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2345 | `src/compiler/sys.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2591 | `src/compiler/sys.ts` | 44 | 0 | 0 | 0 | 0 |
| TS7006 | `src/compiler/sys.ts` | 1 | 0 | 0 | 0 | 0 |
| TS7031 | `src/compiler/sys.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2591 | `src/compiler/tracing.ts` | 4 | 0 | 0 | 0 | 0 |
| TS2322 | `src/compiler/transformers/esDecorators.ts` | 2 | 0 | 0 | 0 | 0 |
| TS18048 | `src/compiler/tsbuildPublic.ts` | 3 | 0 | 0 | 0 | 0 |
| TS2379 | `src/compiler/watchPublic.ts` | 2 | 0 | 0 | 0 | 0 |
| TS2412 | `src/compiler/watchPublic.ts` | 8 | 0 | 0 | 0 | 0 |

Diagnostic count check: 315; diagnosed casts: 2004. Exact adapted line/column locations and messages for every diagnostic are in diagnostic-actions.json. Locations refer to the source hashes at e7f42029, not a newer adaptation tree.

| Outcome | Tagged | Untagged | Total |
|---|---:|---:|---:|
| Free | 0 | 0 | 0 |
| Readiness only | 0 | 0 | 0 |
| Conforms-if | 0 | 0 | 0 |
| Unknown: host metadata | 0 | 1 | 1 |
| Unknown: unsupported flow | 605 | 326 | 931 |
| Unknown: diagnosed body/dependency | 1153 | 851 | 2004 |
| Total | 1758 | 1178 | 2936 |

Reproduce: `python3 stage3/shape-conformance/latent/diagnostic-actions.py stage3/shape-conformance/projection/latent-share.json.gz stage3/shape-conformance/diagnostic-actions.json /tmp/shape-diagnostic-actions-table.md > stage3/shape-conformance/logs/diagnostic-actions.log 2>&1`. The inventory asserts that grouped diagnostics equal the raw checker inventory and that every cast dependency names a recorded diagnostic. No compiler option or adapted source was changed. The measurement stays at 44 own-function and 1960 dependency sites.

Toolchain setup passed in 52.727s: Node 0.029s, Go 0.029s, Markdown dependencies 0.102s, submodules 0.151s, clang 0.312s, Go build 52.495s, cache warm 52.689s. nproc=5, cgroup quota=4 CPUs. Existing full-package/oracle evidence remains at the implementation checkpoint; this report-only checkpoint reruns the diagnostic inventory and overlap control.
