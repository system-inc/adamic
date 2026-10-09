Rebuilt stricter options and optional guards on compiler/optional-presence a774d316 for #k881crd.
Lane commits: fdf3a30e1, 32d36590f, d79beadc3; main interaction repair: d28193f9e; delivery audit SHA is in the handoff.
Own Node-held backend fixtures, native sanitizers, selected regression tests, build/vet, stage 1 gaps, stage 3 and Linux counts pass.
Attribution, production, library, per-witness runtime, optional presence/storage/cleanup and namespace visitor mutants are caught.
Whole native tsc acceptance and the 2,720/2,936 corpus measurement remain pending on the dependencies below.

The branch starts at a774d31656ebbb183d259e445b56135a71d5a1fb, whose parent is main 031a1259. No old bases, area merges, reference-loader history or views implementation were imported. No protected compiler file was changed.

| Lane | Source SHA | Rebuild commit | Selection |
|---|---|---|---|
| codex/stricter-options-next | 8f32e51e | fdf3a30e1 | canonical reconciled implementation |
| codex/stricter-catch-variables | ded85b2d | fdf3a30e1 | duplicate; use reconciled 8f32e51e |
| codex/stricter-indexed-a | cf8048da | fdf3a30e1 | duplicate; use reconciled 8f32e51e |
| codex/stricter-indexed-b | 3ccc3268 | fdf3a30e1 | duplicate; use reconciled 8f32e51e |
| codex/stricter-indexed-c | 00fa8dfc | fdf3a30e1 | duplicate; use reconciled 8f32e51e |
| codex/stricter-indexed-d | 1a83ba1a | fdf3a30e1 | duplicate; use reconciled 8f32e51e |
| codex/stricter-indexed-all | aee98c83 | fdf3a30e1 | duplicate; use reconciled 8f32e51e |
| codex/stricter-records | b150f83c | fdf3a30e1 | duplicate; use reconciled 8f32e51e |
| codex/stricter-optional-writes | e2f994418 | 32d36590f | only own guard delta against d335b9b19 |
| codex/stricter-options-checks | c09c22b8 | d79beadc3 | optional scheduling on shared ledger framework |

The first lane is the net 337aa466..8f32e51e change. Its TRANSPLANT.csv identifies the reconciled sibling work; applying the siblings again would duplicate it. The optional guard patch is restricted to paths owned by its commits and excludes unrelated multi-root, project-entry and weak-read changes. The checks framework already occurs in the first lane; c09c22b8 supplies its missing optional scheduling policy, keeping compilation and census on one dispatcher.

Conflicts: the first lane's nine conflicts keep main checked enumeration, assertions, namespace readiness, lazy initializers, contextual array inference and representations alongside nullable fields, tagged caught values and readonly records. Main localRead retains its checks exactly once. The optional lane's eight conflicts retain the current CLI, one scheduler, ordinary diagnostics, structural metadata, catchProperty emission and counts registry alongside its guards. The missing old area node_process.c and parallel.c owners were excluded. Three absent-write mutants now target main adamic_object_write_field rather than obsolete initialization emission. D069's old presence block is removed only after its release, sanitized and JavaScript witnesses pass.

The sweep found three main entries fixtures incorrectly reclassified by the stricter dictionary rule. The repair keeps main boxed growing literals and scalar-union readonly enumeration views, while supported stricter dictionaries keep their ordered store. Main entries, misfit and readiness mutants pass, as do the stricter record boundaries. Two optional-indexing assertions now expect the precise unsupported lookup/storage NotYet reason. No unsupported lookup was admitted.

The entry overlay exposed a real compiler panic when namespace preflight cast a ClassExpression as ClassDeclaration. Both class forms now retain their own heritage and static-member initialization checks. The reduced program is internal/lower/testdata/namespaces_notyet/class_expression_before_namespace.a. Node stops with TypeError before N initializes; lowering reports NotYet. Restoring the blind cast fails TestNamespaceClosedCallGraphEdges through the original panic, with no build failure counted as a kill.

The production census uses area/stage3 3b255125's scratch apply.sh tree at pinned TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8. All 79 source hashes match whole-program/source-hashes.json byte for byte. Install only Node 25.3.3 and source-map-support declarations/runtime in a scratch dependency directory. Run:

```
go run ./stage3/stricter-options/whole-program <ledger-adapted> stage3/stricter-options/whole-program/rows.csv
```

Result: 173 scheduled contracts: indexed 99, optional 67, caught 5, JSON 2; remaining 0, ordinary errors 0, identity drift 0. These are loader obligations, not emitted whole-program guards. An initial measurement against current main adaptations correctly rejected an unlisted site; it is not counted as reproducing the pinned ledger. A missing source-map-support declaration was installed before the final census. The normalized current report and regenerated states are committed.

Actual project-reference entry acceptance is pending: the pinned base lacks codex/project-references-source 88fe8de4 and reports TS6305 at src/tsc/_namespaces/ts.ts:3:15. For the requested measurement, a scratch Go overlay adds that branch's compatible-reference source-root flattening and ambient types union to the current production loader and option audit. It does not waive ordinary diagnostics or change guard admission. Its script is entry-overlay.py. Run the resulting CLI on current stage3/apply.sh output with ADAMIC_NATIVE_SPLIT=0 and 1. Both exit 1 with identical explicit NotYet at corePublic.ts:14:5, MapLike's mutable index signature. The records landing b58629a2 is also absent from the pinned base. No native tsc execution is claimed.

The table is the first fifteen error headers in 88fe8de4's stored entry-0.stderr. Every earlier diagnostic is absent in the new source-reference measurement; body acceptance remains pending, since MapLike stops emission. The old report attributes its first iterator diagnostic to the embedded prelude, so this table does not claim fifteen emitted strictness guards.

| Earlier location | Earlier code | Current loader measurement | Whole body |
|---|---|---|---|
| src/compiler/builder.ts:1246:69 | TS2345 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:1258:65 | TS2488 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:1292:61 | TS2488 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:1345:64 | TS2345 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:1347:41 | TS2345 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:1347:97 | TS2345 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:1493:60 | TS2345 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:1495:26 | TS2345 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:1496:45 | TS2345 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:1562:64 | TS2345 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:2273:9 | TS2375 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:2310:9 | TS2375 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:2377:13 | TS2769 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:2379:45 | TS2345 | Earlier diagnostic absent | Pending |
| src/compiler/builder.ts:395:118 | TS2345 | Earlier diagnostic absent | Pending |

Lane 3's afb8449b harness was built through a scratch overlay using its exact latent/main.go.txt. It cannot build here: load.LatentLoad, lower.LatentShapeSite and lower.MeasureLatentShapes are absent. The historical 2,720/2,936 is not a current measurement. Whole-body shape acceptance and checked-view optional writes remain pending: acceptance dependency: compiler/views-rehearsal 2b6c032a. TestOptionalFieldCheckedViewPending remains skipped, never passed. OptionalWideningCensus lacks its configured external corpus; nullable timing is opt-in. The main entries acceptance suite also skips because its six-fixture branch is absent from the base; existing main entries witnesses run and pass.

Current validation commands, with output in evidence logs:

```
go test ./stage3/stricter-options ./stage3/stricter-records ./stage3/stricter-indexed-a ./stage3/stricter-indexed-b ./stage3/stricter-indexed-c ./stage3/stricter-indexed-d ./stage3/stricter-indexed-all -count=1 -timeout 30m -v
python3 stage3/stricter-options/mutants.py
python3 stage3/stricter-options/production_mutants.py
python3 internal/load/testdata/project-lib/run_mutants.py /tmp/stricter-options-project-lib-mutants
go test ./internal/load ./cmd/adamic ./internal/lower ./internal/native ./internal/oracle -run 'Test(Project|Production|OptionLedger|ReadOptionLedger|Explain|Caught|CollectionIterator|LibraryIterator|Optional|JSON|Sparse|NumericTypedArray|ReadonlyRecord)' -count=1 -timeout 20m -v
go test ./internal/oracle ./stage3/stricter-records -run 'TestEntries|TestRecord' -count=1 -timeout 15m -v
go test ./internal/lower -run '^TestNamespace' -count=1 -v
go test ./stage1/... -run 'Gap|Gaps|Probes' -count=1 -timeout 30m -v
npm ci --prefix stage3/api --ignore-scripts --no-audit --no-fund
go test ./stage3/fixtures -count=1 -timeout 30m
go build ./...
go vet ./...
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m
```

The pinned fast Gate.aCheck implementation fbac28c6 checks all 41 changed .a against origin/main, including the built-ahead dependency. All pass. Two imported negative witnesses now have measured first-line headers; no executable source or Node observation was rewritten. Gofmt checks 122 changed Go files. No whole oracle, whole native package or full gate was run.

Mutation evidence: six attribution overlays, four production overlays and seven library overlays fail through their intended assertions. Every supported indexed witness checks exact failure text, native release/sanitizers and JavaScript; guard-erasure mutants compile and differ in exit, message or Node output. D069 now also runs. Sparse-array, typed-array, record/prototype, JSON, caught-value/host-brand, nullable-sentinel and Map-key mutants run in the own suites. The thirteen optional fixtures match Node, and their absent construction/write, storage-kind, missing producer and cleanup/leak mutants are caught. The visitor mutant restores the observed compiler panic. A mutant's deliberately wrong exit-0 output is evidence of the test catching it, not a production miscompile.

The stage 3 audit updates only eleven stage0 objects: ten unsupported dictionary diagnostics advance from Refused to NotYet; one array-rest diagnostic advances within NotYet. No Compiles/CheckedStop fixture regresses. Every Node observation and every byte outside stage0 stays unchanged. status-audit.json lists all eleven entries and reasons. Counts are regenerated on Linux and verified, with no old row removed; counts-audit.json lists additions and changes.

Setup succeeded with GOPROXY=https://proxy.golang.org|direct. Timing seconds: Go .025, Node .026, submodules .065, Markdown .073, clang .172, build 39.971, deferred binaries 40.087, cache 40.089, done 40.116. nproc=5; CPU quota=4; memory=17.6 GB. Env /workspace/adamic-tools/env.sh; Go 1.27.1, Node 24.19.0, clang 20.1.8. These are observed setup timings, not estimates.
