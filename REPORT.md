# Paired Adamic census for #s8gy89r, family 1, cause U2

Base `1a6b689e009ad27575887a18c938a6342a6cf2ca` is the sole parent of head `793e2b25d267969957ecbb7ebfb0a6d551bade13` from `platforms/enum-tag-kind-helper`. The named `main` fetch resolved to `a59db917a3a401bd6712a703b94c821ff4864440`. The latent, entry, ordinary census and scan-all executables were built from their respective checkouts after recursive submodule initialization. The head also changes two fast-gate files.

Stage 3 used TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. Base and head applied the same keyed product after submodule initialization. A recursive file comparison outside Git metadata found zero differences, and all 312 checkpoint refs matched. The last checkpoint was tree `8f61e6699b44dd0e54523a760720a37acdd8cecb`. Both measurements use the base-applied tree.

The latent runs use `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_FULL=0`, `LATENT_TAIL=1`, `LATENT_ASSERT_NO_OUTPUT=1` and `GOMAXPROCS=4`. [Evidence](evidence/) is pushed after each pass. Both latent pairs are complete. The root census is pending.

## 1. Family 1 site counts

Exact reason is “an object refinement using an open numeric enum as a literal tag”. Raw positions and distinct TypeScript property-read identities come from `stage3/census/pair`.

| Pass | Base distinct | Base raw | Head distinct | Head raw |
| --- | ---: | ---: | ---: | ---: |
| Compiler directory | 503 | 671 | 493 | 661 |
| `tsc.ts` entry | 503 | 671 | 493 | 661 |

Both pairs have the same **10 cleared reads and zero reads that move to a later stop**. The 110-site historical count is recorded at `dc9d089c`. That abbreviated commit could not be resolved, and the per-site inventory was unavailable. Overlap with that exact cohort remains unverified. [U1 commit](https://github.com/system-inc/adamic/commit/faf9b5093cfbcf5f4c8e373a3285ffbd268c7fa5)

## 2. Units whose first stop moved

Each latent pair has zero units that now lower whole, zero units that move to a later stop, and zero units that move to an earlier or different stop. This holds for both directly recorded unit stops and the first stop when the source-file refusal scan is attributed to the containing top-level unit.

All ten removed reads belong to `createTypeChecker` at `src/compiler/checker.ts:1490:1`. Its first scan stop stays at `src/compiler/checker.ts:1712:51`. Adamic 0.1 refuses an unproven predicate argument for parameter nodeTest (argument "isExportSpecifier"). Its direct body stop stays at `src/compiler/checker.ts:2972:55`. Stage 0 cannot lower a value of type ParameterPropertyDeclaration, an intersection whose field jsDoc has no representation because the union member JSDocArray is an array with its own member jsDocCache.

The grades use kind, phase, reason and message with message positions stripped, along with the paired source position. Reached is a later finding newly observed in a unit that already had a stop.

| Pass | New | Gone | Changed | Reached |
| --- | ---: | ---: | ---: | ---: |
| Compiler directory | 0 | 10 | 0 | 0 |
| `tsc.ts` entry | 0 | 10 | 0 | 0 |

## 3. Regressions

Each latent pair has zero regressions. Zero units lowered at base and stop at head, zero new findings appeared under any reason, and zero new panics appeared. Every stream recorded zero panics. Findings across all reasons fell from 2,983 to 2,973 in each pair.

## 4. Sites by predicate

Both latent pairs have the same changes. All positions below are in `src/compiler/checker.ts` and have zero remaining findings at the read. Each has the family-1 refusal at base during the refusal scan.

| Predicate | Guard and containing function | Cleared read positions | Sites that move or new stops |
| --- | --- | --- | --- |
| `isFunctionLike` | `31507:13`, `tryGetThisTypeAt` | `31508:90`, `31510:53`, `31510:111`, `31514:67`, `31521:42` | 0 |
| `isFunctionLike` | `31556:13`, `getExplicitThisType` | `31557:59` | 0 |
| `isFunctionLike` | `34636:17`, `getEnclosingClassFromThisParameter` | `34637:59` | 0 |
| `isFunctionLike` | `34648:33`, `getThisParameterFromNodeContext` | `34648:82` | 0 |
| `isFunctionLike` | `49994:17`, `isThisPropertyAndThisTyped` | `49995:70` | 0 |
| `isFunctionLike` | `50285:21`, `ThisKeyword` arm | `50286:61` | 0 |

`isFunctionLikeDeclaration` and every other predicate have zero measured site changes in both latent pairs. Workers' probe count of 16 describes the helper-call class. The measured base-to-head delta is 10.

## 5. Latent timing

Wall time is monotonic process elapsed time. Total CPU is process user plus system CPU. Unit sums come from `LATENT_TIMES` and cover attempted units. Process totals also include loading, checking, source-file scans and stream output.

| Pass | Side | Process wall (s) | Total CPU (s) | Unit wall sum (s) | Unit CPU sum (s) | Timed units |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| Compiler directory | Base | 1,133.808 | 1,333.805 | 717.301 | 832.893 | 4,057 |
| Compiler directory | Head | 1,078.488 | 1,255.163 | 688.142 | 793.536 | 4,057 |
| `tsc.ts` entry | Base | 1,024.967 | 1,191.870 | 659.234 | 757.217 | 4,064 |
| `tsc.ts` entry | Head | 1,048.063 | 1,209.614 | 669.078 | 763.677 | 4,064 |

The compiler-directory head used 55.320 s less wall time (4.9%) and 78.642 s less CPU (5.9%). The entry head used 23.096 s more wall time (2.3%) and 17.744 s more CPU (1.5%). Timed units account for 9.845 s of the entry wall-time increase and 6.460 s of the CPU increase.

The five slowest entry-head units by wall time follow. The largest per-unit wall-time increase was 0.268 s at `src/compiler/factory/emitHelpers.ts:148:1`, from 1.305 s to 1.573 s.

| Unit | Base wall (s) | Head wall (s) | Base CPU (s) | Head CPU (s) |
| --- | ---: | ---: | ---: | ---: |
| `src/tsc/tsc.ts:1:1` | 4.455 | 4.715 | 4.758 | 4.715 |
| `src/compiler/utilities.ts:5484:1` | 2.724 | 2.902 | 2.734 | 3.589 |
| `src/compiler/sys.ts:1451:1` | 2.590 | 2.419 | 2.958 | 2.795 |
| `src/compiler/scanner.ts:1026:1` | 2.201 | 2.128 | 3.026 | 2.935 |
| `src/compiler/parser.ts:505:1` | 1.766 | 1.818 | 2.501 | 2.729 |

## 6. First stops across the 79 roots

Pending.
