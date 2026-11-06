# Paired any-members census (tmdhz1d)

**Head any-member findings: 12. Unit first stops among them: 0.**

All 12 are refusal-scan findings: 11 existing NotYet findings and one new Refused finding. None matches the first lowering stop of its lexical owning unit. Scan `unit` fields can be file-wide bookkeeping locations, so both the raw scan owner and independently resolved lexical top-level owner are retained below and in delta.json.

- `src/compiler/builder.ts:393:42`; lexical owner `src/compiler/builder.ts:321:1` (createBuilderProgramState); raw scan owner `src/compiler/builder.ts:2508:1`; phase `refusal_scan`; unit first stop: no; present on base: yes. Full reason: `a member typed any (keys()next()<1>) related to undefined at the assignment site`
- `src/compiler/commandLineParser.ts:2278:82`; lexical owner `src/compiler/commandLineParser.ts:2278:1` (parseConfigFileTextToJson); raw scan owner `src/compiler/commandLineParser.ts:2278:1`; phase `refusal_scan`; unit first stop: no; present on base: no. Full reason: `a member typed any (config, declared): any has no proven type, so every read of it through a typed view is unchecked`
- `src/compiler/commandLineParser.ts:3839:20`; lexical owner `src/compiler/commandLineParser.ts:3823:1` (convertJsonOption); raw scan owner `src/compiler/commandLineParser.ts:4323:1`; phase `refusal_scan`; unit first stop: no; present on base: yes. Full reason: `a member typed any ([]) related to string at the return site`
- `src/compiler/core.ts:1066:97`; lexical owner `src/compiler/core.ts:1066:1` (KindVariableStatement); raw scan owner `src/compiler/core.ts:2614:1`; phase `refusal_scan`; unit first stop: no; present on base: yes. Full reason: `a member typed any (()) related to T | undefined at the initializer site`
- `src/compiler/debug.ts:1036:39`; lexical owner `src/compiler/debug.ts:111:1` (Debug); raw scan owner `src/compiler/debug.ts:1234:1`; phase `refusal_scan`; unit first stop: no; present on base: yes. Full reason: `a member typed any ([]) related to number at the initializer site`
- `src/compiler/emitter.ts:2941:40`; lexical owner `src/compiler/emitter.ts:1211:1` (createPrinter); raw scan owner `src/compiler/emitter.ts:6379:1`; phase `refusal_scan`; unit first stop: no; present on base: yes. Full reason: `a member typed any ((0)) related to Expression at the assignment site`
- `src/compiler/factory/nodeFactory.ts:6098:17`; lexical owner `src/compiler/factory/nodeFactory.ts:492:1` (createNodeFactory); raw scan owner `src/compiler/factory/nodeFactory.ts:7543:1`; phase `refusal_scan`; unit first stop: no; present on base: yes. Full reason: `a member typed any (set(0)) related to number | undefined at the object field site`
- `src/compiler/parser.ts:1253:17`; lexical owner `src/compiler/parser.ts:1249:1` (forEachChild); raw scan owner `src/compiler/parser.ts:10824:1`; phase `refusal_scan`; unit first stop: no; present on base: yes. Full reason: `a member typed any (167(0)) related to QualifiedName at the assertion site`
- `src/compiler/sys.ts:382:121`; lexical owner `src/compiler/sys.ts:249:1` (createDynamicPriorityPollingWatchFile); raw scan owner `src/compiler/sys.ts:1976:1`; phase `refusal_scan`; unit first stop: no; present on base: yes. Full reason: `a member typed any ((0)) related to string at the relation site`
- `src/compiler/tsbuildPublic.ts:2068:69`; lexical owner `src/compiler/tsbuildPublic.ts:2060:1` (scheduleBuildInvalidatedProject); raw scan owner `src/compiler/tsbuildPublic.ts:2494:1`; phase `refusal_scan`; unit first stop: no; present on base: yes. Full reason: `a member typed any ((0)) related to string at the argument site`
- `src/compiler/utilities.ts:8555:49`; lexical owner `src/compiler/utilities.ts:8555:1` (KindVariableStatement); raw scan owner `src/compiler/utilities.ts:12451:1`; phase `refusal_scan`; unit first stop: no; present on base: yes. Full reason: `a member typed any (getTokenConstructor()) related to new <TKind extends SyntaxKind>(kind: TKind, pos: number, end: number) => Token<TKind> at the initializer site`
- `src/compiler/visitorPublic.ts:603:17`; lexical owner `src/compiler/visitorPublic.ts:598:1` (visitEachChild); raw scan owner `src/compiler/visitorPublic.ts:1809:1`; phase `refusal_scan`; unit first stop: no; present on base: yes. Full reason: `a member typed any (167(0)) related to QualifiedName at the assertion site`

Base `b0efce6dce1cd7445ed8a570aecac387a4a7d684`. Head `66fff4ecc05c3bfb2258cf4bb2a7b0a2b762c59b` (one commit on base).

TypeScript v6.0.3, upstream `050880ce59e30b356b686bd3144efe24f875ebc8`. Cohere `06af2e399c6c1fb93fd570b774027942747400b4`; checker `d92d9bfee114c80be2c375d72edae966176e3a4f`.

Box `3cbd7fe9a660`: 5 available processors, CPU quota 4 cores, cgroup memory limit 16 GiB. Host metadata is retained in `evidence/box.json` and `evidence/cgroup.json`.

## Base union admission check

All 23 base pieces and 23 matching head pieces completed, each exited 0, with all 79 compiler files, 4,555 unique units and 4,056 unique timed attempts per compiler. All 746 adapted source hashes still match the original manifest. The longest child was 639.408 seconds, well inside 30 minutes.

The base union differs from the whole base in **65 deferred-generic instantiated flags only**. All 4,555 units retain the same ordered lowering findings and first stop; all file scans and declaration inventories agree. Exactly 4,490 units agree in every field, including all 4,056 timed units. The 65 differing deferred units are excluded from both pairing streams. `UNION.md` lists every differing unit, its old and new flag, and the reason; `evidence/pieces/union-proof.json` retains exact machine-readable differences.

A piece can change a deferred generic witness because later callers in other pieces are absent from its census-wide witness map. That map annotates `instantiated` after selected calls finish; this metadata does not itself select the first lowering stop. The absence of first-stop differences here is measured, not assumed. Witness bits are neither ORed nor masked to claim equality. The gate also compares ordered lowering findings and every other unit field, not just first stops.

## Regressions and changed first stops

### Raw Findings

Regressions: 1. Changed unit first stops: 0. New finding sites: 1.

- previously_finding_free_scan_site: `src/compiler/commandLineParser.ts:2278:82`; owner `src/compiler/commandLineParser.ts:2278:1`; Refused: a member typed any (config, declared): any has no proven type, so every read of it through a typed view is unchecked
- New site `src/compiler/commandLineParser.ts:2278:82`; owner `src/compiler/commandLineParser.ts:2278:1`; phase `refusal_scan`; Refused: a member typed any (config, declared): any has no proven type, so every read of it through a typed view is unchecked
### Distinct Reads

Regressions: 1. Changed unit first stops: 0. New finding sites: 1.

- previously_finding_free_scan_site: `src/compiler/commandLineParser.ts:2278:82`; owner `src/compiler/commandLineParser.ts:2278:1`; Refused: a member typed any (config, declared): any has no proven type, so every read of it through a typed view is unchecked
- New site `src/compiler/commandLineParser.ts:2278:82`; owner `src/compiler/commandLineParser.ts:2278:1`; phase `refusal_scan`; Refused: a member typed any (config, declared): any has no proven type, so every read of it through a typed view is unchecked

## Unique findings

Counts deduplicate `(kind, where, reason, text)` across attempts, as the meter does. Repeated file scans count once. Counts cover the full piece unions; excluding the 65 deferred units removes no findings, so the paired counts are identical. Whole-base counts also match.

| Kind | Base | Head |
| --- | ---: | ---: |
| NotYet | 620 | 620 |
| Refused | 4708 | 4709 |
| panic | 0 | 0 |
| error | 0 | 0 |

## Unit lowering time on this box

One sample per compiler, assembled from the matching sequential pieces. These are sums of all 4,056 unit timers, including synchronous dependency work. Checker work, file scans and shared process setup are excluded. Base runs before head within each piece; this is not a benchmark distribution.

| Sum | Base seconds | Head seconds | Head minus base |
| --- | ---: | ---: | ---: |
| wall_seconds | 5572.796798 | 5527.646214 | -45.150584 |
| cpu_seconds | 6199.835964 | 6151.315293 | -48.520671 |

Earlier whole-base reference: wall 5902.087133 s; CPU 6887.892272 s. The piece sample repeats setup and resets process caches, so it is not substituted for whole-run timing. Successful child process wall sums: base 7767.960 s; head 7680.524 s. All per-unit deltas are in `delta.json`.

## Method and evidence

Copied the method after reading `typescript/kind-predicate-census` REPORT.md and compare.py at `33dca393`, and the meter invocation. One shared adapted tree was built with base `stage3/apply.sh` and lockfile-pinned API packages. Adaptation and census inputs are unchanged between compiler pins; the head also changes unrelated interface-downcast fixtures. The exact compiler/dependency pins, original binary hashes and source manifest remain under `evidence/`.

Existing speculative-only piece support and its union proof were read. It cannot select first-error attempts, so the new measurement helper adds file and top-level-unit scheduling guards to the same no-output latent overlays. It leaves body lowering, registration, dependencies, generic deferral and witness recording unchanged. Each process still loads the complete compiler program. Mode-off output equality for both compilers and an actual cross-file witness difference were checked on a fixture; dropped pieces, duplicate pieces and a finding moved outside its assigned unit were rejected. See `evidence/pieces/selector-audit.json`, the selector patches and `first-error-piece.go.txt`.

Ordinary files form seven disjoint file groups. Utilities and utilitiesPublic exceed a safe whole-file budget, so their top-level units form sixteen disjoint chunks. Every unit is assigned exactly once. Repeated file scans are checked identical and retained once; declaration inventories are not summed. Cost estimates use the retained base LATENT_TIMES. See `PIECES.md` and `evidence/pieces/plan.json`.

Each process runs sequentially in first-error mode with `LATENT_FULL=0`, `LATENT_SPECULATIVE=0`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TIMES` and the explicit selection file. The hard child limit is 20 minutes; five minutes without timer progress requests a goroutine dump and stops. No successful child hit either limit. Header-only finding buffers are expected while timers progress because Latent finalizes deferred witnesses before flushing. No mode switch or backend invocation occurred.

Every completed child was committed immediately to scratch Git, then losslessly archived and committed on this report branch. Checkpoint IDs and hashes are beside each stream. Two interrupted unfinished piece attempts were retried; neither supplies observations here. A transient remote `commit_refs_rules` error delayed ordinary pushes, and a later ordinary push published the pending commits. Every successful checkpoint has now been pushed. Earlier incomplete whole-head attempts remain as historical evidence and supply no findings to this result.

Before feature pairing, `union.py` checked every base unit against the retained complete whole stream and wrote the admission proof. The actual result was reported before invoking main `stage3/census/pair`. Pairing uses shared source hashes and exact offsets, retaining raw findings and normalized access-chain reads. Only lowering-phase events define unit first stops; scan owners remain separate. `delta.json` retains every transition, regression and any-member owner. `pair-report.md` retains the standard report.

The checker accepted the program: checker_rejected=false, with no diagnostics. The fixed measurement label in the raw streams is retained verbatim. Finding-free applies to these eligible first-error attempts and scans; it does not establish complete lowering past earlier stops or native compilation, nor supply whole-run witness flags for the head.

`validate.py` independently checks union arithmetic, admission exclusions, first stops, kind counts, timings, source/pin provenance and artifact hashes. See `validation.txt`.
