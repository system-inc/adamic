# Watch-file interval recut acceptance

Adaptation commit: `5908a5bd665a246b134ea1df097706310cdb833a`. Parent exactly `13549c787e027a72afc4a9a6b0e10f07857abcfd`. Source `63f46490764862fef5af3ce6793557d4000994bc`. No merge with main and no PR or Loom submission. This supersedes the earlier watch-file-check-1532-3-stop report.

## What 93 --pin read

The original --pin path reads sys.ts from its TREE argument, extracts the predecessor owner, plans edits including constructWriter, parses that planned text, and hashes changedOwners[0].getText(changed). It does not construct ownerAfterText: {...manifest[0]} inherits that field from proofs/targets.json. There was no correct invocation capable of updating it. The tool was wrong for pins carrying exact text. A one-line fix adds ownerAfterText:changedOwners[0].getText(changed) to the generated pin. The commit message says so.

The fixed --pin ran on the real apply.py refs/stage3/93-builtin-imports-before snapshot, not the final adapted tree. Its exact after-text and hash match refs/stage3/93-builtin-imports-after, including patchedWriteFile and number | undefined. 99-wasm-crypto-import used its unmodified own --pin on the composed predecessor. All ten before/after proof snapshot trees equal the independent fresh full apply snapshots (logs/snapshot-match.json). No hash was typed by hand.

Core-multimap and timer-closures owners.json were byte-identical between the source commit's parent and the new parent; the actual step snapshots and their own adapters/contracts validate the carried source pins. These source-change paths remain byte-identical to 63f46490.

## Setup and full results

Detached scratch worktrees only. git submodule update --init --recursive --depth 1 completed. npm ci in stage3/api completed. ADAMIC_TYPESCRIPT_SOURCE was TypeScript v6.0.3 at 050880ce59e30b356b686bd3144efe24f875ebc8. TMPDIR, GOTMPDIR, Go cache and npm cache were scratch paths. Parent measurement used the evidence worktree; compiler/apply/adapter/API inputs there are byte-identical to 13549c78. The production compiler and census templates in the candidate are also identical to that parent.

| Check | Exit/result |
|---|---|
| Parent full apply | 0 |
| Fresh candidate full apply, empty product store | 0 |
| Parent full sys.ts census | 0; no diagnostics |
| Candidate full sys.ts census | 0; no diagnostics |
| Stock TypeScript 6.0.3, 79 roots, parent and candidate | 0 / 0; zero diagnostics |
| Production adamic types, 79 roots, parent and candidate | 0 / 0 |
| Actual dynamic boundary witness / guard-removal mutant | 0 / 1 |
| All five directory first applies and no-write replays | 0; bytes and nanosecond mtimes equal |

Censuses used LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 CENSUS_MODULE_FORMAT=CommonJSBundle and LATENT_ONLY_FILE selecting the full sys.ts, with all 79 roots loaded. Both have complete relation trailers and zero checker diagnostics. Pairing is the multiset (kind, phase, reason, message with position stripped). One Refused record disappears; zero new and zero changed. The original 1532 site is at 1537:13 on this parent. Raw counts and the exact gone finding are in census-comparison.json.

## Each directory command

Exit 1 is expected for a red control. Planting the core restoring mutant itself exits 0; the subsequent guard rejection exits 1. Embedded contract/binding mutants assert that every planted case is rejected and write nothing.

| Command | Actual exit | Expected exit |
|---|---|---|
| 93-builtin-imports-first-apply | 0 | 0 |
| 93-builtin-imports-replay | 0 | 0 |
| 99-correctness-core-multimap-first-apply | 0 | 0 |
| 99-correctness-core-multimap-replay | 0 | 0 |
| 99-correctness-single-watcher-forwarder-first-apply | 0 | 0 |
| 99-correctness-single-watcher-forwarder-replay | 0 | 0 |
| 99-correctness-timer-closures-first-apply | 0 | 0 |
| 99-correctness-timer-closures-replay | 0 | 0 |
| 99-wasm-crypto-import-first-apply | 0 | 0 |
| 99-wasm-crypto-import-replay | 0 | 0 |
| 93-legacy-check | 1 | 1 |
| 93-bindings-and-mutants | 0 | 0 |
| 93-owner-contract-and-mutants | 0 | 0 |
| 93-scope-mutant | 1 | 1 |
| core-contract-and-mutants | 0 | 0 |
| core-body-proof | 0 | 0 |
| core-runtime-witness | 0 | 0 |
| core-restoring-mutant | 0 | 0 |
| core-restoring-mutant-rejected | 1 | 1 |
| single-watcher-contract-and-mutants | 0 | 0 |
| timer-contract-and-mutants | 0 | 0 |
| timer-restored-slot-mutant | 1 | 1 |
| timer-witness-before | 0 | 0 |
| timer-witness-after | 0 | 0 |
| timer-request-mutant | 1 | 1 |
| timer-label-mutant | 1 | 1 |
| wasm-source-transpile | 0 | 0 |
| wasm-bindings-and-mutants | 0 | 0 |
| wasm-scope-mutant | 1 | 1 |
| wasm-ternary-mutant | 1 | 1 |
| wasm-require-result-contract | 0 | 0 |
| wasm-resolve-host-contract | 0 | 0 |
| wasm-watch-directory-contract | 0 | 0 |
| wasm-watch-directory-witness | 0 | 0 |

93 legacy check exits 1 with TestBuiltinSourceScope because it predates the expanded scope check fixed on main in f8ca4cc9; that fix is outside this parent, as anticipated in the task. The historical 93 scope-mutant exits 1 at its missing declaration target before planting; it is NOT credited as a mutant kill. The valid 93 owner-contract and binding suites pass their actual drift, changed-import, shadowing and bare-namespace mutants. No proof tool was changed in the adaptation commit.

The first optional boundary fixture attempt omitted its watchedFiles closure array; its retained attempt1 logs are fixture setup failures, not compiler failures. The corrected actual-body fixture passes and removing the guard fails by missing the expected rejection.

## Every path difference

The commit changes exactly the source commit's 13 paths. Seven carry over byte-identically. The six recut exceptions are:

- `stage3/adapt/93-builtin-imports/README.md`: Retain parent construction documentation, then original interval amendment.
- `stage3/adapt/93-builtin-imports/adapt.cjs`: Retain parent writer construction and --pin path, add original interval edits and whole-owner check, and fix --pin in one line to store its planned ownerAfterText.
- `stage3/adapt/93-builtin-imports/proofs/targets.json`: Regenerated by fixed own --pin at actual before-93 snapshot; exact after-93 text/hash equality verified.
- `stage3/adapt/99-wasm-crypto-import/README.md`: Retain parent construction pin documentation, then original interval amendment.
- `stage3/adapt/99-wasm-crypto-import/proofs/targets.json`: Regenerated by directory own --pin on composed predecessor, preserving writer plus interval and all earlier crypto/require/resolve/watchDirectory amendments.
- `stage3/upstream/LEDGER.md`: Exact parent bytes plus source commit original two-line suffix; zero parent lines removed.

The ledger is exact parent bytes followed by the source two-line suffix; zero parent lines removed. path-audit.json lists EVERY tree path differing from 63f46490, including all inherited main/parent changes. Every path outside the 13-path adaptation change was verified byte-identical to 13549c78.

drivers/ retains the exact scratch drivers and absolute invocation paths used, not a portable installer. logs/ and proofs/ contain raw command receipts, compressed logs and complete census streams. sources/ contains the parent, candidate and exact after-93 sys.ts snapshots. To reproduce, create new detached parent/source worktrees, install pinned inputs, apply this one commit, run fresh apply.py, and use its step refs with the directory scripts listed in proofs/results.json. Full upstream lane/counting/declaration-bundle builds were not requested in this recut and were not rerun.
