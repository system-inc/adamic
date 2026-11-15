# Timer type-query census 4

**Both intended timer type-query findings are gone; new findings: 0. Exactly two findings are gone and none changed.** The freshly built pinned adapted tree locates those expressions at **sys.ts:1501:57 and sys.ts:1614:57**, rather than the requested 1500:57 and 1598:57. The expected unchanged first refusal-scan site at 1387:5 is **not confirmed**: base's first scan finding is 1501:57; head's is 1548:13.

This is measurement and report only. This branch starts at the requested main and adds only this report and its evidence. No compiler or adaptation source was changed.

## Pins and shared adapted input

- Base: `89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825`.
- Head: `fc677461c7d593cb4eb9bfca86102f001a91d6e1`, resolved from remote `runtime-performance/timer-type-query-4`. Its sole parent is exactly base; ancestry therefore holds.
- Both `stage3` trees: `c4981d46fb8998476032918d2e0003964bc60e3d`.
- TypeScript: shallow clone of `v6.0.3`, confirmed `050880ce59e30b356b686bd3144efe24f875ebc8`.
- Recursive submodules: cohere `5360753e19e4417675a90e92a1823bea66b13918`; its TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`.
- Go 1.27.1 linux/amd64; Node v24.19.0; npm 11.9.0.
- Adapted product key: `25e8779881c2b6bf6b03516f33b3e168909fcc0c7e807c4c76a1d62e78a23071`.
- Adapted `sys.ts` SHA-256: `04de8aeb5f32caadd83d2ae6695d40097d7008a50a3a566d983cefd72a381a31` (verified unchanged after both runs). The complete 747-file `src` manifest is in [evidence/adapted-src.sha256](evidence/adapted-src.sha256).

The head diff was read before building: `inTypeQuery(node)` was added at `nodeLibraryRefusal`'s expression gate, before all subsequent host-member use rules, and in `hostBindingRefusal`. The timer helper changed only its explanatory comment. Test and fixture additions are also present in the head. No `stage3` differences exist, so **one successfully built base adapted product was reused for both commits**.

Fresh detached worktrees were created in `/tmp/timer-type-query-census-4/{base,head}`, from a scratch Git repository; each ran `git submodule update --init --recursive --depth 1` and `npm ci --prefix stage3/api`. `ADAMIC_TYPESCRIPT_SOURCE` named the separately cloned pinned source. Caches and temporary build output were directed to scratch; larger scratch storage under `/workspace/timer-type-query-census-4-scratch` was used after `/tmp` exhaustion. Neither shared checkout nor its working files were used for builds.

Setup failures are retained, not counted as census runs: the first cold overlay builds exhausted `/tmp`; both were rebuilt successfully with `go build -p 2` and a larger scratch cache/temp directory. The first adapted attempt was accidentally interrupted by refreshing API dependencies during its execution, causing adaptation 49 to fail to load TypeScript. It was discarded and a fresh standard `apply.sh --build-product` completed successfully. There was no second adapted build for head. The first npm installs preceded setting the scratch npm cache; they were repeated with the scratch cache. These setup mistakes do not affect the successful product or the paired measurement.

## Method and run completion

Each worktree used its unmodified latent README workflow:

```sh
python3 stage3/census/latent/make_overlay.py "$PWD" "$OVERLAY"
gofmt -w "$OVERLAY"/*.go
go build -p 2 -buildvcs=false -overlay="$OVERLAY/overlay.json" \
  -o "$BINARY" ./stage3/census/latent/tool
```

The base adapted tree was built with:

```sh
stage3/apply.sh --build-product /tmp/timer-type-query-census-4/adapted
```

The runs were sequential, with the complete compiler source-directory argument and exact absolute single-file selection:

```sh
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
LATENT_ONLY_FILE=/tmp/timer-type-query-census-4/adapted/src/compiler/sys.ts \
CENSUS_MODULE_FORMAT=CommonJSBundle \
timeout --signal=TERM --kill-after=30s 3600s "$BINARY" \
  /tmp/timer-type-query-census-4/adapted/src/compiler "$OUTPUT_JSONL"
```

| Commit | Start (UTC) | Wall seconds | Exit | Limit |
|---|---|---:|---:|---:|
| Base | 2026-11-15T03:59:59.601445+00:00 | 97.198493 | 0 | 3600 s |
| Head | 2026-11-15T04:01:47.488155+00:00 | 97.146995 | 0 | 3600 s |

Both streams contain a full-mode, CommonJSBundle header, one selected sys.ts file record with 136 units, and a terminal `relation_report`. All non-finding file fields, including unit eligibility and boundaries, are identical. Both headers are identical: **zero checker diagnostics, `checker_rejected: false`**. The measurement tool still prints its fixed historical label “measured on a checker-rejected program”; that label is not an accurate description of these checker-clean inputs. Both no-output guards were enabled; no backend or usable IR output was produced.

`audit.py` passed on each binary, including continuation, attribution, no-output guards and mutants. The optional `full_audit.py` failed identically on each binary: its `three` probe expects string-condition, number-condition and var refusals, while the actual set is only `{'var'}`. The current production source has no matching string/number-condition refusal messages. No audit or compiler was modified to make that check pass. This remains a validation limitation; the real census runs completed normally. Logs are included.

## Complete finding delta

Findings are deduplicated and paired by **(phase, kind, where, reason, message with its leading diagnostic position stripped)**. All kinds and phases were compared, including Boundary records and any trailer findings, with no unit-based exclusion. A changed finding would retain phase/kind/where with a different reason or stripped message; there are none.

| Phase | Kind | Base | Head | Delta |
|---|---|---:|---:|---:|
| refusal_scan | NotYet | 3 | 1 | -2 |
| refusal_scan | Refused | 3 | 3 | 0 |
| lowering | NotYet | 38 | 38 | 0 |
| lowering | Refused | 10 | 10 | 0 |
| lowering | Boundary | 55 | 55 | 0 |
| **All paired findings** | | **109** | **107** | **-2** |

Raw attempt findings: 180 → 178. Phase-aware unique NotYet: 41 → 39. Phase-aware unique Refused: 13 → 13. New: **0**. Changed: **0**.

### Gone 1 — refusal_scan / NotYet / sys.ts:1501:57

Reason:

```text
node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot
```

Position-stripped message:

```text
stage 0 can't lower node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot yet
```

The expression is the `clearTimeout` identifier in `Parameters<typeof clearTimeout>[0]`, in `getNodeSystem`'s object passed to `createSystemWatchFunctions`. **Inside a type query: yes.** Stock TypeScript 6.0.3 reports `Identifier → TypeQuery → TypeReference → IndexedAccessType → Parameter → ArrowFunction`. The arrow's later direct call `clearTimeout(timeoutId)` is a separate expression. This removed diagnostic is not a runtime value use.

### Gone 2 — refusal_scan / NotYet / sys.ts:1614:57

Reason:

```text
node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot
```

Position-stripped message:

```text
stage 0 can't lower node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not (timeout: unknown) => void with no slot yet
```

The expression is the `clearTimeout` identifier in `Parameters<typeof clearTimeout>[0]`, in `getNodeSystem`'s returned system object. **Inside a type query: yes.** Stock TypeScript gives the same ancestry through the arrow parameter. Its later direct call is a separate expression. This removed diagnostic is not a runtime value use.

These are every gone finding. No vanished finding suppresses a real runtime use in this census. Full AST ancestry and source lines are in [evidence/gone-ast.json](evidence/gone-ast.json).

### New

None, in any phase or kind.

### Changed

None, in any phase or kind.

## First refusal-scan site and requested coordinates

The first scan site in sys.ts source order **does change**, because the first base site is one of the two removed timer type queries:

- Base: `refusal_scan / NotYet / sys.ts:1501:57` (Gone 1 above).
- Head: `refusal_scan / Refused / sys.ts:1548:13`, the `watchFile` property. This full finding already existed unchanged in base.

Head first-site reason:

```text
a function taking PollingInterval seen as one taking number | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take
```

Head first-site position-stripped message:

```text
Adamic 0.1 refuses a function taking PollingInterval seen as one taking number | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)
```

At the specifically requested `sys.ts:1387:5`, head's source reads:

```ts
    sys.writeFile = (path, data, writeBom) =>
```

Neither commit records a **refusal_scan** finding there. Both record an identical **lowering / Boundary** there, with reason `failed statement; state rolled back; continued at next statement or declaration`. Its message refers to the identical **lowering / Refused** at `sys.ts:1392:49`, on `originalWriteFile.call(sys, path, data, writeByteOrderMark)`:

```text
Adamic 0.1 refuses inherited library member call read as an own field; prototype members are not stored in an object's shape; call the method on its receiver, or wrap that call in an arrow (unbound-method)
```

The literal requested timer coordinates were not base timer findings in these newly adapted bytes: line 1500 contains the `setTimeout` forwarder; line 1598 contains `disableCPUProfiler(() => process.exit(exitCode));`. The intended two `clearTimeout` type-query expressions are identified above by their exact reason and stock AST. Therefore the **two-expression removal / zero-new claim is confirmed**, while the supplied coordinate and unchanged-first-site expectations are not reproduced by the requested fresh base build. No input adjustment was made to obtain older coordinates.

## Evidence and limits

[verify.py](verify.py) independently checks the archived completed streams, exact pairing delta, identical non-finding metadata, both timer reasons, type-query evidence, first scan site, timing exits, and the absent 1387:5 scan site. Run `python3 stage3/census/timer-type-query-4/verify.py` from a checkout of this report. Compressed raw streams, successful and failed setup logs, audit logs, exact command metadata, provenance, and source hashes are in [evidence/](evidence/).

This is a guarded single-file observation ledger over the complete compiler project. It does not establish whole-program lowerability, runtime behavior, or exhaustive expression coverage beyond the census's recovery boundaries. The failed synthetic full audit is explicitly retained above. No Go files are added under stage3 by this report.
