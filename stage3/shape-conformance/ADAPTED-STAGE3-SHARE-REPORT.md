Built: census of the exact output of area/stage3 234ab1aa, with checked-view integration f1c91970 merged.
Commits: integration merge 27ba7d7e; this census checkpoint is recorded in branch history.
Commands: exact apply.sh run, immutable-site mapping, independent audit, control audit, diagnostic handoff, owned unit tests and vet pass.
Mutants: dropped/duplicated cast identities, overlapping diagnostic obligations, wrong-shape erasure and readiness erasure are caught.
Limits: Adamic still rejects this adapted program with 260 diagnostics; zero free casts; eight graph-region aborts reproduce on the integration baseline.

Input: `bash /tmp/shape-stage3-234ab1aa/stage3/apply.sh /tmp/shape-stage3-234ab1aa-adapted`, in a detached checkout of origin/area/stage3 **234ab1aa5f728a5221fb6075c35b94f88a2c6437**. This uses every adaptation on that tip, including 10, 20 and 75. Neither parser nor scanner proof branches were merged. The compiler includes the designated integrator **f1c919701ec9b42b78e104fe9cc1ce96db9d7d09** through merge 27ba7d7e. The apply script and adaptation files were run from the isolated area checkout because the integrator still has older stage3 adaptations.

Every conformance and diagnostic number below is **measured on a checker-rejected program** under Adamic checker policy. The reported 106,366 Linux oracle passes are upstream evidence supplied by typescript; they were not rerun here and do not establish that this whole compiler checks under Adamic.

The completed tree contains **4,129 non-generated assertions**, net +28 versus the original 4,101; all are `as` expressions. It also contains 2,130 assertions in generated files, for 6,259 total syntax assertions. These totals include const assertions and upcasts. All **2,936 original downcasts map exactly**, preserving 1,758 tagged and 1,178 untagged ledger identities. No original site was lost or guessed from line numbers. This census measures that unchanged downcast denominator; newly introduced syntax is inventoried, not silently counted as a downcast. Exact final input hashes and rows are in [cast-inventory.json](adapted-stage3/cast-inventory.json).

| Outcome | Tagged | Untagged | Total |
|---|---:|---:|---:|
| Free | 0 | 0 | 0 |
| Readiness-only | 0 | 0 | 0 |
| Conforms-if | 0 | 0 | 0 |
| Unknown: host metadata | 0 | 3 | 3 |
| Unknown: unsupported flow | 585 | 288 | 873 |
| Unknown: diagnosed body or dependency | 1173 | 887 | 2060 |
| Total | 1758 | 1178 | 2936 |

Compared with callback checkpoint 08ceff92, diagnostics fall **315 to 260**, diagnosed casts **2339 to 2060**, and unsupported casts increase **594 to 873** as previously diagnosed paths become inspectable. Primary host sites remain three. Own-function diagnoses remain 44 (9 tagged, 35 untagged); dependency diagnoses fall 2295 to 2016 (1164 tagged, 852 untagged). Allocation schemas increase 2510 to 2522. These are input/compiler changes together, not an isolated adaptation effect.

TS1484 remains zero. TS2412 is **25**, compared with 23 in the earlier 315-diagnostic measurement: nodeFactory TS2412 is cleared, while watchPublic and program account for newly visible rows. The table below is authoritative for this exact input and compiler policy; clearing an adaptation under stock TypeScript is not assumed to clear Adamic diagnostics.

| Diagnostic | Previous | Current |
|---|---:|---:|
| TS18046 | 6 | 6 |
| TS18048 | 49 | 49 |
| TS2304 | 6 | 0 |
| TS2307 | 1 | 1 |
| TS2322 | 30 | 30 |
| TS2339 | 3 | 3 |
| TS2345 | 88 | 87 |
| TS2375 | 14 | 16 |
| TS2379 | 10 | 10 |
| TS2412 | 23 | 25 |
| TS2420 | 1 | 1 |
| TS2488 | 6 | 6 |
| TS2532 | 15 | 15 |
| TS2538 | 2 | 2 |
| TS2556 | 1 | 1 |
| TS2591 | 50 | 0 |
| TS2722 | 1 | 1 |
| TS2740 | 1 | 1 |
| TS2769 | 6 | 6 |
| TS7006 | 1 | 0 |
| TS7031 | 1 | 0 |

Primary host values remain `host.getPackageJsonInfoCache?.()?.getPackageJsonInfo` at the PackageJsonInfo cast, and results of `host.resolveModuleNames!(...).map` at two ResolvedModuleFull casts. Both stay unknown. Overlapping host frontiers, including `compilerHost.getSourceFile`, are named with target types and cast locations in [latent-host-values.json](adapted-stage3/latent-host-values.json); distinct-site counts per host operation are in [latent-share-summary.json](adapted-stage3/latent-share-summary.json).

The largest concrete unsupported boundary is now **generic call substitution: 455 sites (379 tagged, 76 untagged)**. Missing producers affect 498 but combine unrelated sources. Callback escape affects 376, dynamic keys 75, constructors 36 and whole-array contracts 5. These counts overlap. The next flow change will track allocation identities across clean, statically resolved generic function bodies without using generic types as construction certificates.

Validation: owned shape/flow and IR checks pass (0.720s and 0.015s), vet passes, 26 analysis controls pass (four free, one readiness-only, two conforms-if, nineteen unknown). Node/native/JS eraser fixtures and wrong-shape/readiness erasure mutants pass. The independent syntax inventory rejects dropped and duplicated ledger identities. Diagnostic-action controls reject the mutant that clears a site after fixing only one of several dependency obligations.

The 43-row graph-region guard fails on eight fixtures with `free(): invalid pointer`: cache, entries, literals, regression_06, regression_07, regression_08, regression_09 and symbols. Replacing all four production lane 3 files with the published integration versions reproduces the same eight failures and identical other 35 count rows. This is an observed integration baseline regression, not a green guard. Evidence is in [graph-baseline-comparison.json](adapted-stage3/graph-baseline-comparison.json) and logs/adapted-integration-baseline-graph.log. No runtime fix or changed guard counts were made in this lane. No full repository gate was run.

The first census attempt stopped before analysis because the integrated loader requires @types/node 25.3.3 in stage3/api/node_modules. `npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund` installed the declared pin (three packages, 495ms); the retry completed and was independently audited. Toolchain setup timing was already recorded in DIAGNOSTIC-ACTIONS-REPORT.md; nproc remains 5.

The remaining diagnostic handoff follows. All 260 diagnostics are grouped into 73 code/file rows with exact adapted locations/messages in [diagnostic-actions.json](adapted-stage3/diagnostic-actions.json). Rows overlap; affected casts are dependency obligations, and cleared-alone is not an observed free-cast result.

Every number below is measured on a checker-rejected program.

The table covers all 260 diagnostics, grouped by code and file. Affected sites count casts whose own function or a dependency carries that group; these columns overlap across rows. Clearing the group unblocks those recorded dependencies. “Cleared alone” counts casts whose entire recorded diagnostic frontier belongs to this group. It is a counterfactual over current edges: rechecking edited source may expose other dependencies, and host/unsupported/readiness barriers remain. It is not an observed free-cast count.

| Code | File | Diagnostics | Own function sites | Dependency sites | Affected sites | Cleared alone |
|---|---|---:|---:|---:|---:|---:|
| TS2345 | `src/compiler/checker.ts` | 30 | 16 | 1927 | 1943 | 10 |
| TS2345 | `src/compiler/transformers/generators.ts` | 11 | 0 | 1934 | 1934 | 0 |
| TS2532 | `src/compiler/transformers/generators.ts` | 5 | 0 | 1934 | 1934 | 0 |
| TS18048 | `src/compiler/checker.ts` | 27 | 7 | 1920 | 1927 | 1 |
| TS18048 | `src/compiler/transformers/es2017.ts` | 6 | 0 | 1926 | 1926 | 5 |
| TS2769 | `src/compiler/checker.ts` | 5 | 9 | 1914 | 1923 | 2 |
| TS2322 | `src/compiler/transformers/es2017.ts` | 1 | 0 | 1921 | 1921 | 0 |
| TS2532 | `src/compiler/transformers/es2017.ts` | 1 | 0 | 1921 | 1921 | 0 |
| TS2345 | `src/compiler/transformers/module/esnextAnd2015.ts` | 4 | 4 | 1917 | 1921 | 4 |
| TS18048 | `src/compiler/transformers/classFields.ts` | 2 | 0 | 1920 | 1920 | 0 |
| TS2345 | `src/compiler/transformers/classFields.ts` | 4 | 0 | 1920 | 1920 | 0 |
| TS2532 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 1920 | 1920 | 0 |
| TS2538 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 1920 | 1920 | 0 |
| TS2375 | `src/compiler/checker.ts` | 2 | 7 | 1912 | 1919 | 6 |
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
| TS2322 | `src/compiler/commandLineParser.ts` | 4 | 0 | 1906 | 1906 | 8 |
| TS2345 | `src/compiler/tsbuildPublic.ts` | 1 | 4 | 1900 | 1904 | 7 |
| TS2345 | `src/compiler/transformers/esnext.ts` | 6 | 0 | 1901 | 1901 | 4 |
| TS2412 | `src/compiler/tsbuildPublic.ts` | 10 | 2 | 1898 | 1900 | 3 |
| TS2412 | `src/compiler/watchPublic.ts` | 10 | 0 | 1900 | 1900 | 3 |
| TS2345 | `src/compiler/program.ts` | 4 | 0 | 1898 | 1898 | 0 |
| TS2375 | `src/compiler/program.ts` | 4 | 0 | 1898 | 1898 | 0 |
| TS2345 | `src/compiler/transformers/declarations.ts` | 1 | 1 | 1897 | 1898 | 1 |
| TS2345 | `src/compiler/transformers/jsx.ts` | 2 | 1 | 1897 | 1898 | 1 |
| TS2488 | `src/compiler/builder.ts` | 2 | 0 | 1897 | 1897 | 0 |
| TS18046 | `src/compiler/commandLineParser.ts` | 1 | 0 | 1897 | 1897 | 0 |
| TS18046 | `src/compiler/program.ts` | 2 | 0 | 1897 | 1897 | 0 |
| TS18048 | `src/compiler/transformers/esDecorators.ts` | 2 | 0 | 1897 | 1897 | 0 |
| TS2538 | `src/compiler/transformers/esDecorators.ts` | 1 | 0 | 1897 | 1897 | 0 |
| TS18046 | `src/compiler/transformers/jsx.ts` | 2 | 0 | 1897 | 1897 | 0 |
| TS2488 | `src/compiler/transformers/jsx.ts` | 1 | 0 | 1897 | 1897 | 0 |
| TS2322 | `src/compiler/transformers/generators.ts` | 1 | 0 | 1596 | 1596 | 0 |
| TS2375 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 1119 | 1119 | 0 |
| TS2322 | `src/compiler/transformer.ts` | 4 | 0 | 1090 | 1090 | 0 |
| TS2532 | `src/compiler/transformer.ts` | 2 | 0 | 1077 | 1077 | 0 |
| TS2412 | `src/compiler/resolutionCache.ts` | 1 | 0 | 950 | 950 | 0 |
| TS2375 | `src/compiler/watch.ts` | 2 | 0 | 950 | 950 | 1 |
| TS2307 | `src/compiler/sys.ts` | 1 | 0 | 943 | 943 | 0 |
| TS2322 | `src/compiler/sys.ts` | 1 | 0 | 943 | 943 | 0 |
| TS2345 | `src/compiler/sys.ts` | 1 | 0 | 943 | 943 | 0 |
| TS18046 | `src/compiler/sys.ts` | 1 | 0 | 934 | 934 | 0 |
| TS18048 | `src/compiler/tsbuildPublic.ts` | 3 | 0 | 903 | 903 | 0 |
| TS2412 | `src/compiler/program.ts` | 2 | 0 | 808 | 808 | 0 |
| TS2488 | `src/compiler/program.ts` | 1 | 0 | 808 | 808 | 0 |
| TS2322 | `src/compiler/moduleNameResolver.ts` | 4 | 0 | 806 | 806 | 0 |
| TS18048 | `src/compiler/resolutionCache.ts` | 2 | 0 | 731 | 731 | 0 |
| TS2322 | `src/compiler/resolutionCache.ts` | 2 | 0 | 731 | 731 | 0 |
| TS2345 | `src/compiler/resolutionCache.ts` | 3 | 0 | 731 | 731 | 0 |
| TS2375 | `src/compiler/resolutionCache.ts` | 1 | 0 | 731 | 731 | 0 |
| TS2379 | `src/compiler/resolutionCache.ts` | 3 | 0 | 731 | 731 | 0 |
| TS2375 | `src/compiler/builder.ts` | 2 | 0 | 530 | 530 | 0 |
| TS2769 | `src/compiler/builder.ts` | 1 | 0 | 530 | 530 | 0 |
| TS2345 | `src/compiler/commandLineParser.ts` | 3 | 1 | 469 | 470 | 2 |
| TS2375 | `src/compiler/commandLineParser.ts` | 1 | 0 | 468 | 468 | 0 |
| TS2345 | `src/compiler/moduleNameResolver.ts` | 3 | 0 | 385 | 385 | 0 |
| TS2375 | `src/compiler/moduleNameResolver.ts` | 3 | 0 | 385 | 385 | 0 |
| TS2345 | `src/compiler/sourcemap.ts` | 1 | 0 | 385 | 385 | 0 |
| TS2420 | `src/compiler/checker.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2740 | `src/compiler/core.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2322 | `src/compiler/transformers/esDecorators.ts` | 2 | 0 | 0 | 0 | 0 |
| TS2379 | `src/compiler/watchPublic.ts` | 2 | 0 | 0 | 0 | 0 |

Diagnostic count check: 260; diagnosed casts: 2060. Exact adapted line/column locations and messages for every diagnostic are in diagnostic-actions.json. Locations refer to the adapted tree recorded in this census source hashes.
