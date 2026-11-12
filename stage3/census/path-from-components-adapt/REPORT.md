# Path-from-components adaptation: stopped at adapter 30 replay

**Stopped under the user's explicit gate: “If 30 refuses, stop and report.”** No adaptation commit was made or pushed. The proposed four-file adapter and ledger line are retained as text snapshots here, and remain uncommitted in `/workspace/path-components-adapt` on `typescript/path-from-components-adapt`. No existing adapter was amended.

Base: Adamic main **`5243a7e8fc1fa6c0696a3b875881dc02ec6301ef`**, fetched for this task on 2026-11-12. Upstream TypeScript 6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`. The whole `99-correctness-axolotl-k/adapt.cjs`, `owners.json` and README were read before implementing this draft. Measurement output is exclusively on `typescript/path-from-components-evidence`, in this directory.

## The blocker and its scope

The draft's **full apply passed**, but the required adapter-by-adapter replay stopped at **30-indexed-reads**, child exit **1**, with:

```text
declined site changed: core.ts:992:21
```

The exact same full adapter command against the **unmodified main baseline adapted tree** also exits **1** with that message. Its ledger entry (`stage3/adapt/30-indexed-reads/sites.json:614–625`) is the declined `addRange` read `from[i]`, occurrence 2 of 2. The fully adapted baseline has `to.push(from[i]!)` at `core.ts:1002`. Adapter 30's declined-site guard rejects the non-null assertion. This is a pre-existing whole-replay incompatibility, not evidence that the new path owner caused the refusal.

The isolated exported `30.plan(ts, "path.ts", text, true)` check **accepts both baseline and candidate path.ts**: zero assertions/zeros/closures to add, four declined reads, and byte-unchanged planned text in each case. Thus the new `getPathFromPathComponents` owner itself still satisfies the requested read pin. See [30-path-only.json](30-path-only.json), [30-baseline.log](30-baseline.log), [30-baseline-status.json](30-baseline-status.json), and [replay.json](replay.json).

Nevertheless, the requested **whole replay is red**, so the implementation path stopped. Resolving the failed existing pin requires review/amendment within 30 before this full-replay gate can pass; the draft does not bypass or weaken it. The path-only result alone does not justify claiming that all adapter replays pass. No later adapters were replayed after 30 refused.

## Draft implementation and owner contract

[proposed-adaptation/adapt.cjs.txt](proposed-adaptation/adapt.cjs.txt), [owners.json.txt](proposed-adaptation/owners.json.txt), [contract.cjs.txt](proposed-adaptation/contract.cjs.txt), and [README.md.txt](proposed-adaptation/README.md.txt) preserve the proposed directory `stage3/adapt/99-correctness-path-from-components`. [proposed-ledger.diff](proposed-ledger.diff) preserves the proposed ledger addition. These are **reviewable drafts, not landed changes**.

The draft writes exactly the previously measured candidate: two `@internal` Path/string overloads **before the existing `//// Path Formatting` comment and JSDoc**, followed by the string implementation with both `as T` assertions removed. `owners.json` pins the entire before/after region, including the full owner, complete overload set and comments, by exact normalized text and SHA-256. It parses the complete file, verifies the planned after owner and byte-identical transpile, and rereads the original before its sole write. It preserves CRLF.

The contract child exited **0**. It verified a zero-edit, zero-write replay with identical bytes and mtime, unrelated-owner preservation and LF/CRLF handling. Ten drift/partial-owner cases were refused with zero writes: changed before owner, each half-removed cast, missing/moved/duplicate owner, parse error, changed after owner, missing after overload and moved section comment. See [contract.json](contract.json).

The original source-call measurement remains **T = string: 13; T = Path: 2** (including five harness alias calls), on the earlier [report branch](https://github.com/system-inc/adamic/tree/typescript/path-from-components/stage3/census/path-from-components). Both requested public API inventories omit the target function.

## Other pins on the two owners

[owner-pins-grep.txt](owner-pins-grep.txt) is the pre-draft `rg -n 'getPathFromPathComponents|forEachAncestorDirectory' stage3/adapt` output. The only direct named path-function/read pin is 30's declined `pathComponents[0]` ledger entry at `sites.json:3689–3699`. The other name hits are copied import-owner blocks in `99-correctness-core-or-predicate/owners.json` and `99-correctness-core-multimap/owners.json`; they do not pin either path function's body. No other complete `forEachAncestorDirectory` owner pin was found. The draft adds no ancestor-function pin because its candidate fails the stock gate below.

## forEachAncestorDirectory measurement and rejection

[measure-p.cjs](measure-p.cjs) uses the stock TypeScript 6.0.3 checker on the baseline's **604 adapted src roots**, including compiler, services, server, harness and tests. It matches resolved call declarations to the actual path.ts owner. [p-instantiations.json](p-instantiations.json) retains complete call expressions, argument types and resolved signatures.

A distinction matters here: this function already exposes Path and string overloads parameterized only by **T**. **P exists only on the hidden implementation**, so the checker does not infer a caller-visible P type argument. The analogous measurable directory instantiation is the resolved signature's first parameter: **string 3, Path 2**. Actual argument types are string 2, Path 2, and a generic P 1; that P argument resolves to the string overload.

| Adapted file:line:column | Argument type | Resolved directory type | Call prefix |
|---|---|---|---|
| `src/compiler/moduleNameResolver.ts:503:5` | `string` | `string` | `forEachAncestorDirectory(normalizePath(currentDirectory), directory => {` |
| `src/compiler/moduleNameResolver.ts:3057:12` | `P` | `string` | `forEachAncestorDirectory(directory, ancestorDirectory => {` |
| `src/compiler/watchUtilities.ts:296:9` | `Path` | `Path` | `forEachAncestorDirectory(` |
| `src/compiler/program.ts:330:12` | `string` | `string` | `forEachAncestorDirectory(searchPath, ancestor => {` |
| `src/typingsInstallerCore/typingsInstaller.ts:258:21` | `Path` | `Path` | `forEachAncestorDirectory(getDirectoryPath(fileName), directory => {` |

The scratch candidate retained both existing overloads and changed only the implementation to:

```ts
export function forEachAncestorDirectory<T>(directory: string,
    callback: (directory: string) => T | undefined): T | undefined {
    // original body, with getDirectoryPath(directory) and no `as P`
}
```

Stock TypeScript reports **one TS2394**: “This overload signature is not compatible with its implementation signature.” The Path callback overload is incompatible with the string callback implementation. The baseline has zero diagnostics, and the candidate's 79 per-root JavaScript hashes still match, but **the zero-stock-diagnostics condition fails**. The conditional inclusion gate therefore rejects this candidate; the ancestor owner remains unchanged. No broader callback-union rewrite or cast workaround was attempted. Its census was not needed after that failed gate.

See [p-candidate.cjs](p-candidate.cjs), [p-candidate-path.ts.txt](p-candidate-path.ts.txt), [p-candidate-stock.json](p-candidate-stock.json) and [p-candidate.log](p-candidate.log). That experiment's driver exits 0 because it records diagnostic evidence; it does **not** claim that the candidate type-check passes.

## Completed proof results

| Check | Measured result |
|---|---|
| Full main baseline apply | child 0 |
| Full apply including new adapter | child 0, 404.61 seconds |
| New adapter contract / isolated replay | child 0; replay zero edits/writes, bytes and mtime unchanged |
| Every adapter replay | **STOPPED:** 00, 10 and 20 child 0 / zero changed files; 30 child 1 / zero changed files; remaining adapters not run |
| 30 baseline full replay | child 1, same existing core.ts refusal |
| 30 path-only check | child 0, baseline and candidate accepted without edits |
| Actual upstream baseline and candidate builds | both child 0 |
| Ten emitted build JS outputs | all byte-identical |
| Both public declaration files | byte-identical |
| Adamic types, sorted 79 roots | child 0 before and after |
| Stock TypeScript, 79 compiler-project roots | zero diagnostics before and after |
| Full-mode single-file census before/after | both child 0, no checker diagnostics, output guards enabled |
| Restoring mutant census | child 0, records original 513:45 refusal; expected-red detector child 1 |
| generate_shards.py --check | child 0 |
| Whole stage3/lane Go package, all 398 shard declarations | **INTERRUPTED at required stop gate**, child -15 / shell 143, 531.88 seconds; not a passing lane |

[build-comparison.json](build-comparison.json) records exact byte counts and matching before/after SHA-256 values for each of these outputs:

- `_tsc.js`: 6227135 bytes, SHA-256 `cbeb71fde0a41040237b6246f44b6ca8eb41ab5da40e03e6ea33618cbb95e4e8`.
- `_tsserver.js`: 27889 bytes, SHA-256 `1992fc3518a4aa5443638a25794281995ec3ce84820692713296262997d5443b`.
- `_typingsInstaller.js`: 10364 bytes, SHA-256 `71515c1c1aa29c8616309f7dea4d3a32eed662df3a4dbcf4345e7522ece56ac1`.
- `run.js`: 13170484 bytes, SHA-256 `287162d1e70dc447d3b20e7b25d3221a396b6e08f583bd20fcee2edc136ee43d`.
- `tsc.js`: 267 bytes, SHA-256 `2cffde0b8c6760dfb0b5b0382bbb7e00ba6a8b2d981b9205b256a700a481d983`.
- `tsserver.js`: 272 bytes, SHA-256 `e3ccfeec65ec5c470b8ffc5611878a31182650c7e8a062c38a719f83b523edcb`.
- `tsserverlibrary.js`: 1013 bytes, SHA-256 `6a0b343fa3a3f53cf198c83a18035c104b6990b2a260272654f9a6b2f2380fdf`.
- `typescript.js`: 9136687 bytes, SHA-256 `1cd4c38977a16cfacb23c9bfb81b5f46a83f7217dccae9a705a848362ac86c78`.
- `typingsInstaller.js`: 280 bytes, SHA-256 `2d48a038523ea7c626fa2796a30fa5fdc07612ea5fab83544423009d7683a213`.
- `watchGuard.js`: 2314 bytes, SHA-256 `008f2ceb504ee9b8c5d47054d43033e1ec3f2fef03f18d8aa1b2dd70cba6fc88`.
- `typescript.d.ts`: 593506 bytes, SHA-256 `e14a3c25a0ec4f42f05974fe029a3de87c13cdb9b92e823385a5e9ac321c31bd`.
- `tsserverlibrary.d.ts`: 866 bytes, SHA-256 `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd`.

The ten JS files come from actual `npm run build` output, not a comment-stripped or token-normalized comparison. Source maps and internal declarations are outside that byte-identity claim. The builds include the upstream stock checks; [stock-after.json](stock-after.json) separately records the project-config stock checker over exactly 79 roots with zero diagnostics. The baseline's zero diagnostics are also retained in the baseline half of [p-candidate-stock.json](p-candidate-stock.json). Adamic types commands and all 79 explicit root paths are in [types-before-status.json](types-before-status.json) and [types-after-status.json](types-after-status.json).

The Go command was unfiltered:

```sh
go test -json -count=1 -timeout=28m -parallel=4 ./stage3/lane
```

The process group was explicitly terminated under the user's stop rule. The final JSON stream records 28 completed shard passes and no failure entries before interruption; the remaining 370 shards and the package result are unproved. [lane-go.jsonl](lane-go.jsonl), [lane-go.stderr](lane-go.stderr), [lane-go-status.json](lane-go-status.json) and [STOP.json](STOP.json) retain this incomplete run; 28 completed shard passes do not establish the whole lane. It must be rerun after resolving the replay blocker. No shard, expectation, timeout policy, compiler or existing adaptation was changed to get a green result.

## Exact census changes

Both censuses used `LATENT_FULL=1`, an exact absolute `LATENT_ONLY_FILE` for path.ts and `LATENT_ASSERT_NO_OUTPUT=1`, loading the complete compiler directory. The before header and after header have `checker_rejected: false` and empty diagnostic lists. The recorder's fixed label still says “measured on a checker-rejected program”; the actual headers establish that these inputs checked clean.

Before: **81 units, 102 raw findings**. After: **83 units, 84 raw findings**. The new overload units and implementation are attempted and have no findings. The two inserted lines shift later original coordinates by +2; [census-delta.py](census-delta.py) normalizes that shift before comparing unique `(kind, location, exact reason)` tuples. It finds **six removed unique findings and no additions**:

| Baseline path.ts site | Kind | Exact reason removed |
|---|---|---|
| `513:38` | Boundary | `failed statement; state rolled back; continued at next statement or declaration` |
| `516:5` | Boundary | `failed statement; state rolled back; continued at next statement or declaration` |
| `513:45` | Refused | `a cast the runtime can't check` |
| `513:45` | Refused | `an unproven relation from "" to T: the source is not assignable to the target` |
| `516:12` | Refused | `a cast the runtime can't check` |
| `516:12` | Refused | `an unproven relation from string to T: the source is not assignable to the target` |

These account for four distinct source positions: the two casts and their two rolled-back statement boundaries. Repeated dependency attempts explain why 18 raw events disappear rather than only six. No other substantive or boundary site changes after line-shift normalization. In particular, the `forEachAncestorDirectory` cast refusal at original **1103:28** remains at after **1105:28**, with reason `a cast the runtime can't check` and expression `getDirectoryPath(directory) as P`.

The restoring mutant replaces only the new owner's pinned after spelling with its original generic before spelling, restoring `"" as T` (and the original final cast) and removing the two overload lines. Its path.ts is byte-equivalent to baseline. The mutant census again records **path.ts:513:45**, reason `a cast the runtime can't check`; [restoring-mutant-check.log](restoring-mutant-check.log) retains the exact matching diagnostic and expected-red exit 1. No adaptation source is mutated for this proof.

Raw streams: [census-before.jsonl](census-before.jsonl), [census-after.jsonl](census-after.jsonl), [census-mutant.jsonl](census-mutant.jsonl). The exact changes are in [census-delta.json](census-delta.json).

## Process bounds, reproduction and remaining work

[run.py](run.py) bounds each child at 29 minutes, reports that child's own exit code and elapsed time, reaps its process group, and returns the child's status (signal-derived shell status for signals). Every wrapper shell prints its returned status too. Adapter replay children additionally have a 600-second bound. The whole Go test had a 28-minute internal limit and was interrupted after 8.86 minutes; no process reached 30 minutes. The externally requested stop is recorded separately in STOP.json because it was not a timeout. Go overlay evidence filenames end `.go.txt`; [build-overlay.py](build-overlay.py) performs that naming at generation time. Binaries are excluded here; their hashes and compiler/submodule pins are in [provenance.json](provenance.json).

Setup-only mistakes were corrected without compiler changes: the first shard-generation attempt lacked gofmt on PATH; the retry sourced the configured tools and passed. An initial scratch types driver was renamed because `types.py` shadowed Python's standard module; the actual types measurements then passed. These are not reasons for the adapter-30 stop.

To reproduce the blocker, prepare main's adapted tree with the pinned parser and upstream checkout, then run adapter 30 on that completed baseline tree. Its full replay rejects the existing core.ts declined read even without the new adapter. To reproduce the candidate proof, copy the proposed four-file directory into stage3/adapt, apply into a new tree, and use the saved scripts with their recorded scratch paths (or adjust the paths consistently). Keep measurements outside stage3/adapt so they do not change the apply product key.

**Remaining:** resolve the existing adapter-30 whole-replay incompatibility through its owner, then rerun the complete adapter replay and the whole 398-shard Go lane. Only after those gates pass should a new adaptation commit be pushed. The requested separate adaptation push is intentionally not made under the explicit stop instruction.
