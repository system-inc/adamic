# forEachAncestorDirectory adaptation: final proof

Candidate A is implemented in `stage3/adapt/99-correctness-for-each-ancestor` on main **`2f32716054f245622c2ac3c8f692d00fa51e66e9`**, in fresh worktree `/workspace/ancestor-adapt`. Only that adapter directory and its upstream ledger entry belong to the adaptation change. Measurement output is on the separate `typescript/for-each-ancestor-evidence` branch.

**Adaptation commit: `2a7d2cf9cf81bb1e32e4fb589ca99f17ac189b1d`**, the single commit on the pinned main tip, branch `typescript/for-each-ancestor-adapt`. Its four adapter files were compared against the tested SHA-256s through their Git objects; they match exactly.

## Gates

| Gate | Result |
|---|---|
| Complete stage-by-stage apply.sh | child 0; 572.78 seconds |
| Owner/drift contract | child 0; 14 no-write refusals; LF/CRLF and unrelated-owner preservation |
| Replay over complete apply output | child 0; files 0, edits 0, writes 0; bytes and mtime unchanged |
| Existing adaptation 30 path-only check | child 0 before and after; zero edits, four declined reads |
| Existing path-from-components replay | child 0 before and after; zero edits |
| Stock TypeScript 6.0.3 compiler program | 79 roots, 0 diagnostics before and after |
| Extended stock caller program | 80 explicit roots including typingsInstaller; 0 diagnostics before and after |
| Adamic types | 79 sorted roots; child 0 before and after |
| Actual upstream build | child 0 before and after |
| Actual emitted JavaScript outputs | all 10 byte-identical |
| Both public declaration files | byte-identical |
| Extended stock program emission | all 89 JavaScript files byte-identical |
| Full-mode single-file path.ts census | child 0 before, after and mutant; all headers have 0 checker diagnostics |
| Census delta | exactly the cast refusal at 1105:28 removed; no new findings |
| Restoring mutant | restores path.ts bytes and the complete baseline finding set, including 1105:28 |
| generate_shards.py --check | child 0 |
| Sanctioned API freshness in whole Go run | PASS |
| Whole stage3/lane package | **child 0; package PASS; all 398/398 shards pass; no failures or skips** |

The whole, unfiltered command was:

```sh
go test -json -count=1 -timeout=28m -parallel=4 ./stage3/lane
```

It completed in **1397.92 seconds wall time**, with package elapsed **1384.842 seconds**, **430 unique Go tests passed**, **398/398 unique shards passed**, zero failures and zero skips. [lane-summary.json](lane-summary.json), [lane.status.json](lane.status.json) and the complete [lane.log.gz](lane.log.gz) retain the result. **TestLaneSanctionedApiFreshnessPython passed in this same unfiltered run.**

The **additional** Adamic check of all 80 extended roots is not green on the unchanged baseline: it has **32 existing diagnostics**, including TS1484 type-only imports and optional-value errors in jsTyping/typingsInstaller. Its before and after children both exit **1**, with **identical diagnostics after normalizing only the root path**. These failures are retained in `types-before-extended.log.gz`, `types-after-extended.log.gz` and [extra-adamic-extended-comparison.json](extra-adamic-extended-comparison.json). The requested 79-root Adamic checks and the stock extended caller check are clean. The `types-*-driver` wrappers preserve the extra child's exit 1; the after-proof orchestrator explicitly expects that extra probe's recorded failure and verifies equality. No errors were suppressed or fixed outside the target owner.

## Owner decision and exact edit

Before implementation, the sibling's **adapt.cjs, owners.json and contract.cjs were read in full**. [prior-owner-pins-grep.txt](prior-owner-pins-grep.txt) and [pin-audit.json](pin-audit.json) record the prior adaptation audit. There is no existing complete-owner pin for forEachAncestorDirectory. Adaptation 30's path.ts entries cover indexed reads in other functions, including getPathFromPathComponents; none belongs to this loop. Its whole-owner path list is empty. `99-correctness-path-from-components/owners.json` pins only getPathFromPathComponents. A new adapter is therefore appropriate; neither earlier adapter is amended.

The edit is A exactly: remove the Path overload signature; retain the string overload; replace the generic implementation signature with:

```ts
export function forEachAncestorDirectory<T>(directory: string, callback: (directory: string) => T | undefined): T | undefined
```

Remove `as P` from `getDirectoryPath(directory)`. Keep the loop, comments and all callback bodies unchanged. All five source calls now resolve through the string overload; [stock-comparison.json](stock-comparison.json) retains their complete expressions and signatures. [source-comparison.json](source-comparison.json) verifies that only path.ts changes among the 79 compiler source roots.

The whole owner pin includes **Path Traversal, documentation, every overload declaration, and the complete implementation**. Before has three declarations; after has two. Exact LF text and SHA-256 identify both states. The adapter parses the whole file, rejects missing/moved/duplicate/changed/half-adapted owners, checks the complete planned owner and **full-file stock transpile equality before writing**, rereads the original bytes, and preserves CRLF. A replay performs no write and preserves bytes and mtime. The 14 refusing mutants and positive LF/CRLF/unrelated-owner checks are retained in [contract.json](contract.json). [replay.json](replay.json) checks replay against the actual full apply output.

[existing-pins.json](existing-pins.json) records the successful path-only checks and sibling replays against both trees. A wholesale adaptation 30 replay over a fully adapted tree is not substituted for stage-by-stage apply; its historical later-core conflict is unrelated to this owner.

## The Path callers need only string

- **src/compiler/watchUtilities.ts:296**, `clearFirstAncestorEntry`, within `createCachedDirectoryStructureHost`: `ancestor => cachedReadDirectoryResult.delete(ensureTrailingDirectorySeparator(ancestor)) ? true : undefined`. The separator helper accepts string; the cache is `Map<string, MutableFileSystemEntries | false>`. No operation requires Path.
- **src/typingsInstallerCore/typingsInstaller.ts:258**, `TypingsInstaller.installPackage`: the callback checks `fileExists(combinePaths(directory, "package.json"))` and returns directory. Its result, or projectRootPath, is passed to `installWorker` at line 264; that method's cwd parameter is string (line 552). No downstream consumer requires Path.

These locations and callback bodies are unchanged by the adapter. The extended stock check matters because typingsInstaller is outside the 79 compiler roots. It checks the compiler roots plus that entry using the same strict options, with composite disabled for the extended program. Both stock programs use the pinned TypeScript 6.0.3 API. Primary Adamic checks use the explicit sorted 79 roots; their root lists and own exits are saved separately.

## Actual output identity and public API

Both trees ran **npm run build**, using the locked stage3/typescript install. [build-comparison.json](build-comparison.json) contains matching raw-byte lengths and SHA-256s for **all ten top-level built/local JavaScript outputs** and **typescript.d.ts plus tsserverlibrary.d.ts**. The files are `_tsc.js`, `_tsserver.js`, `_typingsInstaller.js`, `run.js`, `tsc.js`, `tsserver.js`, `tsserverlibrary.js`, `typescript.js`, `typingsInstaller.js`, and `watchGuard.js`.

Separately, [stock.cjs](stock.cjs) emits the compiler and extended caller programs through stock Program.emit; **all 89 JavaScript outputs are byte-identical**. Their per-file hashes are in [stock-before.json](stock-before.json) and [stock-after.json](stock-after.json). This includes typingsInstaller and its dependency closure, not only the compiler project. Source maps and internal declaration outputs are outside these JavaScript/public-declaration byte-identity claims.

The function is absent from npm 6.0.3's public typescript.d.ts, both actual public declaration outputs, and stage3/lane/sanctioned-api.json. [public-api.json](public-api.json) records the actual declarations/inventory searches. [sanction-proof-hashes.json](sanction-proof-hashes.json) verifies that **all 28 proof files listed by sanctioned-api.json retain their recorded hashes**. No listed proof file changes, no public API change is introduced, and neither sanctioned-api.json nor shard-plan.json needs regeneration. The unfiltered whole Go run includes and passes **TestLaneSanctionedApiFreshnessPython**; it is not replaced by the hash audit or a filtered run.

## Exact census delta and restoring mutant

All three censuses use the fresh worktree's main-tip binary, `LATENT_FULL=1`, exact absolute `LATENT_ONLY_FILE` selectors, and `LATENT_ASSERT_NO_OUTPUT=1`. They load the complete compiler directory. Each header reports `checker_rejected: false`, empty checker diagnostics and `latent_mode: full`. The recorder's fixed historical label does not override these actual headers.

After mapping after-tree lines at/after 1094 to their before-tree line +1, the unique `(kind, location, exact reason)` difference is exactly:

```text
removed: Refused, path.ts:1105:28, a cast the runtime can't check
added:   none
```

The raw finding multiset, including phases, also has exactly one removed event and zero added events; the mutant restores that multiset exactly.

The implementation is still generic: its ledger is **deferred_generic_instantiation**, with `instantiated: false`. This census proves the recorded cast site is gone and no new path.ts finding appears; it does not claim successful native lowering of a concrete generic instantiation.

The restoring scratch mutant replaces the exact after owner with the pinned before owner, preserving CRLF. Its complete path.ts bytes equal baseline, and its complete normalized finding set equals baseline, including the refusal at **1105:28**. [census-delta.json](census-delta.json) retains the assertion, exact delta and unit ledgers. The three `census-*.jsonl` streams retain the complete observations.

## Provenance, bounds and reproduction

The baseline originally came from a successful main-tip apply. Main is still the same commit, and every one of its prior adaptation input hashes was checked against the fresh worktree; its path.ts bytes also match the prior measurement pin. [baseline-reuse.json](baseline-reuse.json) and [baseline-original-apply.status.json](baseline-original-apply.status.json) document that reuse. **Build, stock types, Adamic types and census were rerun in this task**. The new adaptation's complete apply was run fresh, not synthesized by editing the baseline. [lane-product.json](lane-product.json) also verifies that the whole lane's shared product contains exactly the tested after path.ts and identical built JavaScript/public declarations. [patch-set.md](patch-set.md) retains its full stage-by-stage table.

[provenance.json](provenance.json) pins main, upstream, every adaptation input, before/after source bytes, binaries, sanction and shard-plan hashes. [build-overlay.py](build-overlay.py) generates the measurement-only census overlay with physical Go evidence names ending **.go.txt**. No production compiler source or binary is committed.

[run.py](run.py) records each child's **own exit code**, duration and timeout flag, and bounds its child at **29 minutes**. The whole Go test also has an internal **28-minute** timeout. No recorded child reaches 30 minutes or times out. Shell setup and dependency checkout operations also completed below that bound. Every proof's status JSON retains the exact command. Child stdout/stderr is preserved in compressed `.log.gz` files; the whole Go JSON stream is also the complete lane log.

An early local shared clone could not serve missing promisor objects; dependency worktrees over the already checked-out pins replaced it. A helper was initially named types.py, shadowing Python's standard module; it was renamed check-types.py before the affected baseline checks were rerun. Neither setup issue changed adaptation inputs or the running lane. The retained completed measurement statuses and assertions are authoritative.
