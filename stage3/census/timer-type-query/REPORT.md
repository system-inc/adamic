Both targets gone; **0 new findings**. Exactly two unique NotYet findings disappeared; no other finding changed.

# Timer type-query census — task #7yz6rtp

## Pins and input identity

- Base: `1df98d9828a4da2f00d31b50ecfa516e2c3d0a44`.
- `runtime-performance/timer-type-query` resolved from the remote branch to `9fbb475c55dac1346c716eb79d6c080a6b0caa3d`. Its sole parent is the base above; ancestry succeeds. It is also an ancestor of the observed remote main `9c30403606a7a07b33b64e046c7bc0fb63385646`.
- Both commits have the identical `stage3` Git tree: `8ca56fd05798045489d1dd7c7c22699c83bd83d5`. One completed adapted tree was built at the base using `stage3/apply.sh` and reused for both runs.
- Upstream TypeScript was cloned with `--depth 1 --branch v6.0.3`; HEAD confirmed `050880ce59e30b356b686bd3144efe24f875ebc8`. `ADAMIC_TYPESCRIPT_SOURCE=/tmp/timer-type-query-census/TypeScript`.
- Adapted `src/compiler/sys.ts` SHA-256: `30876cbae0e2cb19eae2525cbed919abdba86fac921ecc22fa6aaad0441a3398`.
- Both fresh detached worktrees ran `git submodule update --init --recursive --depth 1`: cohere `5360753e19e4417675a90e92a1823bea66b13918`, nested TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- `npm ci` ran in the base `stage3/api`; only the base performs adaptation. All writes, caches and temporary files were under `/tmp/timer-type-query-census`. The shared checkout was unchanged. Sparse checkout omitted unused historical assets and submodule test baselines to fit scratch capacity; compiler and adaptation sources were unchanged.

## Method and execution

At each pinned checkout, following `stage3/census/latent/README.md`:

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" "$SCRATCH/$REF-overlay"
gofmt -w "$SCRATCH/$REF-overlay/"*.go
go build -buildvcs=false -overlay="$SCRATCH/$REF-overlay/overlay.json" \
  -o "$SCRATCH/$REF-census" ./stage3/census/latent/tool
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
LATENT_ONLY_FILE="$SCRATCH/adapted/src/compiler/sys.ts" \
CENSUS_MODULE_FORMAT=CommonJSBundle \
  "$SCRATCH/$REF-census" "$SCRATCH/adapted/src/compiler" "$SCRATCH/$REF.jsonl"
```

`SCRATCH=/tmp/timer-type-query-census`; `REF=base` or `head`. The source-directory argument loads the complete compiler program, with full-mode file selection restricted to the exact absolute sys.ts path. Each binary was built independently with its own overlay. `LATENT_PROGRESS_FILE` was also set separately for observational progress. The local scratch runner `/tmp/timer-type-query-census/run-census.py` records the exact environment, wall time, exit and independent 3,600-second bound; its scratch prefix is specific to this run.

`PATH` included `/workspace/adamic-tools/bin` and `/workspace/adamic-tools/go/bin`. `TMPDIR` and `GOTMPDIR` were `$SCRATCH/tmp`, `GOCACHE=$SCRATCH/gocache`, `GOMODCACHE=$SCRATCH/gomodcache`, `ADAMIC_BUILD_CACHE_DIR=$SCRATCH/cache`, `STAGE3_CACHE=$SCRATCH/stage3cache`; npm used `$SCRATCH/npmcache`.

| Run | Wall seconds | Exit | Timeout | UTC start → end |
| --- | ---: | ---: | --- | --- |
| Base | 99.086700 | 0 | No | 2026-11-14 17:43:42.214433 → 17:45:21.301003 |
| Head | 100.417606 | 0 | No | 2026-11-14 17:45:21.310768 → 17:47:01.728151 |

Both streams contain a measurement header, exactly one sys.ts file record, and a terminal relation report. Both no-output guards passed. Both headers report `latent_mode=full`, `module_format=CommonJSBundle`, `checker_rejected=false`, and zero checker diagnostics. The tool's fixed label says “measured on a checker-rejected program”; that label does not override the actual clean header on this adapted input.

## Complete delta

Pair/deduplication key: **(phase, kind, where, reason)**, as requested. Paths below abbreviate the identical absolute prefix `/tmp/timer-type-query-census/adapted/src/compiler/`.

**Gone (2):**

| Phase | Kind | Where | Exact reason |
| --- | --- | --- | --- |
| refusal_scan | NotYet | sys.ts:1500:57 | `node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot` |
| refusal_scan | NotYet | sys.ts:1598:57 | `node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot` |

Both had attempting unit `sys.ts:1473:5`. Each original text is `WHERE: stage 0 can't lower REASON yet`, with WHERE and REASON exactly as listed above. The complete raw finding objects remain in local scratch evidence.

**New: 0. Changed: 0.** No surviving paired finding changed any field. After deleting just the two target observations from the base, the entire decoded file record equals the head record, including ordered raw findings, declarations and units. Headers and terminal relation reports also match exactly. Thus this result is stronger than equality of deduplicated totals.

| Phase | Kind | Base unique | Head unique | Delta |
| --- | --- | ---: | ---: | ---: |
| refusal_scan | NotYet | 3 | 1 | -2 |
| refusal_scan | Refused | 8 | 8 | 0 |
| lowering | NotYet | 38 | 38 | 0 |
| lowering | Refused | 10 | 10 | 0 |
| lowering | Boundary | 55 | 55 | 0 |
| **All** | | **114** | **112** | **-2** |

Total unique NotYet: **41 → 39**; unique Refused: **18 → 18**. Raw finding observations (including repeated attempt contexts and Boundary rows): **185 → 183**. Both files catalogue 136 units: 133 attempted and 3 deferred generic instantiations; unit metadata is identical.

**First refusal-scan site in sys.ts order is unchanged: `790:87`**, phase `refusal_scan`, kind `Refused`, reason `a value of type Map<Path, string[]> seen as InvokeMap, which can write true | string[] where string[] is read`. It is both the first emitted scan finding and the earliest scan finding by source position on both sides; its full finding object is unchanged.

## Validation limits and preparation incident

The README's standard `audit.py` passed on both binaries. The additional `full_audit.py` failed on both: its continuation fixture expected `{'a string as a condition', 'a number as a condition', 'var'}` and observed only `{'var'}`. Base used the default format for this audit; head used CommonJSBundle. These failures are preserved, not repaired. They limit broader confidence in that fixture/coverage contract; the measured same-input delta above is an observation ledger, not a successful compilation or proof of exhaustive blockers. The census emits no IR and invokes no backend.

An initial adaptation attempt failed because my sparse-checkout initialization briefly removed an adaptation script while apply was running. The partial output was discarded, the inputs restored, and apply rerun successfully with stable inputs. Only the completed rebuild supplied either measurement; the failed preparation log is retained. There was no compiler change or adaptation change.

## Local evidence

Raw JSONL, timings, comparison objects, run/build/overlay logs, audit results, and both adaptation logs are retained locally under `/tmp/timer-type-query-census/`. Only this report is published. Raw outputs and detailed repository logs were excluded after automatic approval review rejected their inclusion as outside the report publication authorization.

Uncompressed raw SHA-256:

- Base: `1049850db6d6bf61f8db0a973ab75585805b01f8c4b2ff3c59fe1f3a4d5bfcd8`.
- Head: `c8bc9016d3528fbdb3dfb241f18a7a6462b98744d204bf2ee09203e31ba8f922`.

Measurement binaries SHA-256:

- Base: `75120eae94dfb90e62a09bdc67bfd94d31abc7b4079afedb7600e4951bfc2e05`.
- Head: `924c07511d909cb1d696ef2b23a4492e3712cea56223c6098db7e9a192ea762d`.
