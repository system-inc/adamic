Built: step 09 callable certificates for 10 pairs / 10 candidate reads; 10 additional pairs / 10 reads have pinned code frontiers.
Commits: codex/views-callables-d, based on share-a 18d1c9da9e1889fa3f042cb6997a6734b404ccf2; delivery SHA is reported after commit.
Checks: focused Node, release native, ASan/UBSan, finishing leaks, JavaScript, original-source verification, vet and 30 owned counts rows pass; global counts is outside-share red.
Mutants: 10 arity mutants caught independently by the normal negative tests; 20 native/JavaScript mutant executions caught in the passing suite; one mapped-declaration mutant caught by source verification.
Uncovered: this is a 20-pair checkpoint, not the whole share; 669 pairs / 1040 reads remain pending after the inherited ledger and this checkpoint.

The bottom-up batch comprises the first 20 pending share-a ranks, 2817 through 2760, excluding its certificates and needs-code entries. The latest fetched share-a tip is unchanged, and no pair overlaps. Its inherited 65 certificates / 268 reads and 80 frontiers / 1231 reads remain intact. With this checkpoint the combined ledgers have 75 certificates / 278 reads and 90 frontiers / 1241 reads. No lane or compiler implementation was merged or edited.

The source pin is upstream TypeScript 050880ce59e30b356b686bd3144efe24f875ebc8, independently checked out at /tmp/views-callables-d-original. Stock checker TypeScript 6.0.3 and @types/node 25.3.3 come from the existing stage3/api lock. Witness ledgers retain original read expressions, AST spans, complete callable declarations and source SHA-256 hashes. External declaration paths use api: relative to stage3/api/node_modules so they do not depend on this workspace layout.

These are reduced original-member contract fixtures, not executions of every original application context. Adjacent object carriers become readonly numeric value fields, flags and enum carriers become numbers, NodeArray becomes a readonly array, and reduced CompilerOptions fields are optional booleans. This does not certify original enum domains or original producer bodies. The four computed-option families retain the original mapped computeValue declaration verbatim. Optional members retain their original declaration, including Required/Pick where the original read has already narrowed them.

| Certified rank | Original read | Reads | Node stdout |
| --- | --- | ---: | --- |
| 2814 | `nodeBuilder.symbolToTypeParameterDeclarations` | 1 | `1` |
| 2811 | `nodeBuilder.serializeTypeForExpression` | 1 | `7` |
| 2790 | `sys.readFile` | 1 | `ok` |
| 2781 | `_computedOptions.allowSyntheticDefaultImports.computeValue` | 1 | `true` |
| 2778 | `_computedOptions.resolvePackageJsonExports.computeValue` | 1 | `true` |
| 2775 | `_computedOptions.preserveConstEnums.computeValue` | 1 | `true` |
| 2772 | `_computedOptions.incremental.computeValue` | 1 | `true` |
| 2769 | `colors.brightWhite` | 1 | `ok` |
| 2766 | `tracing?.stopTracing` | 1 | `done` |
| 2763 | `tracing.instant` | 1 | `done` |

Each certificate has good, wrong-arity and wrong-value .a fixtures. Good source Node stdout, stderr and exit status match both backends, release and sanitized native; the good native run also passes the independent leak check. Node permits the reduced wrong-arity producers, whereas Adamic must stop at the checked member read with exit 70 and a complete pinned diagnostic. Wrong-kind source Node raises TypeError; Adamic has its own exact field guard message. No sanitizer crash or invalid C is accepted as mutant evidence.

The arity mutant changes only the expected callable IR parameter list to the malformed producer arity. It leaves producer metadata, calling convention, body, result contract and call untouched. Both backends then finish with Node output and exit 0. The pinned exit-70 matcher catches that behavior. The independent ADAMIC_CALLABLE_D_MUTANT run fails all 10 wrong-arity subtests. The source mutant weakens computeValue to optional in rank 2778 and also alters its self-recorded carrier; verification against the original upstream mapped declaration still exits 1. Restoration returns green.

| Code-needed rank | Original read | Required work |
| --- | --- | --- |
| 2817 | `watchFile` | Demand-time callable representation conversion for an original destructured HostWatchFile binding, preserving higher-order parameter and result contracts. |
| 2808 | `startRecoveryScope` | Demand-time callable representation conversion for destructuring a method whose result is another callable. |
| 2805 | `finalizeBoundary` | Demand-time callable representation conversion for the original destructured boolean-returning method. |
| 2802 | `pendingDeclarations.push` | Rest callable descriptor and real Array receiver transport for the original push signature; preserve its original object item fields. |
| 2799 | `directoryExists` | Demand-time callable representation conversion for the original destructured optional directoryExists result. |
| 2796 | `cachedFunc.assertion` | Sound never-rest function value lowering and descriptor coverage for the exact AnyFunction alias, without inventing admissible arguments. |
| 2793 | `host.setTimeout` | Original signature has an any result and a rest callback; lowering stops at a call returning any. Certification needs a ruled sound boundary without erasing or narrowing the advertised original contract. |
| 2787 | `host.getCurrentDirectory` | Check optional callable fields at demand time while retaining the original optional host member; the whole structural view is currently uncheckable. |
| 2784 | `candidates.push` | Rest callable descriptor and real Array receiver transport for the original push signature; preserve the original item shape. |
| 2760 | `_fs.watchFile` | Demand-time selection and producer coverage for all original watchFile overloads, preserving callback/optional domains. |

All 10 frontier controls have valid source Node observations and a complete exact lowering stop with line and column in needs-code.json. They are not certificates; no native runtime or leak claim is made for them. Initial fixture construction and diagnostic-pin corrections were completed before frontier classification and are not counted as compiler findings.

Exactly 30 counts rows were appended, one for each of the 10 certified families and three variants. Existing counts.md bytes remain unchanged. Good rows finish and balance allocations/frees; stopping rows record the pinned early stop rather than a finishing leak result. The required global update exits 1 on unrelated existing process.exit, overload/template, MaybeNumber, fs option/error/return and graph-region free failures, documented individually in logs/counts-global.log. There is no share-d failure in that log. It did not write a partial global table. The owned append-only updater and subsequent read-only counts check pass.

Commands executed with all output redirected to logs:

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/views-callables-d-setup.log 2>&1
source /workspace/adamic-tools/env.sh
npm ci --prefix stage3/api > /tmp/views-callables-d-api.log 2>&1
node stage3/interface-downcasts/lane5/share-d/prepare.cjs /tmp/views-callables-d-original stage3/api/node_modules/typescript > /tmp/views-callables-d-prepare.log 2>&1
node stage3/interface-downcasts/lane5/share-d/verify.cjs /tmp/views-callables-d-original stage3/api/node_modules/typescript > /tmp/views-callables-d-verify-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareD$|^TestCheckedViewCallableShareDFrontiers$' -count=1 -v -timeout 10m > /tmp/views-callables-d-oracle.log 2>&1
ADAMIC_GATE_UNCACHED=1 ADAMIC_CALLABLE_D_MUTANT=1 go test ./internal/oracle -run '^TestCheckedViewCallableShareD$/^rank-/^wrong-arity$' -count=1 -v -timeout 10m > /tmp/views-callables-d-arity-mutants.log 2>&1
node stage3/interface-downcasts/lane5/share-d/mutant-source.cjs /tmp/views-callables-d-original stage3/api/node_modules/typescript > /tmp/views-callables-d-source-mutant-final.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 10m -args -update-counts > /tmp/views-callables-d-counts-global.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableShareDCounts$' -count=1 -timeout 10m -args -update-counts > /tmp/views-callables-d-counts.log 2>&1
go test ./internal/oracle -run '^TestCheckedViewCallableShareDCounts$' -count=1 > /tmp/views-callables-d-counts-check.log 2>&1
gofmt -w internal/oracle/checked_views_callable_share_d_test.go
go vet ./internal/oracle > /tmp/views-callables-d-vet.log 2>&1
```

Focused oracle PASS 15.547s; independent arity-mutant run exits 1 in 4.517s with all 10 intended catchers failing. Original source verification covers 40 fixtures / 20 pairs / 20 candidate reads. Regeneration preserves every fixture byte. Counts check PASS 3.528s; vet passes. No whole package test or full gate was run.

Setup timing lines: Go ready 0.021s; Node ready 0.022s; submodules ready 0.051s; markdown dependency validation step 0.006s, ready 0.074s; clang ready 0.215s; Go build ready 43.609s; test binaries deferred 43.785s; build cache warm 43.786s; done 43.813s. nproc=5, cpu.max=400000 100000. Toolchain is Go 1.27.1, Node 24.19.0 and clang 20.1.8. Setup succeeded.
