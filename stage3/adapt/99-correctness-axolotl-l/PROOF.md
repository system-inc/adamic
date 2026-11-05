# 99l proof

All required checks passed. Exact type and census commands and exits are in
measurements.json. Compiler build provenance and hashes are in
compiler-provenance.json. Full logs are retained as .gz files.

| Check | Result |
| --- | --- |
| Prescribed timeout-wrapped cloud/setup.sh | Exit 0 |
| Full stage3/apply.sh --build-product | Exit 0 |
| Ordinary keyed second apply to fresh output | Exit 0; all source bytes match |
| Second direct adapter run | 0 files, 0 edits, 0 writes; bytes and mtime preserved |
| Stock TypeScript 6.0.3 parsing and normal compiler builds | Pass before and after |
| Adamic types over complete sorted 79-root inputs | 0 diagnostics before, 0 after; both exit 0 |
| Both public declaration files | Byte-identical |
| core.ts JavaScript, comments kept and removed | Byte-identical |
| All 9 emitted JavaScript files from normal compiler builds | Byte-identical |
| Complete adapted src comparison | 746 files; only the one core.ts annotation differs |
| generate_shards.py --check | Exit 0 |
| Stage3 lane Go fixtures | Exit 0, serial scheduling |
| Owner contract refusals | 7 no-write refusals; unrelated edits and CRLF/LF preserved |
| Ordinary per-file compiler census | All 79 roots measured before and after; 73 move past core.ts:381 |

The lane oracle was skipped because the adapted tree's JavaScript is
byte-identical: all 9 emitted JavaScript files match byte-for-byte. Also,
all src files other than core.ts are byte-identical, and the
complete core.ts transpile is identical with comments kept and removed.
comparison.json records hashes, sizes, both public declarations, and the
replay source comparison. No JavaScript behavior changes.

The lane fixture command is 99k's `go test -count=1 -skip
'TestShard_|TestProduct_' ./stage3/lane/` with `-parallel=1`. The first attempt
omitted the required upstream input variable. A corrected concurrent run hit
two existing Python wrapper deadlines while compiler builds were running.
The serial retry passed without changing any timeout or test selection.
Both earlier logs are retained.

Preparation replaced an oversized incomplete nested checker clone with a
shallow checkout at the same pinned d92d9bfe commit. The prescribed setup
command completed successfully. No compiler sources or dependency pins changed.

Reproduce contract and comparisons with NODE_PATH set to stage3/api/node_modules:

```
node stage3/adapt/99-correctness-axolotl-l/contract.cjs BEFORE_TREE
node stage3/adapt/99-correctness-axolotl-l/compare.cjs BEFORE_TREE AFTER_TREE REPLAY_TREE
```

The adapter's complete before/after owner text and owner-scoped SHA-256 pins
are embedded in adapt.cjs. sort.json records the real Python directory sort:
99l immediately follows 99k. See CENSUS.md for the new first-stop table.
