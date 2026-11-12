# getPathFromPathComponents adaptation: final proof

**All required gates pass.** Adaptation commit **`3675359ce05b7938b2d56d8a6150970b3301e127`** is the single commit on main **`02ebf2b154495843ba2b1c51de63ed8bd9a8e3b9`**, branch `typescript/path-from-components-adapt`. It contains only the new `stage3/adapt/99-correctness-path-from-components` directory and the upstream ledger entry. Measurement output is on `typescript/path-from-components-evidence`, outside stage3/adapt.

## Corrected replay gate and pre-existing refusal

The earlier stop was superseded by the user's correction. Adapter 30's `declined site changed: core.ts:992:21` refusal reproduces on unmodified main when 30 is replayed over the **fully adapted** tree. **That is not apply's path: apply runs each adapter on its predecessor's output, before later adaptations add `from[i]!` in addRange.** This historical refusal is not attributed to the new path owner and does not require an amendment to 30 for this change.

The isolated `30.plan(ts, "path.ts", text, true)` accepts both baseline and new owner text unchanged: zero edits and four declined reads. The valid replay proofs are the successful **complete stage-by-stage apply.sh run**, and this adaptation's **zero-edit, zero-write replay over its own output**, with bytes and mtime unchanged. No existing adapter or declined-site guard was changed. Historical observations are retained in [30-baseline.log](30-baseline.log), [30-path-only.json](30-path-only.json), and the superseded [HISTORICAL-STOP.md](HISTORICAL-STOP.md). The interrupted old Go run is historical; the final run below completed.

## Final verification on main's tip

| Gate | Result |
|---|---|
| Complete apply.sh, stage by stage | child 0, 777.33 seconds |
| New adaptation's owner/drift contract | child 0; ten no-write refusals, unrelated-owner and LF/CRLF checks |
| New adaptation replayed on its output | child 0; files 0, edits 0, writes 0; bytes and mtime unchanged |
| Whole `go test` stage3/lane package | **child 0; all 398 shards pass; zero failures; package PASS** |
| `generate_shards.py --check` | child 0 |
| Adamic types over sorted 79 roots | child 0 before and after |
| Stock TypeScript 6.0.3 compiler project | 79 roots, zero diagnostics |
| Actual upstream build | child 0 |
| Ten emitted JavaScript outputs | all byte-identical to baseline |
| Both public declaration files | byte-identical to baseline |
| Full-mode, single-file path.ts census | child 0 before and after; no checker diagnostics; output guards enabled |
| Restoring mutant | census child 0; original cast refusal restored at path.ts:513:45 |

The completed whole Go command is unfiltered:

```sh
go test -json -count=1 -timeout=28m -parallel=4 ./stage3/lane
```

It completed in **1048.11 seconds**, including all fixture tests and all **398 unique TestShard_ passes**. [final/lane-summary.json](final/lane-summary.json), [final/lane-go-status.json](final/lane-go-status.json) and [final/lane-go.jsonl](final/lane-go.jsonl) retain the count, exact exit and full stream. This was run to completion; it is not the historical interrupted 28-shard observation.

[final/provenance.json](final/provenance.json) pins main, upstream TypeScript (`050880ce59e30b356b686bd3144efe24f875ebc8`), tools, binary hashes, apply product key and every adaptation/ledger input hash. Those hashes were checked against the final commit's Git objects, so the tested working-tree bytes equal the committed files. The lane's shared product contains the new overloads. The earlier baseline adapted tree is reusable because the intervening main changes only affect Adamic lowering, not stage3's source/adaptation inputs; the final main binaries reran its types and census checks.

## Adapter contents

The new adapter changes only getPathFromPathComponents: two `@internal` overloads for `readonly Path[] -> Path` and `readonly string[] -> string`, and a string implementation with no casts. The overloads precede the existing `//// Path Formatting` comment and JSDoc, preserving emitted comment bytes. `owners.json` pins the complete before/after owner region, including its comments and overload set, with exact text and SHA-256.

Planning parses the whole file and verifies all preconditions and the complete after owner before the sole write. Owner drift, partial edits, missing/moved/duplicate owners and parse errors are refused without writing. Replay makes zero edits and writes, preserving bytes and mtime. Unrelated owners and CRLF are preserved. Full path.ts transpile equality is also checked before writing. See the committed adaptation README and [final/contract.json.log](final/contract.json.log), [final/replay.json](final/replay.json).

The upstream ledger entry records this type-only change and the evidence branch. No compiler, oracle, shard, public API sanction, other adapter or forEachAncestorDirectory owner was changed. [owner-pins-grep.txt](owner-pins-grep.txt) retains the complete earlier pin search: 30's declined root read is the direct target-owner pin; other name hits are copied import blocks.

## forEachAncestorDirectory remains unchanged

Its existing overloads resolve five source calls as **string 3, Path 2**. Implementation-only P is not a caller-visible type parameter: one actual generic P argument resolves to the string overload. The stock checker evidence, complete call positions and signatures remain in [p-instantiations.json](p-instantiations.json).

The same-shape candidate retaining the Path/string overloads but using a string directory/string callback implementation and removing `as P` produces **TS2394**, “This overload signature is not compatible with its implementation signature.” Its JavaScript is identical, but the stock zero-diagnostic gate fails. The README records the count and this reason for exclusion. No callback-union workaround was introduced. See [p-candidate-stock.json](p-candidate-stock.json) and [p-candidate-path.ts.txt](p-candidate-path.ts.txt).

## Exact source and output evidence

The established getPathFromPathComponents source instantiations are **string 13, Path 2**, including five harness alias calls. The target is absent from the upstream public API baseline and sanctioned-api inventory; those complete measurements remain on [typescript/path-from-components](https://github.com/system-inc/adamic/tree/typescript/path-from-components/stage3/census/path-from-components).

The final actual build compared every one of these ten emitted JavaScript files and both public declarations byte for byte, retaining sizes and matching hashes in [final/build-comparison.json](final/build-comparison.json):

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

Source maps and internal declarations are outside that byte-identity claim. The comparison uses raw bytes, not normalized tokens or removed comments. [final/stock-after.json](final/stock-after.json) records the exact 79 project roots and empty diagnostic list; the earlier baseline stock project has zero diagnostics in [p-candidate-stock.json](p-candidate-stock.json). Adamic's explicit sorted root lists, own exits and timings are in [final/types-before-status.json](final/types-before-status.json) and [final/types-after-status.json](final/types-after-status.json).

## Exact census delta and restoring mutant

Both final main-tip censuses used `LATENT_FULL=1`, exact `LATENT_ONLY_FILE` selectors, and `LATENT_ASSERT_NO_OUTPUT=1`, loading the complete compiler directory. Both header diagnostic lists are empty (`checker_rejected: false`); the recorder's fixed historical measurement label does not override those headers.

After normalizing the two inserted lines, the unique `(kind, source location, exact reason)` comparison removes these six findings and adds **none**:

| Baseline path.ts site | Kind | Exact reason removed |
|---|---|---|
| `513:38` | Boundary | `failed statement; state rolled back; continued at next statement or declaration` |
| `516:5` | Boundary | `failed statement; state rolled back; continued at next statement or declaration` |
| `513:45` | Refused | `a cast the runtime can't check` |
| `513:45` | Refused | `an unproven relation from "" to T: the source is not assignable to the target` |
| `516:12` | Refused | `a cast the runtime can't check` |
| `516:12` | Refused | `an unproven relation from string to T: the source is not assignable to the target` |

These are four distinct source positions: the two cast expressions and their two statement boundaries. No other diagnostic site changes after line-shift normalization. `forEachAncestorDirectory`'s `getDirectoryPath(directory) as P` refusal remains, shifting from original 1103:28 to after 1105:28. The overloads and new implementation have zero findings. See [final/census-before.jsonl](final/census-before.jsonl), [final/census-after.jsonl](final/census-after.jsonl), [final/census-delta.json](final/census-delta.json).

The restoring scratch mutant replaces the after owner with its original generic before owner, including `"" as T`; it restores the exact original line numbering and the refusal **path.ts:513:45**, reason `a cast the runtime can't check`. The final main binary remeasured this in [final/census-mutant.jsonl](final/census-mutant.jsonl) and [final/mutant-result.json](final/mutant-result.json). The historical expected-red detector also exits 1 in [restoring-mutant-check-status.json](restoring-mutant-check-status.json). No real adaptation source was mutated.

## Bounds and reproduction

Every final process is supervised by [final/run.py](final/run.py): it reports its child's own exit, bounds it at 29 minutes, and reaps its process group. The whole Go test also has a 28-minute internal bound. No process ran for 30 minutes and no final bound was reached. Generated physical Go evidence files use `.go.txt`; [final/build-overlay.py](final/build-overlay.py) applies that naming before generation. Binaries are excluded from evidence; their SHA-256 values are retained in provenance.

Reproduce using the pinned upstream checkout, locked stage3 parser/build dependencies and the committed adapter. Run apply.sh into a new output, then the saved final drivers (adjust their recorded absolute paths together when relocating). Keep every output outside stage3/adapt. The adapter README links this evidence branch. All final status files report child 0; the historical stop and failed ancestor candidate remain explicitly identified as history and rejection evidence.
