# Price sys public — measurement checkpoint

| Site | Road | Target gone | New findings | Declaration diff lines | Calls added per lane run |
|---|---|---|---|---|---|
| 3 | site-local (readonly view) | Yes | 0 | 0 | 0; final candidate oracle pending |
| 3 | declaration | Yes | 0 | 2 typescript.d.ts; 0 tsserverlibrary.d.ts | Erasure/count verification pending |
| 10 | site-local (fresh host) | Yes | 0 | 0 | Counting pending |
| 10 | declaration (ModuleResolutionHost) | No | 0 | 6 adapted / 2 upstream; 0 server shim | 0 (erased) |

Baseline: adamic `8e3a784aa5a601f57342c07b68f145fe56252c21`, detached private worktree. Upstream v6.0.3 confirmed at `050880ce59e30b356b686bd3144efe24f875ebc8`. Recursive shallow submodules initialized; stage3/api npm ci succeeded. Baseline apply and census binary preparation are running. No adaptation or production compiler change is included.

Upstream sys.ts:1380 is in `patchWriteFileEnsuringDirectory`, not the missing-entry watcher. It captures sys.writeFile, replaces sys.writeFile, and subsequently calls the captured function with `.call(sys, ...)`. An arrow reading the same object's writeFile at call time would recurse. There is no identity comparison of originalWriteFile in this function. Census and lane conclusions remain unmeasured.

## Upstream declaration experiment

Pristine upstream `npm ci` and `node_modules/.bin/hereby dts` succeeded. Each public member conversion was built separately, resetting source between variants. The generated `typescript.d.ts` has exactly one removed and one added line per road; `tsserverlibrary.d.ts` is byte-identical (empty diff). Exact diffs are adjacent to this report. These are pristine upstream measurements, not yet the adapted-tree declaration builds or 79-root checks.

The resolved upstream reader inventory is in `readers-upstream.json`: 526 explicit member accesses/destructuring entries, including homonyms; 35 resolve directly to the two requested public member declarations. This inventory includes harness and src test runners. Dynamic computed keys and structural implementations require further audit; this checkpoint does not claim a complete reader proof.

## Adapted baseline control

Baseline apply exited 0. Coordinates match the request: sys.ts:1385:31 and sys.ts:1612:77. Stock tsc 6.0.3 CLI passes the 79 compiler roots (zero diagnostics) with pinned Node ambient types. An initial stock API setup omitted those ambient types and was corrected; its missing-Node diagnostics are not compiler regressions.

The full single-file CommonJSBundle census exited 0 in 350.052 seconds, using LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1 and LATENT_ONLY_FILE selecting the complete sys.ts. It reports zero checker diagnostics and checker_rejected:false. Its historical generic measurement label does not override that header. Both requested refusal sites are present. There are 185 raw finding events, 59 unique non-Boundary findings: 41 NotYet and 18 Refused. The control stream, timings, apply log, exact tsc config/result and source manifest are retained. This is a census, not successful native compilation.

`READERS.md` lists the 35 resolved public-member references on the adapted baseline; the full 526-entry inventory includes homonyms and the harness/tests.

## Site 3 local form

Private candidate: `const originalWriteFile = sys.writeFile.bind(sys);`; the subsequent `.call(sys, ...)` remains unchanged. Its complete sys.ts census exits 0 (228.687 seconds): 58 unique findings, 41 NotYet and 17 Refused. Exactly the target refusal is gone; no finding is new or changed. Stock tsc over 79 roots exits 0. The one-line private candidate diff and full raw census are retained.

The source captures writeFile once, immediately replaces sys.writeFile, and later invokes originalWriteFile.call(sys, ...). There is no identity comparison or other escape of originalWriteFile. A live-receiver arrow would look up the replacement and recurse. Bind captures the function and receiver at the original read point, creates a new function identity and performs a bind call then. The witness retains receiver-sensitive calls after a later replacement and confirms the identity difference and naive-arrow recursion. This is a measured candidate, not a proposed adaptation. The lane count and broader observation review remain pending.

## Site 3 public declaration form

Changing System.writeFile from method syntax to a function property removes exactly the same target refusal. Full sys.ts census exits 0 in 165.422 seconds: 58 unique findings, 41 NotYet and 17 Refused; no new or changed findings. Stock tsc exits 0 over 79 roots. Adapted-tree npm ci and hereby dts both exit 0. The exact adapted generated declaration diff is adjacent to the report: one removed and one added typescript.d.ts line; no tsserverlibrary.d.ts change.

The initial counting control was disqualified because the scratch-copy helper omitted tracked test-fixture node_modules directories. A full control rerun and fresh candidate counting copies are queued with those fixtures preserved. No per-lane call total is claimed from the disqualified run.

## Site 10 local form

The final fresh host forwards fileExists, readFile, directoryExists, realpath, getCurrentDirectory and getDirectories through receiver-preserving arrows, and copies useCaseSensitiveFileNames. The exact private diff is retained. realpath is undefined when absent; the current adapted public interface already admits this. Full sys.ts census exits 0 in 138.431 seconds: exactly the target is gone, no new or changed finding (58 unique: 41 NotYet, 17 Refused). Stock tsc exits 0 over all 79 roots. Source bytes remained unchanged through the census.

An earlier conditional-spread trial also removed the target but introduced one refusal at the partial realpath object (optional trace absent from a structural source that can hide fields). Its raw stream is retained and is not the admitted road being priced.

The adapted ModuleResolutionHost.directoryExists is already an optional property whose function type is indexed from a method. Both its plain-arrow conversion and the alternate public System.directoryExists method-to-property conversion are being measured to disambiguate the declaration road. Counting is being restarted in complete-fixture copies after the source measurements, because a simultaneous compiler scan contributed to a 40-second hook timeout in a disqualified control.

## Value-identical site 3 local road

The primary local road is now a readonly view, not bind:

```ts
const writeFileView: { readonly writeFile: System["writeFile"]; } = sys;
const originalWriteFile = writeFileView.writeFile;
```

This aliases the same receiver and reads the exact original function once. The subsequent reassignment and originalWriteFile.call(sys, ...) are unchanged. It adds no calls, wrapper, allocation or function identity; its only extra runtime operation is a local reference assignment. The witness covers a getter, receiver-sensitive function, later replacement, and an own .call override. All remain identical to control. Full sys.ts census exits 0 in 107.407 seconds: 0 new, exactly target gone, 0 changed; 58 unique (41 NotYet,17 Refused). Actual stock tsc over 79 roots exits 0.

Bind is retained as a rejected trial. Although its census removes the target, a function with its own .call distinguishes it from the original; own-call-witness.json records that counterexample. It is not the value-identical road being priced.

## Completed site 10 declaration measurements

The plain-arrow ModuleResolutionHost.directoryExists property leaves all 59 findings unchanged, including the target. Its generated adapted typescript.d.ts replaces five indexed-method-property lines with one arrow-property line; pristine upstream replaces one method line with one property line. The alternate System.directoryExists property also leaves 59 findings, changing the target's reason to ordinary mutable widening: the target can still write function-or-undefined where a required function is read. This is one changed finding, not a removed refusal. Its public diff is one removed and one added line in both pristine upstream and adapted dts. All actual stock 79-root checks pass. Every tsserverlibrary.d.ts is byte-identical. Stock TypeScript erasure gives byte-identical JavaScript for all three declaration experiments, hence zero inserted executable calls.

## Valid counting control

The isolated complete-fixture control meets the exact lane contract: 106,366 passing, one expected API failure, zero pending, sole baseline diff api/typescript.d.ts. It counted 2,200 captures and 4,627 callback invocations at site 3. Site 10's nodeSystem.require resolution entry counted zero. The final local candidate oracle runs are pending. Final counting copies exclude only root build/dependency directories, audit all 32 affected tracked fixture files, and run sequentially without census/compiler competition.
