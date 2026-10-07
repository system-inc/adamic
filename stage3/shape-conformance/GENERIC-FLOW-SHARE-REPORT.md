Built: allocation identity flow across clean direct generic calls, with conservative rest/spread callee frontiers and production erasure witnesses.
Commits: exact adapted-tree handoff e6676ba2; integration merge 27ba7d7e (f1c91970); generic checkpoint is recorded in branch history.
Commands: independent 2936-site audit, unchanged-source/diagnostic comparison, 35 controls, owned lower/IR checks, vet and Node/native/JS fixtures pass.
Mutants: four valid measurement binaries and wrong-shape/readiness runtime erasures caught by semantic assertions.
Limits: zero free tsc casts; remaining 260 diagnostics; eight unchanged integration graph aborts; generic callable protocols and broader field certificates remain unknown.

Every corpus count below is **measured on a checker-rejected program**. The input is exactly the output of apply.sh on area/stage3 234ab1aa, as documented in [ADAPTED-STAGE3-SHARE-REPORT.md](ADAPTED-STAGE3-SHARE-REPORT.md). The compiler contains integration f1c91970. No peer lane branch was merged. Original 2936 ledger sites all map exactly; the full input has 4129 non-generated assertions. The input hashes, 260 raw diagnostics and 2522 allocation schemas are unchanged by this flow change.

| Outcome | Before tagged | Before untagged | After tagged | After untagged | After total |
|---|---:|---:|---:|---:|---:|
| Free | 0 | 0 | 0 | 0 | 0 |
| Readiness-only | 0 | 0 | 0 | 0 | 0 |
| Conforms-if | 0 | 0 | 0 | 0 | 0 |
| Unknown: host metadata | 0 | 3 | 0 | 3 | 3 |
| Unknown: unsupported flow | 585 | 288 | 258 | 248 | 506 |
| Unknown: diagnosed body/dependency | 1173 | 887 | 1500 | 927 | 2427 |
| Total | 1758 | 1178 | 1758 | 1178 | 2936 |

Exactly 367 unsupported sites move to diagnosed dependencies (327 tagged, 40 untagged); all other primary outcomes remain unchanged. Diagnosed own bodies remain 44 (9 tagged, 35 untagged); dependencies increase 2016 to 2383 (1491 tagged, 892 untagged). This exposes existing diagnostic barriers through generic calls. No new diagnostic was introduced, and this is not a free-cast claim. The 73 code/file rows below use the refreshed generic frontier.

The largest concrete unsupported boundary was direct generic calls: **455 distinct sites (379 tagged, 76 untagged)**. That blanket boundary is now absent. Remaining barriers include other sources on the same sites; one removed boundary does not imply that the cast is proven.

| Overlapping unsupported pattern | Before tagged | Before untagged | After tagged | After untagged |
|---|---:|---:|---:|---:|
| constructors | 0 | 36 | 0 | 36 |
| dynamic keys | 46 | 29 | 70 | 56 |
| function-value escape callback boundary | 335 | 41 | 55 | 20 |
| generic direct calls | 379 | 76 | 0 | 0 |
| unbounded callback identity | 277 | 27 | 40 | 23 |
| whole-array contracts | 0 | 5 | 0 | 4 |

The next concrete flow family is dynamic keys: 126 sites (70 tagged, 56 untagged). Missing producers affect 190 but combine unrelated causes. Callback escapes fall 376 to 75, and unknown callable identities fall 304 to 63, because direct generic helper bodies now propagate existing identities and incoming values. Arbitrary generic callable identities and intrinsic callback protocols remain unknown.

Named additions: `directAllocationCall` supplies a statically resolved, clean function body and its actual argument values to the existing graph, including generic implementations. Runtime type arguments do not select a different body; all call contexts join conservatively. `opaqueAllocationCall` retains an unknown result and taints each known callee parameter when rest/spread mapping is not represented. An ignored return value cannot hide an incoming caller. Rest parameters need a real array allocation; they are never the first object argument. Generic type arguments are not substituted into or treated as semantic field certificates.

Production already represents instantiated direct generic calls in the shared IR graph; no new ownership graph or production flow algorithm was needed. The four production allocation/eraser files are byte-identical to integration merge 27ba7d7e. This change removes a conservative limitation in the analysis source adapter and adds production witnesses. Generic callable-value admission/signature metadata from lane 5 is not required for these direct implementation identities; no code was copied.

proven-generic.a passes a factory value through a generic identity and a Base parameter. Its IR contains zero checked casts and zero checked reads; source Node, sanitized native, release native and JavaScript print hello. nonconforming-generic.a joins good and number-valued ready shapes, retains one cast and one checked read, and exits 70 naming ready, expected boolean and found number. The erasure mutant finishes with wrong output (native false; JavaScript 0), caught by semantic exit/output assertions. Existing uninitialized/readiness erasure mutants also remain caught.

All 35 analysis controls pass: five free, two readiness-only, three conforms-if and 25 unknown. The new clean generic identity is free; a staged generic result is readiness-only; a mixed generic result needs the ready guard because its declared type is number | boolean. Host, diagnosed, rest and spread origins remain unknown. A generic implementation called once with a clean object and later with JSON.parse remains unknown at the clean call as well, because this proof joins every caller.

Four valid Go measurement mutants are caught: drop later generic callers (genericJoinedRead falsely becomes free); treat the rest array as its first object (genericRestRead falsely becomes free); ignore opaque callee inputs (genericSpreadJoinedRead falsely becomes free); and drop diagnosed callee provenance (genericDiagnosedRead is incorrectly classified as unsupported flow). Logs pin each failed assertion; compiler warnings are not the catch. Restored controls pass.

Owned lower/IR checks pass in 1.235s and 0.038s, and vet passes. The uncached eraser/generic oracle passes in 5.779s. Production graph/eraser files were not changed; the graph baseline still has the eight aborting fixtures documented at e6676ba2, reproduced with the integrator versions, with the other 35 rows identical. No counts guard was weakened or refreshed. No full repository gate was run. An interim census was stopped after rest/spread guards were strengthened; only the final completed, audited run supplies these numbers.

Count-row handoff for the integrator; shared counts.md was not edited:

| Fixture | Allocations | Frees | Retains | Releases | Peak | In regions | Graph regions | Graph merges |
|---|---:|---:|---:|---:|---:|---:|---:|---:|
| proven-generic.a | 1 | 1 | 4 | 3 | 1 | 0 | 0 | 0 |
| nonconforming-generic.a | 3 | 1 | 4 | 4 | 3 | 0 | 0 | 0 |

The successful fixture frees its only object; the negative fixture records two objects still owned when the expected panic terminates the process. Host values stay unknown. The three primary host sites remain the PackageJsonInfo cache value and two ResolvedModuleFull results from host.resolveModuleNames(...).map. All overlapping host names, cast locations and target types are in [latent-host-values.json](generic-flow/latent-host-values.json).

Uncovered: dynamic indices, class constructor/field certificates, spreads/clones, arbitrary generic callable implementations, nested field contracts, post-store readiness and the dedicated one-shape-lookup-per-cast emitter. The source adapter is analysis-only and exposes no usable production IR. These limitations retain checks.

Remaining diagnostic handoff: exact adapted lines, columns and messages in [diagnostic-actions.json](generic-flow/diagnostic-actions.json), with all 260 diagnostics grouped into 73 rows. No speculative assignment to tsc adaptations versus language gaps is made; typescript can use the exact messages and current cast dependencies to split them.

Every number below is measured on a checker-rejected program.

The table covers all 260 diagnostics, grouped by code and file. Affected sites count casts whose own function or a dependency carries that group; these columns overlap across rows. Clearing the group unblocks those recorded dependencies. “Cleared alone” counts casts whose entire recorded diagnostic frontier belongs to this group. It is a counterfactual over current edges: rechecking edited source may expose other dependencies, and host/unsupported/readiness barriers remain. It is not an observed free-cast count.

| Code | File | Diagnostics | Own function sites | Dependency sites | Affected sites | Cleared alone |
|---|---|---:|---:|---:|---:|---:|
| TS2345 | `src/compiler/checker.ts` | 30 | 16 | 2320 | 2336 | 10 |
| TS18048 | `src/compiler/checker.ts` | 27 | 7 | 2313 | 2320 | 1 |
| TS18048 | `src/compiler/transformers/es2017.ts` | 6 | 0 | 2319 | 2319 | 4 |
| TS2769 | `src/compiler/checker.ts` | 5 | 9 | 2306 | 2315 | 0 |
| TS2322 | `src/compiler/transformers/es2017.ts` | 1 | 0 | 2315 | 2315 | 0 |
| TS2532 | `src/compiler/transformers/es2017.ts` | 1 | 0 | 2315 | 2315 | 0 |
| TS2345 | `src/compiler/transformers/module/esnextAnd2015.ts` | 4 | 4 | 2311 | 2315 | 4 |
| TS18048 | `src/compiler/transformers/classFields.ts` | 2 | 0 | 2314 | 2314 | 0 |
| TS2345 | `src/compiler/transformers/classFields.ts` | 4 | 0 | 2314 | 2314 | 0 |
| TS2375 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 2314 | 2314 | 0 |
| TS2532 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 2314 | 2314 | 0 |
| TS2538 | `src/compiler/transformers/classFields.ts` | 1 | 0 | 2314 | 2314 | 0 |
| TS2375 | `src/compiler/checker.ts` | 2 | 7 | 2306 | 2313 | 6 |
| TS2345 | `src/compiler/transformers/es2015.ts` | 1 | 0 | 2313 | 2313 | 0 |
| TS2532 | `src/compiler/transformers/es2015.ts` | 4 | 0 | 2313 | 2313 | 0 |
| TS2379 | `src/compiler/checker.ts` | 5 | 8 | 2304 | 2312 | 4 |
| TS18048 | `src/compiler/transformers/generators.ts` | 7 | 0 | 2312 | 2312 | 0 |
| TS2322 | `src/compiler/transformers/generators.ts` | 1 | 0 | 2312 | 2312 | 0 |
| TS2339 | `src/compiler/transformers/generators.ts` | 3 | 0 | 2312 | 2312 | 0 |
| TS2345 | `src/compiler/transformers/generators.ts` | 11 | 0 | 2312 | 2312 | 0 |
| TS2532 | `src/compiler/transformers/generators.ts` | 5 | 0 | 2312 | 2312 | 0 |
| TS2322 | `src/compiler/checker.ts` | 11 | 7 | 2304 | 2311 | 3 |
| TS2532 | `src/compiler/transformers/module/esnextAnd2015.ts` | 1 | 0 | 2311 | 2311 | 0 |
| TS2412 | `src/compiler/checker.ts` | 2 | 8 | 2301 | 2309 | 2 |
| TS2532 | `src/compiler/checker.ts` | 1 | 8 | 2301 | 2309 | 2 |
| TS2488 | `src/compiler/checker.ts` | 2 | 7 | 2301 | 2308 | 1 |
| TS2556 | `src/compiler/checker.ts` | 1 | 6 | 2301 | 2307 | 0 |
| TS2722 | `src/compiler/checker.ts` | 1 | 6 | 2301 | 2307 | 0 |
| TS2345 | `src/compiler/tsbuildPublic.ts` | 1 | 4 | 2293 | 2297 | 6 |
| TS2412 | `src/compiler/tsbuildPublic.ts` | 10 | 2 | 2294 | 2296 | 4 |
| TS2412 | `src/compiler/watchPublic.ts` | 10 | 0 | 2295 | 2295 | 3 |
| TS2345 | `src/compiler/transformers/esnext.ts` | 6 | 0 | 2294 | 2294 | 3 |
| TS2345 | `src/compiler/commandLineParser.ts` | 3 | 1 | 2292 | 2293 | 2 |
| TS2345 | `src/compiler/builder.ts` | 12 | 5 | 2287 | 2292 | 9 |
| TS2322 | `src/compiler/commandLineParser.ts` | 4 | 0 | 2292 | 2292 | 0 |
| TS2412 | `src/compiler/resolutionCache.ts` | 1 | 0 | 2292 | 2292 | 0 |
| TS2345 | `src/compiler/transformers/declarations.ts` | 1 | 1 | 2291 | 2292 | 1 |
| TS2345 | `src/compiler/transformers/jsx.ts` | 2 | 1 | 2291 | 2292 | 1 |
| TS18046 | `src/compiler/commandLineParser.ts` | 1 | 0 | 2291 | 2291 | 0 |
| TS2375 | `src/compiler/commandLineParser.ts` | 1 | 0 | 2291 | 2291 | 0 |
| TS18046 | `src/compiler/program.ts` | 2 | 0 | 2291 | 2291 | 0 |
| TS2345 | `src/compiler/program.ts` | 4 | 0 | 2291 | 2291 | 0 |
| TS2375 | `src/compiler/program.ts` | 4 | 0 | 2291 | 2291 | 0 |
| TS2412 | `src/compiler/program.ts` | 2 | 0 | 2291 | 2291 | 0 |
| TS2488 | `src/compiler/program.ts` | 1 | 0 | 2291 | 2291 | 0 |
| TS2345 | `src/compiler/sourcemap.ts` | 1 | 0 | 2291 | 2291 | 0 |
| TS18048 | `src/compiler/transformers/esDecorators.ts` | 2 | 0 | 2291 | 2291 | 0 |
| TS2538 | `src/compiler/transformers/esDecorators.ts` | 1 | 0 | 2291 | 2291 | 0 |
| TS18048 | `src/compiler/tsbuildPublic.ts` | 3 | 0 | 2291 | 2291 | 0 |
| TS2488 | `src/compiler/builder.ts` | 2 | 0 | 2243 | 2243 | 0 |
| TS2322 | `src/compiler/transformer.ts` | 4 | 0 | 2130 | 2130 | 0 |
| TS2532 | `src/compiler/transformer.ts` | 2 | 0 | 2130 | 2130 | 0 |
| TS2375 | `src/compiler/watch.ts` | 2 | 0 | 1822 | 1822 | 3 |
| TS2322 | `src/compiler/moduleNameResolver.ts` | 4 | 0 | 1522 | 1522 | 0 |
| TS18046 | `src/compiler/transformers/jsx.ts` | 2 | 0 | 1442 | 1442 | 0 |
| TS2488 | `src/compiler/transformers/jsx.ts` | 1 | 0 | 1442 | 1442 | 0 |
| TS2375 | `src/compiler/moduleNameResolver.ts` | 3 | 0 | 1238 | 1238 | 0 |
| TS2345 | `src/compiler/moduleNameResolver.ts` | 3 | 0 | 1030 | 1030 | 0 |
| TS18048 | `src/compiler/resolutionCache.ts` | 2 | 0 | 1030 | 1030 | 0 |
| TS2322 | `src/compiler/resolutionCache.ts` | 2 | 0 | 1030 | 1030 | 0 |
| TS2345 | `src/compiler/resolutionCache.ts` | 3 | 0 | 1030 | 1030 | 0 |
| TS2375 | `src/compiler/resolutionCache.ts` | 1 | 0 | 1030 | 1030 | 0 |
| TS2379 | `src/compiler/resolutionCache.ts` | 3 | 0 | 1030 | 1030 | 0 |
| TS2307 | `src/compiler/sys.ts` | 1 | 0 | 924 | 924 | 0 |
| TS2322 | `src/compiler/sys.ts` | 1 | 0 | 924 | 924 | 0 |
| TS2345 | `src/compiler/sys.ts` | 1 | 0 | 924 | 924 | 0 |
| TS18046 | `src/compiler/sys.ts` | 1 | 0 | 862 | 862 | 0 |
| TS2375 | `src/compiler/builder.ts` | 2 | 0 | 573 | 573 | 0 |
| TS2769 | `src/compiler/builder.ts` | 1 | 0 | 573 | 573 | 0 |
| TS2322 | `src/compiler/transformers/esDecorators.ts` | 2 | 0 | 356 | 356 | 0 |
| TS2420 | `src/compiler/checker.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2740 | `src/compiler/core.ts` | 1 | 0 | 0 | 0 | 0 |
| TS2379 | `src/compiler/watchPublic.ts` | 2 | 0 | 0 | 0 | 0 |

Diagnostic count check: 260; diagnosed casts: 2427. Exact adapted line/column locations and messages for every diagnostic are in diagnostic-actions.json. Locations refer to the adapted tree recorded in this census source hashes.

Reproduction:

```sh
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
python3 stage3/shape-conformance/latent/make-overlay.py /tmp/shape-generic-final-overlay
go build -buildvcs=false -overlay=/tmp/shape-generic-final-overlay/overlay.json -o /tmp/shape-generic-final-census ./stage3/shape-conformance/latent/tool
GOMEMLIMIT=3GiB GOGC=50 /tmp/shape-generic-final-census /tmp/shape-stage3-234ab1aa-adapted /tmp/shape-adapted-current-map.json /tmp/shape-generic-final-result.json > stage3/shape-conformance/logs/generic-census.log 2>&1
python3 stage3/shape-conformance/latent/audit.py /tmp/shape-generic-final-result.json /tmp/shape-adapted-current-map.json /tmp/shape-stage3-234ab1aa-adapted > stage3/shape-conformance/logs/generic-census-audit.log 2>&1
go test ./internal/lower ./internal/ir -run "TestShape|TestAllocationFlow|TestClosureTargets" -count=1 > stage3/shape-conformance/logs/generic-unit.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run "^(TestCheckedViewShapeErasure|TestShapeGenericCountRows)$" -count=1 -v > stage3/shape-conformance/logs/generic-oracle.log 2>&1
python3 stage3/shape-conformance/generic-flow-mutants.py > stage3/shape-conformance/logs/generic-mutants.log 2>&1
```
