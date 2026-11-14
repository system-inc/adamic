# Publication disposition: summaries and receipts only

The user instructed: "Publish summaries and receipts only. The raw evidence stays in your scratch, with each raw file’s sha256 and size recorded in the report. Then stop and wait: the next step for 1533 waits on a ruling."

That instruction is final for this publication. Raw census/type logs, source snapshots, diffs, and owner-search payloads remain in scratch. No raw upload or permission request is pending. The evidence-only decision remains STOP_NEW_CENSUS_FINDINGS; no implementation branch or Loom submission exists. **1533 is on hold pending a ruling.**

The report below describes the completed local proofs. Raw artifact names refer to local scratch files, not published payloads. The exact byte sizes and SHA-256 hashes are recorded in the table below and in [RAW-MANIFEST.json](RAW-MANIFEST.json). Published files are listed in PUBLICATION.json.

# watch-directory 1533: stop at the census gate

Base: `8e3a784aa5a601f57342c07b68f145fe56252c21`.
Upstream: TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.
Task: type-only watchDirectory 1533 adaptation, excluding watchFile 1532. No Loom submission.

**Decision: STOP.** The read gate holds, but a consistent type-only widening creates two new NotYet findings and three associated Boundary records. The requested zero-new-finding gate fails. No implementation branch, adaptation amendment, ledger line, or landing claim is supplied. Only the authorized evidence branch is published.

The final scratch experiment widens `recursive: boolean` to `recursive: boolean | undefined` at 12 signature positions: nine in sys.ts, one in virtualFileSystemWithWatch.ts, and two in watchUtils.ts. This includes the explicitly returned HostWatchDirectory alias and the concrete downstream contracts. It changes no runtime expression. `candidate.diff` and the compressed complete source snapshots retain every byte.

## Every recursive read

All reads either test truthiness, negate/coerce the boolean, or forward to another reader listed here. No read distinguishes undefined from false. `reads.json` retains the exact source text on both trees. Parameter declarations, comments and child-watcher calls with literal false are not incoming-value reads.

| File | Adapted line | Upstream line | Reader |
| --- | ---: | ---: | --- |
| src/compiler/sys.ts | 1111 | 1108 | pass to fsWatch |
| src/compiler/sys.ts | 1129 | 1126 | pass to hostRecursiveDirectoryWatcher |
| src/compiler/sys.ts | 621 | 623 | truthiness branch |
| src/compiler/sys.ts | 623 | 625 | pass to nonRecursiveWatchDirectory |
| src/compiler/sys.ts | 1133 | 1130 | !recursive assertion |
| src/compiler/sys.ts | 1163 | 1160 | pass to fsWatch |
| src/compiler/sys.ts | 1211 | 1207 | truthiness selects watcher map |
| src/compiler/sys.ts | 1215 | 1211 | pass to fsWatchHandlingExistenceOnHost |
| src/compiler/sys.ts | 1269 | 1264 | pass to worker or timestamp worker |
| src/compiler/sys.ts | 1367 | 1362 | pass to worker |
| src/compiler/sys.ts | 1776 | 1782 | !!recursive normalizes before Node |
| src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts | 889 | 889 | pass to watchUtils.fsWatch |
| src/harness/watchUtils.ts | 184 | 182 | truthiness selects watcher map |

The only concrete createSystemWatchFunctions callers are getNodeSystem and TestServerHost. At the Node boundary `!!recursive` is exactly false in both cases. The test worker eventually chooses its map with a ternary. Missing-entry/polling fallbacks do not read recursive. The complete required owners/targets/adapt.cjs search is `owner-search.json.gz`; the factory is pinned by 99-correctness-single-watcher-forwarder, TestServerHost by 99-correctness-timer-closures, and getNodeSystem by 99-wasm-crypto-import (with preceding 93/41 changes). Since the census fails, no pins are amended.

## Proof results

- Full single-file sys.ts census: `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=<absolute sys.ts>`, all 79 compiler roots, CommonJSBundle. All three final runs exited 0 and have the final relation trailer. Main has 185 records; candidate 189. There are **2 new NotYet, 3 new Boundary, 1 gone Refused, 0 changed**. Including every kind, 5 new / 1 gone / 0 changed. The gone finding is refusal_scan 1533:13. Every new/gone record and its phase appears in full in delta.json; no other finding disappears. Main and candidate have zero checker diagnostics.
- New NotYet, phase `lowering`, sys.ts:1215:82: `a boolean | undefined variable a function value captures`.
- New NotYet, phase `lowering`, sys.ts:1216:208: `reading callbacks`.
- New Boundary, phase `lowering`, sys.ts:1210:9 and 1216:170: `failed statement; state rolled back; continued at next statement or declaration`, with the captured-optional-boolean failure text above.
- New Boundary, phase `lowering`, sys.ts:1216:204: the same boundary reason, with the reading-callbacks failure text above.
- 1532:13 remains byte-for-byte the same Refused finding (apart from normalized tree paths) in main, candidate, and mutant. It is not modified.
- Restoring mutant: all three source files are restored to their original exact bytes. It returns 1533:13 and its complete contextual findings equal main. 1532 is still present. Complete streams are main.jsonl.gz, change.jsonl.gz, mutant.jsonl.gz. Comparisons normalize only tree paths and exclude shifted numeric byte offsets from identity; original offsets remain in every raw stream and reported delta record. Unit, phase, kind, location, reason and text remain in the contextual comparison.
- Stock tsc 6.0.3: 79 roots, 0 main diagnostics, 0 candidate diagnostics, 0 new. The script enables the pinned Node ambient types and otherwise follows production checker options. stock-delta.json retains the result.
- Ordinary `adamic types`: all 79 roots, exit 0. Complete declaration/type output is retained losslessly in adamic-types.log.gz.parts/ (concatenate the numerically ordered parts, then gunzip).
- JavaScript: 604 non-declaration TypeScript files emitted independently with stock 6.0.3; every corresponding output is byte-identical. All 747 source files were compared; only the three signature-only source files differ. javascript.json.gz retains each output hash.
- stage3/apply.sh --build-product: unmodified main passed, exit 0; its patch table is baseline-patch-set.md. No candidate adapter was installed because the census gate failed, so candidate apply and adapter contracts are not claimed.
- Declarations: upstream's own npm ci and node_modules/.bin/hereby dts passed, exit 0, with their logs retained. A main/candidate declaration comparison was not run after the stop; there is no fabricated empty diff.
- Whole unfiltered stage3/lane: not run after the census stop; 106,366 passing is not claimed.
- Ledger: unchanged because no adaptation is built.

## Reproduction and provenance

All work took place in the scratch detached worktree and scratch source trees, not the shared checkout. Recursive shallow submodules were initialized. GOTMPDIR, TMPDIR and GOCACHE were scratch paths for Go builds; setup.json records commits and Node version. make_overlay.py generated the guarded measurement binary, with CommonJSBundle selected by the repository's own unchanged census templates. The raw run logs and build logs are included. Every Go source written by the measurement overlay was outside stage3; no new .go file is committed here.

Commands, with R=/workspace/scratch/watch-directory-1533 and W=$R/work:

```sh
ADAMIC_TYPESCRIPT_SOURCE=$R/TypeScript XDG_CACHE_HOME=$R/cache TMPDIR=$R GOTMPDIR=$R GOCACHE=$R/go-cache $W/stage3/apply.sh --build-product $R/baseline
python3 $W/stage3/census/latent/make_overlay.py $W $R/overlay
go build -buildvcs=false -overlay=$R/overlay/overlay.json -o $R/census ./stage3/census/latent/tool
# For each T in baseline, candidate, mutant; O in main, change, mutant respectively:
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=$R/$T/src/compiler/sys.ts $R/census $R/$T/src/compiler $R/$O.jsonl
NODE_PATH=$W/stage3/api/node_modules node $R/stock.cjs $R/baseline $R/stock-main.json
NODE_PATH=$W/stage3/api/node_modules node $R/stock.cjs $R/candidate $R/stock-change.json
NODE_PATH=$W/stage3/api/node_modules node $R/emit.cjs
python3 $R/delta.py
```

`experiment.py SCRATCH_PARENT` reproduces the final candidate from a completed baseline. The initial one-parameter pilot is isolated under pilot/: it widened the implementation parameter but left the explicit result alias unchanged, producing two diagnostics while leaving 1533 present. It was corrected before the final census by consistently widening the already-traced signature contracts. The final exact candidate is the diff and source snapshots; experiment.json records its scope. Once the corrected experiment produced new findings, no further widening or runtime change was attempted.

The first shell push lacked GitHub credentials. The connected GitHub tools published checkpoint c1954286b0fe5c8118177604ea4ba1e541a2463c and publish summaries and receipts by a fast-forward ref update, never force. No PR, message, or Loom submission is created.

## Retained raw evidence byte inventory

Paths are relative to `/workspace/scratch/watch-directory-1533/`. This inventory also includes the complete local bundle’s summaries and scripts, so every retained bundle file is accounted for. Execution binaries and dependency checkouts are outside this evidence inventory. Raw files were hashed without modification.

| Scratch-relative path | Bytes | SHA-256 |
| --- | ---: | --- |
| `TypeScript/src/compiler/sys.ts` | 86142 | `9fe83b83c72ee53545fd63384f67c0beb31e864d1cc315ba126014d801e2212a` |
| `TypeScript/src/harness/watchUtils.ts` | 9258 | `10f15abc55b68da94ab3e9b9a222ed439dc450579edb31fad0fbe4a0a289391b` |
| `TypeScript/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts` | 57378 | `2ed307de9a0e69c9d10b2a5cb7e2702105b3ff4da37327b7bed23f2b9382179a` |
| `adamic-build.log` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `adamic-types-exit.json` | 25 | `4e0f292d4e9b5c7ee955cf61c2b15099e9f18252da51c8d7934cede9f85fcacf` |
| `adamic-types.log` | 7243852 | `b5e2bb553a87bced5448900287553d8b7cd97d604a08e1401111801c8cde958f` |
| `baseline/src/compiler/sys.ts` | 86914 | `30876cbae0e2cb19eae2525cbed919abdba86fac921ecc22fa6aaad0441a3398` |
| `baseline/src/harness/watchUtils.ts` | 9363 | `78b02ee9cb8a5198c0361b4bc4a40ecca5ebc1bdff368afa77176b9898eb8723` |
| `baseline/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts` | 57410 | `3467c9e18da5fd3563d0af02d2d4041e6dff63dbb3594f356fcb0ff02ee04c57` |
| `build.log` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `candidate/src/compiler/sys.ts` | 87022 | `13a115ad8d5f07875f8f5b694e5e85e7d5e241eb452fa2da52f165e9bbab7df9` |
| `candidate/src/harness/watchUtils.ts` | 9387 | `ab2ce0d341edb50bbf32634607a2efb251c0b9e3a80fe29ea68048169e932dd4` |
| `candidate/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts` | 57422 | `1ee73230ceebb962384b8d0fd2f3b474e85d66083e9a6ac1c8577250bc053ddb` |
| `change.jsonl` | 603405 | `36fc3a6323e1662aeeee526bd238931272031c6bced4c31f652125a2c1211812` |
| `change.log` | 117 | `f6bd7bdb8dc568f3185d21f029a1260f0b5fab6b216e95ab4fb987c0741272b9` |
| `delta.json` | 11745 | `1a34379e197f4777676aa1ae58032d71bf9c9cfff817955b527e9bd1b769c565` |
| `delta.log` | 3319 | `7c8abd3015b9756fe3e8f5535b06988a1844a0d4109540e2554b313caf793aaf` |
| `emit.log` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `experiment.json` | 366 | `600f77e133323da85a6d576e619e78eca4d6c0b08e37e51cdf68a3bbd0fd5ebb` |
| `javascript.json` | 112988 | `8a028264cdc9885fff3c7ece232dcf1068112985d0eca4ad283aab97709482ef` |
| `main.jsonl` | 599931 | `202e0af2775a7e7a47debd9ce877fb4f3e349331372aa48beb0107eec6d4175b` |
| `main.log` | 116 | `9e5041643f82debb6bb9b2a129f9376f582b7eb746ac8f7be1aa5e1118d06d45` |
| `mutant/src/compiler/sys.ts` | 86914 | `30876cbae0e2cb19eae2525cbed919abdba86fac921ecc22fa6aaad0441a3398` |
| `mutant/src/harness/watchUtils.ts` | 9363 | `78b02ee9cb8a5198c0361b4bc4a40ecca5ebc1bdff368afa77176b9898eb8723` |
| `mutant/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts` | 57410 | `3467c9e18da5fd3563d0af02d2d4041e6dff63dbb3594f356fcb0ff02ee04c57` |
| `mutant.jsonl` | 596553 | `390cc663000f9b4a6c513db7f3877052ab3d219657e82aee9031b19eeea66eef` |
| `mutant.log` | 114 | `9deab5d4fac411466811809d0b6aefc578fbba67058269ea77db075518def767` |
| `overlay.log` | 104 | `1e311f0df2e1276081992457cecddd1dc00498ad40bc8d9148fd7718868cdda1` |
| `owner-search.json` | 583654 | `ff7243c0429c652edecaeb6fcd09834148f582117ce5f1eae890c08e50a54781` |
| `pilot-adamic-types-exit.json` | 25 | `086dc5840af0823f6a62d6d3b975b21da10180cd428da67779dd8f7443e14bf8` |
| `pilot-adamic-types.log` | 454 | `911512eb6a670fd0f5988f4cc488d91048eaf90281b9089a55c9b01180574a40` |
| `pilot-change.jsonl` | 600170 | `cb857eecc3947e11d4d95769072ce584c12ea8bac516d2e6255977ebacbfbc3e` |
| `pilot-change.log` | 117 | `41bf3c405301b8a591fd270dc07f447360016ecf7089827773d829300f93abd3` |
| `pilot-delta.log` | 424 | `0db97f7b47a0196636aa161ef462d97d686526f2f4aa81a1c015c4cdf0a4e87a` |
| `pilot-experiment.json` | 519 | `a4d43890db8df6b2e8df98114fab3fe1925de4d45476387004b7292566ff7885` |
| `pilot-javascript.json` | 14120 | `1898584e0bab22c204a2d617b9b01f027f567dc28c0e311b2d9066d59752d701` |
| `pilot-stock-change.json` | 616 | `b4bf410f775387cf3c61b39559b4e47d68cafc7d042968dba47719a9cca34193` |
| `pilot-stock-delta.json` | 656 | `640649a72be839c27674e95c33fc7843ff334afa62c1b309995f531b8cbab84e` |
| `pilot-sys.ts.txt` | 86926 | `4b2024f9409704960d346e278c8d8bfa2b7391bbfdf5e177ff398d31645b63c1` |
| `reads-upstream.md` | 2293 | `c4e3681cdc90da8f91c484853e2aa49a1c4feb853ff988c7e540119145af4ef6` |
| `reads.json` | 4094 | `7a8d6c2e44feb2f74777b87ab5fc2eaba060baab89a59e949c09ae8db013e0f0` |
| `setup.json` | 522 | `b189b60ec4820f778f35a2fc155584c75ec9224449c7761b5f498f36dc8f5d23` |
| `stock-change.json` | 64 | `471dc276402b70481156d8842a952f2781d0541d42ede4c430f6069c19613b2c` |
| `stock-delta.json` | 104 | `45128211e5448bb60f1682355b7dbf9d77fe08c5c4db70516d270a808aa3a027` |
| `stock-main.json` | 64 | `471dc276402b70481156d8842a952f2781d0541d42ede4c430f6069c19613b2c` |
| `upstream-dts.log` | 851 | `c18fdca4fb7168d5926b43b66a1fd77160e46c6c55855f10f1aecdd038e6b819` |
| `upstream-npm-ci.log` | 1031 | `f85745cec5e2b0f20baec8fa062f91bfb18879328cab8c8b85567ac242be120a` |
| `work/stage3/census/watch-directory-1533/README.md` | 8231 | `9545299ee2f7bb43f2539abf16205ddb743a91015ed3ae2bea61bb4993a7d500` |
| `work/stage3/census/watch-directory-1533/SHA256.json` | 7649 | `f85adbcb20bb9fc7a6f43b80bfac725f1c2034325f73e1a700544d8315727043` |
| `work/stage3/census/watch-directory-1533/adamic-build.log` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `work/stage3/census/watch-directory-1533/adamic-types-exit.json` | 25 | `4e0f292d4e9b5c7ee955cf61c2b15099e9f18252da51c8d7934cede9f85fcacf` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/000` | 24576 | `6f97839fe4107293af1483519dcb441e8ffbfe811909466e254a4f6b6cdc0b8f` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/001` | 24576 | `811f8caa640a45a88a5002ae1c68956b40fc5a599d0998203ff0409abaecf2a7` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/002` | 24576 | `59e568acdc8f8a3a73ff38b8ca88510b297ca65257b72f4397abc86c9bf4502d` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/003` | 24576 | `dc0ffdcb71d6fd7a6e49aeeb1550d6bf6a3aba2f261206ec8e3f2362b20cfdae` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/004` | 24576 | `0cc85ee645bb328b64d125d75de1ac09894b0d9f303dbf10680464a516a12871` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/005` | 24576 | `81da6b748b47ea3b748d3959e3e3575fad035bfbd93d594224d73142b78a7030` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/006` | 24576 | `100ae43c8daf49f599899faa32c8d4463badc64ff731c99b3bf45c00769799a8` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/007` | 24576 | `08c6e8f9d4a6f2666f17219406bf2b3ea6ea82a49fe8e0e0cf23122dec24a25c` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/008` | 24576 | `68d1fb102c4809ef94103e84c9a45568f8058a9fa8ee13854d79c1593cd8b4a0` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/009` | 24576 | `c53784cc976a74d72df32bc19f1408e0b5aab1f8ab6bd897458da39bf3395e96` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/010` | 24576 | `9adafa34b45b867471ab01a8539ccc8ff055ac8ee202f6961e9003927b29828b` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/011` | 24576 | `71db4445d31e245769c0707a9598d379d8e60b824923880d97fc31120ecec6d9` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/012` | 24576 | `50dcd51363ba97416b511e224e68b99932a93ce0798c2b665cd2ade14e388e78` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/013` | 24576 | `705dafb8abe7c66bb5bc3a3fc638e4825cb450cc22eb8e50579d20810665ea12` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/014` | 24576 | `36a8e7b0e014fff009f341ad3b5fc20b173945c8bca02b50ebfc1e8ff96efb26` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/015` | 24576 | `ade0268eec6dcb8e7591e9b08d8f28753603dd3ac50178620b29b0ebbd303140` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/016` | 24576 | `e1df7d4b52a29856dc8974c7db9fa97261d3769e13512addf075f17a70056d5a` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/017` | 24576 | `600bc53618920949f3902aa12af66de942c4fe30d657ced1fb60c032e50b7ac6` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/018` | 24576 | `a23a36bd41de34f149c9302af92bb8d21560f97c5a80a4ad123fe0fa6a00d7d3` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/019` | 24576 | `ba45ef12493de7a9b32fddd4b45033b3dc0dabafde009bded9fdfd08d10a957a` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/020` | 24576 | `d3dd49711ba474d04aa522b41729af556556baa6331b2e29942f549be14330ad` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/021` | 24576 | `a67201edcf79a81e7347d88cdc18f254af0a288b73587e7e05db43fe9c8a9e75` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/022` | 24576 | `789c19488738c86ecb0ff8e4d2d55dc5d10ca0cc0f1009c3ced1d55e6dceb338` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/023` | 24576 | `e186b451fd045cccd1e7919a31cbc8b93b058ab483d988fe40b3d0daf559aa14` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/024` | 24576 | `4182d51a356f0c5b88242ec7bac5bbbeec81bd2e7b2d7d35879216969f330cdc` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/025` | 24576 | `0a5a7c429fd97a79f45b8bde78edcad0f2c50b4635192a4778aa61c8ad7e6400` |
| `work/stage3/census/watch-directory-1533/adamic-types.log.gz.parts/026` | 14951 | `88d5702ee63ed60dcc44748e27ca26284d0c5cc669ceff0f99974e971d82e4d4` |
| `work/stage3/census/watch-directory-1533/baseline-patch-set.md` | 4849 | `c8f4639b7431e219e252798531b93688924d9587d4f43f66561e276343ffebac` |
| `work/stage3/census/watch-directory-1533/build.log` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `work/stage3/census/watch-directory-1533/candidate.diff` | 7590 | `922290c8e66b738a77738ee3af3c7b727fc6769eb83a99f1e204037f4c79e09f` |
| `work/stage3/census/watch-directory-1533/change.jsonl.gz` | 26265 | `1659a0bffb428e3be803a132a062a03d4305f071073c1de8e87446e3ddf41da5` |
| `work/stage3/census/watch-directory-1533/change.log` | 117 | `f6bd7bdb8dc568f3185d21f029a1260f0b5fab6b216e95ab4fb987c0741272b9` |
| `work/stage3/census/watch-directory-1533/delta.json` | 11745 | `1a34379e197f4777676aa1ae58032d71bf9c9cfff817955b527e9bd1b769c565` |
| `work/stage3/census/watch-directory-1533/delta.log` | 3319 | `7c8abd3015b9756fe3e8f5535b06988a1844a0d4109540e2554b313caf793aaf` |
| `work/stage3/census/watch-directory-1533/delta.py` | 2395 | `232b3c1b89005a8d149602cc89b705c02887d7dab18935e90c74c4acf3776a10` |
| `work/stage3/census/watch-directory-1533/emit.cjs` | 1413 | `53f05018cc9c5cd3497bf430ee1cf3dca4c69ce1a71cf8cacc5db11d729ed358` |
| `work/stage3/census/watch-directory-1533/emit.log` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `work/stage3/census/watch-directory-1533/experiment.json` | 366 | `600f77e133323da85a6d576e619e78eca4d6c0b08e37e51cdf68a3bbd0fd5ebb` |
| `work/stage3/census/watch-directory-1533/experiment.py` | 895 | `39acb6ac8e7aec21000acd8535ab773b287c77750e49b7b7842c20435f684a32` |
| `work/stage3/census/watch-directory-1533/javascript.json.gz` | 30557 | `2c53c16ee944779b0c720cdaf0fcecf270f79869ab9830aea74866b455cb61f2` |
| `work/stage3/census/watch-directory-1533/main.jsonl.gz` | 26116 | `6b1ed44c439fd8604207a2fa8421b84a0a43e97c2ebcf5bc160eb9425752c0a7` |
| `work/stage3/census/watch-directory-1533/main.log` | 116 | `9e5041643f82debb6bb9b2a129f9376f582b7eb746ac8f7be1aa5e1118d06d45` |
| `work/stage3/census/watch-directory-1533/mutant.jsonl.gz` | 26101 | `a7b64c2a4db773374e33b8b5d083b0fb449a22eeb6a930977e458b6b8cb235bb` |
| `work/stage3/census/watch-directory-1533/mutant.log` | 114 | `9deab5d4fac411466811809d0b6aefc578fbba67058269ea77db075518def767` |
| `work/stage3/census/watch-directory-1533/overlay.log` | 104 | `1e311f0df2e1276081992457cecddd1dc00498ad40bc8d9148fd7718868cdda1` |
| `work/stage3/census/watch-directory-1533/owner-search.json.gz` | 87514 | `e867b45ffa0281fab4a201709cbee11bbb684dfcad979b6af3f747f87015ad9f` |
| `work/stage3/census/watch-directory-1533/pilot/experiment.py` | 740 | `52b2ff58b90598e1cee536cd9775414c9ec5432db2e3cd0679a5b3981ff308cf` |
| `work/stage3/census/watch-directory-1533/pilot/pilot-adamic-types-exit.json.gz` | 43 | `e264a5a9d30f70ded08f5cd4b3f53e482a16924aded0a80a4ab7b1197a17bc8f` |
| `work/stage3/census/watch-directory-1533/pilot/pilot-adamic-types.log.gz` | 176 | `82849aa7dfb60566cf5b29b47715d699544b04d22099fd226901ad1434257eb3` |
| `work/stage3/census/watch-directory-1533/pilot/pilot-change.jsonl.gz` | 26306 | `8f961a1a4ca61f8171ce43c42741d2ae99b17ce94421e365c92415365f5cbdbd` |
| `work/stage3/census/watch-directory-1533/pilot/pilot-change.log.gz` | 118 | `ba7878cfb832c0aacb242c63c20f917ef6dc2ff4c021a9ad24a7dd2e3f02dff0` |
| `work/stage3/census/watch-directory-1533/pilot/pilot-experiment.json.gz` | 268 | `e4c560c8ae902cca743bed47b35b07335e139e0e522c6211191f4aebc937b335` |
| `work/stage3/census/watch-directory-1533/pilot/pilot-javascript.json.gz` | 4074 | `4ecd80ad576abe87629e3bf79e813ff2880b8e376eb3472509d2228f92f322fd` |
| `work/stage3/census/watch-directory-1533/pilot/pilot-stock-change.json.gz` | 231 | `0fb71e23d4e6473b96a6b483701aa6ad0e13612d8ebfba88baf305d9ade903d6` |
| `work/stage3/census/watch-directory-1533/pilot/pilot-sys.ts.txt.gz` | 17538 | `c15d3e8a32c1361cf67a6cc8ee0d58e74b84b97889650f109338b2c017428a6e` |
| `work/stage3/census/watch-directory-1533/reads-upstream.md` | 2293 | `c4e3681cdc90da8f91c484853e2aa49a1c4feb853ff988c7e540119145af4ef6` |
| `work/stage3/census/watch-directory-1533/reads.json` | 4094 | `7a8d6c2e44feb2f74777b87ab5fc2eaba060baab89a59e949c09ae8db013e0f0` |
| `work/stage3/census/watch-directory-1533/results.json` | 741 | `aab50665c309f86e05ad3b252de383022d57a246715afdf2c234a76936e67ee1` |
| `work/stage3/census/watch-directory-1533/setup.json` | 522 | `b189b60ec4820f778f35a2fc155584c75ec9224449c7761b5f498f36dc8f5d23` |
| `work/stage3/census/watch-directory-1533/source/baseline/src/compiler/sys.ts.gz` | 17532 | `32aa3899d572f43d9a01eba0eef287f621652b91883120112f8c4855b1b2e40e` |
| `work/stage3/census/watch-directory-1533/source/baseline/src/harness/watchUtils.ts.gz` | 2067 | `625bbf5855aa9910db7571fb4deadfe313d9f579f7e48c43de4d0e7b4cdaebf1` |
| `work/stage3/census/watch-directory-1533/source/baseline/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts.gz` | 11596 | `ef0691697a5586c930af27a9ee845cf49627eb683c030a0f65be8d558ff6ce39` |
| `work/stage3/census/watch-directory-1533/source/candidate/src/compiler/sys.ts.gz` | 17534 | `e3a5c07025d2068b27ba1cfda6d7916d6b7770cf9ab62e41bfd652b03482a613` |
| `work/stage3/census/watch-directory-1533/source/candidate/src/harness/watchUtils.ts.gz` | 2070 | `05571a15414bfdf4deea07c6e34fc585702a0156aaacc79c6687565f5f8b95a9` |
| `work/stage3/census/watch-directory-1533/source/candidate/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts.gz` | 11598 | `1fbc551b0f2dd69c39d07cebb8d7d84ea20deea366c8e04c608db691e55f5b61` |
| `work/stage3/census/watch-directory-1533/source/mutant/src/compiler/sys.ts.gz` | 17532 | `32aa3899d572f43d9a01eba0eef287f621652b91883120112f8c4855b1b2e40e` |
| `work/stage3/census/watch-directory-1533/source/mutant/src/harness/watchUtils.ts.gz` | 2067 | `625bbf5855aa9910db7571fb4deadfe313d9f579f7e48c43de4d0e7b4cdaebf1` |
| `work/stage3/census/watch-directory-1533/source/mutant/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts.gz` | 11596 | `ef0691697a5586c930af27a9ee845cf49627eb683c030a0f65be8d558ff6ce39` |
| `work/stage3/census/watch-directory-1533/stock-change.json` | 64 | `471dc276402b70481156d8842a952f2781d0541d42ede4c430f6069c19613b2c` |
| `work/stage3/census/watch-directory-1533/stock-delta.json` | 104 | `45128211e5448bb60f1682355b7dbf9d77fe08c5c4db70516d270a808aa3a027` |
| `work/stage3/census/watch-directory-1533/stock-main.json` | 64 | `471dc276402b70481156d8842a952f2781d0541d42ede4c430f6069c19613b2c` |
| `work/stage3/census/watch-directory-1533/stock.cjs` | 1051 | `69035e7593418400bd8ae82275129b488ef9a3a82a6ee0ccc09b8debbb8c69d4` |
| `work/stage3/census/watch-directory-1533/upstream-dts.log` | 851 | `c18fdca4fb7168d5926b43b66a1fd77160e46c6c55855f10f1aecdd038e6b819` |
| `work/stage3/census/watch-directory-1533/upstream-npm-ci.log` | 1031 | `f85745cec5e2b0f20baec8fa062f91bfb18879328cab8c8b85567ac242be120a` |
