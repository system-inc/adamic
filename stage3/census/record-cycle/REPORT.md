Units affected: **unknown — the whole-program graph-facts prerequisite stopped before the unit census.** First site in each affected unit: **unavailable**. This is not a measured zero.

# Record-cycle measurement attempt at 1c7f96e6

Head: `1c7f96e65693bdb47ce4c8fa0cf9fc358f957a3b` (the two commits after main `55c375d9d245b580c010c3e0b5659f2849f24fb6`, task #t5q4prw). Measurement only; no new adaptation or production compiler change. The requested all-unit census is **not complete**. This report preserves a validated prototype, its prerequisite stop, and the independent ordinary-root table. It does not establish that another scratch measurement design cannot compute the missing facts.

## Prerequisite stop

whole-program graph facts unavailable before record check: /workspace/scratch/record-cycle/adapted/src/compiler/sys.ts:1385:31: Adamic 0.1 refuses a method read as a value (writeFile would lose its object, and this with it); call it in an arrow that keeps the object: (path, data, writeByteOrderMark) => sys.writeFile(path, data, writeByteOrderMark) (unbound-method)

Quoted prepared source (`src/compiler/sys.ts:1385`):

```typescript
    const originalWriteFile = sys.writeFile;
```

The prototype obtains graph facts by retaining the real `cycleFinder` only when whole-program lowering reaches the new record-check point. It does not disable the preceding refusal scan. Consequently this preparation repeats the known production-tail blind spot on tsc: the refusal above prevents complete facts, so the unit-end observer never runs on tsc. No number of affected units, site/type/reach/source inventory, or main all-unit zero is claimed. The next measurement design needs a validated way to populate whole-program shapes, captures, local types and graph selection through the latent overlay, without treating skipped lowering as absent edges. No source adaptation was added to bypass this stop.

## Seam and controls

The same `prepare-seam.py` generator and identical dedicated driver build on both commits. Exact sources are in `head/` and `main/`; the scratch patch is `SEAM.patch.go.txt`. These are archival `.go.txt` files, not active compiler code. The generator runs after each commit’s normal `stage3/census/latent/make_overlay.py`; run its `importclean` on the restored lower file before building. Its private `LowerWithOptions` remains disabled and returns nil IR. The dedicated preparation API returns only an error; it exposes no IR or backend path.

For the head unit-end observation, each recorded literal is presented alone to the retained finder and the unmodified production `cyclicRecord` is called. The original record list is restored afterward. Main never receives a copied implementation of the missing check and records zero hits. The whole-program fact gate is deliberate: missing graph selection or capture edges must not become a zero. This prototype retains the existing attempted-unit eligibility and cannot count literals beyond failed ordinary lowering.

| Same-driver fixture | Head calls | Head hits | Main hits |
| --- | ---: | ---: | ---: |
| Record reaches its own type | 2 | 2 | pending |
| Graph with acyclic string record | 3 | 0 | pending |

The positive fixture’s type path is `Record<string, Node>` values → `Node` → `Node.box: Box` → `Box.r: Record<string, Node>`. The real allocation rule also judges the empty record in `make`, since it uses the allocation’s record and element types. Its two head hits prove the unit-end seam invokes the real check. The negative fixture has a separate `Node[]` graph cycle, while each record has string elements and cannot reach a record type. Fixture text is preserved as `cycle.a.txt` and `acyclic.a.txt`.

## Latent-tail audit

`make_overlay.py` replaces production lowering’s whole-program body/tail. `pass_ledger.go.txt` marks `findCycles` as `WholeProgram: true`, not `Scoped`. `latentScopedTail` dispatches only entries with `Scoped: true`. The 13 scoped entries are: `forwarderValues`, `finishClassCalls`, `finishAccessors`, `finishToPrimitive`, `checkRecordGuardedReads`, `flow.NormalizeAsync`, `asyncEntrySource`, `checkViewOptionalWrites`, `checkHostMemberWrites`, `asyncPayloadFailures`, `checkViewLogicalWrites`, `borrow`, `counters`. Neither `findCycles` nor its `graphTypes` call runs there; `cyclicRecord` is therefore absent. A plain latent zero cannot answer this question.

The #3zex15c precedent (`typescript/view-guard-roads-census`, `c13c98f7e920082c5ffc1034b095d4147c38998d`) was read, including `rerun-ed3c0354/METHOD.md`. That observer explicitly did not model whole-program graph membership. This task needs additional graph facts; its method cannot simply be copied with an empty finder.

## CommonJSBundle ordinary 79-root table

0/79 roots completed twice with matching first-stop tuples. 0 completed roots have the new reason as their first stop. Table incomplete; remaining roots have no result yet.

- `corePublic.ts`: pending.
- `hostErrors.ts`: pending.

The production driver is built without any overlay from the exact head. The scripts from `typescript/next-stops-8e3a784a` (`d8a5019190cf3c176f1b11385d6437c06f04ba4e`) are retargeted here; the group executor is reduced from three workers to one. Each root is inspected twice by the unchanged ordinary driver. The separately running main 79-root table is not reproduced or claimed here. Roots stopped before the tail do not supply negative record-cycle evidence.

| # | Root | First outcome | Site | Reason |
| ---: | --- | --- | --- | --- |

## Reproduction and receipts

Scratch root: `/workspace/scratch/record-cycle`. Fresh detached worktrees were created at the exact head and main, then recursive depth-1 submodules initialized. Both use cohere `5360753e19e4417675a90e92a1823bea66b13918` and TypeScript-Go `d92d9bfee114c80be2c375d72edae966176e3a4f`. Upstream was cloned with `--depth 1 --branch v6.0.3` and verified as `050880ce59e30b356b686bd3144efe24f875ebc8`. `npm ci` ran in both `stage3/api` directories. Go 1.27.1 is `/workspace/adamic-tools/go/bin/go` (the default `/usr/bin/go` is unrelated and rejected `go run`; its failed receipt is retained).

`GOTMPDIR`, `TMPDIR` and `GOCACHE` point inside scratch. `ADAMIC_TYPESCRIPT_SOURCE` names the verified upstream clone; `CENSUS_MODULE_FORMAT=CommonJSBundle`. The head’s existing `stage3/apply.sh` prepared `adapted/`; stage3 source recipe inputs are byte-identical between head and main. No additional adaptation was made. Raw source hashes and sizes are in scratch `source-manifest.json`, itself hashed in `RAW-MANIFEST.json`.

Raw streams, logs, binary outputs and generated overlays remain in scratch; `RAW-MANIFEST.json` publishes sha256 and byte size for each measurement raw file and overlay source, and `RECEIPTS.json` records commands, exit status, elapsed time and child peak RSS for bounded wrapper jobs. All measurement wrappers use a 3500-second timeout; individual ordinary roots use 1780 seconds. Source preparation completed in under 10 minutes. An ordinary build was briefly launched while apply was active, then terminated to restore one-heavy-job execution; it supplied no measurement evidence. The first seam build failed on two unused imports after disabling the output API; `importclean` corrected the scratch overlay and the subsequent build succeeded.

Run from the corresponding scratch worktree after sourcing `env.sh`: generate its normal latent overlay; run `prepare-seam.py REPO OVERLAY`; run `go run ./stage3/census/latent/importclean OVERLAY/internal_lower_lower.go`; then build `./stage3/census/latent/tool` with `-buildvcs=false -overlay=OVERLAY/overlay.json`. Invoke that binary with an entry and output JSONL. Run each ordinary piece with `python3 measure.py table-FIRST-LAST python3 table/group.py FIRST LAST` (at most ten roots), checking exit receipts before continuing. No benchmark or successful native tsc compilation is claimed.
