# watchDirectory boundary forwarder (#cyjqp5f)

**Built and proved.** Implementation branch `typescript/watch-directory-1533-1`
points to `e94db5d4b44873f03e3c091a1e6067feb45a9aed`, exactly one commit on
`e1161aed0f021d0bfdf8bc744c72d6ba537bc392` (the resolve-host-1612 chain).
The normalized forwarder removes the historical 1533 refusal with 0 new and
0 changed findings. The historical watchFile 1532 hole remains unchanged.
Nothing was submitted to Loom, and no ref was forced.

The whole unfiltered lane and its private counting copy both pass with 106,366
passing and only the sanctioned API row. **The lane counters record undefined=0,
true=0, false=0 (other=0): this lane makes no call through the real Node
watchDirectory boundary.** The counter is installed in the parent plus all eight
test workers; the positive self-test detects each value. The six concrete Node
witness cases exercise the three values with native recursive watch enabled and
disabled. Lane success alone is not presented as boundary-value coverage.

Only summaries and receipts are published here. Raw files stay in
`/workspace/scratch/watch-directory-1533-forwarder/`. Their exact SHA-256 and byte
sizes are in the inventory below and [RAW-MANIFEST.json](RAW-MANIFEST.json).
[checks.json](checks.json) is the machine-readable proof receipt;
[census-delta.json](census-delta.json) lists every finding delta in full;
[counter.json](counter.json) records the private count/control result;
[declarations.diff](declarations.diff) is the genuinely empty declaration diff.

## Implementation and pin scope

At getNodeSystem, the shorthand becomes exactly:

```ts
watchDirectory: (path: string, callback: DirectoryWatcherCallback, recursive?: boolean, options?: WatchOptions) => watchDirectory(path, callback, !!recursive, options)
```

The four names, types and optional markers match System.watchDirectory exactly;
there is no rest or default. The entire createSystemWatchFunctions owner is
byte-identical to the base, including its inner `recursive: boolean` parameter.
Only this boundary expression changes the adapted source and emitted JavaScript.

The required owner search found the 92/93/99 getNodeSystem pin chain. This amends
only 99-wasm-crypto-import's whole-owner after text/SHA-256 after the existing
require-result and resolution-host amendments. Its before pin and imports, and
the earlier 92/93 pins, stay unchanged. The original source/transpile scope oracle
is amended to undo this exact additional forwarder. The ledger has one
JavaScript-changing, value-identical line citing #cyjqp5f.

- Before owner SHA-256 (unchanged): `e45365ef731d397bee5f07cce554f8fc745c5028e614115ede71dfdbb4552941`.
- Previous after SHA-256: `d45d9bcd92b0de37cc6bf5d221c5292cdc1e1a349ea84a0c85ee92d3f48c7652`.
- New after SHA-256: `56dd252d2d6fdfcc264111196b92b5ca6a8567df4375ac70f3a91dbe887b18b8`.

Publication used the exact tested tree `a4739c82d3763e8646e68a23e2efa3b8d4bb9e4b`. Local receipts name
`b9f3938c6fd6ae1fdb7096048f2039a217966bf3`; GitHub publication changes commit metadata,
while the tree is byte-identical and the published commit's sole parent is the
specified base. The remote ref, parent, one-commit count and tree equality were
fetched and independently verified.

## Thirteen recursive readers

All readers treat undefined as false, negate/coerce it, or pass it to another
reader in this table. No undefined-versus-false distinguishing read was found.
The upstream reference is TypeScript v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`.

| File | Base/candidate line | Upstream line | Read |
| --- | ---: | ---: | --- |
| `src/compiler/sys.ts` | 1111 | 1108 | pass to fsWatch |
| `src/compiler/sys.ts` | 1129 | 1126 | pass to hostRecursiveDirectoryWatcher |
| `src/compiler/sys.ts` | 621 | 623 | truthiness branch |
| `src/compiler/sys.ts` | 623 | 625 | pass to nonRecursiveWatchDirectory |
| `src/compiler/sys.ts` | 1133 | 1130 | !recursive assertion |
| `src/compiler/sys.ts` | 1163 | 1160 | pass to fsWatch |
| `src/compiler/sys.ts` | 1211 | 1207 | truthiness selects watcher map |
| `src/compiler/sys.ts` | 1215 | 1211 | pass to fsWatchHandlingExistenceOnHost |
| `src/compiler/sys.ts` | 1269 | 1264 | pass to worker or timestamp worker |
| `src/compiler/sys.ts` | 1367 | 1362 | pass to worker |
| `src/compiler/sys.ts` | 1802 | 1782 | !!recursive normalizes before Node |
| `src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts` | 889 | 889 | pass to watchUtils.fsWatch |
| `src/harness/watchUtils.ts` | 184 | 182 | truthiness selects watcher map |

The concrete factory hosts are getNodeSystem and TestServerHost. The Node worker
uses `!!recursive` before passing the option to fs.watch. The test worker chooses
a watcher map by truthiness. The remaining branch readers are the recursive
wrapper, the `!recursive` assertion, and the fsWatch map selection; closures
forward into the same readers. Missing-entry/polling fallbacks add no recursive
reader. Literal false child-watcher calls are not reads of the incoming value.

## Full census and every delta

The baseline is a separate untouched worktree at e1161aed; the candidate is the
complete official apply of the amendment. Both load all 79 compiler roots and
attempt every eligible sys.ts unit, using CommonJSBundle with
`LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE=<exact adapted sys.ts>`.
Both have 136 units and zero checker diagnostics. Main has 181 findings and the
candidate 180. Both relation trailers have empty after-load/end overflow fields.
The measurement binary produces no usable IR.

**New: 0. Changed: 0. Gone: 1.** The sole gone record is:

- Kind: `Refused`; phase: `refusal_scan`.
- Location: `<tree>/src/compiler/sys.ts:1548:13`.
- Attempting unit: `<tree>/src/compiler/sys.ts:1473:5`.
- Reason: a function taking boolean seen as one taking boolean | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take.
- Full text: `<tree>/src/compiler/sys.ts:1548:13: Adamic 0.1 refuses a function taking boolean seen as one taking boolean | undefined (tsc relates a method's parameters both ways), so it can be handed what it can't take; write the method as a property holding a function (handle: (animal: Animal) => void), which tsc checks one way, or take the wider type in the method (method-signature-style)`.

On this newer base, historical 1533:13 is **1548:13**, and historical 1532:13 is
**1547:13**. The targets are resolved by the properties on nodeSystem, not by
assuming an old line number. The watchFile refusal at 1547:13 is identical in
main, candidate and restoring mutant.

The comparison normalizes only scratch-root paths and excludes mechanically
shifted numeric byte offsets from record identity. Unit, kind, location, phase,
reason and full text remain in contextual comparison. Original offsets and every
raw record remain in the retained streams; no finding kind or phase is filtered.

## Mutants

- **Restoring shorthand:** returns the target refusal at 1548:13. All contextual
  findings equal main exactly, and the watchFile refusal remains present.
- **Unnormalized optional-contract mutant:** passes recursive unnormalized and
  propagates the optional parameter contracts through the same reader graph as
  the earlier rejected widening road. This has zero checker diagnostics and
  restores phase `lowering`, kind `NotYet`, sys.ts:1215:82, reason
  `a boolean | undefined variable a function value captures`.
- **Single-token normalization deletion control:** removes only `!!`, retaining
  the unchanged inner boolean contract. It correctly yields TS2345 at
  sys.ts:1548:159, because boolean | undefined cannot be passed to boolean.
  This checker failure is kept separate from the optional-contract census
  failure; deleting only the two exclamation marks is not claimed to produce
  the 1215 NotYet. All mutation modes and source scopes are explicit in the
  committed mutant helper and retained mutation receipts.

## Checks and Node witnesses

| Proof | Result |
| --- | --- |
| Stock TypeScript 6.0.3, all 79 roots | 0 main diagnostics, 0 candidate diagnostics, 0 new |
| Ordinary adamic types, all 79 roots | Exit 0 |
| Official stage3/apply.sh, base and candidate | Both exit 0 |
| Source scope | All 747 source files compared; only sys.ts differs |
| Stock emitted JavaScript | All 604 emitted files compared; only the forwarder differs; undo restores exact bytes |
| Inner watch factory | Complete owner byte-identical to base |
| New watch-directory owner/parameter contract | Pass; exact System parameters; LF/CRLF; drift, duplicates, mixed/partial input refuse; zero-write replay |
| Prior require-result contract | Pass |
| Prior resolution-host contract | Pass |
| Original complete crypto-chain source/transpile oracle | Pass, including comments kept and removed |
| Upstream own npm ci and hereby dts, both trees | Pass; both declaration files byte-identical |
| Whole unfiltered normal lane | PASS; 106,366 passing, one sanctioned row, zero pending |
| Whole unfiltered private count lane | PASS; identical counts and sanctioned diff |
| Counter positive/disabled-control self-test | Each of undefined/true/false detected for both emitted aliases; disabled control is zero |
| Ledger | One JavaScript-changing, value-identical #cyjqp5f line |

The Node witness uses the exact stock-emitted forwarder, complete actual watch
factory, adapted multi-map helpers and actual Node fsWatchWorker. It controls OS
event delivery and compares registration, callback and close traces to the
original direct call. Every following case passes:

| Input recursive | Native recursive support | Trace equality |
| --- | --- | --- |
| undefined | true | PASS |
| undefined | false | PASS |
| true | true | PASS |
| true | false | PASS |
| false | true | PASS |
| false | false | PASS |

This is a bounded witness of the reader/value behavior, not a statement that all
OS failure paths or function-identity/stack observers were exercised.

## Declaration bytes

Both builds run upstream's own locked install and `node_modules/.bin/hereby dts`.
The retained and published declaration diff is empty (0 bytes; SHA-256
`e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855`).

| File | Bytes | Identical before/after SHA-256 |
| --- | ---: | --- |
| `typescript.d.ts` | 593536 | `f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516` |
| `tsserverlibrary.d.ts` | 866 | `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd` |

## Whole lane, private count and control

Both lanes use all runners, eight workers and no test filter. Their only failure
is `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`,
with only `api/typescript.d.ts` in baseline differences. Both official lane guards
return PASS, with zero timeouts or unexpected exceptions. Their sanctioned diff
is byte-identical: 71,315 bytes, SHA-256
`e990045d4345f8de90d253f87d432bc19ffe274e4deaee90e4ab81eeb97e5c34`.

The normal uninstrumented whole lane is the control. A distinct private lane tree
uses a scratch-only preload hook to count at the exact emitted forwarder. The
hook recognizes the bundler's path3 alias and preserves all four parameters,
normalization, return value and call order. It is installed in nine actual
compiler-bundle copies (parent and eight workers). Counts persist per process
and thread at module installation and each observed call, avoiding cleanup loss.
No hook, count call or runtime instrumentation enters the implementation tree.
All 747 source files in candidate, control and counting trees are byte-identical.

| Incoming recursive | Whole private lane count |
| --- | ---: |
| undefined | 0 |
| true | 0 |
| false | 0 |
| other | 0 |

The final positive counter self-test observes each value twice (original and
bundled aliases). Its disabled control records zero and returns the same values.
The alternate harness VM wrapper evalFile has no executable call site in
harness/testRunner; evaluated fixture snippets are separate from the loaded
compiler bundle. Raw per-process data, the hook, loader audit and self-test
artifacts remain in scratch. Only the final acceptance self-test and runs named
in checks.json are acceptance receipts; earlier preflights remain raw history.

## Retained raw artifact inventory

Paths below are relative to `/workspace/scratch/watch-directory-1533-forwarder/`.
All listed files remain local; only metadata is published. This inventories all
scratch-root proof files, the generated measurement overlay, private counters
and preflights, lane/oracle raw logs and receipts, selected exact source snapshots,
both declaration pairs and the pre-crypto source archive. Dependency checkouts,
full adapted-tree build/dependency outputs, caches and disposable worker
workspaces are excluded execution inputs. Publication summaries and the inventory
itself are excluded to avoid self-reference. No raw file is silently replaced by
a summary. Full absolute paths are available in RAW-MANIFEST.json.

Retained artifacts: **211 files**, **137,970,330 bytes**.

| Scratch-relative path | Bytes | SHA-256 |
| --- | ---: | --- |
| `PROGRESS.json` | 773 | `111853d1b6428e867873f785c0aae8156b5c6a682c29263fa4cf8d989fd7f823` |
| `adamic` | 44160020 | `e724ecfbb0d77497303f3d34b305d821fdea40617b732d368db7fc9af3265ab5` |
| `adamic-build.log` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `adamic-types-receipt.json` | 8602 | `7216d30e34f37cdcc6bea032085a3a6fc7c876c63d71f7e2d3b86d9b43d0ba62` |
| `adamic-types.log` | 8382588 | `690345a3dca619e6e566b9605865a03fb1c2e873d13bb3c853cebe01e21e7e1b` |
| `api-npm-ci.log` | 232 | `312abe6df6d367f5fdddce4ca41919fed54c934bb1b0beecc01997274fe7535b` |
| `apply-change.log` | 844523 | `1131e7d177afc6d351915e6e6fa1578561717ca693aefccbe6dd6fd869ad0757` |
| `apply-main.log` | 843972 | `1600ec1a4bcf3dba6252309ff60b6b144c150ff9d3b139ff51034de0ed0f8491` |
| `base-worktree.log` | 2773 | `42058945ce2d3ec6f26c9c15d001faacaccbbcf1e5c8e460a06b9c08ffb6d9f7` |
| `baseline/built/local/tsserverlibrary.d.ts` | 866 | `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd` |
| `baseline/built/local/typescript.d.ts` | 593536 | `f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516` |
| `baseline/src/compiler/core.ts` | 93355 | `6925a744b191c13ebe1425ec76a994b51cb2ab1c81f19b2c7958a4ed1e89311b` |
| `baseline/src/compiler/sys.ts` | 89088 | `a8a76f3bcca5ab8f670a7b89377531f85df078d96210ac0b429c5c3ca76c4669` |
| `baseline/src/harness/watchUtils.ts` | 9363 | `78b02ee9cb8a5198c0361b4bc4a40ecca5ebc1bdff368afa77176b9898eb8723` |
| `baseline/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts` | 57410 | `3467c9e18da5fd3563d0af02d2d4041e6dff63dbb3594f356fcb0ff02ee04c57` |
| `candidate/built/local/tsserverlibrary.d.ts` | 866 | `cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd` |
| `candidate/built/local/typescript.d.ts` | 593536 | `f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516` |
| `candidate/src/compiler/core.ts` | 93355 | `6925a744b191c13ebe1425ec76a994b51cb2ab1c81f19b2c7958a4ed1e89311b` |
| `candidate/src/compiler/sys.ts` | 89241 | `78a61566d287c8349c378f9765a6d5e938d39d038192093a3c68b989174eba77` |
| `candidate/src/harness/watchUtils.ts` | 9363 | `78b02ee9cb8a5198c0361b4bc4a40ecca5ebc1bdff368afa77176b9898eb8723` |
| `candidate/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts` | 57410 | `3467c9e18da5fd3563d0af02d2d4041e6dff63dbb3594f356fcb0ff02ee04c57` |
| `census` | 37038553 | `8efe35639edf121f2cd2be7ae0d1ec5786ec8f757f3a4d82085b970ead9dc9aa` |
| `census-build.log` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `change-receipt.json` | 750 | `9cfc8c8f02041f9cc0e52e19bb7a30f859faf92f0814c4efc76f0718afa3155e` |
| `change.jsonl` | 628729 | `7f5e607bf4e60587731f912d082a4ef8bad6b412ad5e11014450612eca1a506d` |
| `change.log` | 127 | `6d4194befd2da1e58a28db845369b19a18d8c7a7cd54d5c561579e52c65b121d` |
| `count-hook.cjs` | 1763 | `7c4c89c553fb27db9124169ad5ad7c5d493ca88991514b4455913afd8bfe19de` |
| `counter-load-audit.json` | 613 | `b7392b17f1c6afa42f6a9afb34fdb039be53062d5d1b8af2f6d8caecde765894` |
| `counter-selftest/control/19476.json` | 170 | `ce767fd0a4ce50e8f1ccc04c25adcb0384adb83252917335a59122acaa6ee746` |
| `counter-selftest/count/19465.json` | 168 | `4d090ec3ccb289996bb86b6584d1c4944b638d621911942a445003658c785ab7` |
| `counter-selftest/receipt.json` | 99 | `7c21089b874c7512db7a41f4420853127a794081db86d93efbac97be5f08bbdb` |
| `counter-selftest-final/control/26512-0.json` | 262 | `eb802c46f608fd9fe91e193785c2afea2e55a758aa6fb8f9bf3753361646b43b` |
| `counter-selftest-final/count/26505-0.json` | 260 | `139dee51885bce05d431b5207f02a5d07529944ddfb5770a8077e6f8b62101d3` |
| `counter-selftest-final/receipt.json` | 172 | `a618ca9995c9abfdca25940a07fc35979dd09a47484e1d7192c9a98cf6320beb` |
| `counter-selftest-v2/control/26238.json` | 103 | `bfa3ab04e7eeb44dd66513cbe9d37f401e2ff9d3fb96868d7c94caf6d6c392dc` |
| `counter-selftest-v2/count/26231.json` | 101 | `40b38407e2b2db36bb7226626b8b38dab809ae477b8205171e0c0599dd554eb0` |
| `counter-selftest-v3/control/26257.json` | 249 | `e1429098a8c95a5d30e23dcd2fe9e6fce37680812b9c872ea09ed685c16ad6f4` |
| `counter-selftest-v3/count/26250.json` | 247 | `25b7a0fe8b8bbf76a732a3901ffa0e0ebe16e8a2c0650e02e4d09bce15cf2894` |
| `counter-selftest-v3/receipt.json` | 136 | `2b4635aa0faaeac4526836663a7c76286e55028262345e744ea02a637e53027a` |
| `counter-selftest.cjs` | 567 | `06df04a293e9b11acc85bbe853f8651e67a25de41b4419f200a7232f9d7145cd` |
| `crypto-chain-scope-receipt.json` | 834 | `7b9704d12d272de85ff5adb8d805f4440ac2c02c5632648e64d0572340174323` |
| `crypto-chain-scope.json` | 186004 | `436aa462d09c82fdf909eeb1e9048c9931dac6526e9c82b1e90a4aa159f371c4` |
| `crypto-chain-scope.log` | 716 | `747f1b07537c990bc531e6368aa4a69a798ac6bc1e279fdc37ad77f6588eddc9` |
| `declarations.diff` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `declarations.json` | 668 | `886a36584823bd11e644e4b6dd3232a455b2bf9693fbaeb40f563948f52e3433` |
| `delta.json` | 7114 | `d3ba99d8d2a21a7c09402d1a6164e7cdffbfb6cc11ef350063857f0c54f699da` |
| `delta.log` | 948 | `2662751fc122771eeab7bfedda85b003271f99075ccda113a6fa8162083ce43c` |
| `delta.py` | 2544 | `8adbe17f1c83e3185abc884b76914095942f452a6df1c704f668b156a77973ae` |
| `dts-change-npm-ci-receipt.json` | 573 | `a54d96f588387e2e3eae575e88a1466d1b2f7566baba4cb2a6e93e4d121536cf` |
| `dts-change-npm-ci.log` | 823 | `c23f5fa6cdf2f8e545d5644d1a8cee43839adb640cc8858ff2bb3d784c7d27a8` |
| `dts-change-receipt.json` | 589 | `429f0a351a482d0dc9eda343ae47c4a9c20ac6b165288ab2909a3003ca70e920` |
| `dts-change.log` | 867 | `9148bbe6ae5e75b26ece007db015d8ce7528c6bf8836a0ebbc2f25845f8c71b9` |
| `dts-main-npm-ci.log` | 823 | `2b16d9870ff33e18e2f9fe6f2afa91f09252d22709c6bb0202052a743603e476` |
| `dts-main.log` | 859 | `1f8c351ac22b3b7f03904ceb73b237c311d22fc01aad8fcb8079ea33286fab5d` |
| `env.sh` | 583 | `4e1d9d9ec93f0ee855cafc6692294f4ba43634c53bfccd607b7887ddc86dcae3` |
| `implementation-publication.json` | 374 | `4e0ea21e1a64e8c0d8d1be4ce9ab4c0b24f08ff14617aaa6a75d92f9196146ac` |
| `implementation.diff` | 92771 | `93901563504828914d3d08960ef22b8a2a23e4ff46d65360fc9b4ad2e7a342de` |
| `javascript-receipt.json` | 619 | `dbff47e7df3757d38e65518be6b351a89bc667b17af8603ded336c034e0d0644` |
| `javascript.json` | 158142 | `e3f530fa73cb1461c2156f76b88c4cb9335ed4e59a8510a6fd2c41043d5039d2` |
| `javascript.log` | 142 | `f51b6ccf1c0ebe11273b2e96940eb7139ecc8474cb964c04a7d8f449e38f906a` |
| `lane/apply.log` | 68 | `b56a9ba8803c52940bc5789fc310ec43b91786ee9b8a9e8bd68a1a2793dd44e8` |
| `lane/execution.json` | 207 | `db898ada302f11c70377dad4936b31be02205247b7064c1d203b1cf87f975bfb` |
| `lane/failures.md` | 82183 | `0805022c81a0234975e15db8eb2818e8e45886de8dcff04a732baf9bc24dd3a2` |
| `lane/oracle/baseline.diff` | 71315 | `e990045d4345f8de90d253f87d432bc19ffe274e4deaee90e4ab81eeb97e5c34` |
| `lane/oracle/build.log` | 2856 | `5fc6044ddd659953a4521800bd0f46ec8d789a2990c3e23104f5dcc9142de3bb` |
| `lane/oracle/failures.md` | 82183 | `0805022c81a0234975e15db8eb2818e8e45886de8dcff04a732baf9bc24dd3a2` |
| `lane/oracle/install.log` | 822 | `a357a0c5f25c663f2b739c7f8f04cfcc20aa4dd13ad50ee06beb370806a65669` |
| `lane/oracle/mocha-errors/26098.jsonl` | 165257 | `9f0d20c4a5f84828cca4dd69c11739432009accc81b14cf4fcad8271bb73a7af` |
| `lane/oracle/report.json` | 169526 | `c492eed22a0c4f70c2433ad1f32c95d263adf50818ebd659f853afa9ff9b98cf` |
| `lane/oracle/tests.log` | 82857 | `5a7e42fdb5f5fc898c7b3b75c8c4429daa084fae31f62ab1cdda24ccc51c2555` |
| `lane/oracle.log` | 169526 | `c492eed22a0c4f70c2433ad1f32c95d263adf50818ebd659f853afa9ff9b98cf` |
| `lane/patch-set.md` | 4850 | `214e931c5194f2d32543553a5cc30879293949b67d7c6752c3137f44e13f1012` |
| `lane/report.json` | 169623 | `25a0dba7ca03025447cfb0ecc5f781d256eebc8210f98c759d4d5d39488464bc` |
| `lane/verdict.json` | 168797 | `67c6c55606382403fda63010334e6eed122014b1eedddff1634931b0c79b1729` |
| `lane/verdict.txt` | 107 | `b9ae60564612ebd5a95341f85a73da2667f23c5f91d7b6cd5db80c6dd5f22f89` |
| `lane-count/apply.log` | 74 | `efc691787fa054957dc1e9e2bb62ca822a509f8d6585ff3edcf83e8bac0dad05` |
| `lane-count/execution.json` | 207 | `cd0879c6e580e3e089a0904711a78092362470d7265adace7d31ef17b44af682` |
| `lane-count/failures.md` | 82183 | `0d8a8ea3cc06790af6f36ec35ca6cfcdcb90e76bb5f14def5d183689c69095c5` |
| `lane-count/oracle/baseline.diff` | 71315 | `e990045d4345f8de90d253f87d432bc19ffe274e4deaee90e4ab81eeb97e5c34` |
| `lane-count/oracle/build.log` | 2858 | `c6116046250687a709d5f3016af4980b5922eaf52d0fe77a61d3c3c2e173ce2c` |
| `lane-count/oracle/failures.md` | 82183 | `0d8a8ea3cc06790af6f36ec35ca6cfcdcb90e76bb5f14def5d183689c69095c5` |
| `lane-count/oracle/install.log` | 822 | `a7559adff83103c8eb61f1d959aee27962f2048ad1b4fe6aac124cfdf4457266` |
| `lane-count/oracle/mocha-errors/26812.jsonl` | 165257 | `7e050fa3101699fab3ef93d9d77e9f5e846584b51c501f938b5b0a464a278a99` |
| `lane-count/oracle/report.json` | 169530 | `dc57954f390544b4e31380ccd9f4f395bd06da3fffb9962dd0fe117bf80dd40e` |
| `lane-count/oracle/tests.log` | 82859 | `06dbd02ad75f8e26f2f9d871d1d8f5ab71536d1f81b41e28f34b07d75e1d56d9` |
| `lane-count/oracle.log` | 169530 | `dc57954f390544b4e31380ccd9f4f395bd06da3fffb9962dd0fe117bf80dd40e` |
| `lane-count/patch-set.md` | 4850 | `214e931c5194f2d32543553a5cc30879293949b67d7c6752c3137f44e13f1012` |
| `lane-count/report.json` | 169629 | `35298b985f70c57b25ccdcc03a42cf4834dac4a83b61c5552dc3641cb15b2218` |
| `lane-count/verdict.json` | 168797 | `f47046165449d05805ff09e0da7d42cb2bbb57ba2b8253fa20cfe2ab83bd47fd` |
| `lane-count/verdict.txt` | 107 | `b9ae60564612ebd5a95341f85a73da2667f23c5f91d7b6cd5db80c6dd5f22f89` |
| `lane-count-receipt.json` | 863 | `b21f27e1ec840d29750efaef3378aff03054fc4c501350d049a67f4da91147cc` |
| `lane-count-summary.json` | 1137 | `32c90787b7cb53960d27028fb724fd7861e3a2dda44dcf0d9422dff7619eeac7` |
| `lane-count.log` | 168 | `cd75d890f24cf84ba64a8a986322e6f99731fc8b08fc858b3754460605b9cd87` |
| `lane-receipt.json` | 623 | `84c6fad27fa7523aae6e53fe3317971e28adddf56c30baf9dcc46f0844a5f7b5` |
| `lane-source-identity.json` | 161 | `01111f3258e441213b2a1c518c1afb61c7dded0e9456bb9dff9bc61263ced97f` |
| `lane.log` | 162 | `fe3e729c8da4f818230a9ab16baa39902daae45957703224cf007fada0e01eb3` |
| `main.jsonl` | 627063 | `b2d337de98916d63c3881d9e02d136e8056be74dc2fb762dcd092b403c00f6c5` |
| `main.log` | 126 | `3af0b34cf9707d2549eb061ccf13a86a2addd1a307b63dce867d6634969c37fd` |
| `mutants.json` | 1224 | `1d6a33bf2a7700fc82d7fc862d4f91d481d7741d6a12ee158f0f3c1f8e2e4e88` |
| `node-witness-receipt.json` | 697 | `4c04d4215792a18f01b9d3805db461228052216229a3a99aa85d812e0b12900a` |
| `node-witness.log` | 2542 | `f735ed73697723f845bce2d385d7eff2e7fce154640bc253c1c86bc60c794e8f` |
| `optional-contract/src/compiler/core.ts` | 93355 | `6925a744b191c13ebe1425ec76a994b51cb2ab1c81f19b2c7958a4ed1e89311b` |
| `optional-contract/src/compiler/sys.ts` | 89347 | `026642c93ffe772d8521a3ad116703e8c7ec82845c3933c1b5d141af1265acba` |
| `optional-contract/src/harness/watchUtils.ts` | 9387 | `ab2ce0d341edb50bbf32634607a2efb251c0b9e3a80fe29ea68048169e932dd4` |
| `optional-contract/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts` | 57422 | `1ee73230ceebb962384b8d0fd2f3b474e85d66083e9a6ac1c8577250bc053ddb` |
| `optional-contract-mutation.json` | 97 | `76a4191fc37dd5265bef349ad538f72d234060bf80220d2d29007f798a28ba27` |
| `optional-contract-receipt.json` | 781 | `937b4ddb49484dbd05aa82cdbe2838fa2373cd0cf0c6248edf8980de69417d6e` |
| `optional-contract.jsonl` | 645320 | `5e7d6580cda0dc4f4cdf3970389e302ed66f62579603e0ead96924d5a20f8093` |
| `optional-contract.log` | 135 | `8eba83ff818889058a03f56c7cd8fa5f6680b6eddb122dbb4c8d4f2466a6c471` |
| `overlay/internal_load_latent_hook.go` | 3179 | `154155d62954f677731f302b7dc57a3aab914ff373618fd9efc0a1bf3d68b5a4` |
| `overlay/internal_load_load.go` | 13520 | `fb7e14fd673f689330564b6486c5db6e0722bf5b3a8fab103322bd7d49f3eb07` |
| `overlay/internal_lower_entries_records.go` | 32835 | `50678750c15518faf7d614399542e2014999207a72da48f3f453dfac38900040` |
| `overlay/internal_lower_expression.go` | 96659 | `7eb1c5c94bef9b6ee6db36983061bcec4bdf92d3e662f2c9a86c02c796498259` |
| `overlay/internal_lower_functions.go` | 23616 | `ddcf50645205fc8686bc21bdd6220087965e6463336a2067ce377a5c8b2fd820` |
| `overlay/internal_lower_generic.go` | 18709 | `65fd3364373ced517d0543a94007a37d2f87e45223f20945577b43e8d81a2ed8` |
| `overlay/internal_lower_latent_deferred_probe_test.go` | 2438 | `a59311fe81ce15b510fe53648f7fc6a2c3004061e44d718b2105a46cdb96ea6b` |
| `overlay/internal_lower_latent_full.go` | 7833 | `99b8f7fa928e3dd6884068b9c858e4d8c9a65eb1e3b7a5c2ee0f71c19f45342f` |
| `overlay/internal_lower_latent_hook.go` | 23881 | `abd73e4a70cdaf0f8fccb857e434a92d43bef69864e982ea841ebafde74df830` |
| `overlay/internal_lower_latent_pass_ledger.go` | 3933 | `14b593e1b2c84ae6f14e290ae3d6c6fd0cdf3be5abcf4cb2cfde75bb1e49f41e` |
| `overlay/internal_lower_latent_piece.go` | 4991 | `ae04c70b5cd7c746e9f2ffdd61147ab8d92056bea29342b5be9feee4fd84afee` |
| `overlay/internal_lower_latent_progress.go` | 3587 | `bdb2684110c5e6c160909c0d36cf975778a9b8f8f46273df8b3d75729be494c3` |
| `overlay/internal_lower_latent_relation_overflow.go` | 8424 | `1630d2346899454d170faff069c67e6cdeea959cba13477e4a8a3d6026e4cf63` |
| `overlay/internal_lower_latent_replay.go` | 4197 | `4dd0d8bc81da16d543f3d278d53a7e67307d3113df8f3b6611fe22c62e674b1d` |
| `overlay/internal_lower_latent_scoped_tail.go` | 4617 | `4c84db89648acc9407962025de626bf06615cd7061da57b86628e0ad5c99767d` |
| `overlay/internal_lower_latent_scoped_tail_test.go` | 3804 | `d934c6fe78b5183d756b5b24f6be568ecf28b80f819ab0fd54003c1696f83d0a` |
| `overlay/internal_lower_latent_speculative.go` | 21161 | `6fae169360682ca61b5338feee8fa325d45df9ad4d3fe26e776790ec0378a73e` |
| `overlay/internal_lower_latent_state.go` | 83465 | `83e4abdb5ba91f54ce9d19c390687332f82f69f89cbfdab5bf68b6bb1904d9c0` |
| `overlay/internal_lower_latent_timing.go` | 1241 | `080b7937b2ec03c1837ec639a7916056fbf2c439f5491c321fc972899f41ad3f` |
| `overlay/internal_lower_latent_units.go` | 11056 | `fa099b0021a1541d6969d797274f8709b12d03356429adb5a9a692a45d685b11` |
| `overlay/internal_lower_locals.go` | 17723 | `fa21e0661addb989a3b89a418814a1e7ff8af9d6f89040fc7293c6bc83034602` |
| `overlay/internal_lower_lower.go` | 23464 | `dc835339aba092d393011ce9d338afe632c0df03c74767def36e5c8fc93a2c36` |
| `overlay/internal_lower_refusals.go` | 15071 | `35e3116e30afb2174e1630f2fbbe2649ac41c508ee2edc7676f53da03d387cb5` |
| `overlay/internal_lower_statements.go` | 14567 | `bcfc857a47f0257238d3d3a39096d7aead73f8021c2344c077f2923691be365f` |
| `overlay/overlay.json` | 5131 | `d87d38c509a38f77640a26fd077c1203b6862f450adf056d42ad834232d8f108` |
| `overlay/stage3_census_latent_replay_worker_main.go` | 3270 | `70d936e625ff77e3e1243061f73b60544abf9d17e30b0a6118448283d563b2d5` |
| `overlay/stage3_census_latent_tool_main.go` | 3263 | `64647fa6e6113201cd276b12011b8c6760926061759545e30e7277a3108f6966` |
| `overlay/stage3_census_latent_tool_relation_overflow_test.go` | 7735 | `e63247a585d0722fcbfe1a9f32b9f4e672d48f6f4520b7459100b94d61907c6c` |
| `overlay.log` | 71 | `54f9437b5337d1dc0715afb32e110e262d4e023aa882482d441b99fa6d5a5849` |
| `owner-search.json` | 61456 | `5583d8cd1c8b9815ff202b8b5a819b655c82119366eec0c6ff65bff744011cfb` |
| `pre-crypto/src/compiler/core.ts` | 93355 | `6925a744b191c13ebe1425ec76a994b51cb2ab1c81f19b2c7958a4ed1e89311b` |
| `pre-crypto/src/compiler/sys.ts` | 87217 | `99054c5bc079cc046dee7ada45a3bbda9a59e2fdc66c56b194461b9a336ba5e4` |
| `pre-crypto/src/harness/watchUtils.ts` | 9363 | `78b02ee9cb8a5198c0361b4bc4a40ecca5ebc1bdff368afa77176b9898eb8723` |
| `pre-crypto/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts` | 57410 | `3467c9e18da5fd3563d0af02d2d4041e6dff63dbb3594f356fcb0ff02ee04c57` |
| `pre-crypto-src.tar` | 37488640 | `fead84e129670ee48e085d2a1dccf16de434d890049b2885e8426799f8731041` |
| `primary-census-summary.json` | 778 | `7d7433677b2db2080375da9dffdf635c0c920eaccf9aaab669dcca1978499a44` |
| `private-counts/26524-0.json` | 114 | `88f9c59a05396e1dc3180f10aba2b001aeb4919cf8901851e7bbea7b995bff5c` |
| `private-counts/26561-0.json` | 114 | `7250687d3988e9ba0fbb4f80708537040c0bd650301632a87f538ffd1fa9a68b` |
| `private-counts/26574-0.json` | 114 | `3e9f1a0b3328b14d2ea6bc3194682e99c8eebb54ff7bba1abd96f014072301a7` |
| `private-counts/26576-0.json` | 114 | `2dea7821bee7f5373bc7036a7b72240770dac184f2a9c556e6498056e98d949e` |
| `private-counts/26598-0.json` | 114 | `f151c19a371429ff3a85b16aed2fdeb9c2a0d2aa3977f64a202a191d4f672a15` |
| `private-counts/26610-0.json` | 114 | `6b11fadfb630ec105bfe79a353a65490228e3fd557b0b7e5d23dae44435a7e07` |
| `private-counts/26622-0.json` | 114 | `c23af5237d35030ef480862cc0bd9b165727408043aeb84ed03f34945e5e1294` |
| `private-counts/26629-0.json` | 114 | `76b90a8a492501822a71b1b57170a1547ea1e7dc363a827f60d11f6e884027c4` |
| `private-counts/26651-0.json` | 114 | `f549c31d9aaa95b9219ecb415b17055227cfb0a733d508f204a439d9e86c46e8` |
| `private-counts/26657-0.json` | 114 | `4f6ec3d1aae13d8ea78d2132680eeb737d2402f8e27ba9c20f7a54aa509a0b5f` |
| `private-counts/26668-0.json` | 114 | `1430549ca81f7a41418b44433d64db4332c70de2bdde2765be323b7e9bd54ca7` |
| `private-counts/26697-0.json` | 114 | `97a747cad658bf90e90c86c950e74fc6981a3f596e11bc3318e02b515faee164` |
| `private-counts/26704-0.json` | 114 | `2ec348172d02715eb5387c2a29fe2b79c5c29054d654a6432f3fafd34a585e1b` |
| `private-counts/26716-0.json` | 114 | `851b2c1962bd01c0169cbabde6cb4905f621bef3b32ef7cab91ffe6cc0eadd90` |
| `private-counts/26727-0.json` | 114 | `df9599320faa58413509e8c93c223336c201ed660961d273240111c40d28f039` |
| `private-counts/26747-0.json` | 114 | `a7e4dfcd036f29244f73ec3c5880da596e833bd485714b5fcde636f06482efef` |
| `private-counts/26755-0.json` | 114 | `72aea072ca802ae18c54f341af4a7ca398001469ebe8dd5bc69be8b23b0df93c` |
| `private-counts/26767-0.json` | 114 | `63cc2433cf6f301799e42e62673c3be62dd420204f9b010efb04f9a5c7f33d87` |
| `private-counts/26778-0.json` | 114 | `e1f1a086701d87cb4f80c89a5eff07e3a107aff4803587980feb982dde377c42` |
| `private-counts/26798-0.json` | 114 | `1a1b35e1e6fe1d3ee86e6edacf644c432edb75c59acd452da9111de957406ec8` |
| `private-counts/26805-0.json` | 233 | `97bd3943bf4ac0786ebfc09dde2c21f392f4a3c7b815bcb6602ee9c2387d8720` |
| `private-counts/26812-0.json` | 233 | `fc966c985201fd76d0472c37c0a9ebaea99e85f128dad171d5b0b019d7459445` |
| `private-counts/26813-0.json` | 233 | `be64068e9818c24a3d55224e0268302198eefb436cd2eac0481b0d2fdcde3e50` |
| `private-counts/26815-0.json` | 233 | `3ab4cff1402cd3fe5d0041672fdec12144ea9be2b90680727b5c733ca46ced90` |
| `private-counts/26820-0.json` | 233 | `a0f3430c812700fb26fb6d63204c664f6c35da643fc98d3a053b94598b7e5932` |
| `private-counts/26826-0.json` | 233 | `e088ee5075d3f94a8ebd56397396cbd62f7cf9ba51eb3b866aec21d42e4822a4` |
| `private-counts/26827-0.json` | 233 | `6039b7db5c7473833b94c63035fa0f40e59faa904344fcd5d956bf0a0a064e00` |
| `private-counts/26833-0.json` | 233 | `bf40c52ea95f22c0505099e3810ea106e99b5ea01351234b4289b35450dbb8e5` |
| `private-counts/26834-0.json` | 233 | `f0f66063691e895c07a09b0ade0472ab5caac8c281fbd42ade268956552b5f03` |
| `private-counts/27051-0.json` | 114 | `ad551b570f09e68146e50797ddbb82def672b8657c39fccf8749ce55a338c125` |
| `proofs-receipt.json` | 12614 | `0ee882c16f397991319d699c0a81b8f0bd19de728b71b360cf757e677266dd13` |
| `read-table.log` | 4454 | `ab13054eb05752947556e6dbd3e9a5a7db4b08a264229e7d06e8ab11222ea324` |
| `read-table.py` | 869 | `68bc1eb5b665f57a59d2c463c0f4c0bf37b722699d3f6067be45b3d57e56f586` |
| `reads.json` | 4454 | `ab13054eb05752947556e6dbd3e9a5a7db4b08a264229e7d06e8ab11222ea324` |
| `require-result-contract-receipt.json` | 710 | `84722303466b8308f577b51fb4b616d4dce9f00f3ba7e33b08170dbe6e7a8561` |
| `require-result-contract.log` | 349 | `d9a9dd78d1e426d2fae3b37ef491df1d99d754fc53e1e738066596d22f213db0` |
| `resolve-host-contract-receipt.json` | 637 | `c1d214c91eda9ef02e5a54106ff011da88f29f9e6626ea3b129152ce25a0278d` |
| `resolve-host-contract.log` | 363 | `5ec5ff96a0ba2c60ed880335372e95a675bd711ef38aaac1f73ba5eb9f6eff60` |
| `restoring/src/compiler/core.ts` | 93355 | `6925a744b191c13ebe1425ec76a994b51cb2ab1c81f19b2c7958a4ed1e89311b` |
| `restoring/src/compiler/sys.ts` | 89088 | `a8a76f3bcca5ab8f670a7b89377531f85df078d96210ac0b429c5c3ca76c4669` |
| `restoring/src/harness/watchUtils.ts` | 9363 | `78b02ee9cb8a5198c0361b4bc4a40ecca5ebc1bdff368afa77176b9898eb8723` |
| `restoring/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts` | 57410 | `3467c9e18da5fd3563d0af02d2d4041e6dff63dbb3594f356fcb0ff02ee04c57` |
| `restoring-mutation.json` | 89 | `b1dc7a27b0a3952363b85198db7b584e542c776624ccb609727885f1c10d5ad6` |
| `restoring-receipt.json` | 756 | `398a6b33c76db0d9f04bd05a7927fe3881dafe580029def928aaae808931b1d7` |
| `restoring.jsonl` | 628767 | `800b87cc35b0589d97aa92dd22fe3f0831d5b547082c0ae752bf30e174d08b15` |
| `restoring.log` | 127 | `2db42c7d42c57f4c0ba953984f3b690362b016ed443170cacbcbaaed8096546c` |
| `run_receipt.py` | 826 | `a3235bfdbadd643cb9f1bd26cd35124906adb71a4cd091ee2224b8a2c0ea5d07` |
| `scope.cjs` | 2383 | `d253229bc380875a342e9505e34fbbcbdca9a536d53d65e644c0d1c0336412c2` |
| `setup.json` | 676 | `4c643b617b04f733bc01fb39cf59eea6b2b1233652b8ac60ebb3e3d784f5452f` |
| `stock-change-receipt.json` | 763 | `66d9c0155120aea0024dabde4cd8e90136a65110ed7425c55bdaa41b1978d253` |
| `stock-change.json` | 64 | `471dc276402b70481156d8842a952f2781d0541d42ede4c430f6069c19613b2c` |
| `stock-change.log` | 0 | `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| `stock-main.json` | 64 | `471dc276402b70481156d8842a952f2781d0541d42ede4c430f6069c19613b2c` |
| `stock.cjs` | 1051 | `69035e7593418400bd8ae82275129b488ef9a3a82a6ee0ccc09b8debbb8c69d4` |
| `submodules.log` | 696 | `0bcbfb90d9512fe6157d83dfbd083d6f28548aabc16ff00afefaa93895e4acc4` |
| `targets.json` | 194 | `bc55bb4a3089ccb3f23b1d61addcb9392373013bd4d2c7d36f6f8c6989a118ca` |
| `unnormalized/src/compiler/core.ts` | 93355 | `6925a744b191c13ebe1425ec76a994b51cb2ab1c81f19b2c7958a4ed1e89311b` |
| `unnormalized/src/compiler/sys.ts` | 89239 | `eb2bcb8bd1b063fb0e2251072bc1dab365ab2e9fceb4a22171036f739d768cb2` |
| `unnormalized/src/harness/watchUtils.ts` | 9363 | `78b02ee9cb8a5198c0361b4bc4a40ecca5ebc1bdff368afa77176b9898eb8723` |
| `unnormalized/src/testRunner/unittests/helpers/virtualFileSystemWithWatch.ts` | 57410 | `3467c9e18da5fd3563d0af02d2d4041e6dff63dbb3594f356fcb0ff02ee04c57` |
| `unnormalized-mutation.json` | 93 | `c4c89fd4b671c9cfa240b9a7cbb70e499f05af5d8f5c81e3bb425abea539ff51` |
| `unnormalized-receipt.json` | 766 | `2a06f5196d649530960bf6b5e30132646bffcd6f20d43575c2e52d6aadba5418` |
| `unnormalized.jsonl` | 611703 | `7e6480540c5bc3bf4795981a13cc2813475fabf6a999836f228db8ff0ff3e246` |
| `unnormalized.log` | 130 | `b7ade2b65fd8d6ff4681d06fd3d174643255ff31901a0c4fd0169eddb11f6f61` |
| `watch-directory-contract-receipt.json` | 711 | `ad727e54074710350cb40ee812a518d436f8e3ac151bdd76d640b567ece71266` |
| `watch-directory-contract.log` | 365 | `23056317afa84db5ff583351cd3c743272ae40e9956841f5a8cbbf89a69fc597` |

## Reproduction

Use the specified base in a detached scratch worktree, initialize its recursive
shallow submodules and install stage3/api using npm ci. The TypeScript source pin
is v6.0.3 / 050880ce59e30b356b686bd3144efe24f875ebc8. TMPDIR, GOTMPDIR, GOCACHE,
npm and Node compilation caches are scratch paths. The baseline uses a separate
immutable adapter worktree. Use the committed adaptation and contract/witness
scripts for the candidate; the scratch command receipts record exact argv,
environments, exits and durations. The census settings above are supplied on
every full run. Both lanes use the unchanged stage3/lane/run.sh with fresh result
directories. No new Go file is committed; measurement-only generated Go remains
outside repository stage3 in scratch.

The implementation and evidence refs are published without force. No PR, message
to another person, or Loom submission is created.
