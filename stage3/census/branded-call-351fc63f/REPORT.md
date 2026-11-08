# Paired latent census: branded call 351fc63f against b5c0aa64

Task fx9waqy. Compiler request, November 8, 2026. Report only; no compiler or adaptation change lands.

**Entry comparisons are still pending. No complete entry regression verdict is claimed.** Full-core and focused observations have completed.

## Pins and applied inputs

| Role | Pin |
| --- | --- |
| Parent compiler and A apply | `b5c0aa648a7a35b2f17899895d009eabd4147000` |
| Head compiler | `351fc63fa2573036e194656fb92c35fe03858afe` |
| Parent main | `3fbe3a2ab40f22f8837a66daaad3b5fb845376e7` |
| B adapter source | `eeb87728b743235cf1df68dfed17241f63416734` |
| TypeScript 6.0.3 | `050880ce59e30b356b686bd3144efe24f875ebc8` |

A uses the unmodified parent apply. B uses a separate parent checkout carrying exactly `99-correctness-core-or-predicate` and `99-correctness-core-multimap` from the adapter pin. Both applies succeeded; B needed no pin repair. The two 81-source measured graphs differ only in `src/compiler/core.ts` and `src/compiler/checker.ts`. Every completed run has zero checker diagnostics.

Compiler and measurement source bytes match their own pins. Both compilers use the same pinned checker dependency `06af2e399c6c1fb93fd570b774027942747400b4`. [Provenance](provenance.json), [input comparison](input-comparison.json), [A hashes](A-source-hashes.json), [B hashes](B-source-hashes.json), and [patch sets](A-patch-set.md) ([B](B-patch-set.md)) retain the evidence.

## Entry first-error pairing

These are the guarded meter entry driver over `src/tsc/tsc.ts`, with all 81 reached source files. `LATENT_ASSERT_NO_OUTPUT=1` is required. No speculative, mutant or forced-generic mode is used. Raw below means pair.py raw-position findings, deduplicated by its exact key; distinct means its outermost-access-chain normalization. New findings include replacement reasons and therefore are not disjoint from changed findings. New sites and first-stop changes are shown separately.

| Pairing | View | Cleared | New findings | Changed findings | New sites | First-stop unit changes | Regression entries |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| A | Raw | pending | pending | pending | pending | pending | pending |
| A | Distinct | pending | pending | pending | pending | pending | pending |
| B | Raw | pending | pending | pending | pending | pending | pending |
| B | Distinct | pending | pending | pending | pending | pending | pending |

A previously finding-free unit that now stops, a previously clear scan site that now finds, and a new panic are regressions. A different first stop in an already-blocked unit is reported separately. Normalized identities control the site judgment.

| Pairing | Unit | Parent first stop | Head first stop | Classification |
| --- | --- | --- | --- | --- |
| A | pending | | | |
| B | pending | | | |

Raw streams and complete deltas: [A parent](A-parent-entry.jsonl), [A head](A-head-entry.jsonl), [A paired JSON](A-entry-pair.json), [B parent](B-parent-entry.jsonl), [B head](B-head-entry.jsonl), [B paired JSON](B-entry-pair.json). A link to an unfinished stream is partial evidence, not a complete census.

## Full core.ts pairing

The reference procedure is preserved in [reference-run-census.py](reference-run-census.py) and its [report](reference-core-census-REPORT.md). Each compiler loads the full compiler-directory program and selects only the exact adapted `core.ts` using `LATENT_FULL=1`, `LATENT_ONLY_FILE` and `LATENT_ASSERT_NO_OUTPUT=1`. No speculative or replay flags apply to these four runs. The legacy checker-rejected label remains verbatim despite zero actual checker diagnostics.

Within each pairing the complete parent and head streams are **byte-identical** ([hashes](core-equality.json)). Each has 270 units, 14 NotYet events, 4 Refused events and 12 recorder Boundary events. Thus cleared/new/changed are **0/0/0 in both raw and distinct views**, and no other core.ts site moves.

A does not contain `orPredicate2` or `orPredicate3`. In B, both helper declarations and all three remaining `or` declarations are deferred and uninstantiated. There is no recorded site to quote in any of those bodies. This is not evidence that a deferred body lowers.

### Every requested declaration and recorded site

#### Pairing A

All kinds and phases are retained here, including recorder Boundary events. Repeated contexts share one row; raw streams retain them all.

| Owner | Location | Kind | Phase | Exact reason | Parent | Head |
| --- | --- | --- | --- | --- | ---: | ---: |
| No recorded sites | | | | | 0 | 0 |

Declaration unit statuses:

| Side | Name | Unit | Status | Instantiated |
| --- | --- | --- | --- | --- |
| parent | or | compiler/core.ts:2462:1 | deferred_generic_instantiation | False |
| parent | or | compiler/core.ts:2464:1 | deferred_generic_instantiation | False |
| parent | or | compiler/core.ts:2467:1 | deferred_generic_instantiation | False |
| parent | or | compiler/core.ts:2469:1 | deferred_generic_instantiation | False |
| parent | or | compiler/core.ts:2471:1 | deferred_generic_instantiation | False |
| head | or | compiler/core.ts:2462:1 | deferred_generic_instantiation | False |
| head | or | compiler/core.ts:2464:1 | deferred_generic_instantiation | False |
| head | or | compiler/core.ts:2467:1 | deferred_generic_instantiation | False |
| head | or | compiler/core.ts:2469:1 | deferred_generic_instantiation | False |
| head | or | compiler/core.ts:2471:1 | deferred_generic_instantiation | False |

#### Pairing B

All kinds and phases are retained here, including recorder Boundary events. Repeated contexts share one row; raw streams retain them all.

| Owner | Location | Kind | Phase | Exact reason | Parent | Head |
| --- | --- | --- | --- | --- | ---: | ---: |
| No recorded sites | | | | | 0 | 0 |

Declaration unit statuses:

| Side | Name | Unit | Status | Instantiated |
| --- | --- | --- | --- | --- |
| parent | orPredicate2 | compiler/core.ts:2462:1 | deferred_generic_instantiation | False |
| parent | orPredicate3 | compiler/core.ts:2466:1 | deferred_generic_instantiation | False |
| parent | or | compiler/core.ts:2470:1 | deferred_generic_instantiation | False |
| parent | or | compiler/core.ts:2472:1 | deferred_generic_instantiation | False |
| parent | or | compiler/core.ts:2474:1 | deferred_generic_instantiation | False |
| head | orPredicate2 | compiler/core.ts:2462:1 | deferred_generic_instantiation | False |
| head | orPredicate3 | compiler/core.ts:2466:1 | deferred_generic_instantiation | False |
| head | or | compiler/core.ts:2470:1 | deferred_generic_instantiation | False |
| head | or | compiler/core.ts:2472:1 | deferred_generic_instantiation | False |
| head | or | compiler/core.ts:2474:1 | deferred_generic_instantiation | False |


Full streams: [A parent](A-parent-core.jsonl), [A head](A-head-core.jsonl), [B parent](B-parent-core.jsonl), [B head](B-head-core.jsonl). Paired ledgers: [A](A-core-pair.json), [B](B-core-pair.json).

## Named checker callers and expected moves

The completed parent A entry census stops `createTypeChecker` before these nested sites. It stops at `checker.ts:2969:55`: `a value of type ParameterPropertyDeclaration`. Focused full-mode replay observations retain the complete entry graph and use the existing replay selector. The serializer is additionally observed through the real nongeneric `symbolTableToDeclarationStatements` initializer context, without forcing a generic declaration.

The focused observations are unchanged on parent and head in both pairings. The hoped-for serializer `?.` wall is not reached: the caller prologue refuses `serializePropertySymbolsForClassOrInterface` before that body can be observed. Of the six requested findAncestor source sites, only 4767, 4778 and 50073 have the #448a51k refusal as their enclosing statement stop. Site 4748 stops earlier on a source-slot certificate, 48433 on `AliasDeclarationNode` in its prologue, and 50082 on `reading result`. These earlier observations must not be relabeled as six independent findAncestor failures.

### Exact named-site observations

These use the standard full-mode replay selector. A failed statement can contain more than one call; the ledger identifies the enclosing statement stop, not independent execution of every argument. Unit-prologue failures and unreached/deferred bodies are stated explicitly.

#### Pairing A

| Requested checker line | Observed site | Full-mode unit | Parent | Head |
| ---: | --- | --- | --- | --- |
| 4748 | compiler/checker.ts:4749:17 | resolveExternalModule (compiler/checker.ts:4731:5) | enclosing statement: NotYet: writing a checked object field without its source-slot type certificate | enclosing statement: NotYet: writing a checked object field without its source-slot type certificate |
| 4767 | compiler/checker.ts:4768:21 | resolveExternalModule (compiler/checker.ts:4731:5) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-A/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-A/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) |
| 4778 | compiler/checker.ts:4779:21 | resolveExternalModule (compiler/checker.ts:4731:5) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-A/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-A/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) |
| 10911 | compiler/checker.ts:10912:29 | symbolTableToDeclarationStatements (compiler/checker.ts:9347:9) | unit prologue: Refused: overload 1 of serializePropertySymbolsForClassOrInterface result TypeElement[] cannot be served by implementation result (ClassElement &#124; TypeElement)[] | unit prologue: Refused: overload 1 of serializePropertySymbolsForClassOrInterface result TypeElement[] cannot be served by implementation result (ClassElement &#124; TypeElement)[] |
| 48433 | compiler/checker.ts:48435:47 | checkAliasSymbol (compiler/checker.ts:48396:5) | unit prologue: NotYet: a value of type AliasDeclarationNode | unit prologue: NotYet: a value of type AliasDeclarationNode |
| 50073 | compiler/checker.ts:50075:29 | getSymbolOfNameOrPropertyAccessExpression (compiler/checker.ts:49990:5) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-A/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-A/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) |
| 50082 | compiler/checker.ts:50084:39 | getSymbolOfNameOrPropertyAccessExpression (compiler/checker.ts:49990:5) | enclosing statement: NotYet: reading result | enclosing statement: NotYet: reading result |

#### Pairing B

| Requested checker line | Observed site | Full-mode unit | Parent | Head |
| ---: | --- | --- | --- | --- |
| 4748 | compiler/checker.ts:4751:17 | resolveExternalModule (compiler/checker.ts:4733:5) | enclosing statement: NotYet: writing a checked object field without its source-slot type certificate | enclosing statement: NotYet: writing a checked object field without its source-slot type certificate |
| 4767 | compiler/checker.ts:4770:21 | resolveExternalModule (compiler/checker.ts:4733:5) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-B/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-B/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) |
| 4778 | compiler/checker.ts:4781:21 | resolveExternalModule (compiler/checker.ts:4733:5) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-B/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-B/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) |
| 10911 | compiler/checker.ts:10914:29 | symbolTableToDeclarationStatements (compiler/checker.ts:9349:9) | unit prologue: Refused: overload 1 of serializePropertySymbolsForClassOrInterface result TypeElement[] cannot be served by implementation result (ClassElement &#124; TypeElement)[] | unit prologue: Refused: overload 1 of serializePropertySymbolsForClassOrInterface result TypeElement[] cannot be served by implementation result (ClassElement &#124; TypeElement)[] |
| 48433 | compiler/checker.ts:48437:47 | checkAliasSymbol (compiler/checker.ts:48398:5) | unit prologue: NotYet: a value of type AliasDeclarationNode | unit prologue: NotYet: a value of type AliasDeclarationNode |
| 50073 | compiler/checker.ts:50077:29 | getSymbolOfNameOrPropertyAccessExpression (compiler/checker.ts:49992:5) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-B/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) | enclosing statement: Refused: overload 1 of findAncestor result T &#124; undefined cannot be served by implementation result Node &#124; undefined (return at /tmp/branded-B/src/compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked) |
| 50082 | compiler/checker.ts:50086:39 | getSymbolOfNameOrPropertyAccessExpression (compiler/checker.ts:49992:5) | enclosing statement: NotYet: reading result | enclosing statement: NotYet: reading result |


The raw JSON [named-site attribution](named-sites.json) retains the matching Boundary range and exact underlying finding for every quoted statement/prologue stop. [A observing selectors](A-observing-sites.json) and [B selectors](B-sites.json) preserve requested coordinates, actual coordinates, lexical owners and call text. The parent apply shifts the supplied checker lines by 1 or 2; the B adaptation shifts them by 3 or 4.

## Time and peak memory

Wall seconds are the wrapper observation, with up to one second of polling resolution. Peak RSS uses Linux `getrusage(RUSAGE_CHILDREN)` in KiB; each census is one Go child. CPU user/system time is retained separately. Runs overlap as described in [the concurrency note](concurrency-note.txt); these single observations are not performance benchmarks. Completed short runs are reused by matching command and compiler-binary hash, not rerun.

| Pairing | Scope | Compiler | Child exit | Wall seconds | Peak RSS MiB |
| --- | --- | --- | ---: | ---: | ---: |
| A | entry | parent | 0 | 2480.032 | 7980.3 |
| A | entry | head | running | pending | pending |
| A | core | parent | 0 | 31.073 | 880.4 |
| A | core | head | 0 | 35.379 | 913.6 |
| B | entry | parent | running | pending | pending |
| B | entry | head | running | pending | pending |
| B | core | parent | 0 | 34.144 | 863.0 |
| B | core | head | 0 | 33.069 | 871.6 |

Every focused run, apply, build and preliminary failure is included in [all run resources](resources.md) ([JSON](resources.json)). Each `*.status.json` records the exact child command and exit; each `*.time.json` records peak RSS and CPU time. Logs are files and were not piped through head or tail.

## Setup, verification and limits

Setup was run with `/workspace/adamic-tools/env.sh`; successful [setup retry](setup-retry.log) used Go 1.27.1, Node 24.19.0, clang 20.1.8 and the pinned TypeScript 6.0.3 API ([npm ci](npm-ci.log)). [Build environment](build-environment.json), [parent workspace](parent-workspace.txt), [head workspace](head-workspace.txt), and the generated guarded [overlays](overlays/) are retained. All Go evidence filenames end in `.go.txt`.

The existing pairing suite passed all 15 tests, including normalization and regression mutants ([log](pair-tests.log)). Source hashes are checked before and after each completed pairing. The runtime no-output loader and lowerer assertions execute in every census. No native acceptance, backend agreement or full gate is claimed.

Preliminary infrastructure failures are preserved: the initial A apply lost its optional-declarations Node child to SIGKILL with a cgroup OOM kill recorded; both concurrent census builds were terminated and A was reapplied successfully from a clean tree. An initial scratch-workspace replacement error and a missing bridge directory were corrected before successful builds. Setup was retried after restoring an omitted stage1 package. These are not compiler regressions. No later OOM kill has been observed.

To reserve RAM, unused scratch checkout directories, the upstream source clone after both applies, and the upstream `tests` directories in the adapted trees were removed. No measured source file was dropped or changed; all 81 source hashes are retained. Completed guarded streams remain the evidence.

Checkpoints use separate suffixed branch refs so the final requested branch can contain exactly one commit directly on b5c0aa64 without force-pushing. [Checkpoint records](checkpoints.json) retain the SHAs. No main branch is pushed.
