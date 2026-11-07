# getAllKeys proof

Base: origin/main e8fc6350f0378eeadb866c4a41ef8eaf9638f8e9. TypeScript 6.0.3: 050880ce59e30b356b686bd3144efe24f875ebc8. Checkpoint 0988e806b was plainly pushed to typescript/core-getallkeys-checkpoint-20261107 after both full applies passed; the requested final branch retains one commit on main, with no force push.

| Proof item | Result |
| --- | --- |
| Both fresh full applies, including product packaging | PASS, child exit 0, 94 baseline adapters and 95 adapted adapters |
| Paired replay of every adapter | PASS, all 94 shared outcomes identical after validated source-coordinate mapping; new adapter zero edits and zero writes |
| Owner contracts | PASS, 26 checks, including half trees, complete-owner drift, orphaned comments, no-write replay, LF and CRLF |
| Adamic types | PASS before and after, all 79 sorted compiler roots together, child exit 0 and zero diagnostics |
| Ordinary compiler and harness builds | PASS before and after, child exit 0 |
| typescript.d.ts and tsserverlibrary.d.ts | Byte-identical |
| typescript.internal.d.ts | Only getAllKeys and its internal JSDoc removed; declaration-hunk.diff shows the entire hunk |
| All emitted JavaScript | All 10 files checked; only run.js and typescript.js differ; every changed hunk is in javascript-hunks.diff, 6 and 4 respectively |
| Source and emitted function body | Byte-identical move from core to the module-local harness function |
| Complete source scope | All 746 files checked; exactly core.ts and fourslashImpl.ts differ by the isolated adapter's transformation |
| Whole lane in pieces | PASS, all 398 units in 10 bounded pieces; 106,366 passing, 1 sanctioned API failure, 0 skips |
| Exact unit guards | All selected test identities, multiplicities, counts, failure titles and baseline guards pass; source, build and oracle input hashes unchanged |
| Full core.ts site census | 24 before, 23 after; exactly F10 removed, zero new sites; zero checker diagnostics |
| Original-core mutant | Restoring only the core owner restores F10 and the complete original 24-site set |

The named pushIfUnique import preserves the original function body. Esbuild emits one extra adjacent init_ts8() call in the harness initializer for this second import. compare.cjs pins the complete once-only __esm helper, verifies the owning initializer and both adjacent calls, and accounts for exactly that added call. It also removes only the complete getAllKeys declaration, its exact export entries and its dead module.exports annotation entry before comparing every remaining parser-derived runtime token. No identifier renaming or blanket hunk normalization is used. The helper leaves all compiler bundles and remains local in run.js.

Paired replay preserves all raw outputs, actual exits, changed-file sets and write records. The common completed-tree refusals are identical on both sides; these are not failures of the successful incremental full applies. replay-summary.json retains each coordinate mapping, which must point to an identical complete source line. proof-results.json records the replay exit counts.

The census follows typescript/next-stops-after-wave, ba90529b3d16768a886a7570c47a0d1c07b06a32, core-sites-af42d3dc/run-census.py: LATENT_FULL=1, LATENT_ONLY_FILE set to the complete core.ts path, LATENT_ASSERT_NO_OUTPUT=1, LATENT_TIMES set, with no speculative or replay flags. Both ordinary Load and Lower no-output assertions execute. The exact reference procedure and report are packed. The current main meter tools generate the overlay; only physical scratch Go filenames gain .txt. The report excludes recorder events with kind Boundary, exactly as the cited analyzer does. Raw streams retain every event and their legacy labels. Absence of F10 is not a claim of successful whole-program lowering.

Setup followed 99-correctness-axolotl-w/PROOF.md: source /workspace/adamic-tools/env.sh, bash cloud/setup.sh, npm ci --prefix stage3/api, export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source. The locked upstream dependencies were installed with npm ci --prefix stage3/typescript. Setup, installs, both fresh full applies, checker commands, ordinary builds and final proof commands exited zero. All output went to files. supervise.py retains actual child exits independently from stopping reasons and reaps each process group.

The first baseline attempt failed only product packaging because a build check began before tar finished. Both full applies were rerun without overlapping builds. Two comparison-script corrections, the Boundary event filter and template-token handling, are retained with their original failing logs. They did not change the adapter, raw census streams or ordinary compiler outputs. No failed or excluded attempt is counted as accepted proof.

Prior reach counts, not a new instrumented measurement: 6,566 calls, all from fourslashImpl.ts:461, and zero namespace reads, from f9d57ee2d8213572b02748c323b43b85d48d6a37. The accepted lane ran ordinary JavaScript with no call observer. All actual command records, raw streams, timings, paired replay results, lane reports, source manifests, owner snapshots, reference procedures and .go.txt overlays are packed in proof-evidence.tar.gz with a verified internal SHA256SUMS and external proof-evidence.sha256. EVIDENCE.json indexes the pack. Binary hashes are in provenance.json. Later evidence files change product keys without changing the measured transformation.

Reproduction after the documented setup: fresh full --build-product applies from isolated main and this branch; verify.py; replay.py; lane-pieces.py; finalize.py; archive.py. The drivers retain exact paths and commands. No compiler, other adapter, refusal, test expectation or test selection is changed.
