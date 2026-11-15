# Write-file construction amendment

Accepted evidence: `typescript/write-file-slot-scout` at
`de29fb07566f3eb82a1895fe6561ce13e61f5139`,
`stage3/census/write-file-slot-scout/REPORT.md`, road (a′).

This amendment is JavaScript-changing. The ensuring wrapper is constructed before
`nodeSystem`, calls the hoisted local writer directly, and reads directory methods
live through `nodeSystem`. Its name changes from `""` to `"patchedWriteFile"`
(0 name/length readers in the 479-access audit). The raw writer never occupies the
slot. The direct call's receiver is undefined; the audited writer body reads no
`this`, and the Node path reads no `.call` property. Removing the internal export
removes `ts.patchWriteFileEnsuringDirectory` (0 adapted reads, 27 firing controls).
The harness keeps its original helper body and live receiver-preserving capture.
System's public method declaration does not change.

Ownership is split across the existing whole-owner pins: 93-builtin-imports owns
construction; 99-wasm-crypto-import regenerates its downstream getNodeSystem pin
with the parent 1533 forwarding intact; 99-correctness-write-file-view owns removal
of the whole helper. This directory owns only the unpinned initializer patch
statement and harness import/module-local helper. TestServerHost and its
createServerHost method remain byte-identical; the timer-closures pin is unchanged.

`node adapt.cjs --pin TREE` regenerates this directory's pins without source writes.
`node adapt.cjs TREE` applies the two unpinned edits. `node adapt.cjs --proof TREE`
checks LF/CRLF, no-write replay, refused drift without writes, class identity and
helper identity. Supply the predecessor tree for the proof.
`node mutant.cjs TREE` restores this construction change only, including the
landed readonly helper and pre-construction slot, while preserving parent edits.

Acceptance receipts are recorded in PROOF.md; no evidence files belong here.

## Touched directories

| Directory | Why |
|---|---|
| 93-builtin-imports | Construction belongs to its whole getNodeSystem pin. |
| 99-wasm-crypto-import | Compose its whole getNodeSystem pin with construction and retain 1533. |
| 99-correctness-write-file-view | Its whole helper pin owns deleting the sys.ts declaration. |
| 99-correctness-write-file-construct | Own the unpinned initializer/import/local helper edits and ledger line. |

Fresh acceptance: [PROOF.md](PROOF.md). Evidence: [`6b362fe55551e862c3d8775047e0e2dc41317a45`](https://github.com/system-inc/adamic/tree/6b362fe55551e862c3d8775047e0e2dc41317a45/stage3/census/write-file-slot-scout/construct-1).
All 79 production outcomes, the full census pair, source/JavaScript diffs, the
once-run lane and every owning proof/mutant exit are preserved there.
