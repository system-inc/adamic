# Receiver split, second cuts

Complete; pushed under new names without force. No Loom submission.

| Branch | Commit | Sole parent |
| --- | --- | --- |
| typescript/receiver-parser-2 | `7bc402ece2dda307e451112c8a735d516f9b6720` | `c2f3c21f72b04e4c3acf7731913f0fc8ab89c2ca` |
| typescript/receiver-transformers-2 | `87b09e9ce4971ef9afa2dac5938a213f7af977c2` | `7bc402ece2dda307e451112c8a735d516f9b6720` |
| typescript/receiver-printer-2 | `a77ca4a82271f18c142d8843ea76604b30d3833d` | `87b09e9ce4971ef9afa2dac5938a213f7af977c2` |

Evidence relocation anchor: 42f0c552bb93c523b6c5eb8b4681719cf0bc3197 on typescript/receiver-split-evidence. All 57 moved proofs/*.json and scope.json files remain byte-identical to their first-head Git blobs, under parser/, transformers/, printer/. relocation-byte-manifest.json records original commits, paths, sizes and sha256. adapt.cjs reads only owners.json locally; mutant.cjs calls adapt.cjs; no moved file is a runtime dependency. Each new adaptation directory contains only adapt.cjs, owners.json, mutant.cjs, README.md and PROOF.md. The docs and commit messages cite the relocation evidence path and anchor SHA.

Each ledger is the parent's exact bytes followed by exactly two added lines (blank line plus that part's original entry), with zero parent bytes or lines removed. Parser is one commit on c2f3c21f72b04e4c3acf7731913f0fc8ab89c2ca; transformers and printer each follow the prior new cut. Runtime adapters, mutants and all owner pins are byte-identical to the first heads. No owner pin was regenerated.

## New-stack proof

Printer merged with exact main fe5a1a54f045e9bce4e5aa32ca74885a317fbebd in isolated scratch produced verification merge `fb36b3ab3b8b10af36ab8cf9a39bde34b6d986ce`. The five merge conflicts were earlier copies of the crypto/watchDirectory adaptation and ledger: the requested chain's newer bytes were preserved; all clean main changes, including its native/watchFile changes, merged normally. merge-resolution.json identifies each blob and decision. No replacement source was hand-edited.

A fresh stage3/apply.sh with all three installed exited 0, validating every unchanged owner pin on this merged base. Each actual mutant restores every one of its edited files byte-for-byte to the apply pipeline's own before snapshot; all other compiler files stay unchanged. Mutant replay writes 0; each adapter replay writes 0; reapplication restores all 79 complete files byte-for-byte.

The first stock API attempt ran before the fresh adapted checkout had its dependencies, producing 65 missing-Node-type diagnostics. Its failed receipt and witness are preserved. A locked npm ci --ignore-scripts then supplied the required definitions; the successful stock-api-command-installed and stock-api-installed-witness are the final checker evidence. No compiler source changed for this correction.

Stock TypeScript 6.0.3: API 0 diagnostics over the explicit 79 roots, CLI exit 0 with strictFunctionTypes/noEmit and incremental/composite false. Adamic was rebuilt from the verification merge; adamic types exits 0 over those same 79 roots. Command receipts, stock inventory, whole-source hashes and per-part mutant witnesses are published alongside this report. Raw logs stay in /workspace/scratch/receiver-split-v2 with byte sizes and sha256 in raw-manifest.json. Go/npm/temp caches are scratch-local; one heavy job ran at a time.

Rerun: locked stage3/api install, fresh full apply, stock API and CLI, Adamic build and types, all three actual mutants, mutant replays, adapter replays and exact-byte reapplications. Lanes and full single-file censuses were not rerun: no pins or replacement source changed. Their first-head proofs and every per-site transformer scope/count row are retained byte-identically at the relocation anchor. The 17 sites remain at their factory scopes, with no eligible lift.
