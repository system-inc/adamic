# Adamic receiver freshness census for #n7mgy3s

## 1. New or changed findings

None. The base and head entry streams each contain 3,255 finding records. Every record pairs on kind, phase, reason and message with positions stripped. The full records, including positions, also match. There are zero new, gone, changed or reached findings, where reached means a later stop newly hit in a unit that was already blocked. The repository's `stage3/census/pair` also reports zero first-stop changes and zero regressions.

## 2. Units that lowered at base and stop at head

None. On each revision, 3,274 entry units lowered without a body stop, and 3,272 of those also cleared the scoped tail. The two tail stops have the same units and findings on both revisions. The ordinary root census accepted `src/compiler/corePublic.ts` and `src/compiler/hostErrors.ts` on both revisions.

## 3. Wrapped function-expression properties using `this`

None. A TypeScript 6.0.3 AST walk covered 79 files in adapted `src/compiler` and two in `src/tsc`. It examined object-literal property initializers through parentheses, non-null expressions, `as`, angle-bracket assertions and `satisfies`. It found zero function-expression initializers wrapped in an assertion or `satisfies`.

## 4. First stops across 79 roots

No root's first stop moved. The ordinary per-root records match apart from timing.

| Roots | Count | Ordinary outcome at both revisions |
| --- | ---: | --- |
| `corePublic.ts`, `hostErrors.ts` | 2 | Accepted |
| `debug.ts` | 1 | NotYet at `semver.ts:71:9` |
| `diagnosticInformationMap.generated.ts` | 1 | NotYet at `commandLineParser.ts:293:35` |
| `tracing.ts` | 1 | Refused at `utilitiesPublic.ts:779:16` |
| `types.ts` | 1 | NotYet at `diagnosticInformationMap.generated.ts:13:45` |
| Other roots | 73 | NotYet at `tracing.ts:76:15` |

The meter classified `debug.ts`, `diagnosticInformationMap.generated.ts` and `types.ts` as root-order artifacts at both revisions. All three readiness probes threw under Node v24.19.0. For each of these roots, the meter reports the same next stop at both revisions, `Refused` at `src/compiler/binder.ts:496:42` with the reason “a cast the runtime can't check.”

## Pins and evidence

- Both branches were fetched by name. The measured base was `aa90e3fab745620d0df04b52b3aa2d2bd8db23b4` from `main`. The measured head was `5fe496b1d24595294b6171b09e83858d817b86c3` from `typescript/this-type-getter-fresh-2`, with that base as its sole parent.
- The Go 1.27.1 builds used each revision's own latent and entry overlays, ordinary census, and `scan-all` sources. The measurement commands used the meter's entry overlay and ordinary per-root census. Both revisions pinned cohere at `5360753e19e4417675a90e92a1823bea66b13918` and its TypeScript checker at `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- The shared source tree was produced by base's `stage3/apply.sh` after `git submodule update --init --recursive`. Upstream was cloned with `--depth 1 --branch v6.0.3` and resolved to `050880ce59e30b356b686bd3144efe24f875ebc8`. The sorted file checksum manifest under adapted `src` hashed to `af90a93765451d97c5962209486b0b2263f2680fba9a2579e73891d8916064d1` before and after measurement.
- The entry runs used `CENSUS_MODULE_FORMAT=CommonJSBundle`, `LATENT_FULL=0`, `LATENT_TAIL=1` and `LATENT_ASSERT_NO_OUTPUT=1`. Both measured 81 reachable files and recorded zero checker diagnostics. The directory runs used `CommonJSBundle` and retained all 79 source-root rows, three JSON input rows and the meter's aggregate row.
- Commit `d2ae83ef4` recorded a preliminary header produced against an interrupted adaptation. Commit `ecedad876` withdrew it. The evidence below comes from the completed adaptation, with each completed pass pushed before the next requested pass started.

| Pass | Base | Head |
| --- | --- | --- |
| Entry census | [Stream](evidence/base-entry.jsonl.gz) · [Unit timings](evidence/base-entry-times.jsonl.gz) | [Stream](evidence/head-entry.jsonl.gz) · [Unit timings](evidence/head-entry-times.jsonl.gz) |
| Per-root census | [Stream](evidence/base-roots.jsonl.gz) | [Stream](evidence/head-roots.jsonl.gz) |
