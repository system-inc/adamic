# 95a owner-pin repair evidence

Code branch: typescript/axolotl-checker-a-pin, commit
`d12816b9dc1d425149f5f797c74d31f6c81f4d65`.
Main base: `80c2d3282e44b527555d59690bb413cc883d2bf2`.
This separate branch keeps new proof artifacts out of the adaptation directory.

- Full apply: exit 0, 213 seconds; 95a row 1 file / 13 lines added / 13 removed.
- Main-tip tree with only 95a's owner restored to before: own-file 21, whole 67.
- Repaired 95a: own-file 8, whole 54. Exactly the historical 13 rows removed;
  no diagnostic added anywhere. Root and aggregate observations agree and count once.
- Exactly 13 insertions; complete checker bytes match the full main apply output.
- Independent verifier: 711 source files, one changed file, CRLF 54,434 unchanged.
- Stock JavaScript identical: 2,892,419 bytes with comments; 2,354,716 without.
- Idempotence: zero files, zero assertions.
- Outside-owner mutant: widen getMergedSymbol's temporary to Symbol | undefined;
  all 13 assertions apply, unrelated edit retained; rerun zero edits.
- Inside-owner mutant: final return false changed to return true; exit 1 for
  uninventoried owner, zero writes. Existing one-read and emitter mutant controls
  also pass with the updated owner-guard catcher.
- Setup: 57.936 seconds; nproc 5. No full oracle, native gate or Loom submission.

proof.json contains the exact structural/mutant/byte observations; census.json
contains every removed diagnostic and the empty additions list. The gzip files
retain raw census observations. logs.tar.gz contains the complete apply, setup,
verification and mutant command outputs.

Commands ran from /workspace/adamic with its environment sourced:

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /tmp/axolotl-a-pin-setup.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/axolotl-upstream bash stage3/apply.sh /workspace/scratch/axolotl-a-pin-applied > /tmp/axolotl-a-pin-apply.log 2>&1
go build -o /tmp/axolotl-a-pin-census ./stage3/census/tool > /tmp/axolotl-a-pin-census-build.log 2>&1
NODE_PATH=$PWD/stage3/api/node_modules node PROOF_DIRECTORY/prove.cjs > /tmp/axolotl-a-pin-prove.log 2>&1
/tmp/axolotl-a-pin-census /workspace/scratch/axolotl-a-pin-before/src/tsc/tsc.ts PROOF_DIRECTORY/before.jsonl > /tmp/axolotl-a-pin-before-census.log 2>&1
/tmp/axolotl-a-pin-census /workspace/scratch/axolotl-a-pin-after/src/tsc/tsc.ts PROOF_DIRECTORY/after.jsonl > /tmp/axolotl-a-pin-after-census.log 2>&1
python3 PROOF_DIRECTORY/compare-census.py PROOF_DIRECTORY/before.jsonl PROOF_DIRECTORY/after.jsonl stage3/adapt/95-checker-axolotl-a/proofs/census.json > PROOF_DIRECTORY/census.json
```

prove.cjs uses the retained fresh apply tree and new scratch destinations; use
fresh paths when replaying. It restores just the owner's before text from the
apply Git snapshot while retaining 95b and all later main adaptations.
