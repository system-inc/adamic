# Receiver parser

JavaScript-changing, value-identical; #trxxkge: 46 parser callee rewrites call through factory, measured at 18,022,708 added per-use property reads per complete lane run.

Generated from the accepted split-codemod.cjs on the exact pinned base. owners.json stores whole before/after owners and hashes; adapt.cjs validates all files before writing and refuses drift, duplicates, half adaptation and mixed endings. Replay writes zero files. mutant.cjs restores the baseline.

Run from Adamic with `NODE_PATH=stage3/api/node_modules node stage3/adapt/99-correctness-receiver-parser/adapt.cjs TREE`; use mutant.cjs with the same environment to restore. The stock 6.0.3 parser is supplied automatically by stage3/apply.sh.

See [PROOF.md](PROOF.md) and [stage3/census/receiver-split-evidence/parser/proofs/](https://github.com/system-inc/adamic/tree/42f0c552bb93c523b6c5eb8b4681719cf0bc3197/stage3/census/receiver-split-evidence/parser/proofs) at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197` for all checks, observations, scope decisions and receipts. The typescript/receiver-split-evidence branch retains shared summaries and raw artifact manifests. No Loom submission.

Re-cut base: `c2f3c21f72b04e4c3acf7731913f0fc8ab89c2ca`. Runtime files and pins are byte-identical to the accepted first head. Evidence lives at `stage3/census/receiver-split-evidence/parser` on `typescript/receiver-split-evidence` at `42f0c552bb93c523b6c5eb8b4681719cf0bc3197`. This directory contains only the five house files; moved JSON is not a runtime dependency. New-stack verification receipts are published forward on the same evidence branch.
