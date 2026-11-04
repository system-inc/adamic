# 99j-4: composition and landing lane pass

All checks in the rerun table ran on WebAssembly's scoped-pin commit `b434890646453e5fc0895628169f10f36f253cd9` plus the 99j cherry-pick. Its parent is main `4b8634a4a498a39d254ee8c325eb2711f84359d9`. The measured cherry-pick commit is `dc07dfe2dde26e7ea03afc3f3ff65532a60931d0`; the final amendment updates evidence only. The measured 99j adapter source is unchanged from 0854785b and is pinned by SHA-256 in proof-results.json. No WebAssembly file is changed by the 99j commit.

| Check rerun | Result |
| --- | --- |
| Full stage3/apply.sh | PASS: all 69 adapters, all 138 before/after refs and patch rows present; 99j and then 99-wasm-crypto-import both apply |
| Second full apply | PASS: cached replay is byte-identical, excluding Git metadata |
| Both adapters on completed tree | PASS: 99j has zero files and edits, WebAssembly has zero files, and zero writes occur; bytes and mtimes are unchanged |
| Full landing lane oracle | PASS: 106,366 passing, one sanctioned failure, zero pending |
| Exact oracle identities | Complete pass/fail identity multisets, including multiplicities, equal the retained main reference |
| Oracle source | All src file bytes match the full-apply tree |
| Public declarations | Built typescript.d.ts and local API declaration bytes match main exactly |
| Shard check | python3 -B stage3/lane/generate_shards.py --check passes |
| Go lane fixtures | go test -count=1 -skip 'TestShard_|TestProduct_' ./stage3/lane/ passes, with GOMAXPROCS=1 |

The first oracle attempt had one additional failure: fourslash codeFixClassImplementInterfaceNoTruncation.ts exceeded the unchanged 40-second timeout (106,365 passes, two failures including the sanctioned one). Its complete report is preserved in proof-results.json. The failed run source copy was removed, preserving its logs and report, clean scratch cache pages were released again, and the complete unmodified oracle was rerun. The retry passed with exact main identities. No timeout, worker count, test selection or expectation was changed.

The oracle ran all TypeScript runners with eight workers, no filter and the unchanged 40-second timeout and expected counts. The sole common failure is `unittests:: Public APIs for typescript.d.ts should be acknowledged when they change`. A read-only Mocha host-replay observer recorded the exact identities; it did not select tests or change outcomes. The retained main reference uses the same pinned TypeScript 6.0.3 test suite and platform. The oracle ran on a byte-identical copy of the same keyed full-apply product; its source hashes were checked against that product afterward.

Before the oracle, verified replay and earlier failed/superseded source copies were removed, preserving logs and reports. Clean scratch file-cache pages were released with POSIX_FADV_DONTNEED to avoid the resource-pressure timeouts seen in older attempts. No test setting or expected result was altered. These operations and their outputs are recorded in proof-results.json.

The previous whole-file WebAssembly refusal at base 4b8634a4 is resolved by b4348906. Its historical report is [prior-composition-stop.json](prior-composition-stop.json), with [prior-apply-stop.log.gz](prior-apply-stop.log.gz). Current successful full-apply output is [apply.log.gz](apply.log.gz). Exact current commands, exits, elapsed times, identity results, lane report and declaration hashes are in [proof-results.json](proof-results.json). Raw logs are in /workspace/scratch with the 99j4 prefix.

## Unchanged and historical evidence

The fourteen transformations remain four narrowings and ten checks, with three sites left. Nine enclosing-function pins retain unique owners, lexical ancestry, exact statements, parse checks, after-owner validation, rereads before writes and partial-tree refusal. [contract.cjs](contract.cjs) and [contract-results.json](contract-results.json) retain the unchanged 99j-3 scoped-pin fixtures: twenty refusals, unrelated sys.ts code accepted and preserved, and direct no-op behavior. Those fixtures were not rerun here; real composition and both adapters' no-write checks above were rerun.

The 79-root checker and latent census were not rerun for 99j-4. [CENSUS.md](CENSUS.md), [census-delta.json](census-delta.json) and [prior-proof-results.json](prior-proof-results.json) remain explicitly historical 99j-2 evidence. The old derived census is fourteen removed sites and zero new sites on compiler 5d521b90, not current-base census totals. The README retains all seventeen decisions, both array witnesses and the null-path result "..". No public declaration or other adaptation is changed.
