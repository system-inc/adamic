# Paired overload certificates census — #5fws6q8

Base: `910f21afffd892c747bd82f72229743a899586ea`. Head: `812c342689ded7e1cc8a31c35a86f3546f7b0772`, exactly one commit on base. This report branch changes no adaptation, ledger or internal source.

## Answers for Lower

### 1. Declaration acceptance

**Both complete named sets now pass the declaration census:** arrayFrom overloads **1–2**, and arrayToMap overloads **1–4**. Source accounting is two named sets and six overload declarations; [DECLARATIONS.md](DECLARATIONS.md) lists all ordinals and explains baseline short-circuiting.

The visible base refusals eliminated are arrayFrom overload 1 (U[] versus (T | U)[]) and arrayToMap overload 2 (Map<K,V2> versus Map<K,V1 | V2>). Map overload 4 is no longer blocked by its unwitnessed mapper-result binder. It was masked behind overload 2 at base, so it is not counted as an independently observed old failure. The identity signatures are not all claimed as new independent passes.

Positive evidence is the completed first-error entry pass reaching the certified storage gate for each set in builder.ts: the callee census must have succeeded before that gate. The standalone full core.ts pair defers these generic declarations; their absence there is not itself acceptance evidence.

### 2. The 17 mapped calls

**4 of the 17 compiler/services mapped call sites now have an observed certified-storage NotYet in the required streams.** These are unique static sites, not repeated event counts. Missing storage refusals do not prove success: a scan refusal, prior stop or independently deferred generic body may prevent reaching a call. [MAPPED-CALLS.md](MAPPED-CALLS.md) accounts for every one of the 17 and lists observation contexts.

- `src/compiler/builder.ts:2416:31` (`arrayToMap`):

```text
a certified overload whose implementation stores Map<Path, V2 | IncrementalBuildInfoDiagnosticOfFile> where the call reads Map<Path, readonly ReusableDiagnostic[]>
```

- `src/compiler/builder.ts:1421:25` (`arrayFrom`):

```text
a certified overload whose implementation stores (IncrementalBuildInfoFileId | Path)[] where the call reads IncrementalBuildInfoFileId[]
```

- `src/compiler/builder.ts:1315:23` (`arrayFrom`):

```text
a certified overload whose implementation stores ([Path, FileInfo] | IncrementalMultiFileEmitBuildInfoFileInfo)[] where the call reads IncrementalMultiFileEmitBuildInfoFileInfo[]
```

- `src/services/codefixes/helpers.ts:586:12` (`arrayFrom`):

```text
a certified overload whose implementation stores (string | TypeParameterDeclaration)[] where the call reads TypeParameterDeclaration[]
```

The tsc.ts checker closure itself contains **11** mapped calls. The prior scout's **17** additionally include six service calls outside that closure. Full-src resolution finds **129 calls**, **27 mapped**, across **47 files**; all 47 files (core included) were selected and attempted, so no server/harness/testRunner caller was silently omitted. The 17-call count above excludes those extra mapped sites.

### 3. Other overload sets

The whole-tsc first-error pair has **no other overload-set finding movement**. Its 19 non-target named sets / 67 ordinal overload findings remain identical by unit and position-stripped diagnostic tuple. Every overload-result/parameter/type-constraint reason was inspected, not only the two target names.

**No other ordinal overload-result/parameter/type-constraint finding moved in any completed required full-file pair either.** The complete before/after ledger, including unchanged findings, is [overload-ledger.json](overload-ledger.json). Non-ordinal wording is also audited in [all-overload-mentions.json](all-overload-mentions.json); no unnamed overload movement is omitted.

A syntactic scope cross-check found **25** overload signatures in **15** owners with extra implementation binders, all inside the 81-source entry closure; see [extra-binders.json](extra-binders.json). This is the entry condition for the changed censusInferRemainder path, not a claim that all such bodies were instantiated. Observed movement is not uniquely attributed to that inference change versus newly reached code without a separate ablation.

## Paired totals and grades

The unfiltered meter-style first-error entry pass is **4488 -> 4485 findings**: 0 new, 6 gone, 0 changed, 3 reached. Headers are identical and checker-clean. Both entry relation reports have no load/end overflow.

Across the 41 completed full-file pairs: 0 new, 96 gone, 43 changed, and 53 reached. All 43 changed pairs are Boundary recovery metadata. The six failed harness-file pairs are excluded. [TOTALS.md](TOTALS.md) contains each file's paired counts. [FINDINGS.md](FINDINGS.md) and [paired.json.gz](paired.json.gz) grade every changed event, including recovery Boundary records. A changed Boundary message is administrative metadata, not a separate overload claim.

Tuples are `(kind, phase, reason, full message)` with complete source coordinates replaced by `<position>`. Correspondence is within the same unchanged-source unit. Identical tuples pair first, preserving multiplicities. Different words for the same overload claim family in the same unit are changed. Remaining base observations are gone. A remaining head non-scan stop in a previously blocked unit is reached; otherwise it is new. Reached does not mean compilation succeeded or a new panic is harmless. Raw positions and original words are retained.

## Inputs, builds and runs

Both worktrees ran recursive submodule initialization before building. Cohere is `5360753e19e4417675a90e92a1823bea66b13918`; TypeScript-go is `d92d9bfee114c80be2c375d72edae966176e3a4f`. Go is 1.27.1. Each sha supplied its own latent generator, meter entry extension, scan-all generator and compiler sources. All six probe hashes, generation/build commands and exits are under [builds](builds). No head tool reused a base overlay or binary.

Apply inputs are unchanged between shas. Both keys are `6c54c265b3ddf357d631b5dc1bb7f13cf328782bdf8f8e9db38e24a5aec7fae4`; the source pin is TypeScript 6.0.3 `050880ce59e30b356b686bd3144efe24f875ebc8`. One base-generated adapted tree was used by both probes. All shared TypeScript source hashes still match the pre-run manifest. [SETUP.json](SETUP.json), [source-manifest.json](source-manifest.json) and [VALIDATION.json](VALIDATION.json) retain these checks.

**100 required probe processes were attempted:** two unfiltered first-error entry runs, 94 exact full-file runs, and four scan-all runs (core and tsc roots at each sha). 88 completed successfully; 12 failed. Every successful latent stream has its relation trailer and every successful full selection is exactly the requested single absolute file. Two additional normal-overlay compiler-directory core runs are labeled supplemental and are not substituted for required runs.

Entry flags: `LATENT_FULL=0`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TAIL=1`, CommonJSBundle, rooted at src/tsc/tsc.ts exactly as the meter entry pass. Full flags: `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_TAIL=1`, exact `LATENT_ONLY_FILE`, CommonJSBundle. Entry-reachable full files use the tsc root; extra callers use their own entry closure. [callers.json](callers.json) records the exact roots and getResolvedSignature results.

Failed run attempts are listed in coverage and raw logs. No complete-file conclusions are drawn from their partial streams. Raw streams/timings are gzip-compressed under [runs](runs); logs, exits, wall time, peak child RSS and manifests remain beside them. Scan-all records its real boundary and is blind after its first declaration/body/tail stop; its zeros are not treated as declaration passes. Its complete before/after observations are retained in paired.json.gz and FINDINGS.md.

The initial scheduler was deliberately interrupted to avoid scheduling duplicate full runs when the entry and full batches were separated. Its already-running entry children completed normally with exit 0 and full trailers; only the scheduler exited 130. The full scheduler ran the file/scan batch independently. The interruption log and child receipts are retained; no census child was canceled or its exit masked.

Checker-rejected entry-root contexts are listed in [checker-rejected-runs.json](checker-rejected-runs.json). Their measurements are reported as such, not as checker-clean compilation passes.

All completed latent relation reports have empty load/end overflow answers; [relation-audit.json](relation-audit.json) is empty.

## Reproduction

Set scratch cache paths and ADAMIC_TYPESCRIPT_SOURCE to the pinned upstream. Create the base/head worktrees, initialize submodules, install stage3/api, and build every probe from its own source using build.py SCRATCH LABEL. Build one shared adapted tree after confirming equal apply keys. Resolve callers with the pinned stock API, then run the entry and full batches separately or in all mode:

```sh
python3 run-pair.py /scratch/paired entry
python3 run-pair.py /scratch/paired full
python3 pair-findings.py /scratch/paired
python3 render-report.py /scratch/paired
```

The scripts and exact executed commands are retained here. There is no adaptation, ledger entry, internal change, native-compilation claim or claim that an unobserved call passed.

## Blocked exact-file runs

12 required children failed. Every file holding one of the 17 mapped compiler/services calls has a complete paired stream; these failures affect additional caller contexts. They prevent claiming a complete full-file census across all 47 files. The whole-entry result remains complete.

- base `src/harness/compilerImpl.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/base/full/src__harness__compilerImpl.log`.

- head `src/harness/compilerImpl.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/head/full/src__harness__compilerImpl.log`.

- base `src/harness/documentsUtil.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/base/full/src__harness__documentsUtil.log`.

- head `src/harness/documentsUtil.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/head/full/src__harness__documentsUtil.log`.

- base `src/harness/fourslashImpl.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/base/full/src__harness__fourslashImpl.log`.

- head `src/harness/fourslashImpl.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/head/full/src__harness__fourslashImpl.log`.

- base `src/harness/harnessIO.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/base/full/src__harness__harnessIO.log`.

- head `src/harness/harnessIO.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/head/full/src__harness__harnessIO.log`.

- base `src/harness/projectServiceStateLogger.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/base/full/src__harness__projectServiceStateLogger.log`.

- head `src/harness/projectServiceStateLogger.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/head/full/src__harness__projectServiceStateLogger.log`.

- base `src/harness/vfsUtil.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/base/full/src__harness__vfsUtil.log`.

- head `src/harness/vfsUtil.ts`: child exit 2; panic: Unhandled case in Node.Text: *ast.ComputedPropertyName. Full log: `runs/head/full/src__harness__vfsUtil.log`.

Scan-all is unchanged for both roots: 4,836 stops before and after, with `production abort at lowering.declareModule` on both revisions. It provides no evidence beyond that boundary.
