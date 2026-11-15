# Receiver split: final proof

All three independent adaptations are complete. No Loom submission.

| Branch | One-commit SHA |
| --- | --- |
| typescript/receiver-parser-1 | `3a5dbe2d025478c588c9c5bbc96b4829a81596bf` |
| typescript/receiver-transformers-1 | `2dbef8ed2a76a32f4be306b9885cc0797e6e43b5` |
| typescript/receiver-printer-1 | `457a0447edd202b88af3d86243b5958a045d5bb6` |

Each adaptation commit has sole parent 8ada94542fbdcb5c1d1689e27564e471eae241b8 and contains its own directory plus its own stage3/upstream/LEDGER.md entry citing #trxxkge. The evidence branch contains checkpoint history and final receipts.

## Generation and guards

A fresh detached scratch worktree, recursive shallow submodules, locked stage3/api install and shallow upstream v6.0.3 clone (050880ce59e30b356b686bd3144efe24f875ebc8) were used. The unmodified accepted split-codemod.cjs at 24c61e74d26caa79c0f313112902e325d090afcd regenerated all 19 files: the complete receipt equals the accepted one, and --verify reproduced every generated file byte-for-byte. No new safety-scan failure occurred. Replacement source was never hand-edited.

Whole enclosing owners, changed imports and the 34 nearest parser function owners (full overload groups) are pinned before/after with text and sha256. Required prior greps found no existing whole-owner amendment. Adapters pass drift, duplicates, partial/half-tree, mixed-ending and zero-write replay controls; real restoring mutants restore all 79 baseline compiler files byte-for-byte.

## Proof results

- All 38 complete single-file CommonJSBundle runs completed with zero checker diagnostics and complete trailers. Every file pairs by (kind, phase, reason, position-stripped message), with 0 new and 0 changed outside targets. Exact pairing is published in each directory's proofs/census.json. The initial printer multiplicity attribution conservatively stopped; unchanged-source mapping subsequently proves the two removals are exactly line 1215, with the pairing key unchanged. Both raw attempts remain hashed. Seven historically hidden transformer sites have unchanged finding counts; AST inventories independently prove their bindings removed, without claiming their bodies lowered.
- Each independent tree: stock tsc 6.0.3 API 0 diagnostics over 79 roots, CLI exit 0, adamic types exit 0, fresh single-part stage3/apply.sh exit 0. All 79 applied compiler files equal the independent tree.
- Complete stock emitted JavaScript comparisons are identical outside the rewritten sites. hereby dts declarations are byte-identical: typescript.d.ts 593536 bytes, sha256 f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516; tsserverlibrary.d.ts 866 bytes, sha256 cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd.
- All six unfiltered whole lanes (ordinary and counting for each part): 106366 passing, 1 sanctioned Public APIs/typescript.d.ts acknowledgement row, 0 pending. Sole baseline diff api/typescript.d.ts. All runners, eight workers, unchanged 40000 ms hook; install and build exit 0. No timeout, exception or other failure.
- Each counted lane: 0 observed post-creation setter/assignment/defineProperty writes; 0 direct member identity and length reads, including forwarding callees; 0 new-arrow length reads; 0 implementation this reads by AST, 0 parse failures. Positive controls prove actual instrumentation observations and deliberate this detection. Broad unrelated shared-function bind length and preserved printer sentinel identity observations are reported separately; no blanket zero is claimed. Counting copies preserve all cases/reference inputs.
- Printer controls preserve getter receiver/order, once-only aliases/default resolution, bare callback this, sentinels and printed output; both defaults separately have 0 AST this reads.

Parser: 46 rewritten calls, 18022708 per-use property reads. Transformers: 189211 forwarding calls and 1601433 arrow creations per lane. All 17 destructures and 52 arrows are directly at the original top-level factory scope; no per-node site is eligible for lifting. scope.json reports every site's count before/after the unchanged scope decision. Printer: 739160 alias reads over 73916 entries. All three ledger lines classify JavaScript-changing and value-identical.

Each source directory's PROOF.md and proofs/ contain the full per-part summaries and command receipts. final-proof-status.json, ledger.json and source-commits.json index the result. Raw files remain in /workspace/scratch/receiver-split; raw-manifest-final.json records their paths, byte sizes and sha256. No Go source was added; one heavy job ran at a time, with scratch Go/npm/temp caches.

During final publication, three result/receipt filename collisions were found. Tiny witness checks were rerun into distinct filenames; published witness JSON contains results, while original command receipts and scratch artifacts remain preserved. All assertions passed.

## Second cuts and merged-stack verification

The first-head results above remain historical measured evidence. See [stack-v2/REPORT.md](stack-v2/REPORT.md) for the new parent chain, exact evidence relocation, current pushed SHAs and new-stack proof. No pins were regenerated; lanes/censuses were not rerun.
