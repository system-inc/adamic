# Acceptance proof: construction and harness move

Base: `a77ca4a82271f18c142d8843ea76604b30d3833d`.
Newest main, fetched again before final proof:
`89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825`.
TypeScript v6.0.3: `050880ce59e30b356b686bd3144efe24f875ebc8`.
Private worktrees only; scratch owns TMPDIR, GOTMPDIR and the Go cache.

The control is the requested parent merged with that main, preserving 1965,
1533 and the three receiver adaptations. Comparing against main alone would
attribute those parent changes to this adaptation.

Accepted shape and reader audit: scout `de29fb07566f3eb82a1895fe6561ce13e61f5139`,
`stage3/census/write-file-slot-scout/REPORT.md`, road (a′).
Fresh acceptance evidence: [`6b362fe55551e862c3d8775047e0e2dc41317a45`](https://github.com/system-inc/adamic/tree/6b362fe55551e862c3d8775047e0e2dc41317a45/stage3/census/write-file-slot-scout/construct-1), `stage3/census/write-file-slot-scout/construct-1/`.

| Check | Result |
|---|---|
| Fresh apply, frozen implementation merged with main | exit 0; complete normal adaptation sequence; empty product cache |
| All 79 production roots, CommonJSBundle | 73 move from Refused sys.ts:1387:5 to NotYet sys.ts:1481:57; other six unchanged; complete per-root table in evidence |
| Production tsc entry | NotYet sys.ts:1481:57, node:timers.clearTimeout parameter-name/slot boundary |
| Full single-file sys.ts census | 51 → 50 unique findings, exactly 1 gone / 0 new / 0 changed; 136 → 135 units; 176 → 174 raw records |
| Gone unique finding | Refused / lowering / inherited library member call read as an own field, old sys.ts:1392:49 |
| Stock tsc 6.0.3 | exit 0 over all 79 roots |
| Production adamic types | exit 0 over all 79 roots |
| Public declaration bundle | stripInternal emit + unmodified upstream bundler; byte-identical control/candidate, SHA-256 f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516 |
| Stock public API baseline comparison | inherited difference unchanged; zero additional declaration delta; System remains a public method |
| Real Node wrapper witness | 29 matching events: BOM, directory checks/creation, filesystem receivers, live directory reads and later slot replacement |
| Real harness route | 10 matching events through unchanged createServerHost and writeFile bodies; moved helper body byte-identical |
| Full restoring mutant | exit 0; all 747 source files exactly reproduce parent+main control; replay exit 0, no writes; production entry returns Refused at 1387:5 |
| Whole unmodified lane, once | PASS: 106,366 passing, 0 pending, only the sanctioned API baseline row; lane exit 0, raw oracle exit 1 as expected |
| Ledger | exact parent bytes then appended line; 0 parent lines removed |

Census uses `LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1`, the complete sys.ts via
`LATENT_ONLY_FILE`, and `CENSUS_MODULE_FORMAT=CommonJSBundle`; both streams finish
with a relation trailer and have zero diagnostics, checker_rejected:false.
Pairing is by (kind, phase, reason, message with positions stripped), as a
multiset. Removing the helper removes its Refused finding and a Boundary raw
record. No inherited finding is counted as new merely because its position moves.
The first getNodeSystem refusal-scan stop is 1481:57; the removed helper has no
remaining unit. No originalWriteFile.call is introduced on the Node path.
This does not claim that the whole compiler now lowers successfully.

The observable changes are priced explicitly in LEDGER.md: wrapper `.name`
`""` → `"patchedWriteFile"` (zero name/length readers in the 479-access scout),
the internal helper disappears from `ts` (zero adapted runtime reads, 27 firing
controls), the raw writer never occupies the slot, and the direct call has
receiver undefined (raw body reads no this, Node path reads no .call). The own-call
control observes 5 reads before and 0 after. Actual rebuilt exports change
2256 → 2255, losing only the helper and preserving remaining key order. Wrapper
length stays 3 and descriptor flags are unchanged. Internal declarations lose
the same internal helper; public declarations do not change.

## Owning directories' checks and mutants

| Directory | Check | Exit / observation |
|---|---|---|
| 93-builtin-imports | whole-owner contract | 0; before/after drift, foreign edits, replay, bytes/mtimes and import-anchor guards |
| 93-builtin-imports | bindings and built witness | 0 each; binding mutants caught, built CLI outputs equal |
| 93-builtin-imports | legacy check | 1 on main's own exact 93-before/after snapshots, also 1 here; it still assumes only three imports, excluding its previously landed timer forwarders |
| 93-builtin-imports | legacy scope mutant | 1 here and on main: stale ambient clearTimeout fixture; not claimed as a successful new mutation test; owner-contract and binding mutants above remain executable |
| 99-wasm-crypto-import | source/transpile scope, bindings, built witness | 0 each; construction exists on both sides of this owner's scope comparison |
| 99-wasm-crypto-import | require-result, resolve-host, watch-directory contracts and watch-directory witness | 0 each; LF/CRLF, exact forwarders, no-write replay and malformed-state refusals |
| 99-wasm-crypto-import | scope mutant | 1 at source guard after both JS comparisons pass; stale fixture updated to current System signature |
| 99-wasm-crypto-import | ternary mutant | 1 at semantic binding guard |
| 99-wasm-crypto-import | restoring mutants; watch unnormalized/optional-contract variants | mutation exits 0; adapter refusal exits 1; parent 1533 behavior remains present |
| 99-correctness-write-file-view | contract, moved real-body witness, CLI mutant/replay/remove | 0 each; mutant restores the already landed readonly capture, never the original pre-b8ffarm read |
| 99-correctness-write-file-construct | embedded contract, restoring mutant and replay | 0 each; LF/CRLF, atomic refusal, helper/class identity and full-source exact restoration |

All four touched directories regenerate pins through their own `adapt.cjs --pin`
commands, with sourceWrites:0. Pin continuity receipts retain the original
b8ffarm before and captured texts/hashes. The new directory owns only the unpinned
initializer statement and harness import/local helper; the TestServerHost class
pin in timer-closures is unchanged. See README.md and the evidence ownership audit.

The executable adaptation was frozen at `51104a74ba9e981fe93404e21e9c66298e47425f` before final fresh application.
The final source identity receipt compares all 747 fresh output files against the
once-tested lane tree. Later amendment completes documentation/citations only;
code, pins and ledger remain identical to that frozen commit. Evidence contains
all commands, exits, complete census streams, every production-root result,
source/JavaScript/bundle diffs, public bundles and the unchanged lane verdict.
Nothing is submitted.
